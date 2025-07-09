package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
)

const (
	ChunkSize = 128
)

type AsepriteFile struct {
	Header Header
	Frames []Frame
}

type Loader struct {
	Reader io.Reader
	Buf    []byte
	Buffer *bytes.Buffer
	File   *AsepriteFile
}

const HeaderSize = 128

type Header struct {
	FileSize     uint32
	MagicNumber  uint16
	Frames       uint16
	Width        uint16
	Height       uint16
	ColorDepth   uint16
	Flags        uint32
	FrameSpeed   uint16 // deprecated
	_            [2]uint32
	PaletteEntry byte
	_            [3]byte
	NumberColors uint16
	PixelWidth   byte
	PixelHeight  byte
	GridX        int16
	GridY        int16
	GridWidth    uint16
	GridHeight   uint16
	_            [84]byte
}

type Frame struct {
	Header FrameHeader
	Chunks []Chunk
}

const FrameHeaderSize = 16

type FrameHeader struct {
	FrameBytes     uint32
	MagicNumber    uint16
	OldChunkNumber uint16 // deprecated?
	FrameDuration  uint16
	_              [2]byte
	ChunkNumber    uint32
}

type ChunkDataType uint16

const (
	OldPaletteChunk      ChunkDataType = 0x0004
	OldPaletteChunk2     ChunkDataType = 0x0011
	LayerChunk           ChunkDataType = 0x2004
	CelChunk             ChunkDataType = 0x2005
	CelExtraChunk        ChunkDataType = 0x2006
	ColorProfileChunkHex ChunkDataType = 0x2007
	ExternalFilesChunk   ChunkDataType = 0x2008
	MaskChunk            ChunkDataType = 0x2016 // DEPRECATED
	PathChunk            ChunkDataType = 0x2017 // NEVER USED
	TagsChunk            ChunkDataType = 0x2018
	PaletteChunk         ChunkDataType = 0x2019
	UserDataChunk        ChunkDataType = 0x2020
	SliceChunk           ChunkDataType = 0x2022
	TilesetChunk         ChunkDataType = 0x2023
)

const ChunkHeaderSize = 6

type ChunkHeader struct {
	Size uint32
	Type ChunkDataType
}

type Chunk interface {
	GetHeader() ChunkHeader
	GetType() ChunkDataType
}

// type OldpalettechunkData struct {}
//
// type Oldpalettechunk2Data struct {}
//
// type LayerChunkData struct {}

type ColorProfileType uint16

const (
	ColorProfileNone ColorProfileType = iota
	ColorProfileSRGB
	ColorProfileICC
)

type ChunkColorProfile struct {
	header ChunkHeader
	ChunkColorProfileData
}

type ChunkColorProfileICC struct {
	ChunkColorProfile
	ChunkColorProfileICCData
}

func (c *ChunkColorProfile) GetHeader() ChunkHeader {
	return c.header
}

func (c *ChunkColorProfile) GetType() ChunkDataType {
	return c.header.Type
}

const ChunkColorProfileDataSize = 16

// TODO: criar uma struct Fixed
type ChunkColorProfileData struct {
	Type  ColorProfileType
	Flags uint16
	Gamma int32 // TODO: adjust to FIXED type
	_     [8]byte
}

type ChunkColorProfileICCData struct {
	DataLength uint32
	Data       []byte
}

func checkMagicNumber(magic, number uint16, from string) error {
	if number != magic {
		return fmt.Errorf("%s: magic number fail (got 0x%X, want 0x%X)", from, number, magic)
	}

	return nil
}

func (l *Loader) readToBuffer() error {
	_, err := l.Reader.Read(l.Buf)
	if err != nil {
		return err
	}

	_, err = l.Buffer.Write(l.Buf)
	if err != nil {
		return err
	}

	return nil
}

func (l *Loader) enoughSpaceToRead(size int) bool {
	available := l.Buffer.Len()
	needed := size
	fmt.Println("available: ", available, "needed: ", needed)
	return available >= needed
}

// NOTE: DEPRECATED: use v2
func BytesToStruct[T any](loader *Loader, size int) (T, error) {
	var t T
	if loader.enoughSpaceToRead(size) {
		err := binary.Read(loader.Buffer, binary.LittleEndian, &t)
		if err != nil {
			return t, err
		}
		return t, nil
	}

	fmt.Println("pegar mais dados do fd")
	err := loader.readToBuffer()
	if err != nil {
		return t, err
	}

	return BytesToStruct[T](loader, size)
}

func (l *Loader) BytesToStructV2(size int, t any) error {
	if l.enoughSpaceToRead(size) {
		err := binary.Read(l.Buffer, binary.LittleEndian, t)
		if err != nil {
			return err
		}
		return nil
	}

	fmt.Println("pegar mais dados do fd")
	err := l.readToBuffer()
	if err != nil {
		return err
	}

	return l.BytesToStructV2(size, t)
}

func (l *Loader) loadFrameChunkData(ch ChunkHeader) ([]byte, error) {
	size := ch.Size - ChunkHeaderSize
	if l.enoughSpaceToRead(int(size)) {
		bufchunk := make([]byte, size)
		_, err := io.ReadFull(l.Buffer, bufchunk)
		if err != nil {
			return nil, err
		}
		return bufchunk, nil
	}

	fmt.Println("chunk: pegar mais dados do fd")
	err := l.readToBuffer()
	if err != nil {
		return nil, err
	}

	return l.loadFrameChunkData(ch)
}

func (l *Loader) ParseHeader() (Header, error) {
	fmt.Printf("Parser Header")
	header, err := BytesToStruct[Header](l, ChunkHeaderSize)
	if err != nil {
		return header, err
	}

	err = checkMagicNumber(0xA5E0, header.MagicNumber, "header")
	if err != nil {
		return header, err
	}

	return header, nil
}

func (l *Loader) ParseChunk(ch ChunkHeader) (Chunk, error) {
	var chunk Chunk

	fmt.Printf("chunk type: %d", ch.Type)
	switch ch.Type {
	case ColorProfileChunkHex:
		var err error
		if chunk, err = l.ParseColorProfileChunk(chunk, ch); err != nil {
			return nil, err
		}

	default:
		// NOTE: quando definir todos os chunk types, dar erro aqui:
		// return fmt.Errorf("Invalid chunk type: %d", ch.Type)
		l.loadFrameChunkData(ch)
		cfake := &ChunkColorProfile{header: ch}
		return cfake, nil
	}

	return chunk, nil
}

func (l *Loader) ParseColorProfileChunk(chunk Chunk, ch ChunkHeader) (Chunk, error) {
	var cData ChunkColorProfileData
	err := l.BytesToStructV2(ChunkColorProfileDataSize, &cData)
	if err != nil {
		return nil, err
	}

	if cData.Type == ColorProfileICC {
		var iccSize uint32
		err = l.BytesToStructV2(8, &iccSize)
		if err != nil {
			return nil, err
		}
		iccData := make([]byte, iccSize)
		err = l.BytesToStructV2(int(iccSize), &iccData)
		if err != nil {
			return nil, err
		}

		chunk = &ChunkColorProfileICC{
			ChunkColorProfile: ChunkColorProfile{
				header:                ch,
				ChunkColorProfileData: cData,
			},
			ChunkColorProfileICCData: ChunkColorProfileICCData{
				DataLength: iccSize,
				Data:       iccData,
			},
		}
		return chunk, nil
	}
	chunk = &ChunkColorProfile{
		header:                ch,
		ChunkColorProfileData: cData,
	}
	return chunk, nil
}

func (l *Loader) ParseFrames(header *Header) ([]Frame, error) {
	frames := make([]Frame, 0)

	fmt.Printf("Parser Frames, count: %d", header.Frames)
	for i := range header.Frames {
		fmt.Println("frameheader to struct")
		fh, err := BytesToStruct[FrameHeader](l, FrameHeaderSize)
		if err != nil {
			return nil, err
		}
		err = checkMagicNumber(0xF1FA, fh.MagicNumber, "frameheader "+fmt.Sprint(i))
		if err != nil {
			return nil, err
		}

		fmt.Println("Chunk number: ", fh.ChunkNumber)

		chunkList := make([]Chunk, 0)
		// TODO: verificar o numero antigo
		for range fh.ChunkNumber {
			fmt.Println("chunkheader to struct")
			ch, err := BytesToStruct[ChunkHeader](l, ChunkHeaderSize)
			if err != nil {
				return nil, err
			}

			fmt.Println("chunkdata to struct")
			var c Chunk
			c, err = l.ParseChunk(ch)
			if err != nil {
				return nil, err
			}

			chunkList = append(chunkList, c)
		}

		frames = append(frames, Frame{Header: fh, Chunks: chunkList})
		// NOTE: depois de ler dar um reset no buffer? Vê se tem um resto guardar e depois
		// coloca no buffer de novo
	}

	return frames, nil
}

func DeserializeFile(fd *os.File) (*AsepriteFile, error) {
	var ase *AsepriteFile = new(AsepriteFile)
	loader := new(Loader)

	reader := fd

	loader.Buf = make([]byte, ChunkSize)
	loader.Buffer = new(bytes.Buffer)
	loader.Reader = reader
	loader.File = ase

	header, err := loader.ParseHeader()
	if err != nil {
		return nil, err
	}
	loader.Buffer.Reset()
	frames, err := loader.ParseFrames(&header)
	if err != nil {
		return nil, err
	}

	ase.Header = header
	ase.Frames = frames

	return ase, nil
}

func main() {
	path := os.Args[1]
	fd, err := os.Open(path)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer fd.Close()

	_, err = DeserializeFile(fd)
	if err != nil {
		log.Fatalf("error deserializing file: %v", err)
	}
}

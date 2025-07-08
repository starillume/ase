package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
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

const ChunkHeaderSize = 6

type ChunkHeader struct {
	Size uint32
	Type uint16
}

type Chunk struct {
	Header ChunkHeader
	Data   []byte
}

// TODO: criar uma struct Fixed
type ChunkColorProfileData struct {
	Type  uint16
	Flags uint16
	Gamma int32 // TODO: adjust to FIXED type
	_     [8]byte
}

func checkMagicNumber(magic, number uint16, from ...string) error {
	if number != magic {
		return fmt.Errorf("%s: magic number fail (got 0x%X, want 0x%X)",
			strings.Join(from, ""), number, magic)
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

func (l *Loader) ParseFrames(header *Header) ([]Frame, error) {
	frames := make([]Frame, 0)

	fmt.Printf("Parser Frames, count: %d", header.Frames)
	for i := range header.Frames {
		fmt.Println("frameheader to struct")
		fh, err := BytesToStruct[FrameHeader](l, FrameHeaderSize)
		if err != nil {
			return nil, err
		}
		err = checkMagicNumber(0xF1FA, fh.MagicNumber, "frameheader", fmt.Sprint(i))
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
			bufchunk, err := l.loadFrameChunkData(ch)
			if err != nil {
				return nil, err
			}

			c := Chunk{
				Header: ch,
				Data:   bufchunk,
			}

			fmt.Println("Chunk data: ", c.Data)

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

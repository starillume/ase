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

func checkMagicNumber(magic, number uint16, from ...string) {
	fmt.Println(from, "magic number: ", magic, number)
	if number != magic {
		log.Fatalf(strings.Join(from, ""), "magic number fail")
	}
	fmt.Println(from, "magic number pass")
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

func BytesToStruct[T any](loader *Loader, size int) T {
	var t T
	if loader.enoughSpaceToRead(size) {
		err := binary.Read(loader.Buffer, binary.LittleEndian, &t)
		if err != nil {
			fmt.Println("binary.Read: ", err)
		}
		return t
	}

	fmt.Println("pegar mais dados do fd")
	loader.readToBuffer()
	return BytesToStruct[T](loader, size)
}

func (l *Loader) loadFrameChunkData(ch ChunkHeader) []byte {
	size := ch.Size - ChunkHeaderSize
	if l.enoughSpaceToRead(int(size)) {
		bufchunk := make([]byte, size)
		io.ReadFull(l.Buffer, bufchunk)
		return bufchunk
	}

	fmt.Println("chunk: pegar mais dados do fd")
	l.readToBuffer()
	return l.loadFrameChunkData(ch)
}

func (l *Loader) ParseHeader() (Header, error) {
	fmt.Printf("Parser Header")
	return BytesToStruct[Header](l, ChunkHeaderSize), nil
}

func (l *Loader) ParseFrames(header *Header) ([]Frame, error) {
	frames := make([]Frame, 0)

	fmt.Printf("Parser Frames, count: %d", header.Frames)
	for i := range header.Frames {
		fmt.Println("frameheader to struct")
		fh := BytesToStruct[FrameHeader](l, FrameHeaderSize)
		checkMagicNumber(0xF1FA, fh.MagicNumber, "frameheader", fmt.Sprint(i))
		fmt.Println("Chunk number: ", fh.ChunkNumber)

		chunkList := make([]Chunk, 0)
		// TODO: verificar o numero antigo
		for range fh.ChunkNumber {
			fmt.Println("chunkheader to struct")
			ch := BytesToStruct[ChunkHeader](l, ChunkHeaderSize)

			fmt.Println("chunkdata to struct")
			bufchunk := l.loadFrameChunkData(ch)

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

	header, _ := loader.ParseHeader()
	loader.Buffer.Reset()
	frames, _ := loader.ParseFrames(&header)

	ase.Header = header
	ase.Frames = frames

	return ase, nil
}

func main() {
	path := os.Args[1]
	fd, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer fd.Close()

	_, err = DeserializeFile(fd)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
}

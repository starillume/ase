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

var Reader io.Reader
var ChunkPart = make([]byte, ChunkSize)

const (
	ChunkSize = 128
)

type AsepriteFile struct {
	Header Header
	Frames []Frame
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
	// Chunks []Chunk n precisa por enquanto
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

func readToBuffer(reader io.Reader, buffer *bytes.Buffer, chunk []byte) error {
	_, err := reader.Read(chunk)
	if err != nil {
		return err
	}

	_, err = buffer.Write(chunk)
	if err != nil {
		return err
	}

	return nil
}

func enoughSpaceToRead(size int, buffer *bytes.Buffer) bool {
	available := buffer.Len()
	needed := size
	fmt.Println("available: ", available, "needed: ", needed)
	return available >= needed
}

func BytesToStruct[T any](size int, buffer *bytes.Buffer) T {
	var t T
	if enoughSpaceToRead(size, buffer) {
		err := binary.Read(buffer, binary.LittleEndian, &t)
		if err != nil {
			fmt.Println("binary.Read: ", err)
		}
		return t
	}

	fmt.Println("pegar mais dados do fd")
	readToBuffer(Reader, buffer, ChunkPart)
	return BytesToStruct[T](size, buffer)
}

func loadFrameChunkData(ch ChunkHeader, buffer *bytes.Buffer) []byte {
	size := ch.Size - ChunkHeaderSize
	if enoughSpaceToRead(int(size), buffer) {
		bufchunk := make([]byte, size)
		io.ReadFull(buffer, bufchunk)
		return bufchunk
	}

	fmt.Println("chunk: pegar mais dados do fd")
	readToBuffer(Reader, buffer, ChunkPart)
	return loadFrameChunkData(ch, buffer)
}

func main() {
	path := os.Args[1]
	fd, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer fd.Close()

	buffer := new(bytes.Buffer)
	fmt.Println("buffer available on create: ", buffer.Len())
	Reader = fd
	readToBuffer(Reader, buffer, ChunkPart)
	fmt.Println("buffer available after a read fd: ", buffer.Len())

	var h Header
	fmt.Println("header to struct")
	binary.Read(buffer, binary.LittleEndian, &h)
	fmt.Println("buffer available after a read to struct: ", buffer.Len())

	checkMagicNumber(0xA5E0, h.MagicNumber, "header")

	buffer.Reset()

	for i := range h.Frames {
		readToBuffer(Reader, buffer, ChunkPart)

		var fh FrameHeader
		fmt.Println("framaeheader to struct")
		fh = BytesToStruct[FrameHeader](FrameHeaderSize, buffer)
		checkMagicNumber(0xF1FA, fh.MagicNumber, "frameheader", fmt.Sprint(i))

		fmt.Println("Chunk number: ", fh.ChunkNumber)

		for range fh.ChunkNumber {
			var ch ChunkHeader
			fmt.Println("chunkheader to struct")
			ch = BytesToStruct[ChunkHeader](ChunkHeaderSize, buffer)

			fmt.Println("Chunk type: ", ch.Type)
			fmt.Println("Chunk size: ", ch.Size)

			fmt.Println("chunkdata to struct")
			bufchunk := loadFrameChunkData(ch, buffer)

			c := &Chunk{
				Header: ch,
				Data:   bufchunk,
			}

			fmt.Println("Chunk data: ", c.Data)
		}
	}
}

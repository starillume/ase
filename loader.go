package ase

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"

	"github.com/starillume/ase/chunk"
)

const (
	ChunkSize = 128
)

type rawAseprite struct {
	Header        Header
	Palette       *palette
	Frames        []*frame
	Tags          []*tag
	Layers        []*layer
	Slices        []*slice
	ColorProfile  *colorProfile
	ExternalFiles *externalFiles
}

type slice struct {
	Chunk    *chunk.Slice
	UserData *chunk.UserData
}

type palette struct {
	Chunk    *chunk.Chunk
	UserData *chunk.UserData
}

type frame struct {
	Cels []*cel
	Header FrameHeader
}

type cel struct {
	Chunk    *chunk.Chunk
	Extra    *chunk.CelExtra
	UserData *chunk.UserData
}

type tag struct {
	Chunk    *chunk.Tag
	UserData *chunk.UserData
}

type layer struct {
	Chunk    *chunk.Layer
	UserData *chunk.UserData
}

type colorProfile struct {
	Chunk    *chunk.Chunk
	UserData *chunk.UserData
}

type externalFiles struct {
	Chunk    *chunk.ExternalFiles
	UserData *chunk.UserData
}

type Loader struct {
	Reader io.Reader
	Buf    []byte
	Buffer *bytes.Buffer
	Ase    *rawAseprite
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

	err := l.readToBuffer()
	if err != nil {
		return err
	}

	return l.BytesToStructV2(size, t)
}

func (l *Loader) loadFrameChunkData(ch chunk.Header) ([]byte, error) {
	size := ch.Size - chunk.HeaderSize
	if l.enoughSpaceToRead(int(size)) {
		bufchunk := make([]byte, size)
		_, err := io.ReadFull(l.Buffer, bufchunk)
		if err != nil {
			return nil, err
		}
		return bufchunk, nil
	}

	err := l.readToBuffer()
	if err != nil {
		return nil, err
	}

	return l.loadFrameChunkData(ch)
}

func NewLoader(fd *os.File) *Loader {
	return &Loader{
		Reader: fd,
		Buf:    make([]byte, ChunkSize),
		Buffer: new(bytes.Buffer),
		Ase:    new(rawAseprite),
	}
}

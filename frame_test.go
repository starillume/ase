package ase

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"testing"

	"github.com/starillume/ase/chunk"
	"github.com/starillume/ase/common"
)

// ----------------- Helpers -----------------

func createFrameHeaderWithChunks(chunkCount int, frameSize int) FrameHeader {
	return FrameHeader{
		FrameBytes:     uint32(frameSize),
		MagicNumber:    FrameMagicNumber,
		OldChunkNumber: uint16(chunkCount),
		FrameDuration:  100,
		ChunkNumber:    uint32(chunkCount),
	}
}

func createChunkHeaderBytes(typ chunk.ChunkDataType, dataSize uint32) []byte {
	buf := new(bytes.Buffer)
	size := dataSize + chunk.HeaderSize
	binary.Write(buf, binary.LittleEndian, size)
	binary.Write(buf, binary.LittleEndian, uint16(typ))
	binary.Write(buf, binary.LittleEndian, uint16(0)) // reserved
	return buf.Bytes()
}

func createTestFrameBytesForParseFrame(t *testing.T) []byte {
	fd, err := os.Open("./test.aseprite")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer fd.Close()

	fd.Seek(HeaderSize, io.SeekCurrent)

	fh1buf := make([]byte, FrameHeaderSize)
	fd.Read(fh1buf)

	f1size, _ := common.BytesToStruct[uint32](bytes.NewReader(fh1buf))

	fd.Seek(int64(f1size)-FrameHeaderSize, io.SeekCurrent)

	fh2buf := make([]byte, FrameHeaderSize)
	fd.Read(fh2buf)

	f2size, _ := common.BytesToStruct[uint32](bytes.NewReader(fh2buf))

	fdatabuf := make([]byte, f2size-FrameHeaderSize)

	fd.Read(fdatabuf)

	return append(fh2buf, fdatabuf...)
}

func createTestFrameBytesForFirstFrame(t *testing.T) []byte {
	fd, err := os.Open("./test.aseprite")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer fd.Close()

	fd.Seek(HeaderSize, io.SeekCurrent)

	fhbuf := make([]byte, FrameHeaderSize)
	fd.Read(fhbuf)

	fsize, _ := common.BytesToStruct[uint32](bytes.NewReader(fhbuf))

	fdatabuf := make([]byte, fsize-FrameHeaderSize)

	fd.Read(fdatabuf)

	return append(fhbuf, fdatabuf...)
}

// ----------------- Tests -----------------

func TestParseFrame(t *testing.T) {
	data := createTestFrameBytesForParseFrame(t)
	reader := bytes.NewReader(data)
	var fh FrameHeader
	binary.Read(reader, binary.LittleEndian, &fh)

	frameData := make([]byte, fh.FrameBytes-FrameHeaderSize)
	binary.Read(reader, binary.LittleEndian, &frameData)

	frame, err := parseFrame(fh, frameData)
	if err != nil {
		t.Fatalf("parseFrame failed: %v", err)
	}

	if len(frame.Cels) != 2 {
		t.Errorf("Expected 1 Cel, got %d", len(frame.Cels))
	}
}

func TestParseFirstFrame(t *testing.T) {
	data := createTestFrameBytesForFirstFrame(t)
	reader := bytes.NewReader(data)
	var fh FrameHeader
	binary.Read(reader, binary.LittleEndian, &fh)

	frameData := make([]byte, fh.FrameBytes-FrameHeaderSize)
	binary.Read(reader, binary.LittleEndian, &frameData)

	frame, layers, tags, slices, externalFiles, colorProfile, palette, err := parseFirstFrame(fh, frameData)
	if err != nil {
		t.Fatalf("parseFirstFrame failed: %v", err)
	}

	// Frame
	if len(frame.Cels) != 2 {
		t.Errorf("Expected 2 Cel, got %d", len(frame.Cels))
	}
	cel := frame.Cels[0]
	if cel.UserData == nil {
		t.Errorf("Expected UserData attached")
	}

	// Layers
	if len(layers) != 4 {
		t.Errorf("Expected 4 Layer, got %d", len(layers))
	}

	// Tags
	if len(tags) != 2 {
		t.Errorf("Expected 2 Tag, got %d", len(tags))
	}

	// Slices
	if len(slices) != 2 {
		t.Errorf("Expected 2 Slice, got %d", len(slices))
	}

	// Palette
	if palette == nil {
		t.Errorf("Expected Palette")
	}

	// ExternalFiles
	if externalFiles == nil {
		t.Errorf("Expected ExternalFiles struct (even empty)")
	}

	// ColorProfile
	if colorProfile == nil {
		t.Errorf("Expected ColorProfile struct (even empty)")
	}
}

func TestResolveUserDataTags(t *testing.T) {
	// Criar tags sem userdata
	tags := []*tag{{UserData: nil}, {UserData: nil}}
	ud := &chunk.UserData{}

	resolveUserDataTags(ud, tags)

	if tags[0].UserData != ud && tags[1].UserData != ud {
		t.Errorf("Expected UserData to be attached to first nil tag")
	}
}

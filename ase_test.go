package main

import (
	"os"
	"testing"
	"bytes"
)

const testFilePath = "./test.aseprite"

func TestChunkOldPalette(t *testing.T) {
	data := []byte{0x02, 0x00, 0x00, 0x02, 0xFF, 0x00, 0x00, 0x00, 0xFF, 0x00, 0x01, 0x01, 0x00, 0x00, 0xFF}
	chunkHeader := ChunkHeader{
		Size: uint32(len(data)) + ChunkHeaderSize,
		Type: OldPaletteChunkHex,
	}
	loader := &Loader{Buffer: new(bytes.Buffer)}
	_, err := loader.Buffer.Write(data)
	if err != nil {
		t.Fatalf("failed to write to buffer: %v", err)
	}

	_, err = loader.ParseChunkOldPalette(chunkHeader)
	if err != nil {
		t.Fatalf("failed to parse ChunkOldPalette: %v", err)
	}

	if unread := loader.Buffer.Len(); unread != 0 {
		t.Errorf("expected ChunkOldPalette to be fully read, but %d bytes remain (read %d of %d)", unread, len(data) - unread, len(data))
	}
}

func TestDeserializeFile(t *testing.T) {
	fd, err := os.Open(testFilePath)
	if err != nil {
		t.Fatalf("failed to open file %s: %v", testFilePath, err)
	}
	defer fd.Close()

	ase, err := DeserializeFile(fd)
	if err != nil {
		t.Fatalf("failed to deserialize file %s: %v", testFilePath, err)
	}

	verifyHeader(t, ase)
	verifyFrames(t, ase)
	verifyAllDataRead(t, ase, testFilePath)
}

func TestEmptyFile(t *testing.T) {
	tmp, err := os.CreateTemp("", "empty.aseprite")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	_, err = DeserializeFile(tmp)
	if err == nil {
		t.Errorf("expected error when parsing empty file, got nil")
	}
}

func TestZeroFile(t *testing.T) {
	tmp, err := os.CreateTemp("", "zero.aseprite")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	zeroData := make([]byte, 512)
	if _, err := tmp.Write(zeroData); err != nil {
		t.Fatalf("failed to write to temp zero file: %v", err)
	}
	if _, err := tmp.Seek(0, 0); err != nil {
		t.Fatalf("failed to seek in zero file: %v", err)
	}

	_, err = DeserializeFile(tmp)
	if err == nil {
		t.Errorf("expected error when parsing zero file")
	}
}

func verifyHeader(t *testing.T, ase *AsepriteFile) {
	if ase.Header.MagicNumber != 0xA5E0 {
		t.Errorf("invalid header magic number: got 0x%X, want 0xA5E0", ase.Header.MagicNumber)
	}
}

func verifyFrames(t *testing.T, ase *AsepriteFile) {
	expectedFrameCount := 8
	if len(ase.Frames) != expectedFrameCount {
		t.Errorf("expected %d frames, got %d", expectedFrameCount, len(ase.Frames))
	}

	expectedChunks := []int{4, 1, 1, 1, 1, 1, 1, 1}
	for i, frame := range ase.Frames {
		if frame.Header.MagicNumber != 0xF1FA {
			t.Errorf("frame %d has invalid magic number: got 0x%X, want 0xF1FA", i, frame.Header.MagicNumber)
		}
		if len(frame.Chunks) != expectedChunks[i] {
			t.Errorf("frame %d: expected %d chunks, got %d", i, expectedChunks[i], len(frame.Chunks))
		}
	}
}

func verifyAllDataRead(t *testing.T, ase *AsepriteFile, filepath string) {
	stat, err := os.Stat(filepath)
	if err != nil {
		t.Fatalf("failed to stat file %s: %v", filepath, err)
	}
	fileSize := stat.Size()

	read := int64(HeaderSize)

	for _, frame := range ase.Frames {
		read += int64(FrameHeaderSize)
		for _, chunk := range frame.Chunks {
			read += int64(chunk.GetHeader().Size)
		}
	}

	if fileSize - read > 16 {
		t.Errorf("expected file to be fully read, but %d bytes remain (read %d of %d)", fileSize - read, read, fileSize)
	}
}

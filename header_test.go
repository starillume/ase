package ase

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/starillume/ase/pixel"
)

func createHeaderData() []byte {
	buf := new(bytes.Buffer)

	// FileSize
	binary.Write(buf, binary.LittleEndian, uint32(123456))
	// MagicNumber
	binary.Write(buf, binary.LittleEndian, uint16(0xA5E0))
	// Frames
	binary.Write(buf, binary.LittleEndian, uint16(2))
	// Width, Height
	binary.Write(buf, binary.LittleEndian, uint16(64))
	binary.Write(buf, binary.LittleEndian, uint16(32))
	// ColorDepth (pixel.ColorDepth, uint16)
	var depth pixel.ColorDepth = 32
	binary.Write(buf, binary.LittleEndian, depth)
	// Flags
	binary.Write(buf, binary.LittleEndian, uint32(0))
	// FrameSpeed
	binary.Write(buf, binary.LittleEndian, uint16(16))
	// 2 x uint32 padding
	binary.Write(buf, binary.LittleEndian, uint32(0))
	binary.Write(buf, binary.LittleEndian, uint32(0))
	// PaletteEntry
	binary.Write(buf, binary.LittleEndian, byte(0))
	// 3 byte padding
	buf.Write([]byte{0, 0, 0})
	// NumberColors
	binary.Write(buf, binary.LittleEndian, uint16(256))
	// PixelWidth, PixelHeight
	buf.WriteByte(1)
	buf.WriteByte(1)
	// GridX, GridY
	binary.Write(buf, binary.LittleEndian, int16(0))
	binary.Write(buf, binary.LittleEndian, int16(0))
	// GridWidth, GridHeight
	binary.Write(buf, binary.LittleEndian, uint16(64))
	binary.Write(buf, binary.LittleEndian, uint16(32))
	// 84 bytes padding
	buf.Write(make([]byte, 84))

	return buf.Bytes()
}

func TestParseHeader(t *testing.T) {
	data := createHeaderData()
	header, err := ParseHeader(data)
	if err != nil {
		t.Fatalf("ParseHeader failed: %v", err)
	}

	if header.FileSize != 123456 {
		t.Errorf("FileSize mismatch: expected 123456, got %d", header.FileSize)
	}
	if header.MagicNumber != 0xA5E0 {
		t.Errorf("MagicNumber mismatch: expected 0xA5E0, got 0x%X", header.MagicNumber)
	}
	if header.Frames != 2 {
		t.Errorf("Frames mismatch: expected 2, got %d", header.Frames)
	}
	if header.Width != 64 || header.Height != 32 {
		t.Errorf("Dimensions mismatch: expected 64x32, got %dx%d", header.Width, header.Height)
	}
	if header.ColorDepth != 32 {
		t.Errorf("ColorDepth mismatch: expected 32, got %d", header.ColorDepth)
	}
	if header.NumberColors != 256 {
		t.Errorf("NumberColors mismatch: expected 256, got %d", header.NumberColors)
	}
	if header.PixelWidth != 1 || header.PixelHeight != 1 {
		t.Errorf("Pixel dimensions mismatch: expected 1x1, got %dx%d", header.PixelWidth, header.PixelHeight)
	}
	if header.GridWidth != 64 || header.GridHeight != 32 {
		t.Errorf("Grid size mismatch: expected 64x32, got %dx%d", header.GridWidth, header.GridHeight)
	}
}


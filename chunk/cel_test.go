package chunk

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/starillume/ase/pixel"
)

func createCelData(celType CelType) []byte {
	buf := new(bytes.Buffer)

	cData := ChunkCelData{
		LayerIndex: 9,
		X:          -4,
		Y:          8,
		Opacity:    200,
		CelType:    celType,
		Z:          42,
	}
	binary.Write(buf, binary.LittleEndian, cData)

	if celType == CelTypeLinked {
		link := ChunkCelLinkedData{FramePosition: 13}
		binary.Write(buf, binary.LittleEndian, link)
		return buf.Bytes()
	}

	dim := ChunkCelDimensionData{Width: 2, Height: 3}
	binary.Write(buf, binary.LittleEndian, dim)

	buf.Write([]byte{10, 20, 30, 40, 50, 60})
	return buf.Bytes()
}

func TestParseChunkCel_RawImage(t *testing.T) {
	data := createCelData(CelTypeRawImage)
	chunk, err := ParseChunkCel(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cel, ok := chunk.(*ChunkCelImage)
	if !ok {
		t.Fatalf("expected *ChunkCelImage, got %T", chunk)
	}

	if cel.CelType() != CelTypeRawImage {
		t.Errorf("CelType: got %v, want %v", cel.CelType(), CelTypeRawImage)
	}

	if cel.LayerIndex != 9 {
		t.Errorf("LayerIndex: got %d, want 9", cel.LayerIndex)
	}
	if cel.X != -4 {
		t.Errorf("X: got %d, want -4", cel.X)
	}
	if cel.Y != 8 {
		t.Errorf("Y: got %d, want 8", cel.Y)
	}
	if cel.Opacity != 200 {
		t.Errorf("Opacity: got %d, want 200", cel.Opacity)
	}
	if cel.Z != 42 {
		t.Errorf("Z: got %d, want 42", cel.Z)
	}
	if cel.Width != 2 {
		t.Errorf("Width: got %d, want 2", cel.Width)
	}
	if cel.Height != 3 {
		t.Errorf("Height: got %d, want 3", cel.Height)
	}
	if len(cel.Pixels) != 6 {
		t.Errorf("Pixels length: got %d, want 6", len(cel.Pixels))
	}
}

func TestParseChunkCel_CompressedImage(t *testing.T) {
	data := createCelData(CelTypeCompressedImage)
	chunk, err := ParseChunkCel(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cel, ok := chunk.(*ChunkCelCompressedImage)
	if !ok {
		t.Fatalf("expected *ChunkCelCompressedImage, got %T", chunk)
	}

	if cel.CelType() != CelTypeCompressedImage {
		t.Errorf("CelType: got %v, want %v", cel.CelType(), CelTypeCompressedImage)
	}
	if cel.LayerIndex != 9 {
		t.Errorf("LayerIndex: got %d, want 9", cel.LayerIndex)
	}
	if cel.X != -4 {
		t.Errorf("X: got %d, want -4", cel.X)
	}
	if cel.Y != 8 {
		t.Errorf("Y: got %d, want 8", cel.Y)
	}
	if cel.Opacity != 200 {
		t.Errorf("Opacity: got %d, want 200", cel.Opacity)
	}
	if cel.Z != 42 {
		t.Errorf("Z: got %d, want 42", cel.Z)
	}
	if cel.Width != 2 {
		t.Errorf("Width: got %d, want 2", cel.Width)
	}
	if cel.Height != 3 {
		t.Errorf("Height: got %d, want 3", cel.Height)
	}
	if len(cel.Pixels.(pixel.PixelsZlib)) != 6 {
		t.Errorf("Pixels length: got %d, want 6", len(cel.Pixels.(pixel.PixelsZlib)))
	}
}

func TestParseChunkCel_Linked(t *testing.T) {
	data := createCelData(CelTypeLinked)
	chunk, err := ParseChunkCel(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cel, ok := chunk.(*ChunkCelLinked)
	if !ok {
		t.Fatalf("expected *ChunkCelLinked, got %T", chunk)
	}

	if cel.CelType() != CelTypeLinked {
		t.Errorf("CelType: got %v, want %v", cel.CelType(), CelTypeLinked)
	}
	if cel.LayerIndex != 9 {
		t.Errorf("LayerIndex: got %d, want 9", cel.LayerIndex)
	}
	if cel.X != -4 {
		t.Errorf("X: got %d, want -4", cel.X)
	}
	if cel.Y != 8 {
		t.Errorf("Y: got %d, want 8", cel.Y)
	}
	if cel.Opacity != 200 {
		t.Errorf("Opacity: got %d, want 200", cel.Opacity)
	}
	if cel.Z != 42 {
		t.Errorf("Z: got %d, want 42", cel.Z)
	}
	if cel.FramePosition != 13 {
		t.Errorf("FramePosition: got %d, want 13", cel.FramePosition)
	}
}

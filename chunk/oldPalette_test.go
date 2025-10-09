package chunk

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func createOldPaletteDataMultiplePackets() []byte {
	buf := new(bytes.Buffer)

	packetsNumber := uint16(3)
	binary.Write(buf, binary.LittleEndian, packetsNumber)

	binary.Write(buf, binary.LittleEndian, byte(0))
	binary.Write(buf, binary.LittleEndian, byte(2))
	colors1 := []ChunkOldPaletteColor{
		{R: 10, G: 20, B: 30},
		{R: 40, G: 50, B: 60},
	}
	for _, c := range colors1 {
		binary.Write(buf, binary.LittleEndian, c)
	}

	binary.Write(buf, binary.LittleEndian, byte(1))
	binary.Write(buf, binary.LittleEndian, byte(3))
	colors2 := []ChunkOldPaletteColor{
		{R: 70, G: 80, B: 90},
		{R: 100, G: 110, B: 120},
		{R: 130, G: 140, B: 150},
	}
	for _, c := range colors2 {
		binary.Write(buf, binary.LittleEndian, c)
	}

	binary.Write(buf, binary.LittleEndian, byte(4))
	binary.Write(buf, binary.LittleEndian, byte(1))
	colors3 := []ChunkOldPaletteColor{
		{R: 200, G: 210, B: 220},
	}
	for _, c := range colors3 {
		binary.Write(buf, binary.LittleEndian, c)
	}

	return buf.Bytes()
}

func TestParseChunkOldPalette(t *testing.T) {
	data := createOldPaletteDataMultiplePackets()

	chunk, err := ParseChunkOldPalette(data, OldPaletteChunkHex)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	palette, ok := chunk.(*ChunkOldPalette)
	if !ok {
		t.Fatalf("expected *ChunkOldPalette, got %T", chunk)
	}

	if len(palette.Colors) != 256 {
		t.Errorf("Colors length: got %d, want 256", len(palette.Colors))
	}
}

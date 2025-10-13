package chunk_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/starillume/ase/chunk"
)

func createChunkPaletteData() []byte {
	buf := new(bytes.Buffer)

	// ChunkPaletteData: 3 entries, from 0 to 2
	cData := chunk.ChunkPaletteData{
		EntriesNumber: 3,
		From:          0,
		To:            2,
	}
	binary.Write(buf, binary.LittleEndian, cData)

	// Entry 0: has name
	entry0 := chunk.ChunkPaletteEntryData{
		HasName: 1,
		Red:     10,
		Green:   20,
		Blue:    30,
	}
	binary.Write(buf, binary.LittleEndian, entry0)
	name0 := "Redish"
	nameLen0 := uint16(len(name0))
	binary.Write(buf, binary.LittleEndian, nameLen0)
	buf.Write([]byte(name0))

	// Entry 1: no name
	entry1 := chunk.ChunkPaletteEntryData{
		HasName: 0,
		Red:     40,
		Green:   50,
		Blue:    60,
	}
	binary.Write(buf, binary.LittleEndian, entry1)

	// Entry 2: has name
	entry2 := chunk.ChunkPaletteEntryData{
		HasName: 1,
		Red:     70,
		Green:   80,
		Blue:    90,
	}
	binary.Write(buf, binary.LittleEndian, entry2)
	name2 := "Bluish"
	nameLen2 := uint16(len(name2))
	binary.Write(buf, binary.LittleEndian, nameLen2)
	buf.Write([]byte(name2))

	return buf.Bytes()
}

func TestParseChunkPalette(t *testing.T) {
	data := createChunkPaletteData()
	parsed, err := chunk.ParseChunkPalette(data)
	if err != nil {
		t.Fatalf("ParseChunkPalette failed: %v", err)
	}

	cp := parsed.(*chunk.ChunkPalette)
	if cp.EntriesNumber != 3 {
		t.Fatalf("Expected EntriesNumber 3, got %d", cp.EntriesNumber)
	}
	if len(cp.Entries) != 3 {
		t.Fatalf("Expected 3 entries in slice, got %d", len(cp.Entries))
	}

	// Entry 0
	if cp.Entries[0].ColorName != "Redish" {
		t.Errorf("Entry 0: expected ColorName 'Redish', got '%s'", cp.Entries[0].ColorName)
	}
	if cp.Entries[0].Red != 10 {
		t.Errorf("Entry 0: expected Red 10, got %d", cp.Entries[0].Red)
	}
	if cp.Entries[0].Green != 20 {
		t.Errorf("Entry 0: expected Green 20, got %d", cp.Entries[0].Green)
	}
	if cp.Entries[0].Blue != 30 {
		t.Errorf("Entry 0: expected Blue 30, got %d", cp.Entries[0].Blue)
	}

	// Entry 1
	if cp.Entries[1].ColorName != "" {
		t.Errorf("Entry 1: expected ColorName '', got '%s'", cp.Entries[1].ColorName)
	}
	if cp.Entries[1].Red != 40 {
		t.Errorf("Entry 1: expected Red 40, got %d", cp.Entries[1].Red)
	}
	if cp.Entries[1].Green != 50 {
		t.Errorf("Entry 1: expected Green 50, got %d", cp.Entries[1].Green)
	}
	if cp.Entries[1].Blue != 60 {
		t.Errorf("Entry 1: expected Blue 60, got %d", cp.Entries[1].Blue)
	}

	// Entry 2
	if cp.Entries[2].ColorName != "Bluish" {
		t.Errorf("Entry 2: expected ColorName 'Bluish', got '%s'", cp.Entries[2].ColorName)
	}
	if cp.Entries[2].Red != 70 {
		t.Errorf("Entry 2: expected Red 70, got %d", cp.Entries[2].Red)
	}
	if cp.Entries[2].Green != 80 {
		t.Errorf("Entry 2: expected Green 80, got %d", cp.Entries[2].Green)
	}
	if cp.Entries[2].Blue != 90 {
		t.Errorf("Entry 2: expected Blue 90, got %d", cp.Entries[2].Blue)
	}
}

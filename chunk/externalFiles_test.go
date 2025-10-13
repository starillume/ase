package chunk

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func createExternalFilesData() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, ChunkExternalFilesData{NumberEntries: 2})

	name1 := "fileA.png"
	entry1 := ChunkExternalFilesEntryData{
		ID:         1,
		Type:       0x01,
		NameLength: uint16(len(name1)),
	}
	binary.Write(buf, binary.LittleEndian, entry1)
	buf.Write([]byte(name1))

	name2 := "fileB.txt"
	entry2 := ChunkExternalFilesEntryData{
		ID:         2,
		Type:       0x02,
		NameLength: uint16(len(name2)),
	}
	binary.Write(buf, binary.LittleEndian, entry2)
	buf.Write([]byte(name2))

	return buf.Bytes()
}

func TestParseChunkExternalFiles(t *testing.T) {
	data := createExternalFilesData()
	chunk, err := ParseChunkExternalFiles(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	externalFiles, ok := chunk.(*ExternalFiles)
	if !ok {
		t.Fatalf("expected *ExternalFiles, got %T", chunk)
	}

	if externalFiles.NumberEntries != 2 {
		t.Errorf("NumberEntries: got %d, want 2", externalFiles.NumberEntries)
	}

	entry1 := externalFiles.Entries[0]
	if entry1.ID != 1 {
		t.Errorf("Entry 1 ID: got %d, want 1", entry1.ID)
	}
	if entry1.Type != 0x01 {
		t.Errorf("Entry 1 Type: got 0x%X, want 0x01", entry1.Type)
	}
	if entry1.Name != "fileA.png" {
		t.Errorf("Entry 1 Name: got %s, want fileA.png", entry1.Name)
	}

	entry2 := externalFiles.Entries[1]
	if entry2.ID != 2 {
		t.Errorf("Entry 2 ID: got %d, want 2", entry2.ID)
	}
	if entry2.Type != 0x02 {
		t.Errorf("Entry 2 Type: got 0x%X, want 0x02", entry2.Type)
	}
	if entry2.Name != "fileB.txt" {
		t.Errorf("Entry 2 Name: got %s, want fileB.txt", entry2.Name)
	}
}

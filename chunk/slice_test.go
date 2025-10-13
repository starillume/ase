package chunk

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func createChunkSliceData() []byte {
	buf := new(bytes.Buffer)

	chunkData := ChunkSliceData{
		NumberSliceKeys: 2,
		FlagsBit:        3,
		NameLength:      5,
	}
	binary.Write(buf, binary.LittleEndian, chunkData)

	buf.Write([]byte("Slice"))

	key1 := ChunkSliceKeyData{
		FrameNumber: 0,
		OriginX:     1,
		OriginY:     2,
		Width:       10,
		Height:      20,
	}
	binary.Write(buf, binary.LittleEndian, key1)

	ninePatch1 := ChunkSliceKey9PatchesData{
		CenterX:      3,
		CenterY:      4,
		CenterWidth:  5,
		CenterHeight: 6,
	}
	binary.Write(buf, binary.LittleEndian, ninePatch1)

	pivot1 := ChunkSliceKeyPivotData{
		X: 7,
		Y: 8,
	}
	binary.Write(buf, binary.LittleEndian, pivot1)

	key2 := ChunkSliceKeyData{
		FrameNumber: 1,
		OriginX:     11,
		OriginY:     12,
		Width:       30,
		Height:      40,
	}
	binary.Write(buf, binary.LittleEndian, key2)

	ninePatch2 := ChunkSliceKey9PatchesData{
		CenterX:      13,
		CenterY:      14,
		CenterWidth:  15,
		CenterHeight: 16,
	}
	binary.Write(buf, binary.LittleEndian, ninePatch2)

	pivot2 := ChunkSliceKeyPivotData{
		X: 17,
		Y: 18,
	}
	binary.Write(buf, binary.LittleEndian, pivot2)

	return buf.Bytes()
}

func TestParseChunkSlice(t *testing.T) {
	data := createChunkSliceData()

	chunk, err := ParseChunkSlice(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	slice, ok := chunk.(*Slice)
	if !ok {
		t.Fatalf("expected *Slice, got %T", chunk)
	}

	if slice.Name != "Slice" {
		t.Errorf("Name: got %s, want Slice", slice.Name)
	}

	if len(slice.Keys) != 2 {
		t.Errorf("Keys length: got %d, want 2", len(slice.Keys))
	}

	// Key 1 checks
	k1 := slice.Keys[0]
	if k1.FrameNumber != 0 || k1.OriginX != 1 || k1.OriginY != 2 || k1.Width != 10 || k1.Height != 20 {
		t.Errorf("Key 1 data mismatch: %+v", k1)
	}
	if k1.ChunkSliceKey9PatchesData == nil || k1.ChunkSliceKeyPivotData == nil {
		t.Errorf("Key 1 missing 9Patches or Pivot")
	}

	// Key 2 checks
	k2 := slice.Keys[1]
	if k2.FrameNumber != 1 || k2.OriginX != 11 || k2.OriginY != 12 || k2.Width != 30 || k2.Height != 40 {
		t.Errorf("Key 2 data mismatch: %+v", k2)
	}
	if k2.ChunkSliceKey9PatchesData == nil || k2.ChunkSliceKeyPivotData == nil {
		t.Errorf("Key 2 missing 9Patches or Pivot")
	}
}

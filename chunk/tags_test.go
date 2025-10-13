package chunk

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func createChunkTagData() []byte {
	buf := new(bytes.Buffer)

	// TagData: 2 tags
	cData := TagData{
		NumberTags: 2,
	}
	binary.Write(buf, binary.LittleEndian, cData)

	// Tag 1
	tag1 := TagEntryData{
		FromFrame:         0,
		ToFrame:           5,
		LoopAnimationType: LoopAnimationForward,
		Repeat:            1,
		Color:             [3]byte{10, 20, 30},
		TagNameSize:       6,
	}
	binary.Write(buf, binary.LittleEndian, tag1)
	buf.Write([]byte("TagOne"))

	// Tag 2
	tag2 := TagEntryData{
		FromFrame:         6,
		ToFrame:           10,
		LoopAnimationType: LoopAnimationPingPong,
		Repeat:            2,
		Color:             [3]byte{40, 50, 60},
		TagNameSize:       6,
	}
	binary.Write(buf, binary.LittleEndian, tag2)
	buf.Write([]byte("TagTwo"))

	return buf.Bytes()
}

func TestParseChunkTag(t *testing.T) {
	data := createChunkTagData()

	tagChunk, err := ParseChunkTag(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tagChunk.Entries) != 2 {
		t.Fatalf("Entries length: got %d, want 2", len(tagChunk.Entries))
	}

	// --- Tag 1 ---
	e1 := tagChunk.Entries[0]
	if e1.FromFrame != 0 || e1.ToFrame != 5 || e1.LoopAnimationType != LoopAnimationForward || e1.Repeat != 1 {
		t.Errorf("Tag 1 frame/loop/repeat mismatch: %+v", e1)
	}
	if e1.Color != [3]byte{10, 20, 30} {
		t.Errorf("Tag 1 color mismatch: got %+v, want [10 20 30]", e1.Color)
	}
	if e1.Name != "TagOne" {
		t.Errorf("Tag 1 name mismatch: got %s, want TagOne", e1.Name)
	}

	// --- Tag 2 ---
	e2 := tagChunk.Entries[1]
	if e2.FromFrame != 6 || e2.ToFrame != 10 || e2.LoopAnimationType != LoopAnimationPingPong || e2.Repeat != 2 {
		t.Errorf("Tag 2 frame/loop/repeat mismatch: %+v", e2)
	}
	if e2.Color != [3]byte{40, 50, 60} {
		t.Errorf("Tag 2 color mismatch: got %+v, want [40 50 60]", e2.Color)
	}
	if e2.Name != "TagTwo" {
		t.Errorf("Tag 2 name mismatch: got %s, want TagTwo", e2.Name)
	}
}

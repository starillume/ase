package chunk

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func createLayerData(flagsBit uint16, layerType uint16, name string, includeType2 bool, lockMovement bool) []byte {
	buf := new(bytes.Buffer)

	layerData := ChunkLayerData{
		FlagsBit:      flagsBit,
		Type:          layerType,
		ChildLevel:    1,
		DefaultWidth:  256,
		DefaultHeight: 128,
		BlendMode:     0,
		Opacity:       200,
		NameLength:    uint16(len(name)),
	}
	binary.Write(buf, binary.LittleEndian, layerData)

	buf.Write([]byte(name))

	if includeType2 && layerType == 2 {
		type2 := ChunkLayerType2Data{TilesetIndex: 42}
		binary.Write(buf, binary.LittleEndian, type2)
	}

	if lockMovement && (flagsBit&(1<<2) != 0) {
		var lock ChunkLayerLockMovementData
		copy(lock.LayerUUID[:], []byte("abcdefghijklmnop"))
		binary.Write(buf, binary.LittleEndian, lock)
	}

	return buf.Bytes()
}

func TestParseChunkLayer_Basic(t *testing.T) {
	data := createLayerData(0x1B, 1, "Layer1", false, false)
	layer, err := ParseChunkLayer(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if layer.ChunkLayerData.FlagsBit != 0x1B {
		t.Errorf("FlagsBit: got 0x%X, want 0x1B", layer.ChunkLayerData.FlagsBit)
	}
	if layer.ChunkLayerData.Type != 1 {
		t.Errorf("Type: got %d, want 1", layer.ChunkLayerData.Type)
	}
	if layer.ChunkLayerData.ChildLevel != 1 {
		t.Errorf("ChildLevel: got %d, want 1", layer.ChunkLayerData.ChildLevel)
	}
	if layer.ChunkLayerData.DefaultWidth != 256 {
		t.Errorf("DefaultWidth: got %d, want 256", layer.ChunkLayerData.DefaultWidth)
	}
	if layer.ChunkLayerData.DefaultHeight != 128 {
		t.Errorf("DefaultHeight: got %d, want 128", layer.ChunkLayerData.DefaultHeight)
	}
	if layer.ChunkLayerData.Opacity != 200 {
		t.Errorf("Opacity: got %d, want 200", layer.ChunkLayerData.Opacity)
	}

	if string(layer.ChunkLayerName) != "Layer1" {
		t.Errorf("Name: got %s, want Layer1", layer.ChunkLayerName)
	}

	if !layer.Visible {
		t.Error("Visible: expected true")
	}
	if !layer.Editable {
		t.Error("Editable: expected true")
	}
	if layer.LockMovement {
		t.Error("LockMovement: expected false")
	}
	if !layer.Background {
		t.Error("Background: expected true")
	}
	if !layer.PreferLinkedCels {
		t.Error("PreferLinkedCels: expected true")
	}
	if layer.LayerGroupDisplayCollapsed {
		t.Error("LayerGroupDisplayCollapsed: expected false")
	}
	if layer.ReferenceLayer {
		t.Error("ReferenceLayer: expected false")
	}

	if layer.ChunkLayerType2Data != nil {
		t.Errorf("ChunkLayerType2Data: expected nil, got %+v", layer.ChunkLayerType2Data)
	}
	if layer.ChunkLayerLockMovementData != nil {
		t.Errorf("ChunkLayerLockMovementData: expected nil, got %+v", layer.ChunkLayerLockMovementData)
	}
}

func TestParseChunkLayer_Type2WithLock(t *testing.T) {
	flags := uint16(1<<2)
	data := createLayerData(flags, 2, "Layer2", true, true)
	layer, err := ParseChunkLayer(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if layer.ChunkLayerData.Type != 2 {
		t.Errorf("Type: got %d, want 2", layer.ChunkLayerData.Type)
	}

	if layer.ChunkLayerType2Data == nil {
		t.Fatal("expected ChunkLayerType2Data to be non-nil")
	}
	if layer.ChunkLayerType2Data.TilesetIndex != 42 {
		t.Errorf("TilesetIndex: got %d, want 42", layer.ChunkLayerType2Data.TilesetIndex)
	}

	if !layer.LockMovement {
		t.Error("LockMovement: expected true")
	}

	if layer.ChunkLayerLockMovementData == nil {
		t.Fatal("expected ChunkLayerLockMovementData to be non-nil")
	}

	uuidStr := string(layer.ChunkLayerLockMovementData.LayerUUID[:])
	if uuidStr != "abcdefghijklmnop" {
		t.Errorf("LayerUUID: got %s, want abcdefghijklmnop", uuidStr)
	}
}

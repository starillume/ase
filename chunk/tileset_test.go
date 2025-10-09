package chunk

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func createChunkTilesetData() []byte {
	buf := new(bytes.Buffer)

	tilesetData := ChunkTilesetData{
		ID:          123,
		FlagsBit:    0b111111,
		TilesNumber: 10,
		TileWidth:   32,
		TileHeight:  32,
		BaseIndex:   0,
		NameLength:  7,
	}
	binary.Write(buf, binary.LittleEndian, tilesetData)

	buf.Write([]byte("Tileset"))

	return buf.Bytes()
}

func TestParseChunkTileset(t *testing.T) {
	data := createChunkTilesetData()

	chunk, err := ParseChunkTileset(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tileset, ok := chunk.(*ChunkTileset)
	if !ok {
		t.Fatalf("expected *ChunkTileset, got %T", chunk)
	}

	if tileset.ID != 123 {
		t.Errorf("ID: got %d, want 123", tileset.ID)
	}
	if tileset.TilesNumber != 10 {
		t.Errorf("TilesNumber: got %d, want 10", tileset.TilesNumber)
	}
	if tileset.TileWidth != 32 || tileset.TileHeight != 32 {
		t.Errorf("Tile dimensions: got %dx%d, want 32x32", tileset.TileWidth, tileset.TileHeight)
	}
	if tileset.BaseIndex != 0 {
		t.Errorf("BaseIndex: got %d, want 0", tileset.BaseIndex)
	}
	if tileset.Name != "Tileset" {
		t.Errorf("Name: got %s, want Tileset", tileset.Name)
	}

	// Check flags individually
	flags := tileset.Flags
	if !flags.LinkExternalFile {
		t.Error("LinkExternalFile flag expected true")
	}
	if !flags.LinkTiles {
		t.Error("LinkTiles flag expected true")
	}
	if !flags.UseTileID0 {
		t.Error("UseTileID0 flag expected true")
	}
	if !flags.AutoFlipX {
		t.Error("AutoFlipX flag expected true")
	}
	if !flags.AutoFlipY {
		t.Error("AutoFlipY flag expected true")
	}
	if !flags.AutoFlipD {
		t.Error("AutoFlipD flag expected true")
	}
}

package ase

import (
	"bytes"
	"image/png"
	"os"
	"testing"
)

const testFilePath = "./test.aseprite"

func TestChunkCel(t *testing.T) {
	data := []byte{
		0x03, 0x00,             // Layer index = 3
		0x0A, 0x00,             // X = 10
		0x14, 0x00,             // Y = 20
		0xFF,                   // Opacity = 255
		0x02, 0x00,             // Cel Type = 2 (Compressed Image)
		0x00, 0x00,             // Z-index = 0

		// Reserved (5 bytes)
		0x00, 0x00, 0x00, 0x00, 0x00,

		// Cel data for type 2
		0x22, 0x00,             // Width = 34
		0x22, 0x00,             // Height = 34
		0x78, 0x9c, 0xed, 0x95, 0xb1, 0x0d, 0xc2, 0x30, 0x10, 0x45, 0x19, 0x08, 0x51, 0x40, 0x07, 0x15,
		0xa2, 0xa7, 0xcd, 0x00, 0xd4, 0x88, 0x01, 0x32, 0x00, 0x55, 0x76, 0xa1, 0x85, 0x01, 0x28, 0x69,
		0x11, 0x23, 0x30, 0x45, 0x10, 0x96, 0x1c, 0x45, 0xd6, 0xd9, 0x77, 0xff, 0x7c, 0x38, 0x05, 0xfe,
		0xd2, 0xef, 0x4e, 0xff, 0x3f, 0x9d, 0x7c, 0xc9, 0xb3, 0xbd, 0xf5, 0xb3, 0xaa, 0xaa, 0xaa, 0xaa,
		0x3f, 0xd4, 0xfb, 0xd4, 0xf6, 0x31, 0x97, 0xec, 0x7e, 0x35, 0x87, 0xa8, 0x7f, 0xc5, 0x24, 0xe9,
		0xe6, 0x98, 0x2c, 0x18, 0x34, 0xfd, 0x14, 0x4f, 0xe9, 0x1d, 0x58, 0xed, 0xc6, 0xa2, 0xff, 0x38,
		0x5f, 0x3a, 0x6b, 0x77, 0x93, 0x62, 0x48, 0x65, 0x87, 0x73, 0xf7, 0x73, 0xe7, 0xac, 0x61, 0xe1,
		0x18, 0xb8, 0xec, 0xd8, 0x2c, 0xca, 0x12, 0xe3, 0xf0, 0x59, 0x08, 0x47, 0x68, 0x4b, 0x0e, 0x2e,
		0x37, 0x9c, 0xef, 0x36, 0x3b, 0x67, 0xe4, 0x8e, 0xb8, 0xb7, 0xe9, 0x33, 0x25, 0xef, 0x03, 0xf5,
		0x98, 0xc3, 0x3a, 0x5b, 0xc3, 0x32, 0x35, 0x43, 0xe5, 0xc8, 0xe3, 0x90, 0xde, 0x0c, 0x72, 0x5b,
		0x28, 0x07, 0x75, 0xbb, 0x54, 0x87, 0x74, 0x4e, 0xc3, 0x11, 0xcb, 0x46, 0x3c, 0x25, 0xc7, 0x65,
		0xb1, 0x1e, 0x2c, 0xe1, 0xe0, 0x58, 0xb4, 0xfd, 0x8f, 0xd5, 0x76, 0x70, 0x49, 0x0e, 0xaa, 0x1f,
		0xe5, 0xc8, 0x61, 0x49, 0xf5, 0x4b, 0x38, 0xa8, 0x7f, 0x1d, 0xca, 0xc2, 0xf5, 0xa3, 0xbb, 0x40,
		0xfe, 0x35, 0x08, 0x83, 0x66, 0x17, 0x5e, 0xd7, 0x7d, 0x23, 0xfa, 0xf6, 0xe4, 0xf8, 0xdb, 0x91,
		0x62, 0x18, 0xb3, 0x58, 0xf3, 0xf8, 0x4c, 0x29, 0x03, 0xc5, 0x63, 0xe1, 0x54, 0xcf, 0x07, 0x7d,
		0x11, 0xf7, 0x9b,
	}

	chunkHeader := ChunkHeader{
		Size: uint32(len(data)) + ChunkHeaderSize,
		Type: CelChunkHex,
	}

	tmp, err := os.CreateTemp("", "chunk_cel_test.aseprite")
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := tmp.Write(data); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	if _, err := tmp.Seek(0, 0); err != nil {
		t.Fatalf("failed to seek in temp file: %v", err)
	}

	loader := &Loader{Buffer: new(bytes.Buffer), Reader: tmp, Buf: make([]byte, len(data)), File: &AsepriteFile{Header: Header{Width: 36, Height: 36, ColorDepth: ColorDepthRGBA}}}

	chunk, err := loader.ParseChunkCel(chunkHeader, 0)

	if err != nil {
		t.Fatalf("failed to parse ChunkCel: %v", err)
	}

	if chunk.GetType() != CelChunkHex {
		t.Errorf("unexpected chunk type: got %d, want %d", chunk.GetType(), CelChunkHex)
	}

	chunkCel := chunk.(*ChunkCelImage)
	
	if chunkCel.LayerIndex != 3 {
		t.Errorf("unexpected layer index: got %d, want %d", chunkCel.LayerIndex, 3)
	}

	if chunkCel.CelType != 2 {
		t.Errorf("unexpected cel type: got %d, want %d", chunkCel.CelType, 2)
	}

	if chunkCel.Width != 34 || chunkCel.Height != 34 {
		t.Errorf("unexpected width or height: got %d and %d, want %d and %d", chunkCel.Width, chunkCel.Height, 34, 34)
	}
}

func TestChunkTileset(t *testing.T) {
	data := []byte{
		0x11, 0x00, 0x00, 0x00, // Tileset ID = 17
		0x3F, 0x00, 0x00, 0x00, // Flags = 63
		0x02, 0x00, 0x00, 0x00, // Number of tiles = 2
		0x10, 0x00,             // Tile width = 16
		0x10, 0x00,             // Tile height = 16
		0x01, 0x00,             // Base index = 1

		// Reserved 14 bytes
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00,

		// STRING: "Terrain"
		0x07, 0x00,             // Length = 7
		'T', 'e', 'r', 'r', 'a', 'i', 'n',

		// Flag 1 → external file link
		0x2A, 0x00, 0x00, 0x00, // External file ID = 42
		0x07, 0x00, 0x00, 0x00, // External tileset ID = 7

		// Flag 2 → inline image data
		0xf3, 0x00, 0x00, 0x00, // length = 243 bytes
		0x78, 0x9c, 0xed, 0x95, 0xb1, 0x0d, 0xc2, 0x30, 0x10, 0x45, 0x19, 0x08, 0x51, 0x40, 0x07, 0x15,
		0xa2, 0xa7, 0xcd, 0x00, 0xd4, 0x88, 0x01, 0x32, 0x00, 0x55, 0x76, 0xa1, 0x85, 0x01, 0x28, 0x69,
		0x11, 0x23, 0x30, 0x45, 0x10, 0x96, 0x1c, 0x45, 0xd6, 0xd9, 0x77, 0xff, 0x7c, 0x38, 0x05, 0xfe,
		0xd2, 0xef, 0x4e, 0xff, 0x3f, 0x9d, 0x7c, 0xc9, 0xb3, 0xbd, 0xf5, 0xb3, 0xaa, 0xaa, 0xaa, 0xaa,
		0x3f, 0xd4, 0xfb, 0xd4, 0xf6, 0x31, 0x97, 0xec, 0x7e, 0x35, 0x87, 0xa8, 0x7f, 0xc5, 0x24, 0xe9,
		0xe6, 0x98, 0x2c, 0x18, 0x34, 0xfd, 0x14, 0x4f, 0xe9, 0x1d, 0x58, 0xed, 0xc6, 0xa2, 0xff, 0x38,
		0x5f, 0x3a, 0x6b, 0x77, 0x93, 0x62, 0x48, 0x65, 0x87, 0x73, 0xf7, 0x73, 0xe7, 0xac, 0x61, 0xe1,
		0x18, 0xb8, 0xec, 0xd8, 0x2c, 0xca, 0x12, 0xe3, 0xf0, 0x59, 0x08, 0x47, 0x68, 0x4b, 0x0e, 0x2e,
		0x37, 0x9c, 0xef, 0x36, 0x3b, 0x67, 0xe4, 0x8e, 0xb8, 0xb7, 0xe9, 0x33, 0x25, 0xef, 0x03, 0xf5,
		0x98, 0xc3, 0x3a, 0x5b, 0xc3, 0x32, 0x35, 0x43, 0xe5, 0xc8, 0xe3, 0x90, 0xde, 0x0c, 0x72, 0x5b,
		0x28, 0x07, 0x75, 0xbb, 0x54, 0x87, 0x74, 0x4e, 0xc3, 0x11, 0xcb, 0x46, 0x3c, 0x25, 0xc7, 0x65,
		0xb1, 0x1e, 0x2c, 0xe1, 0xe0, 0x58, 0xb4, 0xfd, 0x8f, 0xd5, 0x76, 0x70, 0x49, 0x0e, 0xaa, 0x1f,
		0xe5, 0xc8, 0x61, 0x49, 0xf5, 0x4b, 0x38, 0xa8, 0x7f, 0x1d, 0xca, 0xc2, 0xf5, 0xa3, 0xbb, 0x40,
		0xfe, 0x35, 0x08, 0x83, 0x66, 0x17, 0x5e, 0xd7, 0x7d, 0x23, 0xfa, 0xf6, 0xe4, 0xf8, 0xdb, 0x91,
		0x62, 0x18, 0xb3, 0x58, 0xf3, 0xf8, 0x4c, 0x29, 0x03, 0xc5, 0x63, 0xe1, 0x54, 0xcf, 0x07, 0x7d,
		0x11, 0xf7, 0x9b,
	}

	chunkHeader := ChunkHeader{
		Size: uint32(len(data)) + ChunkHeaderSize,
		Type: TilesetChunkHex,
	}

	tmp, err := os.CreateTemp("", "chunk_tileset_test.aseprite")
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := tmp.Write(data); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	if _, err := tmp.Seek(0, 0); err != nil {
		t.Fatalf("failed to seek in temp file: %v", err)
	}

	loader := &Loader{Buffer: new(bytes.Buffer), Reader: tmp, Buf: make([]byte, len(data))}

	chunk, err := loader.ParseChunkTileset(chunkHeader)

	if err != nil {
		t.Fatalf("failed to parse ChunkTileset: %v", err)
	}

	if chunk.GetType() != TilesetChunkHex {
		t.Errorf("unexpected chunk type: got %d, want %d", chunk.GetType(), TilesetChunkHex)
	}

	chunkTileset := chunk.(*ChunkTileset)
	
	if chunkTileset.Name != "Terrain" {
		t.Errorf("unexpected name: got %s, want %s", chunkTileset.Name, "Terrain")
	}

	if !chunkTileset.Flags.LinkExternalFile || !chunkTileset.Flags.LinkTiles || !chunkTileset.Flags.UseTileID0 || !chunkTileset.Flags.AutoFlipX || !chunkTileset.Flags.AutoFlipY || !chunkTileset.Flags.AutoFlipD {
		t.Errorf("unexpected layer flags: got (%t, %t, %t, %t, %t, %t), want (true, true, true, true, true, true)", chunkTileset.Flags.LinkExternalFile, chunkTileset.Flags.LinkTiles, chunkTileset.Flags.UseTileID0, chunkTileset.Flags.AutoFlipX, chunkTileset.Flags.AutoFlipY, chunkTileset.Flags.AutoFlipD)
	}

	if chunkTileset.ChunkTilesetLinkExternalFileData.ExternalFileID != 42 {
		t.Errorf("unexpected external file id: got %d, want %d", chunkTileset.ChunkTilesetLinkExternalFileData.ExternalFileID, 42)
	}
}

func TestChunkTag(t *testing.T) {
	data := []byte{
		0x01, 0x00,             // Number of tags = 1
		0x00, 0x00, 0x00, 0x00, // Reserved (8 bytes)
		0x00, 0x00, 0x00, 0x00,

		// Tag 1
		0x00, 0x00,             // From frame = 0
		0x05, 0x00,             // To frame = 5
		0x02,                   // Loop direction = 2 (ping-pong)
		0x03, 0x00,             // Repeat = 3

		// Reserved 6 bytes
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,

		// RGB color (deprecated)
		0xFF, 0x00, 0x80,       // RGB = (255, 0, 128)
		0x00,                   // Extra byte (zero)

		// STRING: "Run"
		0x03, 0x00,             // Length = 3
		'R', 'u', 'n',
	}

	chunkHeader := ChunkHeader{
		Size: uint32(len(data)) + ChunkHeaderSize,
		Type: TagsChunkHex,
	}

	tmp, err := os.CreateTemp("", "chunk_tag_test.aseprite")
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := tmp.Write(data); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	if _, err := tmp.Seek(0, 0); err != nil {
		t.Fatalf("failed to seek in temp file: %v", err)
	}

	loader := &Loader{Buffer: new(bytes.Buffer), Reader: tmp, Buf: make([]byte, len(data))}

	chunk, err := loader.ParseChunkTag(chunkHeader)

	if err != nil {
		t.Fatalf("failed to parse ChunkTag: %v", err)
	}

	if chunk.GetType() != TagsChunkHex {
		t.Errorf("unexpected chunk type: got %d, want %d", chunk.GetType(), TagsChunkHex)
	}

	chunkTag := chunk.(*ChunkTag)

	if len(chunkTag.Entries) != 1 {
		t.Errorf("unexpected number of chunk tag entries: got %d, want %d", len(chunkTag.Entries), 1)
	}

	if chunkTag.Entries[0].Name != "Run" {
		t.Errorf("unexpected chunk tag entry name: got %s, want %s", chunkTag.Entries[0].Name, "Run")
	}

	if chunkTag.Entries[0].Name != "Run" {
		t.Errorf("unexpected chunk tag entry name: got %s, want %s", chunkTag.Entries[0].Name, "Run")
	}

	if chunkTag.Entries[0].LoopAnimationType != 2 {
		t.Errorf("unexpected chunk tag entry loop type: got %d, want %d", chunkTag.Entries[0].LoopAnimationType, 2)
	}
}

func TestChunkSlice(t *testing.T) {
	data := []byte{
		0x01, 0x00, 0x00, 0x00, // Number of slice keys = 1
		0x03, 0x00, 0x00, 0x00, // Flags = 3 (has 9-patch + pivot)
		0x00, 0x00, 0x00, 0x00, // Reserved

		0x06, 0x00, // Name length = 7
		'H', 'P', ' ', 'B', 'a', 'r',

		// Slice key
		0x00, 0x00, 0x00, 0x00, // Frame = 0
		0x0A, 0x00, 0x00, 0x00, // X = 10
		0x14, 0x00, 0x00, 0x00, // Y = 20
		0x64, 0x00, 0x00, 0x00, // Width = 100
		0x1E, 0x00, 0x00, 0x00, // Height = 30

		0x05, 0x00, 0x00, 0x00, // Center X = 5
		0x05, 0x00, 0x00, 0x00, // Center Y = 5
		0x5A, 0x00, 0x00, 0x00, // Center Width = 90
		0x14, 0x00, 0x00, 0x00, // Center Height = 20

		0x32, 0x00, 0x00, 0x00, // Pivot X = 50
		0x0F, 0x00, 0x00, 0x00, // Pivot Y = 15
	}

	chunkHeader := ChunkHeader{
		Size: uint32(len(data)) + ChunkHeaderSize,
		Type: SliceChunkHex,
	}

	tmp, err := os.CreateTemp("", "chunk_slice_test.aseprite")
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := tmp.Write(data); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	if _, err := tmp.Seek(0, 0); err != nil {
		t.Fatalf("failed to seek in temp file: %v", err)
	}

	loader := &Loader{Buffer: new(bytes.Buffer), Reader: tmp, Buf: make([]byte, len(data))}

	chunk, err := loader.ParseChunkSlice(chunkHeader)

	if err != nil {
		t.Fatalf("failed to parse ChunkSlice: %v", err)
	}

	if chunk.GetType() != SliceChunkHex {
		t.Errorf("unexpected chunk type: got %d, want %d", chunk.GetType(), SliceChunkHex)
	}

	chunkSlice := chunk.(*ChunkSlice)
	
	if chunkSlice.Name != "HP Bar" {
		t.Errorf("unexpected name: got %s, want %s", chunkSlice.Name, "HP Bar")
	}

	if chunkSlice.Keys[0].OriginX != 10 {
		t.Errorf("unexpected x key value: got %d, want %d", chunkSlice.Keys[0].X, 10)
	}

	if chunkSlice.Keys[0].ChunkSliceKeyPivotData.X != 50 {
		t.Errorf("unexpected x pivot key value: got %d, want %d", chunkSlice.Keys[0].ChunkSliceKeyPivotData.X, 50)
	}

	if chunkSlice.Keys[0].ChunkSliceKey9PatchesData.CenterWidth != 90 {
		t.Errorf("unexpected slice nine patches center width: got %d, want %d", chunkSlice.Keys[0].ChunkSliceKey9PatchesData.CenterWidth, 90)
	}
}

func TestChunkExternalFiles(t *testing.T) {
	data := []byte{
		0x02, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,

		0x2A, 0x00, 0x00, 0x00,
		0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x10, 0x00,
		'p', 'a', 'l', 'e', 't', 't', 'e', '.', 'a', 's', 'e', 'p', 'r', 'i', 't', 'e',

		0x2B, 0x00, 0x00, 0x00,
		0x01,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x0B, 0x00,
		't', 'i', 'l', 'e', 's', 'e', 't', '.', 't', 's', 'x',
	}

	chunkHeader := ChunkHeader{
		Size: uint32(len(data)) + ChunkHeaderSize,
		Type: ExternalFilesChunkHex,
	}

	tmp, err := os.CreateTemp("", "chunk_external_files_test.aseprite")
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := tmp.Write(data); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	if _, err := tmp.Seek(0, 0); err != nil {
		t.Fatalf("failed to seek in temp file: %v", err)
	}

	loader := &Loader{Buffer: new(bytes.Buffer), Reader: tmp, Buf: make([]byte, len(data))}

	chunk, err := loader.ParseChunkExternalFiles(chunkHeader)

	if err != nil {
		t.Fatalf("failed to parse ChunkExternalFiles: %v", err)
	}

	if chunk.GetType() != ExternalFilesChunkHex {
		t.Errorf("unexpected chunk type: got %d, want %d", chunk.GetType(), ExternalFilesChunkHex)
	}

	chunkExternalFiles := chunk.(*ChunkExternalFiles)
	
	if chunkExternalFiles.NumberEntries != 2 {
		t.Errorf("unexpected number entries value: got %d, want %d", chunkExternalFiles.NumberEntries, 2)
	}

	if chunkExternalFiles.Entries[0].ID != 42 {
		t.Errorf("unexpected entry id: got %d, want %d", chunkExternalFiles.Entries[0].ID, 42)
	}

	if chunkExternalFiles.Entries[0].Type != 0 {
		t.Errorf("unexpected entry type: got %d, want %d", chunkExternalFiles.Entries[0].Type, 0)
	}

	if chunkExternalFiles.Entries[0].Name != "palette.aseprite" {
		t.Errorf("unexpected entry name: got %s, want %s", chunkExternalFiles.Entries[0].Name, "palette.aseprite")
	}
}

func TestChunkCelExtra(t *testing.T) {
	data := []byte{
		0x01, 0x00, 0x00, 0x00,
		0x00, 0x80, 0xFE, 0xFF,
		0x00, 0x40, 0x02, 0x00,
		0x00, 0x00, 0x20, 0x00,
		0x00, 0x00, 0x20, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}

	chunkHeader := ChunkHeader{
		Size: uint32(len(data)) + ChunkHeaderSize,
		Type: CelExtraChunkHex,
	}

	tmp, err := os.CreateTemp("", "chunk_cel_extra_test.aseprite")
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := tmp.Write(data); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	if _, err := tmp.Seek(0, 0); err != nil {
		t.Fatalf("failed to seek in temp file: %v", err)
	}

	loader := &Loader{Buffer: new(bytes.Buffer), Reader: tmp, Buf: make([]byte, len(data))}

	chunk, err := loader.ParseChunkCelExtra(chunkHeader)

	if err != nil {
		t.Fatalf("failed to parse ChunkPalette: %v", err)
	}

	if chunk.GetType() != CelExtraChunkHex {
		t.Errorf("unexpected chunk type: got %d, want %d", chunk.GetType(), CelExtraChunkHex)
	}

	chunkCelExtra := chunk.(*ChunkCelExtra)

	if chunkCelExtra.Flags != 1 {
		t.Errorf("unexpected flags value: got %d, want %d", chunkCelExtra.Flags, 1)
	}

	if chunkCelExtra.X.FixedToFloat() != -1.5 {
		t.Errorf("unexpected X value: got %f, want %f", chunkCelExtra.X.FixedToFloat(), -1.5)
	}

	if chunkCelExtra.Y.FixedToFloat() != 2.25 {
		t.Errorf("unexpected Y value: got %f, want %f", chunkCelExtra.Y.FixedToFloat(), 2.25)
	}

	if chunkCelExtra.Width.FixedToFloat() != 32.0 {
		t.Errorf("unexpected Width value: got %f, want %f", chunkCelExtra.Width.FixedToFloat(), 32.0)
	}

	if chunkCelExtra.Height.FixedToFloat() != 32.0 {
		t.Errorf("unexpected Height value: got %f, want %f", chunkCelExtra.Height.FixedToFloat(), 32.0)
	}
}


func TestChunkPalette(t *testing.T) {
	data := []byte{
		0x03, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x02, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		// --- Entry 0 ---
		0x00, 0x00, // flags
		0xFF, 0x00, 0x00, // red
		// --- Entry 1 ---
		0x00, 0x00, // flags
		0x00, 0xFF, 0x00, // green
		// --- Entry 2 ---
		0x01, 0x00, // flags (has name)
		0x00, 0x00, 0xFF, // blue
		// "Azulão" + null
		0x06, 0x00, 'a', 'z', 'u', 'l', 'a', 'o',
	}

	chunkHeader := ChunkHeader{
		Size: uint32(len(data)) + ChunkHeaderSize,
		Type: PaletteChunkHex,
	}

	tmp, err := os.CreateTemp("", "chunk_palette_test.aseprite")
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := tmp.Write(data); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	if _, err := tmp.Seek(0, 0); err != nil {
		t.Fatalf("failed to seek in temp file: %v", err)
	}

	loader := &Loader{Buffer: new(bytes.Buffer), Reader: tmp, Buf: make([]byte, len(data))}

	chunk, err := loader.ParseChunkPalette(chunkHeader)

	if err != nil {
		t.Fatalf("failed to parse ChunkPalette: %v", err)
	}

	if chunk.GetType() != PaletteChunkHex {
		t.Errorf("unexpected palette type: got %d, want %d", chunk.GetType(), PaletteChunkHex)
	}

	chunkPalette := chunk.(*ChunkPalette)

	if unread := loader.Buffer.Len(); unread != 0 {
		t.Errorf("expected ChunkPalette to be fully read, but %d bytes remain (read %d of %d)", unread, len(data)-unread, len(data))
	}

	if int(chunkPalette.EntriesNumber) != len(chunkPalette.Entries) {
		t.Errorf("unexpected entries number: got %d, want %d", len(chunkPalette.Entries), chunkPalette.EntriesNumber)
	}

	if chunkPalette.From != 0 {
		t.Errorf("unexpected from: got %d, want %d", chunkPalette.From, 0)
	}

	if chunkPalette.To != 2 {
		t.Errorf("unexpected to: got %d, want %d", chunkPalette.To, 2)
	}

	if chunkPalette.Entries[2].ColorName != "azulao" {
		t.Errorf("unexpected entry name: got %s, want %s", chunkPalette.Entries[2].ColorName, "Azulão")
	}
}

func TestChunkLayer(t *testing.T) {
	data := []byte{
		0x0B, 0x00,
		0x00, 0x00,
		0x00, 0x00,
		0x00, 0x00,
		0x00, 0x00,
		0x00, 0x00,
		0xFF,
		0x00, 0x00, 0x00,
		0x05, 0x00,
		'L', 'a', 'y', 'e', 'r',
	}

	chunkHeader := ChunkHeader{
		Size: uint32(len(data)) + ChunkHeaderSize,
		Type: LayerChunkHex,
	}

	tmp, err := os.CreateTemp("", "chunk_layer_test.aseprite")
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := tmp.Write(data); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	if _, err := tmp.Seek(0, 0); err != nil {
		t.Fatalf("failed to seek in temp file: %v", err)
	}

	loader := &Loader{Buffer: new(bytes.Buffer), Reader: tmp, Buf: make([]byte, len(data))}

	chunk, err := loader.ParseChunkLayer(chunkHeader)

	if err != nil {
		t.Fatalf("failed to parse ChunkLayer: %v", err)
	}

	if chunk.GetType() != LayerChunkHex {
		t.Errorf("unexpected layer type: got %d, want %d", chunk.GetType(), LayerChunkHex)
	}

	chunkLayer := chunk.(*ChunkLayer)

	if unread := loader.Buffer.Len(); unread != 0 {
		t.Errorf("expected ChunkLayer to be fully read, but %d bytes remain (read %d of %d)", unread, len(data)-unread, len(data))
	}

	if chunkLayer.ChunkLayerData.FlagsBit != 0x000B {
		t.Errorf("unexpected layer flags: got 0x%X, want 0x000B", chunkLayer.ChunkLayerData.FlagsBit)
	}

	if !chunkLayer.Visible || !chunkLayer.Editable || !chunkLayer.Background {
		t.Errorf("unexpected layer flags: got (Visible: %t, Editable: %t, Background: %t), want (Visible: true, Editable: true and Background: true)", chunkLayer.Visible, chunkLayer.Editable, chunkLayer.Background)
	}

	if chunkLayer.ChunkLayerName != "Layer" {
		t.Errorf("unexpected layer name: got %q, want \"Layer\"", chunkLayer.ChunkLayerName)
	}
}

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
		t.Errorf("expected ChunkOldPalette to be fully read, but %d bytes remain (read %d of %d)", unread, len(data)-unread, len(data))
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

func TestSpriteSheet(t *testing.T) {
	test, err := os.Open("test.aseprite")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer test.Close()

	ase, err := DeserializeFile(test)
	ss, err := ase.SpriteSheet()

	fd, err := os.Create("aa.png")
	defer fd.Close()

	png.Encode(fd, ss)
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

	expectedChunks := []int{5, 1, 1, 1, 1, 1, 1, 1}
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

	if fileSize-read > 16 {
		t.Errorf("expected file to be fully read, but %d bytes remain (read %d of %d)", fileSize-read, read, fileSize)
	}
}

package ase

import (
	"bytes"
	"compress/zlib"
	"image/color"
	"os"
	"testing"

	"github.com/starillume/ase/chunk"
	"github.com/starillume/ase/pixel"
)

// --- Helpers ---

func fakeRawFrame() *frame {
	p := []byte{255, 0, 0, 255}

	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	zw.Write(p)
	zw.Close()

	pc := buf.Bytes()


	c := &chunk.ChunkCelCompressedImage{
		ChunkCelData: chunk.ChunkCelData{
			LayerIndex: 0,
			X:          0,
			Y:          0,
			Opacity:    100,
			CelType:    chunk.CelTypeCompressedImage,
			Z:          0,
		},
		ChunkCelCompressedImageData: chunk.ChunkCelCompressedImageData{
			ChunkCelDimensionData: chunk.ChunkCelDimensionData{
				Width:  1,
				Height: 1,
			},
			Pixels: pixel.PixelsZlib(pc),
		},
	}

	return &frame{
		Cels: []*cel{
			{Chunk: c},
		},
		Header: FrameHeader{
			FrameDuration: 100,
		},
	}
}

func fakeRawLayer() *layer {
	c := &chunk.Layer{
		ChunkLayerData: chunk.ChunkLayerData{
			FlagsBit:      0x00,
			Type:          0,
			ChildLevel:    0,
			DefaultWidth:  1,
			DefaultHeight: 1,
			BlendMode:     0,
			Opacity:       0,
			NameLength:    1,
		},
		ChunkLayerName: "l",
		ChunkLayerFlags: chunk.ChunkLayerFlags{
			Visible: true,
		},
	}

	return &layer{
		Chunk: c,
	}
}

func fakeRawTag() *tag {
	c := &chunk.Tag{
		Entries: []chunk.TagEntry{
			{
				Name: "tag1",
				TagEntryData: chunk.TagEntryData{
					FromFrame:         1,
					ToFrame:           3,
					LoopAnimationType: chunk.LoopAnimationForward,
					Repeat:            16,
					Color:             [3]byte{255, 0, 0},
					TagNameSize:       4,
				},
			},
		},
	}

	return &tag{
		Chunk: c,
	}
}

// --- Tests ---

func TestCreateFrames(t *testing.T) {
	rawFrames := []*frame{fakeRawFrame()}
	frames, images := createFrames(rawFrames, Header{Width: 10, Height: 10, ColorDepth: 32})
	if len(frames) != 1 {
		t.Fatalf("Expected 1 frame, got %d", len(frames))
	}
	if frames[0].Duration != 100 {
		t.Errorf("Expected duration 100, got %d", frames[0].Duration)
	}
	if len(images) != 1 || images[0] == nil {
		t.Errorf("Expected frame image to be created")
	}
}

func TestCreateCel(t *testing.T) {
	c, err := createCel(fakeRawFrame().Cels[0], 0, 10, 10, 32)
	if err != nil {
		t.Fatalf("createCel failed: %v", err)
	}
	if c.FrameIndex != 0 {
		t.Errorf("Expected FrameIndex 0, got %d", c.FrameIndex)
	}
}

func TestCreateFrameImage(t *testing.T) {
	_, frameImage := createFrames(
		[]*frame{fakeRawFrame()},
		Header{Width: 10, Height: 10, ColorDepth: 32},
	)

	cel := &Cel{
		Image: frameImage[0],
	}
	img := createFrameImage([]*Cel{cel})
	if img == nil {
		t.Errorf("Expected image, got nil")
	}
}

func TestCreateTags(t *testing.T) {
	rawTags := []*tag{fakeRawTag()}
	frames := []*Frame{{Index: 0}, {Index: 1}, {Index: 2}, {Index: 3}, {Index: 4}}
	tags := createTags(rawTags, frames)
	if len(tags) != 1 {
		t.Fatalf("Expected 1 tag, got %d", len(tags))
	}
	if tags[0].Name != "tag1" {
		t.Errorf("Expected tag name 'tag1', got %s", tags[0].Name)
	}
	c := tags[0].Color.(color.RGBA)
	if c.R != 255 || c.G != 0 || c.B != 0 {
		t.Errorf("Expected color red, got %v", c)
	}
}

func TestCreateLayerFrameImages(t *testing.T) {
	_, frameImage := createFrames(
		[]*frame{fakeRawFrame()},
		Header{Width: 10, Height: 10, ColorDepth: 32},
	)

	frames := []*Frame{
		{Cels: []*Cel{{LayerIndex: 0, Image: frameImage[0]}}},
	}
	layerImages := createLayerFrameImages(0, frames)
	if len(layerImages) != 1 {
		t.Errorf("Expected 1 layer image, got %d", len(layerImages))
	}
}

func TestCreateLayers(t *testing.T) {
	_, frameImage := createFrames(
		[]*frame{fakeRawFrame()},
		Header{Width: 10, Height: 10, ColorDepth: 32},
	)

	rawLayers := []*layer{fakeRawLayer()}
	frames := []*Frame{{Cels: []*Cel{{LayerIndex: 0, Image: frameImage[0]}}}}
	layers, groups := createLayers(rawLayers, frames)
	if len(layers) != 1 {
		t.Errorf("Expected 1 layer, got %d", len(layers))
	}
	if len(groups) != 0 {
		t.Errorf("Expected 0 groups, got %d", len(groups))
	}
}

func TestCreateAseprite(t *testing.T) {
	raw := &rawAseprite{
		Header: Header{Width: 10, Height: 10, ColorDepth: 32},
		Frames: []*frame{fakeRawFrame(), fakeRawFrame(), fakeRawFrame(), fakeRawFrame()},
		Layers: []*layer{fakeRawLayer()},
		Tags:   []*tag{fakeRawTag()},
	}
	ase, err := createAseprite(raw)
	if err != nil {
		t.Fatalf("createAseprite failed: %v", err)
	}
	if len(ase.Frames) != 4 {
		t.Errorf("Expected 4 frame, got %d", len(ase.Frames))
	}
	if len(ase.Layers) != 1 {
		t.Errorf("Expected 1 layer, got %d", len(ase.Layers))
	}
	if len(ase.Tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(ase.Tags))
	}
}

const testFilePath = "./test.aseprite"

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
	// verifyAllDataRead(t, ase, testFilePath)
}
func verifyHeader(t *testing.T, ase *Aseprite) {
	if ase.Header.MagicNumber != 0xA5E0 {
		t.Errorf("invalid header magic number: got 0x%X, want 0xA5E0", ase.Header.MagicNumber)
	}
}

func verifyFrames(t *testing.T, ase *Aseprite) {
	expectedFrameCount := 7
	if len(ase.Frames) != expectedFrameCount {
		t.Errorf("expected %d frames, got %d", expectedFrameCount, len(ase.Frames))
	}
}

// func verifyAllDataRead(t *testing.T, ase *Aseprite, filepath string) {
// 	stat, err := os.Stat(filepath)
// 	if err != nil {
// 		t.Fatalf("failed to stat file %s: %v", filepath, err)
// 	}
// 	fileSize := stat.Size()
//
// 	read := int64(HeaderSize)
//
// 	for _, frame := range ase.Frames {
// 		read += int64(FrameHeaderSize)
// 		for _, chunk := range frame.Chunks {
// 			read += int64(chunk.GetHeader().Size)
// 		}
// 	}
//
// 	if fileSize-read > 16 {
// 		t.Errorf("expected file to be fully read, but %d bytes remain (read %d of %d)", fileSize-read, read, fileSize)
// 	}
// }

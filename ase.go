package ase

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"sort"

	"github.com/starillume/ase/chunk"
	"github.com/starillume/ase/pixel"
)

// TODO: userdata
// not sure how its type would be defined tho

type Aseprite struct {
	Header Header
	Palette pixel.Palette
	Frames []*Frame
	FrameImages map[int]image.Image
	Tags   []*Tag
	Layers []*Layer
	Groups []*LayerGroup
}

type Frame struct {
	Index int
	Duration int // ms
	Cels []*Cel
	Image image.Image
}

type Layer struct {
	Name string
	Index int
	Visible bool
	Opacity float64
	BlendMode string
	FrameImages map[int]image.Image
}

type Cel struct {
	LayerIndex int
	FrameIndex int
	Image image.Image
}

type Tag struct {
	Name              string
	From              int
	To                int
	Frames            []*Frame
	LoopAnimationType chunk.LoopAnimationType
	Repeat            int
	Color             color.Color
	UserData          any
}

type LayerGroup struct {
	Name string
	Layers []*Layer
}

func createFrames(rawFrames []*frame, rawHeader Header) ([]*Frame, map[int]image.Image)  {
	frames := make([]*Frame, 0)
	frameImages := make(map[int]image.Image, 0)
	for i, rawFrame := range rawFrames {
		cels := make([]*Cel, 0)
		for _, cel := range rawFrame.Cels {
			newCel, _ := createCel(cel, i, int(rawHeader.Width), int(rawHeader.Height))
			cels = append(cels, newCel)
		}

		frame := &Frame{
			Index:    i,
			Duration: int(rawFrame.Header.FrameDuration),
			Cels:     cels,
			Image:    createFrameImage(cels),
		}

		frameImages[i] = frame.Image
		frames = append(frames, frame)
	}

	return frames, frameImages
}

func createCel(cel *cel, frameIndex int, canvasWidth int, canvasHeight int) (*Cel, error) {
	var layerIndex int
	var img image.Image
	switch (*cel.Chunk).(chunk.Cel).CelType() {
		case chunk.CelTypeCompressedImage, chunk.CelTypeRawImage:
			celImageChunk := (*cel.Chunk).(chunk.ChunkCelImage)
			layerIndex = int(celImageChunk.LayerIndex)
			img = celImageChunk.ChunkCelRawImageData.Pixels.ToImage(pixel.PixelToImageOpts{
				CelX: int(celImageChunk.X),
				CelY: int(celImageChunk.Y),
				Width: int(celImageChunk.Width),
				Height: int(celImageChunk.Height),
				CanvasWidth: canvasWidth,
				CanvasHeight: canvasHeight,
				// Palette: palette,
			})
		// TODO:
		// case chunk.CelTypeLinked:
		// case chunk.CelTypeCompressedTilemap:
		default:
			return &Cel{}, nil
	}

	return &Cel{
		LayerIndex: layerIndex,
		FrameIndex: frameIndex,
		Image: img,
	}, nil
}

func createFrameImage(frameCels []*Cel) image.Image {
	if len(frameCels) == 0 {
		return nil
	}
	
	// making sure that it draws respecting layer order
	sort.Slice(frameCels, func(i, j int) bool {
		return frameCels[i].LayerIndex < frameCels[j].LayerIndex
	})

	bounds := frameCels[0].Image.Bounds()
	img := image.NewRGBA(bounds)

	for _, cel := range frameCels {
		draw.Draw(img, bounds, cel.Image, bounds.Min, draw.Over)
	}

	return img
}

func createTags(rawTags []*tag, frames []*Frame) []*Tag {
	tags := make([]*Tag, 0)
	for _, rawTag := range rawTags {
		for _, entry := range rawTag.Chunk.Entries {
			from := int(entry.FromFrame)
			to := int(entry.ToFrame)
			tag := &Tag{
				Name: entry.Name,
				From: from,
				To: to,
				Frames: frames[from:to],
				LoopAnimationType: entry.LoopAnimationType,
				Repeat: int(entry.Repeat),
				Color: color.RGBA{entry.Color[0], entry.Color[1], entry.Color[2], 255},
			}

			tags = append(tags, tag)
		}
	}

	return tags
}

func createLayerFrameImages(layerIndex int, frames []*Frame) map[int]image.Image {
	layerFrameImages := make(map[int]image.Image)
	for _, frame := range frames {
		layeredCels := make([]*Cel, 0)
		for _, cel := range frame.Cels {
			if cel.LayerIndex == layerIndex {
				layeredCels = append(layeredCels, frame.Cels...)
			}
		}

		bounds := layeredCels[0].Image.Bounds()
		img := image.NewRGBA(bounds)

		for _, cel := range layeredCels {
			draw.Draw(img, bounds, cel.Image, bounds.Min, draw.Over)
		}

		layerFrameImages[frame.Index] = img
	}

	return layerFrameImages
}

func createLayers(rawLayers []*layer, frames []*Frame) ([]*Layer, []*LayerGroup) {
	layers := make([]*Layer, 0)
	layerGroups := make([]*LayerGroup, 0)
	groupStack := make([]*LayerGroup, 0)

	for i, rawLayer := range rawLayers {
		layerFrameImages := createLayerFrameImages(i, frames)

		layer := &Layer{
			Name:        string(rawLayer.Chunk.ChunkLayerName),
			Index:       i,
			Visible:     rawLayer.Chunk.Visible,
			Opacity:     float64(rawLayer.Chunk.ChunkLayerData.Opacity),
			FrameImages: layerFrameImages,

			// FIX: preguiça
			BlendMode:   "Normal",
		}
		layers = append(layers, layer)

		switch rawLayer.Chunk.ChunkLayerData.Type {
		case 1:
			group := &LayerGroup{
				Name:   string(rawLayer.Chunk.ChunkLayerName),
				Layers: []*Layer{},
			}
			layerGroups = append(layerGroups, group)
			groupStack = append(groupStack, group)

		default:
			if len(groupStack) > 0 {
				currentGroup := groupStack[len(groupStack)-1]
				currentGroup.Layers = append(currentGroup.Layers, layer)
			}
		}

		for len(groupStack) > 0 {
			if nextIndex := i + 1; nextIndex < len(rawLayers) {
				nextLayer := rawLayers[nextIndex]
				if nextLayer.Chunk.ChunkLayerData.ChildLevel < rawLayer.Chunk.ChunkLayerData.ChildLevel {
					groupStack = groupStack[:len(groupStack)-1]
				} else {
					break
				}
			} else {
				break
			}
		}
	}

	return layers, layerGroups
}
func createAseprite(raw *rawAseprite) (*Aseprite, error) {
	frames, frameImages := createFrames(raw.Frames, raw.Header)
	layers, groups := createLayers(raw.Layers, frames)
	tags := createTags(raw.Tags, frames)

	return &Aseprite{
		Header: raw.Header,
		// Palette: yeah i dont know bro
		Frames: frames,
		FrameImages: frameImages,
		Tags: tags,
		Layers: layers,
		Groups: groups,
	}, nil
}

func DeserializeFile(fd *os.File) (*Aseprite, error) {
	loader := NewLoader(fd)

	var headerBytes = make([]byte, HeaderSize)

	err := loader.BytesToStructV2(HeaderSize, headerBytes)
	if err != nil {
		return nil, err
	}

	header, err := ParseHeader(headerBytes)
	if err != nil {
		return nil, err
	}

	loader.Ase.Header = *header

	if err := loader.ParseFrames(); err != nil {
		return nil, err
	}

	fmt.Printf("ase: %+v\n", loader.Ase)
	// fmt.Printf("tags: %+v\n", loader.Ase.Tags)
	// fmt.Printf("tag: %+v\n", loader.Ase.Tags[0])
	// fmt.Printf("frame: %+v\n", loader.Ase.Frames[0])

	return createAseprite(loader.Ase)
}

func (a *Aseprite) SpriteSheet() (image.Image, error) {
	return joinImagesHorizontally(a.FrameImages), nil
}

func joinImagesHorizontally(images map[int]image.Image) image.Image {
	totalWidth := 0
	maxHeight := 0
	for _, img := range images {
		bounds := img.Bounds()
		totalWidth += bounds.Dx()
		if bounds.Dy() > maxHeight {
			maxHeight = bounds.Dy()
		}
	}

	dst := image.NewRGBA(image.Rect(0, 0, totalWidth, maxHeight))

	draw.Draw(dst, dst.Bounds(), &image.Uniform{color.Transparent}, image.Point{}, draw.Src)

	xOffset := 0
	for _, img := range images {
		bounds := img.Bounds()
		pos := image.Rect(xOffset, 0, xOffset+bounds.Dx(), bounds.Dy())
		draw.Draw(dst, pos, img, bounds.Min, draw.Over)
		xOffset += bounds.Dx()
	}

	return dst
}

package ase

import (
	"fmt"
	"image/color"
	"os"

	"github.com/starillume/ase/chunk"
	"github.com/starillume/ase/pixel"
)

type AsepriteFile struct {
	// CanvasWidth
	// CanvasHeight
	// ColorDepth
	//
	// PixelWidth
	// PixelHeight

	Header Header

	Palette pixel.Palette

	Frames []*Frame
	Tags   []*Tag
	Layers []*Layer
	Groups []*GroupLayers
}

type Frame struct {
	Duration int // ms
	Cels     []*Cel
}

type Cel struct {
	Layer *Layer
	Frame *Frame
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

type Layer struct {
	Cels []*Cel
}

type GroupLayers struct {
	Layers []*Layer
}

func DeserializeFile(fd *os.File) (*AsepriteFile, error) {
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
	fmt.Printf("tags: %+v\n", loader.Ase.Tags)
	// fmt.Printf("tag: %+v\n", loader.Ase.Tags[0])
	// fmt.Printf("frame: %+v\n", loader.Ase.Frames[0])

	return loader.Ase, nil
}

// func (a *AsepriteFile) SpriteSheet() (image.Image, error) {
// 	sprites := make([]image.Image, 0)
//
// 	for _, frame := range a.Frames {
// 		sprite := image.NewRGBA(image.Rect(0, 0, int(a.Header.Width), int(a.Header.Height)))
//
// 		for _, c := range frame.Chunks {
// 			switch c.(type) {
// 			case *chunk.ChunkCelImage:
// 				c := chunk.(*ChunkCelImage)
// 				pixels := c.ChunkCelRawImageData.Pixels.(PixelsRGBA)
// 				img := pixels.ToImage(int(c.X), int(c.Y), int(c.ChunkCelDimensionData.Width), int(c.ChunkCelDimensionData.Height), int(a.Header.Width), int(a.Header.Height))
// 				draw.Draw(sprite, sprite.Bounds(), img, sprite.Bounds().Min, draw.Over)
// 			}
// 		}
//
// 		sprites = append(sprites, sprite)
// 	}
//
// 	spriteSheet := joinImagesHorizontally(sprites)
//
// 	return spriteSheet, nil
// }
//
// func joinImagesHorizontally(images []image.Image) image.Image {
// 	totalWidth := 0
// 	maxHeight := 0
// 	for _, img := range images {
// 		bounds := img.Bounds()
// 		totalWidth += bounds.Dx()
// 		if bounds.Dy() > maxHeight {
// 			maxHeight = bounds.Dy()
// 		}
// 	}
//
// 	dst := image.NewRGBA(image.Rect(0, 0, totalWidth, maxHeight))
//
// 	draw.Draw(dst, dst.Bounds(), &image.Uniform{color.Transparent}, image.Point{}, draw.Src)
//
// 	xOffset := 0
// 	for _, img := range images {
// 		bounds := img.Bounds()
// 		pos := image.Rect(xOffset, 0, xOffset+bounds.Dx(), bounds.Dy())
// 		draw.Draw(dst, pos, img, bounds.Min, draw.Over)
// 		xOffset += bounds.Dx()
// 	}
//
// 	return dst
// }

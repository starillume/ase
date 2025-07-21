package ase

import (
	"image"
	"image/color"
	"image/draw"
)

// type AseFile struct {
// 	CanvasWidth
// 	CanvasHeight
// 	ColorDepth
//
// 	PixelWidth
// 	PixelHeight
//
// 	[]Frame
// 	[]Tags
// 	[]Layers
// }

type AsepriteFile struct {
	Header Header
	Frames []Frame
}

func joinImagesHorizontally(images []image.Image) image.Image {
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

func (a *AsepriteFile) SpriteSheet() (image.Image, error) {
	sprites := make([]image.Image, 0)

	for _, frame := range a.Frames {
		sprite := image.NewRGBA(image.Rect(0, 0, int(a.Header.Width), int(a.Header.Height)))

		for _, chunk := range frame.Chunks {
			switch chunk.(type) {
			case *ChunkCelImage:
				c := chunk.(*ChunkCelImage)
				pixels := c.ChunkCelRawImageData.Pixels.(PixelsRGBA)
				img := pixels.ToImage(int(c.X), int(c.Y), int(c.ChunkCelDimensionData.Width), int(c.ChunkCelDimensionData.Height), int(a.Header.Width), int(a.Header.Height))
				draw.Draw(sprite, sprite.Bounds(), img, sprite.Bounds().Min, draw.Over)
			}
		}

		sprites = append(sprites, sprite)
	}

	spriteSheet := joinImagesHorizontally(sprites)

	return spriteSheet, nil
}

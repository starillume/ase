package pixel

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"io"
)

type ColorDepth uint16

const (
	ColorDepthRGBA      ColorDepth = 32
	ColorDepthGrayscale ColorDepth = 16
	ColorDepthIndexed   ColorDepth = 8
)

type Palette interface {
	RevolveColor(index int) color.Color
}

type PixelToImageOpts struct {
	CelX, CelY, Width, Height, CanvasWidth, CanvasHeight int
	Palette Palette
}

type Pixels interface {
	ToImage(opts PixelToImageOpts) image.Image
}

type PixelsCompressed interface {
	Decompress() ([]byte, error)
}

type PixelsZlib []byte

func (p PixelsZlib) Decompress() ([]byte, error) {
	buffer := bytes.NewBuffer(p)
	r, err := zlib.NewReader(buffer)
	if err != nil {
		return nil, err
	}

	defer r.Close()

	d := new(bytes.Buffer)

	if _, err := io.Copy(d, r); err != nil {
		return nil, err
	}

	return d.Bytes(), nil
}

func ResolvePixelType(buf []byte, colorDepth ColorDepth) Pixels {
	switch colorDepth {
	case ColorDepthRGBA:
		return BytesToPixelsRGBA(buf)
	case ColorDepthGrayscale:
		return BytesToPixelsGrayscale(buf)
	case ColorDepthIndexed:
		return Indexed(buf)
	default:
		panic("unreachable: colordepth possibly not defined: " + fmt.Sprint(colorDepth))
	}
}

func BytesToPixelsRGBA(data []byte) RGBA {
	var chunks [][4]byte
	for i := 0; i < len(data); i += 4 {
		var block [4]byte
		copy(block[:], data[i:i+4])
		chunks = append(chunks, block)
	}
	return chunks
}

func BytesToPixelsGrayscale(data []byte) Grayscale {
	var chunks [][2]byte
	for i := 0; i < len(data); i += 2 {
		var block [2]byte
		copy(block[:], data[i:i+2])
		chunks = append(chunks, block)
	}
	return chunks
}

type RGBA [][4]byte

func (p RGBA) ToImage(opts PixelToImageOpts) image.Image {
	rect := image.Rect(0, 0, opts.CanvasWidth, opts.CanvasHeight)
	img := image.NewRGBA(rect)
	for y := range opts.Height {
		for x := range opts.Width {
			i := y*opts.Width + x
			color := color.RGBA{
				R: p[i][0],
				G: p[i][1],
				B: p[i][2],
				A: p[i][3],
			}
			img.Set(x+opts.CelX, y+opts.CelY, color)
		}
	}

	return img
}

type Grayscale [][2]byte

func (p Grayscale) ToImage(opts PixelToImageOpts) image.Image {
	rect := image.Rect(0, 0, opts.CanvasWidth, opts.CanvasHeight)
	img := image.NewRGBA(rect)
	for y := range opts.Height {
		for x := range opts.Width {
			i := y*opts.Width + x
			color := color.Gray16{
				Y: binary.NativeEndian.Uint16(p[i][:]),
			}
			img.Set(x+opts.CelX, y+opts.CelY, color)
		}
	}

	return img
}

type Indexed []byte

func (p Indexed) ToImage(opts PixelToImageOpts) image.Image {
	rect := image.Rect(0, 0, opts.CanvasWidth, opts.CanvasHeight)
	img := image.NewRGBA(rect)
	for y := range opts.Height {
		for x := range opts.Width {
			i := y*opts.Width + x
			color := opts.Palette.RevolveColor(i)

			img.Set(x+opts.CelX, y+opts.CelY, color)
		}
	}

	return img
}

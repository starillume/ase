package pixel

import (
	"bytes"
	"compress/zlib"
	"image/color"
	"testing"
)

type mockPalette struct{}

func (m mockPalette) RevolveColor(index int) color.Color {
	return color.RGBA{R: uint8(index), G: 0, B: 0, A: 255}
}

func zlibCompress(data []byte, t *testing.T) []byte {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	_, err := w.Write(data)
	if err != nil {
		t.Fatalf("compress error: %v", err)
	}
	w.Close()
	return buf.Bytes()
}

func TestPixelsZlib_Decompress(t *testing.T) {
	original := []byte("hello world")
	compressed := zlibCompress(original, t)

	pz := PixelsZlib(compressed)
	out, err := pz.Decompress()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(out, original) {
		t.Errorf("expected %q, got %q", original, out)
	}
}

func TestResolvePixelType(t *testing.T) {
	data := []byte{1, 2, 3, 4}

	if _, ok := ResolvePixelType(data, ColorDepthRGBA).(RGBA); !ok {
		t.Errorf("expected RGBA type")
	}

	if _, ok := ResolvePixelType(data, ColorDepthGrayscale).(Grayscale); !ok {
		t.Errorf("expected Grayscale type")
	}

	if _, ok := ResolvePixelType(data, ColorDepthIndexed).(Indexed); !ok {
		t.Errorf("expected Indexed type")
	}
}

func TestRGBA_ToImage(t *testing.T) {
	pixels := BytesToPixelsRGBA([]byte{
		255, 0, 0, 255, // red
		0, 255, 0, 255, // green
	})
	opts := PixelToImageOpts{
		Width: 2, Height: 1,
		CanvasWidth:  2,
		CanvasHeight: 1,
	}

	img := pixels.ToImage(opts)
	c1 := img.At(0, 0).(color.RGBA)
	c2 := img.At(1, 0).(color.RGBA)

	if c1.R != 255 || c2.G != 255 {
		t.Errorf("unexpected colors: got %v and %v", c1, c2)
	}
}

func TestGrayscale_ToImage(t *testing.T) {
	data := []byte{0xFF, 0x7F, 0x00, 0x00} // two pixels
	pixels := BytesToPixelsGrayscale(data)
	opts := PixelToImageOpts{
		Width: 2, Height: 1,
		CanvasWidth:  2,
		CanvasHeight: 1,
	}

	img := pixels.ToImage(opts)
	c1 := color.Gray16Model.Convert(img.At(0, 0)).(color.Gray16)
	if c1.Y == 0 {
		t.Errorf("expected nonzero grayscale value, got %v", c1)
	}
}

func TestIndexed_ToImage(t *testing.T) {
	data := []byte{0, 1, 2, 3}
	p := Indexed(data)

	opts := PixelToImageOpts{
		Width: 2, Height: 2,
		CanvasWidth:  2,
		CanvasHeight: 2,
		Palette: mockPalette{},
	}

	img := p.ToImage(opts)
	if img.At(0, 0).(color.RGBA).R != 0 {
		t.Errorf("expected first pixel to match palette index 0")
	}
	if img.At(1, 1).(color.RGBA).R != 3 {
		t.Errorf("expected pixel (1,1) to match palette index 3")
	}
}

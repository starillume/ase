package chunk

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/starillume/ase/common"
)

func createCelExtraData() []byte {
	buf := new(bytes.Buffer)

	extra := CelExtra{
		Flags:  0xDEADBEEF,
		X:      common.FloatToFixed(1.5),
		Y:      common.FloatToFixed(-2.25),
		Width:  common.FloatToFixed(10.0),
		Height: common.FloatToFixed(20.5),
	}

	binary.Write(buf, binary.LittleEndian, extra)

	return buf.Bytes()
}

func TestParseChunkCelExtra(t *testing.T) {
	data := createCelExtraData()
	chunk, err := ParseChunkCelExtra(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	celExtra, ok := chunk.(*CelExtra)
	if !ok {
		t.Fatalf("expected *CelExtra, got %T", chunk)
	}

	if celExtra.Flags != 0xDEADBEEF {
		t.Errorf("Flags: got 0x%X, want 0xDEADBEEF", celExtra.Flags)
	}

	if x := celExtra.X.FixedToFloat(); x != 1.5 {
		t.Errorf("X: got %f, want 1.5", x)
	}
	if y := celExtra.Y.FixedToFloat(); y != -2.25 {
		t.Errorf("Y: got %f, want -2.25", y)
	}
	if w := celExtra.Width.FixedToFloat(); w != 10.0 {
		t.Errorf("Width: got %f, want 10.0", w)
	}
	if h := celExtra.Height.FixedToFloat(); h != 20.5 {
		t.Errorf("Height: got %f, want 20.5", h)
	}
}


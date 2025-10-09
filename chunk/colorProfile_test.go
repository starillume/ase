package chunk

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/starillume/ase/common"
)

func createColorProfileData(profileType ColorProfileType, gamma float64) []byte {
	buf := new(bytes.Buffer)

	cData := ChunkColorProfileData{
		Type:  profileType,
		Flags: 0x1234,
		Gamma: common.FloatToFixed(gamma),
	}
	binary.Write(buf, binary.LittleEndian, cData)

	if profileType == ColorProfileICC {
		icc := []byte{1, 2, 3, 4, 5}
		size := uint32(len(icc))
		binary.Write(buf, binary.LittleEndian, size)
		binary.Write(buf, binary.LittleEndian, icc)
	}

	return buf.Bytes()
}

func TestParseChunkColorProfile_NoneOrSRGB(t *testing.T) {
	data := createColorProfileData(ColorProfileSRGB, 2.2)
	chunk, err := ParseChunkColorProfile(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cProfile, ok := chunk.(*ChunkColorProfile)
	if !ok {
		t.Fatalf("expected *ChunkColorProfile, got %T", chunk)
	}

	if cProfile.Type != ColorProfileSRGB {
		t.Errorf("Type: got %v, want %v", cProfile.Type, ColorProfileSRGB)
	}
	if cProfile.Flags != 0x1234 {
		t.Errorf("Flags: got 0x%X, want 0x1234", cProfile.Flags)
	}
	if gamma := cProfile.Gamma.FixedToFloat(); gamma <= 2.1 || gamma >= 2.2 {
		t.Errorf("Gamma: got %f, want 2.2", gamma)
	}
}

func TestParseChunkColorProfile_ICC(t *testing.T) {
	data := createColorProfileData(ColorProfileICC, 1.8)
	chunk, err := ParseChunkColorProfile(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cProfileICC, ok := chunk.(*ChunkColorProfileICC)
	if !ok {
		t.Fatalf("expected *ChunkColorProfileICC, got %T", chunk)
	}

	if cProfileICC.Type != ColorProfileICC {
		t.Errorf("Type: got %v, want %v", cProfileICC.Type, ColorProfileICC)
	}
	if cProfileICC.Flags != 0x1234 {
		t.Errorf("Flags: got 0x%X, want 0x1234", cProfileICC.Flags)
	}
	if gamma := cProfileICC.Gamma.FixedToFloat(); gamma <= 1.7 || gamma >= 1.8 {
		t.Errorf("Gamma: got %f, want 1.8", gamma)
	}
	if cProfileICC.DataLength != 5 {
		t.Errorf("DataLength: got %d, want 5", cProfileICC.DataLength)
	}
	expectedData := []byte{1, 2, 3, 4, 5}
	for i := range expectedData {
		if cProfileICC.Data[i] != expectedData[i] {
			t.Errorf("Data[%d]: got %d, want %d", i, cProfileICC.Data[i], expectedData[i])
		}
	}
}

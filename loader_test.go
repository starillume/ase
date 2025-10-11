package ase

import (
	"bytes"
	"encoding/binary"
	"testing"
)

type testStruct struct {
	A uint32
	B uint16
	C uint8
}

func createTestData() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint32(123456))
	binary.Write(buf, binary.LittleEndian, uint16(789))
	binary.Write(buf, binary.LittleEndian, uint8(42))
	return buf.Bytes()
}

func TestLoaderBytesToStructV2(t *testing.T) {
	data := createTestData()
	reader := bytes.NewReader(data)
	loader := &Loader{
		Reader: reader,
		Buf:    make([]byte, 4), // buffer menor que struct intencional
		Buffer: new(bytes.Buffer),
	}

	var s testStruct
	err := loader.BytesToStructV2(7, &s)
	if err != nil {
		t.Fatalf("BytesToStructV2 failed: %v", err)
	}

	if s.A != 123456 {
		t.Errorf("Expected A=123456, got %d", s.A)
	}
	if s.B != 789 {
		t.Errorf("Expected B=789, got %d", s.B)
	}
	if s.C != 42 {
		t.Errorf("Expected C=42, got %d", s.C)
	}
}

func TestLoaderReadToBuffer(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	reader := bytes.NewReader(data)
	loader := &Loader{
		Reader: reader,
		Buf:    make([]byte, 2),
		Buffer: new(bytes.Buffer),
	}

	// First read
	if err := loader.readToBuffer(); err != nil {
		t.Fatalf("readToBuffer failed: %v", err)
	}
	if loader.Buffer.Len() != 2 {
		t.Errorf("Expected buffer length 2, got %d", loader.Buffer.Len())
	}

	// Second read
	if err := loader.readToBuffer(); err != nil {
		t.Fatalf("readToBuffer failed: %v", err)
	}
	if loader.Buffer.Len() != 4 {
		t.Errorf("Expected buffer length 4, got %d", loader.Buffer.Len())
	}

	// Third read
	if err := loader.readToBuffer(); err != nil {
		t.Fatalf("readToBuffer failed: %v", err)
	}
	if loader.Buffer.Len() != 5 {
		t.Errorf("Expected buffer length 5, got %d", loader.Buffer.Len())
	}
}

func TestLoaderEnoughSpaceToRead(t *testing.T) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint32(1234))
	reader := bytes.NewReader(buf.Bytes())
	loader := &Loader{
		Reader: reader,
		Buf:    make([]byte, 4),
		Buffer: new(bytes.Buffer),
	}

	if loader.enoughSpaceToRead(1) {
		t.Errorf("Expected false, buffer is empty")
	}

	loader.readToBuffer()
	if !loader.enoughSpaceToRead(1) {
		t.Errorf("Expected true, buffer has data")
	}
	if !loader.enoughSpaceToRead(4) {
		t.Errorf("Expected true, buffer has 4 bytes")
	}
	if loader.enoughSpaceToRead(5) {
		t.Errorf("Expected false, buffer has only 4 bytes")
	}
}

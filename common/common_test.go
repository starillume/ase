package common

import (
	"bytes"
	"encoding/binary"
	"testing"
)

type sample struct {
	A uint16
	B int32
}

func TestBytesToStruct(t *testing.T) {
	buf := new(bytes.Buffer)
	expected := sample{A: 0x1234, B: 42}
	if err := binary.Write(buf, binary.LittleEndian, expected); err != nil {
		t.Fatalf("failed to write binary data: %v", err)
	}

	reader := bytes.NewReader(buf.Bytes())
	got, err := BytesToStruct[sample](reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != expected {
		t.Errorf("expected %+v, got %+v", expected, got)
	}
}

func TestBytesToStruct2(t *testing.T) {
	buf := new(bytes.Buffer)
	expected := sample{A: 0xABCD, B: -123}
	if err := binary.Write(buf, binary.LittleEndian, expected); err != nil {
		t.Fatalf("failed to write binary data: %v", err)
	}

	reader := bytes.NewReader(buf.Bytes())
	var got sample
	if err := BytesToStruct2(reader, &got); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != expected {
		t.Errorf("expected %+v, got %+v", expected, got)
	}
}

func TestCheckMagicNumber_Valid(t *testing.T) {
	err := CheckMagicNumber(0xAABB, 0xAABB, "testcase")
	if err != nil {
		t.Errorf("unexpected error for valid magic number: %v", err)
	}
}

func TestCheckMagicNumber_Invalid(t *testing.T) {
	err := CheckMagicNumber(0xAABB, 0xCCDD, "testcase")
	if err == nil {
		t.Fatal("expected error for invalid magic number, got nil")
	}

	want := "testcase: magic number fail"
	if err.Error()[:len(want)] != want {
		t.Errorf("unexpected error message: %v", err)
	}
}

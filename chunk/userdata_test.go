package chunk_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/starillume/ase/chunk"
)

func createChunkUserDataWithVectors() []byte {
	buf := new(bytes.Buffer)

	var flag chunk.UserDataFlag = chunk.UserDataHasText | chunk.UserDataHasColor | chunk.UserDataHasProperties
	binary.Write(buf, binary.LittleEndian, flag)

	text := "HelloVector"
	textLen := uint16(len(text))
	binary.Write(buf, binary.LittleEndian, textLen)
	buf.Write([]byte(text))

	color := chunk.ChunkUserDataColor{R: 10, G: 20, B: 30, A: 255}
	binary.Write(buf, binary.LittleEndian, color)

	propsBuf := new(bytes.Buffer)

	propMapData := chunk.ChunkUserDataPropMapData{
		PropKey:     0,
		PropNumbers: 3,
	}
	binary.Write(propsBuf, binary.LittleEndian, propMapData)

	propName := "testProp"
	nameLen := uint16(len(propName))
	binary.Write(propsBuf, binary.LittleEndian, nameLen)
	propsBuf.Write([]byte(propName))
	typeValue := chunk.UserDataInt32
	binary.Write(propsBuf, binary.LittleEndian, typeValue)
	binary.Write(propsBuf, binary.LittleEndian, int32(123))

	propName = "homVec"
	nameLen = uint16(len(propName))
	binary.Write(propsBuf, binary.LittleEndian, nameLen)
	propsBuf.Write([]byte(propName))
	typeValue = chunk.UserDataVector
	binary.Write(propsBuf, binary.LittleEndian, typeValue)
	count := uint32(3)
	elemType := uint16(chunk.UserDataInt32)
	binary.Write(propsBuf, binary.LittleEndian, count)
	binary.Write(propsBuf, binary.LittleEndian, elemType)
	binary.Write(propsBuf, binary.LittleEndian, int32(1))
	binary.Write(propsBuf, binary.LittleEndian, int32(2))
	binary.Write(propsBuf, binary.LittleEndian, int32(3))

	propName = "hetVec"
	nameLen = uint16(len(propName))
	binary.Write(propsBuf, binary.LittleEndian, nameLen)
	propsBuf.Write([]byte(propName))
	typeValue = chunk.UserDataVector
	binary.Write(propsBuf, binary.LittleEndian, typeValue)
	count = uint32(2)
	elemType = 0
	binary.Write(propsBuf, binary.LittleEndian, count)
	binary.Write(propsBuf, binary.LittleEndian, elemType)
	elemType0 := uint16(chunk.UserDataInt32)
	elemType1 := uint16(chunk.UserDataFloat)
	binary.Write(propsBuf, binary.LittleEndian, elemType0)
	binary.Write(propsBuf, binary.LittleEndian, int32(5))
	binary.Write(propsBuf, binary.LittleEndian, elemType1)
	binary.Write(propsBuf, binary.LittleEndian, float32(1.5))

	propMapHeader := chunk.ChunkUserDataPropMapHeader{
		SizeInBytes:    uint32(propsBuf.Len()),
		PropMapNumbers: 1,
	}
	binary.Write(buf, binary.LittleEndian, propMapHeader)
	binary.Write(buf, binary.LittleEndian, propsBuf.Bytes())

	return buf.Bytes()
}

func TestParseChunkUserDataWithVectors(t *testing.T) {
	data := createChunkUserDataWithVectors()
	parsed, err := chunk.ParseChunkUserData(data)
	if err != nil {
		t.Fatalf("ParseChunkUserData failed: %v", err)
	}

	ud := parsed.(*chunk.UserData)

	if ud.Text != "HelloVector" {
		t.Errorf("Text mismatch: expected 'HelloVector', got '%s'", ud.Text)
	}

	if ud.Color == nil {
		t.Fatal("Color is nil")
	}
	if ud.Color.R != 10 || ud.Color.G != 20 || ud.Color.B != 30 || ud.Color.A != 255 {
		t.Errorf("Color mismatch: %+v", ud.Color)
	}

	if ud.Maps == nil || len(*ud.Maps) != 1 {
		t.Fatalf("Expected 1 property map, got %+v", ud.Maps)
	}
	propMap := (*ud.Maps)[0]

	val, ok := propMap.Props["testProp"]
	if !ok {
		t.Fatal("Property 'testProp' not found")
	}
	if intVal, ok := val.(int32); !ok || intVal != 123 {
		t.Errorf("testProp mismatch: expected 123, got %v", val)
	}

	val, ok = propMap.Props["homVec"]
	if !ok {
		t.Fatal("Property 'homVec' not found")
	}
	elems, ok := val.([]any)
	if !ok || len(elems) != 3 {
		t.Fatalf("homVec expected 3 elements, got %v", val)
	}
	expected := []int32{1, 2, 3}
	for i := 0; i < 3; i++ {
		if elems[i].(int32) != expected[i] {
			t.Errorf("homVec element %d mismatch: expected %d, got %v", i, expected[i], elems[i])
		}
	}

	val, ok = propMap.Props["hetVec"]
	if !ok {
		t.Fatal("Property 'hetVec' not found")
	}
	elems, ok = val.([]any)
	if !ok || len(elems) != 2 {
		t.Fatalf("hetVec expected 2 elements, got %v", val)
	}
	if elems[0].(int32) != 5 {
		t.Errorf("hetVec element 0 mismatch: expected 5, got %v", elems[0])
	}
	if elems[1].(float32) != 1.5 {
		t.Errorf("hetVec element 1 mismatch: expected 1.5, got %v", elems[1])
	}
}

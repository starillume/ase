package chunk_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/starillume/ase/chunk"
)

// createChunkUserDataWithVectors generates a byte slice with:
// - text: "HelloVector"
// - color: {10,20,30,255}
// - properties: 1 map with:
//   - "testProp": int32=123
//   - "homVec": vector of int32 [1,2,3]
//   - "hetVec": vector hetero [int32=5, float32=1.5]
func createChunkUserDataWithVectors() []byte {
	buf := new(bytes.Buffer)

	// Flags
	var flag chunk.UserDataFlag = chunk.UserDataHasText | chunk.UserDataHasColor | chunk.UserDataHasProperties
	binary.Write(buf, binary.LittleEndian, flag)

	// Text
	text := "HelloVector"
	textLen := uint16(len(text))
	binary.Write(buf, binary.LittleEndian, textLen)
	buf.Write([]byte(text))

	// Color
	color := chunk.ChunkUserDataColor{R: 10, G: 20, B: 30, A: 255}
	binary.Write(buf, binary.LittleEndian, color)

	// Property map header: 1 map
	propMapHeader := chunk.ChunkUserDataPropMapHeader{
		SizeInBytes:    0,
		PropMapNumbers: 3, // testProp + homVec + hetVec
	}
	binary.Write(buf, binary.LittleEndian, propMapHeader)

	// Property map data
	propMapData := chunk.ChunkUserDataPropMapData{
		PropKey:     0,
		PropNumbers: 3,
	}
	binary.Write(buf, binary.LittleEndian, propMapData)

	// Property 1: "testProp" int32=123
	propName := "testProp"
	nameLen := uint16(len(propName))
	binary.Write(buf, binary.LittleEndian, nameLen)
	buf.Write([]byte(propName))
	typeValue := chunk.UserDataInt32
	binary.Write(buf, binary.LittleEndian, typeValue)
	binary.Write(buf, binary.LittleEndian, int32(123))

	// Property 2: "homVec" vector int32 [1,2,3]
	propName = "homVec"
	nameLen = uint16(len(propName))
	binary.Write(buf, binary.LittleEndian, nameLen)
	buf.Write([]byte(propName))
	typeValue = chunk.UserDataVector
	binary.Write(buf, binary.LittleEndian, typeValue)
	// Vector header
	count := uint32(3)
	elemType := uint16(chunk.UserDataInt32) // homogeneous
	binary.Write(buf, binary.LittleEndian, count)
	binary.Write(buf, binary.LittleEndian, elemType)
	// Elements
	binary.Write(buf, binary.LittleEndian, int32(1))
	binary.Write(buf, binary.LittleEndian, int32(2))
	binary.Write(buf, binary.LittleEndian, int32(3))

	// Property 3: "hetVec" vector hetero [int32=5, float32=1.5]
	propName = "hetVec"
	nameLen = uint16(len(propName))
	binary.Write(buf, binary.LittleEndian, nameLen)
	buf.Write([]byte(propName))
	typeValue = chunk.UserDataVector
	binary.Write(buf, binary.LittleEndian, typeValue)
	// Vector header
	count = uint32(2)
	elemType = 0 // heterogeneous
	binary.Write(buf, binary.LittleEndian, count)
	binary.Write(buf, binary.LittleEndian, elemType)
	// Elements: first int32, then float32
	elemType0 := uint16(chunk.UserDataInt32)
	elemType1 := uint16(chunk.UserDataFloat)
	binary.Write(buf, binary.LittleEndian, elemType0)
	binary.Write(buf, binary.LittleEndian, int32(5))
	binary.Write(buf, binary.LittleEndian, elemType1)
	binary.Write(buf, binary.LittleEndian, float32(1.5))

	return buf.Bytes()
}

func TestParseChunkUserDataWithVectors(t *testing.T) {
	data := createChunkUserDataWithVectors()
	parsed, err := chunk.ParseChunkUserData(data)
	if err != nil {
		t.Fatalf("ParseChunkUserData failed: %v", err)
	}

	ud := parsed.(*chunk.UserData)

	// Text
	if ud.Text != "HelloVector" {
		t.Errorf("Text mismatch: expected 'HelloVector', got '%s'", ud.Text)
	}

	// Color
	if ud.Color == nil {
		t.Fatal("Color is nil")
	}
	if ud.Color.R != 10 || ud.Color.G != 20 || ud.Color.B != 30 || ud.Color.A != 255 {
		t.Errorf("Color mismatch: %+v", ud.Color)
	}

	// Property map
	if ud.Maps == nil || len(*ud.Maps) != 1 {
		t.Fatalf("Expected 1 property map, got %v", ud.Maps)
	}
	propMap := (*ud.Maps)[0]

	// testProp
	val, ok := propMap.Props["testProp"]
	if !ok {
		t.Fatal("Property 'testProp' not found")
	}
	if intVal, ok := val.(int32); !ok || intVal != 123 {
		t.Errorf("testProp mismatch: expected 123, got %v", val)
	}

	// homVec
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

	// hetVec
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

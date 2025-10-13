package chunk

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/starillume/ase/common"
)

type readerFunc func([]byte) (any, int, error)

var (
	readers        map[UserDataPropType]readerFunc
	initReaderSync sync.Once
)

const UserDataFlagSize = 4

type UserDataFlag uint32

const (
	UserDataHasText       UserDataFlag = 1 << 0
	UserDataHasColor      UserDataFlag = 1 << 1
	UserDataHasProperties UserDataFlag = 1 << 2
)

type UserData struct {
	Text  string
	Color *ChunkUserDataColor
	Maps  *[]ChunkUserDataPropMap
}

type ChunkUserDataPropMap struct {
	ExternalId int
	Props      map[string]any
}

type ChunkUserDataTextSize uint32

const ChunkUserDataColorSize = 4

type ChunkUserDataColor struct {
	R byte
	G byte
	B byte
	A byte
}

const ChunkUserDataPropMapHeaderSize = 8

type ChunkUserDataPropMapHeader struct {
	SizeInBytes    uint32
	PropMapNumbers uint32
}

const ChunkUserDataPropMapDataSize = 8

type ChunkUserDataPropMapData struct {
	PropKey     uint32
	PropNumbers uint32
}

type UserDataPropType uint16

const (
	UserDataNil UserDataPropType = iota
	UserDataBool
	UserDataInt8
	UserDataUint8
	UserDataInt16
	UserDataUint16
	UserDataInt32
	UserDataUint32
	UserDataInt64
	UserDataUint64
	UserDataFixed
	UserDataFloat
	UserDataDouble
	UserDataString
	UserDataPoint
	UserDataSize
	UserDataRect
	UserDataVector
	UserDataProp
	UserDataUUID
)

func getReaders() map[UserDataPropType]readerFunc {
	initReaderSync.Do(initReaders)
	return readers
}

func initReaders() {
	readers = map[UserDataPropType]readerFunc{
		UserDataBool: func(data []byte) (any, int, error) {
			var b uint8 // BYTE
			reader := bytes.NewReader(data)
			if err := common.BytesToStruct2(reader, &b); err != nil {
				return nil, 0, err
			}
			return b != 0, 1, nil
		},

		UserDataInt8: func(data []byte) (any, int, error) {
			var v int8 // BYTE interpretado como signed
			reader := bytes.NewReader(data)
			return v, 1, common.BytesToStruct2(reader, &v)
		},

		UserDataUint8: func(data []byte) (any, int, error) {
			var v uint8
			reader := bytes.NewReader(data)
			return v, 1, common.BytesToStruct2(reader, &v)
		},

		UserDataInt16: func(data []byte) (any, int, error) {
			var v int16 // SHORT
			reader := bytes.NewReader(data)
			return v, 2, common.BytesToStruct2(reader, &v)
		},

		UserDataUint16: func(data []byte) (any, int, error) {
			var v uint16 // WORD
			reader := bytes.NewReader(data)
			return v, 2, common.BytesToStruct2(reader, &v)
		},

		UserDataInt32: func(data []byte) (any, int, error) {
			var v int32 // LONG
			reader := bytes.NewReader(data)
			return v, 4, common.BytesToStruct2(reader, &v)
		},

		UserDataUint32: func(data []byte) (any, int, error) {
			var v uint32 // DWORD
			reader := bytes.NewReader(data)
			return v, 4, common.BytesToStruct2(reader, &v)
		},

		UserDataInt64: func(data []byte) (any, int, error) {
			var v int64 // LONG64
			reader := bytes.NewReader(data)
			return v, 8, common.BytesToStruct2(reader, &v)
		},

		UserDataUint64: func(data []byte) (any, int, error) {
			var v uint64 // QWORD
			reader := bytes.NewReader(data)
			return v, 8, common.BytesToStruct2(reader, &v)
		},

		UserDataFixed: func(data []byte) (any, int, error) {
			var v common.Fixed // FIXED (16.16)
			reader := bytes.NewReader(data)
			return v, 4, common.BytesToStruct2(reader, &v)
		},

		UserDataFloat: func(data []byte) (any, int, error) {
			var v float32 // FLOAT
			reader := bytes.NewReader(data)
			return v, 4, common.BytesToStruct2(reader, &v)
		},

		UserDataDouble: func(data []byte) (any, int, error) {
			var v float64 // DOUBLE
			reader := bytes.NewReader(data)
			return v, 8, common.BytesToStruct2(reader, &v)
		},

		UserDataString: func(data []byte) (any, int, error) {
			var length uint16 // WORD
			reader := bytes.NewReader(data)
			if err := common.BytesToStruct2(reader, &length); err != nil {
				return nil, 0, err
			}
			stringData := make([]byte, length)
			if err := common.BytesToStruct2(reader, &stringData); err != nil {
				return nil, 0, err
			}

			readed := int(2 + length)
			return string(data), readed, nil // UTF-8, sem \0
		},

		UserDataPoint: func(data []byte) (any, int, error) {
			var p struct {
				X int32 // LONG
				Y int32 // LONG
			}
			reader := bytes.NewReader(data)
			return p, 8, common.BytesToStruct2(reader, &p)
		},

		UserDataSize: func(data []byte) (any, int, error) {
			var s struct {
				W int32 // LONG
				H int32 // LONG
			}
			reader := bytes.NewReader(data)
			return s, 8, common.BytesToStruct2(reader, &s)
		},

		UserDataRect: func(data []byte) (any, int, error) {
			var r struct {
				X int32
				Y int32
				W int32
				H int32
			}
			reader := bytes.NewReader(data)
			return r, 16, common.BytesToStruct2(reader, &r)
		},

		UserDataVector: func(data []byte) (any, int, error) {
			var count uint32 // DWORD
			reader := bytes.NewReader(data)
			if err := common.BytesToStruct2(reader, &count); err != nil {
				return nil, 0, err
			}

			var elemType uint16 // WORD
			if err := common.BytesToStruct2(reader, &elemType); err != nil {
				return nil, 0, err
			}

			readed := 4 + 2
			var elements []any

			if elemType == 0 {
				// Heterogêneo: cada elemento tem tipo próprio
				for i := 0; i < int(count); i++ {
					var etype uint16
					if err := common.BytesToStruct2(reader, &etype); err != nil {
						return nil, 0, err
					}
					readerUserData, ok := getReaders()[UserDataPropType(etype)]
					if !ok {
						return nil, 0, fmt.Errorf("unsupported vector element type %d", etype)
					}
					unread := data[len(data)-reader.Len():]
					val, n, err := readerUserData(unread)
					if err != nil {
						return nil, 0, err
					}
					elements = append(elements, val)
					readed += n + 2
					reader.Seek(int64(n), io.SeekCurrent)
				}
			} else {
				// Homogêneo: todos elementos têm o mesmo tipo
				readerUserData, ok := getReaders()[UserDataPropType(elemType)]
				if !ok {
					return nil, 0, fmt.Errorf("unsupported vector element type %d", elemType)
				}
				for i := 0; i < int(count); i++ {
					unread := data[len(data)-reader.Len():]
					val, n, err := readerUserData(unread)
					if err != nil {
						return nil, 0, err
					}
					elements = append(elements, val)
					readed += n
					reader.Seek(int64(n), io.SeekCurrent)
				}
			}

			return elements, readed, nil
		},

		UserDataUUID: func(data []byte) (any, int, error) {
			var uuid [16]byte
			reader := bytes.NewReader(data)
			return uuid, 16, common.BytesToStruct2(reader, &uuid)
		},

		UserDataProp: func(data []byte) (any, int, error) {
			var propsLen uint32
			reader := bytes.NewReader(data)
			if err := common.BytesToStruct2(reader, &propsLen); err != nil {
				return nil, 0, err
			}

			unread := data[len(data)-reader.Len():]
			p, n, err := ParseUserDataProps(unread, int(propsLen))
			if err != nil {
				return nil, 0, err
			}

			m := &ChunkUserDataPropMap{
				ExternalId: 0,
				Props:      p,
			}

			readed := 4 + n

			return m, readed, nil
		},
	}
}

func ParseChunkUserData(data []byte) (Chunk, error) {
	reader := bytes.NewReader(data)

	var flag UserDataFlag
	if err := common.BytesToStruct2(reader, &flag); err != nil {
		return nil, err
	}

	chunk := &UserData{}

	if flag&UserDataHasText == UserDataHasText {
		var textLen uint16
		if err := common.BytesToStruct2(reader, &textLen); err != nil {
			return nil, err
		}

		text := make([]byte, textLen)
		if err := common.BytesToStruct2(reader, &text); err != nil {
			return nil, err
		}

		chunk.Text = string(text)
	}

	if flag&UserDataHasColor == UserDataHasColor {
		var userDataColor ChunkUserDataColor
		if err := common.BytesToStruct2(reader, &userDataColor); err != nil {
			return nil, err
		}

		chunk.Color = &userDataColor
	}

	if flag&UserDataHasProperties == UserDataHasProperties {
		var mapHeader ChunkUserDataPropMapHeader
		if err := common.BytesToStruct2(reader, &mapHeader); err != nil {
			return nil, err
		}

		propMaps := make([]ChunkUserDataPropMap, mapHeader.PropMapNumbers)

		propsSize := int(mapHeader.SizeInBytes)

		for i := range mapHeader.PropMapNumbers {
			if propsSize <= 0 {
				return nil, fmt.Errorf("malformed file, SizeInBytes is smaller than the total size")
			}

			unread := data[len(data)-reader.Len():]
			pm, n, err := ParseUserDataPropMap(unread)
			if err != nil {
				return nil, err
			}

			propsSize -= n

			propMaps[i] = *pm
			reader.Seek(int64(n), io.SeekCurrent)
		}

		chunk.Maps = &propMaps
	}

	return chunk, nil
}

func ParseUserDataPropMap(data []byte) (*ChunkUserDataPropMap, int, error) {
	reader := bytes.NewReader(data)
	readed := 0

	var propMapData ChunkUserDataPropMapData
	if err := common.BytesToStruct2(reader, &propMapData); err != nil {
		return nil, 0, err
	}

	readed += 8

	propMap := &ChunkUserDataPropMap{
		ExternalId: int(propMapData.PropKey),
	}

	unread := data[len(data)-reader.Len():]
	propsmap, n, err := ParseUserDataProps(unread, int(propMapData.PropNumbers))
	if err != nil {
		return nil, 0, err
	}

	readed += n

	propMap.Props = propsmap

	return propMap, readed, nil
}

func ParseUserDataProps(data []byte, count int) (map[string]any, int, error) {
	reader := bytes.NewReader(data)

	readed := 0

	props := make(map[string]any, count)
	for range count {
		var nameLen uint16
		if err := common.BytesToStruct2(reader, &nameLen); err != nil {
			return nil, 0, err
		}

		readed += 2

		nameBytes := make([]byte, nameLen)
		if err := common.BytesToStruct2(reader, &nameBytes); err != nil {
			return nil, 0, err
		}

		readed += int(nameLen)

		key := string(nameBytes)

		var typeValue UserDataPropType
		if err := common.BytesToStruct2(reader, &typeValue); err != nil {
			return nil, 0, err
		}

		readed += 2

		if readerUserData, ok := getReaders()[typeValue]; ok {
			unread := data[len(data)-reader.Len():]
			value, n, err := readerUserData(unread)
			if err != nil {
				return nil, 0, err
			}
			props[key] = value
			readed += n
			reader.Seek(int64(n), io.SeekCurrent)
		} else {
			return nil, 0, fmt.Errorf("unsupported type %d", typeValue)
		}
	}

	return props, readed, nil
}

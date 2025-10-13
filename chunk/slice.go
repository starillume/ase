package chunk

import (
	"bytes"

	"github.com/starillume/ase/common"
)

type Slice struct {
	ChunkSliceData
	Name string
	Keys []ChunkSliceKey
}

const ChunkSliceDataSize = 14

type ChunkSliceData struct {
	NumberSliceKeys uint32
	FlagsBit        uint32
	_               uint32
	NameLength      uint16
}

type ChunkSliceKey struct {
	ChunkSliceKeyData
	*ChunkSliceKey9PatchesData
	*ChunkSliceKeyPivotData
}

const ChunkSliceKeyDataSize = 20

type ChunkSliceKeyData struct {
	FrameNumber uint32
	OriginX     int32
	OriginY     int32
	Width       uint32 // (can be 0 if this slice hidden in the animation from the given frame)
	Height      uint32
}

const ChunkSliceKey9PatchesDataSize = 16

type ChunkSliceKey9PatchesData struct {
	CenterX      int32
	CenterY      int32
	CenterWidth  uint32
	CenterHeight uint32
}

const ChunkSliceKeyPivotDataSize = 8

type ChunkSliceKeyPivotData struct {
	X int32
	Y int32
}

func ParseChunkSlice(data []byte) (Chunk, error) {
	reader := bytes.NewReader(data)

	var chunkSliceData ChunkSliceData
	if err := common.BytesToStruct2(reader, &chunkSliceData); err != nil {
		return nil, err
	}

	chunkSliceName := make([]byte, chunkSliceData.NameLength)
	if err := common.BytesToStruct2(reader, &chunkSliceName); err != nil {
		return nil, err
	}

	sliceKeys := make([]ChunkSliceKey, chunkSliceData.NumberSliceKeys)
	for i := 0; i < int(chunkSliceData.NumberSliceKeys); i++ {
		var sliceKeyData ChunkSliceKeyData
		if err := common.BytesToStruct2(reader, &sliceKeyData); err != nil {
			return nil, err
		}

		sliceKey := ChunkSliceKey{ChunkSliceKeyData: sliceKeyData, ChunkSliceKeyPivotData: nil, ChunkSliceKey9PatchesData: nil}

		if chunkSliceData.FlagsBit&1 == 1 {
			var ninePatchesData ChunkSliceKey9PatchesData
			if err := common.BytesToStruct2(reader, &ninePatchesData); err != nil {
				return nil, err
			}

			sliceKey.ChunkSliceKey9PatchesData = &ninePatchesData
		}

		if chunkSliceData.FlagsBit&2 == 2 {
			var pivotData ChunkSliceKeyPivotData
			if err := common.BytesToStruct2(reader, &pivotData); err != nil {
				return nil, err
			}

			sliceKey.ChunkSliceKeyPivotData = &pivotData
		}

		sliceKeys[i] = sliceKey
	}

	return &Slice{
		ChunkSliceData: chunkSliceData,
		Name:           string(chunkSliceName),
		Keys:           sliceKeys,
	}, nil
}

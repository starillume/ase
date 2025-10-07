package chunk

import (
	"bytes"

	"github.com/starillume/ase/common"
)

type ChunkCelExtra struct {
	ChunkCelExtraData
}

const ChunkCelExtraDataSize = 36

type ChunkCelExtraData struct {
	Flags  uint32
	X      common.Fixed
	Y      common.Fixed
	Width  common.Fixed
	Height common.Fixed
	_      [16]byte
}

func ParseChunkCelExtra(data []byte) (Chunk, error) {
	reader := bytes.NewReader(data)

	var celExtraData ChunkCelExtraData
	if err := common.BytesToStruct2(reader, &celExtraData); err != nil {
		return nil, err
	}

	return &ChunkCelExtra{
		ChunkCelExtraData: celExtraData,
	}, nil
}

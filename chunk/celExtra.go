package chunk

import (
	"bytes"

	"github.com/starillume/ase/common"
)

type CelExtra struct {
	Flags  uint32
	X      common.Fixed
	Y      common.Fixed
	Width  common.Fixed
	Height common.Fixed
	_      [16]byte
}

const ChunkCelExtraDataSize = 36

func ParseChunkCelExtra(data []byte) (Chunk, error) {
	reader := bytes.NewReader(data)

	var celExtraData CelExtra
	if err := common.BytesToStruct2(reader, &celExtraData); err != nil {
		return nil, err
	}

	return &celExtraData, nil
}

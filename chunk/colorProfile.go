package chunk

import (
	"bytes"

	"github.com/starillume/ase/common"
)

type ColorProfileType uint16

const (
	ColorProfileNone ColorProfileType = iota
	ColorProfileSRGB
	ColorProfileICC
)

type ChunkColorProfile struct {
	ChunkColorProfileData
}

type ChunkColorProfileICC struct {
	ChunkColorProfile
	ChunkColorProfileICCData
}

const ChunkColorProfileDataSize = 16

type ChunkColorProfileData struct {
	Type  ColorProfileType
	Flags uint16
	Gamma common.Fixed
	_     [8]byte
}

type ChunkColorProfileICCData struct {
	DataLength uint32
	Data       []byte
}

func ParseChunkColorProfile(data []byte) (Chunk, error) {
	reader := bytes.NewReader(data)

	var cData ChunkColorProfileData
	err := common.BytesToStruct2(reader, &cData)
	if err != nil {
		return nil, err
	}

	if cData.Type == ColorProfileICC {
		var iccSize uint32
		err = common.BytesToStruct2(reader, &iccSize)
		if err != nil {
			return nil, err
		}
		iccData := make([]byte, iccSize)
		err = common.BytesToStruct2(reader, &iccData)
		if err != nil {
			return nil, err
		}

		chunk := &ChunkColorProfileICC{
			ChunkColorProfile: ChunkColorProfile{
				ChunkColorProfileData: cData,
			},
			ChunkColorProfileICCData: ChunkColorProfileICCData{
				DataLength: iccSize,
				Data:       iccData,
			},
		}
		return chunk, nil
	}
	chunk := &ChunkColorProfile{
		ChunkColorProfileData: cData,
	}
	return chunk, nil
}

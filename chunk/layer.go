package chunk

import (
	"bytes"
	"reflect"

	"github.com/starillume/ase/common"
)

type Layer struct {
	ChunkLayerData ChunkLayerData
	ChunkLayerName
	ChunkLayerFlags
	*ChunkLayerType2Data
	*ChunkLayerLockMovementData
}

const ChunkLayerDataSize = 18

type ChunkLayerData struct {
	FlagsBit      uint16
	Type          uint16
	ChildLevel    uint16
	DefaultWidth  uint16
	DefaultHeight uint16
	BlendMode     uint16
	Opacity       byte
	_             [3]byte
	NameLength    uint16
}

type ChunkLayerFlags struct {
	Visible                    bool
	Editable                   bool
	LockMovement               bool
	Background                 bool
	PreferLinkedCels           bool
	LayerGroupDisplayCollapsed bool
	ReferenceLayer             bool
}

type ChunkLayerName string

type ChunkLayerType2Data struct {
	TilesetIndex uint32
}

type ChunkLayerLockMovementData struct {
	LayerUUID [16]byte
}

func ParseChunkLayer(data []byte) (*Layer, error) {
	reader := bytes.NewReader(data)

	var layerData ChunkLayerData
	if err := common.BytesToStruct2(reader, &layerData); err != nil {
		return nil, err
	}

	nameBytes := make([]byte, layerData.NameLength)
	if err := common.BytesToStruct2(reader, &nameBytes); err != nil {
		return nil, err
	}
	var name ChunkLayerName = ChunkLayerName(string(nameBytes))
	chunk := &Layer{
		ChunkLayerData:             layerData,
		ChunkLayerName:             name,
		ChunkLayerType2Data:        nil,
		ChunkLayerLockMovementData: nil,
	}
	if layerData.Type == 2 {
		var type2Data ChunkLayerType2Data
		if err := common.BytesToStruct2(reader, &type2Data); err != nil {
			return nil, err
		}

		chunk.ChunkLayerType2Data = &type2Data
	}

	var flags ChunkLayerFlags = ChunkLayerFlags{}
	value := reflect.ValueOf(&flags).Elem()
	fieldIndex := 0
	for i := uint16(1); i < 65; i *= 2 {
		field := value.Field(fieldIndex)
		if layerData.FlagsBit&i == i && field.CanSet() {
			field.SetBool(true)
		}
		fieldIndex++
	}

	chunk.ChunkLayerFlags = flags

	if flags.LockMovement {
		var lockData ChunkLayerLockMovementData
		if err := common.BytesToStruct2(reader, &lockData); err != nil {
			return nil, err
		}

		chunk.ChunkLayerLockMovementData = &lockData
	}

	return chunk, nil
}

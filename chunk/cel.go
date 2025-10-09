package chunk

import (
	"bytes"
	"io"

	"github.com/starillume/ase/common"
	"github.com/starillume/ase/pixel"
)

type CelType uint16

type Cel interface {
	CelType() CelType
}

const (
	CelTypeRawImage CelType = iota
	CelTypeLinked
	CelTypeCompressedImage
	CelTypeCompressedTilemap
)

type ChunkCelImage struct {
	ChunkCelData
	ChunkCelRawImageData
}

func (c *ChunkCelImage) CelType() CelType {
	return CelTypeRawImage
}

type ChunkCelCompressedImage struct {
	ChunkCelData
	ChunkCelCompressedImageData
}

func (c *ChunkCelCompressedImage) CelType() CelType {
	return CelTypeCompressedImage
}

type ChunkCelLinked struct {
	ChunkCelData
	ChunkCelLinkedData
}

func (c *ChunkCelLinked) CelType() CelType {
	return CelTypeLinked
}

type ChunkCelTilemap struct {
	ChunkCelData
	ChunkCelCompressedTilemapData
}

func (c *ChunkCelTilemap) CelType() CelType {
	return CelTypeCompressedTilemap
}

const ChunkCelDataSize = 16

type ChunkCelData struct {
	LayerIndex uint16
	X          int16
	Y          int16
	Opacity    byte
	CelType    // not flag
	Z          int16
	_          [5]byte
}

const ChunkCelDimensionSize = 4

type ChunkCelDimensionData struct {
	Width  uint16
	Height uint16
}

type ChunkCelRawImageData struct {
	ChunkCelDimensionData
	Pixels []byte
}

const ChunkCelLinkedDataSize = 2

type ChunkCelLinkedData struct {
	FramePosition int16
}

type ChunkCelCompressedImageData struct {
	ChunkCelDimensionData
	Pixels pixel.PixelsCompressed
}

const ChunkCelCompressedTilemapStaticDataSize = 28

type ChunkCelCompressedTilemapStaticData struct {
	BitsPerTile      uint16
	MaskTileId       uint32
	MaskXFlip        uint32
	MaskYFlip        uint32
	MaskDiagonalFlip uint32
	_                [10]byte
}

type ChunkCelCompressedTilemapData struct {
	ChunkCelDimensionData
	ChunkCelCompressedTilemapStaticData
	Tiles pixel.PixelsCompressed
	// NOTE: não sei como fazer esse negocio, voltar depois
	// NOTE: dica pro lerdinho acima: tile não é pixel !
}

func ParseChunkCel(data []byte) (Chunk, error) {
	reader := bytes.NewReader(data)

	var cData ChunkCelData
	if err := common.BytesToStruct2(reader, &cData); err != nil {
		return nil, err
	}

	if cData.CelType == CelTypeLinked {
		var cLinkeData ChunkCelLinkedData
		if err := common.BytesToStruct2(reader, &cLinkeData); err != nil {
			return nil, err
		}
		return &ChunkCelLinked{
			ChunkCelData:       cData,
			ChunkCelLinkedData: cLinkeData,
		}, nil
	}

	var dimensions ChunkCelDimensionData
	if err := common.BytesToStruct2(reader, &dimensions); err != nil {
		return nil, err
	}

	// pixelDataSize := int(ch.Size - ChunkHeaderSize - ChunkCelDataSize - ChunkCelDimensionSize)
	switch cData.CelType {
	case CelTypeRawImage:
		pixels := new(bytes.Buffer)

		io.Copy(pixels, reader)

		return &ChunkCelImage{
			ChunkCelData: cData,
			ChunkCelRawImageData: ChunkCelRawImageData{
				ChunkCelDimensionData: dimensions,
				Pixels:                pixels.Bytes(),
			},
		}, nil
	case CelTypeCompressedImage:
		pixels := new(bytes.Buffer)

		io.Copy(pixels, reader)

		return &ChunkCelCompressedImage{
			ChunkCelData: cData,
			ChunkCelCompressedImageData: ChunkCelCompressedImageData{
				ChunkCelDimensionData: dimensions,
				Pixels:                pixel.PixelsZlib(pixels.Bytes()),
			},
		}, nil
		// 	case CelTypeCompressedTilemap:
		// 		// TODO: ver dps dado errado @Carto1a
		// 		var ctilemapStatic ChunkCelCompressedTilemapStaticData
		// 		if err := common.BytesToStruct2(reader, &ctilemapStatic); err != nil {
		// 			return nil, err
		// 		}
		//
		// 		cTilemapData := ChunkCelCompressedTilemapData{
		// 			ChunkCelDimensionData:               dimensions,
		// 			ChunkCelCompressedTilemapStaticData: ctilemapStatic,
		// 		}
		// 		return &ChunkCelTilemap{
		// 			ChunkCelData:                  cData,
		// 			ChunkCelCompressedTilemapData: cTilemapData,
		// 		}, nil
	}

	return nil, nil
}

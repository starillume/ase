package chunk

import "fmt"

type ChunkDataType uint16

const (
	OldPaletteChunkHex    ChunkDataType = 0x0004
	OldPaletteChunk2Hex   ChunkDataType = 0x0011
	LayerChunkHex         ChunkDataType = 0x2004
	CelChunkHex           ChunkDataType = 0x2005
	CelExtraChunkHex      ChunkDataType = 0x2006
	ColorProfileChunkHex  ChunkDataType = 0x2007
	ExternalFilesChunkHex ChunkDataType = 0x2008
	MaskChunkHex          ChunkDataType = 0x2016 // DEPRECATED
	PathChunkHex          ChunkDataType = 0x2017 // NEVER USED
	TagsChunkHex          ChunkDataType = 0x2018
	PaletteChunkHex       ChunkDataType = 0x2019
	UserDataChunkHex      ChunkDataType = 0x2020
	SliceChunkHex         ChunkDataType = 0x2022
	TilesetChunkHex       ChunkDataType = 0x2023
)

const HeaderSize = 6

type Header struct {
	Size uint32
	Type ChunkDataType
}

type Chunk interface {}

func Parse(ctype ChunkDataType, data []byte) (Chunk, error) {
	switch ctype {
	case OldPaletteChunkHex:
		return ParseChunkOldPalette(data, ctype)
	case OldPaletteChunk2Hex:
		return ParseChunkOldPalette(data, ctype)
	case LayerChunkHex:
		return ParseChunkLayer(data)
	case CelChunkHex:
		return ParseChunkCel(data)
	case CelExtraChunkHex:
		return ParseChunkCelExtra(data)
	case ColorProfileChunkHex:
		return ParseChunkColorProfile(data)
	case ExternalFilesChunkHex:
		return ParseChunkExternalFiles(data)
	case MaskChunkHex:
		panic("mask chunk deprecated")
	case PathChunkHex:
		panic("path chunk never used")
	case TagsChunkHex:
		return ParseChunkTag(data)
	case PaletteChunkHex:
		return ParseChunkPalette(data)
	case UserDataChunkHex:
		return ParseChunkUserData(data)
	case SliceChunkHex:
		return ParseChunkSlice(data)
	case TilesetChunkHex:
		return ParseChunkTileset(data)
	default:
		panic("unreachable: Invalid chunk type: " + fmt.Sprint(ctype))
	}
}

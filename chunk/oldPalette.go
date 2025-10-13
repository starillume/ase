package chunk

import (
	"bytes"
	"fmt"

	"github.com/starillume/ase/common"
)

type ChunkOldPalette struct {
	Colors []ChunkOldPaletteColor
}

type ChunkOldPalette2 struct{ *ChunkOldPalette } // NOTE: same thing (memory-wise), but each color has values between 0-63

type ChunkOldPaletteData struct {
	PacketsNumber uint16
	Packets       []ChunkOldPalettePacket
}

type ChunkOldPalettePacket struct {
	PaletteEntriesNumber byte
	ColorsNumber         byte
}

type ChunkOldPaletteColor struct {
	// values between 0-255
	R byte
	G byte
	B byte
}

func ParseChunkOldPalette(data []byte, paletteType ChunkDataType) (Chunk, error) {
	reader := bytes.NewReader(data)

	var packetsNumber uint16
	if err := common.BytesToStruct2(reader, &packetsNumber); err != nil {
		return nil, err
	}

	colors := make([]ChunkOldPaletteColor, 256)
	for i := range packetsNumber {
		var paletteEntriesNumber byte
		if err := common.BytesToStruct2(reader, &paletteEntriesNumber); err != nil {
			return nil, err
		}

		var colorsNumber byte
		if err := common.BytesToStruct2(reader, &colorsNumber); err != nil {
			return nil, err
		}
		i += uint16(paletteEntriesNumber)
		for range colorsNumber {
			var color ChunkOldPaletteColor
			if err := common.BytesToStruct2(reader, &color); err != nil {
				return nil, err
			}

			colors[i] = color
		}
	}

	switch paletteType {
	case OldPaletteChunkHex:
		return &ChunkOldPalette{
			Colors: colors,
		}, nil
	case OldPaletteChunk2Hex:
		return &ChunkOldPalette2{
			ChunkOldPalette: &ChunkOldPalette{
				Colors: colors,
			},
		}, nil
	default:
		// should be unreachable
		return nil, fmt.Errorf("Invalid chunk type for parsing OldPaletteChunk: 0x%X", paletteType)
	}
}

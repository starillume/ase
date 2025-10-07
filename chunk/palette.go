package chunk

import (
	"bytes"

	"github.com/starillume/ase/common"
)

type ChunkPalette struct {
	ChunkPaletteData
	Entries []ChunkPaletteEntry
}

type ChunkPaletteEntry struct {
	Red       byte
	Green     byte
	Blue      byte
	ColorName string
}

const ChunkPaletteDataSize = 20

type ChunkPaletteData struct {
	EntriesNumber uint32
	From          uint32
	To            uint32
	_             [8]byte
}

const ChunkPaletteEntryDataSize = 5

type ChunkPaletteEntryData struct {
	HasName uint16
	Red     byte
	Green   byte
	Blue    byte
}

func ParseChunkPalette(data []byte) (Chunk, error) {
	reader := bytes.NewReader(data)

	var cData ChunkPaletteData
	if err := common.BytesToStruct2(reader, &cData); err != nil {
		return nil, err
	}

	entries := make([]ChunkPaletteEntry, 0)
	for range cData.To - cData.From + 1 {
		entry := ChunkPaletteEntry{}
		var entryData ChunkPaletteEntryData
		if err := common.BytesToStruct2(reader, &entryData); err != nil {
			return nil, err
		}

		if entryData.HasName == 1 {
			var nameLen uint16
			if err := common.BytesToStruct2(reader, &nameLen); err != nil {
				return nil, err
			}

			colorName := make([]byte, nameLen)
			if err := common.BytesToStruct2(reader, &colorName); err != nil {
				return nil, err
			}

			entry.ColorName = string(colorName)
		}

		entries = append(entries, entry)
	}

	return &ChunkPalette{
		ChunkPaletteData: cData,
		Entries:          entries,
	}, nil
}

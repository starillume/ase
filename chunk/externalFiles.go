package chunk

import (
	"bytes"

	"github.com/starillume/ase/common"
)

type ChunkExternalFiles struct {
	ChunkExternalFilesData
	Entries []ChunkExternalFilesEntry
}

const ChunkExternalFilesDataSize = 12

type ChunkExternalFilesData struct {
	NumberEntries uint32
	_             [8]byte
}

type ChunkExternalFilesEntry struct {
	ChunkExternalFilesEntryData
	Name string
}

const ChunkExternalFilesEntryDataSize = 14

type ChunkExternalFilesEntryData struct {
	ID         uint32
	Type       byte
	_          [7]byte
	NameLength uint16
}

func ParseChunkExternalFiles(data []byte) (Chunk, error) {
	reader := bytes.NewReader(data)

	var externalFilesData ChunkExternalFilesData
	if err := common.BytesToStruct2(reader, &externalFilesData); err != nil {
		return nil, err
	}

	entries := make([]ChunkExternalFilesEntry, externalFilesData.NumberEntries)
	for i := 0; i < int(externalFilesData.NumberEntries); i++ {
		var entryData ChunkExternalFilesEntryData
		if err := common.BytesToStruct2(reader, &entryData); err != nil {
			return nil, err
		}

		entryName := make([]byte, entryData.NameLength)
		if err := common.BytesToStruct2(reader, &entryName); err != nil {
			return nil, err
		}

		entry := ChunkExternalFilesEntry{
			ChunkExternalFilesEntryData: entryData,
			Name:                        string(entryName),
		}
		entries[i] = entry
	}

	return &ChunkExternalFiles{ChunkExternalFilesData: externalFilesData, Entries: entries}, nil
}

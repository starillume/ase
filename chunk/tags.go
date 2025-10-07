package chunk

import (
	"bytes"
	"fmt"

	"github.com/starillume/ase/common"
)

type Tag struct {
	Entries []TagEntry
}

type TagEntry struct {
	TagEntryData
	Name string
}

const TagDataSize = 10

type TagData struct {
	NumberTags uint16
	_          [8]byte
}

type LoopAnimationType byte

const (
	LoopAnimationForward LoopAnimationType = iota
	LoopAnimationReverse
	LoopAnimationPingPong
	LoopAnimationPingPongReverse
)

const TagEntryDataSize = 19

type TagEntryData struct {
	FromFrame         uint16
	ToFrame           uint16
	LoopAnimationType LoopAnimationType
	Repeat            uint16
	_                 [6]byte
	Color             [3]byte
	_                 byte
	TagNameSize       uint16
}

func ParseChunkTag(data []byte) (*Tag, error) {
	reader := bytes.NewReader(data)

	cData, err := common.BytesToStruct[TagData](reader)
	if err != nil {
		return nil, err
	}

	entries := make([]TagEntry, cData.NumberTags)
	for i := range cData.NumberTags {
		entryData, err := common.BytesToStruct[TagEntryData](reader)
		if err != nil {
			return nil, err
		}

		tagName := make([]byte, entryData.TagNameSize)
		if err := common.BytesToStruct2(reader, &tagName); err != nil {
			return nil, err
		}

		fmt.Printf("b\n")
		entry := TagEntry{
			TagEntryData: entryData,
			Name:         string(tagName),
		}

		entries[i] = entry
	}

	return &Tag{
		Entries: entries,
	}, nil
}

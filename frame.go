package ase

import (
	"bytes"
	"fmt"

	"github.com/starillume/ase/chunk"
	"github.com/starillume/ase/common"
)

type FrameData struct {
	Header FrameHeader
	Chunks []chunk.Chunk
}

const FrameHeaderSize = 16
const FrameMagicNumber = 0xF1FA

type FrameHeader struct {
	FrameBytes     uint32
	MagicNumber    uint16
	OldChunkNumber uint16 // deprecated?
	FrameDuration  uint16
	_              [2]byte
	ChunkNumber    uint32
}

func resolveUserDataTags(c *chunk.UserData, tags []*tag) {
	for _, tag := range tags {
		if tag.UserData == nil {
			tag.UserData = c

			return
		}
	}
}

func parseFrame(fh FrameHeader, data []byte) (*frame, error) {
	reader := bytes.NewReader(data)

	frame := &frame{
		Cels: make([]*cel, 0),
	}

	chunkCount := 0
	if fh.ChunkNumber != 0 {
		chunkCount = int(fh.ChunkNumber)
	} else {
		chunkCount = int(fh.OldChunkNumber)
	}

	fmt.Printf("chunk count, %d\n", chunkCount)

	for range chunkCount {
		ch, err := common.BytesToStruct[chunk.Header](reader)
		if err != nil {
			return nil, err
		}

		fmt.Printf("type chunk: %x\n", ch.Type)
		chunkData := make([]byte, ch.Size-chunk.HeaderSize)

		fmt.Printf("ch, %+v\n", ch)

		if err := common.BytesToStruct2(reader, chunkData); err != nil {
			return nil, err
		}

		c, err := chunk.Parse(ch.Type, chunkData)
		if err != nil {
			return nil, err
		}

		if ch.Type == chunk.UserDataChunkHex {
			fmt.Printf("userdata: %+v\n", c)
		}
	}

	return frame, nil
}

func parseFirstFrame(fh FrameHeader, data []byte) (*frame, []*layer, []*tag, *colorProfile, error) {
	reader := bytes.NewReader(data)

	frame := &frame{
		Cels: make([]*cel, 0),
	}

	layers := make([]*layer, 0)
	tags := make([]*tag, 0)
	colorProfile := &colorProfile{}

	chunkCount := 0
	if fh.ChunkNumber != 0 {
		chunkCount = int(fh.ChunkNumber)
	} else {
		chunkCount = int(fh.OldChunkNumber)
	}

	fmt.Printf("chunk count, %d\n", chunkCount)

	var lastChunkType chunk.ChunkDataType

	for range chunkCount {
		ch, err := common.BytesToStruct[chunk.Header](reader)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		fmt.Printf("type chunk: %x\n", ch.Type)
		chunkData := make([]byte, ch.Size-chunk.HeaderSize)

		fmt.Printf("ch, %+v\n", ch)

		if err := common.BytesToStruct2(reader, chunkData); err != nil {
			return nil, nil, nil, nil, err
		}

		c, err := chunk.Parse(ch.Type, chunkData)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		switch ch.Type {
		case chunk.ColorProfileChunkHex:
			colorProfile.Chunk = &c
			lastChunkType = ch.Type
		// case chunk.OldPaletteChunkHex:
		// case chunk.OldPaletteChunk2Hex:
		// case chunk.LayerChunkHex:
		// case chunk.CelChunkHex:
		// case chunk.CelExtraChunkHex:
		// case chunk.ExternalFilesChunkHex:
		// case chunk.MaskChunkHex:
		// case chunk.PathChunkHex:
		case chunk.TagsChunkHex:
			chunkTag, ok := c.(*chunk.Tag)
			if !ok {
				panic("chunk tag cant cast")
			}
			t := &tag{
				Chunk: chunkTag,
			}
			tags = append(tags, t)
			lastChunkType = ch.Type
		// case chunk.PaletteChunkHex:
		case chunk.UserDataChunkHex:
			chunkUserData, ok := c.(*chunk.UserData)
			if !ok {
				panic("chunk userdata cant cast")
			}

			switch lastChunkType {
			case chunk.TagsChunkHex:
				fmt.Printf("userdata: %+v\n", chunkUserData)
				resolveUserDataTags(chunkUserData, tags)
			default:
				lastChunkType = chunk.UserDataChunkHex
			}
		// // case chunk.SliceChunkHex:
		// // case chunk.TilesetChunkHex:
		default:
			lastChunkType = ch.Type
		}
	}

	return frame, layers, tags, colorProfile, nil
}

func (l *Loader) ParseFrames() error {
	fh, err := BytesToStruct[FrameHeader](l, FrameHeaderSize)
	if err != nil {
		return err
	}
	if err = common.CheckMagicNumber(FrameMagicNumber, fh.MagicNumber, "frameheader "+fmt.Sprint(1)); err != nil {
		return err
	}

	chunksSize := int(fh.FrameBytes) - FrameHeaderSize
	chunksBytes := make([]byte, chunksSize)
	if err := l.BytesToStructV2(chunksSize, chunksBytes); err != nil {
		return err
	}

	fmt.Printf("\nframeId: 0\n")
	parseFirstFrame(fh, chunksBytes)

	header := l.Ase.Header

	for i := range header.Frames - 1 {
		index := i + 1
		fmt.Printf("\nframeId: %d\n", index)

		fh, err := BytesToStruct[FrameHeader](l, FrameHeaderSize)
		if err != nil {
			return err
		}
		if err = common.CheckMagicNumber(FrameMagicNumber, fh.MagicNumber, "frameheader "+fmt.Sprint(1)); err != nil {
			return err
		}

		chunksSize := int(fh.FrameBytes) - FrameHeaderSize
		chunksBytes := make([]byte, chunksSize)
		if err := l.BytesToStructV2(chunksSize, chunksBytes); err != nil {
			return err
		}

		parseFrame(fh, chunksBytes)
	}

	return nil
}

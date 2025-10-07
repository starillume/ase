package ase

import (
	"fmt"
	"image/color"

	"github.com/starillume/ase/chunk"
	"github.com/starillume/ase/common"
)

type FrameData struct {
	Header FrameHeader
	Chunks []chunk.Chunk
}

const FrameHeaderSize = 16

type FrameHeader struct {
	FrameBytes     uint32
	MagicNumber    uint16
	OldChunkNumber uint16 // deprecated?
	FrameDuration  uint16
	_              [2]byte
	ChunkNumber    uint32
}

func resolveChunkTag(c *chunk.Tag, frames []*Frame) []*Tag {
	tags := make([]*Tag, 0, len(c.Entries))

	for _, chunkTag := range c.Entries {
		from := int(chunkTag.FromFrame)
		to := int(chunkTag.ToFrame)
		tagColor := color.RGBA{R: chunkTag.Color[0], B: chunkTag.Color[1], G: chunkTag.Color[2], A: 1}

		tag := &Tag{
			Name:              chunkTag.Name,
			From:              from,
			To:                to,
			Frames:            frames[from:to],
			LoopAnimationType: chunkTag.LoopAnimationType,
			Repeat:            int(chunkTag.Repeat),
			Color:             tagColor,
		}

		tags = append(tags, tag)
	}

	return tags
}

func resolveUserDataTags(c *chunk.UserData, tags []*Tag) {
	for _, tag := range tags {
		if tag.UserData == nil {
			tag.UserData = c

			return
		}
	}
}

func resolveChunkUserData(c *chunk.UserData, asefile *AsepriteFile, lastChunkType chunk.ChunkDataType) chunk.ChunkDataType {
	fmt.Printf("userdata last type: %x\n", lastChunkType)
	switch lastChunkType {
	case chunk.TagsChunkHex:
		fmt.Printf("userdata: %+v\n", c)
		tags := asefile.Tags

		resolveUserDataTags(c, tags)
		return chunk.TagsChunkHex
	}

	return chunk.UserDataChunkHex
}

func (l *Loader) ParseFirstFrame(header *Header) error {
	index := 0
	fh, err := BytesToStruct[FrameHeader](l, FrameHeaderSize)
	if err != nil {
		return err
	}
	if err = common.CheckMagicNumber(0xF1FA, fh.MagicNumber, "frameheader "+fmt.Sprint(index)); err != nil {
		return err
	}

	chunkList := make([]chunk.Chunk, 0)
	frame := l.Ase.Frames[0]
	frame.Duration = int(fh.FrameDuration)

	var lastChunkType chunk.ChunkDataType = 0

	// TODO: verificar o numero antigo
	for range fh.ChunkNumber {
		ch, err := BytesToStruct[chunk.Header](l, chunk.HeaderSize)
		if err != nil {
			return err
		}

		data, err := l.loadFrameChunkData(ch)
		if err != nil {
			return err
		}

		c, err := chunk.Parse(ch.Type, data)
		if err != nil {
			return err
		}

		fmt.Printf("type chunk: %x\n", ch.Type)

		switch ch.Type {
		// case chunk.OldPaletteChunkHex:
		// case chunk.OldPaletteChunk2Hex:
		// case chunk.LayerChunkHex:
		// case chunk.CelChunkHex:
		// case chunk.CelExtraChunkHex:
		// case chunk.ColorProfileChunkHex:
		// case chunk.ExternalFilesChunkHex:
		// case chunk.MaskChunkHex:
		// case chunk.PathChunkHex:
		case chunk.TagsChunkHex:
			chunkTags, ok := c.(*chunk.Tag)
			if !ok {
				panic("chunk tag cant cast")
			}

			l.Ase.Tags = resolveChunkTag(chunkTags, l.Ase.Frames)
			lastChunkType = ch.Type
		// case chunk.PaletteChunkHex:
		case chunk.UserDataChunkHex:
			chunkUserData, ok := c.(*chunk.UserData)
			if !ok {
				panic("chunk userdata cant cast")
			}

			lastChunkType = resolveChunkUserData(chunkUserData, l.Ase, lastChunkType)
		// case chunk.SliceChunkHex:
		// case chunk.TilesetChunkHex:
		default:
			lastChunkType = ch.Type
		}

		chunkList = append(chunkList, c)
	}

	return nil
}

func (l *Loader) ParseFrame(header *Header, index int) (*FrameData, error) {
	fh, err := BytesToStruct[FrameHeader](l, FrameHeaderSize)
	if err != nil {
		return nil, err
	}
	if err = common.CheckMagicNumber(0xF1FA, fh.MagicNumber, "frameheader "+fmt.Sprint(index)); err != nil {
		return nil, err
	}

	chunkList := make([]chunk.Chunk, 0)
	frame := &FrameData{Header: fh, Chunks: chunkList}

	// TODO: verificar o numero antigo
	for range fh.ChunkNumber {
		ch, err := BytesToStruct[chunk.Header](l, chunk.HeaderSize)
		if err != nil {
			return nil, err
		}

		data, err := l.loadFrameChunkData(ch)
		if err != nil {
			return nil, err
		}

		c, err := chunk.Parse(ch.Type, data)
		if err != nil {
			return nil, err
		}

		chunkList = append(chunkList, c)
	}

	return frame, nil
}

func (l *Loader) ParseFrames() error {
	header := l.Ase.Header
	frames := make([]*Frame, header.Frames)
	for i := range frames {
		frames[i] = new(Frame)
	}
	fmt.Printf("frames a: %+v\n", frames)
	l.Ase.Frames = frames

	if err := l.ParseFirstFrame(&header); err != nil {
		return err
	}

	for i := range header.Frames - 1 {
		index := i + 1
		fmt.Printf("\nframeId: %d\n", index)

		data, err := l.ParseFrame(&header, int(index))
		if err != nil {
			return err
		}

		frame := &Frame{
			Duration: int(data.Header.FrameDuration),
			// Cels: data.,
		}

		frames[index] = frame
	}

	return nil
}

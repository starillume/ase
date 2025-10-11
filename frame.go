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
		Header: fh,
	}

	chunkCount := 0
	if fh.ChunkNumber != 0 {
		chunkCount = int(fh.ChunkNumber)
	} else {
		chunkCount = int(fh.OldChunkNumber)
	}

	for range chunkCount {
		ch, err := common.BytesToStruct[chunk.Header](reader)
		if err != nil {
			return nil, err
		}

		chunkData := make([]byte, ch.Size-chunk.HeaderSize)

		if err := common.BytesToStruct2(reader, chunkData); err != nil {
			return nil, err
		}

		c, err := chunk.Parse(ch.Type, chunkData)
		if err != nil {
			return nil, err
		}

		var lastChunkType chunk.ChunkDataType

		switch ch.Type {
		case chunk.CelChunkHex:
			cl, ok := c.(chunk.Cel)
			if !ok {
				panic("chunk cel couldn't cast")
			}
			cel := &cel{
				Chunk: cl,
			}

			frame.Cels = append(frame.Cels, cel)
			lastChunkType = ch.Type
		case chunk.CelExtraChunkHex:
			celExtra, ok := c.(*chunk.CelExtra)
			if !ok {
				panic("chunk celextra couldn't cast")
			}

			if lastChunkType != chunk.CelChunkHex {
				panic("so deus sabe irmao")
			}
			
			lastCel := frame.Cels[len(frame.Cels) - 1]
			lastCel.Extra = celExtra
			frame.Cels[len(frame.Cels) - 1] = lastCel
			lastChunkType = ch.Type

		case chunk.UserDataChunkHex:
			chunkUserData, ok := c.(*chunk.UserData)
			if !ok {
				panic("chunk userdata couldn't cast")
			}

			switch lastChunkType {
			case chunk.CelChunkHex, chunk.CelExtraChunkHex:
				lastCel := frame.Cels[len(frame.Cels) - 1]
				lastCel.UserData = chunkUserData
				frame.Cels[len(frame.Cels) - 1] = lastCel
				lastChunkType = chunk.UserDataChunkHex
			default:
				lastChunkType = chunk.UserDataChunkHex
			}
		default:
			lastChunkType = ch.Type
		}
	}

	return frame, nil
}

func parseFirstFrame(fh FrameHeader, data []byte) (*frame, []*layer, []*tag, []*slice, *externalFiles, *colorProfile, *palette, error) {
	reader := bytes.NewReader(data)

	frame := &frame{
		Cels: make([]*cel, 0),
		Header: fh,
	}

	layers := make([]*layer, 0)
	tags := make([]*tag, 0)
	slices := make([]*slice, 0)
	externalFiles := &externalFiles{}
	colorProfile := &colorProfile{}
	palette := &palette{}

	chunkCount := 0
	if fh.ChunkNumber != 0 {
		chunkCount = int(fh.ChunkNumber)
	} else {
		chunkCount = int(fh.OldChunkNumber)
	}

	var lastChunkType chunk.ChunkDataType

	for range chunkCount {
		ch, err := common.BytesToStruct[chunk.Header](reader)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}

		chunkData := make([]byte, ch.Size-chunk.HeaderSize)

		if err := common.BytesToStruct2(reader, chunkData); err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}

		c, err := chunk.Parse(ch.Type, chunkData)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}

		switch ch.Type {
		case chunk.ColorProfileChunkHex:
			colorProfile.Chunk = &c
			lastChunkType = ch.Type
		case chunk.LayerChunkHex:
			layerChunk, ok := c.(*chunk.Layer)
			if !ok {
				panic("chunk layer couldn't cast")
			}
			
			layer := &layer{
				Chunk: layerChunk,
			}

			layers = append(layers, layer)
			lastChunkType = ch.Type
		case chunk.CelChunkHex:
			cl, ok := c.(chunk.Cel)
			if !ok {
				panic("chunk cel couldn't cast")
			}

			cel := &cel{
				Chunk: cl,
			}

			frame.Cels = append(frame.Cels, cel)
			lastChunkType = ch.Type
		case chunk.CelExtraChunkHex:
			celExtra, ok := c.(*chunk.CelExtra)
			if !ok {
				panic("chunk celextra couldn't cast")
			}

			if lastChunkType != chunk.CelChunkHex {
				panic("so deus sabe irmao")
			}
			
			lastCel := frame.Cels[len(frame.Cels)]
			lastCel.Extra = celExtra
			frame.Cels[len(frame.Cels)] = lastCel
			lastChunkType = ch.Type

		case chunk.ExternalFilesChunkHex:
			externalFilesChunk, ok := c.(*chunk.ExternalFiles)
			if !ok {
				panic("chunk externalfiles couldn't cast")
			}

			externalFiles.Chunk = externalFilesChunk
			lastChunkType = ch.Type
		case chunk.TagsChunkHex:
			chunkTag, ok := c.(*chunk.Tag)
			if !ok {
				panic("chunk tag couldn't cast")
			}
			t := &tag{
				Chunk: chunkTag,
			}
			tags = append(tags, t)
			lastChunkType = ch.Type
		case chunk.OldPaletteChunkHex, chunk.OldPaletteChunk2Hex, chunk.PaletteChunkHex:
			palette.Chunk = &c
			lastChunkType = ch.Type
		case chunk.UserDataChunkHex:
			chunkUserData, ok := c.(*chunk.UserData)
			if !ok {
				panic("chunk userdata couldn't cast")
			}

			switch lastChunkType {
			case chunk.TagsChunkHex:
				resolveUserDataTags(chunkUserData, tags)
			case chunk.ColorProfileChunkHex:
				colorProfile.UserData = chunkUserData
				lastChunkType = chunk.UserDataChunkHex
			case chunk.LayerChunkHex:
				lastLayer := layers[len(layers) - 1]
				lastLayer.UserData = chunkUserData
				layers[len(layers) - 1] = lastLayer
				lastChunkType = chunk.UserDataChunkHex
			case chunk.CelChunkHex, chunk.CelExtraChunkHex:
				lastCel := frame.Cels[len(frame.Cels) - 1]
				lastCel.UserData = chunkUserData
				frame.Cels[len(frame.Cels) - 1] = lastCel
				lastChunkType = chunk.UserDataChunkHex
			case chunk.OldPaletteChunkHex, chunk.OldPaletteChunk2Hex, chunk.PaletteChunkHex:
				palette.UserData = chunkUserData 
				lastChunkType = chunk.UserDataChunkHex
			default:
				lastChunkType = chunk.UserDataChunkHex
			}
		case chunk.SliceChunkHex:
			sliceChunk := c.(*chunk.Slice)
			s := &slice{
				Chunk: sliceChunk,
			}

			slices = append(slices, s)
			lastChunkType = ch.Type

		// case chunk.TilesetChunkHex: mo trampo kk
		// case chunk.MaskChunkHex: deprecated
		// case chunk.PathChunkHex: unused
		default:
			lastChunkType = ch.Type
		}
	}

	return frame, layers, tags, slices, externalFiles, colorProfile, palette, nil
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

	frame, layers, tags, slices, externalFiles, colorProfile, palette, err := parseFirstFrame(fh, chunksBytes)
	if err != nil {
		return err
	}

	l.Ase.Frames = append(l.Ase.Frames, frame)
	l.Ase.Layers = layers
	l.Ase.Tags = tags
	l.Ase.Slices = slices
	l.Ase.ExternalFiles = externalFiles
	l.Ase.ColorProfile = colorProfile
	l.Ase.Palette = palette

	header := l.Ase.Header
	for range header.Frames - 1 {
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

		frame, err := parseFrame(fh, chunksBytes)
		if err != nil {
			return err
		}

		l.Ase.Frames = append(l.Ase.Frames, frame)
	}

	return nil
}

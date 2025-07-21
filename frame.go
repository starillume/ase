package ase

import "fmt"

type Frame struct {
	Header FrameHeader
	Chunks []Chunk
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

func (l *Loader) ParseFrames(header *Header) ([]Frame, error) {
	frames := make([]Frame, 0)

	for i := range header.Frames {
		fh, err := BytesToStruct[FrameHeader](l, FrameHeaderSize)
		if err != nil {
			return nil, err
		}
		err = checkMagicNumber(0xF1FA, fh.MagicNumber, "frameheader "+fmt.Sprint(i))
		if err != nil {
			return nil, err
		}

		fmt.Printf("\nframeId: %d\n", i)

		chunkList := make([]Chunk, 0)
		// TODO: verificar o numero antigo
		for range fh.ChunkNumber {
			ch, err := BytesToStruct[ChunkHeader](l, ChunkHeaderSize)
			if err != nil {
				return nil, err
			}

			var c Chunk
			c, err = l.ParseChunk(ch, int(i))
			if err != nil {
				return nil, err
			}

			chunkList = append(chunkList, c)
		}

		frames = append(frames, Frame{Header: fh, Chunks: chunkList})
		// NOTE: depois de ler dar um reset no buffer? Vê se tem um resto guardar e depois
		// coloca no buffer de novo
	}

	return frames, nil
}

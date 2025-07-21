package ase

const HeaderSize = 128

type Header struct {
	FileSize     uint32
	MagicNumber  uint16
	Frames       uint16
	Width        uint16
	Height       uint16
	ColorDepth   ColorDepth
	Flags        uint32
	FrameSpeed   uint16 // deprecated
	_            [2]uint32
	PaletteEntry byte
	_            [3]byte
	NumberColors uint16
	PixelWidth   byte
	PixelHeight  byte
	GridX        int16
	GridY        int16
	GridWidth    uint16
	GridHeight   uint16
	_            [84]byte
}

func (l *Loader) ParseHeader() (Header, error) {
	header, err := BytesToStruct[Header](l, ChunkHeaderSize)
	if err != nil {
		return header, err
	}

	err = checkMagicNumber(0xA5E0, header.MagicNumber, "header")
	if err != nil {
		return header, err
	}

	return header, nil
}

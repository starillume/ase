package ase

import (
	"github.com/starillume/ase/common"
	"github.com/starillume/ase/pixel"
)

const HeaderSize = 128

type Header struct {
	FileSize     uint32
	MagicNumber  uint16
	Frames       uint16
	Width        uint16
	Height       uint16
	ColorDepth   pixel.ColorDepth
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

func (l *Loader) ParseHeader() error {
	header, err := BytesToStruct[Header](l, HeaderSize)
	if err != nil {
		return err
	}

	err = common.CheckMagicNumber(0xA5E0, header.MagicNumber, "header")
	if err != nil {
		return err
	}

	l.Ase.Header = header

	return nil
}

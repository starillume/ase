package common

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func BytesToStruct[T any](reader *bytes.Reader) (T, error) {
	var t T
	err := binary.Read(reader, binary.LittleEndian, &t)
	if err != nil {
		return t, err
	}

	return t, nil
}

func BytesToStruct2(reader *bytes.Reader, t any) error {
	err := binary.Read(reader, binary.LittleEndian, t)
	if err != nil {
		return err
	}

	return nil
}

func CheckMagicNumber(magic, number uint16, from string) error {
	if number != magic {
		return fmt.Errorf("%s: magic number fail (got 0x%X, want 0x%X)", from, number, magic)
	}

	return nil
}

// func (l *Loader) GetPixels(ch ChunkHeader, compressed bool, pixelDataSize int) (Pixels, error) {
// 	var pbuf []byte
// 	if compressed {
// 		fmt.Printf("pixel data size compressed: %d\n", pixelDataSize)
//
// 		pixelsCompressed := make(PixelsZlib, pixelDataSize)
// 		if err := l.BytesToStructV2(pixelDataSize, &pixelsCompressed); err != nil {
// 			return nil, err
// 		}
//
// 		var err error
// 		pbuf, err = pixelsCompressed.Decompress()
// 		if err != nil {
// 			return nil, err
// 		}
// 	} else {
// 		pbuf = make([]byte, pixelDataSize)
// 		if err := l.BytesToStructV2(pixelDataSize, &pbuf); err != nil {
// 			return nil, err
// 		}
// 	}
//
// 	fmt.Printf("pixels decompressed size: %d\n", len(pbuf))
//
// 	return l.ResolvePixelType(pbuf), nil
// }
//

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

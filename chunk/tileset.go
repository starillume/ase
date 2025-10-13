package chunk

import (
	"bytes"
	// "compress/zlib"
	// "fmt"
	// "io"
	"reflect"

	"github.com/starillume/ase/common"
)

type ChunkTileset struct {
	ChunkTilesetData
	Name  string
	Flags ChunkTilesetFlags
	*ChunkTilesetLinkExternalFileData
	// TilesetImage *Pixels
}

const ChunkTilesetDataSize = 34

type ChunkTilesetData struct {
	ID          uint32
	FlagsBit    uint32
	TilesNumber uint32
	TileWidth   uint16
	TileHeight  uint16
	BaseIndex   int16
	_           [14]byte
	NameLength  uint16
}

type ChunkTilesetFlags struct {
	LinkExternalFile bool
	LinkTiles        bool
	UseTileID0       bool
	AutoFlipX        bool
	AutoFlipY        bool
	AutoFlipD        bool
}

const ChunkTilesetLinkExternalFileDataSize = 8

type ChunkTilesetLinkExternalFileData struct {
	ExternalFileID        uint32
	ExternalFileTilesetID uint32
}

func ParseChunkTileset(data []byte) (Chunk, error) {
	reader := bytes.NewReader(data)

	var tilesetData ChunkTilesetData
	if err := common.BytesToStruct2(reader, &tilesetData); err != nil {
		return nil, err
	}

	nameBytes := make([]byte, tilesetData.NameLength)
	if err := common.BytesToStruct2(reader, &nameBytes); err != nil {
		return nil, err
	}

	var flags ChunkTilesetFlags = ChunkTilesetFlags{}
	value := reflect.ValueOf(&flags).Elem()
	fieldIndex := 0
	for i := uint32(1); i < 33; i *= 2 {
		field := value.Field(fieldIndex)
		if tilesetData.FlagsBit&i == i && field.CanSet() {
			field.SetBool(true)
		}
		fieldIndex++
	}

	chunk := ChunkTileset{
		ChunkTilesetData:                 tilesetData,
		Name:                             string(nameBytes),
		Flags:                            flags,
		ChunkTilesetLinkExternalFileData: nil,
		// TilesetImage:                     nil,
	}

	// if chunk.Flags.LinkExternalFile {
	// 	var externalFileData ChunkTilesetLinkExternalFileData
	// 	if err := common.BytesToStruct2(reader, &externalFileData); err != nil {
	// 		return nil, err
	// 	}
	//
	// 	chunk.ChunkTilesetLinkExternalFileData = &externalFileData
	// }
	//
	// if chunk.Flags.LinkTiles {
	// 	var pixelDataSize uint32
	// 	if err := common.BytesToStruct2(reader, &pixelDataSize); err != nil {
	// 		return nil, err
	// 	}
	//
	// 	fmt.Printf("pixel data size compressed: %d\n", pixelDataSize)
	//
	// 	pixelsCompressed := make(PixelsZlib, pixelDataSize)
	// 	if err := reader.BytesToStruct2(reader, &pixelsCompressed); err != nil {
	// 		return nil, err
	// 	}
	//
	// 	buffer := bytes.NewBuffer(pixelsCompressed)
	// 	r, err := zlib.NewReader(buffer)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	//
	// 	defer r.Close()
	//
	// 	d := new(bytes.Buffer)
	//
	// 	if _, err := io.Copy(d, r); err != nil {
	// 		return nil, err
	// 	}
	//
	// 	fmt.Printf("pixels decompressed size: %d\n", len(d.Bytes()))
	//
	// 	bytesTo4ByteChunks := func(data []byte) [][4]byte {
	// 		var chunks [][4]byte
	// 		for i := 0; i < len(data); i += 4 {
	// 			var block [4]byte
	// 			copy(block[:], data[i:i+4])
	// 			chunks = append(chunks, block)
	// 		}
	// 		return chunks
	// 	}
	//
	// 	t := bytesTo4ByteChunks(d.Bytes())
	//
	// 	fmt.Printf("teste: %v\n", PixelsRGBA(t)[0])
	//
	// 	tilesetImage := Pixels(t)
	//
	// 	chunk.TilesetImage = &tilesetImage
	// }

	return &chunk, nil
}

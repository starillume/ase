package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"os"
	"reflect"
	"sync"
)

const (
	ChunkSize = 128
)

type AsepriteFile struct {
	Header Header
	Frames []Frame
}

type Loader struct {
	Reader io.Reader
	Buf    []byte
	Buffer *bytes.Buffer
	File   *AsepriteFile
}

type ColorDepth uint16

const (
	ColorDepthRGBA      ColorDepth = 32
	ColorDepthGrayscale ColorDepth = 16
	ColorDepthIndexed   ColorDepth = 8
)

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

type ChunkDataType uint16

const (
	OldPaletteChunkHex    ChunkDataType = 0x0004
	OldPaletteChunk2Hex   ChunkDataType = 0x0011
	LayerChunkHex         ChunkDataType = 0x2004
	CelChunkHex           ChunkDataType = 0x2005
	CelExtraChunkHex      ChunkDataType = 0x2006
	ColorProfileChunkHex  ChunkDataType = 0x2007
	ExternalFilesChunkHex ChunkDataType = 0x2008
	MaskChunkHex          ChunkDataType = 0x2016 // DEPRECATED
	PathChunkHex          ChunkDataType = 0x2017 // NEVER USED
	TagsChunkHex          ChunkDataType = 0x2018
	PaletteChunkHex       ChunkDataType = 0x2019
	UserDataChunkHex      ChunkDataType = 0x2020
	SliceChunkHex         ChunkDataType = 0x2022
	TilesetChunkHex       ChunkDataType = 0x2023
)

const ChunkHeaderSize = 6

type ChunkHeader struct {
	Size uint32
	Type ChunkDataType
}

type Chunk interface {
	GetHeader() ChunkHeader
	GetType() ChunkDataType
}

type ColorProfileType uint16

const (
	ColorProfileNone ColorProfileType = iota
	ColorProfileSRGB
	ColorProfileICC
)

type ChunkColorProfile struct {
	header ChunkHeader
	ChunkColorProfileData
}

type ChunkColorProfileICC struct {
	ChunkColorProfile
	ChunkColorProfileICCData
}

func (c *ChunkColorProfile) GetHeader() ChunkHeader {
	return c.header
}

func (c *ChunkColorProfile) GetType() ChunkDataType {
	return c.header.Type
}

const ChunkColorProfileDataSize = 16

type Fixed int32

func FloatToFixed(f float64) Fixed {
	return Fixed(f * 65536)
}

func (f *Fixed) FixedToFloat() float64 {
	return float64(*f) / 65536
}

type ChunkColorProfileData struct {
	Type  ColorProfileType
	Flags uint16
	Gamma Fixed
	_     [8]byte
}

type ChunkColorProfileICCData struct {
	DataLength uint32
	Data       []byte
}

type ChunkOldPalette struct {
	header ChunkHeader
	Colors []ChunkOldPaletteColor
}

type ChunkOldPalette2 struct{ *ChunkOldPalette } // NOTE: same thing (memory-wise), but each color has values between 0-63

type ChunkOldPaletteData struct {
	PacketsNumber uint16
	Packets       []ChunkOldPalettePacket
}

type ChunkOldPalettePacket struct {
	PaletteEntriesNumber byte
	ColorsNumber         byte
}

type ChunkOldPaletteColor struct {
	// values between 0-255
	R byte
	G byte
	B byte
}

func (c *ChunkOldPalette) GetHeader() ChunkHeader {
	return c.header
}

func (c *ChunkOldPalette) GetType() ChunkDataType {
	return c.header.Type
}

type ChunkLayer struct {
	header         ChunkHeader
	ChunkLayerData ChunkLayerData
	ChunkLayerName
	ChunkLayerFlags
	*ChunkLayerType2Data
	*ChunkLayerLockMovementData
}

const ChunkLayerDataSize = 18

type ChunkLayerData struct {
	FlagsBit      uint16
	Type          uint16
	ChildLevel    uint16
	DefaultWidth  uint16
	DefaultHeight uint16
	BlendMode     uint16
	Opacity       byte
	_             [3]byte
	NameLength    uint16
}

type ChunkLayerFlags struct {
	Visible                    bool
	Editable                   bool
	LockMovement               bool
	Background                 bool
	PreferLinkedCels           bool
	LayerGroupDisplayCollapsed bool
	ReferenceLayer             bool
}

type ChunkLayerName string

type ChunkLayerType2Data struct {
	TilesetIndex uint32
}

type ChunkLayerLockMovementData struct {
	LayerUUID [16]byte
}

func (c *ChunkLayer) GetHeader() ChunkHeader {
	return c.header
}

func (c *ChunkLayer) GetType() ChunkDataType {
	return c.header.Type
}

type ChunkPalette struct {
	header ChunkHeader
	ChunkPaletteData
	Entries []ChunkPaletteEntry
}

func (c *ChunkPalette) GetHeader() ChunkHeader {
	return c.header
}

func (c *ChunkPalette) GetType() ChunkDataType {
	return c.header.Type
}

type ChunkPaletteEntry struct {
	Red       byte
	Green     byte
	Blue      byte
	ColorName string
}

const ChunkPaletteDataSize = 20

type ChunkPaletteData struct {
	EntriesNumber uint32
	From          uint32
	To            uint32
	_             [8]byte
}

const ChunkPaletteEntryDataSize = 5

type ChunkPaletteEntryData struct {
	HasName uint16
	Red     byte
	Green   byte
	Blue    byte
}

type ChunkCelExtra struct {
	header ChunkHeader
	ChunkCelExtraData
}

const ChunkCelExtraDataSize = 36

type ChunkCelExtraData struct {
	Flags  uint32
	X      Fixed
	Y      Fixed
	Width  Fixed
	Height Fixed
	_      [16]byte
}

func (c *ChunkCelExtra) GetHeader() ChunkHeader {
	return c.header
}

func (c *ChunkCelExtra) GetType() ChunkDataType {
	return c.header.Type
}

type ChunkExternalFiles struct {
	header ChunkHeader
	ChunkExternalFilesData
	Entries []ChunkExternalFilesEntry
}

const ChunkExternalFilesDataSize = 12

type ChunkExternalFilesData struct {
	NumberEntries uint32
	_             [8]byte
}

type ChunkExternalFilesEntry struct {
	ChunkExternalFilesEntryData
	Name string
}

const ChunkExternalFilesEntryDataSize = 14

type ChunkExternalFilesEntryData struct {
	ID         uint32
	Type       byte
	_          [7]byte
	NameLength uint16
}

func (c *ChunkExternalFiles) GetHeader() ChunkHeader {
	return c.header
}

func (c *ChunkExternalFiles) GetType() ChunkDataType {
	return c.header.Type
}

type ChunkTag struct {
	header  ChunkHeader
	Entries []ChunkTagEntry
}

func (c *ChunkTag) GetHeader() ChunkHeader {
	return c.header
}

func (c *ChunkTag) GetType() ChunkDataType {
	return c.header.Type
}

type ChunkTagEntry struct {
	ChunkTagEntryData
	Name string
}

const ChunkTagDataSize = 10

type ChunkTagData struct {
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

const ChunkTagEntryDataSize = 19

type ChunkTagEntryData struct {
	FromFrame         uint16
	ToFrame           uint16
	LoopAnimationType LoopAnimationType
	Repeat            uint16
	_                 [6]byte
	Color             [3]byte
	_                 byte
	TagNameSize       uint16
}

type ChunkSlice struct {
	header ChunkHeader
	ChunkSliceData
	Name string
	Keys []ChunkSliceKey
}

const ChunkSliceDataSize = 14

type ChunkSliceData struct {
	NumberSliceKeys uint32
	FlagsBit        uint32
	_               uint32
	NameLength      uint16
}

type ChunkSliceKey struct {
	ChunkSliceKeyData
	*ChunkSliceKey9PatchesData
	*ChunkSliceKeyPivotData
}

const ChunkSliceKeyDataSize = 20

type ChunkSliceKeyData struct {
	FrameNumber uint32
	OriginX     int32
	OriginY     int32
	Width       uint32 // (can be 0 if this slice hidden in the animation from the given frame)
	Height      uint32
}

const ChunkSliceKey9PatchesDataSize = 16

type ChunkSliceKey9PatchesData struct {
	CenterX      int32
	CenterY      int32
	CenterWidth  uint32
	CenterHeight uint32
}

const ChunkSliceKeyPivotDataSize = 8

type ChunkSliceKeyPivotData struct {
	X int32
	Y int32
}

func (c *ChunkSlice) GetHeader() ChunkHeader {
	return c.header
}

func (c *ChunkSlice) GetType() ChunkDataType {
	return c.header.Type
}

type ChunkTileset struct {
	header ChunkHeader
	ChunkTilesetData
	Name  string
	Flags ChunkTilesetFlags
	*ChunkTilesetLinkExternalFileData
	TilesetImage *Pixels
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

func (c *ChunkTileset) GetHeader() ChunkHeader {
	return c.header
}

func (c *ChunkTileset) GetType() ChunkDataType {
	return c.header.Type
}

type CelType uint16

const (
	CelTypeRawImage CelType = iota
	CelTypeLinked
	CelTypeCompressedImage
	CelTypeCompressedTilemap
)

type ChunkCelImage struct {
	header ChunkHeader
	ChunkCelData
	ChunkCelRawImageData
}

func (c *ChunkCelImage) GetHeader() ChunkHeader {
	return c.header
}

func (c *ChunkCelImage) GetType() ChunkDataType {
	return c.header.Type
}

type ChunkCelTilemap struct {
	header ChunkHeader
	ChunkCelData
	ChunkCelCompressedTilemapData
}

func (c *ChunkCelTilemap) GetHeader() ChunkHeader {
	return c.header
}

func (c *ChunkCelTilemap) GetType() ChunkDataType {
	return c.header.Type
}

type ChunkCelLinked struct {
	header ChunkHeader
	ChunkCelData
	ChunkCelLinkedData
}

func (c *ChunkCelLinked) GetHeader() ChunkHeader {
	return c.header
}

func (c *ChunkCelLinked) GetType() ChunkDataType {
	return c.header.Type
}

const ChunkCelDataSize = 16

type ChunkCelData struct {
	LayerIndex uint16
	X          int16
	Y          int16
	Opacity    byte
	CelType    // not flag
	Z          int16
	_          [5]byte
}

type Pixels any

type PixelsCompressed interface {
	Decompress() (Pixels, error)
}

type PixelsIndexed []byte
type PixelsGrayscale [][2]byte
type PixelsRGBA [][4]byte
type PixelsZlib []byte

func (p *PixelsZlib) Decompress() ([]byte, error) {
	buffer := bytes.NewBuffer(*p)
	r, err := zlib.NewReader(buffer)
	if err != nil {
		return nil, err
	}

	defer r.Close()

	d := new(bytes.Buffer)

	if _, err := io.Copy(d, r); err != nil {
		return nil, err
	}

	return d.Bytes(), nil
}

func (p *PixelsRGBA) ToImage(celX, celY, width, height, canvasWidth, canvasHeight int, frameId int) image.Image {
	rect := image.Rect(0, 0, canvasWidth, canvasHeight)
	img := image.NewRGBA(rect)
	pixels := *p
	for y := range height {
		for x := range width {
			i := y*width + x
			color := color.RGBA{
				R: pixels[i][0],
				G: pixels[i][1],
				B: pixels[i][2],
				A: pixels[i][3],
			}
			img.Set(x+celX, y+celY, color)
		}
	}

	file, _ := os.Create(fmt.Sprintf("teste%d.png", frameId))
	defer file.Close()
	png.Encode(file, img)

	return img
}

const ChunkCelDimensionSize = 4

type ChunkCelDimensionData struct {
	Width  uint16
	Height uint16
}

type ChunkCelRawImageData struct {
	ChunkCelDimensionData
	Pixels Pixels
}

const ChunkCelLinkedDataSize = 2

type ChunkCelLinkedData struct {
	FramePosition int16
}

type ChunkCelCompressedImageData struct {
	ChunkCelDimensionData
	Pixels PixelsCompressed
}

const ChunkCelCompressedTilemapStaticDataSize = 28

type ChunkCelCompressedTilemapStaticData struct {
	BitsPerTile      uint16
	MaskTileId       uint32
	MaskXFlip        uint32
	MaskYFlip        uint32
	MaskDiagonalFlip uint32
	_                [10]byte
}

type ChunkCelCompressedTilemapData struct {
	ChunkCelDimensionData
	ChunkCelCompressedTilemapStaticData
	Tiles PixelsCompressed
	// NOTE: não sei como fazer esse negocio, voltar depois
	// NOTE: dica pro lerdinho acima: tile não é pixel !
}

const UserDataFlagSize = 4

type UserDataFlag uint32

const (
	UserDataHasText       UserDataFlag = 1 << 0
	UserDataHasColor      UserDataFlag = 1 << 1
	UserDataHasProperties UserDataFlag = 1 << 2
)

type ChunkUserData struct {
	header ChunkHeader
	Text   string
	Color  *ChunkUserDataColor
	Maps   *[]ChunkUserDataPropMap
}

func (c *ChunkUserData) GetHeader() ChunkHeader {
	return c.header
}

func (c *ChunkUserData) GetType() ChunkDataType {
	return c.header.Type
}

type ChunkUserDataPropMap struct {
	External bool
	Props    map[string]any
}

type ChunkUserDataTextSize uint32

const ChunkUserDataColorSize = 4

type ChunkUserDataColor struct {
	R byte
	G byte
	B byte
	A byte
}

const ChunkUserDataPropMapHeaderSize = 8

type ChunkUserDataPropMapHeader struct {
	SizeInBytes    uint32
	PropMapNumbers uint32
}

const ChunkUserDataPropMapDataSize = 8

type ChunkUserDataPropMapData struct {
	PropKey     uint32
	PropNumbers uint32
}

type UserDataPropType uint16

const (
	UserDataNil UserDataPropType = iota
	UserDataBool
	UserDataInt8
	UserDataUint8
	UserDataInt16
	UserDataUint16
	UserDataInt32
	UserDataUint32
	UserDataInt64
	UserDataUint64
	UserDataFixed
	UserDataFloat
	UserDataDouble
	UserDataString
	UserDataPoint
	UserDataSize
	UserDataRect
	UserDataVector
	UserDataProp
	UserDataUUID
)

func checkMagicNumber(magic, number uint16, from string) error {
	if number != magic {
		return fmt.Errorf("%s: magic number fail (got 0x%X, want 0x%X)", from, number, magic)
	}

	return nil
}

func (l *Loader) readToBuffer() error {
	_, err := l.Reader.Read(l.Buf)
	if err != nil {
		return err
	}

	_, err = l.Buffer.Write(l.Buf)
	if err != nil {
		return err
	}

	return nil
}

func (l *Loader) enoughSpaceToRead(size int) bool {
	available := l.Buffer.Len()
	needed := size
	return available >= needed
}

// NOTE: DEPRECATED: use v2
func BytesToStruct[T any](loader *Loader, size int) (T, error) {
	var t T
	if loader.enoughSpaceToRead(size) {
		err := binary.Read(loader.Buffer, binary.LittleEndian, &t)
		if err != nil {
			return t, err
		}
		return t, nil
	}

	err := loader.readToBuffer()
	if err != nil {
		return t, err
	}

	return BytesToStruct[T](loader, size)
}

func (l *Loader) BytesToStructV2(size int, t any) error {
	if l.enoughSpaceToRead(size) {
		err := binary.Read(l.Buffer, binary.LittleEndian, t)
		if err != nil {
			return err
		}
		return nil
	}

	err := l.readToBuffer()
	if err != nil {
		return err
	}

	return l.BytesToStructV2(size, t)
}

func (l *Loader) loadFrameChunkData(ch ChunkHeader) ([]byte, error) {
	size := ch.Size - ChunkHeaderSize
	if l.enoughSpaceToRead(int(size)) {
		bufchunk := make([]byte, size)
		_, err := io.ReadFull(l.Buffer, bufchunk)
		if err != nil {
			return nil, err
		}
		return bufchunk, nil
	}

	err := l.readToBuffer()
	if err != nil {
		return nil, err
	}

	return l.loadFrameChunkData(ch)
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

	fmt.Printf("header width: %d, height: %d\n", header.Width, header.Height)

	return header, nil
}

func (l *Loader) ParseChunk(ch ChunkHeader, frameId int) (Chunk, error) {
	var chunk Chunk

	// fmt.Printf("frameId: %d chunk type: %x\n", frameId, ch.Type)
	switch ch.Type {
	case ColorProfileChunkHex:
		return l.ParseChunkColorProfile(ch)
	case OldPaletteChunkHex:
		return l.ParseChunkOldPalette(ch)
	case OldPaletteChunk2Hex:
		return l.ParseChunkOldPalette(ch)
	case LayerChunkHex:
		return l.ParseChunkLayer(ch)
	case PaletteChunkHex:
		return l.ParseChunkPalette(ch)
	case CelExtraChunkHex:
		return l.ParseChunkCelExtra(ch)
	case ExternalFilesChunkHex:
		return l.ParseChunkExternalFiles(ch)
	case TagsChunkHex:
		return l.ParseChunkTag(ch)
	case SliceChunkHex:
		return l.ParseChunkSlice(ch)
	case TilesetChunkHex:
		return l.ParseChunkTileset(ch)
	case CelChunkHex:
		return l.ParseChunkCel(ch, frameId)
	case UserDataChunkHex:
		return l.ParseChunkUserData(ch)
	case MaskChunkHex:
		l.loadFrameChunkData(ch)
		return chunk, nil
	case PathChunkHex:
		l.loadFrameChunkData(ch)
		return chunk, nil
	default:
		panic("unreachable: Invalid chunk type: " + fmt.Sprint(ch.Type))
	}
}

func (l *Loader) ParseChunkTileset(ch ChunkHeader) (Chunk, error) {
	var tilesetData ChunkTilesetData
	if err := l.BytesToStructV2(ChunkTilesetDataSize, &tilesetData); err != nil {
		return nil, err
	}

	nameBytes := make([]byte, tilesetData.NameLength)
	if err := l.BytesToStructV2(int(tilesetData.NameLength), &nameBytes); err != nil {
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
		header:                           ch,
		ChunkTilesetData:                 tilesetData,
		Name:                             string(nameBytes),
		Flags:                            flags,
		ChunkTilesetLinkExternalFileData: nil,
		TilesetImage:                     nil,
	}

	if chunk.Flags.LinkExternalFile {
		var externalFileData ChunkTilesetLinkExternalFileData
		if err := l.BytesToStructV2(ChunkTilesetLinkExternalFileDataSize, &externalFileData); err != nil {
			return nil, err
		}

		chunk.ChunkTilesetLinkExternalFileData = &externalFileData
	}

	if chunk.Flags.LinkTiles {
		var pixelDataSize uint32
		if err := l.BytesToStructV2(4, &pixelDataSize); err != nil {
			return nil, err
		}

		fmt.Printf("pixel data size compressed: %d\n", pixelDataSize)

		pixelsCompressed := make(PixelsZlib, pixelDataSize)
		if err := l.BytesToStructV2(int(pixelDataSize), &pixelsCompressed); err != nil {
			return nil, err
		}

		buffer := bytes.NewBuffer(pixelsCompressed)
		r, err := zlib.NewReader(buffer)
		if err != nil {
			return nil, err
		}

		defer r.Close()

		d := new(bytes.Buffer)

		if _, err := io.Copy(d, r); err != nil {
			return nil, err
		}

		fmt.Printf("pixels decompressed size: %d\n", len(d.Bytes()))

		bytesTo4ByteChunks := func(data []byte) [][4]byte {
			var chunks [][4]byte
			for i := 0; i < len(data); i += 4 {
				var block [4]byte
				copy(block[:], data[i:i+4])
				chunks = append(chunks, block)
			}
			return chunks
		}

		t := bytesTo4ByteChunks(d.Bytes())

		fmt.Printf("teste: %v\n", PixelsRGBA(t)[0])

		tilesetImage := Pixels(t)

		chunk.TilesetImage = &tilesetImage
	}

	return &chunk, nil
}

func (l *Loader) ParseChunkSlice(ch ChunkHeader) (Chunk, error) {
	var chunkSliceData ChunkSliceData
	if err := l.BytesToStructV2(ChunkSliceDataSize, &chunkSliceData); err != nil {
		return nil, err
	}

	chunkSliceName := make([]byte, chunkSliceData.NameLength)
	if err := l.BytesToStructV2(int(chunkSliceData.NameLength), &chunkSliceName); err != nil {
		return nil, err
	}

	sliceKeys := make([]ChunkSliceKey, chunkSliceData.NumberSliceKeys)
	for i := 0; i < int(chunkSliceData.NumberSliceKeys); i++ {
		var sliceKeyData ChunkSliceKeyData
		if err := l.BytesToStructV2(ChunkSliceKeyDataSize, &sliceKeyData); err != nil {
			return nil, err
		}

		sliceKey := ChunkSliceKey{ChunkSliceKeyData: sliceKeyData, ChunkSliceKeyPivotData: nil, ChunkSliceKey9PatchesData: nil}

		if chunkSliceData.FlagsBit&1 == 1 {
			var ninePatchesData ChunkSliceKey9PatchesData
			if err := l.BytesToStructV2(ChunkSliceKey9PatchesDataSize, &ninePatchesData); err != nil {
				return nil, err
			}

			sliceKey.ChunkSliceKey9PatchesData = &ninePatchesData
		}

		if chunkSliceData.FlagsBit&2 == 2 {
			var pivotData ChunkSliceKeyPivotData
			if err := l.BytesToStructV2(ChunkSliceKeyPivotDataSize, &pivotData); err != nil {
				return nil, err
			}

			sliceKey.ChunkSliceKeyPivotData = &pivotData
		}

		sliceKeys[i] = sliceKey
	}

	return &ChunkSlice{
		header:         ch,
		ChunkSliceData: chunkSliceData,
		Name:           string(chunkSliceName),
		Keys:           sliceKeys,
	}, nil
}

type readerFunc func(*Loader) (any, error)

var (
	readers        map[UserDataPropType]readerFunc
	initReaderSync sync.Once
)

func getReaders() map[UserDataPropType]readerFunc {
	initReaderSync.Do(initReaders)
	return readers
}

func initReaders() {
	readers = map[UserDataPropType]readerFunc{
		UserDataBool: func(l *Loader) (any, error) {
			var b uint8 // BYTE
			if err := l.BytesToStructV2(1, &b); err != nil {
				return nil, err
			}
			return b != 0, nil
		},

		UserDataInt8: func(l *Loader) (any, error) {
			var v int8 // BYTE interpretado como signed
			return v, l.BytesToStructV2(1, &v)
		},

		UserDataUint8: func(l *Loader) (any, error) {
			var v uint8
			return v, l.BytesToStructV2(1, &v)
		},

		UserDataInt16: func(l *Loader) (any, error) {
			var v int16 // SHORT
			return v, l.BytesToStructV2(2, &v)
		},

		UserDataUint16: func(l *Loader) (any, error) {
			var v uint16 // WORD
			return v, l.BytesToStructV2(2, &v)
		},

		UserDataInt32: func(l *Loader) (any, error) {
			var v int32 // LONG
			return v, l.BytesToStructV2(4, &v)
		},

		UserDataUint32: func(l *Loader) (any, error) {
			var v uint32 // DWORD
			return v, l.BytesToStructV2(4, &v)
		},

		UserDataInt64: func(l *Loader) (any, error) {
			var v int64 // LONG64
			return v, l.BytesToStructV2(8, &v)
		},

		UserDataUint64: func(l *Loader) (any, error) {
			var v uint64 // QWORD
			return v, l.BytesToStructV2(8, &v)
		},

		UserDataFixed: func(l *Loader) (any, error) {
			var v Fixed // FIXED (16.16)
			return v, l.BytesToStructV2(4, &v)
		},

		UserDataFloat: func(l *Loader) (any, error) {
			var v float32 // FLOAT
			return v, l.BytesToStructV2(4, &v)
		},

		UserDataDouble: func(l *Loader) (any, error) {
			var v float64 // DOUBLE
			return v, l.BytesToStructV2(8, &v)
		},

		UserDataString: func(l *Loader) (any, error) {
			var length uint16 // WORD
			if err := l.BytesToStructV2(2, &length); err != nil {
				return nil, err
			}
			data := make([]byte, length)
			if err := l.BytesToStructV2(int(length), &data); err != nil {
				return nil, err
			}
			return string(data), nil // UTF-8, sem \0
		},

		UserDataPoint: func(l *Loader) (any, error) {
			var p struct {
				X int32 // LONG
				Y int32 // LONG
			}
			return p, l.BytesToStructV2(8, &p)
		},

		UserDataSize: func(l *Loader) (any, error) {
			var s struct {
				W int32 // LONG
				H int32 // LONG
			}
			return s, l.BytesToStructV2(8, &s)
		},

		UserDataRect: func(l *Loader) (any, error) {
			var r struct {
				X int32
				Y int32
				W int32
				H int32
			}
			return r, l.BytesToStructV2(16, &r)
		},

		UserDataVector: func(l *Loader) (any, error) {
			var count uint32 // DWORD
			if err := l.BytesToStructV2(4, &count); err != nil {
				return nil, err
			}

			var elemType uint16 // WORD
			if err := l.BytesToStructV2(2, &elemType); err != nil {
				return nil, err
			}

			var elements []any

			if elemType == 0 {
				// Heterogêneo: cada elemento tem tipo próprio
				for i := 0; i < int(count); i++ {
					var etype uint16
					if err := l.BytesToStructV2(2, &etype); err != nil {
						return nil, err
					}
					reader, ok := getReaders()[UserDataPropType(etype)]
					if !ok {
						return nil, fmt.Errorf("unsupported vector element type %d", etype)
					}
					val, err := reader(l)
					if err != nil {
						return nil, err
					}
					elements = append(elements, val)
				}
			} else {
				// Homogêneo: todos elementos têm o mesmo tipo
				reader, ok := getReaders()[UserDataPropType(elemType)]
				if !ok {
					return nil, fmt.Errorf("unsupported vector element type %d", elemType)
				}
				for i := 0; i < int(count); i++ {
					val, err := reader(l)
					if err != nil {
						return nil, err
					}
					elements = append(elements, val)
				}
			}

			return elements, nil
		},

		UserDataUUID: func(l *Loader) (any, error) {
			var uuid [16]byte
			return uuid, l.BytesToStructV2(16, &uuid)
		},

		UserDataProp: func(l *Loader) (any, error) {
			var propsLen uint32
			if err := l.BytesToStructV2(4, &propsLen); err != nil {
				return nil, err
			}

			return l.ParseUserDataProps(int(propsLen))
		},
	}
}

func (l *Loader) ParseUserDataProps(count int) (map[string]any, error) {
	props := make(map[string]any, count)
	for range count {
		var nameLen uint16
		if err := l.BytesToStructV2(2, &nameLen); err != nil {
			return nil, err
		}

		nameBytes := make([]byte, nameLen)
		if err := l.BytesToStructV2(int(nameLen), &nameBytes); err != nil {
			return nil, err
		}

		key := string(nameBytes)

		var typeValue UserDataPropType
		if err := l.BytesToStructV2(2, &typeValue); err != nil {
			return nil, err
		}

		if reader, ok := getReaders()[typeValue]; ok {
			value, err := reader(l)
			if err != nil {
				return nil, err
			}
			props[key] = value
		} else {
			return nil, fmt.Errorf("unsupported type %d", typeValue)
		}
	}

	return props, nil
}

func (l *Loader) ParseUserDataPropMap() (*ChunkUserDataPropMap, error) {
	var propMapData ChunkUserDataPropMapData
	if err := l.BytesToStructV2(ChunkUserDataPropMapHeaderSize, &propMapData); err != nil {
		return nil, err
	}

	propMap := &ChunkUserDataPropMap{
		External: (propMapData.PropKey != 0),
	}

	propsmap, err := l.ParseUserDataProps(int(propMapData.PropNumbers))
	if err != nil {
		return nil, err
	}

	propMap.Props = propsmap

	return propMap, nil
}

func (l *Loader) ParseChunkUserData(ch ChunkHeader) (Chunk, error) {
	var flag UserDataFlag
	if err := l.BytesToStructV2(4, &flag); err != nil {
		return nil, err
	}

	chunk := &ChunkUserData{
		header: ch,
	}

	if flag&UserDataHasText == UserDataHasText {
		var textLen uint16
		if err := l.BytesToStructV2(2, &textLen); err != nil {
			return nil, err
		}

		text := make([]byte, textLen)
		if err := l.BytesToStructV2(int(textLen), &text); err != nil {
			return nil, err
		}

		chunk.Text = string(text)
	}

	if flag&UserDataHasColor == UserDataHasColor {
		var userDataColor ChunkUserDataColor
		if err := l.BytesToStructV2(ChunkUserDataColorSize, &userDataColor); err != nil {
			return nil, err
		}

		chunk.Color = &userDataColor
	}

	if flag&UserDataHasProperties == UserDataHasProperties {
		var mapHeader ChunkUserDataPropMapHeader
		if err := l.BytesToStructV2(ChunkUserDataPropMapHeaderSize, &mapHeader); err != nil {
			return nil, err
		}

		propMaps := make([]ChunkUserDataPropMap, mapHeader.PropMapNumbers)
		for range mapHeader.PropMapNumbers {
			l.ParseUserDataPropMap()
		}

		chunk.Maps = &propMaps
	}

	return chunk, nil
}

func (l *Loader) ParseChunkExternalFiles(ch ChunkHeader) (Chunk, error) {
	var externalFilesData ChunkExternalFilesData
	if err := l.BytesToStructV2(ChunkExternalFilesDataSize, &externalFilesData); err != nil {
		return nil, err
	}

	entries := make([]ChunkExternalFilesEntry, externalFilesData.NumberEntries)
	for i := 0; i < int(externalFilesData.NumberEntries); i++ {
		var entryData ChunkExternalFilesEntryData
		if err := l.BytesToStructV2(ChunkExternalFilesEntryDataSize, &entryData); err != nil {
			return nil, err
		}

		entryName := make([]byte, entryData.NameLength)
		if err := l.BytesToStructV2(int(entryData.NameLength), &entryName); err != nil {
			return nil, err
		}

		entry := ChunkExternalFilesEntry{
			ChunkExternalFilesEntryData: entryData,
			Name:                        string(entryName),
		}
		entries[i] = entry
	}

	return &ChunkExternalFiles{header: ch, ChunkExternalFilesData: externalFilesData, Entries: entries}, nil
}

func (l *Loader) ParseChunkCelExtra(ch ChunkHeader) (Chunk, error) {
	var celExtraData ChunkCelExtraData
	if err := l.BytesToStructV2(ChunkCelExtraDataSize, &celExtraData); err != nil {
		return nil, err
	}

	return &ChunkCelExtra{
		header:            ch,
		ChunkCelExtraData: celExtraData,
	}, nil
}

func (l *Loader) ParseChunkLayer(ch ChunkHeader) (Chunk, error) {
	var layerData ChunkLayerData
	if err := l.BytesToStructV2(ChunkLayerDataSize, &layerData); err != nil {
		return nil, err
	}

	nameBytes := make([]byte, layerData.NameLength)
	if err := l.BytesToStructV2(int(layerData.NameLength), &nameBytes); err != nil {
		return nil, err
	}
	var name ChunkLayerName = ChunkLayerName(string(nameBytes))
	chunk := &ChunkLayer{
		header:                     ch,
		ChunkLayerData:             layerData,
		ChunkLayerName:             name,
		ChunkLayerType2Data:        nil,
		ChunkLayerLockMovementData: nil,
	}
	if layerData.Type == 2 {
		var type2Data ChunkLayerType2Data
		if err := l.BytesToStructV2(4, &type2Data); err != nil {
			return nil, err
		}

		chunk.ChunkLayerType2Data = &type2Data
	}

	var flags ChunkLayerFlags = ChunkLayerFlags{}
	value := reflect.ValueOf(&flags).Elem()
	fieldIndex := 0
	for i := uint16(1); i < 65; i *= 2 {
		field := value.Field(fieldIndex)
		if layerData.FlagsBit&i == i && field.CanSet() {
			field.SetBool(true)
		}
		fieldIndex++
	}

	chunk.ChunkLayerFlags = flags

	if flags.LockMovement {
		var lockData ChunkLayerLockMovementData
		if err := l.BytesToStructV2(16, &lockData); err != nil {
			return nil, err
		}

		chunk.ChunkLayerLockMovementData = &lockData
	}

	return chunk, nil
}

func (l *Loader) ParseChunkOldPalette(ch ChunkHeader) (Chunk, error) {
	var packetsNumber uint16
	if err := l.BytesToStructV2(2, &packetsNumber); err != nil {
		return nil, err
	}

	colors := make([]ChunkOldPaletteColor, 256)
	for i := range packetsNumber {
		var paletteEntriesNumber byte
		if err := l.BytesToStructV2(1, &paletteEntriesNumber); err != nil {
			return nil, err
		}

		var colorsNumber byte
		if err := l.BytesToStructV2(1, &colorsNumber); err != nil {
			return nil, err
		}
		i += uint16(paletteEntriesNumber)
		for range colorsNumber {
			var color ChunkOldPaletteColor
			if err := l.BytesToStructV2(3, &color); err != nil {
				return nil, err
			}

			colors[i] = color
		}
	}

	switch ch.Type {
	case OldPaletteChunkHex:
		return &ChunkOldPalette{
			header: ch,
			Colors: colors,
		}, nil
	case OldPaletteChunk2Hex:
		return &ChunkOldPalette2{
			ChunkOldPalette: &ChunkOldPalette{
				header: ch,
				Colors: colors,
			},
		}, nil
	default:
		// should be unreachable
		return nil, fmt.Errorf("Invalid chunk type for parsing OldPaletteChunk: 0x%X", ch.Type)
	}
}

func BytesToPixelsRGBA(data []byte) PixelsRGBA {
	var chunks [][4]byte
	for i := 0; i < len(data); i += 4 {
		var block [4]byte
		copy(block[:], data[i:i+4])
		chunks = append(chunks, block)
	}
	return chunks
}

func BytesToPixelsGrayscale(data []byte) PixelsGrayscale {
	var chunks [][2]byte
	for i := 0; i < len(data); i += 2 {
		var block [2]byte
		copy(block[:], data[i:i+2])
		chunks = append(chunks, block)
	}
	return chunks
}

func (l *Loader) ResolvePixelType(buf []byte) Pixels {
	var pixels Pixels
	colorDepth := l.File.Header.ColorDepth

	switch colorDepth {
	case ColorDepthRGBA:
		pixels = PixelsRGBA(BytesToPixelsRGBA(buf))
	case ColorDepthGrayscale:
		pixels = PixelsGrayscale(BytesToPixelsGrayscale(buf))
	case ColorDepthIndexed:
		pixels = buf
	default:
		panic("unreachable: colordepth possibly not defined: " + fmt.Sprint(colorDepth))
	}

	return pixels
}

func (l *Loader) GetPixels(ch ChunkHeader, compressed bool, pixelDataSize int) (Pixels, error) {
	var pbuf []byte
	if compressed {
		fmt.Printf("pixel data size compressed: %d\n", pixelDataSize)

		pixelsCompressed := make(PixelsZlib, pixelDataSize)
		if err := l.BytesToStructV2(pixelDataSize, &pixelsCompressed); err != nil {
			return nil, err
		}

		var err error
		pbuf, err = pixelsCompressed.Decompress()
		if err != nil {
			return nil, err
		}
	} else {
		pbuf = make([]byte, pixelDataSize)
		if err := l.BytesToStructV2(pixelDataSize, &pbuf); err != nil {
			return nil, err
		}
	}

	fmt.Printf("pixels decompressed size: %d\n", len(pbuf))

	return l.ResolvePixelType(pbuf), nil
}

func (l *Loader) ParseChunkCel(ch ChunkHeader, frameId int) (Chunk, error) {
	var cData ChunkCelData
	if err := l.BytesToStructV2(ChunkCelDataSize, &cData); err != nil {
		return nil, err
	}

	if cData.CelType == CelTypeLinked {
		var cLinkeData ChunkCelLinkedData
		if err := l.BytesToStructV2(ChunkCelLinkedDataSize, &cLinkeData); err != nil {
			return nil, err
		}
		return &ChunkCelLinked{
			header:             ch,
			ChunkCelData:       cData,
			ChunkCelLinkedData: cLinkeData,
		}, nil
	}

	var dimensions ChunkCelDimensionData
	if err := l.BytesToStructV2(ChunkCelDimensionSize, &dimensions); err != nil {
		return nil, err
	}

	fmt.Printf("width: %d, height: %d\n", dimensions.Width, dimensions.Width)

	pixelDataSize := int(ch.Size - ChunkHeaderSize - ChunkCelDataSize - ChunkCelDimensionSize)
	switch cData.CelType {
	case CelTypeRawImage:
		var pixels Pixels
		var err error
		if pixels, err = l.GetPixels(ch, false, pixelDataSize); err != nil {
			return nil, err
		}
		return &ChunkCelImage{
			header:       ch,
			ChunkCelData: cData,
			ChunkCelRawImageData: ChunkCelRawImageData{
				ChunkCelDimensionData: dimensions,
				Pixels:                pixels,
			},
		}, nil
	case CelTypeCompressedImage:
		var pixels Pixels
		var err error
		if pixels, err = l.GetPixels(ch, true, pixelDataSize); err != nil {
			return nil, err
		}

		// p := pixels.(PixelsRGBA)
		// p.ToImage(int(cData.X), int(cData.Y), int(dimensions.Width), int(dimensions.Height), int(l.File.Header.Width), int(l.File.Header.Height), frameId)

		return &ChunkCelImage{
			header:       ch,
			ChunkCelData: cData,
			ChunkCelRawImageData: ChunkCelRawImageData{
				ChunkCelDimensionData: dimensions,
				Pixels:                pixels,
			},
		}, nil
	case CelTypeCompressedTilemap:
		// TODO: ver dps dado errado @Carto1a
		var ctilemapStatic ChunkCelCompressedTilemapStaticData
		if err := l.BytesToStructV2(ChunkCelCompressedTilemapStaticDataSize, &ctilemapStatic); err != nil {
			return nil, err
		}

		cTilemapData := ChunkCelCompressedTilemapData{
			ChunkCelDimensionData:               dimensions,
			ChunkCelCompressedTilemapStaticData: ctilemapStatic,
		}
		return &ChunkCelTilemap{
			header:                        ch,
			ChunkCelData:                  cData,
			ChunkCelCompressedTilemapData: cTilemapData,
		}, nil
	}

	return nil, nil
}

func (l *Loader) ParseChunkColorProfile(ch ChunkHeader) (Chunk, error) {
	var cData ChunkColorProfileData
	err := l.BytesToStructV2(ChunkColorProfileDataSize, &cData)
	if err != nil {
		return nil, err
	}

	if cData.Type == ColorProfileICC {
		var iccSize uint32
		err = l.BytesToStructV2(8, &iccSize)
		if err != nil {
			return nil, err
		}
		iccData := make([]byte, iccSize)
		err = l.BytesToStructV2(int(iccSize), &iccData)
		if err != nil {
			return nil, err
		}

		chunk := &ChunkColorProfileICC{
			ChunkColorProfile: ChunkColorProfile{
				header:                ch,
				ChunkColorProfileData: cData,
			},
			ChunkColorProfileICCData: ChunkColorProfileICCData{
				DataLength: iccSize,
				Data:       iccData,
			},
		}
		return chunk, nil
	}
	chunk := &ChunkColorProfile{
		header:                ch,
		ChunkColorProfileData: cData,
	}
	return chunk, nil
}

func (l *Loader) ParseChunkPalette(ch ChunkHeader) (Chunk, error) {
	var cData ChunkPaletteData
	if err := l.BytesToStructV2(ChunkPaletteDataSize, &cData); err != nil {
		return nil, err
	}

	entries := make([]ChunkPaletteEntry, 0)
	for range cData.To - cData.From + 1 {
		entry := ChunkPaletteEntry{}
		var entryData ChunkPaletteEntryData
		if err := l.BytesToStructV2(ChunkPaletteEntryDataSize, &entryData); err != nil {
			return nil, err
		}

		if entryData.HasName == 1 {
			var nameLen uint16
			if err := l.BytesToStructV2(2, &nameLen); err != nil {
				return nil, err
			}

			colorName := make([]byte, nameLen)
			if err := l.BytesToStructV2(len(colorName), &colorName); err != nil {
				return nil, err
			}

			entry.ColorName = string(colorName)
		}

		entries = append(entries, entry)
	}

	return &ChunkPalette{
		header:           ch,
		ChunkPaletteData: cData,
		Entries:          entries,
	}, nil
}

func (l *Loader) ParseChunkTag(ch ChunkHeader) (Chunk, error) {
	var cData ChunkTagData
	if err := l.BytesToStructV2(ChunkTagDataSize, &cData); err != nil {
		return nil, err
	}

	entries := make([]ChunkTagEntry, cData.NumberTags)
	for i := range cData.NumberTags {
		var entryData ChunkTagEntryData
		if err := l.BytesToStructV2(ChunkTagEntryDataSize, &entryData); err != nil {
			return nil, err
		}

		tagName := make([]byte, entryData.TagNameSize)
		if err := l.BytesToStructV2(len(tagName), &tagName); err != nil {
			return nil, err
		}

		entry := ChunkTagEntry{
			ChunkTagEntryData: entryData,
			Name:              string(tagName),
		}

		entries[i] = entry
	}

	return &ChunkTag{
		header:  ch,
		Entries: entries,
	}, nil
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

func DeserializeFile(fd *os.File) (*AsepriteFile, error) {
	var ase *AsepriteFile = new(AsepriteFile)
	loader := new(Loader)

	reader := fd

	loader.Buf = make([]byte, ChunkSize)
	loader.Buffer = new(bytes.Buffer)
	loader.Reader = reader
	loader.File = ase

	header, err := loader.ParseHeader()
	if err != nil {
		return nil, err
	}
	ase.Header = header
	loader.Buffer.Reset()
	frames, err := loader.ParseFrames(&header)
	if err != nil {
		return nil, err
	}
	ase.Frames = frames

	return ase, nil
}


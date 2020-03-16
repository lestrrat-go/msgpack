package msgpack

import (
	"bufio"
	"io"
	"reflect"
	"sync"
	"time"
)

const MaxPositiveFixNum = byte(0x7f)
const MinNegativeFixNum = byte(0xe0)

// Code represents the first by in a msgpack element. It tell us
// the data layout that follows it
type Code byte

const (
	InvalidCode    Code = 0
	FixMap0        Code = 0x80
	FixMap1        Code = 0x81
	FixMap2        Code = 0x82
	FixMap3        Code = 0x83
	FixMap4        Code = 0x84
	FixMap5        Code = 0x85
	FixMap6        Code = 0x86
	FixMap7        Code = 0x87
	FixMap8        Code = 0x88
	FixMap9        Code = 0x89
	FixMap10       Code = 0x8a
	FixMap11       Code = 0x8b
	FixMap12       Code = 0x8c
	FixMap13       Code = 0x8d
	FixMap14       Code = 0x8e
	FixMap15       Code = 0x8f
	FixArray0      Code = 0x90
	FixArray1      Code = 0x91
	FixArray2      Code = 0x92
	FixArray3      Code = 0x93
	FixArray4      Code = 0x94
	FixArray5      Code = 0x95
	FixArray6      Code = 0x96
	FixArray7      Code = 0x97
	FixArray8      Code = 0x98
	FixArray9      Code = 0x99
	FixArray10     Code = 0x9a
	FixArray11     Code = 0x9b
	FixArray12     Code = 0x9c
	FixArray13     Code = 0x9d
	FixArray14     Code = 0x9e
	FixArray15     Code = 0x9f
	NegFixedNumLow Code = 0xe0
	FixStr0        Code = 0xa0
	FixStr1        Code = 0xa1
	FixStr2        Code = 0xa2
	FixStr3        Code = 0xa3
	FixStr4        Code = 0xa4
	FixStr5        Code = 0xa5
	FixStr6        Code = 0xa6
	FixStr7        Code = 0xa7
	FixStr8        Code = 0xa8
	FixStr9        Code = 0xa9
	FixStr10       Code = 0xaa
	FixStr11       Code = 0xab
	FixStr12       Code = 0xac
	FixStr13       Code = 0xad
	FixStr14       Code = 0xae
	FixStr15       Code = 0xaf
	FixStr16       Code = 0xb0
	FixStr17       Code = 0xb1
	FixStr18       Code = 0xb2
	FixStr19       Code = 0xb3
	FixStr20       Code = 0xb4
	FixStr21       Code = 0xb5
	FixStr22       Code = 0xb6
	FixStr23       Code = 0xb7
	FixStr24       Code = 0xb8
	FixStr25       Code = 0xb9
	FixStr26       Code = 0xba
	FixStr27       Code = 0xbb
	FixStr28       Code = 0xbc
	FixStr29       Code = 0xbd
	FixStr30       Code = 0xbe
	FixStr31       Code = 0xbf
	Nil            Code = 0xc0
	False          Code = 0xc2
	True           Code = 0xc3
	Bin8           Code = 0xc4
	Bin16          Code = 0xc5
	Bin32          Code = 0xc6
	Ext8           Code = 0xc7
	Ext16          Code = 0xc8
	Ext32          Code = 0xc9
	Float          Code = 0xca
	Double         Code = 0xcb
	Uint8          Code = 0xcc
	Uint16         Code = 0xcd
	Uint32         Code = 0xce
	Uint64         Code = 0xcf
	Int8           Code = 0xd0
	Int16          Code = 0xd1
	Int32          Code = 0xd2
	Int64          Code = 0xd3
	FixExt1        Code = 0xd4
	FixExt2        Code = 0xd5
	FixExt4        Code = 0xd6
	FixExt8        Code = 0xd7
	FixExt16       Code = 0xd8
	Str8           Code = 0xd9
	Str16          Code = 0xda
	Str32          Code = 0xdb
	Array16        Code = 0xdc
	Array32        Code = 0xdd
	Map16          Code = 0xde
	Map32          Code = 0xdf
	FixedArrayMask Code = 0xf
)

type InvalidDecodeError struct {
	Type reflect.Type
}

// EncodeMsgpacker is an interface for those objects that provide
// their own serialization. The objects are responsible for providing
// the complete msgpack payload, including the code, payload length
// (if applicable), and payload (if applicable)
//
// The Encoder instance that is passed is NOT safe to be used
// concurrently with other goroutines. DO NOT share it between multiple
// goroutines, and DO NOT attempt to use it once you return from
// your DecodeMsgpack method.
type EncodeMsgpacker interface {
	EncodeMsgpack(Encoder) error
}

// DecodeMsgpacker is an interface for those objects that provide
// their own deserialization. The objects are responsible for handling
// the code, payload length (if applicable), and payload (if applicable)
//
// The Decoder instance that is passed is NOT safe to be used
// concurrently with other goroutines. DO NOT share it between multiple
// goroutines, and DO NOT attempt to use it once you return from
// your DecodeMsgpack method.
type DecodeMsgpacker interface {
	DecodeMsgpack(Decoder) error
}

// ArrayBuilder is used to build a msgpack array
type ArrayBuilder interface {
	Add(interface{})
	Bytes() ([]byte, error)
	Count() int
	Encode(io.Writer) error
	Reset()
}

// MapBuilder is used to build a msgpack map
type MapBuilder interface {
	Add(string, interface{})
	Bytes() ([]byte, error)
	Count() int
	Encode(io.Writer) error
	Reset()
}

// Writer handles low-level writing to an io.Writer.
// Note that Writers are NEVER meant to be shared concurrently
// between goroutines. You DO NOT write serialized data concurrently
// to the same destination.
type Writer interface {
	io.Writer
	WriteByte(byte) error
	WriteByteUint8(byte, uint8) error
	WriteByteUint16(byte, uint16) error
	WriteByteUint32(byte, uint32) error
	WriteByteUint64(byte, uint64) error
	WriteString(string) (int, error)
	WriteUint8(uint8) error
	WriteUint16(uint16) error
	WriteUint32(uint32) error
	WriteUint64(uint64) error
}

// Reader handles low-level reading from an io.Reader.
// Note that Readers are NEVER meant to be shared concurrently
// between goroutines. You DO NOT read data concurrently
// from the same serialized source.
type Reader interface {
	io.Reader
	ReadByte() (byte, error)
	ReadUint8() (uint8, error)
	ReadUint16() (uint16, error)
	ReadUint32() (uint32, error)
	ReadUint64() (uint64, error)
	ReadByteUint8() (byte, uint8, error)
	ReadByteUint16() (byte, uint16, error)
	ReadByteUint32() (byte, uint32, error)
	ReadByteUint64() (byte, uint64, error)
}

// Encoder writes serialized data to a destination pointed to by
// an io.Writer
type Encoder interface {
	Encode(interface{}) error
	EncodeArrayHeader(int) error
	EncodeArray(interface{}) error
	EncodeBool(bool) error
	EncodeBytes([]byte) error
	EncodeExt(EncodeMsgpacker) error
	EncodeExtHeader(int) error
	EncodeExtType(EncodeMsgpacker) error
	EncodeMap(interface{}) error
	EncodeNegativeFixNum(int8) error
	EncodeNil() error
	EncodePositiveFixNum(uint8) error
	EncodeInt(int) error
	EncodeInt8(int8) error
	EncodeInt16(int16) error
	EncodeInt32(int32) error
	EncodeInt64(int64) error
	EncodeUint(uint) error
	EncodeUint8(uint8) error
	EncodeUint16(uint16) error
	EncodeUint32(uint32) error
	EncodeUint64(uint64) error
	EncodeFloat32(float32) error
	EncodeFloat64(float64) error
	EncodeString(string) error
	EncodeStruct(interface{}) error
	EncodeTime(time.Time) error
	Writer() Writer

	// SetDestination is a utility tool that allows the user to swap out the
	// reader object the Encoder is writing to, there by saving the
	// extra cost of re-instantiaion.
	SetDestination(io.Writer)
}

type encoder struct {
	nl *encoderNL
	mu sync.RWMutex
}

type encoderNL struct {
	dst Writer
}

// Decoder reads serialized data from a source pointed to by
// an io.Reader, and meterializes the data structure
type Decoder interface {
	// Decode takes a pointer to a variable, and populates it with the value
	// that was unmarshaled from the stream.
	//
	// If the variable is a non-pointer or nil, an error is returned.
	Decode(interface{}) error
	DecodeArray(interface{}) error
	DecodeArrayLength(*int) error
	DecodeBool(b *bool) error
	DecodeBytes(*[]byte) error
	DecodeExt(DecodeMsgpacker) error
	DecodeExtLength(*int) error
	DecodeExtType(*reflect.Type) error
	DecodeFloat32(*float32) error
	DecodeFloat64(*float64) error
	DecodeInt(*int) error
	DecodeInt8(*int8) error
	DecodeInt16(*int16) error
	DecodeInt32(*int32) error
	DecodeInt64(*int64) error
	DecodeMap(*map[string]interface{}) error
	DecodeMapLength(*int) error
	DecodeNil(*interface{}) error
	DecodeString(*string) error
	DecodeStruct(interface{}) error
	DecodeTime(*time.Time) error
	DecodeUint(*uint) error
	DecodeUint8(*uint8) error
	DecodeUint16(*uint16) error
	DecodeUint32(*uint32) error
	DecodeUint64(*uint64) error
	PeekCode() (Code, error)
	ReadCode() (Code, error)
	Reader() Reader

	// SetSource is a utility tool that allows the user to swap out the
	// reader object the Decoder is reading from, there by saving the
	// extra cost of re-instantiaion.
	SetSource(io.Reader)
}

type decoder struct {
	nl *decoderNL
	mu sync.RWMutex
}

// decoderNL is a type of decoder that does no locking of
// the member fields when public methods are called.
// in contrast, calling public methods for the equivalent
// locking version will force other method calls to wait while
// an operation is progressing on the object.
type decoderNL struct {
	raw *bufio.Reader
	src Reader
}

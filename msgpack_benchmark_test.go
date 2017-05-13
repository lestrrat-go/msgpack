// +build bench

package msgpack_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"testing"

	lestrrat "github.com/lestrrat/go-msgpack"
	vmihailenco "gopkg.in/vmihailenco/msgpack.v2"
)

type Decoder interface {
	Decode(interface{}) error
}

type DecodeUinter interface {
	DecodeUint(*uint) error
}

type DecodeUint8er interface {
	DecodeUint8(*uint8) error
}

type DecodeUint8Returner interface {
	DecodeUint8() (uint8, error)
}

type DecodeUint16er interface {
	DecodeUint16(*uint16) error
}

type DecodeUint16Returner interface {
	DecodeUint16() (uint16, error)
}

type DecodeUint32er interface {
	DecodeUint32(*uint32) error
}

type DecodeUint32Returner interface {
	DecodeUint32() (uint32, error)
}
type DecodeUint64er interface {
	DecodeUint64(*uint64) error
}

type DecodeUint64Returner interface {
	DecodeUint64() (uint64, error)
}

type Encoder interface {
	Encode(interface{}) error
}

type EncodeStringer interface {
	EncodeString(string) error
}

type EncodeUinter interface {
	EncodeUint(uint) error
}

type EncodeUint8er interface {
	EncodeUint8(uint8) error
}

type EncodeUint16er interface {
	EncodeUint16(uint16) error
}

type EncodeUint32er interface {
	EncodeUint32(uint32) error
}

type EncodeUint64er interface {
	EncodeUint64(uint64) error
}

type EncodeFloat32er interface {
	EncodeFloat32(float32) error
}

type EncodeFloat64er interface {
	EncodeFloat64(float64) error
}

type EncodeInter interface {
	EncodeInt(int) error
}

type EncodeInt8er interface {
	EncodeInt8(int8) error
}

type EncodeInt16er interface {
	EncodeInt16(int16) error
}

type EncodeInt32er interface {
	EncodeInt32(int32) error
}

type EncodeInt64er interface {
	EncodeInt64(int64) error
}

// Oh why, why did you need to declare your Decode with variadic
// input list?
type VmihailencoDecoder struct {
	*vmihailenco.Decoder
}

func (e VmihailencoDecoder) Decode(v interface{}) error {
	return e.Decoder.Decode(v)
}

// Oh why, why did you need to declare your Encode with variadic
// input list?
type VmihailencoEncoder struct {
	*vmihailenco.Encoder
}

func (e VmihailencoEncoder) Encode(v interface{}) error {
	return e.Encoder.Encode(v)
}

var encoders = []struct {
	Name        string
	Encoder     interface{}
	MakeDecoder func(io.Reader) interface{}
}{
	{
		Name:    "___lestrrat",
		Encoder: lestrrat.NewEncoder(ioutil.Discard),
		MakeDecoder: func(r io.Reader) interface{} {
			return lestrrat.NewDecoder(r)
		},
	},
	{
		Name:    "vmihailenco",
		Encoder: VmihailencoEncoder{Encoder: vmihailenco.NewEncoder(ioutil.Discard)},
		MakeDecoder: func(r io.Reader) interface{} {
			return VmihailencoDecoder{Decoder: vmihailenco.NewDecoder(r)}
		},
	},
}

func BenchmarkEncodeFloat32(b *testing.B) {
	for _, data := range encoders {
		if enc, ok := data.Encoder.(Encoder); ok {
			b.Run(fmt.Sprintf("%s/encode float32 via Encode()", data.Name), func(b *testing.B) {
				var v float32 = math.MaxFloat32
				for i := 0; i < b.N; i++ {
					if err := enc.Encode(v); err != nil {
						panic(err)
					}
				}
			})
		}
		if enc, ok := data.Encoder.(EncodeFloat32er); ok {
			b.Run(fmt.Sprintf("%s/encode float32 via EncodeFloat32()", data.Name), func(b *testing.B) {
				var v float32 = math.MaxFloat32
				for i := 0; i < b.N; i++ {
					if err := enc.EncodeFloat32(v); err != nil {
						panic(err)
					}
				}
			})
		}
	}
}

func BenchmarkEncodeFloat64(b *testing.B) {
	for _, data := range encoders {
		if enc, ok := data.Encoder.(Encoder); ok {
			b.Run(fmt.Sprintf("%s/encode float64 via Encode()", data.Name), func(b *testing.B) {
				var v float64 = math.MaxFloat64
				for i := 0; i < b.N; i++ {
					if err := enc.Encode(v); err != nil {
						panic(err)
					}
				}
			})
		}
		if enc, ok := data.Encoder.(EncodeFloat64er); ok {
			b.Run(fmt.Sprintf("%s/encode float64 via EncodeFloat64()", data.Name), func(b *testing.B) {
				var v float64 = math.MaxFloat64
				for i := 0; i < b.N; i++ {
					if err := enc.EncodeFloat64(v); err != nil {
						panic(err)
					}
				}
			})
		}
	}
}

func BenchmarkEncodeUint8(b *testing.B) {
	for _, data := range encoders {
		if enc, ok := data.Encoder.(Encoder); ok {
			b.Run(fmt.Sprintf("%s/encode uint8 via Encode()", data.Name), func(b *testing.B) {
				var v uint8 = math.MaxUint8
				for i := 0; i < b.N; i++ {
					if err := enc.Encode(v); err != nil {
						panic(err)
					}
				}
			})
		}
		if enc, ok := data.Encoder.(EncodeUint8er); ok {
			b.Run(fmt.Sprintf("%s/encode uint8 via EncodeUint8()", data.Name), func(b *testing.B) {
				var v uint8 = math.MaxUint8
				for i := 0; i < b.N; i++ {
					if err := enc.EncodeUint8(v); err != nil {
						panic(err)
					}
				}
			})
		}
	}
}

func BenchmarkEncodeUint16(b *testing.B) {
	for _, data := range encoders {
		if enc, ok := data.Encoder.(Encoder); ok {
			b.Run(fmt.Sprintf("%s/encode uint16 via Encode()", data.Name), func(b *testing.B) {
				var v uint16 = math.MaxUint16
				for i := 0; i < b.N; i++ {
					if err := enc.Encode(v); err != nil {
						panic(err)
					}
				}
			})
		}
		if enc, ok := data.Encoder.(EncodeUint16er); ok {
			b.Run(fmt.Sprintf("%s/encode uint16 via EncodeUint16()", data.Name), func(b *testing.B) {
				var v uint16 = math.MaxUint16
				for i := 0; i < b.N; i++ {
					if err := enc.EncodeUint16(v); err != nil {
						panic(err)
					}
				}
			})
		}
	}
}

func BenchmarkEncodeUint32(b *testing.B) {
	for _, data := range encoders {
		if enc, ok := data.Encoder.(Encoder); ok {
			b.Run(fmt.Sprintf("%s/encode uint32 via Encode()", data.Name), func(b *testing.B) {
				var v uint32 = math.MaxUint32
				for i := 0; i < b.N; i++ {
					if err := enc.Encode(v); err != nil {
						panic(err)
					}
				}
			})
		}
		if enc, ok := data.Encoder.(EncodeUint32er); ok {
			b.Run(fmt.Sprintf("%s/encode uint32 via EncodeUint32()", data.Name), func(b *testing.B) {
				var v uint32 = math.MaxUint32
				for i := 0; i < b.N; i++ {
					if err := enc.EncodeUint32(v); err != nil {
						panic(err)
					}
				}
			})
		}
	}
}

func BenchmarkEncodeUint64(b *testing.B) {
	for _, data := range encoders {
		if enc, ok := data.Encoder.(Encoder); ok {
			b.Run(fmt.Sprintf("%s/encode uint64 via Encode()", data.Name), func(b *testing.B) {
				var v uint64 = math.MaxUint64
				for i := 0; i < b.N; i++ {
					if err := enc.Encode(v); err != nil {
						panic(err)
					}
				}
			})
		}
		if enc, ok := data.Encoder.(EncodeUint64er); ok {
			b.Run(fmt.Sprintf("%s/encode uint64 via EncodeUint64()", data.Name), func(b *testing.B) {
				var v uint64 = math.MaxUint64
				for i := 0; i < b.N; i++ {
					if err := enc.EncodeUint64(v); err != nil {
						panic(err)
					}
				}
			})
		}
	}
}

func BenchmarkEncodeInt8(b *testing.B) {
	for _, data := range encoders {
		if enc, ok := data.Encoder.(Encoder); ok {
			b.Run(fmt.Sprintf("%s/encode int8 via Encode()", data.Name), func(b *testing.B) {
				var v int8 = math.MaxInt8
				for i := 0; i < b.N; i++ {
					if err := enc.Encode(v); err != nil {
						panic(err)
					}
				}
			})
		}
		if enc, ok := data.Encoder.(EncodeInt8er); ok {
			b.Run(fmt.Sprintf("%s/encode int8 via EncodeInt8()", data.Name), func(b *testing.B) {
				var v int8 = math.MaxInt8
				for i := 0; i < b.N; i++ {
					if err := enc.EncodeInt8(v); err != nil {
						panic(err)
					}
				}
			})
		}
	}
}

func BenchmarkEncodeInt16(b *testing.B) {
	for _, data := range encoders {
		if enc, ok := data.Encoder.(Encoder); ok {
			b.Run(fmt.Sprintf("%s/encode int16 via Encode()", data.Name), func(b *testing.B) {
				var v int16 = math.MaxInt16
				for i := 0; i < b.N; i++ {
					if err := enc.Encode(v); err != nil {
						panic(err)
					}
				}
			})
		}
		if enc, ok := data.Encoder.(EncodeInt16er); ok {
			b.Run(fmt.Sprintf("%s/encode int16 via EncodeInt16()", data.Name), func(b *testing.B) {
				var v int16 = math.MaxInt16
				for i := 0; i < b.N; i++ {
					if err := enc.EncodeInt16(v); err != nil {
						panic(err)
					}
				}
			})
		}
	}
}

func BenchmarkEncodeInt32(b *testing.B) {
	for _, data := range encoders {
		if enc, ok := data.Encoder.(Encoder); ok {
			b.Run(fmt.Sprintf("%s/encode int32 via Encode()", data.Name), func(b *testing.B) {
				var v int32 = math.MaxInt32
				for i := 0; i < b.N; i++ {
					if err := enc.Encode(v); err != nil {
						panic(err)
					}
				}
			})
		}
		if enc, ok := data.Encoder.(EncodeInt32er); ok {
			b.Run(fmt.Sprintf("%s/encode int32 via EncodeInt32()", data.Name), func(b *testing.B) {
				var v int32 = math.MaxInt32
				for i := 0; i < b.N; i++ {
					if err := enc.EncodeInt32(v); err != nil {
						panic(err)
					}
				}
			})
		}
	}
}

func BenchmarkEncodeInt64(b *testing.B) {
	for _, data := range encoders {
		if enc, ok := data.Encoder.(Encoder); ok {
			b.Run(fmt.Sprintf("%s/encode int64 via Encode()", data.Name), func(b *testing.B) {
				var v int64 = math.MaxInt64
				for i := 0; i < b.N; i++ {
					if err := enc.Encode(v); err != nil {
						panic(err)
					}
				}
			})
		}
		if enc, ok := data.Encoder.(EncodeInt64er); ok {
			b.Run(fmt.Sprintf("%s/encode int64 via EncodeInt64()", data.Name), func(b *testing.B) {
				var v int64 = math.MaxInt64
				for i := 0; i < b.N; i++ {
					if err := enc.EncodeInt64(v); err != nil {
						panic(err)
					}
				}
			})
		}
	}
}

func BenchmarkEncodeString(b *testing.B) {
	for _, data := range encoders {
		for _, size := range []int{math.MaxUint8, math.MaxUint8 + 1, math.MaxUint16 + 1} {
			s := makeString(size)
			if enc, ok := data.Encoder.(Encoder); ok {
				b.Run(fmt.Sprintf("%s/encode string (%d bytes) via Encode()", data.Name, size), func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						if err := enc.Encode(s); err != nil {
							panic(err)
						}
					}
				})
			}
			if enc, ok := data.Encoder.(EncodeStringer); ok {
				b.Run(fmt.Sprintf("%s/encode string (%d bytes) via EncodeString()", data.Name, size), func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						if err := enc.EncodeString(s); err != nil {
							panic(err)
						}
					}
				})
			}
		}
	}
}

func BenchmarkDecodeUint8(b *testing.B) {
	for _, data := range encoders {
		canary := data.MakeDecoder(&bytes.Buffer{})
		serialized := []byte{lestrrat.Uint8.Byte(), byte(math.MaxUint8)}
		rdr := bytes.NewReader(serialized)
		if _, ok := canary.(DecodeUint8er); ok {
			b.Run(fmt.Sprintf("%s/decode uint8 via DecodeUint8()", data.Name), func(b *testing.B) {
				var v uint8
				dec := data.MakeDecoder(rdr).(DecodeUint8er)
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					rdr.Seek(0, 0)
					b.StartTimer()
					if err := dec.DecodeUint8(&v); err != nil {
						panic(err)
					}
					if v != math.MaxUint8 {
						panic("v should be math.MaxUint :/")
					}
				}
			})
		} else if _, ok := canary.(DecodeUint8Returner); ok {
			b.Run(fmt.Sprintf("%s/decode uint8 via DecodeUint8() (return)", data.Name), func(b *testing.B) {
				dec := data.MakeDecoder(rdr).(DecodeUint8Returner)
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					rdr.Seek(0, 0)
					b.StartTimer()
					v, err := dec.DecodeUint8()
					if err != nil {
						panic(err)
					}
					if v != math.MaxUint8 {
						panic("v should be math.MaxUint :/")
					}
				}
			})
		}
	}
}

func BenchmarkDecodeUint16(b *testing.B) {
	for _, data := range encoders {
		canary := data.MakeDecoder(&bytes.Buffer{})
		serialized := make([]byte, 3)
		serialized[0] = lestrrat.Uint16.Byte()
		binary.BigEndian.PutUint16(serialized[1:], math.MaxUint16)
		rdr := bytes.NewReader(serialized)
		if _, ok := canary.(DecodeUint16er); ok {
			b.Run(fmt.Sprintf("%s/decode uint16 via DecodeUint16()", data.Name), func(b *testing.B) {
				var v uint16
				dec := data.MakeDecoder(rdr).(DecodeUint16er)
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					rdr.Seek(0, 0)
					b.StartTimer()
					if err := dec.DecodeUint16(&v); err != nil {
						panic(err)
					}
					if v != math.MaxUint16 {
						panic("v should be math.MaxUint :/")
					}
				}
			})
		} else if _, ok := canary.(DecodeUint16Returner); ok {
			b.Run(fmt.Sprintf("%s/decode uint16 via DecodeUint16() (return)", data.Name), func(b *testing.B) {
				dec := data.MakeDecoder(rdr).(DecodeUint16Returner)
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					rdr.Seek(0, 0)
					b.StartTimer()
					v, err := dec.DecodeUint16()
					if err != nil {
						panic(err)
					}
					if v != math.MaxUint16 {
						panic("v should be math.MaxUint :/")
					}
				}
			})
		}
	}
}

func BenchmarkDecodeUint32(b *testing.B) {
	for _, data := range encoders {
		canary := data.MakeDecoder(&bytes.Buffer{})
		serialized := make([]byte, 5)
		serialized[0] = lestrrat.Uint32.Byte()
		binary.BigEndian.PutUint32(serialized[1:], math.MaxUint32)
		rdr := bytes.NewReader(serialized)
		if _, ok := canary.(DecodeUint32er); ok {
			b.Run(fmt.Sprintf("%s/decode uint32 via DecodeUint32()", data.Name), func(b *testing.B) {
				var v uint32
				dec := data.MakeDecoder(rdr).(DecodeUint32er)
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					rdr.Seek(0, 0)
					b.StartTimer()
					if err := dec.DecodeUint32(&v); err != nil {
						panic(err)
					}
					if v != math.MaxUint32 {
						panic("v should be math.MaxUint :/")
					}
				}
			})
		} else if _, ok := canary.(DecodeUint32Returner); ok {
			b.Run(fmt.Sprintf("%s/decode uint32 via DecodeUint32() (return)", data.Name), func(b *testing.B) {
				dec := data.MakeDecoder(rdr).(DecodeUint32Returner)
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					rdr.Seek(0, 0)
					b.StartTimer()
					v, err := dec.DecodeUint32()
					if err != nil {
						panic(err)
					}
					if v != math.MaxUint32 {
						panic("v should be math.MaxUint :/")
					}
				}
			})
		}
	}
}

func BenchmarkDecodeUint64(b *testing.B) {
	for _, data := range encoders {
		serialized := make([]byte, 9)
		serialized[0] = lestrrat.Uint64.Byte()
		binary.BigEndian.PutUint64(serialized[1:], math.MaxUint64)
		rdr := bytes.NewReader(serialized)
		canary := data.MakeDecoder(rdr)

		switch dec := canary.(type) {
		case DecodeUint64er:
			b.Run(fmt.Sprintf("%s/decode uint64 via DecodeUint64()", data.Name), func(b *testing.B) {
				var v uint64
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					rdr.Seek(0, 0)
					b.StartTimer()
					if err := dec.DecodeUint64(&v); err != nil {
						panic(err)
					}
					if v != math.MaxUint64 {
						panic("v should be math.MaxUint :/")
					}
				}
			})
		case DecodeUint64Returner:
			b.Run(fmt.Sprintf("%s/decode uint64 via DecodeUint64() (return)", data.Name), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					rdr.Seek(0, 0)
					b.StartTimer()
					v, err := dec.DecodeUint64()
					if err != nil {
						panic(err)
					}
					if v != math.MaxUint64 {
						panic("v should be math.MaxUint :/")
					}
				}
			})
		default:
			panic("couldn't figure out type")
		}
	}
}

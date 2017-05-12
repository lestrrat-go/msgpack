// +build bench

package msgpack_test

import (
	"bytes"
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

type DecodeUint32er interface {
	DecodeUint32(*uint32) error
}

type DecodeUint64er interface {
	DecodeUint64(*uint64) error
}

type Encoder interface {
	Encode(interface{}) error
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

func BenchmarkInt(b *testing.B) {
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

		canary := data.MakeDecoder(&bytes.Buffer{})
		serialized := []byte{lestrrat.Uint8.Byte(), byte(math.MaxUint8)}
		if _, ok := canary.(DecodeUint8er); ok {
			b.Run(fmt.Sprintf("%s/decode uint8 via DecodeUint8()", data.Name), func(b *testing.B) {
				var v uint8
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					dec := data.MakeDecoder(bytes.NewReader(serialized)).(DecodeUint8er)
					b.StartTimer()
					if err := dec.DecodeUint8(&v); err != nil {
						panic(err)
					}
					b.StopTimer()
					if v != math.MaxUint8 {
						panic("v should be math.MaxUint :/")
					}
					b.StartTimer()
				}
			})
		} else if _, ok := canary.(DecodeUint8Returner); ok {
			b.Run(fmt.Sprintf("%s/decode uint8 via DecodeUint8() (return)", data.Name), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					dec := data.MakeDecoder(bytes.NewReader(serialized)).(DecodeUint8Returner)
					b.StartTimer()
					v, err := dec.DecodeUint8()
					if err != nil {
						panic(err)
					}
					b.StopTimer()
					if v != math.MaxUint8 {
						panic("v should be math.MaxUint :/")
					}
					b.StartTimer()
				}
			})
		}
	}
}

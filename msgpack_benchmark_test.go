// +build bench

package msgpack_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"testing"

	lestrrat "github.com/lestrrat/go-msgpack"
	"github.com/pkg/errors"
	vmihailenco "gopkg.in/vmihailenco/msgpack.v2"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func randString(l int) string {
	buf := make([]byte, l)
	for i := 0; i < (l+1)/2; i++ {
		buf[i] = byte(rand.Intn(256))
	}
	return fmt.Sprintf("%x", buf)[:l]
}

type Encoder interface {
	Encode(interface{}) error
}

type Decoder interface {
	Decode(interface{}) error
}

type Marshaler interface {
	Marshal(interface{}) ([]byte, error)
}

type Unmarshaler interface {
	Unmarshal([]byte, interface{}) error
}

type MarshalFunc func(interface{}) ([]byte, error)

func (f MarshalFunc) Marshal(v interface{}) ([]byte, error) {
	return f(v)
}

var strvar16 = randString(16)
var strvar256 = randString(256)
var strvar65536 = randString(65536)

func BenchmarkLestrrat(b *testing.B) {
	b.StopTimer()
	e := lestrrat.NewEncoder(ioutil.Discard)
	b.StartTimer()
	b.ReportAllocs()
	benchEncodeDecode(b, e, func(in io.Reader) Decoder { return lestrrat.NewDecoder(in) })
	benchMarshalUnmarshal(b, MarshalFunc(lestrrat.Marshal))
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

func BenchmarkVmihailenco(b *testing.B) {
	b.StopTimer()
	e := VmihailencoEncoder{Encoder: vmihailenco.NewEncoder(ioutil.Discard)}
	b.ReportAllocs()
	b.StartTimer()
	benchEncodeDecode(b, e, func(in io.Reader) Decoder {
		return VmihailencoDecoder{Decoder: vmihailenco.NewDecoder(in)}
	})
	benchMarshalUnmarshal(b, MarshalFunc(func(v interface{}) ([]byte, error) {
		return vmihailenco.Marshal(v)
	}))
}

func benchMarshalUnmarshal(b *testing.B, m Marshaler) {
	benchFloatMarshal(b, m)
	benchIntMarshal(b, m)
}

func benchMarshal(b *testing.B, m Marshaler, v interface{}) {
	for i := 0; i < b.N; i++ {
		if _, err := m.Marshal(v); err != nil {
			panic(err)
		}
	}
}

func benchEncodeDecode(b *testing.B, e Encoder, newDecoder func(io.Reader) Decoder) {
	benchNilEncode(b, e)
	benchBool(b, e)
	benchStrings(b, e)
	benchFloatEncode(b, e)
	benchIntEncode(b, e)

	benchNilDecode(b, newDecoder)
}

func benchEncode(b *testing.B, e Encoder, v interface{}) {
	for i := 0; i < b.N; i++ {
		if err := e.Encode(v); err != nil {
			panic(err)
		}
	}
}

func benchDecode(b *testing.B, makeDecoder func() Decoder) {
	var v interface{}
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		d := makeDecoder()
		b.StartTimer()
		if err := d.Decode(&v); err != nil {
			if err, ok := err.(stackTracer); ok {
				for _, f := range err.StackTrace() {
					fmt.Printf("%v\n", f)
				}
			}
			panic(err)
		}
	}
}

func benchNilEncode(b *testing.B, e Encoder) {
	b.Run("encode nil", func(b *testing.B) {
		benchEncode(b, e, nil)
	})
}

func benchNilDecode(b *testing.B, newDecoder func(io.Reader) Decoder) {
	data := []byte{lestrrat.Nil.Byte()}
	b.Run("decode nil", func(b *testing.B) {
		benchDecode(b, func() Decoder {
			return newDecoder(bytes.NewBuffer(data))
		})
	})
}

func benchBool(b *testing.B, e Encoder) {
	b.Run("encode true", func(b *testing.B) {
		benchEncode(b, e, true)
	})
	b.Run("encode false", func(b *testing.B) {
		benchEncode(b, e, false)
	})
}

func benchStrings(b *testing.B, e Encoder) {
	for _, v := range []string{strvar16, strvar256, strvar65536} {
		b.Run(fmt.Sprintf("encode string (len=%d)", len(v)), func(b *testing.B) {
			benchEncode(b, e, v)
		})
	}
}

func benchFloatEncode(b *testing.B, e Encoder) {
	b.Run("encode float32", func(b *testing.B) {
		benchEncode(b, e, math.MaxFloat32)
	})
	b.Run("encode float64", func(b *testing.B) {
		benchEncode(b, e, math.MaxFloat64)
	})
}

func benchFloatMarshal(b *testing.B, m Marshaler) {
	b.Run("marshal float32", func(b *testing.B) {
		benchMarshal(b, m, math.MaxFloat32)
	})
	b.Run("marshal float64", func(b *testing.B) {
		benchMarshal(b, m, math.MaxFloat64)
	})
}

func benchIntEncode(b *testing.B, e Encoder) {
	b.Run("encode uint8", func(b *testing.B) {
		benchEncode(b, e, math.MaxUint8)
	})
	b.Run("encode uint16", func(b *testing.B) {
		benchEncode(b, e, math.MaxUint16)
	})
	b.Run("encode uint32", func(b *testing.B) {
		benchEncode(b, e, math.MaxUint32)
	})
	b.Run("encode uint64", func(b *testing.B) {
		benchEncode(b, e, uint64(math.MaxUint64))
	})
	b.Run("encode int8", func(b *testing.B) {
		benchEncode(b, e, math.MaxInt8)
	})
	b.Run("encode int16", func(b *testing.B) {
		benchEncode(b, e, math.MaxInt16)
	})
	b.Run("encode int32", func(b *testing.B) {
		benchEncode(b, e, math.MaxInt32)
	})
	b.Run("encode int64", func(b *testing.B) {
		benchEncode(b, e, math.MaxInt64)
	})
}

func benchIntMarshal(b *testing.B, e Marshaler) {
	b.Run("marshal uint8", func(b *testing.B) {
		benchMarshal(b, e, math.MaxUint8)
	})
	b.Run("marshal uint16", func(b *testing.B) {
		benchMarshal(b, e, math.MaxUint16)
	})
	b.Run("marshal uint32", func(b *testing.B) {
		benchMarshal(b, e, math.MaxUint32)
	})
	b.Run("marshal uint64", func(b *testing.B) {
		benchMarshal(b, e, uint64(math.MaxUint64))
	})
	b.Run("marshal int8", func(b *testing.B) {
		benchMarshal(b, e, math.MaxInt8)
	})
	b.Run("marshal int16", func(b *testing.B) {
		benchMarshal(b, e, math.MaxInt16)
	})
	b.Run("marshal int32", func(b *testing.B) {
		benchMarshal(b, e, math.MaxInt32)
	})
	b.Run("marshal int64", func(b *testing.B) {
		benchMarshal(b, e, math.MaxInt64)
	})
}

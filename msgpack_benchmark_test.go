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

var strvar16 = randString(16)
var strvar256 = randString(256)
var strvar65536 = randString(65536)

func BenchmarkLestrrat(b *testing.B) {
	b.StopTimer()
	e := lestrrat.NewEncoder(ioutil.Discard)
	b.StartTimer()
	b.ReportAllocs()
	bench(b, e, func(in io.Reader) Decoder { return lestrrat.NewDecoder(in) })
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
	bench(b, e, func(in io.Reader) Decoder {
		return VmihailencoDecoder{Decoder: vmihailenco.NewDecoder(in)}
	})
}

func bench(b *testing.B, e Encoder, newDecoder func(io.Reader) Decoder) {
	benchNilEncode(b, e)
	benchBool(b, e)
	benchStrings(b, e)
	benchFloats(b, e)
	benchInts(b, e)

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
	b.Run("serialize nil", func(b *testing.B) {
		benchEncode(b, e, nil)
	})
}

func benchNilDecode(b *testing.B, newDecoder func(io.Reader) Decoder) {
	b.Run("deserialize nil", func(b *testing.B) {
		benchDecode(b, func () Decoder {
			return newDecoder(bytes.NewBuffer([]byte{lestrrat.Nil.Byte()}))
		})
	})
}

func benchBool(b *testing.B, e Encoder) {
	b.Run("serialize true", func(b *testing.B) {
		benchEncode(b, e, true)
	})
	b.Run("serialize false", func(b *testing.B) {
		benchEncode(b, e, false)
	})
}

func benchStrings(b *testing.B, e Encoder) {
	for _, v := range []string{strvar16, strvar256, strvar65536} {
		b.Run(fmt.Sprintf("serialize string (len = %d)", len(v)), func(b *testing.B) {
			benchEncode(b, e, v)
		})
	}
}

func benchFloats(b *testing.B, e Encoder) {
	b.Run("serialize float32", func(b *testing.B) {
		benchEncode(b, e, math.MaxFloat32)
	})
	b.Run("serialize float64", func(b *testing.B) {
		benchEncode(b, e, math.MaxFloat64)
	})
}

func benchInts(b *testing.B, e Encoder) {
	b.Run("serialize uint8", func(b *testing.B) {
		benchEncode(b, e, math.MaxUint8)
	})
	b.Run("serialize uint16", func(b *testing.B) {
		benchEncode(b, e, math.MaxUint16)
	})
	b.Run("serialize uint32", func(b *testing.B) {
		benchEncode(b, e, math.MaxUint32)
	})
	b.Run("serialize uint64", func(b *testing.B) {
		benchEncode(b, e, uint64(math.MaxUint64))
	})
	b.Run("serialize int8", func(b *testing.B) {
		benchEncode(b, e, math.MaxInt8)
	})
	b.Run("serialize int16", func(b *testing.B) {
		benchEncode(b, e, math.MaxInt16)
	})
	b.Run("serialize int32", func(b *testing.B) {
		benchEncode(b, e, math.MaxInt32)
	})
	b.Run("serialize int64", func(b *testing.B) {
		benchEncode(b, e, math.MaxInt64)
	})
}

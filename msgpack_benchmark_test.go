// +build bench

package msgpack_test

import (
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"testing"

	lestrrat "github.com/lestrrat/go-msgpack"
	vmihailenco "gopkg.in/vmihailenco/msgpack.v2"
)

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

var strvar16 = randString(16)
var strvar256 = randString(256)
var strvar65536 = randString(65536)

func BenchmarkLestrrat(b *testing.B) {
	b.StopTimer()
	e := lestrrat.NewEncoder(ioutil.Discard)
	b.StartTimer()
	b.ReportAllocs()
	bench(b, e)
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
	bench(b, e)
}

func bench(b *testing.B, e Encoder) {
	benchNil(b, e)
	benchBool(b, e)
	benchStrings(b, e)
	benchFloats(b, e)
	benchInts(b, e)
}

func benchEncode(b *testing.B, e Encoder, v interface{}) {
	for i := 0; i < b.N; i++ {
		if err := e.Encode(v); err != nil {
			panic(err)
		}
	}
}

func benchNil(b *testing.B, e Encoder) {
	b.Run("serialize nil", func(b *testing.B) {
		benchEncode(b, e, nil)
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

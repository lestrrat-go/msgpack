// +build bench

package msgpack_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"reflect"
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
	var b1 EncodeDecodeBenchmarker
	b1.B = b
	b1.Encoder = e
	b1.MakeDecoder = func(in io.Reader) Decoder {
		return lestrrat.NewDecoder(in)
	}
	b1.Run()

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

type EncodeDecodeBenchmarker struct {
	B               *testing.B
	Encoder         Encoder
	MakeDecoder     func(io.Reader) Decoder
	SkipDecodeTrue  bool
	SkipDecodeFalse bool
}

func BenchmarkVmihailenco(b *testing.B) {
	b.StopTimer()
	e := VmihailencoEncoder{Encoder: vmihailenco.NewEncoder(ioutil.Discard)}
	b.ReportAllocs()
	b.StartTimer()

	var b1 EncodeDecodeBenchmarker
	b1.B = b
	b1.Encoder = e
	b1.MakeDecoder = func(in io.Reader) Decoder {
		return VmihailencoDecoder{Decoder: vmihailenco.NewDecoder(in)}
	}
	b1.SkipDecodeTrue = true
	b1.SkipDecodeFalse = true
	b1.Run()

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

func (bench *EncodeDecodeBenchmarker) Run() {
	bench.NilEncode()
	bench.BoolEncode()
	bench.StringEncode()
	bench.FloatEncode()
	bench.IntEncode()
	bench.MapEncode()
	bench.ArrayEncode()

	bench.NilDecode()
	bench.BoolDecode()
}

func (bench *EncodeDecodeBenchmarker) Encode(b *testing.B, v interface{}) {
	for i := 0; i < b.N; i++ {
		if err := bench.Encoder.Encode(v); err != nil {
			panic(err)
		}
	}
}

func handleErr(err error) {
	if err, ok := err.(stackTracer); ok {
		for _, f := range err.StackTrace() {
			fmt.Printf("%v\n", f)
		}
	}
	panic(err)
}

func (bench *EncodeDecodeBenchmarker) Decode(b *testing.B, in *bytes.Reader) {
	b.StopTimer()
	d := bench.MakeDecoder(in)
	b.StartTimer()

	var v interface{}
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		in.Seek(0, 0)
		b.StartTimer()
		if err := d.Decode(&v); err != nil {
			handleErr(err)
		}
	}
}

func (bench *EncodeDecodeBenchmarker) NilEncode() {
	bench.B.Run("encode nil", func(b *testing.B) {
		bench.Encode(b, nil)
	})
}

func (bench *EncodeDecodeBenchmarker) NilDecode() {
	data := []byte{lestrrat.Nil.Byte()}
	bench.B.Run("decode nil", func(b *testing.B) {
		bench.Decode(b, bytes.NewReader(data))
	})
}

func (bench *EncodeDecodeBenchmarker) BoolEncode() {
	bench.B.Run("encode true", func(b *testing.B) {
		bench.Encode(b, true)
	})
	bench.B.Run("encode false", func(b *testing.B) {
		bench.Encode(b, false)
	})
}

func (bench *EncodeDecodeBenchmarker) BoolDecode() {
	dataTrue := []byte{lestrrat.True.Byte()}
	dataFalse := []byte{lestrrat.False.Byte()}
	bench.B.Run("decode true", func(b *testing.B) {
		if bench.SkipDecodeFalse {
			b.Skip("Decode bool (true) skipped, as it panics")
		}
		bench.Decode(b, bytes.NewReader(dataTrue))
	})
	bench.B.Run("decode false", func(b *testing.B) {
		if bench.SkipDecodeFalse {
			b.Skip("Decode bool (false) skipped, as it panics")
		}
		bench.Decode(b, bytes.NewReader(dataFalse))
	})
}

func (bench *EncodeDecodeBenchmarker) StringEncode() {
	for _, v := range []string{strvar16, strvar256, strvar65536} {
		bench.B.Run(fmt.Sprintf("encode string (len=%d)", len(v)), func(b *testing.B) {
			bench.Encode(b, v)
		})
	}
}

func (bench *EncodeDecodeBenchmarker) FloatEncode() {
	bench.B.Run("encode float32", func(b *testing.B) {
		bench.Encode(b, math.MaxFloat32)
	})
	bench.B.Run("encode float64", func(b *testing.B) {
		bench.Encode(b, math.MaxFloat64)
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

func (bench *EncodeDecodeBenchmarker) IntEncode() {
	bench.B.Run("encode uint8", func(b *testing.B) {
		bench.Encode(b, math.MaxUint8)
	})
	bench.B.Run("encode uint16", func(b *testing.B) {
		bench.Encode(b, math.MaxUint16)
	})
	bench.B.Run("encode uint32", func(b *testing.B) {
		bench.Encode(b, math.MaxUint32)
	})
	bench.B.Run("encode uint64", func(b *testing.B) {
		bench.Encode(b, uint64(math.MaxUint64))
	})
	bench.B.Run("encode int8", func(b *testing.B) {
		bench.Encode(b, math.MaxInt8)
	})
	bench.B.Run("encode int16", func(b *testing.B) {
		bench.Encode(b, math.MaxInt16)
	})
	bench.B.Run("encode int32", func(b *testing.B) {
		bench.Encode(b, math.MaxInt32)
	})
	bench.B.Run("encode int64", func(b *testing.B) {
		bench.Encode(b, math.MaxInt64)
	})
}

func (bench *EncodeDecodeBenchmarker) MapEncode() {
	types := []reflect.Type{
		reflect.TypeOf(true),
		reflect.TypeOf(int(0)),
		reflect.TypeOf(int8(0)),
		reflect.TypeOf(int16(0)),
		reflect.TypeOf(int32(0)),
		reflect.TypeOf(int64(0)),
		reflect.TypeOf(uint(0)),
		reflect.TypeOf(uint8(0)),
		reflect.TypeOf(uint16(0)),
		reflect.TypeOf(uint32(0)),
		reflect.TypeOf(uint64(0)),
		reflect.TypeOf(float32(0)),
		reflect.TypeOf(float64(0)),
		reflect.TypeOf(""),
	}

	stype := reflect.TypeOf("")
	for _, typ := range types {
		mtype := reflect.MapOf(stype, typ)
		mv := reflect.MakeMap(mtype)
		for i := 0; i < 32; i++ {
			mv.SetMapIndex(reflect.ValueOf(fmt.Sprintf("%d", i)), reflect.New(typ).Elem())
		}
		bench.B.Run(fmt.Sprintf("encode map[string]%s", typ), func(b *testing.B) {
			bench.Encode(b, mv.Interface())
		})
	}
}

func (bench *EncodeDecodeBenchmarker) ArrayEncode() {
	types := []reflect.Type{
		reflect.TypeOf(true),
		reflect.TypeOf(int(0)),
		reflect.TypeOf(int8(0)),
		reflect.TypeOf(int16(0)),
		reflect.TypeOf(int32(0)),
		reflect.TypeOf(int64(0)),
		reflect.TypeOf(uint(0)),
		reflect.TypeOf(uint8(0)),
		reflect.TypeOf(uint16(0)),
		reflect.TypeOf(uint32(0)),
		reflect.TypeOf(uint64(0)),
		reflect.TypeOf(float32(0)),
		reflect.TypeOf(float64(0)),
		reflect.TypeOf(""),
	}

	for _, typ := range types {
		stype := reflect.SliceOf(typ)
		sv := reflect.MakeSlice(stype, 32, 32)
		for i := 0; i < 32; i++ {
			sv.Index(i).Set(reflect.New(typ).Elem())
		}
		bench.B.Run(fmt.Sprintf("encode []%s", typ), func(b *testing.B) {
			bench.Encode(b, sv.Interface())
		})
	}
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

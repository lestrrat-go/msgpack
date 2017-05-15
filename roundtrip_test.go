package msgpack_test

import (
	"math"
	"reflect"
	"testing"

	msgpack "github.com/lestrrat/go-msgpack"
	"github.com/stretchr/testify/assert"
)

func TestRoundTrip(t *testing.T) {
	var list = []interface{}{
		int8(-31),
		int8(127),
		int8(math.MaxInt8),
		int16(math.MaxInt16),
		int32(math.MaxInt32),
		int64(math.MaxInt64),
		uint8(math.MaxUint8),
		uint16(math.MaxUint16),
		uint32(math.MaxUint32),
		uint64(math.MaxUint64),
		float32(math.MaxFloat32),
		float64(math.MaxFloat64),
		"Hello, World!",
		[]byte("Hello, World!"),
	}

	for _, data := range list {
		t.Run(reflect.TypeOf(data).String(), func(t *testing.T) {
			b, err := msgpack.Marshal(data)
			if !assert.NoError(t, err, "Marshal should succeed") {
				return
			}
			var v interface{} = reflect.New(reflect.TypeOf(data)).Interface()
			if !assert.NoError(t, msgpack.Unmarshal(b, &v), "Unmarshal should succeed") {
				return
			}

			if !assert.Equal(t, data, v, "RoundTrip should succeed") {
				return
			}
		})
	}
}

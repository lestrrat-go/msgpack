package msgpack_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"testing"

	"github.com/lestrrat/go-msgpack"
	"github.com/stretchr/testify/assert"
)

func TestDecodeNil(t *testing.T) {
	var e interface{}
	var b = []byte{msgpack.Nil.Byte()}

	t.Run("decode via Unmarshal", func(t *testing.T) {
		var v interface{}
		if !assert.NoError(t, msgpack.Unmarshal(b, &v), "Unmarshal should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})
	t.Run("decode via DecodeNil", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		if !assert.NoError(t, msgpack.NewDecoder(buf).DecodeNil(), "DecodeNil should succeed") {
			return
		}
	})
	t.Run("decode via Decoder (interface{})", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		var v interface{} = 0xdeadcafe
		if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})
}

func TestDecodeBool(t *testing.T) {
	for _, code := range []msgpack.Code{msgpack.True, msgpack.False} {
		var e bool
		if code == msgpack.True {
			e = true
		}
		var b = []byte{code.Byte()}

		t.Run(fmt.Sprintf("decode %s via Unmarshal", code), func(t *testing.T) {
			var v bool
			if !assert.NoError(t, msgpack.Unmarshal(b, &v), "Unmarshal should succeed") {
				return
			}
			if !assert.Equal(t, e, v, "value should be %t", e) {
				return
			}
		})
		t.Run(fmt.Sprintf("decode %s via Unmarshal (interface{})", code), func(t *testing.T) {
			var v interface{}
			if !assert.NoError(t, msgpack.Unmarshal(b, &v), "Unmarshal (interface{}) should succeed") {
				return
			}
			if !assert.Equal(t, e, v, "value should be %t", e) {
				return
			}
		})
		t.Run(fmt.Sprintf("decode %s via DecodeBool", code), func(t *testing.T) {
			buf := bytes.NewBuffer(b)
			v, err := msgpack.NewDecoder(buf).DecodeBool()
			if !assert.NoError(t, err, "DecodeBool should succeed") {
				return
			}
			if !assert.Equal(t, e, v, "value should be %t", e) {
				return
			}
		})
		t.Run(fmt.Sprintf("decode %s via Decoder (interface{})", code), func(t *testing.T) {
			buf := bytes.NewBuffer(b)
			var v interface{} = 0xdeadcafe
			if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
				return
			}

			if !assert.Equal(t, e, v, "value should be %d", e) {
				return
			}
		})
	}
}

func TestDecodeFloat32(t *testing.T) {
	var e = float32(math.MaxFloat32)
	var b = make([]byte, 5)
	b[0] = msgpack.Float.Byte()
	binary.BigEndian.PutUint32(b[1:], math.Float32bits(e))

	t.Run("decode via Unmarshal", func(t *testing.T) {
		var v float32
		if !assert.NoError(t, msgpack.Unmarshal(b, &v), "Unmarshal should succeed") {
			return
		}
		if !assert.Equal(t, e, v, "value should be %f", e) {
			return
		}
	})
	t.Run("decode via Unmarshal (interface{})", func(t *testing.T) {
		var v interface{}
		if !assert.NoError(t, msgpack.Unmarshal(b, &v), "Unmarshal (interface{}) should succeed") {
			return
		}
		if !assert.Equal(t, e, v, "value should be %f", e) {
			return
		}
	})
	t.Run("decode via DecodeFloat32", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		v, err := msgpack.NewDecoder(buf).DecodeFloat32()
		if !assert.NoError(t, err, "DecodeFloat32 should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %f", e) {
			return
		}
	})
	t.Run("decode via Decoder (concrete)", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		var v float32
		if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %f", e) {
			return
		}
	})

	t.Run("decode via Decoder (interface{})", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		var v interface{}
		if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %f", e) {
			return
		}
	})
}

func TestDecodeFloat64(t *testing.T) {
	var e = float64(math.MaxFloat64)
	var b = make([]byte, 9)
	b[0] = msgpack.Double.Byte()
	binary.BigEndian.PutUint64(b[1:], math.Float64bits(e))

	t.Run("decode via Marshal", func(t *testing.T) {
		var v float64
		if !assert.NoError(t, msgpack.Unmarshal(b, &v), "Marshal should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %f", e) {
			return
		}
	})

	t.Run("decode via Marshal (interface{})", func(t *testing.T) {
		var v interface{}
		if !assert.NoError(t, msgpack.Unmarshal(b, &v), "Marshal should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %f", e) {
			return
		}
	})

	t.Run("decode via DecodeFloat64", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		v, err := msgpack.NewDecoder(buf).DecodeFloat64()
		if !assert.NoError(t, err, "DecodeFloat64 should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %f", e) {
			return
		}
	})

	t.Run("decode via Decoder (concrete)", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		var v float64
		if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %f", e) {
			return
		}
	})

	t.Run("decode via Decoder (interface{})", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		var v interface{}
		if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})
}

func TestDecodeUint8(t *testing.T) {
	var e = uint8(math.MaxUint8)
	var b = []byte{msgpack.Uint8.Byte(), byte(e)}

	t.Run("decode via DecodeUint8", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		v, err := msgpack.NewDecoder(buf).DecodeUint8()
		if !assert.NoError(t, err, "DecodeUint8 should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})

	t.Run("decode via Decoder (concrete)", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		var v uint8
		if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})

	t.Run("decode via Decoder (interface{})", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		var v interface{}
		if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})
}

func TestDecodeUint16(t *testing.T) {
	var e = uint16(math.MaxUint16)
	var b = make([]byte, 3)
	b[0] = msgpack.Uint16.Byte()
	binary.BigEndian.PutUint16(b[1:], uint16(e))

	t.Run("decode via DecodeUint16", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		v, err := msgpack.NewDecoder(buf).DecodeUint16()
		if !assert.NoError(t, err, "DecodeUint16 should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})

	t.Run("decode via Decode (concrete)", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		var v uint16
		if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})

	t.Run("decode via Decoder (interface{})", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		var v interface{}
		if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})
}

func TestDecodeUint32(t *testing.T) {
	var e = uint32(math.MaxUint32)
	var b = make([]byte, 5)
	b[0] = msgpack.Uint32.Byte()
	binary.BigEndian.PutUint32(b[1:], uint32(e))

	t.Run("decode via DecodeUint32", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		v, err := msgpack.NewDecoder(buf).DecodeUint32()
		if !assert.NoError(t, err, "DecodeUint32 should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})

	t.Run("decode via Decode (concrete)", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		var v uint32
		if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})

	t.Run("decode via Decoder (interface{})", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		var v interface{}
		if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})
}

func TestDecodeUint64(t *testing.T) {
	var e = uint64(math.MaxUint64)
	var b = make([]byte, 9)
	b[0] = msgpack.Uint64.Byte()
	binary.BigEndian.PutUint64(b[1:], uint64(e))

	t.Run("decode via DecodeUint64", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		v, err := msgpack.NewDecoder(buf).DecodeUint64()
		if !assert.NoError(t, err, "DecodeUint64 should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})

	t.Run("decode via Decode (concrete)", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		var v uint64
		if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})

	t.Run("decode via Decoder (interface{})", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		var v interface{}
		if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})
}

func TestDecodeInt8(t *testing.T) {
	var e = int8(math.MaxInt8)
	var b = []byte{msgpack.Int8.Byte(), byte(e)}

	t.Run("decode via DecodeInt8", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		v, err := msgpack.NewDecoder(buf).DecodeInt8()
		if !assert.NoError(t, err, "DecodeInt8 should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})

	t.Run("decode via Decoder (concrete)", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		var v int8
		if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})

	t.Run("decode via Decoder (interface{})", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		var v interface{}
		if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})
}

func TestDecodeInt16(t *testing.T) {
	var e = int16(math.MaxInt16)
	var b = make([]byte, 3)
	b[0] = msgpack.Int16.Byte()
	binary.BigEndian.PutUint16(b[1:], uint16(e))

	t.Run("decode via DecodeInt16", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		v, err := msgpack.NewDecoder(buf).DecodeInt16()
		if !assert.NoError(t, err, "DecodeInt16 should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})

	t.Run("decode via Decode (concrete)", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		var v int16
		if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})

	t.Run("decode via Decoder (interface{})", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		var v interface{}
		if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})
}

func TestDecodeInt32(t *testing.T) {
	var e = int32(math.MaxInt32)
	var b = make([]byte, 5)
	b[0] = msgpack.Int32.Byte()
	binary.BigEndian.PutUint32(b[1:], uint32(e))

	t.Run("decode via Unmarshal", func(t *testing.T) {
		var v int32
		if !assert.NoError(t, msgpack.Unmarshal(b, &v), "Unmarshal should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})

	t.Run("decode via DecodeInt32", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		v, err := msgpack.NewDecoder(buf).DecodeInt32()
		if !assert.NoError(t, err, "DecodeInt32 should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})

	t.Run("decode via Decode (concrete)", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		var v int32
		if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})

	t.Run("decode via Decoder (interface{})", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		var v interface{}
		if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})
}

func TestDecodeInt64(t *testing.T) {
	var e = int64(math.MaxInt64)
	var b = make([]byte, 9)
	b[0] = msgpack.Int64.Byte()
	binary.BigEndian.PutUint64(b[1:], uint64(e))

	t.Run("decode via Unmarshal", func(t *testing.T) {
		var v int64
		if !assert.NoError(t, msgpack.Unmarshal(b, &v), "Marshal should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})

	t.Run("decode via DecodeInt64", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		v, err := msgpack.NewDecoder(buf).DecodeInt64()
		if !assert.NoError(t, err, "DecodeInt64 should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})

	t.Run("decode via Decode (concrete)", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		var v int64
		if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})

	t.Run("decode via Decoder (interface{})", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		var v interface{}
		if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %d", e) {
			return
		}
	})
}

func TestDecodeStr8(t *testing.T) {
	var l = math.MaxUint8
	var e = makeString(l)
	var b = make([]byte, l+2)
	b[0] = msgpack.Str8.Byte()
	b[1] = byte(l)
	copy(b[2:], []byte(e))

	t.Run("decode via Unmarshal", func(t *testing.T) {
		var v string
		if !assert.NoError(t, msgpack.Unmarshal(b, &v), "Unmarshal should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %s", e) {
			return
		}
	})

	t.Run("decode via DecodeString", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		v, err := msgpack.NewDecoder(buf).DecodeString()
		if !assert.NoError(t, err, "DecodeString should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %s", e) {
			return
		}
	})

	t.Run("decode via Decode (concrete)", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		var v string
		if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %s", e) {
			return
		}
	})

	t.Run("decode via Decoder (interface{})", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		var v interface{}
		if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %s", e) {
			return
		}
	})
}

func TestDecodeStr16(t *testing.T) {
	var l = math.MaxUint16
	var e = makeString(l)
	var b = make([]byte, l+3)
	b[0] = msgpack.Str16.Byte()
	binary.BigEndian.PutUint16(b[1:], uint16(l))
	copy(b[3:], []byte(e))

	t.Run("decode via Unmarshal", func(t *testing.T) {
		var v string
		if !assert.NoError(t, msgpack.Unmarshal(b, &v), "Unmarshal should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %s", e) {
			return
		}
	})

	t.Run("decode via DecodeString", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		v, err := msgpack.NewDecoder(buf).DecodeString()
		if !assert.NoError(t, err, "DecodeString should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %s", e) {
			return
		}
	})

	t.Run("decode via Decode (concrete)", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		var v string
		if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %s", e) {
			return
		}
	})

	t.Run("decode via Decoder (interface{})", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		var v interface{}
		if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %s", e) {
			return
		}
	})
}

func TestDecodeStr32(t *testing.T) {
	var l = math.MaxUint16 + 1
	var e = makeString(l)
	var b = make([]byte, l+5)
	b[0] = msgpack.Str32.Byte()
	binary.BigEndian.PutUint32(b[1:], uint32(l))
	copy(b[5:], []byte(e))

	t.Run("decode via Unmarshal", func(t *testing.T) {
		var v string
		if !assert.NoError(t, msgpack.Unmarshal(b, &v), "Unmarshal should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %s", e) {
			return
		}
	})

	t.Run("decode via DecodeString", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		v, err := msgpack.NewDecoder(buf).DecodeString()
		if !assert.NoError(t, err, "DecodeString should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %s", e) {
			return
		}
	})

	t.Run("decode via Decode (concrete)", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		var v string
		if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %s", e) {
			return
		}
	})

	t.Run("decode via Decoder (interface{})", func(t *testing.T) {
		buf := bytes.NewBuffer(b)
		var v interface{}
		if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
			return
		}

		if !assert.Equal(t, e, v, "value should be %s", e) {
			return
		}
	})
}

func TestDecodeFixStr(t *testing.T) {
	for l := 1; l < 32; l++ {
		var e = makeString(l)
		var b = make([]byte, l+1)
		b[0] = byte(msgpack.FixStr0.Byte() + byte(l))
		copy(b[1:], []byte(e))

		t.Run(fmt.Sprintf("decode via Unmarshal (fixstr%d)", l), func(t *testing.T) {
			var v string
			if !assert.NoError(t, msgpack.Unmarshal(b, &v), "Unmarshal should succeed") {
				return
			}

			if !assert.Equal(t, e, v, "value should be %s", e) {
				return
			}
		})

		t.Run(fmt.Sprintf("decode via DecodeString (fixstr%d)", l), func(t *testing.T) {
			buf := bytes.NewBuffer(b)
			v, err := msgpack.NewDecoder(buf).DecodeString()
			if !assert.NoError(t, err, "DecodeString should succeed") {
				return
			}

			if !assert.Equal(t, e, v, "value should be %s", e) {
				return
			}
		})

		t.Run(fmt.Sprintf("decode via Decode (concrete) (fixstr%d)", l), func(t *testing.T) {
			buf := bytes.NewBuffer(b)
			var v string
			if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
				return
			}

			if !assert.Equal(t, e, v, "value should be %s", e) {
				return
			}
		})

		t.Run(fmt.Sprintf("decode via Decoder (interface{}) (fixstr%d)", l), func(t *testing.T) {
			buf := bytes.NewBuffer(b)
			var v interface{}
			if !assert.NoError(t, msgpack.NewDecoder(buf).Decode(&v), "Decode should succeed") {
				return
			}

			if !assert.Equal(t, e, v, "value should be %s", e) {
				return
			}
		})
	}
}

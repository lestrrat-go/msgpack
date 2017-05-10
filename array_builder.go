package msgpack

import (
	"io"
	"math"
	"reflect"

	"github.com/pkg/errors"
)

type arrayBuilder struct {
	buffer []interface{}
}

func NewArrayBuilder() ArrayBuilder {
	return &arrayBuilder{}
}

func (e *arrayBuilder) Add(v interface{}) {
	e.buffer = append(e.buffer, v)
}

func (e arrayBuilder) Encode(dst io.Writer) error {
	w := NewWriter(dst)
	switch c := e.Count(); {
	case c < 16:
		w.WriteByte(FixArray0.Byte() + byte(c))
	case c < math.MaxUint16:
		w.WriteByte(Array16.Byte())
		w.WriteUint16(uint16(c))
	case c < math.MaxUint32:
		w.WriteByte(Array32.Byte())
		w.WriteUint32(uint32(c))
	default:
		return errors.Errorf(`msgpack: array element count out of range (%d)`, c)
	}

	enc := NewEncoder(dst)
	for _, v := range e.buffer {
		if err := enc.Encode(v); err != nil {
			return errors.Wrapf(err, `msgpack: failed to encode array element %s`, reflect.TypeOf(v))
		}
	}
	return nil
}

func (e arrayBuilder) Count() int {
	return len(e.buffer)
}

func (e *arrayBuilder) Reset() {
	e.buffer = e.buffer[:0]
}

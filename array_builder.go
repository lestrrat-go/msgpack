package msgpack

import (
	"bytes"
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

func WriteArrayHeader(dst io.Writer, c int) error {
	var w Writer
	var ok bool
	if w, ok = dst.(Writer); !ok {
		w = NewWriter(dst)
	}

	switch {
	case c < 16:
		if err := w.WriteByte(FixArray0.Byte() + byte(c)); err != nil {
			return errors.Wrap(err, `msgpack: failed to write fixed array header`)
		}
	case c < math.MaxUint16:
		if err := w.WriteByte(Array16.Byte()); err != nil {
			return errors.Wrap(err, `msgpack: failed to write 16-bit array header prefix`)
		}
		if err := w.WriteUint16(uint16(c)); err != nil {
			return errors.Wrap(err, `msgpack: failed to write 16-bit array header`)
		}
	case c < math.MaxUint32:
		if err := w.WriteByte(Array32.Byte()); err != nil {
			return errors.Wrap(err, `msgpack: failed to write 32-bit array header prefix`)
		}
		if err := w.WriteUint32(uint32(c)); err != nil {
			return errors.Wrap(err, `msgpack: failed to write 32-bit array header`)
		}
	default:
		return errors.Errorf(`msgpack: array element count out of range (%d)`, c)
	}
	return nil
}

func (e arrayBuilder) Encode(dst io.Writer) error {
	if err := WriteArrayHeader(dst, e.Count()); err != nil {
		return errors.Wrap(err, `msgpack: failed to write array header`)
	}

	enc := NewEncoder(dst)
	for _, v := range e.buffer {
		if err := enc.Encode(v); err != nil {
			return errors.Wrapf(err, `msgpack: failed to encode array element %s`, reflect.TypeOf(v))
		}
	}
	return nil
}

func (e arrayBuilder) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := e.Encode(&buf); err != nil {
		return nil, errors.Wrap(err, `msgpack: failed to encode array`)
	}
	return buf.Bytes(), nil
}

func (e arrayBuilder) Count() int {
	return len(e.buffer)
}

func (e *arrayBuilder) Reset() {
	e.buffer = e.buffer[:0]
}

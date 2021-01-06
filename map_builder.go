package msgpack

import (
	"bytes"
	"io"
	"math"

	"github.com/pkg/errors"
)

type mapBuilder struct {
	buffer []interface{}
}

func NewMapBuilder() MapBuilder {
	return &mapBuilder{}
}

func (b *mapBuilder) Reset() {
	b.buffer = b.buffer[:0]
}

func (b *mapBuilder) Add(key string, value interface{}) {
	b.buffer = append(b.buffer, key, value)
}

func (b *mapBuilder) Count() int {
	return len(b.buffer) / 2
}

func WriteMapHeader(dst io.Writer, c int) error {
	var w Writer
	var ok bool
	if w, ok = dst.(Writer); !ok {
		w = NewWriter(dst)
	}

	switch {
	case c < 16:
		if err := w.WriteByte(FixMap0.Byte() + byte(c)); err != nil {
			return errors.Wrap(err, `failed to write element size prefix`)
		}
	case c < math.MaxUint16:
		if err := w.WriteByte(Map16.Byte()); err != nil {
			return errors.Wrap(err, `failed to write 16-bit element size prefix`)
		}
		if err := w.WriteUint16(uint16(c)); err != nil {
			return errors.Wrap(err, `failed to write 16-bit element size`)
		}
	case c < math.MaxUint32:
		if err := w.WriteByte(Map32.Byte()); err != nil {
			return errors.Wrap(err, `failed to write 32-bit element size prefix`)
		}

		if err := w.WriteUint32(uint32(c)); err != nil {
			return errors.Wrap(err, `failed to write 32-bit element size`)
		}
	default:
		return errors.Errorf(`map builder: map element count out of range (%d)`, c)
	}
	return nil
}

func (b *mapBuilder) Encode(dst io.Writer) error {
	if err := WriteMapHeader(dst, b.Count()); err != nil {
		return errors.Wrap(err, `failed to write map header`)
	}

	e := NewEncoder(dst)
	for i := 0; i < b.Count(); i++ {
		if err := e.Encode(b.buffer[i*2]); err != nil {
			return errors.Wrapf(err, `map builder: failed to encode map key %s`, b.buffer[i])
		}
		if err := e.Encode(b.buffer[i*2+1]); err != nil {
			return errors.Wrapf(err, `map builder: failed to encode map element for %s`, b.buffer[i])
		}
	}
	return nil
}

func (b *mapBuilder) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := b.Encode(&buf); err != nil {
		return nil, errors.Wrap(err, `map builder: failed to write map`)
	}

	return buf.Bytes(), nil
}

package msgpack

import (
	"bytes"
	"io"
	"math"

	"github.com/pkg/errors"
)

type mapBuilder struct {
	buffer  *bytes.Buffer
	count   int
	encoder *Encoder
}

func NewMapBuilder() MapBuilder {
	dst := &bytes.Buffer{}
	return &mapBuilder{
		buffer:  dst,
		encoder: NewEncoder(dst),
	}
}

func (b *mapBuilder) Reset() {
	b.buffer.Reset()
	b.count = 0
}

func (b *mapBuilder) Encode(key string, value interface{}) error {
	if err := b.encoder.Encode(key); err != nil {
		return errors.Wrapf(err, `map builder: failed to encode map key %s`, key)
	}
	if err := b.encoder.Encode(value); err != nil {
		return errors.Wrapf(err, `map builder: failed to encode map element for %s`, key)
	}

	b.count++
	return nil
}

func (b *mapBuilder) Count() int {
	return b.count
}

func (b *mapBuilder) WriteTo(dst io.Writer) (int64, error) {
	w := NewWriter(dst)

	switch c := b.Count(); {
	case c < 16:
		w.WriteByte(FixMap0.Byte() + byte(c))
	case c < math.MaxUint16:
		w.WriteByte(Map16.Byte())
		w.WriteUint16(uint16(c))
	case c < math.MaxUint32:
		w.WriteByte(Map32.Byte())
		w.WriteUint32(uint32(c))
	default:
		return 0, errors.Errorf(`map builder: map element count out of range (%d)`, c)
	}

	return b.buffer.WriteTo(w)
}

func (b *mapBuilder) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	if _, err := b.WriteTo(&buf); err != nil {
		return nil, errors.Wrap(err, `map builder: failed to write map`)
	}

	return buf.Bytes(), nil
}

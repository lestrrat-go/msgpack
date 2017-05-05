package msgpack

import (
	"io"

	"github.com/pkg/errors"
)

type mapBuilder struct {
	count   int
	encoder *Encoder
}

func NewMapBuilder(dst io.Writer) MapBuilder {
	return &mapBuilder{
		encoder: NewEncoder(dst),
	}
}

func (e *mapBuilder) Encode(key string, value interface{}) error {
	if err := e.encoder.Encode(key); err != nil {
		return errors.Wrapf(err, `msgpack: failed to encode map key %s`, key)
	}
	if err := e.encoder.Encode(value); err != nil {
		return errors.Wrapf(err, `msgpack: failed to encode map element for %s`, key)
	}

	e.count++
	return nil
}

func (e *mapBuilder) Count() int {
	return e.count
}

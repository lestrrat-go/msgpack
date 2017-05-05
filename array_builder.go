package msgpack

import (
	"io"
	"reflect"

	"github.com/pkg/errors"
)

type arrayBuilder struct {
	count   int
	encoder *Encoder
}

func NewArrayBuilder(dst io.Writer) ArrayBuilder {
	return &arrayBuilder{
		encoder: NewEncoder(dst),
	}
}

func (e *arrayBuilder) Encode(v interface{}) error {
	if err := e.encoder.Encode(v); err != nil {
		return errors.Wrapf(err, `msgpack: failed to encode array element %s`, reflect.TypeOf(v))
	}

	e.count++
	return nil
}

func (e *arrayBuilder) Count() int {
	return e.count
}

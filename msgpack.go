//go:generate stringer -type Code
//go:generate go run internal/cmd/gencontainer/gencontainer.go - encoder_container_gen.go
//go:generate go run internal/cmd/gendecoder/gendecoder.go
//go:generate go run internal/cmd/genencoder/genencoder.go

package msgpack

import (
	"bytes"
	"sync"

	"github.com/pkg/errors"
)

var appendingWriterPool = sync.Pool{
	New: allocAppendingWriter,
}

func allocAppendingWriter() interface{} {
	return newAppendingWriter(9)
}

func releaseAppendingWriter(w *appendingWriter) {
	w.buf = w.buf[0:0]
	appendingWriterPool.Put(w)
}

var encoderPool = sync.Pool{
	New: func() interface{} { return NewEncoder(nil) },
}

// Marshal takes a Go value and serializes it in msgpack format.
func Marshal(v interface{}) ([]byte, error) {
	var buf = appendingWriterPool.Get().(*appendingWriter)
	defer releaseAppendingWriter(buf)

	var enc = encoderPool.Get().(Encoder)
	enc.SetDestination(buf)
	if err := enc.Encode(v); err != nil {
		return nil, errors.Wrap(err, `failed to marshal`)
	}
	raw := buf.Bytes()
	ret := make([]byte, len(raw))
	copy(ret, raw)
	return ret, nil
}

// Unmarshal takes a byte slice and a pointer to a Go value and
// deserializes the Go value from the data in msgpack format.
func Unmarshal(data []byte, v interface{}) error {
	buf := bytes.NewBuffer(data)
	if err := NewDecoder(buf).Decode(v); err != nil {
		return errors.Wrap(err, `failed to unmarshal`)
	}
	return nil
}

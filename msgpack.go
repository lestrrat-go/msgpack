//go:generate stringer -type Code
//go:generate go run internal/cmd/genmap/genmap.go - encoder_map_gen.go

package msgpack

import (
	"bytes"

	"github.com/pkg/errors"
)

// Marshal takes a Go value and serializes it in msgpack format.
func Marshal(v interface{}) ([]byte, error) {
	var buf = newAppendingWriter(9)
	if err := NewEncoder(buf).Encode(v); err != nil {
		return nil, errors.Wrap(err, `failed to marshal`)
	}
	return buf.Bytes(), nil
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

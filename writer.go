package msgpack

import (
	"encoding/binary"
	"io"
	"sync"

	"github.com/pkg/errors"
)

type writer struct {
	dst io.Writer
}

func NewWriter(w io.Writer) Writer {
	return &writer{dst: w}
}

// sbytepool holds buffers of size = 1, only used when
// writing a single byte to an io.Writer
var sbytepool sync.Pool

func init() {
	sbytepool.New = allocsbyte
}

func allocsbyte() interface{} {
	return make([]byte, 1)
}

func (w *writer) Write(buf []byte) (int, error)  {
	return w.dst.Write(buf)
}

func (w *writer) WriteByte(v byte) error {
	buf := sbytepool.Get().([]byte)
	defer sbytepool.Put(buf)

	buf[0] = byte(v)
	n, err := w.Write(buf)
	if n != 1 {
		return errors.Wrap(err, `writer: failed to write byte`)
	}
	return nil
}

func (w *writer) WriteUint8(v uint8) error {
	return w.WriteByte(byte(v))
}

func (w *writer) WriteUint16(v uint16) error {
	return binary.Write(w.dst, binary.BigEndian, v)
}

func (w *writer) WriteUint32(v uint32) error {
	return binary.Write(w.dst, binary.BigEndian, v)
}

func (w *writer) WriteUint64(v uint64) error {
	return binary.Write(w.dst, binary.BigEndian, v)
}

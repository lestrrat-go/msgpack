package msgpack

import (
	"encoding/binary"
	"io"
)

type writer struct {
	dst io.Writer
	// Note: accessing buf concurrently is a mistake. But you DO NOT
	// write to a writer concurrently, or otherwise you can't guarantee
	// the correct memory layout. We assume that the caller doesn't do
	// anything silly.
	buf []byte
}

func NewWriter(w io.Writer) Writer {
	return &writer{
		dst: w,
		buf : make([]byte, 9),
	}
}

func (w writer) Write(buf []byte) (int, error) {
	return w.dst.Write(buf)
}

func (w writer) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func (w writer) WriteByte(v byte) error {
	b := w.buf[:1]
	b[0] = v
	_, err := w.Write(b)
	return err
}

func (w writer) WriteUint8(v uint8) error {
	return w.WriteByte(byte(v))
}

func (w writer) WriteUint16(v uint16) error {
	b := w.buf[:2]
	binary.BigEndian.PutUint16(b, v)
	_, err := w.Write(b)
	return err
}

func (w writer) WriteUint32(v uint32) error {
	b := w.buf[:4]
	binary.BigEndian.PutUint32(b, v)
	_, err := w.Write(b)
	return err
}

func (w writer) WriteUint64(v uint64) error {
	b := w.buf[:8]
	binary.BigEndian.PutUint64(b, v)
	_, err := w.Write(b)
	return err
}

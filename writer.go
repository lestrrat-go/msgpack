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

type appendingWriter struct {
	buf []byte
}

var _ Writer = &appendingWriter{}

func newAppendingWriter(size int) *appendingWriter {
	return &appendingWriter{
		buf: make([]byte, 0, size),
	}
}

func (w *appendingWriter) Write(buf []byte) (int, error) {
	w.buf = append(w.buf, buf...)
	return len(buf), nil
}

func (w *appendingWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func (w *appendingWriter) WriteByte(v byte) error {
	w.buf = append(w.buf, v)
	return nil
}

func (w *appendingWriter) WriteUint8(v uint8) error {
	return w.WriteByte(byte(v))
}

func (w *appendingWriter) WriteUint16(v uint16) error {
	const size = 2
	for i := 0; i < size; i++ {
		w.buf = append(w.buf, 0)
	}
	binary.BigEndian.PutUint16(w.buf[len(w.buf)-size:], v)
	return nil
}

func (w *appendingWriter) WriteUint32(v uint32) error {
	const size = 4
	for i := 0; i < size; i++ {
		w.buf = append(w.buf, 0)
	}
	binary.BigEndian.PutUint32(w.buf[len(w.buf)-size:], v)
	return nil
}

func (w *appendingWriter) WriteUint64(v uint64) error {
	const size = 8
	for i := 0; i < size; i++ {
		w.buf = append(w.buf, 0)
	}
	binary.BigEndian.PutUint64(w.buf[len(w.buf)-size:], v)
	return nil
}

func (w appendingWriter) Bytes() []byte {
	return w.buf
}

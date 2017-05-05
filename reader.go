package msgpack

import (
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
)

type reader struct {
	src io.Reader
}

func NewReader(r io.Reader) Reader {
	return &reader{
		src: r,
	}
}

func (r *reader) Read(buf []byte) (int, error) {
	return r.src.Read(buf)
}

func (r *reader) ReadByte() (byte, error) {
	b := sbytepool.Get().([]byte)
	defer sbytepool.Put(b)

	n, err := r.src.Read(b)
	if n != 1 {
		return byte(0), errors.Wrap(err, `reader: failed to read byte`)
	}
	return b[0], nil
}

func (r *reader) ReadUint8() (uint8, error) {
	b, err := r.ReadByte()
	if err != nil {
		return uint8(0), errors.Wrap(err, `reader: failed to read uint8`)
	}
	return uint8(b), nil
}

func (r *reader) ReadUint16() (uint16, error) {
	var v uint16
	if err := binary.Read(r.src, binary.BigEndian, &v); err != nil {
		return uint16(0), errors.Wrap(err, `reader: failed to read uint16`)
	}
	return v, nil
}

func (r *reader) ReadUint32() (uint32, error) {
	var v uint32
	if err := binary.Read(r.src, binary.BigEndian, &v); err != nil {
		return uint32(0), errors.Wrap(err, `reader: failed to read uint32`)
	}
	return v, nil
}

func (r *reader) ReadUint64() (uint64, error) {
	var v uint64
	if err := binary.Read(r.src, binary.BigEndian, &v); err != nil {
		return uint64(0), errors.Wrap(err, `reader: failed to read uint64`)
	}
	return v, nil
}

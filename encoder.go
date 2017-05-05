package msgpack

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"reflect"

	bufferpool "github.com/lestrrat/go-bufferpool"
	"github.com/pkg/errors"
)

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w: NewWriter(w),
	}
}

func isExtType(v interface{}) (int, bool) {
	rt := reflect.TypeOf(v)
	muExtEncode.RLock()
	typ, ok := extEncodeRegistry[rt]
	muExtEncode.RUnlock()
	if ok {
		return typ, true
	}

	return 0, false
}

var byteType = reflect.TypeOf(byte(0))

func (e *Encoder) Encode(v interface{}) error {
	if typ, ok := isExtType(v); ok {
		return e.EncodeExt(typ, v.(EncodeMsgpackExter))
	}

	if em, ok := v.(EncodeMsgpacker); ok {
		return em.EncodeMsgpack(e)
	}

	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return e.EncodeNil()
	}

	switch rv.Kind() {
	case reflect.Bool:
		return e.EncodeBool(v.(bool))
	case reflect.Float32:
		return e.EncodeFloat32(v.(float32))
	case reflect.Float64:
		return e.EncodeFloat64(v.(float64))
	case reflect.Uint8:
		return e.EncodeUint8(v.(uint8))
	case reflect.Uint16:
		return e.EncodeUint16(v.(uint16))
	case reflect.Uint32:
		return e.EncodeUint32(v.(uint32))
	case reflect.Uint64:
		return e.EncodeUint64(v.(uint64))
	case reflect.Int8:
		return e.EncodeInt8(v.(int8))
	case reflect.Int16:
		return e.EncodeInt16(v.(int16))
	case reflect.Int32:
		return e.EncodeInt32(v.(int32))
	case reflect.Int64:
		return e.EncodeInt64(v.(int64))
	case reflect.Int:
		return e.EncodeInt64(int64(v.(int)))
	case reflect.String:
		return e.EncodeString(v.(string))
	case reflect.Slice:
		if rv.Type().Elem() == byteType {
			return e.EncodeBytes(v.([]byte))
		}
		// XXX Is there a better way to do this...?
		var l = make([]interface{}, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			l[i] = rv.Index(i).Interface()
		}
		return e.EncodeArray(l)
	case reflect.Map:
		return e.EncodeMap(v)
	}

	//	return enc.Encode(e.w)
	return errors.Errorf(`msgpack: encode unimplemented for type %s`, rv.Type())
}

func (e *Encoder) EncodeNil() error {
	return e.w.WriteByte(Nil.Byte())
}

func (e *Encoder) EncodeBool(b bool) error {
	var code Code
	if b {
		code = True
	} else {
		code = False
	}
	return e.w.WriteByte(code.Byte())
}

func (e *Encoder) EncodeFloat32(f float32) error {
	if err := e.w.WriteByte(Float.Byte()); err != nil {
		return errors.Wrap(err, `msgpack: failed to write Float code`)
	}

	if err := e.w.WriteUint32(math.Float32bits(f)); err != nil {
		return errors.Wrap(err, `msgpack: failed to write Float payload`)
	}
	return nil
}

func (e *Encoder) EncodeFloat64(f float64) error {
	if err := e.w.WriteByte(Double.Byte()); err != nil {
		return errors.Wrap(err, `msgpack: failed to write Double code`)
	}

	if err := e.w.WriteUint64(math.Float64bits(f)); err != nil {
		return errors.Wrap(err, `msgpack: failed to write Double payload`)
	}
	return nil
}

func (e *Encoder) EncodeUint8(i uint8) error {
	if err := e.w.WriteByte(Uint8.Byte()); err != nil {
		return errors.Wrap(err, `msgpack: failed to write Uint8 code`)
	}

	if err := e.w.WriteUint8(i); err != nil {
		return errors.Wrap(err, `msgpack: failed to write Uint8 payload`)
	}
	return nil
}

func (e *Encoder) EncodeUint16(i uint16) error {
	if err := e.w.WriteByte(Uint16.Byte()); err != nil {
		return errors.Wrap(err, `msgpack: failed to write Uint16 code`)
	}

	if err := e.w.WriteUint16(i); err != nil {
		return errors.Wrap(err, `msgpack: failed to write Uint16 payload`)
	}
	return nil
}

func (e *Encoder) EncodeUint32(i uint32) error {
	if err := e.w.WriteByte(Uint32.Byte()); err != nil {
		return errors.Wrap(err, `msgpack: failed to write Uint32 code`)
	}

	if err := e.w.WriteUint32(i); err != nil {
		return errors.Wrap(err, `msgpack: failed to write Uint32 payload`)
	}
	return nil
}

func (e *Encoder) EncodeUint64(i uint64) error {
	if err := e.w.WriteByte(Uint64.Byte()); err != nil {
		return errors.Wrap(err, `msgpack: failed to write Uint64 code`)
	}

	if err := e.w.WriteUint64(i); err != nil {
		return errors.Wrap(err, `msgpack: failed to write Uint64 payload`)
	}
	return nil
}

func (e *Encoder) EncodeInt8(i int8) error {
	if err := e.w.WriteByte(Int8.Byte()); err != nil {
		return errors.Wrap(err, `msgpack: failed to write Int8 code`)
	}

	if err := e.w.WriteByte(byte(i)); err != nil {
		return errors.Wrap(err, `msgpack: failed to write Int8 payload`)
	}
	return nil
}

func (e *Encoder) EncodeInt16(i int16) error {
	if err := e.w.WriteByte(Int16.Byte()); err != nil {
		return errors.Wrap(err, `msgpack: failed to write Int16 code`)
	}

	if err := e.w.WriteUint16(uint16(i)); err != nil {
		return errors.Wrap(err, `msgpack: failed to write Int16 payload`)
	}
	return nil
}

func (e *Encoder) EncodeInt32(i int32) error {
	if err := e.w.WriteByte(Int32.Byte()); err != nil {
		return errors.Wrap(err, `msgpack: failed to write Int32 code`)
	}

	if err := e.w.WriteUint32(uint32(i)); err != nil {
		return errors.Wrap(err, `msgpack: failed to write Int32 payload`)
	}
	return nil
}

func (e *Encoder) EncodeInt64(i int64) error {
	if err := e.w.WriteByte(Int64.Byte()); err != nil {
		return errors.Wrap(err, `msgpack: failed to write Int64 code`)
	}

	if err := e.w.WriteUint64(uint64(i)); err != nil {
		return errors.Wrap(err, `msgpack: failed to write Int64 payload`)
	}
	return nil
}

func (e *Encoder) EncodeBytes(b []byte) error {
	l := len(b)

	var w int
	var code Code
	switch {
	case l <= math.MaxUint8:
		code = Bin8
		w = 1
	case l <= math.MaxUint16:
		code = Bin16
		w = 2
	case l <= math.MaxUint32:
		code = Bin32
		w = 4
	default:
		return errors.Errorf(`msgpack: string is too long (len=%d)`, l)
	}

	return encodeBuffer(code, w, e.w, bytes.NewBuffer(b))
}

func (e *Encoder) EncodeString(s string) error {
	l := len(s)

	var w int
	var code Code
	switch {
	case l < 32:
		code = Code(byte(l) + FixStr0.Byte())
	case l <= math.MaxUint8:
		code = Str8
		w = 1
	case l <= math.MaxUint16:
		code = Str16
		w = 2
	case l <= math.MaxUint32:
		code = Str32
		w = 4
	default:
		return errors.Errorf(`msgpack: string is too long (len=%d)`, l)
	}

	return encodeBuffer(code, w, e.w, bytes.NewBufferString(s))
}

func encodeBuffer(code Code, w int, dst Writer, buf *bytes.Buffer) error {
	if err := dst.WriteByte(code.Byte()); err != nil {
		return errors.Wrap(err, `msgpack: failed to write code`)
	}

	l := buf.Len()
	switch w {
	case 1:
		dst.WriteUint8(uint8(l))
	case 2:
		if err := binary.Write(dst, binary.BigEndian, uint16(l)); err != nil {
			return errors.Wrap(err, `msgpack: failed to write length`)
		}
	case 4:
		if err := binary.Write(dst, binary.BigEndian, uint32(l)); err != nil {
			return errors.Wrap(err, `msgpack: failed to write length`)
		}
	}

	if _, err := io.Copy(dst, buf); err != nil {
		return errors.Wrap(err, `msgpack: failed to write Str/Bin body`)
	}
	return nil
}

func (e *Encoder) EncodeArray(v []interface{}) error {
	buf := bufferpool.Get()
	defer bufferpool.Release(buf)

	// XXX: We could just as easily implement this without using
	// ArrayBuilder, but I think I'll leave this as it is for now
	// because this code path automatically tests it for us.
	// In reality, we only need to use an ArrayBuilder in case
	// we do not know the number of elements before hand.
	arrayb := NewArrayBuilder(buf)
	for i, x := range v {
		if err := arrayb.Encode(x); err != nil {
			return errors.Wrapf(err, `msgpack: failed to encode array element %d`, i)
		}
	}

	switch c := arrayb.Count(); {
	case c < 16:
		e.w.WriteByte(FixArray0.Byte() + byte(c))
	case c < math.MaxUint16:
		e.w.WriteByte(Array16.Byte())
		e.w.WriteUint16(uint16(c))
	case c < math.MaxUint32:
		e.w.WriteByte(Array32.Byte())
		e.w.WriteUint32(uint32(c))
	default:
		return errors.Errorf(`msgpack: array element count out of range (%d)`, c)
	}

	if _, err := buf.WriteTo(e.w); err != nil {
		return errors.Wrap(err, `msgpack: failed to write array payload`)
	}
	return nil
}

func (e *Encoder) EncodeMap(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Map {
		return errors.Errorf(`msgpack: argument to EncodeMap must be a map (not %s)`, rv.Type())
	}
	if rv.Type().Key().Kind() != reflect.String {
		return errors.Errorf(`msgpack: keys to maps must be strings (not %s)`, rv.Type().Key())
	}

	buf := bufferpool.Get()
	defer bufferpool.Release(buf)
	mapb := NewMapBuilder(buf)
	for _, key := range rv.MapKeys() {
		value := rv.MapIndex(key)
		if err := mapb.Encode(key.Interface().(string), value.Interface()); err != nil {
			return errors.Wrap(err, `msgpack: failed to encode map element`)
		}
	}

	switch c := mapb.Count(); {
	case c < 16:
		e.w.WriteByte(FixMap0.Byte() + byte(c))
	case c < math.MaxUint16:
		e.w.WriteByte(Map16.Byte())
		e.w.WriteUint16(uint16(c))
	case c < math.MaxUint32:
		e.w.WriteByte(Map32.Byte())
		e.w.WriteUint32(uint32(c))
	default:
		return errors.Errorf(`msgpack: map element count out of range (%d)`, c)
	}

	if _, err := buf.WriteTo(e.w); err != nil {
		return errors.Wrap(err, `msgpack: failed to write map payload`)
	}
	return nil
}

func (e *Encoder) EncodeExt(typ int, v EncodeMsgpackExter) error {
	buf := bufferpool.Get()
	defer bufferpool.Release(buf)
	var w = NewWriter(buf)
	if err := v.EncodeMsgpackExt(w); err != nil {
		return errors.Wrapf(err, `msgpack: failed during call to EncodeMsgpackExt for %s`, reflect.TypeOf(v))
	}

	switch l := buf.Len(); {
	case l == 1:
		if err := e.w.WriteByte(FixExt1.Byte()); err != nil {
			return errors.Wrap(err, `msgpack: failed to write fixext1 code`)
		}
	case l == 2:
		if err := e.w.WriteByte(FixExt2.Byte()); err != nil {
			return errors.Wrap(err, `msgpack: failed to write fixext2 code`)
		}
	case l == 4:
		if err := e.w.WriteByte(FixExt4.Byte()); err != nil {
			return errors.Wrap(err, `msgpack: failed to write fixext4 code`)
		}
	case l == 8:
		if err := e.w.WriteByte(FixExt8.Byte()); err != nil {
			return errors.Wrap(err, `msgpack: failed to write fixext8 code`)
		}
	case l == 16:
		if err := e.w.WriteByte(FixExt16.Byte()); err != nil {
			return errors.Wrap(err, `msgpack: failed to write fixext16 code`)
		}
	case l <= math.MaxUint8:
		if err := e.w.WriteByte(Ext8.Byte()); err != nil {
			return errors.Wrap(err, `msgpack: failed to write ext8 code`)
		}
		if err := e.w.WriteByte(byte(l)); err != nil {
			return errors.Wrap(err, `msgpack: failed to write ext8 payload length`)
		}
	case l <= math.MaxUint16:
		if err := e.w.WriteByte(Ext16.Byte()); err != nil {
			return errors.Wrap(err, `msgpack: failed to write ext16 code`)
		}
		if err := e.w.WriteUint16(uint16(l)); err != nil {
			return errors.Wrap(err, `msgpack: failed to write ext16 payload length`)
		}
	case l <= math.MaxUint32:
		if err := e.w.WriteByte(Ext32.Byte()); err != nil {
			return errors.Wrap(err, `msgpack: failed to write ext32 code`)
		}
		if err := e.w.WriteUint32(uint32(l)); err != nil {
			return errors.Wrap(err, `msgpack: failed to write ext32 payload length`)
		}
	default:
		return errors.Errorf(`msgpack: extension payload too large: %d bytes`, l)
	}

	if err := e.w.WriteByte(byte(typ)); err != nil {
		return errors.Wrap(err, `msgpack: failed to write typ code`)
	}

	if _, err := buf.WriteTo(e.w); err != nil {
		return errors.Wrap(err, `msgpack: failed to write extention payload`)
	}
	return nil
}

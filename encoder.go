package msgpack

import (
	"io"
	"math"
	"reflect"
	"strings"

	bufferpool "github.com/lestrrat/go-bufferpool"
	"github.com/pkg/errors"
)

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w: NewWriter(w),
	}
}

func isExtType(t reflect.Type) (int, bool) {
	muExtEncode.RLock()
	typ, ok := extEncodeRegistry[t]
	muExtEncode.RUnlock()
	if ok {
		return typ, true
	}

	return 0, false
}

var encodeMsgpackerType = reflect.TypeOf((*EncodeMsgpacker)(nil)).Elem()

func isEncodeMsgpacker(t reflect.Type) bool {
	return t.Implements(encodeMsgpackerType)
}

var byteType = reflect.TypeOf(byte(0))

func (e *Encoder) Encode(v interface{}) error {
	switch v := v.(type) {
	case string:
		return e.EncodeString(v)
	case bool:
		return e.EncodeBool(v)
	case float32:
		return e.EncodeFloat32(v)
	case float64:
		return e.EncodeFloat64(v)
	case uint:
		return e.EncodeUint64(uint64(v))
	case uint8:
		return e.EncodeUint8(v)
	case uint16:
		return e.EncodeUint16(v)
	case uint32:
		return e.EncodeUint32(v)
	case uint64:
		return e.EncodeUint64(v)
	case int:
		return e.EncodeInt64(int64(v))
	case int8:
		return e.EncodeInt8(v)
	case int16:
		return e.EncodeInt16(v)
	case int32:
		return e.EncodeInt32(v)
	case int64:
		return e.EncodeInt64(v)
	}

	// Find the first non-pointer, non-interface{}
	rv := reflect.ValueOf(v)
INDIRECT:
	for {
		if !rv.IsValid() {
			return e.EncodeNil()
		}
		if typ, ok := isExtType(rv.Type()); ok {
			return e.EncodeExt(typ, rv.Interface().(EncodeMsgpackExter))
		}

		if ok := isEncodeMsgpacker(rv.Type()); ok {
			return rv.Interface().(EncodeMsgpacker).EncodeMsgpack(e)
		}
		switch rv.Kind() {
		case reflect.Ptr, reflect.Interface:
			rv = rv.Elem()
		default:
			break INDIRECT
		}
	}

	if !rv.IsValid() {
		return e.EncodeNil()
	}

	v = rv.Interface()
	switch rv.Kind() {
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
	case reflect.Struct:
		return e.EncodeStruct(v)
	}

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

	if err := e.writePreamble(code, w, l); err != nil {
		return errors.Wrap(err, `msgpack: failed to write []byte preamble`)
	}
	e.w.Write(b)
	return nil
}

func (e *Encoder) EncodeString(s string) error {
	l := len(s)
	switch {
	case l < 32:
		e.w.WriteByte(FixStr0.Byte() | uint8(l))
	case l <= math.MaxUint8:
		e.w.WriteByte(Str8.Byte())
		e.w.WriteUint8(uint8(l))
	case l <= math.MaxUint16:
		e.w.WriteByte(Str16.Byte())
		e.w.WriteUint16(uint16(l))
	case l <= math.MaxUint32:
		e.w.WriteByte(Str32.Byte())
		e.w.WriteUint32(uint32(l))
	default:
		return errors.Errorf(`msgpack: string is too long (len=%d)`, l)
	}

	e.w.WriteString(s)
	return nil
}

func (e *Encoder) writePreamble(code Code, w int, l int) error {
	if err := e.w.WriteByte(code.Byte()); err != nil {
		return errors.Wrap(err, `msgpack: failed to write code`)
	}

	switch w {
	case 1:
		if err := e.w.WriteUint8(uint8(l)); err != nil {
			return errors.Wrap(err, `msgpack: failed to write length`)
		}
	case 2:
		if err := e.w.WriteUint16(uint16(l)); err != nil {
			return errors.Wrap(err, `msgpack: failed to write length`)
		}
	case 4:
		if err := e.w.WriteUint32(uint32(l)); err != nil {
			return errors.Wrap(err, `msgpack: failed to write length`)
		}
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

	mapb := NewMapBuilder()
	for _, key := range rv.MapKeys() {
		value := rv.MapIndex(key)
		if err := mapb.Encode(key.Interface().(string), value.Interface()); err != nil {
			return errors.Wrap(err, `msgpack: failed to encode map element`)
		}
	}

	if _, err := mapb.WriteTo(e.w); err != nil {
		return errors.Wrap(err, `msgpack: failed to write map payload`)
	}
	return nil
}

func parseMsgpackTag(rv reflect.StructField) (string, bool) {
	var name = rv.Name
	var omitempty bool
	if tag := rv.Tag.Get(`msgpack`); tag != "" {
		l := strings.Split(tag, ",")
		if len(l) > 0 && l[0] != "" {
			name = l[0]
		}

		if len(l) > 1 && l[1] == "omitempty" {
			omitempty = true
		}
	}
	return name, omitempty
}

func (e *Encoder) EncodeStruct(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Struct {
		return errors.Errorf(`msgpack: argument to EncodeStruct must be a struct (not %s)`, rv.Type())
	}
	mapb := NewMapBuilder()

	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		ft := rt.Field(i)
		if ft.PkgPath != "" {
			continue
		}

		name, omitempty := parseMsgpackTag(ft)
		if name == "-" {
			continue
		}

		field := rv.Field(i)
		if omitempty {
			if reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface()) {
				continue
			}
		}

		if err := mapb.Encode(name, field.Interface()); err != nil {
			return errors.Wrap(err, `msgpack: failed to encode struct field`)
		}
	}

	if _, err := mapb.WriteTo(e.w); err != nil {
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

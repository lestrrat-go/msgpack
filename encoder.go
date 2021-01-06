package msgpack

import (
	"io"
	"math"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// NewEncoder creates a new Encoder that writes serialized forms
// to the specified io.Writer
//
// Note that Encoders are NEVER meant to be shared concurrently
// between goroutines. You DO NOT write serialized data concurrently
// to the same destination.
func NewEncoder(w io.Writer) Encoder {
	enc := &encoder{nl: &encoderNL{}}
	enc.nl.SetDestination(w)
	return enc
}

// NewEncoderNoLock creates a new Encoder that DOES NOT protect
// users against accidental race conditions caused by concurrent
// method access. If you have complete control over the usage of
// this object, then the object returned by this constructor will
// shorten a whopping 30~50ns per method call. Use at your own peril
func NewEncoderNoLock(w io.Writer) Encoder {
	enc := &encoderNL{}
	enc.SetDestination(w)
	return enc
}

func (e *encoder) SetDestination(r io.Writer) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.nl.SetDestination(r)
}

func (enl *encoderNL) SetDestination(w io.Writer) {
	var dst Writer
	if x, ok := w.(Writer); ok {
		dst = x
	} else {
		dst = NewWriter(w)
	}

	enl.dst = dst
}

func inPositiveFixNumRange(i int64) bool {
	return i >= 0 && i <= 127
}

func inNegativeFixNumRange(i int64) bool {
	return i >= -31 && i <= -1
}

func isExtType(t reflect.Type) bool {
	muExtEncode.RLock()
	_, ok := extEncodeRegistry[t]
	muExtEncode.RUnlock()
	return ok
}

func isEncodeMsgpacker(t reflect.Type) bool {
	return t.Implements(encodeMsgpackerType)
}

func (enl *encoderNL) Writer() Writer {
	return enl.dst
}

//nolint:stylecheck,golint
func (enl *encoderNL) encodeBuiltin(v interface{}) (error, bool) {
	switch v := v.(type) {
	case string:
		return enl.EncodeString(v), true
	case []byte:
		return enl.EncodeBytes(v), true
	case bool:
		return enl.EncodeBool(v), true
	case float32:
		return enl.EncodeFloat32(v), true
	case float64:
		return enl.EncodeFloat64(v), true
	case uint:
		return enl.EncodeUint64(uint64(v)), true
	case uint8:
		return enl.EncodeUint8(v), true
	case uint16:
		return enl.EncodeUint16(v), true
	case uint32:
		return enl.EncodeUint32(v), true
	case uint64:
		return enl.EncodeUint64(v), true
	case int:
		return enl.EncodeInt64(int64(v)), true
	case int8:
		return enl.EncodeInt8(v), true
	case int16:
		return enl.EncodeInt16(v), true
	case int32:
		return enl.EncodeInt32(v), true
	case int64:
		return enl.EncodeInt64(v), true
	}

	return nil, false
}

func (enl *encoderNL) Encode(v interface{}) error {
	if err, ok := enl.encodeBuiltin(v); ok {
		return err
	}

	// Find the first non-pointer, non-interface{}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr && rv.Elem().IsValid() {
		if err, ok := enl.encodeBuiltin(rv.Elem().Interface()); ok {
			return err
		}
	}

INDIRECT:
	for {
		if !rv.IsValid() {
			return enl.EncodeNil()
		}

		if isExtType(rv.Type()) {
			return enl.EncodeExt(rv.Interface().(EncodeMsgpacker))
		}

		if ok := isEncodeMsgpacker(rv.Type()); ok {
			return rv.Interface().(EncodeMsgpacker).EncodeMsgpack(enl)
		}
		switch rv.Kind() {
		case reflect.Ptr, reflect.Interface:
			rv = rv.Elem()
		default:
			break INDIRECT
		}
	}

	if !rv.IsValid() {
		return enl.EncodeNil()
	}

	v = rv.Interface()
	switch rv.Kind() {
	case reflect.Slice: // , reflect.Array:
		return enl.EncodeArray(v)
	case reflect.Map:
		return enl.EncodeMap(v)
	case reflect.Struct:
		return enl.EncodeStruct(v)
	}

	return errors.Errorf(`msgpack: encode unimplemented for type %s`, rv.Type())
}

func (enl *encoderNL) encodePositiveFixNum(i uint8) error {
	return enl.dst.WriteByte(byte(i))
}

func (enl *encoderNL) encodeNegativeFixNum(i int8) error {
	return enl.dst.WriteByte(byte(i))
}

func (enl *encoderNL) EncodeNil() error {
	return enl.dst.WriteByte(Nil.Byte())
}

func (enl *encoderNL) EncodeBool(b bool) error {
	var code Code
	if b {
		code = True
	} else {
		code = False
	}
	return enl.dst.WriteByte(code.Byte())
}

func (enl *encoderNL) EncodePositiveFixNum(i uint8) error {
	if i > uint8(MaxPositiveFixNum) {
		return errors.Errorf(`msgpack: value %d is not in range for positive FixNum (127 >= x >= 0)`, i)
	}

	if err := enl.dst.WriteByte(byte(i)); err != nil {
		return errors.Wrap(err, `msgpack: failed to write FixNum`)
	}
	return nil
}

func (enl *encoderNL) EncodeNegativeFixNum(i int8) error {
	if i < -31 || i >= 0 {
		return errors.Errorf(`msgpack: value %d is not in range for positive FixNum (0 > x >= -31)`, i)
	}

	if err := enl.dst.WriteByte(byte(i)); err != nil {
		return errors.Wrap(err, `msgpack: failed to write FixNum`)
	}
	return nil
}

func (enl *encoderNL) EncodeBytes(b []byte) error {
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

	if err := enl.writePreamble(code, w, l); err != nil {
		return errors.Wrap(err, `msgpack: failed to write []byte preamble`)
	}
	if _, err := enl.dst.Write(b); err != nil {
		return errors.Wrap(err, `msgpack: failed to write []byte`)
	}
	return nil
}

func (enl *encoderNL) EncodeString(s string) error {
	l := len(s)
	switch {
	case l < 32:
		if err := enl.dst.WriteByte(FixStr0.Byte() | uint8(l)); err != nil {
			return errors.Wrap(err, `failed to encode fixed string length`)
		}
	case l <= math.MaxUint8:
		if err := enl.dst.WriteByte(Str8.Byte()); err != nil {
			return errors.Wrap(err, `msgpack: failed to encode 8-bit string length prefix`)
		}
		if err := enl.dst.WriteUint8(uint8(l)); err != nil {
			return errors.Wrap(err, `msgpack: failed to encode 8-bit string length`)
		}
	case l <= math.MaxUint16:
		if err := enl.dst.WriteByte(Str16.Byte()); err != nil {
			return errors.Wrap(err, `msgpack: failed to encode 16-bit string length prefix`)
		}
		if err := enl.dst.WriteUint16(uint16(l)); err != nil {
			return errors.Wrap(err, `msgpack: failed to encode 16-bit string length`)
		}
	case l <= math.MaxUint32:
		if err := enl.dst.WriteByte(Str32.Byte()); err != nil {
			return errors.Wrap(err, `msgpack: failed to encode 32-bit string length prefix`)
		}
		if err := enl.dst.WriteUint32(uint32(l)); err != nil {
			return errors.Wrap(err, `msgpack: failed to encode 32-bit string length`)
		}
	default:
		return errors.Errorf(`msgpack: string is too long (len=%d)`, l)
	}

	if _, err := enl.dst.WriteString(s); err != nil {
		return errors.Wrap(err, `msgpack: failed to write string`)
	}
	return nil
}

func (enl *encoderNL) writePreamble(code Code, w int, l int) error {
	if err := enl.dst.WriteByte(code.Byte()); err != nil {
		return errors.Wrap(err, `msgpack: failed to write code`)
	}

	switch w {
	case 1:
		if err := enl.dst.WriteUint8(uint8(l)); err != nil {
			return errors.Wrap(err, `msgpack: failed to write length`)
		}
	case 2:
		if err := enl.dst.WriteUint16(uint16(l)); err != nil {
			return errors.Wrap(err, `msgpack: failed to write length`)
		}
	case 4:
		if err := enl.dst.WriteUint32(uint32(l)); err != nil {
			return errors.Wrap(err, `msgpack: failed to write length`)
		}
	}
	return nil
}

func (enl *encoderNL) EncodeArrayHeader(l int) error {
	if err := WriteArrayHeader(enl.dst, l); err != nil {
		return errors.Wrap(err, `msgpack: failed to write array header`)
	}
	return nil
}

func (enl *encoderNL) EncodeArray(v interface{}) error {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
	default:
		return errors.Errorf(`msgpack: argument must be an array or a slice`)
	}

	if err := enl.EncodeArrayHeader(rv.Len()); err != nil {
		return err
	}

	switch rv.Type().Elem().Kind() {
	case reflect.String:
		return enl.encodeArrayString(rv.Convert(reflect.TypeOf([]string{})).Interface())
	case reflect.Bool:
		return enl.encodeArrayBool(v)
	case reflect.Int:
		return enl.encodeArrayInt(v)
	case reflect.Int8:
		return enl.encodeArrayInt8(v)
	case reflect.Int16:
		return enl.encodeArrayInt16(v)
	case reflect.Int32:
		return enl.encodeArrayInt32(v)
	case reflect.Int64:
		return enl.encodeArrayInt64(v)
	case reflect.Uint:
		return enl.encodeArrayUint(v)
	case reflect.Uint8:
		return enl.encodeArrayUint8(v)
	case reflect.Uint16:
		return enl.encodeArrayUint16(v)
	case reflect.Uint32:
		return enl.encodeArrayUint32(v)
	case reflect.Uint64:
		return enl.encodeArrayUint64(v)
	case reflect.Float32:
		return enl.encodeArrayFloat32(v)
	case reflect.Float64:
		return enl.encodeArrayFloat64(v)
	}

	for i := 0; i < rv.Len(); i++ {
		if err := enl.Encode(rv.Index(i).Interface()); err != nil {
			return errors.Wrap(err, `msgpack: failed to write array payload`)
		}
	}
	return nil
}

func (enl *encoderNL) EncodeMap(v interface{}) error {
	rv := reflect.ValueOf(v)

	if !rv.IsValid() {
		return enl.EncodeNil()
	}

	if rv.Kind() != reflect.Map {
		var typ string
		if !rv.IsValid() {
			typ = "invalid"
		} else {
			typ = rv.Type().String()
		}
		return errors.Errorf(`msgpack: argument to EncodeMap must be a map (not %s)`, typ)
	}

	if rv.IsNil() {
		return enl.EncodeNil()
	}

	if rv.Type().Key().Kind() != reflect.String {
		return errors.Errorf(`msgpack: keys to maps must be strings (not %s)`, rv.Type().Key())
	}

	// XXX We do NOT use MapBuilder's convenience methods except for the
	// WriteHeader bit, purely for performance reasons.
	keys := rv.MapKeys()
	if err := WriteMapHeader(enl.dst, len(keys)); err != nil {
		return errors.Wrap(err, `msgpack: failed to encode map header`)
	}

	// These are silly fast paths for common cases
	switch rv.Type().Elem().Kind() {
	case reflect.String:
		return enl.encodeMapString(v)
	case reflect.Bool:
		return enl.encodeMapBool(v)
	case reflect.Uint:
		return enl.encodeMapUint(v)
	case reflect.Uint8:
		return enl.encodeMapUint8(v)
	case reflect.Uint16:
		return enl.encodeMapUint16(v)
	case reflect.Uint32:
		return enl.encodeMapUint32(v)
	case reflect.Uint64:
		return enl.encodeMapUint64(v)
	case reflect.Int:
		return enl.encodeMapInt(v)
	case reflect.Int8:
		return enl.encodeMapInt8(v)
	case reflect.Int16:
		return enl.encodeMapInt16(v)
	case reflect.Int32:
		return enl.encodeMapInt32(v)
	case reflect.Int64:
		return enl.encodeMapInt64(v)
	case reflect.Float32:
		return enl.encodeMapFloat32(v)
	case reflect.Float64:
		return enl.encodeMapFloat64(v)
	default:
		for _, key := range keys {
			if err := enl.EncodeString(key.Interface().(string)); err != nil {
				return errors.Wrap(err, `failed to encode map key`)
			}

			if err := enl.Encode(rv.MapIndex(key).Interface()); err != nil {
				return errors.Wrap(err, `failed to encode map value`)
			}
		}
	}
	return nil
}

func parseMsgpackTag(rv reflect.StructField) (string, bool) {
	var tags = []string{`msgpack`, `msg`}

	var name = rv.Name
	var omitempty bool

	// We will support both msg and msgpack tags, the former
	// is used by tinylib/msgp, and the latter vmihailenco/msgpack
LOOP:
	for _, tagName := range tags {
		if tag, ok := rv.Tag.Lookup(tagName); ok && tag != "" {
			l := strings.Split(tag, ",")
			if len(l) > 0 && l[0] != "" {
				name = l[0]
			}

			if len(l) > 1 && l[1] == "omitempty" {
				omitempty = true
			}
			break LOOP
		}
	}
	return name, omitempty
}

// EncodeTime encodes time.Time as a sequence of two integers
func (enl *encoderNL) EncodeTime(t time.Time) error {
	if err := enl.dst.WriteByte(FixArray0.Byte() + byte(2)); err != nil {
		return errors.Wrap(err, `msgpack: failed to encode time header`)
	}

	if err := enl.EncodeInt64(t.Unix()); err != nil {
		return errors.Wrap(err, `msgpack: failed to encode seconds for time.Time`)
	}

	if err := enl.EncodeInt(t.Nanosecond()); err != nil {
		return errors.Wrap(err, `msgpack: failed to encode nanoseconds for time.Time`)
	}
	return nil
}

// EncodeStruct encodes a struct value as a map object.
func (enl *encoderNL) EncodeStruct(v interface{}) error {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return enl.EncodeNil()
	}

	if isExtType(rv.Type()) {
		return enl.EncodeExt(v.(EncodeMsgpacker))
	}

	if v, ok := v.(EncodeMsgpacker); ok {
		return v.EncodeMsgpack(enl)
	}

	// Special case
	if v, ok := v.(time.Time); ok {
		return enl.EncodeTime(v)
	}

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

		mapb.Add(name, field.Interface())
	}

	if err := mapb.Encode(enl.dst); err != nil {
		return errors.Wrap(err, `msgpack: failed to write map payload`)
	}
	return nil
}

func (enl *encoderNL) EncodeExtType(v EncodeMsgpacker) error {
	t := reflect.TypeOf(v)

	muExtDecode.RLock()
	typ, ok := extEncodeRegistry[t]
	muExtDecode.RUnlock()

	if !ok {
		return errors.Errorf(`msgpack: type %s has not been registered as an extension`, reflect.TypeOf(v))
	}

	if err := enl.dst.WriteByte(byte(typ)); err != nil {
		return errors.Wrapf(err, `msgpack: failed to write ext type for %s`, t)
	}
	return nil
}

func (enl *encoderNL) EncodeExt(v EncodeMsgpacker) error {
	w := newAppendingWriter(9)
	elocal := NewEncoder(w)

	if err := v.EncodeMsgpack(elocal); err != nil {
		return errors.Wrapf(err, `msgpack: failed during call to EncodeMsgpack for %s`, reflect.TypeOf(v))
	}

	buf := w.Bytes()
	if err := enl.EncodeExtHeader(len(buf)); err != nil {
		return errors.Wrap(err, `failed to encode ext header`)
	}
	if err := enl.EncodeExtType(v); err != nil {
		return errors.Wrap(err, `failed to encode ext type`)
	}

	for b := buf; len(b) > 0; {
		n, err := enl.dst.Write(buf)
		b = b[n:]
		if err != nil {
			return errors.Wrap(err, `msgpack: failed to write extension payload`)
		}
	}

	return nil
}

func (enl *encoderNL) EncodeExtHeader(l int) error {
	switch {
	case l == 1:
		if err := enl.dst.WriteByte(FixExt1.Byte()); err != nil {
			return errors.Wrap(err, `msgpack: failed to write fixext1 code`)
		}
	case l == 2:
		if err := enl.dst.WriteByte(FixExt2.Byte()); err != nil {
			return errors.Wrap(err, `msgpack: failed to write fixext2 code`)
		}
	case l == 4:
		if err := enl.dst.WriteByte(FixExt4.Byte()); err != nil {
			return errors.Wrap(err, `msgpack: failed to write fixext4 code`)
		}
	case l == 8:
		if err := enl.dst.WriteByte(FixExt8.Byte()); err != nil {
			return errors.Wrap(err, `msgpack: failed to write fixext8 code`)
		}
	case l == 16:
		if err := enl.dst.WriteByte(FixExt16.Byte()); err != nil {
			return errors.Wrap(err, `msgpack: failed to write fixext16 code`)
		}
	case l <= math.MaxUint8:
		if err := enl.dst.WriteByteUint8(Ext8.Byte(), uint8(l)); err != nil {
			return errors.Wrap(err, `msgpack: failed to write ext8 code and payload length`)
		}
	case l <= math.MaxUint16:
		if err := enl.dst.WriteByteUint16(Ext16.Byte(), uint16(l)); err != nil {
			return errors.Wrap(err, `msgpack: failed to write ext16 code and payload length`)
		}
	case l <= math.MaxUint32:
		if err := enl.dst.WriteByteUint32(Ext32.Byte(), uint32(l)); err != nil {
			return errors.Wrap(err, `msgpack: failed to write ext32 code and payload length`)
		}
	default:
		return errors.Errorf(`msgpack: extension payload too large: %d bytes`, l)
	}

	return nil
}

package msgpack

import (
	"bufio"
	"io"
	"reflect"
	"time"

	bufferpool "github.com/lestrrat-go/bufferpool"
	"github.com/pkg/errors"
)

// NewDecoder creates a Decoder instance
func NewDecoder(r io.Reader) Decoder {
	d := &decoder{nl: &decoderNL{}}
	d.nl.SetSource(r)
	return d
}

// NewDecoderNoLock creates a new Decoder that DOES NOT protect
// users against accidental race conditions caused by concurrent
// method access. If you have complete control over the usage of
// this object, then the object returned by this constructor will
// shorten a whopping 30~50ns per method call. Use at your own peril
func NewDecoderNoLock(r io.Reader) Decoder {
	d := &decoderNL{}
	d.SetSource(r)
	return d
}

func (d *decoder) SetSource(r io.Reader) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.nl.SetSource(r)
}

func (dnl *decoderNL) SetSource(r io.Reader) {
	dnl.raw = bufio.NewReader(r)
	dnl.src = NewReader(dnl.raw)
}

func (dnl *decoderNL) Reader() Reader {
	return dnl.src
}

func (dnl *decoderNL) ReadCode() (Code, error) {
	b, err := dnl.raw.ReadByte()
	if err != nil {
		return Code(0), errors.Wrap(err, `msgpack: failed to read code`)
	}

	return Code(b), nil
}

func (dnl *decoderNL) PeekCode() (Code, error) {
	code, err := dnl.ReadCode()
	if err != nil {
		return code, errors.Wrap(err, `msgpack: failed to peek code`)
	}

	if err := dnl.raw.UnreadByte(); err != nil {
		return Code(0), errors.Wrap(err, `msgpack: failed to unread code`)
	}
	return code, nil
}

func (dnl *decoderNL) isNil() bool {
	code, err := dnl.PeekCode()
	if err != nil {
		return false
	}
	return code == Nil
}

func (dnl *decoderNL) DecodeNil(v *interface{}) error {
	code, err := dnl.ReadCode()
	if err != nil {
		return errors.Wrap(err, `msgpack: failed to read code`)
	}
	if code != Nil {
		return errors.Errorf(`msgpack: expected Nil, got %s`, code)
	}
	if v != nil {
		*v = nil
	}
	return nil
}

func (dnl *decoderNL) DecodeBool(b *bool) error {
	code, err := dnl.ReadCode()
	if err != nil {
		return errors.Wrap(err, `msgpack: failed to read code`)
	}

	switch code {
	case True:
		*b = true
		return nil
	case False:
		*b = false
		return nil
	default:
		return errors.Errorf(`msgpack: expected True/False, got %s`, code)
	}
}

func (dnl *decoderNL) DecodeBytes(v *[]byte) error {
	code, err := dnl.ReadCode()
	if err != nil {
		return errors.Wrap(err, `msgpack: failed to read code`)
	}

	var l int64
	switch {
	case code == Bin8:
		v, err := dnl.src.ReadUint8()
		if err != nil {
			return errors.Wrap(err, `msgpack: failed to read length for string/byte slice`)
		}
		l = int64(v)
	case code == Bin16:
		v, err := dnl.src.ReadUint16()
		if err != nil {
			return errors.Wrap(err, `msgpack: failed to read length for string/byte slice`)
		}
		l = int64(v)
	case code == Bin32:
		v, err := dnl.src.ReadUint32()
		if err != nil {
			return errors.Wrap(err, `msgpack: failed to read length for string/byte slice`)
		}
		l = int64(v)
	default:
		return errors.Errorf(`msgpack: invalid code: expected Bin8/Bin16/Bin32, got %s`, code)
	}

	// Sanity check
	if l < 0 {
		return errors.Errorf(`msgpack: invalid byte slice length %d`, l)
	}

	b := make([]byte, l)
	for x := b; len(x) > 0; {
		n, err := dnl.raw.Read(x)
		if err != nil {
			return errors.Wrap(err, `msgpack: failed to read byte slice`)
		}
		x = x[n:]
	}

	*v = b
	return nil
}

func (dnl *decoderNL) DecodeString(s *string) error {
	code, err := dnl.ReadCode()
	if err != nil {
		return errors.Wrap(err, `msgpack: failed to read code`)
	}

	var l int64
	switch {
	case code >= FixStr0 && code <= FixStr31:
		l = int64(code.Byte() - FixStr0.Byte())
	case code == Str8:
		v, err := dnl.src.ReadUint8()
		if err != nil {
			return errors.Wrap(err, `msgpack: failed to read length for string/byte slice`)
		}
		l = int64(v)
	case code == Str16:
		v, err := dnl.src.ReadUint16()
		if err != nil {
			return errors.Wrap(err, `msgpack: failed to read length for string/byte slice`)
		}
		l = int64(v)
	case code == Str32:
		v, err := dnl.src.ReadUint32()
		if err != nil {
			return errors.Wrap(err, `msgpack: failed to read length for string/byte slice`)
		}
		l = int64(v)
	default:
		return errors.Errorf(`msgpack: invalid code: expected FixStr/Str8/Str16/Str32, got %s`, code)
	}

	// Sanity check
	if l < 0 {
		return errors.Errorf(`msgpack: invalid string length %d`, l)
	}

	// Read the contents of the string.
	// Now, here's the tricky part: conversion from byte slice to string is
	// just going to create a copy of b as an immutable string, and so this
	// byte slice is just thrown away. It would be nice if we could reuse
	// this memory later...
	buf := bufferpool.Get()
	defer bufferpool.Release(buf)

	// Make sure we can write l bytes
	buf.Grow(int(l))
	b := buf.Bytes()
	for x := b[:l]; len(x) > 0; {
		n, err := dnl.raw.Read(x)
		if err != nil {
			return errors.Wrap(err, `msgpack: failed to read string`)
		}
		x = x[n:]
	}

	*s = string(b[:l])
	return nil
}

func (dnl *decoderNL) DecodeArrayLength(l *int) error {
	code, err := dnl.ReadCode()
	if err != nil {
		return errors.Wrap(err, `msgpack: failed to read code`)
	}

	if code >= FixArray0 && code <= FixArray15 {
		*l = int(code.Byte() - FixArray0.Byte())
		return nil
	}

	switch code {
	case Array16:
		s, err := dnl.src.ReadUint16()
		if err != nil {
			return errors.Wrap(err, `msgpack: failed to read array size for Array16`)
		}
		*l = int(s)
	case Array32:
		s, err := dnl.src.ReadUint32()
		if err != nil {
			return errors.Wrap(err, `msgpack: failed to read array size for Array32`)
		}
		*l = int(s)
	default:
		return errors.Errorf(`msgpack: unsupported array type %s`, code)
	}

	return nil
}

func (dnl *decoderNL) DecodeArray(v interface{}) error {
	var size int
	if err := dnl.DecodeArrayLength(&size); err != nil {
		return errors.Wrap(err, `msgpack: failed to decode array length`)
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return errors.Errorf(`msgpack: DecodeArray expected pointer to slice, got %s`, rv.Type())
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Slice {
		return errors.Errorf(`msgpack: DecodeArray expected slice, got %s`, rv.Type())
	}

	slice := reflect.MakeSlice(rv.Type(), size, size)
	for i := 0; i < size; i++ {
		e := slice.Index(i)
		if e.Kind() == reflect.Ptr {
			if e.IsNil() {
				e.Set(reflect.New(e.Type().Elem()))
			}
		} else {
			e = e.Addr()
		}
		if err := dnl.Decode(e.Interface()); err != nil {
			return errors.Wrapf(err, `msgpack: failed to decode array element %d`, i)
		}
	}

	rv.Set(slice)
	return nil
}

func (dnl *decoderNL) DecodeMapLength(l *int) error {
	code, err := dnl.ReadCode()
	if err != nil {
		return errors.Wrap(err, `msgpack: failed to read code`)
	}

	if code == Nil {
		*l = -1
		return nil
	}

	if code >= FixMap0 && code <= FixMap15 {
		*l = int(code.Byte() - FixMap0.Byte())
		return nil
	}

	switch code {
	case Map16:
		s, err := dnl.src.ReadUint16()
		if err != nil {
			return errors.Wrap(err, `msgpack: failed to read array size for Map16`)
		}
		*l = int(s)
	case Map32:
		s, err := dnl.src.ReadUint32()
		if err != nil {
			return errors.Wrap(err, `msgpack: failed to read array size for Map32`)
		}
		*l = int(s)
	default:
		return errors.Errorf(`msgpack: unsupported map type %s`, code)
	}

	return nil
}

func (dnl *decoderNL) DecodeMap(v *map[string]interface{}) error {
	var size int
	if err := dnl.DecodeMapLength(&size); err != nil {
		return errors.Wrap(err, `msgpack: failed to decode map length`)
	}

	if size == -1 {
		*v = nil
		return nil
	}

	m := make(map[string]interface{})
	for i := 0; i < size; i++ {
		var s string
		if err := dnl.DecodeString(&s); err != nil {
			return errors.Wrap(err, `msgpack: failed to decode map key`)
		}

		var v interface{}
		if err := dnl.Decode(&v); err != nil {
			return errors.Wrapf(err, `msgpack: failed to decode map element for key %s`, s)
		}
		m[s] = v
	}
	*v = m
	return nil
}

func (dnl *decoderNL) DecodeTime(v *time.Time) error {
	var size int
	if err := dnl.DecodeArrayLength(&size); err != nil {
		return errors.Wrap(err, `msgpack: failed to decode array length for time.Time`)
	}
	if size != 2 {
		return errors.Errorf(`msgpack: expected array of size 2 (got %d)`, size)
	}

	var seconds int64
	if err := dnl.DecodeInt64(&seconds); err != nil {
		return errors.Wrap(err, `msgpack: failed to decode seconds part for time.Time`)
	}
	var nanosecs int
	if err := dnl.DecodeInt(&nanosecs); err != nil {
		return errors.Wrap(err, `msgpack: failed to decode nanoseconds part for time.Time`)
	}

	*v = time.Unix(seconds, int64(nanosecs))
	return nil
}

func (dnl *decoderNL) DecodeStruct(v interface{}) error {
	if v, ok := v.(DecodeMsgpacker); ok {
		return dnl.DecodeExt(v)
	}

	if v, ok := v.(*time.Time); ok {
		return dnl.DecodeTime(v)
	}

	var size int
	if err := dnl.DecodeMapLength(&size); err != nil {
		return errors.Wrap(err, `msgpack: failed to decode map length`)
	}

	var rv = reflect.ValueOf(v)
	// You better be a pointer to a struct, damnit
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return errors.New(`msgpack: expected pointer to struct`)
	}

	if size == -1 {
		if rv.CanSet() {
			rv.Set(reflect.Value{})
		}
		return nil
	}

	var rt = rv.Elem().Type()
	// Find the fields
	name2field := map[string]reflect.Value{}
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if field.PkgPath != "" {
			continue
		}

		name, _ := parseMsgpackTag(field)
		if name == "-" {
			continue
		}

		name2field[name] = rv.Elem().Field(i)
	}

	var key string
	for i := 0; i < size; i++ {
		if err := dnl.Decode(&key); err != nil {
			return errors.Wrapf(err, `msgpack: failed to decode struct key at index %d`, i)
		}

		f, ok := name2field[key]
		if !ok {
			continue
		}
		if dnl.isNil() {
			if err := dnl.DecodeNil(nil); err != nil {
				return errors.Wrapf(err, `msgpack: failed to decode nil field %s`, key)
			}
			continue
		}

		if f.Kind() == reflect.Slice {
			r := reflect.New(f.Type()).Elem()
			if err := dnl.Decode(r.Addr().Interface()); err != nil {
				return errors.Wrapf(err, `msgpack: failed to decode slice value for key %s`, key)
			}
			f.Set(r)
		} else if f.Kind() == reflect.Struct {
			if err := dnl.Decode(f.Addr().Interface()); err != nil {
				return errors.Wrapf(err, `msgpack: failed to decode struct value for key %s (struct)`, key)
			}
		} else if f.Kind() == reflect.Ptr && f.Type().Elem().Kind() == reflect.Struct {
			r := reflect.New(f.Type().Elem())
			if err := dnl.Decode(r.Interface()); err != nil {
				return errors.Wrapf(err, `msgpack: failed to decode struct value for key %s (pointer to struct)`, key)
			}
			f.Set(r)
		} else {
			var fv reflect.Value
			if f.Kind() == reflect.Ptr {
				fv = reflect.New(f.Type().Elem())
			} else {
				fv = reflect.New(f.Type())
			}
			if err := dnl.Decode(fv.Interface()); err != nil {
				return errors.Wrapf(err, `msgpack: failed to decode struct value for key %s (not struct/pointer to struct)`, key)
			}

			if err := assignIfCompatible(f, fv.Elem()); err != nil {
				return errors.Wrapf(err, `msgpack: failed to assign struct value for key %s`, key)
			}
		}
	}

	return nil
}

func assignIfCompatible(dst, src reflect.Value) (err error) {
	// src will always be from result of a Decode. therefore
	// we will have no pointers. But dst can be either a
	// a pointer or the actual type
	var dstlist = []reflect.Value{dst}
	if dst.Kind() == reflect.Ptr {
		if dst.Elem().IsValid() {
			dstlist = append(dstlist, dst.Elem())
		} else {
			v := reflect.New(dst.Type().Elem())
			dstlist = append(dstlist, v.Elem())
			defer func() {
				if err == nil {
					dst.Set(v)
				}
			}()
		}
	}

	for _, dst := range dstlist {
		if !dst.IsValid() {
			continue
		}

		if dst.Type() == emptyInterfaceType {
			dst.Set(reflect.ValueOf(src.Interface()))
			return nil
		}

		// Unmarshalers need to assign in case of pointers, too

		if src.Type().AssignableTo(dst.Type()) {
			dst.Set(src)
			return nil
		}

		if src.Type().ConvertibleTo(dst.Type()) {
			dst.Set(src.Convert(dst.Type()))
			return nil
		}

		// We may have a container...
		if dst.Kind() == reflect.Slice && src.Kind() == reflect.Slice {
			slice := reflect.MakeSlice(dst.Type(), src.Len(), src.Len())
			if dst.Type().Elem() == emptyInterfaceType {
				// if this is the case, we can assign everything from
				// src to dst
				for i := 0; i < src.Len(); i++ {
					dst.Index(i).Set(src.Index(i))
				}
				return nil
			}

			if src.Type().Elem() == emptyInterfaceType {
				sliceElemType := dst.Type().Elem() // []string -> string
				isSliceElemPtr := dst.Type().Elem().Kind() == reflect.Ptr

				// See if we can install src's contents into dst
			SLICE:
				for i := 0; i < src.Len(); i++ {
					e := src.Index(i)

					var assignErr error
					switch {
					case sliceElemType == e.Elem().Type():
						if assignErr = assignIfCompatible(slice.Index(i), e.Elem()); assignErr == nil {
							continue SLICE
						}
					case isSliceElemPtr:
						if sliceElemType.Elem() == e.Elem().Type() {
							if assignErr = assignIfCompatible(slice.Index(i), e.Elem().Addr()); assignErr == nil {
								continue SLICE
							}
						} else if e.Elem().Type().ConvertibleTo(sliceElemType.Elem()) {
							v := reflect.New(sliceElemType.Elem())
							v.Elem().Set(e.Elem().Convert(sliceElemType.Elem()))
							if assignErr = assignIfCompatible(slice.Index(i), v); assignErr == nil {
								continue SLICE
							}
						}
					}

					return errors.Wrapf(assignErr, `msgpack: cannot assign slice element on index %d (slice type = %s, element type = %s)`, i, dst.Type(), e.Elem().Type())
				}
				dst.Set(slice)
				return nil
			}
		}
	}
	return errors.Errorf(`invalid type for assignment: dst = %s, src = %s`, dst.Type(), src.Type())
}

var emptyInterfaceType = reflect.TypeOf((*interface{})(nil)).Elem()

func (dnl *decoderNL) Decode(v interface{}) error {
	rv := reflect.ValueOf(v)

	// The result of decoding must be assigned to v, and v
	// should be a pointer
	if rv.Kind() == reflect.Interface {
		// if it's an interface, get the underlying type
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		// report error
		var typ reflect.Type
		if rv.IsValid() {
			typ = rv.Type()
		}
		return &InvalidDecodeError{
			Type: typ,
		}
	}

	// First, try guessing what to do by checking the type of the
	// incoming payload. These are the easy choices
	switch v := v.(type) {
	case *interface{}:
		goto FromCode
	case *int:
		return dnl.DecodeInt(v)
	case *int8:
		return dnl.DecodeInt8(v)
	case *int16:
		return dnl.DecodeInt16(v)
	case *int32:
		return dnl.DecodeInt32(v)
	case *int64:
		return dnl.DecodeInt64(v)
	case *uint:
		return dnl.DecodeUint(v)
	case *uint8:
		return dnl.DecodeUint8(v)
	case *uint16:
		return dnl.DecodeUint16(v)
	case *uint32:
		return dnl.DecodeUint32(v)
	case *uint64:
		return dnl.DecodeUint64(v)
	case *float32:
		return dnl.DecodeFloat32(v)
	case *float64:
		return dnl.DecodeFloat64(v)
	case *[]byte:
		return dnl.DecodeBytes(v)
	case *string:
		return dnl.DecodeString(v)
	case *map[string]interface{}:
		return dnl.DecodeMap(v)
	case DecodeMsgpacker:
		// If we know this object does its own decoding, we bypass everything
		// and just let it handle itself
		return v.DecodeMsgpack(dnl)
	}

	// Next up: try using reflect to find out the general family of
	// the payload.
	switch rv.Elem().Kind() {
	case reflect.Struct:
		return dnl.DecodeStruct(v)
	case reflect.Slice:
		list := reflect.New(rv.Elem().Type())
		if err := dnl.DecodeArray(list.Interface()); err != nil {
			return errors.Wrap(err, `msgpack: failed to decode array`)
		}
		if err := assignIfCompatible(reflect.ValueOf(v).Elem(), list.Elem()); err != nil {
			return errors.Wrap(err, `msgpack: error while assigning slice elements`)
		}
		return nil
	}

FromCode:
	decoded, err := dnl.decodeInterface(v)
	if err != nil {
		return errors.Wrap(err, `msgpack: failed to decode interface value`)
	}

	// if decoded == nil, then we have a special case, where we need
	// to assign a nil to v, but the type of the nil must match v
	// (if you get what I mean)
	if decoded == nil {
		// Note: I wish I could just return without doing anything, but
		// because the encoded value is explicitly nil, it's only right
		// to properly assign a nil to whatever value that was passed to
		// this method.
		rv.Elem().Set(reflect.Zero(rv.Elem().Type()))
		return nil
	}

	dv := reflect.ValueOf(decoded)

	// Since we know rv to be a pointer, we must set the new value
	// to the destination of the pointer.
	dst := rv.Elem()

	// If it's assignable, assign, and we're done.
	if err := assignIfCompatible(dst, dv); err == nil {
		return nil
	}

	// This could only happen if we have a decoder that creates
	// the value dynamically, such as in the case of struct
	// decoder or extension decoder.
	if reflect.PtrTo(dst.Type()) == dv.Type() {
		dst.Set(dv.Elem())
		return nil
	}

	return errors.Errorf(`msgpack: cannot assign %s to %s`, dv.Type(), dst.Type())
}

// Note: v is only used as a hint. do not assign to it inside this method
func (dnl *decoderNL) decodeInterface(v interface{}) (interface{}, error) {
	code, err := dnl.PeekCode()
	if err != nil {
		return nil, errors.Wrap(err, `msgpack: failed to peek code`)
	}

	switch {
	case IsExtFamily(code):
		var size int
		if err := dnl.DecodeExtLength(&size); err != nil {
			return nil, errors.Wrap(err, `msgpack: failed to read extension sizes`)
		}

		var typ reflect.Type
		if err := dnl.DecodeExtType(&typ); err != nil {
			return nil, errors.Wrap(err, `msgpack: faied to read extension type`)
		}

		rv := reflect.New(typ).Interface().(DecodeMsgpacker)
		if err := rv.DecodeMsgpack(dnl); err != nil {
			return nil, errors.Wrap(err, `msgpack: failed to decode extension`)
		}
		return rv, nil
	case IsFixNumFamily(code):
		return int8(code), nil
	case code == Nil:
		// Optimization: doesn't require any more handling than to
		// throw away the code
		if _, err := dnl.raw.ReadByte(); err != nil {
			return nil, errors.Wrap(err, `msgpack: failed to ready byte`)
		}
		return nil, nil
	case code == True:
		// Optimization: doesn't require any more handling than to
		// throw away the code
		if _, err := dnl.raw.ReadByte(); err != nil {
			return false, errors.Wrap(err, `msgpack: failed to read byte`)
		}
		return true, nil
	case code == False:
		// Optimization: doesn't require any more handling than to
		// throw away the code
		if _, err := dnl.raw.ReadByte(); err != nil {
			return false, errors.Wrap(err, `msgpack: failed to read byte`)
		}
		return false, nil
	case code == Int8:
		var x int8
		if err := dnl.DecodeInt8(&x); err != nil {
			return nil, errors.Wrap(err, `msgpack: failed to decode Int8`)
		}
		return x, nil
	case code == Int16:
		var x int16
		if err := dnl.DecodeInt16(&x); err != nil {
			return nil, errors.Wrap(err, `msgpack: failed to decode Int16`)
		}
		return x, nil
	case code == Int32:
		var x int32
		if err := dnl.DecodeInt32(&x); err != nil {
			return nil, errors.Wrap(err, `msgpack: failed to decode Int32`)
		}
		return x, nil
	case code == Int64:
		var x int64
		if err := dnl.DecodeInt64(&x); err != nil {
			return nil, errors.Wrap(err, `msgpack: failed to decode Int64`)
		}
		return x, nil
	case code == Uint8:
		var x uint8
		if err := dnl.DecodeUint8(&x); err != nil {
			return nil, errors.Wrap(err, `msgpack: failed to decode Uint8`)
		}
		return x, nil
	case code == Uint16:
		var x uint16
		if err := dnl.DecodeUint16(&x); err != nil {
			return nil, errors.Wrap(err, `msgpack: failed to decode Uint16`)
		}
		return x, nil
	case code == Uint32:
		var x uint32
		if err := dnl.DecodeUint32(&x); err != nil {
			return nil, errors.Wrap(err, `msgpack: failed to decode Uint32`)
		}
		return x, nil
	case code == Uint64:
		var x uint64
		if err := dnl.DecodeUint64(&x); err != nil {
			return nil, errors.Wrap(err, `msgpack: failed to decode Uint64`)
		}
		return x, nil
	case code == Float:
		var x float32
		if err := dnl.DecodeFloat32(&x); err != nil {
			return nil, errors.Wrap(err, `msgpack: failed to decode Float`)
		}
		return x, nil
	case code == Double:
		var x float64
		if err := dnl.DecodeFloat64(&x); err != nil {
			return nil, errors.Wrap(err, `msgpack: failed to decode Double`)
		}
		return x, nil
	case IsBinFamily(code):
		var b []byte
		if err := dnl.DecodeBytes(&b); err != nil {
			return nil, errors.Wrapf(err, `msgpack: failed to decode %s`, code)
		}
		return b, nil
	case IsStrFamily(code):
		var s string
		if err := dnl.DecodeString(&s); err != nil {
			return nil, errors.Wrapf(err, `msgpack: failed to decode %s`, code)
		}
		return s, nil
	case IsArrayFamily(code):
		var l []interface{}
		if err := dnl.DecodeArray(&l); err != nil {
			return nil, errors.Wrapf(err, `msgpack: failed to decode %s`, code)
		}
		return l, nil
	case IsMapFamily(code):
		// Special case: If the object is a Map type, and the target object
		// is a Struct, we do the struct decoding bit.
		// could be &struct, interface{}(&struct{}), or interface{}(&interface{}(struct{}))
		rv := reflect.ValueOf(v)
		if rv.Type().Kind() == reflect.Interface {
			rv = rv.Elem()
		}

		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
			if rv.Kind() == reflect.Interface {
				rv = rv.Elem()
			}
			if rv.Kind() == reflect.Struct {
				v := reflect.New(rv.Type()).Interface()
				if err := dnl.DecodeStruct(v); err != nil {
					return nil, errors.Wrap(err, `msgpack: failed to decode struct`)
				}
				return reflect.ValueOf(v).Elem().Interface(), nil
			}
		}

		var v = make(map[string]interface{})
		if err := dnl.DecodeMap(&v); err != nil {
			return nil, errors.Wrap(err, `msgpack: failed to decode map`)
		}
		return v, nil
	default:
		return nil, errors.Errorf(`msgpack: invalid code %s`, code)
	}
}

func (dnl *decoderNL) DecodeExtLength(l *int) error {
	code, err := dnl.ReadCode()
	if err != nil {
		return errors.Wrap(err, `msgpack: failed to read code`)
	}

	var payloadSize int
	switch code {
	case FixExt1:
		payloadSize = 1
	case FixExt2:
		payloadSize = 2
	case FixExt4:
		payloadSize = 1
	case FixExt8:
		payloadSize = 8
	case FixExt16:
		payloadSize = 16
	case Ext8:
		s, err := dnl.src.ReadUint8()
		if err != nil {
			return errors.Wrap(err, `msgpack: failed to read size for ext8 value`)
		}
		payloadSize = int(s)
	case Ext16:
		s, err := dnl.src.ReadUint16()
		if err != nil {
			return errors.Wrap(err, `msgpack: failed to read size for ext16 value`)
		}
		payloadSize = int(s)
	case Ext32:
		s, err := dnl.src.ReadUint32()
		if err != nil {
			return errors.Wrap(err, `msgpack: failed to read size for ext32 value`)
		}
		payloadSize = int(s)
	default:
		return errors.Errorf(`msgpack: invalid ext code %s`, code)
	}
	*l = payloadSize
	return nil
}

func (dnl *decoderNL) DecodeExt(v DecodeMsgpacker) error {
	var size int
	if err := dnl.DecodeExtLength(&size); err != nil {
		return errors.Wrap(err, `msgpack: failed to read extension sizes`)
	}

	var typ reflect.Type
	if err := dnl.DecodeExtType(&typ); err != nil {
		return errors.Wrap(err, `msgpack: faied to read extension type`)
	}

	if rt := reflect.TypeOf(v); rt != reflect.PtrTo(typ) {
		return errors.Errorf(`msgpack: extension should be %s, got %s`, typ, rt)
	}

	if err := v.DecodeMsgpack(dnl); err != nil {
		return errors.Wrap(err, `msgpack: failed to call DecodeMsgpack`)
	}
	return nil
}

func (dnl *decoderNL) DecodeExtType(v *reflect.Type) error {
	t, err := dnl.src.ReadUint8()
	if err != nil {
		return errors.Wrap(err, `msgpack: failed to read type for extension`)
	}

	muExtDecode.Lock()
	typ, ok := extDecodeRegistry[int(t)]
	muExtDecode.Unlock()

	if !ok {
		return errors.Errorf(`msgpack: type %d is not registered as an extension`, int(t))
	}

	*v = typ
	return nil
}

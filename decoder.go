package msgpack

import (
	"bufio"
	"io"
	"math"
	"reflect"

	bufferpool "github.com/lestrrat/go-bufferpool"
	"github.com/pkg/errors"
)

type valueDecoder interface {
	Decode(io.Reader) (interface{}, error)
}

var decoders = map[Code]valueDecoder{
	Float:   &floatDecoder{code: Float},
	Double:  &floatDecoder{code: Double},
	Uint8:   &uintDecoder{code: Uint8},
	Uint16:  &uintDecoder{code: Uint16},
	Uint32:  &uintDecoder{code: Uint32},
	Uint64:  &uintDecoder{code: Uint64},
	Int8:    &intDecoder{code: Int8},
	Int16:   &intDecoder{code: Int16},
	Int32:   &intDecoder{code: Int32},
	Int64:   &intDecoder{code: Int64},
	Ext8:    &extDecoder{code: Ext8},
	FixExt8: &extDecoder{code: FixExt8},
	Str8:    &strDecoder{code: Str8},
	Str16:   &strDecoder{code: Str16},
	Str32:   &strDecoder{code: Str32},
	Bin8:    &strDecoder{code: Bin8},
	Bin16:   &strDecoder{code: Bin16},
	Bin32:   &strDecoder{code: Bin32},
	Array16: &arrayDecoder{code: Array16},
	Array32: &arrayDecoder{code: Array32},
	Map16:   &mapDecoder{code: Map16},
	Map32:   &mapDecoder{code: Map32},
}

func init() {
	for i := 0; i < 32; i++ {
		code := Code(FixStr0.Byte() + byte(i))
		decoders[code] = &fixstrDecoder{code: code}
	}
	for i := 0; i < 16; i++ {
		code := Code(FixArray0.Byte() + byte(i))
		decoders[code] = &arrayDecoder{code: code}

		code = Code(FixMap0.Byte() + byte(i))
		decoders[code] = &mapDecoder{code: code}
	}
}

type floatDecoder struct {
	code Code
}

func (d *floatDecoder) Decode(r io.Reader) (interface{}, error) {
	rdr := NewReader(r)
	switch d.code {
	case Float:
		n, err := rdr.ReadUint32()
		if err != nil {
			return nil, errors.Wrap(err, `msgpack: failed to read uint32 for Float`)
		}
		return math.Float32frombits(n), nil
	case Double:
		n, err := rdr.ReadUint64()
		if err != nil {
			return nil, errors.Wrap(err, `msgpack: failed to read uint64 for Float`)
		}
		return math.Float64frombits(n), nil
	default:
		return nil, errors.Errorf(`msgpack: unknown float code %s`, d.code)
	}
}

type uintDecoder struct {
	code Code
}

func (d *uintDecoder) Decode(r io.Reader) (interface{}, error) {
	rdr := NewReader(r)
	switch d.code {
	case Uint8:
		v, err := rdr.ReadUint8()
		if err != nil {
			return nil, errors.Wrap(err, `msgpack: failed to decode uint8`)
		}
		return v, nil
	case Uint16:
		v, err := rdr.ReadUint16()
		if err != nil {
			return nil, errors.Wrap(err, `msgpack: failed to decode uint16`)
		}
		return v, nil
	case Uint32:
		v, err := rdr.ReadUint32()
		if err != nil {
			return nil, errors.Wrap(err, `msgpack: failed to decode uint32`)
		}
		return v, nil
	case Uint64:
		v, err := rdr.ReadUint64()
		if err != nil {
			return nil, errors.Wrap(err, `msgpack: failed to decode uint64`)
		}
		return v, nil
	default:
		return nil, errors.Errorf(`msgpack: invalid code %s for uint`, d.code)
	}
}

type intDecoder struct {
	code Code
}

func (d *intDecoder) Decode(r io.Reader) (interface{}, error) {
	rdr := NewReader(r)
	switch d.code {
	case Int8:
		v, err := rdr.ReadUint8()
		if err != nil {
			return nil, errors.Wrap(err, `msgpack: failed to decode int8`)
		}
		return int8(v), nil
	case Int16:
		v, err := rdr.ReadUint16()
		if err != nil {
			return nil, errors.Wrap(err, `msgpack: failed to decode int16`)
		}
		return int16(v), nil
	case Int32:
		v, err := rdr.ReadUint32()
		if err != nil {
			return nil, errors.Wrap(err, `msgpack: failed to decode uint32`)
		}
		return int32(v), nil
	case Int64:
		v, err := rdr.ReadUint64()
		if err != nil {
			return nil, errors.Wrap(err, `msgpack: failed to decode int64`)
		}
		return int64(v), nil
	default:
		return nil, errors.Errorf(`msgpack: invalid code %s for int`, d.code)
	}
}

type strDecoder struct {
	code Code
}

func (d *strDecoder) Decode(r io.Reader) (interface{}, error) {
	rdr := NewReader(r)
	var l int64
	switch d.code {
	case Str8, Bin8:
		v, err := rdr.ReadUint8()
		if err != nil {
			return nil, errors.Wrap(err, `msgpack: failed to read length for string/byte slice`)
		}
		l = int64(v)
	case Str16, Bin16:
		v, err := rdr.ReadUint16()
		if err != nil {
			return nil, errors.Wrap(err, `msgpack: failed to read length for string/byte slice`)
		}
		l = int64(v)
	case Str32, Bin32:
		v, err := rdr.ReadUint32()
		if err != nil {
			return nil, errors.Wrap(err, `msgpack: failed to read length for string/byte slice`)
		}
		l = int64(v)
	}

	buf := bufferpool.Get()
	switch d.code {
	case Bin8, Bin16, Bin32:
		// Note: no defer, because the callee wants to use this buffer
		return buf.Bytes(), nil
	}

	// buf.String() is an immutable copy, so we don't need to have
	// the buffer lying around
	defer bufferpool.Release(buf)
	if _, err := io.CopyN(buf, r, l); err != nil {
		return nil, errors.Wrap(err, `msgpack: failed to read string`)
	}

	return buf.String(), nil
}

type fixstrDecoder struct {
	code Code
}

func (d *fixstrDecoder) Decode(r io.Reader) (interface{}, error) {
	l := int64(d.code.Byte() - FixStr0.Byte())

	buf := bufferpool.Get()
	bufferpool.Release(buf)
	n, err := io.CopyN(buf, r, l)
	if n != l && err != nil {
		return nil, errors.Wrap(err, `msgpack: failed to decode FixStr (body)`)
	}

	return buf.String(), nil
}

type arrayDecoder struct {
	code Code
}

func (d *arrayDecoder) Decode(r io.Reader) (interface{}, error) {
	var size int
	if d.code >= FixArray0 && d.code <= FixArray15 {
		size = int(d.code.Byte() - FixArray0.Byte())
	} else {
		rdr := NewReader(r)
		switch d.code {
		case Array16:
			s, err := rdr.ReadUint16()
			if err != nil {
				return nil, errors.Wrap(err, `msgpack: failed to read array size for Array16`)
			}
			size = int(s)
		case Array32:
			s, err := rdr.ReadUint32()
			if err != nil {
				return nil, errors.Wrap(err, `msgpack: failed to read array size for Array32`)
			}
			size = int(s)
		default:
			return nil, errors.Errorf(`msgpack: unsupported array type %s`, d.code)
		}
	}

	l := make([]interface{}, size)
	dec := NewDecoder(r)
	for i := 0; i < size; i++ {
		if err := dec.Decode(&l[i]); err != nil {
			return nil, errors.Wrapf(err, `msgpack: failed to decode array at index %d`, i)
		}
	}

	return l, nil
}

type mapDecoder struct {
	code Code
}

func (d *mapDecoder) Decode(r io.Reader) (interface{}, error) {
	var size int
	if d.code >= FixMap0 && d.code <= FixMap15 {
		size = int(d.code.Byte() - FixMap0.Byte())
	} else {
		rdr := NewReader(r)
		switch d.code {
		case Map16:
			s, err := rdr.ReadUint16()
			if err != nil {
				return nil, errors.Wrap(err, `msgpack: failed to read map size for Map16`)
			}
			size = int(s)
		case Map32:
			s, err := rdr.ReadUint32()
			if err != nil {
				return nil, errors.Wrap(err, `msgpack: failed to read map size for Map32`)
			}
			size = int(s)
		default:
			return nil, errors.Errorf(`msgpack: unsupported map type %s`, d.code)
		}
	}

	dec := NewDecoder(r)
	var m = map[string]interface{}{}
	var key string
	var value interface{}
	for i := 0; i < size; i++ {
		if err := dec.Decode(&key); err != nil {
			return nil, errors.Wrapf(err, `msgpack: failed to decode fixmap key at index %d`, i)
		}
		if err := dec.Decode(&value); err != nil {
			return nil, errors.Wrapf(err, `msgpack: failed to decode fixmap value for key %s`, key)
		}

		m[key] = value
	}

	return m, nil
}

type structDecoder struct {
	code   Code
	target reflect.Type
}

func (d *structDecoder) Decode(r io.Reader) (interface{}, error) {
	var size int
	if d.code >= FixMap0 && d.code <= FixMap15 {
		size = int(d.code.Byte() - FixMap0.Byte())
	} else {
		rdr := NewReader(r)
		switch d.code {
		case Map16:
			s, err := rdr.ReadUint16()
			if err != nil {
				return nil, errors.Wrap(err, `msgpack: failed to read map size for Map16`)
			}
			size = int(s)
		case Map32:
			s, err := rdr.ReadUint32()
			if err != nil {
				return nil, errors.Wrap(err, `msgpack: failed to read map size for Map32`)
			}
			size = int(s)
		default:
			return nil, errors.Errorf(`msgpack: unsupported map type %s`, d.code)
		}
	}

	dec := NewDecoder(r)
	var s = reflect.New(d.target)

	// XXX: This needs caching
	name2field := map[string]reflect.Value{}
	for i := 0; i < d.target.NumField(); i++ {
		field := d.target.Field(i)
		if field.PkgPath != "" {
			continue
		}

		name, _ := parseMsgpackTag(field)
		if name == "-" {
			continue
		}

		name2field[name] = s.Elem().Field(i)
	}

	var key string
	var value interface{}
	for i := 0; i < size; i++ {
		if err := dec.Decode(&key); err != nil {
			return nil, errors.Wrapf(err, `msgpack: failed to decode struct key at index %d`, i)
		}

		f, ok := name2field[key]
		if !ok {
			continue
		}

		if f.Kind() == reflect.Struct {
			if err := dec.Decode(f.Addr().Interface()); err != nil {
				return nil, errors.Wrapf(err, `msgpack: failed to decode struct value for key %s`, key)
			}
		} else if f.Kind() == reflect.Ptr && f.Type().Elem().Kind() == reflect.Struct {
			if err := dec.Decode(f.Interface()); err != nil {
				return nil, errors.Wrapf(err, `msgpack: failed to decode struct value for key %s`, key)
			}
		} else {
			if err := dec.Decode(&value); err != nil {
				return nil, errors.Wrapf(err, `msgpack: failed to decode struct value for key %s`, key)
			}

			fv := reflect.ValueOf(value)
			if !fv.Type().ConvertibleTo(f.Type()) {
				return nil, errors.Errorf(`msgpack: cannot convert from %s to %s`, fv.Type(), f.Type())
			}
			f.Set(reflect.ValueOf(value).Convert(f.Type()))
		}

	}

	return s.Elem().Interface(), nil
}

type extDecoder struct {
	code Code
}

var decodeMsgpackExterType = reflect.TypeOf((*DecodeMsgpackExter)(nil)).Elem()

func (d *extDecoder) Decode(r io.Reader) (interface{}, error) {
	rdr := NewReader(r)

	var size int
	switch d.code {
	case Ext8:
		size = 1
	}

	var payloadSize int64
	if size > 0 {
		switch d.code {
		case Ext8:
			s, err := rdr.ReadUint8()
			if err != nil {
				return nil, errors.Wrap(err, `msgpack: failed to read size for ext8 value`)
			}
			payloadSize = int64(s)
		case Ext16:
			s, err := rdr.ReadUint16()
			if err != nil {
				return nil, errors.Wrap(err, `msgpack: failed to read size for ext16 value`)
			}
			payloadSize = int64(s)
		case Ext32:
			s, err := rdr.ReadUint32()
			if err != nil {
				return nil, errors.Wrap(err, `msgpack: failed to read size for ext32 value`)
			}
			payloadSize = int64(s)
		default:
			return nil, errors.Errorf(`msgpack: unsupported ext %s`, d.code)
		}
	} else {
		switch d.code {
		case FixExt8:
			payloadSize = 8
		}
	}
	_ = payloadSize

	// lookup the Go type from Msgpack type
	b, err := rdr.ReadByte()
	if err != nil {
		return nil, errors.Wrap(err, `msgpack: failed to read type byte`)
	}
	exttyp := int(b)

	muExtDecode.RLock()
	typ, ok := extDecodeRegistry[exttyp]
	muExtDecode.RUnlock()

	if !ok {
		return nil, errors.Wrapf(err, `msgpack: failed to lookup msgpack type %d`, exttyp)
	}

	if reflect.PtrTo(typ).Implements(decodeMsgpackExterType) {
		rv := reflect.New(typ).Interface().(DecodeMsgpackExter)
		// At this point we delegate to the underlying object, but
		// we should limit reading to the payload size
		if err := rv.DecodeMsgpackExt(NewReader(io.LimitReader(r, payloadSize))); err != nil {
			return nil, errors.Wrap(err, `msgpack: failed to call DecodeMsgpackExt`)
		}

		return rv, nil
	}

	return nil, errors.Errorf(`msgpack: %s does not implement DecodeMsgpackExter`, typ)
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r: bufio.NewReader(r),
	}
}

func (d *Decoder) PeekCode() (Code, error) {
	b, err := d.r.ReadByte()
	if err != nil {
		return Code(0), errors.Wrap(err, `msgpack: failed to peek code`)
	}

	if err := d.r.UnreadByte(); err != nil {
		return Code(0), errors.Wrap(err, `msgpack: failed to unread code`)
	}
	return Code(b), nil
}

func (d *Decoder) DecodeNil() error {
	var v interface{}
	if err := d.Decode(&v); err != nil {
		return errors.Wrap(err, `failed to decode nil`)
	}
	return nil
}

func (d *Decoder) DecodeBool() (bool, error) {
	var v bool
	if err := d.Decode(&v); err != nil {
		return false, errors.Wrap(err, `failed to decode bool`)
	}
	return v, nil
}

func (d *Decoder) DecodeFloat32() (float32, error) {
	var v float32
	if err := d.Decode(&v); err != nil {
		return float32(0), errors.Wrap(err, `failed to decode float32`)
	}
	return v, nil
}

func (d *Decoder) DecodeFloat64() (float64, error) {
	var v float64
	if err := d.Decode(&v); err != nil {
		return float64(0), errors.Wrap(err, `failed to decode float64`)
	}
	return v, nil
}

func (d *Decoder) DecodeUint8() (uint8, error) {
	var v uint8
	if err := d.Decode(&v); err != nil {
		return uint8(0), errors.Wrap(err, `failed to decode uint8`)
	}
	return v, nil
}

func (d *Decoder) DecodeUint16() (uint16, error) {
	var v uint16
	if err := d.Decode(&v); err != nil {
		return uint16(0), errors.Wrap(err, `failed to decode uint16`)
	}
	return v, nil
}

func (d *Decoder) DecodeUint32() (uint32, error) {
	var v uint32
	if err := d.Decode(&v); err != nil {
		return uint32(0), errors.Wrap(err, `failed to decode uint32`)
	}
	return v, nil
}

func (d *Decoder) DecodeUint64() (uint64, error) {
	var v uint64
	if err := d.Decode(&v); err != nil {
		return uint64(0), errors.Wrap(err, `failed to decode uint64`)
	}
	return v, nil
}

func (d *Decoder) DecodeInt8() (int8, error) {
	var v int8
	if err := d.Decode(&v); err != nil {
		return int8(0), errors.Wrap(err, `failed to decode int8`)
	}
	return v, nil
}

func (d *Decoder) DecodeInt16() (int16, error) {
	var v int16
	if err := d.Decode(&v); err != nil {
		return int16(0), errors.Wrap(err, `failed to decode int16`)
	}
	return v, nil
}

func (d *Decoder) DecodeInt32() (int32, error) {
	var v int32
	if err := d.Decode(&v); err != nil {
		return int32(0), errors.Wrap(err, `failed to decode int32`)
	}
	return v, nil
}

func (d *Decoder) DecodeString() (string, error) {
	var v string
	if err := d.Decode(&v); err != nil {
		return "", errors.Wrap(err, `failed to decode string`)
	}
	return v, nil
}

func (d *Decoder) DecodeInt64() (int64, error) {
	var v int64
	if err := d.Decode(&v); err != nil {
		return int64(0), errors.Wrap(err, `failed to decode int64`)
	}
	return v, nil
}

func (d *Decoder) DecodeArray() ([]interface{}, error) {
	var v []interface{}
	if err := d.Decode(&v); err != nil {
		return nil, errors.Wrap(err, `msgpack: failed to decode array`)
	}
	return v, nil
}

func lookupDecoder(code Code) (valueDecoder, error) {
	dec, ok := decoders[code]
	if !ok {
		return nil, errors.Errorf(`msgpack: decoder for %s not found`, code)
	}
	return dec, nil
}

// Decode takes a pointer to a variable, and populates it with the value
// that was unmarshaled from the stream.
// If the variable is a non-pointer or nil, an error is returned.
func (d *Decoder) Decode(v interface{}) error {
	if dm, ok := v.(DecodeMsgpacker); ok {
		return dm.DecodeMsgpack(d)
	}

	decoded, err := d.decodeInterface(v)
	if err != nil {
		return errors.Wrap(err, `msgpack: failed to decode value`)
	}

	rv := reflect.ValueOf(v)

	// The result of decoding must be assigned to v, and v
	// should be a pointer
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
	if dv.Type().AssignableTo(dst.Type()) {
		dst.Set(dv)
		return nil
	}

	// Can we convert it then?
	if dv.Type().ConvertibleTo(dst.Type()) {
		dst.Set(dv.Convert(dst.Type()))
		return nil
	}

	// This could only happen if we have a decoder that creates
	// the value dynamically, such asin the case of struct
	// decoder or extension decoder.
	if reflect.PtrTo(dst.Type()) == dv.Type() {
		dst.Set(dv.Elem())
		return nil
	}

	return errors.Errorf(`msgpack: cannot assign %s to %s`, dv.Type(), dst.Type())
}

// Note: v is only used as a hint. do not assign in this method
func (d *Decoder) decodeInterface(v interface{}) (interface{}, error) {
	code, err := d.PeekCode()
	if err != nil {
		return nil, errors.Wrap(err, `msgpack: failed to peek code`)
	}
	d.r.ReadByte() // throw away code

	switch code {
	case Nil:
		return nil, nil
	case True:
		return true, nil
	case False:
		return false, nil
	}

	var dec valueDecoder
	// Special case: If the object is a Map type, and the target object
	// is a Struct, we do the struct decoding bit.
	if IsMapFamily(code) {
		rv := reflect.ValueOf(v)
		if rv.Type().Elem().Kind() == reflect.Struct {
			dec = &structDecoder{code: code, target: rv.Type().Elem()}
		}
	}

	if dec == nil {
		var err error
		dec, err = lookupDecoder(code)
		if err != nil {
			return nil, errors.Wrapf(err, `msgpack: failed to lookup decoder for code %s`, code)
		}
	}

	decoded, err := dec.Decode(d.r)
	if err != nil {
		return nil, errors.Wrap(err, `msgpack: failed to decode value`)
	}

	return decoded, nil
}

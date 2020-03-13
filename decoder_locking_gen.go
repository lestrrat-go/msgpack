package msgpack

// Auto-generated by internal/cmd/gendecoder/gendecoder.go. DO NOT EDIT!

import (
	"reflect"
	"time"
)

func (d *decoder) Decode(v interface{}) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.Decode(v)
}

func (d *decoder) DecodeArray(v interface{}) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.DecodeArray(v)
}

func (d *decoder) DecodeArrayLength(v *int) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.DecodeArrayLength(v)
}

func (d *decoder) DecodeBool(v *bool) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.DecodeBool(v)
}

func (d *decoder) DecodeBytes(v *[]byte) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.DecodeBytes(v)
}

func (d *decoder) DecodeExt(v DecodeMsgpacker) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.DecodeExt(v)
}

func (d *decoder) DecodeExtLength(v *int) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.DecodeExtLength(v)
}

func (d *decoder) DecodeExtType(v *reflect.Type) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.DecodeExtType(v)
}

func (d *decoder) DecodeMap(v *map[string]interface{}) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.DecodeMap(v)
}

func (d *decoder) DecodeMapLength(v *int) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.DecodeMapLength(v)
}

func (d *decoder) DecodeNil(v *interface{}) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.DecodeNil(v)
}

func (d *decoder) DecodeString(v *string) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.DecodeString(v)
}

func (d *decoder) DecodeStruct(v interface{}) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.DecodeStruct(v)
}

func (d *decoder) DecodeTime(v *time.Time) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.DecodeTime(v)
}

func (d *decoder) PeekCode() (Code, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.PeekCode()
}

func (d *decoder) ReadCode() (Code, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.ReadCode()
}

func (d *decoder) Reader() Reader {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.Reader()
}

func (d *decoder) DecodeInt(v *int) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.DecodeInt(v)
}

func (d *decoder) DecodeInt8(v *int8) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.DecodeInt8(v)
}

func (d *decoder) DecodeInt16(v *int16) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.DecodeInt16(v)
}

func (d *decoder) DecodeInt32(v *int32) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.DecodeInt32(v)
}

func (d *decoder) DecodeInt64(v *int64) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.DecodeInt64(v)
}

func (d *decoder) DecodeUint(v *uint) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.DecodeUint(v)
}

func (d *decoder) DecodeUint8(v *uint8) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.DecodeUint8(v)
}

func (d *decoder) DecodeUint16(v *uint16) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.DecodeUint16(v)
}

func (d *decoder) DecodeUint32(v *uint32) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.DecodeUint32(v)
}

func (d *decoder) DecodeUint64(v *uint64) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.DecodeUint64(v)
}

func (d *decoder) DecodeFloat32(v *float32) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.DecodeFloat32(v)
}

func (d *decoder) DecodeFloat64(v *float64) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nl.DecodeFloat64(v)
}
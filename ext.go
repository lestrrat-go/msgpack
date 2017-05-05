package msgpack

import (
	"reflect"
	"sync"
)

var muExtDecode sync.RWMutex
var muExtEncode sync.RWMutex
var extDecodeRegistry = make(map[int]reflect.Type)
var extEncodeRegistry = make(map[reflect.Type]int)

func RegisterExt(typ int, v interface{}) error {
	/*
		if _, ok := v.(Marshaler); !ok {
			return errors.New(`value does not implement msgpack.Marshaler`)
		}

		if _, ok := v.(Unmarshaler); !ok {
			return errors.New(`value does not implement msgpack.Unmarshaler`)
		}
	*/

	t := reflect.TypeOf(v)

	muExtDecode.Lock()
	extDecodeRegistry[typ] = t
	muExtDecode.Unlock()

	muExtEncode.Lock()
	extEncodeRegistry[t] = typ
	muExtEncode.Unlock()

	return nil
}

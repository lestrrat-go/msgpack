# go-msgpack

A Work-In-Progress `msgpack` serializer and deserializer

[![Build Status](https://travis-ci.org/lestrrat/go-msgpack.png?branch=master)](https://travis-ci.org/lestrrat/go-msgpack)

[![GoDoc](https://godoc.org/github.com/lestrrat/go-msgpack?status.svg)](https://godoc.org/github.com/lestrrat/go-msgpack)

# SYNOPSIS

```go
package msgpack_test

import (
  "fmt"
  "time"

  msgpack "github.com/lestrrat/go-msgpack"
  "github.com/pkg/errors"
)

type EventTime struct {
  time.Time
}

func init() {
  if err := msgpack.RegisterExt(0, EventTime{}); err != nil {
    panic(err)
  }
}

func (t *EventTime) DecodeMsgpackExt(r msgpack.Reader) error {
  sec, err := r.ReadUint32()
  if err != nil {
    return errors.Wrap(err, `failed to read uint32 from first 4 bytes`)
  }

  nsec, err := r.ReadUint32()
  if err != nil {
    return errors.Wrap(err, `failed to read uint32 from second 4 bytes`)
  }

  t.Time = time.Unix(int64(sec), int64(nsec)).UTC()
  return nil
}

func (t EventTime) EncodeMsgpackExt(w msgpack.Writer) error {
  if err := w.WriteUint32(uint32(t.Unix())); err != nil {
    return errors.Wrap(err, `failed to write EventTime seconds payload`)
  }

  if err := w.WriteUint32(uint32(t.Nanosecond())); err != nil {
    return errors.Wrap(err, `failed to write EventTime nanoseconds payload`)
  }

  return nil
}

func ExampleMsgpackExt_MarshalUnmarshal() {
  var t1 EventTime
  t1.Time = time.Unix(1234567890, 123).UTC()

  b, err := msgpack.Marshal(t1)
  if err != nil {
    fmt.Printf("%s\n", err)
    return
  }

  var t2 EventTime
  if err := msgpack.Unmarshal(b, &t2); err != nil {
    fmt.Printf("%s\n", err)
    return
  }

  fmt.Printf("%s\n", t2.UTC())
  // OUTPUT:
  // 2009-02-13 23:31:30.000000123 +0000 UTC
}

type FluentdMessage struct {
  Tag    string
  Time   EventTime
  Record map[string]interface{}
  Option interface{}
}

func (m FluentdMessage) EncodeMsgpack(e *msgpack.Encoder) error {
  if err := e.EncodeArrayHeader(4); err != nil {
    return errors.Wrap(err, `failed to encode array header`)
  }
  if err := e.EncodeString(m.Tag); err != nil {
    return errors.Wrap(err, `failed to encode tag`)
  }
  if err := e.Encode(m.Time); err != nil {
    return errors.Wrap(err, `failed to encode time`)
  }
  if err := e.Encode(m.Record); err != nil {
    return errors.Wrap(err, `failed to encode record`)
  }
  if err := e.Encode(m.Option); err != nil {
    return errors.Wrap(err, `failed to encode option`)
  }
  return nil
}

func (m *FluentdMessage) DecodeMsgpack(e *msgpack.Decoder) error {
  var l int
  if err := e.DecodeArrayLength(&l); err != nil {
    return errors.Wrap(err, `failed to decode msgpack array length`)
  }

  if l != 4 {
    return errors.Errorf(`invalid array length %d (expected 4)`, l)
  }

  if err := e.DecodeString(&m.Tag); err != nil {
    return errors.Wrap(err, `failed to decode fluentd message tag`)
  }

  if err := e.Decode(&m.Time); err != nil {
    return errors.Wrap(err, `failed to decode fluentd time`)
  }

  if err := e.Decode(&m.Record); err != nil {
    return errors.Wrap(err, `failed to decode fluentd record`)
  }

  if err := e.Decode(&m.Option); err != nil {
    return errors.Wrap(err, `failed to decode fluentd option`)
  }

  return nil
}

func ExampleFluentdMessage() {
  var f1 = FluentdMessage{
    Tag:  "foo",
    Time: EventTime{Time: time.Unix(1234567890, 123).UTC()},
    Record: map[string]interface{}{
      "count": 1000,
    },
  }

  b, err := msgpack.Marshal(f1)
  if err != nil {
    fmt.Printf("%s\n", err)
    return
  }

  var f2 FluentdMessage
  if err := msgpack.Unmarshal(b, &f2); err != nil {
    fmt.Printf("%s\n", err)
    return
  }

  fmt.Printf("%s %s %v %v\n", f2.Tag, f2.Time, f2.Record, f2.Option)
  // OUTPUT:
  // foo 2009-02-13 23:31:30.000000123 +0000 UTC map[count:1000] <nil>
}
# STATUS

* Requires more testing for array/map/struct types

# DESCRIPTION

While tinkering with low-level `msgpack` stuff for the first time,
I realized that I didn't know enough about its internal workings to make
suggestions of have confidence producing bug reports, and I really
should: So I wrote one for my own amusement and education.

# PROS/CONS

## PROS

As most late comers are, I believe the project is a little bit cleaner than my predecessors, which **possibly** could mean a slightly easier experience for the users to hack and tweak it. I know, it's very subjective.

As far as comparisons against `gopkg.in/vmihailenco/msgpack.v2` goes, this library tries to keep the API as compatible as possible to the standard library's `encoding/*` packages. For example, `encoding/json` allows:

```go
  b, _ := json.Marshal(true)

  // using uninitialized empty interface
  var v interface{}
  json.Unmarshal(b, &v)
```

But if you do the same with `gopkg.in/vmihailenco/msgpack.v2`, this throws a panic:

```go
  b, _ := msgpack.Marshal(true)

  // using uninitialized empty interface
  var v interface{}
  msgpack.Unmarshal(b, &v)
```

This library follows the semantics for `encoding/json`, and you can safely pass an uninitialized empty inteface to Unmarsha/Decode

## CONS

As previously described, I have been learning by implementing this library.
I intend to work on it until I'm satisfied, but unless you are the type of
person who likes to live on the bleeding edge, you probably want to use another library.

# BENCHMARKS

Current status

```
$ go test -run=none -tags bench -benchmem -bench .     
BenchmarkEncodeFloat32/___lestrrat/encode_float32_via_Encode()-4         	30000000	        56.8 ns/op	       4 B/op	       1 allocs/op
BenchmarkEncodeFloat32/___lestrrat/encode_float32_via_EncodeFloat32()-4  	50000000	        25.2 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeFloat32/vmihailenco/encode_float32_via_Encode()-4         	20000000	        60.2 ns/op	       4 B/op	       1 allocs/op
BenchmarkEncodeFloat32/vmihailenco/encode_float32_via_EncodeFloat32()-4  	50000000	        25.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeFloat64/___lestrrat/encode_float64_via_Encode()-4         	30000000	        54.6 ns/op	       8 B/op	       1 allocs/op
BenchmarkEncodeFloat64/___lestrrat/encode_float64_via_EncodeFloat64()-4  	50000000	        25.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeFloat64/vmihailenco/encode_float64_via_Encode()-4         	20000000	        61.6 ns/op	       8 B/op	       1 allocs/op
BenchmarkEncodeFloat64/vmihailenco/encode_float64_via_EncodeFloat64()-4  	50000000	        28.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeUint8/___lestrrat/encode_uint8_via_Encode()-4             	30000000	        48.1 ns/op	       1 B/op	       1 allocs/op
BenchmarkEncodeUint8/___lestrrat/encode_uint8_via_EncodeUint8()-4        	100000000	        23.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeUint8/vmihailenco/encode_uint8_via_Encode()-4             	10000000	       184 ns/op	       1 B/op	       1 allocs/op
BenchmarkEncodeUint8/vmihailenco/encode_uint8_via_EncodeUint8()-4        	50000000	        25.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeUint16/___lestrrat/encode_uint16_via_Encode()-4           	20000000	        52.0 ns/op	       2 B/op	       1 allocs/op
BenchmarkEncodeUint16/___lestrrat/encode_uint16_via_EncodeUint16()-4     	100000000	        22.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeUint16/vmihailenco/encode_uint16_via_Encode()-4           	10000000	       176 ns/op	       2 B/op	       1 allocs/op
BenchmarkEncodeUint16/vmihailenco/encode_uint16_via_EncodeUint16()-4     	50000000	        27.2 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeUint32/___lestrrat/encode_uint32_via_Encode()-4           	30000000	        53.3 ns/op	       4 B/op	       1 allocs/op
BenchmarkEncodeUint32/___lestrrat/encode_uint32_via_EncodeUint32()-4     	50000000	        26.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeUint32/vmihailenco/encode_uint32_via_Encode()-4           	10000000	       191 ns/op	       4 B/op	       1 allocs/op
BenchmarkEncodeUint32/vmihailenco/encode_uint32_via_EncodeUint32()-4     	50000000	        29.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeUint64/___lestrrat/encode_uint64_via_Encode()-4           	20000000	        58.3 ns/op	       8 B/op	       1 allocs/op
BenchmarkEncodeUint64/___lestrrat/encode_uint64_via_EncodeUint64()-4     	50000000	        27.0 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeUint64/vmihailenco/encode_uint64_via_Encode()-4           	20000000	        70.0 ns/op	       8 B/op	       1 allocs/op
BenchmarkEncodeUint64/vmihailenco/encode_uint64_via_EncodeUint64()-4     	50000000	        29.1 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeInt8/___lestrrat/encode_int8_via_Encode()-4               	30000000	        46.5 ns/op	       1 B/op	       1 allocs/op
BenchmarkEncodeInt8/___lestrrat/encode_int8_via_EncodeInt8()-4           	50000000	        25.0 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeInt8/vmihailenco/encode_int8_via_Encode()-4               	10000000	       181 ns/op	       2 B/op	       2 allocs/op
BenchmarkEncodeInt8/vmihailenco/encode_int8_via_EncodeInt8()-4           	50000000	        34.7 ns/op	       1 B/op	       1 allocs/op
BenchmarkEncodeInt16/___lestrrat/encode_int16_via_Encode()-4             	30000000	        50.3 ns/op	       2 B/op	       1 allocs/op
BenchmarkEncodeInt16/___lestrrat/encode_int16_via_EncodeInt16()-4        	100000000	        23.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeInt16/vmihailenco/encode_int16_via_Encode()-4             	10000000	       183 ns/op	       2 B/op	       1 allocs/op
BenchmarkEncodeInt16/vmihailenco/encode_int16_via_EncodeInt16()-4        	50000000	        31.0 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeInt32/___lestrrat/encode_int32_via_Encode()-4             	30000000	        51.6 ns/op	       4 B/op	       1 allocs/op
BenchmarkEncodeInt32/___lestrrat/encode_int32_via_EncodeInt32()-4        	50000000	        24.1 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeInt32/vmihailenco/encode_int32_via_Encode()-4             	10000000	       186 ns/op	       4 B/op	       1 allocs/op
BenchmarkEncodeInt32/vmihailenco/encode_int32_via_EncodeInt32()-4        	50000000	        31.2 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeInt64/___lestrrat/encode_int64_via_Encode()-4             	30000000	        55.7 ns/op	       8 B/op	       1 allocs/op
BenchmarkEncodeInt64/___lestrrat/encode_int64_via_EncodeInt64()-4        	50000000	        26.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeInt64/vmihailenco/encode_int64_via_Encode()-4             	20000000	        67.3 ns/op	       8 B/op	       1 allocs/op
BenchmarkEncodeInt64/vmihailenco/encode_int64_via_EncodeInt64()-4        	50000000	        33.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeString/___lestrrat/encode_string_(255_bytes)_via_Encode()-4         	10000000	       224 ns/op	     272 B/op	       2 allocs/op
BenchmarkEncodeString/___lestrrat/encode_string_(255_bytes)_via_EncodeString()-4   	10000000	       164 ns/op	     256 B/op	       1 allocs/op
BenchmarkEncodeString/___lestrrat/encode_string_(256_bytes)_via_Encode()-4         	10000000	       207 ns/op	     272 B/op	       2 allocs/op
BenchmarkEncodeString/___lestrrat/encode_string_(256_bytes)_via_EncodeString()-4   	10000000	       161 ns/op	     256 B/op	       1 allocs/op
BenchmarkEncodeString/___lestrrat/encode_string_(65536_bytes)_via_Encode()-4       	  100000	     14252 ns/op	   65552 B/op	       2 allocs/op
BenchmarkEncodeString/___lestrrat/encode_string_(65536_bytes)_via_EncodeString()-4 	  100000	     14646 ns/op	   65536 B/op	       1 allocs/op
BenchmarkEncodeString/vmihailenco/encode_string_(255_bytes)_via_Encode()-4         	10000000	       218 ns/op	     272 B/op	       2 allocs/op
BenchmarkEncodeString/vmihailenco/encode_string_(255_bytes)_via_EncodeString()-4   	10000000	       142 ns/op	     256 B/op	       1 allocs/op
BenchmarkEncodeString/vmihailenco/encode_string_(256_bytes)_via_Encode()-4         	10000000	       199 ns/op	     272 B/op	       2 allocs/op
BenchmarkEncodeString/vmihailenco/encode_string_(256_bytes)_via_EncodeString()-4   	10000000	       138 ns/op	     256 B/op	       1 allocs/op
BenchmarkEncodeString/vmihailenco/encode_string_(65536_bytes)_via_Encode()-4       	  100000	     14009 ns/op	   65552 B/op	       2 allocs/op
BenchmarkEncodeString/vmihailenco/encode_string_(65536_bytes)_via_EncodeString()-4 	  100000	     13645 ns/op	   65536 B/op	       1 allocs/op
BenchmarkDecodeUint8/___lestrrat/decode_uint8_via_DecodeUint8()-4                  	30000000	        39.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkDecodeUint8/vmihailenco/decode_uint8_via_DecodeUint8()_(return)-4         	30000000	        49.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkDecodeUint16/___lestrrat/decode_uint16_via_DecodeUint16()-4               	30000000	        41.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkDecodeUint16/vmihailenco/decode_uint16_via_DecodeUint16()_(return)-4      	20000000	        90.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkDecodeUint32/___lestrrat/decode_uint32_via_DecodeUint32()-4               	30000000	        39.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkDecodeUint32/vmihailenco/decode_uint32_via_DecodeUint32()_(return)-4      	20000000	        92.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkDecodeUint64/___lestrrat/decode_uint64_via_DecodeUint64()-4               	30000000	        39.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkDecodeUint64/vmihailenco/decode_uint64_via_DecodeUint64()_(return)-4      	20000000	        89.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkDecodeInt8/___lestrrat/decode_int8_via_DecodeInt8()-4                     	50000000	        38.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkDecodeInt8/vmihailenco/decode_int8_via_DecodeInt8()_(return)-4            	30000000	        52.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkDecodeInt16/___lestrrat/decode_int16_via_DecodeInt16()-4                  	30000000	        39.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkDecodeInt16/vmihailenco/decode_int16_via_DecodeInt16()_(return)-4         	20000000	        93.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkDecodeInt32/___lestrrat/decode_int32_via_DecodeInt32()-4                  	30000000	        38.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkDecodeInt32/vmihailenco/decode_int32_via_DecodeInt32()_(return)-4         	20000000	        90.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkDecodeInt64/___lestrrat/decode_int64_via_DecodeInt64()-4                  	30000000	        38.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkDecodeInt64/vmihailenco/decode_int64_via_DecodeInt64()_(return)-4         	20000000	        87.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkDecodeFloat32/___lestrrat/decode_float32_via_DecodeFloat32()-4            	30000000	        38.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkDecodeFloat32/vmihailenco/decode_float32_via_DecodeFloat32()_(return)-4   	20000000	        84.1 ns/op	       0 B/op	       0 allocs/op
BenchmarkDecodeFloat64/___lestrrat/decode_float64_via_DecodeFloat64()-4            	30000000	        41.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkDecodeFloat64/vmihailenco/decode_float64_via_DecodeFloat64()_(return)-4   	20000000	        91.5 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/lestrrat/go-msgpack	118.435s
```

# ACKNOWLEDGEMENTS

Much has been stolen from https://github.com/vmihailenco/msgpack
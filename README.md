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
go test -run=none -tags=bench -benchmem -bench Encode    
BenchmarkEncodeFloat32/___lestrrat/encode_float32_via_Encode()-4         	30000000	        49.2 ns/op	       4 B/op	       1 allocs/op
BenchmarkEncodeFloat32/___lestrrat/encode_float32_via_EncodeFloat32()-4  	100000000	        22.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeFloat32/vmihailenco/encode_float32_via_Encode()-4         	20000000	        55.2 ns/op	       4 B/op	       1 allocs/op
BenchmarkEncodeFloat32/vmihailenco/encode_float32_via_EncodeFloat32()-4  	100000000	        22.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeFloat64/___lestrrat/encode_float64_via_Encode()-4         	30000000	        50.3 ns/op	       8 B/op	       1 allocs/op
BenchmarkEncodeFloat64/___lestrrat/encode_float64_via_EncodeFloat64()-4  	100000000	        23.2 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeFloat64/vmihailenco/encode_float64_via_Encode()-4         	20000000	        59.9 ns/op	       8 B/op	       1 allocs/op
BenchmarkEncodeFloat64/vmihailenco/encode_float64_via_EncodeFloat64()-4  	50000000	        25.2 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeUint8/___lestrrat/encode_uint8_via_Encode()-4             	30000000	        42.3 ns/op	       1 B/op	       1 allocs/op
BenchmarkEncodeUint8/___lestrrat/encode_uint8_via_EncodeUint8()-4        	100000000	        21.1 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeUint8/vmihailenco/encode_uint8_via_Encode()-4             	10000000	       158 ns/op	       1 B/op	       1 allocs/op
BenchmarkEncodeUint8/vmihailenco/encode_uint8_via_EncodeUint8()-4        	100000000	        23.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeUint16/___lestrrat/encode_uint16_via_Encode()-4           	30000000	        43.8 ns/op	       2 B/op	       1 allocs/op
BenchmarkEncodeUint16/___lestrrat/encode_uint16_via_EncodeUint16()-4     	100000000	        20.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeUint16/vmihailenco/encode_uint16_via_Encode()-4           	10000000	       160 ns/op	       2 B/op	       1 allocs/op
BenchmarkEncodeUint16/vmihailenco/encode_uint16_via_EncodeUint16()-4     	50000000	        23.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeUint32/___lestrrat/encode_uint32_via_Encode()-4           	30000000	        47.0 ns/op	       4 B/op	       1 allocs/op
BenchmarkEncodeUint32/___lestrrat/encode_uint32_via_EncodeUint32()-4     	100000000	        21.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeUint32/vmihailenco/encode_uint32_via_Encode()-4           	10000000	       165 ns/op	       4 B/op	       1 allocs/op
BenchmarkEncodeUint32/vmihailenco/encode_uint32_via_EncodeUint32()-4     	50000000	        25.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeUint64/___lestrrat/encode_uint64_via_Encode()-4           	30000000	        48.6 ns/op	       8 B/op	       1 allocs/op
BenchmarkEncodeUint64/___lestrrat/encode_uint64_via_EncodeUint64()-4     	100000000	        23.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeUint64/vmihailenco/encode_uint64_via_Encode()-4           	20000000	        56.8 ns/op	       8 B/op	       1 allocs/op
BenchmarkEncodeUint64/vmihailenco/encode_uint64_via_EncodeUint64()-4     	100000000	        23.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeInt8/___lestrrat/encode_int8_via_Encode()-4               	30000000	        43.7 ns/op	       1 B/op	       1 allocs/op
BenchmarkEncodeInt8/___lestrrat/encode_int8_via_EncodeInt8()-4           	100000000	        21.0 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeInt8/vmihailenco/encode_int8_via_Encode()-4               	10000000	       169 ns/op	       2 B/op	       2 allocs/op
BenchmarkEncodeInt8/vmihailenco/encode_int8_via_EncodeInt8()-4           	50000000	        31.4 ns/op	       1 B/op	       1 allocs/op
BenchmarkEncodeInt16/___lestrrat/encode_int16_via_Encode()-4             	30000000	        47.4 ns/op	       2 B/op	       1 allocs/op
BenchmarkEncodeInt16/___lestrrat/encode_int16_via_EncodeInt16()-4        	100000000	        21.1 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeInt16/vmihailenco/encode_int16_via_Encode()-4             	10000000	       164 ns/op	       2 B/op	       1 allocs/op
BenchmarkEncodeInt16/vmihailenco/encode_int16_via_EncodeInt16()-4        	50000000	        26.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeInt32/___lestrrat/encode_int32_via_Encode()-4             	30000000	        47.6 ns/op	       4 B/op	       1 allocs/op
BenchmarkEncodeInt32/___lestrrat/encode_int32_via_EncodeInt32()-4        	100000000	        21.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeInt32/vmihailenco/encode_int32_via_Encode()-4             	10000000	       169 ns/op	       4 B/op	       1 allocs/op
BenchmarkEncodeInt32/vmihailenco/encode_int32_via_EncodeInt32()-4        	50000000	        28.2 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeInt64/___lestrrat/encode_int64_via_Encode()-4             	30000000	        50.2 ns/op	       8 B/op	       1 allocs/op
BenchmarkEncodeInt64/___lestrrat/encode_int64_via_EncodeInt64()-4        	50000000	        23.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkEncodeInt64/vmihailenco/encode_int64_via_Encode()-4             	20000000	        62.2 ns/op	       8 B/op	       1 allocs/op
BenchmarkEncodeInt64/vmihailenco/encode_int64_via_EncodeInt64()-4        	50000000	        30.9 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/lestrrat/go-msgpack	69.471s
```

# ACKNOWLEDGEMENTS

Much has been stolen from https://github.com/vmihailenco/msgpack
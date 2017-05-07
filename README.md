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
  return e.EncodeArray([]interface{}{
    m.Tag,
    m.Time,
    m.Record,
    m.Option,
  })
}

func (m *FluentdMessage) DecodeMsgpack(e *msgpack.Decoder) error {
  l, err := e.DecodeArray()
  if err != nil {
    return errors.Wrap(err, `failed to decode msgpack array`)
  }
  m.Tag = l[0].(string)
  m.Time = *(l[1].(*EventTime))
  m.Record = l[2].(map[string]interface{})
  m.Option = l[3]
  return nil
}

func ExampleFluentdMessage() {
  var f1 = FluentdMessage{
    Tag:    "foo",
    Time:   EventTime{Time: time.Unix(1234567890, 123).UTC()},
    Record: map[string]interface{}{
      "count": 100,
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
  // foo 2009-02-13 23:31:30.000000123 +0000 UTC map[count:100] <nil>
}
```

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
$ go test -run=none -tags "bench" -bench=. -benchmem -v
BenchmarkLestrrat/encode_nil-4                       50000000          29.8 ns/op         0 B/op         0 allocs/op
BenchmarkLestrrat/encode_true-4                     100000000          24.4 ns/op         0 B/op         0 allocs/op
BenchmarkLestrrat/encode_false-4                    100000000          23.1 ns/op         0 B/op         0 allocs/op
BenchmarkLestrrat/encode_string_(len=16)-4           20000000          77.6 ns/op        16 B/op         1 allocs/op
BenchmarkLestrrat/encode_string_(len=256)-4          10000000           156 ns/op       256 B/op         1 allocs/op
BenchmarkLestrrat/encode_string_(len=65536)-4          100000         15584 ns/op     65536 B/op         1 allocs/op
BenchmarkLestrrat/encode_float32-4                   30000000          42.4 ns/op         0 B/op         0 allocs/op
BenchmarkLestrrat/encode_float64-4                   30000000          39.1 ns/op         0 B/op         0 allocs/op
BenchmarkLestrrat/encode_uint8-4                     30000000          37.1 ns/op         0 B/op         0 allocs/op
BenchmarkLestrrat/encode_uint16-4                    30000000          41.5 ns/op         0 B/op         0 allocs/op
BenchmarkLestrrat/encode_uint32-4                    30000000          37.0 ns/op         0 B/op         0 allocs/op
BenchmarkLestrrat/encode_uint64-4                    50000000          37.2 ns/op         0 B/op         0 allocs/op
BenchmarkLestrrat/encode_int8-4                      30000000          39.3 ns/op         0 B/op         0 allocs/op
BenchmarkLestrrat/encode_int16-4                     30000000          41.1 ns/op         0 B/op         0 allocs/op
BenchmarkLestrrat/encode_int32-4                     30000000          38.6 ns/op         0 B/op         0 allocs/op
BenchmarkLestrrat/encode_int64-4                     30000000          36.9 ns/op         0 B/op         0 allocs/op
BenchmarkLestrrat/decode_nil-4                        1000000          1018 ns/op        16 B/op         1 allocs/op
BenchmarkLestrrat/marshal_float32-4                  10000000           187 ns/op        64 B/op         3 allocs/op
BenchmarkLestrrat/marshal_float64-4                  10000000           165 ns/op        64 B/op         3 allocs/op
BenchmarkLestrrat/marshal_uint8-4                    10000000           162 ns/op        64 B/op         3 allocs/op
BenchmarkLestrrat/marshal_uint16-4                   10000000           157 ns/op        64 B/op         3 allocs/op
BenchmarkLestrrat/marshal_uint32-4                   10000000           152 ns/op        64 B/op         3 allocs/op
BenchmarkLestrrat/marshal_uint64-4                   10000000           157 ns/op        64 B/op         3 allocs/op
BenchmarkLestrrat/marshal_int8-4                     10000000           147 ns/op        64 B/op         3 allocs/op
BenchmarkLestrrat/marshal_int16-4                    10000000           143 ns/op        64 B/op         3 allocs/op
BenchmarkLestrrat/marshal_int32-4                    10000000           142 ns/op        64 B/op         3 allocs/op
BenchmarkLestrrat/marshal_int64-4                    10000000           143 ns/op        64 B/op         3 allocs/op
BenchmarkVmihailenco/encode_nil-4                    50000000          33.5 ns/op         1 B/op         1 allocs/op
BenchmarkVmihailenco/encode_true-4                   50000000          35.5 ns/op         1 B/op         1 allocs/op
BenchmarkVmihailenco/encode_false-4                  50000000          34.6 ns/op         1 B/op         1 allocs/op
BenchmarkVmihailenco/encode_string_(len=16)-4        20000000          87.5 ns/op        17 B/op         2 allocs/op
BenchmarkVmihailenco/encode_string_(len=256)-4       10000000           132 ns/op       256 B/op         1 allocs/op
BenchmarkVmihailenco/encode_string_(len=65536)-4       100000         13289 ns/op     65536 B/op         1 allocs/op
BenchmarkVmihailenco/encode_float32-4                50000000          33.5 ns/op         0 B/op         0 allocs/op
BenchmarkVmihailenco/encode_float64-4                50000000          33.6 ns/op         0 B/op         0 allocs/op
BenchmarkVmihailenco/encode_uint8-4                  50000000          33.1 ns/op         0 B/op         0 allocs/op
BenchmarkVmihailenco/encode_uint16-4                 50000000          31.8 ns/op         0 B/op         0 allocs/op
BenchmarkVmihailenco/encode_uint32-4                 50000000          33.3 ns/op         0 B/op         0 allocs/op
BenchmarkVmihailenco/encode_uint64-4                 50000000          33.3 ns/op         0 B/op         0 allocs/op
BenchmarkVmihailenco/encode_int8-4                   30000000          37.5 ns/op         1 B/op         1 allocs/op
BenchmarkVmihailenco/encode_int16-4                  50000000          32.2 ns/op         0 B/op         0 allocs/op
BenchmarkVmihailenco/encode_int32-4                  50000000          33.4 ns/op         0 B/op         0 allocs/op
BenchmarkVmihailenco/encode_int64-4                  30000000          36.0 ns/op         0 B/op         0 allocs/op
BenchmarkVmihailenco/decode_nil-4                     2000000           710 ns/op         0 B/op         0 allocs/op
BenchmarkVmihailenco/marshal_float32-4                5000000           225 ns/op       176 B/op         3 allocs/op
BenchmarkVmihailenco/marshal_float64-4               10000000           220 ns/op       176 B/op         3 allocs/op
BenchmarkVmihailenco/marshal_uint8-4                 10000000           192 ns/op       176 B/op         3 allocs/op
BenchmarkVmihailenco/marshal_uint16-4                10000000           195 ns/op       176 B/op         3 allocs/op
BenchmarkVmihailenco/marshal_uint32-4                10000000           198 ns/op       176 B/op         3 allocs/op
BenchmarkVmihailenco/marshal_uint64-4                10000000           227 ns/op       176 B/op         3 allocs/op
BenchmarkVmihailenco/marshal_int8-4                  10000000           179 ns/op       176 B/op         3 allocs/op
BenchmarkVmihailenco/marshal_int16-4                 10000000           198 ns/op       176 B/op         3 allocs/op
BenchmarkVmihailenco/marshal_int32-4                 10000000           196 ns/op       176 B/op         3 allocs/op
BenchmarkVmihailenco/marshal_int64-4                 10000000           196 ns/op       176 B/op         3 allocs/op
PASS
ok    github.com/lestrrat/go-msgpack  245.376s
```

# ACKNOWLEDGEMENTS

Much has been stolen from https://github.com/vmihailenco/msgpack
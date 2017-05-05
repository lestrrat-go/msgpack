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
		return errors.Wrap(err, `failed to read uint32 from first 32 bytes`)
	}

	nsec, err := r.ReadUint32()
	if err != nil {
		return errors.Wrap(err, `failed to read uint32 from second 32 bytes`)
	}

	t.Time = time.Unix(int64(sec), int64(nsec))
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
	t1.Time = time.Unix(1234567890, 123)

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
		Time:   EventTime{Time: time.Unix(1234567890, 123)},
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
	// foo 2009-02-14 08:31:30.000000123 +0900 JST map[count:100] <nil>
}

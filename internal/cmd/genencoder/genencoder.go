package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"sort"

	"github.com/lestrrat-go/msgpack/internal/util"
	"github.com/pkg/errors"
)

type argument struct {
	name string
	typ  string
}

type numericType struct {
	Code string
	Bits int
}

var floatTypes = map[reflect.Kind]numericType{
	reflect.Float32: {Code: "Float", Bits: 32},
	reflect.Float64: {Code: "Double", Bits: 64},
}

var integerTypes = map[reflect.Kind]numericType{
	reflect.Int:    {Code: "Int64", Bits: 64},
	reflect.Int8:   {Code: "Int8", Bits: 8},
	reflect.Int16:  {Code: "Int16", Bits: 16},
	reflect.Int32:  {Code: "Int32", Bits: 32},
	reflect.Int64:  {Code: "Int64", Bits: 64},
	reflect.Uint:   {Code: "Uint64", Bits: 64},
	reflect.Uint8:  {Code: "Uint8", Bits: 8},
	reflect.Uint16: {Code: "Uint16", Bits: 16},
	reflect.Uint32: {Code: "Uint32", Bits: 32},
	reflect.Uint64: {Code: "Uint64", Bits: 64},
}

func main() {
	if err := _main(); err != nil {
		log.Printf("%s", err)
		os.Exit(1)
	}
}

func _main() error {
	if err := generateNumericEncoders(); err != nil {
		return errors.Wrap(err, `failed to generate numeric encoders`)
	}
	if err := generateLockingWrappers(); err != nil {
		return errors.Wrap(err, `failed to generate locking wrappers`)
	}
	return nil
}

func writeLockingWrapper(dst io.Writer, methodName string, args []argument, rets []string) error {
	fmt.Fprintf(dst, "\n\nfunc (d *encoder) %s(", methodName)
	for i, arg := range args {
		fmt.Fprintf(dst, "%s %s", arg.name, arg.typ)
		if i != len(args)-1 {
			fmt.Fprint(dst, ", ")
		}
	}
	fmt.Fprint(dst, ") ")

	if len(rets) > 1 {
		fmt.Fprint(dst, "(")
	}
	for i, ret := range rets {
		fmt.Fprint(dst, ret)
		if i != len(rets)-1 {
			fmt.Fprint(dst, ", ")
		}
	}
	if len(rets) > 1 {
		fmt.Fprint(dst, ")")
	}
	fmt.Fprint(dst, " {")
	fmt.Fprintf(dst, "\nd.mu.RLock()")
	fmt.Fprintf(dst, "\ndefer d.mu.RUnlock()")
	fmt.Fprintf(dst, "\nreturn d.nl.%s(", methodName)
	for i, arg := range args {
		fmt.Fprint(dst, arg.name)
		if i != len(args)-1 {
			fmt.Fprint(dst, ", ")
		}
	}
	fmt.Fprint(dst, ")")

	fmt.Fprintf(dst, "\n}")
	return nil
}

func generateLockingWrappers() error {
	var dst = &bytes.Buffer{}

	dst.WriteString("package msgpack")
	dst.WriteString("\n\n// Auto-generated by internal/cmd/genencoder/genencoder.go. DO NOT EDIT!")
	dst.WriteString("\n\nimport (")
	//	dst.WriteString("\n\"reflect\"")
	dst.WriteString("\n\"time\"")
	dst.WriteString("\n)")

	wrappers := []struct {
		name string
		args []argument
		rets []string
	}{
		{
			name: "Encode",
			args: []argument{
				{name: "v", typ: "interface{}"},
			},
			rets: []string{"error"},
		},
		{
			name: "EncodeArray",
			args: []argument{
				{name: "v", typ: "interface{}"},
			},
			rets: []string{"error"},
		},
		{
			name: "EncodeArrayHeader",
			args: []argument{
				{name: "v", typ: "int"},
			},
			rets: []string{"error"},
		},
		{
			name: "EncodeBool",
			args: []argument{
				{name: "v", typ: "bool"},
			},
			rets: []string{"error"},
		},
		{
			name: "EncodeBytes",
			args: []argument{
				{name: "v", typ: "[]byte"},
			},
			rets: []string{"error"},
		},
		{
			name: "EncodeExt",
			args: []argument{
				{name: "v", typ: "EncodeMsgpacker"},
			},
			rets: []string{"error"},
		},
		{
			name: "EncodeExtHeader",
			args: []argument{
				{name: "v", typ: "int"},
			},
			rets: []string{"error"},
		},
		{
			name: "EncodeExtType",
			args: []argument{
				{name: "v", typ: "EncodeMsgpacker"},
			},
			rets: []string{"error"},
		},
		{
			name: "EncodeMap",
			args: []argument{
				{name: "v", typ: "interface{}"},
			},
			rets: []string{"error"},
		},
		{
			name: "EncodeNegativeFixNum",
			args: []argument{
				{name: "v", typ: "int8"},
			},
			rets: []string{"error"},
		},
		{
			name: "EncodeNil",
			rets: []string{"error"},
		},
		{
			name: "EncodePositiveFixNum",
			args: []argument{
				{name: "v", typ: "uint8"},
			},
			rets: []string{"error"},
		},
		{
			name: "EncodeString",
			args: []argument{
				{name: "v", typ: "string"},
			},
			rets: []string{"error"},
		},
		{
			name: "EncodeStruct",
			args: []argument{
				{name: "v", typ: "interface{}"},
			},
			rets: []string{"error"},
		},
		{
			name: "EncodeTime",
			args: []argument{
				{name: "v", typ: "time.Time"},
			},
			rets: []string{"error"},
		},
		{
			name: "Writer",
			rets: []string{"Writer"},
		},
	}

	for _, w := range wrappers {
		writeLockingWrapper(dst, w.name, w.args, w.rets)
	}

	// numeric stuff
	keys := make([]reflect.Kind, 0, len(integerTypes)+len(floatTypes))
	for k := range integerTypes {
		keys = append(keys, k)
	}
	for k := range floatTypes {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return uint(keys[i]) < uint(keys[j])
	})

	for _, typ := range keys {
		writeLockingWrapper(
			dst,
			fmt.Sprintf("Encode%s", util.Ucfirst(typ.String())),
			[]argument{
				{name: "v", typ: typ.String()},
			},
			[]string{"error"},
		)
	}

	if err := util.WriteFormattedFile("encoder_locking_gen.go", dst.Bytes()); err != nil {
		return errors.Wrap(err, `failed to write to file`)
	}
	return nil
}

func generateNumericEncoders() error {
	var buf bytes.Buffer

	buf.WriteString("package msgpack")
	buf.WriteString("\n\n// Auto-generated by internal/cmd/genencoder/genencoder.go. DO NOT EDIT!")
	buf.WriteString("\n\nimport (")
	buf.WriteString("\n\"math\"")
	buf.WriteString("\n\n\"github.com/pkg/errors\"")
	buf.WriteString("\n)")

	if err := generateIntegerTypes(&buf); err != nil {
		return errors.Wrap(err, `failed to generate integer encoders`)
	}

	if err := generateFloatTypes(&buf); err != nil {
		return errors.Wrap(err, `failed to generate float encoders`)
	}

	if err := util.WriteFormattedFile("encoder_numeric_gen.go", buf.Bytes()); err != nil {
		return errors.Wrap(err, `failed to write to file`)
	}
	return nil
}

func generateIntegerTypes(dst io.Writer) error {
	types := integerTypes

	keys := make([]reflect.Kind, 0, len(types))
	for k := range types {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return uint(keys[i]) < uint(keys[j])
	})
	for _, typ := range keys {
		data := types[typ]
		fmt.Fprintf(dst, "\n\nfunc (e *encoderNL) Encode%s(v %s) error {", util.Ucfirst(typ.String()), typ)
		switch typ {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			fmt.Fprintf(dst, "\nif inPositiveFixNumRange(int64(v)) {")
			fmt.Fprintf(dst, "\nreturn e.encodePositiveFixNum(uint8(0xff & v))")
			fmt.Fprintf(dst, "\n}")
		case reflect.Int8:
			fmt.Fprintf(dst, "\nif inNegativeFixNumRange(int64(v)) {")
			fmt.Fprintf(dst, "\nreturn e.encodeNegativeFixNum(v)")
			fmt.Fprintf(dst, "\n}")
		default:
			fmt.Fprintf(dst, "\nif inNegativeFixNumRange(int64(v)) {")
			fmt.Fprintf(dst, "\nreturn e.encodeNegativeFixNum(int8(byte(0xff &v)))")
			fmt.Fprintf(dst, "\n}")
		}

		fmt.Fprintf(dst, "\n\nif err := e.dst.WriteByteUint%d(%s.Byte(), ", data.Bits, data.Code)
		switch typ {
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			fmt.Fprintf(dst, "v")
		default:
			fmt.Fprintf(dst, "uint%d(v)", data.Bits)
		}
		fmt.Fprintf(dst, "); err != nil {")
		fmt.Fprintf(dst, "\nreturn errors.Wrap(err, `msgpack: failed to write %s`)", data.Code)
		fmt.Fprintf(dst, "\n}")
		fmt.Fprintf(dst, "\nreturn nil")
		fmt.Fprintf(dst, "\n}")
	}
	return nil
}

func generateFloatTypes(dst io.Writer) error {
	types := floatTypes
	keys := make([]reflect.Kind, 0, len(types))
	for k := range types {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return uint(keys[i]) < uint(keys[j])
	})
	for _, typ := range keys {
		data := types[typ]
		fmt.Fprintf(dst, "\n\nfunc (e *encoderNL) EncodeFloat%d(f float%d) error {", data.Bits, data.Bits)
		fmt.Fprintf(dst, "\nif err := e.dst.WriteByteUint%d(%s.Byte(), math.Float%dbits(f)); err != nil {", data.Bits, data.Code, data.Bits)
		fmt.Fprintf(dst, "\nreturn errors.Wrap(err, `msgpack: failed to write %s`)", data.Code)
		fmt.Fprintf(dst, "\n}")
		fmt.Fprintf(dst, "\nreturn nil")
		fmt.Fprintf(dst, "\n}")
	}
	return nil
}

// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Copyright 2015-2016, Cyrill @ Schumacher.fm and the CoreStore contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"fmt"
	"strconv"
	"time"

	"github.com/corestoreio/errors"
	"github.com/corestoreio/pkg/util/bufferpool"
)

type fieldType uint8

// Type* constants define all available types which a field can contain.
const (
	typeBool fieldType = iota + 1
	typeInt
	typeInts
	typeInt64
	typeInt64s
	typeUint64
	typeFloat64
	typeString
	typeStrings
	typeStringer
	typeStringFn
	typeGoStringer
	typeObject
	typeObjectTypeOf
	typeMarshaler
	typeFields
)

// textMarshaler a copy of encoding.TextMarshaler
type textMarshaler interface {
	MarshalText() (text []byte, err error)
}

// jsonMarshaler is the interface implemented by types that
// can marshal themselves into valid JSON.
type jsonMarshaler interface {
	MarshalJSON() ([]byte, error)
}

// Marshaler allows user-defined types to efficiently add themselves to the
// logging context, and to selectively omit information which shouldn't be
// included in logs (e.g., passwords).
// Compatible to github.com/uber-go/zap
type Marshaler interface {
	MarshalLog(KeyValuer) error
}

// KeyValuer is an encoding-agnostic interface to add structured data to the
// logging context. Like maps, KeyValues aren't safe for concurrent use (though
// typical use shouldn't require locks).
//
// Compatible to github.com/uber-go/zap
type KeyValuer interface {
	AddBool(string, bool)
	AddFloat64(string, float64)
	AddInt(string, int)
	AddInt64(string, int64)
	AddUint64(string, uint64)
	AddMarshaler(string, Marshaler) error
	// AddObject uses reflection to serialize arbitrary objects, so it's slow and
	// allocation-heavy. Consider implementing the LogMarshaler interface instead.
	AddObject(string, interface{})
	AddString(string, string)
	Nest(string, func(KeyValuer) error) error
}

// AddStringFn same as KeyValuer.AddString to allow creating 3rd party log packages
// which can log very different types for which we do not want to create a
// Marshaler.
type AddStringFn func(string, string)

// Field is a deferred marshaling operation used to add a key-value pair to a
// logger's context. Keys and values are appropriately escaped for the current
// encoding scheme (e.g., JSON).
type Field interface {
	// make creates a new non-pointer field. If it would be a pointer this
	// function name would be new ;-).
	make() field
	// AddTo encodes a field to a string
	AddTo(kv KeyValuer) error
}

// Fields a slice of n Field types. Fields implements the Field interface and
// can be added to a function in the Logger interface in a simple way.
type Fields []Field

func (fs Fields) make() field {
	return field{fieldType: typeFields, obj: fs}
}

// AddTo adds all fields within this slice to a KeyValue encoder.
// Breaks on first error.
func (fs Fields) AddTo(kv KeyValuer) error {
	for _, f := range fs {
		if err := f.AddTo(kv); err != nil {
			return errors.Wrap(err, "[log] Fields.AddTo")
		}
	}
	return nil
}

// MarshalLog satisfies the interface of log.LogMarshaler
func (fs Fields) MarshalLog(kv KeyValuer) error {
	return errors.Wrap(fs.AddTo(kv), "[log] Fields.Marshalog")
}

// ToString transforms multiple fields into a single string using the
// format of the type KVStringify.
func (fs Fields) ToString(msg string) string {
	buf := bufferpool.Get()
	defer bufferpool.Put(buf)
	wt := WriteTypes{W: buf}

	_, _ = buf.WriteString(msg)
	if err := fs.AddTo(wt); err != nil {
		_, _ = buf.WriteString(Separator)
		_, _ = buf.WriteString(KeyNameError)
		_, _ = buf.WriteString(AssignmentChar)
		_, _ = buf.WriteString(fmt.Sprintf("%+v", err))
	}
	_, _ = buf.WriteRune('\n')
	return buf.String()
}

// field is a deferred marshaling operation used to add a key-value pair to
// a logger's context. Keys and values are appropriately escaped for the current
// encoding scheme (e.g., JSON).
type field struct {
	key string
	// fieldType specifies the used type. If 0 this struct is empty
	fieldType
	int64
	uint64
	float64
	string
	strFn func(AddStringFn) error
	obj   interface{}
}

func (f field) make() field {
	return f
}

// AddTo adds a field to KeyValue encoder
func (f field) AddTo(kv KeyValuer) error {
	switch f.fieldType {
	case typeBool:
		kv.AddBool(f.key, f.int64 == 1)
	case typeFloat64:
		kv.AddFloat64(f.key, f.float64)
	case typeInt:
		kv.AddInt(f.key, int(f.int64))
	case typeInts:
		buf := bufferpool.Get()
		vals := f.obj.([]int)
		for i, v := range vals {
			_, _ = buf.WriteString(strconv.Itoa(v))
			if i < len(vals)-1 {
				_, _ = buf.WriteString(", ")
			}
		}
		kv.AddString(f.key, buf.String())
		bufferpool.Put(buf)
	case typeInt64:
		kv.AddInt64(f.key, f.int64)
	case typeUint64:
		kv.AddUint64(f.key, f.uint64)
	case typeInt64s:
		buf := bufferpool.Get()
		vals := f.obj.([]int64)
		for i, v := range vals {
			_, _ = buf.WriteString(strconv.FormatInt(v, 10))
			if i < len(vals)-1 {
				_, _ = buf.WriteString(", ")
			}
		}
		kv.AddString(f.key, buf.String())
		bufferpool.Put(buf)
	case typeString:
		kv.AddString(f.key, f.string)
	case typeStrings:
		buf := bufferpool.Get()
		vals := f.obj.([]string)
		for i, s := range vals {
			_, _ = buf.WriteString(s)
			if i < len(vals)-1 {
				_, _ = buf.WriteString(", ")
			}
		}
		kv.AddString(f.key, buf.String())
		bufferpool.Put(buf)
	case typeStringer:
		kv.AddString(f.key, f.obj.(fmt.Stringer).String())
	case typeGoStringer:
		kv.AddString(f.key, f.obj.(fmt.GoStringer).GoString())
	case typeObject:
		kv.AddObject(f.key, f.obj)
	case typeObjectTypeOf:
		kv.AddString(f.key, fmt.Sprintf("%T", f.obj))
	case typeMarshaler:
		return kv.AddMarshaler(f.key, f.obj.(Marshaler))
	case typeStringFn:
		return errors.Wrap(f.strFn(kv.AddString), "[log] AddTo.StringFn")
	case typeFields:
		for _, f := range f.obj.(Fields) {
			if err := f.AddTo(kv); err != nil {
				return errors.Wrap(err, "[log] AddTo.FieldCollection")
			}
		}
	default:
		return errors.NotSupported.Newf("[log] Field type not supported: %v", f)
	}
	return nil
}

// Bool constructs a Field with the given key and value.
func Bool(key string, value bool) Field {
	var val int64
	if value {
		val = 1
	}
	return field{key: key, fieldType: typeBool, int64: val}
}

// Float64 constructs a Field with the given key and value.
func Float64(key string, value float64) Field {
	return field{key: key, fieldType: typeFloat64, float64: value}
}

// Int constructs a Field with the given key and value.
func Int(key string, val int) Field {
	return field{key: key, fieldType: typeInt, int64: int64(val)}
}

// Ints constructs a Field with the given key and multiple values.
// Values will be joined together via a comma.
func Ints(key string, vals ...int) Field {
	return field{key: key, fieldType: typeInts, obj: vals}
}

// Int64 constructs a Field with the given key and value.
func Int64(key string, val int64) Field {
	return field{key: key, fieldType: typeInt64, int64: val}
}

// Int64s constructs a Field with the given key and multiple values.
// Values will be joined together via a comma.
func Int64s(key string, vals ...int64) Field {
	return field{key: key, fieldType: typeInt64s, obj: vals}
}

// Uint constructs a Field with the given key and value.
func Uint(key string, val uint) Field {
	return field{key: key, fieldType: typeUint64, uint64: uint64(val)}
}

// Uint64 constructs a Field with the given key and value.
func Uint64(key string, val uint64) Field {
	return field{key: key, fieldType: typeUint64, uint64: val}
}

// String constructs a Field with the given key and value.
func String(key string, val string) Field {
	return field{key: key, fieldType: typeString, string: val}
}

// Strings constructs a Field with the given key and multiple values.
// Values will be joined together via a comma.
func Strings(key string, vals ...string) Field {
	return field{key: key, fieldType: typeStrings, obj: vals}
}

// StringFn constructs a Field with the given key and a closure to the
// AddStringFn.
func StringFn(key string, fn func(AddStringFn) error) Field {
	return field{key: key, fieldType: typeStringFn, strFn: fn}
}

// Stringer constructs a Field with the given key and value. The value
// is the result of the String() method.
func Stringer(key string, val fmt.Stringer) Field {
	return field{key: key, fieldType: typeStringer, obj: val}
}

// GoStringer constructs a Field with the given key and value. The value
// is the result of the GoString() method.
func GoStringer(key string, val fmt.GoStringer) Field {
	return field{key: key, fieldType: typeGoStringer, obj: val}
}

// Text constructs a Field with the given key and value. The value is the result
// of the MarshalText() method. See package encoding in the standard library for
// encoding.TextMarshaler.
func Text(key string, val textMarshaler) Field {
	return field{key: key, fieldType: typeStringFn, strFn: func(addString AddStringFn) error {
		txt, err := val.MarshalText()
		if err != nil {
			return errors.Wrap(err, "[log] AddTo.TextMarshaler")
		}
		addString(key, string(txt))
		return nil
	}}
}

// JSON constructs a Field with the given key and value. The value is the result
// of the MarshalJSON() method. See package encoding/json in the standard
// library for json.Marshaler.
func JSON(key string, val jsonMarshaler) Field {
	return field{key: key, fieldType: typeStringFn, strFn: func(addString AddStringFn) error {
		j, err := val.MarshalJSON()
		if err != nil {
			return errors.Wrap(err, "[log] JSON.MarshalJSON")
		}
		addString(key, string(j))
		return nil
	}}
}

// Time constructs a Field with the given key and value. It represents a
// time.Time as nanoseconds since epoch. If you need a human formatted time, use
// the Stringer function.
func Time(key string, val time.Time) Field {
	return Int64(key, val.UnixNano())
}

// Duration constructs a Field with the given key and value. It represents
// durations as an integer number of nanoseconds.
func Duration(key string, val time.Duration) Field {
	return field{key: key, fieldType: typeInt64, int64: val.Nanoseconds()}
}

// UnixNanoHuman converts a unix nano value using the format string
//	"2006-01-02 15:04:05.999999999 -0700 MST"
func UnixNanoHuman(key string, val int64) Field {
	secs := val / int64(time.Second)
	var nsecs int64
	if secs > 0 {
		nsecs = val % secs
	}
	return String(key, time.Unix(secs, nsecs).String())
}

// Err constructs a Field that stores err under the key log.ErrorKeyName. Prints
// <nil> if the error is nil.
func Err(err error) Field {
	if err == nil {
		return String(KeyNameError, "<nil>")
	}
	return String(KeyNameError, err.Error())
}

// ErrWithKey constructs a Field that stores err under a key. Prints
// <nil> if the error is nil.
func ErrWithKey(key string, err error) Field {
	if err == nil {
		return String(key, "<nil>")
	}
	return String(key, err.Error())
}

// Object constructs a field with the given key and an arbitrary object. It uses
// an encoding-appropriate, reflection-based function to serialize nearly any
// object into the logging context, but it's relatively slow and allocation-heavy.
//
// If encoding fails (e.g., trying to serialize a map[int]string to JSON), Object
// includes the error message in the final log output.
func Object(key string, val interface{}) Field {
	return field{key: key, fieldType: typeObject, obj: val}
}

// ObjectTypeOf logs the type of the `val` interface. Same as %T in fmt.Printf or
// reflect.TypeOf.
func ObjectTypeOf(key string, val interface{}) Field {
	return field{key: key, fieldType: typeObjectTypeOf, obj: val}
}

// Marshal constructs a field with the given key and log.Marshaler. It
// provides a flexible, but still type-safe and efficient, way to add
// user-defined types to the logging context.
func Marshal(key string, val Marshaler) Field {
	return field{key: key, fieldType: typeMarshaler, obj: val}
}

// Nest takes a key and a variadic number of Fields and creates a nested
// namespace.
func Nest(key string, fields ...Field) Field {
	return field{key: key, fieldType: typeMarshaler, obj: Fields(fields)}
}

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
	"bytes"
	"encoding"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"testing"
	"time"

	"github.com/corestoreio/pkg/util/assert"
)

type iFaceCheck struct{}

func (iFaceCheck) MarshalText() ([]byte, error) {
	return nil, nil
}

func (iFaceCheck) MarshalJSON() ([]byte, error) {
	return nil, nil
}

// Test if our internal interface definition is up to par with the stdlib.
var _ textMarshaler = (*iFaceCheck)(nil)

var (
	_ encoding.TextMarshaler = (*iFaceCheck)(nil)
	_ jsonMarshaler          = (*iFaceCheck)(nil)
	_ json.Marshaler         = (*iFaceCheck)(nil)
	_ Field                  = (*field)(nil)
	_ Field                  = (*Fields)(nil)
)

const testKey = "MyTestKey"

func TestFields_ToString(t *testing.T) {
	xChan := make(chan uint64)
	fs := Fields{
		String("k1", "v1"),
		Int("k2", 2),
		Float64("k3", 3.14159),
		ObjectTypeOf("k4", xChan),
	}
	str := fs.ToString("fieldsKey")
	assert.Exactly(t, "fieldsKey k1: \"v1\" k2: 2 k3: 3.14159 k4: \"chan uint64\"\n", str)
}

func TestFields_ToString_Error(t *testing.T) {
	fs := Fields{
		Text("o1", gs{err: errors.New("ErrToString")}),
		Int("k2", 2),
		Float64("k3", 3.14159),
	}
	str := fs.ToString("fieldsKey")
	assert.Contains(t, str, "fieldsKey error: ErrToString\n")
	assert.Contains(t, str, "[log] AddTo.TextMarshaler\n")
}

func TestFields_Add(t *testing.T) {
	f := Fields{String("a", "b"), Int("c", 3)}.make()
	assert.Exactly(t, typeFields, f.fieldType)
	assert.Exactly(t, Fields{String("a", "b"), Int("c", 3)}, f.obj.(Fields))

	buf := &bytes.Buffer{}
	wt := WriteTypes{W: buf}
	if err := f.AddTo(wt); err != nil {
		t.Fatal(err)
	}
	assert.Exactly(t, " a: \"b\" c: 3", buf.String())
}

func TestField_Bool(t *testing.T) {
	f := Bool(testKey, true).make()
	assert.Exactly(t, typeBool, f.fieldType)
	assert.Exactly(t, int64(1), f.int64)
	assert.Exactly(t, testKey, f.key)
}

func TestField_Float64(t *testing.T) {
	f := Float64(testKey, math.Pi).make()
	assert.Exactly(t, typeFloat64, f.fieldType)
	assert.Exactly(t, math.Pi, f.float64)
	assert.Exactly(t, testKey, f.key)
}

func TestField_Int(t *testing.T) {
	f := Int(testKey, math.MaxInt32).make()
	assert.Exactly(t, typeInt, f.fieldType)
	assert.Exactly(t, int64(math.MaxInt32), f.int64)
	assert.Exactly(t, testKey, f.key)
}

func TestField_Ints(t *testing.T) {
	f := Ints(testKey, 4, 5, 6, 7, 8).make()
	assert.Exactly(t, typeInts, f.fieldType)
	assert.Empty(t, f.int64)
	assert.Exactly(t, testKey, f.key)
	buf := &bytes.Buffer{}
	wt := WriteTypes{W: buf}
	if err := f.AddTo(wt); err != nil {
		t.Fatal(err)
	}
	assert.Exactly(t, " MyTestKey: \"4, 5, 6, 7, 8\"", buf.String())
}

func TestField_Int64(t *testing.T) {
	f := Int64(testKey, math.MaxInt64).make()
	assert.Exactly(t, typeInt64, f.fieldType)
	assert.Exactly(t, int64(math.MaxInt64), f.int64)
	assert.Exactly(t, testKey, f.key)
}

func TestField_Int64s(t *testing.T) {
	f := Int64s(testKey, 4, 5, 6, 7, 8).make()
	assert.Exactly(t, typeInt64s, f.fieldType)
	assert.Empty(t, f.int64)
	assert.Exactly(t, testKey, f.key)
	buf := &bytes.Buffer{}
	wt := WriteTypes{W: buf}
	if err := f.AddTo(wt); err != nil {
		t.Fatal(err)
	}
	assert.Exactly(t, " MyTestKey: \"4, 5, 6, 7, 8\"", buf.String())
}

func TestField_Uint(t *testing.T) {
	f := Uint(testKey, math.MaxUint32).make()
	assert.Exactly(t, typeUint64, f.fieldType)
	assert.Exactly(t, uint64(math.MaxUint32), f.uint64)
	assert.Exactly(t, testKey, f.key)
}

func TestField_Uint64(t *testing.T) {
	f := Uint64(testKey, math.MaxUint64).make()
	assert.Exactly(t, typeUint64, f.fieldType)
	assert.Exactly(t, uint64(math.MaxUint64), f.uint64)
	assert.Exactly(t, testKey, f.key)
}

func TestField_String(t *testing.T) {
	const data = `16. “One is never alone with a rubber duck.” Douglas Adams`
	f := String(testKey, data).make()
	assert.Exactly(t, typeString, f.fieldType)
	assert.Exactly(t, data, f.string)
	assert.Exactly(t, testKey, f.key)
}

func TestField_Strings(t *testing.T) {
	f := Strings(testKey, "a", "b", "c", "d", "e").make()
	assert.Exactly(t, typeStrings, f.fieldType)
	assert.Empty(t, f.string)
	assert.Exactly(t, testKey, f.key)
	buf := &bytes.Buffer{}
	wt := WriteTypes{W: buf}
	if err := f.AddTo(wt); err != nil {
		t.Fatal(err)
	}
	assert.Exactly(t, " MyTestKey: \"a, b, c, d, e\"", buf.String())
}

func TestField_Stringer(t *testing.T) {
	const data = `27. “Anything invented after you're thirty-five is against the natural order of things.” Douglas Adams`
	f := Stringer(testKey, bytes.NewBufferString(data)).make()
	assert.Exactly(t, typeStringer, f.fieldType)
	assert.Empty(t, f.string)
	assert.Exactly(t, testKey, f.key)

	buf := &bytes.Buffer{}
	wt := WriteTypes{W: buf}
	if err := f.AddTo(wt); err != nil {
		t.Fatal(err)
	}
	assert.Exactly(t, " MyTestKey: \"27. “Anything invented after you're thirty-five is against the natural order of things.” Douglas Adams\"", buf.String())
}

type gs struct {
	data interface{}
	err  error
}

func (g gs) MarshalText() ([]byte, error) {
	if g.err != nil {
		return nil, g.err
	}
	return g.data.([]byte), nil
}
func (gs) GoString() string { return "gs struct {}" }
func (g gs) MarshalJSON() ([]byte, error) {
	d, err := json.Marshal(g.data)
	if err != nil {
		g.err = err
	}
	return d, g.err
}

func (g gs) MarshalLog(kv KeyValuer) error {
	if g.err != nil {
		return g.err
	}
	kv.AddObject("MarshalLogKey", g.data)
	return nil
}

func TestField_GoStringer(t *testing.T) {
	f := GoStringer(testKey, gs{}).make()
	assert.Exactly(t, typeGoStringer, f.fieldType)
	assert.Empty(t, f.string)
	assert.Exactly(t, testKey, f.key)
	buf := &bytes.Buffer{}
	wt := WriteTypes{W: buf}
	if err := f.AddTo(wt); err != nil {
		t.Fatal(err)
	}
	assert.Exactly(t, " MyTestKey: \"gs struct {}\"", buf.String())
}

func TestField_Marshaler(t *testing.T) {
	f := Marshal(testKey, gs{data: "MarshalerMarshaler"}).make()
	assert.Exactly(t, typeMarshaler, f.fieldType)
	assert.Empty(t, f.string)
	assert.Exactly(t, testKey, f.key)
	buf := &bytes.Buffer{}
	wt := WriteTypes{W: buf}
	if err := f.AddTo(wt); err != nil {
		t.Fatal(err)
	}
	assert.Exactly(t, " MarshalLogKey: \"MarshalerMarshaler\"", buf.String())
}

type textMar string

func (tm textMar) MarshalText() (text []byte, err error) {
	return []byte(tm), nil
}

func TestField_Text(t *testing.T) {
	const data = `35. “My universe is my eyes and my ears. Anything else is hearsay.” Douglas Adams`
	f := Text(testKey, textMar(data)).make()
	assert.Exactly(t, typeStringFn, f.fieldType)
	assert.Empty(t, f.string)
	assert.Exactly(t, testKey, f.key)
	buf := &bytes.Buffer{}
	wt := WriteTypes{W: buf}
	if err := f.AddTo(wt); err != nil {
		t.Fatal(err)
	}
	assert.Exactly(t, " MyTestKey: \"35. “My universe is my eyes and my ears. Anything else is hearsay.” Douglas Adams\"", buf.String())
}

func TestField_TextError(t *testing.T) {
	data := gs{data: nil, err: errors.New("Errr")}
	f := Text(testKey, data).make()
	assert.Exactly(t, typeStringFn, f.fieldType)
	assert.Empty(t, f.string)
	assert.Exactly(t, testKey, f.key)
	buf := &bytes.Buffer{}
	wt := WriteTypes{W: buf}
	err := f.AddTo(wt)
	assert.Empty(t, buf.String())
	assert.EqualError(t, err, "[log] AddTo.StringFn: [log] AddTo.TextMarshaler: Errr")
}

func TestField_JSON(t *testing.T) {
	const data = `12. “Reality is frequently inaccurate.” Douglas Adams`
	f := JSON(testKey, gs{data: data}).make()
	assert.Exactly(t, typeStringFn, f.fieldType)
	assert.Empty(t, f.string)
	assert.Exactly(t, testKey, f.key)

	buf := &bytes.Buffer{}
	wt := WriteTypes{W: buf}
	assert.NoError(t, f.AddTo(wt))
	assert.Exactly(t, ` MyTestKey: "\"12. “Reality is frequently inaccurate.” Douglas Adams\""`, buf.String())
}

func TestField_JSONError(t *testing.T) {
	f := JSON(testKey, gs{data: make(chan struct{})}).make()
	assert.Exactly(t, typeStringFn, f.fieldType)
	assert.Exactly(t, testKey, f.key)
	buf := &bytes.Buffer{}
	wt := WriteTypes{W: buf}
	err := f.AddTo(wt)
	assert.Empty(t, buf.String())
	assert.EqualError(t, err, "[log] AddTo.StringFn: [log] JSON.MarshalJSON: json: unsupported type: chan struct {}")
}

func TestField_Time(t *testing.T) {
	now := time.Now()
	f := Time(testKey, now).make()
	assert.Exactly(t, typeInt64, f.fieldType)
	assert.Exactly(t, now.UnixNano(), f.int64)
	assert.Exactly(t, testKey, f.key)
}

func TestField_UnixNanoHuman(t *testing.T) {
	un := time.Unix(1234567890, 10).In(time.UTC).UnixNano()
	f := UnixNanoHuman(testKey, un).make()
	assert.Exactly(t, typeString, f.fieldType)
	assert.Exactly(t, `2009-02-14 00:31:30.00000001 +0100 CET`, f.string)
	assert.Exactly(t, testKey, f.key)
}

func TestField_UnixNanoHuman_Zero(t *testing.T) {
	f := UnixNanoHuman(testKey, 0).make()
	assert.Exactly(t, typeString, f.fieldType)
	assert.Exactly(t, `1970-01-01 01:00:00 +0100 CET`, f.string)
	assert.Exactly(t, testKey, f.key)
}

func TestField_Duration(t *testing.T) {
	now := time.Hour * 2
	f := Duration(testKey, now).make()
	assert.Exactly(t, typeInt64, f.fieldType)
	assert.Exactly(t, now.Nanoseconds(), f.int64)
	assert.Exactly(t, testKey, f.key)
}

func TestField_Error(t *testing.T) {
	const data = `15. “There is no point in using the word 'impossible' to describe something that has clearly happened.” Douglas Adams`
	err := errors.New(data)
	f := Err(err).make()
	assert.Exactly(t, typeString, f.fieldType)
	assert.Exactly(t, data, f.string)
	assert.Exactly(t, KeyNameError, f.key)
}

func TestField_Error_Nil(t *testing.T) {
	f := Err(nil).make()
	assert.Exactly(t, typeString, f.fieldType)
	assert.Exactly(t, `<nil>`, f.string)
	assert.Exactly(t, KeyNameError, f.key)
}

func TestField_ErrorWithKey(t *testing.T) {
	const data = `15. “There is no point in using the word 'impossible' to describe something that has clearly happened.” Douglas Adams`
	err := errors.New(data)
	f := ErrWithKey("e1", err).make()
	assert.Exactly(t, typeString, f.fieldType)
	assert.Exactly(t, data, f.string)
	assert.Exactly(t, `e1`, f.key)
}

func TestField_ErrorWithKey_Nil(t *testing.T) {
	f := ErrWithKey(`e2`, nil).make()
	assert.Exactly(t, typeString, f.fieldType)
	assert.Exactly(t, `<nil>`, f.string)
	assert.Exactly(t, `e2`, f.key)
}

func TestField_Object(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://corestore.io", nil)
	req.RemoteAddr = "192.168.0.42"
	f := Object(testKey, req).make()
	assert.Exactly(t, typeObject, f.fieldType)
	assert.Exactly(t, req, f.obj)
	assert.Exactly(t, testKey, f.key)
}

func TestField_ObjectTypeOf(t *testing.T) {
	i := 99
	f := ObjectTypeOf(testKey, i).make()
	assert.Exactly(t, typeObjectTypeOf, f.fieldType)
	assert.Exactly(t, i, f.obj)
	assert.Exactly(t, testKey, f.key)
}

func TestField_Nest(t *testing.T) {
	f := Nest("nest0",
		String("nest1", "1"),
		Int("nest2", 2),
		Int64("", 3),
		Float64("nest4", math.Log2E),
	).make()
	assert.Exactly(t, typeMarshaler, f.fieldType)
	assert.Exactly(t, `nest0`, f.key)
	buf := &bytes.Buffer{}
	wt := WriteTypes{W: buf}
	if err := f.AddTo(wt); err != nil {
		t.Fatal(err)
	}
	assert.Exactly(t, " nest1: \"1\" nest2: 2 _: 3 nest4: 1.4426950408889634", buf.String())
}

func TestField_Nest_Error(t *testing.T) {
	f := Nest("nest0",
		String("nest1", "1"),
		Text("nest2", gs{err: errors.New("NestError. Smoke Alarm on ;-)")}),
	).make()
	assert.Exactly(t, typeMarshaler, f.fieldType)
	assert.Exactly(t, `nest0`, f.key)
	buf := &bytes.Buffer{}
	wt := WriteTypes{W: buf}
	if err := f.AddTo(wt); err != nil {
		t.Fatal(err)
	}
	assert.Contains(t, buf.String(), `nest1: "1" error: NestError. Smoke Alarm on ;-)`)
	assert.Contains(t, buf.String(), `[log] AddTo.TextMarshaler`)
}

var benchmarkFieldsToString string

func BenchmarkFields_ToString(b *testing.B) {
	anError := errors.New("I'm an error")
	fs := Fields{
		String("a", "b"), Int("c", 3),
		Int64("d", 33), Uint("f", 9), Bool("true", true),
		Err(anError),
		Ints("multiple_ints", 1, 2, 3, 4, 5, 6, 7, 8, 9),
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchmarkFieldsToString = fs.ToString("Convert to string")
	}
}

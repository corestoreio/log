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

package logw_test

import (
	"bytes"
	std "log"
	"math"
	"testing"

	"github.com/corestoreio/errors"
	"github.com/corestoreio/log"
	"github.com/corestoreio/log/logw"
	"github.com/corestoreio/pkg/util/assert"
)

var _ log.Logger = (*logw.Log)(nil)

func TestStdLog(t *testing.T) {
	buf := new(bytes.Buffer)

	sl := logw.NewLog(
		logw.WithLevel(logw.LevelInfo),
		logw.WithDebug(buf, "TEST-DEBUG ", std.LstdFlags),
		logw.WithInfo(buf, "TEST-INFO ", std.LstdFlags),
	)

	assert.False(t, sl.IsDebug())
	assert.True(t, sl.IsInfo())

	sl.Debug("my Debug", log.Float64("float", 3.14152))
	sl.Debug("my Debug2", log.Float64("float2", 2.14152))
	sl.Info("InfoTEST")

	logs := buf.String()

	assert.Contains(t, logs, "InfoTEST")
	assert.NotContains(t, logs, "Debug2")

	buf.Reset()

	sl = logw.NewLog(
		logw.WithLevel(logw.LevelDebug),
		logw.WithDebug(buf, "TEST-DEBUG ", std.LstdFlags),
		logw.WithInfo(buf, "TEST-INFO ", std.LstdFlags),
	)

	assert.True(t, sl.IsDebug())
	assert.True(t, sl.IsInfo())
	sl.Debug("my Debug", log.Float64("float", 3.14152))
	sl.Debug("my Debug2", log.Float64("float2", 2.14152))
	sl.Info("InfoTEST")

	logs = buf.String()

	assert.Contains(t, logs, "InfoTEST")
	assert.Contains(t, logs, "Debug2")
}

func TestStdLogGlobals(t *testing.T) {
	buf := new(bytes.Buffer)
	sl := logw.NewLog(
		logw.WithLevel(logw.LevelDebug),
		logw.WithWriter(buf),
		logw.WithFlag(std.Ldate),
	)
	sl.Debug("my Debug", log.Float64("float", 3.14152))
	sl.Debug("my Debug2", log.Float64("float2", 2.14152))
	sl.Info("InfoTEST")

	logs := buf.String()

	assert.NotContains(t, logs, "trace2")
	assert.Contains(t, logs, "InfoTEST")
	assert.NotContains(t, logs, "trace1")
	assert.Contains(t, logs, "Debug2")
}

func TestStdLogFormat(t *testing.T) {
	buf := new(bytes.Buffer)
	bufInfo := new(bytes.Buffer)
	sl := logw.NewLog(
		logw.WithLevel(logw.LevelDebug),
		logw.WithWriter(buf),
		logw.WithInfo(bufInfo, "TEST-INFO ", std.LstdFlags),
	)

	sl.Debug("my Debug", log.Float64("float1", 3.14152))
	sl.Debug("my Debug2", log.Float64("", 2.14152))
	sl.Debug("my Debug3", log.Int("key3", 3105), log.Int64("Hello", 4711))
	sl.Info("InfoTEST")
	sl.Info("InfoTEST", log.Int("keyI", 117), log.Int64("year", 2009))
	sl.Info("InfoTEST", log.String("", "Now we have the salad"))

	logs := buf.String()
	logsInfo := bufInfo.String()

	assert.Contains(t, logs, "Debug2")
	assert.NotContains(t, logs, "BAD_KEY_AT_INDEX_0")
	assert.NotContains(t, logs, `key3: 3105 BAD_KEY_AT_INDEX_2: "Hello"`)

	assert.Contains(t, logsInfo, "InfoTEST")
	assert.Contains(t, logsInfo, `_: "Now we have the salad`)
}

type marshalMock struct {
	string
	float64
	bool
	error
}

func (mm marshalMock) MarshalLog(kv log.KeyValuer) error {
	kv.AddBool("kvbool", mm.bool)
	kv.AddString("kvstring", mm.string)
	kv.AddFloat64("kvfloat64", mm.float64)
	return mm.error
}

func TestAddMarshaler(t *testing.T) {
	buf := new(bytes.Buffer)
	sl := logw.NewLog(
		logw.WithLevel(logw.LevelDebug),
		logw.WithWriter(buf),
	)

	sl.Debug("my Debug", log.Float64("float1", math.SqrtE))
	sl.Debug("marshalling", log.Object("anObject", 42), log.Marshal("marshalLogMock", marshalMock{
		string:  "s1",
		float64: math.Ln2,
		bool:    true,
	}))
	assert.Contains(t, buf.String(), `my Debug float1: 1.6487212707001282`)
	assert.Contains(t, buf.String(), `marshalling anObject: 42 kvbool: true kvstring: "s1" kvfloat64: 0.6931471805599453`)
}

func TestAddMarshaler_Error(t *testing.T) {
	buf := new(bytes.Buffer)
	sl := logw.NewLog(
		logw.WithLevel(logw.LevelDebug),
		logw.WithWriter(buf),
	)

	sl.Debug("my Debug", log.Float64("float1", math.SqrtE))
	sl.Debug("marshalling", log.Marshal("marshalLogMock", marshalMock{
		error: errors.New("Whooops"),
	}))
	assert.Regexp(t,
		`marshalling kvbool: false kvstring: "" kvfloat64: 0 error: Whooops\s+github.com/corestoreio/log/logw_test.TestAddMarshaler_Error`,
		buf.String())
}

func TestLog_With(t *testing.T) {
	buf := new(bytes.Buffer)
	pLog := logw.NewLog(
		logw.WithWriter(buf),
		logw.WithLevel(logw.LevelInfo),
		logw.WithFields(log.Int("parent_info1_level", 2)),
	)
	cLog := pLog.With(log.Int("child_debug1_level", 1))
	assert.Empty(t, buf.String(), "Expecting no logging output")

	pLog.Debug("Root: Debug Message")
	cLog.Info("Child1: Info Message", log.Int("info_child_key", 815))

	assert.NotContains(t, buf.String(), "Root: Debug Message")
	assert.Contains(t, buf.String(), "Child1: Info Message")
	assert.Contains(t, buf.String(), "parent_info1_level: 2 child_debug1_level: 1 info_child_key: 815")

	pLog.Info("Parent Info", log.Int("parent_info2", 457))
	assert.Contains(t, buf.String(), `Parent Info parent_info1_level: 2 parent_info2: 457`)
}

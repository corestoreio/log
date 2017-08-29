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
	"time"

	"github.com/corestoreio/errors"
)

// Now returns the current time including a monotonic time part. This variable
// can be changed for testing purposes.
var Now = time.Now

// KeyNameError whenever an error occurs during marshaling this value defines
// the official key name in the log stream.
const KeyNameError = `error`

// KeyNameDuration defines the name of the duration field logging with the
// struct type "Deferred".
const KeyNameDuration = `duration`

// Logger defines the minimum requirements for logging. See doc.go for more
// details.
type Logger interface {
	// With returns a new Logger that has this logger's context plus the given
	// Fields.
	With(...Field) Logger
	// Debug outputs information for developers including a stack trace.
	Debug(msg string, fields ...Field)
	// Info outputs information for users of the app
	Info(msg string, fields ...Field)
	// IsDebug returns true if Debug level is enabled
	IsDebug() bool
	// IsInfo returns true if Info level is enabled
	IsInfo() bool
}

// AssignmentChar represents the assignment character between key-value pairs
var AssignmentChar = ": "

// Separator is the separator to use between key value pairs
var Separator = " "

// WriteTypes satisfies the interface KeyValuer. It uses under the hood the
// function Sprintf("%#v", val) to print the values. This costs performance.
type WriteTypes struct {
	// AssignmentChar represents the assignment character between key-value
	// pairs
	AssignmentChar string
	// Separator is the separator to use between key value pairs
	Separator string
	// W used as writer. Must be a pointer. At the moment returned errors are
	// getting ignored.
	W interface {
		WriteString(s string) (n int, err error)
	}
}

// todo: check for errors in WriteString and remove _,_=

func (wt WriteTypes) stdSetKV(key string, value interface{}) {
	if wt.Separator == "" {
		wt.Separator = Separator
	}
	_, _ = wt.W.WriteString(wt.Separator)
	if key == "" {
		key = "_"
	}
	_, _ = wt.W.WriteString(key)
	if wt.AssignmentChar == "" {
		wt.AssignmentChar = AssignmentChar
	}
	_, _ = wt.W.WriteString(wt.AssignmentChar)
	_, _ = wt.W.WriteString(fmt.Sprintf("%#v", value)) // can be refactored into the different functions
}

func (wt WriteTypes) AddBool(key string, value bool) {
	wt.stdSetKV(key, value)
}
func (wt WriteTypes) AddFloat64(key string, value float64) {
	wt.stdSetKV(key, value)
}
func (wt WriteTypes) AddInt(key string, value int) {
	wt.stdSetKV(key, value)
}
func (wt WriteTypes) AddInt64(key string, value int64) {
	wt.stdSetKV(key, value)
}
func (wt WriteTypes) AddUint64(key string, value uint64) {
	wt.stdSetKV(key, value)
}
func (wt WriteTypes) AddMarshaler(key string, value Marshaler) error {
	if err := value.MarshalLog(wt); err != nil {
		if wt.Separator == "" {
			wt.Separator = Separator
		}
		_, _ = wt.W.WriteString(wt.Separator)
		_, _ = wt.W.WriteString(KeyNameError)
		if wt.AssignmentChar == "" {
			wt.AssignmentChar = AssignmentChar
		}
		_, _ = wt.W.WriteString(wt.AssignmentChar)
		_, _ = wt.W.WriteString(fmt.Sprintf("%+v", err))
	}
	return nil
}
func (wt WriteTypes) AddObject(key string, value interface{}) {
	wt.stdSetKV(key, value)
}
func (wt WriteTypes) AddString(key string, value string) {
	wt.stdSetKV(key, value)
}

// Nest allows the caller to populate a nested object under the provided key.
func (wt WriteTypes) Nest(key string, f func(KeyValuer) error) error {
	if wt.Separator == "" {
		wt.Separator = Separator
	}
	_, _ = wt.W.WriteString(wt.Separator)
	if key == "" {
		key = "_"
	}
	_, _ = wt.W.WriteString(key)
	if wt.AssignmentChar == "" {
		wt.AssignmentChar = AssignmentChar
	}
	_, _ = wt.W.WriteString(wt.AssignmentChar)
	return errors.Wrap(f(wt), "[log] WriteType.Nest.f")
}

// Deferred defines a logger type which can be used to trace the duration.
// Usage:
//		function main(){
//			lg := log.NewStdLog()
//			// my code ...
// 			defer log.WhenDone(lg).Info("Stats", log.String("Package", "main"))
//			...
// 		}
// Outputs the duration for the main action.
type Deferred struct {
	Info  func(msg string, fields ...Field)
	Debug func(msg string, fields ...Field)
}

// WhenDone returns a Logger which tracks the duration.
func WhenDone(l Logger) Deferred {
	// @see http://play.golang.org/p/K53LV16F9e from @francesc
	start := Now()
	return Deferred{
		Info: func(msg string, fields ...Field) {
			l.Info(msg, Duration(KeyNameDuration, Now().Sub(start)), Fields(fields))
		},
		Debug: func(msg string, fields ...Field) {
			l.Debug(msg, Duration(KeyNameDuration, Now().Sub(start)), Fields(fields))
		},
	}
}

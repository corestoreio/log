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

package logzero

import (
	"fmt"

	"github.com/corestoreio/errors"
	"github.com/corestoreio/log"
	"github.com/rs/zerolog"
)

type Wrap struct {
	level  zerolog.Level
	logger zerolog.Logger
	// ctx is only set when we act as a child logger
	ctx log.Fields
}

// New creates a new https://godoc.org/github.com/rs/zerolog logger.
func New(lvl zerolog.Level, zl zerolog.Logger) *Wrap {
	l := &Wrap{
		level:  lvl,
		logger: zl,
	}
	return l
}

// With creates a new inherited and shallow copied Logger with additional fields
// added to the logging context.
func (l *Wrap) With(fields ...log.Field) log.Logger {
	l2 := *l
	l2.ctx = append(l2.ctx, fields...)
	return &l2
}

// Info outputs information for users of the app
func (l *Wrap) Info(msg string, fields ...log.Field) {
	doZLFieldWrap(l.ctx, l.logger.Info(), msg, fields...)
}

// Debug outputs information for developers.
func (l *Wrap) Debug(msg string, fields ...log.Field) {
	doZLFieldWrap(l.ctx, l.logger.Debug(), msg, fields...)
}

// IsDebug returns true if Debug level is enabled
func (l *Wrap) IsDebug() bool {
	return l.level >= zerolog.DebugLevel
}

// IsInfo returns true if Info level is enabled
func (l *Wrap) IsInfo() bool {
	return l.level >= zerolog.InfoLevel
}

type log15FieldWrap struct {
	ifaces []interface{}
}

func doZLFieldWrap(ctx log.Fields, zl *zerolog.Event, msg string, fs ...log.Field) {
	if ctxl := len(ctx); ctxl > 0 {
		all := make(log.Fields, 0, ctxl+len(fs))
		all = append(all, ctx...)
		all = append(all, fs...)
		fs = all
	}

	fw := &log15FieldWrap{
		ifaces: make([]interface{}, 0, len(fs)*2),
	}

	if err := log.Fields(fs).AddTo(fw); err != nil {
		fw.AddString(log.KeyNameError, fmt.Sprintf("%+v", err))
	}

	zl.Fields(fw.ifaces).Msg(msg)
}

func (se *log15FieldWrap) append(key string, val interface{}) {
	se.ifaces = append(se.ifaces, key, val)
}

func (se *log15FieldWrap) AddBool(k string, v bool) {
	se.append(k, v)
}

func (se *log15FieldWrap) AddFloat64(k string, v float64) {
	se.append(k, v)
}

func (se *log15FieldWrap) AddInt(k string, v int) {
	se.append(k, v)
}

func (se *log15FieldWrap) AddInt64(k string, v int64) {
	se.append(k, v)
}

func (se *log15FieldWrap) AddUint64(k string, v uint64) {
	se.append(k, v)
}

func (se *log15FieldWrap) AddMarshaler(k string, v log.Marshaler) error {
	if err := v.MarshalLog(se); err != nil {
		se.AddString(log.KeyNameError, fmt.Sprintf("%+v", err))
	}
	return nil
}

func (se *log15FieldWrap) AddObject(k string, v interface{}) {
	se.append(k, v)
}

func (se *log15FieldWrap) AddString(k string, v string) {
	se.append(k, v)
}

func (se *log15FieldWrap) Nest(key string, f func(log.KeyValuer) error) error {
	se.append(key, "nest")
	return errors.WithStack(f(se))
}

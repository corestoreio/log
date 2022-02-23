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

package log15w

import (
	"fmt"

	"github.com/corestoreio/errors"
	"github.com/corestoreio/log"
	"github.com/inconshreveable/log15"
)

type Wrap struct {
	level  log15.Lvl
	logger log15.Logger
	// ctx is only set when we act as a child logger
	ctx log.Fields
}

// New creates a new https://godoc.org/github.com/inconshreveable/log15 logger.
func New(lvl log15.Lvl, h log15.Handler, ctx ...interface{}) *Wrap {
	l := &Wrap{
		level:  lvl,
		logger: log15.New(ctx...),
	}
	l.logger.SetHandler(h)
	return l
}

// With creates a new inherited and shallow copied Logger with additional fields
// added to the logging context.
func (l *Wrap) With(fields ...log.Field) log.Logger {
	l2 := new(Wrap)
	*l2 = *l
	l2.ctx = append(l2.ctx, fields...)
	return l2
}

// Info outputs information for users of the app
func (l *Wrap) Info(msg string, fields ...log.Field) {
	l.logger.Info(msg, doLog15FieldWrap(l.ctx, fields...)...)
}

// Debug outputs information for developers.
func (l *Wrap) Debug(msg string, fields ...log.Field) {
	l.logger.Debug(msg, doLog15FieldWrap(l.ctx, fields...)...)
}

// IsDebug returns true if Debug level is enabled
func (l *Wrap) IsDebug() bool {
	return l.level >= log15.LvlDebug
}

// IsInfo returns true if Info level is enabled
func (l *Wrap) IsInfo() bool {
	return l.level >= log15.LvlInfo
}

type log15FieldWrap struct {
	ifaces []interface{}
}

func doLog15FieldWrap(ctx log.Fields, fs ...log.Field) []interface{} {
	if ctxl := len(ctx); ctxl > 0 {
		all := make(log.Fields, 0, ctxl+len(fs))
		all = append(all, ctx...)
		all = append(all, fs...)
		fs = all
	}

	fw := &log15FieldWrap{
		ifaces: make([]interface{}, 0, 6), // just guessing not more than 6 args / 3 Fields
	}

	if err := log.Fields(fs).AddTo(fw); err != nil {
		fw.AddString(log.KeyNameError, fmt.Sprintf("%+v", err))
	}
	return fw.ifaces
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

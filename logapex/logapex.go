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

package logapex

import (
	"fmt"

	apx "github.com/apex/log"
	"github.com/corestoreio/errors"
	"github.com/corestoreio/log"
)

type Wrap struct {
	level apx.Level
	wrap  *apx.Logger
	// ctx is only set when we act as a child logger
	ctx log.Fields
}

// New creates a new https://godoc.org/github.com/apex/log logger.
func New(lvl apx.Level, l *apx.Logger, fields ...log.Field) *Wrap {
	return &Wrap{
		level: lvl,
		wrap:  l,
		ctx:   fields,
	}
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
	l.wrap.WithFields(wrapFields(l.ctx, fields...)).Info(msg)
}

// Debug outputs information for developers.
func (l *Wrap) Debug(msg string, fields ...log.Field) {
	l.wrap.WithFields(wrapFields(l.ctx, fields...)).Debug(msg)
}

// IsDebug returns true if Debug level is enabled
func (l *Wrap) IsDebug() bool {
	return l.level <= apx.DebugLevel
}

// IsInfo returns true if Info level is enabled
func (l *Wrap) IsInfo() bool {
	return l.level <= apx.InfoLevel
}

type fieldWrap struct {
	apx apx.Fields
}

func wrapFields(ctx log.Fields, fs ...log.Field) apx.Fields {
	if ctxl := len(ctx); ctxl > 0 {
		all := make(log.Fields, 0, ctxl+len(fs))
		all = append(all, ctx...)
		all = append(all, fs...)
		fs = all
	}

	fw := &fieldWrap{
		apx: apx.Fields{},
	}

	if err := log.Fields(fs).AddTo(fw); err != nil {
		fw.AddString(log.KeyNameError, fmt.Sprintf("%+v", err))
	}
	return fw.apx
}

func (se *fieldWrap) AddBool(k string, v bool) {
	se.apx[k] = v
}
func (se *fieldWrap) AddFloat64(k string, v float64) {
	se.apx[k] = v
}
func (se *fieldWrap) AddInt(k string, v int) {
	se.apx[k] = v
}
func (se *fieldWrap) AddInt64(k string, v int64) {
	se.apx[k] = v
}
func (se *fieldWrap) AddUint64(k string, v uint64) {
	se.apx[k] = v
}
func (se *fieldWrap) AddMarshaler(k string, v log.Marshaler) error {
	if err := v.MarshalLog(se); err != nil {
		se.AddString(log.KeyNameError, fmt.Sprintf("%+v", err))
	}
	return nil
}
func (se *fieldWrap) AddObject(k string, v interface{}) {
	se.apx[k] = v
}
func (se *fieldWrap) AddString(k string, v string) {
	se.apx[k] = v
}

func (se *fieldWrap) Nest(key string, f func(log.KeyValuer) error) error {
	se.apx[key] = "nest"
	return errors.WithStack(f(se))
}

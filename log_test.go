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

package log_test

import (
	"bytes"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/corestoreio/log"
	"github.com/corestoreio/log/logw"
	"github.com/corestoreio/pkg/util/assert"
)

var (
	_ log.Logger    = (*log.BlackHole)(nil)
	_ log.KeyValuer = (*log.WriteTypes)(nil)
)

func TestWhenDone(t *testing.T) {
	t.Run("Level_Debug", testWhenDone(logw.LevelDebug))
	t.Run("Level_Info", testWhenDone(logw.LevelInfo))
	t.Run("Level_Fatal", testWhenDone(logw.LevelFatal))
}

func testWhenDone(lvl int) func(*testing.T) {
	return func(t *testing.T) {
		buf := &bytes.Buffer{}
		l := logw.NewLog(logw.WithWriter(buf), logw.WithLevel(lvl))
		var wg sync.WaitGroup
		wg.Add(1)
		wd := log.WhenDone(l)
		go func(wg2 *sync.WaitGroup) {
			defer wg2.Done()
			defer wd.Debug("WhenDoneDebug", log.Int("key1", 123))
			defer wd.Info("WhenDoneInfo", log.Int("key2", 321))
			time.Sleep(time.Millisecond * 250)
		}(&wg)
		wg.Wait()

		if lvl == logw.LevelDebug {
			assert.Contains(t, buf.String(), `WhenDoneDebug`)
			assert.Contains(t, buf.String(), `key1: 123`)
			assert.Contains(t, buf.String(), log.KeyNameDuration+`: 25`)
		} else {
			assert.NotContains(t, buf.String(), `WhenDoneDebug`)
			assert.NotContains(t, buf.String(), `key1: 123`)
			//assert.NotContains(t, buf.String(), log.KeyNameDuration+`: 25`)
		}
		if lvl >= logw.LevelInfo {
			assert.Contains(t, buf.String(), `WhenDoneInfo`)
			assert.Contains(t, buf.String(), `key2: 321`)
			assert.Contains(t, buf.String(), log.KeyNameDuration+`: 25`)
		} else {
			assert.NotContains(t, buf.String(), `WhenDoneInfo`)
			assert.NotContains(t, buf.String(), `key2: 321`)
			assert.NotContains(t, buf.String(), log.KeyNameDuration+`: 25`)
		}
	}
}

func TestWriteTypes_Nest(t *testing.T) {
	buf := &bytes.Buffer{}
	wt := log.WriteTypes{W: buf}

	if err := wt.Nest("nestedKey", func(kv log.KeyValuer) error {
		kv.AddBool("nbool", true)
		kv.AddInt("nint", 3)
		kv.AddObject("nobj", []string{"sl1", "sl2"})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	assert.Exactly(t, " nestedKey:  nbool: true nint: 3 nobj: []string{\"sl1\", \"sl2\"}", buf.String())
}

func TestWriteTypes_Nest_EmptyKey(t *testing.T) {
	buf := &bytes.Buffer{}
	wt := log.WriteTypes{W: buf}

	if err := wt.Nest("", func(kv log.KeyValuer) error {
		kv.AddBool("nbool", true)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	assert.Exactly(t, " _:  nbool: true", buf.String())
}

func TestWriteTypes_Nest_Error(t *testing.T) {
	buf := &bytes.Buffer{}
	wt := log.WriteTypes{W: buf}

	err := wt.Nest("nestedKey", func(kv log.KeyValuer) error {
		return errors.New("NestErr")
	})
	assert.Exactly(t, " nestedKey: ", buf.String())
	assert.EqualError(t, err, "[log] WriteType.Nest.f: NestErr")
}

// Copyright 2015-present, Cyrill @ Schumacher.fm and the CoreStore contributors
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

package logmem_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/corestoreio/log"
	"github.com/corestoreio/log/logmem"
	"github.com/corestoreio/log/logw"
	"github.com/corestoreio/pkg/util/assert"
)

func TestPeriodically(t *testing.T) {
	var buf bytes.Buffer
	l := logw.NewLog(
		logw.WithWriter(&buf),
		logw.WithLevel(logw.LevelInfo),
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	logmem.Periodically(ctx, l, time.Second*2, 1024)

	{
		// generate some noise. maybe there are better ways but for now this seems
		// enough for demo purposes.
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			_, _ = w.Write(bytes.Repeat([]byte(`Response `), 100))
		}))
		c := srv.Client()
		for i := 0; i < 100; i++ {
			resp, err := c.Get(srv.URL)
			assert.NoError(t, err)
			_ = resp
		}
		srv.CloseClientConnections()
		srv.Close()
	}
	time.Sleep(time.Second * 11)

	logmem.Now(l, "[testlogmem]", log.String("testkey", "testvalue"))

	t.Log("\n", buf.String())
	assert.True(t, bytes.Count(buf.Bytes(), []byte(`[logmem]`)) >= 5, "Should find at least five log entries")
	assert.True(t, bytes.Count(buf.Bytes(), []byte(`[testlogmem]`)) == 1, "Should find at least one log entry for testlogmem")
	assert.True(t, bytes.Count(buf.Bytes(), []byte(`testkey`)) == 1, "Should find at least one log entry for testkey")
	assert.True(t, bytes.Count(buf.Bytes(), []byte(`testvalue`)) == 1, "Should find at least one log entry for testvalue")
}

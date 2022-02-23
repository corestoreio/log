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

package logmem

import (
	"context"
	"math"
	"runtime"
	"sync"
	"time"

	"github.com/corestoreio/log"
)

// Periodically logs the memory statistic in a defined interval. It calls `Now`
// after the interval duration has been passed. The context cancels the internal
// goroutine. minDiffBytes sets the minimum amount of bytes between the current
// and previous runtime.ReadMemStat call. For example setting minDiffBytes to
// 128MB logs every 128MB change the memory data. This function is meant to be
// run during the whole life time of a program to track is memory consumption.
// If Info level in the logger has been disabled, no logging will happen. Debug
// level logs the termination of the goroutine.
func Periodically(ctx context.Context, l log.Logger, interval time.Duration, minDiffBytes uint64, fields ...log.Field) {
	hasChanges := func(current, previous uint64) bool {
		return previous-current > minDiffBytes
	}

	go func() {
		tick := time.NewTicker(interval)
		defer tick.Stop()
		var prev runtime.MemStats
		var cur runtime.MemStats
		runtime.ReadMemStats(&prev)
		for {
			select {
			case <-tick.C:
				if ctx.Err() != nil {
					if l.IsDebug() {
						l.Debug("logmem.Periodically.terminated.error", log.Err(ctx.Err()))
					}
					return
				}
				runtime.ReadMemStats(&cur)

				if l.IsInfo() && (hasChanges(cur.HeapAlloc, prev.HeapAlloc) || hasChanges(cur.Sys, prev.Sys)) {
					logMemoryStats(l, prev, cur, "[logmem] memory diff", fields...)
				}
				prev = cur
			case <-ctx.Done():
				if l.IsDebug() {
					l.Debug("logmem.Periodically.terminated.done")
				}
				return
			}
		}
	}()
}

type statistician struct {
	mu   sync.Mutex
	prev runtime.MemStats
	cur  runtime.MemStats
}

var stats = new(statistician)

// Now logs the difference of the memory consumption between the first call and
// subsequent calls of Now. It stores the previous state of runtime.ReadMemStats
// in a global variable. If Info level in the logger has been disabled, no
// logging will happen.
func Now(l log.Logger, message string, fields ...log.Field) {
	if l.IsInfo() {
		stats.mu.Lock()
		runtime.ReadMemStats(&stats.cur)
		stats.prev = logMemoryStats(l, stats.prev, stats.cur, message, fields...)
		stats.mu.Unlock()
	}
}

func logMemoryStats(l log.Logger, prev, cur runtime.MemStats, message string, fields ...log.Field) runtime.MemStats {
	nolo := cur.Mallocs - cur.Frees
	if nolo > math.MaxUint64-cur.Mallocs { // check overflow
		nolo = 0
	}

	noloPrev := prev.Mallocs - prev.Frees
	if noloPrev > math.MaxUint64-prev.Mallocs { // check overflow
		noloPrev = 0
	}

	// Add more fields if desired.
	l.Info(message,
		append(log.Fields{
			log.Float64("alloc_mb", toMB(cur.Alloc)),
			log.Float64("heap_alloc_mb", toMB(cur.HeapAlloc)),
			log.Uint64("heap_objects", cur.HeapObjects),
			log.Float64("sys_mb", toMB(cur.Sys)),
			log.Uint64("number_of_live_objects", nolo),
			log.Uint("num_gc", uint(cur.NumGC)),

			log.Float64("diff_alloc_mb", toMB(cur.Alloc-prev.Alloc)),
			log.Float64("diff_heap_alloc_mb", toMB(cur.HeapAlloc-prev.HeapAlloc)),
			log.Uint64("diff_heap_objects", cur.HeapObjects-prev.HeapObjects),
			log.Float64("diff_sys_mb", toMB(cur.Sys-prev.Sys)),
			log.Uint64("diff_number_of_live_objects", noloPrev),
		}, fields...)...,
	)

	return cur
}

func toMB(b uint64) float64 {
	return math.Round(float64(b)/1024/1024*1000) / 1000
}

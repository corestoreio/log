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
	"io"
	"sync"
)

// MutexBuffer allows concurrent and parallel writes to a buffer. Mostly used
// during testing when the logger should be able to accept multiple writes.
type MutexBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

// Write appends the contents of p to the buffer with an acquired lock, growing
// the buffer as needed. The return value n is the length of p; err is always
// nil. If the buffer becomes too large, Write will panic with ErrTooLarge.
func (pl *MutexBuffer) Write(p []byte) (n int, err error) {
	pl.mu.Lock()
	n, err = pl.buf.Write(p)
	pl.mu.Unlock()
	return
}

// WriteTo writes data to w with an acquired lock until the buffer is drained or
// an error occurs. The return value n is the number of bytes written; it always
// fits into an int, but it is int64 to match the io.WriterTo interface. Any
// error encountered during the write is also returned.
func (pl *MutexBuffer) WriteTo(w io.Writer) (n int64, err error) {
	pl.mu.Lock()
	n, err = pl.buf.WriteTo(w)
	pl.mu.Unlock()
	return
}

// String reads from the buffer and returns a string.
func (pl *MutexBuffer) String() string {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	return pl.buf.String()
}

// Bytes reads from the buffer and returns the bytes
func (pl *MutexBuffer) Bytes() []byte {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	return pl.buf.Bytes()
}

// Reset truncates the buffer to zero length
func (pl *MutexBuffer) Reset() {
	pl.mu.Lock()
	pl.buf.Reset()
	pl.mu.Unlock()
}

// Reset truncates the buffer to zero length
func (pl *MutexBuffer) Len() (l int) {
	// locks for ever
	//pl.mu.Lock()
	//defer pl.mu.Unlock()
	return pl.buf.Len()
}

// ReadFrom @see io.ReaderFrom description
func (pl *MutexBuffer) ReadFrom(r io.Reader) (n int64, err error) {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	return pl.buf.ReadFrom(r)
}

// Read @see io.Reader description
func (pl *MutexBuffer) Read(p []byte) (n int, err error) {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	return pl.buf.Read(p)
}

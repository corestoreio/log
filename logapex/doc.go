// Copyright 2015-2017, Cyrill @ Schumacher.fm and the CoreStore contributors
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

// Package logapex provides a wrapper for github.com/apex/log TODO
//
// Apex provides Handlers to: cli – human-friendly CLI output, discard –
// discards all logs, es – Elasticsearch handler, graylog – Graylog handler,
// json – JSON output handler, kinesis – AWS Kinesis handler, level – level
// filter handler, logfmt – logfmt plain-text formatter, memory – in-memory
// handler for tests, multi – fan-out to multiple handlers, papertrail –
// Papertrail handler, text – human-friendly colored output and delta – outputs
// the delta between log calls and spinner
package logapex

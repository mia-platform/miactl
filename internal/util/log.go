// Copyright Mia srl
// SPDX-License-Identifier: Apache-2.0
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

package util

import (
	"fmt"
	"io"

	"github.com/go-logr/logr"
)

var _ logr.LogSink = &stdSink{}

var LogLevel int

type stdSink struct {
	name      string
	writer    io.Writer
	callDepth int
}

func (sink *stdSink) Init(info logr.RuntimeInfo) {
	sink.callDepth = info.CallDepth
}

func (sink *stdSink) WithName(name string) logr.LogSink {
	return &stdSink{
		name:      fmt.Sprintf("%s.%s", sink.name, name),
		writer:    sink.writer,
		callDepth: sink.callDepth,
	}
}

func (sink *stdSink) WithValues(_ ...any) logr.LogSink {
	// TODO: actually get and use the values
	return &stdSink{
		name:      sink.name,
		writer:    sink.writer,
		callDepth: sink.callDepth,
	}
}

func (sink *stdSink) Enabled(level int) bool {
	return LogLevel >= level
}

func (sink *stdSink) Error(err error, msg string, kvs ...any) {
	// TODO: handle error as an additional value when we will support them
	newMsg := fmt.Sprintf("%s: %s", msg, err)
	sink.Info(0, newMsg, kvs...)
}

func (sink *stdSink) Info(_ int, msg string, _ ...any) {
	fmt.Fprintf(sink.writer, "%s: %s", sink.name, msg)
	fmt.Fprintln(sink.writer)
}

func NewLogger(w io.Writer) logr.Logger {
	sink := &stdSink{
		name:   "miactl",
		writer: w,
	}

	return logr.New(sink)
}

func NewTestLogger(w io.Writer, level int) logr.Logger {
	sink := &testSink{
		stdSink: stdSink{
			name:   "test",
			writer: w,
		},
		level: level,
	}

	return logr.New(sink)
}

var _ logr.LogSink = &testSink{}

type testSink struct {
	stdSink
	level int
}

func (sink *testSink) Enabled(level int) bool {
	return sink.level >= level
}

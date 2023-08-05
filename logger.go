// Copyright 2021 Tetrate
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

// Package log provides an implementation of the telemetry.Logger interface
// that allows defining named loggers
package log

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/tetratelabs/telemetry"
	"github.com/tetratelabs/telemetry/function"
)

// compile time check for compatibility with the telemetry.Logger interface.
var _ telemetry.Logger = (*logger)(nil)

// logger is an implementation of the telemetry.Logger that allows configuring named
// loggers that can be configured independently and referenced by name.
type logger struct {
	telemetry.Logger
	// writer where the log messages will be written
	writer io.Writer
	// now is a function used to obtain the timestamps for the logs
	now func() time.Time
}

// New creates a new structured logger.
func New() telemetry.Logger {
	lg := &logger{
		writer: os.Stdout,
		now:    time.Now,
	}
	lg.Logger = function.NewLogger(lg.structuredLog)
	return lg
}

// NewFlattened creates a new flattened logger for structured data.
func NewFlattened() telemetry.Logger {
	lg := &logger{
		writer: os.Stdout,
		now:    time.Now,
	}
	lg.Logger = function.NewLogger(lg.flattenedLog)
	return lg
}

// NewUnstructured creates a new unstructured logger.
// NOTE: This logger is not to be used for new development as it goes against
// the contract of the telemetry.Logger interface. It expects Printf style data
// from calls to Error, Info, and Debug. It has been added for converting legacy
// code to use this new logging subsystem.
//
//	Example: log.Debug("values should be between %d and %d.", 12, 24)
//
// If you want to have a normal style log output but provide proper structured
// data, please use the flattened logger by instantiating NewFlattened().
func NewUnstructured() telemetry.Logger {
	lg := &logger{
		writer: os.Stdout,
		now:    time.Now,
	}
	lg.Logger = function.NewLogger(lg.unstructuredLog)
	return lg
}

// structuredLog emits the structured log at the given level
func (l *logger) structuredLog(level telemetry.Level, msg string, err error, values function.Values) {
	var (
		t   = l.now()
		out bytes.Buffer
	)

	_, _ = out.WriteString(
		fmt.Sprintf(`time="%d/%02d/%02d %02d:%02d:%02d" level=%s msg=%q`,
			t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second(),
			level, msg))

	writeArgs(&out, values.FromContext)
	writeArgs(&out, values.FromLogger)
	writeArgs(&out, values.FromMethod)

	if err != nil {
		_, _ = out.WriteString(" error=\"")
		_, _ = out.WriteString(err.Error())
		_, _ = out.WriteString("\"")
	}

	_, _ = out.WriteString("\n")
	_, _ = out.WriteTo(l.writer)
}

// unstructuredLog emits the unstructured log at the given level
func (l *logger) unstructuredLog(level telemetry.Level, msg string, err error, values function.Values) {
	var (
		t   = l.now()
		out bytes.Buffer
	)

	_, _ = out.WriteString(fmt.Sprintf("%d/%02d/%02d %02d:%02d:%02d  %-5v  ",
		t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second(), level))

	if err != nil {
		values.FromMethod = append(values.FromMethod, err)
	}
	_, _ = out.WriteString(fmt.Sprintf(msg, values.FromMethod...))

	nCtx := len(values.FromContext)
	nArgs := len(values.FromLogger)

	if nCtx > 0 { // write the context args without a Leading whitespace
		_, _ = out.WriteString(" [")
		writeKeyValue(&out, values.FromContext, 0, nCtx)
		writeArgs(&out, values.FromContext[2:])
	}

	if nArgs > 0 {
		if nCtx > 0 {
			writeArgs(&out, values.FromLogger)
		} else { // write the logger args without a Leading whitespace
			_, _ = out.WriteString(" [")
			writeKeyValue(&out, values.FromLogger, 0, nArgs)
			writeArgs(&out, values.FromLogger[2:])
		}
	}

	if nCtx > 0 || nArgs > 0 {
		_, _ = out.WriteString("]")
	}

	_, _ = out.WriteString("\n")
	_, _ = out.WriteTo(l.writer)
}

// writeArgs writes the list of arguments to the buffer
func writeArgs(b *bytes.Buffer, args []interface{}) {
	n := len(args)
	for i := 0; i < n; i += 2 {
		if _, ok := args[i].(string); !ok {
			continue
		}
		_, _ = b.WriteString(" ")
		writeKeyValue(b, args, i, n)
	}
}

// writeKeyValue writes a key value pair to the buffer
func writeKeyValue(b *bytes.Buffer, args []interface{}, i, n int) {
	k := args[i].(string) // precondition: this has already been checked, and it is safe to cast
	_, _ = b.WriteString(k)
	_, _ = b.WriteString("=")
	if i+1 >= n {
		_, _ = b.WriteString("\"(MISSING)\"")
		return
	}

	if v, ok := args[i+1].(string); ok {
		_, _ = b.WriteString(fmt.Sprintf("%q", v))
	} else {
		_, _ = b.WriteString(fmt.Sprintf("%v", args[i+1]))
	}
}

// flattenedLog emits the structured log as a flattened log at the given level
func (l *logger) flattenedLog(level telemetry.Level, msg string, err error, values function.Values) {
	var (
		t   = l.now()
		out bytes.Buffer
	)

	_, _ = out.WriteString(fmt.Sprintf("%d/%02d/%02d %02d:%02d:%02d  %-5v  %s",
		t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second(), level, msg))

	nCtx := len(values.FromContext)
	nWith := len(values.FromLogger)
	nMeth := len(values.FromMethod)

	if err != nil || nCtx > 0 || nWith > 0 || nMeth > 0 {
		if err != nil {
			_, _ = out.WriteString(fmt.Sprintf(" [error=%q", err.Error()))
			if nCtx > 0 || nWith > 0 || nMeth > 0 {
				_, _ = out.WriteString(" ")
			}
		} else {
			_, _ = out.WriteString(" [")
		}

		writeArgs2(&out, values.FromContext, values.FromLogger, values.FromMethod)

		_, _ = out.WriteString("]")

	}

	_, _ = out.WriteString("\n")
	_, _ = out.WriteTo(l.writer)
}

// writeArgs2 writes the collection of argument lists to the buffer.
func writeArgs2(b *bytes.Buffer, args ...[]interface{}) {
	t := len(args)
	firstValue := true
	for i := 0; i < t; i++ {
		n := len(args[i])
		if n%2 == 1 {
			args[i] = append(args[i], "(MISSING)")
			n++
		}
		for j := 0; j < n; j += 2 {
			if !firstValue {
				_, _ = b.WriteString(" ")
			}
			firstValue = false
			if k, ok := args[i][j].(string); ok {
				_, _ = b.WriteString(k + "=")
			} else {
				_, _ = b.WriteString(fmt.Sprintf("%v=", args[i][j]))
			}
			if v, ok := args[i][j+1].(string); ok {
				_, _ = b.WriteString(fmt.Sprintf("%q", v))
			} else {
				_, _ = b.WriteString(fmt.Sprintf("%v", args[i][j+1]))
			}
		}
	}
}

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

// Copyright (c) Tetrate, Inc 2021 All Rights Reserved.

package log

import (
	"bytes"
	"fmt"

	"github.com/tetratelabs/telemetry"
)

// unstructuredFormatString is the string used to format logging messages in the logger.
const unstructuredFormatString = "%d/%02d/%02d %02d:%02d:%02d  %-5v\t%-10s\t"

// newUnstructured creates a new unstructured logger. It is meant for internal use only.
// To instantiate a new logger use the log.register method.
func newUnstructured(name, description string) *Logger {
	logger := newLogger(name, description)
	logger.emit = unstructuredLog
	return logger
}

// unstructuredLog emits the unstructured log at the given level
func unstructuredLog(l *Logger, level Level, msg string, err error, keyValues ...interface{}) {
	t := l.now()

	var out bytes.Buffer
	_, _ = out.WriteString(fmt.Sprintf(unstructuredFormatString,
		t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second(),
		level, l.name))

	if err != nil {
		keyValues = append(keyValues, err)
	}
	_, _ = out.WriteString(fmt.Sprintf(msg, keyValues...))

	fromCtx := telemetry.KeyValuesFromContext(l.ctx)
	nCtx := len(fromCtx)
	nArgs := len(l.args)

	if nCtx > 0 { // write the context args without a Leading whitespace
		_, _ = out.WriteString(" [")
		writeKeyValue(&out, fromCtx, 0, nCtx)
		writeArgs(&out, fromCtx[2:])
	}

	if nArgs > 0 {
		if nCtx > 0 {
			writeArgs(&out, l.args)
		} else { // write the logger args without a Leading whitespace
			_, _ = out.WriteString(" [")
			writeKeyValue(&out, l.args, 0, nArgs)
			writeArgs(&out, l.args[2:])
		}
	}

	if nCtx > 0 || nArgs > 0 {
		_, _ = out.WriteString("]")
	}

	_, _ = out.WriteString("\n")
	_, _ = out.WriteTo(l.writer)
}

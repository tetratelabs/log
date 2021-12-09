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
	"time"

	"github.com/tetratelabs/telemetry"
)

// unstructuredFormatString is the string used to format logging messages in the logger.
const unstructuredFormatString = "%d/%02d/%02d %02d:%02d:%02d  %-5v\t%-10s\t%v"

// newUnstructured creates a new unstructured logger. It is meant for internal use only.
// To instantiate a new logger use the log.register method.
func newUnstructured(name, description string) *Logger {
	logger := newLogger(name, description)
	logger.emit = unstructuredLog
	return logger
}

func unstructuredLog(l *Logger, level Level, msg string, err error, keyValues ...interface{}) {
	t := time.Now()

	// merge contextual args
	contextualArgs := append([]interface{}{}, telemetry.KeyValuesFromContext(l.ctx)...)
	contextualArgs = append(contextualArgs, l.args...)

	// args for the unstructured message format string
	args := append([]interface{}{}, keyValues...)
	if err != nil {
		args = append(args, err)
	}

	var out bytes.Buffer
	_, _ = out.WriteString(fmt.Sprintf(unstructuredFormatString,
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(),
		level, l.name, fmt.Sprintf(msg, args...)))

	if len(contextualArgs) > 0 {
		_, _ = out.WriteString(" [")
		for i := 0; i < len(contextualArgs); i += 2 {
			if k, ok := contextualArgs[i].(string); ok {
				if v, ok := contextualArgs[i+1].(string); ok {
					_, _ = out.WriteString(fmt.Sprintf(` %s=%q`, k, v))
				} else {
					_, _ = out.WriteString(fmt.Sprintf(` %s=%v`, k, contextualArgs[i+1]))
				}
			}
		}
		_, _ = out.WriteString(" ]")
	}

	_, _ = out.WriteString("\n")

	_, _ = fmt.Fprintln(l.writer, out.String())
}

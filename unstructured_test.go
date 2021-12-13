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

package log

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/tetratelabs/telemetry"
)

func TestUnstructured(t *testing.T) {
	tests := []struct {
		name        string
		level       Level
		logfunc     func(telemetry.Logger)
		expected    *regexp.Regexp
		metricCount float64
		skipContext bool
	}{
		{"unstructured-none", LevelNone, func(l telemetry.Logger) { l.Error("text", errors.New("error")) }, regexp.MustCompile("^$"), 1, false},
		{"unstructured-disabled-info", LevelNone, func(l telemetry.Logger) { l.Info("text") }, regexp.MustCompile("^$"), 1, false},
		{"unstructured-disabled-debug", LevelNone, func(l telemetry.Logger) { l.Debug("text") }, regexp.MustCompile("^$"), 0, false},
		{"unstructured-disabled-error", LevelNone, func(l telemetry.Logger) { l.Error("text", errors.New("error")) }, regexp.MustCompile("^$"), 1, false},
		{"unstructured-info", LevelInfo, func(l telemetry.Logger) { l.Info("hi %s", "there") },
			matchUnstructured("unstructured-info", LevelInfo, "hi there"), 1, false},
		{"unstructured-error", LevelInfo, func(l telemetry.Logger) { l.Error("text %s: %v", errors.New("error"), "there") },
			matchUnstructured("unstructured-error", LevelError, "text there: error"), 1, false},
		{"unstructured-debug", LevelDebug, func(l telemetry.Logger) { l.Debug("hi %s", "there") },
			matchUnstructured("unstructured-debug", LevelDebug, "hi there"), 0, false},
		{"unstructured-nocontext", LevelInfo, func(l telemetry.Logger) { l.Info("hi %s", "there") },
			matchUnstructuredNoCtx("unstructured-nocontext", LevelInfo, "hi there"), 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := RegisterUnstructured(tt.name, "test logger")
			logger.SetLevel(tt.level)
			if logger.Name() != tt.name {
				t.Fatalf("loger.Name()=%s, want: %s", logger.Name(), tt.name)
			}
			if logger.Description() != "test logger" {
				t.Fatalf(`loger.Description()=%s, want: "test logger""`, logger.Description())
			}
			if logger.Level() != tt.level {
				t.Fatalf("loger.Level()=%s, want: %s", logger.Level(), tt.level)
			}

			// Overwrite the output of the loggers to check the output messages
			var out bytes.Buffer
			logger.writer = &out

			metric := mockMetric{}
			if !tt.skipContext {
				ctx := telemetry.KeyValuesToContext(context.Background(), "ctx", "value")
				logger = logger.Context(ctx).(*Logger)
			}

			l := logger.Metric(&metric).With().With(1, "").With("lvl", LevelInfo).With("missing")

			tt.logfunc(l)

			str := out.String()
			if !tt.expected.MatchString(str) {
				t.Fatalf("expected %v to match %s", str, tt.expected)
			}
			if metric.count != tt.metricCount {
				t.Fatalf("metric.count=%v, want %v", metric.count, tt.metricCount)
			}
		})
	}
}

func BenchmarkUnstructuredLog0Args(b *testing.B) {
	benchmarkLogger(b, 0, newUnstructured, unstructuredLog)
}

func BenchmarkUnstructuredLog3Args(b *testing.B) {
	benchmarkLogger(b, 1, newUnstructured, unstructuredLog)
}

func BenchmarkUnstructuredLog9Args(b *testing.B) {
	benchmarkLogger(b, 3, newUnstructured, unstructuredLog)
}

func BenchmarkUnstructuredLog15Args(b *testing.B) {
	benchmarkLogger(b, 5, newUnstructured, unstructuredLog)
}

func BenchmarkUnstructuredLog30Args(b *testing.B) {
	benchmarkLogger(b, 10, newUnstructured, unstructuredLog)
}

func matchUnstructured(n string, l Level, msg string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf("^%s  %-5v\t%-10s\\t%s \\[ctx=\"value\" lvl=info missing=\"\\(MISSING\\)\"\\]\\n$", rprefix, l, n, msg))
}

func matchUnstructuredNoCtx(n string, l Level, msg string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf("^%s  %-5v\t%-10s\\t%s \\[lvl=info missing=\"\\(MISSING\\)\"\\]\\n$", rprefix, l, n, msg))
}

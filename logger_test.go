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
	"io"
	"regexp"
	"testing"

	"github.com/tetratelabs/telemetry"
	"github.com/tetratelabs/telemetry/function"
)

func TestLogger(t *testing.T) {
	tests := []struct {
		name        string
		level       telemetry.Level
		logfunc     func(telemetry.Logger)
		expected    *regexp.Regexp
		metricCount float64
	}{
		{"none", telemetry.LevelNone, func(l telemetry.Logger) { l.Error("text", errors.New("error")) }, regexp.MustCompile("^$"), 1},
		{"disabled-info", telemetry.LevelNone, func(l telemetry.Logger) { l.Info("text") }, regexp.MustCompile("^$"), 1},
		{"disabled-debug", telemetry.LevelNone, func(l telemetry.Logger) { l.Debug("text") }, regexp.MustCompile("^$"), 0},
		{"disabled-error", telemetry.LevelNone, func(l telemetry.Logger) { l.Error("text", errors.New("error")) }, regexp.MustCompile("^$"), 1},
		{"info", telemetry.LevelInfo, func(l telemetry.Logger) { l.Info("text") }, match(telemetry.LevelInfo, ""), 1},
		{"info-missing", telemetry.LevelInfo, func(l telemetry.Logger) { l.Info("text", "where") }, match(telemetry.LevelInfo, ` where="\(MISSING\)"`), 1},
		{"info-with-values", telemetry.LevelInfo, func(l telemetry.Logger) { l.Info("text", "where", "there", 1, "1") },
			match(telemetry.LevelInfo, ` where="there"`), 1},
		{"error", telemetry.LevelInfo, func(l telemetry.Logger) { l.Error("text", errors.New("error")) }, match(telemetry.LevelError, ` error="error"`), 1},
		{"error-missing", telemetry.LevelInfo, func(l telemetry.Logger) { l.Error("text", errors.New("error"), "where") },
			match(telemetry.LevelError, ` where="\(MISSING\)" error="error"`), 1},
		{"error-with-values", telemetry.LevelInfo, func(l telemetry.Logger) { l.Error("text", errors.New("error"), "where", "there", 1, "1") },
			match(telemetry.LevelError, ` where="there" error="error"`), 1},
		{"debug", telemetry.LevelDebug, func(l telemetry.Logger) { l.Debug("text") }, match(telemetry.LevelDebug, ""), 0},
		{"debug-missing", telemetry.LevelDebug, func(l telemetry.Logger) { l.Debug("text", "where") }, match(telemetry.LevelDebug, ` where="\(MISSING\)"`), 0},
		{"debug-with-values", telemetry.LevelDebug, func(l telemetry.Logger) { l.Debug("text", "where", "there", 1, "1") },
			match(telemetry.LevelDebug, ` where="there"`), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New()

			l.SetLevel(tt.level)
			if l.Level() != tt.level {
				t.Fatalf("loger.Level()=%s, want: %s", l.Level(), tt.level)
			}

			// Overwrite the output of the loggers to check the output messages
			var out bytes.Buffer
			l.(*logger).writer = &out

			metric := mockMetric{}
			ctx := telemetry.KeyValuesToContext(context.Background(), "ctx", "value")
			l = l.Context(ctx).Metric(&metric).With().With(1, "").With("lvl", telemetry.LevelInfo).With("missing")

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

func TestUnstructured(t *testing.T) {
	tests := []struct {
		name        string
		level       telemetry.Level
		logfunc     func(telemetry.Logger)
		expected    *regexp.Regexp
		metricCount float64
		skipContext bool
	}{
		{"unstructured-none", telemetry.LevelNone, func(l telemetry.Logger) { l.Error("text", errors.New("error")) }, regexp.MustCompile("^$"), 1, false},
		{"unstructured-disabled-info", telemetry.LevelNone, func(l telemetry.Logger) { l.Info("text") }, regexp.MustCompile("^$"), 1, false},
		{"unstructured-disabled-debug", telemetry.LevelNone, func(l telemetry.Logger) { l.Debug("text") }, regexp.MustCompile("^$"), 0, false},
		{"unstructured-disabled-error", telemetry.LevelNone, func(l telemetry.Logger) { l.Error("text", errors.New("error")) }, regexp.MustCompile("^$"), 1, false},
		{"unstructured-info", telemetry.LevelInfo, func(l telemetry.Logger) { l.Info("hi %s", "there") },
			matchUnstructured(telemetry.LevelInfo, "hi there"), 1, false},
		{"unstructured-error", telemetry.LevelInfo, func(l telemetry.Logger) { l.Error("text %s: %v", errors.New("error"), "there") },
			matchUnstructured(telemetry.LevelError, "text there: error"), 1, false},
		{"unstructured-debug", telemetry.LevelDebug, func(l telemetry.Logger) { l.Debug("hi %s", "there") },
			matchUnstructured(telemetry.LevelDebug, "hi there"), 0, false},
		{"unstructured-nocontext", telemetry.LevelInfo, func(l telemetry.Logger) { l.Info("hi %s", "there") },
			matchUnstructuredNoCtx(telemetry.LevelInfo, "hi there"), 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewUnstructured()

			l.SetLevel(tt.level)
			if l.Level() != tt.level {
				t.Fatalf("loger.Level()=%s, want: %s", l.Level(), tt.level)
			}

			// Overwrite the output of the loggers to check the output messages
			var out bytes.Buffer
			l.(*logger).writer = &out

			metric := mockMetric{}
			if !tt.skipContext {
				ctx := telemetry.KeyValuesToContext(context.Background(), "ctx", "value")
				l = l.Context(ctx)
			}

			l = l.Metric(&metric).With().With(1, "").With("lvl", telemetry.LevelInfo).With("missing")

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

func BenchmarkStructuredLog0Args(b *testing.B) {
	l := New().(*logger)
	benchmarkLogger(b, 0, l, l.structuredLog)
}

func BenchmarkStructuredLog3Args(b *testing.B) {
	l := New().(*logger)
	benchmarkLogger(b, 1, l, l.structuredLog)
}

func BenchmarkStructuredLog9Args(b *testing.B) {
	l := New().(*logger)
	benchmarkLogger(b, 3, l, l.structuredLog)
}

func BenchmarkStructuredLog15Args(b *testing.B) {
	l := New().(*logger)
	benchmarkLogger(b, 5, l, l.structuredLog)
}

func BenchmarkStructuredLog30Args(b *testing.B) {
	l := New().(*logger)
	benchmarkLogger(b, 10, l, l.structuredLog)
}

func BenchmarkUnstructuredLog0Args(b *testing.B) {
	l := New().(*logger)
	benchmarkLogger(b, 0, l, l.unstructuredLog)
}

func BenchmarkUnstructuredLog3Args(b *testing.B) {
	l := New().(*logger)
	benchmarkLogger(b, 1, l, l.unstructuredLog)
}

func BenchmarkUnstructuredLog9Args(b *testing.B) {
	l := New().(*logger)
	benchmarkLogger(b, 3, l, l.unstructuredLog)
}

func BenchmarkUnstructuredLog15Args(b *testing.B) {
	l := New().(*logger)
	benchmarkLogger(b, 5, l, l.unstructuredLog)
}

func BenchmarkUnstructuredLog30Args(b *testing.B) {
	l := New().(*logger)
	benchmarkLogger(b, 10, l, l.unstructuredLog)
}

func benchmarkLogger(b *testing.B, nargs int, l *logger, logFunc function.Emit) {
	var (
		ctx    = context.Background()
		args   = make([]interface{}, 0, nargs*2)
		method = make([]interface{}, 0, nargs*2)
	)

	l.writer = io.Discard

	for i := 0; i < nargs; i++ {
		ctx = telemetry.KeyValuesToContext(ctx, fmt.Sprintf("ctx%d", i), i)
		args = append(args, fmt.Sprintf("with%d", i), i)
		method = append(method, fmt.Sprintf("arg%d", i), i)
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logFunc(telemetry.LevelInfo, "bench", nil, function.Values{
			FromContext: telemetry.KeyValuesFromContext(ctx),
			FromLogger:  args,
			FromMethod:  method,
		})
	}
}

const (
	rdate   = `[0-9]{4}/[0-9]{2}/[0-9]{2}` // matches the date part of a log
	rtime   = `[0-9]{2}:[0-9]{2}:[0-9]{2}` // matches the time part of a log
	rprefix = rdate + " " + rtime          // log prefix match
)

func match(l telemetry.Level, keyvalues string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf("^time=%q level=%s msg=\"text\" ctx=\"value\" lvl=info missing=\"\\(MISSING\\)\"%s\\n$", rprefix, l, keyvalues))
}

func matchUnstructured(l telemetry.Level, msg string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf("^%s  %-5v  %s \\[ctx=\"value\" lvl=info missing=\"\\(MISSING\\)\"\\]\\n$", rprefix, l, msg))
}

func matchUnstructuredNoCtx(l telemetry.Level, msg string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf("^%s  %-5v  %s \\[lvl=info missing=\"\\(MISSING\\)\"\\]\\n$", rprefix, l, msg))
}

type mockMetric struct {
	telemetry.Metric
	count float64
}

func (m *mockMetric) RecordContext(_ context.Context, value float64) { m.count += value }

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

func TestLogger(t *testing.T) {
	tests := []struct {
		name        string
		level       Level
		logfunc     func(telemetry.Logger)
		expected    *regexp.Regexp
		metricCount float64
	}{
		{"none", LevelNone, func(l telemetry.Logger) { l.Error("text", errors.New("error")) }, regexp.MustCompile("^$"), 1},
		{"disabled-info", LevelNone, func(l telemetry.Logger) { l.Info("text") }, regexp.MustCompile("^$"), 1},
		{"disabled-debug", LevelNone, func(l telemetry.Logger) { l.Debug("text") }, regexp.MustCompile("^$"), 0},
		{"disabled-error", LevelNone, func(l telemetry.Logger) { l.Error("text", errors.New("error")) }, regexp.MustCompile("^$"), 1},
		{"info", LevelInfo, func(l telemetry.Logger) { l.Info("text") }, match("info", LevelInfo, ""), 1},
		{"info-missing", LevelInfo, func(l telemetry.Logger) { l.Info("text", "where") }, match("info-missing", LevelInfo, ` where="\(MISSING\)"`), 1},
		{"info-with-values", LevelInfo, func(l telemetry.Logger) { l.Info("text", "where", "there", 1, "1") },
			match("info-with-values", LevelInfo, ` where="there"`), 1},
		{"error", LevelInfo, func(l telemetry.Logger) { l.Error("text", errors.New("error")) }, match("error", LevelError, ` error="error"`), 1},
		{"error-missing", LevelInfo, func(l telemetry.Logger) { l.Error("text", errors.New("error"), "where") },
			match("error-missing", LevelError, ` where="\(MISSING\)" error="error"`), 1},
		{"error-with-values", LevelInfo, func(l telemetry.Logger) { l.Error("text", errors.New("error"), "where", "there", 1, "1") },
			match("error-with-values", LevelError, ` where="there" error="error"`), 1},
		{"debug", LevelDebug, func(l telemetry.Logger) { l.Debug("text") }, match("debug", LevelDebug, ""), 0},
		{"debug-missing", LevelDebug, func(l telemetry.Logger) { l.Debug("text", "where") }, match("debug-missing", LevelDebug, ` where="\(MISSING\)"`), 0},
		{"debug-with-values", LevelDebug, func(l telemetry.Logger) { l.Debug("text", "where", "there", 1, "1") },
			match("debug-with-values", LevelDebug, ` where="there"`), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := Register(tt.name, "test logger")
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
			ctx := telemetry.KeyValuesToContext(context.Background(), "ctx", "value")
			l := logger.Context(ctx).Metric(&metric).With().With(1, "").With("lvl", LevelInfo).With("missing")

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

func TestAsLevel(t *testing.T) {
	tests := []struct {
		level string
		want  Level
		ok    bool
	}{
		{"none", LevelNone, true},
		{"error", LevelError, true},
		{"info", LevelInfo, true},
		{"debug", LevelDebug, true},
		{"invalid", LevelNone, false},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			level, ok := AsLevel(tt.level)

			if level != tt.want {
				t.Fatalf("AsLevel(%s)=%s, want: %s", tt.level, level, tt.want)
			}
			if ok != tt.ok {
				t.Fatalf("AsLevel(%s)=%t, want: %t", tt.level, ok, tt.ok)
			}
		})
	}
}

func TestSetLevel(t *testing.T) {
	logger := Register("level", "test logger")
	withvalues := logger.With("key", "value").(*Logger)
	logger.SetLevel(LevelDebug)

	if withvalues.Level() != LevelDebug {
		t.Fatalf("logger.Level()=%v, want: %v", withvalues.Level(), LevelDebug)
	}
}

const (
	rdate   = `[0-9]{4}/[0-9]{2}/[0-9]{2}` // matches the date part of a log
	rtime   = `[0-9]{2}:[0-9]{2}:[0-9]{2}` // matches the time part of a log
	rprefix = rdate + " " + rtime          // log prefix match
)

func match(n string, l Level, keyvalues string) *regexp.Regexp {
	return regexp.MustCompile(
		fmt.Sprintf("^time=%q level=%s scope=%q msg=\"text\" ctx=\"value\" lvl=info missing=\"\\(MISSING\\)\"%s\\n$",
			rprefix, l, n, keyvalues))
}

type mockMetric struct {
	telemetry.Metric
	count float64
}

func (m *mockMetric) RecordContext(_ context.Context, value float64) { m.count += value }

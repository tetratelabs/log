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
	"context"
	"errors"
	"testing"

	"github.com/tetratelabs/telemetry"
)

func TestDiscard(t *testing.T) {
	tests := []struct {
		name        string
		logfunc     func(telemetry.Logger)
		metricCount float64
	}{
		{"info-", func(l telemetry.Logger) { l.Info("text", "where", "there") }, 1},
		{"error", func(l telemetry.Logger) { l.Error("text", errors.New("error"), "where", "there") }, 1},
		{"debug", func(l telemetry.Logger) { l.Debug("text", "where", "there") }, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewDiscard()

			metric := mockMetric{}
			ctx := telemetry.KeyValuesToContext(context.Background(), "ctx", "value")
			l := logger.Context(ctx).Metric(&metric).With().With("lvl", LevelInfo).With("missing")

			tt.logfunc(l)

			if metric.count != tt.metricCount {
				t.Fatalf("metric.count=%v, want %v", metric.count, tt.metricCount)
			}
		})
	}
}

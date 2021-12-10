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

	"github.com/tetratelabs/telemetry"
)

// compile time check for compatibility with the telemetry.Logger interface.
var _ telemetry.Logger = (*discard)(nil)

// discard is a logger that does not emit log messages but still emits metrics.
type discard struct {
	// ctx holds the Context to extract key-value pairs from to be added to each
	// log line.
	ctx context.Context
	// metric holds the Metric to increment each time Info() or Error() is called.
	metric telemetry.Metric
}

// Discard returns a nre logger that discards log messages but still emits the corresponding metrics, if set.
func Discard() telemetry.Logger {
	return &discard{ctx: context.Background()}
}

// Info emits the configured metric, if any.
func (d *discard) Info(string, ...interface{}) {
	// even if we don't output the log line due to the level configuration,
	// we always emit the Metric if it is set.
	if d.metric != nil {
		d.metric.RecordContext(d.ctx, 1)
	}
}

// Error emits the configured metric, if any.
func (d *discard) Error(string, error, ...interface{}) {
	// even if we don't output the log line due to the level configuration,
	// we always emit the Metric if it is set.
	if d.metric != nil {
		d.metric.RecordContext(d.ctx, 1)
	}
}

// Debug is a noop
func (d *discard) Debug(string, ...interface{}) {

}

// With is a noop
func (d *discard) With(...interface{}) telemetry.Logger {
	return d
}

// Context attaches provided Context to the Logger allowing metadata found in
// this context to be used for metrics labels.
func (d *discard) Context(ctx context.Context) telemetry.Logger {
	return &discard{ctx: ctx, metric: d.metric}
}

// Metric attaches provided Metric to the Logger allowing this metric to
// record each invocation of Info and Error log lines. If context is available
// in the logger, it can be used for Metrics labels.
func (d *discard) Metric(m telemetry.Metric) telemetry.Logger {
	return &discard{ctx: d.ctx, metric: m}
}

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
var _ telemetry.Logger = (*Discard)(nil)

// Discard is a logger that does not emit log messages but still emits metrics.
type Discard struct {
	// ctx holds the Context to extract key-value pairs from to be added to each
	// log line.
	ctx context.Context
	// metric holds the Metric to increment each time Info() or Error() is called.
	metric telemetry.Metric
}

func (d *Discard) Info(string, ...interface{}) {
	// even if we don't output the log line due to the level configuration,
	// we always emit the Metric if it is set.
	if d.metric != nil {
		d.metric.RecordContext(d.ctx, 1)
	}
}

func (d *Discard) Error(string, error, ...interface{}) {
	// even if we don't output the log line due to the level configuration,
	// we always emit the Metric if it is set.
	if d.metric != nil {
		d.metric.RecordContext(d.ctx, 1)
	}
}

func (d *Discard) Debug(string, ...interface{}) {

}

func (d *Discard) With(...interface{}) telemetry.Logger {
	return d
}

func (d *Discard) Context(ctx context.Context) telemetry.Logger {
	return &Discard{ctx: ctx, metric: d.metric}
}

func (d *Discard) Metric(m telemetry.Metric) telemetry.Logger {
	return &Discard{ctx: d.ctx, metric: m}
}

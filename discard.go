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

// Discard is a noop logger. It does not emit log message and does not record metric values.
// Adding values to the logger binding it to a Context will do nothing.
var Discard ScopedLogger = &discard{}

// discardName is the name of the discarding logger.
const discardName = "/dev/null"

type discard struct{}

func (discard) Name() string                                { return discardName }
func (discard) Info(string, ...interface{})                 {}
func (discard) Error(string, error, ...interface{})         {}
func (discard) Debug(string, ...interface{})                {}
func (d *discard) With(...interface{}) telemetry.Logger     { return d }
func (d *discard) Context(context.Context) telemetry.Logger { return d }
func (d *discard) Metric(telemetry.Metric) telemetry.Logger { return d }

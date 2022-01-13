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
	"time"

	"github.com/tetratelabs/telemetry"
)

func mockTime() time.Time {
	return time.Date(2021, 12, 9, 17, 37, 46, 0, time.UTC) // 2021/12/09 17:37:46
}

func ExampleLogger() {
	log := New()
	log.(*logger).now = mockTime // Mock time to have a consistent output

	// Normal and error logging
	log.Info("an info message with values", "key", "value")
	log.Error("validation error", errors.New("validation failed"), "arg1", "invalid-value")

	// Changing log levels at runtime
	log.Debug("a debug message")
	log.SetLevel(telemetry.LevelDebug)
	log.Debug("an enabled debug message")

	// Propagating values
	ctx := telemetry.KeyValuesToContext(context.Background(), "request-id", 123)
	log.Context(ctx).With("component", "middleware").Info("enriched message")

	// Output:
	// time="2021/12/09 17:37:46" level=info msg="an info message with values" key="value"
	// time="2021/12/09 17:37:46" level=error msg="validation error" arg1="invalid-value" error="validation failed"
	// time="2021/12/09 17:37:46" level=debug msg="an enabled debug message"
	// time="2021/12/09 17:37:46" level=info msg="enriched message" request-id=123 component="middleware"
}

func ExampleLogger_unstructured() {
	unstructured := NewUnstructured()
	unstructured.(*logger).now = mockTime // Mock time to have a consistent output

	// Normal and error logging
	unstructured.Info("an info message with %s", "a value")
	unstructured.Error("validation error in %s: %v", errors.New("validation failed"), "arg1")

	// Changing log levels at runtime
	unstructured.Debug("a debug message")
	unstructured.SetLevel(telemetry.LevelDebug)
	unstructured.Debug("an enabled debug message")

	// Propagating values
	ctx := telemetry.KeyValuesToContext(context.Background(), "request-id", 123)
	unstructured.Context(ctx).With("component", "middleware").Info("enriched message")

	// Output:
	// 2021/12/09 17:37:46  info   an info message with a value
	// 2021/12/09 17:37:46  error  validation error in arg1: validation failed
	// 2021/12/09 17:37:46  debug  an enabled debug message
	// 2021/12/09 17:37:46  info   enriched message [request-id=123 component="middleware"]
}

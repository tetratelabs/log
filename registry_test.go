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

import "testing"

func TestLoggers(t *testing.T) {
	custom := Register("custom-logger", "test logger")

	// Verify loggers are registered
	if GetLogger("custom-logger") == nil {
		t.Fatal("custom-logger was not registered")
	}
	if GetLogger("unexisting") != nil {
		t.Fatal("unexisting logger was not expected to be registered")
	}

	// Verify logger list
	found := false
	for _, l := range Loggers() {
		if l.Name() == custom.Name() {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("custom-logger not found in the registered loggers")
	}

	// Verify that re-registering a logger returns the existing one
	custom2 := Register(custom.Name(), "another description")
	if custom != custom2 { // Compare the pointers
		t.Fatal("Register returned a new logger")
	}
}

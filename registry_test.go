// Copyright (c) Tetrate, Inc 2021 All Rights Reserved.

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

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
	"sync"

	"github.com/tetratelabs/telemetry"
)

// Scope is a named logger that can be configured independently
type Scope interface {
	telemetry.Logger
	Name() string
}

// Register a new logger with the given name.
// If the logger already exists, the already registered logger is returned.
func Register(name, description string) Scope {
	if l, ok := reg.GetLogger(name); ok {
		return l
	}

	l := newLogger(name, description)
	reg.Register(l)

	return l
}

// RegisterUnstructured a new unstructured logger with the given name.
// If the logger already exists, the already registered logger is returned.
func RegisterUnstructured(name, description string) Scope {
	if l, ok := reg.GetLogger(name); ok {
		return l
	}

	l := newUnstructured(name, description)
	reg.Register(l)

	return l
}

// Loggers returns the list of all registered loggers.
func Loggers() []Scope { return reg.Loggers() }

// GetLogger gets a registered logger by name.
func GetLogger(name string) Scope {
	logger, _ := reg.GetLogger(name)
	return logger
}

// reg is a registry of all registered loggers.
var reg = &registry{loggers: make(map[string]Scope)}

// registry keeps track of all registered loggers.
type registry struct {
	mu      sync.RWMutex
	loggers map[string]Scope
}

// Loggers returns the list of all registered loggers.
func (r *registry) Loggers() []Scope {
	r.mu.RLock()
	defer r.mu.RUnlock()

	loggers := make([]Scope, 0, len(r.loggers))
	for _, l := range r.loggers {
		loggers = append(loggers, l)
	}

	return loggers
}

// GetLogger a registered logger by name.
func (r *registry) GetLogger(name string) (Scope, bool) {
	r.mu.RLock()
	logger, ok := r.loggers[name]
	r.mu.RUnlock()

	if !ok {
		return NewDiscard(), false
	}

	return logger, true
}

// Register a logger into the registry.
func (r *registry) Register(logger Scope) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.loggers[logger.Name()] = logger
}

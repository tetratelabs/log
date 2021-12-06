// Copyright (c) Tetrate, Inc 2021 All Rights Reserved.

package log

import (
	"sync"
)

// Register a new logger with the given name.
// If the logger already exists, the already registered logger is returned.
func Register(name, description string) *Logger {
	l := reg.GetLogger(name)
	if l != nil {
		return l
	}

	l = newLogger(name, description)
	reg.Register(l)

	return l
}

// Loggers returns the list of all registered loggers.
func Loggers() []*Logger { return reg.Loggers() }

// GetLogger gets a registered logger by name.
func GetLogger(name string) *Logger { return reg.GetLogger(name) }

// reg is a registry of all registered loggers.
var reg = &registry{loggers: make(map[string]*Logger)}

// registry keeps track of all registered loggers.
type registry struct {
	mu      sync.RWMutex
	loggers map[string]*Logger
}

// Loggers returns the list of all registered loggers.
func (r *registry) Loggers() []*Logger {
	r.mu.RLock()
	defer r.mu.RUnlock()

	loggers := make([]*Logger, 0, len(r.loggers))
	for _, l := range r.loggers {
		loggers = append(loggers, l)
	}

	return loggers
}

// GetLogger a registered logger by name.
func (r *registry) GetLogger(name string) *Logger {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.loggers[name]
}

// Register a logger into the registry.
func (r *registry) Register(logger *Logger) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.loggers[logger.Name()] = logger
}

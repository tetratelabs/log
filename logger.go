// Copyright (c) Tetrate, Inc 2021 All Rights Reserved.

// Package log provides an implementation of the telemetry.Logger interface
// that allows defining named loggers
package log

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"

	"github.com/tetratelabs/telemetry"
)

// compile time check for compatibility with the telemetry.Logger interface.
var _ telemetry.Logger = (*Logger)(nil)

// Logger is an implementation of the telemetry.Logger that allows configuring named
// loggers that can be configured independently and referenced by name.
type Logger struct {
	// ctx holds the Context to extract key-value pairs from to be added to each
	// log line.
	ctx context.Context
	// args holds the key-value pairs to be added to each log line.
	args []interface{}
	// metric holds the Metric to increment each time Info() or Error() is called.
	metric telemetry.Metric
	// name of the scoped logger
	name string
	// description for the scoped logger.
	description string
	// level holds the configured log level.
	level int32
	// writer where the log messages will be written
	writer io.Writer
}

// newLogger creates a new logger. It is meant for internal use only.
// To instantiate a new logger use the log.register method.
func newLogger(name, description string) *Logger {
	return &Logger{
		ctx:         context.Background(),
		name:        name,
		description: description,
		level:       int32(LevelInfo),
		writer:      os.Stdout,
	}
}

// Name returns the name of the logger
func (l *Logger) Name() string { return l.name }

// Description of the logger
func (l *Logger) Description() string { return l.description }

// Level returns the logging level configured for this logger.
func (l *Logger) Level() Level { return Level(atomic.LoadInt32(&l.level)) }

// SetLevel configures the logging level for the logger.
func (l *Logger) SetLevel(level Level) { atomic.StoreInt32(&l.level, int32(level)) }

// enabled checks if the current logger should emit log messages for the given
// logging level.
func (l *Logger) enabled(level Level) bool { return level <= l.Level() }

// Debug emits a log message at debug level with the given key value pairs.
func (l *Logger) Debug(msg string, keyValues ...interface{}) {
	if !l.enabled(LevelDebug) {
		return
	}

	args := append([]interface{}{}, telemetry.KeyValuesFromContext(l.ctx)...)
	args = append(args, l.args...)
	args = append(args, keyValues...)
	if len(keyValues)%2 != 0 {
		args = append(args, "(MISSING)")
	}

	l.log(LevelDebug, msg, args...)
}

// Info emits a log message at info level with the given key value pairs.
func (l *Logger) Info(msg string, keyValues ...interface{}) {
	// even if we don't output the log line due to the level configuration,
	// we always emit the Metric if it is set.
	if l.metric != nil {
		l.metric.RecordContext(l.ctx, 1)
	}

	if !l.enabled(LevelInfo) {
		return
	}

	args := append([]interface{}{}, telemetry.KeyValuesFromContext(l.ctx)...)
	args = append(args, l.args...)
	args = append(args, keyValues...)
	if len(keyValues)%2 != 0 {
		args = append(args, "(MISSING)")
	}

	l.log(LevelInfo, msg, args...)
}

// Error emits a log message at error level with the given key value pairs.
// The given error will be used as the last parameter in the message format
// string.
func (l *Logger) Error(msg string, err error, keyValues ...interface{}) {
	// even if we don't output the log line due to the level configuration,
	// we always emit the Metric if it is set.
	if l.metric != nil {
		l.metric.RecordContext(l.ctx, 1)
	}

	if !l.enabled(LevelError) {
		return
	}

	args := append([]interface{}{}, telemetry.KeyValuesFromContext(l.ctx)...)
	args = append(args, l.args...)
	args = append(args, keyValues...)
	if len(keyValues)%2 != 0 {
		args = append(args, "(MISSING)")
	}
	args = append(args, "error", err.Error())

	l.log(LevelError, msg, args...)
}

// log the message at the given level
func (l *Logger) log(level Level, msg string, keyValues ...interface{}) {
	t := time.Now()

	var out bytes.Buffer
	_, _ = out.WriteString(fmt.Sprintf(`time="%d/%02d/%02d %02d:%02d:%02d"`,
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second()))
	_, _ = out.WriteString(fmt.Sprintf(" level=%v", level))
	_, _ = out.WriteString(fmt.Sprintf(" scope=%q", l.name))
	_, _ = out.WriteString(fmt.Sprintf(" msg=%q", msg))

	for i := 0; i < len(keyValues); i += 2 {
		if k, ok := keyValues[i].(string); ok {
			if v, ok := keyValues[i+1].(string); ok {
				_, _ = out.WriteString(fmt.Sprintf(` %s=%q`, k, v))
			} else {
				_, _ = out.WriteString(fmt.Sprintf(` %s=%v`, k, keyValues[i+1]))
			}
		}
	}

	_, _ = fmt.Fprintln(l.writer, out.String())
}

// With returns Logger with provided key value pairs attached.
func (l *Logger) With(keyValues ...interface{}) telemetry.Logger {
	if len(keyValues) == 0 {
		return l
	}
	if len(keyValues)%2 != 0 {
		keyValues = append(keyValues, "(MISSING)")
	}

	newLogger := &Logger{
		args:        make([]interface{}, len(l.args), len(l.args)+len(keyValues)),
		ctx:         l.ctx,
		metric:      l.metric,
		level:       atomic.LoadInt32(&l.level),
		writer:      l.writer,
		name:        l.name,
		description: l.description,
	}

	copy(newLogger.args, l.args)

	for i := 0; i < len(keyValues); i += 2 {
		if k, ok := keyValues[i].(string); ok {
			newLogger.args = append(newLogger.args, k, keyValues[i+1])
		}
	}

	return newLogger
}

// KeyValuesToContext takes provided key-value pairs and places them in Context.
func (l *Logger) KeyValuesToContext(ctx context.Context, keyValues ...interface{}) context.Context {
	return telemetry.KeyValuesToContext(ctx, keyValues...)
}

// Context attaches provided Context to the Logger allowing metadata found in
// this context to be used for log lines and metrics labels.
func (l *Logger) Context(ctx context.Context) telemetry.Logger {
	newLogger := &Logger{
		args:        make([]interface{}, len(l.args), len(l.args)),
		ctx:         ctx,
		metric:      l.metric,
		level:       atomic.LoadInt32(&l.level),
		writer:      l.writer,
		name:        l.name,
		description: l.description,
	}

	copy(newLogger.args, l.args)

	return newLogger
}

// Metric attaches provided Metric to the Logger allowing this metric to
// record each invocation of Info and Error log lines. If context is available
// in the logger, it can be used for Metrics labels.
func (l *Logger) Metric(m telemetry.Metric) telemetry.Logger {
	newLogger := &Logger{
		args:        make([]interface{}, len(l.args), len(l.args)),
		ctx:         l.ctx,
		metric:      m,
		level:       atomic.LoadInt32(&l.level),
		writer:      l.writer,
		name:        l.name,
		description: l.description,
	}

	copy(newLogger.args, l.args)

	return newLogger
}

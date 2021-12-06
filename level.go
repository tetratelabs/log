// Copyright (c) Tetrate, Inc 2021 All Rights Reserved.

package log

// Level configures the verbosity of the logging framework and
// the messages that are emitted or ignored.
type Level int32

const (
	// LevelNone will silent the logger.
	LevelNone Level = iota
	// LevelError configures the logger to only emit error messages.
	LevelError
	// LevelInfo configures the logger to only emit info messages.
	LevelInfo
	// LevelDebug configures the logger to only emit debug messages.
	LevelDebug
)

// levelToString maps each logging level to its string representation.
var levelToString = map[Level]string{
	LevelNone:  "none",
	LevelError: "error",
	LevelInfo:  "info",
	LevelDebug: "debug",
}

// levelToString maps string representations to the corresponding level
var stringToLevel = map[string]Level{
	"none":  LevelNone,
	"error": LevelError,
	"info":  LevelInfo,
	"debug": LevelDebug,
}

// String returns the string representation of the logging level.
func (l Level) String() string { return levelToString[l] }

// AsLevel returns the logging level corresponding to the given string representation.
func AsLevel(level string) (Level, bool) {
	l, ok := stringToLevel[level]
	return l, ok
}

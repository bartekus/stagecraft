// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - A Go-based CLI for orchestrating local-first multi-service deployments using Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package logging

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Feature: CORE_LOGGING
// Spec: spec/core/logging.md

// Level represents a log level.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// String returns the string representation of the log level.
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging.
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	WithFields(fields ...Field) Logger
}

// Field represents a key-value pair in structured logging.
type Field struct {
	Key   string
	Value interface{}
}

// NewField creates a new field.
func NewField(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// loggerImpl is the default logger implementation.
type loggerImpl struct {
	level  Level
	out    io.Writer
	errOut io.Writer
	fields []Field
}

// NewLogger creates a new logger.
// If verbose is true, Debug level logs are shown.
func NewLogger(verbose bool) Logger {
	level := LevelInfo
	if verbose {
		level = LevelDebug
	}

	return &loggerImpl{
		level:  level,
		out:    os.Stdout,
		errOut: os.Stderr,
		fields: []Field{},
	}
}

// Debug logs a debug message.
func (l *loggerImpl) Debug(msg string, fields ...Field) {
	if l.level <= LevelDebug {
		l.log(LevelDebug, msg, fields...)
	}
}

// Info logs an info message.
func (l *loggerImpl) Info(msg string, fields ...Field) {
	if l.level <= LevelInfo {
		l.log(LevelInfo, msg, fields...)
	}
}

// Warn logs a warning message.
func (l *loggerImpl) Warn(msg string, fields ...Field) {
	if l.level <= LevelWarn {
		l.log(LevelWarn, msg, fields...)
	}
}

// Error logs an error message (always shown).
func (l *loggerImpl) Error(msg string, fields ...Field) {
	l.log(LevelError, msg, fields...)
}

// WithFields returns a new logger with additional fields.
func (l *loggerImpl) WithFields(fields ...Field) Logger {
	return &loggerImpl{
		level:  l.level,
		out:    l.out,
		errOut: l.errOut,
		fields: append(l.fields, fields...),
	}
}

// log writes a log message.
func (l *loggerImpl) log(level Level, msg string, fields ...Field) {
	writer := l.out
	if level == LevelError {
		writer = l.errOut
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	prefix := fmt.Sprintf("[%s] %s: ", timestamp, level.String())

	// Combine base fields with message fields
	allFields := append(l.fields, fields...)

	// Format message
	if len(allFields) > 0 {
		fieldStrs := make([]string, 0, len(allFields))
		for _, f := range allFields {
			fieldStrs = append(fieldStrs, fmt.Sprintf("%s=%v", f.Key, f.Value))
		}
		fmt.Fprintf(writer, "%s%s %s\n", prefix, msg, fmt.Sprintf("(%s)", fmt.Sprint(fieldStrs)))
	} else {
		fmt.Fprintf(writer, "%s%s\n", prefix, msg)
	}
}


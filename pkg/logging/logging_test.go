// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package logging

import (
	"bytes"
	"strings"
	"testing"
)

// Feature: CORE_LOGGING
// Spec: spec/core/logging.md

func TestLogger_Levels(t *testing.T) {
	var buf bytes.Buffer
	logger := &loggerImpl{
		level:  LevelInfo,
		out:    &buf,
		errOut: &buf,
		fields: []Field{},
	}

	logger.Debug("debug message")
	if buf.Len() > 0 {
		t.Errorf("expected no output for debug at Info level, got: %q", buf.String())
	}

	buf.Reset()
	logger.Info("info message")
	if !strings.Contains(buf.String(), "INFO") {
		t.Errorf("expected INFO in output, got: %q", buf.String())
	}

	buf.Reset()
	logger.Warn("warn message")
	if !strings.Contains(buf.String(), "WARN") {
		t.Errorf("expected WARN in output, got: %q", buf.String())
	}

	buf.Reset()
	logger.Error("error message")
	if !strings.Contains(buf.String(), "ERROR") {
		t.Errorf("expected ERROR in output, got: %q", buf.String())
	}
}

func TestLogger_Verbose(t *testing.T) {
	var buf bytes.Buffer
	logger := &loggerImpl{
		level:  LevelDebug,
		out:    &buf,
		errOut: &buf,
		fields: []Field{},
	}

	logger.Debug("debug message")
	if !strings.Contains(buf.String(), "DEBUG") {
		t.Errorf("expected DEBUG in output when verbose, got: %q", buf.String())
	}
}

func TestLogger_WithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := &loggerImpl{
		level:  LevelInfo,
		out:    &buf,
		errOut: &buf,
		fields: []Field{},
	}

	logger = logger.WithFields(NewField("env", "prod"), NewField("version", "1.0.0")).(*loggerImpl)
	logger.Info("deploying")

	output := buf.String()
	if !strings.Contains(output, "env=prod") {
		t.Errorf("expected 'env=prod' in output, got: %q", output)
	}
	if !strings.Contains(output, "version=1.0.0") {
		t.Errorf("expected 'version=1.0.0' in output, got: %q", output)
	}
}

func TestNewLogger(t *testing.T) {
	logger := NewLogger(false)
	if logger == nil {
		t.Fatalf("expected non-nil logger")
	}

	verboseLogger := NewLogger(true)
	if verboseLogger == nil {
		t.Fatalf("expected non-nil logger")
	}
}

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{Level(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		if got := tt.level.String(); got != tt.expected {
			t.Errorf("Level(%d).String() = %q, want %q", tt.level, got, tt.expected)
		}
	}
}

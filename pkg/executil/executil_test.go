// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Feature: CORE_EXECUTIL
// Spec: spec/core/executil.md

package executil

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestNewRunner(t *testing.T) {
	runner := NewRunner()
	if runner == nil {
		t.Fatal("NewRunner() returned nil")
	}
}

func TestNewCommand(t *testing.T) {
	cmd := NewCommand("echo", "hello", "world")
	if cmd.Name != "echo" {
		t.Errorf("expected Name to be 'echo', got %q", cmd.Name)
	}
	if len(cmd.Args) != 2 {
		t.Errorf("expected 2 args, got %d", len(cmd.Args))
	}
	if cmd.Args[0] != "hello" || cmd.Args[1] != "world" {
		t.Errorf("expected args ['hello', 'world'], got %v", cmd.Args)
	}
}

func TestRunner_Run_Success(t *testing.T) {
	runner := NewRunner()
	ctx := context.Background()

	var cmd Command
	if runtime.GOOS == "windows" {
		cmd = NewCommand("cmd", "/c", "echo", "test-output")
	} else {
		cmd = NewCommand("echo", "test-output")
	}

	result, err := runner.Run(ctx, cmd)
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}

	output := strings.TrimSpace(string(result.Stdout))
	if output != "test-output" {
		t.Errorf("expected stdout 'test-output', got %q", output)
	}
}

func TestRunner_Run_Failure(t *testing.T) {
	runner := NewRunner()
	ctx := context.Background()

	var cmd Command
	if runtime.GOOS == "windows" {
		cmd = NewCommand("cmd", "/c", "exit", "/b", "42")
	} else {
		cmd = NewCommand("sh", "-c", "exit 42")
	}

	result, err := runner.Run(ctx, cmd)
	if err == nil {
		t.Fatal("expected Run() to return error for non-zero exit code")
	}

	if result.ExitCode != 42 {
		t.Errorf("expected exit code 42, got %d", result.ExitCode)
	}

	// Check that error contains exit code information
	if !strings.Contains(err.Error(), "42") {
		t.Errorf("expected error to contain exit code, got: %v", err)
	}
}

func TestRunner_Run_CommandNotFound(t *testing.T) {
	runner := NewRunner()
	ctx := context.Background()

	cmd := NewCommand("nonexistent-command-that-does-not-exist-12345")

	_, err := runner.Run(ctx, cmd)
	if err == nil {
		t.Fatal("expected Run() to return error for non-existent command")
	}

	// Should be an execution error, not a command failure
	var execErr *exec.Error
	if !errors.As(err, &execErr) {
		t.Errorf("expected exec.Error, got: %T: %v", err, err)
	}
}

func TestRunner_Run_StderrCapture(t *testing.T) {
	runner := NewRunner()
	ctx := context.Background()

	var cmd Command
	if runtime.GOOS == "windows" {
		cmd = NewCommand("cmd", "/c", "echo", "error-output", ">&2")
	} else {
		cmd = NewCommand("sh", "-c", "echo 'error-output' >&2")
	}

	result, err := runner.Run(ctx, cmd)
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	stderr := strings.TrimSpace(string(result.Stderr))
	if !strings.Contains(stderr, "error-output") {
		t.Errorf("expected stderr to contain 'error-output', got %q", stderr)
	}
}

func TestRunner_Run_EnvironmentVariables(t *testing.T) {
	runner := NewRunner()
	ctx := context.Background()

	var cmd Command
	if runtime.GOOS == "windows" {
		cmd = NewCommand("cmd", "/c", "echo", "%TEST_VAR%")
		cmd.Env = map[string]string{"TEST_VAR": "test-value"}
	} else {
		cmd = NewCommand("sh", "-c", "echo $TEST_VAR")
		cmd.Env = map[string]string{"TEST_VAR": "test-value"}
	}

	result, err := runner.Run(ctx, cmd)
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	output := strings.TrimSpace(string(result.Stdout))
	if output != "test-value" {
		t.Errorf("expected stdout 'test-value', got %q", output)
	}
}

func TestRunner_Run_WorkingDirectory(t *testing.T) {
	runner := NewRunner()
	ctx := context.Background()

	// Get a temporary directory
	tmpDir := t.TempDir()

	var cmd Command
	if runtime.GOOS == "windows" {
		cmd = NewCommand("cmd", "/c", "cd")
	} else {
		cmd = NewCommand("pwd")
	}
	cmd.Dir = tmpDir

	result, err := runner.Run(ctx, cmd)
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	output := strings.TrimSpace(string(result.Stdout))
	// Normalize paths for comparison
	output = filepath.Clean(output)
	tmpDir = filepath.Clean(tmpDir)

	if output != tmpDir {
		t.Errorf("expected working directory %q, got %q", tmpDir, output)
	}
}

func TestRunner_Run_Stdin(t *testing.T) {
	runner := NewRunner()
	ctx := context.Background()

	var cmd Command
	if runtime.GOOS == "windows" {
		cmd = NewCommand("cmd", "/c", "more")
	} else {
		cmd = NewCommand("cat")
	}
	cmd.Stdin = strings.NewReader("input-data")

	result, err := runner.Run(ctx, cmd)
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	output := strings.TrimSpace(string(result.Stdout))
	if output != "input-data" {
		t.Errorf("expected stdout 'input-data', got %q", output)
	}
}

func TestRunner_Run_ContextCancellation(t *testing.T) {
	runner := NewRunner()
	ctx, cancel := context.WithCancel(context.Background())

	var cmd Command
	if runtime.GOOS == "windows" {
		cmd = NewCommand("cmd", "/c", "timeout", "/t", "10")
	} else {
		cmd = NewCommand("sleep", "10")
	}

	// Cancel context after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	_, err := runner.Run(ctx, cmd)
	if err == nil {
		t.Fatal("expected Run() to return error when context is cancelled")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got: %v", err)
	}
}

func TestRunner_RunStream_Success(t *testing.T) {
	runner := NewRunner()
	ctx := context.Background()

	var cmd Command
	if runtime.GOOS == "windows" {
		cmd = NewCommand("cmd", "/c", "echo", "stream-output")
	} else {
		cmd = NewCommand("echo", "stream-output")
	}

	var buf bytes.Buffer
	err := runner.RunStream(ctx, cmd, &buf)
	if err != nil {
		t.Fatalf("RunStream() returned error: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if output != "stream-output" {
		t.Errorf("expected output 'stream-output', got %q", output)
	}
}

func TestRunner_RunStream_Failure(t *testing.T) {
	runner := NewRunner()
	ctx := context.Background()

	var cmd Command
	if runtime.GOOS == "windows" {
		cmd = NewCommand("cmd", "/c", "exit", "/b", "1")
	} else {
		cmd = NewCommand("sh", "-c", "exit 1")
	}

	var buf bytes.Buffer
	err := runner.RunStream(ctx, cmd, &buf)
	if err == nil {
		t.Fatal("expected RunStream() to return error for non-zero exit code")
	}
}

func TestRunner_RunStream_ContextCancellation(t *testing.T) {
	runner := NewRunner()
	ctx, cancel := context.WithCancel(context.Background())

	var cmd Command
	if runtime.GOOS == "windows" {
		cmd = NewCommand("cmd", "/c", "timeout", "/t", "10")
	} else {
		cmd = NewCommand("sleep", "10")
	}

	// Cancel context after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	var buf bytes.Buffer
	err := runner.RunStream(ctx, cmd, &buf)
	if err == nil {
		t.Fatal("expected RunStream() to return error when context is cancelled")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got: %v", err)
	}
}

func TestRunner_RunStream_MultipleLines(t *testing.T) {
	runner := NewRunner()
	ctx := context.Background()

	var cmd Command
	if runtime.GOOS == "windows" {
		cmd = NewCommand("cmd", "/c", "(echo line1 & echo line2 & echo line3)")
	} else {
		cmd = NewCommand("sh", "-c", "echo 'line1'; echo 'line2'; echo 'line3'")
	}

	var buf bytes.Buffer
	err := runner.RunStream(ctx, cmd, &buf)
	if err != nil {
		t.Fatalf("RunStream() returned error: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d: %q", len(lines), output)
	}
}

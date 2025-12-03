// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Feature: CORE_EXECUTIL
// Spec: spec/core/executil.md

// Package executil provides utilities for executing external commands.
package executil

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Runner is an interface for executing commands.
type Runner interface {
	// Run executes a command and returns the result.
	// Returns an error if the command fails (non-zero exit code) or if execution fails.
	Run(ctx context.Context, cmd Command) (*Result, error)

	// RunStream executes a command and streams output to the provided writer.
	// Returns an error if the command fails (non-zero exit code) or if execution fails.
	RunStream(ctx context.Context, cmd Command, output io.Writer) error
}

// Command represents a command to execute.
type Command struct {
	Name  string
	Args  []string
	Dir   string
	Env   map[string]string
	Stdin io.Reader
}

// Result contains the result of a command execution.
type Result struct {
	ExitCode int
	Stdout   []byte
	Stderr   []byte
}

// runner is the default implementation of Runner.
type runner struct{}

// NewRunner creates a new Runner instance.
func NewRunner() Runner {
	return &runner{}
}

// NewCommand creates a new Command with the given name and arguments.
func NewCommand(name string, args ...string) Command {
	return Command{
		Name: name,
		Args: args,
	}
}

// Run executes a command and returns the result.
func (r *runner) Run(ctx context.Context, cmd Command) (*Result, error) { //nolint:gocritic // hugeParam: intentional for immutability
	//nolint:gosec // This package is designed to execute arbitrary commands;
	// validation should be done by callers.
	execCmd := exec.CommandContext(ctx, cmd.Name, cmd.Args...)

	// Set working directory if specified
	if cmd.Dir != "" {
		execCmd.Dir = cmd.Dir
	}

	// Set environment variables
	if len(cmd.Env) > 0 {
		execCmd.Env = os.Environ()
		for k, v := range cmd.Env {
			execCmd.Env = append(execCmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Set stdin if provided
	if cmd.Stdin != nil {
		execCmd.Stdin = cmd.Stdin
	}

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	// Execute the command
	err := execCmd.Run()

	result := &Result{
		ExitCode: execCmd.ProcessState.ExitCode(),
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
	}

	// Check for context cancellation first
	if ctx.Err() != nil {
		return result, fmt.Errorf("command cancelled: %w", ctx.Err())
	}

	// Check for execution errors (command not found, permission denied, etc.)
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// Command executed but returned non-zero exit code
			return result, fmt.Errorf("command failed with exit code %d: %w", result.ExitCode, err)
		}
		// Execution error (command not found, etc.)
		return result, fmt.Errorf("executing command: %w", err)
	}

	return result, nil
}

// RunStream executes a command and streams output to the provided writer.
func (r *runner) RunStream(ctx context.Context, cmd Command, output io.Writer) error { //nolint:gocritic // hugeParam: intentional for immutability
	//nolint:gosec // This package is designed to execute arbitrary commands;
	// validation should be done by callers.
	execCmd := exec.CommandContext(ctx, cmd.Name, cmd.Args...)

	// Set working directory if specified
	if cmd.Dir != "" {
		execCmd.Dir = cmd.Dir
	}

	// Set environment variables
	if len(cmd.Env) > 0 {
		execCmd.Env = os.Environ()
		for k, v := range cmd.Env {
			execCmd.Env = append(execCmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Set stdin if provided
	if cmd.Stdin != nil {
		execCmd.Stdin = cmd.Stdin
	}

	// Stream both stdout and stderr to the output writer
	execCmd.Stdout = output
	execCmd.Stderr = output

	// Execute the command
	err := execCmd.Run()

	// Check for context cancellation first
	if ctx.Err() != nil {
		return fmt.Errorf("command cancelled: %w", ctx.Err())
	}

	// Check for execution errors
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// Command executed but returned non-zero exit code
			return fmt.Errorf("command failed with exit code %d: %w", exitErr.ExitCode(), err)
		}
		// Execution error (command not found, etc.)
		return fmt.Errorf("executing command: %w", err)
	}

	return nil
}

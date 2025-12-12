// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

// Feature: DEPLOY_ROLLOUT
// Spec: spec/deploy/rollout.md
package deploy

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"stagecraft/pkg/executil"
)

type mockRunner struct {
	runFunc func(ctx context.Context, cmd executil.Command) (*executil.Result, error)
}

func (m *mockRunner) Run(ctx context.Context, cmd executil.Command) (*executil.Result, error) {
	if m.runFunc != nil {
		return m.runFunc(ctx, cmd)
	}
	return &executil.Result{ExitCode: 0}, nil
}

func (m *mockRunner) RunStream(ctx context.Context, cmd executil.Command, output io.Writer) error {
	result, err := m.Run(ctx, cmd)
	if err != nil {
		return err
	}
	if result != nil && len(result.Stdout) > 0 {
		_, _ = output.Write(result.Stdout)
	}
	return nil
}

func TestRolloutExecutor_IsAvailable_CommandFound(t *testing.T) {
	mock := &mockRunner{
		runFunc: func(ctx context.Context, cmd executil.Command) (*executil.Result, error) {
			if cmd.Name == "docker-rollout" && len(cmd.Args) > 0 && cmd.Args[0] == "--version" {
				return &executil.Result{ExitCode: 0}, nil
			}
			return nil, errors.New("unexpected command")
		},
	}

	executor := NewRolloutExecutorWithRunner(mock)
	available, err := executor.IsAvailable(context.Background())
	if err != nil {
		t.Fatalf("IsAvailable returned error: %v", err)
	}
	if !available {
		t.Error("IsAvailable returned false, expected true")
	}
}

func TestRolloutExecutor_IsAvailable_CommandNotFound(t *testing.T) {
	mock := &mockRunner{
		runFunc: func(ctx context.Context, cmd executil.Command) (*executil.Result, error) {
			return nil, errors.New("executable file not found")
		},
	}

	executor := NewRolloutExecutorWithRunner(mock)
	available, err := executor.IsAvailable(context.Background())
	if err != nil {
		t.Fatalf("IsAvailable returned error: %v", err)
	}
	if available {
		t.Error("IsAvailable returned true, expected false")
	}
}

func TestRolloutExecutor_IsAvailable_NonZeroExit(t *testing.T) {
	mock := &mockRunner{
		runFunc: func(ctx context.Context, cmd executil.Command) (*executil.Result, error) {
			return &executil.Result{ExitCode: 1}, nil
		},
	}

	executor := NewRolloutExecutorWithRunner(mock)
	available, err := executor.IsAvailable(context.Background())
	if err != nil {
		t.Fatalf("IsAvailable returned error: %v", err)
	}
	if available {
		t.Error("IsAvailable returned true, expected false")
	}
}

func TestRolloutExecutor_IsAvailable_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mock := &mockRunner{
		runFunc: func(ctx context.Context, cmd executil.Command) (*executil.Result, error) {
			return nil, ctx.Err()
		},
	}

	executor := NewRolloutExecutorWithRunner(mock)
	available, err := executor.IsAvailable(ctx)
	if err == nil {
		t.Error("IsAvailable should return error when context is cancelled")
	}
	if available {
		t.Error("IsAvailable returned true, expected false")
	}
}

func TestRolloutExecutor_Execute_Success(t *testing.T) {
	mock := &mockRunner{
		runFunc: func(ctx context.Context, cmd executil.Command) (*executil.Result, error) {
			if cmd.Name == "docker-rollout" && len(cmd.Args) > 1 && cmd.Args[0] == "up" && cmd.Args[1] == "-f" {
				return &executil.Result{ExitCode: 0}, nil
			}
			return nil, errors.New("unexpected command")
		},
	}

	executor := NewRolloutExecutorWithRunner(mock)
	err := executor.Execute(context.Background(), "/path/to/compose.yml")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
}

func TestRolloutExecutor_Execute_Failure(t *testing.T) {
	mock := &mockRunner{
		runFunc: func(ctx context.Context, cmd executil.Command) (*executil.Result, error) {
			return &executil.Result{
				ExitCode: 1,
				Stderr:   []byte("rollout failed"),
			}, nil
		},
	}

	executor := NewRolloutExecutorWithRunner(mock)
	err := executor.Execute(context.Background(), "/path/to/compose.yml")
	if err == nil {
		t.Error("Execute should return error on non-zero exit")
	}
	if !strings.Contains(err.Error(), "rollout failed") {
		t.Errorf("Error should contain stderr, got: %v", err)
	}
}

func TestRolloutExecutor_Execute_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mock := &mockRunner{
		runFunc: func(ctx context.Context, cmd executil.Command) (*executil.Result, error) {
			return nil, ctx.Err()
		},
	}

	executor := NewRolloutExecutorWithRunner(mock)
	err := executor.Execute(ctx, "/path/to/compose.yml")
	if err == nil {
		t.Error("Execute should return error when context is cancelled")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Error should be context.Canceled, got: %v", err)
	}
}

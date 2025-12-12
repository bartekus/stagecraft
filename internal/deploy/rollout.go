// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

package deploy

import (
	"context"
	"fmt"

	"stagecraft/pkg/executil"
)

// Feature: DEPLOY_ROLLOUT
// Spec: spec/deploy/rollout.md

// RolloutNotInstalledMessage is returned when rollout is enabled but docker-rollout is not installed.
// Spec: spec/deploy/rollout.md (Error message format)
const RolloutNotInstalledMessage = "docker-rollout is required but not installed; install it from the docker-rollout repository"

// RolloutExecutor executes docker-rollout deployments.
type RolloutExecutor struct {
	runner executil.Runner
}

// NewRolloutExecutor creates a new rollout executor.
func NewRolloutExecutor() *RolloutExecutor {
	return &RolloutExecutor{
		runner: executil.NewRunner(),
	}
}

// NewRolloutExecutorWithRunner allows injecting runner for tests.
func NewRolloutExecutorWithRunner(runner executil.Runner) *RolloutExecutor {
	return &RolloutExecutor{
		runner: runner,
	}
}

// IsAvailable checks if docker-rollout is installed.
// Returns (available, error).
// Error is returned only for context cancellation/deadline exceeded.
func (e *RolloutExecutor) IsAvailable(ctx context.Context) (bool, error) {
	cmd := executil.NewCommand("docker-rollout", "--version")
	result, err := e.runner.Run(ctx, cmd)

	// Check context after Run returns
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	if err != nil {
		// Command not found or exec error -> not available
		return false, nil
	}

	return result.ExitCode == 0, nil
}

// Execute runs docker-rollout up.
func (e *RolloutExecutor) Execute(ctx context.Context, composePath string) error {
	cmd := executil.NewCommand("docker-rollout", "up", "-f", composePath)
	result, err := e.runner.Run(ctx, cmd)

	if ctx.Err() != nil {
		return ctx.Err()
	}

	if err != nil {
		return fmt.Errorf("running docker-rollout: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("docker-rollout failed with exit code %d: %s",
			result.ExitCode, string(result.Stderr))
	}

	return nil
}

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
	"strings"
	"testing"

	"stagecraft/pkg/executil"
)

// TestRolloutExecutor_RolloutEnabledButNotAvailable verifies that when docker-rollout is not available,
// IsAvailable returns (false, nil).
func TestRolloutExecutor_RolloutEnabledButNotAvailable(t *testing.T) {
	mock := &mockRunner{
		runFunc: func(ctx context.Context, cmd executil.Command) (*executil.Result, error) {
			// Simulate docker-rollout not found
			if cmd.Name == "docker-rollout" {
				return nil, errors.New("executable file not found")
			}
			return &executil.Result{ExitCode: 0}, nil
		},
	}

	executor := NewRolloutExecutorWithRunner(mock)
	available, err := executor.IsAvailable(context.Background())
	if err != nil {
		t.Fatalf("IsAvailable should not return error for missing command: %v", err)
	}
	if available {
		t.Error("IsAvailable should return false when command not found")
	}
}

// TestRolloutNotInstalledMessage_ConformsToSpec verifies the exported constant matches the v1 spec:
// - no raw URLs
// - includes an actionable install hint
// The CLI uses this constant when rollout is enabled but docker-rollout is unavailable.
func TestRolloutNotInstalledMessage_ConformsToSpec(t *testing.T) {
	msg := RolloutNotInstalledMessage

	// Verify message contains required fragment
	if !strings.Contains(msg, "docker-rollout is required") {
		t.Error("Error message must contain 'docker-rollout is required'")
	}

	// Verify no URL in message (spec requirement)
	if strings.Contains(msg, "http://") || strings.Contains(msg, "https://") {
		t.Error("Error message must not contain URLs")
	}

	// Verify actionable hint is present
	if !strings.Contains(msg, "install it") {
		t.Error("Error message should include actionable installation hint")
	}
}

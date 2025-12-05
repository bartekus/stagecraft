// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"stagecraft/internal/core"
	"stagecraft/internal/core/state"
	"stagecraft/pkg/logging"
)

// Feature: CLI_PHASE_EXECUTION_COMMON
// Spec: spec/core/phase-execution-common.md

// TestExecutePhasesCommon_AllSuccess tests the happy path where all phases succeed.
func TestExecutePhasesCommon_AllSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, ".stagecraft", "releases.json")

	// Ensure .stagecraft directory exists
	if err := os.MkdirAll(filepath.Dir(stateFile), 0o700); err != nil {
		t.Fatalf("failed to create .stagecraft directory: %v", err)
	}

	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	ctx := context.Background()
	stateMgr := state.NewManager(stateFile)
	logger := logging.NewLogger(false)

	// Create a release
	release, err := stateMgr.CreateRelease(ctx, "staging", "v1.0.0", "commit1")
	if err != nil {
		t.Fatalf("failed to create release: %v", err)
	}

	// Track phase execution order
	var executionOrder []state.ReleasePhase

	// Create PhaseFns that record calls and succeed
	fns := PhaseFns{
		Build: func(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
			executionOrder = append(executionOrder, state.PhaseBuild)
			return nil
		},
		Push: func(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
			executionOrder = append(executionOrder, state.PhasePush)
			return nil
		},
		MigratePre: func(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
			executionOrder = append(executionOrder, state.PhaseMigratePre)
			return nil
		},
		Rollout: func(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
			executionOrder = append(executionOrder, state.PhaseRollout)
			return nil
		},
		MigratePost: func(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
			executionOrder = append(executionOrder, state.PhaseMigratePost)
			return nil
		},
		Finalize: func(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
			executionOrder = append(executionOrder, state.PhaseFinalize)
			return nil
		},
	}

	// Create a minimal plan
	plan := &core.Plan{
		Operations: []core.Operation{},
	}

	// Execute phases
	err = executePhasesCommon(ctx, stateMgr, release.ID, plan, logger, fns)
	if err != nil {
		t.Fatalf("executePhasesCommon should succeed, got: %v", err)
	}

	// Verify execution order
	expectedOrder := []state.ReleasePhase{
		state.PhaseBuild,
		state.PhasePush,
		state.PhaseMigratePre,
		state.PhaseRollout,
		state.PhaseMigratePost,
		state.PhaseFinalize,
	}

	if len(executionOrder) != len(expectedOrder) {
		t.Fatalf("expected %d phases to execute, got %d", len(expectedOrder), len(executionOrder))
	}

	for i, expected := range expectedOrder {
		if executionOrder[i] != expected {
			t.Errorf("expected phase %d to be %q, got %q", i, expected, executionOrder[i])
		}
	}

	// Verify all phases are completed
	updatedRelease, err := stateMgr.GetRelease(ctx, release.ID)
	if err != nil {
		t.Fatalf("failed to get release: %v", err)
	}

	for _, phase := range expectedOrder {
		status, ok := updatedRelease.Phases[phase]
		if !ok {
			t.Errorf("expected phase %q to be present", phase)
			continue
		}
		if status != state.StatusCompleted {
			t.Errorf("expected phase %q to be %q, got %q", phase, state.StatusCompleted, status)
		}
	}
}

// TestExecutePhasesCommon_RolloutFailureSkipsDownstream tests that when rollout fails,
// upstream phases are completed, the failing phase is marked failed, and downstream phases are skipped.
func TestExecutePhasesCommon_RolloutFailureSkipsDownstream(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, ".stagecraft", "releases.json")

	// Ensure .stagecraft directory exists
	if err := os.MkdirAll(filepath.Dir(stateFile), 0o700); err != nil {
		t.Fatalf("failed to create .stagecraft directory: %v", err)
	}

	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	ctx := context.Background()
	stateMgr := state.NewManager(stateFile)
	logger := logging.NewLogger(false)

	// Create a release
	release, err := stateMgr.CreateRelease(ctx, "staging", "v1.0.0", "commit1")
	if err != nil {
		t.Fatalf("failed to create release: %v", err)
	}

	// Create PhaseFns where Rollout fails
	fns := PhaseFns{
		Build: func(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
			return nil
		},
		Push: func(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
			return nil
		},
		MigratePre: func(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
			return nil
		},
		Rollout: func(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
			return fmt.Errorf("forced rollout failure")
		},
		MigratePost: func(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
			return nil
		},
		Finalize: func(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
			return nil
		},
	}

	// Create a minimal plan
	plan := &core.Plan{
		Operations: []core.Operation{},
	}

	// Execute phases - should fail
	err = executePhasesCommon(ctx, stateMgr, release.ID, plan, logger, fns)
	if err == nil {
		t.Fatalf("executePhasesCommon should fail when rollout fails")
	}

	if !strings.Contains(err.Error(), "phase \"rollout\" failed") {
		t.Errorf("expected error to mention rollout failure, got: %v", err)
	}

	// Verify phase statuses
	updatedRelease, err := stateMgr.GetRelease(ctx, release.ID)
	if err != nil {
		t.Fatalf("failed to get release: %v", err)
	}

	type phaseStatus struct {
		phase  state.ReleasePhase
		expect state.PhaseStatus
	}

	expected := []phaseStatus{
		{state.PhaseBuild, state.StatusCompleted},
		{state.PhasePush, state.StatusCompleted},
		{state.PhaseMigratePre, state.StatusCompleted},
		{state.PhaseRollout, state.StatusFailed},
		{state.PhaseMigratePost, state.StatusSkipped},
		{state.PhaseFinalize, state.StatusSkipped},
	}

	for _, ps := range expected {
		got, ok := updatedRelease.Phases[ps.phase]
		if !ok {
			t.Errorf("expected phase %q to be present", ps.phase)
			continue
		}
		if got != ps.expect {
			t.Errorf("expected phase %q to be %q, got %q", ps.phase, ps.expect, got)
		}
	}
}

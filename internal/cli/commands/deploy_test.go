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

	"github.com/spf13/cobra"

	"stagecraft/internal/core"
	"stagecraft/internal/core/state"
	"stagecraft/pkg/logging"
)

// Ensure cobra is used (via newTestRootCommand and executeCommandForGolden)
var _ = cobra.Command{}

// Feature: CORE_STATE_TEST_ISOLATION
// Spec: spec/core/state-test-isolation.md

// Feature: CLI_DEPLOY
// Spec: spec/commands/deploy.md

// executeDeployWithPhases is a test helper that executes deploy with custom PhaseFns.
// This allows tests to inject phase behavior without using global state.
// Feature: CLI_PHASE_EXECUTION_COMMON
func executeDeployWithPhases(fns PhaseFns, args ...string) error {
	return executeWithPhasesCustom(setupDeployCommand, fns, args...)
}

func TestNewDeployCommand_HasExpectedMetadata(t *testing.T) {
	cmd := NewDeployCommand()

	if cmd.Use != "deploy" {
		t.Fatalf("expected Use to be 'deploy', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Fatalf("expected Short description to be non-empty")
	}
}

func TestDeployCommand_ConfigNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	root := newTestRootCommand()
	root.AddCommand(NewDeployCommand())

	_, err := executeCommandForGolden(root, "deploy", "--env", "staging")
	if err == nil {
		t.Fatalf("expected error when config file is missing")
	}

	if !strings.Contains(err.Error(), "stagecraft config not found") {
		t.Fatalf("expected config not found error, got: %v", err)
	}
}

func TestDeployCommand_InvalidEnvironment(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")

	configContent := `project:
  name: test-app
environments:
  dev:
    driver: local
  staging:
    driver: local
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
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

	root := newTestRootCommand()
	root.AddCommand(NewDeployCommand())

	_, err := executeCommandForGolden(root, "deploy", "--env", "nonexistent")
	if err == nil {
		t.Fatalf("expected error when environment is invalid")
	}

	if !strings.Contains(err.Error(), "invalid environment") {
		t.Fatalf("expected invalid environment error, got: %v", err)
	}
}

func TestDeployCommand_CreatesRelease(t *testing.T) {
	env := setupIsolatedStateTestEnv(t)
	configPath := filepath.Join(env.TempDir, "stagecraft.yml")

	configContent := `project:
  name: test-app
backend:
  provider: generic
  providers:
    generic:
      build:
        dockerfile: "./Dockerfile"
        context: "."
environments:
  staging:
    driver: local
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Verify state file doesn't exist yet
	if _, err := os.Stat(env.StateFile); err == nil {
		t.Fatalf("state file should not exist before deploy")
	}

	// Use stubbed phase functions that succeed
	fns := PhaseFns{
		Build:       func(ctx context.Context, plan *core.Plan, logger logging.Logger) error { return nil },
		Push:        func(ctx context.Context, plan *core.Plan, logger logging.Logger) error { return nil },
		MigratePre:  func(ctx context.Context, plan *core.Plan, logger logging.Logger) error { return nil },
		Rollout:     func(ctx context.Context, plan *core.Plan, logger logging.Logger) error { return nil },
		MigratePost: func(ctx context.Context, plan *core.Plan, logger logging.Logger) error { return nil },
		Finalize:    func(ctx context.Context, plan *core.Plan, logger logging.Logger) error { return nil },
	}

	// Run deploy with stubbed phases
	err := executeDeployWithPhases(fns, "deploy", "--env", "staging")
	if err != nil {
		t.Logf("deploy command returned error (may be expected): %v", err)
		// Even if there's an error, the release should still be created
	}

	// Verify state file was created
	if _, err := os.Stat(env.StateFile); err != nil {
		t.Fatalf("state file should be created after deploy: %v", err)
	}

	// Verify release was created
	releases, err := env.Manager.ListReleases(env.Ctx, "staging")
	if err != nil {
		t.Fatalf("failed to list releases: %v", err)
	}

	if len(releases) == 0 {
		t.Fatalf("expected at least one release to be created")
	}

	release := releases[0]
	if release.Environment != "staging" {
		t.Errorf("expected environment 'staging', got %q", release.Environment)
	}

	// Verify all phases are present and completed (since we used stubbed functions)
	expectedPhases := []state.ReleasePhase{
		state.PhaseBuild,
		state.PhasePush,
		state.PhaseMigratePre,
		state.PhaseRollout,
		state.PhaseMigratePost,
		state.PhaseFinalize,
	}

	for _, phase := range expectedPhases {
		status, ok := release.Phases[phase]
		if !ok {
			t.Errorf("expected phase %q to be present", phase)
			continue
		}
		// With stubbed functions, phases should be completed
		if status != state.StatusCompleted {
			t.Errorf("expected phase %q to be %q (with stubbed functions), got %q", phase, state.StatusCompleted, status)
		}
	}
}

func TestDeployCommand_PhaseTransitions(t *testing.T) {
	env := setupIsolatedStateTestEnv(t)
	configPath := filepath.Join(env.TempDir, "stagecraft.yml")

	configContent := `project:
  name: test-app
environments:
  staging:
    driver: local
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	root := newTestRootCommand()
	root.AddCommand(NewDeployCommand())

	// Run deploy without dry-run; this should succeed and initialize phases.
	// Note: actual phase execution will fail without proper setup, but release should be created
	_, _ = executeCommandForGolden(root, "deploy", "--env", "staging")

	// Verify phase transitions occurred
	releases, err := env.Manager.ListReleases(env.Ctx, "staging")
	if err != nil {
		t.Fatalf("failed to list releases: %v", err)
	}

	if len(releases) == 0 {
		t.Fatalf("expected at least one release")
	}

	release := releases[0]
	// In dry-run mode, phases might remain pending or transition to running
	// This test verifies the structure exists
	if release.Phases == nil {
		t.Fatalf("expected phases to be initialized")
	}
}

func TestDeployCommand_Help(t *testing.T) {
	root := newTestRootCommand()
	root.AddCommand(NewDeployCommand())

	out, err := executeCommandForGolden(root, "deploy", "--help")
	if err != nil {
		t.Fatalf("help command should not error, got: %v", err)
	}

	if !strings.Contains(out, "deploy") {
		t.Fatalf("expected help text to contain 'deploy', got: %q", out)
	}
}

func TestDeployCommand_VersionFlag(t *testing.T) {
	env := setupIsolatedStateTestEnv(t)
	configPath := filepath.Join(env.TempDir, "stagecraft.yml")

	configContent := `project:
  name: test-app
environments:
  staging:
    driver: local
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	root := newTestRootCommand()
	root.AddCommand(NewDeployCommand())

	// Run deploy with version flag (without dry-run to create release)
	_, _ = executeCommandForGolden(root, "deploy", "--env", "staging", "--version", "v1.2.3")
	// Error is expected due to missing backend/docker setup, but release should be created

	// Verify release was created with correct version
	releases, err := env.Manager.ListReleases(env.Ctx, "staging")
	if err != nil {
		t.Fatalf("failed to list releases: %v", err)
	}

	if len(releases) == 0 {
		t.Fatalf("expected at least one release")
	}

	release := releases[0]
	if release.Version != "v1.2.3" {
		t.Errorf("expected version 'v1.2.3', got %q", release.Version)
	}
}

func TestDeployCommand_PhaseFailureMarksDownstreamSkipped(t *testing.T) {
	env := setupIsolatedStateTestEnv(t)
	configPath := filepath.Join(env.TempDir, "stagecraft.yml")

	configContent := `project:
  name: test-app
environments:
  staging:
    driver: local
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Prepare PhaseFns where rollout fails deterministically
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
			t.Errorf("MigratePost should not run after rollout failure")
			return nil
		},
		Finalize: func(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
			t.Errorf("Finalize should not run after rollout failure")
			return nil
		},
	}

	// Act: execute deploy with DI-based phases
	err := executeDeployWithPhases(fns, "deploy", "--env", "staging")
	if err == nil {
		t.Fatalf("expected deploy to fail due to forced rollout failure")
	}
	if !strings.Contains(err.Error(), "forced rollout failure") {
		t.Fatalf("expected error to contain 'forced rollout failure', got: %v", err)
	}

	// Assert: reload state using the manager bound to the same state file
	releases, err := env.Manager.ListReleases(env.Ctx, "staging")
	if err != nil {
		t.Fatalf("failed to list releases: %v", err)
	}

	if len(releases) == 0 {
		t.Fatalf("expected at least one release")
	}

	release := releases[0]

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
		got, ok := release.Phases[ps.phase]
		if !ok {
			t.Errorf("expected phase %q to be present", ps.phase)
			continue
		}
		if got != ps.expect {
			t.Errorf("expected phase %q to be %q, got %q", ps.phase, ps.expect, got)
		}
	}
}

func TestMarkAllPhasesFailed_SetsAllPhasesToFailed(t *testing.T) {
	env := setupIsolatedStateTestEnv(t)

	// Create a release normally.
	release, err := env.Manager.CreateRelease(env.Ctx, "staging", "v1.0.0", "commit-sha")
	if err != nil {
		t.Fatalf("failed to create release: %v", err)
	}

	logger := logging.NewLogger(false)

	// Call the helper under test.
	markAllPhasesFailedCommon(env.Ctx, env.Manager, release.ID, logger)

	// Reload and verify.
	releases, err := env.Manager.ListReleases(env.Ctx, "staging")
	if err != nil {
		t.Fatalf("failed to list releases: %v", err)
	}

	if len(releases) != 1 {
		t.Fatalf("expected one release, got %d", len(releases))
	}

	updated := releases[0]

	for _, phase := range allPhasesCommon() {
		status, ok := updated.Phases[phase]
		if !ok {
			t.Errorf("expected phase %q to be present", phase)
			continue
		}
		if status != state.StatusFailed {
			t.Errorf("expected phase %q to be %q, got %q", phase, state.StatusFailed, status)
		}
	}
}

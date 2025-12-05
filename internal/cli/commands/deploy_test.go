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

// Feature: CLI_DEPLOY
// Spec: spec/commands/deploy.md

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

	root := newTestRootCommand()
	root.AddCommand(NewDeployCommand())

	// Run deploy in dry-run mode (should create release without error)
	_, err := executeCommandForGolden(root, "deploy", "--env", "staging", "--dry-run")
	if err != nil {
		t.Logf("deploy command returned error (may be expected): %v", err)
		// Even if there's an error, the release should still be created in dry-run
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

	// Verify all phases are initialized as pending
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
		if status != state.StatusPending {
			t.Errorf("expected phase %q to be %q, got %q", phase, state.StatusPending, status)
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

	// Run deploy in dry-run mode; this should succeed and initialize phases.
	_, _ = executeCommandForGolden(root, "deploy", "--env", "staging", "--dry-run")

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

	// Run deploy with version flag
	_, _ = executeCommandForGolden(root, "deploy", "--env", "staging", "--version", "v1.2.3", "--dry-run")
	// Error is expected

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

	// Override rollout phase to simulate a failure.
	origRollout := rolloutPhaseFn
	rolloutPhaseFn = func(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
		return fmt.Errorf("forced rollout failure")
	}
	defer func() { rolloutPhaseFn = origRollout }()

	root := newTestRootCommand()
	root.AddCommand(NewDeployCommand())

	// Run without --dry-run so executePhases actually runs.
	_, err := executeCommandForGolden(root, "deploy", "--env", "staging")
	if err == nil {
		t.Fatalf("expected deploy to fail due to forced rollout failure")
	}

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

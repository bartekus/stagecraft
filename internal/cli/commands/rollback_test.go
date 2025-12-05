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

// Feature: CLI_ROLLBACK
// Spec: spec/commands/rollback.md

// rollbackTestEnv holds common test environment setup for rollback tests.
type rollbackTestEnv struct {
	t         *testing.T
	ctx       context.Context
	mgr       *state.Manager
	allPhases []state.ReleasePhase
	tmpDir    string
	stateFile string
}

// executeRollbackWithPhases is a test helper that executes rollback with custom PhaseFns.
// This allows tests to inject phase behavior without using global state.
func executeRollbackWithPhases(fns PhaseFns, args ...string) error {
	// Create a fresh root command for this test
	root := newTestRootCommand()

	// Create rollback command with custom PhaseFns
	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "Rollback environment to a previous release",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRollbackWithPhases(cmd, args, fns)
		},
	}
	cmd.Flags().Bool("to-previous", false, "Rollback to immediately previous release")
	cmd.Flags().String("to-release", "", "Rollback to specific release ID")
	cmd.Flags().String("to-version", "", "Rollback to most recent release with matching version")

	root.AddCommand(cmd)
	root.SetArgs(args)

	return root.Execute()
}

// newRollbackTestEnv sets up a test environment with config file, state manager, and phase list.
// It changes to the temp directory and restores the original directory on cleanup.
func newRollbackTestEnv(t *testing.T) *rollbackTestEnv {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")
	stateFile := filepath.Join(tmpDir, ".stagecraft", "releases.json")

	configContent := `project:
  name: test-app
environments:
  staging:
    driver: local
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	originalDir, _ := os.Getwd()
	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	})

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	mgr := state.NewManager(stateFile)
	ctx := context.Background()

	allPhases := []state.ReleasePhase{
		state.PhaseBuild,
		state.PhasePush,
		state.PhaseMigratePre,
		state.PhaseRollout,
		state.PhaseMigratePost,
		state.PhaseFinalize,
	}

	return &rollbackTestEnv{
		t:         t,
		ctx:       ctx,
		mgr:       mgr,
		allPhases: allPhases,
		tmpDir:    tmpDir,
		stateFile: stateFile,
	}
}

func TestNewRollbackCommand_HasExpectedMetadata(t *testing.T) {
	cmd := NewRollbackCommand()

	if cmd.Use != "rollback" {
		t.Fatalf("expected Use to be 'rollback', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Fatalf("expected Short description to be non-empty")
	}
}

func TestRollbackCommand_ToPrevious_ValidPreviousRelease(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")
	stateFile := filepath.Join(tmpDir, ".stagecraft", "releases.json")

	configContent := `project:
  name: test-app
environments:
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

	// Create test releases
	mgr := state.NewManager(stateFile)
	ctx := context.Background()

	// Create previous release (fully deployed)
	previous, err := mgr.CreateRelease(ctx, "staging", "v1.0.0", "commit1")
	if err != nil {
		t.Fatalf("failed to create previous release: %v", err)
	}

	// Mark all phases as completed for previous release
	allPhases := []state.ReleasePhase{
		state.PhaseBuild,
		state.PhasePush,
		state.PhaseMigratePre,
		state.PhaseRollout,
		state.PhaseMigratePost,
		state.PhaseFinalize,
	}
	for _, phase := range allPhases {
		if err := mgr.UpdatePhase(ctx, previous.ID, phase, state.StatusCompleted); err != nil {
			t.Fatalf("failed to update phase: %v", err)
		}
	}

	// Create current release
	current, err := mgr.CreateRelease(ctx, "staging", "v1.1.0", "commit2")
	if err != nil {
		t.Fatalf("failed to create current release: %v", err)
	}

	// Mark all phases as completed for current release
	for _, phase := range allPhases {
		if err := mgr.UpdatePhase(ctx, current.ID, phase, state.StatusCompleted); err != nil {
			t.Fatalf("failed to update phase: %v", err)
		}
	}

	root := newTestRootCommand()
	root.AddCommand(NewRollbackCommand())

	// Run rollback in dry-run mode
	_, err = executeCommandForGolden(root, "rollback", "--env", "staging", "--to-previous", "--dry-run")
	if err != nil {
		t.Fatalf("rollback should succeed in dry-run mode, got: %v", err)
	}

	// Verify no new release was created in dry-run
	releases, err := mgr.ListReleases(ctx, "staging")
	if err != nil {
		t.Fatalf("failed to list releases: %v", err)
	}
	if len(releases) != 2 {
		t.Fatalf("expected 2 releases (dry-run should not create release), got %d", len(releases))
	}
}

func TestRollbackCommand_ToPrevious_NoPreviousRelease(t *testing.T) {
	env := newRollbackTestEnv(t)

	// Create only one release (no previous)
	_, err := env.mgr.CreateRelease(env.ctx, "staging", "v1.0.0", "commit1")
	if err != nil {
		t.Fatalf("failed to create release: %v", err)
	}

	root := newTestRootCommand()
	root.AddCommand(NewRollbackCommand())

	_, err = executeCommandForGolden(root, "rollback", "--env", "staging", "--to-previous")
	if err == nil {
		t.Fatalf("expected error when no previous release exists")
	}

	if !strings.Contains(err.Error(), "no previous release") {
		t.Fatalf("expected error to contain 'no previous release', got: %v", err)
	}
}

func TestRollbackCommand_ToRelease_ValidReleaseID(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")
	stateFile := filepath.Join(tmpDir, ".stagecraft", "releases.json")

	configContent := `project:
  name: test-app
environments:
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

	// Create test releases
	mgr := state.NewManager(stateFile)
	ctx := context.Background()

	// Create target release (fully deployed)
	target, err := mgr.CreateRelease(ctx, "staging", "v1.0.0", "commit1")
	if err != nil {
		t.Fatalf("failed to create target release: %v", err)
	}

	// Mark all phases as completed
	allPhases := []state.ReleasePhase{
		state.PhaseBuild,
		state.PhasePush,
		state.PhaseMigratePre,
		state.PhaseRollout,
		state.PhaseMigratePost,
		state.PhaseFinalize,
	}
	for _, phase := range allPhases {
		if err := mgr.UpdatePhase(ctx, target.ID, phase, state.StatusCompleted); err != nil {
			t.Fatalf("failed to update phase: %v", err)
		}
	}

	// Create current release
	_, err = mgr.CreateRelease(ctx, "staging", "v1.1.0", "commit2")
	if err != nil {
		t.Fatalf("failed to create current release: %v", err)
	}

	root := newTestRootCommand()
	root.AddCommand(NewRollbackCommand())

	// Run rollback in dry-run mode
	_, err = executeCommandForGolden(root, "rollback", "--env", "staging", "--to-release", target.ID, "--dry-run")
	if err != nil {
		t.Fatalf("rollback should succeed in dry-run mode, got: %v", err)
	}
}

func TestRollbackCommand_ToRelease_InvalidReleaseID(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")

	configContent := `project:
  name: test-app
environments:
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
	root.AddCommand(NewRollbackCommand())

	_, err := executeCommandForGolden(root, "rollback", "--env", "staging", "--to-release", "invalid-release-id")
	if err == nil {
		t.Fatalf("expected error when release ID is invalid")
	}

	if !strings.Contains(err.Error(), "rollback target not found") {
		t.Fatalf("expected error to contain 'rollback target not found', got: %v", err)
	}
}

func TestRollbackCommand_ToRelease_EnvironmentMismatch(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")
	stateFile := filepath.Join(tmpDir, ".stagecraft", "releases.json")

	configContent := `project:
  name: test-app
environments:
  staging:
    driver: local
  prod:
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

	// Create test release in different environment
	mgr := state.NewManager(stateFile)
	ctx := context.Background()

	prodRelease, err := mgr.CreateRelease(ctx, "prod", "v1.0.0", "commit1")
	if err != nil {
		t.Fatalf("failed to create release: %v", err)
	}

	root := newTestRootCommand()
	root.AddCommand(NewRollbackCommand())

	_, err = executeCommandForGolden(root, "rollback", "--env", "staging", "--to-release", prodRelease.ID)
	if err == nil {
		t.Fatalf("expected error when release belongs to different environment")
	}

	if !strings.Contains(err.Error(), "belongs to environment") {
		t.Fatalf("expected error to contain 'belongs to environment', got: %v", err)
	}
}

func TestRollbackCommand_ToVersion_MatchingVersion(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")
	stateFile := filepath.Join(tmpDir, ".stagecraft", "releases.json")

	configContent := `project:
  name: test-app
environments:
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

	// Create test releases
	mgr := state.NewManager(stateFile)
	ctx := context.Background()

	// Create target release with specific version (fully deployed)
	target, err := mgr.CreateRelease(ctx, "staging", "v1.0.0", "commit1")
	if err != nil {
		t.Fatalf("failed to create target release: %v", err)
	}

	// Mark all phases as completed
	allPhases := []state.ReleasePhase{
		state.PhaseBuild,
		state.PhasePush,
		state.PhaseMigratePre,
		state.PhaseRollout,
		state.PhaseMigratePost,
		state.PhaseFinalize,
	}
	for _, phase := range allPhases {
		if err := mgr.UpdatePhase(ctx, target.ID, phase, state.StatusCompleted); err != nil {
			t.Fatalf("failed to update phase: %v", err)
		}
	}

	// Create current release with different version
	_, err = mgr.CreateRelease(ctx, "staging", "v1.1.0", "commit2")
	if err != nil {
		t.Fatalf("failed to create current release: %v", err)
	}

	root := newTestRootCommand()
	root.AddCommand(NewRollbackCommand())

	// Run rollback in dry-run mode
	_, err = executeCommandForGolden(root, "rollback", "--env", "staging", "--to-version", "v1.0.0", "--dry-run")
	if err != nil {
		t.Fatalf("rollback should succeed in dry-run mode, got: %v", err)
	}
}

func TestRollbackCommand_ToVersion_NoMatchingVersion(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")

	configContent := `project:
  name: test-app
environments:
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
	root.AddCommand(NewRollbackCommand())

	_, err := executeCommandForGolden(root, "rollback", "--env", "staging", "--to-version", "v999.0.0")
	if err == nil {
		t.Fatalf("expected error when no matching version exists")
	}

	if !strings.Contains(err.Error(), "no release found with version") {
		t.Fatalf("expected error to contain 'no release found with version', got: %v", err)
	}
}

func TestRollbackCommand_TargetValidation_CannotRollbackToCurrent(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")
	stateFile := filepath.Join(tmpDir, ".stagecraft", "releases.json")

	configContent := `project:
  name: test-app
environments:
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

	// Create test release
	mgr := state.NewManager(stateFile)
	ctx := context.Background()

	current, err := mgr.CreateRelease(ctx, "staging", "v1.0.0", "commit1")
	if err != nil {
		t.Fatalf("failed to create release: %v", err)
	}

	// Mark all phases as completed
	allPhases := []state.ReleasePhase{
		state.PhaseBuild,
		state.PhasePush,
		state.PhaseMigratePre,
		state.PhaseRollout,
		state.PhaseMigratePost,
		state.PhaseFinalize,
	}
	for _, phase := range allPhases {
		if err := mgr.UpdatePhase(ctx, current.ID, phase, state.StatusCompleted); err != nil {
			t.Fatalf("failed to update phase: %v", err)
		}
	}

	root := newTestRootCommand()
	root.AddCommand(NewRollbackCommand())

	_, err = executeCommandForGolden(root, "rollback", "--env", "staging", "--to-release", current.ID)
	if err == nil {
		t.Fatalf("expected error when trying to rollback to current release")
	}

	if !strings.Contains(err.Error(), "cannot rollback to current release") {
		t.Fatalf("expected error to contain 'cannot rollback to current release', got: %v", err)
	}
}

func TestRollbackCommand_TargetValidation_TargetMustBeFullyDeployed(t *testing.T) {
	env := setupIsolatedStateTestEnv(t)
	mgr := env.Manager
	ctx := env.Ctx

	// Create previous release (incomplete deployment)
	previous, err := mgr.CreateRelease(ctx, "staging", "v1.0.0", "commit1")
	if err != nil {
		t.Fatalf("failed to create previous release: %v", err)
	}

	// Mark only some phases as completed (not all)
	if err := mgr.UpdatePhase(ctx, previous.ID, state.PhaseBuild, state.StatusCompleted); err != nil {
		t.Fatalf("failed to update phase: %v", err)
	}
	if err := mgr.UpdatePhase(ctx, previous.ID, state.PhasePush, state.StatusCompleted); err != nil {
		t.Fatalf("failed to update phase: %v", err)
	}
	// Leave other phases as pending

	// Create current release
	_, err = mgr.CreateRelease(ctx, "staging", "v1.1.0", "commit2")
	if err != nil {
		t.Fatalf("failed to create current release: %v", err)
	}

	root := newTestRootCommand()
	root.AddCommand(NewRollbackCommand())

	_, err = executeCommandForGolden(root, "rollback", "--env", "staging", "--to-previous")
	if err == nil {
		t.Fatalf("expected error when target is not fully deployed")
	}

	if !strings.Contains(err.Error(), "not fully deployed") {
		t.Fatalf("expected error to contain 'not fully deployed', got: %v", err)
	}
}

func TestRollbackCommand_CreatesNewReleaseWithTargetVersion(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")
	stateFile := filepath.Join(tmpDir, ".stagecraft", "releases.json")

	configContent := `project:
  name: test-app
environments:
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

	// Create test releases
	mgr := state.NewManager(stateFile)
	ctx := context.Background()

	// Create previous release (fully deployed)
	previous, err := mgr.CreateRelease(ctx, "staging", "v1.0.0", "commit1")
	if err != nil {
		t.Fatalf("failed to create previous release: %v", err)
	}

	// Mark all phases as completed
	allPhases := []state.ReleasePhase{
		state.PhaseBuild,
		state.PhasePush,
		state.PhaseMigratePre,
		state.PhaseRollout,
		state.PhaseMigratePost,
		state.PhaseFinalize,
	}
	for _, phase := range allPhases {
		if err := mgr.UpdatePhase(ctx, previous.ID, phase, state.StatusCompleted); err != nil {
			t.Fatalf("failed to update phase: %v", err)
		}
	}

	// Create current release
	current, err := mgr.CreateRelease(ctx, "staging", "v1.1.0", "commit2")
	if err != nil {
		t.Fatalf("failed to create current release: %v", err)
	}

	// Mark all phases as completed for current
	for _, phase := range allPhases {
		if err := mgr.UpdatePhase(ctx, current.ID, phase, state.StatusCompleted); err != nil {
			t.Fatalf("failed to update phase: %v", err)
		}
	}

	root := newTestRootCommand()
	root.AddCommand(NewRollbackCommand())

	// Override phase execution to avoid actual deployment
	err = executeRollbackWithPhases(PhaseFns{
		Build:       func(ctx context.Context, plan *core.Plan, logger logging.Logger) error { return nil },
		Push:        defaultPhaseFns.Push,
		MigratePre:  defaultPhaseFns.MigratePre,
		Rollout:     defaultPhaseFns.Rollout,
		MigratePost: defaultPhaseFns.MigratePost,
		Finalize:    defaultPhaseFns.Finalize,
	}, "rollback", "--env", "staging", "--to-previous")
	if err != nil {
		t.Fatalf("rollback should succeed, got: %v", err)
	}

	// Verify new release was created with target's version
	releases, err := mgr.ListReleases(ctx, "staging")
	if err != nil {
		t.Fatalf("failed to list releases: %v", err)
	}

	if len(releases) < 3 {
		t.Fatalf("expected at least 3 releases (previous, current, rollback), got %d", len(releases))
	}

	// Newest release should be the rollback
	rollbackRelease := releases[0]
	if rollbackRelease.Version != previous.Version {
		t.Errorf("expected rollback release version to be %q, got %q", previous.Version, rollbackRelease.Version)
	}
	if rollbackRelease.CommitSHA != previous.CommitSHA {
		t.Errorf("expected rollback release commit SHA to be %q, got %q", previous.CommitSHA, rollbackRelease.CommitSHA)
	}
}

func TestRollbackCommand_DryRun_DoesNotCreateRelease(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")
	stateFile := filepath.Join(tmpDir, ".stagecraft", "releases.json")

	configContent := `project:
  name: test-app
environments:
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

	// Create test releases
	mgr := state.NewManager(stateFile)
	ctx := context.Background()

	// Create previous release (fully deployed)
	previous, err := mgr.CreateRelease(ctx, "staging", "v1.0.0", "commit1")
	if err != nil {
		t.Fatalf("failed to create previous release: %v", err)
	}

	// Mark all phases as completed
	allPhases := []state.ReleasePhase{
		state.PhaseBuild,
		state.PhasePush,
		state.PhaseMigratePre,
		state.PhaseRollout,
		state.PhaseMigratePost,
		state.PhaseFinalize,
	}
	for _, phase := range allPhases {
		if err := mgr.UpdatePhase(ctx, previous.ID, phase, state.StatusCompleted); err != nil {
			t.Fatalf("failed to update phase: %v", err)
		}
	}

	// Create current release
	_, err = mgr.CreateRelease(ctx, "staging", "v1.1.0", "commit2")
	if err != nil {
		t.Fatalf("failed to create current release: %v", err)
	}

	// Count releases before rollback
	releasesBefore, err := mgr.ListReleases(ctx, "staging")
	if err != nil {
		t.Fatalf("failed to list releases: %v", err)
	}
	countBefore := len(releasesBefore)

	root := newTestRootCommand()
	root.AddCommand(NewRollbackCommand())

	// Run rollback in dry-run mode
	_, err = executeCommandForGolden(root, "rollback", "--env", "staging", "--to-previous", "--dry-run")
	if err != nil {
		t.Fatalf("rollback should succeed in dry-run mode, got: %v", err)
	}

	// Verify no new release was created
	releasesAfter, err := mgr.ListReleases(ctx, "staging")
	if err != nil {
		t.Fatalf("failed to list releases: %v", err)
	}
	countAfter := len(releasesAfter)

	if countAfter != countBefore {
		t.Fatalf("dry-run should not create release: before=%d, after=%d", countBefore, countAfter)
	}
}

func TestRollbackCommand_MultipleTargetFlags_Error(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")

	configContent := `project:
  name: test-app
environments:
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
	root.AddCommand(NewRollbackCommand())

	_, err := executeCommandForGolden(root, "rollback", "--env", "staging", "--to-previous", "--to-release", "rel-123")
	if err == nil {
		t.Fatalf("expected error when multiple target flags are specified")
	}

	if !strings.Contains(err.Error(), "only one rollback target flag may be specified") {
		t.Fatalf("expected error to contain 'only one rollback target flag may be specified', got: %v", err)
	}
}

func TestRollbackCommand_NoTargetFlags_Error(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")

	configContent := `project:
  name: test-app
environments:
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
	root.AddCommand(NewRollbackCommand())

	_, err := executeCommandForGolden(root, "rollback", "--env", "staging")
	if err == nil {
		t.Fatalf("expected error when no target flag is specified")
	}

	if !strings.Contains(err.Error(), "rollback target required") {
		t.Fatalf("expected error to contain 'rollback target required', got: %v", err)
	}
}

func TestRollbackCommand_PhaseFailureHandling(t *testing.T) {
	env := setupIsolatedStateTestEnv(t)
	mgr := env.Manager
	ctx := env.Ctx

	// Create previous release (fully deployed)
	previous, err := mgr.CreateRelease(ctx, "staging", "v1.0.0", "commit1")
	if err != nil {
		t.Fatalf("failed to create previous release: %v", err)
	}

	// Mark all phases as completed
	allPhases := []state.ReleasePhase{
		state.PhaseBuild,
		state.PhasePush,
		state.PhaseMigratePre,
		state.PhaseRollout,
		state.PhaseMigratePost,
		state.PhaseFinalize,
	}
	for _, phase := range allPhases {
		if err := mgr.UpdatePhase(ctx, previous.ID, phase, state.StatusCompleted); err != nil {
			t.Fatalf("failed to update phase: %v", err)
		}
	}

	// Create current release
	_, err = mgr.CreateRelease(ctx, "staging", "v1.1.0", "commit2")
	if err != nil {
		t.Fatalf("failed to create current release: %v", err)
	}

	root := newTestRootCommand()
	root.AddCommand(NewRollbackCommand())

	// Override rollout phase to simulate a failure
	rollbackErr := executeRollbackWithPhases(PhaseFns{
		Build:      defaultPhaseFns.Build,
		Push:       defaultPhaseFns.Push,
		MigratePre: defaultPhaseFns.MigratePre,
		Rollout: func(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
			return fmt.Errorf("forced rollout failure")
		},
		MigratePost: defaultPhaseFns.MigratePost,
		Finalize:    defaultPhaseFns.Finalize,
	}, "rollback", "--env", "staging", "--to-previous")

	if rollbackErr == nil {
		t.Fatalf("expected rollback to fail due to forced rollout failure")
	}

	// Verify phase statuses
	// Create a new manager to ensure we read fresh state from disk
	// Use absolute path to match what the command used
	verifyMgr := state.NewManager(env.StateFile)
	releases, err := verifyMgr.ListReleases(ctx, "staging")
	if err != nil {
		t.Fatalf("failed to list releases: %v", err)
	}

	if len(releases) == 0 {
		t.Fatalf("expected at least one release")
	}

	// Find the rollback release (newest)
	rollbackRelease := releases[0]

	// Verify rollout phase is failed
	if rollbackRelease.Phases[state.PhaseRollout] != state.StatusFailed {
		t.Errorf("expected rollout phase to be failed, got %q (release: %s, all phases: %v)",
			rollbackRelease.Phases[state.PhaseRollout], rollbackRelease.ID, rollbackRelease.Phases)
	}

	// Verify downstream phases are skipped
	if rollbackRelease.Phases[state.PhaseMigratePost] != state.StatusSkipped {
		t.Errorf("expected migrate_post phase to be skipped, got %q", rollbackRelease.Phases[state.PhaseMigratePost])
	}
	if rollbackRelease.Phases[state.PhaseFinalize] != state.StatusSkipped {
		t.Errorf("expected finalize phase to be skipped, got %q", rollbackRelease.Phases[state.PhaseFinalize])
	}
}

func TestRollbackCommand_SuccessfulRollback_AllPhasesCompleted(t *testing.T) {
	env := setupIsolatedStateTestEnv(t)
	mgr := env.Manager
	ctx := env.Ctx

	// Create previous release (fully deployed)
	previous, err := mgr.CreateRelease(ctx, "staging", "v1.0.0", "commit1")
	if err != nil {
		t.Fatalf("failed to create previous release: %v", err)
	}

	// Mark all phases as completed
	allPhases := []state.ReleasePhase{
		state.PhaseBuild,
		state.PhasePush,
		state.PhaseMigratePre,
		state.PhaseRollout,
		state.PhaseMigratePost,
		state.PhaseFinalize,
	}
	for _, phase := range allPhases {
		if err := mgr.UpdatePhase(ctx, previous.ID, phase, state.StatusCompleted); err != nil {
			t.Fatalf("failed to update phase: %v", err)
		}
	}

	// Create current release
	_, err = mgr.CreateRelease(ctx, "staging", "v1.1.0", "commit2")
	if err != nil {
		t.Fatalf("failed to create current release: %v", err)
	}

	root := newTestRootCommand()
	root.AddCommand(NewRollbackCommand())

	// Run rollback (should succeed and complete all phases)
	_, err = executeCommandForGolden(root, "rollback", "--env", "staging", "--to-previous")
	if err != nil {
		t.Fatalf("rollback should succeed, got: %v", err)
	}

	// Verify rollback release was created with all phases completed
	// Create a new manager to ensure we read fresh state from disk
	// Use absolute path to match what the command used
	verifyMgr := state.NewManager(env.StateFile)
	releases, err := verifyMgr.ListReleases(ctx, "staging")
	if err != nil {
		t.Fatalf("failed to list releases: %v", err)
	}

	if len(releases) < 3 {
		t.Fatalf("expected at least 3 releases (previous, current, rollback), got %d", len(releases))
	}

	// Find the rollback release by matching version and commit SHA (more robust than assuming releases[0])
	var rollbackRelease *state.Release
	for _, r := range releases {
		if r.Version == previous.Version && r.CommitSHA == previous.CommitSHA && r.ID != previous.ID {
			rollbackRelease = r
			break
		}
	}

	if rollbackRelease == nil {
		t.Fatalf("rollback release not found (expected version %q, commit %q)", previous.Version, previous.CommitSHA)
	}

	// Verify all phases are completed
	for _, phase := range allPhases {
		status := rollbackRelease.Phases[phase]
		if status != state.StatusCompleted {
			t.Errorf("expected phase %q to be %q, got %q", phase, state.StatusCompleted, status)
		}
	}

	// Verify version and commit SHA match target
	if rollbackRelease.Version != previous.Version {
		t.Errorf("expected rollback release version to be %q, got %q", previous.Version, rollbackRelease.Version)
	}
	if rollbackRelease.CommitSHA != previous.CommitSHA {
		t.Errorf("expected rollback release commit SHA to be %q, got %q", previous.CommitSHA, rollbackRelease.CommitSHA)
	}
}

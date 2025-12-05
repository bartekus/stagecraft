// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package commands

import (
	"os"
	"strings"
	"testing"
	"time"

	"stagecraft/internal/core/state"
)

// Feature: CLI_RELEASES
// Spec: spec/commands/releases.md

func TestNewReleasesCommand_HasExpectedMetadata(t *testing.T) {
	cmd := NewReleasesCommand()

	if cmd.Use != "releases" {
		t.Fatalf("expected Use to be 'releases', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Fatalf("expected Short description to be non-empty")
	}

	// Check that subcommands exist
	if len(cmd.Commands()) != 2 {
		t.Fatalf("expected 2 subcommands, got %d", len(cmd.Commands()))
	}

	subcommandNames := make(map[string]bool)
	for _, subcmd := range cmd.Commands() {
		// Use field may include arguments (e.g., "show <release-id>")
		// Extract just the command name
		name := strings.Fields(subcmd.Use)[0]
		subcommandNames[name] = true
	}

	if !subcommandNames["list"] {
		t.Fatalf("expected 'list' subcommand to exist")
	}
	if !subcommandNames["show"] {
		t.Fatalf("expected 'show' subcommand to exist")
	}
}

func TestReleasesList_EmptyState(t *testing.T) {
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
	root.AddCommand(NewReleasesCommand())

	out, err := executeCommandForGolden(root, "releases", "list", "--env", "prod")
	if err != nil {
		t.Fatalf("releases list should not error on empty state, got: %v", err)
	}

	if !strings.Contains(out, "No releases found") {
		t.Fatalf("expected output to contain 'No releases found', got: %q", out)
	}
}

func TestReleasesList_FiltersByEnvironment(t *testing.T) {
	env := setupIsolatedStateTestEnv(t)

	// Create releases for different environments
	_, err := env.Manager.CreateRelease(env.Ctx, "prod", "v1.0.0", "commit1")
	if err != nil {
		t.Fatalf("failed to create release: %v", err)
	}

	_, err = env.Manager.CreateRelease(env.Ctx, "staging", "v1.0.0", "commit2")
	if err != nil {
		t.Fatalf("failed to create release: %v", err)
	}

	root := newTestRootCommand()
	root.AddCommand(NewReleasesCommand())

	// List only prod releases
	out, err := executeCommandForGolden(root, "releases", "list", "--env", "prod")
	if err != nil {
		t.Fatalf("releases list should not error, got: %v", err)
	}

	// Should contain prod release (check for version which will be in output)
	// but not staging (staging release has different commit, but same version)
	// Actually, both have v1.0.0, so let's check that we have exactly one release
	// by checking the output has the expected structure
	if !strings.Contains(out, "v1.0.0") {
		t.Fatalf("expected output to contain version 'v1.0.0', got: %q", out)
	}
	// Verify we're filtering correctly - staging should not appear
	// Since both have same version, we check that we only have one release line
	// (header + one data line = 2 lines with "v1.0.0" would mean both, but we filter)
	lines := strings.Split(out, "\n")
	releaseLines := 0
	for _, line := range lines {
		if strings.Contains(line, "v1.0.0") && !strings.Contains(line, "VERSION") {
			releaseLines++
		}
	}
	if releaseLines != 1 {
		t.Fatalf("expected exactly 1 release line, got %d. Output: %q", releaseLines, out)
	}
}

func TestReleasesList_ShowsReleasesInCorrectOrder(t *testing.T) {
	env := setupIsolatedStateTestEnv(t)

	// Create first release
	release1, err := env.Manager.CreateRelease(env.Ctx, "prod", "v1.0.0", "commit1")
	if err != nil {
		t.Fatalf("failed to create release: %v", err)
	}

	// Wait a bit to ensure different timestamp
	time.Sleep(10 * time.Millisecond)

	// Create second release (should be newer)
	release2, err := env.Manager.CreateRelease(env.Ctx, "prod", "v1.1.0", "commit2")
	if err != nil {
		t.Fatalf("failed to create release: %v", err)
	}

	root := newTestRootCommand()
	root.AddCommand(NewReleasesCommand())

	out, err := executeCommandForGolden(root, "releases", "list", "--env", "prod")
	if err != nil {
		t.Fatalf("releases list should not error, got: %v", err)
	}

	// Newer release should appear first
	idx1 := strings.Index(out, release1.ID)
	idx2 := strings.Index(out, release2.ID)

	if idx1 == -1 || idx2 == -1 {
		t.Fatalf("expected both release IDs in output, got: %q", out)
	}

	if idx2 > idx1 {
		t.Fatalf("expected newer release %q to appear before older release %q", release2.ID, release1.ID)
	}
}

func TestReleasesShow_DisplaysFullReleaseDetails(t *testing.T) {
	env := setupIsolatedStateTestEnv(t)

	release, err := env.Manager.CreateRelease(env.Ctx, "prod", "v1.0.0", "commit123")
	if err != nil {
		t.Fatalf("failed to create release: %v", err)
	}

	// Update some phases to test status display
	_ = env.Manager.UpdatePhase(env.Ctx, release.ID, state.PhaseBuild, state.StatusCompleted)
	_ = env.Manager.UpdatePhase(env.Ctx, release.ID, state.PhasePush, state.StatusCompleted)

	root := newTestRootCommand()
	root.AddCommand(NewReleasesCommand())

	out, err := executeCommandForGolden(root, "releases", "show", release.ID)
	if err != nil {
		t.Fatalf("releases show should not error, got: %v", err)
	}

	// Check that all expected fields are present
	expectedFields := []string{
		"Release ID:",
		"Environment:",
		"Version:",
		"Commit SHA:",
		"Timestamp:",
		"Phases:",
		"build:",
		"push:",
		"migrate_pre:",
		"rollout:",
		"migrate_post:",
		"finalize:",
	}

	for _, field := range expectedFields {
		if !strings.Contains(out, field) {
			t.Errorf("expected output to contain %q, got: %q", field, out)
		}
	}

	// Check that release ID is in output
	if !strings.Contains(out, release.ID) {
		t.Errorf("expected output to contain release ID %q, got: %q", release.ID, out)
	}

	// Check that environment is in output
	if !strings.Contains(out, "prod") {
		t.Errorf("expected output to contain environment 'prod', got: %q", out)
	}
}

func TestReleasesShow_InvalidReleaseID(t *testing.T) {
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
	root.AddCommand(NewReleasesCommand())

	_, err := executeCommandForGolden(root, "releases", "show", "invalid-release-id")
	if err == nil {
		t.Fatalf("expected error when showing invalid release ID")
	}

	if !strings.Contains(err.Error(), "release not found") {
		t.Fatalf("expected error to contain 'release not found', got: %v", err)
	}
}

func TestReleasesList_ShowsPhaseStatus(t *testing.T) {
	env := setupIsolatedStateTestEnv(t)

	release, err := env.Manager.CreateRelease(env.Ctx, "prod", "v1.0.0", "commit1")
	if err != nil {
		t.Fatalf("failed to create release: %v", err)
	}

	// Mark some phases as completed, one as failed
	_ = env.Manager.UpdatePhase(env.Ctx, release.ID, state.PhaseBuild, state.StatusCompleted)
	_ = env.Manager.UpdatePhase(env.Ctx, release.ID, state.PhasePush, state.StatusCompleted)
	_ = env.Manager.UpdatePhase(env.Ctx, release.ID, state.PhaseMigratePre, state.StatusFailed)

	root := newTestRootCommand()
	root.AddCommand(NewReleasesCommand())

	out, err := executeCommandForGolden(root, "releases", "list", "--env", "prod")
	if err != nil {
		t.Fatalf("releases list should not error, got: %v", err)
	}

	// Should show overall status (failed in this case)
	if !strings.Contains(out, "failed") {
		t.Fatalf("expected output to contain 'failed' status, got: %q", out)
	}
}

func TestReleasesShow_ShowsPhaseStatuses(t *testing.T) {
	env := setupIsolatedStateTestEnv(t)

	release, err := env.Manager.CreateRelease(env.Ctx, "prod", "v1.0.0", "commit1")
	if err != nil {
		t.Fatalf("failed to create release: %v", err)
	}

	// Update phases to different statuses
	_ = env.Manager.UpdatePhase(env.Ctx, release.ID, state.PhaseBuild, state.StatusCompleted)
	_ = env.Manager.UpdatePhase(env.Ctx, release.ID, state.PhasePush, state.StatusRunning)
	_ = env.Manager.UpdatePhase(env.Ctx, release.ID, state.PhaseMigratePre, state.StatusPending)

	root := newTestRootCommand()
	root.AddCommand(NewReleasesCommand())

	out, err := executeCommandForGolden(root, "releases", "show", release.ID)
	if err != nil {
		t.Fatalf("releases show should not error, got: %v", err)
	}

	// Check that phase statuses are displayed
	if !strings.Contains(out, "completed") {
		t.Errorf("expected output to contain 'completed' status, got: %q", out)
	}
	if !strings.Contains(out, "running") {
		t.Errorf("expected output to contain 'running' status, got: %q", out)
	}
	if !strings.Contains(out, "pending") {
		t.Errorf("expected output to contain 'pending' status, got: %q", out)
	}
}

func TestReleasesCommand_Help(t *testing.T) {
	root := newTestRootCommand()
	root.AddCommand(NewReleasesCommand())

	// Test main command help
	out, err := executeCommandForGolden(root, "releases", "--help")
	if err != nil {
		t.Fatalf("help command should not error, got: %v", err)
	}

	if !strings.Contains(out, "releases") {
		t.Fatalf("expected help text to contain 'releases', got: %q", out)
	}

	// Test list subcommand help
	out, err = executeCommandForGolden(root, "releases", "list", "--help")
	if err != nil {
		t.Fatalf("help command should not error, got: %v", err)
	}

	if !strings.Contains(out, "list") {
		t.Fatalf("expected help text to contain 'list', got: %q", out)
	}

	// Test show subcommand help
	out, err = executeCommandForGolden(root, "releases", "show", "--help")
	if err != nil {
		t.Fatalf("help command should not error, got: %v", err)
	}

	if !strings.Contains(out, "show") {
		t.Fatalf("expected help text to contain 'show', got: %q", out)
	}
}

func TestReleasesList_AllEnvironmentsWhenEnvNotSet(t *testing.T) {
	env := setupIsolatedStateTestEnv(t)

	// Create releases for different environments
	_, err := env.Manager.CreateRelease(env.Ctx, "prod", "v1.0.0", "commit1")
	if err != nil {
		t.Fatalf("failed to create release: %v", err)
	}

	_, err = env.Manager.CreateRelease(env.Ctx, "staging", "v1.0.0", "commit2")
	if err != nil {
		t.Fatalf("failed to create release: %v", err)
	}

	_, err = env.Manager.CreateRelease(env.Ctx, "dev", "v1.0.0", "commit3")
	if err != nil {
		t.Fatalf("failed to create release: %v", err)
	}

	root := newTestRootCommand()
	root.AddCommand(NewReleasesCommand())

	// List without --env flag (should show all environments)
	out, err := executeCommandForGolden(root, "releases", "list")
	if err != nil {
		t.Fatalf("releases list should not error, got: %v", err)
	}

	// Should contain ENVIRONMENT column
	if !strings.Contains(out, "ENVIRONMENT") {
		t.Fatalf("expected output to contain 'ENVIRONMENT' column, got: %q", out)
	}

	// Should contain all three environments
	if !strings.Contains(out, "prod") {
		t.Fatalf("expected output to contain 'prod', got: %q", out)
	}
	if !strings.Contains(out, "staging") {
		t.Fatalf("expected output to contain 'staging', got: %q", out)
	}
	if !strings.Contains(out, "dev") {
		t.Fatalf("expected output to contain 'dev', got: %q", out)
	}

	// Verify releases are grouped by environment (alphabetically: dev, prod, staging)
	// and sorted newest first within each environment
	lines := strings.Split(out, "\n")
	var releaseLines []string
	for _, line := range lines {
		if strings.Contains(line, "rel-") && !strings.Contains(line, "RELEASE ID") {
			releaseLines = append(releaseLines, line)
		}
	}

	if len(releaseLines) != 3 {
		t.Fatalf("expected 3 release lines, got %d. Output: %q", len(releaseLines), out)
	}

	// Verify sort order: should be grouped by environment alphabetically
	// Extract environments from output lines
	envs := make([]string, 0, len(releaseLines))
	for _, line := range releaseLines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			// Format: RELEASE_ID ENVIRONMENT VERSION TIMESTAMP STATUS
			envs = append(envs, fields[1])
		}
	}

	// Verify environments are in alphabetical order (dev, prod, staging)
	expectedOrder := []string{"dev", "prod", "staging"}
	if len(envs) != len(expectedOrder) {
		t.Fatalf("expected %d environments, got %d: %v", len(expectedOrder), len(envs), envs)
	}
	for i, env := range envs {
		if env != expectedOrder[i] {
			t.Errorf("expected environment at position %d to be %q, got %q. Order: %v", i, expectedOrder[i], env, envs)
		}
	}
}

func TestReleasesList_OverallStatus_Pending(t *testing.T) {
	env := setupIsolatedStateTestEnv(t)

	// Create a release without updating any phases (all should be pending)
	release, err := env.Manager.CreateRelease(env.Ctx, "prod", "v1.0.0", "commit1")
	if err != nil {
		t.Fatalf("failed to create release: %v", err)
	}

	root := newTestRootCommand()
	root.AddCommand(NewReleasesCommand())

	out, err := executeCommandForGolden(root, "releases", "list", "--env", "prod")
	if err != nil {
		t.Fatalf("releases list should not error, got: %v", err)
	}

	// Should show pending status
	if !strings.Contains(out, "pending") {
		t.Fatalf("expected output to contain 'pending' status, got: %q", out)
	}

	// Verify the release ID is in the output
	if !strings.Contains(out, release.ID) {
		t.Fatalf("expected output to contain release ID %q, got: %q", release.ID, out)
	}
}

func TestReleasesList_OverallStatus_PartialCompletion(t *testing.T) {
	env := setupIsolatedStateTestEnv(t)

	release, err := env.Manager.CreateRelease(env.Ctx, "prod", "v1.0.0", "commit1")
	if err != nil {
		t.Fatalf("failed to create release: %v", err)
	}

	// Mark only build as completed
	_ = env.Manager.UpdatePhase(env.Ctx, release.ID, state.PhaseBuild, state.StatusCompleted)

	root := newTestRootCommand()
	root.AddCommand(NewReleasesCommand())

	out, err := executeCommandForGolden(root, "releases", "list", "--env", "prod")
	if err != nil {
		t.Fatalf("releases list should not error, got: %v", err)
	}

	// Should show pending status (not all phases completed)
	if !strings.Contains(out, "pending") {
		t.Fatalf("expected output to contain 'pending' status for partially completed release, got: %q", out)
	}
	if strings.Contains(out, "completed") {
		t.Fatalf("expected output to not contain 'completed' status for partially completed release, got: %q", out)
	}
}

func TestReleasesList_OverallStatus_Running(t *testing.T) {
	env := setupIsolatedStateTestEnv(t)

	release, err := env.Manager.CreateRelease(env.Ctx, "prod", "v1.0.0", "commit1")
	if err != nil {
		t.Fatalf("failed to create release: %v", err)
	}

	// Mark build as completed, push as running
	_ = env.Manager.UpdatePhase(env.Ctx, release.ID, state.PhaseBuild, state.StatusCompleted)
	_ = env.Manager.UpdatePhase(env.Ctx, release.ID, state.PhasePush, state.StatusRunning)

	root := newTestRootCommand()
	root.AddCommand(NewReleasesCommand())

	out, err := executeCommandForGolden(root, "releases", "list", "--env", "prod")
	if err != nil {
		t.Fatalf("releases list should not error, got: %v", err)
	}

	// Should show running status
	if !strings.Contains(out, "running") {
		t.Fatalf("expected output to contain 'running' status, got: %q", out)
	}
}

func TestReleasesList_OverallStatus_AllCompleted(t *testing.T) {
	env := setupIsolatedStateTestEnv(t)

	release, err := env.Manager.CreateRelease(env.Ctx, "prod", "v1.0.0", "commit1")
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
		_ = env.Manager.UpdatePhase(env.Ctx, release.ID, phase, state.StatusCompleted)
	}

	root := newTestRootCommand()
	root.AddCommand(NewReleasesCommand())

	out, err := executeCommandForGolden(root, "releases", "list", "--env", "prod")
	if err != nil {
		t.Fatalf("releases list should not error, got: %v", err)
	}

	// Should show completed status
	if !strings.Contains(out, "completed") {
		t.Fatalf("expected output to contain 'completed' status, got: %q", out)
	}
}

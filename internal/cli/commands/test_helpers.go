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
	"os"
	"path/filepath"
	"testing"

	"stagecraft/internal/core/state"
)

// Feature: CORE_STATE_TEST_ISOLATION
// Spec: spec/core/state-test-isolation.md

// isolatedStateTestEnv provides an isolated test environment with its own state file,
// working directory, and environment variables. All cleanup is handled automatically
// via t.Cleanup.
type isolatedStateTestEnv struct {
	Ctx       context.Context
	StateFile string
	Manager   *state.Manager
	TempDir   string
}

// setupIsolatedStateTestEnv creates an isolated test environment for state-touching tests.
// It:
// - Creates a temporary directory for the test
// - Sets up an isolated state file at .stagecraft/releases.json in the temp directory
// - Sets STAGECRAFT_STATE_FILE environment variable (test-scoped, auto-cleanup)
// - Changes working directory to the temp directory (with cleanup)
// - Returns a manager configured to use the isolated state file
//
// This ensures each test has its own isolated state file and prevents suite-level cross-talk.
func setupIsolatedStateTestEnv(t *testing.T) *isolatedStateTestEnv {
	t.Helper()

	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, ".stagecraft", "releases.json")

	// Get absolute path for state file
	absStateFile, err := filepath.Abs(stateFile)
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}

	// Set env var (test-scoped, auto-cleanup)
	t.Setenv("STAGECRAFT_STATE_FILE", absStateFile)

	// Change directory (with cleanup)
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
		// Note: t.Setenv cleanup is handled automatically by testing package
	})

	// Create manager with absolute path
	mgr := state.NewManager(absStateFile)

	return &isolatedStateTestEnv{
		Ctx:       context.Background(),
		StateFile: absStateFile,
		Manager:   mgr,
		TempDir:   tmpDir,
	}
}

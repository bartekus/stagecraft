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

// isolatedStateTestEnv provides an isolated test environment with its own state file,
// working directory, and environment variables. This ensures tests don't interfere
// with each other when running in parallel or sequentially.
type isolatedStateTestEnv struct {
	// Ctx is a background context for the test
	Ctx context.Context

	// TmpDir is the temporary directory created for this test
	TmpDir string

	// StateFile is the absolute path to the isolated state file
	StateFile string

	// Manager is a state manager configured to use the isolated state file
	Manager *state.Manager

	// ConfigPath is the path to the stagecraft.yml config file in the temp directory
	ConfigPath string
}

// setupIsolatedStateTestEnv creates an isolated test environment with:
// - A unique temporary directory
// - An isolated state file (via STAGECRAFT_STATE_FILE env var)
// - Working directory changed to the temp directory (restored on cleanup)
// - A basic stagecraft.yml config file
//
// This helper ensures complete test isolation by:
// 1. Using absolute paths for state files (prevents CWD-related issues)
// 2. Setting STAGECRAFT_STATE_FILE env var (ensures CLI uses the test's state file)
// 3. Isolating working directory (prevents path resolution conflicts)
// 4. Cleaning up env vars and CWD on test completion
func setupIsolatedStateTestEnv(t *testing.T) *isolatedStateTestEnv {
	t.Helper()

	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, ".stagecraft", "releases.json")
	configPath := filepath.Join(tmpDir, "stagecraft.yml")

	// Ensure .stagecraft directory exists
	if err := os.MkdirAll(filepath.Dir(stateFile), 0o700); err != nil {
		t.Fatalf("failed to create .stagecraft directory: %v", err)
	}

	// Get absolute path for state file to avoid CWD-related issues
	absStateFile, err := filepath.Abs(stateFile)
	if err != nil {
		t.Fatalf("failed to get absolute path for state file: %v", err)
	}

	// Set environment variable to ensure state.NewDefaultManager() uses our test state file
	// This prevents test isolation issues when running with the full test suite
	t.Setenv("STAGECRAFT_STATE_FILE", absStateFile)

	// Explicitly unset at end of test to prevent interference with other tests
	t.Cleanup(func() {
		_ = os.Unsetenv("STAGECRAFT_STATE_FILE")
	})

	// Save original working directory and change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Restore original directory on cleanup
	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	})

	// Create a basic stagecraft.yml config file
	configContent := `project:
  name: test-app
environments:
  staging:
    driver: local
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Create state manager using absolute path
	mgr := state.NewManager(absStateFile)

	return &isolatedStateTestEnv{
		Ctx:        context.Background(),
		TmpDir:     tmpDir,
		StateFile:  absStateFile,
		Manager:    mgr,
		ConfigPath: configPath,
	}
}

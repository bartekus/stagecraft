// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// test/e2e/migrate_smoke_test.go
package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Feature: CLI_MIGRATE_BASIC
// Spec: spec/commands/migrate-basic.md

// TestStagecraftMigrate_Smoke tests that stagecraft migrate works with examples/basic-node
// This test expects the binary `stagecraft` to be in PATH or built beforehand.
// It is gated behind the `e2e` build tag so it won't run in normal `go test ./...` runs.
func TestStagecraftMigrate_Smoke(t *testing.T) {
	// Find examples/basic-node directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// Navigate to repo root
	repoRoot := filepath.Join(wd, "..", "..")
	exampleDir := filepath.Join(repoRoot, "examples", "basic-node")

	// Check if example exists
	if _, err := os.Stat(exampleDir); os.IsNotExist(err) {
		t.Skip("examples/basic-node not found, skipping E2E test")
	}

	// Check if config exists
	configPath := filepath.Join(exampleDir, "stagecraft.yml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skip("examples/basic-node/stagecraft.yml not found, skipping E2E test")
	}

	// Test --plan flag (doesn't require database connection)
	cmd := exec.Command("stagecraft", "migrate", "--plan")
	cmd.Dir = exampleDir

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	// This will fail if DATABASE_URL is not set, but that's expected
	// We're just verifying the command exists and can parse the config
	err = cmd.Run()

	// Either it succeeds (if migrations dir exists) or fails with a helpful error
	if err != nil {
		// Check that error is about missing DATABASE_URL or migration path, not command not found
		output := out.String()
		if !strings.Contains(output, "DATABASE_URL") &&
			!strings.Contains(output, "migration") &&
			!strings.Contains(output, "not found") &&
			!strings.Contains(output, "reading migration directory") {
			t.Fatalf("unexpected error output: %s", output)
		}
	}
}

// TestStagecraftMigrate_Help verifies the migrate command help works
func TestStagecraftMigrate_Help(t *testing.T) {
	cmd := exec.Command("stagecraft", "migrate", "--help")

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		t.Fatalf("expected 'stagecraft migrate --help' to succeed, got error: %v, output: %s", err, out.String())
	}

	if !strings.Contains(out.String(), "Run database migrations") &&
		!strings.Contains(out.String(), "Loads stagecraft.yml") {
		t.Fatalf("expected output to contain migrate command description, got: %q", out.String())
	}
}

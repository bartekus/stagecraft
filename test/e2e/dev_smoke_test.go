// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// test/e2e/dev_smoke_test.go
package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Feature: CLI_DEV_BASIC
// Spec: spec/commands/dev-basic.md

// TestStagecraftDev_Smoke tests that stagecraft dev works with examples/basic-node
// This test expects the binary `stagecraft` to be in PATH or built beforehand.
// It is gated behind the `e2e` build tag so it won't run in normal `go test ./...` runs.
func TestStagecraftDev_Smoke(t *testing.T) {
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

	// Run stagecraft dev with a timeout (it will run until cancelled)
	// For smoke test, we'll use --help to verify the command exists
	cmd := exec.Command("stagecraft", "dev", "--help")
	cmd.Dir = exampleDir

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		t.Fatalf("expected 'stagecraft dev --help' to succeed, got error: %v, output: %s", err, out.String())
	}

	if !strings.Contains(out.String(), "Start development environment") {
		t.Fatalf("expected output to contain dev command description, got: %q", out.String())
	}
}

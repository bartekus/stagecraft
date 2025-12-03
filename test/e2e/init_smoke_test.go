// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build e2e

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Feature: CLI_INIT
// Spec: spec/commands/init.md

// This is a stub E2E test that expects the binary `stagecraft` to be
// in PATH or built beforehand. It is gated behind the `e2e` build tag
// so it won't run in normal `go test ./...` runs.
func TestStagecraftInit_Smoke(t *testing.T) {
	tmpDir := t.TempDir()

	// Run init in an isolated directory so stagecraft.yml does not already exist.
	cmd := exec.Command("stagecraft", "init", "--non-interactive", "--project-name", "e2e")
	cmd.Dir = tmpDir

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		t.Fatalf("expected 'stagecraft init' to succeed, got error: %v, output: %s", err, out.String())
	}

	output := out.String()
	if !strings.Contains(output, "Created Stagecraft config") {
		t.Fatalf("expected output to contain init success message, got: %q", output)
	}

	// Ensure the config file was actually created.
	configPath := filepath.Join(tmpDir, "stagecraft.yml")
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("expected stagecraft.yml to be created at %s, got error: %v", configPath, err)
	}
}

// go test ./test/e2e -tags=e2e

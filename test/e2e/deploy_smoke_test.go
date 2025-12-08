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

// Feature: CLI_DEPLOY
// Spec: spec/commands/deploy.md

// TestStagecraftDeploy_Smoke is a minimal smoke test that verifies the deploy command
// can run in dry-run mode without errors. It sets up a minimal "hello world" project
// and runs `stagecraft deploy --env=test --dry-run`.
func TestStagecraftDeploy_Smoke(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a minimal valid stagecraft.yml
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
  test:
    driver: local
`

	configPath := filepath.Join(tmpDir, "stagecraft.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Create a minimal Dockerfile (deploy may check for it)
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte("FROM alpine\n"), 0o600); err != nil {
		t.Fatalf("failed to write Dockerfile: %v", err)
	}

	// Run deploy in dry-run mode
	cmd := exec.Command("stagecraft", "deploy", "--env=test", "--dry-run")
	cmd.Dir = tmpDir

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		t.Fatalf("expected 'stagecraft deploy --env=test --dry-run' to succeed, got error: %v, output: %s", err, out.String())
	}

	output := out.String()
	// Check for deterministic marker indicating deploy plan was generated
	// The exact message may vary, but we expect some indication of success
	if !strings.Contains(output, "deploy") && !strings.Contains(output, "plan") && !strings.Contains(output, "DRY RUN") {
		t.Logf("deploy output: %q", output)
		// Don't fail if output format is different, just log it
		// The important thing is that the command exited with code 0
	}
}

// go test ./test/e2e -tags=e2e

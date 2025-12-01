// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - A Go-based CLI for orchestrating local-first multi-service deployments using Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Feature: CLI_DEV_BASIC
// Spec: spec/commands/dev-basic.md

func TestNewDevCommand_HasExpectedMetadata(t *testing.T) {
	cmd := NewDevCommand()

	if cmd.Use != "dev" {
		t.Fatalf("expected Use to be 'dev', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Fatalf("expected Short description to be non-empty")
	}
}

func TestDevCommand_ConfigNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	root := &cobra.Command{Use: "stagecraft"}
	root.AddCommand(NewDevCommand())

	_, err := executeCommandForGolden(root, "dev")
	if err == nil {
		t.Fatalf("expected error when config file is missing")
	}

	if !strings.Contains(err.Error(), "stagecraft config not found") {
		t.Fatalf("expected config not found error, got: %v", err)
	}
}

func TestDevCommand_NoBackendConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")

	// Write config without backend section
	configContent := `project:
  name: test-app
environments:
  dev:
    driver: local
`
	os.WriteFile(configPath, []byte(configContent), 0644)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	root := &cobra.Command{Use: "stagecraft"}
	root.AddCommand(NewDevCommand())

	_, err := executeCommandForGolden(root, "dev")
	if err == nil {
		t.Fatalf("expected error when backend config is missing")
	}

	if !strings.Contains(err.Error(), "no backend configuration") {
		t.Fatalf("expected no backend config error, got: %v", err)
	}
}

func TestDevCommand_UnknownProvider(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")

	configContent := `project:
  name: test-app
backend:
  provider: unknown-provider
  providers:
    unknown-provider:
      dev:
        command: ["echo", "test"]
environments:
  dev:
    driver: local
`
	os.WriteFile(configPath, []byte(configContent), 0644)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	root := &cobra.Command{Use: "stagecraft"}
	root.AddCommand(NewDevCommand())

	_, err := executeCommandForGolden(root, "dev")
	if err == nil {
		t.Fatalf("expected error for unknown provider")
	}

	if !strings.Contains(err.Error(), "unknown backend provider") {
		t.Fatalf("expected unknown provider error, got: %v", err)
	}
}

func TestDevCommand_Help(t *testing.T) {
	root := &cobra.Command{Use: "stagecraft"}
	root.AddCommand(NewDevCommand())

	out, err := executeCommandForGolden(root, "dev", "--help")
	if err != nil {
		t.Fatalf("help command should not error, got: %v", err)
	}

	if !strings.Contains(out, "Loads stagecraft.yml") && !strings.Contains(out, "dev") {
		t.Fatalf("expected help text, got: %q", out)
	}
}


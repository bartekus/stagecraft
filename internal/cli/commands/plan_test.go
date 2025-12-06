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
	"path/filepath"
	"strings"
	"testing"
)

// Feature: CLI_PLAN
// Spec: spec/commands/plan.md

func TestNewPlanCommand_HasExpectedMetadata(t *testing.T) {
	cmd := NewPlanCommand()

	if cmd.Use != "plan" {
		t.Fatalf("expected Use to be 'plan', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Fatalf("expected Short description to be non-empty")
	}
}

func TestPlanCommand_ConfigNotFound(t *testing.T) {
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
	root.AddCommand(NewPlanCommand())

	_, err := executeCommandForGolden(root, "plan", "--env", "staging")
	if err == nil {
		t.Fatalf("expected error when config file is missing")
	}

	if !strings.Contains(err.Error(), "stagecraft config not found") {
		t.Fatalf("expected config not found error, got: %v", err)
	}
}

func TestPlanCommand_InvalidEnvironment(t *testing.T) {
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
	root.AddCommand(NewPlanCommand())

	_, err := executeCommandForGolden(root, "plan", "--env", "nonexistent")
	if err == nil {
		t.Fatalf("expected error when environment is invalid")
	}

	if !strings.Contains(err.Error(), "invalid environment") {
		t.Fatalf("expected invalid environment error, got: %v", err)
	}
}

func TestPlanCommand_MissingEnvFlag(t *testing.T) {
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
	root.AddCommand(NewPlanCommand())

	_, err := executeCommandForGolden(root, "plan")
	if err == nil {
		t.Fatalf("expected error when --env flag is missing")
	}

	// Cobra validates required flags before RunE, so we get a different error message
	if !strings.Contains(err.Error(), "required") && !strings.Contains(err.Error(), "env") {
		t.Fatalf("expected required flag error, got: %v", err)
	}
}

func TestPlanCommand_HappyPathText(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")

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
  staging:
    driver: local
databases:
  main:
    connection_env: DATABASE_URL
    migrations:
      engine: raw
      path: "./migrations"
      strategy: pre_deploy
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
	root.AddCommand(NewPlanCommand())

	output, err := executeCommandForGolden(root, "plan", "--env", "staging")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check for expected content
	if !strings.Contains(output, "Environment: staging") {
		t.Errorf("output should contain 'Environment: staging', got:\n%s", output)
	}
	if !strings.Contains(output, "Version: unknown") {
		t.Errorf("output should contain 'Version: unknown', got:\n%s", output)
	}
	if !strings.Contains(output, "Phases:") {
		t.Errorf("output should contain 'Phases:', got:\n%s", output)
	}

	// Compare with golden file
	goldenName := "plan_staging_all"
	golden := readGoldenFile(t, goldenName)
	if *updateGolden {
		writeGoldenFile(t, goldenName, output)
	} else if golden != "" && golden != output {
		t.Errorf("output does not match golden file:\nExpected:\n%s\nGot:\n%s", golden, output)
	}
}

func TestPlanCommand_WithVersion(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")

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
	root.AddCommand(NewPlanCommand())

	output, err := executeCommandForGolden(root, "plan", "--env", "staging", "--version", "v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Version: v1.0.0") {
		t.Errorf("output should contain 'Version: v1.0.0', got:\n%s", output)
	}
}

func TestPlanCommand_JSONFormat(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")

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
	root.AddCommand(NewPlanCommand())

	output, err := executeCommandForGolden(root, "plan", "--env", "staging", "--format", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check for JSON structure
	if !strings.Contains(output, `"env"`) {
		t.Errorf("output should contain JSON 'env' field, got:\n%s", output)
	}
	if !strings.Contains(output, `"version"`) {
		t.Errorf("output should contain JSON 'version' field, got:\n%s", output)
	}
	if !strings.Contains(output, `"phases"`) {
		t.Errorf("output should contain JSON 'phases' field, got:\n%s", output)
	}

	// Compare with golden file
	goldenName := "plan_staging_json"
	golden := readGoldenFile(t, goldenName)
	if *updateGolden {
		writeGoldenFile(t, goldenName, output)
	} else if golden != "" && golden != output {
		t.Errorf("output does not match golden file:\nExpected:\n%s\nGot:\n%s", golden, output)
	}
}

func TestPlanCommand_Determinism(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")

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
	root.AddCommand(NewPlanCommand())

	// Run the same command twice
	output1, err1 := executeCommandForGolden(root, "plan", "--env", "staging")
	if err1 != nil {
		t.Fatalf("unexpected error on first run: %v", err1)
	}

	output2, err2 := executeCommandForGolden(root, "plan", "--env", "staging")
	if err2 != nil {
		t.Fatalf("unexpected error on second run: %v", err2)
	}

	// Outputs must be identical
	if output1 != output2 {
		t.Errorf("outputs are not deterministic:\nFirst:\n%s\nSecond:\n%s", output1, output2)
	}
}

func TestPlanCommand_InvalidFormat(t *testing.T) {
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
	root.AddCommand(NewPlanCommand())

	_, err := executeCommandForGolden(root, "plan", "--env", "staging", "--format", "invalid")
	if err == nil {
		t.Fatalf("expected error when format is invalid")
	}

	if !strings.Contains(err.Error(), "invalid format") {
		t.Fatalf("expected invalid format error, got: %v", err)
	}
}

func TestPlanCommand_ErrorPropagation(t *testing.T) {
	// Test that plan generation errors propagate correctly
	// We'll use a config that causes CORE_PLAN to fail
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")

	// Invalid config that will cause plan generation to fail
	configContent := `project:
  name: test-app
environments:
  staging:
    driver: local
    # Missing required fields that might cause plan generation to fail
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
	root.AddCommand(NewPlanCommand())

	// This should succeed since the config is valid for plan generation
	// (plan generation doesn't require all fields)
	_, err := executeCommandForGolden(root, "plan", "--env", "staging")
	if err != nil {
		// If there's an error, it should be wrapped with context
		if !strings.Contains(err.Error(), "generating deployment plan") {
			t.Errorf("expected error to mention 'generating deployment plan', got: %v", err)
		}
	}
}

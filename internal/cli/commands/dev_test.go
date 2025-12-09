// SPDX-License-Identifier: AGPL-3.0-or-later

package commands

import (
	"os"
	"path/filepath"
	"testing"
)

// Feature: CLI_DEV
// Spec: spec/commands/dev.md

func TestNewDevCommand_HasExpectedFlags(t *testing.T) {
	t.Helper()

	cmd := NewDevCommand()

	flags := cmd.Flags()

	tests := []struct {
		name         string
		flagName     string
		expectedType string
	}{
		{name: "env string flag", flagName: devFlagEnv, expectedType: "string"},
		{name: "config string flag", flagName: devFlagConfig, expectedType: "string"},
		{name: "no-https bool flag", flagName: devFlagNoHTTPS, expectedType: "bool"},
		{name: "no-hosts bool flag", flagName: devFlagNoHosts, expectedType: "bool"},
		{name: "no-traefik bool flag", flagName: devFlagNoTraefik, expectedType: "bool"},
		{name: "detach bool flag", flagName: devFlagDetach, expectedType: "bool"},
		{name: "verbose bool flag", flagName: devFlagVerbose, expectedType: "bool"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			f := flags.Lookup(tt.flagName)
			if f == nil {
				t.Fatalf("expected flag %q to be defined", tt.flagName)
			}
			if f.Value.Type() != tt.expectedType {
				t.Errorf("flag %q type = %q, want %q", tt.flagName, f.Value.Type(), tt.expectedType)
			}
		})
	}
}

func TestNewDevCommand_DefaultsAndRun(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")

	// Write a minimal valid config with the dev environment
	configContent := `project:
  name: test-app
environments:
  dev:
    driver: local
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cmd := NewDevCommand()

	// Set config path explicitly; defaults should work otherwise
	cmd.SetArgs([]string{"--" + devFlagConfig, configPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}
}

func TestNewDevCommand_EmptyEnvFails(t *testing.T) {
	t.Helper()

	cmd := NewDevCommand()

	// Explicitly set --env to the empty string, which should be rejected.
	cmd.SetArgs([]string{"--" + devFlagEnv, ""})

	if err := cmd.Execute(); err == nil {
		t.Fatalf("Execute() error = nil, want non-nil error for empty env")
	}
}

func TestRunDevWithOptions_EmptyEnvFails(t *testing.T) {
	t.Helper()

	opts := devOptions{
		Env: "",
	}

	if err := runDevWithOptions(opts); err == nil {
		t.Fatalf("runDevWithOptions() error = nil, want non-nil for empty env")
	}
}

func TestRunDevWithOptions_BuildsTopology(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")

	// Write a minimal valid config with the dev environment
	configContent := `project:
  name: test-app
environments:
  dev:
    driver: local
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	opts := devOptions{
		Env:    "dev",
		Config: configPath,
	}

	if err := runDevWithOptions(opts); err != nil {
		t.Fatalf("runDevWithOptions() error = %v, want nil", err)
	}
}

func TestRunDevWithOptions_InvalidEnvFails(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")

	// Write a config with only "prod" environment
	configContent := `project:
  name: test-app
environments:
  prod:
    driver: local
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	opts := devOptions{
		Env:    "dev",
		Config: configPath,
	}

	if err := runDevWithOptions(opts); err == nil {
		t.Fatalf("runDevWithOptions() error = nil, want non-nil for invalid env")
	}
}

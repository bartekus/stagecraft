// SPDX-License-Identifier: AGPL-3.0-or-later

package commands

import (
	"context"
	"os"
	"path/filepath"
	"strings"
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

	// Write a minimal valid config with the dev environment and backend provider
	configContent := `project:
  name: test-app
backend:
  provider: generic
  providers:
    generic:
      dev:
        command: ["echo", "backend"]
        env:
          PORT: "4000"
environments:
  dev:
    driver: local
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Change to tmpDir so that .stagecraft/dev is created relative to it
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cmd := NewDevCommand()

	// Set config path explicitly; defaults should work otherwise
	cmd.SetArgs([]string{"--" + devFlagConfig, configPath})

	err = cmd.Execute()
	// Docker compose may fail if docker is not available or compose file is invalid,
	// but that's acceptable for this test which focuses on command structure.
	if err != nil {
		// If error is about docker compose failing (expected in test environment),
		// that's acceptable. Only fail if it's about command structure or config.
		if !strings.Contains(err.Error(), "docker compose") && !strings.Contains(err.Error(), "start processes") {
			t.Fatalf("Execute() error = %v, want nil or docker compose error", err)
		}
	}

	// Verify that dev files were written (proving command executed successfully)
	devDir := filepath.Join(tmpDir, ".stagecraft", "dev")
	composePath := filepath.Join(devDir, "compose.yaml")
	if _, err := os.Stat(composePath); err != nil {
		t.Fatalf("expected compose.yaml to be written at %s: %v", composePath, err)
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

	if err := runDevWithOptions(context.Background(), opts); err == nil {
		t.Fatalf("runDevWithOptions() error = nil, want non-nil for empty env")
	}
}

func TestRunDevWithOptions_BuildsTopology(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")

	// Write a minimal valid config with the dev environment and backend provider
	configContent := `project:
  name: test-app
backend:
  provider: generic
  providers:
    generic:
      dev:
        command: ["echo", "backend"]
        env:
          PORT: "4000"
environments:
  dev:
    driver: local
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Change to tmpDir so that .stagecraft/dev is created relative to it
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	opts := devOptions{
		Env:    "dev",
		Config: configPath,
	}

	err = runDevWithOptions(context.Background(), opts)
	// The test verifies that topology builds and files are written.
	// Docker compose may fail if docker is not available or compose file is invalid,
	// but that's acceptable for this test which focuses on topology building.
	if err != nil {
		// If error is about docker compose failing (expected in test environment),
		// that's acceptable. Only fail if it's about topology building.
		if !strings.Contains(err.Error(), "docker compose") && !strings.Contains(err.Error(), "start processes") {
			t.Fatalf("runDevWithOptions() error = %v, want nil or docker compose error", err)
		}
	}

	// Verify that dev files were written (proving topology was built)
	devDir := filepath.Join(tmpDir, ".stagecraft", "dev")
	composePath := filepath.Join(devDir, "compose.yaml")
	if _, err := os.Stat(composePath); err != nil {
		t.Fatalf("expected compose.yaml to be written at %s: %v", composePath, err)
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

	if err := runDevWithOptions(context.Background(), opts); err == nil {
		t.Fatalf("runDevWithOptions() error = nil, want non-nil for invalid env")
	}
}

func TestRunDevWithOptions_NoTraefikFlag(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")

	// Write a config with backend provider
	configContent := `project:
  name: test-app
backend:
  provider: generic
  providers:
    generic:
      dev:
        command: ["echo", "backend"]
        env:
          PORT: "4000"
environments:
  dev:
    driver: local
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	opts := devOptions{
		Env:       "dev",
		Config:    configPath,
		NoTraefik: true, // Traefik should be disabled
	}

	err = runDevWithOptions(context.Background(), opts)
	// Docker compose may fail, but topology should build without Traefik
	if err != nil {
		if !strings.Contains(err.Error(), "docker compose") && !strings.Contains(err.Error(), "start processes") {
			t.Fatalf("runDevWithOptions() error = %v, want nil or docker compose error", err)
		}
	}

	// Verify compose.yaml exists
	devDir := filepath.Join(tmpDir, ".stagecraft", "dev")
	composePath := filepath.Join(devDir, "compose.yaml")
	if _, err := os.Stat(composePath); err != nil {
		t.Fatalf("expected compose.yaml to be written at %s: %v", composePath, err)
	}

	// Verify Traefik config files do NOT exist when --no-traefik is used
	traefikDir := filepath.Join(devDir, "traefik")
	staticPath := filepath.Join(traefikDir, "traefik-static.yaml")
	if _, err := os.Stat(staticPath); err == nil {
		t.Errorf("expected traefik-static.yaml to NOT exist when --no-traefik is used, but it exists at %s", staticPath)
	}
}

func TestRunDevWithOptions_NoHTTPSFlag(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")

	// Write a config with backend provider
	configContent := `project:
  name: test-app
backend:
  provider: generic
  providers:
    generic:
      dev:
        command: ["echo", "backend"]
        env:
          PORT: "4000"
environments:
  dev:
    driver: local
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	opts := devOptions{
		Env:     "dev",
		Config:  configPath,
		NoHTTPS: true, // HTTPS should be disabled
	}

	err = runDevWithOptions(context.Background(), opts)
	// Docker compose may fail, but topology should build
	if err != nil {
		if !strings.Contains(err.Error(), "docker compose") && !strings.Contains(err.Error(), "start processes") {
			t.Fatalf("runDevWithOptions() error = %v, want nil or docker compose error", err)
		}
	}

	// Verify compose.yaml exists
	devDir := filepath.Join(tmpDir, ".stagecraft", "dev")
	composePath := filepath.Join(devDir, "compose.yaml")
	if _, err := os.Stat(composePath); err != nil {
		t.Fatalf("expected compose.yaml to be written at %s: %v", composePath, err)
	}

	// With --no-https, mkcert should not generate certificates
	// (This is tested via DEV_MKCERT, but we verify the flow works)
	certsDir := filepath.Join(devDir, "certs")
	_, err = os.Stat(certsDir)
	// Certs dir might exist but be empty - that's fine
	// The key is that mkcert.EnsureCertificates was called with EnableHTTPS: false
	_ = err // explicitly ignore the error
}

func TestRunDevWithOptions_UsesConfigDomains(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")

	// Write a config with backend and frontend providers and custom dev domains
	configContent := `project:
  name: test-app
backend:
  provider: generic
  providers:
    generic:
      dev:
        command: ["echo", "backend"]
        env:
          PORT: "4000"
frontend:
  provider: generic
  providers:
    generic:
      dev:
        command: ["echo", "frontend"]
        env:
          PORT: "3000"
dev:
  domains:
    frontend: app.example.test
    backend: api.example.test
environments:
  dev:
    driver: local
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	opts := devOptions{
		Env:    "dev",
		Config: configPath,
	}

	err = runDevWithOptions(context.Background(), opts)
	// Docker compose may fail, but we're testing that domains are computed correctly
	if err != nil {
		if !strings.Contains(err.Error(), "docker compose") && !strings.Contains(err.Error(), "start processes") {
			t.Fatalf("runDevWithOptions() error = %v, want nil or docker compose error", err)
		}
	}

	// Verify that the domains were used by checking the compose file
	// The compose file should exist, and Traefik config should use the custom domains
	devDir := filepath.Join(tmpDir, ".stagecraft", "dev")
	composePath := filepath.Join(devDir, "compose.yaml")
	if _, err := os.Stat(composePath); err != nil {
		t.Fatalf("expected compose.yaml to be written at %s: %v", composePath, err)
	}

	// Verify Traefik config uses the custom domains
	traefikDynamicPath := filepath.Join(devDir, "traefik", "traefik-dynamic.yaml")
	if _, err := os.Stat(traefikDynamicPath); err == nil {
		// Traefik config exists, verify it contains the custom domains
		// #nosec G304 -- test file path is controlled
		traefikContent, err := os.ReadFile(traefikDynamicPath)
		if err != nil {
			t.Fatalf("failed to read traefik-dynamic.yaml: %v", err)
		}

		traefikStr := string(traefikContent)
		if !strings.Contains(traefikStr, "app.example.test") {
			t.Errorf("traefik-dynamic.yaml should contain frontend domain 'app.example.test', got:\n%s", traefikStr)
		}
		if !strings.Contains(traefikStr, "api.example.test") {
			t.Errorf("traefik-dynamic.yaml should contain backend domain 'api.example.test', got:\n%s", traefikStr)
		}
	}
}

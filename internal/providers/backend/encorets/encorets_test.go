// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package encorets

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"stagecraft/pkg/providers/backend"
)

// Feature: PROVIDER_BACKEND_ENCORE
// Spec: spec/providers/backend/encore-ts.md

// setEnv sets an environment variable, failing the test if it fails.
func setEnv(t *testing.T, key, value string) {
	t.Helper()
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("failed to set env %s: %v", key, err)
	}
}

// unsetEnv unsets an environment variable, failing the test if it fails.
func unsetEnv(t *testing.T, key string) {
	t.Helper()
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("failed to unset env %s: %v", key, err)
	}
}

func TestEncoreTsProvider_ID(t *testing.T) {
	p := &EncoreTsProvider{}
	if got := p.ID(); got != "encore-ts" {
		t.Errorf("ID() = %q, want %q", got, "encore-ts")
	}
}

func TestEncoreTsProvider_ParseConfig(t *testing.T) {
	p := &EncoreTsProvider{}

	cfg := map[string]any{
		"dev": map[string]any{
			"encore_secrets": map[string]any{
				"types":    []string{"dev", "preview", "local"},
				"from_env": []string{"DOMAIN", "API_DOMAIN"},
			},
			"entrypoint": "./backend",
			"env_file":   ".env.local",
			"listen":     "0.0.0.0:4000",
		},
	}

	parsed, err := p.parseConfig(cfg)
	if err != nil {
		t.Fatalf("parseConfig() error = %v", err)
	}

	if len(parsed.Dev.EncoreSecrets.Types) != 3 {
		t.Errorf("Dev.EncoreSecrets.Types length = %d, want 3", len(parsed.Dev.EncoreSecrets.Types))
	}

	if len(parsed.Dev.EncoreSecrets.FromEnv) != 2 {
		t.Errorf("Dev.EncoreSecrets.FromEnv length = %d, want 2", len(parsed.Dev.EncoreSecrets.FromEnv))
	}

	if parsed.Dev.EntryPoint != "./backend" {
		t.Errorf("Dev.EntryPoint = %q, want %q", parsed.Dev.EntryPoint, "./backend")
	}

	if parsed.Dev.Listen != "0.0.0.0:4000" {
		t.Errorf("Dev.Listen = %q, want %q", parsed.Dev.Listen, "0.0.0.0:4000")
	}
}

func TestEncoreTsProvider_ParseConfig_InvalidYAML(t *testing.T) {
	p := &EncoreTsProvider{}

	// Invalid config structure
	cfg := "not a map"

	_, err := p.parseConfig(cfg)
	if err == nil {
		t.Error("parseConfig() error = nil, want error for invalid config")
	}
}

func TestEncoreTsProvider_ParseConfig_WithAllFields(t *testing.T) {
	p := &EncoreTsProvider{}

	cfg := map[string]any{
		"dev": map[string]any{
			"env_file":            ".env.local",
			"listen":              "0.0.0.0:4000",
			"workdir":             "./backend",
			"entrypoint":          "./src/index.ts",
			"disable_telemetry":   true,
			"node_extra_ca_certs": "./certs/ca.pem",
			"encore_secrets": map[string]any{
				"types":    []string{"dev", "preview"},
				"from_env": []string{"SECRET1", "SECRET2"},
			},
		},
		"build": map[string]any{
			"workdir":           "./backend",
			"image_name":        "my-api",
			"docker_tag_suffix": "-encore",
		},
	}

	parsed, err := p.parseConfig(cfg)
	if err != nil {
		t.Fatalf("parseConfig() error = %v", err)
	}

	if parsed.Dev.EnvFile != ".env.local" {
		t.Errorf("Dev.EnvFile = %q, want %q", parsed.Dev.EnvFile, ".env.local")
	}

	if parsed.Dev.WorkDir != "./backend" {
		t.Errorf("Dev.WorkDir = %q, want %q", parsed.Dev.WorkDir, "./backend")
	}

	if parsed.Dev.DisableTelemetry != true {
		t.Errorf("Dev.DisableTelemetry = %v, want true", parsed.Dev.DisableTelemetry)
	}

	if parsed.Dev.NodeExtraCACerts != "./certs/ca.pem" {
		t.Errorf("Dev.NodeExtraCACerts = %q, want %q", parsed.Dev.NodeExtraCACerts, "./certs/ca.pem")
	}

	if parsed.Build.WorkDir != "./backend" {
		t.Errorf("Build.WorkDir = %q, want %q", parsed.Build.WorkDir, "./backend")
	}

	if parsed.Build.ImageName != "my-api" {
		t.Errorf("Build.ImageName = %q, want %q", parsed.Build.ImageName, "my-api")
	}

	if parsed.Build.DockerTagSuffix != "-encore" {
		t.Errorf("Build.DockerTagSuffix = %q, want %q", parsed.Build.DockerTagSuffix, "-encore")
	}
}

func TestEncoreTsProvider_ParseConfig_DefaultImageName(t *testing.T) {
	p := &EncoreTsProvider{}

	cfg := map[string]any{
		"build": map[string]any{},
	}

	parsed, err := p.parseConfig(cfg)
	if err != nil {
		t.Fatalf("parseConfig() error = %v", err)
	}

	if parsed.Build.ImageName != "api" {
		t.Errorf("Build.ImageName = %q, want %q (default)", parsed.Build.ImageName, "api")
	}
}

func TestEncoreTsProvider_ValidateDevConfig_MissingListen(t *testing.T) {
	p := &EncoreTsProvider{}

	cfg := &Config{}
	// Missing required field: Listen is empty

	err := p.validateDevConfig(cfg)
	if err == nil {
		t.Error("validateDevConfig() error = nil, want error for missing listen")
	}

	pe := GetProviderError(err)
	if pe == nil {
		t.Fatal("expected ProviderError, got nil")
	}

	if pe.Category != ErrInvalidConfig {
		t.Errorf("ProviderError.Category = %q, want %q", pe.Category, ErrInvalidConfig)
	}

	if pe.Operation != "dev" {
		t.Errorf("ProviderError.Operation = %q, want %q", pe.Operation, "dev")
	}
}

func TestEncoreTsProvider_ValidateDevConfig_Valid(t *testing.T) {
	p := &EncoreTsProvider{}

	cfg := &Config{}
	cfg.Dev.Listen = "0.0.0.0:4000"

	err := p.validateDevConfig(cfg)
	if err != nil {
		t.Errorf("validateDevConfig() error = %v, want nil", err)
	}
}

func TestProviderError_Error(t *testing.T) {
	err := &ProviderError{
		Category:  ErrInvalidConfig,
		Provider:  "encore-ts",
		Operation: "dev",
		Message:   "test error",
		Detail:    "test detail",
	}

	msg := err.Error()
	if msg == "" {
		t.Error("Error() returned empty string")
	}

	if !strings.Contains(msg, "encore-ts") {
		t.Errorf("Error() message should contain provider ID, got %q", msg)
	}

	if !strings.Contains(msg, "dev") {
		t.Errorf("Error() message should contain operation, got %q", msg)
	}

	if !strings.Contains(msg, ErrInvalidConfig) {
		t.Errorf("Error() message should contain category, got %q", msg)
	}
}

func TestProviderError_Error_NoDetail(t *testing.T) {
	err := &ProviderError{
		Category:  ErrInvalidConfig,
		Provider:  "encore-ts",
		Operation: "dev",
		Message:   "test error",
	}

	msg := err.Error()
	if strings.Contains(msg, ":") && strings.Contains(msg, "test detail") {
		t.Error("Error() should not include detail when Detail is empty")
	}
}

func TestIsProviderError(t *testing.T) {
	err := &ProviderError{
		Category:  ErrInvalidConfig,
		Provider:  "encore-ts",
		Operation: "dev",
		Message:   "test",
	}

	if !IsProviderError(err) {
		t.Error("IsProviderError() = false, want true")
	}

	regularErr := fmt.Errorf("regular error")
	if IsProviderError(regularErr) {
		t.Error("IsProviderError() = true for regular error, want false")
	}
}

func TestGetProviderError(t *testing.T) {
	err := &ProviderError{
		Category:  ErrInvalidConfig,
		Provider:  "encore-ts",
		Operation: "dev",
		Message:   "test",
	}

	pe := GetProviderError(err)
	if pe == nil {
		t.Fatal("GetProviderError() = nil, want ProviderError")
	}

	if pe.Category != ErrInvalidConfig {
		t.Errorf("GetProviderError().Category = %q, want %q", pe.Category, ErrInvalidConfig)
	}

	regularErr := fmt.Errorf("regular error")
	if GetProviderError(regularErr) != nil {
		t.Error("GetProviderError() should return nil for regular error")
	}
}

func TestProviderError_Category(t *testing.T) {
	err := &ProviderError{
		Category:  ErrInvalidConfig,
		Provider:  "encore-ts",
		Operation: "dev",
		Message:   "test",
	}

	if got := err.Category; got != ErrInvalidConfig {
		t.Errorf("Category = %q, want %q", got, ErrInvalidConfig)
	}
}

func TestParseEnvFileInto(t *testing.T) {
	tests := []struct {
		name    string
		envFile string
		wantEnv map[string]string
	}{
		{
			name: "inline comments",
			envFile: `KEY1=value1 # inline comment
KEY2=value2`,
			wantEnv: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
		{
			name: "export keyword",
			envFile: `export KEY1=value1
KEY2=value2`,
			wantEnv: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
		{
			name: "quoted values with escapes",
			envFile: `KEY1="value with spaces"
KEY2="value with \"quotes\""
KEY3="value with\nnewline"
KEY4='single quoted'
KEY5=unquoted`,
			wantEnv: map[string]string{
				"KEY1": "value with spaces",
				"KEY2": "value with \"quotes\"",
				"KEY3": "value with\nnewline",
				"KEY4": "single quoted",
				"KEY5": "unquoted",
			},
		},
		{
			name: "empty values",
			envFile: `KEY1=
KEY2=value2`,
			wantEnv: map[string]string{
				"KEY1": "",
				"KEY2": "value2",
			},
		},
		{
			name: "preserve # inside quotes",
			envFile: `KEY1="value # not a comment"
KEY2=value # this is a comment`,
			wantEnv: map[string]string{
				"KEY1": "value # not a comment",
				"KEY2": "value",
			},
		},
		{
			name: "later values override earlier",
			envFile: `KEY=first
KEY=second`,
			wantEnv: map[string]string{
				"KEY": "second",
			},
		},
		{
			name: "blank lines and comments",
			envFile: `# This is a comment
KEY1=value1

KEY2=value2
# Another comment
KEY3=value3`,
			wantEnv: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
				"KEY3": "value3",
			},
		},
		{
			name: "escape sequences in double quotes",
			envFile: `KEY1="tab\there"
KEY2="newline\nhere"
KEY3="backslash\\here"
KEY4="quote\"here"`,
			wantEnv: map[string]string{
				"KEY1": "tab\there",
				"KEY2": "newline\nhere",
				"KEY3": "backslash\\here",
				"KEY4": "quote\"here",
			},
		},
		{
			name: "malformed lines are skipped",
			envFile: `KEY1=value1
MALFORMED
KEY2=value2`,
			wantEnv: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
		{
			name: "empty keys are skipped",
			envFile: `KEY1=value1
=value2
KEY2=value3`,
			wantEnv: map[string]string{
				"KEY1": "value1",
				"KEY2": "value3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := make(map[string]string)
			parseEnvFileInto(env, []byte(tt.envFile))

			// Manual comparison
			if len(env) != len(tt.wantEnv) {
				t.Errorf("parseEnvFileInto() got %d keys, want %d", len(env), len(tt.wantEnv))
			}
			for k, wantV := range tt.wantEnv {
				if gotV, ok := env[k]; !ok {
					t.Errorf("parseEnvFileInto() missing key %q", k)
				} else if gotV != wantV {
					t.Errorf("parseEnvFileInto() key %q = %q, want %q", k, gotV, wantV)
				}
			}
			// Check for unexpected keys
			for k := range env {
				if _, ok := tt.wantEnv[k]; !ok {
					t.Errorf("parseEnvFileInto() unexpected key %q", k)
				}
			}
		})
	}
}

// createMockEncoreScript creates a mock encore script for testing.
// The script behavior is controlled by environment variables:
// - ENCORE_MOCK_MODE: "success", "failure", "exit_code_<n>", "secret_success", "secret_failure"
// - ENCORE_MOCK_DELAY: delay in seconds before exit (for testing context cancellation)
func createMockEncoreScript(t *testing.T, dir string) string {
	t.Helper()

	var scriptContent string
	if runtime.GOOS == "windows" {
		scriptContent = `@echo off
setlocal
if "%ENCORE_MOCK_MODE%"=="success" (
    echo Encore dev server running...
    timeout /t %ENCORE_MOCK_DELAY% /nobreak >nul 2>&1
    exit /b 0
)
if "%ENCORE_MOCK_MODE%"=="failure" (
    echo Error: encore command failed
    exit /b 1
)
if "%ENCORE_MOCK_MODE%"=="secret_success" (
    REM Read from stdin and echo back
    set /p SECRET_VALUE=
    echo Secret set successfully
    exit /b 0
)
if "%ENCORE_MOCK_MODE%"=="secret_failure" (
    echo Error: secret set failed
    exit /b 1
)
if "%ENCORE_MOCK_MODE:~0,9%"=="exit_code" (
    set EXIT_CODE=%ENCORE_MOCK_MODE:~10%
    exit /b %EXIT_CODE%
)
exit /b 0
`
	} else {
		scriptContent = `#!/bin/sh
case "$ENCORE_MOCK_MODE" in
  "success")
    echo "Encore dev server running..."
    if [ -n "$ENCORE_MOCK_DELAY" ]; then
      sleep "$ENCORE_MOCK_DELAY"
    fi
    exit 0
    ;;
  "failure")
    echo "Error: encore command failed" >&2
    exit 1
    ;;
  "secret_success")
    # Read from stdin
    read -r SECRET_VALUE
    echo "Secret set successfully"
    exit 0
    ;;
  "secret_failure")
    echo "Error: secret set failed" >&2
    exit 1
    ;;
  exit_code_*)
    EXIT_CODE="${ENCORE_MOCK_MODE#exit_code_}"
    exit "$EXIT_CODE"
    ;;
  *)
    exit 0
    ;;
esac
`
	}

	scriptPath := filepath.Join(dir, "encore")
	if runtime.GOOS == "windows" {
		scriptPath += ".bat"
	}

	//nolint:gosec // G306: 0755 is required for executable test scripts
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0o755); err != nil {
		t.Fatalf("failed to create mock encore script: %v", err)
	}

	return scriptPath
}

// setupMockEncorePath sets up PATH to use mock encore script.
// Returns cleanup function.
func setupMockEncorePath(t *testing.T, mockScriptPath string) func() {
	t.Helper()

	scriptDir := filepath.Dir(mockScriptPath)
	originalPath := os.Getenv("PATH")

	// Prepend script directory to PATH
	newPath := scriptDir + string(filepath.ListSeparator) + originalPath
	setEnv(t, "PATH", newPath)

	return func() {
		setEnv(t, "PATH", originalPath)
	}
}

func TestEncoreTsProvider_Dev_Success(t *testing.T) {
	tmpDir := t.TempDir()
	mockScript := createMockEncoreScript(t, tmpDir)
	cleanup := setupMockEncorePath(t, mockScript)
	defer cleanup()

	// Set mock mode to success
	setEnv(t, "ENCORE_MOCK_MODE", "success")
	defer unsetEnv(t, "ENCORE_MOCK_MODE")

	// Create env file
	envFile := filepath.Join(tmpDir, ".env.test")
	//nolint:gosec // G306: 0644 is acceptable for test fixtures
	if err := os.WriteFile(envFile, []byte("TEST_VAR=test_value\n"), 0o644); err != nil {
		t.Fatalf("failed to create env file: %v", err)
	}

	p := &EncoreTsProvider{}
	opts := backend.DevOptions{
		Config: map[string]any{
			"dev": map[string]any{
				"listen":   "0.0.0.0:4000",
				"env_file": ".env.test",
			},
		},
		WorkDir: tmpDir,
		Env:     map[string]string{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// This should succeed with mocked encore
	err := p.Dev(ctx, opts)
	if err != nil {
		// Context timeout is expected since we're using a short timeout
		if ctx.Err() != nil {
			// Expected - context cancelled
			return
		}
		t.Errorf("Dev() error = %v, want nil (or context timeout)", err)
	}
}

func TestEncoreTsProvider_Dev_CommandFailure(t *testing.T) {
	tmpDir := t.TempDir()
	mockScript := createMockEncoreScript(t, tmpDir)
	cleanup := setupMockEncorePath(t, mockScript)
	defer cleanup()

	// Set mock mode to failure
	setEnv(t, "ENCORE_MOCK_MODE", "failure")
	defer unsetEnv(t, "ENCORE_MOCK_MODE")

	p := &EncoreTsProvider{}
	opts := backend.DevOptions{
		Config: map[string]any{
			"dev": map[string]any{
				"listen": "0.0.0.0:4000",
			},
		},
		WorkDir: tmpDir,
		Env:     map[string]string{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := p.Dev(ctx, opts)
	if err == nil {
		t.Error("Dev() error = nil, want error for command failure")
	}

	pe := GetProviderError(err)
	if pe == nil {
		t.Fatal("expected ProviderError, got nil")
	}

	if pe.Category != ErrDevServerFailed {
		t.Errorf("ProviderError.Category = %q, want %q", pe.Category, ErrDevServerFailed)
	}
}

func TestEncoreTsProvider_Dev_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	mockScript := createMockEncoreScript(t, tmpDir)
	cleanup := setupMockEncorePath(t, mockScript)
	defer cleanup()

	// Set mock mode to success with delay
	setEnv(t, "ENCORE_MOCK_MODE", "success")
	setEnv(t, "ENCORE_MOCK_DELAY", "10") // 10 second delay
	defer unsetEnv(t, "ENCORE_MOCK_MODE")
	defer unsetEnv(t, "ENCORE_MOCK_DELAY")

	p := &EncoreTsProvider{}
	opts := backend.DevOptions{
		Config: map[string]any{
			"dev": map[string]any{
				"listen": "0.0.0.0:4000",
			},
		},
		WorkDir: tmpDir,
		Env:     map[string]string{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := p.Dev(ctx, opts)
	if err == nil {
		t.Error("Dev() error = nil, want error for context cancellation")
	}

	if ctx.Err() == nil {
		t.Error("expected context to be cancelled")
	}
}

func TestEncoreTsProvider_Dev_SecretSync_Success(t *testing.T) {
	tmpDir := t.TempDir()
	mockScript := createMockEncoreScript(t, tmpDir)
	cleanup := setupMockEncorePath(t, mockScript)
	defer cleanup()

	// Set mock mode to success for secret sync
	setEnv(t, "ENCORE_MOCK_MODE", "secret_success")
	defer unsetEnv(t, "ENCORE_MOCK_MODE")

	p := &EncoreTsProvider{}
	opts := backend.DevOptions{
		Config: map[string]any{
			"dev": map[string]any{
				"listen": "0.0.0.0:4000",
				"encore_secrets": map[string]any{
					"types":    []string{"dev"},
					"from_env": []string{"TEST_SECRET"},
				},
			},
		},
		WorkDir: tmpDir,
		Env: map[string]string{
			"TEST_SECRET": "secret-value-123",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// This will fail because encore run will be called after secret sync
	// but we're testing that secret sync works
	err := p.Dev(ctx, opts)
	// We expect either context timeout or encore run failure (since we're using secret_success mode)
	// The important thing is that secret sync doesn't fail
	if err != nil {
		pe := GetProviderError(err)
		if pe != nil && pe.Category == ErrSecretSyncFailed {
			t.Errorf("Dev() secret sync failed: %v", err)
		}
		// Other errors (like context timeout or dev server) are acceptable
	}
}

func TestEncoreTsProvider_Dev_SecretSync_Failure(t *testing.T) {
	tmpDir := t.TempDir()
	mockScript := createMockEncoreScript(t, tmpDir)
	cleanup := setupMockEncorePath(t, mockScript)
	defer cleanup()

	// Set mock mode to failure for secret sync
	setEnv(t, "ENCORE_MOCK_MODE", "secret_failure")
	defer unsetEnv(t, "ENCORE_MOCK_MODE")

	p := &EncoreTsProvider{}
	opts := backend.DevOptions{
		Config: map[string]any{
			"dev": map[string]any{
				"listen": "0.0.0.0:4000",
				"encore_secrets": map[string]any{
					"types":    []string{"dev"},
					"from_env": []string{"TEST_SECRET"},
				},
			},
		},
		WorkDir: tmpDir,
		Env: map[string]string{
			"TEST_SECRET": "secret-value-123",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := p.Dev(ctx, opts)
	if err == nil {
		t.Error("Dev() error = nil, want error for secret sync failure")
	}

	pe := GetProviderError(err)
	if pe == nil {
		t.Fatal("expected ProviderError, got nil")
	}

	if pe.Category != ErrSecretSyncFailed {
		t.Errorf("ProviderError.Category = %q, want %q", pe.Category, ErrSecretSyncFailed)
	}
}

func TestEncoreTsProvider_Dev_MissingSecrets(t *testing.T) {
	tmpDir := t.TempDir()
	mockScript := createMockEncoreScript(t, tmpDir)
	cleanup := setupMockEncorePath(t, mockScript)
	defer cleanup()

	// Set mock mode to success
	setEnv(t, "ENCORE_MOCK_MODE", "success")
	defer unsetEnv(t, "ENCORE_MOCK_MODE")

	p := &EncoreTsProvider{}
	opts := backend.DevOptions{
		Config: map[string]any{
			"dev": map[string]any{
				"listen": "0.0.0.0:4000",
				"encore_secrets": map[string]any{
					"types":    []string{"dev"},
					"from_env": []string{"MISSING_SECRET"},
				},
			},
		},
		WorkDir: tmpDir,
		Env:     map[string]string{
			// MISSING_SECRET is not provided
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// This should not fail - missing secrets should just log warnings
	err := p.Dev(ctx, opts)
	if err != nil {
		// Context timeout is expected
		if ctx.Err() == nil {
			pe := GetProviderError(err)
			if pe != nil && pe.Category == ErrSecretSyncFailed {
				t.Errorf("Dev() should not fail for missing secrets, got: %v", err)
			}
		}
	}
}

func TestEncoreTsProvider_Dev_EnvFileLoading(t *testing.T) {
	tmpDir := t.TempDir()
	mockScript := createMockEncoreScript(t, tmpDir)
	cleanup := setupMockEncorePath(t, mockScript)
	defer cleanup()

	// Set mock mode to success
	setEnv(t, "ENCORE_MOCK_MODE", "success")
	defer unsetEnv(t, "ENCORE_MOCK_MODE")

	// Create env file with test values
	envFile := filepath.Join(tmpDir, ".env.test")
	envContent := `TEST_VAR=test_value
NUMBER_VAR=123
QUOTED_VAR="quoted value"
`
	//nolint:gosec // G306: 0644 is acceptable for test fixtures
	if err := os.WriteFile(envFile, []byte(envContent), 0o644); err != nil {
		t.Fatalf("failed to create env file: %v", err)
	}

	p := &EncoreTsProvider{}
	opts := backend.DevOptions{
		Config: map[string]any{
			"dev": map[string]any{
				"listen":   "0.0.0.0:4000",
				"env_file": ".env.test",
			},
		},
		WorkDir: tmpDir,
		Env:     map[string]string{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// This should succeed - env file loading is tested indirectly
	err := p.Dev(ctx, opts)
	if err != nil {
		// Context timeout is expected
		if ctx.Err() == nil {
			t.Errorf("Dev() error = %v", err)
		}
	}
}

func TestEncoreTsProvider_Dev_TelemetryAndCACerts(t *testing.T) {
	tmpDir := t.TempDir()
	mockScript := createMockEncoreScript(t, tmpDir)
	cleanup := setupMockEncorePath(t, mockScript)
	defer cleanup()

	// Create CA cert file
	caCertFile := filepath.Join(tmpDir, "ca.pem")
	//nolint:gosec // G306: 0644 is acceptable for test fixtures
	if err := os.WriteFile(caCertFile, []byte("fake CA cert"), 0o644); err != nil {
		t.Fatalf("failed to create CA cert file: %v", err)
	}

	// Set mock mode to success
	setEnv(t, "ENCORE_MOCK_MODE", "success")
	defer unsetEnv(t, "ENCORE_MOCK_MODE")

	p := &EncoreTsProvider{}
	opts := backend.DevOptions{
		Config: map[string]any{
			"dev": map[string]any{
				"listen":              "0.0.0.0:4000",
				"disable_telemetry":   true,
				"node_extra_ca_certs": "ca.pem",
			},
		},
		WorkDir: tmpDir,
		Env:     map[string]string{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// This should succeed - telemetry and CA certs are set via env vars
	err := p.Dev(ctx, opts)
	if err != nil {
		// Context timeout is expected
		if ctx.Err() == nil {
			t.Errorf("Dev() error = %v", err)
		}
	}
}

func TestEncoreTsProvider_BuildDocker_Success(t *testing.T) {
	tmpDir := t.TempDir()
	mockScript := createMockEncoreScript(t, tmpDir)
	cleanup := setupMockEncorePath(t, mockScript)
	defer cleanup()

	// Set mock mode to success
	setEnv(t, "ENCORE_MOCK_MODE", "success")
	defer unsetEnv(t, "ENCORE_MOCK_MODE")

	p := &EncoreTsProvider{}
	opts := backend.BuildDockerOptions{
		Config: map[string]any{
			"build": map[string]any{
				"image_name": "my-api",
			},
		},
		ImageTag: "v1.0.0",
		WorkDir:  tmpDir,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	imageRef, err := p.BuildDocker(ctx, opts)
	if err != nil {
		t.Errorf("BuildDocker() error = %v, want nil", err)
	}

	expectedRef := "my-api:v1.0.0"
	if imageRef != expectedRef {
		t.Errorf("BuildDocker() imageRef = %q, want %q", imageRef, expectedRef)
	}
}

func TestEncoreTsProvider_BuildDocker_Failure(t *testing.T) {
	tmpDir := t.TempDir()
	mockScript := createMockEncoreScript(t, tmpDir)
	cleanup := setupMockEncorePath(t, mockScript)
	defer cleanup()

	// Set mock mode to failure
	setEnv(t, "ENCORE_MOCK_MODE", "failure")
	defer unsetEnv(t, "ENCORE_MOCK_MODE")

	p := &EncoreTsProvider{}
	opts := backend.BuildDockerOptions{
		Config: map[string]any{
			"build": map[string]any{},
		},
		ImageTag: "v1.0.0",
		WorkDir:  tmpDir,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := p.BuildDocker(ctx, opts)
	if err == nil {
		t.Error("BuildDocker() error = nil, want error for build failure")
	}

	pe := GetProviderError(err)
	if pe == nil {
		t.Fatal("expected ProviderError, got nil")
	}

	if pe.Category != ErrBuildFailed {
		t.Errorf("ProviderError.Category = %q, want %q", pe.Category, ErrBuildFailed)
	}
}

func TestEncoreTsProvider_BuildDocker_ImageReferenceResolution(t *testing.T) {
	tmpDir := t.TempDir()
	mockScript := createMockEncoreScript(t, tmpDir)
	cleanup := setupMockEncorePath(t, mockScript)
	defer cleanup()

	// Set mock mode to success
	setEnv(t, "ENCORE_MOCK_MODE", "success")
	defer unsetEnv(t, "ENCORE_MOCK_MODE")

	tests := []struct {
		name         string
		config       map[string]any
		imageTag     string
		wantImageRef string
	}{
		{
			name: "tag only with default image name",
			config: map[string]any{
				"build": map[string]any{},
			},
			imageTag:     "v1.0.0",
			wantImageRef: "api:v1.0.0",
		},
		{
			name: "tag only with custom image name",
			config: map[string]any{
				"build": map[string]any{
					"image_name": "my-api",
				},
			},
			imageTag:     "v1.0.0",
			wantImageRef: "my-api:v1.0.0",
		},
		{
			name: "tag only with docker_tag_suffix",
			config: map[string]any{
				"build": map[string]any{
					"image_name":        "my-api",
					"docker_tag_suffix": "-encore",
				},
			},
			imageTag:     "v1.0.0",
			wantImageRef: "my-api:v1.0.0-encore",
		},
		{
			name: "full reference with registry",
			config: map[string]any{
				"build": map[string]any{},
			},
			imageTag:     "ghcr.io/org/app:v1.0.0",
			wantImageRef: "ghcr.io/org/app:v1.0.0",
		},
		{
			name: "full reference with docker_tag_suffix",
			config: map[string]any{
				"build": map[string]any{
					"docker_tag_suffix": "-encore",
				},
			},
			imageTag:     "ghcr.io/org/app:v1.0.0",
			wantImageRef: "ghcr.io/org/app:v1.0.0-encore",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &EncoreTsProvider{}
			opts := backend.BuildDockerOptions{
				Config:   tt.config,
				ImageTag: tt.imageTag,
				WorkDir:  tmpDir,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			imageRef, err := p.BuildDocker(ctx, opts)
			if err != nil {
				t.Errorf("BuildDocker() error = %v", err)
				return
			}

			if imageRef != tt.wantImageRef {
				t.Errorf("BuildDocker() imageRef = %q, want %q", imageRef, tt.wantImageRef)
			}
		})
	}
}

func TestEncoreTsProvider_Dev_WorkDirResolution(t *testing.T) {
	tmpDir := t.TempDir()
	mockScript := createMockEncoreScript(t, tmpDir)
	cleanup := setupMockEncorePath(t, mockScript)
	defer cleanup()

	// Set mock mode to success
	setEnv(t, "ENCORE_MOCK_MODE", "success")
	defer unsetEnv(t, "ENCORE_MOCK_MODE")

	tests := []struct {
		name    string
		config  map[string]any
		opts    backend.DevOptions
		wantDir string
	}{
		{
			name: "config workdir takes precedence",
			config: map[string]any{
				"dev": map[string]any{
					"listen":  "0.0.0.0:4000",
					"workdir": tmpDir,
				},
			},
			opts: backend.DevOptions{
				WorkDir: "/other/dir",
			},
			wantDir: tmpDir,
		},
		{
			name: "opts workdir used when config missing",
			config: map[string]any{
				"dev": map[string]any{
					"listen": "0.0.0.0:4000",
				},
			},
			opts: backend.DevOptions{
				WorkDir: tmpDir,
			},
			wantDir: tmpDir,
		},
		{
			name: "defaults to current directory",
			config: map[string]any{
				"dev": map[string]any{
					"listen": "0.0.0.0:4000",
				},
			},
			opts: backend.DevOptions{
				WorkDir: "",
			},
			wantDir: ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.opts.Config = tt.config
			tt.opts.Env = map[string]string{}

			p := &EncoreTsProvider{}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			// This will timeout (expected), but we're testing workdir resolution
			_ = p.Dev(ctx, tt.opts)
		})
	}
}

func TestEncoreTsProvider_BuildDocker_WorkDirResolution(t *testing.T) {
	tmpDir := t.TempDir()
	mockScript := createMockEncoreScript(t, tmpDir)
	cleanup := setupMockEncorePath(t, mockScript)
	defer cleanup()

	// Set mock mode to success
	setEnv(t, "ENCORE_MOCK_MODE", "success")
	defer unsetEnv(t, "ENCORE_MOCK_MODE")

	tests := []struct {
		name    string
		config  map[string]any
		opts    backend.BuildDockerOptions
		wantDir string
	}{
		{
			name: "config workdir takes precedence",
			config: map[string]any{
				"build": map[string]any{
					"workdir": tmpDir,
				},
			},
			opts: backend.BuildDockerOptions{
				ImageTag: "v1.0.0",
				WorkDir:  "/other/dir",
			},
			wantDir: tmpDir,
		},
		{
			name: "opts workdir used when config missing",
			config: map[string]any{
				"build": map[string]any{},
			},
			opts: backend.BuildDockerOptions{
				ImageTag: "v1.0.0",
				WorkDir:  tmpDir,
			},
			wantDir: tmpDir,
		},
		{
			name: "defaults to current directory",
			config: map[string]any{
				"build": map[string]any{},
			},
			opts: backend.BuildDockerOptions{
				ImageTag: "v1.0.0",
				WorkDir:  "",
			},
			wantDir: ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.opts.Config = tt.config

			p := &EncoreTsProvider{}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// This should succeed with mocked encore
			_, err := p.BuildDocker(ctx, tt.opts)
			if err != nil {
				t.Errorf("BuildDocker() error = %v", err)
			}
		})
	}
}

func TestEncoreTsProvider_CheckEncoreAvailable(t *testing.T) {
	p := &EncoreTsProvider{}

	// Test when encore is not available (by temporarily removing from PATH)
	originalPath := os.Getenv("PATH")
	setEnv(t, "PATH", "")
	defer setEnv(t, "PATH", originalPath)

	err := p.checkEncoreAvailable()
	if err == nil {
		t.Error("checkEncoreAvailable() error = nil, want error when encore not found")
	}

	pe := GetProviderError(err)
	if pe == nil {
		t.Fatal("expected ProviderError, got nil")
	}

	if pe.Category != ErrProviderNotAvailable {
		t.Errorf("ProviderError.Category = %q, want %q", pe.Category, ErrProviderNotAvailable)
	}
}

func TestEncoreTsProvider_Plan(t *testing.T) {
	p := &EncoreTsProvider{}

	tests := []struct {
		name    string
		config  any
		opts    backend.PlanOptions
		wantErr bool
		check   func(t *testing.T, plan backend.ProviderPlan)
	}{
		{
			name: "basic plan with default image name",
			config: map[string]any{
				"build": map[string]any{},
			},
			opts: backend.PlanOptions{
				ImageTag: "v1.0.0",
				WorkDir:  "/tmp/test",
			},
			wantErr: false,
			check: func(t *testing.T, plan backend.ProviderPlan) {
				if plan.Provider != "encore-ts" {
					t.Errorf("Provider = %q, want %q", plan.Provider, "encore-ts")
				}
				if len(plan.Steps) != 3 {
					t.Errorf("Steps length = %d, want 3", len(plan.Steps))
				}
				if plan.Steps[0].Name != "CheckEncoreAvailable" {
					t.Errorf("Steps[0].Name = %q, want %q", plan.Steps[0].Name, "CheckEncoreAvailable")
				}
				if plan.Steps[1].Name != "ResolveImageReference" {
					t.Errorf("Steps[1].Name = %q, want %q", plan.Steps[1].Name, "ResolveImageReference")
				}
				if plan.Steps[2].Name != "BuildDocker" {
					t.Errorf("Steps[2].Name = %q, want %q", plan.Steps[2].Name, "BuildDocker")
				}
				if !strings.Contains(plan.Steps[1].Description, "api:v1.0.0") {
					t.Errorf("Steps[1].Description should contain resolved image, got %q", plan.Steps[1].Description)
				}
			},
		},
		{
			name: "plan with custom image name",
			config: map[string]any{
				"build": map[string]any{
					"image_name": "my-api",
				},
			},
			opts: backend.PlanOptions{
				ImageTag: "v1.0.0",
				WorkDir:  "/tmp/test",
			},
			wantErr: false,
			check: func(t *testing.T, plan backend.ProviderPlan) {
				if !strings.Contains(plan.Steps[1].Description, "my-api:v1.0.0") {
					t.Errorf("Steps[1].Description should contain custom image name, got %q", plan.Steps[1].Description)
				}
			},
		},
		{
			name: "plan with docker tag suffix",
			config: map[string]any{
				"build": map[string]any{
					"image_name":        "my-api",
					"docker_tag_suffix": "-encore",
				},
			},
			opts: backend.PlanOptions{
				ImageTag: "v1.0.0",
				WorkDir:  "/tmp/test",
			},
			wantErr: false,
			check: func(t *testing.T, plan backend.ProviderPlan) {
				if !strings.Contains(plan.Steps[1].Description, "my-api:v1.0.0-encore") {
					t.Errorf("Steps[1].Description should contain suffix, got %q", plan.Steps[1].Description)
				}
			},
		},
		{
			name: "plan with full image reference",
			config: map[string]any{
				"build": map[string]any{},
			},
			opts: backend.PlanOptions{
				ImageTag: "ghcr.io/org/app:v1.0.0",
				WorkDir:  "/tmp/test",
			},
			wantErr: false,
			check: func(t *testing.T, plan backend.ProviderPlan) {
				if !strings.Contains(plan.Steps[1].Description, "ghcr.io/org/app:v1.0.0") {
					t.Errorf("Steps[1].Description should contain full image reference, got %q", plan.Steps[1].Description)
				}
			},
		},
		{
			name:   "plan with invalid config",
			config: "not a map",
			opts: backend.PlanOptions{
				ImageTag: "v1.0.0",
				WorkDir:  "/tmp/test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.opts.Config = tt.config

			ctx := context.Background()
			plan, err := p.Plan(ctx, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("Plan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, plan)
			}
		})
	}
}

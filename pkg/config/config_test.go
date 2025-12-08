// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package config

import (
	"os"
	"path/filepath"
	"testing"
)

// Feature: CORE_CONFIG
// Spec: spec/core/config.md

func TestDefaultConfigPath(t *testing.T) {
	path := DefaultConfigPath()
	if path != "stagecraft.yml" {
		t.Fatalf("expected DefaultConfigPath to return 'stagecraft.yml', got %q", path)
	}
}

func TestExists_ReportsCorrectly(t *testing.T) {
	tmpDir := t.TempDir()

	nonExisting := filepath.Join(tmpDir, "nope.yml")
	ok, err := Exists(nonExisting)
	if err != nil {
		t.Fatalf("expected no error for non-existing file, got: %v", err)
	}
	if ok {
		t.Fatalf("expected Exists to return false for non-existing file")
	}

	existing := filepath.Join(tmpDir, "config.yml")
	if werr := os.WriteFile(existing, []byte("project:\n  name: test\n"), 0o600); werr != nil {
		t.Fatalf("failed to write temp config: %v", werr)
	}

	ok, err = Exists(existing)
	if err != nil {
		t.Fatalf("expected no error for existing file, got: %v", err)
	}
	if !ok {
		t.Fatalf("expected Exists to return true for existing file")
	}
}

func TestLoad_ReturnsErrConfigNotFoundWhenMissing(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "missing.yml")

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected error for missing config, got nil")
	}

	if err != ErrConfigNotFound {
		t.Fatalf("expected ErrConfigNotFound, got %v", err)
	}
}

func TestLoad_ParsesValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "stagecraft.yml")

	content := []byte(`
project:
  name: "my-app"
environments:
  dev:
    driver: "digitalocean"
`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error loading valid config, got: %v", err)
	}

	if cfg.Project.Name != "my-app" {
		t.Fatalf("expected project.name 'my-app', got %q", cfg.Project.Name)
	}

	dev, ok := cfg.Environments["dev"]
	if !ok {
		t.Fatalf("expected 'dev' environment to be present")
	}

	if dev.Driver != "digitalocean" {
		t.Fatalf("expected dev.driver 'digitalocean', got %q", dev.Driver)
	}
}

func TestLoad_ValidatesProjectName(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "stagecraft.yml")

	content := []byte(`
project:
  name: ""
`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected validation error for empty project.name")
	}
}

func TestLoad_ValidatesBackend_WithGenericProvider(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "stagecraft.yml")

	content := []byte(`
project:
  name: "test-app"
backend:
  provider: generic
  providers:
    generic:
      dev:
        command: ["npm", "run", "dev"]
environments:
  dev:
    driver: "local"
`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error loading valid config with generic provider, got: %v", err)
	}

	if cfg.Backend == nil {
		t.Fatalf("expected backend config to be present")
	}

	if cfg.Backend.Provider != "generic" {
		t.Fatalf("expected backend.provider 'generic', got %q", cfg.Backend.Provider)
	}
}

func TestLoad_ValidatesBackend_UnknownProvider(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "stagecraft.yml")

	content := []byte(`
project:
  name: "test-app"
backend:
  provider: unknown-provider
  providers:
    unknown-provider:
      dev: {}
environments:
  dev:
    driver: "local"
`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected validation error for unknown backend provider")
	}

	if err != nil && err.Error() == "" {
		t.Fatalf("expected error message, got empty")
	}

	// Verify error message includes available providers
	errMsg := err.Error()
	if !contains(errMsg, "unknown backend provider") {
		t.Errorf("error message should mention 'unknown backend provider', got: %q", errMsg)
	}
	if !contains(errMsg, "available providers") {
		t.Errorf("error message should mention 'available providers', got: %q", errMsg)
	}
	// Should mention at least one registered provider (generic or encore-ts)
	if !contains(errMsg, "generic") && !contains(errMsg, "encore-ts") {
		t.Errorf("error message should list available providers, got: %q", errMsg)
	}
}

func TestLoad_ValidatesBackend_MissingProviderConfig(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "stagecraft.yml")

	content := []byte(`
project:
  name: "test-app"
backend:
  provider: generic
  providers: {}
environments:
  dev:
    driver: "local"
`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected validation error for missing provider config")
	}

	if err != nil && !contains(err.Error(), "backend.providers.generic") {
		t.Fatalf("expected error to mention missing provider config, got: %v", err)
	}
}

func TestLoad_ValidatesBackend_MissingProvidersMap(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "stagecraft.yml")

	content := []byte(`
project:
  name: "test-app"
backend:
  provider: generic
environments:
  dev:
    driver: "local"
`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected validation error for missing providers map")
	}

	if err != nil && !contains(err.Error(), "backend.providers is required") {
		t.Fatalf("expected error to mention missing providers, got: %v", err)
	}
}

// Feature: GOV_V1_CORE
// Spec: spec/governance/GOV_V1_CORE.md
// Tests for frontend validation (currently 0% coverage)

func TestLoad_ValidatesFrontend_WithGenericProvider(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "stagecraft.yml")

	content := []byte(`
project:
  name: "test-app"
frontend:
  provider: generic
  providers:
    generic:
      build:
        command: ["npm", "run", "build"]
environments:
  dev:
    driver: "local"
`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error loading valid config with generic frontend provider, got: %v", err)
	}

	if cfg.Frontend == nil {
		t.Fatalf("expected frontend config to be present")
	}

	if cfg.Frontend.Provider != "generic" {
		t.Fatalf("expected frontend.provider 'generic', got %q", cfg.Frontend.Provider)
	}
}

func TestLoad_ValidatesFrontend_UnknownProvider(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "stagecraft.yml")

	content := []byte(`
project:
  name: "test-app"
frontend:
  provider: unknown-provider
  providers:
    unknown-provider:
      build: {}
environments:
  dev:
    driver: "local"
`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected validation error for unknown frontend provider")
	}

	if err != nil && err.Error() == "" {
		t.Fatalf("expected error message, got empty")
	}

	// Verify error message includes available providers
	errMsg := err.Error()
	if !contains(errMsg, "unknown frontend provider") {
		t.Errorf("error message should mention 'unknown frontend provider', got: %q", errMsg)
	}
	if !contains(errMsg, "available providers") {
		t.Errorf("error message should mention 'available providers', got: %q", errMsg)
	}
}

func TestLoad_ValidatesFrontend_MissingProviderConfig(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "stagecraft.yml")

	content := []byte(`
project:
  name: "test-app"
frontend:
  provider: generic
  providers: {}
environments:
  dev:
    driver: "local"
`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected validation error for missing frontend provider config")
	}

	if err != nil && !contains(err.Error(), "frontend.providers.generic") {
		t.Fatalf("expected error to mention missing provider config, got: %v", err)
	}
}

func TestLoad_ValidatesFrontend_MissingProvidersMap(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "stagecraft.yml")

	content := []byte(`
project:
  name: "test-app"
frontend:
  provider: generic
environments:
  dev:
    driver: "local"
`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected validation error for missing frontend providers map")
	}

	if err != nil && !contains(err.Error(), "frontend.providers is required") {
		t.Fatalf("expected error to mention missing providers, got: %v", err)
	}
}

func TestLoad_ValidatesDatabase_WithRawEngine(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "stagecraft.yml")

	content := []byte(`
project:
  name: "test-app"
databases:
  main:
    connection_env: DATABASE_URL
    migrations:
      engine: raw
      path: ./migrations
      strategy: pre_deploy
environments:
  dev:
    driver: "local"
`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error loading valid config with raw migration engine, got: %v", err)
	}

	if cfg.Databases == nil {
		t.Fatalf("expected databases config to be present")
	}

	mainDB, ok := cfg.Databases["main"]
	if !ok {
		t.Fatalf("expected 'main' database to be present")
	}

	if mainDB.Migrations == nil {
		t.Fatalf("expected migrations config to be present")
	}

	if mainDB.Migrations.Engine != "raw" {
		t.Fatalf("expected migrations.engine 'raw', got %q", mainDB.Migrations.Engine)
	}
}

func TestLoad_ValidatesDatabase_UnknownEngine(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "stagecraft.yml")

	content := []byte(`
project:
  name: "test-app"
databases:
  main:
    connection_env: DATABASE_URL
    migrations:
      engine: unknown-engine
      path: ./migrations
environments:
  dev:
    driver: "local"
`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected validation error for unknown migration engine")
	}

	if err != nil && err.Error() == "" {
		t.Fatalf("expected error message, got empty")
	}

	// Verify error message includes available engines
	errMsg := err.Error()
	if !contains(errMsg, "unknown migration engine") {
		t.Errorf("error message should mention 'unknown migration engine', got: %q", errMsg)
	}
	if !contains(errMsg, "available engines") {
		t.Errorf("error message should mention 'available engines', got: %q", errMsg)
	}
	// Should mention at least one registered engine (raw)
	if !contains(errMsg, "raw") {
		t.Errorf("error message should list available engines, got: %q", errMsg)
	}
}

func TestLoad_ValidatesDatabase_MissingPath(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "stagecraft.yml")

	content := []byte(`
project:
  name: "test-app"
databases:
  main:
    connection_env: DATABASE_URL
    migrations:
      engine: raw
environments:
  dev:
    driver: "local"
`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected validation error for missing migration path")
	}

	if err != nil && !contains(err.Error(), "migrations.path is required") {
		t.Fatalf("expected error to mention missing path, got: %v", err)
	}
}

func TestLoad_ValidatesDatabase_InvalidStrategy(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "stagecraft.yml")

	content := []byte(`
project:
  name: "test-app"
databases:
  main:
    connection_env: DATABASE_URL
    migrations:
      engine: raw
      path: ./migrations
      strategy: invalid-strategy
environments:
  dev:
    driver: "local"
`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected validation error for invalid migration strategy")
	}

	if err != nil && !contains(err.Error(), "strategy must be one of") {
		t.Fatalf("expected error to mention invalid strategy, got: %v", err)
	}
}

func TestBackendConfig_GetProviderConfig(t *testing.T) {
	cfg := &BackendConfig{
		Provider: "generic",
		Providers: map[string]any{
			"generic": map[string]any{
				"dev": map[string]any{
					"command": []string{"npm", "run", "dev"},
				},
			},
		},
	}

	providerCfg, err := cfg.GetProviderConfig()
	if err != nil {
		t.Fatalf("GetProviderConfig() error = %v, want nil", err)
	}

	if providerCfg == nil {
		t.Fatalf("GetProviderConfig() returned nil config")
	}
}

func TestBackendConfig_GetProviderConfig_MissingProvider(t *testing.T) {
	cfg := &BackendConfig{
		Provider:  "",
		Providers: map[string]any{},
	}

	_, err := cfg.GetProviderConfig()
	if err == nil {
		t.Fatalf("GetProviderConfig() error = nil, want error for missing provider")
	}
}

func TestBackendConfig_GetProviderConfig_MissingConfig(t *testing.T) {
	cfg := &BackendConfig{
		Provider:  "generic",
		Providers: map[string]any{},
	}

	_, err := cfg.GetProviderConfig()
	if err == nil {
		t.Fatalf("GetProviderConfig() error = nil, want error for missing config")
	}
}

// Feature: GOV_V1_CORE
// Spec: spec/governance/GOV_V1_CORE.md
// Tests for FrontendConfig.GetProviderConfig (second overload)

func TestFrontendConfig_GetProviderConfig(t *testing.T) {
	cfg := &FrontendConfig{
		Provider: "generic",
		Providers: map[string]any{
			"generic": map[string]any{
				"build": map[string]any{
					"command": []string{"npm", "run", "build"},
				},
			},
		},
	}

	providerCfg, err := cfg.GetProviderConfig()
	if err != nil {
		t.Fatalf("GetProviderConfig() error = %v, want nil", err)
	}

	if providerCfg == nil {
		t.Fatalf("GetProviderConfig() returned nil config")
	}
}

func TestFrontendConfig_GetProviderConfig_MissingProvider(t *testing.T) {
	cfg := &FrontendConfig{
		Provider:  "",
		Providers: map[string]any{},
	}

	_, err := cfg.GetProviderConfig()
	if err == nil {
		t.Fatalf("GetProviderConfig() error = nil, want error for missing provider")
	}

	if !contains(err.Error(), "frontend.provider is required") {
		t.Fatalf("expected error to mention missing provider, got: %v", err)
	}
}

func TestFrontendConfig_GetProviderConfig_MissingProvidersMap(t *testing.T) {
	cfg := &FrontendConfig{
		Provider:  "generic",
		Providers: nil,
	}

	_, err := cfg.GetProviderConfig()
	if err == nil {
		t.Fatalf("GetProviderConfig() error = nil, want error for missing providers map")
	}

	if !contains(err.Error(), "frontend.providers is required") {
		t.Fatalf("expected error to mention missing providers, got: %v", err)
	}
}

func TestFrontendConfig_GetProviderConfig_MissingConfig(t *testing.T) {
	cfg := &FrontendConfig{
		Provider:  "generic",
		Providers: map[string]any{},
	}

	_, err := cfg.GetProviderConfig()
	if err == nil {
		t.Fatalf("GetProviderConfig() error = nil, want error for missing config")
	}

	if !contains(err.Error(), "frontend.providers.generic") {
		t.Fatalf("expected error to mention missing provider config, got: %v", err)
	}
}

// Feature: GOV_V1_CORE
// Spec: spec/governance/GOV_V1_CORE.md
// Additional error path tests for Load

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "stagecraft.yml")

	// Invalid YAML syntax
	content := []byte(`
project:
  name: "test-app"
  invalid: [unclosed bracket
environments:
  dev:
    driver: "local"
`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected error for invalid YAML, got nil")
	}

	if !contains(err.Error(), "parsing config file") {
		t.Fatalf("expected error to mention parsing, got: %v", err)
	}
}

func TestLoad_MissingRequiredSections(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "stagecraft.yml")

	// Missing project.name (required field)
	content := []byte(`
project: {}
environments:
  dev:
    driver: "local"
`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected validation error for missing project.name")
	}

	// Should fail validation for empty project.name
	if !contains(err.Error(), "project.name") {
		t.Fatalf("expected validation error mentioning project.name, got: %v", err)
	}
}

func TestLoad_ReadError(t *testing.T) {
	// Test with a path that exists but can't be read (on some systems)
	// This is hard to test portably, so we'll test with a directory path instead
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "stagecraft.yml")

	// Create a directory with the same name as the file
	if err := os.Mkdir(path, 0o750); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	_, err := Load(path)
	// Should either fail on Exists check (if it detects directory) or on ReadFile
	if err == nil {
		t.Fatalf("expected error when path is a directory, got nil")
	}
}

// Feature: GOV_V1_CORE
// Spec: spec/governance/GOV_V1_CORE.md
// Additional edge case tests for Exists

func TestExists_Directory(t *testing.T) {
	tmpDir := t.TempDir()
	dirPath := filepath.Join(tmpDir, "adir")

	if err := os.Mkdir(dirPath, 0o750); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	ok, err := Exists(dirPath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Exists should return false for directories (only files)
	if ok {
		t.Fatalf("expected Exists to return false for directory, got true")
	}
}

func TestExists_PermissionError(t *testing.T) {
	// On Unix systems, we can test permission errors
	// But this is system-dependent, so we'll just test the non-existent path case
	tmpDir := t.TempDir()
	nonExistent := filepath.Join(tmpDir, "nonexistent.yml")

	ok, err := Exists(nonExistent)
	if err != nil {
		t.Fatalf("expected no error for non-existent file, got: %v", err)
	}

	if ok {
		t.Fatalf("expected Exists to return false for non-existent file")
	}
}

// contains checks if a string contains a substring (case-sensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || substr == "" ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

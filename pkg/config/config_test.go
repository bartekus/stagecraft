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
	if err := os.WriteFile(existing, []byte("project:\n  name: test\n"), 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
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

// contains checks if a string contains a substring (case-sensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
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

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

// TestRegistryIntegration_EndToEnd tests that registry-based validation works
// end-to-end with actual provider registration.
func TestRegistryIntegration_EndToEnd(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "stagecraft.yml")

	// Test config with both backend and migration providers
	content := []byte(`
project:
  name: "integration-test"
backend:
  provider: generic
  providers:
    generic:
      dev:
        command: ["npm", "run", "dev"]
        workdir: "./backend"
databases:
  main:
    connection_env: DATABASE_URL
    migrations:
      engine: raw
      path: ./migrations
      strategy: pre_deploy
environments:
  dev:
    driver: local
`)

	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	// Verify backend config
	if cfg.Backend == nil {
		t.Fatal("expected backend config to be present")
	}
	if cfg.Backend.Provider != "generic" {
		t.Errorf("Backend.Provider = %q, want %q", cfg.Backend.Provider, "generic")
	}

	// Verify database config
	if cfg.Databases == nil {
		t.Fatal("expected databases config to be present")
	}
	mainDB, ok := cfg.Databases["main"]
	if !ok {
		t.Fatal("expected 'main' database to be present")
	}
	if mainDB.Migrations == nil {
		t.Fatal("expected migrations config to be present")
	}
	if mainDB.Migrations.Engine != "raw" {
		t.Errorf("Migrations.Engine = %q, want %q", mainDB.Migrations.Engine, "raw")
	}

	// Verify GetProviderConfig works
	providerCfg, err := cfg.Backend.GetProviderConfig()
	if err != nil {
		t.Fatalf("GetProviderConfig() error = %v, want nil", err)
	}
	if providerCfg == nil {
		t.Fatal("GetProviderConfig() returned nil config")
	}
}

// TestRegistryIntegration_ErrorMessages tests that error messages are helpful
// and include available options from registries.
func TestRegistryIntegration_ErrorMessages(t *testing.T) {
	tests := []struct {
		name          string
		configContent string
		wantContains  []string // Error message must contain all of these
	}{
		{
			name: "unknown backend provider shows available options",
			configContent: `
project:
  name: test
backend:
  provider: invalid-provider
  providers:
    invalid-provider: {}
environments:
  dev:
    driver: local
`,
			wantContains: []string{
				"unknown backend provider",
				"available providers",
				// Should list actual registered providers (from registry, not hardcoded)
				"generic", // At least one registered provider should be listed
			},
		},
		{
			name: "unknown migration engine shows available options",
			configContent: `
project:
  name: test
databases:
  main:
    connection_env: DATABASE_URL
    migrations:
      engine: invalid-engine
      path: ./migrations
environments:
  dev:
    driver: local
`,
			wantContains: []string{
				"unknown migration engine",
				"available engines",
				// Should list actual registered engines (from registry, not hardcoded)
				"raw", // At least one registered engine should be listed
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "stagecraft.yml")

			if err := os.WriteFile(path, []byte(tt.configContent), 0o644); err != nil {
				t.Fatalf("failed to write temp config: %v", err)
			}

			_, err := Load(path)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}

			errMsg := err.Error()

			// Verify error contains required phrases
			for _, phrase := range tt.wantContains {
				if !contains(errMsg, phrase) {
					t.Errorf("error message should contain %q, got: %q", phrase, errMsg)
				}
			}

			// Verify error message comes from registry (shows actual registered providers/engines)
			// This proves we're using the registry, not hardcoded lists
			t.Logf("Error message: %s", errMsg)
		})
	}
}

// TestRegistryIntegration_ProviderRegistrationOrder verifies that providers
// are registered before config validation runs.
func TestRegistryIntegration_ProviderRegistrationOrder(t *testing.T) {
	// This test verifies that importing config package causes providers to register
	// The actual registration happens via init() functions in provider packages
	// which are imported in pkg/config/config.go

	// Just verify that Load() can validate against registered providers
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "stagecraft.yml")

	content := []byte(`
project:
  name: test
backend:
  provider: generic
  providers:
    generic:
      dev:
        command: ["echo", "test"]
environments:
  dev:
    driver: local
`)

	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	// If providers weren't registered, this would fail
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v (providers may not be registered)", err)
	}

	if cfg.Backend.Provider != "generic" {
		t.Errorf("Backend.Provider = %q, want %q", cfg.Backend.Provider, "generic")
	}
}

// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"os"
	"path/filepath"
	"testing"
)

// Feature: MIGRATION_CONFIG
// Spec: spec/migrations/config.md

func TestLoad_ValidatesMigrationsConfig_MinimalValid(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "stagecraft.yml")

	content := []byte(`
project:
  name: "test-app"
migrations:
  default_engine: "raw"
  sources:
    raw_sql_dir: "migrations/sql"
  selection:
    all: true
environments:
  dev:
    driver: "local"
`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error loading valid migrations config, got: %v", err)
	}

	if cfg.Migrations == nil {
		t.Fatalf("expected migrations config to be present")
	}
	if cfg.Migrations.DefaultEngine != "raw" {
		t.Fatalf("expected migrations.default_engine 'raw', got %q", cfg.Migrations.DefaultEngine)
	}
	if cfg.Migrations.Sources == nil || cfg.Migrations.Sources.RawSQLDir != "migrations/sql" {
		t.Fatalf("expected migrations.sources.raw_sql_dir 'migrations/sql', got %+v", cfg.Migrations.Sources)
	}
	if cfg.Migrations.Selection == nil || !cfg.Migrations.Selection.All {
		t.Fatalf("expected migrations.selection.all true, got %+v", cfg.Migrations.Selection)
	}
}

func TestLoad_ValidatesMigrationsConfig_SelectionAllCannotCombine(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "stagecraft.yml")

	content := []byte(`
project:
  name: "test-app"
migrations:
  default_engine: "raw"
  selection:
    all: true
    tags: ["schema"]
environments:
  dev:
    driver: "local"
`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}
	if !containsSubstring(err.Error(), "migrations.selection") {
		t.Fatalf("expected error to mention migrations.selection, got: %v", err)
	}
}

func TestLoad_NormalizesMigrationsConfig_SortsLists(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "stagecraft.yml")

	content := []byte(`
project:
  name: "test-app"
migrations:
  default_engine: "raw"
  sources:
    raw_sql_files: ["b.sql", "a.sql"]
  selection:
    all: false
    ids: ["m2", "m1"]
    tags: ["z", "a"]
environments:
  dev:
    driver: "local"
`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error loading valid config, got: %v", err)
	}

	if cfg.Migrations == nil || cfg.Migrations.Selection == nil || cfg.Migrations.Sources == nil {
		t.Fatalf("expected migrations config, selection, and sources to be present")
	}

	if cfg.Migrations.Sources.RawSQLFiles[0] != "a.sql" || cfg.Migrations.Sources.RawSQLFiles[1] != "b.sql" {
		t.Fatalf("expected raw_sql_files sorted, got: %+v", cfg.Migrations.Sources.RawSQLFiles)
	}
	if cfg.Migrations.Selection.IDs[0] != "m1" || cfg.Migrations.Selection.IDs[1] != "m2" {
		t.Fatalf("expected selection.ids sorted, got: %+v", cfg.Migrations.Selection.IDs)
	}
	if cfg.Migrations.Selection.Tags[0] != "a" || cfg.Migrations.Selection.Tags[1] != "z" {
		t.Fatalf("expected selection.tags sorted, got: %+v", cfg.Migrations.Selection.Tags)
	}
}

func TestLoad_ValidatesMigrationsConfig_RejectsDotDotPaths(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "stagecraft.yml")

	content := []byte(`
project:
  name: "test-app"
migrations:
  default_engine: "raw"
  sources:
    raw_sql_dir: "../migrations"
environments:
  dev:
    driver: "local"
`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}
	if !containsSubstring(err.Error(), "must not contain '..'") {
		t.Fatalf("expected error to mention '..' segments, got: %v", err)
	}
}

// containsSubstring checks if a string contains a substring (case-sensitive).
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || substr == "" ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsMiddleStr(s, substr))))
}

func containsMiddleStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

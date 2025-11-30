// pkg/config/config_test.go
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
	if err := os.WriteFile(existing, []byte("project:\n  name: test\n"), 0o644); err != nil {
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

	if err := os.WriteFile(path, content, 0o644); err != nil {
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

	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected validation error for empty project.name")
	}
}

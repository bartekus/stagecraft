package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Feature: CLI_MIGRATE_BASIC
// Spec: spec/commands/migrate-basic.md

func TestNewMigrateCommand_HasExpectedMetadata(t *testing.T) {
	cmd := NewMigrateCommand()

	if cmd.Use != "migrate" {
		t.Fatalf("expected Use to be 'migrate', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Fatalf("expected Short description to be non-empty")
	}
}

func TestMigrateCommand_ConfigNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	root := &cobra.Command{Use: "stagecraft"}
	root.AddCommand(NewMigrateCommand())

	_, err := executeCommandForGolden(root, "migrate")
	if err == nil {
		t.Fatalf("expected error when config file is missing")
	}

	if !strings.Contains(err.Error(), "stagecraft config not found") {
		t.Fatalf("expected config not found error, got: %v", err)
	}
}

func TestMigrateCommand_DatabaseNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")

	configContent := `project:
  name: test-app
databases:
  main:
    connection_env: DATABASE_URL
    migrations:
      engine: raw
      path: ./migrations
environments:
  dev:
    driver: local
`
	os.WriteFile(configPath, []byte(configContent), 0644)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	root := &cobra.Command{Use: "stagecraft"}
	root.AddCommand(NewMigrateCommand())

	_, err := executeCommandForGolden(root, "migrate", "--database", "nonexistent")
	if err == nil {
		t.Fatalf("expected error when database is not found")
	}

	if !strings.Contains(err.Error(), "database") || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected database not found error, got: %v", err)
	}
}

func TestMigrateCommand_NoMigrationsConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")

	configContent := `project:
  name: test-app
databases:
  main:
    connection_env: DATABASE_URL
environments:
  dev:
    driver: local
`
	os.WriteFile(configPath, []byte(configContent), 0644)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	root := &cobra.Command{Use: "stagecraft"}
	root.AddCommand(NewMigrateCommand())

	_, err := executeCommandForGolden(root, "migrate")
	if err == nil {
		t.Fatalf("expected error when migrations config is missing")
	}

	if !strings.Contains(err.Error(), "no migrations configured") {
		t.Fatalf("expected no migrations config error, got: %v", err)
	}
}

func TestMigrateCommand_UnknownEngine(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")

	configContent := `project:
  name: test-app
databases:
  main:
    connection_env: DATABASE_URL
    migrations:
      engine: unknown-engine
      path: ./migrations
environments:
  dev:
    driver: local
`
	os.WriteFile(configPath, []byte(configContent), 0644)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	root := &cobra.Command{Use: "stagecraft"}
	root.AddCommand(NewMigrateCommand())

	_, err := executeCommandForGolden(root, "migrate")
	if err == nil {
		t.Fatalf("expected error for unknown engine")
	}

	if !strings.Contains(err.Error(), "unknown migration engine") {
		t.Fatalf("expected unknown engine error, got: %v", err)
	}
}

func TestMigrateCommand_Help(t *testing.T) {
	root := &cobra.Command{Use: "stagecraft"}
	root.AddCommand(NewMigrateCommand())

	out, err := executeCommandForGolden(root, "migrate", "--help")
	if err != nil {
		t.Fatalf("help command should not error, got: %v", err)
	}

	if !strings.Contains(out, "Loads stagecraft.yml") && !strings.Contains(out, "migrate") {
		t.Fatalf("expected help text, got: %q", out)
	}
}


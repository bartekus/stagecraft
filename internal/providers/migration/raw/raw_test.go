// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package raw

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"stagecraft/pkg/providers/migration"
)

// Feature: MIGRATION_ENGINE_RAW

func TestRawEngine_ID(t *testing.T) {
	e := &Engine{}
	if got := e.ID(); got != "raw" {
		t.Errorf("ID() = %q, want %q", got, "raw")
	}
}

func TestRawEngine_Plan(t *testing.T) {
	e := &Engine{}
	tmpDir := t.TempDir()

	// Create some SQL migration files
	migrationFiles := []string{
		"001_initial.sql",
		"002_add_users.sql",
		"003_add_posts.sql",
	}

	for _, name := range migrationFiles {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte("-- migration: "+name), 0o600); err != nil {
			t.Fatalf("failed to create migration file: %v", err)
		}
	}

	opts := migration.PlanOptions{
		MigrationPath: tmpDir,
		ConnectionEnv: "DATABASE_URL",
		WorkDir:       ".",
	}

	migrations, err := e.Plan(context.Background(), opts)
	if err != nil {
		t.Fatalf("Plan() error = %v, want nil", err)
	}

	if len(migrations) != 3 {
		t.Errorf("Plan() returned %d migrations, want 3", len(migrations))
	}

	// Verify migrations are sorted
	if migrations[0].ID != "001_initial.sql" {
		t.Errorf("First migration ID = %q, want %q", migrations[0].ID, "001_initial.sql")
	}
}

func TestRawEngine_Plan_EmptyDirectory(t *testing.T) {
	e := &Engine{}
	tmpDir := t.TempDir()

	opts := migration.PlanOptions{
		MigrationPath: tmpDir,
		ConnectionEnv: "DATABASE_URL",
		WorkDir:       ".",
	}

	migrations, err := e.Plan(context.Background(), opts)
	if err != nil {
		t.Fatalf("Plan() error = %v, want nil", err)
	}

	if len(migrations) != 0 {
		t.Errorf("Plan() returned %d migrations for empty directory, want 0", len(migrations))
	}
}

func TestRawEngine_Plan_NonExistentDirectory(t *testing.T) {
	e := &Engine{}

	opts := migration.PlanOptions{
		MigrationPath: "/nonexistent/path",
		ConnectionEnv: "DATABASE_URL",
		WorkDir:       ".",
	}

	_, err := e.Plan(context.Background(), opts)
	if err == nil {
		t.Error("Plan() error = nil, want error for non-existent directory")
	}
}

func TestRawEngine_Plan_IgnoresNonSQLFiles(t *testing.T) {
	e := &Engine{}
	tmpDir := t.TempDir()

	// Create SQL and non-SQL files
	files := map[string]string{
		"001_initial.sql":   "-- SQL migration",
		"002_add_users.sql": "-- SQL migration",
		"README.md":         "# Documentation",
		"config.json":       "{}",
	}

	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
	}

	opts := migration.PlanOptions{
		MigrationPath: tmpDir,
		ConnectionEnv: "DATABASE_URL",
		WorkDir:       ".",
	}

	migrations, err := e.Plan(context.Background(), opts)
	if err != nil {
		t.Fatalf("Plan() error = %v, want nil", err)
	}

	// Should only return SQL files
	if len(migrations) != 2 {
		t.Errorf("Plan() returned %d migrations, want 2 (only SQL files)", len(migrations))
	}
}

func TestRawEngine_Run_MissingMigrationPath(t *testing.T) {
	t.Parallel()

	e := &Engine{}

	opts := migration.RunOptions{
		MigrationPath: "",
		ConnectionEnv: "DATABASE_URL",
		WorkDir:       ".",
		Direction:     "up",
		Steps:         0,
	}

	err := e.Run(context.Background(), opts)
	if err == nil {
		t.Error("Run() error = nil, want error for missing migration path")
	}

	if err != nil && !strings.Contains(err.Error(), "migration path is required") {
		t.Errorf("expected error to mention 'migration path is required', got: %v", err)
	}
}

func TestRawEngine_Run_EmptyMigrations(t *testing.T) {
	t.Parallel()

	e := &Engine{}
	tmpDir := t.TempDir()

	// Note: This test can't fully exercise the "no migrations" path without a real DB
	// because Run() connects to the DB before checking for migrations.
	// We test the validation that happens before DB connection instead.
	// The actual "no migrations" check happens after DB connection, which requires
	// a real database or more sophisticated mocking than we want for Phase 2.

	opts := migration.RunOptions{
		MigrationPath: tmpDir,
		ConnectionEnv: "DATABASE_URL",
		WorkDir:       ".",
		Direction:     "up",
		Steps:         0,
	}

	// Ensure env var is not set to test the env validation path
	originalEnv := os.Getenv(opts.ConnectionEnv)
	defer func() {
		if originalEnv != "" {
			_ = os.Setenv(opts.ConnectionEnv, originalEnv)
		} else {
			_ = os.Unsetenv(opts.ConnectionEnv)
		}
	}()
	_ = os.Unsetenv(opts.ConnectionEnv)

	err := e.Run(context.Background(), opts)
	if err == nil {
		t.Error("Run() error = nil, want error")
	}

	// Should fail at connection env check (before checking migrations)
	if err != nil && !strings.Contains(err.Error(), "is not set") {
		t.Errorf("expected error about connection env, got: %v", err)
	}
}

func TestRawEngine_Run_MissingConnectionEnv(t *testing.T) {
	t.Parallel()

	e := &Engine{}
	tmpDir := t.TempDir()

	// Create a SQL file
	sqlFile := filepath.Join(tmpDir, "001_initial.sql")
	if err := os.WriteFile(sqlFile, []byte("CREATE TABLE test (id INT);"), 0o600); err != nil {
		t.Fatalf("failed to create migration file: %v", err)
	}

	// Ensure env var is not set
	originalEnv := os.Getenv("DATABASE_URL")
	defer func() {
		if originalEnv != "" {
			_ = os.Setenv("DATABASE_URL", originalEnv)
		} else {
			_ = os.Unsetenv("DATABASE_URL")
		}
	}()
	_ = os.Unsetenv("DATABASE_URL")

	opts := migration.RunOptions{
		MigrationPath: tmpDir,
		ConnectionEnv: "DATABASE_URL",
		WorkDir:       ".",
		Direction:     "up",
		Steps:         0,
	}

	err := e.Run(context.Background(), opts)
	if err == nil {
		t.Error("Run() error = nil, want error for missing connection env")
	}

	if err != nil && !strings.Contains(err.Error(), "is not set") {
		t.Errorf("expected error to mention connection env not set, got: %v", err)
	}
}

func TestRawEngine_Plan_MissingMigrationPath(t *testing.T) {
	t.Parallel()

	e := &Engine{}

	opts := migration.PlanOptions{
		MigrationPath: "",
		ConnectionEnv: "DATABASE_URL",
		WorkDir:       ".",
	}

	_, err := e.Plan(context.Background(), opts)
	if err == nil {
		t.Error("Plan() error = nil, want error for missing migration path")
	}

	if err != nil && !strings.Contains(err.Error(), "migration path is required") {
		t.Errorf("expected error to mention 'migration path is required', got: %v", err)
	}
}

func TestRawEngine_Plan_Sorting(t *testing.T) {
	t.Parallel()

	e := &Engine{}
	tmpDir := t.TempDir()

	// Create migrations in non-lexicographic order
	migrationFiles := []string{
		"003_third.sql",
		"001_first.sql",
		"002_second.sql",
	}

	for _, name := range migrationFiles {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte("-- migration: "+name), 0o600); err != nil {
			t.Fatalf("failed to create migration file: %v", err)
		}
	}

	opts := migration.PlanOptions{
		MigrationPath: tmpDir,
		ConnectionEnv: "DATABASE_URL",
		WorkDir:       ".",
	}

	migrations, err := e.Plan(context.Background(), opts)
	if err != nil {
		t.Fatalf("Plan() error = %v, want nil", err)
	}

	if len(migrations) != 3 {
		t.Fatalf("Plan() returned %d migrations, want 3", len(migrations))
	}

	// Verify migrations are sorted lexicographically
	expectedOrder := []string{"001_first.sql", "002_second.sql", "003_third.sql"}
	for i, expected := range expectedOrder {
		if migrations[i].ID != expected {
			t.Errorf("migrations[%d].ID = %q, want %q", i, migrations[i].ID, expected)
		}
	}
}

func TestRawEngine_Plan_IgnoresDirectories(t *testing.T) {
	t.Parallel()

	e := &Engine{}
	tmpDir := t.TempDir()

	// Create a subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0o750); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	// Create SQL file in subdirectory (should be ignored)
	sqlFile := filepath.Join(subDir, "001_ignored.sql")
	if err := os.WriteFile(sqlFile, []byte("-- ignored"), 0o600); err != nil {
		t.Fatalf("failed to create SQL file in subdir: %v", err)
	}

	// Create SQL file in root (should be included)
	rootFile := filepath.Join(tmpDir, "001_included.sql")
	if err := os.WriteFile(rootFile, []byte("-- included"), 0o600); err != nil {
		t.Fatalf("failed to create SQL file in root: %v", err)
	}

	opts := migration.PlanOptions{
		MigrationPath: tmpDir,
		ConnectionEnv: "DATABASE_URL",
		WorkDir:       ".",
	}

	migrations, err := e.Plan(context.Background(), opts)
	if err != nil {
		t.Fatalf("Plan() error = %v, want nil", err)
	}

	// Should only include the root file, not the subdirectory file
	if len(migrations) != 1 {
		t.Errorf("Plan() returned %d migrations, want 1", len(migrations))
	}

	if len(migrations) > 0 && migrations[0].ID != "001_included.sql" {
		t.Errorf("Plan() returned migration %q, want %q", migrations[0].ID, "001_included.sql")
	}
}

func TestRawEngine_Run_NonExistentDirectory(t *testing.T) {
	e := &Engine{}

	opts := migration.RunOptions{
		MigrationPath: "/nonexistent/path",
		ConnectionEnv: "DATABASE_URL",
		WorkDir:       ".",
		Direction:     "up",
		Steps:         0,
	}

	err := e.Run(context.Background(), opts)
	if err == nil {
		t.Error("Run() error = nil, want error for non-existent directory")
	}
}

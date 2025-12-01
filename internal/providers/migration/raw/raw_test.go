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
	"testing"

	"stagecraft/pkg/providers/migration"
)

// Feature: MIGRATION_ENGINE_RAW
// Spec: spec/providers/migration/raw.md

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
		if err := os.WriteFile(path, []byte("-- migration: "+name), 0644); err != nil {
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
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
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

func TestRawEngine_Run_NotImplemented(t *testing.T) {
	e := &Engine{}
	tmpDir := t.TempDir()

	// Create a SQL file
	sqlFile := filepath.Join(tmpDir, "001_initial.sql")
	if err := os.WriteFile(sqlFile, []byte("CREATE TABLE test (id INT);"), 0644); err != nil {
		t.Fatalf("failed to create migration file: %v", err)
	}

	opts := migration.RunOptions{
		MigrationPath: tmpDir,
		ConnectionEnv: "DATABASE_URL",
		WorkDir:       ".",
		Direction:     "up",
		Steps:         0,
	}

	err := e.Run(context.Background(), opts)
	if err == nil {
		t.Error("Run() error = nil, want error (not yet implemented)")
	}

	// Should mention that execution is not yet implemented
	if err != nil && err.Error() == "" {
		t.Error("expected error message, got empty")
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

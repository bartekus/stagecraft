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
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"

	"stagecraft/pkg/providers/migration"
)

// Feature: MIGRATION_ENGINE_RAW
// Spec: spec/providers/migration/raw.md

// RawEngine implements a simple SQL file-based migration engine.
type RawEngine struct{}

// Ensure RawEngine implements Engine
var _ migration.Engine = (*RawEngine)(nil)

// ID returns the engine identifier.
func (e *RawEngine) ID() string {
	return "raw"
}

// Config represents the raw engine configuration.
type Config struct {
	// Additional engine-specific config can be added here
	// For now, raw engine uses the standard migration path
}

// Plan analyzes migration files and returns a list of pending migrations.
func (e *RawEngine) Plan(ctx context.Context, opts migration.PlanOptions) ([]migration.Migration, error) {
	// For raw engine, we simply list all SQL files in the migration directory
	// In a real implementation, we'd check which ones have been applied

	migrationPath := opts.MigrationPath
	if migrationPath == "" {
		return nil, fmt.Errorf("migration path is required")
	}

	// Read directory
	entries, err := os.ReadDir(migrationPath)
	if err != nil {
		return nil, fmt.Errorf("reading migration directory: %w", err)
	}

	var migrations []migration.Migration

	// Collect SQL files
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(strings.ToLower(entry.Name()), ".sql") {
			continue
		}

		migrations = append(migrations, migration.Migration{
			ID:          entry.Name(),
			Description: fmt.Sprintf("SQL migration: %s", entry.Name()),
			Applied:     false, // Raw engine doesn't track state in v1
		})
	}

	// Sort by filename (lexicographic)
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].ID < migrations[j].ID
	})

	return migrations, nil
}

// Run executes migrations.
func (e *RawEngine) Run(ctx context.Context, opts migration.RunOptions) error {
	migrationPath := opts.MigrationPath
	if migrationPath == "" {
		return fmt.Errorf("migration path is required")
	}

	// Verify migration directory exists
	if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
		return fmt.Errorf("migration directory does not exist: %s", migrationPath)
	}

	// Get connection string from environment
	dbURL := os.Getenv(opts.ConnectionEnv)
	if dbURL == "" {
		return fmt.Errorf("connection environment variable %q is not set", opts.ConnectionEnv)
	}

	// Parse database URL and connect
	// For v1, assume PostgreSQL (pgx driver)
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Verify connection
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("pinging database: %w", err)
	}

	// Get migration files (reuse Plan logic)
	planOpts := migration.PlanOptions{
		MigrationPath: migrationPath,
	}
	migrations, err := e.Plan(ctx, planOpts)
	if err != nil {
		return fmt.Errorf("planning migrations: %w", err)
	}

	if len(migrations) == 0 {
		return fmt.Errorf("no SQL migration files found in %s", migrationPath)
	}

	// Create migrations table if it doesn't exist
	if err := e.ensureMigrationsTable(ctx, db); err != nil {
		return fmt.Errorf("ensuring migrations table: %w", err)
	}

	// Execute each migration
	for _, m := range migrations {
		// Check if already applied
		applied, err := e.isApplied(ctx, db, m.ID)
		if err != nil {
			return fmt.Errorf("checking migration status: %w", err)
		}
		if applied {
			fmt.Printf("Skipping already applied migration: %s\n", m.ID)
			continue
		}

		// Read and execute SQL file
		sqlPath := filepath.Join(migrationPath, m.ID)
		sqlContent, err := os.ReadFile(sqlPath)
		if err != nil {
			return fmt.Errorf("reading migration file %s: %w", m.ID, err)
		}

		// Execute in transaction
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("starting transaction: %w", err)
		}

		if _, err := tx.ExecContext(ctx, string(sqlContent)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("executing migration %s: %w", m.ID, err)
		}

		// Record migration
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO stagecraft_migrations (id, applied_at) VALUES ($1, NOW())",
			m.ID,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("recording migration %s: %w", m.ID, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("committing migration %s: %w", m.ID, err)
		}

		fmt.Printf("Applied migration: %s\n", m.ID)
	}

	return nil
}

// ensureMigrationsTable creates the migrations tracking table if it doesn't exist.
func (e *RawEngine) ensureMigrationsTable(ctx context.Context, db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS stagecraft_migrations (
			id VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`
	_, err := db.ExecContext(ctx, query)
	return err
}

// isApplied checks if a migration has already been applied.
func (e *RawEngine) isApplied(ctx context.Context, db *sql.DB, id string) (bool, error) {
	var count int
	err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM stagecraft_migrations WHERE id = $1",
		id,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func init() {
	migration.Register(&RawEngine{})
}

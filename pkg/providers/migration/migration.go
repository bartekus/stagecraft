// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - A Go-based CLI for orchestrating local-first multi-service deployments using Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package migration

import "context"

// Feature: MIGRATION_INTERFACE
// Spec: spec/core/migration-registry.md

// Migration represents a single migration step.
type Migration struct {
	ID          string
	Description string
	Applied     bool
	// Additional fields as needed by specific engines
}

// PlanOptions contains options for planning migrations.
type PlanOptions struct {
	// Engine-specific configuration decoded from
	// databases[dbName].migrations in stagecraft.yml
	Config any

	// MigrationPath is the path to migration files
	MigrationPath string

	// ConnectionEnv is the environment variable name for DB connection
	ConnectionEnv string

	// WorkDir is the working directory
	WorkDir string
}

// RunOptions contains options for running migrations.
type RunOptions struct {
	// Config is the engine-specific configuration
	Config any

	// MigrationPath is the path to migration files
	MigrationPath string

	// ConnectionEnv is the environment variable name for DB connection
	ConnectionEnv string

	// WorkDir is the working directory
	WorkDir string

	// Direction specifies migration direction (up, down, etc.)
	Direction string

	// Steps limits the number of migrations to run (0 = all)
	Steps int
}

// Engine is the interface that all migration engines must implement.
type Engine interface {
	// ID returns the unique identifier for this engine (e.g., "drizzle", "prisma", "knex", "raw").
	ID() string

	// Plan analyzes migration files and returns a list of pending migrations.
	Plan(ctx context.Context, opts PlanOptions) ([]Migration, error)

	// Run executes migrations.
	Run(ctx context.Context, opts RunOptions) error
}


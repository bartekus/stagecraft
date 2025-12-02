// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"stagecraft/pkg/config"
	"stagecraft/pkg/logging"
	migrationengines "stagecraft/pkg/providers/migration"
)

// Feature: CLI_MIGRATE_BASIC
// Spec: spec/commands/migrate-basic.md

// NewMigrateCommand returns the `stagecraft migrate` command.
func NewMigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Long:  "Loads stagecraft.yml, resolves migration engine, and runs migrations",
		RunE:  runMigrate,
	}

	// Global flags (--config, --env, --verbose, --dry-run) are inherited from root
	cmd.Flags().String("database", "main", "Database name to migrate")
	cmd.Flags().Bool("plan", false, "Show migration plan without executing")

	return cmd
}

func runMigrate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Resolve global flags
	flags, err := ResolveFlags(cmd, nil)
	if err != nil {
		return fmt.Errorf("resolving flags: %w", err)
	}

	// Load config to validate environment if needed
	cfg, err := config.Load(flags.Config)
	if err != nil {
		if err == config.ErrConfigNotFound {
			return fmt.Errorf("stagecraft config not found at %s", flags.Config)
		}
		return fmt.Errorf("loading config: %w", err)
	}

	// Re-resolve flags with config for environment validation
	flags, err = ResolveFlags(cmd, cfg)
	if err != nil {
		return fmt.Errorf("resolving flags: %w", err)
	}

	absPath, err := filepath.Abs(flags.Config)
	if err != nil {
		return fmt.Errorf("resolving config path: %w", err)
	}

	dbName, _ := cmd.Flags().GetString("database")
	dbCfg, ok := cfg.Databases[dbName]
	if !ok {
		return fmt.Errorf("database %q not found in config; available: %v",
			dbName, getDatabaseNames(cfg))
	}

	if dbCfg.Migrations == nil {
		return fmt.Errorf("database %q has no migrations configured", dbName)
	}

	engineID := dbCfg.Migrations.Engine
	engine, err := migrationengines.Get(engineID)
	if err != nil {
		// Enhance error message with available engines
		available := migrationengines.DefaultRegistry.IDs()
		return fmt.Errorf("unknown migration engine %q for database %s; available engines: %v",
			engineID, dbName, available)
	}

	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	// Resolve migration path (relative to workDir)
	migrationPath := dbCfg.Migrations.Path
	if !filepath.IsAbs(migrationPath) {
		migrationPath = filepath.Join(workDir, migrationPath)
	}

	// Initialize logger
	logger := logging.NewLogger(flags.Verbose)
	logger.Info("Running migrations",
		logging.NewField("engine", engineID),
		logging.NewField("database", dbName),
		logging.NewField("env", flags.Env),
	)
	logger.Debug("Migration details",
		logging.NewField("config", absPath),
		logging.NewField("path", migrationPath),
	)

	planOnly, _ := cmd.Flags().GetBool("plan")

	// Check for dry-run mode (treat as plan if not already planning)
	if flags.DryRun && !planOnly {
		logger.Info("Dry-run mode: would run migrations")
		planOnly = true
	}

	if planOnly {
		opts := migrationengines.PlanOptions{
			Config:        dbCfg.Migrations, // Pass the whole migration config
			MigrationPath: migrationPath,
			ConnectionEnv: dbCfg.ConnectionEnv,
			WorkDir:       workDir,
		}

		migrations, err := engine.Plan(ctx, opts)
		if err != nil {
			return fmt.Errorf("planning migrations: %w", err)
		}

		out := cmd.OutOrStdout()
		_, _ = fmt.Fprintf(out, "Migration plan (%d pending):\n", len(migrations))
		for _, m := range migrations {
			status := "pending"
			if m.Applied {
				status = "applied"
			}
			_, _ = fmt.Fprintf(out, "  - %s: %s [%s]\n", m.ID, m.Description, status)
		}

		return nil
	}

	opts := migrationengines.RunOptions{
		Config:        dbCfg.Migrations,
		MigrationPath: migrationPath,
		ConnectionEnv: dbCfg.ConnectionEnv,
		WorkDir:       workDir,
		Direction:     "up",
		Steps:         0, // All
	}

	return engine.Run(ctx, opts)
}

func getDatabaseNames(cfg *config.Config) []string {
	names := make([]string, 0, len(cfg.Databases))
	for name := range cfg.Databases {
		names = append(names, name)
	}
	return names
}

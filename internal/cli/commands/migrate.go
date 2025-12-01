// internal/cli/commands/migrate.go
package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"stagecraft/pkg/config"
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

	cmd.Flags().String("config", "", "path to Stagecraft config file (default: stagecraft.yml)")
	cmd.Flags().String("database", "main", "Database name to migrate")
	cmd.Flags().Bool("plan", false, "Show migration plan without executing")

	return cmd
}

func runMigrate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Get config path from flag or use default
	configPath, _ := cmd.Flags().GetString("config")
	if configPath == "" {
		configPath = config.DefaultConfigPath()
	}
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return fmt.Errorf("resolving config path: %w", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		if err == config.ErrConfigNotFound {
			return fmt.Errorf("stagecraft config not found at %s", configPath)
		}
		return fmt.Errorf("loading config: %w", err)
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

	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		fmt.Fprintf(cmd.OutOrStdout(), "Using migration engine: %s\n", engineID)
		fmt.Fprintf(cmd.OutOrStdout(), "Config file: %s\n", absPath)
		fmt.Fprintf(cmd.OutOrStdout(), "Database: %s\n", dbName)
		fmt.Fprintf(cmd.OutOrStdout(), "Migration path: %s\n", migrationPath)
	}

	planOnly, _ := cmd.Flags().GetBool("plan")

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

		fmt.Fprintf(cmd.OutOrStdout(), "Migration plan (%d pending):\n", len(migrations))
		for _, m := range migrations {
			status := "pending"
			if m.Applied {
				status = "applied"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  - %s: %s [%s]\n", m.ID, m.Description, status)
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


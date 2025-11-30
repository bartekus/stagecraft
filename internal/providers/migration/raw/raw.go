// internal/providers/migration/raw/raw.go
package raw

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

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
	// For v1, raw engine is a placeholder
	// Full implementation would:
	// 1. Parse SQL files
	// 2. Execute them against the database
	// 3. Track applied migrations
	
	migrationPath := opts.MigrationPath
	if migrationPath == "" {
		return fmt.Errorf("migration path is required")
	}

	// Verify migration directory exists
	if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
		return fmt.Errorf("migration directory does not exist: %s", migrationPath)
	}

	// For now, just verify we can read the directory
	entries, err := os.ReadDir(migrationPath)
	if err != nil {
		return fmt.Errorf("reading migration directory: %w", err)
	}

	// Count SQL files
	sqlCount := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".sql") {
			sqlCount++
		}
	}

	if sqlCount == 0 {
		return fmt.Errorf("no SQL migration files found in %s", migrationPath)
	}

	// In a full implementation, we would:
	// - Connect to database using opts.ConnectionEnv
	// - Execute SQL files in order
	// - Track which migrations have been applied
	
	return fmt.Errorf("raw migration engine execution not yet implemented (found %d SQL files)", sqlCount)
}

// parseConfig unmarshals the engine config.
func (e *RawEngine) parseConfig(cfg any) (*Config, error) {
	if cfg == nil {
		return &Config{}, nil // Config is optional for raw engine
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshaling config: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("invalid raw engine config: %w", err)
	}

	return &config, nil
}

func init() {
	migration.Register(&RawEngine{})
}


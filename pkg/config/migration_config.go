// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// Feature: MIGRATION_CONFIG
// Spec: spec/migrations/config.md

// MigrationsRootConfig describes top-level migration configuration.
type MigrationsRootConfig struct {
	Enabled       *bool                                  `yaml:"enabled,omitempty"`
	DefaultEngine string                                 `yaml:"default_engine,omitempty"`
	Sources       *MigrationSourcesConfig                `yaml:"sources,omitempty"`
	Selection     *MigrationSelectionConfig              `yaml:"selection,omitempty"`
	EngineConfig  map[string]map[string]any              `yaml:"engine_config,omitempty"`
	Env           map[string]MigrationsEnvOverrideConfig `yaml:"env,omitempty"`
}

// MigrationSourcesConfig describes migration sources.
type MigrationSourcesConfig struct {
	RawSQLDir   string   `yaml:"raw_sql_dir,omitempty"`
	RawSQLFiles []string `yaml:"raw_sql_files,omitempty"`
}

// MigrationSelectionConfig describes migration selection.
type MigrationSelectionConfig struct {
	All  bool     `yaml:"all,omitempty"`
	IDs  []string `yaml:"ids,omitempty"`
	Tags []string `yaml:"tags,omitempty"`
}

// MigrationsEnvOverrideConfig describes per-environment overrides for migrations config.
// Any field that is set replaces the corresponding global field; unset fields inherit.
type MigrationsEnvOverrideConfig struct {
	Enabled       *bool                     `yaml:"enabled,omitempty"`
	DefaultEngine *string                   `yaml:"default_engine,omitempty"`
	Sources       *MigrationSourcesConfig   `yaml:"sources,omitempty"`
	Selection     *MigrationSelectionConfig `yaml:"selection,omitempty"`
	EngineConfig  map[string]map[string]any `yaml:"engine_config,omitempty"`
}

func validateMigrations(cfg *MigrationsRootConfig) error {
	enabled := true
	if cfg.Enabled != nil {
		enabled = *cfg.Enabled
	}

	if !enabled {
		return nil
	}

	if cfg.DefaultEngine == "" {
		return errors.New("migrations.default_engine is required when migrations.enabled is true")
	}

	if cfg.Selection != nil {
		if err := normalizeAndValidateSelection("migrations.selection", cfg.Selection); err != nil {
			return err
		}
	}

	if cfg.Sources != nil {
		if err := normalizeAndValidateSources("migrations.sources", cfg.Sources); err != nil {
			return err
		}
	}

	for envName, ov := range cfg.Env {
		if envName == "" {
			return errors.New("migrations.env: environment name must be non-empty")
		}
		if ov.DefaultEngine != nil && *ov.DefaultEngine == "" {
			return fmt.Errorf("migrations.env.%s.default_engine must be non-empty when set", envName)
		}
		if ov.Selection != nil {
			if err := normalizeAndValidateSelection(fmt.Sprintf("migrations.env.%s.selection", envName), ov.Selection); err != nil {
				return err
			}
		}
		if ov.Sources != nil {
			if err := normalizeAndValidateSources(fmt.Sprintf("migrations.env.%s.sources", envName), ov.Sources); err != nil {
				return err
			}
		}
	}

	return nil
}

func normalizeAndValidateSelection(prefix string, sel *MigrationSelectionConfig) error {
	for i := range sel.IDs {
		sel.IDs[i] = strings.TrimSpace(sel.IDs[i])
	}
	for i := range sel.Tags {
		sel.Tags[i] = strings.TrimSpace(sel.Tags[i])
	}

	sort.Strings(sel.IDs)
	sort.Strings(sel.Tags)

	if sel.All && (len(sel.IDs) > 0 || len(sel.Tags) > 0) {
		return fmt.Errorf("%s: all=true cannot be combined with ids or tags", prefix)
	}

	if hasDuplicates(sel.IDs) {
		return fmt.Errorf("%s.ids must not contain duplicates", prefix)
	}
	if hasDuplicates(sel.Tags) {
		return fmt.Errorf("%s.tags must not contain duplicates", prefix)
	}

	return nil
}

func normalizeAndValidateSources(prefix string, src *MigrationSourcesConfig) error {
	src.RawSQLDir = normalizeRelPath(src.RawSQLDir)
	for i := range src.RawSQLFiles {
		src.RawSQLFiles[i] = normalizeRelPath(src.RawSQLFiles[i])
	}
	sort.Strings(src.RawSQLFiles)

	if src.RawSQLDir != "" {
		if err := validateRelPath(prefix+".raw_sql_dir", src.RawSQLDir); err != nil {
			return err
		}
	}
	for _, p := range src.RawSQLFiles {
		if err := validateRelPath(prefix+".raw_sql_files", p); err != nil {
			return err
		}
	}
	if hasDuplicates(src.RawSQLFiles) {
		return fmt.Errorf("%s.raw_sql_files must not contain duplicates", prefix)
	}
	return nil
}

func normalizeRelPath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	p = strings.ReplaceAll(p, "\\", "/")
	return p
}

func validateRelPath(prefix, p string) error {
	if p == "" {
		return fmt.Errorf("%s must be non-empty", prefix)
	}
	if filepath.IsAbs(p) || strings.HasPrefix(p, "~") {
		return fmt.Errorf("%s must be a relative path", prefix)
	}
	for _, seg := range strings.Split(p, "/") {
		if seg == ".." {
			return fmt.Errorf("%s must not contain '..' segments", prefix)
		}
	}
	return nil
}

func hasDuplicates(sorted []string) bool {
	for i := 1; i < len(sorted); i++ {
		if sorted[i] == sorted[i-1] {
			return true
		}
	}
	return false
}

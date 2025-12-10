// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package config defines the Stagecraft configuration schema and helpers for loading and validating config files.
package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	// Import providers to ensure they register themselves
	_ "stagecraft/internal/providers/backend/encorets"
	_ "stagecraft/internal/providers/backend/generic"
	_ "stagecraft/internal/providers/frontend/generic"
	_ "stagecraft/internal/providers/migration/raw"

	backendproviders "stagecraft/pkg/providers/backend"
	frontendproviders "stagecraft/pkg/providers/frontend"
	migrationengines "stagecraft/pkg/providers/migration"
)

// Feature: CORE_BACKEND_PROVIDER_CONFIG_SCHEMA
// Spec: spec/core/backend-provider-config.md

// Feature: CORE_CONFIG
// Spec: spec/core/config.md

// ErrConfigNotFound is returned when the config file does not exist at the given path.
var ErrConfigNotFound = errors.New("stagecraft config not found")

// Config represents the top-level Stagecraft configuration.
type Config struct {
	Project      ProjectConfig                `yaml:"project"`
	Backend      *BackendConfig               `yaml:"backend,omitempty"`
	Frontend     *FrontendConfig              `yaml:"frontend,omitempty"`
	Dev          *DevConfig                   `yaml:"dev,omitempty"`
	Databases    map[string]DatabaseConfig    `yaml:"databases,omitempty"`
	Environments map[string]EnvironmentConfig `yaml:"environments"`
}

// ProjectConfig describes project-level settings.
type ProjectConfig struct {
	Name string `yaml:"name"`
}

// BackendConfig describes backend provider configuration.
type BackendConfig struct {
	Provider  string         `yaml:"provider"`
	Providers map[string]any `yaml:"providers"`
}

// FrontendConfig describes frontend provider configuration.
type FrontendConfig struct {
	Provider  string         `yaml:"provider"`
	Providers map[string]any `yaml:"providers"`
}

// DevConfig describes development environment configuration.
// Feature: CLI_DEV
// Spec: spec/commands/dev.md
type DevConfig struct {
	Domains *DevDomains `yaml:"domains,omitempty"`
}

// DevDomains describes development domain configuration.
// Feature: CLI_DEV
// Spec: spec/commands/dev.md
type DevDomains struct {
	Frontend string `yaml:"frontend,omitempty"`
	Backend  string `yaml:"backend,omitempty"`
}

// DatabaseConfig describes database configuration including migrations.
type DatabaseConfig struct {
	Migrations    *MigrationConfig `yaml:"migrations,omitempty"`
	ConnectionEnv string           `yaml:"connection_env"`
}

// MigrationConfig describes migration engine configuration.
type MigrationConfig struct {
	Engine   string `yaml:"engine"`
	Path     string `yaml:"path"`
	Strategy string `yaml:"strategy"` // pre_deploy, post_deploy, manual
}

// EnvironmentConfig describes per-environment settings.
type EnvironmentConfig struct {
	Driver  string `yaml:"driver"`
	EnvFile string `yaml:"env_file,omitempty"` // Path to environment file
	// Future: region, registry, etc.
}

// GetProviderConfig returns the config for the selected backend provider.
func (c *BackendConfig) GetProviderConfig() (any, error) {
	if c.Provider == "" {
		return nil, fmt.Errorf("backend.provider is required")
	}

	if c.Providers == nil {
		return nil, fmt.Errorf("backend.providers is required")
	}

	cfg, ok := c.Providers[c.Provider]
	if !ok {
		return nil, fmt.Errorf(
			"backend.providers.%s is missing; provider-specific config is required",
			c.Provider,
		)
	}

	return cfg, nil
}

// GetProviderConfig returns the config for the selected frontend provider.
func (c *FrontendConfig) GetProviderConfig() (any, error) {
	if c.Provider == "" {
		return nil, fmt.Errorf("frontend.provider is required")
	}

	if c.Providers == nil {
		return nil, fmt.Errorf("frontend.providers is required")
	}

	cfg, ok := c.Providers[c.Provider]
	if !ok {
		return nil, fmt.Errorf(
			"frontend.providers.%s is missing; provider-specific config is required",
			c.Provider,
		)
	}

	return cfg, nil
}

// DefaultConfigPath returns the default config path for the current working directory.
func DefaultConfigPath() string {
	return "stagecraft.yml"
}

// Exists reports whether a config file exists at the given path.
// It returns (false, nil) if the file does not exist.
func Exists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		return !info.IsDir(), nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

// Load reads and validates the config from the given path.
//
// It returns ErrConfigNotFound if the file does not exist.
func Load(path string) (*Config, error) {
	exists, err := Exists(path)
	if err != nil {
		return nil, fmt.Errorf("checking config existence: %w", err)
	}

	if !exists {
		return nil, ErrConfigNotFound
	}

	// nolint:gosec // G304: reading config file from user-specified path is expected behavior
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func validate(cfg *Config) error {
	if cfg.Project.Name == "" {
		return errors.New("config: project.name must be non-empty")
	}

	// Validate backend configuration (if present)
	if cfg.Backend != nil {
		if err := validateBackend(cfg.Backend); err != nil {
			return err
		}
	}

	// Validate frontend configuration (if present)
	if cfg.Frontend != nil {
		if err := validateFrontend(cfg.Frontend); err != nil {
			return err
		}
	}

	// Validate database configurations (if present)
	for dbName, dbCfg := range cfg.Databases {
		if err := validateDatabase(dbName, dbCfg); err != nil {
			return err
		}
	}

	// Validate environments
	for envName, envCfg := range cfg.Environments {
		if envName == "" {
			return errors.New("config: environment name must be non-empty")
		}
		if envCfg.Driver == "" {
			return fmt.Errorf("config: environment %q: driver must be non-empty", envName)
		}
	}

	return nil
}

// validateBackend validates backend configuration using the registry.
func validateBackend(cfg *BackendConfig) error {
	if cfg.Provider == "" {
		return fmt.Errorf("backend.provider is required")
	}

	if !backendproviders.Has(cfg.Provider) {
		return fmt.Errorf(
			"unknown backend provider %q; available providers: %v",
			cfg.Provider,
			backendproviders.DefaultRegistry.IDs(),
		)
	}

	if cfg.Providers == nil {
		return fmt.Errorf("backend.providers is required")
	}

	if _, ok := cfg.Providers[cfg.Provider]; !ok {
		return fmt.Errorf(
			"backend.providers.%s is missing; provider-specific config is required",
			cfg.Provider,
		)
	}

	return nil
}

// validateFrontend validates frontend configuration using the registry.
func validateFrontend(cfg *FrontendConfig) error {
	if cfg.Provider == "" {
		return fmt.Errorf("frontend.provider is required")
	}

	if !frontendproviders.Has(cfg.Provider) {
		return fmt.Errorf(
			"unknown frontend provider %q; available providers: %v",
			cfg.Provider,
			frontendproviders.DefaultRegistry.IDs(),
		)
	}

	if cfg.Providers == nil {
		return fmt.Errorf("frontend.providers is required")
	}

	if _, ok := cfg.Providers[cfg.Provider]; !ok {
		return fmt.Errorf(
			"frontend.providers.%s is missing; provider-specific config is required",
			cfg.Provider,
		)
	}

	return nil
}

// validateDatabase validates database configuration including migrations.
func validateDatabase(name string, db DatabaseConfig) error {
	if db.Migrations == nil {
		return nil // Migrations are optional
	}

	engine := db.Migrations.Engine
	if engine == "" {
		return fmt.Errorf("databases.%s.migrations.engine is required", name)
	}

	if !migrationengines.Has(engine) {
		return fmt.Errorf(
			"unknown migration engine %q for database %s; available engines: %v",
			engine,
			name,
			migrationengines.DefaultRegistry.IDs(),
		)
	}

	if db.Migrations.Path == "" {
		return fmt.Errorf("databases.%s.migrations.path is required", name)
	}

	// Validate strategy if present
	if db.Migrations.Strategy != "" {
		validStrategies := map[string]bool{
			"pre_deploy":  true,
			"post_deploy": true,
			"manual":      true,
		}
		if !validStrategies[db.Migrations.Strategy] {
			return fmt.Errorf(
				"databases.%s.migrations.strategy must be one of: pre_deploy, post_deploy, manual",
				name,
			)
		}
	}

	return nil
}

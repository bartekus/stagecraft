package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Feature: CORE_CONFIG
// Spec: spec/core/config.md

// ErrConfigNotFound is returned when the config file does not exist at the given path.
var ErrConfigNotFound = errors.New("stagecraft config not found")

// Config represents the top-level Stagecraft configuration.
type Config struct {
	Project      ProjectConfig                `yaml:"project"`
	Environments map[string]EnvironmentConfig `yaml:"environments"`
}

// ProjectConfig describes project-level settings.
type ProjectConfig struct {
	Name string `yaml:"name"`
}

// EnvironmentConfig describes per-environment settings.
type EnvironmentConfig struct {
	Driver string `yaml:"driver"`
	// Future: region, registry, etc.
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

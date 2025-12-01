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
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"stagecraft/pkg/config"
	"stagecraft/pkg/logging"
)

// Feature: CLI_INIT
// Spec: spec/commands/init.md

// NewInitCommand returns the `stagecraft init` command.
func NewInitCommand() *cobra.Command {
	var nonInteractive bool
	var configPath string
	var projectName string
	var envName string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Bootstrap Stagecraft into the current project",
		Long: `Initialize Stagecraft configuration in the current directory.

This command will create a minimal Stagecraft config file and guide you
through initial setup.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			verbose, _ := cmd.Flags().GetBool("verbose")
			logger := logging.NewLogger(verbose)

			if configPath == "" {
				configPath = config.DefaultConfigPath()
			}

			// Check if config already exists
			exists, err := config.Exists(configPath)
			if err != nil {
				return fmt.Errorf("checking existing config at %s: %w", configPath, err)
			}

			if exists {
				logger.Warn("Config file already exists", logging.NewField("path", configPath))
				fmt.Fprintf(out, "A Stagecraft config file already exists at %s.\n", configPath)
				fmt.Fprintf(out, "Run 'stagecraft init --config <path>' to create a config at a different location.\n")
				return nil
			}

			// Gather configuration - use os.Stdout for interactive prompts
			cfg, err := gatherConfig(os.Stdout, nonInteractive, projectName, envName)
			if err != nil {
				return fmt.Errorf("gathering configuration: %w", err)
			}

			// Write config file
			if err := writeConfig(configPath, cfg); err != nil {
				return fmt.Errorf("writing config file: %w", err)
			}

			logger.Info("Created Stagecraft config",
				logging.NewField("path", configPath),
				logging.NewField("project", cfg.Project.Name),
			)
			fmt.Fprintf(out, "✓ Created Stagecraft config at %s\n", configPath)
			fmt.Fprintf(out, "You can now run 'stagecraft dev' to start development.\n")

			return nil
		},
	}

	cmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "run without interactive prompts and use defaults")
	cmd.Flags().StringVar(&configPath, "config", "", "path to Stagecraft config file (default: stagecraft.yml)")
	cmd.Flags().StringVar(&projectName, "project-name", "", "project name (default: directory name)")
	cmd.Flags().StringVar(&envName, "env", "dev", "default environment name")

	return cmd
}

// gatherConfig collects configuration from user or uses defaults.
func gatherConfig(out *os.File, nonInteractive bool, projectName, envName string) (*config.Config, error) {
	// Get project name
	if projectName == "" {
		if nonInteractive {
			// Use current directory name as default
			wd, err := os.Getwd()
			if err != nil {
				return nil, fmt.Errorf("getting working directory: %w", err)
			}
			projectName = filepath.Base(wd)
		} else {
			wd, err := os.Getwd()
			if err != nil {
				return nil, fmt.Errorf("getting working directory: %w", err)
			}
			defaultName := filepath.Base(wd)
			fmt.Fprintf(out, "Project name [%s]: ", defaultName)
			var input string
			fmt.Scanln(&input)
			if strings.TrimSpace(input) == "" {
				projectName = defaultName
			} else {
				projectName = strings.TrimSpace(input)
			}
		}
	}

	if envName == "" {
		envName = "dev"
	}

	// Create minimal valid config
	cfg := &config.Config{
		Project: config.ProjectConfig{
			Name: projectName,
		},
		Environments: map[string]config.EnvironmentConfig{
			envName: {
				Driver: "local", // Default to local for initial setup
			},
		},
	}

	return cfg, nil
}

// writeConfig writes the config to disk as YAML.
func writeConfig(path string, cfg *config.Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package commands contains Cobra subcommands for the Stagecraft CLI.
package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"stagecraft/pkg/config"
	"stagecraft/pkg/logging"
	backendproviders "stagecraft/pkg/providers/backend"
)

// Feature: CLI_DEV_BASIC
// Spec: spec/commands/dev-basic.md

// NewDevCommand returns the `stagecraft dev` command.
func NewDevCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dev",
		Short: "Start development environment",
		Long:  "Loads stagecraft.yml, resolves backend provider, and runs dev mode",
		RunE:  runDev,
	}

	cmd.Flags().String("config", "", "path to Stagecraft config file (default: stagecraft.yml)")

	return cmd
}

func runDev(cmd *cobra.Command, args []string) error {
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

	// Load and validate config
	cfg, err := config.Load(configPath)
	if err != nil {
		if err == config.ErrConfigNotFound {
			return fmt.Errorf("stagecraft config not found at %s", configPath)
		}
		return fmt.Errorf("loading config: %w", err)
	}

	if cfg.Backend == nil {
		return fmt.Errorf("no backend configuration found in %s", configPath)
	}

	// Resolve backend provider
	backendID := cfg.Backend.Provider
	provider, err := backendproviders.Get(backendID)
	if err != nil {
		// Enhance error message with available providers
		available := backendproviders.DefaultRegistry.IDs()
		return fmt.Errorf("unknown backend provider %q; available providers: %v", backendID, available)
	}

	// Get provider-specific config
	providerCfg, err := cfg.Backend.GetProviderConfig()
	if err != nil {
		return fmt.Errorf("getting provider config: %w", err)
	}

	// Determine workdir (project root)
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	// Initialize logger
	verbose, _ := cmd.Flags().GetBool("verbose")
	logger := logging.NewLogger(verbose)
	logger.Info("Starting development environment",
		logging.NewField("provider", backendID),
		logging.NewField("config", absPath),
	)
	logger.Debug("Working directory", logging.NewField("workdir", workDir))

	// Call provider
	opts := backendproviders.DevOptions{
		Config:  providerCfg,
		WorkDir: workDir,
		Env:     make(map[string]string), // Future: load from env files
	}

	return provider.Dev(ctx, opts)
}

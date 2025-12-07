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
	frontendproviders "stagecraft/pkg/providers/frontend"
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

	// Global flags (--config, --env, --verbose, --dry-run) are inherited from root

	return cmd
}

func runDev(cmd *cobra.Command, args []string) error {
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

	// Check for dry-run mode
	if flags.DryRun {
		logger := logging.NewLogger(flags.Verbose)
		logger.Info("Dry-run mode: would start development environment",
			logging.NewField("env", flags.Env),
			logging.NewField("config", flags.Config),
		)
		return nil
	}

	if cfg.Backend == nil {
		return fmt.Errorf("no backend configuration found in %s", flags.Config)
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

	absPath, err := filepath.Abs(flags.Config)
	if err != nil {
		return fmt.Errorf("resolving config path: %w", err)
	}

	// Initialize logger
	logger := logging.NewLogger(flags.Verbose)
	logger.Info("Starting development environment",
		logging.NewField("provider", backendID),
		logging.NewField("config", absPath),
		logging.NewField("env", flags.Env),
	)
	logger.Debug("Working directory", logging.NewField("workdir", workDir))

	// Call backend provider
	backendOpts := backendproviders.DevOptions{
		Config:  providerCfg,
		WorkDir: workDir,
		Env:     make(map[string]string), // Future: load from env files
	}

	// If frontend is configured, start it as well
	if cfg.Frontend != nil {
		frontendID := cfg.Frontend.Provider
		frontendProvider, err := frontendproviders.Get(frontendID)
		if err != nil {
			available := frontendproviders.DefaultRegistry.IDs()
			return fmt.Errorf("unknown frontend provider %q; available providers: %v", frontendID, available)
		}

		frontendProviderCfg, err := cfg.Frontend.GetProviderConfig()
		if err != nil {
			return fmt.Errorf("getting frontend provider config: %w", err)
		}

		logger.Info("Starting frontend",
			logging.NewField("provider", frontendID),
		)

		frontendOpts := frontendproviders.DevOptions{
			Config:  frontendProviderCfg,
			WorkDir: workDir,
			Env:     make(map[string]string), // Future: load from env files
		}

		// For now, run frontend in a goroutine (simple parallel execution)
		// TODO: Replace with proper process management (DEV_PROCESS_MGMT)
		go func() {
			if err := frontendProvider.Dev(ctx, frontendOpts); err != nil {
				logger.Error("Frontend exited with error", logging.NewField("error", err))
			}
		}()
	}

	return provider.Dev(ctx, backendOpts)
}

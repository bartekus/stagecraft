// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Feature: CLI_GLOBAL_FLAGS
// Spec: spec/core/global-flags.md

package commands

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"stagecraft/pkg/config"
)

// ResolvedFlags contains the resolved values for all global flags.
type ResolvedFlags struct {
	Env     string
	Config  string
	Verbose bool
	DryRun  bool
}

// ResolveFlags resolves global flags with the following precedence:
// 1. Command-line flags (highest priority)
// 2. Environment variables
// 3. Config file defaults
// 4. Built-in defaults (lowest priority)
func ResolveFlags(cmd *cobra.Command, cfg *config.Config) (*ResolvedFlags, error) {
	flags := &ResolvedFlags{}

	// Resolve --env flag
	envFlag, _ := cmd.Flags().GetString("env")
	envEnv := os.Getenv("STAGECRAFT_ENV")
	envDefault := "dev" // Built-in default

	flags.Env = resolveString(envFlag, envEnv, envDefault)

	// Validate --env if config is available
	if cfg != nil {
		if flags.Env != "" {
			if _, exists := cfg.Environments[flags.Env]; !exists {
				available := make([]string, 0, len(cfg.Environments))
				for name := range cfg.Environments {
					available = append(available, name)
				}
				return nil, fmt.Errorf("invalid environment %q; available environments: %v", flags.Env, available)
			}
		}
	}

	// Resolve --config flag
	configFlag, _ := cmd.Flags().GetString("config")
	configEnv := os.Getenv("STAGECRAFT_CONFIG")
	configDefault := config.DefaultConfigPath() // Built-in default

	flags.Config = resolveString(configFlag, configEnv, configDefault)

	// Note: Config file existence validation is done by commands that actually use the config,
	// not here, as the file may not exist yet (e.g., during init)

	// Resolve --verbose flag
	verboseFlag, _ := cmd.Flags().GetBool("verbose")
	verboseEnv := parseBoolEnv(os.Getenv("STAGECRAFT_VERBOSE"))
	verboseDefault := false // Built-in default

	flags.Verbose = resolveBool(verboseFlag, verboseEnv, verboseDefault)

	// Resolve --dry-run flag
	dryRunFlag, _ := cmd.Flags().GetBool("dry-run")
	dryRunEnv := parseBoolEnv(os.Getenv("STAGECRAFT_DRY_RUN"))
	dryRunDefault := false // Built-in default

	flags.DryRun = resolveBool(dryRunFlag, dryRunEnv, dryRunDefault)

	return flags, nil
}

// resolveString resolves a string value with precedence: flag > env > default.
func resolveString(flag, env, defaultValue string) string {
	if flag != "" {
		return flag
	}
	if env != "" {
		return env
	}
	return defaultValue
}

// resolveBool resolves a boolean value with precedence: flag > env > default.
func resolveBool(flag, env, defaultValue bool) bool {
	if flag {
		return true
	}
	if env {
		return true
	}
	return defaultValue
}

// parseBoolEnv parses a boolean from an environment variable.
// Returns false if the env var is not set or cannot be parsed.
func parseBoolEnv(value string) bool {
	if value == "" {
		return false
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}
	return parsed
}


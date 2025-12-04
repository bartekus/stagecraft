// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package env provides environment resolution and context management.
package env

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"stagecraft/pkg/config"
)

// Feature: CORE_ENV_RESOLUTION
// Spec: spec/core/env-resolution.md

// ErrEnvironmentNotFound is returned when an environment is not found in config.
var ErrEnvironmentNotFound = errors.New("environment not found")

// Context represents an environment context with resolved settings.
type Context struct {
	// Name is the environment name (e.g., "dev", "staging", "prod")
	Name string

	// Config is the resolved environment configuration
	Config config.EnvironmentConfig

	// EnvFile is the path to the environment file
	EnvFile string

	// Variables are the resolved environment variables
	Variables map[string]string
}

// Resolver resolves environment contexts from configuration.
type Resolver struct {
	cfg     *config.Config
	workDir string
}

// NewResolver creates a new environment resolver.
func NewResolver(cfg *config.Config) *Resolver {
	return &Resolver{
		cfg:     cfg,
		workDir: ".",
	}
}

// SetWorkDir sets the working directory for resolving relative paths.
func (r *Resolver) SetWorkDir(workDir string) {
	r.workDir = workDir
}

// Resolve resolves an environment context by name.
func (r *Resolver) Resolve(ctx context.Context, name string) (*Context, error) {
	_ = ctx // Reserved for future cancellation/timeout support

	// 1. Look up environment in config
	envCfg, ok := r.cfg.Environments[name]
	if !ok {
		available := make([]string, 0, len(r.cfg.Environments))
		for envName := range r.cfg.Environments {
			available = append(available, envName)
		}
		return nil, fmt.Errorf("%w: %q (available: %v)", ErrEnvironmentNotFound, name, available)
	}

	// 2. Resolve env file path
	envFile := envCfg.EnvFile
	if envFile != "" && !filepath.IsAbs(envFile) {
		envFile = filepath.Join(r.workDir, envFile)
	}

	// 3. Load and merge environment variables
	variables, err := r.loadVariables(envFile)
	if err != nil {
		return nil, fmt.Errorf("loading environment variables: %w", err)
	}

	// 4. Apply variable interpolation
	variables = r.interpolateVariables(variables)

	// 5. Return resolved context
	return &Context{
		Name:      name,
		Config:    envCfg, // Value, not pointer
		EnvFile:   envFile,
		Variables: variables,
	}, nil
}

// ResolveFromFlags resolves an environment context from CLI flags.
// It uses envFlag if provided, otherwise defaults to "dev".
func (r *Resolver) ResolveFromFlags(ctx context.Context, envFlag string) (*Context, error) {
	_ = ctx // Reserved for future cancellation/timeout support

	envName := envFlag
	if envName == "" {
		envName = "dev"
	}

	return r.Resolve(ctx, envName)
}

// loadVariables loads and merges environment variables.
//
// Precedence (lowest to highest):
//  1. Env file variables
//  2. System environment variables (highest precedence)
func (r *Resolver) loadVariables(envFilePath string) (map[string]string, error) {
	variables := make(map[string]string)

	// 1. Load from env file if it exists (lowest precedence)
	if envFilePath != "" {
		if _, err := os.Stat(envFilePath); err == nil {
			// File exists, read and parse it
			//nolint:gosec // G304: envFilePath comes from trusted stagecraft.yml config
			data, err := os.ReadFile(envFilePath)
			if err != nil {
				return nil, fmt.Errorf("reading env file %q: %w", envFilePath, err)
			}

			// Parse dotenv format
			parseEnvFileInto(variables, data)
		}
		// If file doesn't exist, we continue (it's optional)
	}

	// 2. Override with system environment variables (highest precedence)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			variables[parts[0]] = parts[1]
		}
	}

	return variables, nil
}

// Package-level regex for variable interpolation (compiled once)
var varPattern = regexp.MustCompile(`\$\{([^}:]+)\}`)

// interpolateVariables performs variable interpolation using ${VAR} syntax.
// Supports nested/chained interpolation through multiple passes.
// Maximum of 5 passes to prevent infinite loops from circular references.
// Circular references will partially resolve and then stop after maxPasses.
func (r *Resolver) interpolateVariables(vars map[string]string) map[string]string {
	// Create a copy to avoid modifying the original
	result := make(map[string]string, len(vars))
	for k, v := range vars {
		result[k] = v
	}

	const maxPasses = 5
	for pass := 0; pass < maxPasses; pass++ {
		changed := false

		for key, value := range result {
			newValue := varPattern.ReplaceAllStringFunc(value, func(match string) string {
				varName := varPattern.FindStringSubmatch(match)[1]
				if interpolated, ok := result[varName]; ok {
					return interpolated
				}
				// If variable not found, return the original match (don't interpolate)
				return match
			})

			if newValue != value {
				result[key] = newValue
				changed = true
			}
		}

		// If no changes were made, we've converged
		if !changed {
			break
		}
	}

	return result
}

// parseEnvFileInto parses a dotenv-format file and merges key-value pairs into env.
// Semantics intentionally mirror the encorets provider parser for consistency.
// Handles: comments, export keyword, quoted values, inline comments,
// escaped characters in quoted strings, and empty values.
func parseEnvFileInto(env map[string]string, data []byte) {
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle export keyword (e.g., "export KEY=value")
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimPrefix(line, "export ")
			line = strings.TrimSpace(line)
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			// Skip malformed lines (no = found)
			continue
		}

		key := strings.TrimSpace(parts[0])
		if key == "" {
			// Skip lines with empty keys (e.g., "=value")
			continue
		}
		value := strings.TrimSpace(parts[1])

		// Handle inline comments (but preserve # inside quoted strings)
		commentIdx := -1
		inDoubleQuote := false
		inSingleQuote := false
		for i, r := range value {
			if r == '"' && (i == 0 || value[i-1] != '\\') {
				inDoubleQuote = !inDoubleQuote
			} else if r == '\'' && (i == 0 || value[i-1] != '\\') {
				inSingleQuote = !inSingleQuote
			} else if r == '#' && !inDoubleQuote && !inSingleQuote {
				commentIdx = i
				break
			}
		}
		if commentIdx >= 0 {
			value = strings.TrimSpace(value[:commentIdx])
		}

		// Handle quoted values with escaped characters
		if len(value) >= 2 {
			if value[0] == '"' && value[len(value)-1] == '"' {
				// Double-quoted string: handle escaped characters
				unquoted := value[1 : len(value)-1]
				unquoted = strings.ReplaceAll(unquoted, "\\\\", "\\")
				unquoted = strings.ReplaceAll(unquoted, "\\\"", "\"")
				unquoted = strings.ReplaceAll(unquoted, "\\n", "\n")
				unquoted = strings.ReplaceAll(unquoted, "\\t", "\t")
				unquoted = strings.ReplaceAll(unquoted, "\\r", "\r")
				value = unquoted
			} else if value[0] == '\'' && value[len(value)-1] == '\'' {
				// Single-quoted string: no escape sequences (remove quotes only)
				value = value[1 : len(value)-1]
			}
		}

		// Later values override earlier ones (map behavior)
		env[key] = value
	}
}

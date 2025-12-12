// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

// Package deploy provides deployment-related functionality including compose file generation and rollout execution.
//
// Feature: DEPLOY_COMPOSE_GEN
// Spec: spec/deploy/compose-gen.md
package deploy

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"stagecraft/internal/compose"
	"stagecraft/pkg/config"
)

// Feature: DEPLOY_COMPOSE_GEN
// Spec: spec/deploy/compose-gen.md

// ComposeGenerator generates compose files for deployment environments.
type ComposeGenerator struct {
	loader    *compose.Loader
	writeFile func(string, []byte, os.FileMode) error
	mkdirAll  func(string, os.FileMode) error
}

// NewComposeGenerator creates a new compose generator.
func NewComposeGenerator() *ComposeGenerator {
	return &ComposeGenerator{
		loader:    compose.NewLoader(),
		writeFile: os.WriteFile,
		mkdirAll:  os.MkdirAll,
	}
}

// NewComposeGeneratorWithFS allows injecting file operations for tests.
func NewComposeGeneratorWithFS(
	writeFn func(string, []byte, os.FileMode) error,
	mkdirFn func(string, os.FileMode) error,
) *ComposeGenerator {
	g := NewComposeGenerator()
	g.writeFile = writeFn
	g.mkdirAll = mkdirFn
	return g
}

// Generate generates a compose file for the environment.
// v1: single-host only, generates .stagecraft/rendered/<env>/docker-compose.yml
func (g *ComposeGenerator) Generate(
	cfg *config.Config,
	envName string,
	baseComposePath string,
	builtImageTag string,
	workdir string,
) (outputPath, hash string, err error) {
	// 1. Load base compose file
	composeFile, err := g.loader.Load(baseComposePath)
	if err != nil {
		return "", "", fmt.Errorf("loading base compose file: %w", err)
	}

	// 2. Load env_file variables if configured
	var envVars map[string]string
	envCfg, exists := cfg.Environments[envName]
	if exists && envCfg.EnvFile != "" {
		envFilePath := envCfg.EnvFile
		if !filepath.IsAbs(envFilePath) {
			envFilePath = filepath.Join(workdir, envFilePath)
		}

		// Parse env file (graceful if missing - log debug, continue)
		if data, err := os.ReadFile(envFilePath); err == nil {
			envVars = make(map[string]string)
			parseEnvFileInto(envVars, data)
		}
		// If file missing: no error, just continue without env vars
	}

	// 3. Mutate compose file: inject image tags and merge env vars
	// This preserves all fields (version, networks, volumes, configs, secrets, x-*)
	err = composeFile.Mutate(func(data map[string]any) error {
		services, ok := data["services"].(map[string]any)
		if !ok {
			return fmt.Errorf("compose file has no services section")
		}

		// Get sorted service names for deterministic processing
		serviceNames := composeFile.GetServices()

		for _, svcName := range serviceNames {
			svcData, ok := services[svcName].(map[string]any)
			if !ok {
				continue
			}

			// Always set image (forces Stagecraft's built tag, even if build: exists)
			svcData["image"] = builtImageTag

			// Merge env_file variables (existing env vars win)
			if len(envVars) > 0 {
				envMap, ok := svcData["environment"].(map[string]any)
				if !ok {
					envMap = make(map[string]any)
					svcData["environment"] = envMap
				}

				for k, v := range envVars {
					if _, exists := envMap[k]; !exists {
						envMap[k] = v
					}
				}

				// Normalize environment map keys (sort for determinism)
				svcData["environment"] = g.normalizeMap(envMap)
			}
		}

		return nil
	})
	if err != nil {
		return "", "", fmt.Errorf("mutating compose file: %w", err)
	}

	// 4. Marshal deterministically using ToYAML()
	yamlBytes, err := composeFile.ToYAML()
	if err != nil {
		return "", "", fmt.Errorf("marshaling compose file: %w", err)
	}

	// 5. Write to output path
	outputPath = filepath.Join(workdir, ".stagecraft", "rendered", envName, "docker-compose.yml")
	if err := g.mkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return "", "", fmt.Errorf("creating output directory: %w", err)
	}

	if err := g.writeFile(outputPath, yamlBytes, 0o644); err != nil {
		return "", "", fmt.Errorf("writing compose file: %w", err)
	}

	// 6. Compute hash of exact rendered bytes
	hashBytes := sha256.Sum256(yamlBytes)
	hash = hex.EncodeToString(hashBytes[:])

	return outputPath, hash, nil
}

// normalizeMap sorts map keys for deterministic output.
func (g *ComposeGenerator) normalizeMap(m map[string]any) map[string]any {
	if len(m) == 0 {
		return m
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	normalized := make(map[string]any, len(m))
	for _, k := range keys {
		normalized[k] = m[k]
	}

	return normalized
}

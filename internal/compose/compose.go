// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package compose provides Docker Compose file loading, parsing, and manipulation.
package compose

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"stagecraft/pkg/config"
)

// Feature: CORE_COMPOSE
// Spec: spec/core/compose.md

// ErrComposeNotFound is returned when the Compose file does not exist.
var ErrComposeNotFound = errors.New("compose file not found")

// ComposeFile represents a parsed Docker Compose file.
//
//nolint:revive // ComposeFile is intentionally descriptive, not stuttering
type ComposeFile struct {
	data map[string]any
	path string
}

// Loader loads and parses Docker Compose files.
type Loader struct{}

// NewLoader creates a new Compose file loader.
func NewLoader() *Loader {
	return &Loader{}
}

// Load loads a Compose file from the given path.
func (l *Loader) Load(path string) (*ComposeFile, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolving compose file path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%w: %s", ErrComposeNotFound, absPath)
		}
		return nil, fmt.Errorf("checking compose file: %w", err)
	}

	// nolint:gosec // G304: reading compose file from user-specified path is expected behavior
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("reading compose file: %w", err)
	}

	var composeData map[string]any
	if err := yaml.Unmarshal(data, &composeData); err != nil {
		return nil, fmt.Errorf("parsing compose file: %w", err)
	}

	return &ComposeFile{
		data: composeData,
		path: absPath,
	}, nil
}

// GetServices returns all service names from the Compose file.
func (c *ComposeFile) GetServices() []string {
	services, ok := c.data["services"].(map[string]any)
	if !ok {
		return []string{}
	}

	serviceNames := make([]string, 0, len(services))
	for name := range services {
		serviceNames = append(serviceNames, name)
	}

	return serviceNames
}

// GetServiceRoles returns the role mapping for services from config.
func (c *ComposeFile) GetServiceRoles(cfg *config.Config) map[string]string {
	roles := make(map[string]string)

	// Services config is not yet in the Config struct, so we return empty for now
	// This will be populated when services config is added to Config
	_ = cfg

	return roles
}

// FilterServices filters services by role or environment configuration.
// For now, returns all services. Full filtering will be implemented when
// service roles are added to config.
func (c *ComposeFile) FilterServices(roles []string) []string {
	// For v1, return all services
	// Future: filter by roles when service role mapping is available
	_ = roles
	return c.GetServices()
}

// GenerateOverride generates an environment-specific override file.
func (c *ComposeFile) GenerateOverride(env string, cfg *config.Config) ([]byte, error) {
	envCfg, ok := cfg.Environments[env]
	if !ok {
		return nil, fmt.Errorf("environment %q not found in config", env)
	}

	// Create override structure
	override := map[string]any{
		"version":  c.data["version"],
		"services": make(map[string]any),
	}

	services, ok := c.data["services"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("compose file missing services section")
	}

	// Process each service
	for serviceName, serviceData := range services {
		serviceMap, ok := serviceData.(map[string]any)
		if !ok {
			continue
		}

		// Check if service should be excluded (mode: external)
		if shouldExcludeService(serviceName, env, cfg) {
			continue
		}

		serviceOverride := c.generateServiceOverride(serviceName, serviceMap, env, envCfg, cfg)
		if serviceOverride != nil {
			servicesOverride := override["services"].(map[string]any)
			servicesOverride[serviceName] = serviceOverride
		}
	}

	// Marshal to YAML
	data, err := yaml.Marshal(override)
	if err != nil {
		return nil, fmt.Errorf("marshaling override: %w", err)
	}

	return data, nil
}

// generateServiceOverride generates override for a single service.
func (c *ComposeFile) generateServiceOverride(
	serviceName string,
	serviceMap map[string]any,
	env string,
	envCfg config.EnvironmentConfig,
	cfg *config.Config,
) map[string]any {
	override := make(map[string]any)

	// Resolve volumes
	if volumes := c.resolveVolumes(serviceMap, env, envCfg, cfg); volumes != nil {
		override["volumes"] = volumes
	}

	// Resolve ports
	if ports := c.resolvePorts(serviceName, serviceMap, env, envCfg, cfg); ports != nil {
		override["ports"] = ports
	}

	// Resolve environment variables
	if envVars := c.resolveEnvironment(serviceMap, env, envCfg, cfg); envVars != nil {
		override["environment"] = envVars
	}

	// Only return override if it has content
	if len(override) == 0 {
		return nil
	}

	return override
}

// resolveVolumes resolves volume paths from environment config.
func (c *ComposeFile) resolveVolumes(
	serviceMap map[string]any,
	env string,
	envCfg config.EnvironmentConfig,
	cfg *config.Config,
) []any {
	volumes, ok := serviceMap["volumes"].([]any)
	if !ok {
		return nil
	}

	resolved := make([]any, 0, len(volumes))
	for _, vol := range volumes {
		volStr, ok := vol.(string)
		if !ok {
			// Keep non-string volumes as-is
			resolved = append(resolved, vol)
			continue
		}

		// Resolve volume path
		resolvedVol := c.resolveVolumePath(volStr, env, envCfg, cfg)
		resolved = append(resolved, resolvedVol)
	}

	return resolved
}

// resolveVolumePath resolves a single volume path.
func (c *ComposeFile) resolveVolumePath(
	volumeSpec string,
	env string,
	envCfg config.EnvironmentConfig,
	cfg *config.Config,
) string {
	// Handle volume mount format: "source:target:options"
	// If source starts with ${, we need to find the closing } before splitting
	var sourceEnd int
	if strings.HasPrefix(volumeSpec, "${") {
		// Find the matching closing brace
		braceCount := 0
		for i, r := range volumeSpec {
			if r == '{' {
				braceCount++
			} else if r == '}' {
				braceCount--
				if braceCount == 0 {
					sourceEnd = i + 1
					break
				}
			}
		}
		// If we found the closing brace, extract source and resolve it
		if sourceEnd > 0 && sourceEnd < len(volumeSpec) && volumeSpec[sourceEnd] == ':' {
			source := volumeSpec[:sourceEnd]
			rest := volumeSpec[sourceEnd+1:]
			resolvedSource := c.resolveVolumeVariable(source, env, envCfg, cfg)
			return resolvedSource + ":" + rest
		}
	}

	// Fallback to simple split if no variable pattern found
	parts := strings.Split(volumeSpec, ":")
	if len(parts) < 2 {
		return volumeSpec
	}

	source := parts[0]
	target := parts[1]
	options := ""
	if len(parts) > 2 {
		options = ":" + strings.Join(parts[2:], ":")
	}

	// Resolve common volume variables
	// ${POSTGRES_VOLUME:-postgres_data} -> resolved value
	resolvedSource := c.resolveVolumeVariable(source, env, envCfg, cfg)

	return resolvedSource + ":" + target + options
}

// resolveVolumeVariable resolves volume variable references.
// For v1, this does basic variable resolution. Full environment config
// support will be added when EnvironmentConfig is extended.
func (c *ComposeFile) resolveVolumeVariable(
	varRef string,
	env string,
	envCfg config.EnvironmentConfig,
	cfg *config.Config,
) string {
	// Check for ${VAR:-default} or ${VAR} pattern
	if strings.HasPrefix(varRef, "${") && strings.HasSuffix(varRef, "}") {
		inner := strings.TrimPrefix(strings.TrimSuffix(varRef, "}"), "${")

		// Handle default value: VAR:-default
		if idx := strings.Index(inner, ":-"); idx > 0 {
			defaultValue := inner[idx+2:]
			// For v1, return default value
			// Future: resolve from environment config when EnvironmentConfig is extended
			_ = env
			_ = envCfg
			_ = cfg
			return defaultValue
		}

		// No default provided; for v1, preserve the original reference
		// until we have proper env resolution.
		_ = env
		_ = envCfg
		_ = cfg
		return varRef
	}

	return varRef
}

// resolvePorts resolves port publishing from environment config.
func (c *ComposeFile) resolvePorts(
	serviceName string,
	serviceMap map[string]any,
	env string,
	envCfg config.EnvironmentConfig,
	cfg *config.Config,
) []any {
	ports, ok := serviceMap["ports"].([]any)
	if !ok {
		return nil
	}

	resolved := make([]any, 0, len(ports))
	for _, port := range ports {
		portStr, ok := port.(string)
		if !ok {
			// Keep non-string ports as-is
			resolved = append(resolved, port)
			continue
		}

		// Resolve port variable
		resolvedPort := c.resolvePortVariable(portStr, serviceName, env, envCfg, cfg)
		if resolvedPort != "" {
			resolved = append(resolved, resolvedPort)
		}
		// Empty string means don't publish (remove port)
	}

	return resolved
}

// resolvePortVariable resolves port variable references.
// For v1, this does basic variable resolution. Full environment config
// support will be added when EnvironmentConfig is extended.
func (c *ComposeFile) resolvePortVariable(
	varRef string,
	serviceName string,
	env string,
	envCfg config.EnvironmentConfig,
	cfg *config.Config,
) string {
	// Check for ${VAR:-default} or ${VAR} pattern
	if strings.HasPrefix(varRef, "${") && strings.HasSuffix(varRef, "}") {
		inner := strings.TrimPrefix(strings.TrimSuffix(varRef, "}"), "${")

		if idx := strings.Index(inner, ":-"); idx > 0 {
			defaultValue := inner[idx+2:]
			// For v1, return default value
			// Future: resolve from environment config when EnvironmentConfig is extended
			// Empty default string still means "do not publish".
			_ = serviceName
			_ = env
			_ = envCfg
			_ = cfg
			return defaultValue
		}

		// No default; keep as-is for now so we don't accidentally
		// drop ports that should be resolved later.
		_ = serviceName
		_ = env
		_ = envCfg
		_ = cfg
		return varRef
	}

	return varRef
}

// resolveEnvironment resolves environment variables.
// Supports both list and map forms of environment configuration.
func (c *ComposeFile) resolveEnvironment(
	serviceMap map[string]any,
	env string,
	envCfg config.EnvironmentConfig,
	cfg *config.Config,
) any {
	if envSlice, ok := serviceMap["environment"].([]any); ok {
		// For v1, return environment as-is
		// Future: full variable interpolation
		_ = env
		_ = envCfg
		_ = cfg
		return envSlice
	}

	if envMap, ok := serviceMap["environment"].(map[string]any); ok {
		// For v1, return environment as-is
		// Future: full variable interpolation
		_ = env
		_ = envCfg
		_ = cfg
		return envMap
	}

	return nil
}

// shouldExcludeService checks if a service should be excluded from override.
func shouldExcludeService(serviceName, env string, cfg *config.Config) bool {
	// For v1, check if service has mode: external in environment config
	// This will be implemented when environment config includes service modes
	_ = serviceName
	_ = env
	_ = cfg

	return false
}

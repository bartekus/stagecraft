// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Feature: PROVIDER_NETWORK_TAILSCALE
// Spec: spec/providers/network/tailscale.md

package tailscale

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Config represents Tailscale provider configuration.
type Config struct {
	AuthKeyEnv    string              `yaml:"auth_key_env"`
	TailnetDomain string              `yaml:"tailnet_domain"`
	DefaultTags   []string            `yaml:"default_tags"`
	RoleTags      map[string][]string `yaml:"role_tags"`
	Install       InstallConfig       `yaml:"install"`
}

// InstallConfig contains Tailscale installation settings.
type InstallConfig struct {
	Method     string `yaml:"method"`      // "auto" or "skip"
	MinVersion string `yaml:"min_version"` // e.g., "1.78.0"
}

// parseConfig unmarshals provider config from generic interface.
func parseConfig(cfg any) (*Config, error) {
	// Convert to YAML bytes and unmarshal
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshaling config: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConfigInvalid, err)
	}

	// Validate required fields
	if config.AuthKeyEnv == "" {
		return nil, fmt.Errorf("%w: auth_key_env is required", ErrConfigInvalid)
	}
	if config.TailnetDomain == "" {
		return nil, fmt.Errorf("%w: tailnet_domain is required", ErrConfigInvalid)
	}

	// Set defaults
	if config.Install.Method == "" {
		config.Install.Method = "auto"
	}

	return &config, nil
}

// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Feature: PROVIDER_CLOUD_DO
// Spec: spec/providers/cloud/digitalocean.md

package digitalocean

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Config represents DigitalOcean provider configuration.
type Config struct {
	TokenEnv      string                           `yaml:"token_env"`      // Required: env var name for DO API token (token never stored)
	SSHKeyName    string                           `yaml:"ssh_key_name"`   // Required: SSH key name in DO account (must exist, validated via API)
	DefaultRegion string                           `yaml:"default_region"` // Optional: default region
	DefaultSize   string                           `yaml:"default_size"`   // Optional: default size
	Regions       []string                         `yaml:"regions"`        // Optional: allowed regions
	Sizes         []string                         `yaml:"sizes"`          // Optional: allowed sizes
	Hosts         map[string]map[string]HostConfig `yaml:"hosts"`          // Required: host definitions per environment
}

// HostConfig represents configuration for a single host.
type HostConfig struct {
	Role   string `yaml:"role"`   // Required: role (e.g., "gateway", "app", "db")
	Size   string `yaml:"size"`   // Optional: size (defaults to default_size)
	Region string `yaml:"region"` // Optional: region (defaults to default_region)
}

// parseConfig unmarshals provider config from generic interface.
func parseConfig(cfg any) (*Config, error) {
	// Convert to YAML bytes and unmarshal
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: marshaling config: %v", ErrConfigInvalid, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConfigInvalid, err)
	}

	// Validate required fields
	if config.TokenEnv == "" {
		return nil, fmt.Errorf("%w: token_env is required", ErrConfigInvalid)
	}
	if config.SSHKeyName == "" {
		return nil, fmt.Errorf("%w: ssh_key_name is required", ErrConfigInvalid)
	}
	if len(config.Hosts) == 0 {
		return nil, fmt.Errorf("%w: hosts configuration is required", ErrConfigInvalid)
	}

	// Validate host configs
	for env, hosts := range config.Hosts {
		for hostname, hostConfig := range hosts {
			if hostConfig.Role == "" {
				return nil, fmt.Errorf("%w: host %s.%s: role is required", ErrConfigInvalid, env, hostname)
			}
		}
	}

	return &config, nil
}

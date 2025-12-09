// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Feature: CLI_DEV
Specs: spec/commands/dev.md
Docs:
  - docs/engine/analysis/CLI_DEV.md
*/

// Package dev provides development environment topology and file generation.
package dev

import "stagecraft/pkg/config"

const (
	defaultFrontendDomain = "app.localdev.test"
	defaultBackendDomain  = "api.localdev.test"
)

// ComputeDomains computes dev domains from config with deterministic defaults.
//
// For v1, this reads from top-level dev.domains.* in the config.
// If dev.domains is not present or a domain is empty, the corresponding
// default is used.
//
// The env parameter is accepted for future environment-specific overrides,
// but v1 only uses global dev.domains.
//
// Returns deterministic domains that are used for:
//   - mkcert certificate generation (when HTTPS is enabled)
//   - Traefik router rules (when Traefik is enabled)
func ComputeDomains(cfg *config.Config, env string) (Domains, error) {
	domains := Domains{
		Frontend: defaultFrontendDomain,
		Backend:  defaultBackendDomain,
	}

	// v1: use top-level dev.domains.* if present.
	if cfg.Dev != nil && cfg.Dev.Domains != nil {
		if cfg.Dev.Domains.Frontend != "" {
			domains.Frontend = cfg.Dev.Domains.Frontend
		}
		if cfg.Dev.Domains.Backend != "" {
			domains.Backend = cfg.Dev.Domains.Backend
		}
	}

	// env parameter is accepted for future env-specific overrides.
	_ = env

	return domains, nil
}

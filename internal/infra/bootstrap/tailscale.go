// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

// Package bootstrap implements the INFRA_HOST_BOOTSTRAP engine.
//
// Feature: INFRA_HOST_BOOTSTRAP
// Spec: spec/infra/bootstrap.md
package bootstrap

import (
	"context"
	"fmt"

	"stagecraft/pkg/providers/network"
)

// ensureTailscale ensures Tailscale is installed and joined on the host via NetworkProvider.
//
// It performs:
// 1. EnsureInstalled: Ensures Tailscale client is installed on the host
// 2. EnsureJoined: Ensures the host is joined to the Tailscale network with appropriate tags
//
// Both operations are idempotent - safe to call multiple times.
//
// Returns (true, nil) if Tailscale is working, (false, error) otherwise.
//
//nolint:gocritic // hugeParam: host is passed by value for consistency with interface methods
func (s *service) ensureTailscale(ctx context.Context, host Host, cfg Config) (bool, error) { //nolint:unparam // cfg is kept for future use (see TODO below)
	// Map bootstrap.Host to network provider hostname
	// Use Name as the hostname (e.g., "app-1")
	hostname := host.Name
	if hostname == "" {
		hostname = host.ID
	}

	// Get network provider config from bootstrap config
	// For now, we'll need to pass this through - it should come from cfg.Infra.Bootstrap
	// but for v1, we'll use an empty config and let the provider handle defaults
	var networkConfig any
	// TODO: Extract network config from bootstrap.Config when it's added

	// Step 1: Ensure Tailscale is installed
	installOpts := network.EnsureInstalledOptions{
		Config: networkConfig,
		Host:   hostname,
	}
	if err := s.networkProvider.EnsureInstalled(ctx, installOpts); err != nil {
		return false, fmt.Errorf("tailscale install failed: %w", err)
	}

	// Step 2: Ensure host is joined to the network
	joinOpts := network.EnsureJoinedOptions{
		Config: networkConfig,
		Host:   hostname,
		Tags:   host.Tags,
	}
	if err := s.networkProvider.EnsureJoined(ctx, joinOpts); err != nil {
		return false, fmt.Errorf("tailscale join failed: %w", err)
	}

	return true, nil
}

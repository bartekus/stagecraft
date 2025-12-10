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
	"testing"

	"stagecraft/pkg/providers/network"
)

func TestTailscaleProvider_Registration(t *testing.T) {
	// Import the package to trigger init()
	_ = "imported"

	// Check if provider is registered
	provider, err := network.Get("tailscale")
	if err != nil {
		t.Fatalf("Get(\"tailscale\") error = %v, want nil", err)
	}

	if provider.ID() != "tailscale" {
		t.Errorf("provider.ID() = %q, want %q", provider.ID(), "tailscale")
	}
}

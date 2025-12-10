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
	"encoding/json"
	"fmt"
)

// TailscaleStatus represents the output of `tailscale status --json`.
//
//nolint:revive // TailscaleStatus is intentionally named for clarity in provider package
type TailscaleStatus struct {
	TailnetName string   `json:"TailnetName"`
	Self        NodeInfo `json:"Self"`
}

// NodeInfo contains information about a Tailscale node.
type NodeInfo struct {
	Online       bool     `json:"Online"`
	TailscaleIPs []string `json:"TailscaleIPs"`
	Tags         []string `json:"Tags"`
}

// parseStatus parses JSON output from `tailscale status --json`.
func parseStatus(jsonData string) (*TailscaleStatus, error) {
	var status TailscaleStatus
	if err := json.Unmarshal([]byte(jsonData), &status); err != nil {
		return nil, fmt.Errorf("parsing tailscale status: %w", err)
	}
	return &status, nil
}

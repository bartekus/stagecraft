// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package network provides interfaces and types for network providers.
package network

import "context"

// Feature: PROVIDER_NETWORK_INTERFACE
// Spec: spec/providers/network/interface.md

// EnsureInstalledOptions contains options for ensuring network client is installed.
type EnsureInstalledOptions struct {
	// Config is the provider-specific configuration decoded from
	// network.providers[providerID] in stagecraft.yml.
	// The provider implementation is responsible for unmarshaling this.
	Config any

	// Host is the hostname or Tailscale node name where to ensure installation
	Host string
}

// EnsureJoinedOptions contains options for ensuring a host is joined to the network.
type EnsureJoinedOptions struct {
	// Config is the provider-specific configuration
	Config any

	// Host is the hostname or Tailscale node name
	Host string

	// Tags are the tags to apply to the node (e.g., ["tag:gateway", "tag:app"])
	Tags []string
}

// NetworkProvider is the interface that all network providers must implement.
//
//nolint:revive // NetworkProvider is the preferred name for clarity
type NetworkProvider interface {
	// ID returns the unique identifier for this provider (e.g., "tailscale", "headscale").
	ID() string

	// EnsureInstalled ensures the network client is installed on the given host.
	EnsureInstalled(ctx context.Context, opts EnsureInstalledOptions) error

	// EnsureJoined ensures the host is joined to the mesh network with the given tags.
	EnsureJoined(ctx context.Context, opts EnsureJoinedOptions) error

	// NodeFQDN returns the fully qualified domain name for a node in the mesh network.
	// For example, "plat-db-1.mytailnet.ts.net" for Tailscale.
	NodeFQDN(host string) (string, error)
}

// ProviderMetadata contains metadata about a provider.
type ProviderMetadata struct {
	Name         string
	Description  string
	Version      string
	Author       string
	Experimental bool
}

// MetadataProvider is an optional interface that providers can implement
// to expose descriptive metadata.
type MetadataProvider interface {
	// Base provider interface
	NetworkProvider

	// Metadata returns descriptive metadata about the provider.
	Metadata() ProviderMetadata
}

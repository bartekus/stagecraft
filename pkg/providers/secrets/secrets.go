// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package secrets provides interfaces and types for secrets providers.
package secrets

import "context"

// Feature: PROVIDER_SECRETS_INTERFACE
// Spec: spec/providers/secrets/interface.md

// SyncOptions contains options for syncing secrets.
type SyncOptions struct {
	// Config is the provider-specific configuration decoded from
	// secrets.providers[providerID] in stagecraft.yml.
	// The provider implementation is responsible for unmarshaling this.
	Config any

	// Source is the source environment or location (e.g., "dev", ".env.local")
	Source string

	// Target is the target environment or location (e.g., "staging", "encore")
	Target string

	// Keys are the specific secret keys to sync (empty means sync all)
	Keys []string
}

// SecretsProvider is the interface that all secrets providers must implement.
//
//nolint:revive // SecretsProvider is the preferred name for clarity
type SecretsProvider interface {
	// ID returns the unique identifier for this provider (e.g., "envfile", "encore").
	ID() string

	// Sync syncs secrets from source to target.
	Sync(ctx context.Context, opts SyncOptions) error
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
	SecretsProvider

	// Metadata returns descriptive metadata about the provider.
	Metadata() ProviderMetadata
}

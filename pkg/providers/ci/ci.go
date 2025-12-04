// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package ci provides interfaces and types for CI providers.
package ci

import "context"

// Feature: PROVIDER_CI_INTERFACE
// Spec: spec/providers/ci/interface.md

// InitOptions contains options for initializing CI pipelines.
type InitOptions struct {
	// Config is the provider-specific configuration decoded from
	// ci.providers[providerID] in stagecraft.yml.
	// The provider implementation is responsible for unmarshaling this.
	Config any

	// WorkDir is the working directory (typically repository root)
	WorkDir string
}

// TriggerOptions contains options for triggering a CI run.
type TriggerOptions struct {
	// Config is the provider-specific configuration
	Config any

	// Environment is the environment to deploy to (e.g., "staging", "prod")
	Environment string

	// Version is the version to deploy (e.g., "v1.2.3" or git SHA)
	Version string
}

// CIProvider is the interface that all CI providers must implement.
//
//nolint:revive // CIProvider is the preferred name for clarity
type CIProvider interface {
	// ID returns the unique identifier for this provider (e.g., "github", "gitlab").
	ID() string

	// Init initializes CI pipelines in the repository.
	// This typically creates workflow files (e.g., .github/workflows/deploy.yml).
	Init(ctx context.Context, opts InitOptions) error

	// Trigger triggers a CI run for the given environment and version.
	Trigger(ctx context.Context, opts TriggerOptions) error
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
	CIProvider

	// Metadata returns descriptive metadata about the provider.
	Metadata() ProviderMetadata
}

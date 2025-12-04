// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package frontend provides interfaces and types for frontend providers.
package frontend

import "context"

// Feature: PROVIDER_FRONTEND_INTERFACE
// Spec: spec/providers/frontend/interface.md

// DevOptions contains options for running a frontend in development mode.
type DevOptions struct {
	// Config is the provider-specific configuration decoded from
	// frontend.providers[providerID] in stagecraft.yml.
	// The provider implementation is responsible for unmarshaling this.
	Config any

	// WorkDir is the working directory for the frontend
	WorkDir string

	// Env is the environment variables to pass to the dev process
	Env map[string]string
}

// FrontendProvider is the interface that all frontend providers must implement.
//
//nolint:revive // FrontendProvider is the preferred name for clarity
type FrontendProvider interface {
	// ID returns the unique identifier for this provider (e.g., "generic", "vite").
	ID() string

	// Dev runs the frontend in development mode.
	Dev(ctx context.Context, opts DevOptions) error
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
	FrontendProvider

	// Metadata returns descriptive metadata about the provider.
	Metadata() ProviderMetadata
}

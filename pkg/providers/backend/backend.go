// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package backend

import "context"

// Feature: PROVIDER_BACKEND_INTERFACE
// Spec: spec/core/backend-registry.md

// DevOptions contains options for running a backend in development mode.
type DevOptions struct {
	// Config is the provider-specific configuration decoded from
	// backend.providers[providerID] in stagecraft.yml.
	// The provider implementation is responsible for unmarshaling this.
	Config any

	// WorkDir is the working directory for the backend
	WorkDir string

	// Env is the environment variables to pass to the dev process
	Env map[string]string
}

// BuildDockerOptions contains options for building a Docker image.
type BuildDockerOptions struct {
	// Config is the provider-specific configuration
	Config any

	// ImageTag is the full image tag (e.g., "ghcr.io/org/app:tag")
	ImageTag string

	// WorkDir is the working directory for the build
	WorkDir string
}

// BackendProvider is the interface that all backend providers must implement.
type BackendProvider interface {
	// ID returns the unique identifier for this provider (e.g., "encore-ts", "generic").
	ID() string

	// Dev runs the backend in development mode.
	Dev(ctx context.Context, opts DevOptions) error

	// BuildDocker builds a Docker image for the backend.
	BuildDocker(ctx context.Context, opts BuildDockerOptions) (string, error)
}

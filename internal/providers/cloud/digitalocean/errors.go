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

import "errors"

// Error definitions for DigitalOcean provider.

// Config errors (local, deterministic, no API calls).
var (
	// ErrConfigInvalid indicates invalid provider configuration.
	ErrConfigInvalid = errors.New("digitalocean provider: invalid config")
)

// Authentication errors (API calls required).
var (
	// ErrTokenMissing indicates API token is missing from environment.
	ErrTokenMissing = errors.New("digitalocean provider: API token missing from environment")

	// ErrSSHKeyNotFound indicates SSH key is not found in DigitalOcean account.
	ErrSSHKeyNotFound = errors.New("digitalocean provider: SSH key not found")
)

// Resource errors (API operations).
var (
	// ErrDropletExists indicates droplet already exists (when reconciliation needed).
	ErrDropletExists = errors.New("digitalocean provider: droplet already exists")

	// ErrDropletNotFound indicates droplet not found.
	ErrDropletNotFound = errors.New("digitalocean provider: droplet not found")

	// ErrDropletCreateFailed indicates droplet creation failed.
	ErrDropletCreateFailed = errors.New("digitalocean provider: droplet creation failed")

	// ErrDropletDeleteFailed indicates droplet deletion failed.
	ErrDropletDeleteFailed = errors.New("digitalocean provider: droplet deletion failed")

	// ErrDropletTimeout indicates droplet operation timeout.
	ErrDropletTimeout = errors.New("digitalocean provider: droplet operation timeout")
)

// API errors (infrastructure/rate limiting).
var (
	// ErrAPIError indicates DigitalOcean API error (wraps underlying API errors).
	ErrAPIError = errors.New("digitalocean provider: API error")

	// ErrRateLimit indicates API rate limit exceeded (with retry logic).
	ErrRateLimit = errors.New("digitalocean provider: API rate limit exceeded")
)

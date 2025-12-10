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

import "errors"

// Error definitions for Tailscale provider.
var (
	// ErrConfigInvalid indicates invalid provider configuration.
	ErrConfigInvalid = errors.New("invalid config")

	// ErrAuthKeyMissing indicates auth key is missing from environment.
	ErrAuthKeyMissing = errors.New("auth key missing from environment")

	// ErrAuthKeyInvalid indicates auth key is invalid or expired.
	ErrAuthKeyInvalid = errors.New("invalid or expired auth key")

	// ErrTailnetMismatch indicates host is in different tailnet than expected.
	ErrTailnetMismatch = errors.New("tailnet mismatch")

	// ErrTagMismatch indicates host tags do not match expected tags.
	ErrTagMismatch = errors.New("tag mismatch")

	// ErrInstallFailed indicates Tailscale installation failed.
	ErrInstallFailed = errors.New("tailscale installation failed")

	// ErrUnsupportedOS indicates unsupported operating system.
	ErrUnsupportedOS = errors.New("unsupported operating system")
)

// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package frontend

import (
	"testing"
)

// Feature: PROVIDER_FRONTEND_INTERFACE
// Spec: spec/providers/frontend/interface.md

// Note: Registry tests are in frontend_test.go
// This file exists for consistency with backend/migration package structure
// and can be used for additional registry-specific tests if needed.

func TestRegistry_ImplementsInterface(t *testing.T) {
	// This test ensures the registry methods match the expected interface
	reg := NewRegistry()

	// Verify registry has expected methods
	_ = reg.Register
	_ = reg.Get
	_ = reg.Has
	_ = reg.IDs
}

// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package state

import "testing"

// Feature: CORE_STATE
// Spec: spec/core/state.md

func TestCoreStateFeatureAnchor_NewManagerNonNil(t *testing.T) {
	mgr := NewManager("test-releases.json")
	if mgr == nil {
		t.Fatal("expected NewManager to return non-nil manager")
	}
}

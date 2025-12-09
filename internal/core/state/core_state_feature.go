// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package state

// Feature: CORE_STATE
// Spec: spec/core/state.md

// coreStateFeatureAnchor is a no-op used to anchor the CORE_STATE feature mapping.
//
// It documents that the state.Manager and related types in this package implement
// the CORE_STATE feature. The real behavior is exercised in state_test.go and by
// callers in the CLI layer.
func coreStateFeatureAnchor() {} //nolint:unused // Anchor function for feature mapping

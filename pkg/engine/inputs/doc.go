// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

// Package inputs provides typed, validated, and normalized Inputs structs
// for each engine.StepAction.
//
// Contract:
//
// Producers (Planner/Adapter):
//  1. Create typed struct from metadata
//  2. Call Normalize() - sorts set-like fields, normalizes paths
//  3. Call Validate() - enforces required fields and constraints
//  4. Marshal to JSON
//
// Consumers (Agent/Executor):
//  1. UnmarshalStrict() - rejects unknown fields (DisallowUnknownFields)
//  2. Validate() - re-validate defensively
//  3. Use inputs
//
// Determinism: all set-like lists are sorted, paths are normalized,
// hashes are validated, and JSON output is deterministic.
package inputs

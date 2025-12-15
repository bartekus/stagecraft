// Package migrations defines the interface and types for migration engines.
//
// SPDX-License-Identifier: AGPL-3.0-or-later
package migrations

import "context"

// Engine Identity
// Each engine MUST expose a stable Name() string.
type Engine interface {
	// Name returns the stable name of the engine (e.g., "raw", "prisma").
	Name() string

	// List returns the candidate migrations for the given environment and selection scope.
	// Returned list MUST be deterministically ordered.
	List(ctx context.Context, req *MigrationRequest) ([]Migration, error)

	// Plan returns a deterministic migration plan without mutating the target.
	Plan(ctx context.Context, req *MigrationRequest) (MigrationPlan, error)

	// Apply executes migrations and returns the result. May mutate the target.
	Apply(ctx context.Context, req *MigrationRequest) (MigrationApplyResult, error)
}

// ValidatingEngine is an optional interface for engines that support validation.
type ValidatingEngine interface {
	Engine
	// Validate checks if the engine is configured correctly for the request.
	Validate(ctx context.Context, req *MigrationRequest) (ValidationResult, error)
}

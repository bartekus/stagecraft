// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package migration

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

// Feature: CORE_MIGRATION_REGISTRY
// Spec: spec/core/migration-registry.md

const registryName = "migration.Registry"

var (
	// ErrUnknownProvider is returned when Get() is called with an unknown provider ID.
	ErrUnknownProvider = errors.New("unknown provider")
	// ErrDuplicateProvider is used when attempting to register a provider with a duplicate ID.
	ErrDuplicateProvider = errors.New("duplicate provider ID")
	// ErrEmptyProviderID is used when attempting to register a provider with an empty ID.
	ErrEmptyProviderID = errors.New("empty provider ID")
)

// Instrumentation hooks for observability (optional).
var (
	OnProviderRegistered func(kind, id string)
	OnProviderLookup     func(kind, id string, found bool)
)

// Registry manages migration engine registration and lookup.
type Registry struct {
	mu      sync.RWMutex
	engines map[string]Engine
}

// NewRegistry creates a new empty registry.
func NewRegistry() *Registry {
	return &Registry{
		engines: make(map[string]Engine),
	}
}

// Register registers a migration engine.
// Panics if the engine ID is empty or already registered.
func (r *Registry) Register(e Engine) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := e.ID()
	if id == "" {
		panic(fmt.Sprintf("%s.Register: %v", registryName, ErrEmptyProviderID))
	}
	if _, exists := r.engines[id]; exists {
		panic(fmt.Sprintf("%s.Register: %v: %q", registryName, ErrDuplicateProvider, id))
	}

	r.engines[id] = e

	if OnProviderRegistered != nil {
		OnProviderRegistered(registryName, id)
	}
}

// Get retrieves an engine by ID.
// Returns an error if the engine is not found.
func (r *Registry) Get(id string) (Engine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	e, ok := r.engines[id]
	if OnProviderLookup != nil {
		OnProviderLookup(registryName, id, ok)
	}
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownProvider, id)
	}
	return e, nil
}

// Has checks if an engine with the given ID is registered.
func (r *Registry) Has(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.engines[id]
	return ok
}

// IDs returns all registered engine IDs in deterministic lexicographic order.
// Determinism is required by Agent.md for stable output and golden tests.
func (r *Registry) IDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.engines))
	for id := range r.engines {
		ids = append(ids, id)
	}
	sort.Strings(ids) // Ensure deterministic lexicographic ordering
	return ids
}

// List returns all registered engines in lexicographic order by ID.
func (r *Registry) List() []Engine {
	r.mu.RLock()
	defer r.mu.RUnlock()

	engines := make([]Engine, 0, len(r.engines))
	for _, e := range r.engines {
		engines = append(engines, e)
	}

	// Deterministic order by ID
	sort.Slice(engines, func(i, j int) bool {
		return engines[i].ID() < engines[j].ID()
	})

	return engines
}

// DefaultRegistry is the global default registry.
var DefaultRegistry = NewRegistry()

// Register registers an engine in the default registry.
func Register(e Engine) {
	DefaultRegistry.Register(e)
}

// Get retrieves an engine from the default registry.
func Get(id string) (Engine, error) {
	return DefaultRegistry.Get(id)
}

// Has checks if an engine exists in the default registry.
func Has(id string) bool {
	return DefaultRegistry.Has(id)
}

// List returns all engines from the default registry.
func List() []Engine {
	return DefaultRegistry.List()
}

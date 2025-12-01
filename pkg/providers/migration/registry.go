// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - A Go-based CLI for orchestrating local-first multi-service deployments using Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package migration

import (
	"fmt"
	"sync"
)

// Feature: CORE_MIGRATION_REGISTRY
// Spec: spec/core/migration-registry.md

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
		panic("migration engine registration: empty ID")
	}
	if _, exists := r.engines[id]; exists {
		panic(fmt.Sprintf("migration engine registration: duplicate ID %q", id))
	}

	r.engines[id] = e
}

// Get retrieves an engine by ID.
// Returns an error if the engine is not found.
func (r *Registry) Get(id string) (Engine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	e, ok := r.engines[id]
	if !ok {
		return nil, fmt.Errorf("unknown migration engine %q", id)
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

// IDs returns all registered engine IDs.
func (r *Registry) IDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.engines))
	for id := range r.engines {
		ids = append(ids, id)
	}
	return ids
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


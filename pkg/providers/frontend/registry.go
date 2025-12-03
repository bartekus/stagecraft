// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package frontend

import (
	"fmt"
	"sort"
	"sync"
)

// Feature: PROVIDER_FRONTEND_INTERFACE
// Spec: spec/providers/frontend/interface.md

// Registry manages frontend provider registration and lookup.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]FrontendProvider
}

// NewRegistry creates a new empty registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]FrontendProvider),
	}
}

// Register registers a frontend provider.
// Panics if the provider ID is empty or already registered.
func (r *Registry) Register(p FrontendProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := p.ID()
	if id == "" {
		panic("frontend provider registration: empty ID")
	}
	if _, exists := r.providers[id]; exists {
		panic(fmt.Sprintf("frontend provider registration: duplicate ID %q", id))
	}

	r.providers[id] = p
}

// Get retrieves a provider by ID.
// Returns an error if the provider is not found.
func (r *Registry) Get(id string) (FrontendProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.providers[id]
	if !ok {
		return nil, fmt.Errorf("unknown frontend provider %q", id)
	}
	return p, nil
}

// Has checks if a provider with the given ID is registered.
func (r *Registry) Has(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.providers[id]
	return ok
}

// IDs returns all registered provider IDs in deterministic lexicographic order.
// Determinism is required by Agent.md for stable output and golden tests.
func (r *Registry) IDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.providers))
	for id := range r.providers {
		ids = append(ids, id)
	}
	sort.Strings(ids) // Ensure deterministic lexicographic ordering
	return ids
}

// DefaultRegistry is the global default registry.
var DefaultRegistry = NewRegistry()

// Register registers a provider in the default registry.
func Register(p FrontendProvider) {
	DefaultRegistry.Register(p)
}

// Get retrieves a provider from the default registry.
func Get(id string) (FrontendProvider, error) {
	return DefaultRegistry.Get(id)
}

// Has checks if a provider exists in the default registry.
func Has(id string) bool {
	return DefaultRegistry.Has(id)
}


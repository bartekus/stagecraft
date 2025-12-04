// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package backend

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

// Feature: CORE_BACKEND_REGISTRY
// Spec: spec/core/backend-registry.md

const registryName = "backend.Registry"

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

// Registry manages backend provider registration and lookup.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]BackendProvider
}

// NewRegistry creates a new empty registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]BackendProvider),
	}
}

// Register registers a backend provider.
// Panics if the provider ID is empty or already registered.
func (r *Registry) Register(p BackendProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := p.ID()
	if id == "" {
		panic(fmt.Sprintf("%s.Register: %v", registryName, ErrEmptyProviderID))
	}
	if _, exists := r.providers[id]; exists {
		panic(fmt.Sprintf("%s.Register: %v: %q", registryName, ErrDuplicateProvider, id))
	}

	r.providers[id] = p

	if OnProviderRegistered != nil {
		OnProviderRegistered(registryName, id)
	}
}

// Get retrieves a provider by ID.
// Returns an error if the provider is not found.
func (r *Registry) Get(id string) (BackendProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.providers[id]
	if OnProviderLookup != nil {
		OnProviderLookup(registryName, id, ok)
	}
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownProvider, id)
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

// IDs returns all registered provider IDs.
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

// List returns all registered providers in lexicographic order by ID.
func (r *Registry) List() []BackendProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]BackendProvider, 0, len(r.providers))
	for _, p := range r.providers {
		providers = append(providers, p)
	}

	// Deterministic order by ID
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].ID() < providers[j].ID()
	})

	return providers
}

// DefaultRegistry is the global default registry.
var DefaultRegistry = NewRegistry()

// Register registers a provider in the default registry.
func Register(p BackendProvider) {
	DefaultRegistry.Register(p)
}

// Get retrieves a provider from the default registry.
func Get(id string) (BackendProvider, error) {
	return DefaultRegistry.Get(id)
}

// Has checks if a provider exists in the default registry.
func Has(id string) bool {
	return DefaultRegistry.Has(id)
}

// List returns all providers from the default registry.
func List() []BackendProvider {
	return DefaultRegistry.List()
}

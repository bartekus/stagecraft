// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package ci

import (
	"context"
	"sync"
	"testing"
)

// Feature: CORE_CI_REGISTRY
// Spec: spec/providers/ci/interface.md

// mockProvider is a test implementation of CIProvider.
type mockProvider struct {
	id string
}

func (m *mockProvider) ID() string {
	return m.id
}

func (m *mockProvider) Init(ctx context.Context, opts InitOptions) error {
	return nil
}

func (m *mockProvider) Trigger(ctx context.Context, opts TriggerOptions) error {
	return nil
}

func TestRegistry_Register(t *testing.T) {
	reg := NewRegistry()

	p1 := &mockProvider{id: "test-provider-1"}
	p2 := &mockProvider{id: "test-provider-2"}

	reg.Register(p1)
	reg.Register(p2)

	// Verify both are registered
	if !reg.Has("test-provider-1") {
		t.Error("expected test-provider-1 to be registered")
	}
	if !reg.Has("test-provider-2") {
		t.Error("expected test-provider-2 to be registered")
	}
}

func TestRegistry_Register_PanicsOnEmptyID(t *testing.T) {
	reg := NewRegistry()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when registering provider with empty ID")
		}
	}()

	p := &mockProvider{id: ""}
	reg.Register(p)
}

func TestRegistry_Register_PanicsOnDuplicateID(t *testing.T) {
	reg := NewRegistry()

	p1 := &mockProvider{id: "duplicate"}
	p2 := &mockProvider{id: "duplicate"}

	reg.Register(p1)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when registering duplicate provider ID")
		}
	}()

	reg.Register(p2)
}

func TestRegistry_Get(t *testing.T) {
	reg := NewRegistry()

	p := &mockProvider{id: "test-provider"}
	reg.Register(p)

	got, err := reg.Get("test-provider")
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}

	if got.ID() != "test-provider" {
		t.Errorf("Get() returned provider with ID %q, want %q", got.ID(), "test-provider")
	}
}

func TestRegistry_Get_ReturnsErrorForUnknownID(t *testing.T) {
	reg := NewRegistry()

	_, err := reg.Get("unknown-provider")
	if err == nil {
		t.Error("Get() error = nil, want error for unknown provider")
	}

	if reg.Has("unknown-provider") {
		t.Error("Has() = true for unknown provider, want false")
	}
}

func TestRegistry_Has(t *testing.T) {
	reg := NewRegistry()

	if reg.Has("test-provider") {
		t.Error("Has() = true for unregistered provider, want false")
	}

	p := &mockProvider{id: "test-provider"}
	reg.Register(p)

	if !reg.Has("test-provider") {
		t.Error("Has() = false for registered provider, want true")
	}
}

func TestRegistry_IDs(t *testing.T) {
	reg := NewRegistry()

	// Empty registry should return empty slice
	ids := reg.IDs()
	if len(ids) != 0 {
		t.Errorf("IDs() length = %d, want 0", len(ids))
	}

	// Register multiple providers in non-alphabetical order
	providers := []*mockProvider{
		{id: "provider-3"},
		{id: "provider-1"},
		{id: "provider-2"},
	}

	for _, p := range providers {
		reg.Register(p)
	}

	ids = reg.IDs()
	if len(ids) != 3 {
		t.Errorf("IDs() length = %d, want 3", len(ids))
	}

	// Verify IDs are sorted lexicographically
	expected := []string{"provider-1", "provider-2", "provider-3"}
	for i, id := range ids {
		if id != expected[i] {
			t.Errorf("IDs()[%d] = %q, want %q (IDs must be sorted)", i, id, expected[i])
		}
	}

	// Verify all IDs are present
	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[id] = true
	}

	for _, p := range providers {
		if !idMap[p.id] {
			t.Errorf("IDs() missing provider ID %q", p.id)
		}
	}
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	reg := NewRegistry()

	// Test concurrent registration
	var wg sync.WaitGroup
	numProviders := 10

	for i := 0; i < numProviders; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			p := &mockProvider{id: string(rune('a' + id))}
			reg.Register(p)
		}(i)
	}

	wg.Wait()

	// Verify all providers registered
	if len(reg.IDs()) != numProviders {
		t.Errorf("concurrent registration: got %d providers, want %d", len(reg.IDs()), numProviders)
	}

	// Test concurrent reads
	wg = sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			reg.Has("a")
			_, _ = reg.Get("a") // Ignore error in concurrent test
			reg.IDs()
		}()
	}

	wg.Wait()
}

func TestDefaultRegistry(t *testing.T) {
	p := &mockProvider{id: "default-test"}

	DefaultRegistry.Register(p)

	if !Has("default-test") {
		t.Error("Has() = false for provider in DefaultRegistry, want true")
	}

	got, err := Get("default-test")
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}

	if got.ID() != "default-test" {
		t.Errorf("Get() returned provider with ID %q, want %q", got.ID(), "default-test")
	}
}

func TestDefaultRegistry_Register(t *testing.T) {
	p := &mockProvider{id: "global-test"}

	Register(p)

	if !Has("global-test") {
		t.Error("Has() = false after Register(), want true")
	}
}

func TestDefaultRegistry_Get(t *testing.T) {
	p := &mockProvider{id: "global-get-test"}
	Register(p)

	got, err := Get("global-get-test")
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}

	if got.ID() != "global-get-test" {
		t.Errorf("Get() returned provider with ID %q, want %q", got.ID(), "global-get-test")
	}
}

func TestDefaultRegistry_Has(t *testing.T) {
	p := &mockProvider{id: "global-has-test"}
	Register(p)

	if !Has("global-has-test") {
		t.Error("Has() = false for registered provider, want true")
	}

	if Has("not-registered") {
		t.Error("Has() = true for unregistered provider, want false")
	}
}

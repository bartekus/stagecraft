// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package secrets

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

// Feature: PROVIDER_SECRETS_INTERFACE
// Spec: spec/providers/secrets/interface.md

// resetDefaultRegistry resets the global registry for testing.
// This prevents test pollution when tests run in parallel.
func resetDefaultRegistry() {
	DefaultRegistry = NewRegistry()
}

// mockProvider is a test implementation of SecretsProvider.
type mockProvider struct {
	id string
}

func (m *mockProvider) ID() string {
	return m.id
}

func (m *mockProvider) Sync(ctx context.Context, opts SyncOptions) error {
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
			p := &mockProvider{id: fmt.Sprintf("p-%d", id)}
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
			reg.Has("p-0")
			_, _ = reg.Get("p-0") // Ignore error in concurrent test
			reg.IDs()
		}()
	}

	wg.Wait()
}

func TestDefaultRegistry(t *testing.T) {
	resetDefaultRegistry()

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
	resetDefaultRegistry()

	p := &mockProvider{id: "global-test"}

	Register(p)

	if !Has("global-test") {
		t.Error("Has() = false after Register(), want true")
	}
}

func TestDefaultRegistry_Get(t *testing.T) {
	resetDefaultRegistry()

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
	resetDefaultRegistry()

	p := &mockProvider{id: "global-has-test"}
	Register(p)

	if !Has("global-has-test") {
		t.Error("Has() = false for registered provider, want true")
	}

	if Has("not-registered") {
		t.Error("Has() = true for unregistered provider, want false")
	}
}

func TestRegistry_List(t *testing.T) {
	reg := NewRegistry()

	p1 := &mockProvider{id: "b-provider"}
	p2 := &mockProvider{id: "a-provider"}
	p3 := &mockProvider{id: "c-provider"}

	reg.Register(p1)
	reg.Register(p2)
	reg.Register(p3)

	list := reg.List()
	if len(list) != 3 {
		t.Fatalf("List() length = %d, want 3", len(list))
	}

	if list[0].ID() != "a-provider" || list[1].ID() != "b-provider" || list[2].ID() != "c-provider" {
		t.Errorf("List() order = [%s, %s, %s], want [a-provider, b-provider, c-provider]",
			list[0].ID(), list[1].ID(), list[2].ID())
	}
}

func TestDefaultRegistry_List(t *testing.T) {
	resetDefaultRegistry()

	p1 := &mockProvider{id: "z-provider"}
	p2 := &mockProvider{id: "a-provider"}

	Register(p1)
	Register(p2)

	list := List()
	if len(list) != 2 {
		t.Fatalf("List() length = %d, want 2", len(list))
	}
	if list[0].ID() != "a-provider" || list[1].ID() != "z-provider" {
		t.Errorf("List() order incorrect")
	}
}

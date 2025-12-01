// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package migration

import (
	"context"
	"sync"
	"testing"
)

// Feature: CORE_MIGRATION_REGISTRY
// Spec: spec/core/migration-registry.md

// mockEngine is a test implementation of Engine.
type mockEngine struct {
	id string
}

func (m *mockEngine) ID() string {
	return m.id
}

func (m *mockEngine) Plan(ctx context.Context, opts PlanOptions) ([]Migration, error) {
	return nil, nil
}

func (m *mockEngine) Run(ctx context.Context, opts RunOptions) error {
	return nil
}

func TestRegistry_Register(t *testing.T) {
	reg := NewRegistry()

	e1 := &mockEngine{id: "drizzle"}
	e2 := &mockEngine{id: "prisma"}

	reg.Register(e1)
	reg.Register(e2)

	// Verify both are registered
	if !reg.Has("drizzle") {
		t.Error("expected drizzle to be registered")
	}
	if !reg.Has("prisma") {
		t.Error("expected prisma to be registered")
	}
}

func TestRegistry_Register_PanicsOnEmptyID(t *testing.T) {
	reg := NewRegistry()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when registering engine with empty ID")
		}
	}()

	e := &mockEngine{id: ""}
	reg.Register(e)
}

func TestRegistry_Register_PanicsOnDuplicateID(t *testing.T) {
	reg := NewRegistry()

	e1 := &mockEngine{id: "duplicate"}
	e2 := &mockEngine{id: "duplicate"}

	reg.Register(e1)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when registering duplicate engine ID")
		}
	}()

	reg.Register(e2)
}

func TestRegistry_Get(t *testing.T) {
	reg := NewRegistry()

	e := &mockEngine{id: "drizzle"}
	reg.Register(e)

	got, err := reg.Get("drizzle")
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}

	if got.ID() != "drizzle" {
		t.Errorf("Get() returned engine with ID %q, want %q", got.ID(), "drizzle")
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

	if reg.Has("drizzle") {
		t.Error("Has() = true for unregistered engine, want false")
	}

	e := &mockEngine{id: "drizzle"}
	reg.Register(e)

	if !reg.Has("drizzle") {
		t.Error("Has() = false for registered engine, want true")
	}
}

func TestRegistry_IDs(t *testing.T) {
	reg := NewRegistry()

	// Empty registry should return empty slice
	ids := reg.IDs()
	if len(ids) != 0 {
		t.Errorf("IDs() length = %d, want 0", len(ids))
	}

	// Register multiple engines
	engines := []*mockEngine{
		{id: "drizzle"},
		{id: "prisma"},
		{id: "knex"},
		{id: "raw"},
	}

	for _, e := range engines {
		reg.Register(e)
	}

	ids = reg.IDs()
	if len(ids) != 4 {
		t.Errorf("IDs() length = %d, want 4", len(ids))
	}

	// Verify all IDs are present
	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[id] = true
	}

	for _, e := range engines {
		if !idMap[e.id] {
			t.Errorf("IDs() missing engine ID %q", e.id)
		}
	}
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	reg := NewRegistry()

	// Test concurrent registration
	var wg sync.WaitGroup
	numEngines := 10

	for i := 0; i < numEngines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			e := &mockEngine{id: string(rune('a' + id))}
			reg.Register(e)
		}(i)
	}

	wg.Wait()

	// Verify all engines registered
	if len(reg.IDs()) != numEngines {
		t.Errorf("concurrent registration: got %d engines, want %d", len(reg.IDs()), numEngines)
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
	e := &mockEngine{id: "default-test"}

	// This would normally be called in init(), but for testing we call directly
	DefaultRegistry.Register(e)

	if !Has("default-test") {
		t.Error("Has() = false for engine in DefaultRegistry, want true")
	}

	got, err := Get("default-test")
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}

	if got.ID() != "default-test" {
		t.Errorf("Get() returned engine with ID %q, want %q", got.ID(), "default-test")
	}
}

func TestDefaultRegistry_Register(t *testing.T) {
	e := &mockEngine{id: "global-test"}

	Register(e)

	if !Has("global-test") {
		t.Error("Has() = false after Register(), want true")
	}
}

func TestDefaultRegistry_Get(t *testing.T) {
	e := &mockEngine{id: "global-get-test"}
	Register(e)

	got, err := Get("global-get-test")
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}

	if got.ID() != "global-get-test" {
		t.Errorf("Get() returned engine with ID %q, want %q", got.ID(), "global-get-test")
	}
}

func TestDefaultRegistry_Has(t *testing.T) {
	e := &mockEngine{id: "global-has-test"}
	Register(e)

	if !Has("global-has-test") {
		t.Error("Has() = false for registered engine, want true")
	}

	if Has("not-registered") {
		t.Error("Has() = true for unregistered engine, want false")
	}
}

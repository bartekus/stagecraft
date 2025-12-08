// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package backend

import (
	"context"
	"testing"
)

// Feature: PROVIDER_BACKEND_INTERFACE
// Spec: spec/core/backend-registry.md

// testProvider is a test implementation of BackendProvider for interface testing.
type testProvider struct {
	id string
}

func (t *testProvider) ID() string {
	return t.id
}

func (t *testProvider) Dev(ctx context.Context, opts DevOptions) error {
	return nil
}

func (t *testProvider) BuildDocker(ctx context.Context, opts BuildDockerOptions) (string, error) {
	return "test-image:tag", nil
}

func (t *testProvider) Plan(ctx context.Context, opts PlanOptions) (ProviderPlan, error) {
	return ProviderPlan{
		Provider: t.id,
		Steps: []ProviderStep{
			{Name: "Build", Description: "Build Docker image"},
		},
	}, nil
}

// TestBackendProvider_Interface verifies that BackendProvider interface can be implemented
// and that the interface contract is satisfied at compile time.
func TestBackendProvider_Interface(t *testing.T) {
	var provider BackendProvider = &testProvider{id: "test-provider"}

	if provider.ID() != "test-provider" {
		t.Errorf("ID() = %q, want %q", provider.ID(), "test-provider")
	}

	// Verify all interface methods can be called
	ctx := context.Background()

	err := provider.Dev(ctx, DevOptions{})
	if err != nil {
		t.Errorf("Dev() error = %v, want nil", err)
	}

	imageTag, err := provider.BuildDocker(ctx, BuildDockerOptions{})
	if err != nil {
		t.Errorf("BuildDocker() error = %v, want nil", err)
	}
	if imageTag == "" {
		t.Error("BuildDocker() returned empty image tag")
	}

	plan, err := provider.Plan(ctx, PlanOptions{})
	if err != nil {
		t.Errorf("Plan() error = %v, want nil", err)
	}
	if plan.Provider != "test-provider" {
		t.Errorf("Plan().Provider = %q, want %q", plan.Provider, "test-provider")
	}
	if len(plan.Steps) == 0 {
		t.Error("Plan().Steps is empty, want at least one step")
	}
}

// TestBackendProvider_Types verifies that the option types are properly defined.
func TestBackendProvider_Types(t *testing.T) {
	// Verify DevOptions can be constructed
	devOpts := DevOptions{
		Config:  map[string]any{"key": "value"},
		WorkDir: "/tmp/test",
		Env:     map[string]string{"VAR": "value"},
	}

	if devOpts.WorkDir == "" {
		t.Error("DevOptions.WorkDir should be settable")
	}

	// Verify BuildDockerOptions can be constructed
	buildOpts := BuildDockerOptions{
		Config:   map[string]any{"key": "value"},
		ImageTag: "test:tag",
		WorkDir:  "/tmp/test",
	}

	if buildOpts.ImageTag == "" {
		t.Error("BuildDockerOptions.ImageTag should be settable")
	}

	// Verify PlanOptions can be constructed
	planOpts := PlanOptions{
		Config:   map[string]any{"key": "value"},
		ImageTag: "test:tag",
		WorkDir:  "/tmp/test",
	}

	if planOpts.ImageTag == "" {
		t.Error("PlanOptions.ImageTag should be settable")
	}

	// Verify ProviderStep can be constructed
	step := ProviderStep{
		Name:        "TestStep",
		Description: "Test description",
	}

	if step.Name == "" {
		t.Error("ProviderStep.Name should be settable")
	}

	// Verify ProviderPlan can be constructed
	plan := ProviderPlan{
		Provider: "test-provider",
		Steps:    []ProviderStep{step},
	}

	if plan.Provider == "" {
		t.Error("ProviderPlan.Provider should be settable")
	}
	if len(plan.Steps) == 0 {
		t.Error("ProviderPlan.Steps should be settable")
	}
}

// metadataTestProvider is a test implementation that implements both BackendProvider and MetadataProvider.
type metadataTestProvider struct {
	testProvider
	metadata ProviderMetadata
}

// Metadata implements the MetadataProvider interface.
func (m *metadataTestProvider) Metadata() ProviderMetadata {
	return m.metadata
}

// TestMetadataProvider_Interface verifies that MetadataProvider interface can be implemented.
func TestMetadataProvider_Interface(t *testing.T) {
	metadataProvider := &metadataTestProvider{
		testProvider: testProvider{id: "test-metadata"},
		metadata: ProviderMetadata{
			Name:        "Test Provider",
			Description: "Test metadata provider",
			Version:     "1.0.0",
		},
	}

	// Verify it implements BackendProvider
	var bp BackendProvider = metadataProvider
	if bp.ID() != "test-metadata" {
		t.Errorf("BackendProvider.ID() = %q, want %q", bp.ID(), "test-metadata")
	}

	// Verify it implements MetadataProvider
	var mp MetadataProvider = metadataProvider
	_ = mp // Verify assignment compiles (proves interface is satisfied)

	// Verify Metadata() can be called
	meta := mp.Metadata()
	if meta.Name == "" {
		t.Error("Metadata() should return non-empty metadata")
	}
}

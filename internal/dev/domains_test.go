// SPDX-License-Identifier: AGPL-3.0-or-later

package dev

import (
	"testing"

	"stagecraft/pkg/config"
)

// Feature: CLI_DEV
// Spec: spec/commands/dev.md

func TestComputeDomains_Defaults(t *testing.T) {
	t.Helper()

	cfg := &config.Config{}

	domains, err := ComputeDomains(cfg, "dev")
	if err != nil {
		t.Fatalf("ComputeDomains() error = %v, want nil", err)
	}

	if domains.Frontend != defaultFrontendDomain {
		t.Errorf("ComputeDomains() Frontend = %q, want %q", domains.Frontend, defaultFrontendDomain)
	}

	if domains.Backend != defaultBackendDomain {
		t.Errorf("ComputeDomains() Backend = %q, want %q", domains.Backend, defaultBackendDomain)
	}
}

func TestComputeDomains_PartialOverride(t *testing.T) {
	t.Helper()

	cfg := &config.Config{
		Dev: &config.DevConfig{
			Domains: &config.DevDomains{
				Frontend: "frontend.example.test",
				// Backend not set, should use default
			},
		},
	}

	domains, err := ComputeDomains(cfg, "dev")
	if err != nil {
		t.Fatalf("ComputeDomains() error = %v, want nil", err)
	}

	if domains.Frontend != "frontend.example.test" {
		t.Errorf("ComputeDomains() Frontend = %q, want %q", domains.Frontend, "frontend.example.test")
	}

	if domains.Backend != defaultBackendDomain {
		t.Errorf("ComputeDomains() Backend = %q, want %q", domains.Backend, defaultBackendDomain)
	}
}

func TestComputeDomains_FullOverride(t *testing.T) {
	t.Helper()

	cfg := &config.Config{
		Dev: &config.DevConfig{
			Domains: &config.DevDomains{
				Frontend: "app.example.test",
				Backend:  "api.example.test",
			},
		},
	}

	domains, err := ComputeDomains(cfg, "dev")
	if err != nil {
		t.Fatalf("ComputeDomains() error = %v, want nil", err)
	}

	if domains.Frontend != "app.example.test" {
		t.Errorf("ComputeDomains() Frontend = %q, want %q", domains.Frontend, "app.example.test")
	}

	if domains.Backend != "api.example.test" {
		t.Errorf("ComputeDomains() Backend = %q, want %q", domains.Backend, "api.example.test")
	}
}

func TestComputeDomains_BackendOnlyOverride(t *testing.T) {
	t.Helper()

	cfg := &config.Config{
		Dev: &config.DevConfig{
			Domains: &config.DevDomains{
				// Frontend not set, should use default
				Backend: "api.example.test",
			},
		},
	}

	domains, err := ComputeDomains(cfg, "dev")
	if err != nil {
		t.Fatalf("ComputeDomains() error = %v, want nil", err)
	}

	if domains.Frontend != defaultFrontendDomain {
		t.Errorf("ComputeDomains() Frontend = %q, want %q", domains.Frontend, defaultFrontendDomain)
	}

	if domains.Backend != "api.example.test" {
		t.Errorf("ComputeDomains() Backend = %q, want %q", domains.Backend, "api.example.test")
	}
}

func TestComputeDomains_EmptyStringsUseDefaults(t *testing.T) {
	t.Helper()

	cfg := &config.Config{
		Dev: &config.DevConfig{
			Domains: &config.DevDomains{
				Frontend: "", // Empty string should use default
				Backend:  "", // Empty string should use default
			},
		},
	}

	domains, err := ComputeDomains(cfg, "dev")
	if err != nil {
		t.Fatalf("ComputeDomains() error = %v, want nil", err)
	}

	if domains.Frontend != defaultFrontendDomain {
		t.Errorf("ComputeDomains() Frontend = %q, want %q (empty string should use default)", domains.Frontend, defaultFrontendDomain)
	}

	if domains.Backend != defaultBackendDomain {
		t.Errorf("ComputeDomains() Backend = %q, want %q (empty string should use default)", domains.Backend, defaultBackendDomain)
	}
}

func TestComputeDomains_EnvParameterAccepted(t *testing.T) {
	t.Helper()

	// Test that env parameter doesn't cause issues even though it's unused in v1
	cfg := &config.Config{}

	domains1, err1 := ComputeDomains(cfg, "dev")
	if err1 != nil {
		t.Fatalf("ComputeDomains(dev) error = %v, want nil", err1)
	}

	domains2, err2 := ComputeDomains(cfg, "prod")
	if err2 != nil {
		t.Fatalf("ComputeDomains(prod) error = %v, want nil", err2)
	}

	// Both should return same defaults (env not used in v1)
	if domains1.Frontend != domains2.Frontend {
		t.Errorf("ComputeDomains() should return same domains for different env in v1, got Frontend %q vs %q", domains1.Frontend, domains2.Frontend)
	}

	if domains1.Backend != domains2.Backend {
		t.Errorf("ComputeDomains() should return same domains for different env in v1, got Backend %q vs %q", domains1.Backend, domains2.Backend)
	}
}

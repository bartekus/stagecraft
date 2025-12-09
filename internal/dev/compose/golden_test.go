// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package compose

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"stagecraft/pkg/config"
)

// Feature: DEV_COMPOSE_INFRA
// Spec: spec/dev/compose-infra.md

func TestGenerateCompose_Golden_BackendFrontendTraefik(t *testing.T) {
	t.Helper()

	cfg := &config.Config{}
	backend := &ServiceDefinition{
		Name: "backend",
		Ports: []PortMapping{
			{Host: "8080", Container: "3000", Protocol: "tcp"},
			{Host: "9090", Container: "4000", Protocol: "tcp"},
		},
		Environment: map[string]string{
			"B": "2",
			"A": "1",
		},
	}
	frontend := &ServiceDefinition{
		Name: "frontend",
		Ports: []PortMapping{
			{Host: "3000", Container: "3000", Protocol: "tcp"},
		},
		Environment: map[string]string{
			"B": "2",
			"A": "1",
		},
	}
	traefik := &ServiceDefinition{
		Name: "traefik",
	}

	gen := NewGenerator()

	composeFile, err := gen.GenerateCompose(cfg, backend, frontend, traefik)
	if err != nil {
		t.Fatalf("GenerateCompose() error = %v, want nil", err)
	}

	if composeFile == nil {
		t.Fatalf("GenerateCompose() got = nil, want non-nil *ComposeFile")
	}

	gotYAML, err := composeFile.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML() error = %v, want nil", err)
	}

	goldenPath := filepath.Join("testdata", "dev_compose_backend_frontend_traefik.yaml")

	// #nosec G304 -- test file path is controlled
	wantYAML, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("failed to read golden file %q: %v", goldenPath, err)
	}

	if !bytes.Equal(gotYAML, wantYAML) {
		t.Fatalf("generated compose YAML does not match golden file\n\n=== got ===\n%s\n\n=== want ===\n%s", gotYAML, wantYAML)
	}
}

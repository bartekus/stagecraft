// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package core

import (
	"os"
	"path/filepath"
	"testing"

	"stagecraft/pkg/config"
)

// Feature: CORE_PLAN
// Spec: spec/core/plan.md

func TestPlanner_PlanDeploy(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")

	content := []byte(`
project:
  name: test-app
backend:
  provider: generic
  providers:
    generic:
      dev:
        command: ["npm", "run", "dev"]
environments:
  dev:
    driver: local
  prod:
    driver: digitalocean
databases:
  main:
    connection_env: DATABASE_URL
    migrations:
      engine: raw
      path: ./migrations
      strategy: pre_deploy
`)

	if err := os.WriteFile(configPath, content, 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	planner := NewPlanner(cfg)
	plan, err := planner.PlanDeploy("prod")
	if err != nil {
		t.Fatalf("expected no error planning deployment, got: %v", err)
	}

	if plan.Environment != "prod" {
		t.Fatalf("expected environment 'prod', got %q", plan.Environment)
	}

	if len(plan.Operations) == 0 {
		t.Fatalf("expected at least one operation in plan")
	}

	// Check for migration operation
	foundMigration := false
	for _, op := range plan.Operations {
		if op.Type == OpTypeMigration {
			foundMigration = true
			if op.Metadata["database"] != "main" {
				t.Errorf("expected migration for database 'main', got %v", op.Metadata["database"])
			}
		}
	}

	if !foundMigration {
		t.Errorf("expected to find migration operation in plan")
	}
}

func TestPlanner_PlanDeploy_UnknownEnvironment(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")

	content := []byte(`
project:
  name: test-app
environments:
  dev:
    driver: local
`)

	if err := os.WriteFile(configPath, content, 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	planner := NewPlanner(cfg)
	_, err = planner.PlanDeploy("unknown")
	if err == nil {
		t.Fatalf("expected error for unknown environment")
	}
}

func TestPlanner_PlanDeploy_IncludesAllOperationTypes(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stagecraft.yml")

	content := []byte(`
project:
  name: test-app
backend:
  provider: generic
  providers:
    generic:
      dev:
        command: ["npm", "run", "dev"]
environments:
  prod:
    driver: digitalocean
databases:
  main:
    connection_env: DATABASE_URL
    migrations:
      engine: raw
      path: ./migrations
      strategy: pre_deploy
`)

	if err := os.WriteFile(configPath, content, 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	planner := NewPlanner(cfg)
	plan, err := planner.PlanDeploy("prod")
	if err != nil {
		t.Fatalf("expected no error planning deployment, got: %v", err)
	}

	// Check that we have expected operation types
	opTypes := make(map[OperationType]bool)
	for _, op := range plan.Operations {
		opTypes[op.Type] = true
	}

	if !opTypes[OpTypeMigration] {
		t.Errorf("expected migration operation in plan")
	}
	if !opTypes[OpTypeBuild] {
		t.Errorf("expected build operation in plan")
	}
	if !opTypes[OpTypeDeploy] {
		t.Errorf("expected deploy operation in plan")
	}
	if !opTypes[OpTypeHealthCheck] {
		t.Errorf("expected health check operation in plan")
	}
}

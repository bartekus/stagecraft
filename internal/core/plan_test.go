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

func TestPlanner_PlanDeploy_DeterministicOperationOrder(t *testing.T) {
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
  z_database:
    connection_env: DATABASE_URL
    migrations:
      engine: raw
      path: ./migrations
      strategy: pre_deploy
  a_database:
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

	// Generate plan twice
	plan1, err := planner.PlanDeploy("prod")
	if err != nil {
		t.Fatalf("expected no error planning deployment, got: %v", err)
	}

	plan2, err := planner.PlanDeploy("prod")
	if err != nil {
		t.Fatalf("expected no error planning deployment, got: %v", err)
	}

	// Operation order must be deterministic (sorted by database name)
	if len(plan1.Operations) != len(plan2.Operations) {
		t.Fatalf("expected same number of operations, got %d and %d", len(plan1.Operations), len(plan2.Operations))
	}

	// Verify migration operations are in sorted order (a_database before z_database)
	var migrationOps []Operation
	for _, op := range plan1.Operations {
		if op.Type == OpTypeMigration {
			strategy, ok := op.Metadata["strategy"].(string)
			if ok && strategy == "pre_deploy" {
				migrationOps = append(migrationOps, op)
			}
		}
	}

	if len(migrationOps) != 2 {
		t.Fatalf("expected 2 pre_deploy migration operations, got %d", len(migrationOps))
	}

	// First migration should be a_database (alphabetically first)
	if migrationOps[0].ID != "migration_a_database_pre_deploy" {
		t.Errorf("expected first migration ID 'migration_a_database_pre_deploy', got %q", migrationOps[0].ID)
	}
	if migrationOps[1].ID != "migration_z_database_pre_deploy" {
		t.Errorf("expected second migration ID 'migration_z_database_pre_deploy', got %q", migrationOps[1].ID)
	}

	// Verify operation IDs match across runs
	for i := range plan1.Operations {
		if plan1.Operations[i].ID != plan2.Operations[i].ID {
			t.Errorf("operation %d: expected ID %q, got %q", i, plan1.Operations[i].ID, plan2.Operations[i].ID)
		}
	}
}

func TestPlanner_PlanDeploy_WiresDependencies(t *testing.T) {
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

	// Find deploy operation
	var deployOp *Operation
	for i := range plan.Operations {
		if plan.Operations[i].Type == OpTypeDeploy {
			deployOp = &plan.Operations[i]
			break
		}
	}

	if deployOp == nil {
		t.Fatal("expected deploy operation in plan")
	}

	// Deploy should depend on build_backend and pre_deploy migration
	if len(deployOp.Dependencies) == 0 {
		t.Error("expected deploy operation to have dependencies")
	}

	hasBuildDep := false
	hasMigrationDep := false
	for _, dep := range deployOp.Dependencies {
		if dep == "build_backend" {
			hasBuildDep = true
		}
		if dep == "migration_main_pre_deploy" {
			hasMigrationDep = true
		}
	}

	if !hasBuildDep {
		t.Error("expected deploy to depend on build_backend")
	}
	if !hasMigrationDep {
		t.Error("expected deploy to depend on migration_main_pre_deploy")
	}

	// Find health check operation
	var healthCheckOp *Operation
	for i := range plan.Operations {
		if plan.Operations[i].Type == OpTypeHealthCheck {
			healthCheckOp = &plan.Operations[i]
			break
		}
	}

	if healthCheckOp == nil {
		t.Fatal("expected health check operation in plan")
	}

	// Health check should depend on deploy
	if len(healthCheckOp.Dependencies) != 1 {
		t.Errorf("expected health check to have 1 dependency, got %d", len(healthCheckOp.Dependencies))
	}
	if healthCheckOp.Dependencies[0] != "deploy_prod" {
		t.Errorf("expected health check to depend on deploy_prod, got %q", healthCheckOp.Dependencies[0])
	}
}

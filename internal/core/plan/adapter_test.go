// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

package plan

import (
	"encoding/json"
	"fmt"
	"testing"

	"stagecraft/internal/core"
	"stagecraft/pkg/engine"
)

func TestToEnginePlan_DeterministicPlanID(t *testing.T) {
	// Create a plan with known operations
	corePlan := &core.Plan{
		Environment: "prod",
		Operations: []core.Operation{
			{
				ID:           "migration_main_pre_deploy",
				Type:         core.OpTypeMigration,
				Description:  "Run pre_deploy migrations",
				Dependencies: []string{},
				Metadata: map[string]interface{}{
					"database": "main",
					"strategy": "pre_deploy",
					"engine":   "raw",
					"path":     "./migrations",
				},
			},
			{
				ID:           "build_backend",
				Type:         core.OpTypeBuild,
				Description:  "Build backend",
				Dependencies: []string{},
				Metadata: map[string]interface{}{
					"provider": "generic",
				},
			},
		},
	}

	// Convert twice
	plan1, err := ToEnginePlan(corePlan, "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	plan2, err := ToEnginePlan(corePlan, "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Plan IDs must be identical
	if plan1.ID != plan2.ID {
		t.Errorf("plan IDs must be deterministic: got %q and %q", plan1.ID, plan2.ID)
	}

	// Plan IDs must be non-empty
	if plan1.ID == "" {
		t.Error("plan ID must not be empty")
	}

	// Plan ID should be 24 hex characters (12 bytes)
	if len(plan1.ID) != 24 {
		t.Errorf("plan ID should be 24 hex characters, got %d: %q", len(plan1.ID), plan1.ID)
	}
}

func TestToEnginePlan_StableStepOrdering(t *testing.T) {
	corePlan := &core.Plan{
		Environment: "prod",
		Operations: []core.Operation{
			{
				ID:          "build_backend",
				Type:        core.OpTypeBuild,
				Description: "Build backend",
				Metadata:    map[string]interface{}{"provider": "generic"},
			},
			{
				ID:          "deploy_prod",
				Type:        core.OpTypeDeploy,
				Description: "Deploy",
				Metadata:    map[string]interface{}{"environment": "prod"},
			},
		},
	}

	plan, err := ToEnginePlan(corePlan, "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Steps must be ordered by Index
	if len(plan.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(plan.Steps))
	}

	if plan.Steps[0].Index != 0 {
		t.Errorf("expected first step Index=0, got %d", plan.Steps[0].Index)
	}
	if plan.Steps[1].Index != 1 {
		t.Errorf("expected second step Index=1, got %d", plan.Steps[1].Index)
	}

	// Step IDs should match Operation IDs
	if plan.Steps[0].ID != "build_backend" {
		t.Errorf("expected step ID 'build_backend', got %q", plan.Steps[0].ID)
	}
	if plan.Steps[1].ID != "deploy_prod" {
		t.Errorf("expected step ID 'deploy_prod', got %q", plan.Steps[1].ID)
	}
}

func TestToEnginePlan_HostAssignment(t *testing.T) {
	corePlan := &core.Plan{
		Environment: "prod",
		Operations: []core.Operation{
			{
				ID:       "build_backend",
				Type:     core.OpTypeBuild,
				Metadata: map[string]interface{}{},
			},
		},
	}

	plan, err := ToEnginePlan(corePlan, "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All steps must have host assigned (even if "local" for v1)
	if len(plan.Steps) == 0 {
		t.Fatal("expected at least one step")
	}

	for i, step := range plan.Steps {
		if step.Host.LogicalID != "local" {
			t.Errorf("step %d: expected host LogicalID='local', got %q", i, step.Host.LogicalID)
		}
	}
}

func TestToEnginePlan_ActionMapping(t *testing.T) {
	tests := []struct {
		name           string
		opType         core.OperationType
		expectedAction engine.StepAction
	}{
		{
			name:           "build maps to build",
			opType:         core.OpTypeBuild,
			expectedAction: engine.StepActionBuild,
		},
		{
			name:           "deploy maps to apply_compose",
			opType:         core.OpTypeDeploy,
			expectedAction: engine.StepActionApplyCompose,
		},
		{
			name:           "migration maps to migrate",
			opType:         core.OpTypeMigration,
			expectedAction: engine.StepActionMigrate,
		},
		{
			name:           "health_check maps to health_check",
			opType:         core.OpTypeHealthCheck,
			expectedAction: engine.StepActionHealthCheck,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			corePlan := &core.Plan{
				Environment: "prod",
				Operations: []core.Operation{
					{
						ID:       fmt.Sprintf("op_%s", tt.opType),
						Type:     tt.opType,
						Metadata: map[string]interface{}{},
					},
				},
			}

			plan, err := ToEnginePlan(corePlan, "prod")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(plan.Steps) != 1 {
				t.Fatalf("expected 1 step, got %d", len(plan.Steps))
			}

			if plan.Steps[0].Action != tt.expectedAction {
				t.Errorf("expected action %q, got %q", tt.expectedAction, plan.Steps[0].Action)
			}
		})
	}
}

func TestToEnginePlan_InputsAreDeterministicJSON(t *testing.T) {
	corePlan := &core.Plan{
		Environment: "prod",
		Operations: []core.Operation{
			{
				ID:   "migration_main_pre_deploy",
				Type: core.OpTypeMigration,
				Metadata: map[string]interface{}{
					"database": "main",
					"strategy": "pre_deploy",
					"engine":   "raw",
					"path":     "./migrations",
					"conn_env": "DATABASE_URL",
				},
			},
		},
	}

	plan1, err := ToEnginePlan(corePlan, "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	plan2, err := ToEnginePlan(corePlan, "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Inputs must be identical byte-for-byte
	if len(plan1.Steps) != 1 || len(plan2.Steps) != 1 {
		t.Fatal("expected 1 step in each plan")
	}

	inputs1 := plan1.Steps[0].Inputs
	inputs2 := plan2.Steps[0].Inputs

	if string(inputs1) != string(inputs2) {
		t.Errorf("inputs must be deterministic:\n  plan1: %s\n  plan2: %s", string(inputs1), string(inputs2))
	}

	// Verify inputs are valid JSON
	var inputsObj map[string]interface{}
	if err := json.Unmarshal(inputs1, &inputsObj); err != nil {
		t.Errorf("inputs must be valid JSON: %v", err)
	}

	// Verify expected fields are present
	if inputsObj["database"] != "main" {
		t.Errorf("expected database='main', got %v", inputsObj["database"])
	}
	if inputsObj["strategy"] != "pre_deploy" {
		t.Errorf("expected strategy='pre_deploy', got %v", inputsObj["strategy"])
	}
}

func TestToEnginePlan_NoTimestampsOrRandomness(t *testing.T) {
	corePlan := &core.Plan{
		Environment: "prod",
		Operations: []core.Operation{
			{
				ID:       "build_backend",
				Type:     core.OpTypeBuild,
				Metadata: map[string]interface{}{"provider": "generic"},
			},
		},
	}

	plan1, err := ToEnginePlan(corePlan, "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	plan2, err := ToEnginePlan(corePlan, "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Marshal both plans to JSON and compare
	json1, err := json.Marshal(plan1)
	if err != nil {
		t.Fatalf("marshaling plan1: %v", err)
	}

	json2, err := json.Marshal(plan2)
	if err != nil {
		t.Fatalf("marshaling plan2: %v", err)
	}

	if string(json1) != string(json2) {
		t.Errorf("plans must be identical (no timestamps/randomness):\n  plan1: %s\n  plan2: %s", string(json1), string(json2))
	}
}

func TestToEnginePlan_ResourceKindMapping(t *testing.T) {
	tests := []struct {
		name         string
		opType       core.OperationType
		expectedKind string
	}{
		{
			name:         "build maps to image",
			opType:       core.OpTypeBuild,
			expectedKind: "image",
		},
		{
			name:         "deploy maps to service",
			opType:       core.OpTypeDeploy,
			expectedKind: "service",
		},
		{
			name:         "migration maps to database",
			opType:       core.OpTypeMigration,
			expectedKind: "database",
		},
		{
			name:         "health_check maps to service",
			opType:       core.OpTypeHealthCheck,
			expectedKind: "service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			corePlan := &core.Plan{
				Environment: "prod",
				Operations: []core.Operation{
					{
						ID:       fmt.Sprintf("op_%s", tt.opType),
						Type:     tt.opType,
						Metadata: map[string]interface{}{},
					},
				},
			}

			plan, err := ToEnginePlan(corePlan, "prod")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(plan.Steps) != 1 {
				t.Fatalf("expected 1 step, got %d", len(plan.Steps))
			}

			if plan.Steps[0].Target.Kind != tt.expectedKind {
				t.Errorf("expected resource kind %q, got %q", tt.expectedKind, plan.Steps[0].Target.Kind)
			}
		})
	}
}

func TestToEnginePlan_DuplicateOperationIDsError(t *testing.T) {
	corePlan := &core.Plan{
		Environment: "prod",
		Operations: []core.Operation{
			{
				ID:       "dup",
				Type:     core.OpTypeBuild,
				Metadata: map[string]interface{}{},
			},
			{
				ID:       "dup",
				Type:     core.OpTypeDeploy,
				Metadata: map[string]interface{}{},
			},
		},
	}

	_, err := ToEnginePlan(corePlan, "prod")
	if err == nil {
		t.Fatal("expected error for duplicate operation ids")
	}

	if err.Error() != `duplicate operation id: "dup"` {
		t.Errorf("expected duplicate ID error, got: %v", err)
	}
}

func TestToEnginePlan_EmptyOperationIDError(t *testing.T) {
	corePlan := &core.Plan{
		Environment: "prod",
		Operations: []core.Operation{
			{
				// ID intentionally empty
				Type:     core.OpTypeBuild,
				Metadata: map[string]interface{}{},
			},
		},
	}

	_, err := ToEnginePlan(corePlan, "prod")
	if err == nil {
		t.Fatal("expected error for empty operation id")
	}

	if err.Error() != "operation id is empty at index 0 (planner bug)" {
		t.Errorf("expected empty ID error, got: %v", err)
	}
}

func TestToEnginePlan_PreservesDependencyIDs(t *testing.T) {
	corePlan := &core.Plan{
		Environment: "prod",
		Operations: []core.Operation{
			{
				ID:           "build_backend",
				Type:         core.OpTypeBuild,
				Dependencies: []string{},
				Metadata:     map[string]interface{}{"provider": "generic"},
			},
			{
				ID:           "deploy_prod",
				Type:         core.OpTypeDeploy,
				Dependencies: []string{"build_backend"}, // References Operation.ID
				Metadata:     map[string]interface{}{"environment": "prod"},
			},
		},
	}

	plan, err := ToEnginePlan(corePlan, "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(plan.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(plan.Steps))
	}

	// Step IDs should match Operation IDs
	if plan.Steps[0].ID != "build_backend" {
		t.Errorf("expected step ID 'build_backend', got %q", plan.Steps[0].ID)
	}
	if plan.Steps[1].ID != "deploy_prod" {
		t.Errorf("expected step ID 'deploy_prod', got %q", plan.Steps[1].ID)
	}

	// Dependencies should map directly (no remapping needed)
	if len(plan.Steps[1].DependsOn) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(plan.Steps[1].DependsOn))
	}
	if plan.Steps[1].DependsOn[0] != "build_backend" {
		t.Errorf("expected dependency 'build_backend', got %q", plan.Steps[1].DependsOn[0])
	}
}

func TestToEnginePlan_DependenciesWithMultiplePreDeployMigrations(t *testing.T) {
	corePlan := &core.Plan{
		Environment: "prod",
		Operations: []core.Operation{
			{
				ID:           "migration_main_pre_deploy",
				Type:         core.OpTypeMigration,
				Dependencies: []string{},
				Metadata:     map[string]interface{}{"database": "main", "strategy": "pre_deploy"},
			},
			{
				ID:           "migration_analytics_pre_deploy",
				Type:         core.OpTypeMigration,
				Dependencies: []string{},
				Metadata:     map[string]interface{}{"database": "analytics", "strategy": "pre_deploy"},
			},
			{
				ID:           "build_backend",
				Type:         core.OpTypeBuild,
				Dependencies: []string{},
				Metadata:     map[string]interface{}{},
			},
			{
				ID:           "deploy_prod",
				Type:         core.OpTypeDeploy,
				Dependencies: []string{"build_backend", "migration_main_pre_deploy", "migration_analytics_pre_deploy"},
				Metadata:     map[string]interface{}{"environment": "prod"},
			},
		},
	}

	plan, err := ToEnginePlan(corePlan, "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Find deploy step
	var deployStep *engine.PlanStep
	for i := range plan.Steps {
		if plan.Steps[i].ID == "deploy_prod" {
			deployStep = &plan.Steps[i]
			break
		}
	}

	if deployStep == nil {
		t.Fatal("expected to find deploy_prod step")
	}

	// Verify all dependencies are preserved
	expectedDeps := map[string]bool{
		"build_backend":                  true,
		"migration_main_pre_deploy":      true,
		"migration_analytics_pre_deploy": true,
	}

	if len(deployStep.DependsOn) != len(expectedDeps) {
		t.Fatalf("expected %d dependencies, got %d", len(expectedDeps), len(deployStep.DependsOn))
	}

	for _, dep := range deployStep.DependsOn {
		if !expectedDeps[dep] {
			t.Errorf("unexpected dependency: %q", dep)
		}
		delete(expectedDeps, dep)
	}

	if len(expectedDeps) > 0 {
		t.Errorf("missing dependencies: %v", expectedDeps)
	}
}

func TestToEnginePlan_SortsDependenciesDeterministically(t *testing.T) {
	// Test that dependencies are sorted even if planner produces unsorted list
	corePlan := &core.Plan{
		Environment: "prod",
		Operations: []core.Operation{
			{
				ID:           "migration_z_pre_deploy",
				Type:         core.OpTypeMigration,
				Dependencies: []string{},
				Metadata:     map[string]interface{}{},
			},
			{
				ID:           "migration_a_pre_deploy",
				Type:         core.OpTypeMigration,
				Dependencies: []string{},
				Metadata:     map[string]interface{}{},
			},
			{
				ID:           "build_backend",
				Type:         core.OpTypeBuild,
				Dependencies: []string{},
				Metadata:     map[string]interface{}{},
			},
			{
				ID:   "deploy_prod",
				Type: core.OpTypeDeploy,
				// Dependencies in non-sorted order (z before a, build in middle)
				Dependencies: []string{"migration_z_pre_deploy", "build_backend", "migration_a_pre_deploy"},
				Metadata:     map[string]interface{}{"environment": "prod"},
			},
		},
	}

	plan1, err := ToEnginePlan(corePlan, "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	plan2, err := ToEnginePlan(corePlan, "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Find deploy step
	var deployStep1, deployStep2 *engine.PlanStep
	for i := range plan1.Steps {
		if plan1.Steps[i].ID == "deploy_prod" {
			deployStep1 = &plan1.Steps[i]
			break
		}
	}
	for i := range plan2.Steps {
		if plan2.Steps[i].ID == "deploy_prod" {
			deployStep2 = &plan2.Steps[i]
			break
		}
	}

	if deployStep1 == nil || deployStep2 == nil {
		t.Fatal("expected to find deploy_prod step")
	}

	// Dependencies should be sorted deterministically
	expectedDeps := []string{"build_backend", "migration_a_pre_deploy", "migration_z_pre_deploy"}
	if len(deployStep1.DependsOn) != len(expectedDeps) {
		t.Fatalf("expected %d dependencies, got %d", len(expectedDeps), len(deployStep1.DependsOn))
	}

	for i, expected := range expectedDeps {
		if deployStep1.DependsOn[i] != expected {
			t.Errorf("dependency %d: expected %q, got %q", i, expected, deployStep1.DependsOn[i])
		}
	}

	// Both plans should have identical dependency lists
	if len(deployStep1.DependsOn) != len(deployStep2.DependsOn) {
		t.Fatalf("dependency lists should match: %v vs %v", deployStep1.DependsOn, deployStep2.DependsOn)
	}
	for i := range deployStep1.DependsOn {
		if deployStep1.DependsOn[i] != deployStep2.DependsOn[i] {
			t.Errorf("dependency %d mismatch: %q vs %q", i, deployStep1.DependsOn[i], deployStep2.DependsOn[i])
		}
	}
}

func TestToEnginePlan_NilPlan(t *testing.T) {
	_, err := ToEnginePlan(nil, "prod")
	if err == nil {
		t.Error("expected error for nil plan")
	}
}

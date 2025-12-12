// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

package engine

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestSlicePlan_RejectsCrossHostDependencies(t *testing.T) {
	plan := Plan{
		Version: PlanSchemaVersion,
		ID:      "test-plan",
		Steps: []PlanStep{
			{
				ID:     "step-host-a",
				Index:  0,
				Action: StepActionBuild,
				Host:   HostRef{LogicalID: "host-a"},
			},
			{
				ID:        "step-host-b",
				Index:     1,
				Action:    StepActionApplyCompose,
				Host:      HostRef{LogicalID: "host-b"},
				DependsOn: []string{"step-host-a"}, // Cross-host dependency!
			},
		},
	}

	_, err := SlicePlan(plan)
	if err == nil {
		t.Fatal("expected error for cross-host dependency")
	}

	expectedErr := "step \"step-host-b\" on host \"host-b\" depends on step \"step-host-a\" on host \"host-a\" (cross-host dependencies not allowed in v1)"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestSlicePlan_AllowsGlobalStepDependencies(t *testing.T) {
	plan := Plan{
		Version: PlanSchemaVersion,
		ID:      "test-plan",
		Steps: []PlanStep{
			{
				ID:     "global-step",
				Index:  0,
				Action: StepActionNoop,
				Host:   HostRef{LogicalID: ""}, // Global step
			},
			{
				ID:        "step-host-a",
				Index:     1,
				Action:    StepActionBuild,
				Host:      HostRef{LogicalID: "host-a"},
				DependsOn: []string{"global-step"}, // Dependency on global step is allowed
			},
		},
	}

	result, err := SlicePlan(plan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Global step should be in GlobalSteps
	if len(result.GlobalSteps) != 1 {
		t.Fatalf("expected 1 global step, got %d", len(result.GlobalSteps))
	}
	if result.GlobalSteps[0].ID != "global-step" {
		t.Errorf("expected global step ID 'global-step', got %q", result.GlobalSteps[0].ID)
	}

	// Host step should be in HostPlans, but dependency removed (controller handles global deps)
	hp, ok := result.HostPlans["host-a"]
	if !ok {
		t.Fatal("expected host plan for host-a")
	}
	if len(hp.Steps) != 1 {
		t.Fatalf("expected 1 step in host plan, got %d", len(hp.Steps))
	}
	// Dependency on global step should be removed from DependsOn (controller handles it)
	if len(hp.Steps[0].DependsOn) != 0 {
		t.Errorf("expected no dependencies in host plan (global deps handled by controller), got %v", hp.Steps[0].DependsOn)
	}

	// Global dependency should be tracked explicitly
	if len(result.GlobalStepIDs) != 1 {
		t.Fatalf("expected 1 global step ID, got %d", len(result.GlobalStepIDs))
	}
	if result.GlobalStepIDs[0] != "global-step" {
		t.Errorf("expected global step ID 'global-step', got %q", result.GlobalStepIDs[0])
	}

	// Global dependency refs should track the host step's dependency on global step
	if len(result.GlobalDependencyRefs) != 1 {
		t.Fatalf("expected 1 global dependency ref, got %d", len(result.GlobalDependencyRefs))
	}
	globalDeps, ok := result.GlobalDependencyRefs["step-host-a"]
	if !ok {
		t.Fatal("expected global dependency ref for step-host-a")
	}
	if len(globalDeps) != 1 || globalDeps[0] != "global-step" {
		t.Errorf("expected global dependency ['global-step'], got %v", globalDeps)
	}
}

func TestSlicePlan_GlobalStepsDeterministicOrdering(t *testing.T) {
	plan := Plan{
		Version: PlanSchemaVersion,
		ID:      "test-plan",
		Steps: []PlanStep{
			{
				ID:     "global-z",
				Index:  2,
				Action: StepActionNoop,
				Host:   HostRef{LogicalID: ""},
			},
			{
				ID:     "global-a",
				Index:  0,
				Action: StepActionNoop,
				Host:   HostRef{LogicalID: ""},
			},
			{
				ID:     "global-m",
				Index:  1,
				Action: StepActionNoop,
				Host:   HostRef{LogicalID: ""},
			},
		},
	}

	result1, err := SlicePlan(plan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result2, err := SlicePlan(plan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Global steps should be sorted by Index then ID
	if len(result1.GlobalSteps) != 3 {
		t.Fatalf("expected 3 global steps, got %d", len(result1.GlobalSteps))
	}

	expectedOrder := []string{"global-a", "global-m", "global-z"}
	for i, expectedID := range expectedOrder {
		if result1.GlobalSteps[i].ID != expectedID {
			t.Errorf("global step %d: expected ID %q, got %q", i, expectedID, result1.GlobalSteps[i].ID)
		}
		if result2.GlobalSteps[i].ID != expectedID {
			t.Errorf("global step %d (second run): expected ID %q, got %q", i, expectedID, result2.GlobalSteps[i].ID)
		}
	}

	// Results should be identical
	json1, _ := json.Marshal(result1.GlobalSteps)
	json2, _ := json.Marshal(result2.GlobalSteps)
	if !bytes.Equal(json1, json2) {
		t.Error("global steps ordering must be deterministic")
	}
}

func TestSlicePlan_PreservesLocalDependencies(t *testing.T) {
	plan := Plan{
		Version: PlanSchemaVersion,
		ID:      "test-plan",
		Steps: []PlanStep{
			{
				ID:     "step-1",
				Index:  0,
				Action: StepActionBuild,
				Host:   HostRef{LogicalID: "host-a"},
			},
			{
				ID:        "step-2",
				Index:     1,
				Action:    StepActionApplyCompose,
				Host:      HostRef{LogicalID: "host-a"},
				DependsOn: []string{"step-1"}, // Same host - should be preserved
			},
		},
	}

	result, err := SlicePlan(plan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hp, ok := result.HostPlans["host-a"]
	if !ok {
		t.Fatal("expected host plan for host-a")
	}

	if len(hp.Steps) != 2 {
		t.Fatalf("expected 2 steps in host plan, got %d", len(hp.Steps))
	}

	// Find step-2
	var step2 *HostPlanStep
	for i := range hp.Steps {
		if hp.Steps[i].ID == "step-2" {
			step2 = &hp.Steps[i]
			break
		}
	}

	if step2 == nil {
		t.Fatal("expected to find step-2")
	}

	if len(step2.DependsOn) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(step2.DependsOn))
	}
	if step2.DependsOn[0] != "step-1" {
		t.Errorf("expected dependency 'step-1', got %q", step2.DependsOn[0])
	}
}

func TestSlicePlan_SortsDependenciesDeterministically(t *testing.T) {
	plan := Plan{
		Version: PlanSchemaVersion,
		ID:      "test-plan",
		Steps: []PlanStep{
			{
				ID:     "step-z",
				Index:  0,
				Action: StepActionBuild,
				Host:   HostRef{LogicalID: "host-a"},
			},
			{
				ID:     "step-a",
				Index:  1,
				Action: StepActionBuild,
				Host:   HostRef{LogicalID: "host-a"},
			},
			{
				ID:     "step-m",
				Index:  2,
				Action: StepActionBuild,
				Host:   HostRef{LogicalID: "host-a"},
			},
			{
				ID:        "step-final",
				Index:     3,
				Action:    StepActionApplyCompose,
				Host:      HostRef{LogicalID: "host-a"},
				DependsOn: []string{"step-z", "step-m", "step-a"}, // Unsorted
			},
		},
	}

	result, err := SlicePlan(plan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hp := result.HostPlans["host-a"]
	var stepFinal *HostPlanStep
	for i := range hp.Steps {
		if hp.Steps[i].ID == "step-final" {
			stepFinal = &hp.Steps[i]
			break
		}
	}

	if stepFinal == nil {
		t.Fatal("expected to find step-final")
	}

	// Dependencies should be sorted
	expectedDeps := []string{"step-a", "step-m", "step-z"}
	if len(stepFinal.DependsOn) != len(expectedDeps) {
		t.Fatalf("expected %d dependencies, got %d", len(expectedDeps), len(stepFinal.DependsOn))
	}
	for i, expected := range expectedDeps {
		if stepFinal.DependsOn[i] != expected {
			t.Errorf("dependency %d: expected %q, got %q", i, expected, stepFinal.DependsOn[i])
		}
	}
}

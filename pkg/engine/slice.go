// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

package engine

import (
	"fmt"
	"sort"
)

// SlicePlan deterministically partitions a Plan into HostPlans and GlobalSteps.
// Rules:
// - Steps are assigned by PlanStep.Host.LogicalID
// - Steps with empty Host.LogicalID are collected as GlobalSteps (controller/CLI handles separately)
// - Steps within HostPlan are ordered by Index then ID (stable)
// - Cross-host dependencies are rejected (Rule A: strict validation)
//
// Returns an error if any step depends on a step that is not in the same HostPlan.
// nolint:gocritic // passed by value intentionally; treated as immutable and keeps call sites simple.
func SlicePlan(plan Plan) (SliceResult, error) {
	result := SliceResult{
		HostPlans:            make(map[string]HostPlan),
		GlobalSteps:          nil,
		GlobalStepIDs:        nil,
		GlobalDependencyRefs: make(map[string][]string),
	}

	// Build step ID to host ID mapping for dependency validation
	stepToHost := make(map[string]string, len(plan.Steps))
	// globalStepIDs is a set (map[string]bool) for O(1) lookup during dependency validation.
	// Deterministic ordering is derived from GlobalSteps (sorted below), not from this map.
	// GlobalStepIDs (the ordered list) is built from GlobalSteps after deterministic sorting.
	globalStepIDs := make(map[string]bool)

	// First pass: collect steps and build mappings
	for i := range plan.Steps {
		step := &plan.Steps[i]
		hostID := step.Host.LogicalID
		if hostID == "" {
			globalStepIDs[step.ID] = true
			result.GlobalSteps = append(result.GlobalSteps, *step)
			continue
		}
		stepToHost[step.ID] = hostID
	}

	// Sort global steps deterministically and build GlobalStepIDs list
	sort.SliceStable(result.GlobalSteps, func(i, j int) bool {
		if result.GlobalSteps[i].Index != result.GlobalSteps[j].Index {
			return result.GlobalSteps[i].Index < result.GlobalSteps[j].Index
		}
		return result.GlobalSteps[i].ID < result.GlobalSteps[j].ID
	})

	// Build sorted GlobalStepIDs list for explicit tracking
	result.GlobalStepIDs = make([]string, 0, len(globalStepIDs))
	for i := range result.GlobalSteps {
		result.GlobalStepIDs = append(result.GlobalStepIDs, result.GlobalSteps[i].ID)
	}

	// Second pass: assign steps to host plans and validate dependencies
	for i := range plan.Steps {
		step := &plan.Steps[i]
		hostID := step.Host.LogicalID
		if hostID == "" {
			continue // Already handled as global step
		}

		hp, ok := result.HostPlans[hostID]
		if !ok {
			hp = HostPlan{
				Version: HostPlanSchemaVersion,
				PlanID:  plan.ID,
				Host:    step.Host,
				Steps:   nil,
				Meta:    nil,
			}
		}

		// Validate dependencies: all must be in the same host plan
		localDeps := make([]string, 0, len(step.DependsOn))
		globalDeps := make([]string, 0)
		for _, depID := range step.DependsOn {
			depHostID, exists := stepToHost[depID]
			if !exists {
				// Check if it's a global step (allowed)
				if globalStepIDs[depID] {
					// Global step dependency - track explicitly for controller enforcement
					globalDeps = append(globalDeps, depID)
					continue
				}
				return result, fmt.Errorf("step %q depends on unknown step %q", step.ID, depID)
			}

			if depHostID != hostID {
				return result, fmt.Errorf("step %q on host %q depends on step %q on host %q (cross-host dependencies not allowed in v1)", step.ID, hostID, depID, depHostID)
			}

			localDeps = append(localDeps, depID)
		}

		// Track global dependencies explicitly
		if len(globalDeps) > 0 {
			sort.Strings(globalDeps) // Deterministic ordering
			result.GlobalDependencyRefs[step.ID] = globalDeps
		}

		// Sort dependencies deterministically
		sort.Strings(localDeps)
		if len(localDeps) == 0 {
			localDeps = nil
		}

		hp.Steps = append(hp.Steps, HostPlanStep{
			ID:        step.ID,
			Index:     step.Index,
			Action:    step.Action,
			Target:    step.Target,
			Inputs:    step.Inputs,
			DependsOn: localDeps,
			Meta:      cloneStringMap(step.Meta),
		})

		result.HostPlans[hostID] = hp
	}

	// Sort steps within each host plan deterministically
	for hostID, hp := range result.HostPlans {
		sort.SliceStable(hp.Steps, func(i, j int) bool {
			if hp.Steps[i].Index != hp.Steps[j].Index {
				return hp.Steps[i].Index < hp.Steps[j].Index
			}
			return hp.Steps[i].ID < hp.Steps[j].ID
		})
		result.HostPlans[hostID] = hp
	}

	return result, nil
}

// SlicePlanByHost is a convenience function that returns only HostPlans (legacy compatibility).
// It calls SlicePlan and returns only the HostPlans map, ignoring global steps.
//
// Deprecated: Use SlicePlan instead for explicit global step handling.
// This function ignores errors from SlicePlan for backward compatibility.
// Once all callers migrate to SlicePlan, this function should be removed.
//
// nolint:gocritic // passed by value intentionally; treated as immutable and keeps call sites simple.
//
//nolint:staticcheck // Deprecated function kept for backward compatibility
func SlicePlanByHost(plan Plan) map[string]HostPlan {
	result, err := SlicePlan(plan)
	if err != nil {
		// For backward compatibility, return empty map on error
		// Callers should migrate to SlicePlan() for proper error handling
		return make(map[string]HostPlan)
	}
	return result.HostPlans
}

func cloneStringMap(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

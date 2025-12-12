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
func SlicePlan(plan Plan) (SliceResult, error) {
	result := SliceResult{
		HostPlans:   make(map[string]HostPlan),
		GlobalSteps: nil,
	}

	// Build step ID to host ID mapping for dependency validation
	stepToHost := make(map[string]string, len(plan.Steps))
	globalStepIDs := make(map[string]bool)

	// First pass: collect steps and build mappings
	for _, step := range plan.Steps {
		hostID := step.Host.LogicalID
		if hostID == "" {
			globalStepIDs[step.ID] = true
			result.GlobalSteps = append(result.GlobalSteps, step)
			continue
		}
		stepToHost[step.ID] = hostID
	}

	// Sort global steps deterministically
	sort.SliceStable(result.GlobalSteps, func(i, j int) bool {
		if result.GlobalSteps[i].Index != result.GlobalSteps[j].Index {
			return result.GlobalSteps[i].Index < result.GlobalSteps[j].Index
		}
		return result.GlobalSteps[i].ID < result.GlobalSteps[j].ID
	})

	// Second pass: assign steps to host plans and validate dependencies
	for _, step := range plan.Steps {
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
		for _, depID := range step.DependsOn {
			depHostID, exists := stepToHost[depID]
			if !exists {
				// Check if it's a global step (allowed)
				if globalStepIDs[depID] {
					// Global step dependency - controller handles this
					// For v1 strict mode, we reject cross-host but allow global deps
					// (global steps are handled separately by controller)
					continue
				}
				return result, fmt.Errorf("step %q depends on unknown step %q", step.ID, depID)
			}

			if depHostID != hostID {
				return result, fmt.Errorf("step %q on host %q depends on step %q on host %q (cross-host dependencies not allowed in v1)", step.ID, hostID, depID, depHostID)
			}

			localDeps = append(localDeps, depID)
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

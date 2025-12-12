// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

package plan

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"stagecraft/internal/core"
	"stagecraft/pkg/engine"
)

// ToEnginePlan converts a core.Plan to an engine.Plan.
// This adapter bridges the existing planner to the wire contract.
//
// Mapping rules:
// - Operation.ID → PlanStep.ID (required, fallback only for defensive purposes)
// - Operation.Dependencies → PlanStep.DependsOn (direct mapping, IDs are stable)
// - Operation.Metadata → PlanStep.Inputs (as typed JSON structs)
// - Host assignment defaults to "local" for single-host v1 mode
//
// Returns an error if duplicate operation IDs are detected or if any operation ID is empty.
func ToEnginePlan(corePlan *core.Plan, envName string) (*engine.Plan, error) {
	if corePlan == nil {
		return nil, fmt.Errorf("core plan is nil")
	}

	seen := make(map[string]struct{}, len(corePlan.Operations))
	steps := make([]engine.PlanStep, 0, len(corePlan.Operations))

	for i, op := range corePlan.Operations {
		stepID := op.ID
		if stepID == "" {
			// Strict: treat empty ID as planner bug
			return nil, fmt.Errorf("operation id is empty at index %d (planner bug)", i)
		}

		// Check for duplicate in single pass
		if _, ok := seen[stepID]; ok {
			return nil, fmt.Errorf("duplicate operation id: %q", stepID)
		}
		seen[stepID] = struct{}{}

		action := mapOperationTypeToAction(op.Type)

		inputsJSON, err := marshalOperationInputs(op.Type, op.Metadata)
		if err != nil {
			return nil, fmt.Errorf("marshaling operation inputs for op index %d: %w", i, err)
		}

		target := engine.ResourceRef{
			Kind:     mapOperationTypeToResourceKind(op.Type),
			Name:     stepID,
			Provider: "stagecraft",
		}

		// Dependencies reference Operation.ID values, so they map directly to step IDs
		// Sort defensively to ensure deterministic ordering even if planner changes
		dependsOn := append([]string(nil), op.Dependencies...)
		if len(dependsOn) == 0 {
			dependsOn = nil // Normalize empty slice to nil for JSON
		} else {
			sort.Strings(dependsOn)
		}

		steps = append(steps, engine.PlanStep{
			ID:        stepID,
			Index:     i,
			Action:    action,
			Target:    target,
			Host:      engine.HostRef{LogicalID: "local"},
			Inputs:    inputsJSON,
			DependsOn: dependsOn,
			Meta:      nil,
		})
	}

	planID, err := generatePlanID(steps, envName)
	if err != nil {
		return nil, fmt.Errorf("generating plan ID: %w", err)
	}

	return &engine.Plan{
		Version: engine.PlanSchemaVersion,
		ID:      planID,
		Summary: fmt.Sprintf("Deploy to %s", envName),
		Steps:   steps,
		Meta:    nil,
	}, nil
}

// mapOperationTypeToAction maps core.OperationType to engine.StepAction.
func mapOperationTypeToAction(opType core.OperationType) engine.StepAction {
	switch opType {
	case core.OpTypeBuild:
		return engine.StepActionBuild
	case core.OpTypeDeploy:
		return engine.StepActionApplyCompose
	case core.OpTypeMigration:
		return engine.StepActionMigrate
	case core.OpTypeHealthCheck:
		return engine.StepActionHealthCheck
	case core.OpTypeInfraProvision:
		return engine.StepActionCreate
	default:
		return engine.StepActionNoop
	}
}

// mapOperationTypeToResourceKind maps core.OperationType to resource kind string.
func mapOperationTypeToResourceKind(opType core.OperationType) string {
	switch opType {
	case core.OpTypeBuild:
		return "image"
	case core.OpTypeDeploy:
		return "service"
	case core.OpTypeMigration:
		return "database"
	case core.OpTypeHealthCheck:
		return "service"
	case core.OpTypeInfraProvision:
		return "infrastructure"
	default:
		return "resource"
	}
}

// OperationInputs represents the typed input structure for operations.
// Each operation type has a specific struct to ensure deterministic JSON.
type OperationInputs struct {
	Environment string `json:"environment,omitempty"`
	Provider    string `json:"provider,omitempty"`

	Database string `json:"database,omitempty"`
	Strategy string `json:"strategy,omitempty"`
	Engine   string `json:"engine,omitempty"`
	Path     string `json:"path,omitempty"`
	ConnEnv  string `json:"conn_env,omitempty"`
}

// marshalOperationInputs converts operation metadata to typed struct, then JSON.
// This ensures deterministic JSON output without map[string]interface{}.
func marshalOperationInputs(_ core.OperationType, metadata map[string]interface{}) (json.RawMessage, error) {
	var inputs OperationInputs

	if metadata != nil {
		if env, ok := metadata["environment"].(string); ok {
			inputs.Environment = env
		}
		if provider, ok := metadata["provider"].(string); ok {
			inputs.Provider = provider
		}
		if database, ok := metadata["database"].(string); ok {
			inputs.Database = database
		}
		if strategy, ok := metadata["strategy"].(string); ok {
			inputs.Strategy = strategy
		}
		if engineName, ok := metadata["engine"].(string); ok {
			inputs.Engine = engineName
		}
		if path, ok := metadata["path"].(string); ok {
			inputs.Path = path
		}
		if connEnv, ok := metadata["conn_env"].(string); ok {
			inputs.ConnEnv = connEnv
		}
	}

	return json.Marshal(inputs)
}

// planIDInput represents the structure used for deterministic plan ID generation.
type planIDInput struct {
	Environment string       `json:"environment"`
	Steps       []planIDStep `json:"steps"`
}

type planIDStep struct {
	ID        string          `json:"id"`
	Index     int             `json:"index"`
	Action    string          `json:"action"`
	Target    planIDResource  `json:"target"`
	Host      string          `json:"host"`
	Inputs    json.RawMessage `json:"inputs"`
	DependsOn []string        `json:"dependsOn"`
}

type planIDResource struct {
	Kind     string `json:"kind"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
}

// generatePlanID creates a deterministic plan ID using SHA256 hash of canonical JSON.
// The plan ID is stable across repeated runs with the same inputs.
func generatePlanID(steps []engine.PlanStep, envName string) (string, error) {
	psteps := make([]planIDStep, 0, len(steps))
	for _, step := range steps {
		psteps = append(psteps, planIDStep{
			ID:     step.ID,
			Index:  step.Index,
			Action: string(step.Action),
			Target: planIDResource{
				Kind:     step.Target.Kind,
				Name:     step.Target.Name,
				Provider: step.Target.Provider,
			},
			Host:      step.Host.LogicalID,
			Inputs:    step.Inputs,
			DependsOn: step.DependsOn,
		})
	}

	// Ensure deterministic ordering
	sort.Slice(psteps, func(i, j int) bool {
		if psteps[i].Index != psteps[j].Index {
			return psteps[i].Index < psteps[j].Index
		}
		return psteps[i].ID < psteps[j].ID
	})

	b, err := json.Marshal(planIDInput{
		Environment: envName,
		Steps:       psteps,
	})
	if err != nil {
		return "", fmt.Errorf("marshaling plan ID input: %w", err)
	}

	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:12]), nil
}

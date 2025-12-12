// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

package agent

import (
	"context"
	"fmt"

	"stagecraft/pkg/engine"
)

// Executor executes a HostPlan step by step, respecting dependencies.
type Executor struct {
	// executors maps StepAction to action-specific executors
	executors map[engine.StepAction]StepExecutor
}

// StepExecutor executes a single step.
type StepExecutor interface {
	Execute(ctx context.Context, step engine.HostPlanStep, inputs []byte) error
}

// NewExecutor creates a new executor with default action handlers.
func NewExecutor() *Executor {
	return &Executor{
		executors: make(map[engine.StepAction]StepExecutor),
	}
}

// RegisterExecutor registers an executor for a specific action.
func (e *Executor) RegisterExecutor(action engine.StepAction, executor StepExecutor) {
	e.executors[action] = executor
}

// ExecuteHostPlan executes a HostPlan step by step, respecting dependencies.
// Steps are executed in topological order based on DependsOn relationships.
func (e *Executor) ExecuteHostPlan(ctx context.Context, plan engine.HostPlan) (*engine.ExecutionReport, error) {
	report := &engine.ExecutionReport{
		PlanID: plan.PlanID,
		Status: engine.ExecStatusSucceeded,
		Steps:  make([]engine.StepExecution, 0, len(plan.Steps)),
	}

	// Track completed steps
	completed := make(map[string]bool, len(plan.Steps))
	stepMap := make(map[string]engine.HostPlanStep, len(plan.Steps))
	for _, step := range plan.Steps {
		stepMap[step.ID] = step
	}

	// Execute steps in order (already sorted by Index, dependencies validated)
	for _, step := range plan.Steps {
		// Check dependencies are completed
		for _, depID := range step.DependsOn {
			if !completed[depID] {
				return nil, fmt.Errorf("step %q depends on %q which has not completed", step.ID, depID)
			}
		}

		// Execute step
		stepExec := engine.StepExecution{
			StepID: step.ID,
			Host:   plan.Host,
			Status: engine.StepStatusRunning,
		}

		executor, ok := e.executors[step.Action]
		if !ok {
			stepExec.Status = engine.StepStatusSkipped
			stepExec.Error = &engine.ExecutionError{
				Code:    "NO_EXECUTOR",
				Message: fmt.Sprintf("no executor registered for action %q", step.Action),
			}
			report.Status = engine.ExecStatusPartial
		} else {
			err := executor.Execute(ctx, step, step.Inputs)
			if err != nil {
				stepExec.Status = engine.StepStatusFailed
				stepExec.Error = &engine.ExecutionError{
					Code:    "EXECUTION_ERROR",
					Message: err.Error(),
				}
				report.Status = engine.ExecStatusFailed
			} else {
				stepExec.Status = engine.StepStatusSucceeded
			}
		}

		completed[step.ID] = true
		report.Steps = append(report.Steps, stepExec)

		// If step failed and we're in strict mode, stop execution
		if stepExec.Status == engine.StepStatusFailed {
			break
		}
	}

	return report, nil
}

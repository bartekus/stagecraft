// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

package engine

import "context"

// Engine is the authoritative planner/executor contract.
// This interface anchors both CLI Mode A (local execution) and controller Mode B (remote execution).
type Engine interface {
	// ComputePlan generates a Plan from topology and current state.
	ComputePlan(ctx context.Context, req ComputePlanRequest) (*ComputePlanResponse, error)

	// ExecutePlan executes a plan locally (for CLI Mode A).
	ExecutePlan(ctx context.Context, req ExecutePlanRequest) (*ExecutePlanResponse, error)

	// InspectState reads current state from runtime.
	InspectState(ctx context.Context, req InspectStateRequest) (*InspectStateResponse, error)
}

// ComputePlanRequest contains inputs for plan computation.
type ComputePlanRequest struct {
	Topology TopologySnapshot `json:"topology"`
	State    StateSnapshot    `json:"state"`
	Options  PlanOptions      `json:"options,omitempty"`
}

// ComputePlanResponse contains the computed plan.
type ComputePlanResponse struct {
	Plan Plan `json:"plan"`
	// Diff can come later; keep v1 small if you want.
}

// InspectStateRequest contains inputs for state inspection.
type InspectStateRequest struct {
	// e.g. host ref + runtime selector
	Host    HostRef `json:"host"`
	Runtime string  `json:"runtime"`
}

// InspectStateResponse contains the inspected state.
type InspectStateResponse struct {
	State StateSnapshot `json:"state"`
}

// ExecutePlanRequest contains inputs for plan execution.
type ExecutePlanRequest struct {
	Plan    Plan        `json:"plan"`
	Options ExecOptions `json:"options,omitempty"`
}

// ExecutePlanResponse contains the execution report.
type ExecutePlanResponse struct {
	Report ExecutionReport `json:"report"`
}

// PlanOptions contains options for plan computation.
type PlanOptions struct {
	// Future: dry-run mode, filters, etc.
}

// ExecOptions contains options for plan execution.
type ExecOptions struct {
	DryRun      bool     `json:"dryRun,omitempty"`
	MaxParallel int      `json:"maxParallel,omitempty"`
	StepFilter  []string `json:"stepFilter,omitempty"`
}

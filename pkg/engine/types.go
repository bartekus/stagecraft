// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

package engine

import (
	"encoding/json"
)

// Schema versions - treat as wire contract.
const (
	PlanSchemaVersion     = "v1"
	HostPlanSchemaVersion = "v1"
)

// TopologySnapshot is the desired state input to planning.
// Keep this minimal and portable - avoid controller concepts.
type TopologySnapshot struct {
	// Schema version for this snapshot (separate from plan schema if needed later).
	Version string `json:"version"`

	// Arbitrary metadata that does not leak multi-tenant/controller details.
	// Prefer small stable keys. Leave empty for most uses.
	Meta map[string]string `json:"meta,omitempty"`

	// Resources is a provider/runtime-neutral desired state graph.
	// Keep ordering deterministic: callers should sort by Kind+Name.
	Resources []ResourceSpec `json:"resources"`
}

// StateSnapshot represents current runtime state.
type StateSnapshot struct {
	// Schema version for this snapshot.
	Version string `json:"version"`

	Meta map[string]string `json:"meta,omitempty"`

	// Resources is the observed state. Keep ordering deterministic.
	Resources []ResourceState `json:"resources"`
}

// ResourceSpec represents a desired resource in the topology.
type ResourceSpec struct {
	Ref  ResourceRef       `json:"ref"`
	Data json.RawMessage   `json:"data"`           // provider/runtime-specific desired payload
	Meta map[string]string `json:"meta,omitempty"` // safe annotations
}

// ResourceState represents an observed resource state.
type ResourceState struct {
	Ref  ResourceRef       `json:"ref"`
	Data json.RawMessage   `json:"data"`           // provider/runtime-specific observed payload
	Meta map[string]string `json:"meta,omitempty"` // safe annotations
}

// ResourceRef uniquely identifies a resource.
type ResourceRef struct {
	Kind      string `json:"kind"`                // e.g. "service", "network", "volume", "droplet"
	Name      string `json:"name"`                // logical name
	Provider  string `json:"provider"`            // e.g. "docker-compose", "kubernetes", "digitalocean"
	Namespace string `json:"namespace,omitempty"` // optional grouping
}

// Plan is the OSS wire contract produced by Engine.ComputePlan.
// Determinism rules: Steps must be emitted in stable order; callers can rely on Index.
type Plan struct {
	Version string `json:"version"` // must be PlanSchemaVersion

	// ID should be deterministic when computed from the same topology+state+options.
	// Implementation can use a canonical JSON hash. The type does not enforce it.
	ID string `json:"id"`

	Summary string `json:"summary,omitempty"`

	Steps []PlanStep `json:"steps"`

	Meta map[string]string `json:"meta,omitempty"`
}

// PlanStep represents a single step in a deployment plan.
type PlanStep struct {
	ID    string `json:"id"`    // stable step id
	Index int    `json:"index"` // total order across the full plan

	Action StepAction  `json:"action"`
	Target ResourceRef `json:"target"`

	// Host is required for anything that must execute "somewhere".
	// For purely global steps, leave Host.LogicalID empty.
	Host HostRef `json:"host"`

	// Inputs is opaque payload owned by provider/runtime implementation.
	// Use json.RawMessage to avoid map ordering issues.
	Inputs json.RawMessage `json:"inputs"`

	DependsOn []string          `json:"dependsOn,omitempty"` // step IDs
	Meta      map[string]string `json:"meta,omitempty"`
}

// StepAction represents the action to take in a plan step.
type StepAction string

const (
	// StepActionCreate creates a resource.
	StepActionCreate StepAction = "create"
	// StepActionUpdate updates a resource.
	StepActionUpdate StepAction = "update"
	// StepActionDelete deletes a resource.
	StepActionDelete StepAction = "delete"
	// StepActionNoop performs no operation.
	StepActionNoop StepAction = "noop"

	// StepActionRenderCompose renders compose artifacts for deployment.
	StepActionRenderCompose StepAction = "render_compose"
	// StepActionApplyCompose applies a rendered compose file.
	StepActionApplyCompose StepAction = "apply_compose"
	// StepActionRollout performs a rollout deployment.
	StepActionRollout StepAction = "rollout"

	// StepActionBuild builds container images for deployment.
	StepActionBuild StepAction = "build"
	// StepActionMigrate runs database migrations.
	StepActionMigrate StepAction = "migrate"
	// StepActionHealthCheck performs health checks on services.
	StepActionHealthCheck StepAction = "health_check"
)

// HostRef identifies a host where steps execute.
type HostRef struct {
	LogicalID string            `json:"logicalId"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// HostPlan is the portable "sub-plan blob" sent to an Agent.
// It must be self-contained: the agent should not need global topology/state.
type HostPlan struct {
	Version string `json:"version"` // must be HostPlanSchemaVersion
	PlanID  string `json:"planId"`

	Host HostRef `json:"host"`

	Steps []HostPlanStep `json:"steps"`

	Meta map[string]string `json:"meta,omitempty"`
}

// HostPlanStep represents a step in a host-specific plan.
type HostPlanStep struct {
	ID     string      `json:"id"`
	Index  int         `json:"index"`
	Action StepAction  `json:"action"`
	Target ResourceRef `json:"target"`

	Inputs    json.RawMessage   `json:"inputs"`
	DependsOn []string          `json:"dependsOn,omitempty"`
	Meta      map[string]string `json:"meta,omitempty"`
}

// SliceResult contains the result of slicing a Plan by host.
// It separates host-specific plans from global steps that must be handled by the controller/CLI.
type SliceResult struct {
	// HostPlans maps host LogicalID to the HostPlan for that host.
	HostPlans map[string]HostPlan `json:"hostPlans"`

	// GlobalSteps are steps with empty Host.LogicalID that must be handled separately.
	// They are ordered by Index then ID (stable).
	GlobalSteps []PlanStep `json:"globalSteps,omitempty"`

	// GlobalStepIDs is the ordered list of global step IDs (Index then ID).
	// Useful for fast lookup and explicit tracking of which steps must complete before host plans can execute.
	GlobalStepIDs []string `json:"globalStepIds,omitempty"`

	// GlobalDependencyRefs maps host step IDs to the global step IDs they depend on.
	// This makes explicit which host steps require global step completion before execution.
	// Key: host step ID, Value: list of global step IDs that step depends on.
	GlobalDependencyRefs map[string][]string `json:"globalDependencyRefs,omitempty"`
}

// ExecutionReport is the portable execution result.
// Avoid time.Time in the wire schema for now (keeps determinism + cross-lang easier).
// Use RFC3339 strings or unix millis when you implement execution.
type ExecutionReport struct {
	PlanID string `json:"planId"`

	Status ExecutionStatus `json:"status"`

	Steps []StepExecution `json:"steps"`

	Meta map[string]string `json:"meta,omitempty"`
}

// ExecutionStatus represents the overall status of plan execution.
type ExecutionStatus string

const (
	// ExecStatusSucceeded indicates all steps completed successfully.
	ExecStatusSucceeded ExecutionStatus = "succeeded"
	// ExecStatusFailed indicates execution failed.
	ExecStatusFailed ExecutionStatus = "failed"
	// ExecStatusPartial indicates partial success.
	ExecStatusPartial ExecutionStatus = "partial"
)

// StepStatus represents the status of an individual step execution.
type StepStatus string

const (
	// StepStatusPending indicates the step is pending execution.
	StepStatusPending StepStatus = "pending"
	// StepStatusRunning indicates the step is currently running.
	StepStatusRunning StepStatus = "running"
	// StepStatusSucceeded indicates the step completed successfully.
	StepStatusSucceeded StepStatus = "succeeded"
	// StepStatusFailed indicates the step failed.
	StepStatusFailed StepStatus = "failed"
	// StepStatusSkipped indicates the step was skipped.
	StepStatusSkipped StepStatus = "skipped"
)

// StepExecution represents the execution result of a single step.
type StepExecution struct {
	StepID string  `json:"stepId"`
	Host   HostRef `json:"host"`

	Status StepStatus `json:"status"`

	// Optional timestamps as strings to keep schema simple and deterministic in tests.
	StartedAt   string `json:"startedAt,omitempty"`
	CompletedAt string `json:"completedAt,omitempty"`

	Error *ExecutionError `json:"error,omitempty"`

	// Logs are optional; streaming can also be done out-of-band via emitter.
	Logs []LogLine `json:"logs,omitempty"`

	Meta map[string]string `json:"meta,omitempty"`
}

// ExecutionError represents an error that occurred during step execution.
type ExecutionError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

// LogLine represents a single log line from execution.
type LogLine struct {
	// Optional timestamp string if you want it.
	Time    string `json:"time,omitempty"`
	Stream  string `json:"stream"`  // "stdout" | "stderr" | "system"
	Message string `json:"message"` // single line
}

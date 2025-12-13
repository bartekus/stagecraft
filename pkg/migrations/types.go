// Package migrations defines the interface and types for migration engines.
//
// SPDX-License-Identifier: AGPL-3.0-or-later
package migrations

// MigrationID is a stable identifier for a migration.
type MigrationID string

// Migration is the minimal unit the engine can plan/apply.
type Migration struct {
	// ID is the stable identifier for the migration.
	ID MigrationID `json:"id"`

	// Description is a human-readable description of the migration.
	Description string `json:"description"`

	// Tags are optional categorization tags. MUST be sorted.
	Tags []string `json:"tags,omitempty"`

	// Source is the stable logical source descriptor (e.g., "sql:db/main").
	Source string `json:"source"`

	// DependsOn is an optional list of migration IDs this migration depends on.
	DependsOn []MigrationID `json:"depends_on,omitempty"`
}

// Selection defines which migrations to consider.
type Selection struct {
	// All includes all available migrations.
	All bool `json:"all"`

	// IDs includes specific migrations by ID.
	IDs []MigrationID `json:"ids,omitempty"`

	// Tags includes migrations matching specific tags.
	Tags []string `json:"tags,omitempty"`
}

// MigrationMode defines the execution mode.
type MigrationMode string

const (
	// ModePlan indicates a dry-run to determine what would be applied.
	ModePlan MigrationMode = "plan"
	// ModeApply indicates that migrations should be executed.
	ModeApply MigrationMode = "apply"
)

// MigrationRequest defines the input for Engine operations.
type MigrationRequest struct {
	// Environment is the resolved environment name (e.g. "dev", "prod").
	Environment string `json:"environment"`

	// Mode specifies whether to plan or apply.
	Mode MigrationMode `json:"mode"`

	// Selection defines which migrations to include.
	Selection Selection `json:"selection"`

	// FailFast, if true, stops execution at the first failure.
	FailFast bool `json:"fail_fast"`

	// AllowNoop, if true, treats "no migrations selected" as success.
	AllowNoop bool `json:"allow_noop"`

	// DryRun, if true, forces plan-only behavior even in Apply mode.
	DryRun bool `json:"dry_run"`
}

// StepOutcome defines the result of a single migration step.
type StepOutcome string

const (
	// OutcomeApplied indicates the migration was successfully applied.
	OutcomeApplied StepOutcome = "applied"
	// OutcomeSkipped indicates the migration was skipped (e.g. already applied).
	OutcomeSkipped StepOutcome = "skipped"
	// OutcomeFailed indicates the migration failed to apply.
	OutcomeFailed StepOutcome = "failed"
)

// MigrationStepResult represents the result of a single migration step.
type MigrationStepResult struct {
	// ID is the migration ID.
	ID MigrationID `json:"id"`

	// Outcome is the result of the step.
	Outcome StepOutcome `json:"outcome"`

	// Message is an optional sanitized status message.
	Message string `json:"message,omitempty"`

	// Warnings is an optional list of sorted warning messages.
	Warnings []string `json:"warnings,omitempty"`
}

// PlanSummary contains aggregate counts for a plan.
type PlanSummary struct {
	Total      int `json:"total"`
	WouldApply int `json:"would_apply"`
	WouldSkip  int `json:"would_skip"`
}

// MigrationPlan represents the output of a Plan operation.
type MigrationPlan struct {
	// Engine is the engine identifier.
	Engine string `json:"engine"`

	// Environment is the target environment.
	Environment string `json:"environment"`

	// Steps is the ordered list of planned steps.
	Steps []MigrationStepResult `json:"steps"`

	// Summary is the aggregate plan summary.
	Summary PlanSummary `json:"summary"`
}

// ApplySummary contains aggregate counts for an apply operation.
type ApplySummary struct {
	Total   int `json:"total"`
	Applied int `json:"applied"`
	Skipped int `json:"skipped"`
	Failed  int `json:"failed"`
}

// MigrationApplyResult represents the output of an Apply operation.
type MigrationApplyResult struct {
	// Engine is the engine identifier.
	Engine string `json:"engine"`

	// Environment is the target environment.
	Environment string `json:"environment"`

	// Steps is the ordered list of executed steps.
	Steps []MigrationStepResult `json:"steps"`

	// Summary is the aggregate apply summary.
	Summary ApplySummary `json:"summary"`
}

// ValidationResult represents the output of a Validate operation.
type ValidationResult struct {
	// Engine is the engine identifier.
	Engine string `json:"engine"`

	// Environment is the target environment.
	Environment string `json:"environment"`

	// OK indicates if validation passed.
	OK bool `json:"ok"`

	// Warnings is a sorted list of warning messages.
	Warnings []string `json:"warnings,omitempty"`

	// Message is a sanitized status message.
	Message string `json:"message,omitempty"`
}

---
feature: MIGRATION_INTERFACE
version: v1
status: done
domain: migrations
inputs:
  flags: []
outputs:
  exit_codes: {}
---

# MIGRATION_INTERFACE - Specification

## Goal

Define the canonical, deterministic interface between Stagecraft and migration engines so the migration registry and CLI can plan and apply migrations without depending on a specific engine implementation.

## Non-Goals

- Defining CLI flags or UX (covered by CLI migrate features).
- Implementing a migration engine (raw engine exists separately).
- Distributed locking or multi-run concurrency management (future).
- Rollback semantics (explicitly deferred unless added below).

---

## Normative Language

The key words **MUST**, **MUST NOT**, **SHOULD**, and **MAY** are to be interpreted as described in RFC 2119.

---

## Determinism Rules

1. All lists of migrations and steps **MUST** be sorted deterministically.
2. Plans and results **MUST NOT** include timestamps, random IDs, or nondeterministic fields by default.
3. If maps are used internally, any serialization output **MUST** be rendered with sorted keys, or converted to ordered lists.
4. Any optional non-deterministic fields (example: durations) **MUST** be disabled by default and excluded from deterministic outputs and golden tests.

---

## Data Model

### MigrationID
A stable identifier for a migration.

- Type: string
- Rules:
  - **MUST** be stable across runs.
  - **MUST** be unique within a project.
  - **MUST NOT** depend on absolute paths or machine-specific locations.

### Migration
A migration is the minimal unit the engine can plan/apply.

Required fields:

- `ID` (MigrationID) - required
- `Description` (string) - required (may be empty but should exist)
- `Tags` ([]string) - optional, but if present **MUST** be sorted lexicographically
- `Source` (string) - required, stable logical source descriptor (example: `sql:db/main`, `tool:prisma`, `raw:/migrations`)

Optional fields:
- `DependsOn` ([]MigrationID) - optional; for v1, if present, must remain acyclic and deterministically resolvable.
  - If DAG is not supported in v1 implementation, the engine **MUST** return an Unsupported error if dependencies are present.

### Selection
Defines which migrations to consider.

- `All` (bool)
- `IDs` ([]MigrationID)
- `Tags` ([]string)

Rules:
- If `All` is true, `IDs` and `Tags` are ignored.
- `IDs` and `Tags` if present **MUST** be treated as sets but results **MUST** preserve deterministic ordering in outputs.

---

## Requests and Results

### MigrationMode
- `plan`
- `apply`

### MigrationRequest
Fields:

- `Environment` (string) - required; resolved environment name
- `Mode` (MigrationMode) - required
- `Selection` (Selection) - required
- `FailFast` (bool) - optional; default false
- `AllowNoop` (bool) - optional; default true
- `DryRun` (bool) - optional; for Mode=plan this is implicitly true; for Mode=apply if true then behavior MUST be plan-only

The request **MUST NOT** carry secrets in plain strings intended for output. Engines may receive secrets via internal handles, not via printable fields.

### StepOutcome
- `applied`
- `skipped`
- `failed`

### MigrationStepResult
Fields:

- `ID` (MigrationID) - required
- `Outcome` (StepOutcome) - required
- `Message` (string) - optional; **MUST** be sanitized (no secrets)
- `Warnings` ([]string) - optional; if present **MUST** be sorted

### MigrationPlan
Fields:

- `Engine` (string) - required; stable engine identifier
- `Environment` (string) - required
- `Steps` ([]MigrationStepResult) - required; ordered deterministically
- `Summary` (PlanSummary) - required

### MigrationApplyResult
Fields:

- `Engine` (string) - required
- `Environment` (string) - required
- `Steps` ([]MigrationStepResult) - required; ordered deterministically
- `Summary` (ApplySummary) - required

### Summaries
Summaries **MUST** include counts:

PlanSummary:
- `Total`
- `WouldApply`
- `WouldSkip`

ApplySummary:
- `Total`
- `Applied`
- `Skipped`
- `Failed`

---

## Error Semantics

### ErrorKind
Engines **MUST** classify errors into one of:

- `invalid_config`
- `unsupported`
- `dependency_missing`
- `connection_failed`
- `migration_failed`
- `internal`

### MigrationError
Engines **SHOULD** return structured errors. If the implementation language returns an error interface, it **MUST** be possible to extract:

- `Kind` (ErrorKind)
- `Message` (string, sanitized)
- `Cause` (optional underlying error, not for user output)
- `StepID` (optional MigrationID) when the error pertains to a specific migration

---

## Engine Interface Contract

Stagecraft defines a migration engine contract.

### Engine Identity
- Each engine **MUST** expose a stable `Name()` string.

### Required Methods
An engine **MUST** implement:

- `List(ctx, req) -> ([]Migration, error)`
  - Returns the candidate migrations for the given environment and selection scope.
  - Returned list **MUST** be deterministically ordered (lexicographic by `ID` unless the engine defines another stable ordering rule and documents it).
- `Plan(ctx, req) -> (MigrationPlan, error)`
  - Must not mutate the target.
- `Apply(ctx, req) -> (MigrationApplyResult, error)`
  - May mutate the target.

### Optional Method
- `Validate(ctx, req) -> (ValidationResult, error)`
  - If not implemented, validation is treated as a no-op success.

### ValidationResult
- `Engine` (string)
- `Environment` (string)
- `OK` (bool)
- `Warnings` ([]string) sorted
- `Message` (string) sanitized

---

## Execution Rules

1. For Mode=plan:
   - Engine **MUST NOT** mutate targets.
   - Returned `Steps[].Outcome` **MUST** be either `applied` (meaning "would apply") or `skipped` (meaning "would skip"), but **MUST NOT** be `failed` unless the plan itself cannot be computed.

2. For Mode=apply:
   - Engine executes migrations in `Steps` order.
   - If `FailFast` is true, engine **MUST** stop at first failure and return `Failed >= 1`.
   - If `FailFast` is false, engine **MAY** continue applying subsequent migrations only if the engine can guarantee safety. If it cannot guarantee safety, it **MUST** behave as fail-fast.

3. No migrations case:
   - If selection yields zero migrations, engine **MUST** return a plan/apply result with `Total=0`.
   - If `AllowNoop` is true (default), this is not an error.
   - If `AllowNoop` is false, engine **MUST** return an error of kind `migration_failed` with a message indicating no migrations selected.

4. Sanitization:
   - `Message` and `Warnings` **MUST** not contain secrets.
   - If an engine captures raw stdout/stderr, it **MUST** either omit it or redact it.

---

## Suggested Go Shape (Non-Normative)

This section is illustrative only. Implementation may differ, but behavior must match the contract.

```go
// type MigrationID string
//
// type Migration struct {
//   ID          MigrationID
//   Description string
//   Tags        []string
//   Source      string
//   DependsOn   []MigrationID // optional
// }
//
// type Selection struct {
//   All  bool
//   IDs  []MigrationID
//   Tags []string
// }
//
// type MigrationMode string
// const (
//   ModePlan  MigrationMode = "plan"
//   ModeApply MigrationMode = "apply"
// )
//
// type MigrationRequest struct {
//   Environment string
//   Mode        MigrationMode
//   Selection   Selection
//   FailFast    bool
//   AllowNoop   bool
//   DryRun      bool
// }
//
// type Engine interface {
//   Name() string
//   List(ctx context.Context, req *MigrationRequest) ([]Migration, error)
//   Plan(ctx context.Context, req *MigrationRequest) (MigrationPlan, error)
//   Apply(ctx context.Context, req *MigrationRequest) (MigrationApplyResult, error)
// }
//
// type ValidatingEngine interface {
//   Engine
//   Validate(ctx context.Context, req *MigrationRequest) (ValidationResult, error)
// }

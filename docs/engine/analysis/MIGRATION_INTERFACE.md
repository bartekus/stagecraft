<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

# MIGRATION_INTERFACE - Analysis

## Purpose

Define the canonical migration contract for Stagecraft so that:
- CLI migration commands can be implemented deterministically.
- The migration registry (already present) can execute migrations through a stable interface.
- Raw migration engines (already present) can be standardized behind a contract.
- Deploy workflows can run pre-deploy and post-deploy migration hooks consistently.

This document is analysis-only. It does not define final APIs, but it does define required behaviors, invariants, and design constraints that the spec must encode.

---

## Scope

This feature defines:
- The migration unit model (identity, ordering, metadata).
- The execution lifecycle (plan, apply, validate).
- Inputs and outputs (context, environment, IO).
- Determinism and reporting guarantees.
- Error semantics suitable for CLI and orchestration.
- How migration engines plug in.

This feature does not implement:
- CLI commands (that is a later feature).
- A specific migration engine (raw engine exists; others may exist later).
- Remote execution or distributed locking (future).

---

## Definitions

### Migration
A deterministic, addressable unit of schema or data change that can be planned and applied in a target environment.

### Migration Engine
A runtime implementation capable of planning and applying migrations (example: raw SQL runner, tool wrapper, provider-specific runner).

### Plan
A deterministic preview of what would be applied, in what order, and with what risks, without mutating the target.

### Apply
A deterministic execution that mutates the target and returns a stable result report.

### Environment
A resolved Stagecraft environment (dev, staging, prod) including connection and credential material, but never leaking secrets in logs.

---

## Core Requirements

### R1. Standard contract
Stagecraft must define a stable interface for migration engines so the core orchestration and CLI do not depend on a specific engine.

### R2. Deterministic planning
Given the same inputs, the migration plan output must be bit-identical:
- Stable ordering.
- Stable IDs.
- Stable formatting.
- No timestamps, random IDs, or nondeterministic map iteration.

### R3. Deterministic application reporting
Applying migrations must produce a deterministic report:
- The set of attempted migrations, in stable order.
- Per migration outcome: applied, skipped, failed.
- A stable summary.

### R4. Explicit lifecycle
Engines must support a clear lifecycle with explicit semantics:
- Plan-only (no mutation).
- Apply (mutation).
- Optional validate step (may run in both plan and apply modes).

### R5. Error semantics
Errors must be typed or classified so callers can:
- Present meaningful CLI output.
- Distinguish user misconfig from runtime failure.
- Handle “no-op” safely.

### R6. Minimal leak surface
Migration outputs must not include secrets. Any engine-provided raw output must be sanitized or treated as sensitive with strict controls.

### R7. Registry compatibility
The interface must be compatible with the already implemented migration registry feature:
- The registry selects engine(s) and migration sources.
- The registry coordinates ordering and execution.

---

## Key Design Decisions

### D1. Migration identity
Migration identity must be:
- Globally unique within a Stagecraft project.
- Stable across runs.
- Derived from deterministic inputs.

Recommended approach:
- `MigrationID` is a string that includes a stable prefix and a deterministic name.
- For file-based migrations, use a stable path-based identifier (not absolute paths).
- For tool-based migrations, use a stable logical ID defined in config.

### D2. Ordering rules
Ordering must be deterministic:
- A migration set MUST define a stable order (lexicographic by ID unless explicitly specified).
- Explicit ordering is allowed but must be deterministic and validated.
- “Natural filesystem order” is forbidden unless explicitly sorted.

### D3. Plan output shape
Plan output must provide:
- Ordered list of steps.
- Per step metadata (id, description, risk flags).
- Whether each step is expected to apply or be skipped.

The plan must not require live mutation. If a plan needs state inspection, it can read, but must not write.

### D4. Apply idempotency expectations
The interface cannot guarantee idempotency for all engines, but it must:
- Explicitly represent whether a migration was applied vs skipped.
- Allow engines to consult state to avoid re-applying.
- Allow callers to decide whether re-apply is an error or a no-op based on mode.

### D5. Hook integration
Later deploy flows will call migration execution in predictable points:
- Pre-deploy migrations.
- Post-deploy migrations.
The interface must not assume only one usage context.

---

## Interface Concepts

The spec should define three layers:

1. **Core types**
   - Migration identity and metadata.
   - Plan and apply results.

2. **Engine contract**
   - `Plan(ctx, req) -> PlanResult`
   - `Apply(ctx, req) -> ApplyResult`
   - Optional `Validate(ctx, req) -> ValidationResult`

3. **Host/environment bindings**
   - The request includes resolved environment details and any required connections.
   - The interface must support different migration targets (database, file store, service API) without assuming Postgres-only.

---

## Execution Model

### Inputs
A migration operation is executed with:
- A resolved environment key/name.
- Target connection material or handles.
- A migration selection (all, subset by tags, subset by IDs).
- An execution mode:
  - Plan (dry-run)
  - Apply (execute)
- Safety settings:
  - Require confirmation (CLI concern later)
  - Fail fast vs continue (policy)

### Outputs
Outputs must be structured, deterministic, and serialization-friendly:
- Plain Go structs with stable JSON/YAML.
- Stable ordered arrays, no maps in final output unless keys are sorted.

---

## Failure Modes

Engines must classify errors at least into:
- Invalid configuration or missing required values.
- Unsupported platform or missing dependency (binary not found, version mismatch).
- Connection failure (target unavailable).
- Migration failure (engine executed but migration failed).
- Internal bug (should be rare and clearly marked).

Callers must be able to render a stable summary and an actionable primary error.

---

## Observability

Minimum required fields for results:
- Engine name
- Environment name
- Ordered list of steps
- Per step outcome and sanitized message
- Stable summary counts

Optional fields:
- Sanitized engine stdout/stderr excerpts with size limits
- Per step duration (durations are allowed if they are optional and not used for determinism comparisons; ideally omitted from deterministic output)

Recommendation:
- Keep deterministic output free of durations. If durations are included, they must be behind a non-default flag and excluded from golden tests.

---

## Security Considerations

- Never embed secrets in output.
- Avoid echoing DSNs unless redacted.
- If an engine consumes env vars, results must not print them.

---

## Testing Strategy

The spec should enable:
- Golden tests for plan/apply result formatting.
- Unit tests for ordering rules.
- Tests for error classification and stable rendering.
- Tests for “no migrations” behavior (must be deterministic and non-error by default).

---

## Open Questions

These must be resolved in the spec (or explicitly deferred):
1. Do we support rollback at the interface level or treat rollback as a separate feature?
2. Are migrations always linear or can there be branches (DAG)? For v1, linear ordering is recommended.
3. How do we represent “state store” (tracking applied migrations)? Engine-specific or core-managed?
4. Do we allow multiple targets (db + other) in one run? If yes, ordering between targets must be deterministic.
5. Do we allow tagging and selection (by tag, group, role)? Likely yes for deploy hooks.

---

## Acceptance Criteria

This feature is done when:
- A spec exists defining the migration interface contract with explicit types and semantics.
- Determinism rules are explicit and testable.
- The contract can wrap the existing raw engine without leaking raw assumptions into core orchestration.
- The contract is sufficient to implement CLI migrate plan and run later without redesign.

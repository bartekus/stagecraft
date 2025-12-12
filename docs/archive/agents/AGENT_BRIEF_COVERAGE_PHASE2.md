> **Superseded by** `docs/engine/agents/AGENT_BRIEF_COVERAGE.md` section "Phase 2: Quality Lift". Kept for historical reference. New coverage execution notes MUST go into the merged agent brief.

# Agent Brief: Test Coverage Quality Lift - Phase 2

**Feature ID**: GOV_V1_CORE  
**Spec**: `spec/governance/GOV_V1_CORE.md`  
**Context**: Non-blocking coverage improvements for low-coverage packages. See `docs/coverage/TEST_COVERAGE_ANALYSIS.md`, `docs/coverage/COVERAGE_COMPLIANCE_PLAN.md`, and `docs/agents/AGENT_BRIEF_COVERAGE_PHASE1.md`.

---

## Mission

Raise coverage for the lowest-coverage, non-core packages to stable, maintainable baselines **without** introducing new behaviour, changing user-facing semantics, or weakening validation.

This is a **quality lift**, not a compliance unblock. Phase 2 may only start after Phase 1 compliance is achieved.

---

## Scope

Focus strictly on the three low-coverage packages called out in `docs/coverage/TEST_COVERAGE_ANALYSIS.md`:

1. **`internal/git`**
   - **Current**: 46.9%
   - **Target**: ≥ 70%

2. **`internal/tools/docs`**
   - **Current**: 37.9%
   - **Target**: ≥ 60%

3. **`internal/providers/migration/raw`**
   - **Current**: 33.3%
   - **Target**: ≥ 70%

> **Note**: These targets are **per-package internal goals**, not tied to the core coverage threshold used in the CI script. Do **not** change thresholds or scripts.

---

## Non-Goals (Out of Scope)

- Changing any coverage thresholds or behaviour in `scripts/check-coverage.sh`.
- Altering behaviour of core packages (`pkg/config`, `internal/core`) - those were Phase 1.
- Introducing new public APIs or user-visible features.
- Large refactors or re-organizing packages.
- Adding new external dependencies (libraries, tools).

If a test need suggests a refactor, keep it **surgical and local**, and only when required to make behaviour testable.

---

## Package-Specific Objectives

### 1. `internal/git` (Target ≥ 70%)

**Intent**: Improve confidence in git integration behaviour by covering existing branches and error paths.

**Focus Areas** (based on typical git helper patterns):

- **Happy-path flows:**
  - Repository detection
  - Branch name resolution
  - Dirty/clean working tree detection

- **Error paths:**
  - Non-git directory handling
  - Command execution failures (mocked or faked)
  - Malformed git output (where applicable)

- **Determinism:**
  - Ensure tests assert exact, stable outputs and error messages.

**Strategy**:

- Use dependency injection or small internal helpers to avoid spawning real git where possible.
- If shelling out is currently hard-coded, introduce **minimal** internal abstraction only if absolutely needed for testability, and keep behaviour identical.

---

### 2. `internal/tools/docs` (Target ≥ 60%)

**Intent**: Ensure documentation tool behaviour is validated, especially where it touches specs, features, or derived docs.

**Focus Areas**:

- **Happy-path processing:**
  - Loading and parsing documentation/spec inputs.
  - Generating derived artifacts (indexes, summaries, or validations).

- **Failure paths:**
  - Missing or unreadable input files.
  - Invalid document structures.
  - Mismatched references (e.g., missing spec files).

- **Determinism:**
  - File walking must use deterministic ordering (sorted).
  - Outputs must be stable given the same repo state.

**Strategy**:

- Add tests that run the smallest possible slice of docs tooling over tiny, synthetic fixtures (testdata).
- Avoid relying on full-repo state; prefer local `testdata/` directories where practical.

---

### 3. `internal/providers/migration/raw` (Target ≥ 70%)

**Intent**: Increase confidence in raw SQL migration engine behaviour - especially error handling and ordering.

**Focus Areas**:

- **Happy path:**
  - Discovering migrations from the expected location(s).
  - Applying migrations in deterministic, correct order.
  - Idempotent behaviour where specified by the engine design.

- **Error paths:**
  - Missing or unreadable migration files.
  - Invalid/parse-failing SQL migrations (when applicable).
  - Partial application failures and error surfacing.

- **Determinism:**
  - File reading order (sorted).
  - Stable error messages and logging patterns.

**Strategy**:

- Use temporary directories and synthetic SQL files under `testdata/`.
- Avoid real database connections; prefer mocks, fakes, or the smallest possible embedded test harness already present in the package.
- The goal is to exercise the engine control flow, not the database itself.

---

## Constraints (From Agent.md & GOV_V1_CORE)

### You MUST:

- Treat **GOV_V1_CORE** as the single Feature ID for this work.
- Follow spec-first, test-first where any behavioural ambiguity is encountered.
- Keep diffs **minimal and deterministic**.
- Keep changes scoped strictly to:
  - `internal/git`
  - `internal/tools/docs`
  - `internal/providers/migration/raw`
  - and their **direct** test files and `testdata/` folders.

### You MUST NOT:

- Change any coverage thresholds or scripts.
- Change CLI behaviour or external interfaces.
- Introduce new dependencies.
- Modify protected files (LICENSE, top-level README, ADRs, global governance docs).
- Reorganize packages or move files across directories.

---

## Success Criteria

Phase 2 is complete when **all** of the following hold:

1. **Per-package coverage targets reached:**
   - `internal/git` ≥ 70% line coverage.
   - `internal/tools/docs` ≥ 60% line coverage.
   - `internal/providers/migration/raw` ≥ 70% line coverage.

2. **Test suite health:**
   - `go test ./...` passes.
   - `./scripts/run-all-checks.sh` passes.
   - `./scripts/check-coverage.sh --fail-on-warning` still passes (no new violations introduced).

3. **Behavioural stability:**
   - No user-facing behaviour changes, except clearly documented bug fixes (if any are found).
   - No changes to CI thresholds or coverage scripts.

4. **Governance alignment:**
   - All new/updated test files include proper Feature ID header:
     ```go
     // Feature: GOV_V1_CORE
     // Spec: spec/governance/GOV_V1_CORE.md
     ```
   - Changes are traceable to GOV_V1_CORE as the governance/quality feature.

---

## Execution Checklist

### Common Setup

- [ ] Confirm Feature ID: `GOV_V1_CORE`
- [ ] Verify hooks (`./scripts/install-hooks.sh` if needed)
- [ ] Ensure clean working directory
- [ ] On appropriate feature branch (e.g. `test/GOV_V1_CORE-coverage-phase2`)

---

### `internal/git`

- [ ] Inventory all public functions and key internal helpers.
- [ ] Add tests covering:
  - [ ] Normal repository detection and branch resolution.
  - [ ] Dirty vs clean tree detection.
  - [ ] Non-git directory handling and error paths.
  - [ ] Command failure/malformed output handling (using fakes/mocks as needed).
- [ ] Re-run coverage and confirm `internal/git` ≥ 70%.

---

### `internal/tools/docs`

- [ ] Identify main doc-processing entry points.
- [ ] Create minimal fixtures under `internal/tools/docs/testdata/` (if not already present).
- [ ] Add tests covering:
  - [ ] Successful processing over a tiny, representative input.
  - [ ] Missing/invalid inputs.
  - [ ] Deterministic ordering and output.
- [ ] Re-run coverage and confirm `internal/tools/docs` ≥ 60%.

---

### `internal/providers/migration/raw`

- [ ] Identify migration discovery and execution paths.
- [ ] Create minimal SQL fixtures under `internal/providers/migration/raw/testdata/`.
- [ ] Add tests covering:
  - [ ] Happy-path migration application sequence.
  - [ ] Missing/malformed migration files.
  - [ ] Error handling when a migration fails.
- [ ] Re-run coverage and confirm `internal/providers/migration/raw` ≥ 70%.

---

### Final Validation

- [ ] Run `./scripts/run-all-checks.sh`
- [ ] Run `./scripts/check-coverage.sh --fail-on-warning`
- [ ] Capture updated coverage summary (for future `docs/coverage/TEST_COVERAGE_ANALYSIS.md` refresh)
- [ ] Ensure working directory is clean and branch is ready for PR

---

## Reference Documents

- `docs/coverage/TEST_COVERAGE_ANALYSIS.md` – Original coverage analysis and metrics
- `docs/coverage/COVERAGE_COMPLIANCE_PLAN.md` – Phase 1 and Phase 2 framing
- `docs/coverage/COVERAGE_COMPLIANCE_PLAN_PHASE2.md` – Detailed Phase 2 plan
- `docs/agents/AGENT_BRIEF_COVERAGE_PHASE1.md` – Phase 1 compliance brief
- `spec/governance/GOV_V1_CORE.md` – Governance specification
- `Agent.md` – Development protocol and constraints


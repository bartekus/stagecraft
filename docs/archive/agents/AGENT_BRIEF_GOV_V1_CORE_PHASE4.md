> **Superseded by** `docs/engine/agents/AGENT_BRIEF_GOV_V1_CORE.md` section "Phase 4: Feature Mapping Invariant Enforcement". Kept for historical reference. New GOV_V1_CORE execution notes MUST go into the merged agent brief.

# AGENT BRIEF — GOV_V1_CORE — Phase 4

## Multi-Feature Cross-Validation & Feature Mapping Invariant Enforcement

**Status:** In progress — Scaffold complete, enforcement pending

**Feature ID:** GOV_V1_CORE

**Spec:** spec/governance/GOV_V1_CORE.md

**Current Phase:** Scaffold implemented — `cmd/feature-map-check` and `internal/tools/features` package are functional. Enforcement and CI integration are tracked in follow-up work.

---

## 🎯 Mission

Implement **Phase 4** of GOV_V1_CORE governance hardening:

> Enforce the **Feature Mapping Invariant** across specs, features.yaml, implementation code, and tests — deterministically and in CI.

This phase introduces a new tool and CI rule that ensures:

1. Every Feature ID has exactly **one spec**.

2. Every spec corresponds to exactly one Feature ID.

3. Every implementation file contains the correct `Feature:` + `Spec:` headers.

4. Tests reference the correct Feature ID.

5. Every `done` or `wip` feature has the required supporting artifacts.

6. No dangling Feature IDs or orphan specs exist anywhere in the repository.

This phase **strengthens repository integrity and traceability**, ensuring that Stagecraft remains a fully-spec'd, test-first, deterministic system.

---

## 🧱 Scope (What is included)

### ✔ Must implement

- A new Go tool:

  `cmd/feature-map-check` or `internal/tools/features/mapcheck`

- Deterministic feature-mapping analysis:

  - Parse `spec/features.yaml`

  - Parse all spec files under `spec/**`

  - Parse all `.go` files (excluding tests and testdata)

  - Parse all test files under `*_test.go`

### ✔ Validate per Feature Status

- **todo**

  - Warnings allowed

  - Missing spec allowed

  - No code/tests required

- **wip**

  - Spec MUST exist

  - At least one implementation OR test MUST exist

  - Incorrect `Spec:` headers = hard error

- **done**

  - Spec MUST exist

  - Implementation MUST reference correct `Feature:` & `Spec:`

  - Tests MUST reference correct `Feature:`

  - No mismatches allowed

### ✔ Output

- Deterministic sorted results

- CI-enforceable error reports

- Zero false positives

---

## 🚫 Out of Scope

- No behavioral changes to CLI commands

- No rewriting of specs themselves

- No provider logic changes

- No multi-feature refactors

- No modification of docs outside governance

Everything stays within **governance tooling**.

---

## 🧪 Test Requirements

The new tool MUST include:

### Unit Tests

- Missing spec errors

- Duplicate spec mapping error

- Implementation missing Feature header

- Tests missing Feature header

- Mismatched spec paths

- Dangling Feature IDs

- Orphan specs

- todo/wip/done status rules

- Deterministic ordering tests

### Golden Tests (optional but recommended)

- Full output of error reports for multiple features

- Stable ordering enforced

### Integration Tests (via run-all-checks.sh)

- Tool executes and returns correct exit code

- Hooks into CI with no nondeterminism

---

## 🧩 Constraints

- MUST be deterministic

- MUST use stdlib only

- MUST follow existing spec-parser/feature-parser patterns

- MUST be integrated into `scripts/run-all-checks.sh`

  **before docs checks** and after spec validation

- MUST NOT modify behavior outside governance

- MUST NOT modify provider config or code

- MUST NOT produce nondeterministic filesystem traversal

---

## ✔ Success Criteria

The phase is complete when:

- All feature → spec → code → test mappings are validated

- CI fails on invalid mappings for **done** features

- Warnings for **todo** features are non-blocking

- Complete tooling + tests exist

- Fully deterministic, reproducible outputs

- No false positives remain

- The repository meets the Feature Mapping Invariant globally

---

## 📎 Execution Checklist

### Pre-work

- [x] Confirm Feature ID: `GOV_V1_CORE`

- [x] Clean branch under:

  `feature/GOV_V1_CORE-phase4-feature-mapping`

- [x] Hook verification

- [x] Read spec and plan documents

### Implementation (Scaffold — Complete)

- [x] Create tool: `cmd/feature-map-check`

- [x] Implement core scanner logic

- [x] Implement status-aware validation rules

- [x] Implement deterministic error reporting

- [x] Add tests (unit scaffolding)

- [ ] Integrate into run-all-checks.sh (follow-up)

### Verification (Scaffold — Complete)

- [x] Run goformat

- [x] Run go build ./... (PASSES)

- [x] Run go test ./... (PASSES)

- [x] Manual verification: `go run ./cmd/feature-map-check` (runs and detects issues)

- [ ] Run run-all-checks.sh with integration (follow-up)

- [ ] Commit with correct message format

- [ ] Produce PR with spec, tests, rationale

### Follow-up Work (Enforcement & Rollout)

See follow-up issue: **[GOV_V1_CORE][Phase 4] Feature Mapping Invariant enforcement & rollout**

- [ ] CI integration into `scripts/run-all-checks.sh`

- [ ] Repository alignment (add missing headers, fix mismatches)

- [ ] Test suite expansion (golden tests, comprehensive unit tests)

- [ ] Governance flip to strict mode (hard-fail on violations)

---

## 📌 Notes for the Agent

- DO NOT guess missing behavior

- STOP if spec parsing is ambiguous

- Thread the Feature ID through ALL commits

- All rule enforcement is deterministic

- Minimal diffs only

- No unrelated cleanup

**Proceed only after this brief is approved.**


# Governance Almanac

> Canonical reference for Stagecraft governance.
> This document replaces scattered governance checklists, commit discipline notes, and provider governance summaries.
>
> **This document is canonical and should evolve slowly; changes here usually imply governance or workflow shifts.**

## 1. Purpose and Scope

The Governance Almanac:

- Defines how governance works across Stagecraft
- Describes commit and PR discipline
- Describes provider governance and coverage enforcement
- Explains documentation lifecycle and ownership
- Links to specs, ADRs, and tooling that enforce governance

It consolidates content that previously lived in (non exhaustive):

- `docs/governance/GOV_V1_CORE_PHASE3_PLAN.md`
- `docs/governance/GOV_V1_TEST_REQUIREMENTS.md`
- `docs/governance/COMMIT_GUIDANCE_PROVIDER_GOVERNANCE.md`
- `docs/governance/COMMIT_MESSAGE_ANALYSIS.md`
- `docs/governance/COMMIT_READY_SUMMARY.md`
- `docs/governance/CI_PROVIDER_COVERAGE_ENFORCEMENT.md`
- `docs/governance/PROVIDER_GOVERNANCE_SUMMARY.md`
- `docs/governance/PHASE5_VIOLATION_FIX_CHECKLIST.md`
- `docs/governance/STRATEGIC_DOC_MIGRATION.md`
- `docs/governance/INFRA_UP_SPEC_FIX.md`
- `docs/governance/PROVIDER_*_COVERAGE_PLAN.md`
- `docs/todo/COMMIT_MESSAGE_ENFORCEMENT_PHASE1.md`
- `docs/todo/COMMIT_MESSAGE_ENFORCEMENT_PHASE2.md`

---

## 2. Core Governance Principles

> Short list of stable principles that govern the project.

- Spec first, test first, deterministic behaviour
- One feature per PR and per commit scope
- Traceability from spec to tests to code to docs to commit
- No hidden behaviour and no untracked changes
- Governance enforced by tools, not memory

### 2.1 Governance Specs and ADRs

- Governance spec: `spec/governance/GOV_V1_CORE.md`
- ADRs:
  - `spec/adr/0002-docs-lifecycle-and-ownership.md`
  - `spec/adr/0001-architecture.md`
  - Other ADRs as they are added

---

## 3. Commit and PR Discipline

### 3.1 Commit Message Rules

> Summarize the rules from commit message enforcement docs and Agent.md.

- **Format**: `<type>(<FEATURE_ID>): <summary>`
- **Allowed types**: `feat`, `fix`, `refactor`, `docs`, `test`, `ci`, `chore`
- **`<FEATURE_ID>`**: SCREAMING_SNAKE_CASE (e.g., `PROVIDER_FRONTEND_GENERIC`)
- **Summary constraints**:
  - ≤ 72 characters
  - No trailing period
  - Lowercase after colon
  - No emojis or unicode decorations
  - ASCII-only characters in subject line
  - Literal, precise, minimal descriptions

- **Why this matters**: Commit messages are a critical link in Stagecraft's deterministic traceability chain: `spec → tests → code → docs → commit → PR`. Without proper format, traceability breaks, feature lifecycle integrity fails, and automation fails.

- **Enforcement**: Git hook `.hooks/commit-msg` validates format (can be bypassed with `STAGECRAFT_SKIP_HOOKS=1` or `SKIP_HOOKS=1`, but this should be avoided)

- **Examples**:
  - ✅ Valid: `feat(PROVIDER_FRONTEND_GENERIC): implement provider`
  - ✅ Valid: `fix(PROVIDER_FRONTEND_GENERIC): address review feedback`
  - ❌ Invalid: `feat: implement PROVIDER_FRONTEND_GENERIC` (missing parentheses)
  - ❌ Invalid: `feat(CLI_PLAN, CLI_DEPLOY): refactor` (multiple Feature IDs - forbidden)

### 3.2 Commit Preconditions

- Clean working tree
- Hooks installed and passing
- `./scripts/run-all-checks.sh` passing
- Feature state and spec alignment confirmed

### 3.3 PR Requirements

- PR title: `[FEATURE_ID] Short human description`
- PR body:
  - Feature ID
  - Spec reference
  - Tests reference
  - Summary and rationale
  - Constraints and limitations
- All tests and governance checks green before review

---

## 4. Provider Governance

### 4.1 Provider Lifecycle

> Summarize rules from provider governance docs.

- Providers must have:
  - Spec
  - Analysis brief
  - Implementation outline
  - Coverage plan
- Provider registration and configuration rules:
  - Registry based
  - No hardcoded providers in core
  - Opaque config

### 4.2 Provider Coverage Requirements

> Integrates with `docs/coverage/COVERAGE_LEDGER.md`.

- **Minimum coverage thresholds**:
  - Providers: ≥ 80% target for v1 complete
  - Minimum acceptable: 75% for v1
  - Stretch goal: 85%+

- **Enforcement rules in CI**:
  - `scripts/check-provider-governance.sh` validates coverage strategy presence
  - All providers marked `done` MUST have `COVERAGE_STRATEGY.md`
  - Providers claiming "V1 Complete" MUST have corresponding status document
  - Coverage threshold checking can be added to CI (future enhancement)

- **Expected slice or phase structure to reach v1**:
  - Providers typically use slice-based approach (Slice 1, Slice 2, etc.)
  - Each slice focuses on specific coverage areas (helpers, error paths, etc.)
  - See `PROVIDER_NETWORK_TAILSCALE_EVOLUTION.md` for example slice structure

### 4.3 Provider Specific Notes

> Short index that points to evolution logs and specific decisions.

- `PROVIDER_NETWORK_TAILSCALE`:
  - Evolution log: `docs/engine/history/PROVIDER_NETWORK_TAILSCALE_EVOLUTION.md`
  - Status: v1 plan (79.6% coverage, 2 slices complete)
- `PROVIDER_FRONTEND_GENERIC`:
  - Evolution log: `docs/engine/history/PROVIDER_FRONTEND_GENERIC_EVOLUTION.md`
  - Status: v1 complete (87.7% coverage, reference model)
- `PROVIDER_BACKEND_GENERIC`:
  - Evolution log: `docs/engine/history/PROVIDER_BACKEND_GENERIC_EVOLUTION.md`
  - Status: v1 complete (84.1% coverage)
- `PROVIDER_BACKEND_ENCORE`:
  - Evolution log: `docs/engine/history/PROVIDER_BACKEND_ENCORE_EVOLUTION.md`
  - Status: v1 complete (90.6% coverage)
- `PROVIDER_CLOUD_DO`:
  - Evolution log: `docs/engine/history/PROVIDER_CLOUD_DO_EVOLUTION.md`
  - Status: v1 complete (80.5% coverage)

---

## 5. Testing and Coverage Governance

> High level rules that complement the coverage ledger.

- **All behaviour changes must include tests**: Happy path, failure path, and edge conditions required

- **Determinism first**:
  - No `time.Sleep()` in tests (except explicitly documented integration scenarios)
  - No goroutine-based tests without proper synchronization
  - No OS-level behavior dependencies (pipe buffering, process scheduling)
  - No randomness or timestamps in test logic
  - All tests must pass with `-race` and `-count=20` for determinism verification

- **Separation of concerns**:
  - Unit tests cover: Pure functions, isolated components, error paths, deterministic state transitions
  - Integration tests cover: Process lifecycle, provider interface contracts, end-to-end workflows
  - Integration tests MUST NOT validate logic that can be unit-tested deterministically

- **No test seams**: If a component requires a "test seam" (injectable dependency) to be testable, the design is incomplete. Preferred: Extract pure, testable primitives.

- **Coverage thresholds**:
  - Core packages (`pkg/config`, `internal/core`): Minimum 80%, target 85%+
  - Provider implementations: Minimum 75% (acceptable for v1), target 80%+, stretch 85%+
  - Interface definitions (`pkg/providers/*`): Minimum 90%

- **Golden file rules**: TBD (location, update policy, review expectations)

- **Coverage thresholds and failure behaviour in CI**: See `CI_PROVIDER_COVERAGE_ENFORCEMENT.md`

See also: `docs/coverage/COVERAGE_LEDGER.md` and `docs/governance/GOV_V1_TEST_REQUIREMENTS.md`.

---

## 6. Documentation Governance

### 6.1 Doc Types and Ownership

> Base this on `docs/README.md` and docs lifecycle ADR.

- **Narrative docs**: under `docs/narrative/` - Human-facing planning and overview
- **Engine docs** (analysis, outlines, status, history): under `docs/engine/` - Implementation-aligned, AI-critical
- **Governance docs**: under `docs/governance/` - Process, discipline, workflow
- **Coverage docs**: under `docs/coverage/` - Coverage tracking and compliance
- **Context handoff and AI collaboration docs**: under `docs/context-handoff/` - Multi-step task context (consolidating into `CONTEXT_LOG.md`)
- **Archive docs**: under `docs/archive/` - Historical, superseded docs

- **Source of Truth Hierarchy**:
  1. `spec/` - Behavioral truth (what the system does)
  2. `docs/engine/` - Implementation truth (how to build it)
  3. `docs/narrative/` - Planning truth (why and when)
  4. `docs/governance/` - Process truth (how we work)
  5. `docs/archive/` - Historical record (how we did it)

### 6.2 Lifecycle

- **Active docs**: Spec aligned and maintained
- **Historical docs**: Marked as superseded, kept for reference (prepend "Superseded by..." notice)
- **Archived docs**: Moved under `docs/archive/` when no longer needed for day to day work

- **Frontmatter tracking**: All non-spec documentation uses frontmatter to track lifecycle:
  ```yaml
  ---
  status: active | canonical | archived
  scope: v1 | v2 | meta
  feature: CLI_PLAN          # optional
  spec: ../spec/commands/plan.md  # optional
  superseded_by: ../Agent.md # optional
  ---
  ```

---

## 7. Violation Handling and Fixes

> Summarize PHASE5 violation fix checklist and related docs.

- **Detection sources**:
  - Validation reports (e.g., `VALIDATION_REPORT.md`)
  - Governance tools (`scripts/check-provider-governance.sh`, `scripts/validate-feature-integrity.sh`)
  - CI workflows (governance checks, coverage checks)

- **Common violation types**:
  - Spec status mismatches (frontmatter status doesn't match `spec/features.yaml`)
  - Feature ID header mismatches (wrong Feature ID in implementation files)
  - Orphan spec references (feature references non-existent spec file)
  - Missing test files (feature marked `done` but test file missing)
  - Coverage strategy missing (provider marked `done` but no `COVERAGE_STRATEGY.md`)

- **Fix process**:
  1. Identify scope (which files/features affected)
  2. Confirm spec alignment (verify against `spec/features.yaml` and actual spec files)
  3. Apply fix in minimal diff (fix only what's needed, preserve existing content)
  4. Update docs and status (update status documents, mark as fixed)

- **Reporting**: Capture fixes in validation or status docs (e.g., update `VALIDATION_REPORT.md` or provider status docs)

---

## 8. Tooling and Automation

> Index of governance related tooling.

- `./scripts/check-provider-governance.sh`
- `./scripts/validate-feature-integrity.sh`
- `./scripts/check-required-tests.sh`
- CI workflows:
  - `.github/workflows/docs-governance.yml`
  - `.github/workflows/ci.yml`
  - `.github/workflows/nightly.yml`

---

## 9. Migration Notes

> Temporary section while folding in legacy governance content.

- [x] Migrate commit message enforcement docs
- [x] Migrate provider governance summary
- [x] Migrate coverage enforcement docs
- [x] Migrate violation fix checklist and status
- [x] Link evolution logs and coverage ledger

Once complete, this checklist can be removed or marked as done.

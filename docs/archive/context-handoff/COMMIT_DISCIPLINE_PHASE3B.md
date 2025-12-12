> **Superseded by** `docs/context-handoff/CONTEXT_LOG.md` section 4.3. Kept for historical reference. New context handoffs MUST be added to the context log.

# 🔥 Agent Promo: Phase 3.B – Commit Health Generators & CLI Integration

**Task:** Implement Phase 3.B of the commit discipline system using the Phase 3.A type layer as the foundation.

**Feature IDs:**

- `PROVIDER_FRONTEND_GENERIC`

- `GOV_V1_CORE`

**Prerequisites:**

- Phase 3.A complete:

  - Go type layer for commit health / traceability reports

  - Golden roundtrip tests for all report types

  - Specs and JSON schemas aligned (see `COMMIT_REPORT_TYPES_PHASE3.md` and associated spec files)

- `COMMIT_DISCIPLINE_PHASE3.md` read and understood (overall mission + non-goals)

- All work follows top-level `Agent.md` rules (spec-first, test-first, deterministic, provider-agnostic, single-feature PRs)

⸻

## 🎯 Mission (Phase 3.B)

Turn the static type layer from Phase 3.A into **working, deterministic generators** wired into the CLI:

1. **Commit history → commit-health report**

2. **Repo + feature map → feature-traceability report**

3. **Internal helper(s) for commit suggestions (optional but recommended)**

4. **CLI entrypoints that produce the JSON reports under `.stagecraft/reports/`**

**Key constraint:**

All logic MUST be deterministic and testable without relying on ambient git config or machine-specific state.

⸻

## 📦 Scope (Phase 3.B Implementation)

The agent MUST deliver the following, in this order:

### 1. Commit History Scanner → Commit Health Report

Implement a pure(ish) generator that, given a deterministic view of commit history, produces a `CommitHealthReport` (actual type name as defined in Phase 3.A).

**Responsibilities:**

- Define a small internal abstraction for commit history, e.g.:

  ```go
  // internal/reports/commithealth/history.go
  
  type CommitMetadata struct {
      SHA        string
      Message    string
      AuthorName string
      AuthorEmail string
      // ... anything needed that is deterministic and spec-approved
  }
  
  type HistorySource interface {
      Commits() ([]CommitMetadata, error)
  }
  ```

- Implement a generator function (exact names may differ; follow spec and Phase 3.A):

  ```go
  // internal/reports/commithealth/generate.go
  
  func GenerateCommitHealthReport(commits []CommitMetadata, features FeatureRegistry) (CommitHealthReport, error) {
      // No I/O here; pure analysis.
  }
  ```

- The generator MUST:

  - Classify commits according to the rules in `COMMIT_DISCIPLINE_PHASE3.md`:

    - Valid/invalid commit message format

    - Missing or unknown Feature IDs

    - Legacy format detection (if defined in spec)

    - Branch naming anomalies (where derivable from commit data or inputs)

  - Populate all required aggregate fields:

    - Counts per violation type

    - Per-violation commit SHA lists

    - Any summary metrics defined by the spec

**Important:**

The generator MUST NOT shell out to git. Git access (if any) belongs in a thin adapter used by CLI commands, not in the core generator logic.

⸻

### 2. Feature Traceability Index → Feature Traceability Report

Implement a second generator that produces a feature-level traceability report.

**Inputs (deterministic):**

- Feature registry derived from `spec/features.yaml`

- Code/test/spec scan results, e.g.:

  ```go
  // internal/reports/traceability/inputs.go
  
  type FeaturePresence struct {
      FeatureID      string
      HasSpec        bool
      HasImplementation bool
      HasTests       bool
      HasCommits     bool
  }
  ```

- You MAY define one or more scanning helpers that:

  - Walk the repo tree (with lexicographically sorted directory traversal)

  - Detect Feature IDs in:

    - Spec headers

    - Implementation header comments

    - Test header comments

    - Commit messages (via the commit history abstraction)

  - These helpers MUST be deterministic and MUST NOT depend on git config or environment variables.

**Generator responsibilities:**

- Implement:

  ```go
  // internal/reports/traceability/generate.go
  
  func GenerateFeatureTraceabilityReport(features []FeaturePresence) (FeatureTraceabilityReport, error) {
      // Maps raw presence data into the Phase 3.A report type.
  }
  ```

- This function MUST:

  - Satisfy the requirements of `COMMIT_DISCIPLINE_PHASE3.md`, section "Traceability Gap Detection":

    - Per-feature fields: spec presence, tests, implementation, commits

    - Produce a structure that serializes to `.stagecraft/reports/feature-traceability.json`

  - Be purely computational (no I/O, no git commands, no filesystem operations).

⸻

### 3. CLI Integration & I/O Adapters

Wire the generators into one or more CLI commands as defined by the spec (do NOT invent new commands; follow the existing or updated spec).

**Responsibilities:**

- Add internal adapters that:

  - Collect real-world inputs:

    - Commit history from git

    - Repo scan for Feature IDs

    - Feature registry from `spec/features.yaml`

  - Pass them into the pure generators.

- Serialize and write the results into:

  - `.stagecraft/reports/commit-health.json`

  - `.stagecraft/reports/feature-traceability.json`

- I/O layer constraints:

  - All filesystem operations MUST:

    - Use deterministic paths

    - Use atomic writes (temp file + rename) following Stagecraft patterns

  - If git is used:

    - Shelling out MUST be confined to a small, well-documented internal helper (e.g. `internal/git`), NOT core packages.

    - Output parsing MUST be deterministic.

    - Tests MUST NOT require a real git repository; use fakes, stubs, or in-repo fixtures instead.

  - CLI commands:

    - Follow existing naming and layout conventions in `cmd/` and the relevant specs.

    - Commands MUST:

      - Exit with deterministic codes

      - Print deterministic log messages (or none, if reports are file-only)

⸻

### 4. Optional: Feature-Aware Commit Suggestion Helper

Implement the suggestion helper described in `COMMIT_DISCIPLINE_PHASE3.md` as a pure function:

- **Inputs:**

  - Current branch name

  - Changed paths

  - Feature registry (from spec)

- **Output:**

  - Suggested commit message skeleton (struct or plain string) or an explicit "no suggestion" result.

**Constraints:**

- MUST only use Feature IDs that exist in `spec/features.yaml`.

- MUST prefer features whose spec paths or known code/test paths overlap with the changed files.

- MUST be deterministic (same input → same suggestion).

- MUST NOT be enforced in CI.

This can live under `internal/reports/commithealth/suggest.go` or a similarly named file.

⸻

### 5. Documentation Updates

Update the following docs:

- `docs/COMMIT_MESSAGE_ANALYSIS.md`

  - Add a "Phase 3.B – Generators & Reports" section.

  - Document:

    - What each report contains

    - How to run the commands

    - How to interpret the outputs

    - Limitations and non-goals (no auto-rewrite of history)

- `docs/guides/AI_COMMIT_WORKFLOW.md`

  - Add a section "Using Commit Health & Traceability Reports".

  - Show how AI/humans should:

    - Run the reports

    - Use them to improve commit hygiene

    - Identify traceability gaps

⸻

## 🧪 Testing Requirements (Phase 3.B)

The agent MUST implement:

1. **Unit tests for generators**

   - `GenerateCommitHealthReport`:

     - Synthetic commit histories covering:

       - Valid/invalid messages

       - Known/unknown Feature IDs

       - Edge cases (no commits, only invalid commits, mixed histories)

   - `GenerateFeatureTraceabilityReport`:

     - Synthetic feature presence sets for:

       - Fully complete features (spec + impl + tests + commits)

       - Missing-spec features

       - Missing-tests features

       - Orphan specs / orphan commits

2. **Golden tests for JSON reports**

   - Store golden files under appropriate `testdata/` dirs, e.g.:

     - `internal/reports/commithealth/testdata/commit-health_report.json`

     - `internal/reports/traceability/testdata/feature-traceability_report.json`

   - Tests MUST:

     - Call generators with deterministic synthetic inputs

     - Serialize to JSON using the same encoder used in production

     - Compare against golden files byte-for-byte

3. **Determinism & isolation**

   - No tests may depend on:

     - Local git config

     - Author name/email

     - Machine path prefixes

   - If filesystem traversal is required:

     - Use in-test fixtures under `testdata/`

     - Force lexicographical ordering for directory reads

4. **End-to-end smoke tests (optional but recommended)**

   - If CLI commands are introduced in 3.B:

     - Add one or more integration tests invoking the command(s) with a controlled test repo fixture.

     - Assert that the expected report files are created and match expected content (or at least basic invariants).

⸻

## 🧱 Development Flow (Agent Checklists)

The agent MUST:

1. **Feature & branch setup**

   - Confirm Feature IDs: `PROVIDER_FRONTEND_GENERIC`, `GOV_V1_CORE`

   - Use a feature branch, e.g.:

     - `feature/PROVIDER_FRONTEND_GENERIC-commit-discipline-phase3b`

   - Verify hooks per `Agent.md` (Hook Verification section).

   - Ensure clean working directory before starting.

2. **Spec alignment**

   - Re-read:

     - `COMMIT_DISCIPLINE_PHASE3.md`

     - `COMMIT_REPORT_TYPES_PHASE3.md`

     - Any relevant spec files referenced there (for reports & CLI commands)

   - Confirm behaviour is fully defined.

   - If spec gaps exist:

     - Propose exact wording snippets for human approval (per Spec Interpretation Rules).

3. **Test-first implementation**

   - Introduce failing unit + golden tests for:

     - Commit health generator

     - Feature traceability generator

   - Only then implement the generators.

   - Only after generators are stable and tested:

     - Add CLI wiring and I/O adapters.

   - Finally, update documentation.

4. **Pre-commit checks**

   - Run:

     - `./scripts/goformat.sh`

     - `./scripts/run-all-checks.sh`

   - Fix all issues before committing.

5. **Commit discipline**

   - Use commit messages like:

     ```
     feat(PROVIDER_FRONTEND_GENERIC): add commit health generators
     
     Spec: <spec path or COMMIT_DISCIPLINE_PHASE3.md>
     Tests: internal/reports/commithealth/..._test.go
     ```

   - Follow all rules in `Agent.md` ("Commit Message Enforcement & Discipline").

6. **PR guidance**

   - PR Title:

     ```
     [PROVIDER_FRONTEND_GENERIC] Phase 3.B – commit generators & reports
     ```

   - PR Body (minimum):

     - Feature: `PROVIDER_FRONTEND_GENERIC`

     - Governance: `GOV_V1_CORE`

     - Summary of implemented generators + CLI integration

     - Spec/docs references

     - Test coverage summary

     - Explicit note: "Phase 3.B is read-only analysis; no history rewriting."

⸻

## ⚠️ Non-Goals (Phase 3.B)

- No rewriting or mutating git history.

- No automatic commit message fixes.

- No enforcement changes to existing Phase 1/2 validation logic.

- No schema changes to report types (that would be a separate, spec-driven change).

- No new external dependencies without ADR and explicit human approval.


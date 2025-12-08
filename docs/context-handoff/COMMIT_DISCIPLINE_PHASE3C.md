# 🔥 Agent Promo: Phase 3.C – CLI Wiring for Commit Discipline Reports

**Task:** Implement Phase 3.C of the commit discipline system: integrate the Phase 3.B generators into stable, deterministic, Stagecraft CLI commands.

**Feature IDs:**

- `PROVIDER_FRONTEND_GENERIC`

- `GOV_V1_CORE`

**Prerequisites:**

- Phase 3.A complete (types + golden tests)

- Phase 3.B complete (generators + golden tests)

- Specs for command behavior exist:

  - `COMMIT_DISCIPLINE_PHASE3.md`

  - `COMMIT_REPORT_TYPES_PHASE3.md`

- All work MUST follow `Agent.md`:

  - Spec-first, test-first

  - Deterministic I/O

  - Provider-agnostic boundaries

  - No guessing; no touching protected files

⸻

## 🎯 Mission (Phase 3.C)

Expose the commit discipline reporting system as first-class Stagecraft CLI commands that:

1. **Collect deterministic input:**

   - Git commit history

   - Feature registry (from `spec/features.yaml`)

   - Repo tree scan (deterministic lexicographical walk)

2. **Pass that input into the Phase 3.B pure generators**

3. **Produce JSON report files under:**

   - `.stagecraft/reports/commit-health.json`

   - `.stagecraft/reports/feature-traceability.json`

4. **Provide deterministic CLI UX:**

   - Zero nondeterministic logs

   - Stable flag ordering

   - Stable exit codes

This phase does not change generator behavior — only wiring and CLI contract.

⸻

## 📦 Scope (Phase 3.C Implementation)

The agent MUST deliver the following outputs, in this exact order:

⸻

### 1. CLI Commands (spec-defined)

Create or update CLI commands for:

**A. `stagecraft commit report`**

Generates:

- `.stagecraft/reports/commit-health.json`

**Responsibilities:**

- Read commit history via deterministic git adapter

- Load feature registry

- Pass into `GenerateCommitHealthReport`

- Serialize using deterministic encoder

- Atomically write report

**B. `stagecraft feature traceability`**

Generates:

- `.stagecraft/reports/feature-traceability.json`

**Responsibilities:**

- Walk repo tree (deterministic lexicographical traversal)

- Collect `FeaturePresence` information

- Pass into `GenerateFeatureTraceabilityReport`

- Serialize using deterministic encoder

- Atomically write report

**Constraints:**

- Commands MUST NOT do any interpretation of commit discipline rules; generators own all behavior

- Commands MUST have:

  - Deterministic logs (or no logs)

  - Stable ordering of flags

  - Deterministic behavior given identical repo state

No new flags unless explicitly spec'd.

⸻

### 2. Git Adapter (Thin, Deterministic)

Add an internal package at:

- `internal/git`

with a single responsibility:

- Provide `HistorySource` implementation for commit retrieval

**Rules:**

- MUST shell out only using `exec.CommandContext`

- MUST set all environment variables explicitly (never inherit env)

- MUST parse with fully deterministic rules

- MUST sort commits lexicographically by commit time or SHA (spec behavior)

- MUST NOT expose timestamps; only order them deterministically

**Tests MUST use:**

- Fixtures, not actual git

- Fake `HistorySource` for unit tests

- Golden files where appropriate

⸻

### 3. Repository Scanner

Add helper under:

- `internal/reports/featuretrace/scan.go`

**Responsibilities:**

- Deterministically traverse repo tree

- Extract Feature IDs from:

  - Spec headers

  - Implementation headers

  - Test headers

**Rules:**

- MUST sort directories lexicographically

- MUST NOT depend on OS ordering

- MUST be pure: no behavior outside scan + string extraction

⸻

### 4. File Output Layer (Stable, Atomic)

Add deterministic writers under:

- `internal/reports/writer.go`

**Requirements:**

- Atomic writes using:

  - `.tmp` file

  - `os.Rename`

- Deterministic JSON encoding

- Error wrapping (`fmt.Errorf("context: %w", err)`)

⸻

### 5. CLI Tests (Golden + Integration)

The agent MUST add:

**A. Unit tests**

- Tests for git adapter (fake backend)

- Tests for scan helpers

- Tests for file writer

**B. Golden CLI tests**

Example:

- `cmd/commit_report_test.go`

- `cmd/feature_traceability_test.go`

- Run CLI in isolated temp repo

- Use fake git history

- Expect deterministic JSON output

- Compare with golden files

**C. No real git repo dependency**

- All git tests MUST use stubbed `HistorySource`

- Only a small number of integration tests may use temp git repos — and only if isolated, deterministic, and documented

⸻

### 6. Documentation Updates

Update:

**A. `docs/COMMIT_MESSAGE_ANALYSIS.md`**

New section: Phase 3.C: CLI Wiring

**B. `docs/guides/AI_COMMIT_WORKFLOW.md`**

Add:

- How to run the reports

- How to use the output to guide commit discipline

**C. `spec/features.yaml`**

After implementation:

- Mark Phase 3.C as done

⸻

## 🧪 Testing Requirements (Phase 3.C)

The agent MUST implement all of the following:

1. **Unit tests**

   - Git adapter: commit parsing

   - Repo scanner: deterministic ID extraction

   - Writer: atomic file behavior

2. **Golden tests**

   - End-to-end CLI report generation

   - JSON output MUST match goldens byte-for-byte

3. **Determinism**

   - Repo tree traversal sorted

   - Git commit order sorted

   - Flags alphabetically sorted

   - No timestamps

   - No nondeterministic logs

4. **Isolation**

   - Tests MUST NOT use real git config

   - Tests MUST NOT rely on machine paths

   - Tests MUST NOT use the network

⸻

## 🧱 Development Flow (Agent Checklist)

The agent MUST follow:

1. Feature branch created and verified

2. Spec alignment:

   - Confirm CLI contract matches existing spec

   - Confirm no spec drift

3. Write failing tests first

4. Implement minimal code to satisfy tests

5. Run all checks

6. Update docs

7. Update feature state

8. Commit with proper message

⸻

## ⚠️ Non-Goals (Phase 3.C)

- No changes to generator logic

- No changes to report schemas

- No new commit validation rules

- No history rewriting

- No auto-fix suggestions (reserved for Phase 3.D)

⸻

## 🏁 Output of Phase 3.C

After this phase completes, Stagecraft will have:

- Full commit discipline reporting pipeline:

  - Pure core logic (3.B)

  - Deterministic CLI generation (3.C)

  - Actionable reports under `.stagecraft/reports/*`

- Deterministic end-to-end developer experience

- Foundation for Phase 3.D (AI assistant integration + suggestions)


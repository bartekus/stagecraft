> **Superseded by** `docs/context-handoff/CONTEXT_LOG.md` section 4.3. Kept for historical reference. New context handoffs MUST be added to the context log.

# 🔥 Agent Promo: Phase 3 – Commit Intelligence & Historical Analysis

**Task:** Implement Phase 3 of the commit message discipline system, building on Phase 1 (local + hooks) and Phase 2 (CI + CLI).

**Feature IDs:**
- `PROVIDER_FRONTEND_GENERIC`
- `GOV_CORE`

**Prerequisites:**
- Phase 1 and Phase 2 completed and documented
- CI validation and `stagecraft validate-commit` behave as specified

⸻

## 🎯 Mission

Elevate commit discipline from pure validation to **intelligent insight**:
- Provide **feature-aware commit suggestions**
- Analyze **historical commit health**
- Surface **traceability gaps** between spec → tests → code → docs → git

**Goal:** Make commit hygiene observable and improvable over time without compromising determinism.

⸻

## 📦 Scope (Phase 3 Implementation)

The agent MUST:

### 1. Historical Commit Health Analysis (Read-only)

- Add a tooling entrypoint (internal only) to:
  - Scan commit history for:
    - Invalid or legacy-format commit messages
    - Missing Feature IDs
    - Orphan Feature IDs (no spec entry)
    - Branch naming anomalies
  - Generate a deterministic report file (e.g. `.stagecraft/reports/commit-health.json`)

- The report MUST include:
  - Counts per violation type
  - List of affected commit SHAs
  - Suggested remediation strategies (e.g. "future commits must adhere; history left intact")

- This tool MUST be read-only:
  - No history rewriting
  - No automatic fixes

### 2. Feature-Aware Commit Suggestions (Optional helper)

- Introduce an internal helper that, given:
  - Current branch name
  - Changed paths
  - Feature ID registry (`spec/features.yaml`)
  
  can propose a commit message skeleton, for example:

  ```text
  feat(CLI_DEPLOY): adjust deploy rollback behavior
  
  Spec: spec/commands/deploy.md
  Tests: cmd/deploy_test.go
  ```

- This helper MUST:
  - Only suggest Feature IDs that exist in `spec/features.yaml`
  - Prefer Feature IDs whose mapped spec paths overlap with changed files
  - Never guess new Feature IDs
  - Be deterministic (same diff and branch always produce the same suggestion)
- It is a suggestion only; NOT enforced in CI

### 3. Traceability Gap Detection

- Extend existing validation logic to detect gaps where:
  - A Feature ID appears in commits but:
    - No spec exists, or
    - No tests reference it, or
    - No implementation header comments reference it
- Produce a deterministic report:
  - `.stagecraft/reports/feature-traceability.json`
  - Per Feature ID:
    - spec presence
    - tests presence
    - implementation presence
    - commit presence
- No automatic fixing, reporting only

### 4. Documentation

- Add a Phase 3 section to:
  - `docs/COMMIT_MESSAGE_ANALYSIS.md`
  - `docs/guides/AI_COMMIT_WORKFLOW.md`
- Document:
  - Historical analysis
  - Commit suggestion helper
  - Traceability reports
  - Non-goals (no auto-rewrite)

⸻

## 🧪 Testing Requirements

- Unit tests for:
  - Commit history scanner logic
  - Feature-aware suggestion logic
  - Traceability gap detection
- Golden tests for:
  - JSON report files under `testdata/`
- All tests MUST be deterministic and independent of:
  - Local git config
  - Author name or email
  - Machine-specific paths

⸻

## 🧱 AI MUST Follow Standard Development Flow

1. Verify Feature IDs (`PROVIDER_FRONTEND_GENERIC`, `GOV_CORE`)
2. Use feature branch: `feature/PROVIDER_FRONTEND_GENERIC-commit-discipline-phase3`
3. Read:
   - `docs/COMMIT_MESSAGE_ANALYSIS.md`
   - `docs/guides/AI_COMMIT_WORKFLOW.md`
   - `spec/features.yaml`
4. Write failing tests for:
   - History analysis
   - Suggestion helper
   - Traceability report
5. Implement minimal logic to make tests pass
6. Update docs
7. Run `./scripts/run-all-checks.sh`
8. Commit with valid message
9. Provide PR title/body

⸻

## 📌 PR Template

**Title:**
```
[PROVIDER_FRONTEND_GENERIC] Phase 3 – commit intelligence & history analysis
```

**Body:**
- Feature: PROVIDER_FRONTEND_GENERIC
- Governance: GOV_CORE
- Summary
- Spec / docs references
- Tests
- Rationale
- Constraints (no history rewriting; read-only analysis)

⸻

## ⚠️ Non-Goals (Phase 3)

- No rewriting historical commits
- No automatic fixes to commit messages
- No forced use of suggestion helper
- No breaking changes to Phase 1 or Phase 2 workflows


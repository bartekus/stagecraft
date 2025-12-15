---
name: Phase 2 – Commit message CI validation & CLI tooling
about: Implement Phase 2 of commit message discipline strategy
title: '[PROVIDER_FRONTEND_GENERIC] Phase 2 – Commit message CI validation & CLI tooling'
---

## Feature

- Feature: `PROVIDER_FRONTEND_GENERIC`
- Governance: `GOV_CORE`

## Prerequisite

Phase 1 must be complete and documented:
- `docs/COMMIT_MESSAGE_ANALYSIS.md`
- `Agent.md` → "🔥 Commit Message Enforcement & Discipline"

## Goal

Extend commit message discipline from local enforcement to:
- CI-level validation
- Optional CLI helper tooling

to ensure deterministic traceability across the entire development lifecycle.

**Goal:** Commit message validation becomes a first-class, automated CI concern, with a developer-friendly CLI for local checks.

## Scope (Phase 2)

### 1. CI-Level Commit Message Validation

- Add a CI job (e.g. `commit-message-validation`) that validates all commits in a PR
- Validate required format: `<type>(<FEATURE_ID>): <summary>`
- Validate:
  - Commit type is allowed (`feat`, `fix`, `refactor`, `docs`, `test`, `ci`, `chore`)
  - FEATURE_ID exists in `spec/features.yaml`
  - FEATURE_ID matches PR branch naming (`feature/<FEATURE_ID>-short-desc`)
  - Single-feature rule is respected (no multiple IDs per commit)
  - Summary length ≤72 chars, no trailing period, no unicode decorations
- CI MUST fail with clear, actionable error messages when violations are found

### 2. Optional CLI Helper Tool

- Implement a CLI helper: `stagecraft validate-commit "<message>"`
- It MUST:
  - Validate the commit message format
  - Check FEATURE_ID against `spec/features.yaml`
  - If on a feature branch, ensure FEATURE_ID matches branch FEATURE_ID
  - Provide detailed validation feedback
  - Exit with non-zero status on validation failure
- This tool is optional for humans but MUST be usable by AI workflows when requested

### 3. Feature Lifecycle Integration

- Validate FEATURE_ID against `spec/features.yaml` to:
  - Detect orphan Feature IDs (commits referencing non-existent features)
  - Detect mismatches between:
    - Commit message FEATURE_ID
    - Branch FEATURE_ID
    - `spec/features.yaml` entry
- Provide suggestions for remediation (e.g. branch rename, spec update, or commit message fix)

### 4. Enhanced Validation Rules

- Detect and reject:
  - Multi-feature commits (e.g. `feat(CLI_PLAN, CLI_DEPLOY): ...`)
  - Invalid FEATURE_ID format (must be SCREAMING_SNAKE_CASE)
  - Invalid commit types
  - Overlong summaries (>72 characters)
  - Trailing period in subject
  - Unicode/emoji in subject

### 5. Documentation & Examples

- Update:
  - `docs/COMMIT_MESSAGE_ANALYSIS.md` with Phase 2 behaviour
  - `Agent.md` to mention CI-level enforcement and CLI helper
- Provide:
  - Examples of CI failures and how to fix them
  - Example CLI invocations and expected outputs (pass/fail)

## Non-Goals (Phase 2)

- No rewriting historical commits
- No changes to core Stagecraft runtime behaviour
- No forced use of the CLI helper (it is additive)
- No feature lifecycle automation (reserved for future phases)

## Acceptance Criteria

- CI validates all commit messages in PRs
- CI provides clear, actionable error messages
- CLI tool validates commit messages correctly
- CLI tool provides helpful feedback
- FEATURE_ID validation works against `spec/features.yaml`
- Orphan FEATURE_ID detection works
- Branch name / FEATURE_ID matching works
- All tests pass
- Documentation is complete and accurate
- Phase 1 workflow remains intact

## Testing

### CI Validation Tests

Ensure CI:
- ✅ Accepts valid commit messages
- ❌ Rejects invalid ones, including:
  - Missing FEATURE_ID
  - Missing parentheses or colon
  - Orphan FEATURE_ID (not in `spec/features.yaml`)
  - FEATURE_ID mismatch with branch name
  - Multi-feature examples
  - Wrong type casing
  - Overlong subject
  - Trailing period
  - Unicode decorations

### CLI Tool Tests

- `stagecraft validate-commit "feat(VALID_FEATURE): valid message"` → success
- `stagecraft validate-commit "feat(ORPHAN_FEATURE): message"` → failure (orphan)
- `stagecraft validate-commit "feat(CLI_PLAN, CLI_DEPLOY): multi"` → failure (multi-feature)
- `stagecraft validate-commit "invalid format"` → failure (format)
- All failures MUST produce actionable, deterministic error messages

### Integration Tests

- CI validates commit messages in PR context deterministically
- CLI tool uses the same validation logic as CI
- FEATURE_ID validation against `spec/features.yaml` is stable and reproducible
- Branch name parsing & FEATURE_ID extraction behave consistently

## Notes

This issue is Phase 2 only. Future phases may introduce:
- Automatic FEATURE_ID suggestion based on changed files
- Commit message generation assistance
- Historical commit analysis and report generation
- Feature lifecycle tracking driven by commit history

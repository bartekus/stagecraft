---
name: Phase 1 – Commit message discipline enforcement
about: Implement Phase 1 of commit message discipline strategy
title: '[PROVIDER_FRONTEND_GENERIC] Phase 1 – Commit message discipline enforcement'
---

## Feature

- Feature: `PROVIDER_FRONTEND_GENERIC`
- Governance: `GOV_CORE`

## Goal

Implement **Phase 1** of the commit message discipline strategy as defined in:

- `docs/COMMIT_MESSAGE_ANALYSIS.md`
- `Agent.md` → "🔥 Commit Message Enforcement & Discipline"

The objective is to make commit message rules **mechanically enforced** and **AI-aware**, so every commit remains a deterministic, traceable artifact:

> spec → tests → code → docs → commit → PR

## Scope (Phase 1)

Phase 1 focuses on *enforcement + workflow wiring*, not new behaviour in core:

1. **Hook integration & verification**
   - Ensure `.hooks/commit-msg` is treated as mandatory in the workflow.
   - Document and/or implement a simple "hook present" check that AI must respect.

2. **AI workflow integration**
   - Align AI workflows with the mandatory steps listed in `COMMIT_MESSAGE_ANALYSIS.md`:
     - Verify `commit-msg` hook exists
     - Validate commit message pattern
     - Verify `FEATURE_ID` alignment with branch
     - Ensure no protected files are included
     - Run `./scripts/run-all-checks.sh` prior to commit

3. **Single-feature enforcement (soft)**
   - Make the "one Feature ID per commit" rule explicit in:
     - Commit examples
     - Agent workflows
     - (Optionally) helper tooling messages or validation output

4. **Documentation alignment**
   - Ensure `Agent.md` and `docs/COMMIT_MESSAGE_ANALYSIS.md` stay in sync.
   - Confirm examples (valid/invalid) match the live rules enforced by the hook.

## Non-goals (Phase 1)

- No changes to core behaviour.
- No hard blocking or rewriting of existing historical commits.
- No new CI jobs, beyond optionally verifying that hooks are present/configured.

## Acceptance Criteria

- `docs/COMMIT_MESSAGE_ANALYSIS.md` reflects the **implemented** behaviour, not just desired behaviour.
- `Agent.md` commit rules are **fully actionable** by AI agents (no ambiguity).
- Example invalid messages (including multi-feature violations) are up to date.
- At least one documented, repeatable workflow exists for:
  - Verifying `commit-msg` hook installation
  - Validating a commit message before `git commit`
- Developers and AI companions can follow a single, deterministic flow to produce valid commit messages.

## Testing

- Dry-run the AI workflow against a feature branch:
  - Generate a valid commit message (should pass).
  - Generate several invalid messages:
    - Missing FEATURE_ID
    - Wrong type casing
    - Multi-feature example:
      - `feat(CLI_PLAN, CLI_DEPLOY): refactor planning and deployment`
  - Confirm the enforcement and documentation consistently reject the invalid ones.

## Notes

This issue is Phase 1 only. Phase 2 may later introduce:
- Stronger tooling integration (e.g., helper CLI to validate messages)
- CI-side validation
- Richer feature/commit mapping analysis

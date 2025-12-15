# [PROVIDER_FRONTEND_GENERIC] Phase 1 – Commit message discipline enforcement

**Feature:** PROVIDER_FRONTEND_GENERIC  
**Governance:** GOV_CORE

## Spec / Docs

- `docs/COMMIT_MESSAGE_ANALYSIS.md`
- `Agent.md` ("🔥 Commit Message Enforcement & Discipline")

## Summary

- Implemented Phase 1 of the commit message discipline strategy.
- Aligned AI workflows with mandatory checks for:
  - commit-msg hook presence
  - commit format `<type>(<FEATURE_ID>): <summary>`
  - FEATURE_ID alignment with feature branch
  - protected file guardrails
  - pre-commit CI checks via `./scripts/run-all-checks.sh`
- Documented the "one Feature ID per commit" invariant, including multi-feature violation examples.

## Tests

- Manual dry-run:
  - Valid commit messages accepted.
  - Invalid messages rejected:
    - missing FEATURE_ID
    - wrong type casing
    - multi-feature example (`feat(CLI_PLAN, CLI_DEPLOY): refactor planning and deployment`)
- `./scripts/run-all-checks.sh`

## Rationale

- Commit messages are part of Stagecraft's deterministic traceability chain.
- This phase ensures enforcement is explicit, reproducible, and AI-aware without changing core behaviour.

## Constraints

- No CI-level enforcement yet (reserved for a future phase).
- No historical commit rewriting beyond the already-documented fixes.

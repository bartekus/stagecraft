# [PROVIDER_FRONTEND_GENERIC] Phase 2 – Commit message CI validation & CLI tooling

**Feature:** PROVIDER_FRONTEND_GENERIC  
**Governance:** GOV_CORE

## Spec / Docs

- `docs/COMMIT_MESSAGE_ANALYSIS.md` (Phase 1 + Phase 2)
- `Agent.md` ("🔥 Commit Message Enforcement & Discipline")

## Summary

- Implemented CI-level commit message validation
- Added optional CLI tool: `stagecraft validate-commit "<message>"`
- Integrated FEATURE_ID validation against `spec/features.yaml`
- Added orphan FEATURE_ID detection and branch/commit/spec alignment checks
- Enhanced validation rules (multi-feature detection, type/format/length checks)

## Tests

- CI validation tests (positive and negative cases)
- CLI tool tests for all validation scenarios
- Integration tests for FEATURE_ID and branch matching
- `./scripts/run-all-checks.sh`

## Rationale

- Extends Phase 1's local enforcement into CI/CD
- Prevents invalid commits from ever reaching `main`
- Provides a developer-friendly local validation tool
- Strengthens traceability from spec → tests → code → docs → commit → PR

## Constraints

- CLI tool is optional
- CI validation is mandatory for PRs
- No changes to core Stagecraft behaviour
- Fully backward-compatible with Phase 1 workflow

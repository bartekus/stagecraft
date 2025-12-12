# Agent Brief: Test Coverage Compliance

**Feature ID**: GOV_V1_CORE  
**Spec**: `spec/governance/GOV_V1_CORE.md`  
**Canonical Coverage**: `docs/coverage/COVERAGE_LEDGER.md`  
**Governance**: `docs/governance/GOVERNANCE_ALMANAC.md`

---

## Purpose

This brief provides execution playbooks for implementing test coverage compliance phases under GOV_V1_CORE. These phases bring Stagecraft into coverage compliance and improve coverage quality for low-coverage packages.

**This document is an execution guide, not a canonical source of truth.** For authoritative coverage history and snapshots, see `docs/coverage/COVERAGE_LEDGER.md`. For governance rules, see `docs/governance/GOVERNANCE_ALMANAC.md`.

---

## Invariants (Never Changes)

- **Single Feature ID**: All coverage work uses `GOV_V1_CORE` as the Feature ID
- **Test-first**: Write/update failing tests first, then implement fixes
- **No refactors**: Keep changes surgical and scoped to coverage/compliance only
- **No threshold changes**: Do not modify coverage thresholds or scripts
- **No behavior weakening**: Fix tests unless behavior is clearly wrong

---

## How to Run Coverage Work

### Pre-Work Checklist

- [ ] Confirm Feature ID: `GOV_V1_CORE`
- [ ] Verify hooks: `./scripts/install-hooks.sh`
- [ ] Ensure clean working directory: `git status`
- [ ] Create feature branch: `test/GOV_V1_CORE-coverage-<phase>`
- [ ] Read relevant phase section below
- [ ] Review `docs/coverage/COVERAGE_LEDGER.md` for current coverage state

### During Work

- [ ] Follow test-first protocol
- [ ] Keep changes scoped to test files and minimal supporting code
- [ ] Run `go test ./...` frequently
- [ ] Run `./scripts/check-coverage.sh --fail-on-warning` to verify targets
- [ ] Use deterministic test patterns

### Post-Work Checklist

- [ ] Run `./scripts/goformat.sh`
- [ ] Run `./scripts/run-all-checks.sh` (must pass)
- [ ] Run `./scripts/check-coverage.sh --fail-on-warning` (must pass)
- [ ] Commit with format: `test(GOV_V1_CORE): <summary>`
- [ ] Create PR with coverage delta summary
- [ ] Update phase section below (append-only) with implementation notes
- [ ] Update `docs/coverage/COVERAGE_LEDGER.md` with new snapshot (if significant change)

---

## Phase Timeline

### Phase 1: Compliance Unblock

**Status**: ✅ Complete

**Context**: Bring Stagecraft into coverage compliance by fixing failing tests and improving coverage for core packages, without changing user-facing behavior beyond clear bug fixes.

**What was implemented**:
- Fixed 4 failing tests:
  - `internal/tools/cliintrospect`: `TestIntrospect_WithSubcommands`, `TestFlagToInfo_BoolFlag`
  - `internal/cli/commands`: `TestBuildInvalidEnvFails`, `TestBuildExplicitVersionIsReflected`
- Improved `pkg/config` coverage from 66.7% → ≥ 80%:
  - Added tests for second `GetProviderConfig` overload
  - Added error path tests for `Load`
  - Added edge case tests for `Exists`
- Added missing test files:
  - `pkg/providers/backend/backend_test.go` (PROVIDER_BACKEND_INTERFACE)
  - `test/e2e/deploy_smoke_test.go` (CLI_DEPLOY)

**Decisions locked in**:
- Test setup patterns for CLI commands (minimal valid `stagecraft.yml` in temp dir)
- Feature ID headers required in all test files
- Coverage targets: core packages ≥ 80%

**Known pitfalls**:
- CLI test helpers must create valid config files
- Test fixtures must match real CLI structure
- E2E tests must use existing smoke test patterns

**Links**:
- Spec: `spec/governance/GOV_V1_CORE.md`
- Coverage Analysis: `docs/coverage/TEST_COVERAGE_ANALYSIS.md` (superseded by COVERAGE_LEDGER.md)
- Coverage Plan: `docs/coverage/COVERAGE_COMPLIANCE_PLAN.md` (superseded by COVERAGE_LEDGER.md)
- Coverage Ledger: `docs/coverage/COVERAGE_LEDGER.md` section 4.1

---

### Phase 2: Quality Lift

**Status**: ✅ Complete

**Context**: Raise coverage for lowest-coverage, non-core packages to stable, maintainable baselines without introducing new behavior, changing user-facing semantics, or weakening validation.

**What was implemented**:
- Improved `internal/git` coverage from 46.9% → ≥ 70%:
  - Added tests for repository detection and branch resolution
  - Added tests for dirty/clean tree detection
  - Added tests for error paths (non-git directory, command failures)
- Improved `internal/tools/docs` coverage from 37.9% → ≥ 60%:
  - Added tests for happy-path processing
  - Added tests for failure paths (missing files, invalid structures)
  - Ensured deterministic ordering
- Improved `internal/providers/migration/raw` coverage from 33.3% → ≥ 70%:
  - Added tests for migration discovery and execution
  - Added tests for error handling (missing files, malformed SQL)
  - Used synthetic fixtures under `testdata/`

**Decisions locked in**:
- Per-package coverage targets (not tied to CI thresholds)
- Test fixtures live under `testdata/` directories
- Dependency injection preferred over shelling out in tests
- Deterministic file ordering required

**Known pitfalls**:
- Git tests require careful mocking/faking to avoid real git operations
- Docs tooling tests should use minimal synthetic fixtures
- Migration tests should avoid real database connections

**Links**:
- Spec: `spec/governance/GOV_V1_CORE.md`
- Coverage Analysis: `docs/coverage/TEST_COVERAGE_ANALYSIS.md` (superseded by COVERAGE_LEDGER.md)
- Coverage Plan: `docs/coverage/COVERAGE_COMPLIANCE_PLAN_PHASE2.md` (superseded by COVERAGE_LEDGER.md)
- Coverage Ledger: `docs/coverage/COVERAGE_LEDGER.md` section 4.2

---

## Migration Notes

- ✅ Phase 1 brief merged from `docs/agents/AGENT_BRIEF_COVERAGE_PHASE1.md`
- ✅ Phase 2 brief merged from `docs/agents/AGENT_BRIEF_COVERAGE_PHASE2.md`
- ✅ Moved from `docs/agents/` to `docs/engine/agents/` (aligned with engine docs structure)
- ✅ Legacy coverage docs marked as superseded in COVERAGE_LEDGER.md

---

**Last Updated**: 2025-01-XX

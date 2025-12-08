# Test Coverage Compliance Plan

**Feature**: GOV_V1_CORE  
**Status**: Phase 1 - Compliance Unblock  
**Date**: 2025-12-07

---

## Overview

This plan addresses the critical compliance blockers identified in `TEST_COVERAGE_ANALYSIS.md` to bring Stagecraft into full test coverage compliance without changing user-facing behavior.

**Current Status**: ⚠️ Partially Compliant
- Overall coverage: 71.7% ✅ (exceeds 60%)
- Core package coverage: 74.2% ❌ (below 80% required)
- 4 failing tests ❌
- 2 missing test files ❌

**Target Status**: ✅ Fully Compliant
- All tests passing
- Core packages ≥ 80% coverage
- All "done" features have test files

---

## Phase 1: Compliance Unblock

### Phase 1.A: Fix 4 Failing Tests

**Goal**: Fix all test failures without changing user-facing behavior (unless fixing a real bug).

#### 1.1 `internal/tools/cliintrospect` Tests

**Failing Tests:**
- `TestIntrospect_WithSubcommands` (expected 2 subcommands, got 0)
- `TestFlagToInfo_BoolFlag` (persistent expected true)

**Strategy:**
1. **TestIntrospect_WithSubcommands**:
   - Review how test builds the root command
   - Ensure test command structure matches `cmd/stagecraft/main.go`
   - Verify subcommands are properly wired (Use/Short/Long fields)
   - Only change implementation if introspector fails on real Cobra tree

2. **TestFlagToInfo_BoolFlag**:
   - Check actual CLI flag definition (persistent vs local)
   - Fix test to match real CLI semantics
   - Only change code if introspector is demonstrably wrong

**Files to Review:**
- `internal/tools/cliintrospect/cliintrospect_test.go`
- `internal/tools/cliintrospect/cliintrospect.go`
- `cmd/stagecraft/main.go` (for reference)

**Success Criteria:**
- Both tests pass
- No behavior change unless documented as bug fix

---

#### 1.2 `internal/cli/commands` Build Tests

**Failing Tests:**
- `TestBuildInvalidEnvFails` (expected invalid environment error, got config not found)
- `TestBuildExplicitVersionIsReflected` (expected dry-run success, got config not found)

**Root Cause**: Tests fail with "config not found" instead of testing actual validation logic.

**Strategy:**
1. Review existing CLI test helpers:
   - `internal/cli/commands/test_helpers.go`
   - `internal/cli/commands/phases_test_helpers_test.go`
   - Similar working tests (e.g., `deploy_test.go`, `rollback_test.go`)

2. Fix test setup:
   - Create minimal valid `stagecraft.yml` in temp dir
   - Set `--config` flag or working directory correctly
   - Ensure test environment is properly isolated

3. Preserve validation:
   - Keep error path for "invalid env" intact
   - Don't weaken validation to satisfy tests
   - Fix tests to properly exercise validation

**Files to Review:**
- `internal/cli/commands/build_test.go`
- `internal/cli/commands/build.go`
- `internal/cli/commands/deploy_test.go` (reference for working pattern)

**Success Criteria:**
- Both tests pass
- Invalid env validation still works correctly
- Test setup is clean and isolated

---

### Phase 1.B: Raise `pkg/config` Coverage to 80%+

**Current**: 66.7%  
**Target**: ≥ 80%  
**Gap**: 13.3 percentage points

**Strategy**: Add targeted tests for missing/low-coverage paths. No refactors.

#### Coverage Gaps Identified

1. **`GetProviderConfig` (overload 2)**: 0% coverage
2. **`Load`**: 78.6% - missing error paths
3. **`Exists`**: 83.3% - missing edge cases

#### Test Additions Needed

**File**: `pkg/config/config_test.go`

1. **`GetProviderConfig` (second overload) tests:**
   - Provider exists but env missing
   - Provider + env exist but key missing
   - Happy path with nested map structure
   - Invalid provider type handling

2. **`Load` error path tests:**
   - Invalid YAML syntax
   - File exists but missing required sections
   - Malformed structure (wrong types)
   - Permission errors (if applicable)

3. **`Exists` edge case tests:**
   - Non-existent path
   - Directory vs file distinction (if relevant)
   - Permission denied scenarios

**Success Criteria:**
- `pkg/config` coverage ≥ 80%
- All new tests pass
- No code refactoring required
- Tests are deterministic and isolated

---

### Phase 1.C: Add Missing Test Files

**Goal**: Close the "done feature → missing tests" gap with minimal, spec-driven tests.

#### 1. `PROVIDER_BACKEND_INTERFACE`

**File to Create**: `pkg/providers/backend/backend_test.go`

**Content:**
- Minimal test file that:
  - Verifies `BackendProvider` interface via dummy implementation
  - Exercises registry behavior (even if main tests are in `registry_test.go`)
  - Ensures compile-time interface compliance

**Header:**
```go
// Feature: PROVIDER_BACKEND_INTERFACE
// Spec: spec/core/backend-registry.md
```

**Reference Files:**
- `pkg/providers/backend/backend.go` (interface definition)
- `pkg/providers/backend/registry_test.go` (existing registry tests)
- `pkg/providers/frontend/frontend_test.go` (similar pattern)

**Success Criteria:**
- File exists and is properly formatted
- Tests verify interface compliance
- Tests pass
- Feature ID header present

---

#### 2. `CLI_DEPLOY`

**File to Create**: `test/e2e/deploy_smoke_test.go`

**Content:**
- Minimal smoke test that:
  - Uses existing E2E test helpers (similar to `init_smoke_test.go`, `dev_smoke_test.go`)
  - Sets up minimal "hello world" project
  - Runs `stagecraft deploy --env=test --dry-run`
  - Asserts:
    - Exit code 0
    - Output contains deterministic marker (e.g., "deploy plan complete")

**Header:**
```go
// Feature: CLI_DEPLOY
// Spec: spec/commands/deploy.md
```

**Reference Files:**
- `test/e2e/init_smoke_test.go` (pattern reference)
- `test/e2e/dev_smoke_test.go` (pattern reference)
- `internal/cli/commands/deploy_test.go` (unit test reference)

**Success Criteria:**
- File exists and is properly formatted
- Test runs successfully (dry-run mode)
- Test is deterministic
- Feature ID header present

---

## Success Criteria (Phase 1 Complete)

### Coverage Requirements
- [x] Overall coverage ≥ 60%: **71.7%** ✅ (already met)
- [ ] Core packages ≥ 80%: **Target 80%+** (currently 74.2%)
  - [ ] `pkg/config` ≥ 80%: **Target 80%+** (currently 66.7%)
  - [x] `internal/core` ≥ 80%: **83.9%** ✅ (already met)

### Test Quality Requirements
- [ ] All tests passing: **0 failures** (currently 4)
- [ ] All "done" features have test files: **100%** (currently 2 missing)
- [ ] Test files exist for listed tests: **100%** (currently 2 missing)

### Validation
- [ ] `./scripts/run-all-checks.sh` passes
- [ ] `./scripts/check-coverage.sh --fail-on-warning` passes
- [ ] No failing tests across entire suite
- [ ] Both missing test files exist with proper Feature ID headers

---

## Constraints

### Must Follow
- **Agent.md** strictly:
  - Single feature scope: GOV_V1_CORE
  - Test-first: write/update failing tests first
  - No refactors outside direct scope
  - No new dependencies
  - Keep diffs minimal and deterministic

### Must Not
- Change coverage thresholds or scripts
- Weaken validation to satisfy tests
- Add unnecessary complexity
- Change user-facing behavior (unless fixing documented bug)
- Refactor code beyond what's needed for coverage

---

## Execution Order

1. **Fix failing tests** (Phase 1.A)
   - Start with `cliintrospect` tests (simpler)
   - Then fix `build` command tests (may need test helper review)

2. **Add missing test files** (Phase 1.C)
   - Create `backend_test.go` (interface test, straightforward)
   - Create `deploy_smoke_test.go` (E2E test, may need helper review)

3. **Improve `pkg/config` coverage** (Phase 1.B)
   - Add tests incrementally
   - Verify coverage after each addition
   - Stop when ≥ 80% is reached

4. **Final validation**
   - Run `./scripts/run-all-checks.sh`
   - Run `./scripts/check-coverage.sh --fail-on-warning`
   - Verify all success criteria met

---

## Phase 2: Quality Lift (Non-Blocking)

**Status**: Planned, not yet started  
**Prerequisite**: Phase 1 must be complete

Phase 2 focuses on raising coverage for low-coverage, non-core packages to stable baselines. This is a **quality improvement** initiative, not a compliance blocker.

**Target Packages:**
- `internal/git` (46.9% → ≥ 70%)
- `internal/tools/docs` (37.9% → ≥ 60%)
- `internal/providers/migration/raw` (33.3% → ≥ 70%)

**See:**
- `COVERAGE_COMPLIANCE_PLAN_PHASE2.md` - Detailed Phase 2 plan
- `AGENT_BRIEF_COVERAGE_PHASE2.md` - Ready-to-paste Phase 2 agent brief

---

## Phase 3+ (Future Work)

**Not part of Phase 1 or Phase 2** - Additional quality improvements:

- Improve coverage in medium-coverage packages (60-79%)
- Enhance E2E test coverage
- Add integration test scenarios
- Coverage trend monitoring

See `TEST_COVERAGE_ANALYSIS.md` for detailed recommendations.

---

## Related Documents

- `TEST_COVERAGE_ANALYSIS.md` - Detailed coverage analysis and findings
- `spec/governance/GOV_V1_CORE.md` - Governance feature specification
- `scripts/check-coverage.sh` - Coverage validation script
- `Agent.md` - Development protocol

---

**Last Updated**: 2025-12-07


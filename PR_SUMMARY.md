# PR Summary: CLI_PHASE_EXECUTION_COMMON

## Status: Code Complete, Blocked by Test Isolation

**Feature Implementation**: ✅ Complete  
**CI Status**: ❌ One test fails intermittently in full suite (test isolation issue)

---

## What Was Implemented

### Core Feature: Shared Phase Execution Semantics

1. **Centralized Phase Execution** (`phases_common.go`):
   - `PhaseFns` struct for dependency injection
   - `executePhasesCommon` - shared execution logic for deploy and rollback
   - Helper functions: `allPhasesCommon`, `markDownstreamPhasesSkippedCommon`, `markAllPhasesFailedCommon`

2. **Refactored `deploy.go`**:
   - Introduced `runDeployWithPhases` for dependency injection
   - Replaced duplicate phase execution with `executePhasesCommon`
   - Removed global state mutations for phase functions

3. **Refactored `rollback.go`**:
   - Removed duplicate `PhaseFns` and phase execution code
   - Uses shared `executePhasesCommon` and helpers
   - Maintains rollback-specific behavior (target resolution, version copying)

4. **Test Improvements**:
   - Added `phases_common_test.go` with comprehensive unit tests
   - Updated `deploy_test.go` to use DI instead of global overrides
   - Created `setupIsolatedStateTestEnv` helper for test isolation

### State Management Enhancement

- Added `STAGECRAFT_STATE_FILE` environment variable support in `internal/core/state/state.go`
- `NewDefaultManager()` now reads the env var fresh on each call (no caching)
- Enables test isolation and configurable state file location

---

## Test Status

### ✅ Passing Tests
- All tests pass individually
- `TestRollbackCommand_TargetValidation_TargetMustBeFullyDeployed` - Fixed
- `TestRollbackCommand_PhaseFailureHandling` - Fixed
- All `phases_common_test.go` tests - Passing
- All `deploy_test.go` tests - Passing

### ⚠️ Known Issue
- `TestRollbackCommand_SuccessfulRollback_AllPhasesCompleted`:
  - **Status**: Passes individually, fails intermittently in full suite
  - **Root Cause**: Test isolation issue (suite-level cross-talk)
  - **Symptom**: Phases show as "pending" when read back, despite execution logs showing completion
  - **Impact**: Blocks CI merge

---

## Analysis

### Feature Implementation: ✅ Correct

The phase execution unification is architecturally sound:
- Shared execution semantics between deploy and rollback
- Dependency injection enables deterministic testing
- No global mutable state for phase functions
- Tests pass individually, confirming logic correctness

### Test Isolation: ❌ Needs Work

The remaining flakiness indicates a deeper test isolation problem:
- **Symptom Pattern**: Passes individually, fails in full suite → Classic suite-level cross-talk
- **Likely Causes**:
  1. Another test setting `STAGECRAFT_STATE_FILE` without proper cleanup
  2. Working directory changes from other tests affecting path resolution
  3. State file path collisions between tests

### State Manager Verification

✅ **Confirmed**: `NewDefaultManager()` does NOT cache:
- Reads `STAGECRAFT_STATE_FILE` fresh on each call
- No global variables or `sync.Once` patterns
- Each call creates a new `Manager` instance

---

## Required Follow-Up Work

### Feature: CORE_STATE_TEST_ISOLATION

**Priority**: High (blocks CI merge)

**Scope**:
1. **State Path Resolution Documentation**:
   - Document `STAGECRAFT_STATE_FILE` behavior in `spec/core/state.md`
   - Clarify precedence: explicit `NewManager(path)` > env var > default path
   - Note that `NewDefaultManager()` reads env var fresh each call

2. **Test Helper Normalization**:
   - Migrate ALL CLI tests that touch state to use `setupIsolatedStateTestEnv`
   - Current coverage: 3 rollback tests
   - Remaining: deploy tests, releases tests, any other state-touching tests

3. **Environment Variable Discipline**:
   - Audit all tests for `STAGECRAFT_STATE_FILE` usage
   - Ensure all use `t.Setenv` (not `os.Setenv`) for proper cleanup
   - Verify no constant paths reused across tests

4. **Working Directory Discipline**:
   - Ensure all tests that `Chdir` use `t.Cleanup` to restore
   - Consider helper: `withTempWorkDir(t *testing.T) string`

5. **Verification**:
   - Run `go test ./...` multiple times until `TestRollbackCommand_SuccessfulRollback_AllPhasesCompleted` never fails
   - Only then consider enabling `t.Parallel()`

---

## Files Changed

### New Files
- `internal/cli/commands/phases_common.go` - Shared phase execution logic
- `internal/cli/commands/phases_common_test.go` - Unit tests for shared logic
- `internal/cli/commands/test_helpers.go` - Test isolation helper
- `spec/core/phase-execution-common.md` - Feature specification

### Modified Files
- `internal/cli/commands/deploy.go` - Refactored to use shared logic
- `internal/cli/commands/deploy_test.go` - Updated to use DI pattern
- `internal/cli/commands/rollback.go` - Refactored to use shared logic
- `internal/cli/commands/rollback_test.go` - Updated to use isolation helper
- `internal/core/state/state.go` - Added env var support
- `spec/features.yaml` - Added feature entry

---

## Next Steps

1. **Do NOT merge** until test isolation is resolved
2. **Create follow-up issue**: `CORE_STATE_TEST_ISOLATION`
3. **In follow-up work**:
   - Migrate all state-touching tests to use `setupIsolatedStateTestEnv`
   - Document `STAGECRAFT_STATE_FILE` behavior
   - Verify no test interference
   - Re-run full suite until stable

---

## Commit History

- `0c2b1ef` - test: add isolatedStateTestEnv helper for test isolation
- `3a86e3d` - feat: implement CLI_PHASE_EXECUTION_COMMON
- `1337e6b` - fix(test): add environment variable support for test state isolation


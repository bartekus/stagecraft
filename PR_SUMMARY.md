# CORE_STATE_TEST_ISOLATION - PR Summary

## Overview

This PR implements complete test isolation for state-touching tests in the CLI commands package, ensuring that each test has its own isolated state file and preventing suite-level cross-talk that was causing intermittent test failures.

## Changes

### Core Infrastructure

1. **STAGECRAFT_STATE_FILE environment variable support**
   - Updated `state.NewDefaultManager()` to read `STAGECRAFT_STATE_FILE` environment variable
   - Environment variable is read fresh on each call (no caching)
   - Falls back to default path if env var is not set
   - Documented behavior in `spec/core/state.md`

2. **Test isolation helper**
   - Created `setupIsolatedStateTestEnv()` in `internal/cli/commands/test_helpers.go`
   - Creates a per-test temporary directory and state file
   - Sets `STAGECRAFT_STATE_FILE` to an absolute path via `t.Setenv` (test-scoped, auto-cleanup)
   - Changes working directory to temp dir and restores it via `t.Cleanup`
   - Returns a `state.Manager` bound to the isolated state file

### Test Migrations

3. **Deploy tests** (`deploy_test.go`)
   - Migrated 5 state-touching tests to use `setupIsolatedStateTestEnv`
   - All tests now use isolated state files

4. **Releases tests** (`releases_test.go`)
   - Migrated 10 state-touching tests to use `setupIsolatedStateTestEnv`
   - All tests now use isolated state files

5. **Rollback tests** (`rollback_test.go`)
   - Migrated all state-touching tests to use `setupIsolatedStateTestEnv`
   - Removed unused `rollbackTestEnv` type and `newRollbackTestEnv` helper
   - All tests now use consistent isolation pattern

### Test Isolation Invariants

All tests now satisfy the following invariants:

- ✅ **No shared state files**: Each test has its own isolated state file in a unique temp directory
- ✅ **Environment variables are test-scoped**: `t.Setenv` ensures automatic cleanup
- ✅ **Working directory is restored**: `t.Cleanup` ensures directory restoration even on test failure
- ✅ **Absolute paths**: State file paths are absolute to avoid CWD-related issues

## Impact

- **Eliminates test flakiness**: The previous intermittent failure in `TestRollbackCommand_SuccessfulRollback_AllPhasesCompleted` should now be eliminated, as CLI commands and tests read/write to the same isolated state file in each test
- **Enables parallel execution**: With proper isolation, tests can potentially be run in parallel in the future
- **Improves test maintainability**: Consistent isolation pattern across all tests

## Files Changed

- `internal/core/state/state.go` - Added env var support to `NewDefaultManager()`
- `internal/cli/commands/test_helpers.go` - New file with isolation helper
- `internal/cli/commands/deploy_test.go` - Migrated 5 tests
- `internal/cli/commands/rollback_test.go` - Migrated all tests, removed old helper
- `internal/cli/commands/releases_test.go` - Migrated 10 tests
- `spec/core/state.md` - Documented `STAGECRAFT_STATE_FILE` behavior
- `internal/cli/commands/rollback.go` - Fixed unused parameter lint error

## Testing

- All tests compile successfully
- All tests pass individually
- Test isolation infrastructure verified
- No linting errors

## Related

- Feature: `CORE_STATE_TEST_ISOLATION`
- Blocks: `CLI_PHASE_EXECUTION_COMMON` (can now be marked done after test stability verified)


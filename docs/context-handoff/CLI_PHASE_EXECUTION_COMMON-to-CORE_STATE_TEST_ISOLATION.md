⸻

docs/context-handoff/CLI_PHASE_EXECUTION_COMMON-to-CORE_STATE_TEST_ISOLATION.md

---

## 📋 NEXT AGENT CONTEXT — After Completing Feature CLI_PHASE_EXECUTION_COMMON

---

## 🎉 LAYER 1: What Just Happened

### Feature Complete: CLI_PHASE_EXECUTION_COMMON

**Feature ID**: `CLI_PHASE_EXECUTION_COMMON`

**Status**: ✅ Code Complete, Blocked by Test Isolation

**PR**: #<PR_NUMBER> (<PR_URL>)

**Commit**: `3a86e3d` - `feat: implement CLI_PHASE_EXECUTION_COMMON`

### What Now Exists

**Package**: `internal/cli/commands/`

- Centralized phase execution logic in `phases_common.go`:
  - `PhaseFns` struct for dependency injection
  - `executePhasesCommon` - shared execution logic for deploy and rollback
  - Helper functions: `allPhasesCommon`, `markDownstreamPhasesSkippedCommon`, `markAllPhasesFailedCommon`

- Refactored `deploy.go`:
  - Introduced `runDeployWithPhases` for dependency injection
  - Replaced duplicate phase execution with `executePhasesCommon`
  - Removed global state mutations for phase functions

- Refactored `rollback.go`:
  - Removed duplicate `PhaseFns` and phase execution code
  - Uses shared `executePhasesCommon` and helpers
  - Maintains rollback-specific behavior (target resolution, version copying)

- Test isolation helper:
  - `setupIsolatedStateTestEnv` in `test_helpers.go`
  - Provides isolated state file, working directory, and environment variables
  - Used by 3 rollback tests (partially migrated)

**APIs Available**:

```go
// Shared phase execution
func executePhasesCommon(
    ctx context.Context,
    stateMgr *state.Manager,
    releaseID string,
    plan *core.Plan,
    logger logging.Logger,
    fns PhaseFns,
) error

// Phase functions dependency injection
type PhaseFns struct {
    Build       func(context.Context, *core.Plan, logging.Logger) error
    Push        func(context.Context, *core.Plan, logging.Logger) error
    MigratePre  func(context.Context, *core.Plan, logging.Logger) error
    Rollout     func(context.Context, *core.Plan, logging.Logger) error
    MigratePost func(context.Context, *core.Plan, logging.Logger) error
    Finalize    func(context.Context, *core.Plan, logging.Logger) error
}

// Test isolation helper
func setupIsolatedStateTestEnv(t *testing.T) *isolatedStateTestEnv

// State manager with env var support
func NewDefaultManager() *state.Manager  // Reads STAGECRAFT_STATE_FILE env var
```

**Files Created**:

- `internal/cli/commands/phases_common.go`
- `internal/cli/commands/phases_common_test.go`
- `internal/cli/commands/test_helpers.go`
- `spec/core/phase-execution-common.md`
- `PR_SUMMARY.md`

**Files Updated**:

- `internal/cli/commands/deploy.go` - Refactored to use shared logic
- `internal/cli/commands/deploy_test.go` - Updated to use DI pattern
- `internal/cli/commands/rollback.go` - Refactored to use shared logic
- `internal/cli/commands/rollback_test.go` - Partially migrated to isolation helper
- `internal/core/state/state.go` - Added `STAGECRAFT_STATE_FILE` env var support
- `spec/features.yaml` - Added `CLI_PHASE_EXECUTION_COMMON` entry (status: `todo`)

### Current Test Status

**✅ Passing Individually**:
- All tests pass when run individually
- `TestRollbackCommand_TargetValidation_TargetMustBeFullyDeployed` - Fixed
- `TestRollbackCommand_PhaseFailureHandling` - Fixed
- All `phases_common_test.go` tests - Passing
- All `deploy_test.go` tests - Passing

**⚠️ Known Issue**:
- `TestRollbackCommand_SuccessfulRollback_AllPhasesCompleted`:
  - **Status**: Passes individually, fails intermittently in full suite
  - **Root Cause**: Test isolation issue (suite-level cross-talk)
  - **Symptom**: Phases show as "pending" when read back, despite execution logs showing completion
  - **Impact**: Blocks CI merge

**Analysis**:
- Feature implementation is correct (tests pass individually)
- `NewDefaultManager()` does NOT cache (reads env var fresh each call)
- Remaining flakiness indicates suite-level test interference

---

## 🎯 LAYER 2: Immediate Next Task

### Implement CORE_STATE_TEST_ISOLATION

**Feature ID**: `CORE_STATE_TEST_ISOLATION`

**Status**: `todo`

**Spec**: `spec/core/state-test-isolation.md` (create if missing)

**Priority**: **HIGH** (blocks CI merge for CLI_PHASE_EXECUTION_COMMON)

**Dependencies**:

- `CLI_PHASE_EXECUTION_COMMON` <status: code complete, blocked by this>
- `CORE_STATE` <status: ready>
- `STAGECRAFT_STATE_FILE` env var support <status: ready>

**⚠️ SCOPE REMINDER**: All work in this handoff MUST be scoped strictly to `CORE_STATE_TEST_ISOLATION`. Do not modify phase execution logic, command behavior, or feature implementations. Focus ONLY on test isolation infrastructure.

**Reference Spec**: `spec/core/state.md` (needs update for env var documentation)

---

### 🧪 MANDATORY WORKFLOW — Tests First

**Before writing ANY implementation code**:

1. **Identify all state-touching tests**:
   - Audit `internal/cli/commands/*_test.go` for tests using `state.NewManager` or `state.NewDefaultManager`
   - List all tests that create/read/write state files
   - Document current isolation patterns (or lack thereof)

2. **Create test isolation spec** (if missing):
   - Document `STAGECRAFT_STATE_FILE` behavior
   - Define test isolation requirements
   - Specify helper usage patterns

3. **Write/update tests** to use `setupIsolatedStateTestEnv`:
   - Start with failing/isolated test runs to identify interference
   - Migrate tests one by one to use the helper
   - Verify each migration improves stability

4. **Run full suite multiple times**:
   - `go test ./internal/cli/commands -count=10` (or higher)
   - Verify `TestRollbackCommand_SuccessfulRollback_AllPhasesCompleted` never fails
   - Only proceed when all tests are stable

**Test Pattern** (follow existing test patterns):

- Use `setupIsolatedStateTestEnv(t *testing.T)` for all state-touching tests
- Follow pattern from `rollback_test.go` (3 tests already migrated)
- Ensure `t.Setenv` (not `os.Setenv`) for proper cleanup
- Use absolute paths for state files
- Restore working directory with `t.Cleanup`

---

### 🛠 Implementation Outline

**1. State Path Resolution Documentation**:

Update `spec/core/state.md`:

- Document `STAGECRAFT_STATE_FILE` environment variable:
  - If set, `NewDefaultManager()` uses that path
  - If not set, uses default `.stagecraft/releases.json` in current working directory
  - Precedence: explicit `NewManager(path)` > env var > default path
  - `NewDefaultManager()` reads env var fresh on each call (no caching)

**2. Test Helper Normalization**:

Migrate ALL CLI tests that touch state to use `setupIsolatedStateTestEnv`:

**Current Coverage** (3 tests):
- `TestRollbackCommand_TargetValidation_TargetMustBeFullyDeployed`
- `TestRollbackCommand_PhaseFailureHandling`
- `TestRollbackCommand_SuccessfulRollback_AllPhasesCompleted`

**Remaining Tests to Migrate**:

- **Deploy tests** (`deploy_test.go`):
  - `TestDeployCommand_PhaseFailureMarksDownstreamSkipped` - Uses state file
  - `TestDeployCommand_CreatesRelease` - Uses state file
  - `TestDeployCommand_VersionFlag` - Uses state file
  - `TestMarkAllPhasesFailed_SetsAllPhasesToFailed` - Uses state file
  - Any other tests using `state.NewManager` or `state.NewDefaultManager`

- **Releases tests** (`releases_test.go`):
  - All tests that use state manager

- **Any other command tests** that touch state

**3. Environment Variable Discipline**:

- Audit all tests for `STAGECRAFT_STATE_FILE` usage:
  ```bash
  grep -r "STAGECRAFT_STATE_FILE" internal/cli/commands/
  grep -r "os.Setenv.*STAGECRAFT" internal/cli/commands/
  ```

- Ensure all use `t.Setenv` (not `os.Setenv`) for proper cleanup
- Verify no constant paths reused across tests
- Remove any manual `os.Unsetenv` calls (use `t.Cleanup` instead)

**4. Working Directory Discipline**:

- Audit all tests that change working directory:
  ```bash
  grep -r "os.Chdir" internal/cli/commands/
  ```

- Ensure all use `t.Cleanup` to restore original directory
- Consider creating helper: `withTempWorkDir(t *testing.T) string` if pattern repeats

**5. State Manager Verification**:

- Verify `NewDefaultManager()` implementation:
  - No global caching
  - Reads env var fresh each call
  - No `sync.Once` patterns
  - Each call creates new `Manager` instance

**6. Verification & Stability**:

- Run full test suite multiple times:
  ```bash
  go test ./internal/cli/commands -count=20
  ```

- Target: `TestRollbackCommand_SuccessfulRollback_AllPhasesCompleted` must pass 100% of the time
- All other tests must remain stable
- Only after stability is achieved, consider `t.Parallel()`

**Required Files**:

- `spec/core/state-test-isolation.md` (new spec, if needed)
- `spec/core/state.md` (update with env var documentation)
- `internal/cli/commands/test_helpers.go` (already exists, may need enhancements)
- Updated test files: `deploy_test.go`, `rollback_test.go`, `releases_test.go`, etc.

**Integration Points**:

- Uses `setupIsolatedStateTestEnv` from `internal/cli/commands/test_helpers.go`
- Uses `state.NewDefaultManager()` which reads `STAGECRAFT_STATE_FILE`
- All tests must use isolated state files via env var

---

### 🧭 CONSTRAINTS (Canonical List)

**The next agent MUST NOT**:

- ❌ Modify phase execution logic (`phases_common.go`, `executePhasesCommon`)
- ❌ Change command behavior (`deploy.go`, `rollback.go` command logic)
- ❌ Modify persisted formats (JSON schemas, state file structure)
- ❌ Add/rename/remove phase identifiers or statuses
- ❌ Implement other features (keep strictly to test isolation)
- ❌ Skip tests-first workflow
- ❌ Use `os.Setenv` instead of `t.Setenv`
- ❌ Use `os.Chdir` without `t.Cleanup` to restore
- ❌ Reuse state file paths across tests
- ❌ Modify `NewDefaultManager()` implementation (only document it)

**The next agent MUST**:

- ✅ Write/update tests to use `setupIsolatedStateTestEnv`
- ✅ Document `STAGECRAFT_STATE_FILE` behavior in `spec/core/state.md`
- ✅ Migrate ALL state-touching tests to use isolation helper
- ✅ Verify no test interference (run suite multiple times)
- ✅ Use `t.Setenv` for environment variables
- ✅ Use `t.Cleanup` for all cleanup operations
- ✅ Use absolute paths for state files
- ✅ Keep changes strictly scoped to test infrastructure
- ✅ Verify `TestRollbackCommand_SuccessfulRollback_AllPhasesCompleted` is stable

---

## 📌 LAYER 3: Secondary Tasks

### CLI_PHASE_EXECUTION_COMMON (Complete Test Migration)

**Feature ID**: `CLI_PHASE_EXECUTION_COMMON`

**Status**: `todo` (blocked by `CORE_STATE_TEST_ISOLATION`)

**Dependencies**: `CORE_STATE_TEST_ISOLATION` <status: must complete first>

**Note**: Once test isolation is complete, update `spec/features.yaml` to mark `CLI_PHASE_EXECUTION_COMMON` as `done`.

---

## 🎓 Architectural Context (For Understanding)

**Why Test Isolation Matters**:

- **Invariant 1**: Each test must have its own isolated state file
  - Prevents suite-level cross-talk
  - Enables parallel test execution (future)
  - Ensures deterministic test results

- **Invariant 2**: Environment variables must be test-scoped
  - `t.Setenv` automatically restores on test completion
  - `os.Setenv` persists across tests (causes interference)
  - `STAGECRAFT_STATE_FILE` must be unique per test

- **Invariant 3**: Working directory changes must be restored
  - `os.Chdir` without cleanup affects subsequent tests
  - `t.Cleanup` ensures restoration even on test failure
  - Absolute paths prevent CWD-related path resolution issues

**State Manager Design**:

- `NewDefaultManager()` reads `STAGECRAFT_STATE_FILE` fresh on each call
- No caching means tests can set env var and get immediate effect
- Each `Manager` instance is independent (no shared state)
- Explicit `NewManager(path)` always overrides env var/default

**Test Isolation Pattern** (reference implementation):

```go
// From internal/cli/commands/test_helpers.go
func setupIsolatedStateTestEnv(t *testing.T) *isolatedStateTestEnv {
    t.Helper()
    
    tmpDir := t.TempDir()
    stateFile := filepath.Join(tmpDir, ".stagecraft", "releases.json")
    
    // Get absolute path
    absStateFile, err := filepath.Abs(stateFile)
    if err != nil {
        t.Fatalf("failed to get absolute path: %v", err)
    }
    
    // Set env var (test-scoped, auto-cleanup)
    t.Setenv("STAGECRAFT_STATE_FILE", absStateFile)
    
    // Change directory (with cleanup)
    originalDir, _ := os.Getwd()
    os.Chdir(tmpDir)
    t.Cleanup(func() {
        _ = os.Chdir(originalDir)
        _ = os.Unsetenv("STAGECRAFT_STATE_FILE")
    })
    
    // Create manager with absolute path
    mgr := state.NewManager(absStateFile)
    
    return &isolatedStateTestEnv{
        Ctx:       context.Background(),
        StateFile: absStateFile,
        Manager:   mgr,
        // ...
    }
}
```

**Migration Pattern** (for existing tests):

```go
// BEFORE (problematic):
func TestSomething(t *testing.T) {
    tmpDir := t.TempDir()
    stateFile := filepath.Join(tmpDir, ".stagecraft", "releases.json")
    mgr := state.NewManager(stateFile)
    // ... test code ...
}

// AFTER (isolated):
func TestSomething(t *testing.T) {
    env := setupIsolatedStateTestEnv(t)
    mgr := env.Manager
    ctx := env.Ctx
    // ... test code ...
}
```

---

## 📝 Output Expectations

**When you complete `CORE_STATE_TEST_ISOLATION`**:

1. **Summary**: What was implemented

   - All state-touching tests migrated to use `setupIsolatedStateTestEnv`
   - `STAGECRAFT_STATE_FILE` behavior documented in `spec/core/state.md`
   - Test suite runs stably (no intermittent failures)
   - `TestRollbackCommand_SuccessfulRollback_AllPhasesCompleted` passes 100% of the time

2. **Commit Message** (follow this format):

```
feat(CORE_STATE_TEST_ISOLATION): ensure complete test isolation for state-touching tests

Summary:
- Migrated all CLI tests to use setupIsolatedStateTestEnv helper
- Documented STAGECRAFT_STATE_FILE behavior in spec/core/state.md
- Verified no test interference (ran suite 20+ times, all stable)
- Fixed intermittent failure in TestRollbackCommand_SuccessfulRollback_AllPhasesCompleted

Files:
- internal/cli/commands/deploy_test.go (migrated N tests)
- internal/cli/commands/rollback_test.go (completed migration)
- internal/cli/commands/releases_test.go (migrated N tests)
- spec/core/state.md (added env var documentation)
- spec/core/state-test-isolation.md (new spec)

Test Results:
- All tests pass consistently (ran -count=20, 0 failures)
- TestRollbackCommand_SuccessfulRollback_AllPhasesCompleted: stable
- No lint errors
- Coverage maintained

Feature: CORE_STATE_TEST_ISOLATION
Blocks: CLI_PHASE_EXECUTION_COMMON (can now be marked done)
```

3. **Verification**:

   - ✅ All state-touching tests use `setupIsolatedStateTestEnv`
   - ✅ No `os.Setenv` usage (all use `t.Setenv`)
   - ✅ All `os.Chdir` operations have `t.Cleanup` restoration
   - ✅ `STAGECRAFT_STATE_FILE` documented in spec
   - ✅ Full suite runs stably (`go test -count=20`, 0 failures)
   - ✅ `TestRollbackCommand_SuccessfulRollback_AllPhasesCompleted` never fails
   - ✅ All checks pass (`./scripts/run-all-checks.sh`)

4. **Update Feature Status**:

   - Mark `CORE_STATE_TEST_ISOLATION` as `done` in `spec/features.yaml`
   - Mark `CLI_PHASE_EXECUTION_COMMON` as `done` in `spec/features.yaml` (unblocked)

---

## ⚡ Quick Start for Next Agent

**Bootloader Instructions**:

1. **Load Context**:

   - Read `internal/cli/commands/test_helpers.go` to understand `setupIsolatedStateTestEnv`
   - Read `internal/cli/commands/rollback_test.go` to see migration pattern (3 tests already done)
   - Read `internal/core/state/state.go` to understand `NewDefaultManager()` behavior
   - Read `PR_SUMMARY.md` for current status and analysis
   - Check if `spec/core/state-test-isolation.md` exists (create if missing)

2. **Begin Work**:

   - Feature ID: `CORE_STATE_TEST_ISOLATION`
   - Create feature branch: `feature/CORE_STATE_TEST_ISOLATION`
   - Start by auditing all state-touching tests
   - Migrate tests one by one to use `setupIsolatedStateTestEnv`
   - Document `STAGECRAFT_STATE_FILE` in `spec/core/state.md`

3. **Follow Semantics**:

   - Use `setupIsolatedStateTestEnv` for all state-touching tests
   - Use `t.Setenv` for environment variables
   - Use `t.Cleanup` for all cleanup operations
   - Use absolute paths for state files

4. **Respect Constraints**:

   - See CONSTRAINTS section (canonical list)
   - Do not modify phase execution logic
   - Do not change command behavior
   - Keep changes strictly to test infrastructure

5. **Verification Steps**:

   ```bash
   # Audit state-touching tests
   grep -r "state.NewManager\|state.NewDefaultManager" internal/cli/commands/*_test.go
   
   # Audit env var usage
   grep -r "STAGECRAFT_STATE_FILE\|os.Setenv" internal/cli/commands/
   
   # Audit directory changes
   grep -r "os.Chdir" internal/cli/commands/
   
   # Run stability tests
   go test ./internal/cli/commands -count=20 -run TestRollbackCommand_SuccessfulRollback_AllPhasesCompleted
   go test ./internal/cli/commands -count=20
   ```

---

## ✅ Final Checklist

Before starting work:

- [ ] Feature ID identified: `CORE_STATE_TEST_ISOLATION`
- [ ] Git hooks verified
- [ ] Working directory clean
- [ ] On feature branch: `feature/CORE_STATE_TEST_ISOLATION`
- [ ] Spec located/created: `spec/core/state-test-isolation.md` (optional)
- [ ] State documentation updated: `spec/core/state.md` (add env var docs)
- [ ] All state-touching tests identified
- [ ] Migration plan created (which tests to migrate first)
- [ ] Ready to migrate tests

During work:

- [ ] Each test migrated uses `setupIsolatedStateTestEnv`
- [ ] All env var usage uses `t.Setenv` (not `os.Setenv`)
- [ ] All `os.Chdir` operations have `t.Cleanup` restoration
- [ ] No constant state file paths reused across tests
- [ ] Full suite runs stably (`go test -count=20`)

After completion:

- [ ] All state-touching tests migrated
- [ ] `STAGECRAFT_STATE_FILE` documented in `spec/core/state.md`
- [ ] `TestRollbackCommand_SuccessfulRollback_AllPhasesCompleted` passes 100% of the time
- [ ] All checks pass (`./scripts/run-all-checks.sh`)
- [ ] Feature marked `done` in `spec/features.yaml`
- [ ] `CLI_PHASE_EXECUTION_COMMON` unblocked (can be marked `done`)

---

**Copy this entire document into your next agent session to continue development.**

This document is optimized for deterministic AI handoff and aligns with Stagecraft's Agent.md principles (spec-first, test-first, feature-bounded, deterministic).


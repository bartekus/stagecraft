> **Superseded by** `docs/context-handoff/CONTEXT_LOG.md`. Kept for historical reference. New context handoffs MUST be added to the context log.

⸻

docs/context-handoff/CLI_ROLLBACK-to-CLI_PHASE_EXECUTION_COMMON.md

---

## 📋 NEXT AGENT CONTEXT — After Completing Feature CLI_ROLLBACK

---

## 🎉 LAYER 1: What Just Happened

### Feature Complete: CLI_ROLLBACK

**Feature ID**: `CLI_ROLLBACK`

**Status**: ✅ Implemented, fully tested, and merged

**PR**: #<PR_NUMBER> (<PR_URL>)

**Commit**: `<COMMIT_HASH>` - `feat(CLI_ROLLBACK): implement rollback command`

### What Now Exists

**Package**: `internal/cli/commands/`

- `stagecraft rollback` command with three target modes:

  - `--to-previous`

  - `--to-release=<id>`

  - `--to-version=<version>`

- Rollback is modelled as a **new release** that re-deploys the target version (not mutation in place)

- Target resolution + validation:

  - Must exist

  - Must belong to the correct environment

  - Must be fully deployed (all phases completed)

  - Cannot be the current release (when a current release exists)

- Dry-run semantics:

  - Does **not** create a release

  - Does **not** execute phases

  - Logs what *would* happen; may generate a plan for debug output

- Phase execution reuses the deploy pipeline semantics:

  - Phases run in order: build → push → migrate_pre → rollout → migrate_post → finalize

  - On failure: current phase marked failed, downstream phases marked skipped

- Rollback phase execution uses **dependency injection**:

  - `PhaseFns` struct encapsulating phase functions

  - `runRollbackWithPhases` used internally, `runRollback` uses `defaultPhaseFns`

**APIs Available**:

```go

// Command wiring

func NewRollbackCommand() *cobra.Command

// Entry points

func runRollback(cmd *cobra.Command, args []string) error

func runRollbackWithPhases(cmd *cobra.Command, args []string, fns PhaseFns) error

// Phase execution

type PhaseFns struct {

    Build       func(context.Context, *core.Plan, logging.Logger) error

    Push        func(context.Context, *core.Plan, logging.Logger) error

    MigratePre  func(context.Context, *core.Plan, logging.Logger) error

    Rollout     func(context.Context, *core.Plan, logging.Logger) error

    MigratePost func(context.Context, *core.Plan, logging.Logger) error

    Finalize    func(context.Context, *core.Plan, logging.Logger) error

}

func executePhasesRollback(

    ctx context.Context,

    stateMgr *state.Manager,

    releaseID string,

    plan *core.Plan,

    logger logging.Logger,

    fns PhaseFns,

) error

// Helpers

func allPhasesRollback() []state.ReleasePhase

func orderedPhasesRollback() []state.ReleasePhase

func markDownstreamPhasesSkippedRollback(

    ctx context.Context,

    stateMgr *state.Manager,

    releaseID string,

    failedPhase state.ReleasePhase,

    logger logging.Logger,

)

func markAllPhasesFailedRollback(

    ctx context.Context,

    stateMgr *state.Manager,

    releaseID string,

    logger logging.Logger,

) error

```

**Files Created**:

- `internal/cli/commands/rollback.go`

- `internal/cli/commands/rollback_test.go`

- `spec/commands/rollback.md`

**Files Updated**:

- `internal/cli/root.go` — Registered rollback command

- `spec/features.yaml` — Marked CLI_ROLLBACK as done

---

## 🎯 LAYER 2: Immediate Next Task

### Implement CLI_PHASE_EXECUTION_COMMON

**Feature ID**: `CLI_PHASE_EXECUTION_COMMON`

**Status**: `todo`

**Spec**: `spec/core/phase-execution-common.md` (✅ already created)

**Dependencies**:

- ✅ `CORE_STATE` — ready

- ✅ `CORE_PLAN` — ready

- ✅ `CLI_DEPLOY` — partially implemented, uses legacy global phase functions

- ✅ `CLI_ROLLBACK` — done, uses DI (PhaseFns + runRollbackWithPhases)

**⚠️ SCOPE REMINDER**: All work in this handoff MUST be scoped strictly to `CLI_PHASE_EXECUTION_COMMON`.

Do not modify unrelated features or change high-level semantics of deploy/rollback beyond unifying the existing behavior and fixing test/CI flakiness.

**Reference Specs**:

- `spec/core/state.md`

- `spec/commands/deploy.md`

- `spec/commands/rollback.md`

- `spec/core/phase-execution-common.md` (✅ already created)

---

## 🔥 Why This Feature Fixes CI

**Current Problem**: CI breaks because:

- ✅ Rollback has been refactored to use DI and per-test helpers
- ❌ Deploy still uses legacy global phase function overrides in tests
- ❌ Phase execution semantics live in two places: deploy and rollback

**This Leads To**:

- Shared mutable globals (racey and order-dependent)
- Slight behavioral differences between deploy and rollback
- Tests that pass alone but flake when run in the full suite

**CLI_PHASE_EXECUTION_COMMON is the correct surgical feature to**:

- ✅ Centralize phase execution into a single helper
- ✅ Move deploy onto the same DI pattern rollback uses
- ✅ Remove all remaining global overrides around phase functions
- ✅ Make it safe to eventually run tests in parallel

**Note**: `PhaseFns` is currently defined in `rollback.go` (lines 32-41). It needs to be moved to the shared `phases_common.go` file to avoid duplication.

---

## 🧪 MANDATORY WORKFLOW — Tests First

### Step 1: Write Failing Tests

Before writing ANY implementation code:

1. **Create/extend test files**:

   - `internal/cli/commands/phases_common_test.go` (NEW - unit tests for common helper)
   - `internal/cli/commands/deploy_test.go` (UPDATE - remove global overrides, add DI tests)

2. **In `phases_common_test.go`, write tests for common helper semantics**:

   **Happy Path Test**:
   - Use a fake `PhaseFns` where each phase function records calls in a slice and returns `nil`
   - Assert `executePhasesCommon`:
     - Calls phases in order: build → push → migrate_pre → rollout → migrate_post → finalize
     - Updates phase statuses: pending → running → completed
     - All phases end up with `StatusCompleted`

   **Failure Path Test**:
   - Create `PhaseFns` where e.g. `Rollout` returns `fmt.Errorf("boom")`
   - Assert:
     - All earlier phases (build, push, migrate_pre) end up `StatusCompleted`
     - `rollout` is `StatusFailed`
     - `migrate_post` and `finalize` are `StatusSkipped`
     - Error is returned

   These tests should fail initially because `executePhasesCommon` does not exist yet.

3. **In `deploy_test.go`, update tests to express the desired DI shape**:

   - Introduce tests that expect a `runDeployWithPhases` entry and DI:
     - Something like `executeDeployWithPhases(t, PhaseFns{...})` helper
   - Remove or comment out direct usage of global phase function overrides
   - Tests should now refer to a yet-to-exist DI path
   - Let them fail with `undefined: runDeployWithPhases` or similar

   **Goal**: The suite should fail in a way that forces you to implement `PhaseFns`/`executePhasesCommon`/`runDeployWithPhases` to make them pass.

4. **Run tests** - they MUST fail

5. **Only then** begin implementation

**Test Pattern** (follow existing test patterns):

- Follow:
  - `internal/cli/commands/deploy_test.go`
  - `internal/cli/commands/rollback_test.go`
- Use golden tests for CLI behavior where applicable
- Use DI (`PhaseFns`) rather than global overrides
- Test:
  - Phase sequences
  - State transitions in `.stagecraft/releases.json`
  - Dry-run semantics vs non-dry-run
- Use temp directories for config and state (like rollback tests)
- Inspect state using `state.NewManager(path)` and `ListReleases` rather than touching files manually

---

## 🛠 Implementation Outline (Step-by-Step)

### Step 2: Implement PhaseFns and Common Helpers

**File**: `internal/cli/commands/phases_common.go` (NEW)

1. **Move `PhaseFns` from `rollback.go`** (lines 32-41) to this shared file
2. **Implement helper functions**:

   ```go
   // Returns all phases in canonical order
   func allPhasesCommon() []state.ReleasePhase
   
   // Convenience alias
   func orderedPhasesCommon() []state.ReleasePhase
   
   // Marks downstream phases as skipped after failure
   func markDownstreamPhasesSkippedCommon(...) error
   
   // Marks all phases as failed (for planner failures)
   func markAllPhasesFailedCommon(...) error
   
   // Main shared phase execution entry
   func executePhasesCommon(...) error
   
   // Helper to select phase function from PhaseFns
   func phaseFnFor(phase state.ReleasePhase, fns PhaseFns) (func(...) error, error)
   ```

3. **See code skeleton below** for complete implementation

### Step 3: Refactor Deploy to Use DI

**File**: `internal/cli/commands/deploy.go` (UPDATE)

1. **Introduce DI entry point**:

   ```go
   func runDeploy(cmd *cobra.Command, args []string) error {
       return runDeployWithPhases(cmd, args, defaultPhaseFns)
   }
   ```

2. **Implement `runDeployWithPhases`** that:
   - Resolves flags and loads config
   - Sets up logger
   - **Handles deploy dry-run semantics** (as spec'd in `commands/deploy.md`):
     - If dry-run, behave exactly as before (including whether it creates a release)
     - **CRITICAL**: Do NOT call `executePhasesCommon` in dry-run
   - Creates release and plan
   - Calls `executePhasesCommon(ctx, stateMgr, release.ID, plan, logger, fns)`

3. **Remove any direct usage of global phase function pointers** from deploy code
   - If there were `var buildPhaseFn = ...` style globals, they can remain as the default implementation targets
   - But tests should no longer mutate them

### Step 4: Refactor Rollback to Use Common Helper

**File**: `internal/cli/commands/rollback.go` (UPDATE)

1. **Remove `PhaseFns` definition** (now in `phases_common.go`)
2. **Replace `executePhasesRollback` implementation** with:
   - A thin wrapper over `executePhasesCommon`, OR
   - Delete `executePhasesRollback` entirely and call `executePhasesCommon` directly from `runRollbackWithPhases`
3. **Keep all rollback-specific bits intact**:
   - Target resolution and validation
   - Dry-run semantics (no release, no phases)
   - Version/commit copying from target

**Note**: Rollback tests should continue to pass without modification other than import path/name changes if you rename helpers.

### Step 5: Fix Tests to Stop Using Globals

**File**: `internal/cli/commands/deploy_test.go` (UPDATE)

1. **Replace global mutation patterns**:

   **OLD (BAD)**:
   ```go
   old := buildPhaseFn
   buildPhaseFn = func(...) error { ... }
   defer func() { buildPhaseFn = old }()
   ```

   **NEW (GOOD)**:
   ```go
   fns := defaultPhaseFns
   fns.Rollout = func(...) error { return fmt.Errorf("boom") }
   executeDeployWithPhases(t, fns, args...)
   ```

2. **Ensure tests**:
   - Use temp directories for config and state (like rollback tests)
   - Inspect state using `state.NewManager(path)` and `ListReleases` rather than touching files manually

**File**: `internal/cli/commands/phases_common_test.go` (NEW)

- Add assertions to fully exercise the helper in isolation
- See code skeleton below for example tests

### Step 6: Run Full Suite and Iterate

1. Run `go test ./...`
2. If anything still flakes:
   - Look for remaining global mutation patterns (especially in deploy tests)
   - Check any shared temp directories or working directory changes that might collide across tests
3. Keep everything within the scope of `CLI_PHASE_EXECUTION_COMMON` (phase execution, DI, and tests)

---

## 📦 Code Skeletons

### `internal/cli/commands/phases_common.go`

See the complete skeleton provided in the user's feedback. Key points:

- `PhaseFns` struct (move from `rollback.go`)
- `allPhasesCommon()` and `orderedPhasesCommon()` helpers
- `markDownstreamPhasesSkippedCommon()` - marks phases after failure as skipped
- `markAllPhasesFailedCommon()` - marks all phases as failed (for planner failures)
- `executePhasesCommon()` - main shared execution logic
- `phaseFnFor()` - selects appropriate function from `PhaseFns`

**Important**: The skeleton includes proper error handling, state updates, and logging.

### `internal/cli/commands/phases_common_test.go`

Two key tests provided:

1. **`TestExecutePhasesCommon_AllSuccess`**:
   - Happy path: all phases succeed
   - Verifies call order and final statuses

2. **`TestExecutePhasesCommon_RolloutFailureSkipsDownstream`**:
   - Failure path: rollout fails
   - Verifies upstream phases completed, failing phase failed, downstream skipped

**Note**: If `state.NewDefaultManager` expects `.stagecraft` to exist, you may need `os.MkdirAll(filepath.Join(tmpDir, ".stagecraft"), 0o700)` before creating the manager.

---

## 📁 Required Files Summary

- ✅ `spec/core/phase-execution-common.md` — already created
- 🆕 `internal/cli/commands/phases_common.go` — shared helpers for phase ordering + execution
- 🆕 `internal/cli/commands/phases_common_test.go` — unit tests for common helper
- 🔄 `internal/cli/commands/deploy.go` — updated to use `runDeployWithPhases` + `PhaseFns`
- 🔄 `internal/cli/commands/deploy_test.go` — updated to use DI instead of global overrides
- 🔄 `internal/cli/commands/rollback.go` — updated to use `executePhasesCommon` (remove duplicate `PhaseFns`)

---

## 🔗 Integration Points

- Uses `core.NewPlanner(cfg)` and `PlanDeploy` from `internal/core/plan`
- Uses `state.Manager` methods:
  - `CreateRelease`
  - `UpdatePhase`
  - `GetRelease` (for downstream skip logic)
  - `ListReleases` (for tests)
- Uses CLI logging from `pkg/logging`
- Preserves existing CLI flags and behavior for:
  - `deploy`
  - `rollback`

---

## 💻 Code Reference (Implementation Skeletons)

The following code skeletons are provided as reference. They should be adapted to match the actual codebase structure and existing patterns.

### `internal/cli/commands/phases_common.go` (Complete Skeleton)

**Key Components**:

1. **`PhaseFns` struct** - Move from `rollback.go` (lines 32-41)
2. **`allPhasesCommon()`** - Returns phases in canonical order
3. **`orderedPhasesCommon()`** - Convenience alias
4. **`markDownstreamPhasesSkippedCommon()`** - Marks downstream phases as skipped after failure
5. **`markAllPhasesFailedCommon()`** - Marks all phases as failed (for planner failures)
6. **`executePhasesCommon()`** - Main shared execution logic
7. **`phaseFnFor()`** - Selects appropriate function from `PhaseFns`

**Implementation Notes**:
- Use proper error wrapping with `fmt.Errorf("context: %w", err)`
- Log phase transitions using structured logging
- Handle state update failures as fatal (abort immediately)
- Ensure deterministic behavior (no timestamps, no randomness)

### `internal/cli/commands/phases_common_test.go` (Test Skeletons)

**Required Tests**:

1. **`TestExecutePhasesCommon_AllSuccess`**:
   - Happy path: all phases succeed
   - Verifies call order matches canonical order
   - Verifies all phases end with `StatusCompleted`

2. **`TestExecutePhasesCommon_RolloutFailureSkipsDownstream`**:
   - Failure path: rollout fails
   - Verifies upstream phases (build, push, migrate_pre) are `StatusCompleted`
   - Verifies failing phase (rollout) is `StatusFailed`
   - Verifies downstream phases (migrate_post, finalize) are `StatusSkipped`
   - Verifies error is returned

**Test Setup**:
- Use `t.TempDir()` for isolated state files
- Use `os.Chdir(tmpDir)` to ensure state manager uses correct directory
- Create release using `stateMgr.CreateRelease()`
- Use fake `PhaseFns` that record calls and simulate success/failure

**Note**: If `state.NewDefaultManager` expects `.stagecraft` to exist, add:
```go
os.MkdirAll(filepath.Join(tmpDir, ".stagecraft"), 0o700)
```

**See the complete code skeletons in the user's feedback for full implementation details.**

---

## 🧭 CONSTRAINTS (Canonical List)

The next agent MUST NOT:

- ❌ Change the external CLI contract for deploy or rollback (flags, outputs, top-level error semantics)

- ❌ Modify persisted formats (JSON in `.stagecraft/releases.json`, schemas)

- ❌ Add/rename/remove phase identifiers or statuses

- ❌ Introduce new side effects in dry-run behavior beyond what specs already define

- ❌ Mix unrelated features (e.g., `CLI_PLAN`, `CLI_BUILD`, infra commands) into this PR

- ❌ Reintroduce global mutable state for phase execution

- ❌ Write directly to state files; always use `state.Manager`

The next agent MUST:

- ✅ Write failing tests first, covering deploy + shared phase execution semantics

- ✅ Follow existing CLI patterns (`deploy.go`, `rollback.go`)

- ✅ Centralize phase execution logic into a shared helper (or at least shared DI pattern)

- ✅ Ensure rollback continues to pass all existing tests unchanged

- ✅ Remove any remaining test patterns that mutate global phase functions

- ✅ Keep changes strictly scoped to `CLI_PHASE_EXECUTION_COMMON`

- ✅ Update or create the spec `spec/core/phase-execution-common.md`

---

## 📌 LAYER 3: Secondary Tasks

### CLI_PLAN

**Feature ID**: `CLI_PLAN`

**Status**: `todo`

**Dependencies**:

- `CORE_PLAN` (ready)

- `CLI_PHASE_EXECUTION_COMMON` (recommended, so plan output matches shared phase semantics)

**Do NOT begin until `CLI_PHASE_EXECUTION_COMMON` is complete.** (See CONSTRAINTS section)

---

### CLI_BUILD (Design Only)

**Feature ID**: `CLI_BUILD`

**Status**: `todo`

**Dependencies**:

- `PROVIDER_BACKEND_GENERIC`

- `PROVIDER_BACKEND_ENCORE`

- `CLI_PHASE_EXECUTION_COMMON` (for eventual phase alignment)

**Do NOT implement until prerequisites are complete.**

Design can consider:

- Mapping build phases to provider capabilities

- Integrating image tagging/versioning with `CORE_STATE`

- Reusing phase execution semantics for build-only operations

---

## 🎓 Architectural Context (For Understanding)

### Why These Design Decisions Matter:

- **Shared phase semantics reduce drift**: Having a single place that defines how phases execute (build/push/migrate/rollout/etc.) ensures deploy and rollback stay in sync and prevents subtle divergence over time.

- **Dependency injection enables deterministic tests**: Injecting `PhaseFns` lets tests precisely control behavior (success, failure, latency) without mutating globals, which is critical for CI stability and eventual parallel testing.

- **CI reliability**: Removing global mutable state and centralizing behavior is key to eliminating flaky tests that pass in isolation but fail in the full suite.

### Current State Analysis:

- ✅ **Rollback**: Already uses DI (`PhaseFns`, `runRollbackWithPhases`, `executePhasesRollback`)
- ❌ **Deploy**: Still uses legacy global phase function overrides in tests
- ❌ **Phase Execution**: Duplicated between `rollback.go` and `deploy.go`

### What This Feature Achieves:

- ✅ **Single source of truth**: `executePhasesCommon` in `phases_common.go`
- ✅ **Consistent behavior**: Deploy and rollback use identical phase execution logic
- ✅ **Test safety**: No more global mutations, tests are deterministic and parallel-safe
- ✅ **CI stability**: Eliminates flaky tests caused by shared mutable state

### Integration Pattern Example:

```go
// Example: Calling shared phase execution from deploy
func runDeployWithPhases(cmd *cobra.Command, args []string, fns PhaseFns) error {
    ctx := cmd.Context()
    if ctx == nil {
        ctx = context.Background()
    }

    // 1) Resolve flags, load config, create logger, etc.
    flags, err := ResolveFlags(cmd, nil)
    if err != nil {
        return fmt.Errorf("resolving flags: %w", err)
    }

    cfg, err := config.Load(flags.Config)
    if err != nil {
        return fmt.Errorf("loading config: %w", err)
    }

    flags, err = ResolveFlags(cmd, cfg)
    if err != nil {
        return fmt.Errorf("resolving flags with config: %w", err)
    }

    logger := logging.NewLogger(flags.Verbose)

    // 2) Initialize state manager and create release
    stateMgr := state.NewDefaultManager()
    release, err := stateMgr.CreateRelease(ctx, flags.Env, flags.Version, flags.CommitSHA)
    if err != nil {
        return fmt.Errorf("creating release: %w", err)
    }

    // 3) Plan deployment
    planner := core.NewPlanner(cfg)
    plan, err := planner.PlanDeploy(flags.Env)
    if err != nil {
        _ = markAllPhasesFailedCommon(ctx, stateMgr, release.ID, logger)
        return fmt.Errorf("generating deployment plan: %w", err)
    }

    // 4) Execute phases using shared helper
    if err := executePhasesCommon(ctx, stateMgr, release.ID, plan, logger, fns); err != nil {
        return fmt.Errorf("deploy failed: %w", err)
    }

    return nil
}
```

---

## 📝 Output Expectations

When you complete `CLI_PHASE_EXECUTION_COMMON`:

1. **Summary**: What was implemented

2. **Commit Message** (follow this format):

```
feat(CLI_PHASE_EXECUTION_COMMON): unify deploy and rollback phase execution

Summary:
- Extracted shared phase execution helper into phases_common.go
- Updated deploy to use runDeployWithPhases and PhaseFns DI
- Removed global phase function overrides from deploy tests
- Ensured rollback continues to use the shared semantics without behavior changes
- Added spec/core/phase-execution-common.md documenting shared semantics

Files:
- internal/cli/commands/phases_common.go
- internal/cli/commands/phases_common_test.go
- internal/cli/commands/deploy.go
- internal/cli/commands/deploy_test.go
- spec/core/phase-execution-common.md

Test Results:
- All tests pass
- Coverage meets targets
- No lint errors

Feature: CLI_PHASE_EXECUTION_COMMON
Spec: spec/core/phase-execution-common.md
```

3. **Verification**:

   - ✅ Tests were written first (before implementation)

   - ✅ No unrelated changes were made

   - ✅ Feature boundaries respected (only `CLI_PHASE_EXECUTION_COMMON` code)

   - ✅ All checks pass (`./scripts/run-all-checks.sh`)

   - ✅ CI no longer flakes due to phase-execution globals

---

## ⚡ Quick Start for Next Agent

### Bootloader Instructions:

1. **Load Context**:

   - Read `internal/cli/commands/deploy.go` to understand current behavior

   - Read `internal/cli/commands/rollback.go` for the DI pattern

   - Read `internal/core/state/state.go` for state semantics

   - Read `spec/commands/deploy.md` and `spec/commands/rollback.md`

   - Create `spec/core/phase-execution-common.md` to capture shared behavior

2. **Begin Work**:

   - Feature ID: `CLI_PHASE_EXECUTION_COMMON`

   - ✅ Feature branch already exists: `feature/CLI_PHASE_EXECUTION_COMMON-unify-phase-execution`

   - Start with tests (Step 1):

     - 🆕 `internal/cli/commands/phases_common_test.go` (NEW - write first)

     - 🔄 `internal/cli/commands/deploy_test.go` (UPDATE - remove globals, add DI tests)

   - Write failing tests first (they should fail because `executePhasesCommon` doesn't exist yet)

   - Then implement (Steps 2-5):

     - Step 2: Create `phases_common.go` with shared helpers
     - Step 3: Refactor `deploy.go` to use DI
     - Step 4: Refactor `rollback.go` to use common helper
     - Step 5: Fix tests to stop using globals
     - Step 6: Run full suite and iterate

3. **Follow Semantics**:

   - Use existing phase identifiers and statuses (do not rename)

   - Keep rollback behavior identical to current spec

   - Preserve deploy's external behavior (flags, dry-run semantics)

4. **Respect Constraints**:

   - See CONSTRAINTS section (canonical list)

   - Do not implement `CLI_PLAN` or `CLI_BUILD` in this feature

   - Keep feature boundaries clean

---

## ✅ Final Checklist

Before starting work:

- [ ] Feature ID identified: `CLI_PHASE_EXECUTION_COMMON`

- [ ] Git hooks verified

- [ ] Working directory clean

- [ ] On feature branch: `feature/CLI_PHASE_EXECUTION_COMMON-unify-phase-execution`

- [x] Spec located/created: `spec/core/phase-execution-common.md` ✅

- [ ] Tests written first: 
  - [ ] `internal/cli/commands/phases_common_test.go` (NEW)
  - [ ] `internal/cli/commands/deploy_test.go` (UPDATE - remove globals)

- [ ] Tests fail (as expected - `executePhasesCommon` doesn't exist yet)

- [ ] Ready to implement (Steps 2-6)

---

Copy this entire document into your next agent session to continue development.

This document is optimized for deterministic AI handoff and aligns with Stagecraft's Agent.md principles (spec-first, test-first, feature-bounded, deterministic).


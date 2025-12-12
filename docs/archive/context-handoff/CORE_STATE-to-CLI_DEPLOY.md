> **Superseded by** `docs/context-handoff/CONTEXT_LOG.md` section 4.4. Kept for historical reference. New context handoffs MUST be added to the context log.

---

## 📋 NEXT AGENT CONTEXT — After Completing Feature CORE_STATE

---

## 🎉 LAYER 1: What Just Happened

### Feature Complete: CORE_STATE

**Feature ID**: `CORE_STATE`  

**Status**: ✅ Implemented, fully tested, and merged  

**PR**: #3 (https://github.com/bartekus/stagecraft/pull/3)  

**Commit**: `c4d1f32` - `feat(CORE_STATE): Implement state management for release history`

### What Now Exists

**Package**: `internal/core/state/`

- State manager with deterministic behavior

- Millisecond-precision release IDs (`rel-YYYYMMDD-HHMMSSmmm`)

- Immutable snapshot semantics (read-only returns)

- Mutex-protected single-process concurrency

- Atomic temp-file writes with PID scoping

- Ordered release listing (newest first)

- Phase update semantics (validated against spec)

- 88.5% test coverage (exceeds 80% target)

**APIs Available**:

```go
// Create new release (all phases start as StatusPending)
CreateRelease(ctx, env, version, commitSHA) (*Release, error)

// Retrieve releases
GetRelease(ctx, id) (*Release, error)
GetCurrentRelease(ctx, env) (*Release, error)
ListReleases(ctx, env) ([]*Release, error)

// Update deployment phase
UpdatePhase(ctx, releaseID, phase, status) error
```

**Files Created**:

- `internal/core/state/state.go` (458 lines)

- `internal/core/state/state_test.go` (875 lines)

**Files Updated**:

- `spec/core/state.md` - Updated with release ID format, usage example, snapshot semantics

- `spec/features.yaml` - Marked `CORE_STATE` as `done`

---

## 🎯 LAYER 2: Immediate Next Task

### Implement CLI_DEPLOY

**Feature ID**: `CLI_DEPLOY`  

**Status**: `todo`  

**Spec**: `spec/commands/deploy.md` (may need creation)  

**Dependencies**:

- ✅ `CORE_STATE` (ready)

- ✅ `CORE_PLAN` (ready)

- ✅ `CORE_COMPOSE` (ready)

- ⏸ `PROVIDER_NETWORK_TAILSCALE` (todo, not required for first pass)

**⚠️ SCOPE REMINDER**: All work in this handoff MUST be scoped strictly to `CLI_DEPLOY`. Do not implement other features or modify unrelated code.

**Reference Spec**: See `docs/stagecraft-spec.md` section 4.4 for high-level behavior

---

### 🧪 MANDATORY WORKFLOW — Tests First

**Before writing ANY implementation code**:

1. **Create test file**: `internal/cli/commands/deploy_test.go`

2. **Write failing tests** describing:

   - Release creation at deploy start

   - Phase sequencing (Pending → Running → Completed/Failed)

   - Failure semantics (mark failed phase, skip downstream)

   - Integration with `CORE_STATE` via `state.Manager`

   - Error handling for invalid environments/versions

3. **Run tests** - they MUST fail

4. **Only then** begin implementation

**Test Pattern** (follow existing CLI test patterns):

- Use `internal/cli/commands/dev_test.go` as reference

- Use golden tests for CLI output

- Mock external dependencies

- Test phase transitions explicitly

---

### 🛠 Implementation Outline

**1. Create Release at Deploy Start**:

```go
import "stagecraft/internal/core/state"

mgr := state.NewDefaultManager()
release, err := mgr.CreateRelease(ctx, env, version, commitSHA)
// All phases initialized as StatusPending
```

**2. Update Phases During Deployment**:

```go
// Each phase transitions: Pending → Running → Completed (or Failed)
mgr.UpdatePhase(ctx, release.ID, state.PhaseBuild, state.StatusRunning)
// ... build happens ...
mgr.UpdatePhase(ctx, release.ID, state.PhaseBuild, state.StatusCompleted)

// Continue for all phases in order (see CONSTRAINTS section for phase list)
```

**3. Failure Semantics** (implement now):

- Mark current phase as `StatusFailed`

- Mark all downstream phases as `StatusSkipped`

- Only one phase may be `StatusRunning` at a time

- Deploy stops on first failure (do not continue to next phase)

**4. Required Files**:

- `internal/cli/commands/deploy.go` - Main deploy command

- `internal/cli/commands/deploy_test.go` - Tests (write first!)

- `spec/commands/deploy.md` - Spec (create if missing)

- `internal/deploy/` - Deployment orchestration layer (may need creation)

**5. Integration Points**:

- Use `CORE_PLAN` (`internal/core/plan.go`) to generate deployment plan

- Use `CORE_COMPOSE` (`internal/compose/compose.go`) for Docker Compose operations

- Use `state.Manager` for all release tracking

- Call `UpdatePhase()` at each deployment checkpoint

---

### 🧭 CONSTRAINTS (Canonical List)

**The next agent MUST NOT**:

- ❌ Modify existing `CORE_STATE` behavior or code

- ❌ Change JSON state format or structure

- ❌ Add/remove/rename phases (use existing: `PhaseBuild`, `PhasePush`, `PhaseMigratePre`, `PhaseRollout`, `PhaseMigratePost`, `PhaseFinalize`)

- ❌ Implement `CLI_ROLLBACK` now (wait until `CLI_DEPLOY` is complete)

- ❌ Modify unrelated features

- ❌ Write deployment logic without tests first

- ❌ Mix code for multiple features in the same PR

- ❌ Skip the tests-first workflow

- ❌ Write `.stagecraft/releases.json` directly (always use `state.Manager`)

**The next agent MUST**:

- ✅ Write failing tests before implementation

- ✅ Follow existing CLI command patterns (see `internal/cli/commands/dev.go`)

- ✅ Use `state.Manager` for all state operations

- ✅ Update phases in correct order: Build → Push → MigratePre → Rollout → MigratePost → Finalize

- ✅ Handle failures according to the semantics above

- ✅ Create/update spec if missing

- ✅ Keep feature boundaries clean (only `CLI_DEPLOY` code)

---

## 📌 LAYER 3: Secondary Tasks

### CLI_RELEASES

**Feature ID**: `CLI_RELEASES`  

**Status**: `todo`  

**Dependencies**: `CORE_STATE` ✅ (ready)

**Simple Implementation**: List & show releases using `CORE_STATE` APIs

- `stagecraft releases list [--env=ENV]`

- `stagecraft releases show <release-id>`

**Do NOT start until `CLI_DEPLOY` is complete.** (See CONSTRAINTS section)

---

### CLI_ROLLBACK (Design Only)

**Feature ID**: `CLI_ROLLBACK`  

**Status**: `todo`  

**Dependencies**: `CORE_STATE` ✅, `CLI_DEPLOY` (todo)

**Do NOT implement until `CLI_DEPLOY` is complete.** (See CONSTRAINTS section)

Design can consider:

- `--to-previous` (use `PreviousID`)

- `--to-release=<id>` (use `GetRelease()`)

- `--to-version=<version>` (search via `ListReleases()`)

---

## 🎓 Architectural Context (For Understanding)

**Why These Design Decisions Matter**:

- **Snapshot cloning**: Prevents mutation bugs when passing `Release` to other layers

- **Clock injection**: Enables deterministic tests (no `time.Sleep` needed)

- **PID-based temp files**: Reduces multi-process file conflicts

- **Validated phase transitions**: Prevents state drift and invalid phase keys

- **Read-only semantics**: Clear API contract for all consumers

**Integration Pattern Example** (for reference, not required to copy exactly):

```go
// Example: How CLI_DEPLOY should integrate with CORE_STATE
// This is illustrative - adapt to your implementation needs
mgr := state.NewDefaultManager()
release, _ := mgr.CreateRelease(ctx, env, version, commitSHA)

// During deployment
mgr.UpdatePhase(ctx, release.ID, state.PhaseBuild, state.StatusRunning)
// ... build happens ...
mgr.UpdatePhase(ctx, release.ID, state.PhaseBuild, state.StatusCompleted)

// On failure (example pattern - adapt as needed)
mgr.UpdatePhase(ctx, release.ID, state.PhaseBuild, state.StatusFailed)
// Mark downstream as skipped
for _, phase := range []state.ReleasePhase{
    state.PhasePush, state.PhaseMigratePre,
    state.PhaseRollout, state.PhaseMigratePost, state.PhaseFinalize,
} {
    mgr.UpdatePhase(ctx, release.ID, phase, state.StatusSkipped)
}
```

---

## 📝 Output Expectations

**When you complete `CLI_DEPLOY`**:

1. **Summary**: What was implemented

2. **Commit Message** (follow this format):

```
feat(CLI_DEPLOY): implement deploy command with release tracking

Summary:
- Added deploy.go with phase tracking integration
- Implemented phase update semantics (Pending → Running → Completed/Failed)
- Integrated CORE_PLAN and CORE_COMPOSE
- Added failure semantics (mark failed, skip downstream)
- Created deploy_test.go with comprehensive tests
- Created/updated spec/commands/deploy.md

Files:
- internal/cli/commands/deploy.go
- internal/cli/commands/deploy_test.go
- spec/commands/deploy.md
- internal/deploy/ (if created)

Test Results:
- All tests pass
- Coverage meets targets
- No linting errors

Feature: CLI_DEPLOY
Spec: spec/commands/deploy.md
```

3. **Verification**:

   - ✅ Tests were written first (before implementation)

   - ✅ No unrelated changes were made

   - ✅ Feature boundaries respected (only `CLI_DEPLOY` code)

   - ✅ All checks pass (`./scripts/run-all-checks.sh`)

---

## ⚡ Quick Start for Next Agent

**Bootloader Instructions**:

1. **Load Context**:

   - Read `internal/core/state/state.go` to understand API

   - Read `spec/core/state.md` for state semantics

   - Read `internal/cli/commands/dev.go` for CLI pattern reference

   - Check if `spec/commands/deploy.md` exists

2. **Begin Work**:

   - Feature ID: `CLI_DEPLOY`

   - Create feature branch: `feature/CLI_DEPLOY`

   - Start with tests: `internal/cli/commands/deploy_test.go`

   - Write failing tests first

   - Then implement: `internal/cli/commands/deploy.go`

3. **Follow Phase Semantics**:

   - Use existing phases (see CONSTRAINTS section)

   - Follow order: Build → Push → MigratePre → Rollout → MigratePost → Finalize

   - Implement failure semantics as specified

4. **Respect Constraints**:

   - See CONSTRAINTS section (canonical list)

   - Do not modify `CORE_STATE`

   - Do not implement `CLI_ROLLBACK` yet

   - Keep feature boundaries clean

---

## ✅ Final Checklist

Before starting work:

- [ ] Feature ID identified: `CLI_DEPLOY`

- [ ] Git hooks verified

- [ ] Working directory clean

- [ ] On feature branch: `feature/CLI_DEPLOY`

- [ ] Spec located/created: `spec/commands/deploy.md`

- [ ] Tests written first: `internal/cli/commands/deploy_test.go`

- [ ] Tests fail (as expected)

- [ ] Ready to implement

---

**Copy this entire document into your next agent session to continue development.**

This document is optimized for deterministic AI handoff and aligns with Stagecraft's Agent.md principles (spec-first, test-first, feature-bounded, deterministic).


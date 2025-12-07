---
feature: CLI_PHASE_EXECUTION_COMMON
version: v1
status: done
domain: core
inputs:
  flags: []
outputs:
  exit_codes: {}
---
# CORE_PHASE_EXECUTION_COMMON – Shared Phase Execution Semantics

- Feature ID: `CLI_PHASE_EXECUTION_COMMON`
- Status: todo
- Depends on:
  - `CORE_STATE`
  - `CORE_PLAN`
  - `CLI_DEPLOY`
  - `CLI_ROLLBACK`

## Goal

Define a **single, shared contract** for executing deployment phases across CLI commands (`deploy`, `rollback`, and future commands that reuse the same pipeline).

This spec exists to:

- Prevent drift between `deploy` and `rollback` behavior
- Enable deterministic, testable phase execution via dependency injection
- Remove reliance on global mutable state in tests and implementation
- Make CI reliable by eliminating phase-related flakiness

---

## Behaviour

### Phase Set

The shared phase execution logic operates on the following ordered phases:

1. `build`
2. `push`
3. `migrate_pre`
4. `rollout`
5. `migrate_post`
6. `finalize`

These are represented in code as:

```go
type ReleasePhase string

const (
    PhaseBuild       ReleasePhase = "build"
    PhasePush        ReleasePhase = "push"
    PhaseMigratePre  ReleasePhase = "migrate_pre"
    PhaseRollout     ReleasePhase = "rollout"
    PhaseMigratePost ReleasePhase = "migrate_post"
    PhaseFinalize    ReleasePhase = "finalize"
)
```

**Invariant**: The set of phases and their order are canonical. Commands may choose to skip some phases (e.g. not configured, or no-op), but MUST NOT introduce new phase identifiers or change the execution order.

### Phase Statuses

Each phase uses the canonical status enum from `CORE_STATE`:

```go
type PhaseStatus string

const (
    StatusPending   PhaseStatus = "pending"
    StatusRunning   PhaseStatus = "running"
    StatusCompleted PhaseStatus = "completed"
    StatusFailed    PhaseStatus = "failed"
    StatusSkipped   PhaseStatus = "skipped"
)
```

**Invariant**: Shared phase execution MUST ONLY use these statuses.

---

## Shared Execution Semantics

### Input

The shared helper operates with the following inputs:

- `ctx context.Context`
- `stateMgr *state.Manager`
  - Responsible for reading/writing releases to `.stagecraft/releases.json`
- `releaseID string`
  - ID of the release being executed
- `plan *core.Plan`
  - Deployment plan produced by `CORE_PLAN`
- `logger logging.Logger`
- `fns PhaseFns`
  - A struct of phase functions (dependency-injected):

```go
type PhaseFns struct {
    Build       func(context.Context, *core.Plan, logging.Logger) error
    Push        func(context.Context, *core.Plan, logging.Logger) error
    MigratePre  func(context.Context, *core.Plan, logging.Logger) error
    Rollout     func(context.Context, *core.Plan, logging.Logger) error
    MigratePost func(context.Context, *core.Plan, logging.Logger) error
    Finalize    func(context.Context, *core.Plan, logging.Logger) error
}
```

Each function is responsible for performing the work associated with its phase and returning an error if the phase fails.

### Execution Flow

For each phase in order:

1. **Log start**:
   ```
   INFO: Starting phase (phase=<phase_name>)
   ```

2. **Set phase status to running**:
   ```go
   if err := stateMgr.UpdatePhase(ctx, releaseID, phase, state.StatusRunning); err != nil {
       // See failure semantics below
   }
   ```

3. **Execute the phase function**:
   ```go
   err := fns.<PhaseName>(ctx, plan, logger)
   ```

4. **If the function returns no error**:
   - Set phase status to completed:
     ```go
     if err := stateMgr.UpdatePhase(ctx, releaseID, phase, state.StatusCompleted); err != nil {
         // See failure semantics below
     }
     ```
   - Proceed to the next phase.

5. **If the function returns an error**:
   - Set the current phase status to failed
   - Mark all downstream phases as skipped
   - Abort execution and return an error up to the caller

### Failure Semantics

There are three main failure shapes:

#### 1. Phase Work Failure

If a phase function returns an error:

- Update current phase:
  ```go
  _ = stateMgr.UpdatePhase(ctx, releaseID, phase, state.StatusFailed)
  ```
- Mark downstream phases as skipped:
  ```go
  markDownstreamPhasesSkipped(ctx, stateMgr, releaseID, failedPhase, logger)
  ```
  Where `markDownstreamPhasesSkipped`:
  - Iterates over `orderedPhases()`
  - Starts skipping after it has passed the failed phase
  - For each downstream phase:
    - If status is still pending or running, update to skipped
- Return an error like:
  ```go
  return fmt.Errorf("phase %q failed: %w", phase, err)
  ```

#### 2. Planner Failure (Pre-Execution)

If the deployment plan cannot be generated (e.g. `PlanDeploy` fails) before any phases run:

- The caller (e.g. deploy or rollback) MAY mark all phases as failed for the release using:
  ```go
  markAllPhasesFailed(ctx, stateMgr, releaseID, logger)
  ```
- Return a wrapped error like:
  ```go
  return fmt.Errorf("generating deployment plan: %w", err)
  ```

**Note**: The exact behavior here is command-specific:
- `deploy` may choose to mark all phases as failed
- `rollback` may do the same or treat plan failure as an early failure before phases

The shared helper SHOULD NOT assume plan creation; it only operates after a plan exists.

#### 3. State Manager Failure (Status Updates)

If updating a phase status fails (e.g. IO error):

- The helper MUST return an error immediately:
  ```go
  return fmt.Errorf("updating phase %q to %q: %w", phase, status, err)
  ```
- No further attempts should be made to execute more phases.

**Invariant**: State update failures are treated as fatal and abort the entire sequence.

---

## Dry-Run Semantics

Dry-run behavior is command-specific, but shares these invariants:

- The shared phase execution helper MUST NOT be called in dry-run mode.
- No phase status updates should occur in dry-run mode.
- No release should be created in dry-run for rollback (per `CLI_ROLLBACK` spec).
- For deploy, dry-run behavior is defined in `spec/commands/deploy.md` and may differ, but if a release is created in dry-run:
  - The shared helper MUST still be bypassed (no phases executed, no status transitions).

Commands MUST enforce dry-run semantics before invoking shared phase execution.

---

## Integration with CLI Commands

### Deploy

`deploy` uses shared phase execution as follows:

1. Resolve flags (`--env`, `--config`, `--dry-run`, etc.)
2. Load config via `pkg/config`
3. Construct logger via `pkg/logging`
4. Initialize `state.Manager` (default or explicit path)
5. Create new release:
   - Environment from flags
   - Version/commit SHA from flags or auto-detection
   - All phases initialized as pending
6. Generate plan using `core.NewPlanner(cfg).PlanDeploy(env)`
7. If not dry-run:
   - Call shared helper with `defaultPhaseFns`:
     ```go
     err = executePhasesCommon(ctx, stateMgr, release.ID, plan, logger, defaultPhaseFns)
     ```
8. Propagate any error back to the CLI.

### Rollback

`rollback` uses shared phase execution similarly, but:

1. Resolves target release (previous/by-id/by-version)
2. Validates target release:
   - Exists
   - Not current
   - All phases completed
3. If not dry-run:
   - Creates a new rollback release:
     - Environment same as current
     - Version + commit SHA copied from target
   - Generates plan via `PlanDeploy(env)`
   - Calls shared helper with `defaultPhaseFns`:
     ```go
     err = executePhasesCommon(ctx, stateMgr, rollbackRelease.ID, plan, logger, defaultPhaseFns)
     ```
4. Propagates any error back to the CLI.

**Invariant**: Rollback semantics (version/commit copying, validation, dry-run behavior) are governed by `spec/commands/rollback.md`. This spec only governs how phases are executed once a release and plan exist.

---

## Dependency Injection & Testing

### PhaseFns DI

Implementations MUST use `PhaseFns` to decouple phase logic from the orchestration:

- Production code uses `defaultPhaseFns`:
  ```go
  var defaultPhaseFns = PhaseFns{
      Build:       buildPhaseFn,
      Push:        pushPhaseFn,
      MigratePre:  migratePrePhaseFn,
      Rollout:     rolloutPhaseFn,
      MigratePost: migratePostPhaseFn,
      Finalize:    finalizePhaseFn,
  }
  ```
- Command entry points:
  ```go
  func runDeploy(cmd *cobra.Command, args []string) error {
      return runDeployWithPhases(cmd, args, defaultPhaseFns)
  }

  func runRollback(cmd *cobra.Command, args []string) error {
      return runRollbackWithPhases(cmd, args, defaultPhaseFns)
  }
  ```
- Tests construct custom `PhaseFns` to simulate success/failure without mutating globals:
  ```go
  fns := PhaseFns{
      Build:       defaultPhaseFns.Build,
      Push:        defaultPhaseFns.Push,
      MigratePre:  defaultPhaseFns.MigratePre,
      Rollout: func(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
          return fmt.Errorf("forced rollout failure")
      },
      MigratePost: defaultPhaseFns.MigratePost,
      Finalize:    defaultPhaseFns.Finalize,
  }
  ```

### Test Helpers

CLI tests SHOULD use helpers that:

- Construct a fresh root command
- Wire a command's `RunE` to the DI entry (`runDeployWithPhases` / `runRollbackWithPhases`)
- Use a temp directory and explicitly control:
  - `--config` path
  - State file location (via `state.NewManager(path)` in tests when verifying state)

### Parallel Safety

This spec encourages a design where:

- No tests mutate global function pointers
- Command behavior depends only on:
  - Explicit DI (`PhaseFns`)
  - Explicit state manager paths
  - Explicit configuration paths

This is a prerequisite for safely running tests in parallel in the future.

---

## Implementation Requirements

### Required Helpers (Suggested)

Implement (or adapt existing) helpers with the following semantics:

```go
// Returns all phases in canonical order.
func allPhasesCommon() []state.ReleasePhase

// Convenience wrapper (same as allPhasesCommon today).
func orderedPhasesCommon() []state.ReleasePhase

// Marks all downstream phases as skipped after a failure.
func markDownstreamPhasesSkippedCommon(
    ctx context.Context,
    stateMgr *state.Manager,
    releaseID string,
    failedPhase state.ReleasePhase,
    logger logging.Logger,
) error

// Marks all phases as failed (used when planning fails before any execution).
func markAllPhasesFailedCommon(
    ctx context.Context,
    stateMgr *state.Manager,
    releaseID string,
    logger logging.Logger,
) error

// Shared phase execution entry.
func executePhasesCommon(
    ctx context.Context,
    stateMgr *state.Manager,
    releaseID string,
    plan *core.Plan,
    logger logging.Logger,
    fns PhaseFns,
) error
```

Names are indicative – the implementation may keep deploy/rollback-specific wrappers (e.g. `executePhasesDeploy`, `executePhasesRollback`) that internally delegate to a common helper.

---

## Testing

Tests MUST cover:

1. **Happy-path execution**:
   - All phases succeed
   - Each phase transitions: pending → running → completed
   - Final state: all phases completed

2. **Single phase failure** (for each phase individually):
   - Failing phase transitions: pending/running → failed
   - All downstream phases become skipped
   - All upstream phases remain completed
   - Error is returned to the caller

3. **Planner failure** (plan generation error):
   - No phases executed by shared helper
   - Command-specific behavior for marking phases as failed (as per deploy/rollback spec)
   - Error returned to caller

4. **State update failure**:
   - If `UpdatePhase` fails when setting running or completed, execution aborts immediately
   - No further phases are executed
   - Error returned to caller

5. **Rollback-specific invariants** (via `CLI_ROLLBACK` tests):
   - Rollback uses target version + commit SHA
   - Dry-run does not create a release
   - Non-dry-run uses shared phase execution semantics

6. **Deploy-specific invariants** (via `CLI_DEPLOY` tests):
   - Existing deploy semantics (including dry-run) remain unchanged externally
   - Internal tests no longer mutate global phase function variables

---

## Non-Goals (v1)

- Introduce new phases or statuses
- Implement partial deployments or partial rollbacks (subset of services)
- Add cross-phase dependency graphs (beyond simple linear order)
- Implement automatic retries or backoff strategies
- Implement cross-environment rollbacks (e.g. staging → prod)

These may be considered in future specs once core semantics are stable.


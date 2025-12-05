---

## 📋 NEXT AGENT CONTEXT — After Completing Features

CLI_ROLLBACK, CLI_RELEASES, CORE_STATE_CONSISTENCY

---

## 🎉 LAYER 1: What Just Happened

### Features Complete

**Feature IDs**:
- `CLI_ROLLBACK`
- `CLI_RELEASES`
- `CORE_STATE_CONSISTENCY`

**Status**: ✅ Fully implemented, fully tested, and merged

**Relevant Pull Requests**:
- CLI_ROLLBACK: #<PR_NUMBER>
- CLI_RELEASES: #<PR_NUMBER>
- CORE_STATE_CONSISTENCY: #<PR_NUMBER>

(Replace with actual PR numbers if you want automatic linking)

---

### 🧩 What Now Exists

**1. Robust rollback pipeline**:
- Full rollback command implementation
- Rollback target selection (explicit or previous)
- Integration with shared phase execution (`executePhasesCommon`)
- Clean interaction with state manager
- Deterministic release creation and update

**2. Releases inspection commands**:
- `stagecraft releases list`
- `stagecraft releases show`
- Deterministic sorting
- Rich formatting with status field
- Tests covering happy path, empty state, invalid IDs

**3. State consistency guarantees (CORE_STATE_CONSISTENCY)**:
- Atomic writes with file + directory sync
- Read-after-write tests enforce consistency
- No more intermittent stale reads
- Validated in multi-release scenarios
- Deterministic merge of writes during phase updates

**APIs Available**:

```go
// From rollback implementation:
func runRollbackWithPhases(cmd *cobra.Command, args []string, fns PhaseFns) error
func executePhasesCommon(ctx context.Context, stateMgr *state.Manager, releaseID string, plan *core.Plan, logger logging.Logger, fns PhaseFns) error

// From state consistency:
func (m *Manager) CreateRelease(ctx context.Context, env, version, commitSHA string) (*Release, error)
func (m *Manager) ListReleases(ctx context.Context, env string) ([]*Release, error)
func (m *Manager) UpdatePhase(ctx context.Context, releaseID string, phase ReleasePhase, status PhaseStatus) error
func (m *Manager) GetRelease(ctx context.Context, id string) (*Release, error)
func (m *Manager) GetCurrentRelease(ctx context.Context, env string) (*Release, error)
```

**Files Created or Updated**:
- `internal/cli/commands/rollback.go`
- `internal/cli/commands/releases.go`
- `internal/core/state/state.go`
- `internal/core/state/state_test.go`
- `spec/core/state-consistency.md`
- `spec/commands/rollback.md`
- `spec/commands/releases.md`
- Testdata files for releases output

---

## 🎯 LAYER 2: Immediate Next Task

### 🚀 Implement Feature: CLI_DEPLOY

**Feature ID**: `CLI_DEPLOY`

**Status**: `todo` (partial implementation exists)

**Spec**: `spec/commands/deploy.md`

### What Exists Already

- Command wiring under `internal/cli/commands/deploy.go`
- Phase execution pipeline (`executePhasesCommon`)
- State integration + release creation
- Version resolution logic
- Dry-run support
- Tests defined (`deploy_test.go`)
- Phase structure defined:
  - `executeBuildPhase`
  - `executePushPhase`
  - `executeMigratePrePhase`
  - `executeRolloutPhase`
  - `executeMigratePostPhase`
  - `executeFinalizePhase`

### What Is Missing (Critical)

**All phase functions are stubs.**

You must implement:

| Phase | Purpose | Implementation Required |
|-------|---------|------------------------|
| Build | Build backend images | Use `BackendProvider.BuildDocker` |
| Push | Push images to registry | Implement simple push via `executil` |
| MigratePre | DB migrations before deploy | Integrate with `MIGRATION_ENGINE_RAW` or stub |
| Rollout | Deploy to hosts | Minimal `compose up/down` OR docker-rollout |
| MigratePost | Post-deploy migrations | Stub allowed initially |
| Finalize | Mark successful release | Write final state update |

---

### 🧪 MANDATORY WORKFLOW — Tests First

**Before writing any implementation code**, the next agent MUST:

1. **Open**: `internal/cli/commands/deploy_test.go`

2. **Identify unimplemented tests** (they already describe required behavior)

3. **Write missing tests** that describe:
   - Expected build phase behavior
   - Expected push behavior
   - Expected rollout behavior
   - Migration hooks calls (pre/post)
   - State transition expectations
   - Dry-run behavior

4. **Run tests** — they MUST FAIL

5. **Then implement phase functions** until tests PASS

---

### 🛠 Implementation Outline

**1. Build Phase**:

Use `BackendProvider`:
```go
imgTag, err := backend.BuildDocker(ctx, BuildOptions{...})
```

**2. Push Phase**:

Initially use:
```go
executil.Run("docker", "push", imgTag)
```

Later phases may abstract this.

**3. Rollout Phase (Minimal Viable Implementation)**:

MVP version:
```go
docker compose -f generated.yml up -d
```

Later:
```go
docker-rollout up -f generated.yml
```

**4. Compose Generation**:

This is its own Feature ID: `DEPLOY_COMPOSE_GEN`

But a minimal version can inline this logic:
1. Load `docker-compose.yml`
2. Filter services by host role
3. Apply image tag overrides
4. Render into temp directory
5. Upload or use locally for dev env

**5. Migration Hooks**:

For now:
```go
// TODO: integrate MIGRATION_ENGINE_RAW when available
return nil
```

**6. Finalize**:

Mark release as successful in state:
```go
mgr.UpdatePhase(release.ID, "finalize", state.StatusCompleted)
```

---

### 🧭 CONSTRAINTS (Canonical List)

**The next agent MUST NOT**:
- ❌ Modify rollback or releases code
- ❌ Modify state file format
- ❌ Change release ID format
- ❌ Add new providers or engines
- ❌ Add randomness, timestamps, or non-determinism
- ❌ Hardcode provider IDs (must stay registry-driven)

**The next agent MUST**:
- ✅ Implement only `CLI_DEPLOY`
- ✅ Start with failing tests
- ✅ Follow existing CLI command structure
- ✅ Use `executePhasesCommon`
- ✅ Follow provider boundaries
- ✅ Update `spec/features.yaml` only when feature is DONE
- ✅ Use deterministic ordering of hosts/services
- ✅ Keep logs structured and stable

---

## 📌 LAYER 3: Secondary Tasks

### DEPLOY_COMPOSE_GEN

**Feature ID**: `DEPLOY_COMPOSE_GEN`

**Status**: `todo`

**Required for full production deploy.**

**Not required for MVP deploy test suite.**

### DEPLOY_ROLLOUT

**Feature ID**: `DEPLOY_ROLLOUT`

**Status**: `todo`

**Integrates docker-rollout.**

**Can be implemented after MVP deploy.**

### MIGRATION_PRE_DEPLOY / MIGRATION_POST_DEPLOY

**Feature IDs**:
- `MIGRATION_PRE_DEPLOY`
- `MIGRATION_POST_DEPLOY`

**Status**: `todo`

**Hook into deploy phases.**

---

## 🎓 Architectural Context (For Understanding)

### Why CLI_DEPLOY Is Critical

- All later phases depend on a working deploy pipeline:
  - Infra
  - CI
  - Migrations
  - Operations
  - Rollback correctness
  - Releases visibility

### Why Now Is the Ideal Time

- Rollback and releases are finished → deploy has stable downstream consumers
- State consistency is guaranteed → deploy writes are now deterministic
- Phase execution engine is complete → deploy only needs phase bodies implemented

---

## 📝 Expected Output When CLI_DEPLOY Is Complete

### 1. Summary

Clear description of implemented behavior.

### 2. Commit Message

```
feat(CLI_DEPLOY): implement deploy phase functions

Summary:
- Implement build, push, rollout, finalize phases
- Add per-host Compose generation (MVP)
- Integrate backend provider build
- Add rollout execution via compose
- Update deploy tests
- Update spec/commands/deploy.md

Files:
- internal/cli/commands/deploy.go
- internal/cli/commands/deploy_test.go
- spec/commands/deploy.md

Feature: CLI_DEPLOY
Spec: spec/commands/deploy.md
```

### 3. Verification Checklist

- ✅ All tests pass
- ✅ Coverage meets thresholds
- ✅ No lint errors
- ✅ No unrelated diffs
- ✅ Feature state set to `done` only after completion

---

## ⚡ Quick Start for Next Agent

1. **Load context from**:
   - `internal/cli/commands/deploy.go`
   - `spec/commands/deploy.md`
   - `deploy_test.go`

2. **Create feature branch**:
   ```bash
   git checkout -b feature/CLI_DEPLOY-deploy-phase-implementation
   ```

3. **Write failing tests first**.

4. **Implement phase functions in smallest increments**.

5. **Run**:
   ```bash
   ./scripts/run-all-checks.sh
   ```

6. **Produce summary + commit**.

---

## ✅ Final Checklist for Next Agent

- [ ] Feature ID: `CLI_DEPLOY`
- [ ] Branch created correctly
- [ ] Tests written first
- [ ] Tests fail before implementation
- [ ] Phase functions implemented incrementally
- [ ] All CI checks pass
- [ ] `spec/features.yaml` updated to `done`
- [ ] PR ready

---

## 📋 APPENDIX: Test-First Development Outline

Below is a test plan you can use to drive the implementation. Some of these probably partially exist already - the idea is to:

- Fill gaps.
- Tighten semantics to match the spec.
- Make the phase functions safe to implement incrementally.

You can adapt names to whatever convention you already use in `deploy_test.go`.

### 2.1 Core happy path tests

#### 1. `TestDeployCommand_Success_AllPhasesRunInOrder`

**Goal**: Ensure that on a successful deploy, all phases run exactly once, in the expected order: `build -> push -> migrate_pre -> rollout -> migrate_post -> finalize`

**Approach**:
- Use a fake phase executor (or injected hooks) that append to a slice when called.
- Use a fake backend provider that always succeeds.
- Ensure state is written with all phases marked `success` (except maybe migrate phases if you treat them as skipped in v1).

**Assertions**:
- Phase call order equals the declared phase list.
- State store shows a new release with all relevant phases in `success` (or `skipped` for migrate phases if that is your v1 behavior).
- Command returns no error.

#### 2. `TestDeployCommand_DryRun_DoesNotModifyStateOrCallExternalCommands`

**Goal**: Enforce dry-run semantics: no side effects.

**Approach**:
- Use a fake backend provider that records invocations but should not be called in `--dry-run` mode.
- Make the test use the isolated state helper (`setupIsolatedStateTestEnv`).
- Run the command with `--dry-run`.

**Assertions**:
- Backend provider build/push methods were not called.
- No external exec calls (if you can inject a fake executil).
- No new releases written to the state file (state file either unchanged or not created).
- CLI output includes or logs the planned phases.

### 2.2 Failure semantics tests

#### 3. `TestDeployCommand_BuildPhaseFailure_SkipsDownstreamPhases`

**Goal**: Verify that a failure in `build`:
- Marks `build` as `failed`
- Marks subsequent phases as `skipped`
- Returns a non-zero error

**Approach**:
- Fake backend provider that fails build with a deterministic error.
- Phase hooks or counters for each phase.

**Assertions**:
- `build` called and fails.
- No attempt to run `push`, `rollout`, or `finalize`.
- State:
  - `build` = failed
  - `push`, `migrate_pre`, `rollout`, `migrate_post`, `finalize` = skipped (with consistent reason or at least a stable status).
- Command returns error.

#### 4. `TestDeployCommand_PushPhaseFailure_SkipsRolloutAndLaterPhases`

**Goal**: Same structure as build failure, but fail in `push`.

**Assertions**:
- `build` = success
- `push` = failed
- `rollout` and later phases = skipped
- No rollout or finalize side effects executed.

#### 5. `TestDeployCommand_RolloutPhaseFailure_StopsAndSkipsFinalize`

**Goal**: A rollout failure must not finalize the release as success.

**Assertions**:
- `build` = success
- `push` = success
- `rollout` = failed
- `finalize` = skipped
- State indicates failed overall release.

### 2.3 State interaction and consistency tests

These lean on `CORE_STATE_CONSISTENCY` but make sure deploy uses it correctly.

#### 6. `TestDeployCommand_StateUpdatedAfterSuccessfulDeploy`

**Goal**: After a successful deploy, a subsequent read of state returns the completed release with all expected phase statuses.

**Approach**:
- Use `setupIsolatedStateTestEnv`.
- Run a successful deploy.
- After the command returns, create a new state manager pointing at the same file.
- Call `ListReleases` and inspect the latest release.

**Assertions**:
- New release exists with correct env, version, and phase statuses.
- No stale or partial state.

#### 7. `TestDeployCommand_StateNotWrittenOnDryRun`

**Goal**: Reinforce dry-run semantics for state.

**Assertions**:
- After `--dry-run`, state file does not contain new releases.
- If state file does not exist before, it remains absent.

### 2.4 Version and env resolution tests

#### 8. `TestDeployCommand_InvalidEnv_ReturnsErrorAndDoesNotWriteState`

**Goal**: Deploy must fail fast if `--env` is not defined in config.

**Assertions**:
- Error returned.
- No state file created and no phases executed.

#### 9. `TestDeployCommand_MissingVersion_UsesPlanResolution`

**Goal**: When `--version` is not provided, deploy must use whichever resolution strategy `CORE_PLAN` specifies.

Because that strategy can change, keep this test coarse:

**Approach**:
- Configure test so that plan resolution returns a deterministic version string (for example via a fake planner, or environment variable).
- Call deploy without `--version`.

**Assertions**:
- State contains a release with the expected resolved version.
- No direct call to Git or CI-specific logic from CLI (those live in `CORE_PLAN` or helpers).

### 2.5 Migration placeholders

For v1, migrations are no-ops, but you can still capture behaviour.

#### 10. `TestDeployCommand_MigratePhasesExistInState`

**Goal**: Ensure the `migrate_pre` and `migrate_post` phases appear in the release, even if they are not implemented yet.

**Assertions**:
- Release phases list includes both `migrate_pre` and `migrate_post`.
- Status is either:
  - `pending` (if you choose to keep them as not-run), or
  - `skipped` with a deterministic reason.

Document the chosen semantics in the spec and tests.

### 2.6 Provider integration tests (fake provider)

#### 11. `TestDeployCommand_UsesBackendProviderBuild`

**Goal**: Ensure deploy actually calls into the configured backend provider for builds.

**Approach**:
- Register a fake backend provider with:
  - `ID() string`
  - `BuildDocker` that records args and returns a deterministic image tag.
- Configure `stagecraft.yml` for the test to use that provider ID.

**Assertions**:
- Fake provider's `BuildDocker` is called with correct options:
  - Workdir
  - Image tag prefix or full tag
- State and logs reflect that tag.

### 2.7 Golden output tests (optional but consistent with Stagecraft)

If you have golden tests for CLI output, add:

#### 12. `TestDeployCommand_Output_Golden`

**Goal**: Keep CLI output stable and deterministic.

**Approach**:
- Capture CLI output for:
  - Successful deploy
  - Failure in early phase
  - Dry-run
- Store under `testdata/deploy_*.golden`.

**Assertions**:
- Output matches golden files.
- Golden updates only when spec changes explicitly.

---

**Copy this entire document into your next agent session to continue development.**

This document is optimized for deterministic AI handoff and aligns with Stagecraft's Agent.md principles (spec-first, test-first, feature-bounded, deterministic).


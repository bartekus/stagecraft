> **Superseded by** `docs/context-handoff/CONTEXT_LOG.md`. Kept for historical reference. New context handoffs MUST be added to the context log.

---

## 📋 NEXT AGENT CONTEXT — After Completing Feature CLI_ROLLBACK

---

## 🎉 LAYER 1: What Just Happened

### Feature Complete: CLI_ROLLBACK

**Feature ID**: `CLI_ROLLBACK`

**Status**: ✅ Implemented, fully tested, and merged

**PR**: #<PR_NUMBER> (<PR_URL>)

**Commit**: `<COMMIT_HASH>` - `feat(CLI_ROLLBACK): add rollback command with phase tracking`

### What Now Exists

**Package**: `internal/cli/commands/` and `internal/core/`

- Rollback command (`stagecraft rollback`) with phase-aware behaviour

- Shared phase execution logic (`executePhasesCommon`) for deploy/rollback

- Deterministic state updates via `CORE_STATE` (phase statuses, skips, failures)

- Golden tests for rollback output and state transitions

- Provider-agnostic rollback implementation (no hardcoded provider IDs)

- Rollback tests integrated with isolated state test env helpers

**APIs Available**:

```go
// internal/core/state/...
type Manager interface {
    // Release and phase operations used by deploy/rollback.
}

// internal/cli/commands/... (phases_common.go)
type PhaseFns struct {
    // Provider callbacks and side-effect hooks.
}
func ExecutePhasesCommon(...) error
```

**Files Created**:

- `internal/cli/commands/rollback.go`
- `internal/cli/commands/rollback_test.go`
- `internal/cli/commands/phases_common.go` (shared execution semantics)

**Files Updated**:

- `spec/commands/rollback.md`
- `spec/features.yaml` — Marked CLI_ROLLBACK as done

---

## 🎯 LAYER 2: Immediate Next Task

### Implement CLI_BUILD

**Feature ID**: `CLI_BUILD`

**Status**: `todo`

**Spec**: `spec/commands/build.md` (created and populated)

**Dependencies**:

- CLI_DEPLOY <status: ready>
- CORE_PLAN <status: ready>
- CORE_STATE <status: ready>
- PROVIDER_BACKEND_GENERIC <status: ready>
- PROVIDER_BACKEND_ENCORE_TS <status: ready>

**⚠️ SCOPE REMINDER**: All work in this handoff MUST be scoped strictly to CLI_BUILD.

Do not modify deploy/rollback behaviour, provider registration, or unrelated features.

**Reference Spec**: `spec/commands/build.md`

---

## 🧪 MANDATORY WORKFLOW — Tests First

Before writing ANY implementation code for CLI_BUILD:

1. **Create test file**: `internal/cli/commands/build_test.go`

2. **Write failing tests** describing:

   - Missing `--env` is a user error with deterministic message
   - Invalid env yields `invalid environment: <env>`
   - Dry-run prints build plan (env, services, version, provider)
   - `--services` filters the build to a subset of services
   - `--version` is reflected in planning/output
   - Provider/plan errors propagate as deterministic failures

3. **Run tests** - they MUST fail

4. **Only then** begin implementation

**Test Pattern** (follow existing test patterns):

- Use existing CLI helpers from `internal/cli/commands/deploy_test.go` / `internal/cli/commands/rollback_test.go`
- Prefer golden tests for CLI output under `internal/cli/commands/testdata/`
- Mock external/provider behaviour through existing registries and test fixtures
- Assert deterministic error messages and exit behaviour

---

## 🛠 Implementation Outline

**1. Standard Initialization Pattern**:

```go
// internal/cli/commands/build.go (wiring only; no business logic in cmd)
func NewBuildCommand() *cobra.Command {
    // Parse flags: --env, --version, --push, --dry-run, --services
    // Resolve env + config with existing helpers
    // Construct core.BuildOptions
    // Call core.ExecuteBuild(ctx, opts)
}
```

**2. Main Behaviour / Phase Sequence**:

1. Load config + environment
2. Generate full plan via CORE_PLAN
3. Filter to build phases
4. Apply `--services` filtering
5. Resolve version (`--version` or deterministic release ID)
6. If `--dry-run`:
   - Render plan summary
   - Exit 0
7. Else:
   - Execute build phases via `executePhasesCommon`
   - If `--push`, push after build
   - Map failures to deterministic exit codes

**3. Failure Semantics**:

- On failure in a build phase → mark that phase as failed
- Abort any remaining build phases
- Deploy/rollback phases MUST NEVER run from CLI_BUILD
- Return exit codes as defined in `spec/commands/build.md`

**4. Required Files**:

- `internal/cli/commands/build.go`
- `internal/cli/commands/build_test.go`
- `spec/commands/build.md` (already created; may require minor updates)
- `internal/core/phases_build.go` (shared build execution entry point)

**5. Integration Points**:

- Uses CORE_PLAN for orchestration plan generation
- Uses CORE_STATE and release ID generator for default version/tag
- Uses shared phase execution logic (`PhaseFns` / `ExecutePhasesCommon`)
- Integrates with backend provider registry (no hardcoded provider IDs)

---

## 🧭 CONSTRAINTS (Canonical List)

**The next agent MUST NOT**:

- ❌ Modify existing CLI_DEPLOY or CLI_ROLLBACK semantics
- ❌ Change persisted state formats
- ❌ Add/rename/remove phase identifiers
- ❌ Implement CLI_PLAN early
- ❌ Mix multiple features in one PR
- ❌ Skip tests-first workflow
- ❌ Introduce provider-specific conditionals into core

**The next agent MUST**:

- ✅ Write failing tests first (`internal/cli/commands/build_test.go`)
- ✅ Follow existing CLI patterns in `internal/cli/commands/`
- ✅ Use `internal/core` abstractions (plan, state, phase execution) correctly
- ✅ Keep changes strictly scoped to CLI_BUILD
- ✅ Update `spec/commands/build.md` and `spec/features.yaml` when complete

---

## 📌 LAYER 3: Secondary Tasks

### CLI_PLAN

**Feature ID**: `CLI_PLAN`

**Status**: `todo`

**Dependencies**: CORE_PLAN, CLI_DEPLOY

**Do NOT begin until CLI_BUILD is complete.** (See CONSTRAINTS section)

---

### DEPLOY_COMPOSE_GEN (Design Only)

**Feature ID**: `DEPLOY_COMPOSE_GEN`

**Status**: `todo`

**Dependencies**: CORE_PLAN, PROVIDER_BACKEND_GENERIC, PROVIDER_NETWORK_TAILSCALE

**Do NOT implement until prerequisites are complete.** (See CONSTRAINTS section)

Design can consider:

- Multi-host Compose topology generation
- Network provider (Tailscale) integration points
- Idempotent rollout planning and validation

---

## 🎓 Architectural Context (For Understanding)

**Why These Design Decisions Matter**:

- **Invariant 1**: Build semantics are shared between deploy and build.
  - This avoids drift between CLI_DEPLOY and CLI_BUILD.
- **Invariant 2**: Core remains provider-agnostic.
  - Backend build details live in providers, not in core.
- **Invariant 3**: Deterministic behaviour and output.
  - CI and humans must see stable plans, exit codes, and logs.

**Integration Pattern Example** (for reference, not required to copy exactly):

```go
// Example: How CLI_BUILD should integrate with CORE_PLAN and shared phase execution.
func ExecuteBuild(ctx context.Context, opts BuildOptions) error {
    plan, err := planner.BuildPlan(ctx, opts.Env)
    if err != nil {
        return fmt.Errorf("build: plan generation failed: %w", err)
    }
    buildPhases := plan.BuildPhases(opts.Services)
    if opts.DryRun {
        return renderBuildPlan(buildPhases, opts)
    }
    return ExecutePhasesCommon(ctx, buildPhases, phaseFns)
}
```

---

## 📝 Output Expectations

**When you complete CLI_BUILD**:

1. **Summary**: What was implemented

2. **Commit Message** (follow this format):

```
feat(CLI_BUILD): add standalone build command

Summary:
- Add stagecraft build CLI command
- Share build execution path between deploy and build
- Support dry-run, service filtering, and explicit version

Files:
- internal/cli/commands/build.go
- internal/cli/commands/build_test.go
- internal/core/phases_build.go
- spec/commands/build.md
- spec/features.yaml

Test Results:
- ./scripts/run-all-checks.sh
- go test ./...
- Golden tests updated

Feature: CLI_BUILD
Spec: spec/commands/build.md
```

3. **Verification**:

   - ✅ Tests were written first
   - ✅ No unrelated changes were made
   - ✅ Feature boundaries respected (only CLI_BUILD code)
   - ✅ All checks pass (`./scripts/run-all-checks.sh`)

---

## ⚡ Quick Start for Next Agent

**Bootloader Instructions**:

1. **Load Context**:

   - Read `spec/commands/build.md` for behaviour and exit codes
   - Read `internal/cli/commands/deploy.go` and `internal/cli/commands/rollback.go` to mirror CLI patterns
   - Read `internal/cli/commands/phases_common.go` and CORE_STATE for phase/state semantics
   - Read `internal/cli/commands/deploy_test.go` for test structure and helpers

2. **Begin Work**:

   - Feature ID: CLI_BUILD
   - Create feature branch: `feature/CLI_BUILD-build-command`
   - Start with tests: `internal/cli/commands/build_test.go`
   - Write failing tests first
   - Then implement: `internal/cli/commands/build.go`, `internal/core/phases_build.go`

3. **Follow Semantics**:

   - Use shared phase execution engine
   - Only build phases (no deploy/rollback)
   - Deterministic version, output, and exit codes

4. **Respect Constraints**:

   - See CONSTRAINTS section (canonical list)
   - Do not modify CLI_DEPLOY or CLI_ROLLBACK semantics
   - Do not implement CLI_PLAN early
   - Keep feature boundaries clean

---

## ✅ Final Checklist

Before starting work:

- [ ] Feature ID identified: CLI_BUILD
- [ ] Git hooks verified
- [ ] Working directory clean
- [ ] On feature branch: `feature/CLI_BUILD-build-command`
- [ ] Spec located: `spec/commands/build.md`
- [ ] Tests written first: `internal/cli/commands/build_test.go`
- [ ] Tests fail (as expected)
- [ ] Ready to implement

---


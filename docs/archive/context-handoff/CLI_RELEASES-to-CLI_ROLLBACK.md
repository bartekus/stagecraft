> **Superseded by** `docs/context-handoff/CONTEXT_LOG.md`. Kept for historical reference. New context handoffs MUST be added to the context log.

---
## 📋 NEXT AGENT CONTEXT — After Completing Feature CLI_RELEASES
---

## 🎉 LAYER 1: What Just Happened

### Feature Complete: CLI_RELEASES

**Feature ID**: `CLI_RELEASES`

**Status**: ✅ Implemented, fully tested, and merged

**What Now Exists**:

**Package**: `internal/cli/commands/`

- Releases command with `list` and `show` subcommands
- Release listing with environment filtering
- Detailed release display with phase status
- Overall status calculation (pending/running/completed/failed)
- Integration with `CORE_STATE` for release data access
- Comprehensive test coverage

**APIs Available**:

```go
// From internal/cli/commands/releases.go
NewReleasesCommand() *cobra.Command
NewReleasesListCommand() *cobra.Command
NewReleasesShowCommand() *cobra.Command

// Helper functions
calculateOverallStatus(release *state.Release) string
displayReleasesList(cmd *cobra.Command, releases []*state.Release, showEnv bool) error
displayReleaseShow(cmd *cobra.Command, release *state.Release) error
```

**Files Created**:

- `internal/cli/commands/releases.go` (264 lines)
- `internal/cli/commands/releases_test.go`
- `spec/commands/releases.md` (323 lines)

**Files Updated**:

- `internal/cli/root.go` - Registered releases command
- `spec/features.yaml` - CLI_RELEASES status updated to `done`

---

## 🎯 LAYER 2: Immediate Next Task

### Implement CLI_ROLLBACK

**Feature ID**: `CLI_ROLLBACK`

**Status**: `todo`

**Spec**: `spec/commands/rollback.md` (exists; may need to be updated if missing)

**Dependencies**:

- ✅ `CORE_STATE` (ready)
- ✅ `CLI_DEPLOY` (ready)
- ✅ `CLI_RELEASES` (ready)

**⚠️ SCOPE REMINDER**: All work in this handoff MUST be scoped strictly to `CLI_ROLLBACK`. Do not modify `CLI_DEPLOY`, `CLI_RELEASES`, or `CORE_STATE`.

---

### 🧪 MANDATORY WORKFLOW — Tests First

**Before writing ANY implementation code**:

1. **Create test file**: `internal/cli/commands/rollback_test.go`

2. **Write failing tests** describing:

   - `rollback --to-previous` with valid previous release
   - `rollback --to-previous` with no previous release (error)
   - `rollback --to-release=<id>` with valid release ID
   - `rollback --to-release=<id>` with invalid release ID (error)
   - `rollback --to-release=<id>` with environment mismatch (error)
   - `rollback --to-version=<v>` with matching version
   - `rollback --to-version=<v>` with no matching version (error)
   - Target validation: cannot rollback to current release (error)
   - Target validation: target must be fully deployed (error)
   - Rollback creates new release with target's version/commit SHA
   - Rollback executes deployment phases
   - Rollback handles phase failures correctly
   - **Dry-run mode does NOT create a release and does NOT execute phases**
   - Multiple target flags (error)
   - No target flags (error)

3. **Run tests** - they MUST fail

4. **Only then** begin implementation

**Test Pattern** (follow existing CLI test patterns):

- Use `internal/cli/commands/deploy_test.go` as reference
- Use `internal/cli/commands/releases_test.go` for state integration patterns
- Create test releases using `state.Manager.CreateRelease()` in test setup
- Use `state.Manager.UpdatePhase()` to set phase statuses for validation tests
- Test all three target resolution methods
- Test all validation rules

---

### 🛠 Implementation Outline

**1. Command Structure**:

```go
// Feature: CLI_ROLLBACK
// Spec: spec/commands/rollback.md

func NewRollbackCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "rollback",
        Short: "Rollback environment to a previous release",
        Long:  "Rolls back an environment to a previous release by creating a new deployment",
        RunE:  runRollback,
    }
    
    cmd.Flags().Bool("to-previous", false, "Rollback to immediately previous release")
    cmd.Flags().String("to-release", "", "Rollback to specific release ID")
    cmd.Flags().String("to-version", "", "Rollback to most recent release with matching version")
    
    return cmd
}
```

**2. Flag Parsing Helper**:

```go
type rollbackFlags struct {
    ToPrevious bool
    ToRelease  string
    ToVersion  string
}

func parseRollbackFlags(cmd *cobra.Command) (rollbackFlags, error) {
    toPrevious, _ := cmd.Flags().GetBool("to-previous")
    toRelease, _ := cmd.Flags().GetString("to-release")
    toVersion, _ := cmd.Flags().GetString("to-version")
    
    count := 0
    if toPrevious {
        count++
    }
    if toRelease != "" {
        count++
    }
    if toVersion != "" {
        count++
    }
    
    if count == 0 {
        return rollbackFlags{}, fmt.Errorf("rollback target required; use --to-previous, --to-release, or --to-version")
    }
    
    if count > 1 {
        return rollbackFlags{}, fmt.Errorf("only one rollback target flag may be specified")
    }
    
    return rollbackFlags{
        ToPrevious: toPrevious,
        ToRelease:  toRelease,
        ToVersion:  toVersion,
    }, nil
}
```

**3. Target Resolution** (accepts current to avoid double-fetch):

```go
// Resolve rollback target based on flags.
// Accepts current release to avoid redundant GetCurrentRelease call.
func resolveRollbackTarget(ctx context.Context, stateMgr *state.Manager, env string, current *state.Release, flags rollbackFlags) (*state.Release, error) {
    // Determine which flag was set
    if flags.ToPrevious {
        if current.PreviousID == "" {
            return nil, fmt.Errorf("no previous release to rollback to")
        }
        target, err := stateMgr.GetRelease(ctx, current.PreviousID)
        if err != nil {
            return nil, fmt.Errorf("rollback target not found: %q", current.PreviousID)
        }
        return target, nil
    }
    
    if flags.ToRelease != "" {
        target, err := stateMgr.GetRelease(ctx, flags.ToRelease)
        if err != nil {
            return nil, fmt.Errorf("rollback target not found: %q", flags.ToRelease)
        }
        // Validate environment match
        if target.Environment != env {
            return nil, fmt.Errorf("release %q belongs to environment %q, not %q", flags.ToRelease, target.Environment, env)
        }
        return target, nil
    }
    
    if flags.ToVersion != "" {
        releases, err := stateMgr.ListReleases(ctx, env)
        if err != nil {
            return nil, fmt.Errorf("listing releases: %w", err)
        }
        // Find most recent matching version (already sorted newest first)
        for _, r := range releases {
            if r.Version == flags.ToVersion {
                return r, nil
            }
        }
        return nil, fmt.Errorf("no release found with version %q in environment %q", flags.ToVersion, env)
    }
    
    // This should not happen if parseRollbackFlags was called correctly
    return nil, fmt.Errorf("rollback target required; use --to-previous, --to-release, or --to-version")
}
```

**4. Target Validation**:

```go
// Validate target release is eligible for rollback
func validateRollbackTarget(current, target *state.Release) error {
    // Cannot rollback to current
    if current.ID == target.ID {
        return fmt.Errorf("cannot rollback to current release %q", target.ID)
    }
    
    // Must be fully deployed
    requiredPhases := []state.ReleasePhase{
        state.PhaseBuild,
        state.PhasePush,
        state.PhaseMigratePre,
        state.PhaseRollout,
        state.PhaseMigratePost,
        state.PhaseFinalize,
    }
    
    incompletePhases := []string{}
    for _, phase := range requiredPhases {
        if target.Phases[phase] != state.StatusCompleted {
            incompletePhases = append(incompletePhases, string(phase))
        }
    }
    
    if len(incompletePhases) > 0 {
        return fmt.Errorf("rollback target %q is not fully deployed (phases: %v)", target.ID, incompletePhases)
    }
    
    return nil
}
```

**5. Rollback Execution** (optimized to avoid double GetCurrentRelease):

```go
func runRollback(cmd *cobra.Command, args []string) error {
    ctx := cmd.Context()
    
    // Resolve flags
    flags, err := ResolveFlags(cmd, nil)
    if err != nil {
        return fmt.Errorf("resolving flags: %w", err)
    }
    
    // Load config
    cfg, err := config.Load(flags.Config)
    if err != nil {
        return fmt.Errorf("loading config: %w", err)
    }
    
    // Re-resolve flags with config
    flags, err = ResolveFlags(cmd, cfg)
    if err != nil {
        return fmt.Errorf("resolving flags: %w", err)
    }
    
    // Validate environment
    if flags.Env == "" {
        return fmt.Errorf("environment is required; use --env flag")
    }
    
    // Initialize state manager
    stateMgr := state.NewDefaultManager()
    
    // Get current release (single fetch, passed to resolveRollbackTarget)
    current, err := stateMgr.GetCurrentRelease(ctx, flags.Env)
    if err != nil {
        return fmt.Errorf("no current release found for environment %q", flags.Env)
    }
    
    // Parse rollback flags
    rollbackFlags, err := parseRollbackFlags(cmd)
    if err != nil {
        return err
    }
    
    // Resolve rollback target (passes current to avoid double-fetch)
    target, err := resolveRollbackTarget(ctx, stateMgr, flags.Env, current, rollbackFlags)
    if err != nil {
        return err
    }
    
    // Validate target
    if err := validateRollbackTarget(current, target); err != nil {
        return err
    }
    
    // Initialize logger
    logger := logging.NewLogger(flags.Verbose)
    
    logger.Info("Rolling back environment",
        logging.NewField("env", flags.Env),
        logging.NewField("target_release", target.ID),
        logging.NewField("target_version", target.Version),
    )
    
    // Handle dry-run (BEFORE creating release)
    if flags.DryRun {
        logger.Info("Dry-run mode: would rollback to release",
            logging.NewField("env", flags.Env),
            logging.NewField("target_release", target.ID),
            logging.NewField("target_version", target.Version),
            logging.NewField("target_commit", target.CommitSHA),
        )
        // Optionally generate plan to show what would happen (but don't execute)
        planner := core.NewPlanner(cfg)
        plan, err := planner.PlanDeploy(flags.Env)
        if err == nil {
            logger.Debug("Would execute deployment plan",
                logging.NewField("operations", len(plan.Operations)),
            )
        }
        // Do NOT create a release or write state in dry-run
        return nil
    }
    
    // Create new release with target's version/commit SHA (only in non-dry-run)
    release, err := stateMgr.CreateRelease(ctx, flags.Env, target.Version, target.CommitSHA)
    if err != nil {
        return fmt.Errorf("creating rollback release: %w", err)
    }
    
    logger.Info("Rollback release created",
        logging.NewField("release_id", release.ID),
    )
    
    // Generate deployment plan
    planner := core.NewPlanner(cfg)
    plan, err := planner.PlanDeploy(flags.Env)
    if err != nil {
        markAllPhasesFailed(ctx, stateMgr, release.ID, logger)
        return fmt.Errorf("generating deployment plan: %w", err)
    }
    
    // Execute deployment phases (reuse from deploy.go - see executePhases directive)
    err = executePhases(ctx, stateMgr, release.ID, plan, logger)
    if err != nil {
        return fmt.Errorf("rollback deployment failed: %w", err)
    }
    
    logger.Info("Rollback completed successfully",
        logging.NewField("release_id", release.ID),
    )
    
    return nil
}
```

**6. Required Files**:

- `internal/cli/commands/rollback.go` - Main rollback command
- `internal/cli/commands/rollback_test.go` - Tests (write first!)
- `spec/commands/rollback.md` - Spec (exists; may need to be updated if missing)

**7. Integration Points**:

- Use `state.NewDefaultManager()` to create state manager
- Use `GetCurrentRelease(ctx, env)` to get current release (single call, passed to helpers)
- Use `GetRelease(ctx, id)` for `--to-release` and `--to-previous`
- Use `ListReleases(ctx, env)` for `--to-version`
- Use `CreateRelease(ctx, env, version, commitSHA)` to create rollback release (only in non-dry-run)
- Reuse `executePhases()` from `deploy.go` (see executePhases directive below)
- Use `ResolveFlags()` for flag resolution
- Use `core.NewPlanner(cfg)` for deployment planning

**8. State API Reality Check**:

**IMPORTANT**: Method names in this document (`GetCurrentRelease`, `CreateRelease`, `GetRelease`, `ListReleases`, etc.) reflect the actual API in `internal/core/state/state.go`. Agents MUST confirm actual method signatures by reading `internal/core/state/state.go` and adapt accordingly, rather than introducing new duplicate abstractions.

Verified method names (as of this handoff):
- ✅ `state.NewDefaultManager() *Manager`
- ✅ `GetCurrentRelease(ctx context.Context, env string) (*Release, error)`
- ✅ `GetRelease(ctx context.Context, id string) (*Release, error)`
- ✅ `ListReleases(ctx context.Context, env string) ([]*Release, error)`
- ✅ `CreateRelease(ctx context.Context, env, version, commitSHA string) (*Release, error)`
- ✅ `UpdatePhase(ctx context.Context, releaseID string, phase ReleasePhase, status PhaseStatus) error`
- ✅ Phase constants: `PhaseBuild`, `PhasePush`, `PhaseMigratePre`, `PhaseRollout`, `PhaseMigratePost`, `PhaseFinalize`
- ✅ Status constants: `StatusPending`, `StatusRunning`, `StatusCompleted`, `StatusFailed`, `StatusSkipped`

**9. executePhases Directive**:

The `executePhases` function in `deploy.go` is currently private (lowercase). For v1 of `CLI_ROLLBACK`, it is acceptable to copy the `executePhases` logic from `deploy.go` into `rollback.go` to avoid a premature refactor. Future refactors can consolidate them into a shared helper (e.g., `internal/cli/commands/common_phases.go`). Do NOT change behavior, only duplicate the code.

If you choose to copy, also copy:
- `orderedPhases()` helper function
- `markDownstreamPhasesSkipped()` helper function
- `markAllPhasesFailed()` helper function (if it exists)
- Phase execution function variables (buildPhaseFn, pushPhaseFn, etc.)

---

### 🧭 CONSTRAINTS (Canonical List)

**The next agent MUST NOT**:

- ❌ Modify existing `CLI_DEPLOY` behavior or code (may copy `executePhases` logic, but do not modify original)
- ❌ Modify `CLI_RELEASES` behavior or code
- ❌ Modify `CORE_STATE` behavior or code
- ❌ Change the on-disk state schema (e.g., `.stagecraft/releases.json`) or its JSON layout
- ❌ Add/remove/rename phases or statuses
- ❌ Modify unrelated features
- ❌ Skip the tests-first workflow
- ❌ Write directly to `.stagecraft/releases.json` (always use `state.Manager`)
- ❌ Create a release in dry-run mode (dry-run must not mutate state)
- ❌ Attempt to change `CLI_DEPLOY`'s dry-run behavior to match `CLI_ROLLBACK`. The asymmetry is intentional: rollback dry-run is a pure simulation, while deploy dry-run records intent.

**The next agent MUST**:

- ✅ Write failing tests before implementation
- ✅ Follow existing CLI command patterns (see `internal/cli/commands/deploy.go`)
- ✅ Use `state.Manager` for all state operations
- ✅ Validate rollback targets before execution
- ✅ Reuse deployment phase execution from `CLI_DEPLOY` (by copying logic, not modifying original)
- ✅ Handle errors gracefully with clear messages
- ✅ Create/update spec if missing
- ✅ Keep feature boundaries clean (only `CLI_ROLLBACK` code)
- ✅ Register command in `internal/cli/root.go` in lexicographic order
- ✅ Ensure dry-run mode does NOT create a release or execute phases
- ✅ Avoid redundant `GetCurrentRelease` calls (fetch once, pass to helpers)

---

## 📌 LAYER 3: Secondary Tasks

### Future Enhancements (v2)

- Rollback preview/diff
- Partial rollback (specific services)
- Automatic rollback on failure
- Rollback to releases from different environments

**Do NOT implement these now.** Focus strictly on `CLI_ROLLBACK` v1 scope.

---

## 🎓 Architectural Context (For Understanding)

**Why These Design Decisions Matter**:

- **Rollback is a new deployment**: Rollback creates a new release, not a mutation of old releases. This preserves history.
- **Version copying**: Rollback uses the target release's version/commit SHA, ensuring the exact same code is deployed.
- **Phase reuse**: Rollback reuses the same deployment pipeline as `CLI_DEPLOY`, ensuring consistency.
- **Validation**: Target must be fully deployed to ensure rollback is safe and predictable.
- **Current release check**: Prevents accidental rollback to the same version.
- **Dry-run semantics**: Dry-run does not mutate state, only shows what would happen. This differs from `CLI_DEPLOY`'s dry-run which creates a release. **This asymmetry is intentional**: rollback dry-run is a pure simulation, while deploy dry-run records intent.

**Integration Pattern Example** (for reference):

```go
// Example: How CLI_ROLLBACK should integrate with CORE_STATE and CLI_DEPLOY
stateMgr := state.NewDefaultManager()

// Get current release (single fetch)
current, err := stateMgr.GetCurrentRelease(ctx, "prod")
if err != nil {
    return fmt.Errorf("no current release: %w", err)
}

// Resolve target (example: --to-previous, passes current to avoid double-fetch)
target, err := resolveRollbackTarget(ctx, stateMgr, "prod", current, rollbackFlags)
if err != nil {
    return fmt.Errorf("target not found: %w", err)
}

// Validate target
if err := validateRollbackTarget(current, target); err != nil {
    return err
}

// Create new release with target's version (only if not dry-run)
if !dryRun {
    release, err := stateMgr.CreateRelease(ctx, "prod", target.Version, target.CommitSHA)
    if err != nil {
        return fmt.Errorf("creating release: %w", err)
    }
    
    // Execute deployment (reuse from deploy.go)
    // ... same as deploy ...
}
```

---

## 📝 Output Expectations

**When you complete `CLI_ROLLBACK`**:

1. **Summary**: What was implemented

2. **Commit Message** (follow this format):

```
feat(CLI_ROLLBACK): implement rollback command

Summary:
- Added rollback.go with three target resolution methods
- Implemented target validation (current check, completion check)
- Integrated with CLI_DEPLOY phase execution (copied logic)
- Created rollback_test.go with comprehensive tests
- Created/updated spec/commands/rollback.md

Files:
- internal/cli/commands/rollback.go
- internal/cli/commands/rollback_test.go
- spec/commands/rollback.md
- internal/cli/root.go (command registration)

Test Results:
- All tests pass
- Coverage meets targets
- No linting errors

Feature: CLI_ROLLBACK
Spec: spec/commands/rollback.md
```

3. **Verification**:

   - ✅ Tests were written first (before implementation)
   - ✅ No unrelated changes were made
   - ✅ Feature boundaries respected (only `CLI_ROLLBACK` code)
   - ✅ All checks pass (`./scripts/run-all-checks.sh`)
   - ✅ Dry-run does not create releases or execute phases

---

## ⚡ Quick Start for Next Agent

**Bootloader Instructions**:

1. **Load Context**:

   - Read `internal/core/state/state.go` to understand API
   - Read `spec/core/state.md` for state semantics
   - Read `internal/cli/commands/deploy.go` to understand phase execution
   - Read `internal/cli/commands/releases.go` for state integration patterns
   - Read `internal/cli/commands/deploy_test.go` for test patterns
   - Check if `spec/commands/rollback.md` exists (create if missing)

2. **Begin Work**:

   - Feature ID: `CLI_ROLLBACK`
   - Create feature branch: `feature/CLI_ROLLBACK`
   - Start with tests: `internal/cli/commands/rollback_test.go`
   - Write failing tests first
   - Then implement: `internal/cli/commands/rollback.go`

3. **Follow Command Semantics**:

   - Three mutually exclusive target flags
   - Validate targets before execution
   - Reuse deployment pipeline (by copying executePhases)
   - Create new release (don't mutate old ones)
   - Dry-run does NOT create release or execute phases
   - Avoid redundant state fetches (fetch current once, pass to helpers)

4. **Respect Constraints**:

   - See CONSTRAINTS section (canonical list)
   - Do not modify `CLI_DEPLOY`, `CLI_RELEASES`, or `CORE_STATE`
   - Do not attempt to normalize dry-run behavior with `CLI_DEPLOY`
   - Keep feature boundaries clean

---

## ✅ Final Checklist

Before starting work:

- [ ] Feature ID identified: `CLI_ROLLBACK`
- [ ] Git hooks verified
- [ ] Working directory clean
- [ ] On feature branch: `feature/CLI_ROLLBACK`
- [ ] Spec located/created: `spec/commands/rollback.md`
- [ ] Tests written first: `internal/cli/commands/rollback_test.go`
- [ ] Tests fail (as expected)
- [ ] Ready to implement

---

**Copy this entire document into your next agent session to continue development.**

This document is optimized for deterministic AI handoff and aligns with Stagecraft's Agent.md principles (spec-first, test-first, feature-bounded, deterministic).


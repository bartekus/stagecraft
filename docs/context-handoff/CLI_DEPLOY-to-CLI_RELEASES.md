---
## 📋 NEXT AGENT CONTEXT — After Completing Feature CLI_DEPLOY
---

## 🎉 LAYER 1: What Just Happened

### Feature Complete: CLI_DEPLOY

**Feature ID**: `CLI_DEPLOY`

**Status**: ✅ Implemented, fully tested, and PR open

**PR**: #4 (https://github.com/bartekus/stagecraft/pull/4)

**Commits**:
- `40f76a6` - `feat(CLI_DEPLOY): align spec, add failure tests, and refactor phases`
- `ac1e5d3` - `fix(CLI_DEPLOY): fix step numbering in deploy spec`

### What Now Exists

**Package**: `internal/cli/commands/`

- Deploy command with full phase tracking integration
- Release creation at deployment start
- Phase status transitions (Pending → Running → Completed/Failed)
- Failure semantics (mark failed phase, skip downstream phases)
- Version resolution (git SHA or "unknown" fallback)
- Dry-run mode support
- Centralized phase ordering via `orderedPhases()` helper
- Injectable phase executors for testing
- Comprehensive test coverage (8 tests including failure scenarios)

**APIs Available**:

```go
// From internal/cli/commands/deploy.go
NewDeployCommand() *cobra.Command

// Phase ordering helper
orderedPhases() []state.ReleasePhase

// Injectable phase executors (for testing)
var (
    buildPhaseFn, pushPhaseFn, migratePrePhaseFn,
    rolloutPhaseFn, migratePostPhaseFn, finalizePhaseFn
)
```

**Files Created**:

- `internal/cli/commands/deploy.go` (354 lines)
- `internal/cli/commands/deploy_test.go` (424 lines)
- `spec/commands/deploy.md` (257 lines)

**Files Updated**:

- `internal/cli/root.go` - Registered deploy command
- `spec/features.yaml` - CLI_DEPLOY status should be updated to `done` after PR merge

---

## 🎯 LAYER 2: Immediate Next Task

### Implement CLI_RELEASES

**Feature ID**: `CLI_RELEASES`

**Status**: `todo`

**Spec**: `spec/commands/releases.md` (needs creation)

**Dependencies**:

- ✅ `CORE_STATE` (ready)
- ✅ `CLI_DEPLOY` (ready)

**⚠️ SCOPE REMINDER**: All work in this handoff MUST be scoped strictly to `CLI_RELEASES`. Do not modify `CLI_DEPLOY` or other features.

**Reference Spec**: See `docs/context-handoff/CORE_STATE-to-CLI_DEPLOY.md` section "CLI_RELEASES" and `spec/core/state.md` for state APIs

---

### 🧪 MANDATORY WORKFLOW — Tests First

**Before writing ANY implementation code**:

1. **Create test file**: `internal/cli/commands/releases_test.go`

2. **Write failing tests** describing:

   - `releases list` command with no releases (empty state)
   - `releases list --env=ENV` filters by environment
   - `releases list` shows releases in correct order (newest first)
   - `releases show <release-id>` displays full release details
   - `releases show <invalid-id>` returns appropriate error
   - Phase status display in list and show commands
   - Help command output

3. **Run tests** - they MUST fail

4. **Only then** begin implementation

**Test Pattern** (follow existing CLI test patterns):

- Use `internal/cli/commands/deploy_test.go` as reference
- Use `internal/cli/commands/migrate_test.go` for subcommand pattern
- Create test releases using `state.Manager.CreateRelease()` in test setup
- Test output formatting (table/list view for list, detailed view for show)

---

### 🛠 Implementation Outline

**1. Command Structure** (subcommand pattern):

```go
// NewReleasesCommand returns the `stagecraft releases` command group
func NewReleasesCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "releases",
        Short: "List and show deployment releases",
        Long:  "View deployment release history and details",
    }
    
    cmd.AddCommand(NewReleasesListCommand())
    cmd.AddCommand(NewReleasesShowCommand())
    
    return cmd
}

// NewReleasesListCommand returns `stagecraft releases list`
func NewReleasesListCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "list",
        Short: "List releases for an environment",
        RunE:  runReleasesList,
    }
    // --env flag inherited from root
    return cmd
}

// NewReleasesShowCommand returns `stagecraft releases show`
func NewReleasesShowCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "show <release-id>",
        Short: "Show details of a specific release",
        Args:  cobra.ExactArgs(1),
        RunE:  runReleasesShow,
    }
    return cmd
}
```

**2. List Command Behavior**:

- Use `state.NewDefaultManager()`
- Call `ListReleases(ctx, env)` (or all environments if env not specified)
- Display in table format:
  - Release ID
  - Environment
  - Version
  - Timestamp (formatted)
  - Overall status (derived from phases: all completed = success, any failed = failed, etc.)
- Sort: newest first (already handled by `ListReleases`)

**3. Show Command Behavior**:

- Use `state.NewDefaultManager()`
- Call `GetRelease(ctx, releaseID)`
- Display detailed information:
  - Release ID
  - Environment
  - Version
  - Commit SHA (if available)
  - Timestamp
  - Previous release ID (if available)
  - Phase statuses (all 6 phases with their statuses)
- Format: human-readable, structured output

**4. Required Files**:

- `internal/cli/commands/releases.go` - Main releases command and subcommands
- `internal/cli/commands/releases_test.go` - Tests (write first!)
- `spec/commands/releases.md` - Spec (create)

**5. Integration Points**:

- Use `state.NewDefaultManager()` to create state manager
- Use `ListReleases(ctx, env)` for list command
- Use `GetRelease(ctx, id)` for show command
- Handle `state.ErrReleaseNotFound` appropriately
- Use `ResolveFlags()` for environment resolution

---

### 🧭 CONSTRAINTS (Canonical List)

**The next agent MUST NOT**:

- ❌ Modify existing `CLI_DEPLOY` behavior or code
- ❌ Modify `CORE_STATE` behavior or code
- ❌ Change JSON state format or structure
- ❌ Add/remove/rename phases or statuses
- ❌ Implement `CLI_ROLLBACK` now (wait until `CLI_RELEASES` is complete)
- ❌ Modify unrelated features
- ❌ Write directly to `.stagecraft/releases.json` (always use `state.Manager`)
- ❌ Skip the tests-first workflow

**The next agent MUST**:

- ✅ Write failing tests before implementation
- ✅ Follow existing CLI command patterns (see `internal/cli/commands/migrate.go` for subcommand pattern)
- ✅ Use `state.Manager` for all state operations
- ✅ Handle empty state gracefully (no releases)
- ✅ Format output in a user-friendly way
- ✅ Create/update spec if missing
- ✅ Keep feature boundaries clean (only `CLI_RELEASES` code)
- ✅ Register command in `internal/cli/root.go` in lexicographic order

---

## 📌 LAYER 3: Secondary Tasks

### CLI_ROLLBACK

**Feature ID**: `CLI_ROLLBACK`

**Status**: `todo`

**Dependencies**: `CORE_STATE` ✅, `CLI_DEPLOY` ✅ (ready after PR merge), `CLI_RELEASES` (todo)

**Do NOT start until `CLI_RELEASES` is complete.** (See CONSTRAINTS section)

Design can consider:
- `--to-previous` (use `PreviousID` from current release)
- `--to-release=<id>` (use `GetRelease()`)
- `--to-version=<version>` (search via `ListReleases()`)
- Integration with `CLI_DEPLOY` to re-deploy a previous version

---

## 🎓 Architectural Context (For Understanding)

**Why These Design Decisions Matter**:

- **Read-only snapshots**: `state.Manager` returns cloned releases to prevent accidental mutation
- **Environment filtering**: `ListReleases(ctx, env)` filters by environment, enabling multi-env deployments
- **Newest-first ordering**: `ListReleases` already sorts by timestamp (newest first), so list command can display directly
- **Release ID format**: `rel-YYYYMMDD-HHMMSS[mmm]` ensures lexicographic ordering matches chronological

**Integration Pattern Example** (for reference):

```go
// Example: How CLI_RELEASES should integrate with CORE_STATE
mgr := state.NewDefaultManager()

// List releases for environment
releases, err := mgr.ListReleases(ctx, "prod")
if err != nil {
    return fmt.Errorf("listing releases: %w", err)
}

// Display releases (already sorted newest first)
for _, release := range releases {
    // Format and display release info
    fmt.Printf("%s\t%s\t%s\t%s\n", 
        release.ID, release.Environment, release.Version, release.Timestamp)
}

// Show specific release
release, err := mgr.GetRelease(ctx, "rel-20250101-120000")
if err != nil {
    if errors.Is(err, state.ErrReleaseNotFound) {
        return fmt.Errorf("release not found")
    }
    return fmt.Errorf("getting release: %w", err)
}

// Display detailed release information
fmt.Printf("Release ID: %s\n", release.ID)
fmt.Printf("Environment: %s\n", release.Environment)
fmt.Printf("Version: %s\n", release.Version)
// ... display phases, etc.
```

---

## 📝 Output Expectations

**When you complete `CLI_RELEASES`**:

1. **Summary**: What was implemented

2. **Commit Message** (follow this format):

```
feat(CLI_RELEASES): implement releases list and show commands

Summary:
- Added releases.go with list and show subcommands
- Implemented release listing with environment filtering
- Implemented detailed release display
- Created releases_test.go with comprehensive tests
- Created/updated spec/commands/releases.md

Files:
- internal/cli/commands/releases.go
- internal/cli/commands/releases_test.go
- spec/commands/releases.md
- internal/cli/root.go (command registration)

Test Results:
- All tests pass
- Coverage meets targets
- No linting errors

Feature: CLI_RELEASES
Spec: spec/commands/releases.md
```

3. **Verification**:

   - ✅ Tests were written first (before implementation)
   - ✅ No unrelated changes were made
   - ✅ Feature boundaries respected (only `CLI_RELEASES` code)
   - ✅ All checks pass (`./scripts/run-all-checks.sh`)

---

## ⚡ Quick Start for Next Agent

**Bootloader Instructions**:

1. **Load Context**:

   - Read `internal/core/state/state.go` to understand API
   - Read `spec/core/state.md` for state semantics
   - Read `internal/cli/commands/migrate.go` for subcommand pattern reference
   - Read `internal/cli/commands/deploy_test.go` for test patterns
   - Check if `spec/commands/releases.md` exists (create if missing)

2. **Begin Work**:

   - Feature ID: `CLI_RELEASES`
   - Create feature branch: `feature/CLI_RELEASES`
   - Start with tests: `internal/cli/commands/releases_test.go`
   - Write failing tests first
   - Then implement: `internal/cli/commands/releases.go`

3. **Follow Command Semantics**:

   - Use subcommand pattern: `releases list` and `releases show`
   - Use `state.Manager` APIs correctly
   - Format output for readability
   - Handle errors gracefully

4. **Respect Constraints**:

   - See CONSTRAINTS section (canonical list)
   - Do not modify `CLI_DEPLOY` or `CORE_STATE`
   - Do not implement `CLI_ROLLBACK` yet
   - Keep feature boundaries clean

---

## ✅ Final Checklist

Before starting work:

- [ ] Feature ID identified: `CLI_RELEASES`
- [ ] Git hooks verified
- [ ] Working directory clean
- [ ] On feature branch: `feature/CLI_RELEASES`
- [ ] Spec located/created: `spec/commands/releases.md`
- [ ] Tests written first: `internal/cli/commands/releases_test.go`
- [ ] Tests fail (as expected)
- [ ] Ready to implement

---

**Copy this entire document into your next agent session to continue development.**

This document is optimized for deterministic AI handoff and aligns with Stagecraft's Agent.md principles (spec-first, test-first, feature-bounded, deterministic).


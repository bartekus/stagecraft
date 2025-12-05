# `stagecraft rollback` – Rollback Command

- Feature ID: `CLI_ROLLBACK`
- Status: todo
- Depends on: `CORE_STATE`, `CLI_DEPLOY`, `CLI_RELEASES`

## Goal

Provide a functional `stagecraft rollback` command that:
- Reverts an environment to a previous release by creating a new deployment
- Supports multiple rollback target selection methods
- Validates rollback targets before execution
- Reuses the deployment pipeline from `CLI_DEPLOY`
- Tracks rollback as a new release in history

## User Story

As a developer,
I want to run `stagecraft rollback --env=prod --to-previous`,
so that I can quickly revert to the previous working release when a deployment fails.

## Behaviour

### Input

- Environment name (required via `--env` flag)
- One of three mutually exclusive rollback target flags:
  - `--to-previous`: Rollback to the immediately previous release
  - `--to-release=<id>`: Rollback to a specific release ID
  - `--to-version=<version>`: Rollback to the most recent release with matching version
- Global flags: `--dry-run`, `--verbose`, `--config`

### Steps

1. Load config from `stagecraft.yml`
2. Validate environment exists in config
3. Parse rollback target flags (validate exactly one is provided)
4. Get current release if needed:
   - For `--to-previous`: Current release is required (used to find previous release)
   - For `--to-release` and `--to-version`: Current release is optional (used only for validation that target is not current)
5. Resolve rollback target using one of the three methods:
   - **Previous**: Use `current.PreviousID` to get target release via `GetRelease()` (requires current release)
   - **By ID**: Use `state.Manager.GetRelease(id)` to get target release (current release optional)
   - **By Version**: Use `state.Manager.ListReleases(env)` and find most recent matching version (current release optional)
6. Validate target release:
   - Target must exist
   - Target must not be the current release (cannot rollback to current; only checked if current release exists)
   - Target must have all phases completed (`StatusCompleted` for all 6 phases)
   - If `--to-previous` is used, current release must have a `PreviousID` (cannot rollback if only one release exists)
7. If `--dry-run`:
   - Log rollback plan (target release, version, commit SHA)
   - Optionally generate deployment plan to show what would happen
   - Return without creating a release or executing phases
8. If not `--dry-run`:
   - Create new release record using `state.Manager.CreateRelease()`:
     - Environment: same as current
     - Version: copied from target release
     - Commit SHA: copied from target release
     - All phases initialized as `StatusPending`
   - Generate deployment plan using `core.Planner.PlanDeploy()`
   - Execute deployment phases using same logic as `CLI_DEPLOY`
   - Update phase statuses during execution
   - Handle failures same as deploy (mark failed, skip downstream)

### Output

- Rollback target identification
- Validation results
- New release ID created (only in non-dry-run mode)
- Phase execution progress (same as deploy, only in non-dry-run mode)
- Success/failure message

**Example Output (non-dry-run)**:
```
Rolling back environment "prod" to release rel-20241215-134455000 (version: v1.2.3)...
Release rel-20241215-134455000 validated (all phases completed)
Creating rollback release...
Release created: rel-20250116-090122881
Planning...
Deploying...
Phase build: completed
Phase push: completed
Phase migrate_pre: completed
Phase rollout: completed
Phase migrate_post: completed
Phase finalize: completed
Rollback complete. New release: rel-20250116-090122881
```

**Example Output (dry-run)**:
```
Dry-run mode: would rollback environment "prod" to release rel-20241215-134455000
Target release: rel-20241215-134455000
Target version: v1.2.3
Target commit: abc123def456
Release rel-20241215-134455000 validated (all phases completed)
Would create rollback release with version v1.2.3
Would execute deployment phases: build, push, migrate_pre, rollout, migrate_post, finalize
```

### Error Handling

- **No current release**: `"no current release found for environment %q"`
- **No previous release** (when using `--to-previous`): `"no previous release to rollback to"`
- **Target release not found**: `"rollback target not found: %q"`
- **Target is current release**: `"cannot rollback to current release %q"`
- **Target not fully deployed**: `"rollback target %q is not fully deployed (phases: %v)"`
- **Multiple target flags**: `"only one rollback target flag may be specified"`
- **No target flag**: `"rollback target required; use --to-previous, --to-release, or --to-version"`
- **Environment mismatch**: `"release %q belongs to environment %q, not %q"`

## CLI Usage

```bash
# Rollback to previous release
stagecraft rollback --env=prod --to-previous

# Rollback to specific release ID
stagecraft rollback --env=prod --to-release=rel-20241215-134455000

# Rollback to specific version
stagecraft rollback --env=prod --to-version=v1.2.3

# Dry-run mode (does not create release or execute phases)
stagecraft rollback --env=prod --to-previous --dry-run

# Verbose output
stagecraft rollback --env=prod --to-previous --verbose
```

### Flags

- `--env <name>`: Target environment (required, inherited from root)
- `--to-previous`: Rollback to immediately previous release
- `--to-release=<id>`: Rollback to specific release ID
- `--to-version=<version>`: Rollback to most recent release with matching version
- `--dry-run`: Show rollback plan without creating release or executing phases
- `--verbose` / `-v`: Enable verbose output (inherited from root)
- `--config <path>`: Specify config file path (inherited from root)

## Target Resolution

### `--to-previous`

1. Get current release for environment
2. Check if `current.PreviousID` is set
3. If empty: return error "no previous release to rollback to"
4. Use `GetRelease(ctx, current.PreviousID)` to get target

### `--to-release=<id>`

1. Use `GetRelease(ctx, id)` to get target
2. Validate target environment matches `--env` flag
3. If mismatch: return error "release %q belongs to environment %q, not %q"

### `--to-version=<version>`

1. Use `ListReleases(ctx, env)` to get all releases for environment
2. Iterate through releases (already sorted newest first)
3. Find first release where `release.Version == version`
4. If not found: return error "no release found with version %q in environment %q"

## Validation Rules

### Target Must Exist
- All three resolution methods must return a valid release
- Handle `state.ErrReleaseNotFound` appropriately

### Target Must Not Be Current
- Compare target ID with current release ID
- If equal: return error "cannot rollback to current release"

### Target Must Be Fully Deployed
- Check all 6 phases have `StatusCompleted`
- Phases to check: `build`, `push`, `migrate_pre`, `rollout`, `migrate_post`, `finalize`
- If any phase is not `StatusCompleted`: return error with details

### Environment Match
- For `--to-release`, validate target release's environment matches `--env` flag
- For other methods, this is implicit (current release already filtered by env)

### Dry-run Semantics

- **Does NOT create a release**: Dry-run mode does not call `CreateRelease()` or write to state
- **Does NOT execute phases**: No phase execution or status updates
- **Does show plan**: Logs target release, version, commit SHA, and what would happen
- **May generate plan**: Optionally generate deployment plan to show operations (but don't execute)

**Note**: This differs from `CLI_DEPLOY`'s dry-run behavior, which creates a release. This asymmetry is intentional: rollback dry-run is a pure simulation, while deploy dry-run records intent.

## Implementation

### Command Structure

```go
// Feature: CLI_ROLLBACK
// Spec: spec/commands/rollback.md

func NewRollbackCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "rollback",
        Short: "Rollback environment to a previous release",
        Long:  "Rolls back an environment to a previous release by creating a new deployment with the target release's version",
        RunE:  runRollback,
    }
    
    cmd.Flags().Bool("to-previous", false, "Rollback to immediately previous release")
    cmd.Flags().String("to-release", "", "Rollback to specific release ID")
    cmd.Flags().String("to-version", "", "Rollback to most recent release with matching version")
    
    // Global flags (--config, --env, --verbose, --dry-run) are inherited from root
    
    return cmd
}
```

### State Integration

- Use `state.NewDefaultManager()` to create state manager
- Use `GetCurrentRelease(ctx, env)` to get current release
- Use `GetRelease(ctx, id)` to get target by ID
- Use `ListReleases(ctx, env)` to search by version
- Use `CreateRelease(ctx, env, version, commitSHA)` to create rollback release (only in non-dry-run)
- Use `UpdatePhase()` for phase tracking (same as deploy, only in non-dry-run)

**Note**: Method names in this document (`GetCurrentRelease`, `CreateRelease`, etc.) reflect the actual API in `internal/core/state/state.go`. Agents MUST confirm actual method signatures and adapt accordingly, rather than introducing new duplicate abstractions.

### Deploy Integration

- Reuse `executePhases()` function from `deploy.go` (see executePhases directive below)
- Reuse `orderedPhases()` helper
- Reuse phase execution functions (buildPhaseFn, pushPhaseFn, etc.)
- Generate plan using `core.Planner.PlanDeploy(env)`

**executePhases Directive**: The `executePhases` function in `deploy.go` is currently private (lowercase). For v1 of `CLI_ROLLBACK`, it is acceptable to copy the `executePhases` logic from `deploy.go` into `rollback.go` to avoid a premature refactor. Future refactors can consolidate them into a shared helper. Do NOT change behavior, only duplicate the code.

### Version Resolution

- Copy version from target release: `target.Version`
- Copy commit SHA from target release: `target.CommitSHA`
- No git resolution needed (using historical values)

### Flag Parsing

Flags are mutually exclusive. Use this helper:

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

## Validation

### Required Input

- `--env` flag must be provided
- Exactly one of `--to-previous`, `--to-release`, or `--to-version` must be provided
- Target release must exist and be valid

### Error Messages

- No target flag: `"rollback target required; use --to-previous, --to-release, or --to-version"`
- Multiple flags: `"only one rollback target flag may be specified"`
- No current release: `"no current release found for environment %q"`
- No previous: `"no previous release to rollback to"`
- Target not found: `"rollback target not found: %q"`
- Target is current: `"cannot rollback to current release %q"`
- Target incomplete: `"rollback target %q is not fully deployed (phases: %v)"`
- Environment mismatch: `"release %q belongs to environment %q, not %q"`

## Testing

Tests should cover:
- `--to-previous` with valid previous release
- `--to-previous` with no previous release (error)
- `--to-release` with valid release ID
- `--to-release` with invalid release ID (error)
- `--to-release` with environment mismatch (error)
- `--to-version` with matching version
- `--to-version` with no matching version (error)
- Target validation: current release (error)
- Target validation: incomplete deployment (error)
- Rollback creates new release with correct version/commit SHA
- Rollback executes deployment phases
- Rollback handles phase failures correctly
- **Dry-run mode does NOT create a release and does NOT execute phases**
- Multiple target flags (error)
- No target flags (error)

See `spec/features.yaml` entry for `CLI_ROLLBACK`:
- `internal/cli/commands/rollback_test.go` – unit/CLI behaviour tests

## Non-Goals (v1)

- Rollback to releases from different environments (v2)
- Partial rollback (rollback specific services only) (v2)
- Rollback with configuration changes (v2)
- Automatic rollback on deployment failure (v2)
- Rollback preview/diff (v2)

## Related Features

- `CORE_STATE` – State management for release tracking
- `CLI_DEPLOY` – Deploy command that rollback reuses
- `CLI_RELEASES` – Releases command for viewing history


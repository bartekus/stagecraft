# `stagecraft releases` – Releases Command

- Feature ID: `CLI_RELEASES`
- Status: implemented
- Depends on: `CORE_STATE`, `CLI_DEPLOY`

## Goal

Provide a functional `stagecraft releases` command that:
- Lists deployment releases for an environment
- Shows detailed information about a specific release
- Displays release history with phase status information
- Integrates with CORE_STATE for release data access

## User Story

As a developer,
I want to run `stagecraft releases list --env=prod` and `stagecraft releases show <release-id>`,
so that I can view deployment history and inspect specific release details.

## Behaviour

### Subcommands

The `releases` command has two subcommands:
- `list` - List releases for an environment
- `show` - Show details of a specific release

### `releases list` Command

#### Input

- Environment name (optional via `--env` flag)
- If `--env` is explicitly provided, shows releases only for that environment
- If `--env` is not provided, shows releases for all environments

#### Steps

1. Check if `--env` flag was explicitly provided using `cmd.Flags().Changed("env")`
2. Use `state.NewDefaultManager()` to create state manager
3. Get releases:
   - If `--env` was explicitly provided: Call `ListReleases(ctx, env)` to get releases for that environment
   - If `--env` was not provided: Call `ListAllReleases(ctx)` to get releases for all environments
4. Display releases in table format:
   - If single environment: sorted newest first (already sorted by `ListReleases`)
   - If all environments: grouped by environment (alphabetically), then sorted newest first within each environment (already sorted by `ListAllReleases`)

#### Output

Table format with columns:
- Release ID
- Environment
- Version
- Timestamp (formatted as human-readable date/time)
- Overall Status (derived from phases)

**Overall Status Derivation**:
- If any phase has `StatusFailed` → release status = `failed`
- Else if all phases are `StatusCompleted` → release status = `completed`
- Else if any phase is `StatusRunning` → release status = `running`
- Else → release status = `pending`

**Empty State**:
- If no releases exist, display: "No releases found"

**Multi-Environment Display**:
- When `--env` is not set, the table always includes the Environment column
- Releases are grouped by environment (alphabetically) and sorted newest first within each environment
- This grouping is achieved via sorting: environment (ascending), then timestamp (descending), then ID (ascending)

#### Example Output

```
RELEASE ID              ENVIRONMENT  VERSION    TIMESTAMP           STATUS
rel-20250101-120000123  prod        abc123def  2025-01-01 12:00:00  completed
rel-20241231-120000     prod        def456ghi  2024-12-31 12:00:00  completed
rel-20250101-110000     staging     abc123def  2025-01-01 11:00:00  running
```

### `releases show` Command

#### Input

- Release ID (required positional argument)

#### Steps

1. Get release ID from positional argument
2. Use `state.NewDefaultManager()` to create state manager
3. Call `GetRelease(ctx, releaseID)` to get release
4. Handle `state.ErrReleaseNotFound` if release doesn't exist
5. Display detailed release information

#### Output

Structured, human-readable output with:
- Release ID
- Environment
- Version
- Commit SHA (if available, otherwise "N/A")
- Timestamp (formatted)
- Previous Release ID (if available, otherwise "N/A")
- Phase Statuses (all 6 phases with their statuses)

**Phase Display Format**:
```
Phases:
  build:       completed
  push:        completed
  migrate_pre: completed
  rollout:     completed
  migrate_post: completed
  finalize:    completed
```

#### Example Output

```
Release ID:        rel-20250101-120000123
Environment:       prod
Version:           abc123def
Commit SHA:        abc123def456789
Timestamp:         2025-01-01 12:00:00
Previous Release: rel-20241231-120000

Phases:
  build:       completed
  push:        completed
  migrate_pre: completed
  rollout:     completed
  migrate_post: completed
  finalize:    completed
```

### Error Handling

- Release not found: `"release not found: <release-id>"`
- Invalid release ID format: Handled by state manager
- State file read error: Error from state manager

## CLI Usage

```bash
# List releases for all environments
stagecraft releases list

# List releases for specific environment
stagecraft releases list --env=prod

# Show specific release
stagecraft releases show rel-20250101-120000123

# Show release with verbose output
stagecraft releases show rel-20250101-120000123 --verbose
```

### Flags

- `--env <name>`: Filter releases by environment (inherited from root)
- `--verbose` / `-v`: Enable verbose output (inherited from root)
- `--config <path>`: Specify config file path (inherited from root)

## Implementation

### Command Structure

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

### State Integration

- Use `state.NewDefaultManager()` to create state manager
- For list command:
  - If `--env` is explicitly provided: Call `ListReleases(ctx, env)` for that environment
  - If `--env` is not provided: Call `ListAllReleases(ctx)` for all environments
- Call `GetRelease(ctx, id)` for show command
- Handle `state.ErrReleaseNotFound` appropriately
- Use `cmd.Flags().Changed("env")` to detect if `--env` was explicitly provided

### Output Formatting

- **List**: Table format with aligned columns
- **Show**: Key-value pairs with clear labels
- Use consistent date/time formatting
- Handle empty/missing values gracefully (show "N/A")

### Overall Status Calculation

```go
func calculateOverallStatus(release *state.Release) string {
    orderedPhases := []state.ReleasePhase{
        state.PhaseBuild,
        state.PhasePush,
        state.PhaseMigratePre,
        state.PhaseRollout,
        state.PhaseMigratePost,
        state.PhaseFinalize,
    }

    hasFailed := false
    hasRunning := false
    allCompleted := true

    for _, phase := range orderedPhases {
        status := release.Phases[phase]
        switch status {
        case state.StatusFailed:
            hasFailed = true
            allCompleted = false
        case state.StatusRunning:
            hasRunning = true
            allCompleted = false
        case state.StatusCompleted:
            // keep allCompleted as-is
        default:
            // missing or any other value → not completed
            allCompleted = false
        }
    }

    if hasFailed {
        return "failed"
    }
    if allCompleted {
        return "completed"
    }
    if hasRunning {
        return "running"
    }
    return "pending"
}
```

**Key points:**
- Iterates over the canonical ordered phase list (all 6 phases)
- Treats missing phases as "not completed" (consistent with `displayReleaseShow`)
- Ensures new releases with no phase updates show as `pending`, not `completed`

## Validation

### Required Input

- `show` command requires release ID as positional argument
- Release ID must exist in state

### Error Messages

- Release not found: `"release not found: <release-id>"`
- Missing release ID: Handled by Cobra's `ExactArgs(1)` validation

## Testing

Tests should cover:
- `releases list` with no releases (empty state)
- `releases list --env=ENV` filters by environment (single environment, no ENVIRONMENT column)
- `releases list` without `--env` shows releases for all environments (with ENVIRONMENT column)
- `releases list` shows releases in correct order (newest first within environment)
- `releases list` groups releases by environment when listing all environments
- `releases show <release-id>` displays full release details
- `releases show <invalid-id>` returns appropriate error
- Phase status display in list and show commands
- Overall status calculation:
  - All phases missing/unset → `pending`
  - Some phases completed, others missing → `pending`
  - One or more phases running, none failed → `running`
  - All phases completed → `completed`
  - At least one phase failed → `failed`
- Help command output for `releases`, `releases list`, and `releases show`
- Empty state handling

See `spec/features.yaml` entry for `CLI_RELEASES`:
- `internal/cli/commands/releases_test.go` – unit/CLI behaviour tests

## Non-Goals (v1)

- Release filtering by date range (v2)
- Release filtering by status (v2)
- Release comparison (v2)
- Release export/import (v2)
- Interactive release selection (v2)

## Related Features

- `CORE_STATE` – State management for release tracking
- `CLI_DEPLOY` – Deploy command that creates releases
- `CLI_ROLLBACK` – Rollback command that uses release history (future)


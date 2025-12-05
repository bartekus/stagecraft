# State Management (Release History)

- Feature ID: `CORE_STATE`
- Status: done
- Depends on: `CORE_CONFIG`, `CORE_ENV_RESOLUTION`

## Goal

Provide state management for tracking deployment history and release information.

State management enables Stagecraft to:
- Track release history per environment
- Support rollback operations
- Record deployment phases (build, migrate, rollout, etc.)
- Maintain deployment state in `.stagecraft/releases.json`

## Interface

```go
// internal/core/state/state.go

package state

import (
    "context"
    "time"
)

// ReleasePhase represents a deployment phase.
type ReleasePhase string

const (
    PhaseBuild      ReleasePhase = "build"
    PhasePush       ReleasePhase = "push"
    PhaseMigratePre ReleasePhase = "migrate_pre"
    PhaseRollout    ReleasePhase = "rollout"
    PhaseMigratePost ReleasePhase = "migrate_post"
    PhaseFinalize   ReleasePhase = "finalize"
)

// PhaseStatus represents the status of a phase.
type PhaseStatus string

const (
    StatusPending   PhaseStatus = "pending"
    StatusRunning   PhaseStatus = "running"
    StatusCompleted PhaseStatus = "completed"
    StatusFailed    PhaseStatus = "failed"
    StatusSkipped   PhaseStatus = "skipped"
)

// Release represents a single deployment release.
// Note: Release values returned from Manager methods should be treated as read-only snapshots.
// Mutating these values will not affect the stored state. To update state, use UpdatePhase or other Manager methods.
type Release struct {
    // ID is a unique identifier for this release (e.g., "rel-20250101-120000" or "rel-20250101-120000123")
    ID string

    // Environment is the target environment
    Environment string

    // Version is the deployed version (e.g., "v1.2.3" or git SHA)
    Version string

    // CommitSHA is the git commit SHA.
    // MAY be empty for non-git deployments.
    CommitSHA string

    // Timestamp is when the release was created
    Timestamp time.Time

    // Phases tracks the status of each deployment phase
    Phases map[ReleasePhase]PhaseStatus

    // PreviousID is the ID of the previous release (for rollback)
    PreviousID string
}

// Manager manages release state.
// Manager is safe for concurrent use within a single process.
// Note: State is not safe for concurrent modification from multiple processes.
type Manager struct {
    stateFile string
}

// NewManager creates a new state manager.
func NewManager(stateFile string) *Manager {
    return &Manager{stateFile: stateFile}
}

// NewDefaultManager creates a new state manager with the default state file path.
// If STAGECRAFT_STATE_FILE environment variable is set, it uses that path instead.
// The environment variable is read fresh on each call (no caching).
func NewDefaultManager() *Manager {
    // If STAGECRAFT_STATE_FILE is set, use it; otherwise use DefaultStatePath
    if envPath := os.Getenv("STAGECRAFT_STATE_FILE"); envPath != "" {
        return NewManager(envPath)
    }
    return NewManager(DefaultStatePath)
}

// DefaultStatePath is the default path for the state file.
const DefaultStatePath = ".stagecraft/releases.json"

// CreateRelease creates a new release record.
func (m *Manager) CreateRelease(ctx context.Context, env, version, commitSHA string) (*Release, error) {
    // Implementation:
    // 1. Generate release ID
    // 2. Get previous release ID for environment
    // 3. Create release record
    // 4. Save to state file
}

// GetRelease retrieves a release by ID.
// Returns a read-only snapshot of the release.
func (m *Manager) GetRelease(ctx context.Context, id string) (*Release, error) {
    // Implementation:
    // 1. Load state file
    // 2. Find release by ID
    // 3. Return cloned release (read-only snapshot)
}

// GetCurrentRelease retrieves the current release for an environment.
// Returns a read-only snapshot of the release.
func (m *Manager) GetCurrentRelease(ctx context.Context, env string) (*Release, error) {
    // Implementation:
    // 1. Load state file
    // 2. Find latest release for environment
    // 3. Return cloned release (read-only snapshot)
}

// UpdatePhase updates the status of a deployment phase.
func (m *Manager) UpdatePhase(ctx context.Context, releaseID string, phase ReleasePhase, status PhaseStatus) error {
    // Implementation:
    // 1. Load state file
    // 2. Find release
    // 3. Update phase status
    // 4. Save to state file
}

// ListReleases lists all releases for an environment, sorted newest first.
// Returns read-only snapshots of the releases.
func (m *Manager) ListReleases(ctx context.Context, env string) ([]*Release, error) {
    // Implementation:
    // 1. Load state file
    // 2. Filter releases by environment
    // 3. Sort by timestamp (newest first)
    // 4. Return list of cloned releases (read-only snapshots)
}
```

## State File Format

State is stored in `.stagecraft/releases.json`:

```json
{
  "releases": [
    {
      "id": "rel-20250101-120000",
      "environment": "prod",
      "version": "v1.2.3",
      "commit_sha": "abc123def456",
      "timestamp": "2025-01-01T12:00:00Z",
      "phases": {
        "build": "completed",
        "push": "completed",
        "migrate_pre": "completed",
        "rollout": "completed",
        "migrate_post": "completed",
        "finalize": "completed"
      },
      "previous_id": "rel-20241231-120000"
    }
  ]
}
```

## Behavior

### Release ID Generation

Release IDs follow the pattern: `rel-YYYYMMDD-HHMMSS[mmm]`
- Format: `rel-{date}-{time}[{milliseconds}]`
- Examples:
  - `rel-20250101-120000` (without milliseconds)
  - `rel-20250101-120000123` (with milliseconds)
- The optional millisecond suffix (`mmm`) ensures uniqueness for high-frequency operations
- Ensures lexicographic ordering matches chronological ordering
- IDs may be 19 or 22 characters in length depending on whether milliseconds are included

### Phase Tracking

Phases are tracked in order:
1. `build` - Docker image build
2. `push` - Push to registry
3. `migrate_pre` - Pre-deployment migrations
4. `rollout` - Container rollout
5. `migrate_post` - Post-deployment migrations
6. `finalize` - Finalization and cleanup

### State File Management

- State file is created automatically if it doesn't exist
- State file is atomically updated (write to temp, then rename)
- State file is JSON-formatted for readability
- State file is git-ignored by default (contains deployment history)

## State File Path Resolution

The state file path is determined by the following precedence order:

1. **Explicit path**: `NewManager(path)` always uses the provided path
2. **Environment variable**: `STAGECRAFT_STATE_FILE` - if set, `NewDefaultManager()` uses this path
3. **Default path**: `.stagecraft/releases.json` in the current working directory

### Environment Variable Support

The `STAGECRAFT_STATE_FILE` environment variable allows overriding the default state file path:

- If set, `NewDefaultManager()` reads the environment variable fresh on each call (no caching)
- Useful for testing isolation (each test can use its own isolated state file)
- The path should be absolute to avoid issues with working directory changes
- Example: `STAGECRAFT_STATE_FILE=/tmp/test-state.json stagecraft deploy`

## Usage Example

```go
import "stagecraft/internal/core/state"

// Create manager (using default path)
mgr := state.NewDefaultManager()
// or explicitly: mgr := state.NewManager(state.DefaultStatePath)
// or with env var: STAGECRAFT_STATE_FILE=/path/to/state.json stagecraft ...

// Create release
release, err := mgr.CreateRelease(ctx, "prod", "v1.2.3", "abc123")
if err != nil {
    return err
}

// Update phase
err = mgr.UpdatePhase(ctx, release.ID, state.PhaseBuild, state.StatusCompleted)

// Get current release
current, err := mgr.GetCurrentRelease(ctx, "prod")

// List releases
releases, err := mgr.ListReleases(ctx, "prod")
```

## Non-Goals (v1)

- Remote state backend (v1 uses local files)
- Distributed state synchronization
- State locking
- State migration/upgrade

## Related Features

- `CORE_ENV_RESOLUTION` - Environment resolution used for release tracking
- `CLI_DEPLOY` - Deploy command that creates and updates releases
- `CLI_ROLLBACK` - Rollback command that uses release history
- `CLI_RELEASES` - Releases command that lists release history


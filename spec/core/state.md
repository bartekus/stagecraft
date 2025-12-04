# State Management (Release History)

- Feature ID: `CORE_STATE`
- Status: todo
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
type Release struct {
    // ID is a unique identifier for this release (e.g., "rel-20250101-120000")
    ID string

    // Environment is the target environment
    Environment string

    // Version is the deployed version (e.g., "v1.2.3" or git SHA)
    Version string

    // CommitSHA is the git commit SHA
    CommitSHA string

    // Timestamp is when the release was created
    Timestamp time.Time

    // Phases tracks the status of each deployment phase
    Phases map[ReleasePhase]PhaseStatus

    // PreviousID is the ID of the previous release (for rollback)
    PreviousID string
}

// Manager manages release state.
type Manager struct {
    stateFile string
}

// NewManager creates a new state manager.
func NewManager(stateFile string) *Manager {
    return &Manager{stateFile: stateFile}
}

// CreateRelease creates a new release record.
func (m *Manager) CreateRelease(ctx context.Context, env, version, commitSHA string) (*Release, error) {
    // Implementation:
    // 1. Generate release ID
    // 2. Get previous release ID for environment
    // 3. Create release record
    // 4. Save to state file
}

// GetRelease retrieves a release by ID.
func (m *Manager) GetRelease(ctx context.Context, id string) (*Release, error) {
    // Implementation:
    // 1. Load state file
    // 2. Find release by ID
    // 3. Return release
}

// GetCurrentRelease retrieves the current release for an environment.
func (m *Manager) GetCurrentRelease(ctx context.Context, env string) (*Release, error) {
    // Implementation:
    // 1. Load state file
    // 2. Find latest release for environment
    // 3. Return release
}

// UpdatePhase updates the status of a deployment phase.
func (m *Manager) UpdatePhase(ctx context.Context, releaseID string, phase ReleasePhase, status PhaseStatus) error {
    // Implementation:
    // 1. Load state file
    // 2. Find release
    // 3. Update phase status
    // 4. Save to state file
}

// ListReleases lists all releases for an environment.
func (m *Manager) ListReleases(ctx context.Context, env string) ([]*Release, error) {
    // Implementation:
    // 1. Load state file
    // 2. Filter releases by environment
    // 3. Sort by timestamp (newest first)
    // 4. Return list
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

Release IDs follow the pattern: `rel-YYYYMMDD-HHMMSS`
- Format: `rel-{date}-{time}`
- Example: `rel-20250101-120000`
- Ensures lexicographic ordering matches chronological ordering

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

## Usage Example

```go
import "stagecraft/internal/core/state"

// Create manager
mgr := state.NewManager(".stagecraft/releases.json")

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


# `stagecraft deploy` – Deploy Command

- Feature ID: `CLI_DEPLOY`
- Status: implemented
- Depends on: `CORE_STATE`, `CORE_PLAN`, `CORE_COMPOSE`

## Goal

Provide a functional `stagecraft deploy` command that:
- Creates a release record at deployment start
- Tracks deployment phases (build, push, migrate_pre, rollout, migrate_post, finalize)
- Updates phase statuses during deployment (Pending → Running → Completed/Failed)
- Handles failures by marking failed phase and skipping downstream phases
- Integrates with CORE_PLAN for deployment planning
- Integrates with CORE_COMPOSE for Docker Compose operations

## User Story

As a developer,
I want to run `stagecraft deploy --env=prod` in my project,
so that my application is deployed to the target environment
with full phase tracking and release history.

## Behaviour

### Input

- Reads `stagecraft.yml` from current working directory (default)
- Environment name (required via `--env` flag)
- Version (optional via `--version` flag, defaults to git SHA)
- Commit SHA (auto-detected from git, or empty for non-git deployments)

### Steps

1. Load config from `stagecraft.yml`
2. Validate environment exists in config
3. Resolve version and commit SHA:
   - If `--version` is provided:
     - `version` is set to the flag value
     - `commitSHA` is set from `git rev-parse HEAD` if available, otherwise left empty
   - If `--version` is not provided and git is available:
     - `version` and `commitSHA` are set to the current git SHA
   - If `--version` is not provided and git is not available:
     - `version` is set to `"unknown"` to satisfy CORE_STATE's requirement that versions are non-empty
     - `commitSHA` is left empty
4. Create release record using `state.Manager.CreateRelease()`
   - All phases initialized as `StatusPending`
5. Generate deployment plan using `core.Planner.PlanDeploy()`
6. Execute deployment phases in order:
   - **Build**: Update phase to `StatusRunning`, execute build, update to `StatusCompleted` or `StatusFailed`
   - **Push**: Update phase to `StatusRunning`, execute push, update to `StatusCompleted` or `StatusFailed`
   - **MigratePre**: Update phase to `StatusRunning`, execute pre-deployment migrations, update to `StatusCompleted` or `StatusFailed`
   - **Rollout**: Update phase to `StatusRunning`, execute rollout, update to `StatusCompleted` or `StatusFailed`
   - **MigratePost**: Update phase to `StatusRunning`, execute post-deployment migrations, update to `StatusCompleted` or `StatusFailed`
   - **Finalize**: Update phase to `StatusRunning`, execute finalization, update to `StatusCompleted` or `StatusFailed`
7. On any phase failure:
   - Mark current phase as `StatusFailed`
   - Mark all downstream phases as `StatusSkipped`
   - Stop deployment (do not continue to next phase)
8. Return error if any phase failed

### Output

- Release ID created at start
- Phase progress messages
- Non-zero exit code if deployment fails
- Useful log lines (when `--verbose` is set):
  - Release ID
  - Environment
  - Version/Commit SHA
  - Phase status updates

### Error Handling

- Config file not found: Clear error message
- Invalid environment: Error with available environments
- Release creation failure: Error from state manager
- Phase execution failure: Mark phase as failed, skip downstream, return error
- Plan generation failure: Error from planner (see Plan Generation Failure section)

## CLI Usage

```bash
# Deploy to staging environment
stagecraft deploy --env=staging

# Deploy with specific version
stagecraft deploy --env=prod --version=v1.2.3

# Dry-run mode
stagecraft deploy --env=staging --dry-run
```

### Flags

- `--env <name>`: Target environment (required)
- `--version <version>`: Version to deploy (defaults to git SHA)
- `--verbose` / `-v`: Enable verbose output
- `--dry-run`: Show actions without executing
- `--config <path>`: Specify config file path

## Phase Semantics

### Phase Order

Phases must execute in this exact order:
1. `PhaseBuild` - Build Docker images
2. `PhasePush` - Push images to registry
3. `PhaseMigratePre` - Pre-deployment database migrations
4. `PhaseRollout` - Deploy containers (Docker Compose up)
5. `PhaseMigratePost` - Post-deployment database migrations
6. `PhaseFinalize` - Finalization and cleanup

### Phase Status Transitions

- All phases start as `StatusPending` when release is created
- When phase starts: `StatusPending` → `StatusRunning`
- On success: `StatusRunning` → `StatusCompleted`
- On failure: `StatusRunning` → `StatusFailed`
- On upstream failure: `StatusPending` → `StatusSkipped`

### Failure Semantics

- Only one phase may be `StatusRunning` at a time
- If a phase fails:
  - Mark that phase as `StatusFailed`
  - Mark all downstream phases as `StatusSkipped`
  - Stop deployment (do not continue)
- Deployment is considered failed if any phase is `StatusFailed`

### Dry-run Semantics

When `--dry-run` is set:

- The command still:
  - Loads config
  - Validates the environment
  - Resolves version and commit SHA
  - Creates a release record so that a "would be deployed" release is visible in history, with all phases initialized as `StatusPending`

- The command does not:
  - Generate a deployment plan
  - Execute any deployment phases
  - Update any phase statuses beyond `StatusPending`

This makes dry-run safe while still recording the intent to deploy.

### Plan Generation Failure

If deployment plan generation fails (for example, due to invalid configuration):

- The release is still created
- All phases are marked `StatusFailed`
- The command returns a non-zero exit code with an error of the form:
  `generating deployment plan: <underlying error>`

## Implementation

### Command Structure

```go
func NewDeployCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "deploy",
        Short: "Deploy application to environment",
        RunE:  runDeploy,
    }
    cmd.Flags().String("version", "", "Version to deploy (defaults to git SHA)")
    return cmd
}

func runDeploy(cmd *cobra.Command, args []string) error {
    // 1. Load config
    // 2. Resolve version/commit SHA
    // 3. Create release
    // 4. Generate plan
    // 5. Execute phases
    // 6. Update phase statuses
}
```

### State Integration

- Use `state.NewDefaultManager()` to create state manager
- Call `CreateRelease()` at deployment start
- Call `UpdatePhase()` for each phase transition
- Use phase constants: `state.PhaseBuild`, `state.PhasePush`, etc.
- Use status constants: `state.StatusPending`, `state.StatusRunning`, etc.

### Plan Integration

- Use `core.NewPlanner(cfg)` to create planner
- Call `PlanDeploy(envName)` to generate plan
- Plan operations map to deployment phases

### Compose Integration

- Use `compose.NewLoader()` to load compose files
- Generate environment-specific overrides
- Use for rollout phase

### Version Resolution

1. If `--version` is provided:
   - `version` is set to the flag value
   - `commitSHA` is set from `git rev-parse HEAD` if available, otherwise left empty

2. If `--version` is not provided and git is available:
   - `version` and `commitSHA` are set to the current git SHA

3. If `--version` is not provided and git is not available:
   - `version` is set to `"unknown"` to satisfy CORE_STATE's requirement that versions are non-empty
   - `commitSHA` is left empty

## Validation

### Required Config

- `environments[env]` must exist
- Environment must be valid

### Error Messages

- Config not found: `"stagecraft config not found at stagecraft.yml"`
- Invalid environment: `"invalid environment 'foo'; available environments: [dev, staging, prod]"`
- Release creation failed: Error from state manager
- Phase execution failed: `"deployment failed at phase 'build': <error>"`

## Testing

Tests should cover:
- Release creation at deploy start
- Phase sequencing (Pending → Running → Completed)
- Failure semantics (mark failed, skip downstream)
- Integration with state manager
- Integration with planner
- Error handling for invalid environments/versions
- Dry-run mode

See `spec/features.yaml` entry for `CLI_DEPLOY`:
- `internal/cli/commands/deploy_test.go` – unit/CLI behaviour tests

## Non-Goals (v1)

- Multi-host deployment (v2)
- Network provider integration (v2)
- Full Docker Compose orchestration (v2)
- Health checks (v2)
- Rollback (separate feature: `CLI_ROLLBACK`)

## Related Features

- `CORE_STATE` – State management for release tracking
- `CORE_PLAN` – Deployment planning
- `CORE_COMPOSE` – Docker Compose integration
- `CLI_ROLLBACK` – Rollback command (future)


---
feature: CLI_DEPLOY
version: v1
status: done
domain: commands
inputs:
  flags:
    - name: --env
      type: string
      default: ""
      description: "Target environment name (required)"
    - name: -e
      type: string
      default: ""
      description: "Shorthand for --env"
    - name: --version
      type: string
      default: ""
      description: "Deploy a specific version (optional)"
    - name: -v
      type: string
      default: ""
      description: "Shorthand for --version"
    - name: --dry-run
      type: bool
      default: "false"
      description: "Show what would be deployed without creating releases or executing side effects"
    - name: --config
      type: string
      default: ""
      description: "Override config file"
    - name: --verbose
      type: bool
      default: "false"
      description: "Increase logging verbosity"
outputs:
  exit_codes:
    success: 0
    error: 1
---
# CLI_DEPLOY - Deploy command

- **Feature ID**: `CLI_DEPLOY`
- **Domain**: `commands`
- **Status**: `done`
- **Related features**:
  - `CORE_PLAN`
  - `CORE_STATE`
  - `CORE_COMPOSE`
  - `CORE_STATE_CONSISTENCY`
  - `CLI_PHASE_EXECUTION_COMMON`
  - `DEPLOY_COMPOSE_GEN` (planned)
  - `DEPLOY_ROLLOUT` (planned)
  - `MIGRATION_PRE_DEPLOY` (planned)
  - `MIGRATION_POST_DEPLOY` (planned)

---

## 1. Purpose

`stagecraft deploy` deploys a given application version to a named environment using:

- The project configuration (`stagecraft.yml`)
- The canonical `docker-compose.yml`
- Provider implementations (primarily backend)
- A file-based release history managed by `CORE_STATE`

It is the core entry point for moving from built images to a running environment.

The command executes a **phase pipeline** consisting of six sequential phases: `build`, `push`, `migrate_pre`, `rollout`, `migrate_post`, and `finalize`. Each phase updates the release state, and failures cause downstream phases to be skipped. See [Section 5: Phase Model](#5-phase-model) for details.

**Version resolution**: If `--version` is not provided, the command resolves the version using the strategy defined in `CORE_PLAN` (typically the current Git SHA). See [Section 3.2: Flags](#32-flags) for flag details.

**Dry-run mode**: When `--dry-run` is set, the command shows what would be deployed without creating releases or executing any side effects. See [Section 7: Dry-Run Semantics](#7-dry-run-semantics) for complete behavior.

---

## 2. Scope

In v1, `CLI_DEPLOY` supports:

- Deploying to a single logical environment (`staging` or `prod`)
- Building backend images using the configured `BackendProvider`
- Pushing images to the configured registry
- Generating deployment phases and persisting them in state
- Executing a rollout for the target environment (minimum viable implementation may be `docker compose up` on the target host(s))
- Marking release status and phase outcomes in `.stagecraft/releases.json`
- A `--dry-run` mode that shows the plan and phase sequence without executing side effects

**Out of scope for this feature** (covered by other specs):

- Full per-host Compose generation (`DEPLOY_COMPOSE_GEN`)
- docker-rollout based zero downtime deployments (`DEPLOY_ROLLOUT`)
- Pre/post deploy migration execution (`MIGRATION_PRE_DEPLOY`, `MIGRATION_POST_DEPLOY`)
- Infrastructure provisioning (`CLI_INFRA_*`)
- CI workflow integration (`CLI_CI_*`)

---

## 3. CLI Interface

### 3.1 Usage

```text
stagecraft deploy [flags]
```

### 3.2 Flags

- `--env, -e <env>`
  - Required.
  - The target environment name (for example `staging`, `prod`).
  - Must exist under `environments` in `stagecraft.yml`.

- `--version, -v <version>`
  - Optional.
  - Deploy a specific version (for example Git SHA or tag).
  - If omitted, the implementation must use the version resolution strategy defined in `CORE_PLAN` (for example environment variable, current Git SHA, or CI supplied value). The concrete mechanism is defined in `core/plan.md`.

- `--dry-run`
  - Optional.
  - When set, the command:
    - Resolves the plan and phases
    - Creates an in-memory representation of the release and phase list
    - Logs or prints:
      - Target env
      - Version
      - Planned phases
    - Does not:
      - Create or modify the state file.
      - Call any external commands (docker, docker-rollout, migrations).
      - Connect to remote hosts.

- `--config <path>`
  - Optional.
  - Override config file, consistent with `CLI_GLOBAL_FLAGS`.

- `--verbose`
  - Optional.
  - Increase logging verbosity, consistent with `CORE_LOGGING` semantics.

Additional flags (for example `--host`, `--role`, `--plan-only`) may be added in later specs. This spec defines the minimal v1.

---

## 4. Inputs and Outputs

### 4.1 Inputs

- `stagecraft.yml` at repo root:
  - `project.registry`
  - `backend.provider` and backend configuration
  - `environments.<env>` configuration
  - `hosts` and `services` mappings (for future per-host deploy)

- Canonical `docker-compose.yml` at repo root
  - Optional override files (implementation detail of `CORE_COMPOSE`)

- Environment variables:
  - Used by `CORE_CONFIG` and providers as specified in their own specs

- State file:
  - `.stagecraft/releases.json`
  - Location can be overridden via `STAGECRAFT_STATE_FILE` (see `CORE_STATE` spec)

### 4.2 Outputs

- **Side effects**:
  - A new release entry persisted in `.stagecraft/releases.json`
  - Phase statuses updated according to execution results
  - Built and pushed Docker images (if not `--dry-run`)
  - Deployments executed on the target environment (if not `--dry-run`)

- **CLI output**:
  - Human readable summary of:
    - Target env and version
    - Planned phases and their status
    - Errors, if any
  - Output must be deterministic given the same inputs (no random ordering, no timestamps).

---

## 5. Phase Model

`CLI_DEPLOY` uses the shared phase execution semantics from `CLI_PHASE_EXECUTION_COMMON`. The deploy pipeline is modeled as a fixed sequence of phases:

1. `build`
2. `push`
3. `migrate_pre` (future)
4. `rollout`
5. `migrate_post` (future)
6. `finalize`

### 5.1 Phase order and behavior

The default v1 sequence:

1. **Build phase (`build`)**
   - Uses `BackendProvider.BuildDocker` (or equivalent) to build the backend image(s).
   - Must produce deterministic image tags.
   - On success:
     - Persist build phase as success.
   - On failure:
     - Persist build phase as failed.
     - All downstream phases become skipped.
     - Command exits with non-zero code.

2. **Push phase (`push`)**
   - Pushes the built images to `project.registry`.
   - Implementation for v1:
     - May use `executil` to call `docker push`.
   - On success:
     - Persist push phase as success.
   - On failure:
     - Persist push as failed, downstream phases as skipped, and exit non-zero.

3. **Pre-migration phase (`migrate_pre`)** - placeholder in v1
   - For v1:
     - No-op in implementation, but the phase is still created in the release.
   - Future:
     - Runs pre-deploy migrations using the migration interface.
   - Status:
     - In v1, implementation may leave this phase as `pending` or mark as `skipped` with a reason. The spec for `MIGRATION_PRE_DEPLOY` will define final semantics.

4. **Rollout phase (`rollout`)**
   - Deploys the application to the target environment.
   - v1 minimum behavior:
     - Use `CORE_PLAN` to determine which services and hosts are involved.
     - Generate or select appropriate Compose configuration (minimal version can reuse canonical `docker-compose.yml` for a single host).
     - Perform deployment using either:
       - `docker compose up -d` on the target host, or
       - A stubbed equivalent that can be replaced by `DEPLOY_ROLLOUT` later.
   - On success:
     - Persist rollout phase as success.
   - On failure:
     - Persist rollout as failed, downstream phases as skipped, and exit non-zero.

5. **Post-migration phase (`migrate_post`)** - placeholder in v1
   - Same rules as `migrate_pre`, but intended for post-deploy migrations.
   - Actual semantics defined in `MIGRATION_POST_DEPLOY` spec.

6. **Finalize phase (`finalize`)**
   - Performs final bookkeeping:
     - Mark release as complete.
     - Optionally mark it as current for the environment.
   - Must be called after all previous phases succeed.
   - On success:
     - Persist finalize as success.
   - On failure:
     - Persist finalize as failed and exit non-zero.

### 5.2 Phase semantics

- Phases are executed in order.
- On the first phase failure:
  - That phase is marked failed.
  - All remaining phases are marked skipped with a deterministic reason.
  - The deploy command returns an error.
- Phase names are stable identifiers and must not be changed without migration tooling.

---

## 6. State Interaction

`CLI_DEPLOY` writes to the state store using `CORE_STATE`:

- Creates a new release entry:
  - Environment
  - Version
  - Phase list with initial status (for example `pending`)
- Updates phases as they complete or fail.
- The state manager must guarantee read-after-write consistency (`CORE_STATE_CONSISTENCY`).

### 6.1 Release selection rules

- Deploy must always create a new release entry for each invocation.
- Release IDs come from `CORE_STATE`.
- Rollback and releases inspection rely on this structure, so `CLI_DEPLOY` must not:
  - Reuse existing release IDs.
  - Modify phases of prior releases.

---

## 7. Dry-Run Semantics

When `--dry-run` is set:

- `CLI_DEPLOY`:
  - Resolves config and environment.
  - Resolves target version.
  - Builds an in-memory representation of the intended release and phase list.
  - Logs or prints:
    - Target env
    - Version
    - Planned phases
- Does not:
  - Create or modify the state file.
  - Call any external commands (docker, docker-rollout, migrations).
  - Connect to remote hosts.

The dry-run path must share as much logic as possible with the real path, excluding side effects.

---

## 8. Error Handling

- All errors must be wrapped with context (per `CORE_LOGGING` and Go error wrapping conventions).
- CLI exit codes:
  - `0` - success
  - Non-zero - any failure, including:
    - Config or plan resolution errors
    - State read/write errors
    - Phase execution failures
- Common error classes:
  - Misconfigured environment (`--env` not defined in config)
  - Missing registry or provider configuration
  - Backend build errors
  - Docker CLI errors
  - State persistence failures

Errors must not leak sensitive data (for example registry credentials).

---

## 9. Determinism Requirements

`CLI_DEPLOY` must obey global determinism rules:

- No direct use of timestamps, random UUIDs, or environment-dependent ordering in CLI behavior.
- Any list of phases, hosts, or services must be rendered in deterministic order:
  - Lexicographical ordering where applicable.
- State writes must be deterministic given the same inputs and external state.

---

## 10. Dependencies

`CLI_DEPLOY` depends on:

- `CORE_CONFIG` - loading `stagecraft.yml`
- `CORE_PLAN` - building deployment plans
- `CORE_COMPOSE` - Compose file handling
- `CORE_STATE` + `CORE_STATE_CONSISTENCY` - release history
- `CLI_PHASE_EXECUTION_COMMON` - shared phase execution semantics
- Backend provider (for example `PROVIDER_BACKEND_ENCORE`) with a build method

For v1, `CLI_DEPLOY` must not introduce new third-party dependencies beyond those approved in `Agent.md`.

---

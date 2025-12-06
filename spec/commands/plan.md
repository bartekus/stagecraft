# CLI_PLAN - Plan command

- **Feature ID**: `CLI_PLAN`
- **Domain**: `commands`
- **Status**: `todo`
- **Related features**:
  - `CORE_PLAN` (done)
  - `CORE_CONFIG` (done)
  - `CLI_DEPLOY` (done)
  - `CLI_BUILD` (done)

---

## 1. Purpose

`stagecraft plan` is the read-only, side-effect-free view of what `stagecraft deploy` (and partially `stagecraft build`) would do.

The command:

- Uses `CORE_PLAN` to generate the full orchestration plan
- Applies the same filters as deploy (env, services, roles, hosts, version)
- Outputs a deterministic, human-readable (and optionally machine-readable) description of:
  - Phases
  - Hosts/roles
  - Services and images
  - Migrations
  - Compose files / rollout steps (at least conceptually)

**Non-goals for v1:**

- No remote calls (no SSH, no docker, no Tailscale)
- No state mutations (no `.stagecraft/releases.json` writes)
- No infra provisioning; infra is assumed to exist
- No Kubernetes transforms yet

---

## 2. Scope

In v1, `CLI_PLAN` supports:

- Generating deployment plans for a single logical environment (`staging` or `prod`)
- Filtering by services, roles, hosts, phases
- Outputting plans in text or JSON format
- Version resolution (mirroring deploy command semantics)
- Deterministic, stable output suitable for golden file testing

**Out of scope for this feature** (covered by other specs):

- Executing any operations (covered by `CLI_DEPLOY`, `CLI_BUILD`)
- State file writes (covered by `CORE_STATE`)
- Remote host connectivity (covered by deployment drivers)
- Infrastructure provisioning (`CLI_INFRA_*`)

---

## 3. CLI Interface

### 3.1 Usage

```text
stagecraft plan [flags]
```

### 3.2 Flags

#### Required Flags

- `--env, -e <env>`
  - Required
  - The target environment name (for example `staging`, `prod`)
  - Must exist under `environments` in `stagecraft.yml`
  - If omitted, the command MUST exit with code 2 and a deterministic error message

#### Optional Flags

- `--version, -v <version>`
  - Optional
  - Version to plan for (for example Git SHA or tag)
  - If omitted, CLI_PLAN MUST NOT shell out to git or any external command. In that case, it MUST use the string "unknown" as the version marker in output.
  - Unlike `deploy` and `build`, the plan command does not call external commands (including git) to maintain its read-only, side-effect-free guarantee.
  - Used only to annotate the plan output; no state writes occur

- `--services <svc1,svc2,...>`
  - Optional
  - Comma-separated list of services to include
  - Filtering semantics: a phase is kept if it touches at least one of the specified services
  - If a service is specified that doesn't exist in the plan, the command MUST exit with code 2

- `--roles <role1,role2,...>`
  - Optional (v1 minimal; can be marked as future extension)
  - Comma-separated list of host roles to filter by
  - Filtering semantics: deploy phases are filtered by target role, but upstream dependencies (build, migrate) are kept if necessary

- `--hosts <host1,host2,...>`
  - Optional (v1 minimal; can be marked as future extension)
  - Comma-separated list of hostnames to filter by
  - Filtering semantics: similar to `--roles`, but filters by explicit hostname

- `--phases <phase1,phase2,...>`
  - Optional (v1 minimal; can be marked as future extension)
  - Filter by phase IDs or prefixes (e.g. `BUILD_`, `DEPLOY_`, `MIGRATE_`)
  - Filtering semantics: only show phases matching the specified IDs or prefixes

- `--format <format>`
  - Optional
  - Output format: `text` (default) or `json`
  - `text`: Human-readable hierarchical layout
  - `json`: Machine-readable JSON encoding suitable for tooling

- `--verbose, -V`
  - Optional
  - Show more detail (per-phase, per-host, per-service)
  - In text mode, adds additional detail lines
  - In JSON mode, includes additional metadata fields

- `--config <path>`
  - Optional
  - Override config file, consistent with `CLI_GLOBAL_FLAGS`

**Note**: For v1, we can keep the implementation minimal with `--env`, `--version`, `--services`, and `--format`. The remaining flags (`--roles`, `--hosts`, `--phases`, `--verbose`) can be marked as "future extension" in the spec if preferred.

---

## 4. Inputs and Outputs

### 4.1 Inputs

- `stagecraft.yml` at repo root:
  - `project.registry`
  - `backend.provider` and backend configuration
  - `environments.<env>` configuration
  - `hosts` and `services` mappings
  - `databases` and migration configuration

- Environment variables:
  - Used by `CORE_CONFIG` and providers as specified in their own specs
  - Git environment (for version resolution)

### 4.2 Outputs

- **Side effects**: None
  - The command MUST NOT:
    - Create or modify the state file
    - Call any external commands (docker, git, SSH)
    - Connect to remote hosts
    - Write any files

- **CLI output**:
  - Human-readable summary (text format) or JSON (json format)
  - Output MUST be deterministic given the same inputs:
    - No random ordering
    - No timestamps
    - Stable lexicographical ordering for lists
    - Phases displayed in execution order (topological) or by stable ID

---

## 5. Behaviour

### 5.1 High-Level Workflow

1. **Load config**
   - Reads `stagecraft.yml` via `CORE_CONFIG`
   - Validates provider IDs via registries (already in place)

2. **Resolve environment**
   - Ensures `--env` exists in `environments` map
   - Assembles environment context (paths, volumes, traefik mode, etc.)

3. **Resolve version**
   - If `--version` provided, use it
   - Else: use "unknown" (CLI_PLAN does NOT shell out to git or any external command)
   - No state write; purely used to annotate the plan

4. **Generate plan (`CORE_PLAN`)**
   - Call into `internal/core` to generate the full plan for:
     - Selected environment
     - Selected version (stored in plan metadata)
     - Full service graph (or filtered subset)

5. **Apply filters**
   - Services: restrict to the subset requested (`--services`)
   - Phases: restrict by IDs/prefixes (`--phases`)
   - Roles/hosts: restrict as requested (`--roles`, `--hosts`)

6. **Render output**
   - `--format=text`:
     - Deterministic human-readable summary (sorted by phase ID, host, service)
   - `--format=json`:
     - Stable JSON encoding of the plan (subset of `CORE_PLAN` data structures) suitable for tooling
   - No colours, no timestamps; line ordering is stable

7. **Exit codes**
   - `0` – plan successfully generated and rendered
   - `2` – user error (missing/invalid flags, unknown env, invalid filter)
   - `3` – planning failed (invalid config, provider misconfig, `CORE_PLAN` error)
   - `1` – unexpected internal error

### 5.2 Filter Semantics

#### Service Filtering (`--services`)

- Service filtering is **inclusive OR**: a phase is kept if it touches at least one of the specified services
- Upstream dependencies are preserved:
  - If a deploy phase for service `api` is included, its build phase is also included
  - Migration phases that affect the selected services are included
- If a service is specified that doesn't exist in the config, the command MUST exit with code 2

#### Phase Filtering (`--phases`)

- Filter by phase IDs or prefixes
- Examples:
  - `--phases=BUILD_` matches all phases starting with `BUILD_`
  - `--phases=BUILD_BACKEND,DEPLOY_APP` matches specific phases
- Dependencies are preserved: if a phase is included, its dependencies are also included

#### Role/Host Filtering (`--roles`, `--hosts`)

- Filter deploy phases by target role or hostname
- Upstream dependencies (build, migrate) are kept if necessary
- Example: `--roles=gateway` shows deploy phases for gateway role, plus any build/migrate phases needed

### 5.3 Ordering Guarantees

- Phases displayed in execution order (topological sort) or by stable ID (lexicographical)
- Within a phase:
  - Hosts sorted lexicographically
  - Services sorted lexicographically
- JSON output: arrays sorted lexicographically

---

## 6. Output Formats

### 6.1 Text Format

The text format provides a hierarchical, human-readable layout:

```
Environment: <env>
Version: <version>
Services: <all|filtered list>
Hosts: <all|filtered list>

Phases:
  1. <phase_id>
     - kind: <build|deploy|migrate|health_check>
     - services: [<svc1>, <svc2>]
     - hosts: [<host1>, <host2>]
     - description: <description>
     - depends_on: [<phase_id1>, <phase_id2>]

  2. <phase_id>
     ...
```

**Deterministic properties:**
- Phases sorted by execution order (topological) or by ID (lexicographical)
- Services and hosts within each phase sorted lexicographically
- No timestamps
- No random ordering

### 6.2 JSON Format

The JSON format provides a machine-readable encoding:

```json
{
  "env": "<env>",
  "version": "<version>",
  "phases": [
    {
      "id": "<phase_id>",
      "kind": "<build|deploy|migrate|health_check>",
      "services": ["<svc1>", "<svc2>"],
      "hosts": ["<host1>", "<host2>"],
      "description": "<description>",
      "depends_on": ["<phase_id1>", "<phase_id2>"],
      "metadata": {
        "<key>": "<value>"
      }
    }
  ]
}
```

**Deterministic properties:**
- Phases array sorted by execution order (topological) or by ID (lexicographical)
- Services and hosts arrays sorted lexicographically
- Metadata keys sorted lexicographically
- Schema is stable across v1 minor releases

---

## 7. Error Handling

- All errors MUST be wrapped with context (per `CORE_LOGGING` and Go error wrapping conventions)
- CLI exit codes:
  - `0` – success
  - `2` – user error (missing/invalid flags, unknown env, invalid filter)
  - `3` – planning failed (invalid config, provider misconfig, `CORE_PLAN` error)
  - `1` – unexpected internal error
- Common error classes:
  - Missing `--env` flag
  - Unknown environment (`--env` not defined in config)
  - Invalid service filter (service doesn't exist)
  - Config load errors
  - Plan generation errors (`CORE_PLAN` failures)
- Errors MUST NOT leak sensitive data (for example registry credentials)

---

## 8. Determinism Requirements

`CLI_PLAN` MUST obey global determinism rules:

- No direct use of timestamps, random UUIDs, or environment-dependent ordering in CLI behavior
- Any list of phases, hosts, or services MUST be rendered in deterministic order:
  - Lexicographical ordering where applicable
  - Topological ordering for phases (execution order)
- Output MUST be bit-for-bit identical given the same inputs

---

## 9. Examples

### 9.1 Basic Usage

```bash
stagecraft plan --env=staging
```

**Sample text output:**

```
Environment: staging
Version: 2025-12-05T180301Z-abc1234
Services: (all)
Hosts: (all)

Phases:
  1. BUILD_BACKEND
     - kind: build
     - services: [api]
     - hosts: []
     - description: Build backend using provider encore-ts
     - depends_on: []

  2. BUILD_FRONTEND
     - kind: build
     - services: [web]
     - hosts: []
     - description: Build frontend using provider generic-dev-command
     - depends_on: []

  3. MIGRATE_DB
     - kind: migrate
     - services: [db]
     - hosts: []
     - description: Run pre_deploy migrations for database main
     - depends_on: []

  4. DEPLOY_DB
     - kind: deploy
     - services: [db]
     - hosts: [plat-db-1]
     - description: Deploy to environment staging
     - depends_on: [MIGRATE_DB]

  5. DEPLOY_APP
     - kind: deploy
     - services: [api, web]
     - hosts: [plat-api-1, plat-web-1]
     - description: Deploy to environment staging
     - depends_on: [BUILD_BACKEND, BUILD_FRONTEND, MIGRATE_DB]

  6. DEPLOY_GATEWAY
     - kind: deploy
     - services: [traefik]
     - hosts: [plat-gw-1]
     - description: Deploy to environment staging
     - depends_on: [DEPLOY_APP]
```

### 9.2 Filtered Services

```bash
stagecraft plan --env=prod --services=api,web
```

Shows only phases that affect `api` or `web` (build, migrate, deploy subset), with dependencies preserved.

### 9.3 JSON Format

```bash
stagecraft plan --env=staging --format=json > plan.json
```

**Sample JSON output:**

```json
{
  "env": "staging",
  "version": "2025-12-05T180301Z-abc1234",
  "phases": [
    {
      "id": "BUILD_BACKEND",
      "kind": "build",
      "services": ["api"],
      "hosts": [],
      "description": "Build backend using provider encore-ts",
      "depends_on": [],
      "metadata": {
        "provider": "encore-ts"
      }
    },
    {
      "id": "DEPLOY_APP",
      "kind": "deploy",
      "services": ["api", "web"],
      "hosts": ["plat-api-1", "plat-web-1"],
      "description": "Deploy to environment staging",
      "depends_on": ["BUILD_BACKEND", "BUILD_FRONTEND", "MIGRATE_DB"],
      "metadata": {
        "environment": "staging"
      }
    }
  ]
}
```

---

## 10. Dependencies

`CLI_PLAN` depends on:

- `CORE_CONFIG` – loading `stagecraft.yml`
- `CORE_PLAN` – building deployment plans
- `CLI_DEPLOY` – for version resolution semantics (shared logic)
- `CLI_BUILD` – for understanding build phase semantics

For v1, `CLI_PLAN` MUST NOT introduce new third-party dependencies beyond those approved in `Agent.md`.

---

## 11. Testing Requirements

Tests MUST verify:

1. **Missing env**
   - Input: `stagecraft plan` with no `--env`
   - Expect: exit code 2, help or deterministic error

2. **Unknown env**
   - Input: `stagecraft plan --env=foo`
   - Expect: error `unknown environment: foo`

3. **Happy path text**
   - Set up an isolated test env (using `setupIsolatedStateTestEnv` style helper, though plan doesn't write state)
   - Run `stagecraft plan --env=staging`
   - Compare output against golden file `plan_staging_all.golden`

4. **Service filtering**
   - Run `--services=api`
   - Golden file ensures only relevant phases appear, but dependencies (build, migrations) are still present

5. **JSON format**
   - Run `--format=json`
   - Validate:
     - Valid JSON
     - Contains expected env/version and at least one phase
     - Either snapshot JSON (normalized) or parse and assert fields

6. **Determinism**
   - Run the same command twice in a row in the same test env
   - Outputs MUST match bit-for-bit (no timestamps or random ordering)

7. **Error propagation**
   - Arrange `CORE_PLAN` to fail (e.g., invalid config, missing hosts)
   - Ensure error message propagates and exit code is 3 (planning error)

### 11.1 Golden Test Layout

- `internal/cli/commands/testdata/plan_staging_all.txt`
- `internal/cli/commands/testdata/plan_prod_api_only.txt`
- `internal/cli/commands/testdata/plan_staging_json.json` (optional; or construct expected structure in code instead of golden)

All golden files MUST:
- Use Unix newlines
- Have sorted sections
- Have no trailing spaces

---

## 12. Implementation Notes

### 12.1 File Structure

- `spec/commands/plan.md` – this spec (fully fleshed)
- `internal/cli/commands/plan.go` – Cobra wiring plus orchestration
- `internal/cli/commands/plan_test.go` – tests, mostly golden
- `internal/cli/commands/testdata/plan_*.golden` – CLI output snapshots

No new core files required; reuse `CORE_PLAN`.

### 12.2 Command Skeleton

The command should follow the pattern established by `CLI_DEPLOY` and `CLI_BUILD`:

1. Parse flags
2. Load config
3. Resolve env + version
4. Generate plan via `core.NewPlanner(cfg).PlanDeploy(env)`
5. Apply filters
6. Render text/json

### 12.3 Core Integration

The command should delegate most logic into a small helper function to keep Cobra wiring thin:

```go
type PlanOptions struct {
    Env      string
    Version  string
    Services []string
    Hosts    []string
    Roles    []string
    Phases   []string
    Format   string
    Verbose  bool
}

func ExecutePlan(ctx context.Context, cfg *config.Config, opts PlanOptions) error {
    // 1. Build environment context
    // 2. Call core.Plan(...) to obtain a Plan struct
    // 3. Filter the Plan
    // 4. Render according to opts.Format
}
```

This keeps CLI thin and re-usable for future "plan as library" use (tests, other commands, etc.).


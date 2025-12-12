---
feature: ENGINE_PLAN_ACTIONS
version: v1
status: draft
domain: engine
inputs:
  flags: []
outputs:
  exit_codes: {}
---
# Engine Plan Actions and Inputs Schema

- Feature ID: `ENGINE_PLAN_ACTIONS`
- Status: draft
- Domain: engine
- Purpose: Defines the canonical Inputs payload schema for each `engine.StepAction`

This document defines the canonical Inputs payload schema for each `engine.StepAction`.
It is the contract boundary between:
- Planner/Engine (producer)
- CLI/Controller (transport/orchestration)
- Agent/Executors (consumer)

---

## Global Rules

### 1. Wire Format

- `PlanStep.Inputs` and `HostPlanStep.Inputs` are JSON objects.
- Inputs MUST be encoded using deterministic JSON from typed structs (no `map[string]interface{}` beyond planner internals).
- Unknown keys policy is defined per action, but the default is: **reject unknown keys**.
- **Consumers MUST reject unknown fields** by decoding with `json.Decoder` + `DisallowUnknownFields()` (or equivalent strict decoder).

### 2. Determinism Rules (apply to all actions)

- Lists that represent sets MUST be sorted lexicographically unless explicitly stated otherwise.
- Maps MUST NOT appear in Inputs unless the map keys are sorted deterministically before serialization (prefer arrays of `{key,value}` instead).
- If an Inputs field represents a collection of named items, the list MUST be sorted by the item's stable key (usually `name`).
- Hash fields MUST be content-addressed and based on canonical bytes. If canonicalization is unclear, store both:
  - `hash_alg` (string, example: "sha256")
  - `hash` (hex lowercase)
  - **If `hash_alg == "sha256"`, `hash` MUST be exactly 64 lowercase hex characters**.

### 3. Validation Rules (apply to all actions)

- Required fields MUST be present and non-empty.
- Strings MUST NOT contain leading/trailing whitespace (producers should trim).
- Paths MUST be relative to the execution root unless explicitly stated otherwise.
- **Paths MUST use forward slashes (`/`) and MUST NOT contain `..` or `.` segments** (normalize before serialization).
- If an Inputs struct includes an enum field, values MUST match exactly.

### 4. Defaults Rules

- **Defaults MUST be applied by the producer**; consumers MUST NOT invent defaults.
- Producers MUST materialize all default values explicitly in Inputs JSON.
- Default values documented in this spec are for reference only; they MUST appear in the wire format.
- This ensures deterministic behavior and prevents consumer-side drift.

### 5. Forward Compatibility Rules

Each action declares its unknown-field behavior:
- **reject**: unknown keys cause validation failure (default for v1)
- **ignore**: unknown keys are ignored (only use when you expect rapid evolution)

For v1, default is **reject** to prevent silent drift.

---

## Action: build (`StepActionBuild`)

**Purpose:**
Build an artifact (typically a container image) using a provider-specific builder.

**Unknown-field behavior:** reject

### Inputs Schema (v1)

**Required:**
- `provider` (string) - builder provider identifier (example: "generic", "docker", "buildkit")
- `workdir` (string) - working directory relative to execution root

**Optional:**
- `target` (string) - build target name (example: "backend")
- `dockerfile` (string) - path to Dockerfile relative to workdir (producer MUST set explicitly; typical value: "Dockerfile")
- `context` (string) - build context path relative to workdir (producer MUST set explicitly; typical value: ".")
- `tags` ([]string) - image tags to apply; MUST be sorted
- `build_args` ([]BuildArg) - MUST be sorted by `key`
- `labels` ([]BuildLabel) - MUST be sorted by `key`

**Types:**
- `BuildArg`:
  - `key` (string, required)
  - `value` (string, required)
- `BuildLabel`:
  - `key` (string, required)
  - `value` (string, required)

**Determinism:**
- `tags` MUST be sorted.
- `build_args` MUST be sorted by `key`.
- `labels` MUST be sorted by `key`.

**Example:**
```json
{
  "provider": "generic",
  "workdir": "apps/backend",
  "dockerfile": "Dockerfile",
  "context": ".",
  "tags": ["stagecraft/backend:prod", "stagecraft/backend:sha-abc123"],
  "build_args": [
    {"key": "NODE_ENV", "value": "production"}
  ]
}
```

---

## Action: render_compose (`StepActionRenderCompose`)

**Purpose:**
Render a final docker-compose YAML for a specific host, producing a file artifact.

**Unknown-field behavior:** reject

### Inputs Schema (v1)

**Required:**
- `environment` (string) - environment name
- `output_path` (string) - where the rendered compose file will be written (relative path)

**One of the following MUST be provided:**
- `base_compose_path` (string) - path to base compose YAML, or
- `base_compose_inline` (string) - inline YAML contents

**Optional:**
- `overlays` ([]ComposeOverlay) - host or environment overlays; MUST be sorted by name
- `variables` ([]ComposeVar) - variable substitutions; MUST be sorted by key
- `expected_compose_hash_alg` (string) - if present, MUST be "sha256" in v1
- `expected_compose_hash` (string) - expected hash of rendered output (hex lowercase, 64 chars)

**Types:**
- `ComposeOverlay`:
  - `name` (string, required) - stable name
  - `path` (string, required) - overlay YAML path relative to execution root
- `ComposeVar`:
  - `key` (string, required)
  - `value` (string, required)

**Determinism:**
- `overlays` MUST be sorted by `name`.
- `variables` MUST be sorted by `key`.
- If `expected_compose_hash` is provided, it MUST be computed from the rendered file bytes.

**Example:**
```json
{
  "environment": "prod",
  "base_compose_path": "deploy/compose/base.yml",
  "overlays": [
    {"name": "host-local", "path": "deploy/compose/overlays/local.yml"}
  ],
  "variables": [
    {"key": "APP_ENV", "value": "prod"}
  ],
  "output_path": ".stagecraft/rendered/compose.prod.local.yml",
  "expected_compose_hash_alg": "sha256",
  "expected_compose_hash": "a3b2c1d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2"
}
```

---

## Action: apply_compose (`StepActionApplyCompose`)

**Purpose:**
Apply a compose file on the target host (up or update).

**Unknown-field behavior:** reject

### Inputs Schema (v1)

**Required:**
- `environment` (string)
- `compose_path` (string) - path to the compose YAML to apply (relative path)
- `project_name` (string) - stable docker compose project name

**Optional:**
- `pull` (bool) - whether to pull images first (producer MUST set explicitly; typical value: false)
- `detach` (bool) - run detached (producer MUST set explicitly; typical value: true)
- `services` ([]string) - if set, only apply these services; MUST be sorted
- `expected_compose_hash_alg` (string) - if present, MUST be "sha256" in v1
- `expected_compose_hash` (string) - expected hash of compose file contents (hex lowercase, 64 chars)

**Determinism:**
- `services` MUST be sorted.
- If `expected_compose_hash` is provided, executor MUST verify compose file bytes hash before apply.

**Example:**
```json
{
  "environment": "prod",
  "compose_path": ".stagecraft/rendered/compose.prod.local.yml",
  "project_name": "test-app-prod",
  "pull": true,
  "detach": true,
  "services": ["api", "web"],
  "expected_compose_hash_alg": "sha256",
  "expected_compose_hash": "a3b2c1d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2"
}
```

---

## Action: migrate (`StepActionMigrate`)

**Purpose:**
Run database migrations.

**Unknown-field behavior:** reject

### Inputs Schema (v1)

**Required:**
- `database` (string) - logical database name (example: "main")
- `strategy` (string) - migration strategy (example: "pre_deploy", "post_deploy")
- `engine` (string) - migration engine identifier (example: "raw")
- `path` (string) - migrations path relative to execution root
- `conn_env` (string) - env var name containing connection string (example: "DATABASE_URL")

**Optional:**
- `timeout_seconds` (int) - must be > 0 if present
- `args` ([]string) - engine-specific args; order is significant (do not sort)

**Determinism:**
- No set-like lists in v1, so nothing to sort.
- If `args` is used, it MUST be emitted deterministically by producer.

**Example:**
```json
{
  "database": "main",
  "strategy": "pre_deploy",
  "engine": "raw",
  "path": "./migrations",
  "conn_env": "DATABASE_URL",
  "timeout_seconds": 600
}
```

---

## Action: health_check (`StepActionHealthCheck`)

**Purpose:**
Verify system health after deploy.

**Unknown-field behavior:** reject

### Inputs Schema (v1)

**Required:**
- `environment` (string)

**One of the following MUST be provided:**
- `endpoints` ([]HealthEndpoint), or
- `services` ([]string) - interpreted by executor as default service health checks

**Optional:**
- `timeout_seconds` (int) - total timeout; must be > 0 if present
- `interval_seconds` (int) - polling interval; must be > 0 if present
- `retries` (int) - must be >= 0 if present

**Types:**
- `HealthEndpoint`:
  - `name` (string, required) - stable name
  - `url` (string, required)
  - `expected_status` (int, required) - HTTP status code
  - `method` (string, optional) - HTTP method (producer MUST set explicitly; typical value: "GET")
  - `headers` ([]HeaderKV, optional) - MUST be sorted by key
- `HeaderKV`:
  - `key` (string, required)
  - `value` (string, required)

**Determinism:**
- `endpoints` MUST be sorted by `name`.
- `services` MUST be sorted.
- `headers` MUST be sorted by `key`.

**Example:**
```json
{
  "environment": "prod",
  "timeout_seconds": 120,
  "interval_seconds": 5,
  "endpoints": [
    {"name": "api-health", "url": "http://localhost:8080/health", "expected_status": 200, "method": "GET"}
  ]
}
```

---

## Action: rollout (`StepActionRollout`)

**Purpose:**
Perform a rollout orchestration step. In v1, this is a placeholder action for future multi-host or progressive delivery semantics.

**Unknown-field behavior:** reject

### Inputs Schema (v1)

**Required:**
- `mode` (string) - rollout mode (example: "serial", "parallel")

**Optional:**
- `batch_size` (int) - if mode implies batching; must be > 0 if present
- `targets` ([]string) - stable target identifiers; MUST be sorted

**Determinism:**
- `targets` MUST be sorted.

**Example:**
```json
{
  "mode": "serial",
  "targets": ["host-a", "host-b"]
}
```

---

## Non-Goals (v1)

- No timestamps in Inputs.
- No dynamic or free-form JSON.
- No implicit defaults that change behavior across versions without schema changes.

---

## Notes for Implementation

- Implement typed structs per action under `pkg/engine/inputs/`.
- Each struct should implement:
  - `Normalize()` (optional) that sorts set-like fields deterministically
  - `Validate() error` - validates required fields and constraints
- **Normalization order**: `Normalize()` MUST be called before `Validate()`, and both MUST be called before marshaling to JSON.
- `marshalOperationInputs` should:
  - create the correct typed struct
  - call `Normalize()`
  - call `Validate()`
  - marshal to JSON
- Consumers MUST use strict JSON decoding with `DisallowUnknownFields()` to enforce unknown-field rejection.

---

## Related Features

- `CORE_PLAN` - Deployment planning engine
- `DEPLOY_COMPOSE_GEN` - Per-host Compose generation (uses `render_compose` and `apply_compose`)
- `DEPLOY_ROLLOUT` - docker-rollout integration (uses `rollout`)
- `MIGRATION_PRE_DEPLOY` / `MIGRATION_POST_DEPLOY` - Migration execution (uses `migrate`)


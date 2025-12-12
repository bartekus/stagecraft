---
feature: DEPLOY_COMPOSE_GEN
version: v1
status: done
domain: deploy
inputs:
  flags: []
outputs:
  exit_codes: {}
---
# DEPLOY_COMPOSE_GEN - Per-Host Compose Generation

- **Feature ID**: `DEPLOY_COMPOSE_GEN`
- **Domain**: `deploy`
- **Status**: `done`
- **Dependencies**: `CLI_DEPLOY`, `CORE_COMPOSE`

---

## 1. Purpose

Generate per-host Docker Compose files from the canonical `docker-compose.yml`, applying environment-specific overrides, image tags, and deterministic output.

**v1 scope**: Single-host only (no host parameter; generate once per environment).

---

## 2. Scope

### In Scope (v1)

- Generate compose file for environment (single-host)
- Image tag injection (always sets image tag, even if service has `build:` configuration)
- Environment variable injection via `env_file` (parse and merge into service `environment:` maps)
- Deterministic output (services, networks, volumes sorted lexicographically; environment map keys sorted)
- Artifact storage: `.stagecraft/rendered/<env>/docker-compose.yml`
- Hash computation: SHA256 of exact rendered bytes

### Explicitly Not Supported (v1)

- Multi-host service filtering (Phase 7/8)
- Port/volume overrides (unless already in config model)
- Complex overlay system
- Health check configuration

---

## 3. Inputs and Outputs

### Inputs

- `stagecraft.yml` configuration
- Base `docker-compose.yml` file (at project root)
- Environment name
- Built image tag (from build phase)
- Workdir (project root)

### Outputs

- Generated compose file path: `.stagecraft/rendered/<env>/docker-compose.yml`
- SHA256 hash of rendered bytes

---

## 4. Behavior

### Image Tag Injection

- Always sets `image` field for each service
- Forces Stagecraft's built tag for deployment
- Even if service has `build:` configuration, `image` is set

### Environment Variable Injection

- If `cfg.Environments[env].EnvFile` is set:
  - Parse env file using dotenv format (reuse `parseEnvFileInto` helper)
  - Merge parsed variables into each service's `environment:` map
  - Precedence: existing service environment vars win over env_file variables
  - Missing env file: no error (graceful, logs debug and continues)
- Env file path resolution: relative to workdir (project root)

### Determinism Guarantees

- Hash computed from exact rendered bytes
- Environment map keys sorted lexicographically
- Same inputs produce identical output
- All compose file fields preserved during mutation (version, services, networks, volumes, configs, secrets, x-*)

---

## 5. Integration

- Uses `ComposeFile.Mutate()` to safely mutate compose data
- Uses `ComposeFile.ToYAML()` for deterministic marshaling
- Integrated into `CLI_DEPLOY` rollout phase

---

## 6. Related Features

- `CORE_COMPOSE` - Compose file parsing and manipulation
- `CLI_DEPLOY` - Deployment command that uses generated compose files
- `DEPLOY_ROLLOUT` - Uses generated compose files for rollout


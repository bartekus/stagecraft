---
feature: DEPLOY_ROLLOUT
version: v1
status: done
domain: deploy
inputs:
  flags: []
outputs:
  exit_codes: {}
---
# DEPLOY_ROLLOUT - docker-rollout Integration

- **Feature ID**: `DEPLOY_ROLLOUT`
- **Domain**: `deploy`
- **Status**: `done`
- **Dependencies**: `CLI_DEPLOY`, `CORE_COMPOSE`, `DEPLOY_COMPOSE_GEN`

---

## 1. Purpose

Replace basic `docker compose up` with docker-rollout for zero-downtime deployments.

**v1 scope**: Opt-in via config flag, serial mode only, health checks deferred to separate feature.

---

## 2. Scope

### In Scope (v1)

- Opt-in via `environments.<env>.rollout.enabled` config flag
- Capability detection (`docker-rollout --version`)
- Zero-downtime deployments with docker-rollout when enabled
- Fallback to `docker compose up` when disabled
- Clear error when tool missing but rollout enabled

### Explicitly Not Supported (v1)

- Parallel mode (serial only)
- Automatic rollback on health check failure
- Health check integration (separate feature)
- Config-driven rollout modes

---

## 3. Compatibility Matrix

**v1 rules:**

- If `docker-rollout` present AND rollout enabled → use docker-rollout
- If `docker-rollout` absent AND rollout enabled → hard error with actionable install hint
- If rollout not enabled → fallback to `docker compose up` (current behavior)

---

## 4. Configuration

```yaml
environments:
  prod:
    rollout:
      enabled: true  # opt-in flag
```

**v1 config schema:**
- `rollout.enabled` (bool) - Opt-in flag
- Mode, health checks deferred to v2

---

## 5. Error Handling

- Context cancellation: return error (not false)
- Command not found: return `(false, nil)` (not available)
- Non-zero exit: return `(false, nil)` (not available)
- Execution failure: return wrapped error

**Error message format:**
- No raw URLs in error text
- "docker-rollout is required but not installed; install it from the docker-rollout repository"

---

## 6. Integration

- Uses `DEPLOY_COMPOSE_GEN` generated compose files
- Integrated into `CLI_DEPLOY` rollout phase
- Health checks deferred to separate feature (v1: deploy fails if rollout fails)

---

## 7. Related Features

- `DEPLOY_COMPOSE_GEN` - Generates compose files used by rollout
- `CLI_DEPLOY` - Deployment command that orchestrates rollout


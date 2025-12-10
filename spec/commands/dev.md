---
feature: CLI_DEV
version: v1
status: done
domain: commands
---

# CLI_DEV

Full `stagecraft dev` command for running a complete local development topology:

- Backend services
- Frontend(s)
- Supporting infrastructure (Traefik, mkcert, hosts, etc.)
- Process lifecycle management

## Overview

`stagecraft dev` orchestrates all local runtime components for a multi service application:

- Reads Stagecraft config and provider configuration
- Plans the required dev topology
- Brings up services and infrastructure in a deterministic order
- Streams logs and surfaces errors in a consistent way
- Cleans up on exit where possible

CLI_DEV builds on CLI_DEV_BASIC and integrates Phase 3 features:

- DEV_COMPOSE_INFRA
- DEV_TRAEFIK
- DEV_MKCERT
- DEV_HOSTS
- DEV_PROCESS_MGMT
- PROVIDER_FRONTEND_GENERIC

## Scope - v1

### Included

- Single host local development using Docker Compose
- Orchestration of:
  - Backend provider (generic or encore-ts)
  - Frontend provider (PROVIDER_FRONTEND_GENERIC)
  - Local Traefik reverse proxy
  - Local HTTPS via mkcert certificates
  - Hosts file management for dev domains
- Deterministic process lifecycle:
  - Start all components
  - Tear them down on interruption (Ctrl+C) where possible

### Excluded (future)

- Multi host dev environments
- Remote dev containers
- Kubernetes based dev environments
- Hot reload integration beyond what providers already support

## Command

```text
stagecraft dev [flags]
```

### Flags

- `--env string` - environment name to use (default: dev)
- `--config string` - path to config file (optional, defaults to standard search paths)
- `--no-https` - disable mkcert and HTTPS integration
- `--no-hosts` - do not modify /etc/hosts (user opts into manual DNS)
- `--no-traefik` - skip Traefik setup (providers must expose their own ports)
- `--detach` - run dev processes in background, return immediately
- `--verbose` - enable verbose output

All flags must be documented in `internal/cli/commands/dev.go` help text and kept lexicographically sorted.

## Behaviour

### High-level flow

1. Load configuration via `pkg/config`.
2. Resolve environment and providers.
3. Compute dev topology:
   - Backend services
   - Frontend services
   - Shared dev infrastructure (Traefik, mkcert, hosts entries)
4. Delegate infra details to:
   - DEV_COMPOSE_INFRA for compose files
   - DEV_TRAEFIK for router configuration
   - DEV_MKCERT for certificates
   - DEV_HOSTS for host name mapping
   - DEV_PROCESS_MGMT for running and monitoring processes
5. Start processes in a deterministic order:
   1. Infrastructure (Traefik, certificates, etc.)
   2. Backends
   3. Frontends
6. Block until interruption or failure, then tear down according to DEV_PROCESS_MGMT rules.

### Configuration sources

- Stagecraft config file(s) via `pkg/config`.
- Provider specific config remains opaque to CLI_DEV and is handled by providers.

### Dev Domains

**v1 Behavior:**

Dev domains are computed from the Stagecraft config with deterministic defaults.

- If `dev.domains.frontend` and/or `dev.domains.backend` are set in `stagecraft.yml`, those values are used.
- If they are not set, the following defaults are used:
  - Frontend domain: `app.localdev.test`
  - Backend domain: `api.localdev.test`

These domains are used for:
- mkcert certificate generation (when `--no-https` is not set)
- Traefik router rules (when Traefik is enabled)

**Configuration Example:**

```yaml
dev:
  domains:
    frontend: app.example.test
    backend: api.example.test
```

Both `frontend` and `backend` keys are optional. If a key is missing or empty, the corresponding default is used.

**Future Enhancement:**

A future slice may introduce environment-specific domains (e.g., `environments[env].dev.domains.*`).

### Hosts File Management

**v1 Behavior:**

DEV_HOSTS manages hosts file entries for dev domains:

- When `--no-hosts` is not set:
  - DEV_HOSTS automatically adds dev domain entries to `/etc/hosts` (or platform-equivalent)
  - Entries point to `127.0.0.1` and are marked as Stagecraft-managed
  - Entries are automatically removed when `stagecraft dev` exits

- When `--no-hosts` is set:
  - Hosts file modification is skipped (user manages DNS manually)
  - No entries are added or removed

## Exit Codes

See governance CLI exit code spec (GOV_CLI_EXIT_CODES) for exact values. For v1, CLI_DEV MUST:

- Return success when all components start successfully and terminate cleanly.
- Return a non-zero exit code when:
  - Configuration is invalid
  - Providers cannot be initialised
  - Any required process fails to start
  - Any required process exits unexpectedly

Exact mapping of error conditions to exit codes will be aligned with GOV_CLI_EXIT_CODES.

## Determinism

- Dev topology computation must be deterministic for the same inputs.
- Process start order must be stable.
- Any lists or maps must be processed in lexicographical order.
- No random ports, names, or identifiers may be used unless explicitly configured.

## Provider boundaries

- CLI_DEV must not interpret provider specific config.
- All provider logic goes through provider registries and interfaces.
- DEV_COMPOSE_INFRA and DEV_TRAEFIK own details of compose and routing.

## Tests

Minimum required tests:

- Unit tests in `internal/cli/commands/dev_test.go` for:
  - Flag parsing
  - Error paths (invalid config, missing providers)
- Integration or e2e tests:
  - `test/e2e/dev_smoke_test.go` for a minimal dev environment
  - Future tests for HTTPS and hosts behaviour

All tests must be deterministic and not depend on real network access.


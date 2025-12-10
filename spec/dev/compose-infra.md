---
feature: DEV_COMPOSE_INFRA
version: v1
status: todo
domain: dev
---

# DEV_COMPOSE_INFRA

Generate and manage Docker Compose infrastructure for `stagecraft dev`.

## Overview

DEV_COMPOSE_INFRA is responsible for:

- Synthesising Docker Compose configuration for local development.
- Combining backend, frontend, and shared infra containers.
- Producing deterministic compose definitions that can be used by `docker compose` or a compatible engine.

It is consumed by CLI_DEV and does not define a CLI surface itself.

## Scope - v1

### Included

- Generate a single deterministic compose model for:
  - Backend services (with real ports from providers)
  - Frontend services (with real ports from providers)
  - Traefik service with:
    - Image: `traefik:v2.11`
    - Ports: `80:80`, `443:443`
    - Volume mounts:
      - `.stagecraft/dev/certs` → `/certs` (read-only, for mkcert certificates)
      - `.stagecraft/dev/traefik` → `/etc/traefik` (read-only, for Traefik config files)
    - Command: File provider flags pointing to mounted config
- Create shared network: `stagecraft-dev` (all services join this network)
- Enforce deterministic service ordering (lexicographic)
- Support bind mounts and volume definitions as specified in config
- Support port mappings for local dev
- Support environment variables and secrets via provider mechanisms

### Excluded (future)

- Multi compose file layering (override stacks).
- Non Docker runtimes.
- Multi host compose swarms.

## Inputs

- Stagecraft config via `pkg/config`.
- Provider resolved service definitions from:
  - Backend provider
  - Frontend provider
  - DEV_TRAEFIK (for router sidecar)

## Outputs

- In memory compose model type under `internal/compose` (exact type to be re used from CORE_COMPOSE).
- Optional on disk compose file in `.stagecraft/dev/compose.yaml` when explicitly requested by spec.

File generation must be deterministic:

- Stable key ordering
- Stable service ordering
- Stable volume and network sections

## Behaviour

- Merge services from providers into a single compose model.
- **Traefik service generation**: When Traefik is included, generate complete service definition with:
  - Image: `traefik:v2.11`
  - Ports: `80:80`, `443:443`
  - Volumes: Mounts for `.stagecraft/dev/certs` and `.stagecraft/dev/traefik`
  - Command: File provider configuration
  - Network: `stagecraft-dev`
- **Network creation**: Always create `stagecraft-dev` network and ensure all services join it.
- Enforce deterministic names for:
  - Networks: `stagecraft-dev` (fixed for v1)
  - Volumes: Deterministic naming from config or service names
  - Service names: From provider definitions
- **Service ordering**: Services sorted lexicographically (`backend`, `frontend`, `traefik`)
- Do not talk to Docker directly; execution is handled by DEV_PROCESS_MGMT or equivalent.

## Determinism

- Compose model generation must produce identical output for identical inputs.
- **Service ordering**: Services sorted lexicographically by name (`backend`, `frontend`, `traefik`)
- **Port mapping ordering**: Ports sorted by host port (numeric, then lexicographic)
- **Volume mapping ordering**: Volumes sorted by target path (lexicographic)
- **Environment variable ordering**: Environment variables sorted by key (lexicographic)
- **Network ordering**: Networks sorted lexicographically
- **Label ordering**: Labels sorted by key (lexicographic)
- **Dependency ordering**: Dependencies sorted lexicographically
- All maps must be sorted before encoding to YAML.

## Tests

- Unit tests in `internal/compose/` (or `internal/dev/compose_*`) for:
  - Service merging behaviour
  - Deterministic output ordering
- Golden tests for generated compose YAML in `internal/compose/testdata/dev_compose_*.yaml`.


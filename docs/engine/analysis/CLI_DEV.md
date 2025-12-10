# CLI_DEV Feature Analysis Brief

This document captures the high level motivation, constraints, and success definition for CLI_DEV.

It is the starting point for the Implementation Outline and Spec.

This brief must be approved before outline work begins.

⸻

## 1. Problem Statement

Developers need a single command that orchestrates a complete local development environment for multi-service applications. Currently, `stagecraft dev` (CLI_DEV_BASIC) only handles backend services. Developers must manually:

- Start frontend services separately
- Configure reverse proxies (Traefik) for routing
- Set up local HTTPS certificates (mkcert)
- Manage hosts file entries for dev domains
- Coordinate process lifecycle across multiple services

This fragmentation creates friction, reduces determinism, and makes it harder to onboard new developers or demonstrate the full Stagecraft stack.

⸻

## 2. Motivation

### Developer Experience

- **Single command simplicity**: `stagecraft dev` should bring up the entire local stack (backend + frontend + infrastructure) in one deterministic operation
- **Consistent local environment**: Every developer gets the same topology, routing, and HTTPS setup
- **Faster onboarding**: New team members can start developing immediately without manual infrastructure setup
- **Portfolio demonstration**: A working `stagecraft dev` showcases the full orchestration capabilities of Stagecraft

### Operational Reliability

- **Deterministic topology**: Same config produces identical dev environments across machines
- **Clean lifecycle management**: Proper startup order and graceful shutdown on interruption
- **Error visibility**: Centralized logging and error surfacing for all components

### CI Workflows

- **E2E test environments**: CI can use `stagecraft dev` to spin up complete test environments
- **Deterministic test setup**: Same dev topology in CI as locally

### Provider Ecosystems

- **Provider integration showcase**: Demonstrates how backend, frontend, and infrastructure providers work together
- **Extension point**: Establishes patterns for future infrastructure providers (Kubernetes, remote dev containers, etc.)

⸻

## 3. Users and User Stories

### Developers

- As a developer, I want to run `stagecraft dev` and have my entire application stack (backend, frontend, Traefik, HTTPS) start automatically, so I can focus on writing code instead of managing infrastructure
- As a developer, I want my local dev environment to match production routing patterns (HTTPS, custom domains), so I can catch routing issues early
- As a developer, I want `stagecraft dev` to clean up all processes when I press Ctrl+C, so I don't leave orphaned containers or processes running

### Platform Engineers

- As a platform engineer, I want `stagecraft dev` to produce deterministic compose files and Traefik configs, so I can review and version control the generated infrastructure
- As a platform engineer, I want the dev topology to be computed from config, not hardcoded, so I can customize it per project

### Automation and CI

- As a CI pipeline, I want `stagecraft dev --detach` to start the full stack in the background, so I can run e2e tests against a complete environment
- As a CI pipeline, I want deterministic dev topology generation, so test environments are reproducible across runs

⸻

## 4. Success Criteria (v1)

1. **Single command orchestration**: Running `stagecraft dev` starts backend, frontend, Traefik, mkcert, and hosts management in a deterministic order

2. **Deterministic topology computation**: The same `stagecraft.yml` config produces identical compose files, Traefik configs, and service definitions across runs and machines

3. **Provider integration**: CLI_DEV successfully orchestrates:
   - Backend provider (generic or encore-ts) via existing provider interfaces
   - Frontend provider (PROVIDER_FRONTEND_GENERIC) via existing provider interfaces
   - Infrastructure components (Traefik, mkcert, hosts) via new dev domain features

4. **Process lifecycle**: All components start in correct order (infra → backends → frontends) and tear down gracefully on interruption (Ctrl+C)

5. **Error handling**: Invalid config, missing providers, or process failures return appropriate exit codes and clear error messages

6. **E2E smoke test**: A minimal test fixture (backend + frontend) successfully runs `stagecraft dev` and verifies all services are accessible

7. **No breaking changes**: CLI_DEV_BASIC behavior is preserved; CLI_DEV extends it without breaking existing workflows

## 4.1. v1 Implementation Status and Limitations

**Current Implementation (Integration Slice Complete):**

- ✅ Provider discovery and service definition extraction from config
- ✅ DEV_COMPOSE_INFRA integration (compose file generation)
- ✅ DEV_TRAEFIK integration (routing configuration)
- ✅ DEV_MKCERT integration (HTTPS certificate generation)
- ✅ DEV_PROCESS_MGMT integration (process lifecycle)
- ✅ DEV_HOSTS integration (hosts file management)
- ✅ Flag handling (`--no-https`, `--no-hosts`, `--no-traefik`, `--detach`, `--verbose`)
- ✅ Backend-only and backend+frontend topologies supported
- ✅ Domain computation: Dev domains are computed from `dev.domains.*` in config with deterministic defaults (`app.localdev.test` / `api.localdev.test`), exactly as per the spec

**v1 Limitations (Future Slices):**

- ⏸ **Service definition richness**: Current extraction uses environment variables and PORT only. Future slices will extract image, build, volumes, etc. from provider config.
- ⏸ **Provider-specific service shapes**: Current extraction assumes generic provider config structure. Encore.ts and other providers may need provider-specific hooks in future slices.
- ⏸ **Environment-specific domains**: Domain computation currently uses top-level `dev.domains.*` only. Future slice may introduce environment-specific domain overrides (e.g., `environments[env].dev.domains.*`).

⸻

## 5. Risks and Constraints

### Determinism Constraints

- **Compose file generation must be stable**: Service ordering, network names, volume names must be lexicographically sorted
- **Traefik config must be stable**: Router definitions, service mappings must be deterministically ordered
- **No random ports or identifiers**: All ports, service names, network names must come from config or use deterministic defaults
- **No timestamps**: Generated configs must not include timestamps or machine-specific paths

### Provider Constraints

- **Provider interfaces must remain stable**: CLI_DEV must not modify existing provider interfaces (BackendProvider, FrontendProvider)
- **Provider config remains opaque**: CLI_DEV must not interpret provider-specific config; all provider logic goes through provider registries
- **Backward compatibility**: CLI_DEV_BASIC must continue to work; CLI_DEV should be an extension, not a replacement

### Architectural Constraints

- **Reuse CORE_COMPOSE**: DEV_COMPOSE_INFRA must reuse existing `internal/compose` types and patterns
- **No circular dependencies**: CLI_DEV → DEV_COMPOSE_INFRA → providers (not CLI_DEV → providers → CLI_DEV)
- **Separation of concerns**: 
  - CLI_DEV orchestrates and coordinates
  - DEV_COMPOSE_INFRA generates compose models
  - DEV_TRAEFIK generates routing configs
  - DEV_PROCESS_MGMT handles process execution (future feature, may be stubbed in v1)

### Implementation Constraints

- **v1 scope is single-host**: Multi-host dev environments are explicitly out of scope
- **Docker Compose only**: Non-Docker runtimes (Podman, containerd) are out of scope for v1
- **No hot reload orchestration**: Providers handle their own hot reload; CLI_DEV doesn't add additional hot reload logic

⸻

## 6. Alternatives Considered

### Alternative 1: Extend CLI_DEV_BASIC in place

**Rejected because**: CLI_DEV_BASIC is already done and working. Extending it would risk breaking existing workflows. Better to build CLI_DEV as a new command that can eventually replace CLI_DEV_BASIC after validation.

### Alternative 2: Use external tools (docker-compose, traefik CLI)

**Rejected because**: Stagecraft's value proposition is deterministic orchestration from config. External tools introduce non-determinism and reduce portability. CLI_DEV should generate configs and delegate execution, but own the generation logic.

### Alternative 3: Hardcode Traefik and mkcert setup

**Rejected because**: Hardcoding reduces flexibility and makes testing harder. DEV_TRAEFIK and DEV_MKCERT should be configurable and testable in isolation.

⸻

## 7. Dependencies

### Required Features (all done)

- **CLI_DEV_BASIC**: Provides base command structure and backend provider integration
- **CORE_CONFIG**: Config loading and validation
- **CORE_COMPOSE**: Compose file loading and manipulation (reused by DEV_COMPOSE_INFRA)
- **PROVIDER_BACKEND_GENERIC**: Generic backend provider (or PROVIDER_BACKEND_ENCORE)
- **PROVIDER_FRONTEND_GENERIC**: Generic frontend provider
- **CORE_BACKEND_REGISTRY**: Backend provider registry
- **CORE_FRONTEND_REGISTRY**: Frontend provider registry (assumed to exist, similar to backend)

### New Features (to be implemented)

- **DEV_COMPOSE_INFRA**: Compose model generation for dev environments
- **DEV_TRAEFIK**: Traefik configuration generation
- **DEV_MKCERT**: Local certificate generation (may be stubbed in v1)
- **DEV_HOSTS**: Hosts file management (may be stubbed in v1)
- **DEV_PROCESS_MGMT**: Process lifecycle management (may be stubbed in v1, using basic exec for now)

### Spec Dependencies

- `spec/commands/dev.md` (CLI_DEV spec) - already created
- `spec/dev/compose-infra.md` (DEV_COMPOSE_INFRA spec) - already created
- `spec/dev/traefik.md` (DEV_TRAEFIK spec) - already created

⸻

## 8. Approval

- Author: [To be filled]
- Reviewer: [To be filled]
- Date: [To be filled]

Once approved, the Implementation Outline may begin.


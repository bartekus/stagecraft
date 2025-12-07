# Stagecraft Implementation Roadmap

> **Purpose**: This document captures all features discussed across the project's design documents and organizes them into a progressive implementation plan that maintains our core development practices (spec-first, TDD, ADR-driven).

> **Last Updated**: Generated from analysis of blog posts and discussion documents

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [V1 Scope Definition](#v1-scope-definition)
3. [Development Principles](#development-principles)
4. [Feature Organization](#feature-organization)
5. [Implementation Phases](#implementation-phases)
6. [Feature Dependency Graph](#feature-dependency-graph)
7. [Implementation Workflow](#implementation-workflow)
8. [Feature Catalog](#feature-catalog)

---

## Quick Start

If you're new to the project or starting implementation, here's the recommended path:

### First 5 Features to Implement (Phase 0)

These form the foundation for everything else:

1. **`CORE_CONFIG`** (Priority: Critical)
   - Establishes the config system
   - Enables all other features
   - Estimated effort: Medium
   - Dependencies: None

2. **`CORE_LOGGING`** (Priority: High)
   - Needed for all commands
   - Simple, self-contained
   - Estimated effort: Low
   - Dependencies: None

3. **`CORE_EXECUTIL`** (Priority: High)
   - Needed for running external commands
   - Estimated effort: Low
   - Dependencies: `CORE_LOGGING`

4. **`CLI_GLOBAL_FLAGS`** (Priority: High)
   - Enables consistent CLI behavior
   - Estimated effort: Low
   - Dependencies: `CORE_CONFIG`

5. **`CLI_INIT`** (Priority: High)
   - First user-facing command
   - Validates the config system
   - Estimated effort: Medium
   - Dependencies: `CORE_CONFIG`

### Critical Path Features

After Phase 0, focus on these in order:

1. **Provider Interfaces** (Phase 1) - Foundation for all providers
2. **Core Orchestration** (Phase 2) - Planning and state management
3. **Local Development** (Phase 3) - `stagecraft dev` command
4. **Provider Implementations** (Phase 4) - Tailscale and DigitalOcean
5. **Build and Deploy** (Phase 5) - Core deployment capabilities

### Getting Help

- See individual feature specs in `spec/` for detailed requirements
- Check ADRs in `docs/adr/` for architectural decisions
- Review `docs/stagecraft-spec.md` for full application specification
- Follow the [Implementation Workflow](#implementation-workflow) for each feature

---

## V1 Scope Definition

### What's In Scope for v1

v1 focuses on core functionality to achieve a working deployment tool:

**Core Capabilities:**
- ✅ Complete config system (`stagecraft.yml` with full schema)
- ✅ Local development (`stagecraft dev` with mkcert, Traefik, Encore, Vite)
- ✅ Build and deploy (`stagecraft build`, `stagecraft deploy`)
- ✅ Infrastructure provisioning (`stagecraft infra up/down` for DigitalOcean)
- ✅ Migration system (pre/post deploy migrations)
- ✅ Basic operations (`status`, `logs`, `ssh`, `rollback`)
- ✅ CI integration (GitHub Actions workflow generation)

**Provider Support:**
- Backend: Encore.ts (primary)
- Frontend: Generic dev commands (Vite, etc.)
- Network: Tailscale (primary), Headscale (optional)
- Cloud: DigitalOcean (primary)
- CI: GitHub Actions (primary)
- Secrets: Env files and Encore dev secrets

**Deployment Model:**
- Docker Compose-based orchestration
- docker-rollout for zero-downtime updates
- Multi-host via Tailscale mesh networking
- File-based state management (`.stagecraft/releases.json`)

### What's Explicitly Out of Scope for v1

**Advanced Features (v2):**
- ❌ Ephemeral environments
- ❌ Audit ledger and replay
- ❌ Infrastructure recipes
- ❌ Topology visualization
- ❌ AI test harness
- ❌ Advanced secrets orchestrator
- ❌ Health watchdog agent
- ❌ Sync primitives
- ❌ Composable pipelines
- ❌ Snapshot manager
- ❌ Editor plugins
- ❌ Multi-owner/organization support
- ❌ Observability stack
- ❌ Budget guardrails
- ❌ Migration preflight simulator

**Config Features:**
- ❌ Full environment variable interpolation (basic `${VAR}` for migrations only)
- ❌ Remote config loading
- ❌ Config file watching/reloading
- ❌ Advanced schema evolution

**Provider Extensions:**
- ❌ Additional cloud providers (AWS, GCP, etc.)
- ❌ Kubernetes support
- ❌ Additional CI providers
- ❌ Advanced secrets backends (Vault, Doppler, etc.)

**State Management:**
- ❌ Remote state backend (v1 uses local files)
- ❌ Distributed state synchronization
- ❌ State locking

### v1 Success Criteria

v1 is considered complete when:

1. ✅ `stagecraft init` creates a valid, complete `stagecraft.yml`
2. ✅ `stagecraft dev` spins up full local stack (infra + backend + frontend)
3. ✅ `stagecraft build` builds and pushes Docker images
4. ✅ `stagecraft deploy` deploys to staging/prod environments
5. ✅ `stagecraft infra up` provisions DigitalOcean infrastructure
6. ✅ Migrations run automatically in deployment pipeline
7. ✅ All core commands work end-to-end
8. ✅ Test coverage meets targets (80%+ core, 70%+ CLI/drivers)

### Transition to v2

v2 planning begins when:
- All v1 features are complete and tested
- v1 has been used in production for at least one project
- User feedback identifies clear v2 priorities
- Core architecture is stable and extensible

See [V2 Features (Deferred)](#v2-features-deferred) for the planned v2 feature list.

---

## Development Principles

All features must follow these principles (from `05-development-strategy.md`):

1. **Spec-First**: Every feature must have a spec in `spec/` before implementation
2. **Test-First**: Core logic uses TDD; CLI uses golden tests
3. **ADR-Driven**: Major architectural decisions require ADRs
4. **Feature Traceability**: Every change links to a feature ID in `spec/features.yaml`
5. **Quality Gates**: 80%+ coverage on core packages, 70%+ on CLI/drivers

### Workflow for Each Feature

```
1. Add feature to spec/features.yaml (status: todo)
2. Create/update spec in spec/
3. Write tests (TDD for core logic)
4. Implement feature
5. Update feature status (wip → done)
6. Update docs/implementation-status.md
```

---

## Feature Organization

Features are organized into **6 major categories**:

1. **Foundation** - Core infrastructure and configuration
2. **Providers** - Pluggable provider interfaces and implementations
3. **Local Development** - `stagecraft dev` and local orchestration
4. **Deployment** - Build, deploy, rollback workflows
5. **Infrastructure** - Cloud provisioning and management
6. **Operations** - Status, logs, SSH, secrets management

Each category has **dependencies** that must be respected.

---

## Implementation Phases

### Phase 0: Foundation ✅ COMPLETE

**Goal**: Complete the config system and basic CLI infrastructure

| Feature ID | Title | Status | Dependencies | Source |
|------------|-------|--------|--------------|--------|
| `CORE_CONFIG` | Config loading and validation | ✅ done | None | `05-development-strategy.md` |
| `CLI_INIT` | Project bootstrap command | ✅ done | `CORE_CONFIG` | `02-project-scaffold.md` |
| `CORE_LOGGING` | Structured logging helpers | ✅ done | None | `docs/stagecraft-spec.md` |
| `CORE_EXECUTIL` | Process execution utilities | ✅ done | `CORE_LOGGING` | `docs/stagecraft-spec.md` |
| `CLI_GLOBAL_FLAGS` | Global flags (--env, --config, --verbose, --dry-run) | ✅ done | `CORE_CONFIG` | `docs/stagecraft-spec.md` |
| `CORE_BACKEND_REGISTRY` | Backend provider registry system | ✅ done | None | `docs/stagecraft-spec.md` |
| `CORE_MIGRATION_REGISTRY` | Migration engine registry system | ✅ done | None | `docs/stagecraft-spec.md` |
| `CORE_BACKEND_PROVIDER_CONFIG_SCHEMA` | Provider-scoped backend configuration schema | ✅ done | `CORE_CONFIG` | `docs/stagecraft-spec.md` |

**Deliverables**:
- Complete `pkg/config` with full `stagecraft.yml` schema
- Working `stagecraft init` that creates valid configs
- Logging and exec utilities for all future commands
- Global flag handling across all commands

**Success Criteria**:
- `stagecraft init` creates a valid, complete `stagecraft.yml`
- Config validation catches all schema errors
- All tests pass with 80%+ coverage on `pkg/config`

---

### Phase 1: Provider Interfaces (Foundation for Everything) ✅ COMPLETE

**Goal**: Define all provider interfaces and create stub implementations

| Feature ID | Title | Status | Dependencies | Source |
|------------|-------|--------|--------------|--------|
| `PROVIDER_BACKEND_INTERFACE` | BackendProvider interface definition | ✅ done | `CORE_CONFIG` | `01-why-not-kamal.md`, `docs/stagecraft-spec.md` |
| `PROVIDER_FRONTEND_INTERFACE` | FrontendProvider interface definition | ✅ done | `CORE_CONFIG` | `01-why-not-kamal.md`, `docs/stagecraft-spec.md` |
| `PROVIDER_NETWORK_INTERFACE` | NetworkProvider interface definition | ✅ done | `CORE_CONFIG` | `01-why-not-kamal.md`, `docs/stagecraft-spec.md` |
| `PROVIDER_CLOUD_INTERFACE` | CloudProvider interface definition | ✅ done | `CORE_CONFIG` | `01-why-not-kamal.md`, `docs/stagecraft-spec.md` |
| `PROVIDER_CI_INTERFACE` | CIProvider interface definition | ✅ done | `CORE_CONFIG` | `01-why-not-kamal.md`, `docs/stagecraft-spec.md` |
| `PROVIDER_SECRETS_INTERFACE` | SecretsProvider interface definition | ✅ done | `CORE_CONFIG` | `docs/stagecraft-spec.md` |

**Deliverables**:
- All provider interfaces defined in `pkg/providers/` or `internal/providers/`
- Interface documentation and examples
- Mock implementations for testing

**Success Criteria**:
- All interfaces defined with clear contracts
- Interfaces are testable (mockable)
- ADR documenting provider architecture

---

### Phase 2: Core Orchestration Engine ✅ COMPLETE

**Goal**: Build the planning and orchestration engine

| Feature ID | Title | Status | Dependencies | Source |
|------------|-------|--------|--------------|--------|
| `CORE_PLAN` | Deployment planning engine | ✅ done | `CORE_CONFIG`, `PROVIDER_*_INTERFACE` | `05-development-strategy.md` |
| `CORE_ENV_RESOLUTION` | Environment resolution and context | ✅ done | `CORE_CONFIG` | `docs/stagecraft-spec.md` |
| `CORE_STATE` | State management (release history) | ✅ done | `CORE_CONFIG` | `03-migration-strategies.md` |
| `CORE_STATE_TEST_ISOLATION` | State test isolation for CLI commands | ✅ done | `CORE_STATE` | `docs/stagecraft-spec.md` |
| `CORE_STATE_CONSISTENCY` | State durability and read-after-write guarantees | ✅ done | `CORE_STATE` | `docs/stagecraft-spec.md` |
| `CORE_COMPOSE` | Docker Compose integration | ✅ done | `CORE_CONFIG` | `01-why-not-kamal.md`, `docs/stagecraft-spec.md` |
| `CLI_PHASE_EXECUTION_COMMON` | Shared phase execution semantics for deploy and rollback | ✅ done | `CORE_STATE`, `CORE_PLAN` | `docs/stagecraft-spec.md` |

**Deliverables**:
- `internal/core/plan.go` - Deployment planning logic
- `internal/core/env.go` - Environment resolution
- `internal/state/` - State backend (file-based v1)
- `internal/compose/` - Compose file handling

**Success Criteria**:
- Can generate deployment plans from config
- Can resolve environment-specific settings
- Can track release history in `.stagecraft/releases.json`
- Can parse and manipulate `docker-compose.yml`

---

### Phase 3: Local Development (`stagecraft dev`) - 3/10 Complete

**Goal**: Full local development experience

| Feature ID | Title | Status | Dependencies | Source |
|------------|-------|--------|--------------|--------|
| `CLI_DEV_BASIC` | Basic `stagecraft dev` command | ✅ done | `PROVIDER_BACKEND_INTERFACE` | `docs/stagecraft-spec.md` |
| `CLI_DEV` | Full `stagecraft dev` command | 🔄 todo | `CORE_PLAN`, `PROVIDER_BACKEND_INTERFACE`, `PROVIDER_FRONTEND_INTERFACE` | `docs/stagecraft-spec.md` |
| `DEV_MKCERT` | mkcert integration for local HTTPS | 🔄 todo | `CLI_DEV` | `01-why-not-kamal.md`, `docs/stagecraft-spec.md` |
| `DEV_HOSTS` | `/etc/hosts` management | 🔄 todo | `CLI_DEV` | `docs/stagecraft-spec.md` |
| `DEV_TRAEFIK` | Traefik dev config generation | 🔄 todo | `CLI_DEV`, `CORE_COMPOSE` | `docs/stagecraft-spec.md` |
| `DEV_COMPOSE_INFRA` | Compose infra up/down for dev | 🔄 todo | `CLI_DEV`, `CORE_COMPOSE` | `docs/stagecraft-spec.md` |
| `PROVIDER_BACKEND_ENCORE` | Encore.ts BackendProvider implementation | ✅ done | `PROVIDER_BACKEND_INTERFACE`, `CLI_DEV_BASIC` | `01-why-not-kamal.md`, `docs/stagecraft-spec.md` |
| `PROVIDER_BACKEND_GENERIC` | Generic command-based BackendProvider implementation | ✅ done | `PROVIDER_BACKEND_INTERFACE` | `docs/stagecraft-spec.md` |
| `PROVIDER_FRONTEND_GENERIC` | Generic dev command FrontendProvider | 🔄 todo | `PROVIDER_FRONTEND_INTERFACE`, `CLI_DEV` | `docs/stagecraft-spec.md` |
| `DEV_PROCESS_MGMT` | Process lifecycle management | 🔄 todo | `CLI_DEV`, `CORE_EXECUTIL` | `docs/stagecraft-spec.md` |

**Deliverables**:
- Working `stagecraft dev` command
- Local HTTPS with mkcert
- Traefik serving local domains
- Encore dev server integration
- Frontend dev server (Vite) integration
- Process management (start/stop/restart)

**Success Criteria**:
- `stagecraft dev` spins up full local stack
- HTTPS works on local domains
- Backend and frontend hot-reload
- Clean shutdown on Ctrl+C

---

### Phase 4: Provider Implementations (Core) - 0/3 Complete

**Goal**: Implement core provider implementations

| Feature ID | Title | Status | Dependencies | Source |
|------------|-------|--------|--------------|--------|
| `PROVIDER_NETWORK_TAILSCALE` | Tailscale NetworkProvider implementation | 🔄 todo | `PROVIDER_NETWORK_INTERFACE`, `CORE_PLAN` | `01-why-not-kamal.md`, `docs/stagecraft-spec.md` |
| `PROVIDER_CLOUD_DO` | DigitalOcean CloudProvider implementation | 🔄 todo | `PROVIDER_CLOUD_INTERFACE`, `CORE_PLAN` | `01-why-not-kamal.md`, `docs/stagecraft-spec.md` |
| `DRIVER_DO` | DigitalOcean driver (legacy name, may merge) | 🔄 todo | `PROVIDER_CLOUD_DO` | `05-development-strategy.md` |

**Deliverables**:
- Tailscale integration for mesh networking
- DigitalOcean API integration for infrastructure

**Success Criteria**:
- Can join hosts to Tailscale tailnet
- Can provision DO droplets
- Can bootstrap hosts (Docker, Tailscale, etc.)

---

### Phase 5: Build and Deploy - 4/6 Complete

**Goal**: Core deployment capabilities

| Feature ID | Title | Status | Dependencies | Source |
|------------|-------|--------|--------------|--------|
| `CLI_BUILD` | `stagecraft build` command | ✅ done | `PROVIDER_BACKEND_ENCORE`, `CORE_COMPOSE` | `docs/stagecraft-spec.md` |
| `CLI_DEPLOY` | `stagecraft deploy` command | ✅ done | `CORE_PLAN`, `PROVIDER_NETWORK_TAILSCALE`, `CORE_COMPOSE`, `CORE_STATE` | `docs/stagecraft-spec.md`, `05-development-strategy.md` |
| `CLI_ROLLBACK` | `stagecraft rollback` command | ✅ done | `CLI_DEPLOY`, `CORE_STATE` | `docs/stagecraft-spec.md` |
| `CLI_PLAN` | `stagecraft plan` command (dry-run) | ✅ done | `CORE_PLAN` | `05-development-strategy.md` |
| `DEPLOY_COMPOSE_GEN` | Per-host Compose generation | 🔄 todo | `CLI_DEPLOY`, `CORE_COMPOSE` | `docs/stagecraft-spec.md` |
| `DEPLOY_ROLLOUT` | docker-rollout integration | 🔄 todo | `CLI_DEPLOY`, `CORE_COMPOSE` | `01-why-not-kamal.md`, `docs/stagecraft-spec.md` |

**Deliverables**:
- `stagecraft build` builds and pushes images
- `stagecraft deploy` deploys to environments
- `stagecraft plan` shows deployment plan (dry-run)
- `stagecraft rollback` rolls back to previous version
- Per-host Compose file generation
- Zero-downtime deployments with docker-rollout

**Success Criteria**:
- Can build Docker images for backend/frontend
- Can deploy to staging/prod environments
- Can rollback deployments
- Can show deployment plan without executing

---

### Phase 6: Migration System - 3/9 Complete

**Goal**: First-class migration handling

| Feature ID | Title | Status | Dependencies | Source |
|------------|-------|--------|--------------|--------|
| `MIGRATION_ENGINE_RAW` | Raw SQL migration engine | ✅ done | `CORE_COMPOSE` | `03-migration-strategies.md` |
| `CLI_MIGRATE_BASIC` | Basic migrate command | ✅ done | `MIGRATION_ENGINE_RAW` | `03-migration-strategies.md` |
| `CLI_RELEASES` | `stagecraft releases list/show` commands | ✅ done | `CORE_STATE` | `03-migration-strategies.md` |
| `MIGRATION_CONFIG` | Migration config schema in stagecraft.yml | 🔄 todo | `CORE_CONFIG` | `03-migration-strategies.md` |
| `MIGRATION_INTERFACE` | Migrator interface | 🔄 todo | `CORE_PLAN` | `03-migration-strategies.md` |
| `MIGRATION_CONTAINER_RUNNER` | ContainerRunner interface | 🔄 todo | `CORE_COMPOSE` | `03-migration-strategies.md` |
| `MIGRATION_PRE_DEPLOY` | Pre-deploy migration execution | 🔄 todo | `CLI_DEPLOY`, `MIGRATION_INTERFACE` | `03-migration-strategies.md` |
| `MIGRATION_POST_DEPLOY` | Post-deploy migration execution | 🔄 todo | `CLI_DEPLOY`, `MIGRATION_INTERFACE` | `03-migration-strategies.md` |
| `CLI_MIGRATE_PLAN` | `stagecraft migrate plan` command | 🔄 todo | `MIGRATION_INTERFACE` | `03-migration-strategies.md` |
| `CLI_MIGRATE_RUN` | `stagecraft migrate run` command | 🔄 todo | `MIGRATION_INTERFACE` | `03-migration-strategies.md` |

**Deliverables**:
- Migration config in `stagecraft.yml`
- Migration execution in deployment pipeline
- Migration planning and manual execution commands
- Release history inspection

**Success Criteria**:
- Migrations run automatically in deployment pipeline
- Can plan and run migrations manually
- Release history tracks migration phases
- Supports multiple migration engines (Drizzle, Prisma, etc.)

---

### Phase 7: Infrastructure Management

**Goal**: Infrastructure provisioning and management

| Feature ID | Title | Status | Dependencies | Source |
|------------|-------|--------|--------------|--------|
| `CLI_INFRA_UP` | `stagecraft infra up` command | todo | `PROVIDER_CLOUD_DO`, `PROVIDER_NETWORK_TAILSCALE` | `docs/stagecraft-spec.md` |
| `CLI_INFRA_DOWN` | `stagecraft infra down` command | todo | `PROVIDER_CLOUD_DO` | `docs/stagecraft-spec.md` |
| `INFRA_HOST_BOOTSTRAP` | Host bootstrap (Docker, Tailscale, etc.) | todo | `CLI_INFRA_UP` | `docs/stagecraft-spec.md` |
| `INFRA_VOLUME_MGMT` | Volume management | todo | `CLI_INFRA_UP`, `PROVIDER_CLOUD_DO` | `docs/stagecraft-spec.md` |
| `INFRA_FIREWALL` | Firewall configuration | todo | `CLI_INFRA_UP`, `PROVIDER_CLOUD_DO` | `docs/stagecraft-spec.md` |

**Deliverables**:
- `stagecraft infra up` provisions infrastructure
- `stagecraft infra down` destroys infrastructure
- Automated host bootstrap
- Volume and firewall management

**Success Criteria**:
- Can provision complete infrastructure from config
- Hosts are bootstrapped and ready for deployment
- Can tear down infrastructure cleanly

---

### Phase 8: Operations Commands

**Goal**: Operational visibility and management

| Feature ID | Title | Status | Dependencies | Source |
|------------|-------|--------|--------------|--------|
| `CLI_STATUS` | `stagecraft status` command | todo | `CORE_PLAN`, `PROVIDER_NETWORK_TAILSCALE` | `docs/stagecraft-spec.md` |
| `CLI_LOGS` | `stagecraft logs` command | todo | `CORE_COMPOSE`, `PROVIDER_NETWORK_TAILSCALE` | `docs/stagecraft-spec.md` |
| `CLI_SSH` | `stagecraft ssh` command | todo | `PROVIDER_NETWORK_TAILSCALE` | `docs/stagecraft-spec.md` |
| `CLI_SECRETS_SYNC` | `stagecraft secrets sync` command | todo | `PROVIDER_SECRETS_INTERFACE` | `docs/stagecraft-spec.md` |
| `PROVIDER_SECRETS_ENVFILE` | Env file SecretsProvider | todo | `PROVIDER_SECRETS_INTERFACE` | `docs/stagecraft-spec.md` |
| `PROVIDER_SECRETS_ENCORE` | Encore dev secrets SecretsProvider | todo | `PROVIDER_SECRETS_INTERFACE`, `PROVIDER_BACKEND_ENCORE` | `docs/stagecraft-spec.md` |

**Deliverables**:
- `stagecraft status` shows environment status
- `stagecraft logs` tails service logs
- `stagecraft ssh` opens SSH sessions
- `stagecraft secrets sync` syncs secrets

**Success Criteria**:
- Can inspect environment status
- Can tail logs from remote services
- Can SSH into hosts via Tailscale
- Can sync secrets between environments

---

### Phase 9: CI Integration

**Goal**: CI/CD integration

| Feature ID | Title | Status | Dependencies | Source |
|------------|-------|--------|--------------|--------|
| `PROVIDER_CI_GITHUB` | GitHub Actions CIProvider | todo | `PROVIDER_CI_INTERFACE` | `01-why-not-kamal.md`, `docs/stagecraft-spec.md` |
| `CLI_CI_INIT` | `stagecraft ci init` command | todo | `PROVIDER_CI_GITHUB` | `docs/stagecraft-spec.md` |
| `CLI_CI_RUN` | `stagecraft ci run` command | todo | `PROVIDER_CI_GITHUB` | `docs/stagecraft-spec.md` |

**Deliverables**:
- GitHub Actions workflow generation
- CI trigger from CLI
- Secret management in CI

**Success Criteria**:
- Can generate GitHub Actions workflows
- Can trigger CI runs from CLI
- Secrets are managed in GitHub

---

### Phase 10: Project Scaffold (Advanced)

**Goal**: Enhanced project initialization

| Feature ID | Title | Status | Dependencies | Source |
|------------|-------|--------|--------------|--------|
| `CLI_INIT_TEMPLATE` | Template system for `stagecraft init` | todo | `CLI_INIT` | `02-project-scaffold.md` |
| `CLI_NEW` | `stagecraft new --template=platform` | todo | `CLI_INIT_TEMPLATE` | `02-project-scaffold.md` |
| `CLI_ATTACH` | `stagecraft attach` for existing projects | todo | `CLI_INIT` | `02-project-scaffold.md` |
| `TEMPLATE_PLATFORM` | Platform template (embedded) | todo | `CLI_INIT_TEMPLATE` | `02-project-scaffold.md` |
| `SCAFFOLD_STAGECRAFT_DIR` | `.stagecraft/` directory generation | todo | `CLI_INIT` | `02-project-scaffold.md` |

**Deliverables**:
- Template system for project generation
- Platform template with full stack
- Drop-in mode for existing projects
- `.stagecraft/` workspace directory

**Success Criteria**:
- Can generate new projects from templates
- Can add Stagecraft to existing projects
- Templates include full configuration

---

## Implementation Workflow

For each feature, follow this workflow:

### Step 1: Feature Planning
1. Add feature to `spec/features.yaml` with `status: todo`
2. Create spec document in `spec/` (e.g., `spec/commands/dev.md`)
3. If architectural decision needed, create ADR in `docs/adr/`
4. Update this roadmap with feature details

### Step 2: Test Design
1. Write test file (e.g., `internal/cli/commands/dev_test.go`)
2. For core logic: Write tests first (TDD)
3. For CLI: Design golden test outputs
4. Ensure tests are runnable (may fail initially)

### Step 3: Implementation
1. Create/update implementation files
2. Add feature ID comments to code:
   ```go
   // Feature: CLI_DEV
   // Spec: spec/commands/dev.md
   ```
3. Implement feature following spec
4. Run tests frequently

### Step 4: Validation
1. All tests pass
2. Coverage meets targets (80%+ core, 70%+ CLI/drivers)
3. Linting passes
4. Code review (self or peer)

### Step 5: Documentation
1. Update `docs/implementation-status.md`
2. Update feature status in `spec/features.yaml` (`todo` → `wip` → `done`)
3. Update relevant docs in `docs/`
4. Update this roadmap if needed

### Step 6: Integration
1. Ensure feature integrates with existing features
2. Update dependent features if needed
3. Test end-to-end workflows

---

## Progress Tracking

### Current Status Summary

- **Total Features Identified**: 72 features (including Architecture & Documentation, Governance)
- **Architecture & Documentation**: 2/2 complete ✅
- **Phase 0 (Foundation)**: 8/8 complete ✅
- **Phase 1 (Provider Interfaces)**: 6/6 complete ✅
- **Phase 2 (Core Orchestration)**: 7/7 complete ✅
- **Phase 3 (Local Development)**: 3/10 complete (30%)
- **Phase 4 (Provider Implementations)**: 0/3 complete
- **Phase 5 (Build and Deploy)**: 4/6 complete (67%)
- **Phase 6 (Migration System)**: 3/9 complete (33%)
- **Phase 7 (Infrastructure)**: 0/5 complete
- **Phase 8 (Operations)**: 0/6 complete
- **Phase 9 (CI Integration)**: 0/3 complete
- **Phase 10 (Project Scaffold)**: 0/5 complete
- **Governance**: 0/1 complete (1 wip)

**Overall Progress**: 33/72 features complete (~46%)

**Completed Features**:
- Architecture & Documentation: `ARCH_OVERVIEW`, `DOCS_ADR`
- Phase 0: `CORE_CONFIG`, `CLI_INIT`, `CORE_LOGGING`, `CORE_EXECUTIL`, `CLI_GLOBAL_FLAGS`, `CORE_BACKEND_REGISTRY`, `CORE_MIGRATION_REGISTRY`, `CORE_BACKEND_PROVIDER_CONFIG_SCHEMA`
- Phase 1: `PROVIDER_BACKEND_INTERFACE`, `PROVIDER_FRONTEND_INTERFACE`, `PROVIDER_NETWORK_INTERFACE`, `PROVIDER_CLOUD_INTERFACE`, `PROVIDER_CI_INTERFACE`, `PROVIDER_SECRETS_INTERFACE`
- Phase 2: `CORE_PLAN`, `CORE_ENV_RESOLUTION`, `CORE_STATE`, `CORE_STATE_TEST_ISOLATION`, `CORE_STATE_CONSISTENCY`, `CORE_COMPOSE`, `CLI_PHASE_EXECUTION_COMMON`
- Phase 3: `CLI_DEV_BASIC`, `PROVIDER_BACKEND_ENCORE`, `PROVIDER_BACKEND_GENERIC`
- Phase 5: `CLI_BUILD`, `CLI_PLAN`, `CLI_DEPLOY`, `CLI_ROLLBACK`
- Phase 6: `MIGRATION_ENGINE_RAW`, `CLI_MIGRATE_BASIC`, `CLI_RELEASES`

### Next Immediate Steps

1. **Complete Phase 3 (Local Development)** - Implement remaining features:
   - `PROVIDER_FRONTEND_GENERIC` - Generic dev command FrontendProvider (highest priority, unblocks rest of Phase 3)
   - `CLI_DEV` - Full stagecraft dev command
   - `DEV_COMPOSE_INFRA` - Compose infra up/down for dev
   - `DEV_TRAEFIK` - Traefik dev config generation
   - `DEV_MKCERT` - mkcert integration for local HTTPS
   - `DEV_HOSTS` - /etc/hosts management
   - `DEV_PROCESS_MGMT` - Process lifecycle management

2. **Complete Phase 5 (Build and Deploy)** - Implement remaining features:
   - `DEPLOY_COMPOSE_GEN` - Per-host Compose generation
   - `DEPLOY_ROLLOUT` - docker-rollout integration

3. **Complete Phase 4 (Provider Implementations)** - Implement core providers:
   - `PROVIDER_NETWORK_TAILSCALE` - Tailscale NetworkProvider implementation
   - `PROVIDER_CLOUD_DO` - DigitalOcean CloudProvider implementation
   - `DRIVER_DO` - DigitalOcean driver

---

## V2 Features (Deferred)

These features are documented but deferred to v2. They represent future directions for Stagecraft but are not part of the v1 scope:

1. **Ephemeral Environments** - Support for temporary, on-demand environments
2. **Audit Ledger** - Comprehensive build and deployment history tracking
3. **Infrastructure Recipes** - Reusable infrastructure templates
4. **Topology Visualization** - Visual representation of infrastructure
5. **Enhanced Testing** - Advanced testing capabilities
6. **Secrets Management** - Enhanced secrets orchestration
7. **Health Monitoring** - Proactive health monitoring and alerting
8. **Sync Capabilities** - Local and remote synchronization primitives
9. **Pipeline Composition** - Advanced pipeline configuration
10. **Snapshot Management** - Infrastructure snapshot capabilities
11. **Editor Integration** - IDE and editor plugin support
12. **Multi-tenant Support** - Organization and team features
13. **Observability** - Comprehensive monitoring and observability
14. **Cost Management** - Budget tracking and guardrails
15. **Migration Tooling** - Advanced migration planning and simulation

> **Note**: Detailed strategic planning for v2 features is maintained internally. The v2 feature list above provides a high-level overview of future directions without revealing prioritization, competitive positioning, or implementation details.

These will be added to `spec/features.yaml` with `status: v2` when v1 is complete and v2 planning begins.

---

## References

### Design Documents
- [`blog/01-why-not-kamal.md`](../blog/01-why-not-kamal.md) - Core architecture decisions
- [`blog/02-project-scaffold.md`](../blog/02-project-scaffold.md) - Project structure vision
- [`blog/03-migration-strategies.md`](../blog/03-migration-strategies.md) - Migration system design
- [`blog/04-features-now-and-future.md`](../blog/04-features-now-and-future.md) - v1/v2 feature split
- [`blog/05-development-strategy.md`](../blog/05-development-strategy.md) - Development methodology

### Specifications
- [`docs/stagecraft-spec.md`](stagecraft-spec.md) - Full application specification
- [`spec/features.yaml`](../spec/features.yaml) - Feature tracking (source of truth)
- [`docs/implementation-status.md`](implementation-status.md) - Quick reference status

### Architecture
- [`docs/adr/0001-architecture.md`](adr/0001-architecture.md) - Architecture and directory structure
- [`docs/architecture.md`](architecture.md) - System architecture overview
- [`spec/overview.md`](../spec/overview.md) - Project overview

### Related Features
- See [`spec/core/config.md`](../spec/core/config.md) for config schema details
- See [`spec/scaffold/stagecraft-dir.md`](../spec/scaffold/stagecraft-dir.md) for `.stagecraft/` directory structure

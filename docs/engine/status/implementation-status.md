# Implementation Status

> **⚠️ Note**: This document is a snapshot view. For the complete, up-to-date feature list, see [`spec/features.yaml`](../spec/features.yaml).
>
> This document shows a subset of features for quick reference. The full feature catalog with 61+ features organized by phase is available in [`docs/implementation-roadmap.md`](implementation-roadmap.md).

This document tracks the implementation status of Stagecraft features. It should be regenerated from `spec/features.yaml` when needed.

> **Last Updated**: See `spec/features.yaml` for the source of truth.

## Feature Status Legend

- **done** - Feature is complete with tests and documentation
- **wip** - Feature is in progress
- **todo** - Feature is planned but not started
- **blocked** - Feature is blocked by dependencies

## Features

### Architecture & Core

| ID | Title | Status | Owner | Spec | Tests |
|----|-------|--------|-------|------|-------|
| ARCH_OVERVIEW | Architecture documentation and project overview | done | bart | [overview.md](../spec/overview.md) | - |
| DOCS_ADR | ADR process and initial decisions | done | bart | [adr/0001-architecture.md](adr/0001-architecture.md) | - |

### Core Functionality

| ID | Title | Status | Owner | Spec | Tests |
|----|-------|--------|-------|------|-------|
| CORE_CONFIG | Config loading and validation | done | bart | [core/config.md](../spec/core/config.md) | [config_test.go](../pkg/config/config_test.go) |
| CORE_LOGGING | Structured logging helpers | done | bart | [core/logging.md](../spec/core/logging.md) | [logging_test.go](../pkg/logging/logging_test.go) |
| CORE_EXECUTIL | Process execution utilities | done | bart | [core/executil.md](../spec/core/executil.md) | [executil_test.go](../pkg/executil/executil_test.go) |
| CORE_PLAN | Deployment planning engine | done | bart | [core/plan.md](../spec/core/plan.md) | [plan_test.go](../internal/core/plan_test.go) |
| CORE_ENV_RESOLUTION | Environment resolution and context | done | bart | [core/env-resolution.md](../spec/core/env-resolution.md) | [env_test.go](../internal/core/env/env_test.go) |
| CORE_STATE | State management (release history) | done | bart | [core/state.md](../spec/core/state.md) | [state_test.go](../internal/core/state/state_test.go) |
| CORE_STATE_TEST_ISOLATION | State test isolation for CLI commands | done | bart | [core/state-test-isolation.md](../spec/core/state-test-isolation.md) | [test_helpers.go](../internal/cli/commands/test_helpers.go), [deploy_test.go](../internal/cli/commands/deploy_test.go), [rollback_test.go](../internal/cli/commands/rollback_test.go), [releases_test.go](../internal/cli/commands/releases_test.go) |
| CORE_STATE_CONSISTENCY | State durability and read-after-write guarantees | done | bart | [core/state-consistency.md](../spec/core/state-consistency.md) | [state_test.go](../internal/core/state/state_test.go) |
| CORE_COMPOSE | Docker Compose integration | done | bart | [core/compose.md](../spec/core/compose.md) | [compose_test.go](../internal/compose/compose_test.go) |
| CORE_BACKEND_REGISTRY | Backend provider registry system | done | bart | [core/backend-registry.md](../spec/core/backend-registry.md) | [registry_test.go](../pkg/providers/backend/registry_test.go) |
| CORE_MIGRATION_REGISTRY | Migration engine registry system | done | bart | [core/migration-registry.md](../spec/core/migration-registry.md) | [registry_test.go](../pkg/providers/migration/registry_test.go) |
| CORE_BACKEND_PROVIDER_CONFIG_SCHEMA | Provider-scoped backend configuration schema | done | bart | [core/backend-provider-config.md](../spec/core/backend-provider-config.md) | [config_test.go](../pkg/config/config_test.go) |

### CLI Commands

| ID | Title | Status | Owner | Spec | Tests |
|----|-------|--------|-------|------|-------|
| CLI_INIT | Project bootstrap command | done | bart | [commands/init.md](../spec/commands/init.md) | [init_test.go](../internal/cli/commands/init_test.go), [init_smoke_test.go](../test/e2e/init_smoke_test.go) |
| CLI_GLOBAL_FLAGS | Global flags (--env, --config, --verbose, --dry-run) | done | bart | [core/global-flags.md](../spec/core/global-flags.md) | [root_test.go](../internal/cli/root_test.go) |
| CLI_DEV_BASIC | Basic stagecraft dev command that delegates to backend provider | done | bart | [commands/dev-basic.md](../spec/commands/dev-basic.md) | [dev_test.go](../internal/cli/commands/dev_test.go), [dev_smoke_test.go](../test/e2e/dev_smoke_test.go) |
| CLI_BUILD | stagecraft build command | done | bart | [commands/build.md](../spec/commands/build.md) | [build_test.go](../internal/cli/commands/build_test.go) |
| CLI_PLAN | Plan command (dry-run) | done | bart | [commands/plan.md](../spec/commands/plan.md) | [plan_test.go](../internal/cli/commands/plan_test.go) |
| CLI_DEPLOY | Deploy command | done | bart | [commands/deploy.md](../spec/commands/deploy.md) | [deploy_test.go](../internal/cli/commands/deploy_test.go), [deploy_smoke_test.go](../test/e2e/deploy_smoke_test.go) |
| CLI_ROLLBACK | stagecraft rollback command | done | bart | [commands/rollback.md](../spec/commands/rollback.md) | [rollback_test.go](../internal/cli/commands/rollback_test.go) |
| CLI_RELEASES | stagecraft releases list/show commands | done | bart | [commands/releases.md](../spec/commands/releases.md) | [releases_test.go](../internal/cli/commands/releases_test.go) |
| CLI_MIGRATE_BASIC | Basic stagecraft migrate command using registered migration engines | done | bart | [commands/migrate-basic.md](../spec/commands/migrate-basic.md) | [migrate_test.go](../internal/cli/commands/migrate_test.go), [migrate_smoke_test.go](../test/e2e/migrate_smoke_test.go) |
| CLI_PHASE_EXECUTION_COMMON | Shared phase execution semantics for deploy and rollback | done | bart | [core/phase-execution-common.md](../spec/core/phase-execution-common.md) | [phases_common_test.go](../internal/cli/commands/phases_common_test.go), [deploy_test.go](../internal/cli/commands/deploy_test.go) |

### Providers

| ID | Title | Status | Owner | Spec | Tests |
|----|-------|--------|-------|------|-------|
| PROVIDER_BACKEND_INTERFACE | BackendProvider interface definition | done | bart | [core/backend-registry.md](../spec/core/backend-registry.md) | [backend_test.go](../pkg/providers/backend/backend_test.go) |
| PROVIDER_BACKEND_GENERIC | Generic command-based BackendProvider implementation | done | bart | [providers/backend/generic.md](../spec/providers/backend/generic.md) | [generic_test.go](../internal/providers/backend/generic/generic_test.go) |
| PROVIDER_BACKEND_ENCORE | Encore.ts BackendProvider implementation | done | bart | [providers/backend/encore-ts.md](../spec/providers/backend/encore-ts.md) | [encorets_test.go](../internal/providers/backend/encorets/encorets_test.go) |
| PROVIDER_FRONTEND_INTERFACE | FrontendProvider interface definition | done | bart | [providers/frontend/interface.md](../spec/providers/frontend/interface.md) | [frontend_test.go](../pkg/providers/frontend/frontend_test.go) |
| PROVIDER_NETWORK_INTERFACE | NetworkProvider interface definition | done | bart | [providers/network/interface.md](../spec/providers/network/interface.md) | [registry_test.go](../pkg/providers/network/registry_test.go) |
| PROVIDER_CLOUD_INTERFACE | CloudProvider interface definition | done | bart | [providers/cloud/interface.md](../spec/providers/cloud/interface.md) | [registry_test.go](../pkg/providers/cloud/registry_test.go) |
| PROVIDER_CI_INTERFACE | CIProvider interface definition | done | bart | [providers/ci/interface.md](../spec/providers/ci/interface.md) | [registry_test.go](../pkg/providers/ci/registry_test.go) |
| PROVIDER_SECRETS_INTERFACE | SecretsProvider interface definition | done | bart | [providers/secrets/interface.md](../spec/providers/secrets/interface.md) | [registry_test.go](../pkg/providers/secrets/registry_test.go) |
| MIGRATION_ENGINE_RAW | Raw SQL migration engine implementation | done | bart | [providers/migration/raw.md](../spec/providers/migration/raw.md) | [raw_test.go](../internal/providers/migration/raw/raw_test.go) |

### Drivers

| ID | Title | Status | Owner | Spec | Tests |
|----|-------|--------|-------|------|-------|
| DRIVER_DO | DigitalOcean driver | todo | bart | [drivers/do.md](../spec/drivers/do.md) | [do_test.go](../internal/drivers/do/do_test.go) |

## Implementation Notes

### Completed Features

**Architecture & Documentation:**
- **ARCH_OVERVIEW**: Architecture documentation and project overview
- **DOCS_ADR**: ADR process and initial decisions

**Phase 0: Foundation:**
- **CORE_CONFIG**: Config loading and validation with full schema support
- **CLI_INIT**: Project bootstrap command with interactive and non-interactive modes
- **CORE_LOGGING**: Structured logging helpers with verbose mode support
- **CORE_EXECUTIL**: Process execution utilities
- **CLI_GLOBAL_FLAGS**: Global flags (--env, --config, --verbose, --dry-run)
- **CORE_BACKEND_REGISTRY**: Backend provider registry system with registration support
- **CORE_MIGRATION_REGISTRY**: Migration engine registry system with registration support
- **CORE_BACKEND_PROVIDER_CONFIG_SCHEMA**: Provider-scoped backend configuration schema

**Phase 1: Provider Interfaces:**
- **PROVIDER_BACKEND_INTERFACE**: BackendProvider interface definition
- **PROVIDER_FRONTEND_INTERFACE**: FrontendProvider interface definition with registry system
- **PROVIDER_NETWORK_INTERFACE**: NetworkProvider interface definition
- **PROVIDER_CLOUD_INTERFACE**: CloudProvider interface definition
- **PROVIDER_CI_INTERFACE**: CIProvider interface definition
- **PROVIDER_SECRETS_INTERFACE**: SecretsProvider interface definition

**Phase 2: Core Orchestration:**
- **CORE_PLAN**: Deployment planning engine
- **CORE_ENV_RESOLUTION**: Environment resolution and context
- **CORE_STATE**: State management (release history)
- **CORE_STATE_TEST_ISOLATION**: Complete test isolation infrastructure for state-touching CLI tests
- **CORE_STATE_CONSISTENCY**: State durability and read-after-write guarantees
- **CORE_COMPOSE**: Docker Compose file loading, parsing, and environment-specific override generation
- **CLI_PHASE_EXECUTION_COMMON**: Shared phase execution semantics for deploy and rollback commands

**Phase 3: Local Development:**
- **CLI_DEV_BASIC**: Basic dev command that delegates to backend provider
- **PROVIDER_BACKEND_ENCORE**: Encore.ts BackendProvider implementation with secret syncing, env file parsing, and Docker build support
- **PROVIDER_BACKEND_GENERIC**: Generic command-based BackendProvider implementation

**Phase 5: Build and Deploy:**
- **CLI_BUILD**: stagecraft build command
- **CLI_PLAN**: Plan command (dry-run)
- **CLI_DEPLOY**: Deploy command
- **CLI_ROLLBACK**: stagecraft rollback command

**Phase 6: Migration System:**
- **MIGRATION_ENGINE_RAW**: Raw SQL migration engine implementation
- **CLI_MIGRATE_BASIC**: Basic migrate command with plan and run support
- **CLI_RELEASES**: stagecraft releases list/show commands

### In Progress

- **GOV_V1_CORE**: Governance Core for v1 (wip)

### Planned Next Steps

1. **PROVIDER_FRONTEND_GENERIC** - Generic dev command FrontendProvider (highest priority, unblocks rest of Phase 3)
2. **CLI_DEV** - Full stagecraft dev command with infrastructure orchestration
3. **DEV_COMPOSE_INFRA** - Compose infra up/down for dev
4. **DEV_TRAEFIK** - Traefik dev config generation
5. **DEV_MKCERT** - mkcert integration for local HTTPS
6. **DEV_HOSTS** - /etc/hosts management
7. **DEV_PROCESS_MGMT** - Process lifecycle management

## Coverage Status

Current test coverage targets:
- **Core packages** (`pkg/config`, `internal/core`): Target 80%+
- **CLI layer** (`internal/cli`): Target 70%+
- **Drivers** (`internal/drivers`): Target 70%+
- **Overall**: Target 60%+ (increasing to 80% as project matures)

## How to Update

This file should be regenerated when `spec/features.yaml` changes. To update:

```bash
# Run the validation script (which also serves as a template for generation)
./scripts/validate-spec.sh

# Or manually update this file to match spec/features.yaml
```

For detailed feature specifications, see the individual spec files referenced in the table above.


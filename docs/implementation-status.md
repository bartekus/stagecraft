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
| CORE_PLAN | Deployment planning engine | done | bart | [core/plan.md](../spec/core/plan.md) | [plan_test.go](../internal/core/plan_test.go) |
| CORE_COMPOSE | Docker Compose integration | done | bart | [core/compose.md](../spec/core/compose.md) | [compose_test.go](../internal/compose/compose_test.go) |
| CORE_BACKEND_REGISTRY | Backend provider registry system | done | bart | [core/backend-registry.md](../spec/core/backend-registry.md) | [registry_test.go](../pkg/providers/backend/registry_test.go) |
| CORE_MIGRATION_REGISTRY | Migration engine registry system | done | bart | [core/migration-registry.md](../spec/core/migration-registry.md) | [registry_test.go](../pkg/providers/migration/registry_test.go) |
| CORE_BACKEND_PROVIDER_CONFIG_SCHEMA | Provider-scoped backend configuration schema | done | bart | [core/backend-provider-config.md](../spec/core/backend-provider-config.md) | [config_test.go](../pkg/config/config_test.go) |
| CORE_STATE | State management (release history) | done | bart | [core/state.md](../spec/core/state.md) | [state_test.go](../internal/core/state/state_test.go) |
| CORE_STATE_TEST_ISOLATION | State test isolation for CLI commands | done | bart | [core/state-test-isolation.md](../spec/core/state-test-isolation.md) | [test_helpers.go](../internal/cli/commands/test_helpers.go), [deploy_test.go](../internal/cli/commands/deploy_test.go), [rollback_test.go](../internal/cli/commands/rollback_test.go), [releases_test.go](../internal/cli/commands/releases_test.go) |
| CORE_STATE_CONSISTENCY | State durability and read-after-write guarantees | todo | bart | [core/state-consistency.md](../spec/core/state-consistency.md) | [state_test.go](../internal/core/state/state_test.go) |

### CLI Commands

| ID | Title | Status | Owner | Spec | Tests |
|----|-------|--------|-------|------|-------|
| CLI_INIT | Project bootstrap command | done | bart | [commands/init.md](../spec/commands/init.md) | [init_test.go](../internal/cli/commands/init_test.go), [init_smoke_test.go](../test/e2e/init_smoke_test.go) |
| CLI_DEV_BASIC | Basic stagecraft dev command that delegates to backend provider | done | bart | [commands/dev-basic.md](../spec/commands/dev-basic.md) | [dev_test.go](../internal/cli/commands/dev_test.go), [dev_smoke_test.go](../test/e2e/dev_smoke_test.go) |
| CLI_MIGRATE_BASIC | Basic stagecraft migrate command using registered migration engines | done | bart | [commands/migrate-basic.md](../spec/commands/migrate-basic.md) | [migrate_test.go](../internal/cli/commands/migrate_test.go), [migrate_smoke_test.go](../test/e2e/migrate_smoke_test.go) |
| CLI_PLAN | Plan command (dry-run) | todo | bart | [commands/plan.md](../spec/commands/plan.md) | [plan_test.go](../internal/cli/commands/plan_test.go) |
| CLI_PHASE_EXECUTION_COMMON | Shared phase execution semantics for deploy and rollback | done | bart | [core/phase-execution-common.md](../spec/core/phase-execution-common.md) | [phases_common_test.go](../internal/cli/commands/phases_common_test.go), [deploy_test.go](../internal/cli/commands/deploy_test.go) |
| CLI_DEPLOY | Deploy command | todo | bart | [commands/deploy.md](../spec/commands/deploy.md) | [deploy_test.go](../internal/cli/commands/deploy_test.go) |

### Providers

| ID | Title | Status | Owner | Spec | Tests |
|----|-------|--------|-------|------|-------|
| PROVIDER_BACKEND_INTERFACE | BackendProvider interface definition | done | bart | [core/backend-registry.md](../spec/core/backend-registry.md) | [backend_test.go](../pkg/providers/backend/backend_test.go) |
| PROVIDER_BACKEND_GENERIC | Generic command-based BackendProvider implementation | done | bart | [providers/backend/generic.md](../spec/providers/backend/generic.md) | [generic_test.go](../internal/providers/backend/generic/generic_test.go) |
| PROVIDER_BACKEND_ENCORE | Encore.ts BackendProvider implementation | done | bart | [providers/backend/encore-ts.md](../spec/providers/backend/encore-ts.md) | [encorets_test.go](../internal/providers/backend/encorets/encorets_test.go) |
| PROVIDER_FRONTEND_INTERFACE | FrontendProvider interface definition | done | bart | [providers/frontend/interface.md](../spec/providers/frontend/interface.md) | [frontend_test.go](../pkg/providers/frontend/frontend_test.go) |
| MIGRATION_ENGINE_RAW | Raw SQL migration engine implementation | done | bart | [providers/migration/raw.md](../spec/providers/migration/raw.md) | [raw_test.go](../internal/providers/migration/raw/raw_test.go) |

### Drivers

| ID | Title | Status | Owner | Spec | Tests |
|----|-------|--------|-------|------|-------|
| DRIVER_DO | DigitalOcean driver | todo | bart | [drivers/do.md](../spec/drivers/do.md) | [do_test.go](../internal/drivers/do/do_test.go) |

## Implementation Notes

### Completed Features

- **ARCH_OVERVIEW**: Basic architecture documentation and project structure established
- **DOCS_ADR**: ADR process documented with initial architecture decision (ADR 0001)
- **CORE_CONFIG**: Config loading and validation with full schema support
- **CLI_INIT**: Project bootstrap command with interactive and non-interactive modes
- **CORE_LOGGING**: Structured logging helpers with verbose mode support
- **CORE_PLAN**: Deployment planning engine
- **CORE_COMPOSE**: Docker Compose file loading, parsing, and environment-specific override generation
- **CORE_BACKEND_REGISTRY**: Backend provider registry system with registration support
- **CORE_MIGRATION_REGISTRY**: Migration engine registry system with registration support
- **CORE_BACKEND_PROVIDER_CONFIG_SCHEMA**: Provider-scoped backend configuration schema
- **CORE_STATE_TEST_ISOLATION**: Complete test isolation infrastructure for state-touching CLI tests
- **CLI_PHASE_EXECUTION_COMMON**: Shared phase execution semantics for deploy and rollback commands
- **PROVIDER_BACKEND_INTERFACE**: BackendProvider interface definition
- **CLI_DEV_BASIC**: Basic dev command that delegates to backend provider
- **PROVIDER_BACKEND_GENERIC**: Generic command-based BackendProvider implementation
- **PROVIDER_BACKEND_ENCORE**: Encore.ts BackendProvider implementation with secret syncing, env file parsing, and Docker build support
- **PROVIDER_FRONTEND_INTERFACE**: FrontendProvider interface definition with registry system
- **MIGRATION_ENGINE_RAW**: Raw SQL migration engine implementation
- **CLI_MIGRATE_BASIC**: Basic migrate command with plan and run support

### In Progress

_None currently_

### Planned Next Steps

1. Complete `CORE_EXECUTIL` - Process execution utilities
2. Complete `CLI_GLOBAL_FLAGS` - Global flags with precedence support
3. Implement remaining provider interfaces (Frontend, Network, Cloud, CI, Secrets)
4. Build full `CLI_DEV` command with infrastructure orchestration
5. Add `DRIVER_DO` for DigitalOcean integration

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


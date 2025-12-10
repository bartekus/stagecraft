# Implementation Status

> **⚠️ Note**: This document is a snapshot view. For the complete, up-to-date feature list, see [`spec/features.yaml`](../../../spec/features.yaml).
>
> This document shows a subset of features for quick reference. The full feature catalog with 61+ features organized by phase is available in [`docs/narrative/implementation-roadmap.md`](../../narrative/implementation-roadmap.md).

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
| ARCH_OVERVIEW | Architecture documentation and project overview | todo | bart | [overview.md](../../../spec/overview.md) | - |
| DOCS_ADR | ADR process and initial decisions | todo | bart | [0001-architecture.md](../../../spec/adr/0001-architecture.md) | - |

### CLI Commands

| ID | Title | Status | Owner | Spec | Tests |
|----|-------|--------|-------|------|-------|
| CLI_ATTACH | stagecraft attach for existing projects | todo | bart | [attach.md](../../../spec/commands/attach.md) | [attach_test.go](../../../internal/cli/commands/attach_test.go) |
| CLI_BUILD | stagecraft build command | done | bart | [build.md](../../../spec/commands/build.md) | [build_test.go](../../../internal/cli/commands/build_test.go) |
| CLI_CI_INIT | stagecraft ci init command | todo | bart | [ci-init.md](../../../spec/commands/ci-init.md) | [ci_init_test.go](../../../internal/cli/commands/ci_init_test.go) |
| CLI_CI_RUN | stagecraft ci run command | todo | bart | [ci-run.md](../../../spec/commands/ci-run.md) | [ci_run_test.go](../../../internal/cli/commands/ci_run_test.go) |
| CLI_COMMIT_SUGGEST | stagecraft commit suggest command | done | bart | [commit-suggest.md](../../../spec/commands/commit-suggest.md) | [commit_suggest_test.go](../../../internal/cli/commands/commit_suggest_test.go) |
| CLI_DEPLOY | Deploy command | done | bart | [deploy.md](../../../spec/commands/deploy.md) | [deploy_test.go](../../../internal/cli/commands/deploy_test.go), [deploy_smoke_test.go](../../../test/e2e/deploy_smoke_test.go) |
| CLI_DEV | stagecraft dev command (full feature set) | done | bart | [dev.md](../../../spec/commands/dev.md) | [dev_test.go](../../../internal/cli/commands/dev_test.go), [dev_smoke_test.go](../../../test/e2e/dev_smoke_test.go) |
| CLI_DEV_BASIC | Basic stagecraft dev command that delegates to backend provider | done | bart | [dev-basic.md](../../../spec/commands/dev-basic.md) | [dev_test.go](../../../internal/cli/commands/dev_test.go), [dev_smoke_test.go](../../../test/e2e/dev_smoke_test.go) |
| CLI_GLOBAL_FLAGS | Global flags (--env, --config, --verbose, --dry-run) | done | bart | [global-flags.md](../../../spec/core/global-flags.md) | [root_test.go](../../../internal/cli/root_test.go) |
| CLI_INFRA_DOWN | stagecraft infra down command | todo | bart | [infra-down.md](../../../spec/commands/infra-down.md) | [infra_down_test.go](../../../internal/cli/commands/infra_down_test.go) |
| CLI_INFRA_UP | stagecraft infra up command | todo | bart | [infra-up.md](../../../spec/commands/infra-up.md) | [infra_up_test.go](../../../internal/cli/commands/infra_up_test.go) |
| CLI_INIT | Project bootstrap command | done | bart | [init.md](../../../spec/commands/init.md) | [init_test.go](../../../internal/cli/commands/init_test.go), [init_smoke_test.go](../../../test/e2e/init_smoke_test.go) |
| CLI_INIT_TEMPLATE | Template system for stagecraft init | todo | bart | [templates.md](../../../spec/scaffold/templates.md) | [templates_test.go](../../../internal/scaffold/templates_test.go) |
| CLI_LOGS | stagecraft logs command | todo | bart | [logs.md](../../../spec/commands/logs.md) | [logs_test.go](../../../internal/cli/commands/logs_test.go) |
| CLI_MIGRATE_BASIC | Basic stagecraft migrate command using registered migration engines | done | bart | [migrate-basic.md](../../../spec/commands/migrate-basic.md) | [migrate_test.go](../../../internal/cli/commands/migrate_test.go), [migrate_smoke_test.go](../../../test/e2e/migrate_smoke_test.go) |
| CLI_MIGRATE_PLAN | stagecraft migrate plan command (dedicated) | todo | bart | [migrate-plan.md](../../../spec/commands/migrate-plan.md) | [migrate_plan_test.go](../../../internal/cli/commands/migrate_plan_test.go) |
| CLI_MIGRATE_RUN | stagecraft migrate run command (dedicated) | todo | bart | [migrate-run.md](../../../spec/commands/migrate-run.md) | [migrate_run_test.go](../../../internal/cli/commands/migrate_run_test.go) |
| CLI_NEW | stagecraft new --template=platform | todo | bart | [new.md](../../../spec/commands/new.md) | [new_test.go](../../../internal/cli/commands/new_test.go) |
| CLI_PHASE_EXECUTION_COMMON | Shared phase execution semantics for deploy and rollback | done | bart | [phase-execution-common.md](../../../spec/core/phase-execution-common.md) | [phases_common_test.go](../../../internal/cli/commands/phases_common_test.go), [deploy_test.go](../../../internal/cli/commands/deploy_test.go) |
| CLI_PLAN | Plan command (dry-run) | done | bart | [plan.md](../../../spec/commands/plan.md) | [plan_test.go](../../../internal/cli/commands/plan_test.go) |
| CLI_RELEASES | stagecraft releases list/show commands | done | bart | [releases.md](../../../spec/commands/releases.md) | [releases_test.go](../../../internal/cli/commands/releases_test.go) |
| CLI_ROLLBACK | stagecraft rollback command | done | bart | [rollback.md](../../../spec/commands/rollback.md) | [rollback_test.go](../../../internal/cli/commands/rollback_test.go) |
| CLI_SECRETS_SYNC | stagecraft secrets sync command | todo | bart | [secrets-sync.md](../../../spec/commands/secrets-sync.md) | [secrets_sync_test.go](../../../internal/cli/commands/secrets_sync_test.go) |
| CLI_SSH | stagecraft ssh command | todo | bart | [ssh.md](../../../spec/commands/ssh.md) | [ssh_test.go](../../../internal/cli/commands/ssh_test.go) |
| CLI_STATUS | stagecraft status command | todo | bart | [status.md](../../../spec/commands/status.md) | [status_test.go](../../../internal/cli/commands/status_test.go) |

### Core Functionality

| ID | Title | Status | Owner | Spec | Tests |
|----|-------|--------|-------|------|-------|
| CORE_BACKEND_PROVIDER_CONFIG_SCHEMA | Provider-scoped backend configuration schema | done | bart | [backend-provider-config.md](../../../spec/core/backend-provider-config.md) | [config_test.go](../../../pkg/config/config_test.go) |
| CORE_BACKEND_REGISTRY | Backend provider registry system | done | bart | [backend-registry.md](../../../spec/core/backend-registry.md) | [registry_test.go](../../../pkg/providers/backend/registry_test.go) |
| CORE_COMPOSE | Docker Compose integration | done | bart | [compose.md](../../../spec/core/compose.md) | [compose_test.go](../../../internal/compose/compose_test.go) |
| CORE_CONFIG | Config loading and validation | done | bart | [config.md](../../../spec/core/config.md) | [config_test.go](../../../pkg/config/config_test.go) |
| CORE_ENV_RESOLUTION | Environment resolution and context | done | bart | [env-resolution.md](../../../spec/core/env-resolution.md) | [env_test.go](../../../internal/core/env/env_test.go) |
| CORE_EXECUTIL | Process execution utilities | done | bart | [executil.md](../../../spec/core/executil.md) | [executil_test.go](../../../pkg/executil/executil_test.go) |
| CORE_LOGGING | Structured logging helpers | done | bart | [logging.md](../../../spec/core/logging.md) | [logging_test.go](../../../pkg/logging/logging_test.go) |
| CORE_MIGRATION_REGISTRY | Migration engine registry system | done | bart | [migration-registry.md](../../../spec/core/migration-registry.md) | [registry_test.go](../../../pkg/providers/migration/registry_test.go) |
| CORE_PLAN | Deployment planning engine | done | bart | [plan.md](../../../spec/core/plan.md) | [plan_test.go](../../../internal/core/plan_test.go) |
| CORE_STATE | State management (release history) | done | bart | [state.md](../../../spec/core/state.md) | [state_test.go](../../../internal/core/state/state_test.go) |
| CORE_STATE_CONSISTENCY | State durability and read-after-write guarantees | done | bart | [state-consistency.md](../../../spec/core/state-consistency.md) | [state_test.go](../../../internal/core/state/state_test.go) |
| CORE_STATE_TEST_ISOLATION | State test isolation for CLI commands | done | bart | [state-test-isolation.md](../../../spec/core/state-test-isolation.md) | [test_helpers.go](../../../internal/cli/commands/test_helpers.go), [deploy_test.go](../../../internal/cli/commands/deploy_test.go), [rollback_test.go](../../../internal/cli/commands/rollback_test.go), [releases_test.go](../../../internal/cli/commands/releases_test.go) |

### Drivers

| ID | Title | Status | Owner | Spec | Tests |
|----|-------|--------|-------|------|-------|
| DRIVER_DO | DigitalOcean driver | todo | bart | [do.md](../../../spec/drivers/do.md) | [do_test.go](../../../internal/drivers/do/do_test.go) |

### Migration Engines

| ID | Title | Status | Owner | Spec | Tests |
|----|-------|--------|-------|------|-------|
| MIGRATION_CONFIG | Migration config schema in stagecraft.yml | todo | bart | [config.md](../../../spec/migrations/config.md) | [config_test.go](../../../pkg/config/config_test.go) |
| MIGRATION_CONTAINER_RUNNER | ContainerRunner interface | todo | bart | [container-runner.md](../../../spec/migrations/container-runner.md) | [runner_test.go](../../../pkg/migrations/runner_test.go) |
| MIGRATION_ENGINE_RAW | Raw SQL migration engine implementation | done | bart | [raw.md](../../../spec/providers/migration/raw.md) | [raw_test.go](../../../internal/providers/migration/raw/raw_test.go) |
| MIGRATION_INTERFACE | Migrator interface | todo | bart | [interface.md](../../../spec/migrations/interface.md) | [migrator_test.go](../../../pkg/migrations/migrator_test.go) |
| MIGRATION_POST_DEPLOY | Post-deploy migration execution | todo | bart | [post-deploy.md](../../../spec/migrations/post-deploy.md) | [migrations_test.go](../../../internal/deploy/migrations_test.go) |
| MIGRATION_PRE_DEPLOY | Pre-deploy migration execution | todo | bart | [pre-deploy.md](../../../spec/migrations/pre-deploy.md) | [migrations_test.go](../../../internal/deploy/migrations_test.go) |

### Other

| ID | Title | Status | Owner | Spec | Tests |
|----|-------|--------|-------|------|-------|
| DEPLOY_COMPOSE_GEN | Per-host Compose generation | todo | bart | [compose-gen.md](../../../spec/deploy/compose-gen.md) | [compose_test.go](../../../internal/deploy/compose_test.go) |
| DEPLOY_ROLLOUT | docker-rollout integration | todo | bart | [rollout.md](../../../spec/deploy/rollout.md) | [rollout_test.go](../../../internal/deploy/rollout_test.go) |
| DEV_COMPOSE_INFRA | Compose infra up/down for dev | done | bart | [compose-infra.md](../../../spec/dev/compose-infra.md) | [generator_test.go](../../../internal/dev/compose/generator_test.go), [golden_test.go](../../../internal/dev/compose/golden_test.go) |
| DEV_HOSTS | /etc/hosts management | done | bart | [hosts.md](../../../spec/dev/hosts.md) | [hosts_test.go](../../../internal/dev/hosts/hosts_test.go) |
| DEV_MKCERT | mkcert integration for local HTTPS | done | bart | [mkcert.md](../../../spec/dev/mkcert.md) | [generator_test.go](../../../internal/dev/mkcert/generator_test.go) |
| DEV_PROCESS_MGMT | Process lifecycle management | done | bart | [process-mgmt.md](../../../spec/dev/process-mgmt.md) | [runner_test.go](../../../internal/dev/process/runner_test.go) |
| DEV_TRAEFIK | Traefik dev config generation | done | bart | [traefik.md](../../../spec/dev/traefik.md) | [generator_test.go](../../../internal/dev/traefik/generator_test.go) |
| GOV_CLI_EXIT_CODES | CLI exit code governance and standardization | todo | bart | [GOV_CLI_EXIT_CODES.md](../../../spec/governance/GOV_CLI_EXIT_CODES.md) | - |
| GOV_STATUS_ROADMAP | stagecraft status roadmap command | done | bart | [status-roadmap.md](../../../spec/commands/status-roadmap.md) | [status_test.go](../../../internal/cli/commands/status_test.go), [phase_test.go](../../../internal/tools/roadmap/phase_test.go), [stats_test.go](../../../internal/tools/roadmap/stats_test.go), [generator_test.go](../../../internal/tools/roadmap/generator_test.go) |
| GOV_V1_CORE | Governance Core for v1 | done | bart | [GOV_V1_CORE.md](../../../spec/governance/GOV_V1_CORE.md) | [specschema_test.go](../../../internal/tools/specschema/specschema_test.go), [cliintrospect_test.go](../../../internal/tools/cliintrospect/cliintrospect_test.go), [features_test.go](../../../internal/tools/features/features_test.go), [mapping_test.go](../../../internal/governance/mapping/mapping_test.go), [docs_test.go](../../../internal/tools/docs/docs_test.go), [diff_test.go](../../../internal/tools/specvscli/diff_test.go) |
| INFRA_FIREWALL | Firewall configuration | todo | bart | [firewall.md](../../../spec/infra/firewall.md) | [firewall_test.go](../../../internal/infra/firewall_test.go) |
| INFRA_HOST_BOOTSTRAP | Host bootstrap (Docker, Tailscale, etc.) | todo | bart | [bootstrap.md](../../../spec/infra/bootstrap.md) | [bootstrap_test.go](../../../internal/infra/bootstrap_test.go) |
| INFRA_VOLUME_MGMT | Volume management | todo | bart | [volumes.md](../../../spec/infra/volumes.md) | [volumes_test.go](../../../internal/infra/volumes_test.go) |
| SCAFFOLD_STAGECRAFT_DIR | .stagecraft/ directory generation | todo | bart | [stagecraft-dir.md](../../../spec/scaffold/stagecraft-dir.md) | [dir_test.go](../../../internal/scaffold/dir_test.go) |
| TEMPLATE_PLATFORM | Platform template (embedded) | todo | bart | [platform-template.md](../../../spec/scaffold/platform-template.md) | [platform_test.go](../../../internal/scaffold/platform_test.go) |

### Providers

| ID | Title | Status | Owner | Spec | Tests |
|----|-------|--------|-------|------|-------|
| PROVIDER_BACKEND_ENCORE | Encore.ts BackendProvider implementation | done | bart | [encore-ts.md](../../../spec/providers/backend/encore-ts.md) | [encorets_test.go](../../../internal/providers/backend/encorets/encorets_test.go) |
| PROVIDER_BACKEND_GENERIC | Generic command-based BackendProvider implementation | done | bart | [generic.md](../../../spec/providers/backend/generic.md) | [generic_test.go](../../../internal/providers/backend/generic/generic_test.go) |
| PROVIDER_BACKEND_INTERFACE | BackendProvider interface definition | done | bart | [backend-registry.md](../../../spec/core/backend-registry.md) | [backend_test.go](../../../pkg/providers/backend/backend_test.go) |
| PROVIDER_CI_GITHUB | GitHub Actions CIProvider | todo | bart | [github.md](../../../spec/providers/ci/github.md) | [github_test.go](../../../internal/providers/ci/github/github_test.go) |
| PROVIDER_CI_INTERFACE | CIProvider interface definition | done | bart | [interface.md](../../../spec/providers/ci/interface.md) | [registry_test.go](../../../pkg/providers/ci/registry_test.go) |
| PROVIDER_CLOUD_DO | DigitalOcean CloudProvider implementation | todo | bart | [digitalocean.md](../../../spec/providers/cloud/digitalocean.md) | [do_test.go](../../../internal/providers/cloud/digitalocean/do_test.go) |
| PROVIDER_CLOUD_INTERFACE | CloudProvider interface definition | done | bart | [interface.md](../../../spec/providers/cloud/interface.md) | [registry_test.go](../../../pkg/providers/cloud/registry_test.go) |
| PROVIDER_FRONTEND_GENERIC | Generic dev command FrontendProvider | done | bart | [generic.md](../../../spec/providers/frontend/generic.md) | [generic_test.go](../../../internal/providers/frontend/generic/generic_test.go) |
| PROVIDER_FRONTEND_INTERFACE | FrontendProvider interface definition | done | bart | [interface.md](../../../spec/providers/frontend/interface.md) | [frontend_test.go](../../../pkg/providers/frontend/frontend_test.go) |
| PROVIDER_NETWORK_INTERFACE | NetworkProvider interface definition | done | bart | [interface.md](../../../spec/providers/network/interface.md) | [registry_test.go](../../../pkg/providers/network/registry_test.go) |
| PROVIDER_NETWORK_TAILSCALE | Tailscale NetworkProvider implementation | done | bart | [tailscale.md](../../../spec/providers/network/tailscale.md) | [tailscale_test.go](../../../internal/providers/network/tailscale/tailscale_test.go), [registry_test.go](../../../internal/providers/network/tailscale/registry_test.go) |
| PROVIDER_SECRETS_ENCORE | Encore dev secrets SecretsProvider | todo | bart | [encore.md](../../../spec/providers/secrets/encore.md) | [encore_test.go](../../../internal/providers/secrets/encore/encore_test.go) |
| PROVIDER_SECRETS_ENVFILE | Env file SecretsProvider | todo | bart | [envfile.md](../../../spec/providers/secrets/envfile.md) | [envfile_test.go](../../../internal/providers/secrets/envfile/envfile_test.go) |
| PROVIDER_SECRETS_INTERFACE | SecretsProvider interface definition | done | bart | [interface.md](../../../spec/providers/secrets/interface.md) | [registry_test.go](../../../pkg/providers/secrets/registry_test.go) |

## Implementation Notes

### Completed Features

- **CLI_BUILD**: stagecraft build command
- **CLI_COMMIT_SUGGEST**: stagecraft commit suggest command
- **CLI_DEPLOY**: Deploy command
- **CLI_DEV**: stagecraft dev command (full feature set)
- **CLI_DEV_BASIC**: Basic stagecraft dev command that delegates to backend provider
- **CLI_GLOBAL_FLAGS**: Global flags (--env, --config, --verbose, --dry-run)
- **CLI_INIT**: Project bootstrap command
- **CLI_MIGRATE_BASIC**: Basic stagecraft migrate command using registered migration engines
- **CLI_PHASE_EXECUTION_COMMON**: Shared phase execution semantics for deploy and rollback
- **CLI_PLAN**: Plan command (dry-run)
- **CLI_RELEASES**: stagecraft releases list/show commands
- **CLI_ROLLBACK**: stagecraft rollback command
- **CORE_BACKEND_PROVIDER_CONFIG_SCHEMA**: Provider-scoped backend configuration schema
- **CORE_BACKEND_REGISTRY**: Backend provider registry system
- **CORE_COMPOSE**: Docker Compose integration
- **CORE_CONFIG**: Config loading and validation
- **CORE_ENV_RESOLUTION**: Environment resolution and context
- **CORE_EXECUTIL**: Process execution utilities
- **CORE_LOGGING**: Structured logging helpers
- **CORE_MIGRATION_REGISTRY**: Migration engine registry system
- **CORE_PLAN**: Deployment planning engine
- **CORE_STATE**: State management (release history)
- **CORE_STATE_CONSISTENCY**: State durability and read-after-write guarantees
- **CORE_STATE_TEST_ISOLATION**: State test isolation for CLI commands
- **DEV_COMPOSE_INFRA**: Compose infra up/down for dev
- **DEV_HOSTS**: /etc/hosts management
- **DEV_MKCERT**: mkcert integration for local HTTPS
- **DEV_PROCESS_MGMT**: Process lifecycle management
- **DEV_TRAEFIK**: Traefik dev config generation
- **GOV_STATUS_ROADMAP**: stagecraft status roadmap command
- **GOV_V1_CORE**: Governance Core for v1
- **MIGRATION_ENGINE_RAW**: Raw SQL migration engine implementation
- **PROVIDER_BACKEND_ENCORE**: Encore.ts BackendProvider implementation
- **PROVIDER_BACKEND_GENERIC**: Generic command-based BackendProvider implementation
- **PROVIDER_BACKEND_INTERFACE**: BackendProvider interface definition
- **PROVIDER_CI_INTERFACE**: CIProvider interface definition
- **PROVIDER_CLOUD_INTERFACE**: CloudProvider interface definition
- **PROVIDER_FRONTEND_GENERIC**: Generic dev command FrontendProvider
- **PROVIDER_FRONTEND_INTERFACE**: FrontendProvider interface definition
- **PROVIDER_NETWORK_INTERFACE**: NetworkProvider interface definition
- **PROVIDER_NETWORK_TAILSCALE**: Tailscale NetworkProvider implementation
- **PROVIDER_SECRETS_INTERFACE**: SecretsProvider interface definition

## Coverage Status

Current test coverage targets:
- **Core packages** (`pkg/config`, `internal/core`): Target 80%+
- **CLI layer** (`internal/cli`): Target 70%+
- **Drivers** (`internal/drivers`): Target 70%+
- **Overall**: Target 60%+ (increasing to 80% as project matures)

## How to Update

This file should be regenerated when `spec/features.yaml` changes. To update:

```bash
# Run the generator script
./scripts/generate-implementation-status.sh
```

For detailed feature specifications, see the individual spec files referenced in the table above.

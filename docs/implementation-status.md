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
| CORE_CONFIG | Config loading and validation | todo | bart | [core/config.md](../spec/core/config.md) | [config_test.go](../pkg/config/config_test.go) |
| CORE_PLAN | Deployment planning engine | todo | bart | [core/plan.md](../spec/core/plan.md) | [plan_test.go](../internal/core/plan_test.go) |

### CLI Commands

| ID | Title | Status | Owner | Spec | Tests |
|----|-------|--------|-------|------|-------|
| CLI_INIT | Project bootstrap command | todo | bart | [commands/init.md](../spec/commands/init.md) | [init_test.go](../internal/cli/commands/init_test.go), [init_smoke_test.go](../test/e2e/init_smoke_test.go) |
| CLI_PLAN | Plan command (dry-run) | todo | bart | [commands/plan.md](../spec/commands/plan.md) | [plan_test.go](../internal/cli/commands/plan_test.go) |
| CLI_DEPLOY | Deploy command | todo | bart | [commands/deploy.md](../spec/commands/deploy.md) | [deploy_test.go](../internal/cli/commands/deploy_test.go) |

### Drivers

| ID | Title | Status | Owner | Spec | Tests |
|----|-------|--------|-------|------|-------|
| DRIVER_DO | DigitalOcean driver | todo | bart | [drivers/do.md](../spec/drivers/do.md) | [do_test.go](../internal/drivers/do/do_test.go) |

## Implementation Notes

### Completed Features

- **ARCH_OVERVIEW**: Basic architecture documentation and project structure established
- **DOCS_ADR**: ADR process documented with initial architecture decision (ADR 0001)

### In Progress

_None currently_

### Planned Next Steps

1. Complete `CORE_CONFIG` implementation with full validation
2. Implement `CLI_INIT` command with config scaffolding
3. Build `CORE_PLAN` deployment planning engine
4. Add `DRIVER_DO` for DigitalOcean integration

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


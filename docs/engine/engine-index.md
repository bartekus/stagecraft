---
status: canonical
scope: meta
---

<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<!--

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

-->

# Stagecraft Engine Documentation Index

This document enumerates the **AI-critical** technical documentation that should be routinely opened in Cursor when working on different feature types. This index helps maintain cost-efficient AI workflows while ensuring proper context.

## Core Principles

- **One feature per thread** - See [CONTRIBUTING_CURSOR.md](../governance/CONTRIBUTING_CURSOR.md)
- **Spec-first** - Always start with the relevant spec(s)
- **Test-aware** - Include test files in context
- **Minimal scope** - Only open files directly relevant to the current feature

> **Note**: See [docs/README.md](../README.md) for complete documentation structure and navigation guide.

---

## Feature Type: CLI Commands

### Required Specs
- `spec/commands/<command>.md` (e.g., `build.md`, `deploy.md`, `plan.md`)
- `spec/core/global-flags.md` (for any command)
- `spec/governance/GOV_CORE.md` (for commands that interact with state)

### Required Code Files
- `internal/cli/root.go` (command registration)
- `internal/cli/commands/<command>.go` (implementation)
- `internal/cli/commands/<command>_test.go` (tests)

### Related Core Files (as needed)
- `internal/core/phases_*.go` (for build/deploy/rollback)
- `internal/core/plan.go` (for plan command)
- `internal/core/state/state.go` (for stateful commands)

### Related Docs
- `docs/engine/analysis/CLI_<COMMAND>_ANALYSIS.md` (if exists)
- `docs/engine/outlines/CLI_<COMMAND>_IMPLEMENTATION_OUTLINE.md` (if exists)
- `docs/context-handoff/*-to-CLI_<COMMAND>.md` (if exists)

### Example: Working on `CLI_BUILD`
```
Open:
- spec/commands/build.md
- spec/core/global-flags.md
- internal/cli/commands/build.go
- internal/cli/commands/build_test.go
- internal/core/phases_build.go
- internal/cli/root.go (for registration context)
```

---

## Feature Type: Core Engine (State, Plan, Config, etc.)

### Required Specs
- `spec/core/<feature>.md` (e.g., `state.md`, `plan.md`, `config.md`)
- `spec/governance/GOV_CORE.md` (for state-related features)

### Required Code Files
- `internal/core/<feature>.go` (or subdirectory)
- `internal/core/<feature>_test.go` (or `internal/core/<feature>/<feature>_test.go`)
- `pkg/config/*.go` (for config features)
- `pkg/executil/*.go` (for executil features)
- `pkg/logging/*.go` (for logging features)

### Related Docs
- `docs/engine/analysis/*.md` (if exists for this feature)
- `docs/context-handoff/*-to-CORE_<FEATURE>.md` (if exists)

### Example: Working on `CORE_STATE`
```
Open:
- spec/core/state.md
- spec/core/state-consistency.md
- spec/core/state-test-isolation.md
- spec/governance/GOV_CORE.md
- internal/core/state/state.go
- internal/core/state/state_test.go
```

---

## Feature Type: Provider Implementation

### Required Specs
- `spec/providers/<provider-type>/interface.md` (e.g., `backend/interface.md`, `migration/interface.md`)
- `spec/providers/<provider-type>/<provider-name>.md` (for specific provider, e.g., `backend/encore-ts.md`)
- `spec/core/backend-registry.md` (for backend providers)
- `spec/core/migration-registry.md` (for migration providers)

### Required Code Files
- `pkg/providers/<provider-type>/<provider-type>.go` (interface/type definitions, e.g., `backend.go`, `migration.go`)
- `pkg/providers/<provider-type>/registry.go` (registry implementation)
- `internal/providers/<provider-type>/<provider-name>/<provider-name>.go` (implementation, may be in subdirectory)
- `internal/providers/<provider-type>/<provider-name>/<provider-name>_test.go` (tests)

### Related Docs
- `docs/providers/<provider-type>.md` (if exists)
- `spec/core/backend-registry.md` (for backend registry details)
- `spec/core/migration-registry.md` (for migration registry details)

### Example: Working on `PROVIDER_BACKEND_ENCORE`
```
Open:
- spec/providers/backend/interface.md
- spec/providers/backend/encore-ts.md
- spec/core/backend-registry.md
- pkg/providers/backend/backend.go
- pkg/providers/backend/registry.go
- internal/providers/backend/encorets/encorets.go
- internal/providers/backend/encorets/encorets_test.go
```

---

## Feature Type: Migration Engine

### Required Specs
- `spec/providers/migration/raw.md` (for raw SQL migrations)
- `spec/core/migration-registry.md`
- `spec/commands/migrate-basic.md` (if working on CLI integration)

### Required Code Files
- `pkg/providers/migration/migration.go` (interface/type definitions)
- `pkg/providers/migration/registry.go` (registry implementation)
- `internal/providers/migration/raw/raw.go` (implementation)
- `internal/providers/migration/raw/raw_test.go` (tests)
- `internal/cli/commands/migrate.go` (if touching CLI)

### Related Docs
- `docs/providers/migrations.md`

---

## Feature Type: Compose Integration

### Required Specs
- `spec/core/compose.md`

### Required Code Files
- `internal/compose/compose.go`
- `internal/compose/compose_test.go`

---

## Feature Type: Governance / Spec Compliance

### Required Specs
- `spec/governance/GOV_CORE.md`
- All relevant specs being governed

### Required Code Files
- Files implementing the governed features

### Related Docs
- `docs/engine/analysis/GOV_CORE_IMPLEMENTATION_ANALYSIS.md` (if exists)
- `docs/context-handoff/GOV_CORE-to-*.md` (if exists)

---

## Universal Reference Docs

These docs are useful across many feature types but should be opened explicitly when needed:

### Specs
- `spec/overview.md` - High-level architecture
- `spec/features.yaml` - Feature registry and dependencies
- `spec/scaffold/stagecraft-dir.md` - Project structure

### Implementation Docs
- `docs/narrative/stagecraft-spec.md` - Complete spec reference (index)
- `docs/features/OVERVIEW.md` - Feature status overview
- `docs/engine/status/implementation-status.md` - Implementation tracking (generated)

### Context Handoff
- `docs/context-handoff/INDEX.md` - Handoff doc index
- `docs/context-handoff/<specific-handoff>.md` - Feature-specific handoff

---

## Quick Reference by Feature ID Prefix

### CLI_* (Commands)
- Spec: `spec/commands/<command-name>.md`
- Code: `internal/cli/commands/<command>.go` + `_test.go`
- Root: `internal/cli/root.go`

### CORE_* (Core Engine)
- Spec: `spec/core/<feature-name>.md`
- Code: `internal/core/<feature>.go` or `internal/core/<feature>/*.go`
- Tests: Matching `_test.go` files

### PROVIDER_* (Providers)
- Spec: `spec/providers/<type>/interface.md` + `<provider>.md`
- Code: `internal/providers/<type>/<provider>/<provider>.go` + `_test.go` (may be in subdirectory)
- Interface: `pkg/providers/<type>/<type>.go` (e.g., `backend.go`, `migration.go`)
- Registry: `pkg/providers/<type>/registry.go`

### MIGRATION_* (Migrations)
- Spec: `spec/providers/migration/*.md`
- Code: `internal/providers/migration/<engine>/<engine>.go` (e.g., `raw/raw.go`)
- Interface: `pkg/providers/migration/migration.go`

### GOV_* (Governance)
- Spec: `spec/governance/*.md`
- Code: All governed feature implementations

---

## Notes

- **Don't open everything** - Only open files directly relevant to the current feature
- **Use attachments** - For large specs, attach them rather than pasting inline
- **Close threads** - When a feature is done, close the thread and start fresh for the next feature
- **Check context-handoff** - Before starting, check if there's a handoff doc for your feature

For detailed workflow guidance, see [CONTRIBUTING_CURSOR.md](../governance/CONTRIBUTING_CURSOR.md).


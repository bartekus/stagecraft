---
status: canonical
scope: v1
---

# Stagecraft CLI - Specification Index

> **Note**: This is an index to the canonical specification. The actual specification content lives in the [`spec/`](../spec/) directory tree.
>
> **Related Documents:**
> - [`implementation-roadmap.md`](implementation-roadmap.md) - Implementation phases and feature catalog
> - [`../spec/features.yaml`](../spec/features.yaml) - Feature tracking (source of truth)
> - [`../adr/0001-architecture.md`](../adr/0001-architecture.md) - Architecture decisions
> - [`../../blog/01-why-not-kamal.md`](../../blog/01-why-not-kamal.md) - Design rationale

## Overview

Stagecraft is a Go-based orchestration CLI for local development and deployment of multi-service applications.

**Design Principles:**
- Local-first DX - one command to spin up full local infra, HTTPS, backend, frontend
- Docker Compose + docker-rollout for runtime orchestration
- Tailscale or Headscale for multi-host networking
- Provider model so Encore.ts, Vite, DO CLI, GitHub CLI and others plug in cleanly
- Configuration driven through a single `stagecraft.yml` plus one canonical `docker-compose.yml`

**Key Differentiators:**
- Uses Compose instead of one-off docker run
- Supports multi-host via mesh network (Tailscale) instead of static IPs only
- Has a first-class local dev story (Encore dev server, Vite dev server, mkcert HTTPS)

## Commands

- [Init](../spec/commands/init.md) - Initialize Stagecraft in a project
- [Build](../spec/commands/build.md) - Build Docker images for deployment
- [Deploy](../spec/commands/deploy.md) - Deploy a given version to a named environment
- [Rollback](../spec/commands/rollback.md) - Roll back an environment to a previous version
- [Releases](../spec/commands/releases.md) - List and show release history
- [Plan](../spec/commands/plan.md) - Plan command (dry-run)
- [Dev Basic](../spec/commands/dev-basic.md) - Basic dev command that delegates to backend provider
- [Migrate Basic](../spec/commands/migrate-basic.md) - Basic migrate command using registered migration engines
- [Commit Suggest](../spec/commands/commit-suggest.md) - Generate commit discipline suggestions

## Core

- [Overview](../spec/overview.md) - Architecture documentation and project overview
- [Config](../spec/core/config.md) - Config loading and validation
- [Logging](../spec/core/logging.md) - Structured logging helpers
- [Executil](../spec/core/executil.md) - Process execution utilities
- [Plan](../spec/core/plan.md) - Deployment planning engine
- [Environment Resolution](../spec/core/env-resolution.md) - Environment resolution and context
- [State](../spec/core/state.md) - State management (release history)
- [State Test Isolation](../spec/core/state-test-isolation.md) - State test isolation for CLI commands
- [State Consistency](../spec/core/state-consistency.md) - State durability and read-after-write guarantees
- [Compose](../spec/core/compose.md) - Docker Compose integration
- [Backend Registry](../spec/core/backend-registry.md) - Backend provider registry system
- [Migration Registry](../spec/core/migration-registry.md) - Migration engine registry system
- [Backend Provider Config](../spec/core/backend-provider-config.md) - Provider-scoped backend configuration schema
- [Phase Execution Common](../spec/core/phase-execution-common.md) - Shared phase execution semantics
- [Global Flags](../spec/core/global-flags.md) - Global flags (--env, --config, --verbose, --dry-run)

## Providers

### Backend
- [Interface](../spec/providers/backend/generic.md#interface) - BackendProvider interface definition
- [Generic](../spec/providers/backend/generic.md) - Generic command-based BackendProvider implementation
- [Encore.ts](../spec/providers/backend/encore-ts.md) - Encore.ts BackendProvider implementation

### Frontend
- [Interface](../spec/providers/frontend/interface.md) - FrontendProvider interface definition
- [Generic](../spec/providers/frontend/generic.md) - Generic dev command FrontendProvider

### Migration
- [Raw](../spec/providers/migration/raw.md) - Raw SQL migration engine implementation

### Network
- [Interface](../spec/providers/network/interface.md) - NetworkProvider interface definition

### Cloud
- [Interface](../spec/providers/cloud/interface.md) - CloudProvider interface definition

### CI
- [Interface](../spec/providers/ci/interface.md) - CIProvider interface definition

### Secrets
- [Interface](../spec/providers/secrets/interface.md) - SecretsProvider interface definition

## Governance

- [GOV_CORE](../spec/governance/GOV_CORE.md) - Governance Core for v1

## Scaffold

- [Stagecraft Directory](../spec/scaffold/stagecraft-dir.md) - Stagecraft directory structure

---

**Source of Truth**: All specification content is in the [`spec/`](../spec/) directory. This index is for navigation only.

# Stagecraft

<p align="center">
  <img src="https://img.shields.io/badge/License-AGPL_v3%2B-blue.svg" alt="AGPL-3.0-or-later" />
  <a href="https://github.com/bartekus/stagecraft/actions/workflows/governance.yml"><img src="https://github.com/bartekus/stagecraft/actions/workflows/governance.yml/badge.svg" alt="Governance Status" /></a>
</p>

Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

- **Local-first DX** – one command to spin up full local infra, HTTPS, and services.
- **Compose + docker-rollout** – predictable runtime orchestration.
- **Mesh networking by default** – Tailscale for single-host → multi-host evolution.
- **Provider model** – Generic, Encore.ts, Vite, DigitalOcean, etc. plug in cleanly.
- **Configuration-driven** – one `stagecraft.yml` plus a canonical `docker-compose.yml`.

> ⚠️ **Status**: Early WIP / experimental. See [Implementation Status](docs/engine/status/implementation-status.md).

---

## Who is this for?

- **Evaluators & Users**: Read this README and the [Getting Started Guide](docs/guides/getting-started.md).
- **New Contributors**: Go to [`CONTRIBUTING.md`](CONTRIBUTING.md) for setup and rules.
- **AI Agents**: STOP. Read [`Agent.md`](Agent.md) immediately for your constraints and protocol.

---

## Goals

- **Trivial Scaling**: `docker compose up` laptop → single VM → multi-host cluster.
- **Provider Agnostic**: Encore.ts (backend), Vite (frontend), DigitalOcean (cloud), etc.
- **Smooth Path**: Local → Staging → Prod without shifting mental models.

See [`docs/narrative/stagecraft-spec.md`](docs/narrative/stagecraft-spec.md) for the full vision.

---

## Quickstart: basic Node app

```bash
# 1. Clone & Build
git clone https://github.com/bartekus/stagecraft.git
cd stagecraft
go build ./cmd/stagecraft

# 2. Run Example
cd examples/basic-node
../stagecraft dev
```

See `examples/basic-node/stagecraft.yml` for the configuration.

---

## High-level architecture

- **Config**: `stagecraft.yml` (environments, providers, roles).
- **Infra**: Canonical `docker-compose.yml`.
- **Providers**: Pluggable backend, frontend, cloud, and network handling.
- **Mesh**: Tailscale/Headscale for zero-trust networking.

---

## Commands

> [!NOTE]
> This is a high-level reference. See **[`docs/__generated__/ai-agent/COMMAND_CATALOG.md`](docs/__generated__/ai-agent/COMMAND_CATALOG.md)** for the authoritative list.

- `stagecraft dev` - Start development environment
- `stagecraft migrate` - Run database migrations
- `stagecraft build` - Build Docker images
- `stagecraft deploy` - Deploy to environments
- `stagecraft plan` - Dry-run deployment plan
- `stagecraft rollback` - Rollback to previous version
- `stagecraft releases` - Show deployment history
- `stagecraft init` - Bootstrap project
- `stagecraft infra up` - Provision infrastructure
- `stagecraft gov` - Governance checks
- `stagecraft context build` - Build AI context

---

## AI Context Pipeline

Stagecraft includes a deterministic engine to generate AI-readable context.

See **[`docs/__generated__/ai-agent/README.md`](docs/__generated__/ai-agent/README.md)** for:
- Repository Index & Statistics
- Documentation Catalogs
- Spec & Command Indexes

---

## Project structure

```text
stagecraft/
  cmd/             # CLI entry points
  internal/        # Private implementation
  pkg/             # Public libraries
  spec/            # Feature specifications (Source of Truth)
  docs/            # Documentation
    __generated__/ # AI-Agent outputs
    engine/        # Technical docs
    narrative/     # Planning docs
    governance/    # Process rules
  examples/        # Usage examples
```

---

## Documentation

- **[CONTRIBUTING.md](CONTRIBUTING.md)**: Human contributor workflow and setup.
- **[Agent.md](Agent.md)**: AI Agent development protocol (Mandatory for AI).
- **[docs/README.md](docs/README.md)**: Master documentation index.
- **[docs/engine/status/implementation-status.md](docs/engine/status/implementation-status.md)**: Feature matrix.

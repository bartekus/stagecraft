# Reimagining Kamal: A CLI-first Deployment Brain in Go

> Related ADRs: [ADR-0001](../docs/adr/0001-architecture.md)
> Status: Draft
> Date: 2025-01-XX

## Introduction

Stagecraft is a Go-based CLI that orchestrates application deployment and infrastructure workflows. This post outlines the vision, goals, and initial architecture decisions that shaped the project.

## The Problem

Modern deployment tools often force developers to choose between:
- **Local-first tools** that work great on a laptop but don't scale
- **Production-focused tools** that require complex setup and don't support local development well
- **Cloud-specific tools** that lock you into a single provider

We wanted a tool that:
- Works seamlessly from local development to production
- Supports multi-host deployments without complex networking setup
- Integrates cleanly with existing tools (Docker Compose, Tailscale, etc.)
- Provides a world-class developer experience

## The Vision

Stagecraft aims to be "Kamal, but for Docker Compose + Tailscale + Encore.ts + Vite":

- **Local-first DX** – one command to spin up full local infra, HTTPS, backend, and frontend
- **Compose + docker-rollout** – predictable runtime orchestration
- **Mesh networking by default** – Tailscale or Headscale for single-host → multi-host evolution
- **Provider model** – Encore.ts, Vite, DigitalOcean, GitHub Actions and others plug in cleanly
- **Configuration-driven** – one `stagecraft.yml` plus a canonical `docker-compose.yml`

## Architecture Decisions

### Why Go?

Go provides:
- Excellent CLI tooling (Cobra)
- Strong concurrency for orchestrating multiple services
- Single binary deployment
- Great ecosystem for cloud integrations

### Why Spec-First?

We adopted a spec-first, TDD-heavy, ADR-driven workflow:
- Every feature has a spec in `spec/`
- Features are tracked in `spec/features.yaml`
- Tests reference specs
- ADRs document major decisions

This ensures:
- Clear traceability from spec to code
- Better AI-assisted development (Cursor, etc.)
- Living documentation

### Layered Architecture

We chose a clear separation:
- **CLI Layer** (`internal/cli/`) – User interaction, flags, output
- **Core Layer** (`internal/core/`) – Business logic, planning, state
- **Driver Layer** (`internal/drivers/`) – Platform integrations
- **Public APIs** (`pkg/`) – Reusable libraries

See [ADR-0001](../docs/adr/0001-architecture.md) for details.

## Current Status

Stagecraft is in early development. Currently implemented:
- ✅ Basic CLI structure with Cobra
- ✅ Config loading and validation (`pkg/config`)
- ✅ `stagecraft init` command (stub)
- ✅ Test infrastructure
- ✅ CI/CD pipeline with linting and coverage checks

Coming soon:
- Deployment planning engine
- DigitalOcean driver
- Full `init` implementation
- `dev` command for local development

## Next Steps

1. Complete core config system
2. Implement deployment planning
3. Add first driver (DigitalOcean)
4. Build local dev workflow

Follow the journey:
- [GitHub Repository](https://github.com/your-org/stagecraft)
- [Architecture Documentation](../docs/architecture.md)
- [Implementation Status](../docs/implementation-status.md)


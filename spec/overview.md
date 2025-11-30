# Stagecraft – Project Overview

Stagecraft is a deployment and infrastructure orchestration CLI written in Go.

## High-Level Goals

- Provide a **world-class developer UX** for app deployment:
    - Simple commands for common workflows (`init`, `plan`, `deploy`, `rollback`).
    - Sensible defaults, clear logs, strong safety rails.
- Provide a **composable core**:
    - Separation between CLI, core planning engine, and provider drivers (e.g. DigitalOcean, GitHub Actions).
    - Plugin/adaptor system for additional targets and tools.

## Core Concepts

- **Project** – A repository containing one or more services and a Stagecraft config.
- **Environment** – A named target (e.g. `local`, `staging`, `prod`) with its own configuration.
- **Plan** – A dry-run representation of what Stagecraft will do for a given command.
- **Driver** – An integration that knows how to apply a plan to a particular platform or toolchain.
- **Plugin** – An extension providing additional commands, drivers, or config hooks.

## Command Surface (Initial)

1. `stagecraft init` – Bootstrap Stagecraft in an existing project.
2. `stagecraft plan` – Compute a deployment plan for a target environment.
3. `stagecraft deploy` – Apply a plan to a target environment.
4. `stagecraft status` – Show the current state of deployed resources.
5. `stagecraft doctor` – Run diagnostics and checks.

Each of these commands will have its own dedicated spec under `spec/commands/`.

## Non-Goals (for v0)

- No GUI or web dashboard.
- No attempt to be a full CI system (we integrate with GitHub Actions etc., we don’t replace them).
- No proprietary binary formats or opaque state; favour human-readable config and logs.

See `spec/features.yaml` for the authoritative feature list and status.

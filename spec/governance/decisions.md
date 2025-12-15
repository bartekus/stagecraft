<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<!--

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

-->

---
type: decision-log
scope: core
status: canonical
last_updated: 2025-12-10
owner: bart
---

## DECISION-001 — Architecture, Documentation Model, and Provider Boundaries

### Status
Accepted

### Decision

Stagecraft adopts the following permanent architectural constraints:

1. **Canonical truth**
  - Behavioral truth lives in `spec/`
  - Executable truth lives in code
  - AI operational truth lives in `ai.agent/`
  - Human documentation is generated only and lives in `docs/__generated__/`

2. **Documentation model**
  - `docs/` is not a source of truth
  - No archival directories; git history is the archive
  - No ADR files; decisions live here and nowhere else

3. **Provider boundaries**
  - Core (`internal/core`) MUST NOT import providers
  - Providers are the only cloud / infra abstraction layer
  - A separate “driver” layer is explicitly rejected

4. **Terraform usage**
  - Terraform is permitted only for substrate provisioning
  - Stagecraft owns orchestration, planning, execution, and lifecycle

5. **CLI / Engine / Agent roles**
  - CLI: UX, validation, exit codes
  - Engine: deterministic planning and action graph generation
  - Agent/Daemon: execution substrate and long-running orchestration

### Explicit Rejection

The concept of a provider-adjacent “driver” abstraction is rejected.
Any prior reference (for example `DRIVER_DO`) is obsolete and removed.

### Consequences

- No duplicate architectural documents
- No drift between docs and specs
- Clear, enforceable boundaries for humans and AI agents

---

## DECISION-002 — Failure Classification Taxonomy

### Status
Accepted

### Decision
Adoption of the 7-class failure taxonomy and strict exit code mapping defined in `spec/governance/GOV_CLI_EXIT_CODES.md`.

### Consequences
- All active skills (e.g., `failure_lens`) must output one of the 7 defined classes.
- CLI exit codes are now rigidly mapped: 0 (Success), 1 (User/Config), 2 (External/Transient), 3 (Internal/Bug).

---

## DECISION-003 — DRIVER_DO Obsolescence

### Status
Accepted

### Decision
The concept of `DRIVER_DO` is officially obsolete.

### Explicit Constraints
- No feature shall depend on `DRIVER_DO`.
- No spec shall reference `DRIVER_DO`.
- Implementation must use `PROVIDER_CLOUD_DO` exclusively.

### Consequences
- Any existing references must be removed.
- Future "drivers" are rejected in favor of Providers.

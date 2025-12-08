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

# Stagecraft Documentation

This directory contains all project documentation organized by purpose and lifecycle.

## Quick Navigation

- **[engine/](engine/)** - Implementation-aligned, AI-critical technical docs
- **[narrative/](narrative/)** - Human-facing planning, roadmaps, architecture
- **[governance/](governance/)** - Process, discipline, and workflow docs
- **[archive/](archive/)** - Historical, completed work (frozen)
- **[spec/](../spec/)** - Source of truth for all specifications

## Documentation Structure

### 🎯 Start Here

- **[README.md](../README.md)** - High-level project description
- **[Agent.md](../Agent.md)** - Core governance and development protocol (read this first)
- **[CONTRIBUTING.md](../CONTRIBUTING.md)** - Contribution guidelines
- **[engine/engine-index.md](engine/engine-index.md)** - Quick reference for which docs to open per feature type

### 📋 Specifications (Source of Truth)

- **[spec/](../spec/)** - All feature specifications
  - `spec/commands/` - CLI command specifications
  - `spec/core/` - Core engine specifications
  - `spec/providers/` - Provider interface specifications
  - `spec/governance/` - Governance specifications
  - `spec/features.yaml` - Feature registry (source of truth)

- **[narrative/stagecraft-spec.md](narrative/stagecraft-spec.md)** - Specification index (links to spec/ files)

### 🔧 Engine Documentation (`engine/`)

**AI-critical technical documentation** for implementation work:

- **`engine/analysis/`** - Feature implementation analysis
  - `CLI_PLAN_ANALYSIS.md`
  - `PROJECT_STRUCTURE_ANALYSIS.md`
  - `GOV_V1_CORE_IMPLEMENTATION_ANALYSIS.md`
  - `TEMPLATE.md`

- **`engine/outlines/`** - Implementation outlines and templates
  - `CLI_PLAN_IMPLEMENTATION_OUTLINE.md`
  - `IMPLEMENTATION_OUTLINE_TEMPLATE.md`

- **`engine/status/`** - Generated status tracking
  - `implementation-status.md` - Auto-generated from `spec/features.yaml`
  - `README.md` - Explains generation process

- **`engine/engine-index.md`** - Quick reference: which files to open per feature type

### 📖 Narrative Documentation (`narrative/`)

**Human-facing planning and overview docs:**

- `architecture.md` - High-level architecture overview
- `implementation-roadmap.md` - Implementation roadmap and feature catalog
- `stagecraft-spec.md` - Specification index (links to spec/)
- `FUTURE_ENHANCEMENTS.md` - Future feature ideas
- `V2_FEATURES.md` - v2 feature overview

### 🛡️ Governance Documentation (`governance/`)

**Process, discipline, and workflow docs:**

- `CONTRIBUTING_CURSOR.md` - AI-assisted development workflow
- `COMMIT_MESSAGE_ANALYSIS.md` - Commit message format analysis
- `STRATEGIC_DOC_MIGRATION.md` - Strategic document handling guide

### 📦 Archive (`archive/`)

**Historical, completed work (frozen):**

- `registry-implementation-summary.md` - Registry implementation summary (superseded by spec/)

## Other Directories

- **`adr/`** - Architecture Decision Records
- **`context-handoff/`** - Feature handoff documents for AI agents
- **`features/`** - Feature documentation (generated)
- **`guides/`** - User-facing guides
- **`providers/`** - Provider implementation docs
- **`reference/`** - API/reference documentation (auto-generated)
- **`design/`** - Design documents
- **`todo/`** - TODO items and planning

## Documentation Lifecycle

Documents use frontmatter to track lifecycle:

```yaml
---
status: active | canonical | archived
scope: v1 | v2 | meta
feature: CLI_PLAN          # optional
spec: ../spec/commands/plan.md  # optional
superseded_by: ../Agent.md # optional
---
```

- **active** - Still guiding current or near-future work
- **canonical** - Stable, long-lived reference (e.g., architecture, governance)
- **archived** - Work is done, content describes "how we did it" (historical)

## Documentation Workflow

When working on a feature:

1. **Read the spec** - `spec/<domain>/<feature>.md`
2. **Check analysis** - `engine/analysis/<FEATURE_ID>_ANALYSIS.md` (if exists)
3. **Review implementation outline** - `engine/outlines/<FEATURE_ID>_IMPLEMENTATION_OUTLINE.md` (if exists)
4. **Check context-handoff** - `context-handoff/INDEX.md` for handoff docs
5. **Reference engine-index** - `engine/engine-index.md` for file location guidance

## For AI Assistants

When using Cursor or other AI tools:

- See **[governance/CONTRIBUTING_CURSOR.md](governance/CONTRIBUTING_CURSOR.md)** for workflow guidance
- See **[engine/engine-index.md](engine/engine-index.md)** for which files to open per feature type
- Prefer opening "engine" docs over "narrative" docs during implementation work
- Use context-handoff docs when transitioning between features

## Generated Documentation

Some documentation is auto-generated:

- **`engine/status/implementation-status.md`** - Generated from `spec/features.yaml`
  - Regenerate with: `./scripts/generate-implementation-status.sh`
  - See `engine/status/README.md` for details

- **`features/OVERVIEW.md`** - Generated from `spec/features.yaml`
  - Regenerate with: `go run ./cmd/gen-features-overview`

- **`reference/cli.md`** - Auto-generated CLI reference

**Never edit generated files manually.** Update the source (usually `spec/features.yaml`) and regenerate.

## Principles

- **spec/ is source of truth** - All specifications live in `spec/`
- **docs/ is implementation + process** - Strongly structured by purpose
- **archive/ is history** - Completed work, frozen for reference
- **No duplication** - Docs reference spec/, don't duplicate it
- **Deterministic generation** - Status docs are generated, not manually maintained

For questions or suggestions about documentation organization, see [engine/analysis/PROJECT_STRUCTURE_ANALYSIS.md](engine/analysis/PROJECT_STRUCTURE_ANALYSIS.md).

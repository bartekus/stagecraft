<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<!--

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

-->

# Stagecraft Documentation

This directory contains all project documentation. This guide explains the organization and helps you find what you need.

## Documentation Categories

### 🎯 Start Here

- **[Agent.md](../Agent.md)** - Core governance and development protocol (read this first)
- **[CONTRIBUTING.md](../CONTRIBUTING.md)** - Contribution guidelines
- **[CONTRIBUTING_CURSOR.md](./CONTRIBUTING_CURSOR.md)** - AI-assisted development workflow
- **[engine-index.md](./engine-index.md)** - Quick reference for which docs to open per feature type

### 📋 Specifications

- **[stagecraft-spec.md](./stagecraft-spec.md)** - Complete specification reference
- **[spec/](../spec/)** - Source of truth for all feature specifications
  - `spec/commands/` - CLI command specifications
  - `spec/core/` - Core engine specifications
  - `spec/providers/` - Provider interface specifications
  - `spec/governance/` - Governance specifications

### 🔧 Engine Documentation (Technical / Implementation)

These are the "AI-critical" technical docs that should be opened when working on features:

- **Analysis Documents** (`analysis/`)
  - Implementation analysis for specific features
  - Template for creating new analyses
  - Example: `GOV_V1_CORE_IMPLEMENTATION_ANALYSIS.md`

- **Implementation Outlines** (root level)
  - `CLI_*_IMPLEMENTATION_OUTLINE.md` - Command implementation plans
  - `IMPLEMENTATION_OUTLINE_TEMPLATE.md` - Template for new outlines

- **Implementation Status**
  - `implementation-status.md` - Current implementation tracking
  - `registry-implementation-summary.md` - Provider registry details

- **Feature Documentation** (`features/`)
  - `OVERVIEW.md` - Feature registry and status overview

- **Provider Documentation** (`providers/`)
  - Provider-specific implementation guides
  - Example: `backend.md`, `migrations.md`

- **Context Handoff** (`context-handoff/`)
  - Deterministic context for AI agents transitioning between features
  - See `context-handoff/INDEX.md` for navigation

- **Project Structure** (`PROJECT_STRUCTURE_ANALYSIS.md`)
  - Analysis of project structure and improvement recommendations

### 📖 Narrative Documentation (Human-Facing / Planning)

These are planning, roadmap, and high-level docs:

- **Architecture** (`architecture.md`, `adr/`)
  - High-level architecture overview
  - Architecture Decision Records

- **Roadmaps & Planning**
  - `implementation-roadmap.md` - Implementation roadmap
  - `FUTURE_ENHANCEMENTS.md` - Future feature ideas

- **User Guides** (`guides/`)
  - User-facing documentation
  - Example: `getting-started.md`

- **Reference** (`reference/`)
  - API and command reference documentation
  - Example: `cli.md` (auto-generated)

## Documentation Workflow

When working on a feature, follow this order:

1. **Read the spec** - `spec/<domain>/<feature>.md`
2. **Check analysis** - `docs/analysis/<FEATURE_ID>.md` (if exists)
3. **Review implementation outline** - `docs/<FEATURE_ID>_IMPLEMENTATION_OUTLINE.md` (if exists)
4. **Check context-handoff** - `docs/context-handoff/INDEX.md` for handoff docs
5. **Reference engine-index** - `docs/engine-index.md` for file location guidance

## For AI Assistants

When using Cursor or other AI tools:

- See **[CONTRIBUTING_CURSOR.md](./CONTRIBUTING_CURSOR.md)** for workflow guidance
- See **[engine-index.md](./engine-index.md)** for which files to open per feature type
- Prefer opening "engine" docs over "narrative" docs during implementation work
- Use context-handoff docs when transitioning between features

## Directory Structure

```
docs/
├── adr/                          # Architecture Decision Records
├── analysis/                     # Implementation analysis documents
├── context-handoff/              # Feature handoff documents
├── features/                     # Feature documentation
├── guides/                       # User-facing guides
├── providers/                    # Provider implementation docs
├── reference/                    # API/reference documentation
├── architecture.md               # Architecture overview (narrative)
├── CLI_*_ANALYSIS.md            # Command analysis (engine)
├── CLI_*_IMPLEMENTATION_OUTLINE.md  # Implementation outlines (engine)
├── CONTRIBUTING_CURSOR.md        # AI workflow guide
├── engine-index.md               # Quick reference index
├── FUTURE_ENHANCEMENTS.md        # Future plans (narrative)
├── implementation-roadmap.md     # Roadmap (narrative)
├── implementation-status.md      # Status tracking (engine)
├── PROJECT_STRUCTURE_ANALYSIS.md # Structure analysis (engine)
├── registry-implementation-summary.md  # Registry details (engine)
└── stagecraft-spec.md            # Complete spec reference (engine)
```

## Notes

- **Engine docs** = Technical implementation details, specs, analysis, outlines
- **Narrative docs** = Planning, roadmaps, architecture overviews, user guides
- All docs follow the spec-first, test-first, feature-bounded principles from [Agent.md](../Agent.md)

For questions or suggestions about documentation organization, see [PROJECT_STRUCTURE_ANALYSIS.md](./PROJECT_STRUCTURE_ANALYSIS.md).


<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<!--

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

-->

# Context Handoff Documents Index

> **Primary Reference**: All new context handoff notes should be added to `CONTEXT_LOG.md` instead of creating new files. This index is maintained for historical reference.

This index lists all context handoff documents and when to use them. These documents provide deterministic context for AI agents transitioning between feature implementations.

**Note**: Individual handoff documents are being consolidated into `CONTEXT_LOG.md`. New context handoffs MUST be added to the context log rather than creating new files.

## Purpose

Context handoff documents ensure:
- ✅ Clear understanding of what was just completed
- ✅ Explicit next task with all dependencies listed
- ✅ Mandatory workflows (tests-first, spec-first)
- ✅ Canonical constraints (what MUST and MUST NOT be done)
- ✅ Architectural context for understanding design decisions

## When to Use

1. **Starting a new feature** - Check `CONTEXT_LOG.md` for recent handoff entries, or check individual handoff docs below
2. **Completing a feature** - Add entry to `CONTEXT_LOG.md` for the next feature in the chain
3. **Understanding dependencies** - Handoff docs show what's ready and what's blocked

## Available Handoff Documents

### CLI Command Handoffs

#### `CLI_DEPLOY-to-CLI_RELEASES.md`
- **From**: CLI_DEPLOY (Deploy command)
- **To**: CLI_RELEASES (Releases list/show commands)
- **Use when**: Starting work on the releases command after deploy is complete

#### `CLI_RELEASES-to-CLI_ROLLBACK.md`
- **From**: CLI_RELEASES (Releases list/show commands)
- **To**: CLI_ROLLBACK (Rollback command)
- **Use when**: Starting work on rollback after releases is complete

#### `CLI_ROLLBACK-to-CLI_BUILD.md`
- **From**: CLI_ROLLBACK (Rollback command)
- **To**: CLI_BUILD (Build command)
- **Use when**: Starting work on build after rollback is complete

#### `CLI_ROLLBACK-to-CLI_PHASE_EXECUTION_COMMON.md`
- **From**: CLI_ROLLBACK (Rollback command)
- **To**: CLI_PHASE_EXECUTION_COMMON (Shared phase execution)
- **Use when**: Starting work on phase execution common after rollback is complete

#### `CLI_ROLLBACK-CLI_RELEASES-CORE_STATE_CONSISTENCY-to-CLI_DEPLOY.md`
- **From**: CLI_ROLLBACK, CLI_RELEASES, CORE_STATE_CONSISTENCY (Multiple features)
- **To**: CLI_DEPLOY (Deploy command)
- **Use when**: Starting work on deploy after rollback, releases, and state consistency are complete

### Core Engine Handoffs

#### `CORE_STATE-to-CLI_DEPLOY.md`
- **From**: CORE_STATE (State management)
- **To**: CLI_DEPLOY (Deploy command)
- **Use when**: Starting work on deploy after state management is complete

#### `CORE_STATE_TEST_ISOLATION-to-CORE_STATE_CONSISTENCY.md`
- **From**: CORE_STATE_TEST_ISOLATION (State test isolation)
- **To**: CORE_STATE_CONSISTENCY (State durability guarantees)
- **Use when**: Starting work on state consistency after test isolation is complete

#### `CLI_PHASE_EXECUTION_COMMON-to-CORE_STATE_TEST_ISOLATION.md`
- **From**: CLI_PHASE_EXECUTION_COMMON (Shared phase execution)
- **To**: CORE_STATE_TEST_ISOLATION (State test isolation)
- **Use when**: Starting work on state test isolation after phase execution common is complete

### Governance Handoffs

#### `GOV_CORE-to-FRONTMATTER.md`
- **From**: GOV_CORE (Governance Core for v1)
- **To**: FRONTMATTER (Spec frontmatter implementation)
- **Use when**: Starting work on spec frontmatter after governance core is complete

#### `COMMIT_DISCIPLINE_PHASE3.md`
- **From**: Phase 1 & Phase 2 (Commit message discipline enforcement)
- **To**: Phase 3 (Commit intelligence & historical analysis)
- **Use when**: Starting work on Phase 3 commit intelligence features after Phase 1 and Phase 2 are complete

#### `COMMIT_REPORT_TYPES_PHASE3.md`
- **From**: Phase 1 & Phase 2 (Commit message discipline enforcement)
- **To**: Phase 3.A (Commit report Go types & golden tests)
- **Use when**: Starting work on Phase 3.A commit report data model implementation (types + golden roundtrip tests)

#### `COMMIT_DISCIPLINE_PHASE3B.md`
- **From**: Phase 3.A (Commit report Go types & golden tests)
- **To**: Phase 3.B (Commit health generators & CLI integration)
- **Use when**: Starting work on Phase 3.B generators and CLI integration after Phase 3.A types are complete

#### `COMMIT_DISCIPLINE_PHASE3C.md`
- **From**: Phase 3.B (Commit health generators & CLI integration)
- **To**: Phase 3.C (CLI wiring for commit discipline reports)
- **Use when**: Starting work on Phase 3.C CLI commands after Phase 3.B generators are complete

## Template

For creating new handoff documents, use:
- `TEMPLATE.md` - Generic template structure
- `README.md` - Detailed instructions on creating handoff docs

## Workflow

1. **Check `CONTEXT_LOG.md`** first for recent handoff entries
2. **Check this index** for historical handoff docs if needed
3. **Open the relevant handoff doc** if one exists for your feature
4. **Follow the handoff doc's Quick Start** section
5. **Add entry to `CONTEXT_LOG.md`** when completing your feature (instead of creating new handoff files)

## Principles

- **Spec-first**: All handoff docs reference spec locations explicitly
- **Test-first**: Mandatory workflow sections enforce tests before code
- **Feature-bounded**: Clear scope reminders and constraints
- **Deterministic**: Complete context, no ambiguity

For more details on the handoff document structure and creation process, see [README.md](./README.md).


<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

# MIGRATION_CONFIG - Analysis

## Purpose

Define the Stagecraft configuration schema for migrations in `stagecraft.yml` so that:
- Migration engines can be selected and configured deterministically.
- Migration sources can be declared in a stable, portable way (no absolute paths).
- CLI migration commands can operate without engine-specific ad-hoc flags.
- Deploy flows can reference migrations consistently (pre-deploy and post-deploy hooks).

This document describes intent, constraints, and required behavior. The spec encodes the precise schema and validation rules.

---

## Scope

This feature defines:
- `stagecraft.yml` schema for migrations (engines, sources, defaults, selection).
- Environment overrides for migration configuration.
- Deterministic normalization rules (sorting, stable identifiers).
- Validation rules that prevent ambiguous or unsafe configurations.

This feature does not implement:
- A new migration engine (raw exists).
- Deploy pre/post hooks (separate features).
- CLI migrate plan/run dedicated commands (separate features).

---

## Current State

- `CORE_MIGRATION_REGISTRY` exists and can register engines.
- `MIGRATION_ENGINE_RAW` exists as an engine implementation.
- `MIGRATION_INTERFACE` exists as a stable engine contract.
- `MIGRATION_CONFIG` is missing; config has no canonical place to declare engines, sources, and defaults.

---

## Design Constraints

### Determinism
Config handling must be deterministic:
- Sorting rules for lists (tags, selected IDs, sources).
- Normalization must not depend on map iteration order.
- No timestamps or derived random values.

### Portability
Config must be portable across machines:
- No absolute paths.
- Prefer paths relative to project root.
- Engine identity uses stable engine names (registry keys).

### Safety
Config should prevent common footguns:
- Reject contradictory selection settings (example: `all=true` plus `ids` or `tags`).
- Reject unknown engine names when validation is possible.
- Enforce stable ordering in lists that become outputs.

---

## High-Level Model

Migrations config has three primary layers:

1. **Global defaults**
   - Whether migrations are enabled.
   - Default engine name.
   - Default sources and selection.

2. **Engine configurations**
   - Per engine settings, optional, engine-specific.
   - Kept generic at this phase to avoid coupling schema to raw-only.
   - Treated as opaque JSON/YAML values at first, validated lightly (type only).

3. **Environment overrides**
   - Allow `dev` to use a different engine, different sources, or different selection.
   - Designed to compose with the global defaults in a predictable way:
     - Override replaces the field, not merges, unless explicitly stated.

---

## Sources and Selection

Stagecraft needs a stable way to describe migration sources.

Recommended minimal source kinds for v1:
- `raw_sql_dir` (directory of `.sql` files)
- `raw_sql_files` (explicit list of `.sql` files)
- Future: tool engines (prisma, drizzle, etc.) can define their own sources.

Selection semantics should map cleanly onto `pkg/migrations.Selection`:
- All migrations
- By IDs
- By tags

The config layer should be able to define:
- Default selection
- Selection used for pre-deploy and post-deploy later

Do not implement hook wiring here, but reserve schema for it.

---

## Environment Overrides

Environment overrides must be explicit and non-magical:
- `migrations.env.dev` may override engine name, sources, selection.
- If override fields are omitted, they inherit the global defaults.
- If override fields are present, they replace the global value.

This keeps behavior easy to reason about and test.

---

## Validation Rules

Minimum required validation:
- Engine names must be non-empty if migrations are enabled.
- `selection` must be well-formed:
  - If `all=true`, then `ids` and `tags` must be empty.
  - `ids` must be unique; tags must be unique.
  - Tags must be sorted and lowercased (or preserve case but enforce sorted order). For v1, enforce sorted, preserve case.
- Sources must be well-formed:
  - Paths must be relative and must not escape the repo root (`..` segments should be rejected).
  - No duplicate sources of the same type.
- Engine config blobs must be objects (map-like), not scalars.

Optional validation (if registry is available at config validation time):
- Unknown engine name should be rejected or warned.

---

## Example User Story

A user has:
- Raw SQL migrations in `migrations/sql/`
- Wants `dev` to run everything
- Wants `prod` to run only tag `schema` pre-deploy, and `seed` post-deploy later

Config should express that without CLI flags.

---

## Acceptance Criteria

This feature is complete when:
- The `stagecraft.yml` schema for migrations is specified and documented.
- Deterministic normalization and validation rules are explicit.
- `pkg/config` can load and validate migrations config with test coverage.
- `spec/features.yaml` for `MIGRATION_CONFIG` can be marked done after implementation and tests.


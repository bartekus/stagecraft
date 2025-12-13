---
feature: MIGRATION_CONFIG
version: v1
status: done
domain: migrations
inputs:
  flags: []
outputs:
  exit_codes: {}
---

# MIGRATION_CONFIG - Specification

## Goal

Define the canonical migrations configuration schema in `stagecraft.yml` that selects and configures migration engines and sources deterministically across environments.

## Non-Goals

- Implementing pre-deploy or post-deploy execution (separate features).
- Implementing dedicated CLI migrate plan/run commands (separate features).
- Defining engine-specific config schemas beyond basic structural validation (future).

---

## Normative Language

The key words **MUST**, **MUST NOT**, **SHOULD**, and **MAY** are to be interpreted as described in RFC 2119.

---

## Location in Config

Migrations configuration lives under:

- `migrations:`

Environment overrides live under:

- `migrations.env.<envName>:`

---

## Determinism Rules

1. Any list field in migrations config **MUST** be normalized to a deterministic order:
   - `tags` sorted lexicographically
   - `ids` sorted lexicographically
   - `raw_sql_files` sorted lexicographically
2. Duplicate values in list fields **MUST** be rejected.
3. Paths **MUST** be normalized:
   - relative to project root
   - slash-separated in config examples
4. Config parsing and validation **MUST NOT** rely on YAML map iteration order.

---

## Schema

### Root

```yaml
migrations:
  enabled: true
  default_engine: "raw"
  sources:
    raw_sql_dir: "migrations/sql"
    raw_sql_files: []
  selection:
    all: true
    ids: []
    tags: []
  engine_config:
    raw: {}
  env:
    dev: {}
    prod: {}
```

Fields:
- `enabled` (bool, optional; default true)
- `default_engine` (string, required if enabled)
- `sources` (object, optional)
- `selection` (object, optional)
- `engine_config` (object/map, optional)
- `env` (object/map, optional)

---

### Sources

`sources.raw_sql_dir`
- Type: string
- Semantics: directory containing `.sql` files
- Rules:
  - MUST be a relative path
  - MUST NOT contain `..` segments
  - MUST NOT be empty string

`sources.raw_sql_files`
- Type: list of string
- Semantics: explicit list of `.sql` files
- Rules:
  - each entry MUST be a relative path
  - each entry MUST NOT contain `..` segments
  - list MUST be sorted lexicographically after normalization
  - list entries MUST be unique

If both `raw_sql_dir` and `raw_sql_files` are set, both are allowed. Engines decide precedence. For v1, Stagecraft SHOULD treat them as additive sources.

---

### Selection

Selection is a configuration-level representation of `pkg/migrations.Selection`.

```yaml
selection:
  all: true
  ids: []
  tags: []
```

Rules:
1. If `all: true`, then `ids` and `tags` MUST be empty.
2. `ids` entries MUST be unique and sorted.
3. `tags` entries MUST be unique and sorted.

---

### Engine Configuration

`engine_config`
- Type: map from engine name to object
- Example:

```yaml
engine_config:
  raw:
    dialect: "postgres"
    schema_table: "stagecraft_migrations"
```

Rules:
- Keys are engine names (must match registry engine name).
- Values MUST be objects (map-like).
- Values are treated as opaque configuration blobs at this feature stage.
- Only structural validation occurs here.
- Engine-specific validation occurs in the engine implementation or later schema enhancements.

---

### Environment Overrides

`migrations.env.<envName>`

Example:

```yaml
migrations:
  default_engine: "raw"
  sources:
    raw_sql_dir: "migrations/sql"
  selection:
    all: true
  env:
    dev:
      selection:
        all: true
    prod:
      selection:
        all: false
        tags: ["schema"]
```

Rules:
- Overrides are applied per environment name.
- Any override field, when present, replaces the global field of the same name.
- Omitted fields inherit from the global configuration.

Allowed override keys:
- `enabled`
- `default_engine`
- `sources`
- `selection`
- `engine_config`

---

## Validation Requirements

When migrations are enabled:
1. `default_engine` MUST be non-empty.
2. `selection` must obey the selection rules above.
3. All paths in `sources` MUST be relative and must not contain `..`.
4. `engine_config` values MUST be objects.

Optional validation (recommended when feasible):
- Unknown `default_engine` SHOULD be rejected if registry is available at validation time.
- Unknown keys under `sources` MUST be rejected.

---

## Normalization Requirements

During config load/validation, Stagecraft MUST normalize:
- sort `selection.ids`, `selection.tags`
- sort `sources.raw_sql_files`
- trim whitespace in string fields
- convert empty lists to `[]` (not null)

Normalization should be performed before any deterministic output generation.

---

## Examples

### Minimal config

```yaml
migrations:
  default_engine: "raw"
  sources:
    raw_sql_dir: "migrations/sql"
  selection:
    all: true
```

### Prod uses tag selection

```yaml
migrations:
  default_engine: "raw"
  sources:
    raw_sql_dir: "migrations/sql"
  selection:
    all: true
  env:
    prod:
      selection:
        all: false
        tags: ["schema"]
```

---

## Acceptance Criteria

This feature is complete when:
1. This spec exists and is approved.
2. `pkg/config` can load and validate migrations config with tests.
3. Normalization and validation are deterministic and covered by tests.
4. `spec/features.yaml` marks `MIGRATION_CONFIG` as done and points to the relevant tests.


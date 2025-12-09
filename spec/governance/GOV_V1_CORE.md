---
feature: GOV_V1_CORE
version: v1
status: done
domain: governance
inputs:
  flags: []   # No direct CLI flags; this is a governance / tooling feature
outputs:
  exit_codes:
    success: 0
    validation_failed: 1
    internal_error: 2
data_structures:
  - name: SpecFrontmatter
    type: object
    fields:
      - name: feature
        type: string
      - name: version
        type: string
      - name: status
        type: string
      - name: domain
        type: string
      - name: inputs
        type: object
      - name: outputs
        type: object
  - name: CliFlag
    type: object
    fields:
      - name: name
        type: string
      - name: type
        type: string
      - name: default
        type: string
      - name: description
        type: string
  - name: FeatureNode
    type: object
    fields:
      - name: id
        type: string
      - name: domain
        type: string
      - name: status
        type: string
      - name: depends_on
        type: array
        items: string
json_schema:
  # Reserved for future: JSON schema describing SpecFrontmatter and CliFlag
  # For v1, only structural shape is enforced by Go code, not by a public JSON schema.
---

# GOV_V1_CORE — Governance Core for v1

## 1. Summary

Stagecraft v1 must ship with a minimal but powerful governance core that ensures:

- Specs are machine-readable and validated

- CLI flags and exit codes match their specs

- Feature dependencies are explicit and acyclic

- A minimal, always-current overview of the feature set exists

This feature bundles the thin-slice implementations of:

1. Machine-verifiable spec schema (frontmatter + validator)

2. Structural diff (spec vs implementation) for flags and exit codes

3. Feature dependency graph with impact analysis

4. Minimal feature overview page (static report)

## 2. Goals

- Enforce structured YAML frontmatter for all `spec/<domain>/<feature>.md` files.

- Provide a single command to validate all specs and fail CI on violations.

- Guarantee alignment between:

  - CLI flags defined in specs

  - CLI flags registered in Cobra commands

- Guarantee alignment between:

  - Exit codes documented in specs

  - Exit codes defined in shared Go constants for core commands

- Maintain a feature dependency DAG derived from `spec/features.yaml`.

- Provide a minimal machine-generated feature overview document.

## 3. Non-Goals

- No full JSON-schema validation of CLI outputs.

- No behavioral diffing or migration-guide generation.

- No fully interactive feature portal or UI dashboard.

- No coverage or velocity metrics in this feature.

## 4. Design

### 4.1 Spec Schema (Frontmatter)

All spec files `spec/<domain>/<feature>.md` MUST start with YAML frontmatter:

- Required keys:

  - `feature` (string; matches Feature ID)

  - `version` (string; e.g. `v1`)

  - `status` (enum: `todo|wip|done`)

  - `domain` (string; e.g. `commands`, `core`, `governance`)

- Optional keys:

  - `inputs.flags[]` — for CLI features; array of `CliFlag`

  - `outputs.exit_codes` — map of symbolic name → integer code

The validator MUST:

- Ensure required keys are present.

- Ensure `feature` matches the filename (e.g. `GOV_V1_CORE.md`).

- Ensure `status` is one of the allowed values.

- Ensure all `inputs.flags[].name` are non-empty strings when present.

- Ensure all `outputs.exit_codes` values are integers when present.

### 4.2 Structural Diff: Spec vs Implementation (Thin Slice)

#### Flags

- Implementation side:

  - A dev-only tool (`cli-introspect`) introspects the root Cobra command and outputs JSON:

    - For each command: list of flags with `name`, `type`, `default`, `usage`.

- Spec side:

  - For CLI-related specs, `inputs.flags[]` defines the canonical list of flags.

The diff tool MUST:

- Fail if a spec declares a flag that does not exist in the Cobra introspection for that command.

- Fail if Cobra declares a flag that is not present in the corresponding spec.

#### Exit Codes

- Implementation side:

  - Core commands (e.g. `deploy`, `rollback`, `build`) MUST use shared exit-code constants defined in a single Go package.

- Spec side:

  - `outputs.exit_codes` for those features MUST list the same symbolic names and integer values.

The diff tool MUST:

- Fail if a spec exit code name or value does not match the shared constant.

- Fail if a shared exit code constant for a core command is not documented in the spec.

### 4.3 Feature Dependency Graph

- Source of truth:

  - `spec/features.yaml` defines all features and a `depends_on` list.

- Optional enrichment:

  - `// Feature: <FEATURE_ID>` comments in Go files associate files with features.

The graph tool MUST:

- Construct a directed graph of features and their dependencies.

- Detect and fail on dependency cycles.

- Provide an "impact" view: given a feature, list all features that directly or transitively depend on it.

### 4.4 Minimal Feature Overview Page

A dev-only generator MUST produce a static overview at:

- `docs/features/OVERVIEW.md`

This document MUST include at least:

- A table of all features with: `ID`, `Domain`, `Status`, `Short description`.

- An optional section showing the dependency graph in a textual form (e.g. adjacency lists or a DOT snippet).

The overview MUST be regenerated by CI on changes to:

- `spec/features.yaml`

- Any `spec/<domain>/<feature>.md` file

## 5. Validation

This feature is considered **done** when:

1. All specs have valid YAML frontmatter and pass the spec validator.

2. The structural diff tool passes for all CLI features:

   - No missing or extra flags.

   - Exit codes aligned for core commands.

3. The feature dependency graph:

   - Builds successfully and has no cycles.

   - Provides a working "impact" command in dev tooling.

4. The feature overview page is generated by:

   - A single `go run` or `make` command.

   - CI fails if the committed overview is stale.

## 6. Rollout

- Phase 1: Implement tools and validators; run them locally only.

- Phase 2: Add to `scripts/run-all-checks.sh` and CI, but mark as "soft fail" (warning mode).

- Phase 3: Flip to "hard fail": PRs cannot merge if any governance checks fail.

- Phase 4: Feature Mapping Invariant & Cross-Validation (see below).

## 7. Phase 4 — Feature Mapping Invariant & Cross-Validation

**Goal:** Enforce the Feature Mapping Invariant across specs, `spec/features.yaml`, implementation code, and tests.

### 7.1 Invariants

1. Every Feature ID in `spec/features.yaml` has exactly one canonical spec file.

2. Every spec file is bound to exactly one Feature ID.

3. Implementation files declare:

   ```go
   // Feature: <FEATURE_ID>
   // Spec: spec/<domain>/<feature>.md
   ```

4. Test files for `wip` and `done` features declare the same `Feature:` header.

5. `todo` features may exist without specs or code, but:

   - Any `Spec:` reference for a `todo` feature MUST point at an existing spec.

6. `wip` and `done` features:

   - MUST have a spec file.

   - MUST have at least one implementation or test file referencing the Feature ID.

7. `done` features:

   - MUST have both implementation and test coverage.

### 7.2 Tooling

Phase 4 is enforced by:

- `internal/governance/mapping` — feature mapping analysis and validation.

- `stagecraft gov feature-mapping` — CLI command that validates the Feature Mapping Invariant.

- `scripts/run-all-checks.sh` — invokes `stagecraft gov feature-mapping` as part of governance validation.

#### CLI Command: `stagecraft gov feature-mapping`

The command validates the Feature Mapping Invariant and reports violations.

**Usage:**
```bash
stagecraft gov feature-mapping [--format=text|json]
```

**Flags:**
- `--format` (default: `text`): Output format. Use `json` for machine-readable output.

**Exit Codes:**
- `0`: No violations detected, invariant holds.
- `1`: One or more violations detected.
- `2`: Internal error (I/O issue, parse error preventing analysis).

**Output:**
- Text format: Human-readable summary with features and violations grouped by type.
- JSON format: Deterministic JSON report with sorted features and violations.

Violations for `wip` and `done` features are treated as hard errors in CI.

Violations for `todo` features are emitted as warnings only.

#### CLI Exit Codes (Feature Mapping)

The `stagecraft gov feature-mapping` command MUST use explicit exit codes so CI and local scripts can distinguish between validation failures and internal errors.

Exit codes:

- `0` — Success. The Feature Mapping Invariant holds and no violations were found.

- `1` — Validation failed. The command completed successfully, but one or more mapping violations were detected (for example: missing spec, missing implementation, missing tests, mismatched headers).

- `2` — Internal error. The command could not complete due to an internal failure (for example: I/O error, parsing error, unexpected panic, or rendering failure).

**Rules:**

- CI MUST treat exit code `1` as a governance failure and mark the workflow as failed.

- CI MUST treat exit code `2` as a tooling failure and also fail the workflow, with additional follow-up required to fix the tool itself.

- Local scripts (such as `scripts/run-all-checks.sh`) MUST use these exit codes without reinterpretation. Validation decisions belong inside the governance tool, not the shell script.


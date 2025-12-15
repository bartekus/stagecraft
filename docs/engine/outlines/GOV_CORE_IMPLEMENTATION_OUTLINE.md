# GOV_CORE — Phase 4 Implementation Outline

Multi-Feature Cross-Validation & Feature Mapping Invariant

---

## 1. Summary & v1 Scope

**Feature ID:** `GOV_CORE`

**Phase:** 4 – Feature Mapping Invariant & Multi-Feature Cross-Validation

**Status:** In progress — scaffold complete, enforcement pending

### 1.1 Problem

As Stagecraft grows, it becomes harder to guarantee that:

- Every **Feature ID** has a single authoritative spec.

- Every spec maps to a single Feature ID.

- Implementation and tests consistently declare the correct `Feature:` and `Spec:` headers.

- `spec/features.yaml` actually reflects the real state of the codebase.

Without automated enforcement, feature–spec–test drift is inevitable, eroding traceability and governance guarantees.

### 1.2 v1 Goal

Provide a **deterministic, CI-enforced tool** that validates the **Feature Mapping Invariant**:

> For every Feature ID, there must be exactly one spec, and all implementation and tests must correctly reference that feature and spec.

v1 focuses on **reporting and enforcement**, not automatic fixes.

---

## 2. In / Out of Scope

### 2.1 In Scope (v1)

- Go-based feature mapping tool:

  - `internal/tools/features` (scanner, validator, index)

  - `cmd/feature-map-check` (CLI wrapper)

- Parsing and validating:

  - `spec/features.yaml`

  - `spec/**.md` spec files

  - Go sources under `internal/` and `pkg/`

  - Go tests (`*_test.go`)

- Status-aware validation:

  - `todo`, `wip`, `done`, `deprecated`, `removed`

- Deterministic, CI-friendly output:

  - Stable ordering of issues

  - Structured severity (`ERROR`, `WARNING`)

### 2.2 Out of Scope (v1)

- Changing CLI behaviour (commands, flags, exit codes)

- Modifying existing specs beyond small governance sections

- Provider logic, registry semantics, or config schema

- Migration behaviour, deployment behaviour

- Auto-fixing feature mappings

---

## 3. CLI & Tooling Contract

### 3.1 CLI Tool: `feature-map-check`

**Command:**

```bash
go run ./cmd/feature-map-check --root . --features spec/features.yaml
```

**Flags:**

- `--root` (string, default `.`)

  Root directory to scan.

- `--features` (string, default `spec/features.yaml`)

  Path to feature index.

**Exit Codes:**

- `0` – all checks passed; no errors (warnings allowed)

- `1` – validation error(s); Feature Mapping Invariant violated

- `>1` – unexpected internal failure (I/O, parsing, panics)

### 3.2 CI Integration

- Integrated into `scripts/run-all-checks.sh` as governance step:

  - After spec validation

  - Before or after docs checks (non-behavioural)

- Required for PRs touching:

  - `spec/features.yaml`

  - `spec/**`

  - `internal/**`

  - `pkg/**`

---

## 4. Data Structures

NOTE: Type names may already exist in `internal/tools/features`. This outline is descriptive; code must match real types.

### 4.1 Feature Index (from spec/features.yaml)

```go
type FeatureStatus string

const (
    FeatureStatusTodo       FeatureStatus = "todo"
    FeatureStatusWIP        FeatureStatus = "wip"
    FeatureStatusDone       FeatureStatus = "done"
    FeatureStatusDeprecated FeatureStatus = "deprecated"
    FeatureStatusRemoved    FeatureStatus = "removed"
)

type FeatureSpec struct {
    ID     string
    Status FeatureStatus
    Spec   string // canonical path like "spec/commands/deploy.md"
}
```

### 4.2 Source Mapping Model

```go
type FileReference struct {
    File      string
    Line      int
    FeatureID string
    SpecPath  string
    IsTest    bool
}

type FeatureIndex struct {
    Features map[string]*FeatureSpec   // Feature ID -> spec
    Impls    map[string][]FileReference // Feature ID -> impl references
    Tests    map[string][]FileReference // Feature ID -> test references
}
```

### 4.3 Validation Issues

```go
type ValidationSeverity string

const (
    SeverityWarning ValidationSeverity = "WARNING"
    SeverityError   ValidationSeverity = "ERROR"
)

type ValidationIssue struct {
    Severity  ValidationSeverity
    FeatureID string
    File      string
    Line      int
    Message   string
}
```

---

## 5. Algorithms

### 5.1 Load spec/features.yaml

- Parse with `yaml.v3`.

- Build `FeatureIndex`:

  - Map ID → `FeatureSpec`.

- Validate:

  - No duplicate IDs.

  - Spec paths non-empty for `wip`, `done`, `deprecated`.

### 5.2 Scan Source Tree

**Inputs:** `rootDir`, `FeatureIndex`.

1. Walk filesystem under `rootDir`:

   - Include: `internal/`, `pkg/`

   - Exclude: `.git`, `.stagecraft`, `testdata`, `vendor`

   - Include `*_test.go` files.

2. For each Go file:

   - Parse header comments only.

   - Extract:

     - `Feature: <FEATURE_ID>`

     - `Spec: spec/<domain>/<feature>.md`

   - Add entry in `SourceIndex`.

3. Build reverse mappings:

   - Feature ID → headers

   - Spec path → feature IDs

### 5.3 Validation Rules (by Feature Status)

Let F = `FeatureSpec`, H = its headers.

**todo**

- No spec required.

- No code/tests required.

- Violations:

  - Using `Spec:` header with nonexistent spec:

    - `WARNING` (non-blocking)

  - Header referencing unknown Feature ID:

    - `WARNING` (point back at `features.yaml`).

**wip**

- Spec MUST exist.

- At least one implementation OR test header must exist.

- Violations:

  - Missing spec path → `ERROR`.

  - Spec file does not exist → `ERROR`.

  - No headers referencing feature → `ERROR`.

  - Any header with mismatched `Spec:` path → `ERROR`.

**done**

- Spec MUST exist.

- At least one implementation and one test reference.

- All headers must:

  - Use same Feature ID.

  - Use same Spec path.

- Violations:

  - Missing spec path → `ERROR`.

  - Spec file does not exist → `ERROR`.

  - No implementation headers → `ERROR`.

  - No test headers → `ERROR`.

  - Mismatched spec paths → `ERROR`.

  - Spec referenced by multiple Feature IDs → `ERROR`.

**deprecated / removed**

- TODO for v1: treat like `done`, but allow docs-only features

  (for now; may evolve in future phases).

---

## 6. Determinism & Output

- Sort issues by:

  1. FeatureID (lexicographically)

  2. Severity (`ERROR` before `WARNING`)

  3. File path

  4. Line number

- Output format:

```
ERROR [CLI_DEPLOY] internal/cli/commands/deploy.go:42: missing Spec header
WARNING [CLI_DEV] internal/cli/commands/dev.go:37: todo feature has spec but no implementation
```

---

## 7. Test Plan

### 7.1 Unit Tests (`internal/tools/features`)

- `LoadFeaturesYAML`:

  - Missing file

  - Duplicate IDs

  - Invalid statuses

- `ScanSourceTree`:

  - Extracts headers from implementation/test files

  - Excludes `testdata/` and `vendor`

- `ValidateFeatureIndex`:

  - `todo`: no spec + no headers = OK

  - `wip`: missing spec → `ERROR`

  - `done`: missing tests → `ERROR`

  - Orphan spec / dangling Feature ID scenarios

  - Status-specific behaviour

### 7.2 Golden Fixture

- Fixture under `internal/tools/features/testdata/feature-map-fixture/`:

  - Minimal `spec/features.yaml`

  - A few spec files

  - A few Go files with headers

- Golden issue output for regression tests

---

## 8. Completion Criteria

Phase 4 v1 is done when:

- `cmd/feature-map-check` is implemented and tested.

- `internal/tools/features` implements:

  - Loading

  - Scanning

  - Validation

  - Deterministic rendering

- Tool integrated into `scripts/run-all-checks.sh`.

- CI fails on Feature Mapping Invariant violations for `wip`/`done`.

- `todo` features only emit warnings.

- Golden tests ensure stable output.

---

## 9. Future Work (Phase 4.x)

- Dashboard view of feature health (counts, coverage, status matrix).

- Integration with commit report / suggestions systems.

- ADR for multi-feature dependencies and cross-feature impact analysis.


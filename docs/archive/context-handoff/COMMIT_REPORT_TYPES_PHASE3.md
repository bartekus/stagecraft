> **Superseded by** `docs/context-handoff/CONTEXT_LOG.md` section 4.3. Kept for historical reference. New context handoffs MUST be added to the context log.

# 🔥 Agent Promo: Phase 3.A – Commit Report Go Types & Golden JSON Roundtrip

**Task:** Implement the Go types and initial golden tests for the Phase 3 commit discipline reports, based on the design in `docs/design/commit-reports-go-types.md`.

**Feature IDs:**

- `GOV_CORE` (governance & commit discipline)

- `CLI_VALIDATE_COMMIT` (telemetry consumer / integration point, future)

**Prerequisites:**

- Phase 1 and Phase 2 commit discipline artifacts exist (docs, templates, hooks)

- JSON schemas and Go type design are stable:

  - `.stagecraft/reports/commit-health.json` schema

  - `.stagecraft/reports/feature-traceability.json` schema

  - `docs/design/commit-reports-go-types.md`

⸻

## 🎯 Mission

Implement the **data model** for commit discipline reporting – no scanning or generators yet.

**Goal:** Provide strongly-typed Go representations of the commit-health and feature-traceability reports that:

- Mirror the JSON schemas exactly

- Are fully deterministic in their JSON encoding

- Are validated via golden roundtrip tests

This is a **pure modeling + testing** task – no git history analysis, no CLI wiring.

⸻

## 📦 Scope (Phase 3.A Implementation)

The agent MUST:

### 1. Implement `commithealth` report types

Create:

```text
internal/reports/commithealth/types.go
internal/reports/commithealth/types_test.go
```

Implementation rules:

- Follow `docs/design/commit-reports-go-types.md` exactly for:

  - `Report`

  - `RepoInfo`

  - `CommitRange`

  - `Summary`

  - `Rule`

  - `Commit`

  - `Violation`

  - `ViolationCode` enum

  - `Severity` enum

- Struct fields and JSON tags MUST match the commit-health.json schema 1:1.

- Prefer enum types over bare strings:

  - `Rule.Code` MUST use `ViolationCode`.

  - `Rule.Severity` MUST use `Severity`.

  - `Violation.Code` MUST use `ViolationCode`.

  - `Violation.Severity` MUST use `Severity`.

  - `Summary.ViolationsByCode` MAY use `map[ViolationCode]int` for extra type safety (recommended).

### 2. Implement `featuretrace` report types

Create:

```text
internal/reports/featuretrace/types.go
internal/reports/featuretrace/types_test.go
```

Implementation rules:

- Follow `docs/design/commit-reports-go-types.md` exactly for:

  - `Report`

  - `Summary`

  - `Feature`

  - `FeatureStatus` enum

  - `SpecInfo`

  - `ImplementationInfo`

  - `TestsInfo`

  - `CommitsInfo`

  - `Problem`

  - `ProblemCode` enum

  - `Severity` enum

- Struct fields and JSON tags MUST match the feature-traceability.json schema 1:1.

- Prefer enum types over bare strings:

  - `Feature.Status` MUST use `FeatureStatus`.

  - `Problem.Code` MUST use `ProblemCode`.

  - `Problem.Severity` MUST use `Severity`.

### 3. Deterministic JSON roundtrip tests (golden)

For each package:

Create golden testdata:

```text
internal/reports/commithealth/testdata/report_basic.golden.json
internal/reports/featuretrace/testdata/report_basic.golden.json
```

Each `*_test.go` MUST:

1. Construct a minimal but non-trivial `Report` instance:

   - For `commithealth`:

     - 2 commits:

       - 1 valid

       - 1 invalid with at least 2 violation codes

     - `summary.violations_by_code` populated accordingly

     - `rules` includes the codes used

   - For `featuretrace`:

     - 2 Feature IDs:

       - 1 fully traceable (spec+impl+tests+commits, no problems)

       - 1 with missing tests/implementation, with corresponding problems

2. Marshal using `encoding/json` with no custom encoder.

3. Normalize whitespace using `json.Compact`.

4. Compare the result byte-for-byte with the golden file.

5. Tests MUST:

   - Be deterministic (no random data, no timestamps)

   - Not assert on `GeneratedAt` value (either leave it empty or ignore the field)

### 4. Determinism Constraints

The implementation MUST:

- Ensure all slice fields that represent sets are sorted before writing golden files:

  - `Summary.ViolationsByCode` (keys sorted when building the map)

  - `CommitHealth.Report.Commits` (keys sorted when building the map)

  - `FeatureTrace.Report.Features` (keys sorted when building the map)

  - Files and SHAs slices

- Rely on `encoding/json`'s documented behavior of sorted map keys, but:

  - Explicitly sort slices before encoding.

### 5. Documentation Update

Update:

- `docs/design/commit-reports-go-types.md` (if needed) to reflect any enum-usage decisions.

- `docs/guides/AI_COMMIT_WORKFLOW.md` to reference:

  - The existence of commit discipline reports as future telemetry, not yet wired.

No changes to core `Agent.md` are required in this phase.

⸻

## 🧪 Testing Requirements

The agent MUST:

1. Add unit tests:

   - `internal/reports/commithealth/types_test.go`

   - `internal/reports/featuretrace/types_test.go`

2. Tests MUST cover:

   - Construction of a valid `Report` for each package.

   - JSON marshaling that matches the golden files.

   - Roundtrip: `json.Unmarshal` back into Go structs with no errors.

3. All tests MUST obey:

   - No timestamps.

   - No environment-dependent fields.

   - No random data.

4. Tests MUST be integrated with:

   - `go test ./...`

   - `./scripts/run-all-checks.sh`

⸻

## 🧱 AI MUST Follow Standard Development Flow

1. Identify Feature ID:

   - `GOV_CORE` (governance / reporting)

   - Use `CLI_VALIDATE_COMMIT` only as a consumer, not as a primary feature for this task.

2. Ensure correct branch:

   ```bash
   git branch --show-current
   # Expected: feature/GOV_CORE-commit-report-types
   ```

3. Read:

   - `docs/design/commit-reports-go-types.md`

   - JSON schema docs (if present)

   - `Agent.md` commit & determinism rules

4. Implement:

   - `types.go` files for both packages

   - Golden test files

   - Roundtrip tests

5. Run:

   ```bash
   ./scripts/goformat.sh
   ./scripts/run-all-checks.sh
   ```

6. Commit with:

   ```
   feat(GOV_CORE): add commit report Go types
   ```

7. Provide PR title and body:

   - Title:

     ```
     [GOV_CORE] Add commit report Go types and golden roundtrip tests
     ```

   - Body:

     - Feature: `GOV_CORE`

     - Spec / docs: `docs/design/commit-reports-go-types.md`

     - Summary of types and tests

     - Determinism guarantees

     - Non-goals (no generators yet)

⸻

## ⚠️ Non-Goals (Phase 3.A)

- No git history scanning.

- No CLI commands or flags.

- No CI wiring changes.

- No `.stagecraft/reports/*.json` generation yet.

- No integration with `stagecraft validate-commit`.

This phase is strictly about types + golden roundtrip tests to establish a stable foundation for subsequent generator and telemetry work.


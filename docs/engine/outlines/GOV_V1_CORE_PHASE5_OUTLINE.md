# GOV_V1_CORE — Phase 5 Implementation Outline

Repository Stabilization and Governance Golden Tests

---

## 1. Summary and v1 Scope

**Feature ID:** `GOV_V1_CORE` (Phase 5 is part of GOV_V1_CORE, not a separate feature)

**Phase:** 5 - Repository stabilization and governance golden tests  

**Status:** Planned

### 1.1 Problem

Phase 4 delivered the governance tooling:

- Feature Mapping Invariant checks.

- Spec frontmatter integrity checks.

- Feature graph validation.

- CLI vs spec flag alignment.

- Governance wired into `run-all-checks.sh` and CI.

However, the repository may still:

- Contain existing mapping violations.

- Have features whose status does not reflect reality.

- Lack golden tests for governance reports themselves.

This leaves governance in a "tooling enforced but not fully stabilized" state.

### 1.2 v1 Goal

Use the GOV_V1_CORE toolchain to fully stabilize the repository and pin governance outputs in tests.

> After Phase 5, governance should report zero violations on the real repo and governance output formats should be protected by golden tests.

---

## 2. In / Out of Scope

### 2.1 In Scope

- Fixing all existing Feature Mapping violations.

- Aligning `spec/features.yaml` feature statuses with implementation reality.

- Adding a golden test for the Feature Mapping report.

- Minor documentation updates for GOV_V1_CORE and implementation status.

### 2.2 Out of Scope

- Introducing new governance tools or commands.

- Changing non governance behavior (dev, deploy, infra, providers).

- Modifying coverage thresholds or coverage enforcement scripts.

- Adding dashboard or UI layers for governance.

---

## 3. Existing Tools Reused

Phase 5 does not introduce new binaries. It reuses:

- `stagecraft gov feature-mapping`

- `go run ./cmd/spec-validate --check-integrity`

- `go run ./cmd/features-tool graph`

- `go run ./cmd/spec-vs-cli`

- `go run ./cmd/gen-features-overview`

- `./scripts/run-all-checks.sh`

No new CLI flags or commands are required.

---

## 4. Workstream A - Mapping Violations Cleanup

### 4.1 Inputs

- Current mapping violations from:

  ```bash
  ./bin/stagecraft gov feature-mapping
  ```

- Codebase in `internal/` and `pkg/`.

- Specs in `spec/**/*.md`.

- Feature index in `spec/features.yaml`.

### 4.2 Tasks

1. Enumerate all violation codes currently reported by mapping.

2. Group violations by type:

   - `MISSING_SPEC`

   - `MISSING_IMPL`

   - `MISSING_TESTS`

   - `SPEC_PATH_MISMATCH`

   - `FEATURE_NOT_LISTED`

   - `ORPHAN_SPEC`

   - Any others that mapping defines.

3. For each group, apply specific fixes:

   - **Missing spec:**

     - Create missing spec file with correct frontmatter, or

     - Remove or rename the Feature ID in `spec/features.yaml` if dead.

   - **Missing impl / tests:**

     - If feature is genuinely done, add missing headers to the relevant Go files.

     - If not done, downgrade feature status to `wip` or `todo`.

   - **Spec path mismatch:**

     - Fix `Spec:` header to match canonical path from `spec/features.yaml`.

     - Or fix `spec/features.yaml` entry if canonical path is wrong.

   - **Feature not listed:**

     - If the Feature ID is real, add it to `spec/features.yaml` with correct metadata.

     - If it is a typo or obsolete, remove or rename the header.

   - **Orphan spec:**

     - Attach spec to a valid Feature ID in `spec/features.yaml` and add code headers, or

     - Move it to archive or delete it if intentionally unused.

4. Re run:

   ```bash
   ./bin/stagecraft gov feature-mapping
   ```

   until violations are zero for the real repository.

---

## 5. Workstream B - Aligning spec/features.yaml

### 5.1 Behavior Rules

For each feature:

- `done`:

  - Spec file exists and is valid.

  - At least one implementation file with `Feature:` and `Spec:` headers.

  - At least one test file with `Feature:` header.

- `wip`:

  - Spec file exists and is valid.

  - At least one implementation or test header exists.

- `todo`:

  - May have a spec file.

  - Must not already be fully implemented and tested.

### 5.2 Tasks

1. Iterate over all entries in `spec/features.yaml`.

2. For each feature:

   - Cross check against:

     - Headers discovered by mapping.

     - Spec files under `spec/**`.

     - Tests under `internal/` and `pkg/`.

3. Adjust state:

   - If `done` but missing requirements → downgrade to `wip` or `todo`.

   - If `wip` but fully implemented and tested → upgrade to `done`.

   - If `todo` but feature is clearly implemented and tested → upgrade to `wip` or `done`.

4. Run:

   ```bash
   ./bin/stagecraft gov feature-mapping
   ```

   to confirm that changes did not introduce new violations.

---

## 6. Workstream C - Governance Golden Tests

### 6.1 Fixture Layout

Add a synthetic "mini repo" under:

```
internal/governance/mapping/testdata/golden_repo/
  spec/
    features.yaml
    commands/
      dev.md
      deploy.md
  internal/
    cli/
      commands/
        dev.go
        dev_test.go
        deploy.go
        deploy_test.go
  golden/
    feature-mapping-report.json
```

Properties:

- `spec/features.yaml` defines a small set of features (for example `CLI_DEV`, `CLI_DEPLOY`).

- `spec/commands/dev.md` and `spec/commands/deploy.md` represent canonical spec files.

- `dev.go`, `dev_test.go`, `deploy.go`, `deploy_test.go` contain proper `Feature` and `Spec` headers.

- No violations should be present for this fixture.

### 6.2 Test Behavior

Add a new Go test file:

`internal/governance/mapping/mapping_golden_test.go`

The test MUST:

- Use the fixture root `internal/governance/mapping/testdata/golden_repo`.

- Configure mapping options to point at this root and its `spec/features.yaml`.

- Call `mapping.Analyze`.

- Marshal the `Report` to indented JSON.

- Compare against `golden/feature-mapping-report.json`.

Golden updates MUST only happen when behavior or spec intentionally changes, guarded by an explicit flag or environment variable.

---

## 7. Determinism and Output

- The test fixture must be static and fully committed.

- All file walks and report slices must already be sorted by mapping.

- The golden JSON must be formatted with consistent indentation and field ordering.

- The golden test must not rely on timestamps or dynamic data.

---

## 8. Completion Criteria

Phase 5 is complete when:

- `./bin/stagecraft gov feature-mapping` reports zero violations for the real repository.

- The golden test for mapping reports passes.

- `./scripts/run-all-checks.sh` passes locally.

- CI Governance workflow passes.

- `spec/governance/GOV_V1_CORE.md` documents Phase 5 and its status as complete.

- `docs/engine/status/implementation-status.md` marks GOV_V1_CORE as done.

---

## 9. Future Enhancements (Post Phase 5)

- Add golden tests for other governance tools (spec vs CLI diff, feature graph).

- Add a consolidated governance JSON summary for dashboards.

- Integrate governance status into commit reports and suggestions.


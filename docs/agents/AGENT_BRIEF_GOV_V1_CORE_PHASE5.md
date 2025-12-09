# AGENT BRIEF — GOV_V1_CORE — Phase 5

## Repository Stabilization and Governance Golden Tests

**Status:** Planned - ready for implementation  

**Feature ID:** GOV_V1_CORE  

**Spec:** spec/governance/GOV_V1_CORE.md  

**Phase:** 5 - Repository stabilization and governance golden tests

---

## 🎯 Mission

Phase 5 turns GOV_V1_CORE from "enforced tooling" into a fully stabilized, authoritative governance layer.

> Use the existing governance tools (feature mapping, spec validation, graph checks, spec vs CLI) to clean up all remaining violations, align metadata, and add golden tests for the governance reports themselves.

After Phase 5:

- The repository has **zero known governance violations**.

- `spec/features.yaml` is in **exact agreement** with the code, tests, and specs.

- Governance reports have **golden tests**, so their shape and content are stable and refactor safe.

- GOV_V1_CORE can safely be considered **v1 complete**.

---

## 🧱 Scope (What is included)

### Must do

1. **Fix all current Feature Mapping violations**

   - Run `./bin/stagecraft gov feature-mapping`.

   - For each violation group:

     - Add missing `Feature:` and `Spec:` headers.

     - Add missing spec files or remove dead Feature IDs.

     - Fix mismatched spec paths or incorrect headers.

     - Resolve orphan specs and dangling Feature IDs.

2. **Align `spec/features.yaml` with reality**

   - For each feature:

     - `done` - must have spec, implementation headers, and test headers.

     - `wip` - must have spec and at least one implementation or test.

     - `todo` - must not secretly be implemented and tested.

   - Update statuses where they lie.

3. **Add governance golden tests**

   - Add a dedicated golden fixture repo under `internal/governance/mapping/testdata/`.

   - Add a golden JSON file for the Feature Mapping report.

   - Add a Go test that runs the mapping analysis on the fixture and compares the result to the golden JSON.

4. **Mark GOV_V1_CORE as Phase 5 complete**

   - Update `spec/governance/GOV_V1_CORE.md` to document Phase 5 stabilization and golden tests.

   - Update `docs/engine/status/implementation-status.md` to reflect final GOV_V1_CORE status (done) once work is complete.

---

## 🚫 Out of Scope

- Introducing new governance tools beyond GOV_V1_CORE.

- Changing core business behavior (deploy, dev, infra, providers).

- Auto-fixers that rewrite code or specs automatically.

- Coverage enforcement or thresholds (handled in other coverage docs).

Phase 5 is about **using existing governance tools** to clean the repo and make governance outputs stable.

---

## 🧪 Test Requirements

### Governance Golden Tests

- Golden tests MUST live in `internal/governance/mapping/testdata/`.

- They MUST use a fixed, self contained fixture repo layout.

- They MUST:

  - Call `mapping.Analyze` on the fixture.

  - Marshal the `Report` to indented JSON.

  - Compare against a committed golden JSON file.

  - Fail on any difference between actual and golden.

### Behavior

- Re running the golden tests without changing the fixture or governance logic MUST produce the same output every time.

- Golden test updates MUST be done only when:

  - The spec has explicitly changed, or

  - A deliberate governance output change has been approved.

---

## 🧩 Constraints

- Deterministic behavior only.

- Use stdlib only in tests (no new dependencies).

- No changes to `run-all-checks.sh` semantics, except possibly adding the new golden test file to target packages if needed.

- Minimal diffs - no unrelated cleanup.

---

## ✔ Success Criteria

Phase 5 is complete when:

- `./bin/stagecraft gov feature-mapping` reports **zero violations** for the real repo.

- `./scripts/run-all-checks.sh` passes cleanly with no governance warnings.

- Golden tests for mapping reports pass.

- GOV_V1_CORE is marked as done in:

  - `spec/governance/GOV_V1_CORE.md`

  - `docs/engine/status/implementation-status.md`

- Governance output shape is pinned by golden tests and safe to refactor around.

---

## 📎 Execution Checklist

### Pre work

- [ ] Confirm Feature ID: `GOV_V1_CORE`

- [ ] Create branch: `feature/GOV_V1_CORE-phase5-repo-stabilization`

- [ ] Verify hooks: `./scripts/install-hooks.sh`

- [ ] Run baseline: `./scripts/run-all-checks.sh` and capture governance violations

### Workstream A - Mapping violations cleanup

- [ ] Run `./bin/stagecraft gov feature-mapping` and list all violation codes.

- [ ] Fix missing headers (Feature and Spec comments) in implementation and test files.

- [ ] Resolve missing or mismatched spec paths.

- [ ] Remove or fix orphan specs and dangling Feature IDs.

- [ ] Re run `./bin/stagecraft gov feature-mapping` until no violations remain.

### Workstream B - `spec/features.yaml` alignment

- [ ] Walk through done features - confirm spec, impl headers, and test headers.

- [ ] Walk through wip features - confirm spec and at least one impl or test.

- [ ] Walk through todo features - ensure they are not secretly done.

- [ ] Update statuses in `spec/features.yaml` to match reality.

### Workstream C - Golden tests

- [ ] Add fixture repo under `internal/governance/mapping/testdata/`.

- [ ] Add golden JSON report file.

- [ ] Add a Go test that:

  - Runs `mapping.Analyze` on the fixture.

  - Marshals the report to indented JSON.

  - Compares to the golden JSON file.

### Workstream D - Documentation and status

- [ ] Update GOV_V1_CORE spec with a short Phase 5 section (stabilization and golden tests).

- [ ] Update implementation status to reflect GOV_V1_CORE as done when complete.

- [ ] Run `./scripts/run-all-checks.sh` and ensure CI passes.

---

## 🔐 Notes for the Agent

- Do not guess mapping behavior. If something about the mapping tool is unclear, STOP and request clarification.

- Keep changes tightly scoped to GOV_V1_CORE and governance files.

- Use small, focused commits (e.g. "fix mapping violations for CLI commands").

- Do not update golden files unless:

  - The spec has changed, or

  - The governance output format was deliberately updated.

- Always run `./scripts/run-all-checks.sh` before finishing work.


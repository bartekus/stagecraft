# Agent Brief: GOV_CORE Governance Hardening

**Feature ID**: GOV_CORE  
**Spec**: `spec/governance/GOV_CORE.md`  
**Canonical Governance**: `docs/governance/GOVERNANCE_ALMANAC.md`  
**Coverage Ledger**: `docs/coverage/COVERAGE_LEDGER.md`

---

## Purpose

This brief provides execution playbooks for implementing GOV_CORE governance hardening phases. GOV_CORE establishes deterministic, spec-driven governance tooling that enforces repository integrity and traceability.

**This document is an execution guide, not a canonical source of truth.** For authoritative governance rules, see `docs/governance/GOVERNANCE_ALMANAC.md`. For coverage history, see `docs/coverage/COVERAGE_LEDGER.md`.

---

## Invariants (Never Changes)

- **Single Feature ID**: All governance work uses `GOV_CORE` as the Feature ID
- **Spec-first, test-first**: No behavioral changes without spec alignment
- **Deterministic**: All tooling produces identical outputs on repeated runs
- **Minimal diffs**: Changes are surgical and scoped to governance tooling only
- **No protected files**: Never modify LICENSE, README.md, ADRs, or global governance docs

---

## How to Run GOV_CORE Work

### Pre-Work Checklist

- [ ] Confirm Feature ID: `GOV_CORE`
- [ ] Verify hooks: `./scripts/install-hooks.sh`
- [ ] Ensure clean working directory: `git status`
- [ ] Create feature branch: `feature/GOV_CORE-<phase>-<short-desc>`
- [ ] Read relevant phase section below

### During Work

- [ ] Follow spec-first, test-first protocol
- [ ] Keep changes scoped to governance tooling only
- [ ] Run `./scripts/run-all-checks.sh` frequently
- [ ] Use deterministic, sorted outputs
- [ ] Link to canonicals, don't duplicate content

### Post-Work Checklist

- [ ] Run `./scripts/goformat.sh` (if Go changes)
- [ ] Run `./scripts/run-all-checks.sh` (must pass)
- [ ] Commit with format: `feat(GOV_CORE): <summary>` or `fix(GOV_CORE): <summary>`
- [ ] Create PR with spec reference and test coverage summary
- [ ] Update phase section below (append-only) with implementation notes

---

## Phase Timeline

### Phase 3: Spec Reference Checker Hardening

**Status**: ✅ Complete

**Context**: Eliminate false positives in spec reference validation by implementing precise, deterministic validation that only considers legitimate spec references from frontmatter-style comments.

**What was implemented**:
- Created `cmd/spec-reference-check` Go tool
- Implemented comment-only parsing (`// Spec:` style)
- Added path format validation
- Implemented directory exclusion (testdata/, test/e2e/)
- Added file existence checking
- Integrated into `scripts/run-all-checks.sh`

**Decisions locked in**:
- `Spec:` is reserved for `spec/*.md` files only
- Non-spec references use `Docs:` comments
- Validation is deterministic and CI-enforceable

**Known pitfalls**:
- Test fixtures must be excluded from validation
- Path normalization handles both `spec/path.md` and `path.md` formats

**Links**:
- Spec: `spec/governance/GOV_CORE.md`
- Implementation: `cmd/spec-reference-check/main.go`
- Integration: `scripts/run-all-checks.sh` (lines 128-151)
- Coverage: `docs/coverage/COVERAGE_LEDGER.md` section 4.2

---

### Phase 4: Feature Mapping Invariant Enforcement

**Status**: ✅ Scaffold Complete, Enforcement Pending

**Context**: Enforce the Feature Mapping Invariant across specs, features.yaml, implementation code, and tests — deterministically and in CI.

**What was implemented**:
- Created `cmd/feature-map-check` tool
- Implemented core scanner logic
- Added status-aware validation rules (todo/wip/done)
- Implemented deterministic error reporting
- Added unit test scaffolding

**Decisions locked in**:
- Every Feature ID has exactly one spec
- Every spec corresponds to exactly one Feature ID
- Implementation files must contain correct `Feature:` + `Spec:` headers
- Tests must reference correct Feature ID
- Status-based validation rules (todo = warnings, wip/done = hard errors)

**Known pitfalls**:
- Tool must handle filesystem traversal deterministically (sorted)
- Status rules differ: todo allows missing artifacts, wip/done require them

**Links**:
- Spec: `spec/governance/GOV_CORE.md`
- Implementation: `cmd/feature-map-check/main.go`, `internal/tools/features/`
- Integration: `scripts/run-all-checks.sh` (pending)
- Coverage: `docs/coverage/COVERAGE_LEDGER.md` section 4.2

**Follow-up Work**:
- CI integration into `scripts/run-all-checks.sh`
- Repository alignment (add missing headers, fix mismatches)
- Test suite expansion (golden tests, comprehensive unit tests)
- Governance flip to strict mode (hard-fail on violations)

---

### Phase 5: Repository Stabilization and Governance Golden Tests

**Status**: 📋 Planned

**Context**: Turn GOV_CORE from "enforced tooling" into a fully stabilized, authoritative governance layer by cleaning up all remaining violations, aligning metadata, and adding golden tests for governance reports.

**What will be implemented**:
- Fix all Feature Mapping violations
- Align `spec/features.yaml` with reality
- Add governance golden tests
- Mark GOV_CORE as Phase 5 complete

**Decisions to lock in**:
- Golden tests live in `internal/governance/mapping/testdata/`
- Golden JSON format for Feature Mapping reports
- Zero violations policy for done features

**Known pitfalls**:
- Golden test updates only when spec changes or deliberate governance output change
- Must maintain deterministic ordering in all reports

**Links**:
- Spec: `spec/governance/GOV_CORE.md`
- Coverage: `docs/coverage/COVERAGE_LEDGER.md` section 4.2
- Governance: `docs/governance/GOVERNANCE_ALMANAC.md` section 8

---

## Migration Notes

- ✅ Phase 3 brief merged from `docs/agents/AGENT_BRIEF_GOV_CORE_PHASE3.md`
- ✅ Phase 4 brief merged from `docs/agents/AGENT_BRIEF_GOV_CORE_PHASE4.md`
- ✅ Phase 5 brief merged from `docs/agents/AGENT_BRIEF_GOV_CORE_PHASE5.md`
- ✅ Moved from `docs/agents/` to `docs/engine/agents/` (aligned with engine docs structure)

---

**Last Updated**: 2025-01-XX

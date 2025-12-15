> **Superseded by** `docs/engine/agents/AGENT_BRIEF_GOV_CORE.md` section "Phase 3: Spec Reference Checker Hardening". Kept for historical reference. New GOV_CORE execution notes MUST go into the merged agent brief.

# Agent Brief: Spec Reference Checker Hardening - Phase 3

**Feature ID**: GOV_CORE  
**Spec**: `spec/governance/GOV_CORE.md`  
**Context**: Hardening governance tooling to eliminate false positives in spec reference validation. See `docs/coverage/COVERAGE_COMPLIANCE_PLAN_PHASE2.md` Phase 3 section.

**Status**: ✅ Phase 3 implemented and merged. `Spec:` is now reserved for `spec/*.md` files and enforced by `cmd/spec-reference-check`. All non-spec references converted to `Docs:` comments.

---

## Mission

Eliminate false positives in the spec reference checker by implementing precise, deterministic validation that only considers legitimate spec references from frontmatter-style comments, while explicitly excluding test fixtures, debug output, and invalid path formats.

This is a **governance tooling hardening** initiative that improves CI determinism and reduces noise in validation failures.

---

## Scope

Focus strictly on hardening the spec reference checker in `scripts/run-all-checks.sh` (lines 128-151).

### Current Problem

The checker uses a naive regex pattern `[Ss]pec:[[:space:]]+([^[:space:]]+)` that matches:
- ✅ Legitimate frontmatter comments: `// Spec: spec/commands/deploy.md`
- ❌ Test fixtures: `"test/feature1.md"`
- ❌ Debug output: `SpecInfo{...}`
- ❌ Empty strings: `""`
- ❌ Non-string tokens: `true`, `{}`
- ❌ Malformed paths: `spec/commands/deploy.md\n",`

This produces false positives that block valid PRs and introduce non-determinism.

### Target Solution

Implement a hardened checker that:

1. **Only matches legitimate spec references:**
   - Pattern: `// Spec: spec/...` or `// spec: spec/...` (comment-style only)
   - Must be in Go source files (`.go`)
   - Must follow comment syntax (`//` prefix)

2. **Validates path format:**
   - Must start with `spec/`
   - Must end with `.md`
   - Must not contain newlines, tabs, or control characters
   - Must not contain struct-like syntax (`{`, `}`)
   - Must not be empty

3. **Excludes test fixtures:**
   - Ignore paths under `**/testdata/`
   - Ignore paths under `test/e2e/`
   - Ignore paths under `internal/**/testdata/`

4. **Validates file existence:**
   - Check if file exists at the referenced path
   - Support both `spec/path.md` and `path.md` (normalize to `spec/path.md`)

---

## Non-Goals (Out of Scope)

- Changing spec file structure or frontmatter format
- Modifying other validation scripts
- Adding new dependencies or tools
- Changing behavior of other governance checks
- Large refactors of the validation infrastructure

---

## Implementation Strategy

### Approach

Replace the naive regex-based checker with a Go-based validator that:

1. **Scans Go files systematically:**
   - Walk `.go` files only
   - Skip `testdata/` directories
   - Parse comments to extract `Spec:` references

2. **Validates references precisely:**
   - Use structured parsing (not regex)
   - Validate path format before checking existence
   - Filter invalid paths early

3. **Provides clear error messages:**
   - Show file and line number for invalid references
   - Distinguish between "invalid format" and "file not found"

### Implementation Location

Create a new Go tool: `cmd/spec-reference-check` that:
- Takes no arguments (scans current directory)
- Outputs structured errors for invalid references
- Exits with code 1 if any invalid references found
- Integrates into `scripts/run-all-checks.sh`

### Alternative: Shell Script Hardening

If Go tool is too heavy, harden the existing shell script with:
- Better regex pattern (comment-only)
- Path format validation
- Directory exclusion
- Better error messages

**Recommendation**: Go tool is preferred for maintainability and testability.

---

## Success Criteria

Phase 3 is complete when **all** of the following hold:

1. **False positives eliminated:**
   - No false positives from test fixtures
   - No false positives from debug output
   - No false positives from struct stringifications
   - No false positives from invalid path formats

2. **True positives preserved:**
   - Legitimate spec references still validated
   - Missing spec files still detected
   - Invalid path formats still rejected

3. **Determinism:**
   - Checker produces identical results on repeated runs
   - No environment-dependent behavior
   - Clear, actionable error messages

4. **Test coverage:**
   - Tests for valid references
   - Tests for invalid references (false positive scenarios)
   - Tests for missing files
   - Tests for directory exclusions

5. **Integration:**
   - `./scripts/run-all-checks.sh` passes with hardened checker
   - CI validation passes
   - No regressions in other checks

---

## Constraints (From Agent.md & GOV_CORE)

### You MUST:

- Treat **GOV_CORE** as the single Feature ID for this work
- Follow spec-first, test-first where any behavioural ambiguity is encountered
- Keep diffs **minimal and deterministic**
- Keep changes scoped strictly to:
  - Spec reference checker implementation
  - Test files for the checker
  - Integration into `scripts/run-all-checks.sh`
  - Documentation updates

### You MUST NOT:

- Change spec file structure or frontmatter format
- Modify other validation scripts
- Introduce new external dependencies (unless absolutely necessary)
- Change behavior of other governance checks
- Modify protected files (LICENSE, top-level README, ADRs, global governance docs)

---

## Execution Checklist

### Setup

- [ ] Confirm Feature ID: `GOV_CORE`
- [ ] Verify hooks (`./scripts/install-hooks.sh` if needed)
- [ ] Ensure clean working directory
- [ ] On appropriate feature branch (e.g. `test/GOV_CORE-phase3-spec-checker`)

### Implementation

- [ ] Create `cmd/spec-reference-check/main.go` (or harden shell script)
- [ ] Implement comment parsing (only `// Spec:` style)
- [ ] Implement path format validation
- [ ] Implement directory exclusion logic
- [ ] Implement file existence checking
- [ ] Add comprehensive tests
- [ ] Update `scripts/run-all-checks.sh` to use new checker

### Testing

- [ ] Test with legitimate spec references (should pass)
- [ ] Test with test fixtures (should ignore)
- [ ] Test with debug output (should ignore)
- [ ] Test with invalid paths (should reject)
- [ ] Test with missing files (should error)
- [ ] Run `./scripts/run-all-checks.sh` and verify no false positives

### Validation

- [ ] Run `go test ./...`
- [ ] Run `./scripts/run-all-checks.sh`
- [ ] Verify no regressions in other checks
- [ ] Ensure working directory is clean and branch is ready for PR

---

## Reference Documents

- `spec/governance/GOV_CORE.md` – Governance feature specification
- `docs/coverage/COVERAGE_COMPLIANCE_PLAN_PHASE2.md` – Phase 2 plan (Phase 3 section)
- `scripts/run-all-checks.sh` – Current implementation (lines 128-151)
- `Agent.md` – Development protocol and constraints

---

**Last Updated**: 2025-12-08


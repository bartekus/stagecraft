# Spec Reference Checker Hardening - Phase 3 Plan

**Feature**: GOV_CORE  
**Status**: Phase 3 - Governance Tooling Hardening  
**Date**: 2025-12-08

---

## Overview

Phase 3 focuses on hardening the spec reference checker to eliminate false positives and improve CI determinism. This is a **governance tooling improvement** that makes validation more reliable and reduces noise in PR checks.

**Prerequisite**: Phase 2 must be complete before starting Phase 3.

**Current Status**: ⚠️ False Positives in Spec Reference Checker
- Checker produces false positives from test fixtures, debug output, and invalid paths
- Blocks valid PRs for non-spec reasons
- Introduces non-determinism in validation

**Target Status**: ✅ Deterministic, Precise Spec Reference Validation
- Only legitimate spec references validated
- False positives eliminated
- Clear, actionable error messages

---

## Problem Analysis

### Current Implementation

The checker in `scripts/run-all-checks.sh` (lines 128-151) uses:

```bash
grep -r "Spec:" --include="*.go" . | \
  while read line; do
    if [[ $line =~ [Ss]pec:[[:space:]]+([^[:space:]]+) ]]; then
      SPEC_FILE="${BASH_REMATCH[1]}"
      # Check if file exists...
    fi
  done
```

### False Positive Scenarios

1. **Test fixtures:**
   ```
   "test/feature1.md",
   ```

2. **Debug output:**
   ```
   SpecInfo{
   true,
   ```

3. **Empty strings:**
   ```
   "",
   ```

4. **Malformed paths:**
   ```
   spec/commands/deploy.md\n",
   ```

5. **Struct stringifications:**
   ```
   true,
   {}
   ```

### Root Cause

The regex pattern `[Ss]pec:[[:space:]]+([^[:space:]]+)` is too permissive:
- Matches any occurrence of "Spec:" followed by whitespace and non-whitespace
- Doesn't validate comment syntax (`//` prefix)
- Doesn't validate path format
- Doesn't exclude test fixtures
- Doesn't filter invalid characters

---

## Solution Design

### Approach: Go-Based Validator

Create a dedicated Go tool `cmd/spec-reference-check` that:

1. **Scans Go files systematically:**
   - Walk `.go` files recursively
   - Skip `testdata/` directories
   - Parse comments to extract `Spec:` references

2. **Validates references precisely:**
   - Only match comment-style references: `// Spec: spec/...`
   - Validate path format before checking existence
   - Filter invalid paths early

3. **Provides clear error messages:**
   - Show file and line number for invalid references
   - Distinguish between "invalid format" and "file not found"

### Path Validation Rules

A valid spec reference path must:

- Start with `spec/`
- End with `.md`
- Not contain newlines (`\n`), tabs (`\t`), or control characters
- Not contain struct-like syntax (`{`, `}`)
- Not be empty
- Not be in excluded directories:
  - `**/testdata/`
  - `test/e2e/`
  - `internal/**/testdata/`

### Comment Pattern

Only match comments that follow this pattern:

```
// Spec: spec/path/to/file.md
// spec: spec/path/to/file.md
```

Case-insensitive `Spec:` keyword, followed by whitespace, followed by a valid path.

---

## Implementation Plan

### Step 1: Create Go Tool

**File**: `cmd/spec-reference-check/main.go`

**Responsibilities:**
- Walk Go files recursively
- Parse comments for `Spec:` references
- Validate path format
- Check file existence
- Report errors

**Dependencies:**
- Standard library only (`filepath`, `os`, `strings`, `regexp`)

### Step 2: Implement Path Validation

**Function**: `validateSpecPath(path string) bool`

**Rules:**
- Must start with `spec/`
- Must end with `.md`
- Must not contain invalid characters
- Must not be in excluded directories

### Step 3: Implement Comment Parsing

**Function**: `extractSpecReferences(content string, filePath string) []SpecReference`

**Pattern:**
- Match `// Spec:` or `// spec:` (case-insensitive)
- Extract path after colon and whitespace
- Track file and line number

### Step 4: Add Tests

**File**: `cmd/spec-reference-check/main_test.go`

**Test Cases:**
- Valid spec references (should pass)
- Invalid path formats (should reject)
- Test fixtures (should ignore)
- Debug output (should ignore)
- Missing files (should error)
- Directory exclusions (should ignore)

### Step 5: Integrate into run-all-checks.sh

**Change**: Replace lines 128-151 with:

```bash
info "Checking for missing spec files..."
if ! go run ./cmd/spec-reference-check; then
    error "Spec reference validation failed"
    exit 1
fi
info "All spec file references are valid"
```

---

## Test Plan

### Unit Tests

1. **Valid References:**
   ```go
   // Spec: spec/commands/deploy.md
   ```
   Expected: Pass

2. **Invalid Format:**
   ```go
   "Spec: test/feature1.md"
   ```
   Expected: Ignore (not a comment)

3. **Test Fixture:**
   ```go
   testFile := "test/feature1.md"
   ```
   Expected: Ignore (not a comment)

4. **Debug Output:**
   ```go
   fmt.Printf("SpecInfo{%v}", spec)
   ```
   Expected: Ignore (not a comment)

5. **Missing File:**
   ```go
   // Spec: spec/commands/nonexistent.md
   ```
   Expected: Error

6. **Directory Exclusion:**
   ```go
   // Spec: testdata/fixture.md
   ```
   Expected: Ignore (in excluded directory)

### Integration Tests

- Run against entire codebase
- Verify no false positives
- Verify legitimate references still validated
- Verify missing files still detected

---

## Success Criteria

### Functional Requirements

- [ ] No false positives from test fixtures
- [ ] No false positives from debug output
- [ ] No false positives from struct stringifications
- [ ] No false positives from invalid path formats
- [ ] Legitimate spec references still validated
- [ ] Missing spec files still detected
- [ ] Invalid path formats still rejected

### Quality Requirements

- [ ] Tests cover all false positive scenarios
- [ ] Tests cover all valid reference scenarios
- [ ] Error messages are clear and actionable
- [ ] Checker is deterministic (same results on repeated runs)
- [ ] No regressions in other checks

### Integration Requirements

- [ ] `./scripts/run-all-checks.sh` passes
- [ ] CI validation passes
- [ ] No changes to other validation scripts
- [ ] Documentation updated

---

## Risk Mitigation

### Potential Risks

1. **Over-filtering:**
   - **Risk**: Valid references might be filtered out
   - **Mitigation**: Test thoroughly with legitimate references

2. **Performance:**
   - **Risk**: Go tool might be slower than shell script
   - **Mitigation**: Profile and optimize if needed; acceptable if < 1 second

3. **Complexity:**
   - **Risk**: Go tool adds maintenance burden
   - **Mitigation**: Keep implementation simple; well-tested; documented

4. **Regressions:**
   - **Risk**: Changes might break existing validation
   - **Mitigation**: Comprehensive tests; run full check suite

---

## Alternative: Shell Script Hardening

If Go tool is deemed too heavy, harden the existing shell script:

### Improved Regex Pattern

```bash
# Only match comment-style references
if [[ $line =~ ^[[:space:]]*//[[:space:]]+[Ss]pec:[[:space:]]+(spec/[^[:space:]]+\.md) ]]; then
    SPEC_FILE="${BASH_REMATCH[1]}"
    # Validate path format
    if [[ ! $SPEC_FILE =~ ^spec/.*\.md$ ]]; then
        continue
    fi
    # Check excluded directories
    if [[ $SPEC_FILE =~ (testdata/|test/e2e/) ]]; then
        continue
    fi
    # Check file existence...
fi
```

**Pros:**
- No new dependencies
- Minimal changes
- Faster to implement

**Cons:**
- Harder to test
- Less maintainable
- More error-prone

**Recommendation**: Go tool is preferred for long-term maintainability.

---

## Related Documents

- `spec/governance/GOV_CORE.md` – Governance feature specification
- `docs/coverage/COVERAGE_COMPLIANCE_PLAN_PHASE2.md` – Phase 2 plan (Phase 3 section)
- `scripts/run-all-checks.sh` – Current implementation
- `Agent.md` – Development protocol

---

**Last Updated**: 2025-12-08


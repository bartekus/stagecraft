> **Superseded by** `docs/coverage/COVERAGE_LEDGER.md` section 4.2 (Phase 2 - Provider Coverage Enforcement). Kept for historical reference. New coverage snapshots and summaries MUST go into the coverage ledger.

# Test Coverage Quality Lift - Phase 2 Plan

**Feature**: GOV_CORE  
**Status**: Phase 2 - Quality Lift (Non-Blocking)  
**Date**: 2025-12-07

---

## Overview

Phase 2 focuses on raising coverage for low-coverage, non-core packages to stable, maintainable baselines. This is a **quality improvement** initiative, not a compliance blocker.

**Prerequisite**: Phase 1 must be complete before starting Phase 2.

**Current Status**: ⚠️ Quality Improvement Needed
- `internal/git`: 46.9% (target ≥ 70%)
- `internal/tools/docs`: 37.9% (target ≥ 60%)
- `internal/providers/migration/raw`: 33.3% (target ≥ 70%)

**Target Status**: ✅ Quality Baselines Met
- All three packages meet their respective targets
- No new test failures introduced
- Behavioural stability maintained

---

## Scope

### In Scope

**Three low-coverage packages** identified in `docs/coverage/TEST_COVERAGE_ANALYSIS.md`:

1. **`internal/git`** (46.9% → ≥ 70%)
2. **`internal/tools/docs`** (37.9% → ≥ 60%)
3. **`internal/providers/migration/raw`** (33.3% → ≥ 70%)

### Out of Scope

- Core packages (`pkg/config`, `internal/core`) - covered in Phase 1
- Coverage threshold changes in scripts
- User-facing behaviour changes
- New dependencies or large refactors
- Protected files (LICENSE, README, ADRs)

---

## Package-Specific Plans

### Phase 2.A: `internal/git` Coverage Lift

**Current**: 46.9%  
**Target**: ≥ 70%  
**Gap**: 23.1 percentage points

#### Strategy

**Focus Areas:**
1. **Happy-path flows:**
   - Repository detection
   - Branch name resolution
   - Dirty/clean working tree detection

2. **Error paths:**
   - Non-git directory handling
   - Command execution failures
   - Malformed git output

3. **Determinism:**
   - Stable outputs and error messages

#### Test Additions Needed

**File**: `internal/git/git_test.go`

1. **Repository detection tests:**
   - Valid git repository
   - Non-git directory
   - Missing `.git` directory

2. **Branch resolution tests:**
   - Current branch name
   - Detached HEAD state
   - Branch name with special characters

3. **Working tree state tests:**
   - Clean working tree
   - Dirty working tree (modified files)
   - Untracked files

4. **Error handling tests:**
   - Git command failures (mocked)
   - Malformed git output
   - Permission errors

#### Implementation Notes

- Use dependency injection or small internal helpers to avoid spawning real git where possible
- If shelling out is hard-coded, introduce minimal internal abstraction only if needed for testability
- Keep behaviour identical to current implementation

**Success Criteria:**
- `internal/git` coverage ≥ 70%
- All new tests pass
- No behaviour changes
- Tests are deterministic and isolated

---

### Phase 2.B: `internal/tools/docs` Coverage Lift

**Current**: 37.9%  
**Target**: ≥ 60%  
**Gap**: 22.1 percentage points

#### Strategy

**Focus Areas:**
1. **Happy-path processing:**
   - Loading and parsing documentation/spec inputs
   - Generating derived artifacts (indexes, summaries, validations)

2. **Failure paths:**
   - Missing or unreadable input files
   - Invalid document structures
   - Mismatched references (missing spec files)

3. **Determinism:**
   - File walking with deterministic ordering (sorted)
   - Stable outputs given same repo state

#### Test Additions Needed

**File**: `internal/tools/docs/docs_test.go`

1. **Document loading tests:**
   - Valid spec/document files
   - Missing files
   - Unreadable files (permissions)

2. **Parsing tests:**
   - Valid YAML/Markdown structures
   - Invalid document structures
   - Malformed frontmatter

3. **Reference validation tests:**
   - Valid spec references
   - Missing spec files
   - Circular references (if applicable)

4. **Output generation tests:**
   - Deterministic ordering
   - Stable output format
   - Error handling in generation

#### Implementation Notes

- Create minimal fixtures under `internal/tools/docs/testdata/`
- Avoid relying on full-repo state
- Prefer local `testdata/` directories
- Use smallest possible slice of docs tooling

**Success Criteria:**
- `internal/tools/docs` coverage ≥ 60%
- All new tests pass
- No behaviour changes
- Tests use synthetic fixtures

---

### Phase 2.C: `internal/providers/migration/raw` Coverage Lift

**Current**: 33.3%  
**Target**: ≥ 70%  
**Gap**: 36.7 percentage points

#### Strategy

**Focus Areas:**
1. **Happy path:**
   - Discovering migrations from expected locations
   - Applying migrations in deterministic, correct order
   - Idempotent behaviour (where specified)

2. **Error paths:**
   - Missing or unreadable migration files
   - Invalid/parse-failing SQL migrations
   - Partial application failures and error surfacing

3. **Determinism:**
   - File reading order (sorted)
   - Stable error messages and logging patterns

#### Test Additions Needed

**File**: `internal/providers/migration/raw/raw_test.go`

1. **Migration discovery tests:**
   - Valid migration files
   - Missing migration directory
   - Empty migration directory

2. **Migration ordering tests:**
   - Correct sort order (lexicographical)
   - Version-based ordering (if applicable)
   - Deterministic sequence

3. **Migration application tests:**
   - Successful application
   - Idempotent behaviour
   - Partial failure handling

4. **Error handling tests:**
   - Invalid SQL syntax
   - Missing migration files
   - Permission errors
   - Database connection failures (mocked)

#### Implementation Notes

- Use temporary directories and synthetic SQL files under `testdata/`
- Avoid real database connections
- Prefer mocks, fakes, or minimal embedded test harness
- Goal is to exercise engine control flow, not database itself

**Success Criteria:**
- `internal/providers/migration/raw` coverage ≥ 70%
- All new tests pass
- No behaviour changes
- Tests use synthetic fixtures

---

## Execution Order

1. **`internal/git`** (Phase 2.A)
   - Start here as it's likely the most straightforward
   - Git helpers typically have clear, testable interfaces

2. **`internal/tools/docs`** (Phase 2.B)
   - May require more fixture setup
   - Good to do after git (builds momentum)

3. **`internal/providers/migration/raw`** (Phase 2.C)
   - Most complex (may need more mocking)
   - Do last to leverage lessons from previous packages

4. **Final validation**
   - Run all checks
   - Verify no regressions
   - Update coverage documentation

---

## Success Criteria (Phase 2 Complete)

### Coverage Requirements
- [ ] `internal/git` ≥ 70%: **Target** (currently 46.9%)
- [ ] `internal/tools/docs` ≥ 60%: **Target** (currently 37.9%)
- [ ] `internal/providers/migration/raw` ≥ 70%: **Target** (currently 33.3%)

### Test Quality Requirements
- [ ] All tests passing: **0 failures**
- [ ] No new test failures introduced
- [ ] All new tests are deterministic and isolated

### Validation
- [ ] `./scripts/run-all-checks.sh` passes
- [ ] `./scripts/check-coverage.sh --fail-on-warning` still passes
- [ ] No new violations introduced
- [ ] All new/updated test files include Feature ID headers

---

## Constraints

### Must Follow
- **Agent.md** strictly:
  - Single feature scope: GOV_CORE
  - Test-first: write tests before implementation changes
  - No refactors outside direct scope
  - No new dependencies
  - Keep diffs minimal and deterministic

### Must Not
- Change coverage thresholds or scripts
- Change user-facing behaviour (unless fixing documented bug)
- Add unnecessary complexity
- Refactor code beyond what's needed for testability
- Modify protected files

---

## Risk Mitigation

### Potential Risks

1. **Git command dependencies**
   - **Risk**: Tests may require real git commands
   - **Mitigation**: Use mocks/fakes, minimal abstraction if needed

2. **Docs tooling complexity**
   - **Risk**: Full-repo state dependencies
   - **Mitigation**: Use synthetic fixtures in `testdata/`

3. **Migration engine database dependencies**
   - **Risk**: Tests may require real database
   - **Mitigation**: Use mocks/fakes, focus on control flow

4. **Coverage measurement accuracy**
   - **Risk**: Coverage may not reflect actual test quality
   - **Mitigation**: Focus on meaningful test cases, not just coverage numbers

---

## Phase 3+ (Future Work)

**Not part of Phase 2** - These are additional quality improvements:

- Improve coverage in medium-coverage packages (60-79%)
- Add integration test scenarios
- Enhance E2E test coverage
- Add performance/load tests where relevant
- Coverage trend monitoring and alerts

See `TEST_COVERAGE_ANALYSIS.md` for detailed recommendations.

---

## Related Documents

- `docs/coverage/TEST_COVERAGE_ANALYSIS.md` - Detailed coverage analysis and findings
- `docs/coverage/COVERAGE_COMPLIANCE_PLAN.md` - Phase 1 and overall plan
- `docs/agents/AGENT_BRIEF_COVERAGE_PHASE1.md` - Phase 1 agent brief
- `docs/agents/AGENT_BRIEF_COVERAGE_PHASE2.md` - Phase 2 agent brief (ready-to-paste)
- `spec/governance/GOV_CORE.md` - Governance feature specification
- `Agent.md` - Development protocol

---

**Last Updated**: 2025-12-07


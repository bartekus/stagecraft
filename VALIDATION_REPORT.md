# Stagecraft Validation Report

**Generated**: 2025-01-XX  
**Scope**: Complete structural analysis of repository integrity  
**Methodology**: GOV_V1_CORE, STRUC-C/L, AATSE alignment

---

## 1. Governance Consistency

### Summary

The governance framework (GOV_V1_CORE) is implemented and functional, but several inconsistencies exist between `spec/features.yaml` and actual artifacts. The Feature Mapping Invariant has violations that need correction.

### Issues Found

#### 1.1 Spec Status Mismatches

**Critical**: Spec file frontmatter status does not match `spec/features.yaml`:

- **`spec/commands/commit-suggest.md`**: Frontmatter declares `status: todo`, but `spec/features.yaml` shows `CLI_COMMIT_SUGGEST` as `status: done`
  - **Location**: `spec/commands/commit-suggest.md:4`
  - **Impact**: Governance validation may incorrectly flag this as incomplete

- **`spec/dev/process-mgmt.md`**: Frontmatter declares `status: wip`, but `spec/features.yaml` shows `DEV_PROCESS_MGMT` as `status: done`
  - **Location**: `spec/dev/process-mgmt.md:4`
  - **Impact**: Status inconsistency violates governance rules

- **`spec/overview.md`**: Frontmatter declares `status: todo`, matches `spec/features.yaml` (ARCH_OVERVIEW is todo) ✅

#### 1.2 Orphan Spec Reference

**Critical**: `DRIVER_DO` feature references non-existent spec file:

- **Feature**: `DRIVER_DO` (status: cancelled)
- **Referenced spec**: `spec/drivers/do.md` (does not exist)
- **Referenced test**: `internal/drivers/do/do_test.go` (directory does not exist)
- **Location**: `spec/features.yaml:305-312`
- **Impact**: Dead reference violates Feature Mapping Invariant
- **Note**: ADR 0003 documents cancellation, but cleanup incomplete

#### 1.3 Feature ID Header Mismatches

**High Priority**: Spec file declares wrong Feature ID:

- **`spec/commands/commit-suggest.md`**: Frontmatter declares `feature: GOV_V1_CORE` but should be `CLI_COMMIT_SUGGEST`
  - **Location**: `spec/commands/commit-suggest.md:2`
  - **Impact**: Spec validation will fail, feature mapping incorrect
  - **Note**: Implementation files (`commit_suggest.go`, `commit_suggest_test.go`) correctly use `Feature: CLI_COMMIT_SUGGEST` ✅

#### 1.4 Outdated Overview Document

**Medium Priority**: `docs/features/OVERVIEW.md` shows outdated statuses:

- Shows `DRIVER_DO` as `todo` (should reflect `cancelled`)
- Shows `CLI_DEV` as `todo` (should be `done` per features.yaml)
- Shows `CLI_INFRA_UP` as `todo` (should be `done` per features.yaml)
- Shows `PROVIDER_CLOUD_DO` as `todo` (should be `done` per features.yaml)
- Shows `PROVIDER_NETWORK_TAILSCALE` as `todo` (should be `done` per features.yaml)
- Shows `DEV_*` features as `todo` (many are `done` per features.yaml)
- **Location**: `docs/features/OVERVIEW.md`
- **Impact**: Misleading documentation, violates governance requirement for current overview
- **Note**: Document claims to be auto-generated but appears stale

### Recommended Corrections

1. **Fix spec frontmatter status**:
   - Update `spec/commands/commit-suggest.md` frontmatter: `status: done`
   - Update `spec/dev/process-mgmt.md` frontmatter: `status: done`
   - Verify all spec files match `spec/features.yaml` status

2. **Fix Feature ID in commit-suggest spec**:
   - Update `spec/commands/commit-suggest.md` frontmatter: `feature: CLI_COMMIT_SUGGEST`

3. **Fix Feature ID headers in implementation files**:
   - Update all `CLI_COMMIT_SUGGEST` implementation files to use correct Feature ID
   - Change `Feature: GOV_V1_CORE` → `Feature: CLI_COMMIT_SUGGEST` in:
     - `internal/cli/commands/commit_suggest.go`
     - `internal/cli/commands/commit_suggest_test.go` (if exists)
     - `internal/reports/suggestions/suggestions.go`
     - `internal/reports/suggestions/suggestions_test.go`

4. **Clean up DRIVER_DO references**:
   - Remove `spec: "drivers/do.md"` from `spec/features.yaml` (or mark as deprecated)
   - Remove `tests: ["internal/drivers/do/do_test.go"]` from `spec/features.yaml`
   - Update `docs/features/OVERVIEW.md` to reflect `cancelled` status
   - Update `docs/engine/status/implementation-status.md` to remove dead references

5. **Regenerate overview document**:
   - Run `go run ./cmd/gen-features-overview` to regenerate `docs/features/OVERVIEW.md`
   - Verify it matches current `spec/features.yaml` state

---

## 2. Spec/Test/Implementation Traceability

### Missing

#### 2.1 Missing Spec Files

**None identified** - All features in `spec/features.yaml` either have spec files or are `todo` status (which allows missing specs).

#### 2.2 Missing Implementation Files

**None identified** - All `done` features have implementation files with correct Feature ID headers ✅

**Note**: `CORE_BACKEND_PROVIDER_CONFIG_SCHEMA` has correct Feature ID headers in `pkg/config/config.go` and `pkg/config/config_test.go` ✅

#### 2.3 Missing Test Files

**High Priority**: Features marked `done` but missing test files:

- **`PROVIDER_BACKEND_INTERFACE`**: Listed in `spec/features.yaml` with test `pkg/providers/backend/backend_test.go`, but file exists and has `Feature: PROVIDER_BACKEND_INTERFACE` header ✅
  - **Note**: File exists, may have been added since last analysis

- **`CLI_DEPLOY`**: Listed with test `test/e2e/deploy_smoke_test.go`, file exists ✅

### Outdated

#### 2.1 Specs Not Matching Implementation

**Medium Priority**: Specs that may not reflect current implementation:

- **`spec/commands/commit-suggest.md`**: Frontmatter declares wrong feature ID (`GOV_V1_CORE` instead of `CLI_COMMIT_SUGGEST`)
  - **Impact**: Spec validation will fail

- **`spec/overview.md`**: References "Driver" concept (line 29) which was removed per ADR 0003
  - **Location**: `spec/overview.md:29`
  - **Impact**: Outdated architectural documentation

#### 2.2 Tests Referencing Removed Behavior

**None identified** - No tests found referencing removed `driver/do` subsystem.

### Conflicting

#### 2.1 Feature ID Conflicts

**Critical**: `CLI_COMMIT_SUGGEST` spec file declares wrong Feature ID:

- **`spec/commands/commit-suggest.md`**: Frontmatter declares `feature: GOV_V1_CORE` but should be `CLI_COMMIT_SUGGEST`
- **Implementation files**: Correctly use `Feature: CLI_COMMIT_SUGGEST` ✅
  - `internal/cli/commands/commit_suggest.go` ✅
  - `internal/cli/commands/commit_suggest_test.go` ✅
- **Impact**: Spec validation will fail, feature mapping may be confused

#### 2.2 Spec Path Mismatches

**High Priority**: Implementation files declare wrong spec paths:

- Multiple files declare `Spec: spec/commands/commit-suggest.md` but have `Feature: GOV_V1_CORE`
- Should be: `Feature: CLI_COMMIT_SUGGEST`, `Spec: spec/commands/commit-suggest.md`

### Redundant

#### 2.1 Duplicate Feature References

**None identified** - No duplicate feature IDs found.

#### 2.2 Redundant Test Coverage

**None identified** - Test coverage appears appropriately scoped.

---

## 3. Orphan or Dead Artifacts

### Files

#### 3.1 Referenced But Non-Existent

- **`spec/drivers/do.md`**: Referenced in `spec/features.yaml:308` for cancelled feature `DRIVER_DO`
  - **Status**: Feature cancelled per ADR 0003
  - **Action**: Remove reference from `spec/features.yaml`

- **`internal/drivers/do/do_test.go`**: Referenced in `spec/features.yaml:311` for cancelled feature
  - **Status**: Directory does not exist (verified)
  - **Action**: Remove reference from `spec/features.yaml`

#### 3.2 Potentially Unused Files

**None identified** - All Go files appear to be referenced or part of active features.

### Specs

#### 3.1 Orphan Specs

**None identified** - All spec files in `spec/` directory are referenced by `spec/features.yaml` or are ADRs (which are not features).

**Note**: ADRs are correctly placed in `spec/adr/` and should NOT be treated as feature specs per governance rules ✅

### Tests

#### 3.1 Orphan Test Files

**None identified** - All test files appear to be associated with features.

### Scripts

#### 3.1 Deprecated Scripts

**None identified** - All scripts in `scripts/` appear to be actively used.

#### 3.2 Legacy Examples

**None identified** - Example directories appear current.

### Documentation

#### 3.1 Stale Documentation

- **`docs/features/OVERVIEW.md`**: Contains outdated feature statuses (see Section 1.4)
  - **Action**: Regenerate using `go run ./cmd/gen-features-overview`

- **`docs/engine/status/implementation-status.md`**: References `DRIVER_DO` with path to non-existent spec
  - **Location**: Line 78
  - **Action**: Update to reflect cancelled status or remove entry

- **`ASSESSMENT.md`**: References "Driver Layer (`internal/drivers/`)" as "planned, not yet implemented" (line 21)
  - **Status**: ADR 0003 removed driver layer concept
  - **Action**: Update to reflect current architecture

- **`spec/overview.md`**: References "Driver" concept (line 29) which was removed
  - **Action**: Update to use "Provider" terminology

---

## 4. TODO/FIXME/HACK Analysis

### MUST Fix (Blocking Correctness)

#### 4.1 Governance Violations

- **`spec/commands/commit-suggest.md`**: Wrong feature ID in frontmatter
  - **Priority**: MUST
  - **Impact**: Blocks governance validation
  - **Action**: Change `feature: GOV_V1_CORE` → `feature: CLI_COMMIT_SUGGEST`

- **Implementation files for `CLI_COMMIT_SUGGEST`**: Wrong Feature ID headers
  - **Priority**: MUST
  - **Impact**: Feature mapping fails
  - **Action**: Update all files to use `Feature: CLI_COMMIT_SUGGEST`

#### 4.2 Spec Status Mismatches

- **`spec/commands/commit-suggest.md`**: Status mismatch (`todo` vs `done`)
- **`spec/dev/process-mgmt.md`**: Status mismatch (`wip` vs `done`)
  - **Priority**: MUST
  - **Impact**: Governance validation incorrect
  - **Action**: Update frontmatter to match `spec/features.yaml`

### SHOULD Fix (Tech Debt)

#### 4.1 TODO Comments in Code

- **`internal/infra/bootstrap/tailscale.go:49`**: `TODO: Extract network config from bootstrap.Config when it's added`
  - **Priority**: SHOULD
  - **Impact**: Incomplete implementation, may cause issues later
  - **Action**: Either implement or document why deferred

#### 4.2 Documentation TODOs

- **`docs/todo/COMMIT_MESSAGE_ENFORCEMENT_PHASE2.md`**: Phase 2 planning document
  - **Priority**: SHOULD (if Phase 2 is planned)
  - **Impact**: Planning document, not blocking
  - **Action**: Review and either implement or archive

#### 4.3 Dead References

- **`spec/features.yaml`**: `DRIVER_DO` references non-existent files
  - **Priority**: SHOULD
  - **Impact**: Confusing, violates governance
  - **Action**: Remove references or mark clearly as cancelled

### COULD Fix (Cosmetic)

#### 4.1 Note Comments

- Multiple `// Note:` comments throughout codebase
  - **Priority**: COULD
  - **Impact**: Documentation quality
  - **Action**: Review and convert to proper GoDoc where appropriate

#### 4.2 Status Comments in Specs

- **`spec/providers/backend/generic.md:14`**: `Status: todo` in body (should be in frontmatter only)
  - **Priority**: COULD
  - **Impact**: Redundant, may cause confusion
  - **Action**: Remove redundant status from body

### TODOs Requiring Feature Spec Updates

**None identified** - All TODOs appear to be implementation details or future work, not spec violations.

### TODOs Violating Governance

**None identified** - No TODOs found that violate Stagecraft governance or coding discipline.

---

## 5. API & Interface Consistency

### Provider Interfaces

#### 5.1 Interface Definitions Match Specs ✅

**Status**: Provider interfaces in `pkg/providers/*` match their spec definitions:

- **`PROVIDER_BACKEND_INTERFACE`**: ✅ Matches `spec/providers/backend/interface.md` (via `core/backend-registry.md`)
- **`PROVIDER_FRONTEND_INTERFACE`**: ✅ Matches `spec/providers/frontend/interface.md`
- **`PROVIDER_NETWORK_INTERFACE`**: ✅ Matches `spec/providers/network/interface.md`
- **`PROVIDER_CLOUD_INTERFACE`**: ✅ Matches `spec/providers/cloud/interface.md`
- **`PROVIDER_CI_INTERFACE`**: ✅ Matches `spec/providers/ci/interface.md`
- **`PROVIDER_SECRETS_INTERFACE`**: ✅ Matches `spec/providers/secrets/interface.md`

#### 5.2 Interface Status Mismatch

**Medium Priority**: Spec frontmatter shows `status: todo` but implementation exists:

- **`spec/providers/cloud/interface.md:14`**: Body says `Status: todo` but frontmatter says `status: done`
  - **Impact**: Confusing, but frontmatter is authoritative
  - **Action**: Remove redundant status from body

- **`spec/providers/secrets/interface.md:14`**: Body says `Status: todo` but frontmatter says `status: done`
  - **Action**: Remove redundant status from body

### Config Schemas

#### 5.1 Config Schema Consistency ✅

**Status**: Config schemas in `pkg/config` appear consistent with specs.

**Note**: `CORE_BACKEND_PROVIDER_CONFIG_SCHEMA` feature may need Feature ID headers added to config files (see Section 2.2).

### CLI Command Expectations

#### 5.1 CLI Flags vs Specs

**Status**: CLI flags appear to be validated by `internal/tools/specvscli` tool per GOV_V1_CORE.

**Note**: No manual verification performed - relies on automated tooling.

---

## 6. Structural Health

### Layering Violations

#### 6.1 Package Import Rules ✅

**Status**: No violations found. Package boundaries appear respected:
- `pkg/` does not import `internal/` ✅
- `internal/` may import `pkg/` ✅
- `cmd/` stays thin ✅

#### 6.2 ADR References

**Low Priority**: `ASSESSMENT.md` references outdated architecture:

- Mentions "Driver Layer (`internal/drivers/`)" as planned
- ADR 0003 removed driver concept
- **Action**: Update `ASSESSMENT.md` to reflect current architecture

### Coding Anti-Patterns

#### 6.1 God Files

**None identified** - File sizes appear reasonable. Largest files are test files with comprehensive coverage, which is acceptable.

#### 6.2 Redundant Helpers

**None identified** - Helper functions appear appropriately scoped.

### Topology Issues

#### 6.1 Infrastructure Leakage

**None identified** - Core remains provider-agnostic ✅

#### 6.2 Runtime Coupling

**None identified** - Topology boundaries appear respected ✅

---

## 7. Test Suite Quality

### Flaky Patterns

#### 7.1 Time-Based Flakiness

**High Priority**: Multiple tests use `time.Sleep` which can cause flakiness:

- **`internal/cli/commands/releases_test.go:142`**: `time.Sleep(10 * time.Millisecond)`
  - **Impact**: Race condition potential
  - **Action**: Use deterministic synchronization instead

- **`internal/providers/frontend/generic/generic_test.go`**: Multiple `time.Sleep` calls:
  - Line 357: `time.Sleep(500 * time.Millisecond)`
  - Line 407: `time.Sleep(500 * time.Millisecond)`
  - Line 519: `time.Sleep(500 * time.Millisecond)`
  - Line 670: `time.Sleep(500 * time.Millisecond)`
  - Line 717: `time.Sleep(500 * time.Millisecond)`
  - Line 764: `time.Sleep(500 * time.Millisecond)`
  - Line 812: `time.Sleep(50 * time.Millisecond)`
  - **Impact**: High flakiness risk, especially under load
  - **Action**: Refactor to use deterministic synchronization (channels, context cancellation)

#### 7.2 Goroutine-Based Tests

**Medium Priority**: Tests using goroutines without proper synchronization:

- **`internal/core/state/state_test.go:774`**: Uses `go func()` for concurrent testing
  - **Status**: Appears to have synchronization (needs review)
  - **Action**: Verify synchronization is deterministic

- **`internal/providers/frontend/generic/generic_test.go`**: Multiple `go func()` calls with `time.Sleep`
  - **Impact**: Non-deterministic test execution
  - **Action**: Refactor to deterministic patterns per `COVERAGE_STRATEGY.md`

- **`internal/dev/process/runner.go:161`**: Uses `go func()` in implementation
  - **Status**: Implementation code, not test - acceptable if properly synchronized
  - **Action**: Verify synchronization in tests

#### 7.3 Timeout-Based Tests

**Medium Priority**: Tests using `time.After` for timeouts:

- **`internal/providers/frontend/generic/generic_test.go:48`**: `case <-time.After(timeout):`
  - **Impact**: May be flaky under load
  - **Action**: Consider deterministic alternatives

- **`internal/providers/frontend/generic/generic.go:313`**: `case <-time.After(timeout):` in implementation
  - **Status**: Implementation code - acceptable if spec requires timeout behavior
  - **Action**: Verify timeout behavior is spec-compliant

### Missing Coverage

#### 7.1 Error Paths

**Medium Priority**: Some packages have low coverage on error paths:

- **`pkg/config`**: 66.7% coverage (below 80% threshold for core packages)
  - **Impact**: Error handling may be untested
  - **Action**: Add tests for error paths per coverage compliance plan

- **`internal/git`**: 46.9% coverage
  - **Impact**: Low coverage on error handling
  - **Action**: Improve coverage per Phase 2 plan

#### 7.2 Edge Conditions

**Medium Priority**: Edge conditions may be untested:

- Empty config scenarios
- Invalid provider configurations
- Network failures in provider code

**Action**: Review coverage reports for specific gaps.

### Fragile Tests

#### 7.1 OS-Dependent Tests

**Low Priority**: Some tests note OS-dependent behavior:

- **`internal/dev/hosts/hosts_test.go:26`**: Comment notes "This test runs on the current platform"
  - **Status**: Acceptable if documented
  - **Action**: Consider cross-platform testing if critical

### Opportunities for Determinism

#### 7.1 Frontend Generic Provider

**High Priority**: `internal/providers/frontend/generic` has extensive non-deterministic patterns:

- Multiple `time.Sleep` calls in tests
- Goroutine-based test patterns
- **Reference**: `internal/providers/frontend/generic/COVERAGE_STRATEGY.md` documents strategy for deterministic testing
- **Action**: Follow COVERAGE_STRATEGY.md recommendations to refactor tests

---

## 8. Coverage Integrity

### Coverage Gaps

#### 8.1 Core Package Coverage ❌

**Critical**: Core packages below 80% threshold:

- **`pkg/config`**: 66.7% (below 80% threshold)
  - **Gap**: 13.3 percentage points
  - **Impact**: Core package threshold violation
  - **Action**: Add tests per `docs/coverage/COVERAGE_COMPLIANCE_PLAN.md`

- **`internal/core`**: 83.9% ✅ (above threshold)

- **Combined Average**: 74.2% (below 80% threshold) ❌

#### 8.2 Low Coverage Packages

**Medium Priority**: Packages below 60% coverage:

- **`internal/git`**: 46.9%
- **`internal/tools/docs`**: 37.9%
- **`internal/providers/migration/raw`**: 33.3%

**Action**: Improve per Phase 2 coverage plan.

### Missing Branch Tests

#### 8.1 Untested Code Paths

**Medium Priority**: Specific untested paths identified in coverage analysis:

- **`pkg/config`**: `GetProviderConfig` (overload 2): 0.0% coverage
  - **Action**: Add tests for second overload

- **`internal/core`**: `ExecuteBuild`: 0.0% coverage
  - **Action**: Add tests per coverage compliance plan

### Suggested New Tests

#### 8.1 Error Path Tests

- Add tests for config loading errors
- Add tests for provider validation failures
- Add tests for network/provider errors

#### 8.2 Edge Case Tests

- Empty configuration scenarios
- Invalid provider configurations
- Boundary conditions in state management

**Reference**: See `docs/coverage/COVERAGE_COMPLIANCE_PLAN.md` for detailed test requirements.

---

## 9. AATSE Alignment

### Where Applied Correctly ✅

#### 9.1 Spec-Driven Development

**Status**: Strong adherence to spec-first approach:
- All features have specs in `spec/` directory ✅
- Specs have structured frontmatter per GOV_V1_CORE ✅
- Implementation follows specs ✅

#### 9.2 Deterministic Implementation

**Status**: Most code is deterministic:
- No randomness in core logic ✅
- No timestamps in deterministic operations ✅
- Provider-agnostic core ✅

#### 9.3 Test Isolation

**Status**: Tests are well-isolated:
- State test isolation implemented (`CORE_STATE_TEST_ISOLATION`) ✅
- Test helpers provide deterministic state ✅

#### 9.4 Topological Boundaries

**Status**: Boundaries respected:
- Core remains provider-agnostic ✅
- Infrastructure concerns isolated ✅
- Registry pattern enforces boundaries ✅

### Where Violated

#### 9.1 Non-Deterministic Tests

**High Priority**: `internal/providers/frontend/generic` tests violate determinism:

- Multiple `time.Sleep` calls
- Goroutine-based test patterns without proper synchronization
- **Impact**: Flaky tests, violates AATSE principle
- **Action**: Refactor per `COVERAGE_STRATEGY.md`

#### 9.2 Runtime Coupling

**Low Priority**: Some tests may have runtime coupling:

- Tests using actual file system operations
- Tests using actual process execution
- **Status**: May be acceptable if properly isolated
- **Action**: Review for determinism opportunities

### Where to Improve

#### 9.1 Test Determinism

**High Priority**: Improve test determinism:

- Replace `time.Sleep` with deterministic synchronization
- Refactor goroutine-based tests to use channels/contexts
- Follow `COVERAGE_STRATEGY.md` recommendations

#### 9.2 Function Extraction

**Medium Priority**: Some logic could be more isolated:

- Review for opportunities to extract deterministic functions
- Consider `scanStream`-like patterns for other operations

#### 9.3 Infrastructure Leakage

**None identified** - Core remains clean ✅

---

## 10. Final Checklist

### High-Priority Actions (MUST Fix)

1. **Fix `CLI_COMMIT_SUGGEST` spec Feature ID**:
   - [ ] Update `spec/commands/commit-suggest.md` frontmatter: `feature: CLI_COMMIT_SUGGEST` (currently `GOV_V1_CORE`)
   - [ ] Update `spec/commands/commit-suggest.md` frontmatter: `status: done` (currently `todo`)
   - **Note**: Implementation files already have correct Feature ID headers ✅

2. **Fix spec status mismatches**:
   - [ ] Update `spec/dev/process-mgmt.md` frontmatter: `status: done`

3. **Clean up DRIVER_DO references**:
   - [ ] Remove `spec: "drivers/do.md"` from `spec/features.yaml`
   - [ ] Remove `tests: ["internal/drivers/do/do_test.go"]` from `spec/features.yaml`
   - [ ] Update `docs/features/OVERVIEW.md` to reflect cancelled status
   - [ ] Update `docs/engine/status/implementation-status.md`

4. **Verify Feature ID headers** (already present ✅):
   - [x] `Feature: CORE_BACKEND_PROVIDER_CONFIG_SCHEMA` in `pkg/config/config.go` ✅
   - [x] `Feature: CORE_BACKEND_PROVIDER_CONFIG_SCHEMA` in `pkg/config/config_test.go` ✅

5. **Improve test determinism**:
   - [ ] Refactor `internal/providers/frontend/generic/generic_test.go` to remove `time.Sleep`
   - [ ] Replace goroutine patterns with deterministic synchronization
   - [ ] Fix `internal/cli/commands/releases_test.go` sleep pattern

6. **Improve core package coverage**:
   - [ ] Increase `pkg/config` coverage from 66.7% to 80%+
   - [ ] Add tests for `GetProviderConfig` overload 2
   - [ ] Add tests for `ExecuteBuild` in `internal/core`

### Medium-Priority Actions (SHOULD Fix)

7. **Regenerate overview document**:
   - [ ] Run `go run ./cmd/gen-features-overview` to update `docs/features/OVERVIEW.md`

8. **Update outdated documentation**:
   - [ ] Update `ASSESSMENT.md` to remove driver layer references
   - [ ] Update `spec/overview.md` to remove driver terminology

9. **Fix redundant status in spec bodies**:
   - [ ] Remove `Status: todo` from body of `spec/providers/cloud/interface.md`
   - [ ] Remove `Status: todo` from body of `spec/providers/secrets/interface.md`
   - [ ] Remove `Status: todo` from body of `spec/providers/backend/generic.md`

10. **Address TODO in code**:
    - [ ] Review `internal/infra/bootstrap/tailscale.go:49` TODO and either implement or document deferral

11. **Improve low-coverage packages** (Phase 2):
    - [ ] Improve `internal/git` coverage (46.9% → 70%+)
    - [ ] Improve `internal/tools/docs` coverage (37.9% → 60%+)
    - [ ] Improve `internal/providers/migration/raw` coverage (33.3% → 70%+)

### Low-Priority Actions (COULD Fix)

12. **Review and improve Note comments**:
    - [ ] Convert `// Note:` comments to proper GoDoc where appropriate

13. **Cross-platform testing**:
    - [ ] Consider adding cross-platform tests for `internal/dev/hosts` if critical

---

## Summary Statistics

- **Total Features**: 87 (per `spec/features.yaml`)
- **Done Features**: 43
- **WIP Features**: 1 (`DEV_PROCESS_MGMT` - but should be `done`)
- **Todo Features**: 42
- **Cancelled Features**: 1 (`DRIVER_DO`)

- **Governance Violations**: 7+ (Feature ID mismatches, status mismatches, dead references)
- **Coverage Violations**: 1 (core package average below 80%)
- **Test Flakiness Issues**: 8+ (time.Sleep patterns, goroutine tests)
- **Dead Artifacts**: 2 (DRIVER_DO spec/test references)

---

**Report End**


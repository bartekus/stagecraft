# PROVIDER_NETWORK_TAILSCALE Slice 2 - Golden Coverage Expectations

**Expected coverage outcomes after Slice 2 implementation.**

---

## Executive Summary

**Current Coverage**: 71.3%  
**Target Coverage**: ~78-80%  
**Coverage Increase**: +6.7-8.7 percentage points

**Primary Focus**: `EnsureInstalled()` error paths and version enforcement logic.

---

## Coverage Breakdown by Function

### 1. New Functions (100% Coverage Target)

| Function | Current | After Slice 2 | Test Suite |
|----------|---------|---------------|------------|
| `parseTailscaleVersion()` | NEW | **100%** | `TestParseTailscaleVersion` (7 cases) |

**Test Cases**:
- ✅ Simple version (`1.78.0`)
- ✅ Version with build metadata (`1.44.0-123-gabcd`)
- ✅ Version with patch suffix (`1.78.0-1`)
- ✅ Version in output string (`tailscale version 1.78.0`)
- ✅ Unparseable version (`not-a-version`)
- ✅ Empty string
- ✅ Edge cases (various formats)

**Coverage Contribution**: +~0.5% (new function, small code size)

---

### 2. Modified Functions (Improved Coverage)

#### 2.1 `EnsureInstalled()`

| Metric | Current | After Slice 2 | Improvement |
|--------|---------|---------------|-------------|
| **Line Coverage** | ~60% | **~85%** | +25% |
| **Branch Coverage** | ~55% | **~80%** | +25% |
| **Error Path Coverage** | ~40% | **~90%** | +50% |

**Current Gaps** (to be covered):
- ❌ Config validation error paths (missing `auth_key_env`, `tailnet_domain`)
- ❌ Version parsing errors (unparseable versions)
- ❌ Version enforcement errors (below minimum)
- ❌ OS compatibility errors (unsupported OS)
- ❌ Install script failure paths
- ❌ Verification failure paths

**After Slice 2** (covered):
- ✅ Config validation error paths (5 test cases)
- ✅ Version parsing errors (1 test case)
- ✅ Version enforcement errors (1 test case)
- ✅ OS compatibility errors (3 test cases)
- ✅ Install script failure paths (1 test case)
- ✅ Verification failure paths (1 test case)

**Test Suites**:
- `TestEnsureInstalled_ConfigValidation` (5 cases)
- `TestEnsureInstalled_OSCompatibility` (7 cases)
- `TestEnsureInstalled_VersionEnforcement` (7 cases)
- `TestEnsureInstalled_InstallFlow` (4 cases)

**Coverage Contribution**: +~4-5% (large function, many branches)

---

#### 2.2 `checkOSCompatibility()`

| Metric | Current | After Slice 2 | Improvement |
|--------|---------|---------------|-------------|
| **Line Coverage** | ~70% | **~90%** | +20% |
| **Branch Coverage** | ~65% | **~85%** | +20% |

**Current Gaps** (to be covered):
- ❌ Unsupported OS detection (Alpine, CentOS, Darwin)
- ❌ `uname` failure graceful handling
- ❌ `os-release` missing, `lsb_release` fallback

**After Slice 2** (covered):
- ✅ Unsupported OS detection (3 test cases)
- ✅ `uname` failure graceful handling (1 test case)
- ✅ `os-release` missing, `lsb_release` fallback (1 test case)

**Test Suites**:
- `TestEnsureInstalled_OSCompatibility` (exercises all branches)

**Coverage Contribution**: +~1% (medium function, exercised via integration)

---

#### 2.3 `parseConfig()`

| Metric | Current | After Slice 2 | Improvement |
|--------|---------|---------------|-------------|
| **Line Coverage** | ~75% | **~90%** | +15% |
| **Branch Coverage** | ~70% | **~85%** | +15% |

**Current Gaps** (to be covered):
- ❌ Missing `auth_key_env` error path
- ❌ Missing `tailnet_domain` error path
- ❌ Invalid YAML error path

**After Slice 2** (covered):
- ✅ Missing `auth_key_env` error path (1 test case)
- ✅ Missing `tailnet_domain` error path (1 test case)
- ✅ Invalid YAML error path (implicitly via config tests)

**Test Suites**:
- `TestEnsureInstalled_ConfigValidation` (exercises error paths)

**Coverage Contribution**: +~1% (small function, error paths)

---

### 3. Unchanged Functions (No Coverage Change)

| Function | Current Coverage | After Slice 2 | Notes |
|----------|------------------|---------------|-------|
| `buildTailscaleUpCommand()` | 100% | 100% | Slice 1 complete |
| `parseOSRelease()` | 100% | 100% | Slice 1 complete |
| `validateTailnetDomain()` | 100% | 100% | Slice 1 complete |
| `buildNodeFQDN()` | 100% | 100% | Slice 1 complete |
| `NodeFQDN()` | ~85% | ~85% | Not in scope for Slice 2 |
| `EnsureJoined()` | ~65% | ~65% | Not in scope for Slice 2 |
| `computeTags()` | ~80% | ~80% | Not in scope for Slice 2 |
| `tagsMatch()` | ~75% | ~75% | Not in scope for Slice 2 |
| `parseStatus()` | ~85% | ~85% | Slice 1 edge cases complete |

**Coverage Contribution**: 0% (unchanged)

---

## Overall Package Coverage

### Current State (After Slice 1)

```
Total Coverage: 71.3%

Breakdown:
- Helper functions (Slice 1):     100% (4 functions)
- EnsureInstalled():                60% (happy paths + some errors)
- checkOSCompatibility():           70% (supported OS only)
- parseConfig():                     75% (happy paths + some errors)
- Other functions:                  ~65-85% (mixed)
```

### Expected State (After Slice 2)

```
Total Coverage: ~78-80%

Breakdown:
- Helper functions (Slice 1):     100% (4 functions)
- parseTailscaleVersion() (NEW):   100% (1 function)
- EnsureInstalled():               ~85% (happy + all error paths)
- checkOSCompatibility():           ~90% (all OS cases)
- parseConfig():                    ~90% (all error paths)
- Other functions:                 ~65-85% (unchanged)
```

### Coverage Delta

| Component | Before | After | Delta |
|-----------|--------|-------|-------|
| **Overall Package** | 71.3% | **78-80%** | **+6.7-8.7%** |
| `EnsureInstalled()` | 60% | **85%** | **+25%** |
| `checkOSCompatibility()` | 70% | **90%** | **+20%** |
| `parseConfig()` | 75% | **90%** | **+15%** |
| `parseTailscaleVersion()` | NEW | **100%** | **NEW** |

---

## Test Case Coverage Matrix

### Config Validation (5 test cases)

| Test Case | Coverage Target | Expected Result |
|-----------|-----------------|-----------------|
| Missing `auth_key_env` | `parseConfig()` error path | ✅ Error returned |
| Missing `tailnet_domain` | `parseConfig()` error path | ✅ Error returned |
| Invalid YAML | `parseConfig()` error path | ✅ Error returned |
| Valid config | `EnsureInstalled()` happy path | ✅ Success |
| `install.method = "skip"` | `EnsureInstalled()` skip path | ✅ Early return |

**Coverage Impact**: +~1% (config parsing error paths)

---

### OS Compatibility (7 test cases)

| Test Case | Coverage Target | Expected Result |
|-----------|-----------------|-----------------|
| Debian supported | `checkOSCompatibility()` happy path | ✅ Success |
| Ubuntu supported | `checkOSCompatibility()` happy path | ✅ Success |
| Alpine unsupported | `checkOSCompatibility()` error path | ✅ `ErrUnsupportedOS` |
| CentOS unsupported | `checkOSCompatibility()` error path | ✅ `ErrUnsupportedOS` |
| Darwin unsupported | `checkOSCompatibility()` error path | ✅ `ErrUnsupportedOS` |
| `uname` fails gracefully | `checkOSCompatibility()` fallback | ✅ Success (proceeds) |
| `os-release` missing, `lsb_release` fallback | `checkOSCompatibility()` fallback | ✅ Success (proceeds) |

**Coverage Impact**: +~1% (OS compatibility error paths and fallbacks)

---

### Version Enforcement (7 test cases)

| Test Case | Coverage Target | Expected Result |
|-----------|-----------------|-----------------|
| Version meets minimum | `EnsureInstalled()` version check | ✅ Success |
| Version exceeds minimum | `EnsureInstalled()` version check | ✅ Success |
| Version below minimum | `EnsureInstalled()` version check | ✅ `ErrInstallFailed` |
| Version with build metadata | `parseTailscaleVersion()` parsing | ✅ Parses correctly |
| Version with patch suffix | `parseTailscaleVersion()` parsing | ✅ Parses correctly |
| Unparseable version | `parseTailscaleVersion()` error | ✅ `ErrInstallFailed` |
| No `min_version` configured | `EnsureInstalled()` skip check | ✅ Success |

**Coverage Impact**: +~2-3% (version parsing and enforcement logic)

---

### Install Flow (4 test cases)

| Test Case | Coverage Target | Expected Result |
|-----------|-----------------|-----------------|
| Already installed | `EnsureInstalled()` early return | ✅ Success (no install) |
| Install succeeds | `EnsureInstalled()` install path | ✅ Success |
| Install fails | `EnsureInstalled()` error path | ✅ `ErrInstallFailed` |
| Verification fails | `EnsureInstalled()` error path | ✅ `ErrInstallFailed` |

**Coverage Impact**: +~1-2% (install script execution paths)

---

## Coverage Verification Commands

### Baseline (Before Slice 2)

```bash
go test -cover ./internal/providers/network/tailscale
# Expected: coverage: 71.3% of statements
```

### After Slice 2

```bash
go test -cover ./internal/providers/network/tailscale
# Expected: coverage: 78-80% of statements

go test -coverprofile=coverage.out ./internal/providers/network/tailscale
go tool cover -func=coverage.out | grep EnsureInstalled
# Expected: EnsureInstalled: ~85% coverage

go tool cover -func=coverage.out | grep parseTailscaleVersion
# Expected: parseTailscaleVersion: 100% coverage

go tool cover -func=coverage.out | grep checkOSCompatibility
# Expected: checkOSCompatibility: ~90% coverage
```

---

## Coverage by File

### `tailscale.go`

| Function | Current | After Slice 2 | Change |
|----------|---------|---------------|--------|
| `EnsureInstalled()` | 60% | **85%** | **+25%** |
| `checkOSCompatibility()` | 70% | **90%** | **+20%** |
| `parseTailscaleVersion()` | NEW | **100%** | **NEW** |
| Other functions | ~65-85% | ~65-85% | 0% |

**File Coverage**: ~72% → **~80%** (+8%)

---

### `tailscale_test.go`

| Test Suite | Test Cases | Coverage Contribution |
|------------|------------|----------------------|
| `TestParseTailscaleVersion` | 7 | `parseTailscaleVersion()`: 100% |
| `TestEnsureInstalled_ConfigValidation` | 5 | `EnsureInstalled()`: +5% |
| `TestEnsureInstalled_OSCompatibility` | 7 | `EnsureInstalled()`: +5%, `checkOSCompatibility()`: +20% |
| `TestEnsureInstalled_VersionEnforcement` | 7 | `EnsureInstalled()`: +10%, `parseTailscaleVersion()`: 100% |
| `TestEnsureInstalled_InstallFlow` | 4 | `EnsureInstalled()`: +5% |

**Total New Test Cases**: 30 (23 unique scenarios + 7 helper tests)

---

## Success Criteria

### Minimum Acceptable Coverage

- ✅ Overall package: **≥78%** (target: 78-80%)
- ✅ `EnsureInstalled()`: **≥80%** (target: ~85%)
- ✅ `parseTailscaleVersion()`: **100%** (target: 100%)
- ✅ `checkOSCompatibility()`: **≥85%** (target: ~90%)
- ✅ `parseConfig()`: **≥85%** (target: ~90%)

### Coverage Quality Metrics

- ✅ All error paths tested (23/23 test cases)
- ✅ All spec requirements covered
- ✅ No flaky tests (deterministic `LocalCommander`)
- ✅ No external dependencies (no real SSH, no real Tailscale)

---

## Coverage Gap Analysis

### Remaining Gaps (Post Slice 2)

**Not in Scope for Slice 2** (deferred to Slice 3+):

1. **`EnsureJoined()` error paths** (~35% uncovered)
   - Tag validation errors
   - Tailnet mismatch errors
   - Auth key missing errors
   - Status parsing errors

2. **Edge cases in existing functions**
   - `NodeFQDN()` edge cases (~15% uncovered)
   - `computeTags()` edge cases (~20% uncovered)
   - `tagsMatch()` edge cases (~25% uncovered)

**Estimated Coverage After Slice 3**: ~82-85%

---

## Confidence Intervals

### Coverage Prediction

| Scenario | Probability | Coverage Range |
|----------|-------------|----------------|
| **Optimistic** | 30% | 79-81% |
| **Expected** | 50% | 78-80% |
| **Pessimistic** | 20% | 77-79% |

**Rationale**:
- Optimistic: All test cases pass, no unexpected edge cases
- Expected: Most test cases pass, some minor adjustments needed
- Pessimistic: Some test cases need refinement, coverage slightly lower

**Recommendation**: Target **78%** as minimum acceptable, **80%** as stretch goal.

---

## Conclusion

**Expected Coverage Outcome**: **78-80%** (+6.7-8.7 percentage points)

**Primary Contributors**:
1. `EnsureInstalled()` error paths: +4-5%
2. Version parsing and enforcement: +2-3%
3. OS compatibility error paths: +1%
4. Config validation error paths: +1%

**Quality Metrics**: ✅ All criteria met

**Ready for Verification**: ✅ Yes (after implementation)

---

## Next Steps

1. ✅ Coverage expectations defined (this document)
2. ⏭️ Implement Slice 2 per Agent guide
3. ⏭️ Verify actual coverage matches expectations
4. ⏭️ Update expectations if needed based on actual results
5. ⏭️ Proceed to Slice 3 planning

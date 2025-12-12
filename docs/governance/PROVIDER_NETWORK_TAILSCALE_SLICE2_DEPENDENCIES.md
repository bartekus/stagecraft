# PROVIDER_NETWORK_TAILSCALE Slice 2 - Dependency Graph

**Visual representation of Slice 2 dependencies, implementation flow, and outputs.**

---

## Dependency Graph (ASCII)

```
┌─────────────────────────────────────────────────────────────────┐
│                    SLICE 2 IMPLEMENTATION                        │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                         PREREQUISITES                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────────────────┐      ┌──────────────────────┐        │
│  │   SLICE 1 OUTPUTS   │      │   EXISTING CODEBASE   │        │
│  └──────────────────────┘      └──────────────────────┘        │
│           │                              │                        │
│           │                              │                        │
│  ┌────────▼────────┐          ┌─────────▼─────────┐             │
│  │ Helper Functions │          │ Core Infrastructure│             │
│  ├──────────────────┤          ├────────────────────┤             │
│  │ • parseOSRelease │          │ • Commander        │             │
│  │ • validateTailnet│          │   interface       │             │
│  │   Domain         │          │ • LocalCommander  │             │
│  │ • buildNodeFQDN │          │ • Error types      │             │
│  │ • buildTailscale │          │   (ErrInstallFailed│            │
│  │   UpCommand      │          │    ErrUnsupportedOS)│            │
│  └──────────────────┘          └────────────────────┘            │
│           │                              │                        │
│           └──────────────┬──────────────┘                        │
│                          │                                        │
└──────────────────────────┼────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                    SLICE 2 IMPLEMENTATION                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              STEP 1: Version Parsing Helper             │  │
│  └──────────────────────────────────────────────────────────┘  │
│                          │                                        │
│                          ▼                                        │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  parseTailscaleVersion(versionStr string) (string, error) │  │
│  │  • Strips build metadata (1.44.0-123-gabcd → 1.44.0)      │  │
│  │  • Strips patch suffixes (1.78.0-1 → 1.78.0)            │  │
│  │  • Validates semantic version format                     │  │
│  │  • Returns error for unparseable versions                │  │
│  └──────────────────────────────────────────────────────────┘  │
│                          │                                        │
│                          ▼                                        │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              STEP 2: Update EnsureInstalled               │  │
│  └──────────────────────────────────────────────────────────┘  │
│                          │                                        │
│                          ▼                                        │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  EnsureInstalled() version logic                         │  │
│  │  • Replace strings.Contains with parseTailscaleVersion  │  │
│  │  • Add version comparison logic                          │  │
│  │  • Return spec-aligned error messages                    │  │
│  └──────────────────────────────────────────────────────────┘  │
│                          │                                        │
│                          ▼                                        │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              STEP 3-6: Add Test Suites                    │  │
│  └──────────────────────────────────────────────────────────┘  │
│                          │                                        │
│        ┌─────────────────┼─────────────────┐                    │
│        │                 │                 │                    │
│        ▼                 ▼                 ▼                    │
│  ┌──────────┐    ┌──────────────┐    ┌──────────────┐           │
│  │ Config   │    │ OS           │    │ Version     │           │
│  │ Validation│   │ Compatibility│    │ Enforcement │           │
│  │ Tests    │    │ Tests        │    │ Tests       │           │
│  └──────────┘    └──────────────┘    └──────────────┘           │
│        │                 │                 │                    │
│        └─────────────────┼─────────────────┘                    │
│                          │                                        │
│                          ▼                                        │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              Install Flow Tests                           │  │
│  └──────────────────────────────────────────────────────────┘  │
│                          │                                        │
└──────────────────────────┼────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                         OUTPUTS                                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    CODE CHANGES                          │  │
│  ├──────────────────────────────────────────────────────────┤  │
│  │ • tailscale.go: parseTailscaleVersion() helper           │  │
│  │ • tailscale.go: Updated EnsureInstalled() version logic  │  │
│  │ • tailscale_test.go: 4 new test suites (23 test cases)  │  │
│  └──────────────────────────────────────────────────────────┘  │
│                          │                                        │
│                          ▼                                        │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                  COVERAGE IMPROVEMENT                     │  │
│  ├──────────────────────────────────────────────────────────┤  │
│  │ • Before: 71.3%                                          │  │
│  │ • After:  ~78-80%                                        │  │
│  │ • Functions improved:                                    │  │
│  │   - EnsureInstalled(): 60% → ~85%                        │  │
│  │   - checkOSCompatibility(): 70% → ~90%                    │  │
│  │   - parseConfig(): 75% → ~90%                             │  │
│  │   - parseTailscaleVersion(): NEW, 100%                    │  │
│  └──────────────────────────────────────────────────────────┘  │
│                          │                                        │
│                          ▼                                        │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                DOCUMENTATION UPDATES                     │  │
│  ├──────────────────────────────────────────────────────────┤  │
│  │ • COVERAGE_STRATEGY.md: Updated coverage %                │  │
│  │ • PROVIDER_COVERAGE_STATUS.md: Updated status            │  │
│  │ • PROVIDER_NETWORK_TAILSCALE_COVERAGE_PLAN.md: Slice 2   │  │
│  └──────────────────────────────────────────────────────────┘  │
│                          │                                        │
└──────────────────────────┼────────────────────────────────────────┘
                           │
                           ▼
                    ┌──────────────┐
                    │   SLICE 3    │
                    │  (Future)    │
                    └──────────────┘
```

---

## Dependency Matrix

### Input Dependencies

| Dependency | Source | Type | Required For |
|------------|--------|------|--------------|
| `parseOSRelease()` | Slice 1 | Helper function | OS compatibility tests |
| `LocalCommander` | Existing | Test infrastructure | All Commander-based tests |
| `ErrInstallFailed` | Existing | Error type | Version enforcement, install flow tests |
| `ErrUnsupportedOS` | Existing | Error type | OS compatibility tests |
| `parseConfig()` | Existing | Config parsing | Config validation tests |
| `checkOSCompatibility()` | Existing | OS check | OS compatibility tests |
| `EnsureInstalled()` | Existing | Main function | All test suites |

### Output Dependencies

| Output | Type | Used By |
|--------|------|---------|
| `parseTailscaleVersion()` | Helper function | `EnsureInstalled()`, future tests |
| `TestEnsureInstalled_ConfigValidation` | Test suite | Coverage verification |
| `TestEnsureInstalled_OSCompatibility` | Test suite | Coverage verification |
| `TestEnsureInstalled_VersionEnforcement` | Test suite | Coverage verification |
| `TestEnsureInstalled_InstallFlow` | Test suite | Coverage verification |
| `TestParseTailscaleVersion` | Test suite | Coverage verification |

---

## Implementation Flow

### Phase 1: Foundation (Steps 1-2)

```
┌─────────────────┐
│  Slice 1 Helpers│
│  (parseOSRelease)│
└────────┬────────┘
         │
         ▼
┌─────────────────────────┐
│ parseTailscaleVersion() │  ← NEW helper
│ (Step 1)                │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│ EnsureInstalled()       │
│ version logic update    │  ← MODIFIED
│ (Step 2)                │
└────────┬────────────────┘
         │
         ▼
    [Phase 2]
```

### Phase 2: Test Implementation (Steps 3-6)

```
┌─────────────────────────┐
│ Config Validation Tests │  ← Step 3
│ (5 test cases)          │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│ OS Compatibility Tests   │  ← Step 4
│ (7 test cases)          │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│ Version Enforcement Tests│  ← Step 5
│ (7 test cases)          │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│ Install Flow Tests      │  ← Step 6
│ (4 test cases)          │
└────────┬────────────────┘
         │
         ▼
    [Phase 3]
```

### Phase 3: Verification & Documentation (Steps 7-9)

```
┌─────────────────────────┐
│ Coverage Verification  │  ← Step 7
│ (go test -cover)       │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│ Documentation Updates   │  ← Step 8
│ (3 files)               │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│ Commit                  │  ← Step 9
│ (git commit)            │
└─────────────────────────┘
```

---

## Test Dependency Graph

```
┌─────────────────────────────────────────────────────────────┐
│              Test Suite Dependencies                        │
└─────────────────────────────────────────────────────────────┘

TestEnsureInstalled_ConfigValidation
    │
    ├─► parseConfig() [existing]
    ├─► LocalCommander [existing]
    └─► ErrConfigInvalid [existing]

TestEnsureInstalled_OSCompatibility
    │
    ├─► checkOSCompatibility() [existing]
    ├─► parseOSRelease() [Slice 1]
    ├─► LocalCommander [existing]
    └─► ErrUnsupportedOS [existing]

TestEnsureInstalled_VersionEnforcement
    │
    ├─► EnsureInstalled() [modified in Step 2]
    ├─► parseTailscaleVersion() [NEW in Step 1]
    ├─► LocalCommander [existing]
    └─► ErrInstallFailed [existing]

TestEnsureInstalled_InstallFlow
    │
    ├─► EnsureInstalled() [modified in Step 2]
    ├─► checkOSCompatibility() [existing]
    ├─► parseOSRelease() [Slice 1]
    ├─► LocalCommander [existing]
    └─► ErrInstallFailed [existing]

TestParseTailscaleVersion
    │
    └─► parseTailscaleVersion() [NEW in Step 1]
        (pure function, no dependencies)
```

---

## Code Modification Impact

### Files Modified

| File | Changes | Impact |
|------|---------|--------|
| `tailscale.go` | Add `parseTailscaleVersion()` helper | New function, no breaking changes |
| `tailscale.go` | Update `EnsureInstalled()` version logic | Internal change, behavior improved per spec |
| `tailscale_test.go` | Add 5 new test suites | Test-only changes, no production impact |

### Breaking Changes

**None** - All changes are:
- Internal implementation improvements
- Test additions
- Spec-aligned behavior fixes (version parsing)

---

## Coverage Dependency Chain

```
┌─────────────────────────────────────────────────────────────┐
│         Coverage Improvement Dependency Chain                │
└─────────────────────────────────────────────────────────────┘

Current Coverage: 71.3%
    │
    ├─► parseTailscaleVersion() [NEW, 100% coverage]
    │       │
    │       └─► TestParseTailscaleVersion [7 test cases]
    │
    ├─► EnsureInstalled() [60% → ~85%]
    │       │
    │       ├─► TestEnsureInstalled_ConfigValidation [5 cases]
    │       ├─► TestEnsureInstalled_OSCompatibility [7 cases]
    │       ├─► TestEnsureInstalled_VersionEnforcement [7 cases]
    │       └─► TestEnsureInstalled_InstallFlow [4 cases]
    │
    ├─► checkOSCompatibility() [70% → ~90%]
    │       │
    │       └─► TestEnsureInstalled_OSCompatibility [exercises all branches]
    │
    └─► parseConfig() [75% → ~90%]
            │
            └─► TestEnsureInstalled_ConfigValidation [exercises error paths]

Expected Coverage: ~78-80%
```

---

## Critical Path

**Minimum viable implementation path** (if time-constrained):

1. ✅ `parseTailscaleVersion()` helper (required for spec compliance)
2. ✅ Update `EnsureInstalled()` version logic (required for spec compliance)
3. ✅ `TestParseTailscaleVersion` (validates helper)
4. ✅ `TestEnsureInstalled_VersionEnforcement` (validates spec compliance)
5. ✅ Coverage verification

**Full implementation path** (recommended):

All 9 steps from Agent document.

---

## Risk Assessment

### Low Risk Dependencies

| Dependency | Risk Level | Mitigation |
|------------|------------|------------|
| Slice 1 helpers | ✅ Low | Already implemented and tested |
| `LocalCommander` | ✅ Low | Existing, well-tested infrastructure |
| Error types | ✅ Low | Already defined and used |

### Medium Risk Dependencies

| Dependency | Risk Level | Mitigation |
|------------|------------|------------|
| Version parsing logic | ⚠️ Medium | Comprehensive test suite (7 cases) |
| Version comparison | ⚠️ Medium | Simple string comparison for v1 (acceptable) |

### No High Risk Dependencies

All dependencies are either:
- Already implemented (Slice 1)
- Well-defined (spec requirements)
- Testable in isolation (pure functions)

---

## Conclusion

**Dependency Status**: ✅ **All dependencies satisfied**

- Prerequisites from Slice 1: ✅ Complete
- Existing infrastructure: ✅ Available
- Spec requirements: ✅ Documented
- Test infrastructure: ✅ Ready

**Implementation Risk**: ✅ **Low**

All dependencies are low-risk, well-understood, and properly scoped.

**Ready for Implementation**: ✅ **Yes**

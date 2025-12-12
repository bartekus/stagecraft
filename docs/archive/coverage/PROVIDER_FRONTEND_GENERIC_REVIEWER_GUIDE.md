> **Superseded by** `docs/coverage/COVERAGE_LEDGER.md` section 5.2 (PROVIDER_FRONTEND_GENERIC). Kept for historical reference. New coverage snapshots and summaries MUST go into the coverage ledger.

# Reviewer Guide: Test Hardening PR

**Feature**: PROVIDER_FRONTEND_GENERIC  
**PR**: `feat/provider-frontend-generic-test-hardening`

---

## What to Review

### 1. Test Hardening Implementation

**File**: `internal/providers/frontend/generic/generic_test.go`

**Key Changes**:
- `devWithTimeout()` helper function (lines ~32-45)
- Timeout wrappers applied to tests with infinite loops or long-running operations
- Script modifications to use finite loops where appropriate

**Questions to Consider**:
- ✅ Are timeout values appropriate? (10-15s for shutdown tests, 3-5s for others)
- ✅ Does `devWithTimeout()` properly handle both success and timeout cases?
- ✅ Are all potentially problematic tests wrapped?

### 2. Test Seam Addition

**File**: `internal/providers/frontend/generic/generic.go`

**Key Changes**:
- `newScanner` variable (lines ~47-51)
- Updated scanner creation to use `newScanner()` instead of direct `bufio.NewScanner()`

**Questions to Consider**:
- ✅ Is the test seam minimal and well-documented?
- ✅ Does it preserve production behavior?
- ✅ Is it only used in tests?

### 3. Scanner Error Test

**File**: `internal/providers/frontend/generic/generic_test.go`

**Key Changes**:
- `TestGenericProvider_RunWithReadyPattern_ScannerError` (lines ~889-951)
- `errorAfterBytesReader` helper (lines ~953-976)

**Questions to Consider**:
- ✅ Does the test exercise the intended error path?
- ✅ Is the timeout (3s) appropriate?
- ✅ Is the flakiness documented?

### 4. Documentation

**File**: `internal/providers/frontend/generic/COVERAGE_STRATEGY.md`

**Key Changes**:
- "Known Test Debt" section documenting scanner error test flakiness
- Test hardening notes

**Questions to Consider**:
- ✅ Is the test debt clearly explained?
- ✅ Is the resolution path clear?
- ✅ Does it align with governance principles?

## What NOT to Review

- **Coverage numbers**: Already verified, Phase 2 targets exceeded
- **Test logic changes**: No behavior changes, only timeout wrappers
- **Production code changes**: Only test seam addition (documented as test-only)

## Quick Verification

Run these commands to verify:

```bash
# All tests pass
go test ./internal/providers/frontend/generic/... -timeout 90s

# Coverage maintained
go test -coverprofile=coverage.out ./internal/providers/frontend/generic/...
go tool cover -func=coverage.out | grep -E "(runWithReadyPattern|total)"

# No linter errors
golangci-lint run ./internal/providers/frontend/generic/...
```

## Review Focus Areas

### High Priority

1. **CI Safety**: Verify all tests have explicit timeouts
2. **Test Seam**: Ensure it's minimal and doesn't affect production behavior
3. **Documentation**: Verify test debt is clearly documented

### Medium Priority

1. **Timeout Values**: Are they appropriate for the test scenarios?
2. **Script Changes**: Are finite loops appropriate where used?

### Low Priority

1. **Coverage Numbers**: Already verified, but can double-check
2. **Code Style**: Should match existing patterns

## Approval Criteria

✅ All tests pass with timeouts  
✅ Test seam is minimal and documented  
✅ Test debt is explicitly documented  
✅ No production behavior changes  
✅ Coverage targets met  

## Questions or Concerns?

- **Flaky test**: This is documented as known debt with clear resolution path
- **Timeout values**: Can be adjusted if needed, but current values are conservative
- **Test seam**: Minimal and only used in tests, documented as test-only

## Related Issues

- Deflake scanner error test: Tracked in separate issue (to be created)

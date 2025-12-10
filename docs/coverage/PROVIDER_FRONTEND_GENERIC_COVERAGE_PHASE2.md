# PROVIDER_FRONTEND_GENERIC Coverage Phase 2 – Complete

**Feature**: PROVIDER_FRONTEND_GENERIC  
**Status**: ✅ Complete  
**Date**: 2025-01-XX

---

## Summary

Phase 2 successfully raised `runWithReadyPattern` coverage from **74.0%** to **92.0%** (target: ≥80%), exceeding the goal with a minimal, focused set of 3 tests.

**Results:**
- `runWithReadyPattern`: 74.0% → **92.0%** ✅
- Overall provider coverage: 80.2% → **87.7%** ✅
- Zero behavior changes (only minimal test seam added)

---

## Phase 2 Scope

**Objective:** Increase function-level coverage for `runWithReadyPattern` from 74.0% to ≥ 80.0% with a minimal set of focused tests.

**Scope (tiny but sharp):**

1. ✅ **Scanner error handling**
   - Added `TestGenericProvider_RunWithReadyPattern_ScannerError`
   - Forces `scanner.Err()` path via minimal test seam (`newScanner` variable)
   - Validates existing error handling behavior

2. ✅ **No-ready-pattern-before-exit edge case**
   - Added `TestGenericProvider_RunWithReadyPattern_ProcessExitsWithoutReadyPattern`
   - Tests process exiting successfully (exit code 0) without emitting ready pattern
   - Asserts current behavior matches spec

3. ✅ **Ready pattern on stderr**
   - Added `TestGenericProvider_RunWithReadyPattern_ReadyPatternOnStderr`
   - Ensures ready pattern detection works on stderr stream
   - Exercises distinct branch in scan/match logic

**Non-goals (Phase 2):**
- ✅ No changes to `Dev`, `runWithShutdown`, or `shutdownProcess` (already above targets)
- ✅ No refactors or behavior changes to the provider
- ✅ No new golden tests or output format guarantees
- ✅ No new docs beyond coverage strategy update

---

## Implementation Details

### Test Seam

Added minimal test seam in `generic.go`:

```go
// newScanner is a test seam for injecting scanner errors in tests.
// In production, this always returns bufio.NewScanner(reader).
// Tests can override this to inject failing readers.
var newScanner = func(reader interface{ Read([]byte) (int, error) }) *bufio.Scanner {
	return bufio.NewScanner(reader)
}
```

This allows tests to inject failing readers without changing production behavior.

### Tests Added

1. **`TestGenericProvider_RunWithReadyPattern_ScannerError`**
   - Uses test seam to inject `errorAfterBytesReader` that fails after reading some bytes
   - Exercises `scanner.Err()` error path in stdout/stderr monitoring goroutines
   - Validates error propagation and process cleanup

2. **`TestGenericProvider_RunWithReadyPattern_ProcessExitsWithoutReadyPattern`**
   - Creates script that exits successfully (exit 0) without emitting ready pattern
   - Exercises branch where process exits before ready pattern found (lines 220-225)
   - Validates error message matches expected behavior

3. **`TestGenericProvider_RunWithReadyPattern_ReadyPatternOnStderr`**
   - Creates script that outputs ready pattern only on stderr (not stdout)
   - Exercises stderr scanning branch (lines 174-191)
   - Validates ready pattern detection works on both streams

---

## Coverage Results

### Before Phase 2
```
stagecraft/internal/providers/frontend/generic/generic.go:120:	runWithReadyPattern	74.0%
Overall: 80.2% of statements
```

### After Phase 2
```
stagecraft/internal/providers/frontend/generic/generic.go:127:	runWithReadyPattern	92.0%
Overall: 87.7% of statements
```

**Improvement:** +18.0 percentage points for `runWithReadyPattern`, +7.5 percentage points overall.

---

## Test Execution

All tests pass:

```bash
$ go test ./internal/providers/frontend/generic/...
ok  	stagecraft/internal/providers/frontend/generic	8.383s	coverage: 87.7% of statements
```

Coverage verification:

```bash
$ go test -coverprofile=coverage.out ./internal/providers/frontend/generic/...
$ go tool cover -func=coverage.out | grep runWithReadyPattern
stagecraft/internal/providers/frontend/generic/generic.go:127:	runWithReadyPattern	92.0%
```

---

## Files Modified

1. **`internal/providers/frontend/generic/generic.go`**
   - Added `newScanner` test seam variable (lines 47-51)
   - Updated scanner creation to use `newScanner()` instead of direct `bufio.NewScanner()` (lines 155, 175)

2. **`internal/providers/frontend/generic/generic_test.go`**
   - Added `TestGenericProvider_RunWithReadyPattern_ScannerError` (lines 848-959)
   - Added `TestGenericProvider_RunWithReadyPattern_ProcessExitsWithoutReadyPattern` (lines 986-1020)
   - Added `TestGenericProvider_RunWithReadyPattern_ReadyPatternOnStderr` (lines 1022-1056)
   - Added `errorAfterBytesReader` test helper (lines 961-984)
   - Added imports: `bufio`, `fmt`

3. **`internal/providers/frontend/generic/COVERAGE_STRATEGY.md`**
   - Updated current status with Phase 2 results
   - Marked Phase 2 as complete
   - Updated missing coverage analysis

---

## Success Criteria

- ✅ `runWithReadyPattern` coverage ≥ 80.0% (achieved: **92.0%**)
- ✅ Overall provider coverage remains ≥ 80.0% (achieved: **87.7%**)
- ✅ `go test ./internal/providers/frontend/generic/...` passes
- ✅ Zero behavior changes (only test seam added)
- ✅ All tests deterministic and hermetic

---

## Next Steps

Phase 2 is complete. `runWithReadyPattern` now has excellent coverage (92.0%), exceeding all targets.

**Remaining optional work (not required):**
- Pipe creation error paths (would require complex test seams, low value)
- Additional edge cases (coverage already excellent)

**Recommendation:** Phase 2 complete. No further coverage work needed for `runWithReadyPattern`.

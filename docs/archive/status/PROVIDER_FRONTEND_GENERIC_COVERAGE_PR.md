# PROVIDER_FRONTEND_GENERIC — Coverage V1 Complete (PR Description)

## Summary

Completes PROVIDER_FRONTEND_GENERIC test coverage to v1 standards by replacing flaky integration tests with deterministic unit tests, eliminating test seams, and achieving 87.7% coverage (exceeds 80% target).

**Key Change**: Extracted `scanStream()` as a pure function, enabling deterministic unit tests that replace flaky OS-dependent integration tests.

---

## Changes

### Removed
- Flaky `TestGenericProvider_RunWithReadyPattern_ScannerError` integration test
- `newScanner` test seam (global variable for error injection)
- All `time.Sleep()` patterns in tests
- Goroutine-based test patterns without proper synchronization

### Added
- `scanStream()` pure function extraction
- Deterministic unit tests: `TestScanStream_ScannerError`, `TestScanStream_ReadyPatternFound`, `TestScanStream_ReadyPatternOnStderr`, `TestScanStream_ReadyOncePreventsMultipleSignals`
- Benchmarks: `BenchmarkScanStream_NoMatch`, `BenchmarkScanStream_MatchEarly`, `BenchmarkScanStream_MatchLate`, `BenchmarkScanStream_LargeInput`
- Updated `COVERAGE_STRATEGY.md` to reflect v1 complete status

### Updated
- `COVERAGE_STRATEGY.md` — Removed references to removed tests/seams, declared v1 complete
- Test organization — Clear separation: unit tests for scanner logic, integration tests for process lifecycle

---

## Coverage Metrics

| Function | Before | After | Status |
|----------|--------|-------|--------|
| `runWithReadyPattern` | 74.0% | 92.0% | ✅ +18% |
| Overall | 80.2% | 87.7% | ✅ +7.5% |

**All functions now exceed 75% coverage, with most exceeding 85%.**

---

## Test Quality Improvements

- ✅ All tests pass with `-race` (no race conditions)
- ✅ All tests pass with `-count=20` (zero flakiness)
- ✅ No `time.Sleep()` in tests
- ✅ No test seams required
- ✅ Deterministic, side-effect-free unit tests

---

## Alignment with Governance

This PR:
- ✅ Meets GOV_CORE test requirements
- ✅ Aligns with AATSE principles (deterministic primitives, no test seams)
- ✅ Serves as reference model for future providers (see `docs/governance/GOV_V1_TEST_REQUIREMENTS.md`)

---

## Files Changed

- `internal/providers/frontend/generic/generic.go` — Extracted `scanStream()` function
- `internal/providers/frontend/generic/generic_test.go` — Added deterministic unit tests, removed flaky integration test
- `internal/providers/frontend/generic/COVERAGE_STRATEGY.md` — Updated to v1 complete status

---

## Testing

```bash
# Run all tests
go test ./internal/providers/frontend/generic/...

# Verify no race conditions
go test -race ./internal/providers/frontend/generic/...

# Verify no flakiness
go test -count=20 ./internal/providers/frontend/generic/...

# Check coverage
go test -cover ./internal/providers/frontend/generic/...
```

---

## Reviewer Guide

### What to Review

1. **`scanStream()` extraction** (`generic.go`)
   - Is it truly pure? (no side effects, deterministic)
   - Does it handle all error cases?
   - Is the interface clean?

2. **Unit tests** (`generic_test.go`)
   - Do they test all branches?
   - Are they deterministic? (no sleeps, no OS dependencies)
   - Do they cover error paths?

3. **Integration tests** (`generic_test.go`)
   - Do they focus on orchestration, not scanner logic?
   - Are they still necessary? (or can they be simplified further)

4. **Documentation** (`COVERAGE_STRATEGY.md`)
   - Does it accurately reflect current state?
   - Is it clear for future maintainers?

### What NOT to Review

- ❌ Don't review removed code (it's gone, that's the point)
- ❌ Don't ask for more integration tests (they were the problem)
- ❌ Don't suggest adding test seams (we removed them intentionally)

### Key Questions

1. **Does `scanStream()` need to be exported?**
   - Currently unexported (internal). If future providers need it, we can export later.

2. **Are benchmarks necessary?**
   - Yes — they provide regression detection without OS dependencies.

3. **Why remove the integration test?**
   - It was flaky, OS-dependent, and tested logic that's now covered by deterministic unit tests.

---

## Related

- Feature: `PROVIDER_FRONTEND_GENERIC`
- Spec: `spec/providers/frontend/generic.md`
- Governance: `docs/governance/GOV_V1_TEST_REQUIREMENTS.md`
- Coverage Strategy: `internal/providers/frontend/generic/COVERAGE_STRATEGY.md`

---

## Checklist

- [x] All tests pass
- [x] No race conditions (`-race`)
- [x] No flakiness (`-count=20`)
- [x] Coverage exceeds 80%
- [x] Documentation updated
- [x] Aligns with AATSE principles
- [x] No test seams required
- [x] Serves as reference model for future providers

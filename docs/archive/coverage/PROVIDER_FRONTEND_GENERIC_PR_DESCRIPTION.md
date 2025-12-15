> **Superseded by** `docs/coverage/COVERAGE_LEDGER.md` section 5.2 (PROVIDER_FRONTEND_GENERIC). Kept for historical reference. New coverage snapshots and summaries MUST go into the coverage ledger.

# PR Description: Test Hardening and Scanner Error Test Debt Documentation

**Feature**: PROVIDER_FRONTEND_GENERIC  
**Related Governance Feature**: GOV_CORE  
**Branch**: `feat/provider-frontend-generic-test-hardening`

---

## Summary

This PR adds comprehensive test hardening to prevent CI hangs and explicitly documents known test debt for the scanner error test. All tests are now bounded by explicit timeouts, ensuring CI safety while maintaining excellent coverage (92% for `runWithReadyPattern`, 87.7% overall).

## Changes

### Test Hardening

- **Added `devWithTimeout()` helper** (`generic_test.go`)
  - Wraps `p.Dev()` calls with explicit timeouts
  - Prevents tests from hanging CI if there's a deadlock or blocking issue
  - Fails fast with clear timeout messages

- **Applied timeout wrappers** to all potentially problematic tests:
  - Tests with infinite loops: 10-15s timeouts (accounting for shutdown timeouts)
  - Long-running ready pattern tests: 5s timeout
  - Scanner error test: 3s timeout (bounded but potentially flaky)

- **Made scripts more finite** where appropriate:
  - Replaced `sleep 10` with finite loop in `TestGenericProvider_RunWithReadyPattern_ContextAfterReady`
  - Kept infinite loops for shutdown tests (intentional for testing shutdown behavior)

### Test Seam Addition

- **Added minimal `newScanner` test seam** (`generic.go`)
  - Allows injecting scanner errors in tests
  - Documented as test-only, behavior-preserving
  - Used by `TestGenericProvider_RunWithReadyPattern_ScannerError` to exercise `scanner.Err()` path

### Documentation

- **Added "Known Test Debt" section** to `COVERAGE_STRATEGY.md`
  - Documents the flaky scanner error test explicitly
  - Explains the tradeoff (coverage vs determinism)
  - Provides clear path to resolution (wrap real reader or extract helper)
  - Aligns with Stagecraft's "determinism over convenience" principle

- **Created Phase 2 completion documentation**
  - `PROVIDER_FRONTEND_GENERIC_COVERAGE_PHASE2.md` - Phase 2 achievements and results
  - `PROVIDER_FRONTEND_GENERIC_DEFLAKE_FOLLOWUP.md` - Post-deflake update guide

## Coverage Results

- **`runWithReadyPattern`**: 74.0% → **92.0%** ✅ (exceeds Phase 2 target of ≥80%)
- **Overall provider**: 80.2% → **87.7%** ✅ (exceeds Phase 2 target of ≥80%)

## Governance Alignment

✅ **Determinism**: Non-determinism explicitly documented as test debt with clear resolution path  
✅ **CI Safety**: All tests bounded by explicit timeouts, no hanging tests possible  
✅ **Coverage**: Phase 2 goals exceeded  
✅ **Documentation**: Complete coverage strategy with known debt tracking  

## Testing

All tests pass with explicit timeouts:

```bash
$ go test ./internal/providers/frontend/generic/... -timeout 90s
ok  	stagecraft/internal/providers/frontend/generic	12.390s
```

Coverage verification:

```bash
$ go test -coverprofile=coverage.out ./internal/providers/frontend/generic/...
$ go tool cover -func=coverage.out | grep runWithReadyPattern
stagecraft/internal/providers/frontend/generic/generic.go:127:	runWithReadyPattern	92.0%
```

## Known Limitations

The scanner error test (`TestGenericProvider_RunWithReadyPattern_ScannerError`) is documented as potentially flaky but bounded. This is tracked as explicit test debt in `COVERAGE_STRATEGY.md` with a clear path to resolution. See issue: `[PROVIDER_FRONTEND_GENERIC] Deflake scanner error test for runWithReadyPattern`.

## Related

- Phase 1 coverage work: Completed in previous PR
- Phase 2 coverage work: This PR
- Deflake follow-up: Tracked in separate issue

## Checklist

- [x] All tests pass with explicit timeouts
- [x] Coverage targets met (≥80% for `runWithReadyPattern`, ≥80% overall)
- [x] No unbounded tests remain
- [x] Test debt documented with resolution path
- [x] Documentation updated
- [x] Commit message follows governance format
- [x] Pre-commit hooks pass

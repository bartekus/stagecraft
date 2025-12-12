> **Superseded by** `docs/coverage/COVERAGE_LEDGER.md` section 5.2 (PROVIDER_FRONTEND_GENERIC). Kept for historical reference. New coverage snapshots and summaries MUST go into the coverage ledger.

# PROVIDER_FRONTEND_GENERIC Test Hardening - Summary

**Branch**: `feat/provider-frontend-generic-test-hardening`  
**Commit**: `03a5e55`  
**Status**: ✅ Committed and Pushed

---

## What Was Done

### 1. Test Hardening Implementation

- **Added `devWithTimeout()` helper** in `generic_test.go`
  - Wraps `p.Dev()` calls with explicit timeouts
  - Prevents tests from hanging CI
  - Fails fast with clear timeout messages

- **Applied timeout wrappers** to all potentially problematic tests:
  - `TestGenericProvider_Dev_ContextCancellation` - 10s timeout
  - `TestGenericProvider_Dev_DefaultShutdown` - 15s timeout (accounts for 10s default shutdown timeout)
  - `TestGenericProvider_RunWithReadyPattern_ContextAfterReady` - 5s timeout
  - All shutdown process tests - 10s timeout
  - `TestGenericProvider_RunWithReadyPattern_ScannerError` - 3s timeout

- **Made scripts more finite** where appropriate:
  - Replaced `sleep 10` with finite loop in `TestGenericProvider_RunWithReadyPattern_ContextAfterReady`
  - Kept infinite loops for shutdown tests (intentional for testing shutdown behavior)

### 2. Documentation Updates

- **Added "Known Test Debt" section** to `COVERAGE_STRATEGY.md`
  - Documents the flaky scanner error test explicitly
  - Explains the tradeoff (coverage vs determinism)
  - Provides clear path to resolution (wrap real reader or extract helper)

- **Created Phase 2 completion document**: `PROVIDER_FRONTEND_GENERIC_COVERAGE_PHASE2.md`
  - Documents Phase 2 achievements
  - Shows coverage improvements (74% → 92% for `runWithReadyPattern`)
  - Lists all tests added

- **Created follow-up documentation**: `PROVIDER_FRONTEND_GENERIC_DEFLAKE_FOLLOWUP.md`
  - Shows what `COVERAGE_STRATEGY.md` should look like after deflaking
  - Provides alternative text for Option A vs Option B approaches

### 3. Test Seam Addition

- **Added minimal test seam** in `generic.go`:
  - `newScanner` variable allows injecting scanner errors in tests
  - Documented as test-only, behavior-preserving
  - Used by `TestGenericProvider_RunWithReadyPattern_ScannerError`

---

## Coverage Results

- **`runWithReadyPattern`**: 74.0% → **92.0%** ✅
- **Overall provider**: 80.2% → **87.7%** ✅

---

## Governance Alignment

✅ **Determinism**: Non-determinism explicitly documented as test debt  
✅ **CI Safety**: All tests bounded by explicit timeouts  
✅ **Coverage**: Phase 2 goals exceeded  
✅ **Documentation**: Clear path to resolve flakiness  

---

## Next Steps

1. **Open PR** for `feat/provider-frontend-generic-test-hardening` branch
2. **Create GitHub issue** for deflaking scanner error test (see issue template below)
3. **After deflaking**: Update `COVERAGE_STRATEGY.md` using `PROVIDER_FRONTEND_GENERIC_DEFLAKE_FOLLOWUP.md` as reference

---

## GitHub Issue Template

**Title**: `[PROVIDER_FRONTEND_GENERIC] Deflake scanner error test for runWithReadyPattern`

**Body**: See `.github/ISSUE_TEMPLATE/deflake_scanner_error_test.md` (if not filtered) or use the template provided in the user's instructions.

---

## Files Changed

1. `internal/providers/frontend/generic/generic.go` - Added `newScanner` test seam
2. `internal/providers/frontend/generic/generic_test.go` - Added `devWithTimeout()` helper and applied timeouts
3. `internal/providers/frontend/generic/COVERAGE_STRATEGY.md` - Added "Known Test Debt" section
4. `docs/coverage/PROVIDER_FRONTEND_GENERIC_COVERAGE_PHASE2.md` - Phase 2 completion doc
5. `docs/coverage/PROVIDER_FRONTEND_GENERIC_DEFLAKE_FOLLOWUP.md` - Post-deflake update guide

---

## PR URL

Create PR at: https://github.com/bartekus/stagecraft/pull/new/feat/provider-frontend-generic-test-hardening

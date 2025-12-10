# Follow-up: COVERAGE_STRATEGY.md Update After Deflaking Scanner Error Test

This document shows the update to `COVERAGE_STRATEGY.md` that should be applied after the scanner error test is deflaked.

---

## Section to Replace

**Location**: After "Test Hardening (Post-Phase 2)" section, replace the "Known Test Debt" section with:

---

## Resolved Test Debt

### Scanner Error Test for `runWithReadyPattern` (Resolved)

`TestGenericProvider_RunWithReadyPattern_ScannerError` previously used the `newScanner` test seam in a way that could cause occasional timeouts due to interaction between the injected reader and the real process lifetime.

**Resolution**: The test seam was tightened to wrap the real pipe reader instead of replacing it entirely. This eliminates the timing-sensitive race condition while preserving full coverage of the `scanner.Err()` error path.

**Current status**:
- ✅ Test passes reliably without intermittent timeouts
- ✅ Coverage for `scanner.Err()` path maintained
- ✅ No changes to runtime behavior
- ✅ Deterministic test execution

The test now uses a deterministic `io.Reader` wrapper that passes through data from the real stdout/stderr pipes but returns an error after a controlled number of bytes, ensuring the scanner error path is exercised reliably.

---

## Alternative Text (If Using Option B - Extract Helper)

If the deflake work uses Option B (extracting the scan loop into a helper), use this instead:

---

## Resolved Test Debt

### Scanner Error Test for `runWithReadyPattern` (Resolved)

`TestGenericProvider_RunWithReadyPattern_ScannerError` previously used the `newScanner` test seam in a way that could cause occasional timeouts due to interaction between the injected reader and the real process lifetime.

**Resolution**: The scan loop was extracted into a pure helper function (`scanUntilReady`) that can be tested in isolation with deterministic failing readers. This eliminates the timing/race domain entirely while preserving full coverage of the `scanner.Err()` error path.

**Current status**:
- ✅ Test passes reliably without intermittent timeouts
- ✅ Coverage for `scanner.Err()` path maintained
- ✅ No changes to runtime behavior
- ✅ Deterministic test execution
- ✅ Improved testability through isolated helper

The helper function is tested directly with controlled readers, and `runWithReadyPattern` calls the helper, ensuring the scanner error path is exercised reliably without process lifecycle timing dependencies.

---

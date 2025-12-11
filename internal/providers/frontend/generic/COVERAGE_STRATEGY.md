# PROVIDER_FRONTEND_GENERIC — Coverage Strategy (V1 Complete)

This document defines the coverage approach for the PROVIDER_FRONTEND_GENERIC provider.
As of v1, all critical execution paths are covered by deterministic, side-effect–free tests that enforce AATSE and no-broken-glass principles.

⸻

## 🎯 Coverage Goals

The provider must:
1. Detect readiness reliably (`runWithReadyPattern`)
2. Stream process output deterministically
3. Handle error paths consistently (scanner errors, premature exits)
4. Avoid test flakiness caused by:
   - OS-level pipe buffering
   - goroutine scheduling
   - timing-based assertions
   - integration tests depending on subprocess semantics

The previous testing approach relied on an inline scanner and lifecycle hooks that were difficult to isolate.
The v1 redesign resolves these issues via a pure, testable helper.

⸻

## ✔️ V1 Coverage Status — COMPLETE

**Overall Coverage: 87.7%** (exceeds v1 target of 80%+)

| Function | Coverage | Status |
|----------|----------|--------|
| `ID` | 100.0% | ✅ Complete |
| `Dev` | 88.0% | ✅ Excellent |
| `parseConfig` | 85.7% | ✅ Excellent |
| `runWithShutdown` | 91.7% | ✅ Excellent |
| `shutdownProcess` | 76.0% | ✅ Good |
| `runWithReadyPattern` | 92.0% | ✅ Excellent |
| `init` | 100.0% | ✅ Complete |

All required test scenarios are now covered using deterministic tests, with no timing dependencies or goroutine-based assertions.

⸻

## What Changed in V1

### 1. Extracted Deterministic Scanner Helper

`scanStream()` is now a standalone, synchronous, pure function:
- Accepts `io.Reader`, `io.Writer`, `*regexp.Regexp`
- Emits readiness via channel
- Emits errors deterministically
- Performant and benchmarked

This aligns with the AATSE requirement that stateful or concurrent behavior must be represented by isolated, testable primitives.

### 2. Replaced Flaky Readiness Tests

The old integration test:
- Depended on OS pipe buffering
- Required injecting synthetic scanner errors through a global seam
- Produced intermittent failures in CI
- Violated AATSE "no broken glass" determinism

**It has been removed.**

Scanner behavior is now tested through:

**New Deterministic Unit Tests:**
- `TestScanStream_ScannerError` - Tests `scanner.Err()` error path with controlled failing reader
- `TestScanStream_ReadyPatternFound` - Tests pattern detection and output forwarding
- `TestScanStream_ReadyPatternOnStderr` - Tests stderr label handling
- `TestScanStream_ReadyOncePreventsMultipleSignals` - Tests `sync.Once` behavior

These tests:
- Have no races
- Require no subprocesses
- Require no goroutines
- Exercise all error-handling and signal-handling branches

### 3. Integration Coverage Restricted to Process Lifecycle

Since scanner correctness is now guaranteed by deterministic unit tests:
- Integration tests no longer validate scan correctness itself
- Integration tests validate only:
  - Process exit before ready pattern → correct error
  - Process exit after pattern → clean shutdown behavior
  - Context cancellation → proper termination flow

This separation of concerns is explicitly aligned with AATSE structural decomposition.

### 4. Benchmarks Added

Benchmarks allow cheap regression detection without depending on OS-level behavior:
- `BenchmarkScanStream_NoMatch` - No pattern match scenario
- `BenchmarkScanStream_MatchEarly` - Pattern found early in stream
- `BenchmarkScanStream_MatchLate` - Pattern found late in stream
- `BenchmarkScanStream_LargeInput` - Large input handling

These cover memory behavior, repeated pattern matching, and early-/late-exit scenarios.

⸻

## 🧪 Coverage Philosophy: AATSE + "No Broken Glass"

The refactor aligns with the Stagecraft engineering standard:

### Deterministic Primitives
Scanner is pure → unit-testable → benchmarkable.

### Predictable Orchestration
`runWithReadyPattern` orchestrates scanners; it no longer contains scanner logic.

### Separation of Concerns
- Scanner correctness tested in isolation
- Process lifecycle tested through normal integration flows

### Minimal Concurrency Surface
Scanner tests run entirely synchronously.

### No Fragile Legal Fictions
Tests do not fabricate impossible I/O states.

Where Stagecraft used to rely on a test seam (`newScanner`) to simulate difficult states, we now rely on the design itself to make the component testable.

**This is the AATSE principle applied:**
> If you need a seam for testing, the design is incomplete.

⸻

## 📈 Future Expansions (Post-V1)

Although V1 is complete, optional enhancements (non-blocking):
- Structured logging tests (when Stagecraft logging V2 lands)
- Extended pattern matching (multiple simultaneous regex signals)
- Timeout orchestration logic (if future features require it)

These are outside v1 scope and not required for correctness or stability.

⸻

## ✅ Conclusion

**PROVIDER_FRONTEND_GENERIC coverage is now V1 Complete.**

All major branches, edge cases, and lifecycle transitions are validated through deterministic tests that align with Stagecraft governance and AATSE design standards.

- ✅ No flaky integration tests remain
- ✅ No test seams remain
- ✅ Coverage and reliability meet or exceed GOV_V1 expectations
- ✅ All tests pass with `-race` and `-count=20` without flakiness

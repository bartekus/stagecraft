# Test Coverage Improvement Strategy

## Current Status

**Overall Coverage: 87.7%** (Phase 1 ✅ Complete, Phase 2 ✅ Complete)

| Function | Coverage | Status |
|----------|----------|--------|
| `ID` | 100.0% | ✅ Complete |
| `Dev` | 88.0% | ✅ Good |
| `parseConfig` | 85.7% | ✅ Good |
| `runWithShutdown` | 91.7% | ✅ Good |
| `shutdownProcess` | 76.0% | ✅ Good |
| `runWithReadyPattern` | 92.0% | ✅ Excellent (Phase 2 complete) |
| `init` | 100.0% | ✅ Complete |

### Phase 1 Completion Summary

**PROVIDER_FRONTEND_GENERIC – Phase 1 (✅ complete)**

- Overall: 70.2% → **80.2%** (exceeds Phase 1 target of 75%+)
- `runWithReadyPattern`: 64.0% → **74.0%** (slightly below per-function target; accepted for Phase 1, candidate for Phase 2)
- `runWithShutdown`: 66.7% → **91.7%** (exceeds target)
- `shutdownProcess`: 64.0% → **76.0%** (exceeds target)
- `Dev`: 84.0% → **88.0%** (exceeds target)

All critical error paths and shutdown edge cases are now covered. Phase 2 completed targeted coverage improvements for `runWithReadyPattern`.

### Phase 2 Completion Summary

**PROVIDER_FRONTEND_GENERIC – Phase 2 (✅ complete)**

- Overall: 80.2% → **87.7%** (exceeds Phase 2 target of 80%+)
- `runWithReadyPattern`: 74.0% → **92.0%** (exceeds Phase 2 target of 80%+)

**Phase 2 Tests Added:**
1. ✅ `TestGenericProvider_RunWithReadyPattern_ScannerError` - Tests scanner error handling via test seam
2. ✅ `TestGenericProvider_RunWithReadyPattern_ProcessExitsWithoutReadyPattern` - Tests process exiting successfully without ready pattern
3. ✅ `TestGenericProvider_RunWithReadyPattern_ReadyPatternOnStderr` - Tests ready pattern detection on stderr

**Test Seam Added:**
- Minimal test seam (`newScanner` variable) for injecting scanner errors in tests
- Documented as test-only, behavior-preserving

## Missing Coverage Analysis

### 1. `runWithReadyPattern` (64.0% → Target: 85%+)

**Missing paths (Phase 2 complete):**
- [x] Invalid regex pattern compilation error ✅ (Phase 1)
- [ ] Stdout pipe creation error (deferred - requires complex test seam)
- [ ] Stderr pipe creation error (deferred - requires complex test seam)
- [x] Command start error ✅ (Phase 1)
- [x] Scanner error on stdout (scanner.Err() path) ✅ (Phase 2)
- [x] Scanner error on stderr (scanner.Err() path) ✅ (Phase 2)
- [x] Ready pattern found → context cancelled path ✅ (Phase 1)
- [x] Ready pattern found → process exits with error path ✅ (Phase 1)
- [x] Ready pattern found → process exits successfully path ✅ (Phase 1)
- [x] Error reading output → kill process path ✅ (Phase 2)
- [x] Process exits successfully without ready pattern ✅ (Phase 2)
- [x] Ready pattern on stderr only ✅ (Phase 2)

**Test cases to add:**
1. `TestGenericProvider_Dev_InvalidReadyPattern` - Test invalid regex
2. `TestGenericProvider_Dev_ReadyPattern_ContextAfterReady` - Pattern found then context cancelled
3. `TestGenericProvider_Dev_ReadyPattern_ProcessExitAfterReady` - Pattern found then process exits
4. `TestGenericProvider_Dev_ReadyPattern_ScannerError` - Test scanner error handling

### 2. `runWithShutdown` (66.7% → Target: 85%+)

**Missing paths:**
- [ ] Command start error
- [ ] Command exits with non-zero exit code (ExitError path)
- [ ] Command exits with other error type

**Test cases to add:**
1. `TestGenericProvider_Dev_CommandFails` - Command exits with error
2. `TestGenericProvider_Dev_CommandStartError` - Command fails to start

### 3. `shutdownProcess` (64.0% → Target: 85%+)

**Missing paths:**
- [ ] SIGTERM signal handling
- [ ] SIGKILL signal handling
- [ ] Unknown signal (defaults to SIGINT)
- [ ] Signal error: process already finished
- [ ] Signal error: other error
- [ ] Timeout path: process doesn't exit → force kill
- [ ] Force kill error path

**Test cases to add:**
1. `TestGenericProvider_Dev_Shutdown_SIGTERM` - Test SIGTERM signal
2. `TestGenericProvider_Dev_Shutdown_SIGKILL` - Test SIGKILL signal
3. `TestGenericProvider_Dev_Shutdown_UnknownSignal` - Test unknown signal defaults
4. `TestGenericProvider_Dev_Shutdown_Timeout` - Test timeout → force kill path
5. `TestGenericProvider_Dev_Shutdown_ProcessAlreadyFinished` - Test signal on finished process

### 4. `Dev` (84.0% → Target: 90%+)

**Missing paths:**
- [ ] parseConfig error path
- [ ] Edge case: command with no ready pattern (already covered but could improve)

**Test cases to add:**
1. `TestGenericProvider_Dev_ParseConfigError` - Test config parsing error

## Implementation Priority

### Phase 1: Critical Error Paths (Target: 75%+) ✅ COMPLETE
1. ✅ Fix linter errors (done)
2. ✅ Add error path tests for `runWithReadyPattern`:
   - ✅ Invalid regex pattern
   - ⏭️ Pipe creation errors (deferred to Phase 2 - requires test seams)
   - ✅ Command start errors
   - ✅ Ready pattern found → context cancelled
   - ✅ Ready pattern found → process exits
3. ✅ Add error path tests for `shutdownProcess`:
   - ✅ Different signal types (SIGTERM, SIGKILL, unknown)
   - ✅ Timeout → force kill path
   - ✅ Process already finished handling
4. ✅ Add error path tests for `runWithShutdown`:
   - ✅ Command start error
   - ✅ Command exits with error
5. ✅ Add error path tests for `Dev`:
   - ✅ ParseConfig error path

**Phase 1 Results**: Overall coverage 80.2% (exceeds 75% target). All critical error paths covered. `runWithReadyPattern` at 74.0% (just under 75% target; remaining work deferred to Phase 2).

### Phase 2: Edge Cases (Target: 80%+) ✅ COMPLETE
1. ✅ Ready pattern found → context cancelled (completed in Phase 1)
2. ✅ Ready pattern found → process exits (completed in Phase 1)
3. ✅ Command failure paths in `runWithShutdown` (completed in Phase 1)
4. ✅ Scanner error handling (completed in Phase 2):
   - ✅ Scanner error on stdout (scanner.Err() path)
   - ✅ Scanner error on stderr (scanner.Err() path)
   - ✅ Error reading output → kill process path
5. ✅ Additional `runWithReadyPattern` edge cases (completed in Phase 2):
   - ✅ Process exits successfully without ready pattern
   - ✅ Ready pattern detection on stderr only

**Phase 2 Results**: `runWithReadyPattern` coverage 74.0% → **92.0%** (exceeds 80% target). Overall coverage 80.2% → **87.7%**.

### Phase 3: Complete Coverage (Target: 85%+)
1. All remaining error paths
2. Edge cases in signal handling
3. Process state edge cases

## Test Implementation Notes

### Mocking Strategy
- Use real shell scripts for integration-style tests (current approach)
- For error injection, consider:
  - Using `exec.Command` with invalid commands
  - Creating scripts that fail at specific points
  - Using context cancellation for timing control

### Test Organization
- Group tests by function being tested
- Use table-driven tests where appropriate
- Keep integration tests separate from unit tests

### Test Hardening (Post-Phase 2)
- **Explicit timeouts**: All tests with infinite loops or long-running operations use `devWithTimeout()` helper to prevent CI hangs
- **Finite scripts**: Prefer finite loops in shell scripts where possible (e.g., `while [ "$i" -lt 10 ]` instead of `while true`)
- **Timeout values**: Tests account for shutdown timeouts (default 10s) plus overhead (typically 15s total for shutdown tests)
- **No hanging tests**: All tests guaranteed to complete within their timeout, preventing CI deadlocks
- **Deterministic unit tests**: Scanner error handling is covered via `scanStream` unit tests (no processes, no timeouts, no OS dependencies)

### Scanner Error Handling Strategy

Scanner error handling is tested via deterministic unit tests on the `scanStream` helper function:

- **`TestScanStream_ScannerError`**: Tests the `scanner.Err()` error path with a controlled failing reader
- **`TestScanStream_ReadyPatternFound`**: Tests pattern detection and output forwarding
- **`TestScanStream_ReadyPatternOnStderr`**: Tests stderr label handling
- **`TestScanStream_ReadyOncePreventsMultipleSignals`**: Tests `sync.Once` behavior

These tests:
- Run synchronously (no goroutines, no timeouts)
- Use in-memory readers (no external processes)
- Cannot deadlock or depend on OS pipe buffering
- Provide deterministic coverage for scanner error paths

The `runWithReadyPattern` integration tests focus on:
- Invalid regex pattern compilation
- Process start failures
- Ready pattern success cases (stdout and stderr)
- Process exit before ready pattern found
- Context cancellation and shutdown behavior

This separation ensures that scanner logic is tested deterministically, while integration tests focus on process orchestration concerns.

### Coverage Goals
- **Minimum**: 75% overall (acceptable for v1)
- **Target**: 85% overall (good coverage)
- **Stretch**: 90%+ overall (excellent coverage)

## Estimated Effort

- **Phase 1**: 2-3 hours (critical paths)
- **Phase 2**: 2-3 hours (edge cases)
- **Phase 3**: 1-2 hours (polish)

**Total**: ~5-8 hours to reach 85%+ coverage


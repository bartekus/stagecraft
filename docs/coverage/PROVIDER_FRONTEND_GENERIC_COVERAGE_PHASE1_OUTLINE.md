# PROVIDER_FRONTEND_GENERIC Coverage Phase 1 Implementation Outline

> This document defines the Phase 1 implementation plan for test coverage improvement for PROVIDER_FRONTEND_GENERIC. It translates the coverage analysis brief into a concrete, testable delivery plan.

> All details in this outline must align with `spec/providers/frontend/generic.md` and `internal/providers/frontend/generic/COVERAGE_STRATEGY.md`.

⸻

## 1. Feature Summary

**Feature ID:** PROVIDER_FRONTEND_GENERIC

**Domain:** providers

**Goal:**

Improve test coverage for PROVIDER_FRONTEND_GENERIC from 70.2% to 75%+ by adding comprehensive tests for critical error paths and edge cases, ensuring governance compliance and operational reliability.

**Phase 1 Scope:**

- Add tests for critical error paths in `runWithReadyPattern`
- Add tests for error paths in `runWithShutdown`
- Add tests for signal handling edge cases in `shutdownProcess`
- Improve coverage for `Dev` error paths
- Achieve 75%+ overall coverage (minimum acceptable threshold)
- Maintain deterministic, hermetic test execution
- Align all tests with spec-defined behaviors

**Out of scope for Phase 1:**

- Behavior changes or refactoring (tests only)
- Coverage beyond 75% (targeted for Phase 2/3)
- Golden tests (not required for error paths)
- Performance tests
- Integration with CLI_DEV (focus on provider tests)

⸻

## 2. Test Implementation Plan

### 2.1 Test Files Structure

```
internal/providers/frontend/generic/
  ├── generic.go                    # Implementation (no changes)
  ├── generic_test.go              # Existing tests (extend)
  └── testdata/                    # Test fixtures (if needed)
      └── scripts/                 # Shell scripts for integration tests
          ├── test_ready.sh        # Existing
          ├── test_no_ready.sh     # Existing
          ├── test_long.sh         # Existing
          ├── test_invalid_cmd.sh  # New: invalid command for error testing
          └── test_exit_error.sh   # New: command that exits with error
```

### 2.2 Test Organization

Tests will be organized by function being tested:

- `TestGenericProvider_Dev_*` - Tests for `Dev()` method
- `TestGenericProvider_RunWithReadyPattern_*` - Tests for `runWithReadyPattern()`
- `TestGenericProvider_RunWithShutdown_*` - Tests for `runWithShutdown()`
- `TestGenericProvider_ShutdownProcess_*` - Tests for `shutdownProcess()`
- `TestGenericProvider_ParseConfig_*` - Tests for `parseConfig()` (existing)

⸻

## 3. Test Cases to Implement

### 3.1 `runWithReadyPattern` Error Paths (64.0% → 75%+)

#### Test: Invalid Regex Pattern
- **Name**: `TestGenericProvider_RunWithReadyPattern_InvalidRegex`
- **Purpose**: Test error handling when ready_pattern is invalid regex
- **Steps**:
  1. Create provider with invalid regex pattern (e.g., `[invalid`)
  2. Call `Dev()` with valid command
  3. Verify error returned with clear message about invalid regex
- **Expected**: Error returned, no process started

#### Test: Pipe Creation Errors
- **Name**: `TestGenericProvider_RunWithReadyPattern_PipeError`
- **Purpose**: Test error handling when stdout/stderr pipe creation fails
- **Note**: This may require test seam (injecting pipe creation failure)
- **Steps**:
  1. If seam needed, document it
  2. Simulate pipe creation failure
  3. Verify error returned
- **Expected**: Error returned with clear message

#### Test: Command Start Error
- **Name**: `TestGenericProvider_RunWithReadyPattern_CommandStartError`
- **Purpose**: Test error handling when command fails to start
- **Steps**:
  1. Create provider with invalid command (e.g., `/nonexistent/command`)
  2. Call `Dev()` with ready_pattern
  3. Verify error returned
- **Expected**: Error returned with clear message about command start failure

#### Test: Scanner Error Handling
- **Name**: `TestGenericProvider_RunWithReadyPattern_ScannerError`
- **Purpose**: Test error handling when scanner encounters errors
- **Note**: May require test seam or script that causes scanner errors
- **Steps**:
  1. Create script that causes scanner errors (if possible)
  2. Call `Dev()` with ready_pattern
  3. Verify error handling
- **Expected**: Error returned, process killed

#### Test: Ready Pattern Found → Context Cancelled
- **Name**: `TestGenericProvider_RunWithReadyPattern_ContextAfterReady`
- **Purpose**: Test graceful shutdown after ready pattern found
- **Steps**:
  1. Create script that outputs ready pattern
  2. Call `Dev()` with ready_pattern and short timeout
  3. Cancel context after pattern found
  4. Verify graceful shutdown
- **Expected**: Process shutdown gracefully

#### Test: Ready Pattern Found → Process Exits
- **Name**: `TestGenericProvider_RunWithReadyPattern_ProcessExitAfterReady`
- **Purpose**: Test behavior when process exits after ready pattern found
- **Steps**:
  1. Create script that outputs ready pattern then exits
  2. Call `Dev()` with ready_pattern
  3. Verify no error returned (normal exit)
- **Expected**: No error, normal completion

### 3.2 `runWithShutdown` Error Paths (66.7% → 75%+)

#### Test: Command Start Error
- **Name**: `TestGenericProvider_RunWithShutdown_CommandStartError`
- **Purpose**: Test error handling when command fails to start
- **Steps**:
  1. Create provider with invalid command
  2. Call `Dev()` without ready_pattern
  3. Verify error returned
- **Expected**: Error returned with clear message

#### Test: Command Exits with Error
- **Name**: `TestGenericProvider_RunWithShutdown_CommandExitsWithError`
- **Purpose**: Test error handling when command exits with non-zero exit code
- **Steps**:
  1. Create script that exits with error code
  2. Call `Dev()` without ready_pattern
  3. Verify error returned with exit code
- **Expected**: Error returned with exit code information

#### Test: Command Exits with Other Error
- **Name**: `TestGenericProvider_RunWithShutdown_CommandExitsWithOtherError`
- **Purpose**: Test error handling for non-ExitError errors
- **Note**: May require test seam to inject different error types
- **Steps**:
  1. If possible, simulate non-ExitError error
  2. Verify error handling
- **Expected**: Error returned with clear message

### 3.3 `shutdownProcess` Edge Cases (64.0% → 75%+)

#### Test: SIGTERM Signal Handling
- **Name**: `TestGenericProvider_ShutdownProcess_SIGTERM`
- **Purpose**: Test shutdown with SIGTERM signal
- **Steps**:
  1. Create long-running script
  2. Call `Dev()` with shutdown.signal: "SIGTERM"
  3. Cancel context
  4. Verify SIGTERM sent and process exits
- **Expected**: Process receives SIGTERM and exits gracefully

#### Test: SIGKILL Signal Handling
- **Name**: `TestGenericProvider_ShutdownProcess_SIGKILL`
- **Purpose**: Test shutdown with SIGKILL signal
- **Steps**:
  1. Create long-running script
  2. Call `Dev()` with shutdown.signal: "SIGKILL"
  3. Cancel context
  4. Verify SIGKILL sent and process exits immediately
- **Expected**: Process killed immediately

#### Test: Unknown Signal Defaults to SIGINT
- **Name**: `TestGenericProvider_ShutdownProcess_UnknownSignal`
- **Purpose**: Test that unknown signals default to SIGINT
- **Steps**:
  1. Create long-running script
  2. Call `Dev()` with shutdown.signal: "INVALID_SIGNAL"
  3. Cancel context
  4. Verify SIGINT sent (default behavior)
- **Expected**: SIGINT sent (default), process exits

#### Test: Process Already Finished
- **Name**: `TestGenericProvider_ShutdownProcess_ProcessAlreadyFinished`
- **Purpose**: Test graceful handling when process already finished
- **Steps**:
  1. Create script that exits quickly
  2. Call `Dev()` and wait for process to finish
  3. Attempt shutdown (should handle gracefully)
  4. Verify no error returned
- **Expected**: No error, graceful handling

#### Test: Timeout → Force Kill
- **Name**: `TestGenericProvider_ShutdownProcess_TimeoutForceKill`
- **Purpose**: Test timeout path that leads to force kill
- **Steps**:
  1. Create script that ignores signals (trap '' SIGINT SIGTERM)
  2. Call `Dev()` with short timeout (e.g., 100ms)
  3. Cancel context
  4. Verify force kill after timeout
- **Expected**: Process force killed after timeout, error returned

#### Test: Force Kill Error
- **Name**: `TestGenericProvider_ShutdownProcess_ForceKillError`
- **Purpose**: Test error handling when force kill fails
- **Note**: May require test seam to simulate kill failure
- **Steps**:
  1. If possible, simulate kill failure
  2. Verify error handling
- **Expected**: Error returned with clear message

### 3.4 `Dev` Error Paths (84.0% → 85%+)

#### Test: ParseConfig Error Path
- **Name**: `TestGenericProvider_Dev_ParseConfigError`
- **Purpose**: Test error handling when config parsing fails
- **Steps**:
  1. Call `Dev()` with invalid config structure
  2. Verify error returned with clear message
- **Expected**: Error returned, no process started

⸻

## 4. Test Implementation Strategy

### 4.1 TDD Workflow

1. **Write failing tests first** (TDD discipline)
2. **Run tests** - verify they fail for the right reasons
3. **Verify tests pass** with current implementation (no behavior changes)
4. **Document any seams** if created for testing

### 4.2 Test Fixtures

- Use shell scripts for integration-style tests (existing pattern)
- Create scripts that:
  - Exit with specific error codes
  - Ignore signals (for timeout testing)
  - Output ready patterns at specific times
  - Cause scanner errors (if possible)

### 4.3 Determinism Requirements

- No timestamps in test output
- No random values
- Use context cancellation for timing control
- Use deterministic delays in test scripts
- All tests must be hermetic (isolated, no shared state)

### 4.4 Test Seams (if needed)

If test seams are required to inject errors:

1. **Document the seam** in test comments
2. **Justify the seam** (why it's needed)
3. **Keep seams minimal** (only what's necessary)
4. **Ensure seams don't affect production code** (test-only)

Example seam pattern (if needed):
```go
// Test seam: allows injecting pipe creation errors for testing
// This is test-only and does not affect production behavior
type pipeCreator interface {
    StdoutPipe() (io.ReadCloser, error)
    StderrPipe() (io.ReadCloser, error)
}
```

⸻

## 5. Coverage Targets

### Overall Coverage
- **Current**: 70.2%
- **Target**: 75%+ (minimum acceptable threshold)
- **Stretch**: 80%+ (Phase 2 target)

### Function-Level Targets

| Function | Current | Phase 1 Target | Status |
|----------|---------|----------------|--------|
| `ID` | 100.0% | 100% | ✅ Complete |
| `Dev` | 84.0% | 85%+ | 🟡 Minor improvement |
| `parseConfig` | 85.7% | 85%+ | ✅ Complete |
| `runWithShutdown` | 66.7% | 75%+ | 🟠 Needs improvement |
| `shutdownProcess` | 64.0% | 75%+ | 🟠 Needs improvement |
| `runWithReadyPattern` | 64.0% | 75%+ | 🟠 Needs improvement |
| `init` | 100.0% | 100% | ✅ Complete |

⸻

## 6. Testing Strategy

### Unit Tests

- Test individual functions in isolation
- Use table-driven tests where appropriate
- Mock external dependencies if needed (minimal)

### Integration Tests

- Use real shell scripts (existing pattern)
- Test end-to-end behavior
- Verify process lifecycle management

### Error Injection

- Use invalid commands for start errors
- Use scripts that exit with errors
- Use context cancellation for timing control
- Use scripts that ignore signals for timeout testing

⸻

## 7. Implementation Phases

### Phase 1.1: Setup and Analysis
1. Review existing test structure
2. Identify test gaps from coverage report
3. Plan test fixtures needed
4. Document any seams required

### Phase 1.2: Error Path Tests
1. Implement `runWithReadyPattern` error tests
2. Implement `runWithShutdown` error tests
3. Implement `shutdownProcess` edge case tests
4. Implement `Dev` error path tests

### Phase 1.3: Verification
1. Run coverage report: `go test -cover ./internal/providers/frontend/generic/...`
2. Verify 75%+ coverage achieved
3. Run all tests: `go test ./internal/providers/frontend/generic/...`
4. Verify CI checks pass: `./scripts/check-coverage.sh --fail-on-warning`

⸻

## 8. Success Metrics

### Coverage Metrics
- ✅ Overall coverage: 75%+ (from 70.2%)
- ✅ `runWithReadyPattern`: 75%+ (from 64.0%)
- ✅ `runWithShutdown`: 75%+ (from 66.7%)
- ✅ `shutdownProcess`: 75%+ (from 64.0%)
- ✅ `Dev`: 85%+ (from 84.0%)

### Quality Metrics
- ✅ All tests pass
- ✅ Tests are deterministic
- ✅ Tests are hermetic
- ✅ Tests align with spec
- ✅ No behavior changes
- ✅ CI checks pass

### Governance Metrics
- ✅ Coverage check passes with `--fail-on-warning`
- ✅ Tests follow TDD discipline
- ✅ Tests follow spec-first approach
- ✅ No linter errors

⸻

## 9. Dependencies

### Internal Dependencies
- `PROVIDER_FRONTEND_GENERIC`: Implementation must exist (✅ done)
- `GOV_V1_CORE`: Coverage governance framework (✅ done)
- `CORE_EXECUTIL`: Process execution utilities (✅ done)

### External Dependencies
- Go testing framework (standard library)
- Shell scripts for integration tests

### No New Dependencies
Phase 1 should not introduce new external dependencies.

⸻

## 10. Approval

This implementation outline must be approved before proceeding to test implementation.

**Next Steps:**
1. Review and approve this outline
2. Verify spec alignment (`spec/providers/frontend/generic.md`)
3. Implement tests following TDD workflow
4. Verify coverage thresholds met
5. Run CI checks with `--fail-on-warning`

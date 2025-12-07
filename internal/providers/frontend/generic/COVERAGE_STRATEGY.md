# Test Coverage Improvement Strategy

## Current Status

**Overall Coverage: 70.2%**

| Function | Coverage | Status |
|----------|----------|--------|
| `ID` | 100.0% | ✅ Complete |
| `Dev` | 84.0% | 🟡 Good |
| `parseConfig` | 85.7% | 🟡 Good |
| `runWithShutdown` | 66.7% | 🟠 Needs improvement |
| `shutdownProcess` | 64.0% | 🟠 Needs improvement |
| `runWithReadyPattern` | 64.0% | 🟠 Needs improvement |
| `init` | 100.0% | ✅ Complete |

## Missing Coverage Analysis

### 1. `runWithReadyPattern` (64.0% → Target: 85%+)

**Missing paths:**
- [ ] Invalid regex pattern compilation error
- [ ] Stdout pipe creation error
- [ ] Stderr pipe creation error
- [ ] Command start error
- [ ] Scanner error on stdout (scanner.Err() path)
- [ ] Scanner error on stderr (scanner.Err() path)
- [ ] Ready pattern found → context cancelled path
- [ ] Ready pattern found → process exits with error path
- [ ] Ready pattern found → process exits successfully path
- [ ] Error reading output → kill process path

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

### Phase 1: Critical Error Paths (Target: 75%+)
1. ✅ Fix linter errors (done)
2. Add error path tests for `runWithReadyPattern`:
   - Invalid regex pattern
   - Pipe creation errors
   - Command start errors
3. Add error path tests for `shutdownProcess`:
   - Different signal types
   - Timeout → force kill path

### Phase 2: Edge Cases (Target: 80%+)
1. Ready pattern found → context cancelled
2. Ready pattern found → process exits
3. Command failure paths in `runWithShutdown`
4. Scanner error handling

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

### Coverage Goals
- **Minimum**: 75% overall (acceptable for v1)
- **Target**: 85% overall (good coverage)
- **Stretch**: 90%+ overall (excellent coverage)

## Estimated Effort

- **Phase 1**: 2-3 hours (critical paths)
- **Phase 2**: 2-3 hours (edge cases)
- **Phase 3**: 1-2 hours (polish)

**Total**: ~5-8 hours to reach 85%+ coverage


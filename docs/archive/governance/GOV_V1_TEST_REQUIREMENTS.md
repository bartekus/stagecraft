# GOV_V1_TEST_REQUIREMENTS — Test Strategy Standards

This document defines the test strategy standards for Stagecraft providers and core components, aligned with AATSE principles and GOV_V1_CORE governance.

⸻

## Core Principles

### 1. Determinism First

Tests MUST be deterministic:
- No `time.Sleep()` in tests (except explicitly documented integration scenarios)
- No goroutine-based tests without proper synchronization (channels, WaitGroup, context cancellation)
- No OS-level behavior dependencies (pipe buffering, process scheduling)
- No randomness or timestamps in test logic

### 2. Separation of Concerns

**Unit Tests** cover:
- Pure functions and isolated components
- Error paths and edge cases
- Deterministic state transitions

**Integration Tests** cover:
- Process lifecycle and orchestration
- Provider interface contracts
- End-to-end workflows

**Integration tests MUST NOT validate logic that can be unit-tested deterministically.**

### 3. No Test Seams

If a component requires a "test seam" (injectable dependency) to be testable, the design is incomplete.

**Preferred approach:**
- Extract pure, testable primitives
- Test primitives in isolation
- Test orchestration separately

**Example:** `PROVIDER_FRONTEND_GENERIC` extracts `scanStream()` as a pure function, eliminating the need for a `newScanner` test seam.

### 4. Coverage Targets

**Core Packages** (`pkg/config`, `internal/core`):
- Minimum: 80% coverage
- Target: 85%+ coverage

**Provider Implementations**:
- Minimum: 75% coverage (acceptable for v1)
- Target: 80%+ coverage
- Stretch: 85%+ coverage

**Interface Definitions** (`pkg/providers/*`):
- Minimum: 90% coverage (interfaces are small and critical)

⸻

## Test Organization

### File Structure

```
internal/providers/<provider>/<name>/
├── <name>.go              # Implementation
├── <name>_test.go         # Unit tests
├── <name>_integration_test.go  # Integration tests (if needed)
└── COVERAGE_STRATEGY.md   # Coverage strategy document
```

### Test Naming

- Unit tests: `Test<Function>_<Scenario>`
- Integration tests: `Test<Component>_Integration_<Scenario>`
- Benchmarks: `Benchmark<Function>_<Scenario>`

### Test Grouping

Group tests by:
1. Function being tested
2. Error path vs. success path
3. Unit vs. integration

⸻

## Provider Test Strategy Template

For new providers, follow this pattern (derived from `PROVIDER_FRONTEND_GENERIC`):

### Phase 1: Extract Deterministic Primitives

1. Identify stateful or concurrent logic
2. Extract pure, testable functions
3. Write unit tests for extracted primitives
4. Add benchmarks for performance-critical paths

### Phase 2: Integration Tests

1. Test process lifecycle (start, stop, shutdown)
2. Test error handling (invalid config, process failures)
3. Test provider interface contracts
4. **Do NOT duplicate unit test coverage in integration tests**

### Phase 3: Error Path Coverage

1. Invalid configuration scenarios
2. Process start failures
3. Process exit scenarios (success, error, timeout)
4. Context cancellation
5. Resource cleanup

### Success Criteria

- All tests pass with `-race` and `-count=20`
- No flaky tests
- Coverage meets or exceeds target
- No test seams required
- Integration tests focus on orchestration, not logic correctness

⸻

## Anti-Patterns to Avoid

### ❌ Flaky Patterns

```go
// BAD: time.Sleep in test
time.Sleep(10 * time.Millisecond)
if !ready {
    t.Fatal("not ready")
}

// GOOD: Deterministic synchronization
select {
case <-readyCh:
    // proceed
case <-time.After(timeout):
    t.Fatal("timeout")
}
```

### ❌ Test Seams

```go
// BAD: Test seam for injecting errors
var newScanner = func(r io.Reader) *bufio.Scanner {
    return bufio.NewScanner(r)
}

// GOOD: Extract pure function, test directly
func scanStream(r io.Reader, w io.Writer, pattern *regexp.Regexp) (<-chan struct{}, <-chan error) {
    // implementation
}
// Test scanStream directly with controlled inputs
```

### ❌ Integration Tests Validating Unit Logic

```go
// BAD: Integration test validating scanner logic
func TestProvider_Integration_ScannerError(t *testing.T) {
    // Uses real process, OS pipes, etc.
    // Tests scanner error handling
}

// GOOD: Unit test for scanner, integration test for orchestration
func TestScanStream_ScannerError(t *testing.T) {
    // Pure unit test with controlled reader
}

func TestProvider_Integration_ProcessLifecycle(t *testing.T) {
    // Tests process start/stop, not scanner logic
}
```

⸻

## Reference Implementation

See `internal/providers/frontend/generic/COVERAGE_STRATEGY.md` for a complete example of:
- Deterministic test design
- Pure function extraction
- Separation of unit and integration tests
- Coverage strategy documentation

⸻

## Validation

Tests MUST pass:
- `go test ./...` - All tests pass
- `go test -race ./...` - No race conditions
- `go test -count=20 ./...` - No flakiness
- `go test -cover ./<package>` - Meets coverage threshold

Governance checks (via `scripts/gov-pre-commit.sh`) validate:
- Core package coverage ≥ 80%
- Feature mapping (tests have correct Feature ID headers)
- No orphan test files

⸻

## Documentation Requirements

Each provider implementation SHOULD include:
- `COVERAGE_STRATEGY.md` - Coverage approach and status
- Test file headers with `// Feature: <FEATURE_ID>`
- Clear separation between unit and integration tests

⸻

**End of GOV_V1_TEST_REQUIREMENTS**

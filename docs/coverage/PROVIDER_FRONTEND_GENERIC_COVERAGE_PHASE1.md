# PROVIDER_FRONTEND_GENERIC Coverage Phase 1 Analysis Brief

This document captures the high level motivation, constraints, and success definition for Phase 1 of the test coverage strategy for PROVIDER_FRONTEND_GENERIC.

It is the starting point for the Implementation Outline and test plan.

This brief must be approved before outline work begins.

⸻

## 1. Problem Statement

The PROVIDER_FRONTEND_GENERIC provider currently has **70.2% test coverage**, which falls below the governance-aligned coverage thresholds enforced by `scripts/check-coverage.sh`. Critical error paths and edge cases are not adequately tested, creating risk:

- **Error handling gaps**: Invalid regex patterns, pipe creation failures, command start errors, and signal handling edge cases lack test coverage
- **Shutdown edge cases**: Timeout scenarios, force kill paths, and process state transitions are not fully exercised
- **Ready pattern edge cases**: Scanner errors, context cancellation after pattern detection, and process exit scenarios need coverage
- **Governance compliance**: Coverage thresholds must be met to satisfy GOV_V1_CORE requirements

This creates operational risk: production failures may occur in untested code paths, and CI coverage checks may fail as coverage requirements tighten.

⸻

## 2. Motivation

### Governance Compliance

- **GOV_V1_CORE Requirement**: Coverage thresholds are enforced in CI and MUST be maintained locally. Phase 1 targets 75%+ coverage to meet minimum acceptable thresholds.
- **Coverage Strategy Alignment**: This work implements Phase 1 of the coverage strategy defined in `internal/providers/frontend/generic/COVERAGE_STRATEGY.md`.
- **Deterministic Testing**: All tests must be deterministic, hermetic, and aligned with spec-defined behaviors.

### Operational Reliability

- **Error Path Validation**: Critical error paths (invalid config, command failures, signal handling) must be tested to ensure graceful degradation
- **Edge Case Coverage**: Timeout scenarios, process state transitions, and signal handling edge cases need validation
- **Regression Prevention**: Comprehensive test coverage prevents regressions as the codebase evolves

### Developer Experience

- **Test-Driven Development**: Following TDD discipline, tests are written first to define expected behavior
- **Clear Test Boundaries**: Tests establish clear boundaries between provider behavior and external dependencies
- **Maintainable Test Suite**: Well-organized tests make it easier to understand and modify provider behavior

⸻

## 3. Users and User Stories

### Developers

- As a developer, I want comprehensive test coverage for error paths, so I can confidently refactor and extend the provider
- As a developer, I want tests that clearly document expected behavior for edge cases, so I understand how the provider handles failures

### Platform Engineers

- As a platform engineer, I want coverage thresholds met, so CI checks pass and governance requirements are satisfied
- As a platform engineer, I want deterministic, hermetic tests, so test results are consistent across environments

### Automation and CI

- As a CI pipeline, I want coverage checks to pass, so builds succeed and quality gates are enforced
- As a CI pipeline, I want fast, reliable tests, so feedback loops remain short

⸻

## 4. Success Criteria (Phase 1)

1. **Coverage Threshold Met**: Overall coverage increases from 70.2% to **75%+** (minimum acceptable threshold)

2. **Critical Error Paths Covered**: All critical error paths have test coverage:
   - Invalid regex pattern compilation
   - Pipe creation errors (stdout/stderr)
   - Command start errors
   - Signal handling errors (SIGTERM, SIGKILL, unknown signals)
   - Timeout → force kill path

3. **Function-Level Coverage Targets Met**:
   - `runWithReadyPattern`: 64.0% → **75%+**
   - `runWithShutdown`: 66.7% → **75%+**
   - `shutdownProcess`: 64.0% → **75%+**
   - `Dev`: 84.0% → **85%+** (maintain or improve)

4. **Test Quality**: All tests follow TDD discipline:
   - Tests written before implementation changes (if any seams are needed)
   - Tests are deterministic (no timestamps, no random values)
   - Tests are hermetic (isolated, no shared state)
   - Tests align with spec-defined behaviors

5. **No Behavior Changes**: Phase 1 focuses on test coverage only:
   - No refactoring unless required to establish test seams
   - All seams documented if created
   - Existing behavior preserved exactly

6. **CI Compliance**: Coverage check passes with `--fail-on-warning` flag

⸻

## 5. Risks and Constraints

### Determinism Constraints

- Tests MUST NOT use timestamps or random values
- Tests MUST be deterministic across runs and environments
- Process execution timing must be controlled via context cancellation or deterministic delays

### Provider Constraints

- Provider behavior MUST NOT change (tests only, no implementation changes)
- Provider interfaces MUST remain stable
- No breaking changes to existing provider behavior

### Test Constraints

- Tests MUST follow spec-first discipline (align with `spec/providers/frontend/generic.md`)
- Tests MUST be hermetic (isolated, no shared state)
- Tests MUST use real shell scripts for integration-style tests (current approach)
- Tests MUST NOT use `t.Parallel()` unless explicitly allowed

### Coverage Constraints

- Coverage improvements MUST NOT come from removing code
- Coverage improvements MUST come from adding tests
- Minimum threshold: 75% overall (Phase 1 target)

### Architectural Constraints

- Test seams MUST be minimal and documented if created
- No new dependencies introduced for testing
- Existing test patterns MUST be followed

⸻

## 6. Coverage Gap Analysis

### Current Coverage (70.2%)

| Function | Current | Target | Gap |
|----------|---------|--------|-----|
| `ID` | 100.0% | 100% | ✅ Complete |
| `Dev` | 84.0% | 85%+ | 🟡 Minor gap |
| `parseConfig` | 85.7% | 85%+ | ✅ Complete |
| `runWithShutdown` | 66.7% | 75%+ | 🟠 Needs improvement |
| `shutdownProcess` | 64.0% | 75%+ | 🟠 Needs improvement |
| `runWithReadyPattern` | 64.0% | 75%+ | 🟠 Needs improvement |
| `init` | 100.0% | 100% | ✅ Complete |

### Missing Coverage (Phase 1 Priority)

#### `runWithReadyPattern` (64.0% → 75%+)
- Invalid regex pattern compilation error
- Stdout/stderr pipe creation errors
- Command start error
- Scanner error handling (stdout/stderr)
- Ready pattern found → context cancelled path
- Ready pattern found → process exits path

#### `runWithShutdown` (66.7% → 75%+)
- Command start error
- Command exits with non-zero exit code (ExitError path)
- Command exits with other error type

#### `shutdownProcess` (64.0% → 75%+)
- SIGTERM signal handling
- SIGKILL signal handling
- Unknown signal (defaults to SIGINT)
- Signal error: process already finished
- Timeout path: process doesn't exit → force kill
- Force kill error path

#### `Dev` (84.0% → 85%+)
- parseConfig error path (already partially covered, improve)

⸻

## 7. Dependencies

### Required Features

- **PROVIDER_FRONTEND_GENERIC**: The provider implementation must exist (✅ done)
- **GOV_V1_CORE**: Coverage governance framework (✅ done)
- **CORE_EXECUTIL**: Process execution utilities (✅ done)

### Blocking Features

None. This is a test-only change that can proceed independently.

### Enables Features

- **Future Coverage Phases**: Phase 1 establishes foundation for Phase 2 (80%+) and Phase 3 (85%+)
- **Provider Reliability**: Improved test coverage enables confident refactoring and extension

⸻

## 8. Test Strategy

### Test Organization

- Group tests by function being tested (`TestGenericProvider_Dev_*`, `TestGenericProvider_Shutdown_*`, etc.)
- Use table-driven tests where appropriate
- Keep integration tests separate from unit tests
- Follow existing test patterns in `generic_test.go`

### Mocking Strategy

- Use real shell scripts for integration-style tests (current approach)
- For error injection:
  - Use `exec.Command` with invalid commands
  - Create scripts that fail at specific points
  - Use context cancellation for timing control
- Avoid introducing new mocking frameworks

### Test Implementation Approach

1. **Write failing tests first** (TDD discipline)
2. **Verify tests fail** for the right reasons
3. **Ensure tests pass** with current implementation (no behavior changes needed)
4. **Document any test seams** if created

### Golden Tests

- Not required for Phase 1 (focus on error paths, not output formatting)
- May be needed in Phase 2/3 for complex output scenarios

⸻

## 9. Out of Scope (Phase 1)

- **Behavior Changes**: No refactoring or behavior modifications
- **New Features**: No new functionality added
- **Coverage Beyond 75%**: Phase 1 targets minimum threshold only
- **Golden Tests**: Not required for Phase 1
- **Performance Tests**: Not in scope
- **Integration with CLI_DEV**: Focus on provider unit/integration tests only

⸻

## 10. Approval

This analysis brief must be approved before proceeding to implementation outline.

**Next Steps:**
1. Review and approve this brief
2. Generate implementation outline (`docs/coverage/PROVIDER_FRONTEND_GENERIC_COVERAGE_PHASE1_OUTLINE.md`)
3. Implement tests following TDD workflow
4. Verify coverage thresholds met
5. Run CI checks with `--fail-on-warning`

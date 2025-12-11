# PROVIDER_BACKEND_GENERIC - Coverage Strategy (V1 Plan)

This document defines the coverage approach for the PROVIDER_BACKEND_GENERIC provider.
V1 coverage is currently in progress and will follow the same AATSE aligned strategy used for PROVIDER_FRONTEND_GENERIC.

---

## 🎯 Coverage Goals

The generic backend provider must:

1. Execute command-based backend operations reliably (Dev, BuildDocker, Plan).
2. Handle configuration parsing and validation errors deterministically.
3. Handle command execution failures and timeouts consistently.
4. Avoid test flakiness caused by:
   - OS-level process execution timing
   - Unbounded command execution timeouts
   - Environment variable dependencies

Tests SHOULD focus on:

- Configuration parsing and validation
- Command construction and argument building
- Error handling for command failures
- Integration tests that use controlled command execution

---

## ✔️ V1 Coverage Status - Plan

**Current Coverage: 84.1%** (exceeds 80% target ✅)

Target for v1: **≥ 80%** coverage (already met, but strategy needed for governance).

Initial coverage is focused on:

- Configuration parsing (YAML unmarshaling)
- Command execution (Dev, BuildDocker, Plan)
- Error handling for invalid configs and command failures

**Status**: Coverage already exceeds target, but needs:
- Coverage strategy document (this file) ✅
- Review for flaky patterns
- Verification of deterministic test design

---

## What Will Change for V1

### 1. Deterministic Helpers (if needed)

Where practical, generic provider logic SHOULD be factored into helper functions that:

- Accept explicit input structures (config, command args)
- Return explicit results or errors
- Are testable without external command execution

Examples (names illustrative):

- `buildCommandArgs(config)` - builds command and arguments from config
- `validateConfig(config)` - validates configuration structure

These helpers are candidates for unit tests.

### 2. Clear Boundary Between Unit and Integration Tests

Unit tests SHOULD:

- Test configuration parsing in isolation
- Test command argument building
- Use fake executors or test doubles where possible

Integration tests MAY:

- Execute real commands (with explicit opt-in)
- Be guarded behind build tags or environment variables
- Focus on end-to-end behavior

### 3. Avoiding Flakiness

To align with the "no broken glass" principle:

- Do not depend on external command availability in unit tests.
- Keep any real command execution to explicit integration tests with clear labels.
- Use deterministic fake executors or mock command execution for unit tests.

---

## 🧪 Coverage Philosophy: AATSE + "No Broken Glass"

The generic backend provider follows the same coverage philosophy as PROVIDER_FRONTEND_GENERIC:

- Deterministic primitives where possible.
- Clear separation between pure helpers and IO heavy integration code.
- Minimal concurrency surfaces and explicit synchronization.
- No tests that "sometimes fail" by design.

If a test requires real command execution, it should be considered an integration test and documented as such.

---

## 📈 V1 Plan - Gaps and Actions

Before V1 coverage is declared complete, the following SHOULD be done:

1. **Review existing tests** for:
   - Flakiness patterns (raw sleeps, uncontrolled timeouts)
   - External command dependencies that can be replaced with fakes
   - Coverage of all error paths

2. **Verify deterministic test design**:
   - No `time.Sleep` in tests
   - No uncontrolled goroutines
   - All tests pass with `-race` and `-count=20`

3. **Document coverage approach**:
   - This strategy document ✅
   - Update to "V1 Complete" once review confirms deterministic design

When these steps are complete, this document should be updated to:

- Change the title label to `Coverage Strategy (V1 Complete)`
- Include a detailed coverage table similar to PROVIDER_FRONTEND_GENERIC
- Remove or update any remaining "plan" language

---

## ✅ Conclusion

PROVIDER_BACKEND_GENERIC coverage is currently in **V1 Plan** status.

- Current: 84.1% (exceeds 80% target) ✅
- Approach: mirror the deterministic test strategy used for PROVIDER_FRONTEND_GENERIC.
- Next step: review existing tests for flakiness patterns and ensure deterministic design.

Once V1 is complete, a status document MUST be added at:

- `docs/engine/status/PROVIDER_BACKEND_GENERIC_COVERAGE_V1_COMPLETE.md`

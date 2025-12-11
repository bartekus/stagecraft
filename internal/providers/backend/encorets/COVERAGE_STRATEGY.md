# PROVIDER_BACKEND_ENCORE - Coverage Strategy (V1 Plan)

This document defines the coverage approach for the PROVIDER_BACKEND_ENCORE provider.
V1 coverage is currently in progress and will follow the same AATSE aligned strategy used for PROVIDER_FRONTEND_GENERIC.

---

## 🎯 Coverage Goals

The Encore.ts backend provider must:

1. Execute Encore.ts-specific backend operations reliably (Dev, BuildDocker, Plan).
2. Handle Encore.ts project structure detection and validation.
3. Handle command execution failures and Encore.ts-specific errors consistently.
4. Avoid test flakiness caused by:
   - OS-level process execution timing
   - Encore.ts CLI availability dependencies
   - Project structure detection timing

Tests SHOULD focus on:

- Encore.ts project structure detection
- Configuration parsing and validation
- Command construction for Encore.ts CLI
- Error handling for missing Encore.ts projects or CLI

---

## ✔️ V1 Coverage Status - Plan

**Current Coverage: 90.6%** (exceeds 80% target ✅)

Target for v1: **≥ 80%** coverage (already met, but strategy needed for governance).

Initial coverage is focused on:

- Encore.ts project detection (finding `encore.app` files)
- Configuration parsing
- Command execution (Dev, BuildDocker, Plan)
- Error handling for invalid projects and command failures

**Status**: Coverage already exceeds target, but needs:
- Coverage strategy document (this file) ✅
- Review for flaky patterns
- Verification of deterministic test design

---

## What Will Change for V1

### 1. Deterministic Helpers (if needed)

Where practical, Encore.ts provider logic SHOULD be factored into helper functions that:

- Accept explicit input structures (config, project paths)
- Return explicit results or errors
- Are testable without external command execution

Examples (names illustrative):

- `findEncoreApp(rootDir)` - finds Encore.ts project root
- `buildEncoreCommand(cmd, args)` - builds Encore.ts CLI command
- `validateEncoreProject(rootDir)` - validates project structure

These helpers are candidates for unit tests.

### 2. Clear Boundary Between Unit and Integration Tests

Unit tests SHOULD:

- Test project detection logic in isolation
- Test configuration parsing
- Test command argument building
- Use fake executors or test doubles where possible

Integration tests MAY:

- Execute real Encore.ts CLI commands (with explicit opt-in)
- Be guarded behind build tags or environment variables
- Focus on end-to-end behavior

### 3. Avoiding Flakiness

To align with the "no broken glass" principle:

- Do not depend on external Encore.ts CLI availability in unit tests.
- Keep any real command execution to explicit integration tests with clear labels.
- Use deterministic fake executors or mock command execution for unit tests.

---

## 🧪 Coverage Philosophy: AATSE + "No Broken Glass"

The Encore.ts backend provider follows the same coverage philosophy as PROVIDER_FRONTEND_GENERIC:

- Deterministic primitives where possible.
- Clear separation between pure helpers and IO heavy integration code.
- Minimal concurrency surfaces and explicit synchronization.
- No tests that "sometimes fail" by design.

If a test requires real Encore.ts CLI execution, it should be considered an integration test and documented as such.

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

PROVIDER_BACKEND_ENCORE coverage is currently in **V1 Plan** status.

- Current: 90.6% (exceeds 80% target) ✅
- Approach: mirror the deterministic test strategy used for PROVIDER_FRONTEND_GENERIC.
- Next step: review existing tests for flakiness patterns and ensure deterministic design.

Once V1 is complete, a status document MUST be added at:

- `docs/engine/status/PROVIDER_BACKEND_ENCORE_COVERAGE_V1_COMPLETE.md`

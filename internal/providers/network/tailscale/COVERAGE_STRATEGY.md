# PROVIDER_NETWORK_TAILSCALE - Coverage Strategy (V1 Plan)

This document defines the coverage approach for the PROVIDER_NETWORK_TAILSCALE provider.
V1 coverage is currently in progress and will follow the same AATSE aligned strategy used for PROVIDER_FRONTEND_GENERIC.

---

## 🎯 Coverage Goals

The Tailscale network provider must:

1. Correctly bring hosts into the configured Tailscale network topology.
2. Handle authentication and connectivity errors in a deterministic way.
3. Cleanly handle shutdown, disconnect, or teardown flows.
4. Avoid test flakiness caused by:
   - Real external Tailscale network calls in unit tests
   - Timing based assertions for background processes
   - OS specific behavior where avoidable

Tests SHOULD focus on:

- Command construction and arguments (for CLI based integration)
- Configuration parsing and validation
- Error handling and fallback behavior
- Integration tests that use fakes or controlled environments where possible

---

## ✔️ V1 Coverage Status - Plan

**Current Coverage (from recent run): 79.6 percent**

Target for v1: **≥ 80 percent** coverage.

**Slice 1 complete**: Extracted 4 pure helper functions and added comprehensive unit tests.

**Slice 2 complete**: Added comprehensive error path tests for `EnsureInstalled()`:
- Config validation tests (5 test cases)
- OS compatibility tests (9 test cases)
- Version enforcement tests (7 test cases)
- Install flow tests (2 test cases)
- Version parsing helper (`parseTailscaleVersion`) with 11 test cases

Initial coverage is focused on:

- Basic happy path flows
- Some error paths
- Registry integration

Missing or partial coverage:

- Some error branches and corner cases
- More exhaustive coverage of failure modes
- Full coverage of shutdown / teardown behavior

---

## What Will Change for V1

### 1. Deterministic Helpers

Where practical, Tailscale specific logic SHOULD be factored into helper functions that:

- Accept explicit input structures (config, arguments, environment)
- Return explicit results or errors
- Are testable without external Tailscale access

Examples (names illustrative):

- `buildTailscaleUpArgs(config)` - builds argument list for `tailscale up`
- `parseTailscaleStatus(output)` - parses status output into a structured type

These helpers are the primary candidates for unit tests.

### 2. Clear Boundary Between Unit and Integration Tests

Unit tests SHOULD:

- Avoid calling real Tailscale binaries
- Use fake executors or test doubles defined in the package
- Cover all main branches and error conditions

Integration tests MAY:

- Call real binaries (with explicit opt in)
- Be guarded behind build tags or environment variables
- Be optional for local runs and CI, depending on environment

Where previous tests relied on raw sleeps or external state, v1 SHOULD refactor them to deterministic patterns using channels or explicit polling loops with capped attempts.

### 3. Avoiding Flakiness

To align with the "no broken glass" principle:

- Do not depend on external network conditions in unit tests.
- Keep any real network reliance to explicit integration tests with clear labels.
- Use deterministic fake executors or mock responses for Tailscale commands.

---

## 🧪 Coverage Philosophy: AATSE + "No Broken Glass"

The Tailscale provider follows the same coverage philosophy as PROVIDER_FRONTEND_GENERIC:

- Deterministic primitives where possible.
- Clear separation between pure helpers and IO heavy integration code.
- Minimal concurrency surfaces and explicit synchronization.
- No tests that "sometimes fail" by design.

If a test requires `time.Sleep` or real network access, it should be considered an integration test and documented as such.

---

## 📈 V1 Plan - Gaps and Actions

Before V1 coverage is declared complete, the following SHOULD be done:

1. **Identify and list all public functions and critical internal helpers.**
2. **Add unit tests** covering:
   - Happy paths
   - Error paths
   - Edge cases (empty config, invalid config, missing dependencies)
3. **Review existing tests** for:
   - Flakiness patterns (raw sleeps, uncontrolled concurrency)
   - External dependencies that can be replaced with fakes

4. **Reach ≥ 80 percent coverage** for the package:
   - Confirm via:
     ```bash
     go test -cover ./internal/providers/network/tailscale
     ```

When these steps are complete, this document should be updated to:

- Change the title label to `Coverage Strategy (V1 Complete)`
- Include a coverage table similar to PROVIDER_FRONTEND_GENERIC
- Remove or update any remaining "plan" language

---

## ✅ Conclusion

PROVIDER_NETWORK_TAILSCALE coverage is currently in **V1 Plan** status.

- Target: ≥ 80 percent coverage.
- Approach: mirror the deterministic test strategy used for PROVIDER_FRONTEND_GENERIC.
- Next step: close coverage gaps and remove any flaky or externally dependent unit tests.

Once V1 is complete, a status document MUST be added at:

- `docs/engine/status/PROVIDER_NETWORK_TAILSCALE_COVERAGE_V1_COMPLETE.md`

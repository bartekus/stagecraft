# PROVIDER_CLOUD_DO - Coverage Strategy (V1 Plan)

This document defines the coverage approach for the PROVIDER_CLOUD_DO provider.
V1 coverage is currently in progress and will follow the same AATSE aligned strategy used for PROVIDER_FRONTEND_GENERIC.

---

## 🎯 Coverage Goals

The DigitalOcean cloud provider must:

1. Provision and manage infrastructure reliably (Plan, Apply, Hosts).
2. Handle API authentication and connectivity errors deterministically.
3. Handle resource creation/deletion failures consistently.
4. Avoid test flakiness caused by:
   - Real external API calls in unit tests
   - Timing based assertions for async operations
   - Network-dependent behavior

Tests SHOULD focus on:

- Configuration parsing and validation (token, SSH keys, host specs)
- Plan generation and reconciliation logic
- Error handling for API failures
- Integration tests that use fakes or controlled API clients

---

## ✔️ V1 Coverage Status - Plan

**Current Coverage: 79.7%** (just below 80% target)

Target for v1: **≥ 80%** coverage.

Initial coverage is focused on:

- Configuration parsing and validation
- Plan generation with reconciliation
- Apply operations (create/delete hosts)
- API client interface and error handling

Missing or partial coverage:

- Some error branches in API error handling
- Edge cases in plan reconciliation
- Full coverage of all error paths

---

## What Will Change for V1

### 1. Deterministic Helpers

Where practical, DigitalOcean specific logic SHOULD be factored into helper functions that:

- Accept explicit input structures (config, host specs)
- Return explicit results or errors
- Are testable without external API access

Examples (names illustrative):

- `buildDropletCreateRequest(spec)` - builds API request from host spec
- `parseDropletResponse(response)` - parses API response into structured type
- `reconcilePlan(desired, existing)` - deterministic plan reconciliation

These helpers are the primary candidates for unit tests.

### 2. Clear Boundary Between Unit and Integration Tests

Unit tests SHOULD:

- Avoid calling real DigitalOcean API
- Use fake API clients or test doubles defined in the package
- Cover all main branches and error conditions

Integration tests MAY:

- Call real API (with explicit opt-in and environment variables)
- Be guarded behind build tags or environment variables
- Be optional for local runs and CI

Where previous tests relied on raw sleeps or external state, v1 SHOULD refactor them to deterministic patterns.

### 3. Avoiding Flakiness

To align with the "no broken glass" principle:

- Do not depend on external network conditions in unit tests.
- Keep any real API reliance to explicit integration tests with clear labels.
- Use deterministic fake API clients or mock responses for DigitalOcean API calls.

---

## 🧪 Coverage Philosophy: AATSE + "No Broken Glass"

The DigitalOcean provider follows the same coverage philosophy as PROVIDER_FRONTEND_GENERIC:

- Deterministic primitives where possible.
- Clear separation between pure helpers and IO heavy integration code.
- Minimal concurrency surfaces and explicit synchronization.
- No tests that "sometimes fail" by design.

If a test requires real API access, it should be considered an integration test and documented as such.

---

## 📈 V1 Plan - Gaps and Actions

Before V1 coverage is declared complete, the following SHOULD be done:

1. **Identify and list all public functions and critical internal helpers.**
2. **Add unit tests** covering:
   - Happy paths (plan generation, host creation/deletion)
   - Error paths (invalid config, API failures, network errors)
   - Edge cases (empty config, invalid host specs, missing dependencies)
3. **Review existing tests** for:
   - Flakiness patterns (raw sleeps, uncontrolled concurrency)
   - External API dependencies that can be replaced with fakes

4. **Reach ≥ 80 percent coverage** for the package:
   - Confirm via:
     ```bash
     go test -cover ./internal/providers/cloud/digitalocean
     ```

When these steps are complete, this document should be updated to:

- Change the title label to `Coverage Strategy (V1 Complete)`
- Include a coverage table similar to PROVIDER_FRONTEND_GENERIC
- Remove or update any remaining "plan" language

---

## ✅ Conclusion

PROVIDER_CLOUD_DO coverage is currently in **V1 Plan** status.

- Current: 79.7% (just below 80% target)
- Target: ≥ 80 percent coverage.
- Approach: mirror the deterministic test strategy used for PROVIDER_FRONTEND_GENERIC.
- Next step: close coverage gaps and remove any flaky or externally dependent unit tests.

Once V1 is complete, a status document MUST be added at:

- `docs/engine/status/PROVIDER_CLOUD_DO_COVERAGE_V1_COMPLETE.md`

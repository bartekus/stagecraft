# PROVIDER_NETWORK_TAILSCALE - Coverage V1 Plan

**Status**: In Progress  
**Feature**: PROVIDER_NETWORK_TAILSCALE  
**Current Coverage**: ~68.2 percent  
**Target Coverage (V1)**: ≥ 80 percent

---

## Summary

This document tracks the work required to bring PROVIDER_NETWORK_TAILSCALE coverage to the same standard as PROVIDER_FRONTEND_GENERIC.

The goal is to:

- Reach at least 80 percent coverage.
- Remove or isolate any flaky tests.
- Ensure tests are deterministic and AATSE aligned.

---

## Plan

1. Review existing tests in `internal/providers/network/tailscale` for:
   - Time based sleeps
   - Real network dependencies
   - Fragile OS dependent assumptions

2. Identify critical helpers and execution paths and add unit tests:
   - Argument building for Tailscale commands
   - Error handling for missing binaries or configuration
   - Shutdown and disconnect behavior

3. Confirm that all tests pass consistently with:
   ```bash
   go test ./internal/providers/network/tailscale
   go test -race ./internal/providers/network/tailscale
   go test -count=20 ./internal/providers/network/tailscale
   ```

4. Update `COVERAGE_STRATEGY.md` to "V1 Complete" once:
   - Coverage is at or above target
   - Flaky patterns have been removed
   - The test design matches `GOV_V1_TEST_REQUIREMENTS`

---

## Alignment

This coverage plan MUST stay aligned with:
- `spec/providers/network/tailscale.md`
- `spec/features.yaml` (PROVIDER_NETWORK_TAILSCALE status)
- `docs/governance/GOV_V1_TEST_REQUIREMENTS.md`
- AATSE principles and "no broken glass" test philosophy.

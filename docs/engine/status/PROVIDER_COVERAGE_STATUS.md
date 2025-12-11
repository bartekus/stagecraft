# Provider Coverage Status

**Last Updated**: 2025-01-XX  
**Source**: `spec/features.yaml` and coverage strategy files

This document tracks the coverage status of all provider implementations marked as `done` in `spec/features.yaml`.

---

## Coverage Status Summary

| Feature ID | Status (spec) | Coverage Strategy | Coverage Status | Status Doc | Notes |
|------------|---------------|-------------------|-----------------|------------|-------|
| `PROVIDER_FRONTEND_GENERIC` | done | ✅ yes | V1 Complete | ✅ yes | **Reference model** - canonical example |
| `PROVIDER_BACKEND_ENCORE` | done | ✅ yes | V1 Complete | ✅ yes | 90.6% coverage (exceeds target) |
| `PROVIDER_BACKEND_GENERIC` | done | ✅ yes | V1 Complete | ✅ yes | 84.1% coverage (exceeds target) |
| `PROVIDER_CLOUD_DO` | done | ✅ yes | V1 Complete | ✅ yes | 80.5% coverage (exceeds target) |
| `PROVIDER_NETWORK_TAILSCALE` | done | ✅ yes | V1 Plan | ✅ yes | In progress (~68.2% → target ≥80%) |

---

## Provider Details

### ✅ PROVIDER_FRONTEND_GENERIC (V1 Complete)

- **Coverage**: 87.7% (exceeds 80% target)
- **Strategy**: `internal/providers/frontend/generic/COVERAGE_STRATEGY.md`
- **Status Doc**: `docs/engine/status/PROVIDER_FRONTEND_GENERIC_COVERAGE_V1_COMPLETE.md`
- **Reference**: Use as the canonical pattern for other providers
- **Key Pattern**: Extracted `scanStream()` pure helper, deterministic unit tests, no flaky patterns

### 🔄 PROVIDER_NETWORK_TAILSCALE (V1 Plan)

- **Current Coverage**: ~68.2%
- **Target Coverage**: ≥80%
- **Strategy**: `internal/providers/network/tailscale/COVERAGE_STRATEGY.md`
- **Status Doc**: `docs/engine/status/PROVIDER_NETWORK_TAILSCALE_COVERAGE_PLAN.md`
- **Next Steps**: Extract deterministic helpers, add error path tests, remove any flaky patterns

### ✅ PROVIDER_BACKEND_ENCORE (V1 Complete)

- **Coverage**: 90.6% (exceeds 80% target)
- **Strategy**: `internal/providers/backend/encorets/COVERAGE_STRATEGY.md`
- **Status Doc**: `docs/engine/status/PROVIDER_BACKEND_ENCORE_COVERAGE_V1_COMPLETE.md`
- **Key Achievement**: Verified deterministic test design with zero flakiness patterns

### ✅ PROVIDER_BACKEND_GENERIC (V1 Complete)

- **Coverage**: 84.1% (exceeds 80% target)
- **Strategy**: `internal/providers/backend/generic/COVERAGE_STRATEGY.md`
- **Status Doc**: `docs/engine/status/PROVIDER_BACKEND_GENERIC_COVERAGE_V1_COMPLETE.md`
- **Key Achievement**: Verified deterministic test design with zero flakiness patterns

### ✅ PROVIDER_CLOUD_DO (V1 Complete)

- **Coverage**: 80.5% (exceeds 80% target)
- **Strategy**: `internal/providers/cloud/digitalocean/COVERAGE_STRATEGY.md`
- **Status Doc**: `docs/engine/status/PROVIDER_CLOUD_DO_COVERAGE_V1_COMPLETE.md`
- **Key Achievement**: Added test for `Hosts()` stub method, achieving 80% coverage threshold

---

## Governance Requirements

Per `docs/governance/GOV_V1_TEST_REQUIREMENTS.md`:

- **All providers marked `done`** MUST have a `COVERAGE_STRATEGY.md` file
- **Providers claiming "V1 Complete"** MUST have a corresponding status document
- **Coverage target**: ≥80% for provider implementations
- **Test quality**: Deterministic, no flakiness, AATSE-aligned

---

## Next Actions

### Priority 1: Complete PROVIDER_BACKEND_GENERIC

Already exceeds 80% coverage but needs:

1. Review existing tests for flakiness patterns
2. Verify deterministic design (`-race`, `-count=20`)
3. Update strategy to "V1 Complete"
4. Create status document

### Priority 3: Complete PROVIDER_NETWORK_TAILSCALE

1. Extract deterministic helpers (if needed)
2. Add missing error path tests
3. Verify coverage ≥80%
4. Update strategy to "V1 Complete"
5. Create `PROVIDER_NETWORK_TAILSCALE_COVERAGE_V1_COMPLETE.md`

### Workflow

Follow `docs/engine/agents/PROVIDER_COVERAGE_AGENT.md` workflow to systematically bring each provider to V1 Complete status.

---

## Validation

Run provider governance checks:

```bash
./scripts/check-provider-governance.sh
```

This will verify:
- Coverage strategy files exist for all `done` providers
- Status documents exist when "V1 Complete" is claimed
- Feature IDs in strategies match `spec/features.yaml`

---

**Note**: This document is manually maintained. Update when provider coverage status changes.

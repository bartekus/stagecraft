> **Superseded by** `docs/coverage/COVERAGE_LEDGER.md` section 5.5 (PROVIDER_CLOUD_DO) and `docs/governance/GOVERNANCE_ALMANAC.md` section 4 (Provider Governance). Kept for historical reference. New coverage snapshots and summaries MUST go into the coverage ledger.

# PROVIDER_CLOUD_DO - Coverage V1 Completion Plan

**Feature**: PROVIDER_CLOUD_DO  
**Current Coverage**: 79.7%  
**Target Coverage**: ≥80%  
**Gap**: 0.3 percentage points

---

## Summary

PROVIDER_CLOUD_DO is just 0.3% away from the 80% target. This is a quick win to bring it to V1 Complete status.

---

## Current State

**Coverage**: 79.7% (just below 80% target)

**Key Functions** (from `internal/providers/cloud/digitalocean/`):
- `Plan()` - Infrastructure planning with reconciliation
- `Apply()` - Idempotent droplet creation/deletion
- `Hosts()` - List provisioned hosts
- Config parsing and validation
- API client interface and error handling

**Test Files**:
- `do_test.go` - 22 tests covering main scenarios

---

## Missing Coverage Analysis

To reach ≥80%, we need to cover additional error paths:

### 1. API Client Error Paths

**Potential gaps:**
- API client initialization failures
- Network timeout scenarios
- API rate limiting errors
- Invalid API response parsing

**Test ideas:**
- `TestDigitalOceanProvider_Plan_APIClientError` - API client fails to initialize
- `TestDigitalOceanProvider_Apply_APITimeout` - API call times out
- `TestDigitalOceanProvider_Hosts_InvalidResponse` - API returns invalid JSON

### 2. Config Validation Edge Cases

**Potential gaps:**
- Empty host list
- Invalid region/size combinations
- Missing required fields

**Test ideas:**
- `TestParseConfig_EmptyHosts` - Config with no hosts defined
- `TestParseConfig_InvalidRegion` - Invalid region string
- `TestParseConfig_MissingToken` - Missing required token_env

### 3. Plan Reconciliation Edge Cases

**Potential gaps:**
- Plan with no changes needed
- Plan with only deletions
- Plan with only creations

**Test ideas:**
- `TestPlan_NoChanges` - Desired state matches existing
- `TestPlan_OnlyDeletions` - All hosts should be deleted
- `TestPlan_OnlyCreations` - All hosts are new

---

## Implementation Plan

### Step 1: Identify Exact Gaps

Run coverage with function-level detail:

```bash
go test -coverprofile=coverage.out ./internal/providers/cloud/digitalocean
go tool cover -func=coverage.out | grep -E "(digitalocean|do\.go)" | sort -k3 -n
```

This will show which functions are below 80% and need additional tests.

### Step 2: Add Targeted Tests

Add 1-2 tests covering the lowest-coverage error paths:

**Priority 1**: Error paths in `Plan()` or `Apply()` that are currently untested
**Priority 2**: Config validation edge cases
**Priority 3**: API client error handling

### Step 3: Verify Coverage

```bash
go test -cover ./internal/providers/cloud/digitalocean
# Should show ≥80%
```

### Step 4: Verify Determinism

```bash
go test -race ./internal/providers/cloud/digitalocean
go test -count=20 ./internal/providers/cloud/digitalocean
```

### Step 5: Update Documentation

1. Update `COVERAGE_STRATEGY.md` to "V1 Complete"
2. Create `docs/engine/status/PROVIDER_CLOUD_DO_COVERAGE_V1_COMPLETE.md`
3. Update `PROVIDER_COVERAGE_STATUS.md`

---

## Estimated Effort

**Time**: 30-60 minutes
- 15-20 min: Identify exact gaps (coverage analysis)
- 15-20 min: Write 1-2 targeted tests
- 10-15 min: Verify and update docs

**Complexity**: Low - just need to add a few error path tests

---

## Success Criteria

- ✅ Coverage ≥80% (verified via `go test -cover`)
- ✅ All tests pass with `-race` and `-count=20`
- ✅ No flaky patterns introduced
- ✅ Coverage strategy updated to "V1 Complete"
- ✅ Status document created

---

## Reference

- Coverage Strategy: `internal/providers/cloud/digitalocean/COVERAGE_STRATEGY.md`
- Test Requirements: `docs/governance/GOV_V1_TEST_REQUIREMENTS.md`
- Coverage Agent: `docs/engine/agents/PROVIDER_COVERAGE_AGENT.md`
- Reference Model: `internal/providers/frontend/generic/COVERAGE_STRATEGY.md`

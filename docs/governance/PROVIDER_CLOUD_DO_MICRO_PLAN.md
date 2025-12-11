# PROVIDER_CLOUD_DO - Micro-Plan: Quick Win to V1 Complete

**Feature**: PROVIDER_CLOUD_DO  
**Current Coverage**: 79.7%  
**Target Coverage**: ≥80%  
**Gap**: 0.3 percentage points  
**Estimated Time**: 30-45 minutes

---

## Coverage Analysis

**Function-level coverage (sorted by coverage):**
- `Hosts()` - **0.0%** ⚠️ **PRIMARY TARGET**
- `Apply()` - 69.1% (secondary target if needed)
- `firstNonEmpty()` - 75.0%
- `parseConfig()` - 88.2%
- `Plan()` - 91.9%
- `ID()`, `NewDigitalOceanProvider()`, `init()` - 100.0%

**Strategy**: Add a single test for `Hosts()` to push coverage from 79.7% → ≥80%.

---

## Implementation Steps

### Step 1: Examine `Hosts()` Function

**Location**: `internal/providers/cloud/digitalocean/do.go:283`

**Function signature**:
```go
func (p *DigitalOceanProvider) Hosts(ctx context.Context, opts cloud.HostsOptions) ([]cloud.Host, error)
```

**Current implementation** (stub):
```go
func (p *DigitalOceanProvider) Hosts(ctx context.Context, opts cloud.HostsOptions) ([]cloud.Host, error) {
	// TODO: Implement full Hosts method in later slices
	// For now, return empty list to satisfy interface
	return []cloud.Host{}, nil
}
```

**Note**: This is currently a stub that returns an empty list. Testing it is trivial but necessary for coverage. If this doesn't push to ≥80%, we may also need to add an error path test for `Apply()` (currently 69.1%).

### Step 2: Identify Test Scenarios

**Required test** (minimum to hit ≥80%):
1. ✅ **`TestDigitalOceanProvider_Hosts_Stub`** - Test stub implementation
   - Verify it returns empty list
   - Verify no error
   - This should push coverage from 79.7% → ≥80%

**If first test doesn't push to ≥80%**, add:
2. **`TestDigitalOceanProvider_Apply_ErrorPath`** - Error handling in Apply()
   - Test API client error scenario
   - Test config parsing error
   - Apply() is at 69.1%, so error paths are likely untested

### Step 3: Write Test

**File**: `internal/providers/cloud/digitalocean/do_test.go`

**Test structure** (following existing patterns):

**Test 1: Hosts() stub** (trivial, but necessary):
```go
func TestDigitalOceanProvider_Hosts_Stub(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	provider := NewDigitalOceanProvider()
	
	opts := cloud.HostsOptions{
		Environment: "prod",
		Config: map[string]interface{}{
			"token_env": "DO_TOKEN",
		},
	}

	hosts, err := provider.Hosts(ctx, opts)
	if err != nil {
		t.Fatalf("Hosts() failed: %v", err)
	}

	if len(hosts) != 0 {
		t.Errorf("Hosts() returned %d hosts, want 0 (stub)", len(hosts))
	}
}
```

**Test 2: Apply() error path** (if needed):
```go
func TestDigitalOceanProvider_Apply_APIError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClient := &MockAPIClient{
		CreateDropletFunc: func(ctx context.Context, req CreateDropletRequest) (*Droplet, error) {
			return nil, ErrAPIError
		},
	}

	provider := NewDigitalOceanProviderWithClient(mockClient)
	
	opts := cloud.ApplyOptions{
		Environment: "prod",
		Config: map[string]interface{}{
			"token_env": "DO_TOKEN",
			"ssh_key_name": "test-key",
			"hosts": map[string]interface{}{
				"prod": []interface{}{
					map[string]interface{}{
						"name": "prod-web-1",
						"region": "nyc1",
						"size": "s-1vcpu-1gb",
					},
				},
			},
		},
	}

	_, err := provider.Apply(ctx, opts)
	if err == nil {
		t.Fatal("Apply() should return error on API failure")
	}
}
```

**Note**: Adjust mock structure to match existing `MockAPIClient` pattern in the test file.

### Step 4: Verify Coverage

```bash
# Run coverage
go test -cover ./internal/providers/cloud/digitalocean
# Expected: ≥80%

# Verify determinism
go test -race ./internal/providers/cloud/digitalocean
go test -count=20 ./internal/providers/cloud/digitalocean

# Verify governance
./scripts/check-provider-governance.sh
```

### Step 5: Update Documentation

1. **Update `COVERAGE_STRATEGY.md`**:
   - Change status from "V1 Plan" to "V1 Complete"
   - Update coverage percentage
   - Add note about `Hosts()` test coverage

2. **Create status document**: `docs/engine/status/PROVIDER_CLOUD_DO_COVERAGE_V1_COMPLETE.md`

3. **Update `PROVIDER_COVERAGE_STATUS.md`**:
   - Change status from "V1 Plan" to "V1 Complete"
   - Update coverage percentage

---

## Expected Outcome

- ✅ Coverage: 79.7% → ≥80%
- ✅ All tests pass with `-race` and `-count=20`
- ✅ No flaky patterns introduced
- ✅ Status updated to "V1 Complete"
- ✅ Status document created

---

## Commit Message

```
test(PROVIDER_CLOUD_DO): achieve v1 coverage

- Add test for Hosts() function (0% → covered)
- Coverage: 79.7% → ≥80% (V1 Complete)
- Update COVERAGE_STRATEGY.md to V1 Complete
- Add PROVIDER_CLOUD_DO_COVERAGE_V1_COMPLETE.md status doc
```

---

## Reference

- Coverage Strategy: `internal/providers/cloud/digitalocean/COVERAGE_STRATEGY.md`
- Test Requirements: `docs/governance/GOV_V1_TEST_REQUIREMENTS.md`
- Coverage Agent: `docs/engine/agents/PROVIDER_COVERAGE_AGENT.md`
- Reference Model: `internal/providers/frontend/generic/COVERAGE_STRATEGY.md`

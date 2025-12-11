# PROVIDER_NETWORK_TAILSCALE Slice 1 - Pre-Implementation Checklist

**Before starting**, verify these to avoid common pitfalls:

---

## Pre-Flight Checks

- [x] **Confirm `ErrConfigInvalid` exists**: ✅ Verified in `internal/providers/network/tailscale/errors.go`
  - Error is defined: `ErrConfigInvalid = errors.New("invalid config")`
  - Already used in `parseConfig()` and `NodeFQDN()` - reuse this

- [ ] **Verify package structure**: Confirm helpers will be in `package tailscale` (same package)
  - No need to expose helpers outside package
  - Avoids circular dependencies

- [ ] **Check error message style**: Review existing error messages in `tailscale.go`
  - Should match pattern: `"tailscale provider: %w: ..."`
  - Keep consistent with existing style

- [ ] **Review existing test patterns**: Look at `tailscale_test.go` to match:
  - Table-driven test style
  - `t.Parallel()` usage
  - Error assertion patterns

---

## Implementation Checklist

### Step 1: Extract Helpers
- [ ] Add `buildTailscaleUpCommand()` to `tailscale.go`
- [ ] Add `parseOSRelease()` to `tailscale.go`
- [ ] Add `validateTailnetDomain()` to `tailscale.go`
- [ ] Add `buildNodeFQDN()` to `tailscale.go`
- [ ] Verify imports (`fmt`, `strings`) are present

### Step 2: Refactor Existing Code
- [ ] Update `EnsureJoined()` to use `buildTailscaleUpCommand()`
- [ ] Update `checkOSCompatibility()` to use `parseOSRelease()`
- [ ] Update `NodeFQDN()` to use `validateTailnetDomain()` and `buildNodeFQDN()`
- [ ] Run existing tests to ensure behavior unchanged:
  ```bash
  go test ./internal/providers/network/tailscale
  ```

### Step 3: Add Unit Tests
- [ ] Add `TestBuildTailscaleUpCommand` with table-driven tests
- [ ] Add `TestParseOSRelease` with table-driven tests
- [ ] Add `TestValidateTailnetDomain` with table-driven tests
- [ ] Add `TestBuildNodeFQDN` with table-driven tests
- [ ] Add `TestParseStatus_InvalidJSON`
- [ ] Add `TestParseStatus_EmptyJSON`
- [ ] Add `TestParseStatus_MissingFields`

### Step 4: Verify
- [ ] Run coverage: `go test -cover ./internal/providers/network/tailscale`
  - **Record actual coverage %** (don't guess)
- [ ] Run race detector: `go test -race ./internal/providers/network/tailscale`
- [ ] Run flakiness check: `go test -count=20 ./internal/providers/network/tailscale`
- [ ] Run governance check: `./scripts/check-provider-governance.sh`

### Step 5: Update Documentation
- [ ] Update `COVERAGE_STRATEGY.md` with **actual** coverage percentage
- [ ] Add note about Slice 1 completion
- [ ] Update `PROVIDER_COVERAGE_STATUS.md` with **actual** coverage percentage
- [ ] Update `PROVIDER_NETWORK_TAILSCALE_COVERAGE_PLAN.md` with Slice 1 status

### Step 6: Commit
- [ ] Stage all changed files
- [ ] Use commit message from agent prompt
- [ ] Verify commit message format passes hooks

---

## Common Pitfalls to Avoid

❌ **Don't** expose helpers as public functions (keep them package-local)  
❌ **Don't** change error message format (match existing style)  
❌ **Don't** guess coverage percentage (use actual `go test -cover` output)  
❌ **Don't** add CLI invocation tests (deferred to Slice 2)  
❌ **Don't** use `time.Sleep` in tests  
❌ **Don't** make helpers depend on external state (keep them pure)

---

## Quick Verification Commands

```bash
# Before starting - baseline coverage
go test -cover ./internal/providers/network/tailscale

# After helpers extracted - verify behavior unchanged
go test ./internal/providers/network/tailscale

# After tests added - verify coverage increase
go test -cover ./internal/providers/network/tailscale

# Final verification
go test -race ./internal/providers/network/tailscale
go test -count=20 ./internal/providers/network/tailscale
./scripts/check-provider-governance.sh
```

---

**Ready to start?** Use `docs/engine/agents/PROVIDER_NETWORK_TAILSCALE_SLICE1_AGENT.md` as your execution guide.

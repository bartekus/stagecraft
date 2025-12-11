# PROVIDER_NETWORK_TAILSCALE Slice 1 - Ready to Execute

**Status**: ✅ Ready  
**Agent Prompt**: `docs/engine/agents/PROVIDER_NETWORK_TAILSCALE_SLICE1_AGENT.md`  
**Checklist**: `docs/governance/PROVIDER_NETWORK_TAILSCALE_SLICE1_CHECKLIST.md`  
**Plan**: `docs/governance/PROVIDER_NETWORK_TAILSCALE_SLICE1_PLAN.md`

---

## Quick Start

1. **Open the agent prompt**: `docs/engine/agents/PROVIDER_NETWORK_TAILSCALE_SLICE1_AGENT.md`
2. **Follow the checklist**: `docs/governance/PROVIDER_NETWORK_TAILSCALE_SLICE1_CHECKLIST.md`
3. **Execute tasks** in order (extract → refactor → test → verify → docs → commit)

---

## What You're Building

**4 Pure Helper Functions**:
- `buildTailscaleUpCommand()` - Builds Tailscale CLI command string
- `parseOSRelease()` - Parses `/etc/os-release` content
- `validateTailnetDomain()` - Validates Tailnet domain format
- `buildNodeFQDN()` - Builds node FQDN from host + domain

**7 New Unit Tests**:
- 4 tests for new helpers (table-driven)
- 3 tests for `parseStatus()` edge cases

**Expected Result**: Coverage 68.2% → ~75%

---

## Pre-Verified

- ✅ `ErrConfigInvalid` exists in `errors.go`
- ✅ Error message style confirmed: `"tailscale provider: %w: ..."`
- ✅ Test patterns reviewed: table-driven with `t.Parallel()`
- ✅ Package structure confirmed: helpers stay in `package tailscale`

---

## Success Criteria

- ✅ 4 helpers extracted and tested
- ✅ Existing code refactored (behavior unchanged)
- ✅ Coverage increases to ~75%
- ✅ All tests pass with `-race` and `-count=20`
- ✅ Documentation updated with actual coverage %

---

## Next Steps After Slice 1

**Slice 2**: Add error path tests for `EnsureInstalled()` and `EnsureJoined()` using mock Commander  
**Slice 3**: Final push to ≥80% (if needed)

---

**Ready to execute!** 🚀

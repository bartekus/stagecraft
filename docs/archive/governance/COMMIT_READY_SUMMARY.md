> **Superseded by** `docs/governance/GOVERNANCE_ALMANAC.md` section 3 (Commit and PR Discipline). Kept for historical reference. New governance rules MUST be recorded in the almanac.

# Governance Commit - Ready Summary

**Status**: ✅ All files ready for commit  
**Commit Type**: `chore(GOV_V1_CORE)`  
**Scope**: Provider coverage governance enforcement

---

## Quick Commit Commands

```bash
# Stage files
git add \
  scripts/check-provider-governance.sh \
  scripts/provider-coverage-planner.sh \
  scripts/gov-pre-commit.sh \
  scripts/run-all-checks.sh \
  internal/providers/frontend/generic/COVERAGE_STRATEGY.md \
  internal/providers/network/tailscale/COVERAGE_STRATEGY.md \
  internal/providers/backend/generic/COVERAGE_STRATEGY.md \
  internal/providers/backend/encorets/COVERAGE_STRATEGY.md \
  internal/providers/cloud/digitalocean/COVERAGE_STRATEGY.md \
  docs/engine/status/PROVIDER_NETWORK_TAILSCALE_COVERAGE_PLAN.md \
  docs/engine/status/README.md \
  docs/engine/agents/STAGECRAFT_VALIDATION_AGENT.md \
  docs/engine/agents/PROVIDER_COVERAGE_AGENT.md \
  docs/engine/status/PROVIDER_COVERAGE_STATUS.md \
  docs/governance/PROVIDER_BACKEND_ENCORE_COVERAGE_PLAN.md \
  docs/governance/PROVIDER_BACKEND_GENERIC_COVERAGE_PLAN.md \
  docs/governance/PROVIDER_CLOUD_DO_COVERAGE_PLAN.md \
  docs/governance/PROVIDER_GOVERNANCE_SUMMARY.md \
  docs/engine/agents/GOV_PRE_COMMIT_INTEGRATION.md \
  docs/governance/PR_TEMPLATE_PROVIDER_COVERAGE.md \
  docs/governance/CI_PROVIDER_COVERAGE_ENFORCEMENT.md \
  docs/README.md \
  .hooks/pre-commit \
  Agent.md

# Verify (optional but recommended)
./scripts/run-all-checks.sh

# Commit
git commit -m "chore(GOV_V1_CORE): enforce provider coverage governance

- Add provider coverage strategy validation script
- Add provider coverage completion planner
- Create coverage strategies for all done providers
- Add provider coverage governance to validation agent
- Integrate provider checks into run-all-checks.sh
- Add PR template and CI enforcement docs

All providers marked 'done' now require COVERAGE_STRATEGY.md.
Provider coverage governance is CI-enforceable."

# Verify commit
git status
```

---

## Next Steps After Commit

### 1. Quick Follow-Up: infra-up Spec Fix

**File**: `docs/governance/INFRA_UP_SPEC_FIX.md`  
**Time**: 5-10 minutes  
**Commit**: `docs(CLI_INFRA_UP): fix spec frontmatter exit codes`

### 2. Provider Coverage Improvements

**Priority order**:
1. **PROVIDER_CLOUD_DO** (30-45 min) - Quick win to ≥80%
   - See: `docs/governance/PROVIDER_CLOUD_DO_MICRO_PLAN.md`
2. **PROVIDER_BACKEND_ENCORE** (30-45 min) - Formalize V1 Complete
   - See: `docs/governance/PROVIDER_BACKEND_ENCORE_COVERAGE_PLAN.md`
3. **PROVIDER_BACKEND_GENERIC** (30-45 min) - Formalize V1 Complete
   - See: `docs/governance/PROVIDER_BACKEND_GENERIC_COVERAGE_PLAN.md`
4. **PROVIDER_NETWORK_TAILSCALE** (2-3 hours) - Improve to ≥80%
   - See: `docs/engine/status/PROVIDER_NETWORK_TAILSCALE_COVERAGE_PLAN.md`

---

## Verification Checklist

Before committing, verify:
- [x] `./scripts/check-provider-governance.sh` passes
- [x] `./scripts/provider-coverage-planner.sh` shows correct status
- [x] All coverage strategies exist for `done` providers
- [x] Documentation is complete and consistent

---

**Ready to commit!** 🚀

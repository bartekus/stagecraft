> **Superseded by** `docs/governance/GOVERNANCE_ALMANAC.md` sections 3 (Commit and PR Discipline) and 4 (Provider Governance). Kept for historical reference. New governance rules MUST be recorded in the almanac.

# Commit Guidance: Provider Coverage Governance

**Feature**: GOV_CORE  
**Type**: Governance enhancement  
**Scope**: Provider coverage enforcement and tooling

---

## Suggested Commit Message

```
chore(GOV_CORE): enforce provider coverage governance

- Add provider coverage strategy validation script
- Add provider coverage completion planner
- Create coverage strategies for all done providers
- Add provider coverage governance to validation agent
- Integrate provider checks into run-all-checks.sh
- Add PR template and CI enforcement docs

All providers marked 'done' now require COVERAGE_STRATEGY.md.
Provider coverage governance is CI-enforceable.
```

---

## Staging Checklist

### Core Scripts
- [x] `scripts/check-provider-governance.sh` - Validation script
- [x] `scripts/provider-coverage-planner.sh` - Planning tool
- [x] `scripts/gov-pre-commit.sh` - Pre-commit governance checks
- [x] `.hooks/pre-commit-gov-snippet.sh` - Hook snippet

### Coverage Strategies
- [x] `internal/providers/frontend/generic/COVERAGE_STRATEGY.md` - Updated to V1 Complete
- [x] `internal/providers/network/tailscale/COVERAGE_STRATEGY.md` - V1 Plan
- [x] `internal/providers/backend/generic/COVERAGE_STRATEGY.md` - V1 Plan
- [x] `internal/providers/backend/encorets/COVERAGE_STRATEGY.md` - V1 Plan
- [x] `internal/providers/cloud/digitalocean/COVERAGE_STRATEGY.md` - V1 Plan

### Documentation
- [x] `docs/coverage/PROVIDER_COVERAGE_TEMPLATE.md` - Template for new providers
- [x] `docs/engine/agents/PROVIDER_COVERAGE_AGENT.md` - Coverage improvement agent
- [x] `docs/engine/agents/STAGECRAFT_VALIDATION_AGENT.md` - Updated with provider checks
- [x] `docs/engine/status/PROVIDER_COVERAGE_STATUS.md` - Central status dashboard
- [x] `docs/engine/status/PROVIDER_FRONTEND_GENERIC_COVERAGE_V1_COMPLETE.md` - Completion tracking
- [x] `docs/engine/status/PROVIDER_FRONTEND_GENERIC_COVERAGE_PR.md` - PR template example
- [x] `docs/engine/status/PROVIDER_NETWORK_TAILSCALE_COVERAGE_PLAN.md` - Plan tracking
- [x] `docs/governance/GOV_V1_TEST_REQUIREMENTS.md` - Test strategy standards
- [x] `docs/governance/CI_PROVIDER_COVERAGE_ENFORCEMENT.md` - CI integration guide
- [x] `docs/governance/PR_TEMPLATE_PROVIDER_COVERAGE.md` - PR template
- [x] `docs/governance/PROVIDER_CLOUD_DO_COVERAGE_PLAN.md` - Quick win plan
- [x] `docs/governance/PROVIDER_BACKEND_ENCORE_COVERAGE_PLAN.md` - Review plan
- [x] `docs/governance/PROVIDER_BACKEND_GENERIC_COVERAGE_PLAN.md` - Review plan

### Integration
- [x] `scripts/run-all-checks.sh` - Updated to include provider checks
- [x] `.hooks/pre-commit` - Updated with governance checks
- [x] `Agent.md` - Updated with governance wrapper verification
- [x] `docs/engine/status/README.md` - Updated with new status docs
- [x] `docs/README.md` - Updated with provider coverage governance section

### Validation
- [x] `VALIDATION_REPORT.md` - Initial structural analysis

---

## Pre-Commit Verification

Before committing, run:

```bash
# 1. Provider governance check
./scripts/check-provider-governance.sh

# 2. Coverage planner (visual check)
./scripts/provider-coverage-planner.sh

# 3. Full validation suite (optional but recommended)
./scripts/run-all-checks.sh
```

All should pass without errors.

---

## Post-Commit

After committing, the governance layer will:

1. **Enforce on every commit** (via pre-commit hook):
   - All `done` providers must have `COVERAGE_STRATEGY.md`
   - V1 Complete providers must have status documents

2. **Enable systematic improvement**:
   - Use `PROVIDER_COVERAGE_AGENT.md` to improve providers
   - Use `provider-coverage-planner.sh` to track progress

3. **Scale to new providers**:
   - New providers automatically require coverage strategies
   - Template provides consistent structure

---

## Next Steps

After this commit:

1. **Quick win**: Complete `PROVIDER_CLOUD_DO` (79.7% → 80%+)
   - See: `docs/governance/PROVIDER_CLOUD_DO_COVERAGE_PLAN.md`
   - Estimated: 30-60 minutes

2. **Formalize**: Mark `PROVIDER_BACKEND_ENCORE` and `PROVIDER_BACKEND_GENERIC` as V1 Complete
   - See: `docs/governance/PROVIDER_BACKEND_ENCORE_COVERAGE_PLAN.md`
   - See: `docs/governance/PROVIDER_BACKEND_GENERIC_COVERAGE_PLAN.md`
   - Estimated: 30-45 minutes each

3. **Improve**: Complete `PROVIDER_NETWORK_TAILSCALE` (68.2% → 80%+)
   - See: `docs/engine/status/PROVIDER_NETWORK_TAILSCALE_COVERAGE_PLAN.md`
   - Estimated: 2-3 hours

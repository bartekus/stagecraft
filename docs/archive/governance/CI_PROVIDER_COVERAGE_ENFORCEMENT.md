> **Superseded by** `docs/governance/GOVERNANCE_ALMANAC.md` section 4.2 (Provider Coverage Requirements) and section 8 (Tooling and Automation). Kept for historical reference. New governance rules MUST be recorded in the almanac.

# CI Provider Coverage Enforcement

This document describes how provider coverage governance is enforced in CI.

---

## Current Enforcement

Provider coverage governance is enforced through:

1. **`scripts/check-provider-governance.sh`** - Validates coverage strategy presence and status documents
2. **`scripts/run-all-checks.sh`** - Includes provider governance checks as part of full validation suite
3. **`scripts/gov-pre-commit.sh`** - Runs governance checks before commits (optional, can be bypassed)

---

## CI Integration

### Option 1: Add to Existing CI Workflow

If you have a GitHub Actions workflow (e.g., `.github/workflows/ci.yml`), add:

```yaml
- name: Provider Governance Checks
  run: ./scripts/check-provider-governance.sh
```

This should run after tests but before deployment.

### Option 2: Standalone Provider Coverage Job

Create `.github/workflows/provider-coverage.yml`:

```yaml
name: Provider Coverage Governance

on:
  pull_request:
    paths:
      - 'internal/providers/**'
      - 'spec/features.yaml'
      - 'scripts/check-provider-governance.sh'
  push:
    branches: [main]

jobs:
  provider-governance:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      
      - name: Run Provider Governance Checks
        run: ./scripts/check-provider-governance.sh
      
      - name: Run Provider Coverage Planner
        run: ./scripts/provider-coverage-planner.sh
```

---

## Enforcement Rules

### Hard Failures (CI blocks merge)

1. **Missing Coverage Strategy**
   - Provider marked `done` in `spec/features.yaml` but no `COVERAGE_STRATEGY.md`
   - **Action**: Create strategy from template

2. **Invalid Feature ID**
   - Coverage strategy references Feature ID not in `spec/features.yaml`
   - **Action**: Fix Feature ID in strategy heading

### Warnings (CI reports but doesn't block)

1. **Missing Status Doc**
   - Coverage strategy claims "V1 Complete" but status document missing
   - **Action**: Create status document

2. **Coverage Below Target**
   - Provider coverage below 80% (detected via planner, not enforced yet)
   - **Action**: Improve coverage following `PROVIDER_COVERAGE_AGENT.md`

---

## Future Enhancements

### Coverage Threshold Enforcement

Add coverage threshold checking to `check-provider-governance.sh`:

```bash
# For each provider, check actual coverage
coverage=$(go test -cover "./internal/providers/${path}" 2>&1 | grep -oP 'coverage: \K[0-9.]+')
if (( $(echo "$coverage < 80" | bc -l) )); then
    echo "ERROR: ${provider_id} coverage ${coverage}% is below 80% target"
    status=1
fi
```

### Automated Coverage Reporting

Generate coverage reports in CI and post to PR comments or status checks.

---

## Manual Validation

Run locally before pushing:

```bash
# Check provider governance
./scripts/check-provider-governance.sh

# See coverage status
./scripts/provider-coverage-planner.sh

# Full validation suite
./scripts/run-all-checks.sh
```

---

**Note**: CI enforcement is currently manual (run checks in CI). Future enhancement: automate coverage threshold checking.

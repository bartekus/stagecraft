> **Superseded by** `docs/engine/agents/AGENT_BRIEF_COVERAGE.md` section "Phase 1: Compliance Unblock". Kept for historical reference. New coverage execution notes MUST go into the merged agent brief.

# Agent Brief: Test Coverage Compliance - Phase 1

**Feature ID**: GOV_V1_CORE  
**Spec**: `spec/governance/GOV_V1_CORE.md`  
**Context**: Test coverage & governance hardening. See `docs/coverage/TEST_COVERAGE_ANALYSIS.md` and `docs/coverage/COVERAGE_COMPLIANCE_PLAN.md`.

---

## Mission

Bring Stagecraft into coverage compliance by completing **Phase 1 – Compliance Unblock** from `TEST_COVERAGE_ANALYSIS.md`, without changing user-facing behaviour beyond clear bug fixes.

---

## Scope

### 1. Fix 4 Failing Tests

#### `internal/tools/cliintrospect`
- `TestIntrospect_WithSubcommands` (expected 2 subcommands, got 0)
- `TestFlagToInfo_BoolFlag` (persistent expected true)

**Approach**: Prefer fixing test setup/assumptions first. Only adjust implementation if it's clearly wrong relative to the CLI and spec.

**Files**:
- `internal/tools/cliintrospect/cliintrospect_test.go`
- `internal/tools/cliintrospect/cliintrospect.go`
- `cmd/stagecraft/main.go` (reference for real CLI structure)

#### `internal/cli/commands` (build command)
- `TestBuildInvalidEnvFails` (expected invalid environment error, got config not found)
- `TestBuildExplicitVersionIsReflected` (expected dry-run success, got config not found)

**Approach**: 
- Review existing CLI test helpers (`test_helpers.go`, `phases_test_helpers_test.go`)
- Fix test setup to create minimal valid `stagecraft.yml` in temp dir
- Ensure `--config` flag or working directory is set correctly
- Keep error path for "invalid env" intact; don't weaken validation

**Files**:
- `internal/cli/commands/build_test.go`
- `internal/cli/commands/build.go`
- `internal/cli/commands/deploy_test.go` (reference for working pattern)

---

### 2. Improve `pkg/config` Coverage from 66.7% → ≥ 80%

**Targeted additions** to `pkg/config/config_test.go`:

- **Second `GetProviderConfig` overload** (currently 0%):
  - Provider exists but env missing
  - Provider + env exist but key missing
  - Happy path with nested map structure

- **`Load` error paths** (currently 78.6%):
  - Invalid YAML syntax
  - File exists but missing required sections

- **`Exists` edge cases** (currently 83.3%):
  - Non-existent path
  - Directory vs file (if relevant)

**Constraint**: No refactors; just additional tests and tiny supporting helpers if required.

---

### 3. Add Missing Test Files for "done" Features

#### `PROVIDER_BACKEND_INTERFACE`
**File**: `pkg/providers/backend/backend_test.go`

**Content**: Minimal tests that:
- Exercise interface/registry behaviour
- Verify compile-time contracts via dummy implementation
- Include Feature ID header:

```go
// Feature: PROVIDER_BACKEND_INTERFACE
// Spec: spec/core/backend-registry.md
```

**Reference**: `pkg/providers/frontend/frontend_test.go` (similar pattern)

---

#### `CLI_DEPLOY`
**File**: `test/e2e/deploy_smoke_test.go`

**Content**: Minimal smoke test that:
- Uses existing E2E helpers (like `init_smoke_test.go`, `dev_smoke_test.go`)
- Sets up minimal "hello world" project
- Runs `stagecraft deploy --env=test --dry-run`
- Asserts:
  - Exit code 0
  - Output contains deterministic marker (e.g., "deploy plan complete")

**Header**:
```go
// Feature: CLI_DEPLOY
// Spec: spec/commands/deploy.md
```

---

## Constraints

### Must Follow (Agent.md)
- **Single feature scope**: GOV_V1_CORE
- **Test-first**: Write/update failing tests first
- **No refactors** outside direct scope of coverage/compliance
- **No new dependencies**
- **Keep diffs minimal and deterministic**
- **Do not change coverage thresholds or scripts**
- **Do not weaken validation** just to satisfy tests; fix tests unless behaviour is clearly wrong

---

## Success Criteria

- `./scripts/run-all-checks.sh` passes, including:
  - `./scripts/check-coverage.sh --fail-on-warning`
- **No failing tests** across the suite
- **`pkg/config` coverage ≥ 80%**
- **Both missing test files exist** and are wired to their Feature IDs
- `docs/coverage/TEST_COVERAGE_ANALYSIS.md` remains accurate (or minimally updated to reflect "fully compliant" status)

---

## Execution Checklist

- [ ] Fix `TestIntrospect_WithSubcommands`
- [ ] Fix `TestFlagToInfo_BoolFlag`
- [ ] Fix `TestBuildInvalidEnvFails`
- [ ] Fix `TestBuildExplicitVersionIsReflected`
- [ ] Add tests for second `GetProviderConfig` overload
- [ ] Add error path tests for `Load`
- [ ] Add edge case tests for `Exists`
- [ ] Verify `pkg/config` coverage ≥ 80%
- [ ] Create `pkg/providers/backend/backend_test.go`
- [ ] Create `test/e2e/deploy_smoke_test.go`
- [ ] Run `./scripts/run-all-checks.sh` and verify all pass
- [ ] Run `./scripts/check-coverage.sh --fail-on-warning` and verify pass

---

## Reference Documents

- `docs/coverage/TEST_COVERAGE_ANALYSIS.md` - Detailed analysis
- `docs/coverage/COVERAGE_COMPLIANCE_PLAN.md` - Complete action plan
- `spec/governance/GOV_V1_CORE.md` - Feature specification
- `Agent.md` - Development protocol


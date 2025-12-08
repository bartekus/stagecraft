# CLI_PLAN Analysis and Preparation Summary

This document summarizes the analysis and preparation work completed for implementing `CLI_PLAN` as specified.

## Analysis Completed

### 1. Existing Infrastructure Review

**CORE_PLAN Status: ✅ Complete**
- Location: `internal/core/plan.go`
- Provides `Planner` type with `PlanDeploy(envName string) (*Plan, error)` method
- Returns structured `Plan` with:
  - `Environment string`
  - `Operations []Operation` (with Type, Description, Dependencies, Metadata)
  - `Metadata map[string]interface{}`
- Operation types: `infra_provision`, `migration`, `build`, `deploy`, `health_check`
- Already tested and used by `CLI_DEPLOY` and `CLI_BUILD`

**Version Resolution: ✅ Available**
- Function: `resolveVersion(ctx, versionFlag, logger)` in `internal/cli/commands/deploy.go`
- Logic:
  1. If `--version` flag provided, use it (try to get commit SHA from git)
  2. Else, try to get current Git SHA via `git rev-parse HEAD`
  3. Fall back to `"unknown"` if Git unavailable
- Can be reused directly or extracted to shared helper

**Command Patterns: ✅ Established**
- Commands follow consistent pattern: `ResolveFlags` → `Load config` → `Validate` → `Generate plan` → `Execute/Render`
- Test helpers: `setupIsolatedStateTestEnv` for isolated testing
- Golden file testing pattern established in `deploy_test.go` and `build_test.go`
- Flag resolution infrastructure in `internal/cli/commands/flags.go`

**Dependencies: ✅ All Met**
- `CORE_PLAN` - done
- `CORE_CONFIG` - done
- `CLI_DEPLOY` - done (for version resolution semantics)
- `CLI_BUILD` - done (for understanding build phase semantics)

### 2. Key Findings

1. **No plan command exists yet** - Fresh implementation needed
2. **CORE_PLAN is ready** - Can be used as-is without modifications
3. **Version resolution can be reused** - Same logic as deploy command
4. **Test patterns established** - Can follow deploy/build test patterns
5. **Filtering needs implementation** - Service filtering logic needs to be built
6. **Rendering needs implementation** - Text and JSON output formatters need to be built

## Deliverables Created

### 1. Specification Document ✅

**File:** `spec/commands/plan.md`

Comprehensive specification covering:
- Purpose and scope
- CLI interface (flags, usage)
- Inputs and outputs
- Behavior (workflow, filtering semantics, ordering guarantees)
- Output formats (text and JSON)
- Error handling and exit codes
- Determinism requirements
- Examples
- Testing requirements
- Implementation notes

### 2. Features.yaml Update ✅

**File:** `spec/features.yaml`

Updated `CLI_PLAN` entry to include dependencies:
- `CORE_PLAN`
- `CORE_CONFIG`
- `CLI_DEPLOY`
- `CLI_BUILD`

### 3. Implementation Outline ✅

**File:** `docs/CLI_PLAN_IMPLEMENTATION_OUTLINE.md`

Detailed implementation guide including:
- Analysis summary
- Step-by-step implementation plan
- Code skeletons for command structure
- Helper function designs
- Test strategy
- Key design decisions
- Future enhancements

## Implementation Readiness

### Ready to Implement

✅ **All prerequisites met:**
- CORE_PLAN exists and is functional
- Version resolution logic available
- Command patterns established
- Test infrastructure in place
- Specification complete

✅ **Clear implementation path:**
- Follow established command patterns (deploy/build)
- Reuse existing infrastructure (CORE_PLAN, version resolution)
- Implement filtering and rendering logic
- Create comprehensive tests with golden files

### Implementation Steps (When Ready)

1. **Create command structure** (`internal/cli/commands/plan.go`)
   - Command skeleton with flags
   - Main execution function
   - Filtering helpers
   - Rendering helpers (text and JSON)

2. **Register command** (`internal/cli/root.go`)
   - Add `NewPlanCommand()` in lexicographic order

3. **Create tests** (`internal/cli/commands/plan_test.go`)
   - Unit tests for filtering
   - Integration tests for full command
   - Golden file tests for output formats

4. **Create golden files** (`internal/cli/commands/testdata/`)
   - Text format examples
   - JSON format examples

## Key Design Decisions

1. **Reuse over rebuild**
   - Use CORE_PLAN as-is
   - Reuse version resolution from deploy
   - Follow established patterns

2. **Minimal v1 scope**
   - Focus on core flags: `--env`, `--version`, `--services`, `--format`
   - Stub advanced filters (`--roles`, `--hosts`, `--phases`) for future

3. **Determinism first**
   - All output must be deterministic
   - Suitable for golden file testing
   - No timestamps or random ordering

4. **No side effects**
   - Pure read-only operation
   - No state writes
   - No external command execution

## Next Steps

When ready to implement:

1. Review the specification (`spec/commands/plan.md`)
2. Review the implementation outline (`docs/CLI_PLAN_IMPLEMENTATION_OUTLINE.md`)
3. Create `internal/cli/commands/plan.go` following the skeleton
4. Create `internal/cli/commands/plan_test.go` with comprehensive tests
5. Register command in `internal/cli/root.go`
6. Create golden files as tests drive the output shape
7. Run full test suite and verify no regressions

## Notes

- The specification is comprehensive and ready for implementation
- The implementation outline provides concrete code skeletons
- All dependencies are satisfied
- Test patterns are established and can be followed
- The feature is well-scoped for v1 with clear extension points


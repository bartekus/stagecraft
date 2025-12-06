# CLI_PLAN Implementation Outline

This document provides a concrete implementation outline for the `CLI_PLAN` feature based on the specification in `spec/commands/plan.md`.

## Analysis Summary

### Existing Infrastructure

1. **CORE_PLAN exists and is functional**
   - Located in `internal/core/plan.go`
   - Provides `Planner` type with `PlanDeploy(envName string) (*Plan, error)` method
   - Returns `Plan` struct with:
     - `Environment string`
     - `Operations []Operation`
     - `Metadata map[string]interface{}`
   - Operation types: `infra_provision`, `migration`, `build`, `deploy`, `health_check`
   - Already used by `CLI_DEPLOY` and `CLI_BUILD`

2. **Version resolution logic exists**
   - Located in `internal/cli/commands/deploy.go`
   - Function: `resolveVersion(ctx, versionFlag, logger) (version, commitSHA string)`
   - Logic:
     1. If `--version` flag provided, use it (try to get commit SHA from git)
     2. Else, try to get current Git SHA via `git rev-parse HEAD`
     3. Fall back to `"unknown"` if Git unavailable
   - Can be reused or extracted to shared helper

3. **Command patterns established**
   - Commands follow pattern: `ResolveFlags` → `Load config` → `Validate` → `Generate plan` → `Execute/Render`
   - Test helpers: `setupIsolatedStateTestEnv` for isolated testing
   - Golden file testing pattern established in `deploy_test.go` and `build_test.go`

4. **Flag resolution infrastructure**
   - `ResolveFlags(cmd, cfg)` in `internal/cli/commands/flags.go`
   - Handles `--env`, `--config`, `--verbose`, `--dry-run`
   - Environment validation against config

## Implementation Steps

### Step 1: Create Spec File ✅

- [x] Create `spec/commands/plan.md` with comprehensive specification
- [x] Update `spec/features.yaml` to mark dependencies

### Step 2: Create Command Structure

#### 2.1 Create `internal/cli/commands/plan.go`

**Command skeleton:**

```go
// Feature: CLI_PLAN
// Spec: spec/commands/plan.md

func NewPlanCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "plan",
        Short: "Show the deployment plan without executing it",
        Long:  "Generates and displays a read-only deployment plan for the specified environment",
        RunE:  runPlan,
    }

    // Flags
    cmd.Flags().StringP("env", "e", "", "Target environment (e.g. staging, prod)")
    cmd.Flags().StringP("version", "v", "", "Version to plan for (defaults to current deployment version)")
    cmd.Flags().String("services", "", "Comma-separated list of services to include")
    cmd.Flags().String("format", "text", "Output format: text or json")
    cmd.Flags().BoolP("verbose", "V", false, "Show more detail")
    
    // Future extensions (v1 minimal, can be stubbed):
    // cmd.Flags().String("roles", "", "Comma-separated list of host roles")
    // cmd.Flags().String("hosts", "", "Comma-separated list of hostnames")
    // cmd.Flags().String("phases", "", "Comma-separated list of phase IDs/prefixes")

    _ = cmd.MarkFlagRequired("env")
    
    return cmd
}

func runPlan(cmd *cobra.Command, args []string) error {
    // Implementation
}
```

**Implementation function:**

```go
func runPlan(cmd *cobra.Command, args []string) error {
    ctx := cmd.Context()
    if ctx == nil {
        ctx = context.Background()
    }

    // 1. Resolve global flags
    flags, err := ResolveFlags(cmd, nil)
    if err != nil {
        return fmt.Errorf("resolving flags: %w", err)
    }

    // 2. Load config
    cfg, err := config.Load(flags.Config)
    if err != nil {
        if err == config.ErrConfigNotFound {
            return fmt.Errorf("stagecraft config not found at %s", flags.Config)
        }
        return fmt.Errorf("loading config: %w", err)
    }

    // 3. Re-resolve flags with config for environment validation
    flags, err = ResolveFlags(cmd, cfg)
    if err != nil {
        return fmt.Errorf("resolving flags: %w", err)
    }

    // 4. Validate environment is provided
    if flags.Env == "" {
        return fmt.Errorf("environment is required; use --env flag")
    }

    // 5. Initialize logger
    logger := logging.NewLogger(flags.Verbose)

    // 6. Parse plan-specific flags
    versionFlag, _ := cmd.Flags().GetString("version")
    servicesFlag, _ := cmd.Flags().GetString("services")
    formatFlag, _ := cmd.Flags().GetString("format")
    verboseFlag, _ := cmd.Flags().GetBool("verbose")

    // 7. Resolve version (reuse deploy logic)
    version, _ := resolveVersion(ctx, versionFlag, logger)

    // 8. Parse services list
    var services []string
    if servicesFlag != "" {
        services = parseServicesList(servicesFlag)
    }

    // 9. Generate plan
    planner := core.NewPlanner(cfg)
    plan, err := planner.PlanDeploy(flags.Env)
    if err != nil {
        return fmt.Errorf("generating deployment plan: %w", err)
    }

    // 10. Store version in plan metadata for rendering
    if plan.Metadata == nil {
        plan.Metadata = make(map[string]interface{})
    }
    plan.Metadata["version"] = version

    // 11. Apply filters
    filteredPlan := applyFilters(plan, services, nil, nil, nil) // roles, hosts, phases stubbed for v1

    // 12. Render output
    opts := PlanRenderOptions{
        Format:  formatFlag,
        Verbose: verboseFlag,
    }
    return renderPlan(cmd.OutOrStdout(), filteredPlan, flags.Env, version, opts)
}
```

#### 2.2 Helper Functions

**Filtering:**

```go
// applyFilters applies service, role, host, and phase filters to a plan
func applyFilters(plan *core.Plan, services, roles, hosts, phases []string) *core.Plan {
    // For v1, implement service filtering
    // Future: add role/host/phase filtering
    
    if len(services) == 0 {
        return plan
    }

    // Build set of services to include
    serviceSet := make(map[string]bool)
    for _, svc := range services {
        serviceSet[svc] = true
    }

    // Filter operations: keep if they touch at least one service
    filteredOps := []core.Operation{}
    for _, op := range plan.Operations {
        if operationTouchesServices(op, serviceSet) {
            filteredOps = append(filteredOps, op)
        }
    }

    return &core.Plan{
        Environment: plan.Environment,
        Operations:  filteredOps,
        Metadata:    plan.Metadata,
    }
}

// operationTouchesServices checks if an operation touches any of the specified services
func operationTouchesServices(op core.Operation, serviceSet map[string]bool) bool {
    // Check metadata for service information
    // This is a simplified check; actual implementation depends on how
    // CORE_PLAN stores service information in operation metadata
    if services, ok := op.Metadata["services"].([]string); ok {
        for _, svc := range services {
            if serviceSet[svc] {
                return true
            }
        }
    }
    // If no service info, include by default (preserves dependencies)
    return true
}
```

**Rendering:**

```go
type PlanRenderOptions struct {
    Format  string // "text" or "json"
    Verbose bool
}

func renderPlan(out io.Writer, plan *core.Plan, env, version string, opts PlanRenderOptions) error {
    switch opts.Format {
    case "text":
        return renderPlanText(out, plan, env, version, opts)
    case "json":
        return renderPlanJSON(out, plan, env, version, opts)
    default:
        return fmt.Errorf("invalid format: %s (must be 'text' or 'json')", opts.Format)
    }
}

func renderPlanText(out io.Writer, plan *core.Plan, env, version string, opts PlanRenderOptions) error {
    // Render hierarchical text output
    // Ensure deterministic ordering: phases sorted by ID, services/hosts sorted lexicographically
    // ...
}

func renderPlanJSON(out io.Writer, plan *core.Plan, env, version string, opts PlanRenderOptions) error {
    // Render JSON output
    // Use encoding/json with deterministic ordering
    // ...
}
```

#### 2.3 Register Command

Update `internal/cli/root.go`:

```go
cmd.AddCommand(commands.NewPlanCommand())
```

Add in lexicographic order (after `migrate`, before `releases`).

### Step 3: Create Tests

#### 3.1 Create `internal/cli/commands/plan_test.go`

**Test structure (following deploy_test.go pattern):**

```go
func TestNewPlanCommand_HasExpectedMetadata(t *testing.T) {
    cmd := NewPlanCommand()
    // Assertions
}

func TestPlanCommand_ConfigNotFound(t *testing.T) {
    // Test missing config file
}

func TestPlanCommand_InvalidEnvironment(t *testing.T) {
    // Test unknown environment
}

func TestPlanCommand_MissingEnvFlag(t *testing.T) {
    // Test missing --env flag
}

func TestPlanCommand_HappyPathText(t *testing.T) {
    // Test basic plan generation with text output
    // Use golden file comparison
}

func TestPlanCommand_ServiceFiltering(t *testing.T) {
    // Test --services filter
}

func TestPlanCommand_JSONFormat(t *testing.T) {
    // Test JSON output format
}

func TestPlanCommand_Determinism(t *testing.T) {
    // Test that same inputs produce identical output
}

func TestPlanCommand_ErrorPropagation(t *testing.T) {
    // Test that CORE_PLAN errors propagate correctly
}
```

#### 3.2 Create Golden Files

- `internal/cli/commands/testdata/plan_staging_all.txt`
- `internal/cli/commands/testdata/plan_prod_api_only.txt`
- `internal/cli/commands/testdata/plan_staging_json.json` (optional)

### Step 4: Implementation Details

#### 4.1 Version Resolution

Reuse `resolveVersion` from `deploy.go`. Consider extracting to shared helper if needed:

```go
// In plan.go or shared helper
func resolveVersionForPlan(ctx context.Context, versionFlag string, logger logging.Logger) string {
    version, _ := resolveVersion(ctx, versionFlag, logger)
    return version
}
```

#### 4.2 Plan Rendering

**Text format considerations:**
- Use deterministic phase ordering (topological sort or lexicographical by ID)
- Sort services and hosts lexicographically within each phase
- No timestamps
- Stable formatting

**JSON format considerations:**
- Use `encoding/json` with sorted keys
- Ensure arrays are sorted
- Stable schema

#### 4.3 Filtering Logic

**Service filtering:**
- Parse comma-separated list (reuse `parseServicesList` from `build.go`)
- Check if operation metadata contains service information
- Preserve dependencies: if a deploy phase is included, include its build phase

**Future filtering (v1 minimal):**
- Role/host filtering: stubbed for v1
- Phase filtering: stubbed for v1

### Step 5: Integration

1. **Register command in root**
   - Add to `internal/cli/root.go` in lexicographic order

2. **Update documentation**
   - Ensure CLI reference docs mention `plan` command

3. **Test integration**
   - Run full test suite
   - Verify no regressions

## Key Design Decisions

1. **Reuse existing infrastructure**
   - Use `CORE_PLAN` as-is
   - Reuse `resolveVersion` from deploy
   - Follow established command patterns

2. **Minimal v1 scope**
   - Focus on `--env`, `--version`, `--services`, `--format`
   - Stub `--roles`, `--hosts`, `--phases` for future

3. **Determinism first**
   - All output must be deterministic
   - Suitable for golden file testing
   - No timestamps or random ordering

4. **No side effects**
   - No state writes
   - No external command execution
   - Pure read-only operation

## Testing Strategy

1. **Unit tests**
   - Test flag parsing
   - Test filtering logic
   - Test rendering functions

2. **Integration tests**
   - Test full command execution
   - Test with real config files
   - Test error cases

3. **Golden file tests**
   - Test text output format
   - Test JSON output format
   - Test determinism

## Future Enhancements

1. **Role/host filtering**
   - Implement `--roles` and `--hosts` flags
   - Add filtering logic

2. **Phase filtering**
   - Implement `--phases` flag
   - Add prefix matching

3. **Enhanced output**
   - Add `--verbose` detail levels
   - Add more metadata to JSON output

4. **Plan validation**
   - Validate plan completeness
   - Check for missing dependencies


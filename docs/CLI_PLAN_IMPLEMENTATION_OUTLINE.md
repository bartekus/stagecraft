Cool, so the tension is now actually resolved in code:
	•	CLI_PLAN – uses its own resolvePlanVersion helper, never shells out to git, falls back to "unknown" when --version is omitted.
	•	CLI_DEPLOY / CLI_BUILD – free to use the existing resolveVersion (with git), since they’re allowed to have side-effects.

That’s exactly option 2 we discussed, and your agent log confirms the spec + impl match that.

On to the next thing: a detailed implementation plan for CLI_BUILD.

Below is something you can drop into docs/CLI_BUILD_IMPLEMENTATION_OUTLINE.md if you want; it mirrors the style of the CLI_PLAN outline you already have.

⸻

CLI_BUILD Implementation Outline

This document provides a concrete implementation outline for the CLI_BUILD feature, aligned with spec/commands/build.md and the existing Stagecraft architecture.

1. Goal & Role of CLI_BUILD

stagecraft build is responsible for:
	•	Building Docker/OCI images for the configured services for a given environment.
	•	Optionally pushing those images to a registry.
	•	Reusing CORE_PLAN to know what to build (instead of ad-hoc logic).
	•	Reusing CLI_PHASE_EXECUTION_COMMON to execute build phases in a consistent way.

It is not responsible for:
	•	Deploying containers (that’s CLI_DEPLOY).
	•	Writing release state (CORE_STATE) – builds are not releases.
	•	Talking directly to remote hosts – providers/drivers own that.

High-level flow:

CLI_BUILD → Resolve flags → Load config → Resolve env + version
         → CORE_PLAN.PlanDeploy(env) → filter to build ops
         → execute build ops via providers (PhaseFns)


⸻

2. Existing Infrastructure

Before implementing, rely on:
	1.	CORE_PLAN (already done)
	•	internal/core/plan.go
	•	Planner.PlanDeploy(envName string) (*Plan, error)
	•	Returns Plan with Operations []Operation and OperationType enums.
	2.	Version Resolution (deploy-style)
	•	internal/cli/commands/deploy.go
	•	resolveVersion(ctx, versionFlag, logger) (version, commitSHA string)
	•	Uses:
	•	--version flag when present.
	•	Otherwise git rev-parse HEAD.
	•	Falls back to "unknown".
	3.	Phase Execution Common
	•	CLI_PHASE_EXECUTION_COMMON (you’ve already implemented this for deploy/rollback).
	•	Shared executePhasesCommon and PhaseFns abstraction to run phases.
	4.	Flag Resolution & Logging
	•	internal/cli/commands/flags.go – ResolveFlags(cmd, cfg)
	•	Logging: internal/logging (or equivalent) used by deploy/build.
	5.	Test Patterns
	•	deploy_test.go, rollback_test.go, plan_test.go:
	•	Isolated environment helpers.
	•	Golden file tests for CLI output as needed.
	•	Deterministic, spec-driven behaviour.

⸻

3. CLI Contract

The details live in spec/commands/build.md; this summarizes what the implementation must respect.

3.1 Usage

stagecraft build [flags]

3.2 Flags

Required:
	•	--env, -e <env>
	•	Target environment (e.g. staging, prod).
	•	Must exist in environments in stagecraft.yml.

Optional:
	•	--version, -v <version>
	•	Version/tag to use for images.
	•	If omitted, use deploy-style version resolution:
	•	Try git SHA via resolveVersion.
	•	--services <svc1,svc2,...>
	•	Comma-separated services to build.
	•	Service filtering semantics:
	•	Only build operations that touch at least one of these services.
	•	Spec decides whether unknown services are an error.
	•	--push
	•	If true, push built images to the registry.
	•	--no-cache
	•	Build without cache (if provider supports it).
	•	Global flags (--config, --verbose, etc.) via CLI_GLOBAL_FLAGS.

⸻

4. File Structure

Implementation will primarily touch:
	•	spec/commands/build.md – spec must be complete and aligned with this outline.
	•	spec/features.yaml – CLI_BUILD entry must be wip → done once finished.
	•	internal/cli/commands/build.go – command wiring + orchestration.
	•	internal/cli/commands/phases_build.go (or similar) – build-specific phase handlers (if you keep phases in a separate file).
	•	internal/cli/commands/build_test.go – tests.
	•	internal/cli/commands/testdata/build_*.golden – golden outputs if build prints summaries.
	•	internal/cli/root.go – ensure NewBuildCommand() is registered and in lexicographic order.

⸻

5. Step-by-Step Implementation Plan

Step 1 – Ensure Spec and Features Are Ready
	1.	spec/commands/build.md
	•	Confirm it defines:
	•	Purpose & scope.
	•	Flags and semantics.
	•	Version resolution (explicitly says it may use git).
	•	Behaviour: env resolution, service filtering, push/no-cache.
	•	Error handling & exit codes.
	•	Determinism requirements.
	2.	spec/features.yaml
	•	Ensure CLI_BUILD:
	•	Has status wip.
	•	Depends on CORE_PLAN, CORE_CONFIG, provider features as needed.

Only proceed once spec and features.yaml reflect the desired behaviour.

⸻

Step 2 – Command Skeleton in build.go

Implement/confirm:

// Feature: CLI_BUILD
// Spec: spec/commands/build.md

func NewBuildCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "build",
        Short: "Build application images for a given environment",
        Long:  "Builds Docker/OCI images for configured services in the target environment.",
        RunE:  runBuild,
    }

    cmd.Flags().StringP("env", "e", "", "Target environment (e.g. staging, prod)")
    cmd.Flags().StringP("version", "v", "", "Version/tag to build (defaults to git SHA or 'unknown')")
    cmd.Flags().String("services", "", "Comma-separated list of services to build")
    cmd.Flags().Bool("push", false, "Push built images to registry after building")
    cmd.Flags().Bool("no-cache", false, "Build images without cache")

    _ = cmd.MarkFlagRequired("env")

    return cmd
}


⸻

Step 3 – runBuild Orchestration

Implement runBuild to follow the standard pattern:

func runBuild(cmd *cobra.Command, args []string) error {
    ctx := cmd.Context()
    if ctx == nil {
        ctx = context.Background()
    }

    // 1. Resolve global flags (config path, verbosity, env placeholder)
    flags, err := ResolveFlags(cmd, nil)
    if err != nil {
        return fmt.Errorf("resolving flags: %w", err)
    }

    // 2. Load config
    cfg, err := config.Load(flags.Config)
    if err != nil {
        return fmt.Errorf("loading config: %w", err)
    }

    // 3. Re-resolve flags with config to validate env
    flags, err = ResolveFlags(cmd, cfg)
    if err != nil {
        return fmt.Errorf("resolving flags: %w", err)
    }

    if flags.Env == "" {
        return fmt.Errorf("environment is required; use --env flag")
    }

    // 4. Logger
    logger := logging.NewLogger(flags.Verbose)

    // 5. Build-specific flags
    versionFlag, _ := cmd.Flags().GetString("version")
    servicesFlag, _ := cmd.Flags().GetString("services")
    push, _ := cmd.Flags().GetBool("push")
    noCache, _ := cmd.Flags().GetBool("no-cache")

    // 6. Resolve version (deploy-style, allowed to use git)
    version, _ := resolveVersion(ctx, versionFlag, logger)

    // 7. Parse services list
    var services []string
    if servicesFlag != "" {
        services = parseServicesList(servicesFlag) // reuse helper if available
    }

    // 8. Generate full plan for env
    planner := core.NewPlanner(cfg)
    plan, err := planner.PlanDeploy(flags.Env)
    if err != nil {
        return fmt.Errorf("generating plan: %w", err)
    }

    // 9. Attach version metadata
    if plan.Metadata == nil {
        plan.Metadata = make(map[string]any)
    }
    plan.Metadata["version"] = version

    // 10. Filter to build operations (and by services)
    buildPlan, err := filterPlanForBuild(plan, services)
    if err != nil {
        return fmt.Errorf("filtering build plan: %w", err)
    }

    // 11. Execute build plan using shared phase executor
    opts := BuildOptions{
        Env:      flags.Env,
        Version:  version,
        Services: services,
        Push:     push,
        NoCache:  noCache,
    }

    return ExecuteBuild(ctx, cfg, buildPlan, opts, logger)
}

Where BuildOptions and ExecuteBuild are defined either in build.go or a companion phases_build.go.

⸻

Step 4 – Plan Filtering for Build

Define a helper that isolates the logic for “which operations are relevant to build?”.

type BuildOptions struct {
    Env      string
    Version  string
    Services []string
    Push     bool
    NoCache  bool
}

func filterPlanForBuild(plan *core.Plan, services []string) (*core.Plan, error) {
    var serviceSet map[string]bool
    if len(services) > 0 {
        serviceSet = make(map[string]bool, len(services))
        for _, svc := range services {
            serviceSet[svc] = true
        }
    }

    filtered := make([]core.Operation, 0, len(plan.Operations))
    for _, op := range plan.Operations {
        if op.Type != core.OpTypeBuild {
            continue
        }

        if serviceSet == nil {
            filtered = append(filtered, op)
            continue
        }

        if operationTouchesServices(op, serviceSet) {
            filtered = append(filtered, op)
        }
    }

    // Optional: validate that all requested services are represented
    if serviceSet != nil {
        if err := ensureRequestedServicesPresent(services, filtered); err != nil {
            return nil, err
        }
    }

    return &core.Plan{
        Environment: plan.Environment,
        Operations:  filtered,
        Metadata:    plan.Metadata,
    }, nil
}

Where:
	•	operationTouchesServices reuses or mirrors the helper you just wrote for CLI_PLAN (checking op.Metadata["services"] or similar).
	•	ensureRequestedServicesPresent enforces spec semantics for unknown services (either error or ignore).

Important: This logic stays in CLI layer; core remains provider-agnostic.

⸻

Step 5 – Execute Build Phases via Common Executor

Leverage CLI_PHASE_EXECUTION_COMMON so build uses the same semantics as other phase-driven commands.

In phases_build.go (or inside build.go if small enough):

func ExecuteBuild(
    ctx context.Context,
    cfg *config.Config,
    plan *core.Plan,
    opts BuildOptions,
    logger logging.Logger,
) error {
    phaseFns := PhaseFns{
        Build: func(ctx context.Context, op core.Operation) error {
            return executeBuildOp(ctx, cfg, op, opts, logger)
        },
        // Other phase functions can be nil or no-ops for CLI_BUILD
    }

    return executePhasesCommon(ctx, plan, phaseFns, logger)
}

And then implement executeBuildOp:

func executeBuildOp(
    ctx context.Context,
    cfg *config.Config,
    op core.Operation,
    opts BuildOptions,
    logger logging.Logger,
) error {
    // 1. Resolve provider (e.g. backend provider)
    if cfg.Backend == nil {
        return fmt.Errorf("build: backend configuration is missing")
    }

    providerID := cfg.Backend.Provider

    backendProv, err := backendproviders.Get(providerID)
    if err != nil {
        return fmt.Errorf("build: resolving backend provider %q: %w", providerID, err)
    }

    // 2. Determine which service/image this op refers to (from op.Metadata)
    //    e.g., op.Metadata["service"], op.Metadata["image"], etc.
    //    Make this logic deterministic & spec-aligned.

    // 3. Build request to provider (pseudo-code):
    req := backend.BuildRequest{
        Env:      opts.Env,
        Version:  opts.Version,
        Service:  op.Metadata["service"].(string),
        Push:     opts.Push,
        NoCache:  opts.NoCache,
    }

    logger.Infof("Building %s (%s)", req.Service, req.Version)

    if err := backendProv.Build(ctx, cfg, req); err != nil {
        return fmt.Errorf("build: provider %q failed for service %q: %w", providerID, req.Service, err)
    }

    return nil
}

Adjust the details (types, fields) to your actual provider interface; the key is:
	•	CLI_BUILD orchestrates.
	•	Providers perform the actual builds (docker, registry etc).

⸻

Step 6 – Register Command in Root

In internal/cli/root.go:
	•	Ensure NewBuildCommand() is added:

root.AddCommand(
    commands.NewBuildCommand(),
    commands.NewDeployCommand(),
    commands.NewPlanCommand(),
    // ...
)

	•	Keep lexicographic order of subcommands as per Agent.md (and your existing pattern).

⸻

6. Test Plan – build_test.go

Follow patterns from deploy_test.go and plan_test.go.

6.1 Command Metadata
	•	TestNewBuildCommand_HasExpectedMetadata
	•	Use == "build".
	•	Short not empty.

6.2 Error Handling
	1.	Config not found
	•	Run command with --config pointing at a non-existent file.
	•	Expect deterministic “config not found” error.
	2.	Missing env
	•	Invoke build without --env.
	•	Expect an error (mapped to exit code 2 at CLI level).
	3.	Unknown env
	•	Config contains only staging.
	•	Run with --env=prod.
	•	Expect deterministic error string.
	4.	Invalid services filter (if spec says so)
	•	Request a service that doesn’t exist in the plan.
	•	filterPlanForBuild should return an error; test asserts it.

6.3 Happy Path Behaviour
	5.	Happy path (all services)
	•	Use an isolated test helper that writes a simple stagecraft.yml with:
	•	A backend provider.
	•	At least one service.
	•	Stub/mock backend provider so:
	•	It records each build request in memory.
	•	It doesn’t hit docker.
	•	Run:

stagecraft build --env=staging --version=test-version


	•	Assert:
	•	Provider called expected number of times.
	•	Each request has Version == "test-version" and correct service.

	6.	Service filtering
	•	Multiple services (api, web).
	•	Run:

stagecraft build --env=staging --services=api --version=test-version


	•	Assert:
	•	Only build ops for api are executed.

	7.	Push / no-cache flags
	•	Run:

stagecraft build --env=staging --version=test-version --push --no-cache


	•	Assert stub provider sees Push=true, NoCache=true.

6.4 Determinism
	8.	Determinism
	•	Run the same command twice in a test.
	•	If build logs are captured, assert the output strings are identical (no timestamps, no random ordering).
	•	If using golden outputs (e.g. summary text), compare against golden file.

6.5 Error Propagation
	9.	Provider failure propagates
	•	Make stub provider return an error for a specific service.
	•	Assert:
	•	ExecuteBuild returns an error wrapped with context.
	•	runBuild surfaces it without swallowing.

⸻

7. Determinism & Side-Effects
	•	CLI_BUILD is allowed to:
	•	Use git via resolveVersion.
	•	Trigger docker/registry operations through providers.
	•	It must still obey global determinism rules:
	•	No random IDs or timestamps in user-visible output.
	•	Stable ordering of lists (services, operations).
	•	All side-effects are through providers, not through CLI core logic.

⸻

8. Finishing the Feature

Once implementation and tests are complete:
	1.	Run:

./scripts/goformat.sh
./scripts/run-all-checks.sh


	2.	Update spec/features.yaml:
	•	CLI_BUILD → done.
	3.	Ensure docs are consistent:
	•	spec/commands/build.md matches actual behaviour.
	•	Help output (stagecraft build --help) is deterministic and sensible.

⸻

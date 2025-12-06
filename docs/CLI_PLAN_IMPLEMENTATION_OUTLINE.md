CLI_PLAN — Feature Plan

Deterministic Release Planning for Stagecraft

⸻

1. Feature Summary

Feature ID: CLI_PLAN
Command: stagecraft plan
Goal: Provide a deterministic, side-effect-free computation of what would happen on a deployment.
Note: Rollback planning is not yet implemented; this command currently focuses on deployment plans.
Output: A fully ordered, validated, provider-aware execution plan including phases, resources, and backend steps.

This feature parallels tools like Terraform’s plan or Ansible’s check mode, giving users and CI the ability to inspect all derived behavior before executing a real deployment.

⸻

2. User Stories

2.1 Developer
	•	As a developer, I want to preview the phases that stagecraft deploy would run so I can catch configuration errors before execution.
	•	As a developer, I want a stable JSON output format so I can integrate Stagecraft planning into CI pipelines.

2.2 DevOps / Platform Engineer
	•	As a platform engineer, I want to confirm what containers, networks, volumes, and environment changes are planned without mutating infrastructure.
	•	As a platform engineer, I want to verify that provider-specific steps (Encore.ts build, Docker Compose build, Cloud provider operations) are valid before running them.

2.3 Automation / CI
	•	As a CI workflow, I want a deterministic, parseable plan so I can block PRs that would cause invalid deployments.

⸻

3. Scope

3.1 Included
	•	Command: stagecraft plan
	•	Phase graph construction (same as deploy, but without executing phases)
	•	Provider-driven planning hooks
(e.g., backend provider may expose PlanSteps() describing expected actions)
	•	Core plan struct generation
	•	Rendering:
	•	Human-readable ANSI output
	•	Machine-readable JSON (--format=json)
	•	Release metadata synthesis without writing actual state
	•	Validation:
	•	config/compose/env correctness
	•	phase ordering
	•	provider readiness
	•	Deterministic behavior: no timestamps, no external side effects

3.2 Excluded (explicit)
	•	No creation of real releases on disk
	•	No network calls to providers (unless providers explicitly support dry-run mode)
	•	No Docker, no builds, no pushes
	•	No state mutation
	•	No filesystem writes except reading config/compose files
	•	Not a substitute for deploy or rollback - this is read-only computation

⸻

4. Architecture

4.1 Implementation Approach

The implementation reuses the existing `core.Plan` structure and extends it with provider plans stored in metadata:

- `core.Plan` is used as-is (no new ExecutionPlan type)
- Provider plans are stored in `plan.Metadata["provider_plans"]` as `map[string]backend.ProviderPlan` (internal representation)
- For JSON output, provider plans are converted to a sorted slice `[]jsonProviderPlan` for deterministic serialization
- Provider types are defined in `pkg/providers/backend/backend.go`:

type ProviderPlan struct {
    Provider string
    Steps    []ProviderStep
}

type ProviderStep struct {
    Name        string
    Description string
}

No executable code is attached — purely a static representation.

⸻

4.2 CLI Flow

planCmd → Load Config → Resolve Provider → Build Phase Graph → Provider.Plan() → Merge Plan → Render Output

4.3 Interaction with Existing Code
	•	phases_common.go is reused for phase enumeration, but never executes functions.
	•	release_id.go may be reused to construct a release ID, but this ID is not written to state.
	•	state.Manager is loaded read-only if necessary for planning.

⸻

5. CLI Specification

5.1 Command

stagecraft plan [flags]

5.2 Flags

Flag	Default	Description
--format	"text"	Output format: "text" or "json"
--env	""	Required: Target environment name
--version	"unknown"	Version to plan for (no git shell-out)
--services	""	Comma-separated list of services to filter
--verbose	false	Reserved for future verbose output

5.3 Exit Codes

Code	Meaning
0	valid plan generated
1	invalid config / provider failure / planning error
Note: Exit code 2 for "plan contains critical errors" is not yet implemented; all errors currently exit with code 1.


⸻

6. Provider Interfaces

6.1 Required Provider Method

Each backend provider must implement:

Plan(ctx context.Context, opts PlanOptions) (ProviderPlan, error)

Where `PlanOptions` contains:
- `Config`: provider-specific configuration
- `ImageTag`: expected image tag that would be built
- `WorkDir`: working directory for the build

The provider must return deterministic, theoretical steps without performing any actual operations.

Example (Docker Compose Provider):

Steps: [
  { Name: "ResolveComposeFiles", Description: "Parsed docker-compose.yml and overrides" },
  { Name: "BuildImages", Description: "Would build 3 images" },
  { Name: "UpServices", Description: "Would start 4 services" },
]

No builds occur — this is purely declarative.

⸻

7. Rendering

7.1 Human-Readable Output (default)

Example:

Plan for release: rel-20250101-120000000

PHASES:
  • prepare       OK
  • build         Would build backend: 3 steps
  • deploy        Would apply docker-compose up
  • post_deploy   Provider reports 1 follow-up hook

PROVIDER (docker-compose):
  - ResolveComposeFiles
  - BuildImages
  - UpServices

No actions executed.

7.2 JSON Output

--format=json returns:

{
  "env": "staging",
  "version": "v1.0.0",
  "phases": [...],
  "provider_plans": [
    {
      "provider": "generic",
      "steps": [
        { "name": "ResolveDockerfile", "description": "Would use Dockerfile: ./Dockerfile" },
        { "name": "ResolveBuildContext", "description": "Would use build context: ." },
        { "name": "BuildImage", "description": "Would build Docker image: myapp:v1.0.0" }
      ]
    }
  ]
}

Note: Provider plans are exposed as a sorted slice (not a map) for deterministic JSON output. The slice is ordered by provider ID (lexicographically), ensuring consistent serialization across runs. Text and JSON formats share the same ordering semantics.


⸻

8. Validation Semantics

A plan is Valid if:
	•	All providers return without error
	•	The phase graph is consistent
	•	Required files exist and parse cleanly
	•	Env vars resolve
	•	No undefined providers
	•	No circular dependencies detected

A plan is Invalid if:
	•	A provider returns an error
	•	Config is malformed
	•	Compose/env/provider files fail parsing
	•	Required fields are missing

**Note**: Exit code 2 for "plan contains critical errors" is not yet implemented in v1. All errors currently exit with code 1. This is a planned enhancement for future versions.

⸻

9. Testing Requirements

9.1 Unit Tests
	•	Plan construction logic
	•	Phase enumeration without execution
	•	Provider Plan() behaviors (mocked)
	•	JSON output stability tests (golden files)

9.2 Integration Tests
	•	stagecraft plan end-to-end using fixtures
	•	error conditions
	•	deterministic release ID snapshots

9.3 Golden Tests
	•	Human output
	•	JSON output

Golden files MUST be timestamp-free.

⸻

10. Implementation Roadmap

Phase 1 — Core API
	•	Reuse existing core.Plan structure
	•	Store provider plans in plan.Metadata["provider_plans"]
	•	Add helpers for phase introspection

Phase 2 — CLI Command
	•	Register command in root
	•	Implement flag parsing
	•	Wire provider resolution + plan aggregation

Phase 3 — Provider Integrations
	•	Docker Compose provider
	•	Encore.ts provider
	•	Future cloud providers

Phase 4 — Renderers
	•	Human output
	•	JSON output
	•	Verbose output mode

Phase 5 — Tests + Golden Files
	•	Full suite
	•	E2E planning tests

⸻

11. Success Criteria

The feature is complete when:
	•	stagecraft plan deterministically produces a full execution plan
	•	JSON output is stable and CI-friendly
	•	Providers expose meaningful plan steps
	•	No side effects occur during planning
	•	All tests pass, all golden files stable
	•	Documentation exists in spec/commands/plan.md

⸻

12. Documentation Requirements
	•	New spec file: spec/commands/plan.md
	•	Update:
	•	spec/features.yaml
	•	docs/commands/
	•	Provider documentation: each provider must document its Plan() semantics

⸻

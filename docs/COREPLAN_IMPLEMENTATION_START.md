
⸻

2. One subtle thing to decide before coding

There’s a tiny tension in the spec that’s worth locking in mentally before you write code:
• In 3.2 --version, the spec says: “uses the same default logic as deploy” – which (today) shells out to git rev-parse
HEAD.
• In 4.2 Outputs, it says: “The command MUST NOT call any external commands (docker, git, SSH).”

You’ve got three options:

1. Allow a tiny exception for git in CLI_PLAN
   • Interpret “no external commands” as “no infra-affecting commands (docker, ssh, cloud CLIs)” and allow git rev-parse
   as a read-only helper.
   • Implementation: reuse resolveVersion from deploy exactly.
2. Make plan stricter than deploy
   • New semantics:
   • If --version is set, use it.
   • If not, do not shell out; just use "unknown" (or some static marker).
   • Implementation: simple helper:

version := versionFlag
if version == "" {
version = "unknown"
}

	•	Plan output is still deterministic; version is just less “smart”.

	3.	Config/env driven only
	•	e.g., accept STAGECRAFT_VERSION env or project.version in config and don’t touch git at all.

If you don’t want to reopen specs right now, the least-friction path is option 1: “git is allowed because it has no side
effects”, and mentally treat the “no external commands” line as “no docker/ssh/cloud calls.”
If you do care about that purity, then adopt option 2 and adjust the version paragraph later.

Either way, your outline still stands; you just change the resolveVersion bit.

⸻

3. Concrete implementation checklist (from here)

Given your outline, here’s the exact “do this in Cursor” sequence I’d follow.

Step 1 – Add the command (skeleton first)

In internal/cli/commands/plan.go:
• Add the command exactly like in the outline:
• Flags:
• --env, -e (required)
• --version, -v
• --services
• --format (default text)
• --verbose, -V
• RunE: runPlan
• Inside runPlan:

1. ResolveFlags(cmd, nil) – load globals
2. config.Load(flags.Config)
3. ResolveFlags(cmd, cfg) – env validation
4. Initialize logger
5. Grab plan flags (version, services, format, verbose)
6. Resolve version (your choice from section 2)
7. Build planner: core.NewPlanner(cfg)
8. Call planner.PlanDeploy(flags.Env)
9. Attach version to plan.Metadata["version"]
10. Apply service filter (for now only --services)
11. Render via renderPlan(...)
12. Map invalid format/filters to exit code 2, plan failures to exit code 3 (via wrapped errors)

Keep it thin – shove the meat into helpers in the same file so tests can exercise them.

Step 2 – Implement filtering helpers

In the same file:
• parseServicesList(string) []string – reuse pattern from build if it exists (trim, split by ,, ignore empties).
• applyFilters(plan *core.Plan, services, roles, hosts, phases []string) *core.Plan
• In v1:
• If len(services) == 0 – return plan as-is.
• Otherwise:
• Build map[string]bool of wanted services.
• Filter plan.Operations:
• Check some combination of:
• op.Metadata["services"] (likely []string)
• Or any other metadata your CORE_PLAN already uses.
• Keep op if it touches at least one service in the set.
• Important: for any op you keep, you may need to also keep its dependencies (via op.Dependencies); otherwise you can
end up with a deploy phase that refers to a missing build phase. In v1 you can take the simpler route:
• If an op has no service metadata, keep it (infra/migration/etc) – that automatically preserves prerequisites.
• Don’t over-complicate services/roles/hosts/phases yet; it’s fine to leave roles/hosts/phases as nil and “future work”.

Step 3 – Implement renderers

Still in plan.go:
• type PlanRenderOptions { Format string; Verbose bool }
• renderPlan(out io.Writer, plan *core.Plan, env, version string, opts PlanRenderOptions) error
• Dispatch to:
• renderPlanText(...)
• renderPlanJSON(...)

For text:
• Header exactly as in spec:

Environment: <env>
Version: <version>
Services: (all|comma list)
Hosts: (all)   // you can keep this stubbed for now

	•	Then Phases: section where each Operation is rendered in stable order:
	•	Decide on ordering:
	•	If CORE_PLAN already produces operations in execution order, you can trust that and only sort when you need lexicographic ordering.
	•	Otherwise, sort by op.ID or some deterministic key.
	•	Within each phase:
	•	Sort services/hosts arrays lexicographically before printing.

For json:
• Define a small, stable projection type instead of dumping the raw struct, e.g.:

type jsonPlan struct {
Env string          `json:"env"`
Version string          `json:"version"`
Phases  []jsonPhase     `json:"phases"`
}

type jsonPhase struct {
ID string            `json:"id"`
Kind string            `json:"kind"`
Services    []string          `json:"services"`
Hosts       []string          `json:"hosts"`
Description string            `json:"description"`
DependsOn   []string          `json:"depends_on"`
Metadata map[string]any    `json:"metadata,omitempty"`
}

	•	Populate these from core.Plan/core.Operation and sort arrays before json.MarshalIndent.

Step 4 – Wire it into root

In internal/cli/root.go:
• Add NewPlanCommand() in the command tree, in lexicographic order (probably around where you already planned: after
migrate, before releases).

⸻

4. Tests to write (before you code)

In internal/cli/commands/plan_test.go, following your existing patterns:

1. Metadata test
   • Use is "plan"
   • Has the expected Short/Long or at least Short.
2. Missing env
   • Invoke command without --env.
   • Expect user error (mapped to exit code 2 / specific error text).
3. Unknown env
   • Use isolated env helper with a config that only has staging.
   • Run with --env=prod.
   • Expect deterministic error string.
4. Happy path (text)
   • Use setupIsolatedStateTestEnv equivalent to set up:
   • A simple config with staging, a couple of services, and hosts.
   • Run stagecraft plan --env=staging.
   • Assert output matches testdata/plan_staging_all.txt golden.
5. Service filter
   • Same setup, run --services=api.
   • Assert golden output only contains phases touching api (but still includes necessary build/migrate ops).
6. JSON format
   • Run with --format=json.
   • Either:
   • Compare to plan_staging_json.json golden, or
   • json.Unmarshal and assert structure fields.
7. Determinism
   • Run the same command twice in the same test and assert the outputs are byte-for-byte identical.
8. Error propagation
   • Inject a plan failure (e.g., via contrived config that makes CORE_PLAN error) and ensure CLI wraps and surfaces it
   properly.

Once those tests exist and fail, you’re green-lit to implement the command until they pass.

⸻

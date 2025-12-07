---
feature: CLI_BUILD
version: v1
status: done
domain: commands
inputs:
  flags:
    - name: --env
      type: string
      default: ""
      description: "Target environment. MUST refer to a defined environment in config."
    - name: -e
      type: string
      default: ""
      description: "Shorthand for --env"
    - name: --version
      type: string
      default: ""
      description: "Override the build image tag/version"
    - name: --push
      type: bool
      default: "false"
      description: "If set, provider MUST push built images to registry after successful build"
    - name: --dry-run
      type: bool
      default: "false"
      description: "Print the resolved build plan and exit without performing any build"
    - name: --services
      type: string
      default: ""
      description: "Limit the build to specific services (comma-separated)"
outputs:
  exit_codes:
    success: 0
    user_error: 1
    plan_failure: 2
    build_failure: 3
    push_failure: 4
    internal_error: 5
---
# CLI_BUILD

`stagecraft build` builds application container images for a given environment using the configured backend provider, without performing a deploy.  

It extracts and stabilizes the build semantics currently embedded inside `CLI_DEPLOY`, making image builds independently runnable, testable, and automatable (CI/CD safe).

This command MUST be deterministic, provider-agnostic, and fully driven by the Stagecraft config specification.

---

## 1. Summary

`stagecraft build` performs the following high-level workflow:

1. Loads and resolves configuration for the target environment.

2. Computes a build plan via the orchestration engine.

3. Filters the plan to include only build-relevant phases.

4. Applies optional service filtering (`--services`).

5. Resolves an image version/tag (via explicit flag or deterministic default).

6. Either:

   - Executes all build phases, OR

   - Performs a dry-run and prints the build plan without executing.

The command MUST NOT execute any deployment, rollout, or runtime phases.

---

## 2. CLI Definition

### 2.1 Usage

```bash
stagecraft build [flags]
```

### 2.2 Required Flags

| Flag | Description |
|------|-------------|
| `--env, -e <name>` | Target environment. MUST refer to a defined environment in config. |

### 2.3 Optional Flags

| Flag | Description |
|------|-------------|
| `--version <tag>` | Override the build image tag/version. If omitted, a deterministic release ID MUST be generated using the same mechanism as CLI_DEPLOY (via the core state manager or injected clock). |
| `--push` | If set, provider MUST push built images to registry after successful build. MUST NOT push on failure. |
| `--dry-run` | Print the resolved build plan and exit without performing any build. MUST NOT call provider build/push methods. |
| `--services <svc1,svc2,...>` | Limit the build to specific services. Filtering MUST occur after plan generation and before execution. Services not in the plan MUST generate an error. |

---

## 3. Behaviour

### 3.1 Config Resolution

1. Load root config.

2. Resolve the environment specified by `--env`.

3. Apply environment inheritance rules as defined in the config spec.

4. Instantiate provider registries.

If the environment does not exist, the command MUST:

- Exit non-zero.

- Emit deterministic error: `invalid environment: <env>`.

### 3.2 Plan Generation

The command MUST generate a full orchestration plan using existing plan APIs (`core.Plan` or equivalent).

The plan MUST include all phases, but CLI_BUILD MUST filter to:

- Only build-related phases defined by backend provider spec.

- Any provider-specific build steps (opaque to core).

No deploy/rollback/runtime phases may be executed.

If plan generation fails:

- Exit non-zero.

- Error MUST wrap underlying cause with context: `build: plan generation failed: %w`.

### 3.3 Service Filtering (`--services`)

If provided:

1. Parse comma-separated list.

2. For each service:

   - Verify the service appears in the build plan.

   - If not found → deterministic error: `unknown service in build plan: <name>`.

3. Filter the list of build phases to include only those associated with the selected services.

Order MUST be lexicographical where order is not semantically required.

### 3.4 Version Resolution

If `--version` is not provided:

- Use deterministic release ID generation (same method as CLI_DEPLOY).

- If time-based tagging is used, time MUST come from the injected deterministic clock (not direct `time.Now()`).

- The version MUST be passed to the backend provider exactly.

Example generated version: `rel-20250422-153015123`.

### 3.5 Dry-Run Mode (`--dry-run`)

When `--dry-run` is set:

1. Build phases MUST be listed but NOT executed.

2. Output MUST include:

   - Environment

   - Selected services (if any)

   - Resolved version/tag

   - Provider ID

   - Ordered list of build tasks with deterministic formatting

3. No provider methods that cause side effects may be called.

Dry-run MUST always exit with code 0 unless plan generation itself fails.

### 3.6 Execution Mode (default)

When not in dry-run:

1. Execute each build phase using the canonical phase execution engine.

2. Failure semantics:

   - On any failure, mark the phase as failed.

   - Abort remaining phases.

   - Exit non-zero.

3. `--push`:

   - If set, push MUST occur after a successful build.

   - If push fails, the command MUST fail.

   - Push MUST NOT occur if build fails.

### 3.7 Output Determinism

Output MUST:

- Use lexicographical ordering for service lists.

- Not include timestamps unless part of the deterministic version generator.

- Not vary by machine, OS, or environment.

---

## 4. Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Successful build or valid dry-run. |
| 1 | User error (invalid flags, environment, unknown services). |
| 2 | Plan generation failure. |
| 3 | Provider build failure. |
| 4 | Provider push failure. |
| 5 | Internal error (unexpected). |

Exit codes MUST be deterministic and MUST NOT overlap with other CLI commands unless defined consistently across the system.

---

## 5. Examples

### 5.1 Build for dev environment

```bash
stagecraft build --env=dev
```

Example output:

```
Building 3 services for environment "dev"
Version: rel-20250422-153015123
Services:
 - api
 - worker
 - frontend
Build completed successfully.
```

### 5.2 Build with explicit version

```bash
stagecraft build --env=prod --version=v2.1.0
```

```
Building 3 services for environment "prod"
Using version: v2.1.0
...
```

### 5.3 Build and push

```bash
stagecraft build --env=prod --push
```

```
Building images...
Pushing images to registry...
Done.
```

### 5.4 Dry-run

```bash
stagecraft build --env=staging --dry-run
```

```
[DRY RUN] Build plan for environment "staging":
Version: rel-20250422-153015123
Provider: backend.generic
Services:
 - api
 - worker
 - frontend
No images will be built or pushed.
```

### 5.5 Build a subset of services

```bash
stagecraft build --env=dev --services=api,worker
```

```
Building services: api, worker
...
```

---

## 6. Determinism Requirements

- MUST NOT call `time.Now()` directly.

- MUST NOT read environment variables except through config resolution.

- MUST NOT use file-system iteration order without sorting.

- MUST NOT rely on provider-specific assumptions beyond interfaces.

- MUST NOT produce machine-dependent logs.

---

## 7. Provider Contract (Informational)

Backend providers MUST implement:

- `Build(ctx, opts)`

- `Push(ctx, opts)` (if supported)

Where opts includes:

- Service name

- Version/tag

- Build context

- Whether push is requested

Providers MUST be treated as opaque.

---

## 8. Completion Criteria

CLI_BUILD is considered done when:

- Spec is complete (this file).

- Tests fully define behaviour.

- Implementation matches the spec.

- Behaviour is deterministic.

- Docs updated.

- `spec/features.yaml` marks CLI_BUILD as done.


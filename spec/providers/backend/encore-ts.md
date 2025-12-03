# Encore.ts Backend Provider

- Feature ID: `PROVIDER_BACKEND_ENCORE`
- Status: draft
- Depends on: `CORE_BACKEND_REGISTRY`, `PROVIDER_BACKEND_INTERFACE`, `CLI_DEV`, `CLI_BUILD`
- Provider ID: `encore-ts`

Encore.ts is the first-class backend framework supported by Stagecraft. Stagecraft integrates Encore.ts through the
generic `BackendProvider` interface and the backend provider
registry.  [oai_citation:0‡backend.md](sediment://file_00000000f2a471fdb64a7542f37f1041)

This spec defines:

- The configuration schema for the Encore.ts backend provider.
- How the provider implements `BackendProvider`.
- How it is used by `stagecraft dev` and
  `stagecraft build`.  [oai_citation:1‡stagecraft-spec.md](sediment://file_00000000dc8471fd8ebe1ea768672ddf)
- Error handling, logging, and testing expectations.

The generic provider (`PROVIDER_BACKEND_GENERIC`) is the baseline, and this spec extends that pattern for
Encore.ts-specific behavior.  [oai_citation:2‡generic.md](sediment://file_000000002bf471fd8f391aeba39b994d)

---

## 1. Goals

The Encore.ts backend provider MUST:

### 1. Integrate Encore.ts as a backend in a way that is:

- Deterministic and testable.
- Configuration-driven via `stagecraft.yml`.
- Aligned with the core `BackendProvider`
  interface.  [oai_citation:3‡backend.md](sediment://file_00000000f2a471fdb64a7542f37f1041)

### 2. Support:

- Local backend development via `stagecraft dev` (Encore dev server).
- Building backend Docker images via
  `stagecraft build`.  [oai_citation:4‡stagecraft-spec.md](sediment://file_00000000dc8471fd8ebe1ea768672ddf)

### 3. Handle Encore-specific concerns:

- Secret syncing via `encore secret set`.
- Local TLS / CA configuration where needed.
- Telemetry and diagnostics configuration.

Non-goals:

- Defining Encore.ts internals beyond what Stagecraft needs.
- Defining generic backend semantics (these belong to core and generic provider
  specs).  [oai_citation:5‡generic.md](sediment://file_000000002bf471fd8f391aeba39b994d)
- Describing remote deployment mechanics (handled by deploy/infra specs).

---

## 2. Relationship to Core BackendProvider

### 2.1 Interface

All backend providers implement the shared `BackendProvider`
interface:  [oai_citation:6‡backend.md](sediment://file_00000000f2a471fdb64a7542f37f1041)

```go
type BackendProvider interface {
ID() string
Dev(ctx context.Context, opts DevOptions) error
BuildDocker(ctx context.Context, opts BuildDockerOptions) (string, error)
}
```

The Encore.ts provider:

* MUST return encore-ts from ID().
* MUST implement Dev and BuildDocker in accordance with this spec.
* MUST be registered via backend.Register(&EncoreTsProvider{}) at init time and imported from the config package so it
  participates in the registry. ￼

### 2.2 Options Structures

Core defines DevOptions / BuildDockerOptions (or equivalent BuildOptions) with at minimum:  ￼

```go
type DevOptions struct {
Env     map[string]string
Listen  string
WorkDir string
Config  any
}

type BuildDockerOptions struct {
ImageTag string
WorkDir  string
Config   any
}
```

The Encore.ts provider MUST:

* Treat opts.Config as the provider-scoped config for encore-ts.
* Treat opts.Env as base environment constructed by core (env file + environment-specific defaults).
* Honor opts.Listen as the requested listen address for the dev server where applicable.
* Honor opts.WorkDir as the effective Encore project root.

⸻

## 3. Configuration Schema

### 3.1 Provider-Scoped Config

Core backend documentation defines provider-scoped config under backend.providers.<provider-id>. ￼

Encore.ts MUST follow this structure:

```yaml
backend:
  provider: encore-ts
  providers:
    encore-ts:
      dev:
        env_file: .env.local                   # required for dev
        listen: "0.0.0.0:4000"                 # required; dev server bind address
        disable_telemetry: true                # optional; disables Encore telemetry
        node_extra_ca_certs: "./.local-infra/certs/mkcert-rootCA.pem"  # optional
        encore_secrets:
          types: [ "dev", "preview", "local" ]   # optional; secret types to sync
          from_env: # optional; env vars to sync via encore secret
            - DOMAIN
            - API_DOMAIN
            - LOGTO_DOMAIN
            - WEB_DOMAIN
            - LOGTO_APP_ID
            - LOGTO_APP_SECRET
            - LOGTO_MANAGEMENT_API_APPLICATION_ID
            - LOGTO_MANAGEMENT_API_APPLICATION_SECRET
            - LOGTO_APP_API_EVENT_WEBHOOK_SIGNING_KEY
            - STRIPE_API_KEY
            - STRIPE_WEBHOOK_SECRET
            - STRIPE_SERVICE_API_KEY
            - STRIPE_API_VERSION
      build:
        workdir: "./backend"                   # optional override; defaults to project root
        image_name: "api"                      # optional; default "api"
        docker_tag_suffix: ""                  # optional; appended to ImageTag (e.g. "-encore")
```

> Note: The top-level backend.dev block in docs/stagecraft-spec.md is conceptually equivalent to
> backend.providers.encore-ts.dev here; this spec adopts the unified provider-scoped pattern from backend provider docs.

### 3.2 Config Semantics

The Encore.ts provider MUST interpret this config as follows:

* dev.env_file:
  * Path to a dotenv-style file relative to project root or absolute.
  * Core MAY parse this and merge into DevOptions.Env, or provider MAY read it given a path; this MUST be stable and
    clarified in implementation notes.
* dev.listen:
  * Address (host:port) for encore run --listen.
* dev.disable_telemetry:
  * When true, provider MUST set the appropriate Encore environment variable (e.g. DISABLE_ENCORE_TELEMETRY=1) before
    spawning the dev server. ￼
* dev.node_extra_ca_certs:
  * Path to mkcert root CA or equivalent; provider MUST set NODE_EXTRA_CA_CERTS when spawning encore run.
* dev.encore_secrets.types:
  * List of Encore secret types to target when syncing (e.g. dev, preview, local).
* dev.encore_secrets.from_env:
  * List of environment variables that provider MUST read from opts.Env and sync via encore secret set. ￼
* build.workdir:
  * Directory to run Encore build commands in; defaults to opts.WorkDir if absent.
* build.image_name:
  * Logical name of the backend image; combined with project.registry + ImageTag to form final image reference. ￼
* build.docker_tag_suffix:
  * Optional suffix appended to ImageTag (e.g. -encore) for clarity.

### 3.3 Validation Rules

On entry to Dev or BuildDocker, the provider MUST:

* Parse opts.Config into a typed config struct.
* Fail early with readable errors when:
  * dev.listen is missing or invalid for Dev.
  * dev.env_file does not exist (if the provider is responsible for reading it).
  * build section is required for BuildDocker but missing necessary fields.
* Support sensible defaults for optional fields where possible (e.g. default build.workdir to opts.WorkDir or ".").

Error messages MUST clearly indicate:

* Which field is wrong or missing.
* The provider ID (encore-ts).
* Whether the error occurred in dev or build mode.

⸻

## 4. Dev Mode Behavior (Dev)

### 4.1 High-Level Contract

stagecraft dev orchestrator uses the backend provider to run the Encore dev server as part of the full local
environment. ￼

Logical behavior for Encore.ts provider in Dev:

1. Validate configuration.
2. Prepare runtime environment (env, secrets, TLS).
3. Spawn encore run dev server with appropriate flags.
4. Stream logs and respect context cancellation.

The provider MUST be responsible for Encore-specific behavior; the core CLI MUST treat it as an opaque backend engine
accessed through BackendProvider.Dev.

### 4.2 Secrets Sync

Based on Stagecraft spec dev sequence:  ￼

* For each name in dev.encore_secrets.from_env:
  * Provider MUST:
    * Look up the value in opts.Env.
    * If value exists and is non-empty:
      * For each type in dev.encore_secrets.types:
        * Run: encore secret set --type <type> <NAME> and pipe the value via stdin.
* If an environment variable is missing:
  * Provider SHOULD log a warning at WARN level but MUST NOT fail Dev solely for missing optional secrets.
* If encore secret set fails:
  * Provider MUST fail Dev and include:
    * Secret name.
    * Secret type.
    * Exit code and truncated stderr.

### 4.3 Environment Preparation

Provider MUST:

* Start from opts.Env as base environment.
* Merge any modifications implied by config:
* Add DISABLE_ENCORE_TELEMETRY=1 when disable_telemetry is true.
* Add NODE_EXTRA_CA_CERTS=<path> when node_extra_ca_certs is set.
* Set PORT or other standard variables if Encore requires them and they are not already present.
* Ensure secrets are NOT echoed in logs.

### 4.4 Dev Command Invocation

The provider MUST invoke Encore dev server using a command equivalent to:  ￼

```shell
encore run --debug --verbose --watch --listen <listen>
```

Where:

*  <listen> is derived from dev.listen (or opts.Listen if present and allowed to override).
* WorkDir is dev.workdir or opts.WorkDir or ".", in that precedence order.

Behavior:

1. Execute the command as a child process with the prepared environment.
2. Stream stdout and stderr back through Stagecraft’s logging subsystem, tagging records with:
  * Provider: encore-ts
  * Operation: dev
  * Feature ID: PROVIDER_BACKEND_ENCORE
3. Respect context:
   * On ctx.Done(), provider MUST terminate the dev process (gracefully if possible) and return.
4. If process exits unexpectedly while ctx is still active:
   * Provider MUST return an error describing:
   * Exit code.
   * A short, truncated tail of the logs.
   * Stagecraft CLI SHOULD surface this prominently to the user.

### 4.5 Readiness and Health

The provider SHOULD:

* Optionally detect readiness by:
* Watching Encore logs for a stable “server listening on …” message, or
* Probing the listen address for HTTP responsiveness.
* If readiness probing is implemented:
* Timeout MUST be configurable (with sensible default).
* On timeout, provider MUST:
* Terminate the dev process.
* Return an error of category DEV_SERVER_FAILED.

If readiness probing is not implemented in v1:

* Provider MUST still start the dev server and stream logs.
* CLI documentation MUST clearly state that readiness is inferred from log output by the user.

⸻

## 5. Build Behavior (BuildDocker)

### 5.1 High-Level Contract

stagecraft build uses the backend provider to construct Docker images for deployment; Stagecraft spec expects Encore
provider to run encore build docker <backendImage:tag>. ￼

The Encore.ts provider MUST:

1. Validate build config.
2. Derive target image reference from:
   * project.registry and logical image name (e.g. api).
   * opts.ImageTag and optional docker_tag_suffix.
3. Invoke appropriate Encore build commands.
4. Return the built image reference on success.

### 5.2 Image Reference Resolution

Given:

* project.registry (from global config).
* build.image_name (provider config; default "api").
* opts.ImageTag (usually git SHA or version).
* build.docker_tag_suffix (optional).

The provider MUST compute an image reference like:

<registry>/<image_name>:<ImageTag><docker_tag_suffix>

Example:

ghcr.io/your-org/platform/api:abc1234
ghcr.io/your-org/platform/api:abc1234-encore

The provider MUST return this exact reference string from BuildDocker on success.

### 5.3 Build Command

The provider MUST invoke Encore build using a command equivalent to:  ￼

encore build docker <IMAGE_REF>

Where:

* <IMAGE_REF> is the fully qualified reference resolved above.
* Command runs in build.workdir or opts.WorkDir, as per config.

Behavior:

1. Execute build with context-aware cancellation.
2. Stream build logs via Stagecraft logging (provider=encore-ts, operation=build).
3. On success:
   * Return <IMAGE_REF> and nil error.
4. On failure:
   * Return an error of category BUILD_FAILED with:
   * Exit code.
   * Truncated build logs.
   * Image ref attempted.

⸻

## 6. Error Handling

The Encore.ts provider MUST map errors into stable categories, similar in spirit to the generic provider but tailored to
Encore:  ￼

* PROVIDER_NOT_AVAILABLE:
* encore binary not found or not executable.
* INVALID_CONFIG:
* Config cannot be parsed or missing required fields (dev.listen, etc.).
* INVALID_PROJECT:
* Encore project structure invalid; e.g. missing Encore app config.
* SECRET_SYNC_FAILED:
* One or more encore secret set invocations failed.
* DEV_SERVER_FAILED:
* Dev server failed to start, crashed early, or failed readiness check.
* BUILD_FAILED:
* encore build docker exited non-zero.

Each error MUST:

* Include provider ID (encore-ts).
* Include operation (dev or build).
* Include a short, user-friendly message.
* Optionally include a “detail” field with truncated logs.

Underlying Encore output SHOULD be preserved in logs, not in the primary error message, to avoid overwhelming the user.

⸻

## 7. Logging and Observability

The provider MUST integrate with Stagecraft’s logging package:  ￼

* Every log record MUST include:
* provider="encore-ts"
* operation="dev" | "build"
* feature="PROVIDER_BACKEND_ENCORE"
* Log levels:
* DEBUG: underlying Encore debug/verbose output, command details.
* INFO: high-level lifecycle events (starting dev, build begun, build succeeded).
* WARN: non-fatal issues (missing optional secrets, non-critical config).
* ERROR: fatal errors causing Dev or BuildDocker to fail.
* Secrets:
* Provider MUST never log secret values or full env dumps.

For long-running dev sessions, provider SHOULD:

* Periodically log health information (e.g. “encore dev still running” every N minutes) at DEBUG level, if cheap.

⸻

## 8. Security and Isolation

The Encore.ts provider MUST:

* Treat provider config as untrusted input:
* Validate before use.
* Avoid shell injection vulnerabilities when constructing commands.
* Restrict command execution to:
* opts.WorkDir or resolved workdir.
* NEVER:
* Print secret values in logs.
* Write secret values to disk outside Encore’s own secret storage.
* Where possible, support running inside containers or restricted environments when Stagecraft core opts into such
  modes.

Any security-relevant assumptions (e.g. relying on local Docker daemon, requiring outbound network for Encore telemetry
if enabled) SHOULD be documented in implementer comments and user-facing docs.

⸻

## 9. Testing

Tests for PROVIDER_BACKEND_ENCORE MUST cover at least:

### 9.1 Config Parsing and Validation

*	Valid minimal config:
*	dev.env_file + dev.listen.
*	Missing required fields:
*	No dev.listen -> INVALID_CONFIG.
*	Nonexistent dev.env_file (if provider reads it) -> clear error.

### 9.2 Secrets Sync

*	With encore_secrets configured and env values present:
*	Provider calls encore secret set for each (type, name).
*	Missing env vars:
*	Provider logs warnings but continues.
*	Simulated secret command failure:
*	Provider returns SECRET_SYNC_FAILED.

### 9.3 Dev Command

*	Happy path:
*	Dev spawns encore run with correct args.
*	Logs are streamed.
*	Process exit with non-zero:
*	Provider returns DEV_SERVER_FAILED with exit code.
*	Context cancellation:
*	Provider terminates process and returns nil or context-related error (depending on design).

### 9.4 Build Command

*	Happy path:
*	BuildDocker runs encore build docker <IMAGE_REF>.
*	Returns correct image reference.
*	Build failure:
*	Returns BUILD_FAILED with error and truncated logs.

### 9.5 Registry and ID

*	Provider registration:
*	Upon importing the Encore provider package, registry contains an entry with ID encore-ts.  ￼
*	ID() returns exactly encore-ts.

Tests SHOULD mock command execution (for encore binary) to avoid external dependencies and to keep CI stable.

⸻

## 10. Edge Cases and Constraints

Known constraints (non-exhaustive):

* Minimum supported Encore.ts version:
* Provider SHOULD verify encore version and warn or error on unsupported ranges.
* Monorepo support:
* WorkDir / build.workdir MUST allow locating an Encore project nested under a monorepo (e.g. ./apps/api).
* Platform support:
* Provider is expected to work on macOS and Linux (aligned with Stagecraft core support). ￼

Where constraints are not yet enforced, they SHOULD be documented as “future work” in code comments and test TODOs.

⸻

## 11. Related Features and Documents

*	CORE_BACKEND_REGISTRY - global registry of backend providers.  ￼
*	PROVIDER_BACKEND_INTERFACE - Go interface for backend providers.  ￼
*	CLI_DEV - stagecraft dev command orchestration.  ￼
*	CLI_BUILD - stagecraft build command orchestration.  ￼
*	PROVIDER_BACKEND_GENERIC - generic backend provider spec; Encore.ts spec should remain consistent with its patterns.  ￼

⸻

## 12. Open Questions / Future Work

These items are explicitly NOT part of the stable contract yet:

* Support for multiple Encore apps (e.g. multiple services) within one repo.
* Additional build outputs (manifests / metadata) for deployment planning.
* Tight integration with Encore’s observability (traces, metrics) exposed via Stagecraft.
* Automated test hooks (e.g. encore test) invoked via provider-level Test operation.

Once any of these become required behaviors, they MUST be promoted into earlier sections with “MUST/SHOULD/MUST NOT”
language and wired to dedicated Feature IDs (for example PROVIDER_BACKEND_ENCORE_TESTS).


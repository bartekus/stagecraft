---
feature: DEV_PROCESS_MGMT
version: v1
status: wip
domain: dev
---

# DEV_PROCESS_MGMT - Dev Process Management

> Feature ID: DEV_PROCESS_MGMT  
> Domain: dev

⸻

## 1. Overview

DEV_PROCESS_MGMT defines how `stagecraft dev` starts and manages the lifecycle of containers for local development using the dev files generated under `.stagecraft/dev`.

It covers:

- Starting the dev stack via Docker Compose.

- Foreground and detached modes.

- Basic teardown behavior on interruption.

- Respect for selected CLI flags.

It does not define how dev files are generated; that is the responsibility of DEV_COMPOSE_INFRA and DEV_TRAEFIK.

⸻

## 2. Behaviour

### 2.1 Dev Files as Source of Truth

DEV_PROCESS_MGMT uses the following file as its sole source of runtime topology:

- `compose.yaml` under the dev directory, normally:

  - `.stagecraft/dev/compose.yaml`

The dev directory is the same directory used by DEV_FILES (dev files slice).

DEV_PROCESS_MGMT does not recompute topology; it only consumes the compose file.

### 2.2 Command Execution

To start the dev stack, DEV_PROCESS_MGMT invokes the Docker Compose CLI with the following base command:

```text
docker compose -f <DEV_DIR>/compose.yaml up
```

All flags and additional arguments are appended deterministically.

If `<DEV_DIR>/compose.yaml` does not exist or cannot be read, the command fails with:

- Exit code: 2 (external provider failure).

- Error message: `dev: compose file not found at <path>` or a more specific error.

### 2.3 Foreground Mode (default)

If `--detach` is not specified:

1. DEV_PROCESS_MGMT runs:

```text
docker compose -f <DEV_DIR>/compose.yaml up [--scale traefik=0]
```

- `--scale traefik=0` is included when `--no-traefik` is set (see 2.4).

2. Standard output and error streams from the compose process are attached to the current process, so users can see logs.

3. The `stagecraft dev` command remains running as long as the compose process is active.

4. On termination:

- If the compose process exits with code 0:

  - `stagecraft dev` exits with code 0.

- If the compose process exits with non zero code:

  - `stagecraft dev` exits with code 2 and includes an error message summarizing the failure.

5. On interruption (for example Ctrl+C):

- DEV_PROCESS_MGMT sends a termination signal to the compose process via context cancellation.

- After the compose process exits, DEV_PROCESS_MGMT issues a teardown command:

```text
docker compose -f <DEV_DIR>/compose.yaml down
```

- Errors from the teardown step are logged but do not mask the original user initiated interruption unless they indicate a critical failure (for example docker not available).

### 2.4 Traefik Control (--no-traefik)

When the CLI flag `--no-traefik` is set:

- DEV_PROCESS_MGMT modifies the compose command by appending `--scale traefik=0` to all `docker compose up` invocations.

- This instructs docker compose to scale the service named `traefik` to zero instances if that service exists.

- If there is no service named `traefik`, docker compose will ignore this scaling option without failing.

DEV_PROCESS_MGMT does not otherwise modify the compose file.

### 2.5 Detached Mode (--detach)

When the CLI flag `--detach` is set:

1. DEV_PROCESS_MGMT runs:

```text
docker compose -f <DEV_DIR>/compose.yaml up -d [--scale traefik=0]
```

2. The command returns when docker compose exits.

3. On success:

- Exit code 0.

- A message is logged indicating that the dev stack is running in the background.

4. On failure:

- Exit code 2.

- Error message indicates that the dev stack could not be started and suggests running without `--detach` to see logs.

DEV_PROCESS_MGMT does not automatically tear down detached stacks; a separate teardown feature will handle that in the future.

### 2.6 Verbose Mode (--verbose)

When the CLI flag `--verbose` is set:

- DEV_PROCESS_MGMT logs at least:

  - The exact docker compose command line before it is executed.

  - The dev directory being used.

  - The location of `compose.yaml`.

- Logging is performed via the internal logger abstraction so that future enhancements can capture or redirect logs without changing behaviour.

Verbose mode does not change the core behaviour or exit codes.

⸻

## 3. CLI Integration

CLI_DEV integrates DEV_PROCESS_MGMT as follows:

1. CLI_DEV:

   - Builds the dev topology.

   - Writes dev files under `.stagecraft/dev` (including `compose.yaml`).

2. CLI_DEV constructs DEV_PROCESS_MGMT options based on CLI flags:

   - DevDir is `.stagecraft/dev`.

   - NoTraefik mirrors `--no-traefik`.

   - Detach mirrors `--detach`.

   - Verbose mirrors `--verbose`.

3. CLI_DEV calls DEV_PROCESS_MGMT to start the dev stack.

4. Any error from DEV_PROCESS_MGMT is wrapped with the `dev:` prefix and surfaced as the CLI exit error.

⸻

## 4. Error Handling and Exit Codes

DEV_PROCESS_MGMT uses the following exit code mapping via CLI_DEV:

| Exit Code | Meaning |
|-----------|---------|
| 0 | Success; dev stack started successfully (foreground or detached). |
| 1 | Invalid input; for example, dev dir path is empty or malformed. |
| 2 | External provider failure; for example docker or docker compose returned non zero. |
| 3 | Internal error; unexpected invariant violation inside DEV_PROCESS_MGMT. |

Examples:

- Missing compose file:

  - Exit code 2.

  - Message: `dev: compose file not found at .stagecraft/dev/compose.yaml`.

- Docker compose not installed:

  - Exit code 2.

  - Message: `dev: docker compose not available; please install Docker and Docker Compose`.

⸻

## 5. Determinism

DEV_PROCESS_MGMT must satisfy the following determinism guarantees:

- For a given `<DEV_DIR>`, CLI flags, and environment with docker compose available:

  - The constructed command line is always identical.

  - The same container set is started on each run, assuming the compose file is unchanged.

- DEV_PROCESS_MGMT does not:

  - Introduce random identifiers.

  - Infer project names based on timestamps or other non deterministic data.

Note that runtime behaviour of containers and logs is inherently time dependent and is outside the scope of this spec.

⸻

## 6. Non Goals

DEV_PROCESS_MGMT explicitly does not:

- Generate or modify dev files.

- Manage mkcert certificate generation or renewal (DEV_MKCERT).

- Modify hosts files or DNS (DEV_HOSTS).

- Inspect or manage container health directly.

- Interact with Kubernetes or non Docker container runtimes.

These responsibilities belong to other features and providers.

⸻

## 7. Testing Requirements

- Unit tests must:

  - Validate command construction for all combinations of Detach, NoTraefik, and Verbose.

  - Validate error propagation when the underlying command fails.

- CLI level tests must:

  - Verify that CLI_DEV passes the correct options to DEV_PROCESS_MGMT based on CLI flags.

  - Verify that errors are wrapped with `dev:` prefix.

Golden tests are not required for DEV_PROCESS_MGMT because output is primarily process side effects and logs, but tests may assert specific log messages for determinism.

⸻

## 8. Lifecycle and Status

- Feature ID: DEV_PROCESS_MGMT

- Initial state in `spec/features.yaml`: `todo` or `wip`.

- State becomes `done` only when:

  - All behaviour in this spec is implemented.

  - Tests and CLI integration are complete.

  - Dev stacks can be started and stopped as described.


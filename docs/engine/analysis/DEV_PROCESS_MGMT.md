# DEV_PROCESS_MGMT Feature Analysis Brief

This document captures the high level motivation, constraints, and success definition for DEV_PROCESS_MGMT.

It is the starting point for the Implementation Outline and Spec.

This brief must be approved before outline work begins.

⸻

## 1. Problem Statement

CLI_DEV currently computes a deterministic dev topology and writes dev files to disk under `.stagecraft/dev`, but it does not start any processes.

Developers still have to:

- Run `docker compose` themselves

- Start and stop Traefik manually (if it is not using the same compose file)

- Manage lifecycle on Ctrl+C

- Coordinate foreground versus detached runs

This gap means `stagecraft dev` is not yet the one stop entrypoint for a complete dev stack. It also makes it harder to use CLI_DEV in CI pipelines or demos where a single command is expected to bring up the stack.

DEV_PROCESS_MGMT fills this gap by providing a deterministic, testable, and spec governed process management layer for dev stacks.

⸻

## 2. Motivation

### Developer Experience

- Single entrypoint: A developer should be able to run `stagecraft dev` and have containers started, logs streaming in the foreground, and everything shut down cleanly when they press Ctrl+C.

- Optional detachment: CI or advanced users should be able to run `stagecraft dev --detach` and get a running environment without blocking the terminal.

- Predictable behavior: The same configuration and flags should always produce the same process start behavior.

### Operational Reliability

- Clean lifecycle: Containers should not be left running accidentally when a foreground dev session ends.

- Explicit detach semantics: Detached mode should be deliberate and visible in logs so users understand that containers will continue running.

- Composability: Process management must build on top of previously generated dev files rather than recomputing ad hoc state.

### CI and Automation

- CI pipelines should be able to:

  - Run `stagecraft dev --detach` to start the stack in the background.

  - Run tests against that stack.

  - Tear the stack down explicitly at the end of the job.

⸻

## 3. Users and User Stories

### Developers

- As a developer, I want `stagecraft dev` to start my backend, frontend, and Traefik containers using the generated dev files so I do not have to run `docker compose` manually.

- As a developer, I want `stagecraft dev` to terminate containers cleanly when I press Ctrl+C, so I do not leave orphaned containers behind.

- As a developer, I want `stagecraft dev --detach` to start the stack and return control to my shell so I can run other commands.

### Platform Engineers

- As a platform engineer, I want process management to be deterministic and driven by dev files, not by hidden logic, so I can reason about what CLI_DEV will do on any machine.

- As a platform engineer, I want to be able to inspect or override the exact commands that are executed, so I can troubleshoot environment specific issues.

### CI / Automation

- As a CI pipeline, I want `stagecraft dev --detach` to start the stack using the generated dev files and exit with a non zero code if startup fails.

- As a CI pipeline, I want a complementary teardown path (future feature) so that dev stacks do not leak across jobs.

⸻

## 4. Success Criteria (v1)

1. `stagecraft dev` uses the previously generated dev files under `.stagecraft/dev` to start containers via Docker Compose.

2. Foreground runs:

   - Start the stack.

   - Stream logs or remain attached until interruption.

   - On Ctrl+C, issue a deterministic teardown for the same compose project.

3. Detached runs:

   - Start the stack using `docker compose up -d` (or equivalent).

   - Return exit code 0 on success, non zero on failure.

4. Flags are respected:

   - `--no-traefik` prevents starting Traefik related services.

   - `--detach` controls foreground versus background behavior.

   - `--verbose` increases logging for process commands and errors.

5. Process management is isolated in a dedicated package under `internal/dev/process` and is fully testable with mocks.

6. No breaking change to existing topology and file generation behavior.

⸻

## 5. Risks and Constraints

### Determinism Constraints

- Commands and arguments must be fully determined by:

  - The dev files on disk.

  - CLI flags.

  - Well defined defaults.

- No environment specific heuristics. For example, no guessing project names based on current directory name unless explicitly specified in the spec.

- Logs are inherently time dependent, but command invocation and exit codes must be deterministic for given inputs.

### External Dependencies

- DEV_PROCESS_MGMT depends on:

  - Docker and Docker Compose being available on the host.

  - Dev files already written by previous slices.

- Misconfigured or missing Docker environments must be surfaced as clear errors with appropriate exit codes.

### Implementation Constraints

- v1 scope is limited to:

  - Docker Compose based startup and teardown using the compose file generated in `.stagecraft/dev/compose.yaml`.

  - Traefik as a service within the same compose stack (if enabled).

- Functions must be testable without actually starting containers:

  - Use an abstraction around `exec.CommandContext` so tests can assert commands without running them.

⸻

## 6. Alternatives Considered

### Alternative 1: Let users run `docker compose` manually

Rejected because it undermines the promise of `stagecraft dev` as the single entrypoint and makes CI integration piecemeal and error prone.

### Alternative 2: Implement a custom container lifecycle manager

Rejected for v1. Docker Compose already provides mature lifecycle control. For dev scenarios the priority is a thin, deterministic wrapper across providers and infrastructure, not a new scheduler.

### Alternative 3: Fold process management into CLI_DEV directly

Rejected to keep CLI_DEV thin and aligned with governance. DEV_PROCESS_MGMT should live in `internal/dev/process` and be consumed by CLI_DEV so it can be tested and evolved independently.

⸻

## 7. Dependencies

Required upstream work:

- CLI_DEV:

  - Implementation outline and spec exist.

  - Topology builder implemented (`internal/dev/topology.go`).

  - Dev files slice complete (`internal/dev/files.go`, `DevFiles`).

- DEV_COMPOSE_INFRA:

  - Compose model generation is implemented and deterministic.

- DEV_TRAEFIK:

  - Traefik config generation is implemented and deterministic.

DEV_PROCESS_MGMT must not assume mkcert or hosts management behavior; those belong to DEV_MKCERT and DEV_HOSTS.

⸻

## 8. Approval

- Author: [To be filled]

- Reviewer: [To be filled]

- Date: [To be filled]

Once approved, the DEV_PROCESS_MGMT Implementation Outline may begin.


# DEV_PROCESS_MGMT Implementation Outline

> This document defines the v1 implementation plan for DEV_PROCESS_MGMT. It translates the feature analysis brief into a concrete, testable, spec aligned delivery plan.

> All details in this outline must be reflected in `spec/dev/process-mgmt.md` before any tests or code are written.

⸻

## 1. Feature Summary

**Feature ID:** DEV_PROCESS_MGMT  
**Domain:** dev

**Goal:**

Provide deterministic process lifecycle management for `stagecraft dev` by starting and stopping containers using the dev files generated under `.stagecraft/dev`.

**v1 Scope:**

- Process management for a single Docker Compose based dev stack.

- Use `.stagecraft/dev/compose.yaml` as the single source of truth.

- Integrate with CLI_DEV via a dedicated runner in `internal/dev/process`.

- Support:

  - Foreground mode (default).

  - Detached mode via `--detach`.

  - Respect `--no-traefik` for Traefik related services.

  - Respect `--verbose` for logging.

**Out of scope for v1:**

- Multi host orchestration.

- Kubernetes or remote dev containers.

- Health checking and auto restart.

- mkcert invocation and hosts file modification (DEV_MKCERT and DEV_HOSTS).

- Complex log multiplexer or UI.

**Future extensions (not implemented in v1):**

- Separate teardown command (for example `stagecraft dev down`).

- Rich health check and readiness probing.

- Native integration with docker APIs instead of shelling out.

- Process supervision for non container dev processes.

⸻

## 2. Problem Definition and Motivation

CLI_DEV builds a dev topology and writes deterministic dev files but stops short of actually starting the dev environment.

DEV_PROCESS_MGMT is the thin bridge from:

- Config and topology

- To running containers and teardown semantics

while preserving Stagecraft principles of determinism and clear boundaries.

This ensures:

- Developers can use `stagecraft dev` as the one command for a running dev stack.

- CI can reliably start and stop dev environments.

- Process management logic remains isolated and testable.

⸻

## 3. User Stories (v1)

### Developer

- As a developer, when I run `stagecraft dev`, I want the containers defined in `.stagecraft/dev/compose.yaml` to start and stay running until I stop them.

- As a developer, when I press Ctrl+C in foreground mode, I want Stagecraft to shut down the dev stack cleanly via Docker Compose.

- As a developer, when I run `stagecraft dev --detach`, I want the dev stack to start in the background and the command to exit with an informative message.

### Platform Engineer

- As a platform engineer, I want to see exactly which `docker compose` commands Stagecraft runs so I can debug issues in different environments.

- As a platform engineer, I want all process commands to be deterministic and based on the dev files generated earlier.

### CI / Automation

- As a CI pipeline, I want `stagecraft dev --detach` to return a non zero exit code if the dev stack cannot be started, so jobs fail early.

- As a CI pipeline, I want to rely on deterministic dev files plus DEV_PROCESS_MGMT rather than reconfiguring Docker Compose separately.

⸻

## 4. Inputs and API Contract

### 4.1 API Surface (v1)

```go
// internal/dev/process/runner.go (new file)

package process

import "context"

// Options captures process management settings.
type Options struct {
    DevDir     string
    NoTraefik  bool
    Detach     bool
    Verbose    bool
}

// Runner manages the dev process lifecycle.
type Runner struct {
    exec ExecCommander
    log  Logger
}

// ExecCommander abstracts command execution for testability.
type ExecCommander interface {
    CommandContext(ctx context.Context, name string, args ...string) Command
}

// Command abstracts running a command.
type Command interface {
    Run() error
    Start() error
    Wait() error
    SetStdout(w Writer)
    SetStderr(w Writer)
}

// Logger is a minimal logging abstraction.
type Logger interface {
    Infof(format string, args ...any)
    Errorf(format string, args ...any)
}

// NewRunner constructs a Runner with default exec and logger.
func NewRunner() *Runner

// Run starts the dev stack using the dev compose file and handles lifecycle
// according to Options. It returns when the stack is ready or has failed to start.
//
// In foreground mode, it blocks until context is cancelled or a fatal error occurs.
// In detached mode, it returns after the `up -d` command has completed.
func (r *Runner) Run(ctx context.Context, opts Options) error
```

Note: ExecCommander, Command, and Logger may be satisfied by thin wrappers over exec.CommandContext and a simple logger, but the interfaces must live in internal/dev/process to allow fixture based tests.

### 4.2 Input Sources

- DevDir: usually `.stagecraft/dev` as used by `dev.WriteFiles`.

- Dev files:

  - `compose.yaml` under DevDir.

- CLI flags:

  - `--detach` → Options.Detach.

  - `--no-traefik` → Options.NoTraefik.

  - `--verbose` → Options.Verbose.

### 4.3 Output

- Running containers started via docker compose commands based on `.stagecraft/dev/compose.yaml`.

- Exit codes and error messages surfaced up to CLI_DEV.

- No additional files written; DEV_PROCESS_MGMT is process only.

⸻

## 5. Behaviour Details

### 5.1 Command Invocation

The baseline command for starting the stack is:

```text
docker compose -f <DevDir>/compose.yaml up
```

- In foreground mode:

  - Use `up` without `-d` so docker compose stays attached and logs stream to stdout and stderr.

- In detached mode:

  - Append `-d` to run containers in the background:

    ```text
    docker compose -f <DevDir>/compose.yaml up -d
    ```

### 5.2 Traefik Respect (--no-traefik)

v1 will not modify the compose file but will pass a deterministic environment variable to compose that can be used in future to conditionally include Traefik.

Options for v1:

- Option A (simple): ignore `--no-traefik` for now, but log a warning that Traefik is still started and wire actual behavior in a later slice.

- Option B (preferred if minimal change is available): allow compose service names that match `traefik` to be excluded by:

  - Using docker compose with `--scale traefik=0` when `NoTraefik` is true.

This outline prefers Option B for v1:

- Foreground: `docker compose -f compose.yaml up --scale traefik=0`.

- Detached: `docker compose -f compose.yaml up -d --scale traefik=0`.

If Traefik is optional or absent, `--scale traefik=0` has no effect and docker compose will still succeed.

### 5.3 Foreground Lifecycle

- Runner starts `docker compose up` in the foreground.

- Stdout and stderr are wired to the current process (or a logger) to show logs.

- If context is cancelled (Ctrl+C):

  - Runner issues a teardown command:

    ```text
    docker compose -f <DevDir>/compose.yaml down
    ```

  - Errors from teardown are logged but should not mask the original cancellation unless they indicate serious issues.

### 5.4 Detached Lifecycle

- Runner executes `docker compose up -d` and returns when command exits.

- Success criteria:

  - Exit code 0, containers started.

- On error:

  - Runner returns an error explaining that the dev stack could not be started and includes a hint to run without `--detach` for more details.

Tear down in detached mode will be handled by a future feature (for example `stagecraft dev down` or a separate DEV_PROCESS_TEARDOWN feature).

⸻

## 6. Determinism and Side Effects

### 6.1 Determinism Rules

- Commands and arguments are constructed deterministically based on:

  - DevDir.

  - Known file path `compose.yaml`.

  - Known service name `traefik` for optional scaling.

  - Flags Detach, NoTraefik, and Verbose.

- No random identifiers, timestamps, or machine specific paths appear in command construction.

### 6.2 Side Effects

- Starts and stops containers via external docker compose binary.

- No modifications to dev files or configuration.

- No network calls beyond what docker compose does on its own.

⸻

## 7. Integration with CLI_DEV

`runDevWithOptions` will be extended to:

1. Build topology and write dev files (existing slice).

2. Construct `process.Options`:

   - DevDir: `.stagecraft/dev`.

   - NoTraefik: from `opts.NoTraefik`.

   - Detach: from `opts.Detach`.

   - Verbose: from `opts.Verbose`.

3. Create a `process.Runner` via `process.NewRunner()`.

4. Call `runner.Run(cmd.Context(), procOpts)` and propagate errors.

Error wrapping should preserve context such as `"dev: start processes: %w"`.

⸻

## 8. Testing Strategy

### Unit Tests

File: `internal/dev/process/runner_test.go`.

- Command construction:

  - Given specific Options, verify the runner builds the correct docker compose command and arguments (foreground and detached).

  - Verify `--scale traefik=0` is included when `NoTraefik` is true.

- Error propagation:

  - Simulate a failing compose command and verify the error is wrapped with a clear message.

- Verbose behavior:

  - When `Verbose` is true, verify that log messages include the exact command line being run.

Tests will use fake implementations of ExecCommander, Command, and Logger to avoid starting real processes.

### Integration Tests (CLI level)

File: `internal/cli/commands/dev_process_test.go` (or similar).

- Use a fake Runner that records options and returns success.

- Verify that `runDevWithOptions`:

  - Calls `WriteFiles` first.

  - Constructs `process.Options` correctly from CLI flags.

  - Propagates errors from the runner.

### E2E (future)

- Once a test fixture exists that uses real docker compose (behind a feature or build tag), add smoke tests that:

  - Run `stagecraft dev --detach` in a controlled environment.

  - Verify docker compose project is up.

  - Tear down with a complementary command.

⸻

## 9. Implementation Plan Checklist

Before coding:

- Analysis brief for DEV_PROCESS_MGMT approved (`docs/engine/analysis/DEV_PROCESS_MGMT.md`).

- This outline approved.

- Spec created or updated (`spec/dev/process-mgmt.md`) to match this outline.

During implementation:

1. Create package structure

   - `internal/dev/process/runner.go`

   - `internal/dev/process/runner_test.go`

   - Define Options, Runner, ExecCommander, Command, Logger.

2. Implement Runner

   - `NewRunner` with default exec and logger.

   - `Run` for:

     - Foreground mode: `docker compose up`.

     - Detached mode: `docker compose up -d`.

     - Optional `--scale traefik=0` logic.

3. Wire into CLI_DEV

   - Update `internal/cli/commands/dev.go`:

     - After `dev.WriteFiles`, construct `process.Options`.

     - Call `process.NewRunner().Run`.

     - Wrap errors with CLI_DEV context.

4. Add tests

   - Unit tests for command construction and error handling.

   - CLI integration tests using fake Runner.

After implementation:

- Update outline and spec if behavior diverges during testing.

- Ensure lifecycle entry in `spec/features.yaml` is updated.

- Run `./scripts/run-all-checks.sh` and coverage checks.

⸻

## 10. Completion Criteria

DEV_PROCESS_MGMT is considered complete when:

- Runner starts containers via docker compose using `.stagecraft/dev/compose.yaml`.

- Foreground mode blocks and exits on Ctrl+C with a clean `compose down` executed.

- Detached mode starts containers and returns a correct exit code.

- `--no-traefik` and `--detach` flags are respected as described.

- `--verbose` causes the runner to log commands and key lifecycle events.

- CLI_DEV uses DEV_PROCESS_MGMT and no longer leaves process management as a TODO.

- Spec `spec/dev/process-mgmt.md` matches the final behavior.

- Feature status for DEV_PROCESS_MGMT is updated to `done` in `spec/features.yaml` once all tests and docs are in place.

---


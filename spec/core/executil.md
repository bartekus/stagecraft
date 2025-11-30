# Core ExecUtil – Process Execution Utilities

- Feature ID: `CORE_EXECUTIL`
- Status: todo

## Goal

Provide utilities for executing external commands and managing processes in a consistent, testable way.

The exec utilities should:
- Execute external commands (Docker, Encore, mkcert, etc.)
- Stream output in real-time
- Handle errors gracefully
- Be mockable for testing
- Support context cancellation

## API Design

### Package: `pkg/executil`

```go
package executil

type Runner interface {
    Run(ctx context.Context, cmd Command) (*Result, error)
    RunStream(ctx context.Context, cmd Command, output io.Writer) error
}

type Command struct {
    Name    string
    Args    []string
    Dir     string
    Env     map[string]string
    Stdin   io.Reader
}

type Result struct {
    ExitCode int
    Stdout   []byte
    Stderr   []byte
}

func NewRunner() Runner
func NewCommand(name string, args ...string) Command
```

### Usage Example

```go
runner := executil.NewRunner()
cmd := executil.NewCommand("docker", "compose", "up", "-d")
result, err := runner.Run(ctx, cmd)
```

## Behavior

### Command Execution

- Execute commands with proper environment setup
- Capture stdout/stderr
- Return exit codes
- Respect context cancellation

### Streaming

- Stream output in real-time for long-running commands
- Support both stdout and stderr streaming
- Handle line buffering appropriately

### Error Handling

- Return structured errors with exit codes
- Distinguish between execution errors and command failures
- Provide helpful error messages

## Non-Goals (initial version)

- Process management (lifecycle, signals) - separate feature
- Complex command pipelines
- Interactive command support

## Tests

See `spec/features.yaml` entry for `CORE_EXECUTIL`:
- `pkg/executil/executil_test.go` – unit tests for:
  - Command execution
  - Output capture
  - Error handling
  - Context cancellation


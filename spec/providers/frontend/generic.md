---
feature: PROVIDER_FRONTEND_GENERIC
version: v1
status: done
domain: providers
inputs:
  flags: []
outputs:
  exit_codes: {}
---
# Generic Frontend Provider

- Feature ID: `PROVIDER_FRONTEND_GENERIC`
- Status: done
- Depends on: `PROVIDER_FRONTEND_INTERFACE`, `CORE_EXECUTIL`

## Goal

Provide a generic, command-based frontend provider that:
- Runs arbitrary commands for frontend development servers
- Detects when the server is ready via pattern matching
- Handles graceful shutdown with configurable signals and timeouts
- Requires no framework-specific knowledge
- Serves as the baseline for frontend-agnostic operation

## Use Cases

The generic provider is useful for:
- Vite-based frontends (Vue, React, Svelte, etc.)
- Next.js frontends
- Remix frontends
- Any frontend that can be run via a command

## Configuration

### Schema

```yaml
frontend:
  provider: generic
  providers:
    generic:
      dev:
        command: ["npm", "run", "dev"]        # required; command to run
        workdir: "./apps/web"                  # optional; defaults to project root
        env:                                   # optional; environment variables
          VITE_API_URL: "http://localhost:4000"
        ready_pattern: "Local:.*http://localhost:5173"  # optional; regex pattern to detect readiness
        shutdown:
          signal: "SIGINT"                     # optional; default: "SIGINT"
          timeout_ms: 10000                    # optional; default: 10000 (10 seconds)
```

### Examples

#### Vite Frontend

```yaml
frontend:
  provider: generic
  providers:
    generic:
      dev:
        command: ["npm", "run", "dev"]
        workdir: "./apps/web"
        env:
          VITE_API_URL: "http://localhost:4000"
        ready_pattern: "Local:.*http://localhost:5173"
        shutdown:
          signal: "SIGINT"
          timeout_ms: 10000
```

#### Next.js Frontend

```yaml
frontend:
  provider: generic
  providers:
    generic:
      dev:
        command: ["npm", "run", "dev"]
        workdir: "./frontend"
        env:
          NEXT_PUBLIC_API_URL: "http://localhost:4000"
        ready_pattern: "Ready in.*ms"
        shutdown:
          signal: "SIGTERM"
          timeout_ms: 5000
```

#### Remix Frontend

```yaml
frontend:
  provider: generic
  providers:
    generic:
      dev:
        command: ["npm", "run", "dev"]
        workdir: "./remix-app"
        env:
          API_URL: "http://localhost:4000"
        ready_pattern: "Remix App Server started"
        shutdown:
          signal: "SIGINT"
          timeout_ms: 10000
```

## Implementation

### Interface Compliance

The generic provider implements `FrontendProvider`:

```go
type GenericProvider struct{}

func (p *GenericProvider) ID() string {
    return "generic"
}

func (p *GenericProvider) Dev(ctx context.Context, opts DevOptions) error {
    // Parse config from opts.Config
    // Execute command with environment variables
    // Stream output and watch for ready_pattern
    // Handle graceful shutdown on context cancellation
}
```

### Config Parsing

The provider unmarshals its config from `opts.Config`:

```go
type Config struct {
    Dev struct {
        Command      []string          `yaml:"command"`
        WorkDir      string            `yaml:"workdir"`
        Env          map[string]string `yaml:"env"`
        ReadyPattern string            `yaml:"ready_pattern"`
        Shutdown     struct {
            Signal    string `yaml:"signal"`
            TimeoutMS int    `yaml:"timeout_ms"`
        } `yaml:"shutdown"`
    } `yaml:"dev"`
}
```

### Dev Mode Behavior

1. Parse config from `opts.Config`
2. Validate that `dev.command` is non-empty
3. Determine working directory (config > opts.WorkDir > ".")
4. Merge environment variables (config.env + opts.Env)
5. Execute command with context cancellation
6. Stream stdout/stderr to user
7. If `ready_pattern` is specified, watch output for pattern match
8. On context cancellation:
   - Send shutdown signal (default: SIGINT)
   - Wait up to timeout_ms (default: 10000ms)
   - If process still running, force kill

### Ready Pattern Detection

If `ready_pattern` is specified:
- Monitor both stdout and stderr for the regex pattern
- Once pattern is found, the Dev() method continues (but process keeps running)
- If pattern is never found and process exits, return error
- Pattern matching is done line-by-line for efficiency

### Shutdown Behavior

On context cancellation:
1. Send the configured shutdown signal (default: SIGINT)
2. Wait up to `timeout_ms` milliseconds for graceful shutdown
3. If process is still running after timeout, send SIGKILL
4. Return error if shutdown fails

If no shutdown config is provided:
- Default signal: SIGINT
- Default timeout: 10000ms (10 seconds)

## Validation

### Required Fields

- `dev.command`: Must be non-empty array

### Optional Fields

- `dev.workdir`: Defaults to opts.WorkDir or "."
- `dev.env`: Merged with opts.Env (opts.Env takes precedence)
- `dev.ready_pattern`: Optional regex pattern
- `dev.shutdown.signal`: Defaults to "SIGINT"
- `dev.shutdown.timeout_ms`: Defaults to 10000

### Error Handling

- Missing `dev.command`: Return clear error message
- Invalid command: Let exec fail with standard error
- Ready pattern not found: Return error if process exits before pattern is detected
- Shutdown failure: Return error if process cannot be terminated

## Testing

Tests should cover:
- Config parsing (valid and invalid)
- Dev mode command execution
- Environment variable merging
- Working directory resolution
- Ready pattern detection (with and without pattern)
- Context cancellation and graceful shutdown
- Force kill after timeout
- Error handling for missing fields

## Comparison with Other Providers

### vs Future Framework-Specific Providers

- **Generic**: Framework-agnostic, requires manual configuration
- **Framework-specific**: Optimized for specific frameworks (e.g., Vite, Next.js) with automatic detection

## Non-Goals (v1)

- Automatic ready pattern detection (v1)
- Build functionality (handled separately)
- Production deployment (handled by build/deploy commands)
- Multiple frontend providers per project (v1)
- Hot reload detection (v1)

## Related Features

- `PROVIDER_FRONTEND_INTERFACE` - FrontendProvider interface
- `CORE_EXECUTIL` - Process execution utilities
- `CLI_DEV` - Development command that uses frontend providers
- `DEV_PROCESS_MGMT` - Process lifecycle management (future)


---
feature: CORE_ENV_RESOLUTION
version: v1
status: done
domain: core
inputs:
  flags: []
outputs:
  exit_codes: {}
---
# Environment Resolution and Context

- Feature ID: `CORE_ENV_RESOLUTION`
- Status: done
- Depends on: `CORE_CONFIG`

## Goal

Provide environment resolution and context management for Stagecraft operations.

Environment resolution enables Stagecraft to:
- Resolve environment-specific settings from `stagecraft.yml`
- Provide environment context to commands and providers
- Handle environment variable interpolation
- Support environment-specific overrides

## Interface

```go
// internal/core/env/env.go

package env

import (
    "context"
    "stagecraft/pkg/config"
)

// Context represents an environment context with resolved settings.
type Context struct {
    // Name is the environment name (e.g., "dev", "staging", "prod")
    Name string

    // Config is the resolved environment configuration
    Config config.EnvironmentConfig

    // EnvFile is the path to the environment file
    EnvFile string

    // Variables are the resolved environment variables
    Variables map[string]string
}

// Resolver resolves environment contexts from configuration.
type Resolver struct {
    cfg     *config.Config
    workDir string
}

// NewResolver creates a new environment resolver.
func NewResolver(cfg *config.Config) *Resolver

// SetWorkDir sets the working directory for resolving relative paths.
func (r *Resolver) SetWorkDir(workDir string)

// Resolve resolves an environment context by name.
// The ctx parameter is reserved for future cancellation/timeout support.
func (r *Resolver) Resolve(ctx context.Context, name string) (*Context, error)

// ResolveFromFlags resolves an environment context from CLI flags.
// It uses envFlag if provided, otherwise defaults to "dev".
// The ctx parameter is reserved for future cancellation/timeout support.
func (r *Resolver) ResolveFromFlags(ctx context.Context, envFlag string) (*Context, error)
```

## Behavior

### Environment Resolution

1. **Lookup**: Find environment in `config.Environments[name]`
2. **Validation**: Ensure environment exists, return `ErrEnvironmentNotFound` if not found
3. **Env File Loading**: Load environment variables from `env_file` path (if specified and exists)
4. **Variable Merging**: Merge variables with precedence (lowest to highest):
   - Env file variables (lowest precedence)
   - System environment variables (highest precedence)
5. **Variable Interpolation**: Apply `${VAR}` interpolation to all values

### Environment Context

The resolved context provides:
- Environment name
- Full environment configuration (as value, not pointer - a snapshot, not a live link)
- Resolved environment variables
- Env file path (absolute if relative path was provided)

### Variable Interpolation

v1 interpolation supports `${VAR}` syntax with the following behavior:
- **Single-level interpolation**: Direct variable references are supported
- **Nested/chained interpolation**: Multi-pass algorithm supports nested expansions (e.g., `${BASE_URL}/v1` where `BASE_URL` itself may contain interpolations)
- **Maximum passes**: Up to 5 passes to prevent infinite loops from circular references
- **Circular references**: If a circular reference is detected (e.g., `A=${B}`, `B=${A}`), interpolation will partially resolve and then stop after the maximum number of passes. The result is best-effort and deterministic.
- **Unknown variables**: If a variable is not found, the `${VAR}` pattern remains unchanged in the value
- **No defaults**: v1 does not support `${VAR:-default}` syntax (deferred to v2)

### Missing Files and Environments

- **Missing env file**: If `env_file` is specified but the file doesn't exist, resolution continues without error (file is optional)
- **Missing environment**: If an environment name is not found in config, `Resolve` returns `ErrEnvironmentNotFound` with available environments listed

### Env File Parser

The env file parser semantics intentionally mirror the encorets provider parser for consistency. It handles:
- Full-line comments (lines starting with `#`)
- Inline comments (outside of quoted strings)
- `export` keyword prefix
- Double-quoted values with escape sequences (`\n`, `\t`, `\"`, `\\`)
- Single-quoted values (no escape processing)
- Empty values (`KEY=` or `KEY=""`)
- Leading/trailing whitespace around keys and values

## Usage Example

```go
import (
    "errors"
    "stagecraft/internal/core/env"
    "stagecraft/pkg/config"
)

// Load config
cfg, err := config.Load("stagecraft.yml")

// Create resolver
resolver := env.NewResolver(cfg)
resolver.SetWorkDir(".")

// Resolve environment
ctx, err := resolver.Resolve(context.Background(), "staging")
if err != nil {
    if errors.Is(err, env.ErrEnvironmentNotFound) {
        // Handle missing environment
    }
    return err
}

// Use context
fmt.Printf("Environment: %s\n", ctx.Name)
fmt.Printf("Env file: %s\n", ctx.EnvFile)
fmt.Printf("DB URL: %s\n", ctx.Variables["DATABASE_URL"])
```

## Config Schema

Environment configuration in `stagecraft.yml`:

```yaml
environments:
  dev:
    env_file: .env.local
    driver: digitalocean
    # ... other settings

  staging:
    env_file: /etc/platform/env
    driver: digitalocean
    # ... other settings
```

The `env_file` field is optional. If specified:
- Relative paths are resolved relative to the working directory (set via `SetWorkDir`)
- Absolute paths are used as-is
- Missing files are handled gracefully (no error)

## Non-Goals (v1)

- Full variable interpolation with defaults (`${VAR:-default}`)
- Remote config loading
- Config file watching/reloading
- Advanced schema evolution
- Variable interpolation in config file itself (only in env file values)

## Related Features

- `CORE_CONFIG` - Config loading that provides environment definitions
- `CLI_GLOBAL_FLAGS` - Global flags that specify target environment
- `CORE_STATE` - State management that uses environment context
- `CLI_DEPLOY` - Deploy command that uses environment resolution

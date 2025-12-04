# Environment Resolution and Context

- Feature ID: `CORE_ENV_RESOLUTION`
- Status: todo
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
    Config *config.Environment

    // EnvFile is the path to the environment file
    EnvFile string

    // Variables are the resolved environment variables
    Variables map[string]string
}

// Resolver resolves environment contexts from configuration.
type Resolver struct {
    cfg *config.Config
}

// NewResolver creates a new environment resolver.
func NewResolver(cfg *config.Config) *Resolver {
    return &Resolver{cfg: cfg}
}

// Resolve resolves an environment context by name.
func (r *Resolver) Resolve(ctx context.Context, name string) (*Context, error) {
    // Implementation:
    // 1. Look up environment in config
    // 2. Load env file if specified
    // 3. Merge environment variables
    // 4. Return resolved context
}

// ResolveFromFlags resolves an environment context from CLI flags.
func (r *Resolver) ResolveFromFlags(ctx context.Context, envFlag string) (*Context, error) {
    // Implementation:
    // 1. Use envFlag or default to "dev"
    // 2. Validate environment exists
    // 3. Call Resolve
}
```

## Behavior

### Environment Resolution

1. **Lookup**: Find environment in `config.Environments[name]`
2. **Validation**: Ensure environment exists, return error if not found
3. **Env File Loading**: Load environment variables from `env_file` path
4. **Variable Merging**: Merge variables with precedence:
   - System environment variables (highest)
   - Env file variables
   - Config defaults (lowest)

### Environment Context

The resolved context provides:
- Environment name
- Full environment configuration
- Resolved environment variables
- Env file path

### Variable Interpolation

Basic variable interpolation support (v1):
- `${VAR}` syntax for environment variables
- Used primarily in migration contexts
- Full interpolation deferred to v2

## Usage Example

```go
import (
    "stagecraft/internal/core/env"
    "stagecraft/pkg/config"
)

// Load config
cfg, err := config.Load("stagecraft.yml")

// Create resolver
resolver := env.NewResolver(cfg)

// Resolve environment
ctx, err := resolver.Resolve(context.Background(), "staging")
if err != nil {
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
    # ... other settings

  staging:
    env_file: /etc/platform/env
    # ... other settings
```

## Non-Goals (v1)

- Full variable interpolation (basic `${VAR}` only)
- Remote config loading
- Config file watching/reloading
- Advanced schema evolution

## Related Features

- `CORE_CONFIG` - Config loading that provides environment definitions
- `CLI_GLOBAL_FLAGS` - Global flags that specify target environment
- `CORE_STATE` - State management that uses environment context
- `CLI_DEPLOY` - Deploy command that uses environment resolution


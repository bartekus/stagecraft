# Backend Providers

Stagecraft uses a provider-based architecture for backend support. This allows Stagecraft to work with any backend framework or technology stack.

## Architecture

### BackendProvider Interface

All backend providers implement the `BackendProvider` interface:

```go
type BackendProvider interface {
    ID() string
    Dev(ctx context.Context, opts DevOptions) error
    BuildDocker(ctx context.Context, opts BuildDockerOptions) (string, error)
}
```

### Provider Registry

Providers register themselves at package initialization:

```go
func init() {
    backend.Register(&MyProvider{})
}
```

The registry is the single source of truth for available providers. Config validation checks against the registry, not hardcoded lists.

## Configuration

### Provider-Scoped Config

Backend configuration uses a provider-scoped structure:

```yaml
backend:
  provider: generic  # Provider ID (must be registered)
  providers:
    generic:          # Provider-specific config
      dev:
        command: ["npm", "run", "dev"]
        workdir: "./backend"
```

### How Config is Passed

1. User specifies `backend.provider: "generic"`
2. Stagecraft validates the provider exists in the registry
3. Stagecraft extracts `backend.providers.generic` as `any`
4. Provider receives this config blob in `DevOptions.Config` or `BuildDockerOptions.Config`
5. Provider unmarshals the config into its own typed struct

## Available Providers

### Generic Provider

The generic provider runs arbitrary commands for development and builds Docker images.

**Config Example:**
```yaml
backend:
  provider: generic
  providers:
    generic:
      dev:
        command: ["npm", "run", "dev"]
        workdir: "./backend"
        env:
          NODE_ENV: development
      build:
        dockerfile: "./backend/Dockerfile"
        context: "./backend"
```

**Use Cases:**
- Node.js (Express, Fastify, etc.)
- Go applications
- Python (Django, FastAPI, etc.)
- Any backend that can be run via command

### Encore.ts Provider

The Encore.ts provider integrates with the Encore framework.

**Config Example:**
```yaml
backend:
  provider: encore-ts
  providers:
    encore-ts:
      dev:
        secrets:
          types: ["dev", "preview", "local"]
          from_env:
            - DOMAIN
            - API_DOMAIN
        entrypoint: "./backend"
        env_from:
          - .env.local
        listen: "0.0.0.0:4000"
```

**Features:**
- Automatic secret syncing via `encore secret set`
- Encore dev server integration
- Docker builds via `encore build docker`

## Creating a Custom Provider

### Step 1: Implement the Interface

```go
package myprovider

import (
    "context"
    "stagecraft/pkg/providers/backend"
)

type MyProvider struct{}

func (p *MyProvider) ID() string {
    return "my-provider"
}

func (p *MyProvider) Dev(ctx context.Context, opts backend.DevOptions) error {
    // Parse opts.Config into your config struct
    // Run your dev server
    return nil
}

func (p *MyProvider) BuildDocker(ctx context.Context, opts backend.BuildDockerOptions) (string, error) {
    // Parse opts.Config into your config struct
    // Build Docker image
    return opts.ImageTag, nil
}
```

### Step 2: Define Config Struct

```go
type Config struct {
    Dev struct {
        Command []string `yaml:"command"`
        WorkDir string   `yaml:"workdir"`
    } `yaml:"dev"`
    
    Build struct {
        Dockerfile string `yaml:"dockerfile"`
    } `yaml:"build"`
}

func (p *MyProvider) parseConfig(cfg any) (*Config, error) {
    data, err := yaml.Marshal(cfg)
    if err != nil {
        return nil, err
    }
    
    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    
    return &config, nil
}
```

### Step 3: Register the Provider

```go
func init() {
    backend.Register(&MyProvider{})
}
```

### Step 4: Import in Config Package

Add to `pkg/config/config.go`:

```go
import (
    _ "stagecraft/internal/providers/backend/myprovider"
)
```

This ensures the provider registers itself before config validation.

## Validation

### Core Validation (Stagecraft)

Stagecraft core validates:
- `backend.provider` is non-empty
- Provider exists in registry
- `backend.providers` map exists
- `backend.providers[backend.provider]` exists

### Provider-Specific Validation

Provider implementations validate their own config:

```go
func (p *MyProvider) Dev(ctx context.Context, opts backend.DevOptions) error {
    cfg, err := p.parseConfig(opts.Config)
    if err != nil {
        return fmt.Errorf("invalid config: %w", err)
    }
    
    // Validate provider-specific fields
    if len(cfg.Dev.Command) == 0 {
        return fmt.Errorf("dev.command is required")
    }
    
    // ... rest of implementation
}
```

## Best Practices

1. **Validate Early**: Parse and validate config at the start of `Dev()` and `BuildDocker()`
2. **Clear Errors**: Provide helpful error messages with context
3. **Document Config**: Document your provider's config schema
4. **Test Thoroughly**: Test config parsing, validation, and execution
5. **Handle Missing Fields**: Use sensible defaults where appropriate

## Related Documentation

- [Backend Provider Registry](../spec/core/backend-registry.md)
- [Provider Config Schema](../spec/core/backend-provider-config.md)
- [Generic Provider Spec](../spec/providers/backend/generic.md)


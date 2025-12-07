---
feature: PROVIDER_FRONTEND_INTERFACE
version: v1
status: done
domain: providers
inputs:
  flags: []
outputs:
  exit_codes: {}
---
# Frontend Provider Interface

- Feature ID: `PROVIDER_FRONTEND_INTERFACE`
- Status: done
- Depends on: `CORE_CONFIG`

## Goal

Define the interface for frontend providers that handle frontend development server execution.

Frontend providers enable Stagecraft to run frontend development servers (e.g., Vite, Next.js, etc.) as part of the local development workflow.

## Interface

```go
// pkg/providers/frontend/frontend.go

package frontend

import "context"

// DevOptions contains options for running a frontend in development mode.
type DevOptions struct {
    // Config is the provider-specific configuration decoded from
    // frontend.providers[providerID] in stagecraft.yml.
    // The provider implementation is responsible for unmarshaling this.
    Config any
    
    // WorkDir is the working directory for the frontend
    WorkDir string
    
    // Env is the environment variables to pass to the dev process
    Env map[string]string
}

// FrontendProvider is the interface that all frontend providers must implement.
type FrontendProvider interface {
    // ID returns the unique identifier for this provider (e.g., "generic", "vite").
    ID() string
    
    // Dev runs the frontend in development mode.
    Dev(ctx context.Context, opts DevOptions) error
}
```

## Registry Pattern

Frontend providers follow the same registry pattern as backend providers:

- Providers register themselves during initialization via `init()`
- Registry provides thread-safe lookup and registration
- Registry IDs are returned in deterministic lexicographic order
- Duplicate registration panics

## Usage Example

```go
import frontendproviders "stagecraft/pkg/providers/frontend"

// Get a provider
provider, err := frontendproviders.Get("generic")
if err != nil {
    return err
}

// Run dev mode
opts := frontendproviders.DevOptions{
    Config:  providerCfg,
    WorkDir: "./frontend",
    Env:     map[string]string{"NODE_ENV": "development"},
}

err = provider.Dev(ctx, opts)
```

## Provider Registration

Providers register themselves during initialization:

```go
// internal/providers/frontend/generic/generic.go
func init() {
    frontend.Register(&GenericProvider{})
}
```

## Config Schema

Frontend provider configuration in `stagecraft.yml`:

```yaml
frontend:
  provider: generic
  providers:
    generic:
      dev:
        command: ["npm", "run", "dev"]
        workdir: "./frontend"
        env:
          NODE_ENV: development
```

## Non-Goals (v1)

- Build functionality (frontend builds handled separately)
- Production deployment (handled by build/deploy commands)
- Multiple frontend providers per project

## Related Features

- `PROVIDER_FRONTEND_GENERIC` - Generic command-based frontend provider
- `CLI_DEV` - Dev command that uses frontend providers
- `CORE_CONFIG` - Config system that validates frontend provider config


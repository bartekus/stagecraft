---
feature: CORE_BACKEND_REGISTRY
version: v1
status: done
domain: core
inputs:
  flags: []
outputs:
  exit_codes: {}
---
# Backend Provider Registry

- Feature ID: `CORE_BACKEND_REGISTRY`
- Status: todo
- Depends on: `CORE_CONFIG`

## Goal

Provide a registry-based system for backend providers that:
- Eliminates hardcoded provider lists in validation
- Enables dynamic provider registration
- Supports extensibility without core changes
- Maintains thread-safe provider lookup

## Architecture

### Registry Pattern

Backend providers are registered at runtime and selected by ID from configuration.
The registry is the single source of truth for available providers.

### Interface

```go
// pkg/providers/backend/backend.go

package backend

import "context"

// DevOptions contains options for running a backend in development mode.
type DevOptions struct {
    // Config is the provider-specific configuration decoded from
    // backend.providers[providerID] in stagecraft.yml.
    // The provider implementation is responsible for unmarshaling this.
    Config any
    
    // WorkDir is the working directory for the backend
    WorkDir string
    
    // Env is the environment variables to pass to the dev process
    Env map[string]string
}

// BuildDockerOptions contains options for building a Docker image.
type BuildDockerOptions struct {
    // Config is the provider-specific configuration
    Config any
    
    // ImageTag is the full image tag (e.g., "ghcr.io/org/app:tag")
    ImageTag string
    
    // WorkDir is the working directory for the build
    WorkDir string
}

// BackendProvider is the interface that all backend providers must implement.
type BackendProvider interface {
    // ID returns the unique identifier for this provider (e.g., "encore-ts", "generic").
    ID() string
    
    // Dev runs the backend in development mode.
    Dev(ctx context.Context, opts DevOptions) error
    
    // BuildDocker builds a Docker image for the backend.
    BuildDocker(ctx context.Context, opts BuildDockerOptions) (string, error)
}
```

### Registry Implementation

```go
// pkg/providers/backend/registry.go

package backend

import (
    "fmt"
    "sync"
)

// Registry manages backend provider registration and lookup.
type Registry struct {
    mu        sync.RWMutex
    providers map[string]BackendProvider
}

// NewRegistry creates a new empty registry.
func NewRegistry() *Registry {
    return &Registry{
        providers: make(map[string]BackendProvider),
    }
}

// Register registers a backend provider.
// Panics if the provider ID is empty or already registered.
func (r *Registry) Register(p BackendProvider) {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    id := p.ID()
    if id == "" {
        panic("backend provider registration: empty ID")
    }
    if _, exists := r.providers[id]; exists {
        panic(fmt.Sprintf("backend provider registration: duplicate ID %q", id))
    }
    
    r.providers[id] = p
}

// Get retrieves a provider by ID.
// Returns an error if the provider is not found.
func (r *Registry) Get(id string) (BackendProvider, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    p, ok := r.providers[id]
    if !ok {
        return nil, fmt.Errorf("unknown backend provider %q", id)
    }
    return p, nil
}

// Has checks if a provider with the given ID is registered.
func (r *Registry) Has(id string) bool {
    r.mu.RLock()
    defer r.mu.RUnlock()
    _, ok := r.providers[id]
    return ok
}

// IDs returns all registered provider IDs.
func (r *Registry) IDs() []string {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    ids := make([]string, 0, len(r.providers))
    for id := range r.providers {
        ids = append(ids, id)
    }
    return ids
}

// DefaultRegistry is the global default registry.
var DefaultRegistry = NewRegistry()

// Register registers a provider in the default registry.
func Register(p BackendProvider) {
    DefaultRegistry.Register(p)
}

// Get retrieves a provider from the default registry.
func Get(id string) (BackendProvider, error) {
    return DefaultRegistry.Get(id)
}

// Has checks if a provider exists in the default registry.
func Has(id string) bool {
    return DefaultRegistry.Has(id)
}
```

## Usage in Config Validation

Instead of hardcoded validation:

```go
// ❌ OLD: Hardcoded list
if cfg.Backend.Provider != "encore-ts" && cfg.Backend.Provider != "generic" {
    return errors.New("invalid provider")
}

// ✅ NEW: Registry-based
import backendproviders "stagecraft/pkg/providers/backend"

func validateBackend(cfg *Config) error {
    if cfg.Backend.Provider == "" {
        return fmt.Errorf("backend.provider is required")
    }
    
    if !backendproviders.Has(cfg.Backend.Provider) {
        return fmt.Errorf(
            "unknown backend provider %q; available providers: %v",
            cfg.Backend.Provider,
            backendproviders.DefaultRegistry.IDs(),
        )
    }
    
    if cfg.Backend.Providers == nil {
        return fmt.Errorf("backend.providers is required")
    }
    
    if _, ok := cfg.Backend.Providers[cfg.Backend.Provider]; !ok {
        return fmt.Errorf(
            "backend.providers.%s is missing; provider-specific config is required",
            cfg.Backend.Provider,
        )
    }
    
    return nil
}
```

## Provider Registration

Providers register themselves during initialization:

```go
// internal/providers/backend/encorets/encorets.go
func init() {
    backend.Register(&EncoreTsProvider{})
}

// internal/providers/backend/generic/generic.go
func init() {
    backend.Register(&GenericProvider{})
}
```

## Thread Safety

The registry uses `sync.RWMutex` for thread-safe concurrent access:
- Multiple readers can access the registry simultaneously
- Writers (Register) acquire exclusive lock
- All operations are protected by appropriate locks

## Testing

Registry tests should cover:
- Registration of providers
- Duplicate registration panics
- Retrieval of registered providers
- Error handling for unknown providers
- Thread-safety under concurrent access
- IDs() returns all registered IDs

## Non-Goals

- Provider discovery from external sources (v1)
- Dynamic provider loading from plugins (v1)
- Provider versioning (v1)

## Related Features

- `CORE_BACKEND_PROVIDER_CONFIG_SCHEMA` - Provider-specific config structure
- `PROVIDER_BACKEND_INTERFACE` - BackendProvider interface definition
- `PROVIDER_BACKEND_ENCORE` - Encore.ts provider implementation
- `PROVIDER_BACKEND_GENERIC` - Generic command-based provider


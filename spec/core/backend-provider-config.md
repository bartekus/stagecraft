# Backend Provider Configuration Schema

- Feature ID: `CORE_BACKEND_PROVIDER_CONFIG_SCHEMA`
- Status: todo
- Depends on: `CORE_CONFIG`, `CORE_BACKEND_REGISTRY`

## Goal

Make backend configuration backend-agnostic by:
- Moving provider-specific settings under `backend.providers.<id>`
- Treating provider config as opaque to Stagecraft core
- Allowing new providers without schema changes
- Eliminating hardcoded provider-specific fields at the top level

## Schema

### Structure

```yaml
backend:
  provider: <string>            # required; must match a registered backend provider ID
  providers:                     # required; object keyed by provider ID
    <provider-id>:
      # provider-specific configuration
      # interpreted by the BackendProvider implementation
      # Stagecraft core treats this as opaque (map[string]any)
```

### Rules

1. `backend.provider` is required and must be a registered provider ID
2. `backend.providers` is required and must be a map
3. `backend.providers[backend.provider]` must exist
4. Stagecraft core does not validate provider-specific fields
5. Provider implementations are responsible for validating their own config

## Examples

### Encore.ts Provider

```yaml
backend:
  provider: encore-ts
  providers:
    encore-ts:
      dev:
        env_file: .env.local
        listen: "0.0.0.0:4000"
        disable_telemetry: true
        node_extra_ca_certs: "./.local-infra/certs/mkcert-rootCA.pem"
        secrets:
          types: ["dev", "preview", "local"]
          from_env:
            - DOMAIN
            - API_DOMAIN
            - LOGTO_DOMAIN
      build:
        # Encore-specific build options
```

### Generic Command-Based Provider

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
          PORT: "4000"
      build:
        dockerfile: "./backend/Dockerfile"
        context: "./backend"
```

### Go Backend Provider (Future Example)

```yaml
backend:
  provider: go
  providers:
    go:
      dev:
        command: ["go", "run", "main.go"]
        workdir: "./cmd/api"
        env:
          PORT: "4000"
      build:
        dockerfile: "./cmd/api/Dockerfile"
        context: "."
```

## Config Struct

```go
// pkg/config/config.go

type BackendConfig struct {
    // Provider is the selected provider ID
    Provider string `yaml:"provider"`
    
    // Providers is a map of provider ID to provider-specific config
    // Each provider implementation unmarshals its own config
    Providers map[string]any `yaml:"providers"`
}

// GetProviderConfig returns the config for the selected provider.
func (c *BackendConfig) GetProviderConfig() (any, error) {
    if c.Provider == "" {
        return nil, fmt.Errorf("backend.provider is required")
    }
    
    if c.Providers == nil {
        return nil, fmt.Errorf("backend.providers is required")
    }
    
    cfg, ok := c.Providers[c.Provider]
    if !ok {
        return nil, fmt.Errorf(
            "backend.providers.%s is missing; provider-specific config is required",
            c.Provider,
        )
    }
    
    return cfg, nil
}
```

## Validation

### Core Validation (Stagecraft)

```go
func validateBackend(cfg *BackendConfig) error {
    // 1. Provider ID is required
    if cfg.Provider == "" {
        return fmt.Errorf("backend.provider is required")
    }
    
    // 2. Provider must be registered
    if !backendproviders.Has(cfg.Provider) {
        return fmt.Errorf(
            "unknown backend provider %q; available providers: %v",
            cfg.Provider,
            backendproviders.DefaultRegistry.IDs(),
        )
    }
    
    // 3. Providers map is required
    if cfg.Providers == nil {
        return fmt.Errorf("backend.providers is required")
    }
    
    // 4. Selected provider must have config
    if _, ok := cfg.Providers[cfg.Provider]; !ok {
        return fmt.Errorf(
            "backend.providers.%s is missing; provider-specific config is required",
            cfg.Provider,
        )
    }
    
    return nil
}
```

### Provider-Specific Validation

Provider implementations validate their own config:

```go
// internal/providers/backend/encorets/config.go

type EncoreTsConfig struct {
    Dev struct {
        EnvFile           string   `yaml:"env_file"`
        Listen            string   `yaml:"listen"`
        DisableTelemetry bool     `yaml:"disable_telemetry"`
        NodeExtraCACerts string   `yaml:"node_extra_ca_certs"`
        Secrets           struct {
            Types   []string `yaml:"types"`
            FromEnv []string `yaml:"from_env"`
        } `yaml:"secrets"`
    } `yaml:"dev"`
}

func (p *EncoreTsProvider) ValidateConfig(cfg any) error {
    // Unmarshal provider config
    data, err := yaml.Marshal(cfg)
    if err != nil {
        return fmt.Errorf("marshaling config: %w", err)
    }
    
    var encoreCfg EncoreTsConfig
    if err := yaml.Unmarshal(data, &encoreCfg); err != nil {
        return fmt.Errorf("invalid encore-ts config: %w", err)
    }
    
    // Validate Encore-specific fields
    if encoreCfg.Dev.EnvFile == "" {
        return fmt.Errorf("encore-ts.dev.env_file is required")
    }
    
    // ... more validation
    
    return nil
}
```

## Migration from Old Schema

### Old Schema (Encore-centric)

```yaml
backend:
  provider: encore-ts
  dev:
    encore_secrets:  # ❌ Encore-specific at top level
      types: ["dev", "preview", "local"]
      from_env:
        - DOMAIN
```

### New Schema (Provider-scoped)

```yaml
backend:
  provider: encore-ts
  providers:
    encore-ts:
      dev:
        secrets:  # ✅ Provider-specific under providers.encore-ts
          types: ["dev", "preview", "local"]
          from_env:
            - DOMAIN
```

## Benefits

1. **Agnostic Core**: Stagecraft core doesn't need to know about Encore-specific fields
2. **Extensibility**: New providers can be added without changing core schema
3. **Clear Ownership**: Provider implementations own their config validation
4. **Multiple Providers**: Can define config for multiple providers (useful for testing/migration)

## Non-Goals

- Provider config inheritance (v1)
- Provider config validation in core (delegated to providers)
- Provider config schema generation (v1)

## Related Features

- `CORE_BACKEND_REGISTRY` - Provider registry system
- `PROVIDER_BACKEND_INTERFACE` - BackendProvider interface
- `PROVIDER_BACKEND_ENCORE` - Encore.ts implementation
- `PROVIDER_BACKEND_GENERIC` - Generic provider implementation


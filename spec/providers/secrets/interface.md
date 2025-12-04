# Secrets Provider Interface

- Feature ID: `PROVIDER_SECRETS_INTERFACE`
- Status: todo
- Depends on: `CORE_CONFIG`

## Goal

Define the interface for secrets providers that handle secrets storage and synchronization.

Secrets providers enable Stagecraft to sync secrets between different sources (e.g., env files, Encore dev secrets, remote secret stores).

## Interface

```go
// pkg/providers/secrets/secrets.go

package secrets

import "context"

// SyncOptions contains options for syncing secrets.
type SyncOptions struct {
	// Config is the provider-specific configuration decoded from
	// secrets.providers[providerID] in stagecraft.yml.
	// The provider implementation is responsible for unmarshaling this.
	Config any

	// Source is the source environment or location (e.g., "dev", ".env.local")
	Source string

	// Target is the target environment or location (e.g., "staging", "encore")
	Target string

	// Keys are the specific secret keys to sync (empty means sync all)
	Keys []string
}

// SecretsProvider is the interface that all secrets providers must implement.
//
//nolint:revive // SecretsProvider is the preferred name for clarity
type SecretsProvider interface {
	// ID returns the unique identifier for this provider (e.g., "envfile", "encore").
	ID() string

	// Sync syncs secrets from source to target.
	Sync(ctx context.Context, opts SyncOptions) error
}
```

## Registry Pattern

Secrets providers follow the same registry pattern as other providers:

- Providers register themselves during initialization via `init()`
- Registry provides thread-safe lookup and registration
- Registry IDs are returned in deterministic lexicographic order
- Duplicate registration panics

## Registry Contract

All Secrets provider registries follow a unified contract:

1. **Thread Safety**  

   All registry methods use an internal RWMutex to guarantee safe concurrent registration and lookups.

2. **Deterministic Ordering**  

   - `IDs()` returns provider IDs in lexicographic order.  

   - `List()` returns provider instances sorted lexicographically by their `ID()`.

3. **Duplicate & Empty ID Prevention**  

   - Registering a provider with an empty ID panics with `ErrEmptyProviderID`.  

   - Registering a provider with an already-registered ID panics with `ErrDuplicateProvider`.

4. **Error Semantics**  

   - `Get(id)` returns the matching provider or an error.  

   - When no provider exists for the given ID, `Get` returns an error that wraps `ErrUnknownProvider` and includes the ID.

5. **Panic Messages**  

   All panic messages are prefixed with `<package>.Registry.Register` to make stack traces searchable and self-describing.

6. **Instrumentation Hooks**  

   Registries expose two optional hooks:

   - `OnProviderRegistered(kind, id string)` – called after a provider is successfully registered.  

   - `OnProviderLookup(kind, id string, found bool)` – called on each `Get`, indicating lookup success or failure.

### Error Types

- `ErrUnknownProvider`: Base error returned (wrapped) whenever `Get()` is called with an unknown provider ID.

- `ErrDuplicateProvider`: Base error used when attempting to register a provider with a duplicate ID.

- `ErrEmptyProviderID`: Base error used when attempting to register a provider with an empty ID.

## Usage Example

```go
import secretsproviders "stagecraft/pkg/providers/secrets"

// Get a provider
provider, err := secretsproviders.Get("envfile")
if err != nil {
    return err
}

// Sync secrets
opts := secretsproviders.SyncOptions{
    Config: providerCfg,
    Source: ".env.local",
    Target: ".env.staging",
    Keys:   []string{"DATABASE_URL", "API_KEY"},
}
err = provider.Sync(ctx, opts)
```

## Provider Registration

Providers register themselves during initialization:

```go
// internal/providers/secrets/envfile/envfile.go
func init() {
    secrets.Register(&EnvFileProvider{})
}
```

## Config Schema

Secrets provider configuration in `stagecraft.yml`:

```yaml
secrets:
  provider: envfile
  providers:
    envfile:
      source_file: .env.local
      target_file: .env.staging
```

## Non-Goals (v1)

- Secret encryption at rest (handled by secret store)
- Secret rotation
- Secret versioning
- Multiple secrets providers per project

## Related Features

- `PROVIDER_SECRETS_ENVFILE` - Env file SecretsProvider implementation
- `PROVIDER_SECRETS_ENCORE` - Encore dev secrets SecretsProvider implementation
- `CLI_SECRETS_SYNC` - Secrets sync command that uses secrets providers
- `CORE_CONFIG` - Config system that validates secrets provider config


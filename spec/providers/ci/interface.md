---
feature: PROVIDER_CI_INTERFACE
version: v1
status: done
domain: providers
inputs:
  flags: []
outputs:
  exit_codes: {}
---
# CI Provider Interface

- Feature ID: `PROVIDER_CI_INTERFACE`
- Status: todo
- Depends on: `CORE_CONFIG`

## Goal

Define the interface for CI providers that handle CI/CD integration.

CI providers enable Stagecraft to initialize CI pipelines (e.g., GitHub Actions) and trigger CI runs from the CLI.

## Interface

```go
// pkg/providers/ci/ci.go

package ci

import "context"

// InitOptions contains options for initializing CI pipelines.
type InitOptions struct {
	// Config is the provider-specific configuration decoded from
	// ci.providers[providerID] in stagecraft.yml.
	// The provider implementation is responsible for unmarshaling this.
	Config any

	// WorkDir is the working directory (typically repository root)
	WorkDir string
}

// TriggerOptions contains options for triggering a CI run.
type TriggerOptions struct {
	// Config is the provider-specific configuration
	Config any

	// Environment is the environment to deploy to (e.g., "staging", "prod")
	Environment string

	// Version is the version to deploy (e.g., "v1.2.3" or git SHA)
	Version string
}

// CIProvider is the interface that all CI providers must implement.
//
//nolint:revive // CIProvider is the preferred name for clarity
type CIProvider interface {
	// ID returns the unique identifier for this provider (e.g., "github", "gitlab").
	ID() string

	// Init initializes CI pipelines in the repository.
	// This typically creates workflow files (e.g., .github/workflows/deploy.yml).
	Init(ctx context.Context, opts InitOptions) error

	// Trigger triggers a CI run for the given environment and version.
	Trigger(ctx context.Context, opts TriggerOptions) error
}
```

## Registry Pattern

CI providers follow the same registry pattern as other providers:

- Providers register themselves during initialization via `init()`
- Registry provides thread-safe lookup and registration
- Registry IDs are returned in deterministic lexicographic order
- Duplicate registration panics

## Registry Contract

All CI provider registries follow a unified contract:

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
import ciproviders "stagecraft/pkg/providers/ci"

// Get a provider
provider, err := ciproviders.Get("github")
if err != nil {
    return err
}

// Initialize CI pipelines
initOpts := ciproviders.InitOptions{
    Config:  providerCfg,
    WorkDir: ".",
}
err = provider.Init(ctx, initOpts)

// Trigger CI run
triggerOpts := ciproviders.TriggerOptions{
    Config:      providerCfg,
    Environment: "staging",
    Version:     "v1.2.3",
}
err = provider.Trigger(ctx, triggerOpts)
```

## Provider Registration

Providers register themselves during initialization:

```go
// internal/providers/ci/github/github.go
func init() {
    ci.Register(&GitHubProvider{})
}
```

## Config Schema

CI provider configuration in `stagecraft.yml`:

```yaml
ci:
  provider: github
  providers:
    github:
      workflow_file: .github/workflows/deploy.yml
      token_env: GITHUB_TOKEN
```

## Non-Goals (v1)

- CI pipeline execution (handled by CI platform)
- CI status monitoring
- Multiple CI providers per project

## Related Features

- `PROVIDER_CI_GITHUB` - GitHub Actions CIProvider implementation
- `CLI_CI_INIT` - CI init command that uses CI providers
- `CLI_CI_RUN` - CI run command that uses CI providers
- `CORE_CONFIG` - Config system that validates CI provider config


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


# Cloud Provider Interface

- Feature ID: `PROVIDER_CLOUD_INTERFACE`
- Status: todo
- Depends on: `CORE_CONFIG`

## Goal

Define the interface for cloud providers that handle infrastructure provisioning and management.

Cloud providers enable Stagecraft to provision and manage infrastructure (e.g., DigitalOcean, AWS, GCP) for deployment environments.

## Interface

```go
// pkg/providers/cloud/cloud.go

package cloud

import "context"

// HostSpec describes a host to be created or deleted.
type HostSpec struct {
	// Name is the hostname for the host
	Name string

	// Role is the role of the host (e.g., "gateway", "app", "db", "cache")
	Role string

	// Size is the instance size (e.g., "s-2vcpu-4gb" for DigitalOcean)
	Size string

	// Region is the region where the host should be created (e.g., "nyc1")
	Region string
}

// InfraPlan describes the infrastructure changes to be made.
type InfraPlan struct {
	// ToCreate are the hosts that should be created
	ToCreate []HostSpec

	// ToDelete are the hosts that should be deleted
	ToDelete []HostSpec
}

// PlanOptions contains options for planning infrastructure changes.
type PlanOptions struct {
	// Config is the provider-specific configuration decoded from
	// cloud.providers[providerID] in stagecraft.yml.
	// The provider implementation is responsible for unmarshaling this.
	Config any

	// Environment is the environment name (e.g., "staging", "prod")
	Environment string
}

// ApplyOptions contains options for applying infrastructure changes.
type ApplyOptions struct {
	// Config is the provider-specific configuration
	Config any

	// Plan is the infrastructure plan to apply
	Plan InfraPlan
}

// CloudProvider is the interface that all cloud providers must implement.
//
//nolint:revive // CloudProvider is the preferred name for clarity
type CloudProvider interface {
	// ID returns the unique identifier for this provider (e.g., "digitalocean", "aws").
	ID() string

	// Plan generates an infrastructure plan for the given environment.
	// This is a dry-run operation that does not modify infrastructure.
	Plan(ctx context.Context, opts PlanOptions) (InfraPlan, error)

	// Apply applies the given infrastructure plan, creating and deleting hosts as needed.
	Apply(ctx context.Context, opts ApplyOptions) error
}
```

## Registry Pattern

Cloud providers follow the same registry pattern as other providers:

- Providers register themselves during initialization via `init()`
- Registry provides thread-safe lookup and registration
- Registry IDs are returned in deterministic lexicographic order
- Duplicate registration panics

## Usage Example

```go
import cloudproviders "stagecraft/pkg/providers/cloud"

// Get a provider
provider, err := cloudproviders.Get("digitalocean")
if err != nil {
    return err
}

// Plan infrastructure changes
planOpts := cloudproviders.PlanOptions{
    Config:      providerCfg,
    Environment: "staging",
}
plan, err := provider.Plan(ctx, planOpts)

// Apply plan
applyOpts := cloudproviders.ApplyOptions{
    Config: providerCfg,
    Plan:   plan,
}
err = provider.Apply(ctx, applyOpts)
```

## Provider Registration

Providers register themselves during initialization:

```go
// internal/providers/cloud/digitalocean/do.go
func init() {
    cloud.Register(&DigitalOceanProvider{})
}
```

## Config Schema

Cloud provider configuration in `stagecraft.yml`:

```yaml
cloud:
  provider: digitalocean
  providers:
    digitalocean:
      token_env: DO_TOKEN
      ssh_key_name: "my-ssh-key"
```

## Non-Goals (v1)

- Infrastructure state management (handled by provider APIs)
- Multi-cloud deployments
- Infrastructure monitoring and alerting

## Related Features

- `PROVIDER_CLOUD_DO` - DigitalOcean CloudProvider implementation
- `CLI_INFRA_UP` - Infra up command that uses cloud providers
- `CLI_INFRA_DOWN` - Infra down command that uses cloud providers
- `CORE_CONFIG` - Config system that validates cloud provider config


# Network Provider Interface

- Feature ID: `PROVIDER_NETWORK_INTERFACE`
- Status: todo
- Depends on: `CORE_CONFIG`

## Goal

Define the interface for network providers that handle mesh networking for multi-host deployments.

Network providers enable Stagecraft to manage mesh networking (e.g., Tailscale, Headscale) for connecting multiple hosts in a deployment.

## Interface

```go
// pkg/providers/network/network.go

package network

import "context"

// EnsureInstalledOptions contains options for ensuring network client is installed.
type EnsureInstalledOptions struct {
	// Config is the provider-specific configuration decoded from
	// network.providers[providerID] in stagecraft.yml.
	// The provider implementation is responsible for unmarshaling this.
	Config any

	// Host is the hostname or Tailscale node name where to ensure installation
	Host string
}

// EnsureJoinedOptions contains options for ensuring a host is joined to the network.
type EnsureJoinedOptions struct {
	// Config is the provider-specific configuration
	Config any

	// Host is the hostname or Tailscale node name
	Host string

	// Tags are the tags to apply to the node (e.g., ["tag:gateway", "tag:app"])
	Tags []string
}

// NetworkProvider is the interface that all network providers must implement.
//
//nolint:revive // NetworkProvider is the preferred name for clarity
type NetworkProvider interface {
	// ID returns the unique identifier for this provider (e.g., "tailscale", "headscale").
	ID() string

	// EnsureInstalled ensures the network client is installed on the given host.
	EnsureInstalled(ctx context.Context, opts EnsureInstalledOptions) error

	// EnsureJoined ensures the host is joined to the mesh network with the given tags.
	EnsureJoined(ctx context.Context, opts EnsureJoinedOptions) error

	// NodeFQDN returns the fully qualified domain name for a node in the mesh network.
	// For example, "plat-db-1.mytailnet.ts.net" for Tailscale.
	NodeFQDN(host string) (string, error)
}
```

## Registry Pattern

Network providers follow the same registry pattern as backend and frontend providers:

- Providers register themselves during initialization via `init()`
- Registry provides thread-safe lookup and registration
- Registry IDs are returned in deterministic lexicographic order
- Duplicate registration panics

## Usage Example

```go
import networkproviders "stagecraft/pkg/providers/network"

// Get a provider
provider, err := networkproviders.Get("tailscale")
if err != nil {
    return err
}

// Ensure installed
opts := networkproviders.EnsureInstalledOptions{
    Config: providerCfg,
    Host:   "plat-db-1",
}
err = provider.EnsureInstalled(ctx, opts)

// Ensure joined
joinOpts := networkproviders.EnsureJoinedOptions{
    Config: providerCfg,
    Host:   "plat-db-1",
    Tags:   []string{"tag:db"},
}
err = provider.EnsureJoined(ctx, joinOpts)

// Get FQDN
fqdn, err := provider.NodeFQDN("plat-db-1")
// fqdn = "plat-db-1.mytailnet.ts.net"
```

## Provider Registration

Providers register themselves during initialization:

```go
// internal/providers/network/tailscale/tailscale.go
func init() {
    network.Register(&TailscaleProvider{})
}
```

## Config Schema

Network provider configuration in `stagecraft.yml`:

```yaml
network:
  provider: tailscale
  providers:
    tailscale:
      auth_key_env: TS_AUTHKEY
      tailnet_domain: "mytailnet.ts.net"
```

## Non-Goals (v1)

- Network policy management (handled by Tailscale/Headscale admin console)
- Dynamic network reconfiguration
- Multiple network providers per project

## Related Features

- `PROVIDER_NETWORK_TAILSCALE` - Tailscale NetworkProvider implementation
- `CLI_DEPLOY` - Deploy command that uses network providers
- `CORE_CONFIG` - Config system that validates network provider config


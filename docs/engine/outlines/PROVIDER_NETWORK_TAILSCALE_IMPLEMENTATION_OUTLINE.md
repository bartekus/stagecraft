# PROVIDER_NETWORK_TAILSCALE Implementation Outline

> This document defines the v1 implementation plan for PROVIDER_NETWORK_TAILSCALE. It translates the feature analysis brief into a concrete, testable, spec aligned delivery plan.

> All details in this outline must be reflected in `spec/providers/network/tailscale.md` before any tests or code are written.

⸻

## 1. Feature Summary

**Feature ID:** PROVIDER_NETWORK_TAILSCALE  
**Domain:** providers/network

**Goal:**

Implement the NetworkProvider interface for Tailscale, enabling Stagecraft to manage Tailscale mesh networking for deployment hosts. The provider ensures Tailscale is installed and configured on hosts, joins them to the correct tailnet with appropriate tags, and provides deterministic FQDNs for use in Compose generation and infrastructure operations.

**v1 Scope:**

- Single network provider per project (Tailscale only)
- Linux hosts only (Debian/Ubuntu) for Tailscale installation
- SSH-based installation and configuration (hosts reachable via public IP before Tailscale is up)
- Auth key from environment variable (never stored in config)
- Deterministic FQDN generation (pure function, no network calls)
- Tag management (default tags + role-specific tags)
- Idempotent operations (EnsureInstalled, EnsureJoined)

**Out of scope for v1:**

- Tailscale ACL management (handled by Tailscale admin console)
- Auth key creation/rotation (user responsibility)
- Multi-OS support (Linux Debian/Ubuntu only)
- Tailscale API integration (CLI-based approach only)
- Dynamic network reconfiguration
- Multiple network providers per project

**Future extensions (not implemented in v1):**

- Support for macOS, Windows hosts
- Custom install methods
- Tailscale API integration for advanced operations
- Support for Headscale (self-hosted Tailscale)

⸻

## 2. Problem Definition and Motivation

Stagecraft needs a way to:

- Ensure hosts are connected to a private mesh network for secure multi-host communication
- Provide stable FQDNs for hosts that can be used in Compose generation and infrastructure configs
- Handle the lifecycle of network client installation and joining hosts to mesh networks

PROVIDER_NETWORK_TAILSCALE provides this by:

- Implementing the NetworkProvider interface for Tailscale
- Managing Tailscale installation on hosts via SSH
- Joining hosts to tailnet with appropriate tags
- Generating deterministic FQDNs from host names and tailnet domain

This enables Stagecraft to support multi-host deployments with secure private networking, unblocking Phase 7 infrastructure features.

⸻

## 3. User Stories (v1)

### Platform Engineer

- As a platform engineer, I want Stagecraft to automatically ensure Tailscale is installed and configured on deployment hosts so I don't have to manually manage mesh networking.

- As a platform engineer, I want Stagecraft to generate stable FQDNs for hosts so I can reference them in Compose files and infrastructure configs.

- As a platform engineer, I want Tailscale provider to handle host tagging automatically so ACLs can be applied correctly.

### Developer

- As a developer deploying to multi-host environments, I want Stagecraft to handle Tailscale setup automatically so I can focus on application deployment.

- As a developer, I want clear error messages when Tailscale setup fails so I can diagnose issues quickly.

⸻

## 4. Inputs and API Contract

### 4.1 NetworkProvider Interface

The provider implements the NetworkProvider interface from `pkg/providers/network/network.go`:

```go
type NetworkProvider interface {
    ID() string
    EnsureInstalled(ctx context.Context, opts EnsureInstalledOptions) error
    EnsureJoined(ctx context.Context, opts EnsureJoinedOptions) error
    NodeFQDN(host string) (string, error)
}
```

### 4.2 EnsureInstalledOptions

```go
type EnsureInstalledOptions struct {
    Config any  // Provider-specific config (unmarshaled from network.providers.tailscale)
    Host   string  // Hostname or logical host ID
}
```

**Host Connectivity:**

For v1, hosts must be reachable via SSH before Tailscale is up. The SSH connection details are not part of EnsureInstalledOptions - Stagecraft core will handle SSH connectivity separately and pass the host identifier.

**Config Structure:**

```go
type Config struct {
    AuthKeyEnv   string            `yaml:"auth_key_env"`   // Required: env var name
    TailnetDomain string           `yaml:"tailnet_domain"`  // Required: tailnet domain
    DefaultTags  []string          `yaml:"default_tags"`    // Optional: default tags
    RoleTags     map[string][]string `yaml:"role_tags"`    // Optional: role-specific tags
    Install      InstallConfig     `yaml:"install"`        // Optional: install settings
}

type InstallConfig struct {
    Method     string `yaml:"method"`      // "auto" or "skip" (default: "auto")
    MinVersion string `yaml:"min_version"` // Optional: minimum Tailscale version
}
```

### 4.3 EnsureJoinedOptions

```go
type EnsureJoinedOptions struct {
    Config any     // Provider-specific config
    Host   string  // Hostname or logical host ID
    Tags   []string // Tags to apply (e.g., ["tag:gateway", "tag:app"])
}
```

**Tags:**

Tags are the union of:
- Default tags from config (`default_tags`)
- Role-specific tags from config (`role_tags[role]`)
- Tags passed in EnsureJoinedOptions (from host role in plan)

### 4.4 NodeFQDN

```go
NodeFQDN(host string) (string, error)
```

**Behavior:**

- Pure function: no network calls, no side effects
- Returns FQDN in format: `{hostname}.{tailnet_domain}`
- Example: `host="db-1"`, `tailnet_domain="example.ts.net"` → `"db-1.example.ts.net"`

⸻

## 5. Data Structures

### 5.1 TailscaleProvider

```go
// internal/providers/network/tailscale/tailscale.go

type TailscaleProvider struct {
    // No fields needed - stateless provider
}

func (p *TailscaleProvider) ID() string {
    return "tailscale"
}
```

### 5.2 Config

```go
// internal/providers/network/tailscale/config.go

type Config struct {
    AuthKeyEnv    string            `yaml:"auth_key_env"`
    TailnetDomain string            `yaml:"tailnet_domain"`
    DefaultTags   []string          `yaml:"default_tags"`
    RoleTags      map[string][]string `yaml:"role_tags"`
    Install       InstallConfig     `yaml:"install"`
}

type InstallConfig struct {
    Method     string `yaml:"method"`      // "auto" or "skip"
    MinVersion string `yaml:"min_version"` // e.g., "1.78.0"
}
```

### 5.3 TailscaleStatus

```go
// internal/providers/network/tailscale/status.go

type TailscaleStatus struct {
    TailnetName string   `json:"TailnetName"`
    Self        NodeInfo `json:"Self"`
}

type NodeInfo struct {
    Online      bool     `json:"Online"`
    TailscaleIPs []string `json:"TailscaleIPs"`
    Tags        []string `json:"Tags"`
}
```

### 5.4 Commander Interface (for testing)

```go
// internal/providers/network/tailscale/commander.go

type Commander interface {
    Run(ctx context.Context, host string, cmd string, args ...string) (stdout, stderr string, err error)
}
```

**Production Implementation:**

```go
type SSHCommander struct {
    // SSH connection details (injected by Stagecraft core)
}

func (c *SSHCommander) Run(ctx context.Context, host string, cmd string, args ...string) (string, string, error) {
    // Use executil + SSH to run command on remote host
}
```

⸻

## 6. Implementation Details

### 6.1 EnsureInstalled Implementation

**Flow:**

1. Parse config from `opts.Config`
2. Validate config (auth_key_env, tailnet_domain required)
3. Check if install should be skipped (`install.method == "skip"`)
4. SSH to host and check if Tailscale is installed:
   - Run: `tailscale version` or `which tailscale`
   - If command succeeds and version >= min_version, return nil (already installed)
5. If not installed:
   - Run Tailscale install script: `curl -fsSL https://tailscale.com/install.sh | sh`
   - Check exit code and return error if install fails
6. Verify installation by running `tailscale version` again

**Error Cases:**

- Config validation errors (missing required fields)
- SSH connection failures
- Install script failures
- Unsupported OS (for v1, only Linux Debian/Ubuntu supported)

### 6.2 EnsureJoined Implementation

**Flow:**

1. Parse config from `opts.Config`
2. Get auth key from environment variable (`config.AuthKeyEnv`)
3. Compute tags: union of `default_tags`, `role_tags[role]`, and `opts.Tags`
4. SSH to host and check current status:
   - Run: `tailscale status --json`
   - Parse JSON to get current tailnet, tags, online status
5. If already joined correctly:
   - Check tailnet matches config
   - Check tags match expected tags
   - If all match, return nil (already joined)
6. If not joined or wrong configuration:
   - Run: `tailscale up --authkey=${AUTHKEY} --hostname=${hostname} --advertise-tags=${tags}`
   - Re-check status to verify join succeeded
7. Validate final state:
   - Tailnet matches config
   - Tags match expected tags
   - Node is online (or at least configured)

**Error Cases:**

- Auth key missing from environment
- Invalid or expired auth key
- Wrong tailnet (node already in different tailnet)
- Tag mismatch (cannot apply required tags)
- SSH connection failures

### 6.3 NodeFQDN Implementation

**Flow:**

1. Parse config (or use cached config)
2. Validate tailnet_domain is set
3. Return: `fmt.Sprintf("%s.%s", host, config.TailnetDomain)`

**Error Cases:**

- Config not loaded (should not happen in practice)
- Tailnet domain missing from config

### 6.4 Config Parsing

**Helper Function:**

```go
func parseConfig(cfg any) (*Config, error) {
    // Convert to YAML bytes and unmarshal
    data, err := yaml.Marshal(cfg)
    if err != nil {
        return nil, fmt.Errorf("marshaling config: %w", err)
    }
    
    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("invalid tailscale provider config: %w", err)
    }
    
    // Validate required fields
    if config.AuthKeyEnv == "" {
        return nil, fmt.Errorf("auth_key_env is required")
    }
    if config.TailnetDomain == "" {
        return nil, fmt.Errorf("tailnet_domain is required")
    }
    
    return &config, nil
}
```

⸻

## 7. Error Handling

### 7.1 Error Types

```go
// internal/providers/network/tailscale/errors.go

var (
    ErrConfigInvalid     = errors.New("invalid config")
    ErrAuthKeyMissing    = errors.New("auth key missing from environment")
    ErrAuthKeyInvalid    = errors.New("invalid or expired auth key")
    ErrTailnetMismatch   = errors.New("tailnet mismatch")
    ErrTagMismatch       = errors.New("tag mismatch")
    ErrInstallFailed     = errors.New("tailscale installation failed")
    ErrUnsupportedOS     = errors.New("unsupported operating system")
)
```

### 7.2 Error Messages

- Config invalid: `"tailscale provider: invalid config: {reason}"`
- Auth key missing: `"tailscale provider: auth key missing from environment variable {env_var}"`
- Auth key invalid: `"tailscale provider: invalid or expired auth key"`
- Tailnet mismatch: `"tailscale provider: host is in tailnet {actual}, expected {expected}"`
- Tag mismatch: `"tailscale provider: host tags {actual} do not match expected {expected}"`
- Install failed: `"tailscale provider: installation failed: {error}"`
- Unsupported OS: `"tailscale provider: unsupported operating system (v1 supports Linux Debian/Ubuntu only)"`

⸻

## 8. Testing Strategy

### 8.1 Unit Tests

**File:** `internal/providers/network/tailscale/tailscale_test.go`

**Test Cases:**

1. **ID()**:
   - Returns "tailscale"

2. **EnsureInstalled**:
   - Already installed (version check succeeds)
   - Not installed, install succeeds
   - Not installed, install fails
   - Install skipped (method == "skip")
   - Config validation errors

3. **EnsureJoined**:
   - Already joined correctly (tailnet and tags match)
   - Not joined, join succeeds
   - Wrong tailnet, join fails
   - Tag mismatch, join fails
   - Auth key missing
   - Auth key invalid

4. **NodeFQDN**:
   - Simple pattern test: `host="db-1"`, `tailnet_domain="example.ts.net"` → `"db-1.example.ts.net"`
   - Config validation errors

5. **Config Parsing**:
   - Valid config
   - Missing required fields
   - Invalid YAML

**Mocking:**

- Use fake Commander implementation for SSH/exec commands
- Use golden JSON fixtures for `tailscale status --json` responses
- Mock environment variable access for auth key

### 8.2 Integration Tests (Optional)

**File:** `internal/providers/network/tailscale/tailscale_integration_test.go`

**Build Tag:** `//go:build integration`

**Requirements:**

- Environment variable: `STAGECRAFT_TAILSCALE_INTEGRATION=1`
- Environment variable: `TS_AUTHKEY` (valid Tailscale auth key)
- Test tailnet configured for testing

**Test Cases:**

- Real Tailscale installation on localhost (via SSH)
- Real Tailscale join with auth key
- Verify FQDN resolution

⸻

## 9. Package Structure

```
internal/providers/network/tailscale/
├── tailscale.go          # Provider implementation
├── tailscale_test.go     # Unit tests
├── config.go             # Config struct and parsing
├── status.go             # Tailscale status parsing
├── commander.go          # Commander interface and SSH implementation
└── errors.go             # Error definitions
```

⸻

## 10. Registry Integration

**Registration:**

```go
// internal/providers/network/tailscale/tailscale.go

func init() {
    network.Register(&TailscaleProvider{})
}
```

**Import:**

The package must be imported somewhere in the main binary to trigger `init()`. This is typically done in `cmd/stagecraft/main.go` with an anonymous import:

```go
import _ "stagecraft/internal/providers/network/tailscale"
```

⸻

## 11. Config Schema Example

**stagecraft.yml:**

```yaml
network:
  provider: tailscale
  providers:
    tailscale:
      auth_key_env: TS_AUTHKEY
      tailnet_domain: "mytailnet.ts.net"
      default_tags:
        - "tag:stagecraft"
        - "tag:stagecraft-env-prod"
      role_tags:
        app:
          - "tag:stagecraft-app"
        db:
          - "tag:stagecraft-db"
        gateway:
          - "tag:stagecraft-gateway"
      install:
        method: "auto"
        min_version: "1.78.0"
```

⸻

## 12. Determinism Guarantees

- **NodeFQDN**: Pure function, deterministic output for given inputs
- **Config parsing**: Deterministic YAML unmarshaling
- **Tag computation**: Deterministic union of tags (sorted)
- **Error messages**: Stable error messages without timestamps or random data
- **Status parsing**: Deterministic JSON parsing (no ordering dependencies)

⸻

## 13. Implementation Checklist

- [ ] Create package structure: `internal/providers/network/tailscale/`
- [ ] Implement Config struct and parsing
- [ ] Implement TailscaleProvider struct with ID()
- [ ] Implement EnsureInstalled with SSH/exec abstraction
- [ ] Implement EnsureJoined with status parsing
- [ ] Implement NodeFQDN (pure function)
- [ ] Define error types
- [ ] Register provider in init()
- [x] Write unit tests (approximately 70% coverage, with all critical paths and error modes covered)
- [ ] Write integration tests (optional, gated by env var)
- [ ] Update config validation if needed
- [ ] Document usage in spec file

⸻

## 14. Dependencies

### Go Dependencies

- `gopkg.in/yaml.v3` - Config parsing (already in go.mod)
- `stagecraft/pkg/providers/network` - Network provider interface
- `stagecraft/pkg/executil` - Process execution utilities
- `stagecraft/pkg/logging` - Logging helpers

### External Dependencies

- Tailscale CLI installable on target hosts (Linux Debian/Ubuntu)
- SSH access to target hosts (before Tailscale is up)
- Tailscale auth key in environment variable

⸻

## 15. Success Metrics

- Provider registers successfully and can be retrieved from registry
- EnsureInstalled works for Linux hosts (Debian/Ubuntu)
- EnsureJoined works with valid auth keys
- NodeFQDN generates correct FQDNs
- Unit tests achieve approximately 70% coverage, with all critical paths and error modes covered
- All error paths are tested
- Integration tests pass (when run with valid auth key)

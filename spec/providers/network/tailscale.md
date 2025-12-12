---
feature: PROVIDER_NETWORK_TAILSCALE
version: v1
status: done
domain: providers
inputs:
  flags: []
outputs:
  exit_codes: {}
---

# Tailscale NetworkProvider Implementation

⸻

## 1. Overview

PROVIDER_NETWORK_TAILSCALE implements the NetworkProvider interface for Tailscale, enabling Stagecraft to manage Tailscale mesh networking for deployment hosts.

It covers:

- Installing Tailscale client on hosts via SSH
- Joining hosts to Tailscale tailnet with appropriate tags
- Generating deterministic FQDNs for hosts
- Ensuring idempotent operations

PROVIDER_NETWORK_TAILSCALE does not:

- Manage Tailscale ACLs or tailnet configuration (handled by Tailscale admin console)
- Create or rotate auth keys (user responsibility)
- Support every OS (Linux Debian/Ubuntu only for v1)
- Use Tailscale API (CLI-based approach only)

⸻

## 2. Interface Contract

The provider implements the NetworkProvider interface from `spec/providers/network/interface.md`:

```go
type NetworkProvider interface {
    ID() string
    EnsureInstalled(ctx context.Context, opts EnsureInstalledOptions) error
    EnsureJoined(ctx context.Context, opts EnsureJoinedOptions) error
    NodeFQDN(host string) (string, error)
}
```

### 2.1 ID

**Behavior:**

- Returns `"tailscale"` as the provider identifier
- Must match the key used in config: `network.provider: tailscale`

### 2.2 EnsureInstalled

**Behavior:**

Ensures the Tailscale client is installed and enabled on the given host.

**Input:**

```go
type EnsureInstalledOptions struct {
    Config any    // Provider-specific config (unmarshaled from network.providers.tailscale)
    Host   string // Hostname or logical host ID
}
```

**Guarantees:**

For a host that is "successfully ensured":

- `tailscaled` daemon is installed and enabled
- `tailscale` CLI is available on PATH
- Tailscale version meets minimum version requirement (if configured)

**Flow:**

1. Parse config from `opts.Config`
2. Validate config (auth_key_env, tailnet_domain required)
3. Check if install should be skipped (`install.method == "skip"`)
   - If skipped, return nil immediately
4. SSH to host and check if Tailscale is installed:
   - Run: `tailscale version` or `which tailscale`
   - If command succeeds:
     - Parse installed version string as semantic version
     - **Version parsing rules**:
       - Strip build metadata (e.g., `1.44.0-123-gabcd` → `1.44.0`)
       - Accept patch suffixes (e.g., `1.78.0-1` → `1.78.0`)
       - If version cannot be parsed as semantic version, return `ErrInstallFailed` with message: `"tailscale provider: installation failed: cannot parse installed version {version}"`
     - If `install.min_version` is configured and installed version < min_version:
       - Return error: `"tailscale provider: installation failed: installed version {actual} is below minimum {min_version}"`
       - Do not attempt automatic upgrade
     - Otherwise, return nil (already installed at acceptable version)
5. If not installed:
   - Run Tailscale install script: `curl -fsSL https://tailscale.com/install.sh | sh`
   - Check exit code and return error if install fails
6. Verify installation by running `tailscale version` again

**Idempotency:**

- If Tailscale is already installed at acceptable version, does nothing and returns nil
- Running EnsureInstalled multiple times produces identical results

**Supported OS (v1):**

- **Target hosts**: Linux (Debian/Ubuntu) only
- Uses Tailscale's official install script
- **Unsupported target OS**: macOS, Windows, Alpine, CentOS, and all other Linux distributions
  - For unsupported target OS, `EnsureInstalled` MUST return `ErrUnsupportedOS`
- **Orchestrator OS**: Completely irrelevant - the provider operates on remote hosts via SSH, not the orchestrator machine
  - The orchestrator (local machine running Stagecraft) MUST NOT be inspected for OS compatibility
  - Only remote hosts are checked for OS compatibility
- Other OS support is deferred to future versions

**Note**: The orchestrator (the machine running Stagecraft) does not need Tailscale installed. The provider manages Tailscale on remote Linux hosts via SSH. You may install Tailscale on your Mac for your own network access, but this is not a requirement for provider operation.

**Error Cases:**

- Config validation errors (missing required fields)
- SSH connection failures
- Install script failures (non-zero exit code)
- Unsupported OS (for v1, only Linux Debian/Ubuntu supported)
- Installed version below minimum version requirement (if `install.min_version` is configured)
- Installed version cannot be parsed as semantic version

**Error Messages:**

- Config invalid: `"tailscale provider: invalid config: {reason}"`
- Install failed: `"tailscale provider: installation failed: {error}"`
- Unsupported OS: `"tailscale provider: unsupported operating system (v1 supports Linux Debian/Ubuntu only)"`

### 2.3 EnsureJoined

**Behavior:**

Ensures the host is logged into the correct Tailscale tailnet with the configured tags.

**Input:**

```go
type EnsureJoinedOptions struct {
    Config any     // Provider-specific config
    Host   string  // Hostname or logical host ID
    Tags   []string // Final computed tags to apply (includes default_tags, role_tags, and plan tags)
}
```

**Note**: The `Tags` field contains the final computed tag list. Stagecraft core computes this by combining `default_tags`, `role_tags[role]`, and any plan-specific tags before calling `EnsureJoined`. The provider does not access `role_tags` directly.

**Guarantees:**

For a host that is "successfully joined":

- Host is logged into the configured tailnet
- Host has the configured tags applied
- Host's Tailscale node is online or at least successfully configured
- **Offline node handling**: If a node is offline but has the correct tailnet and tags configured, `EnsureJoined` considers this a success
- The provider checks `Self.Online` from status JSON but does not fail if the node is offline, as long as configuration is correct
- This allows deployment workflows to proceed even if nodes are temporarily offline

**Flow:**

1. Parse config from `opts.Config`
2. Get auth key from environment variable (`config.AuthKeyEnv`)
   - If missing, return error
3. Use expected tags from `opts.Tags`:
   - The provider receives the final tag list in `opts.Tags`
   - Stagecraft core (the caller) is responsible for computing the union of:
     - Default tags from config (`default_tags`)
     - Role-specific tags from config (`role_tags[role]`) where role is determined by core
     - Any additional tags from the deployment plan
   - **Forbidden behavior**: The provider MUST NOT attempt to compute `role_tags` or inspect roles directly
   - Only the final tag list in `opts.Tags` is authoritative
   - Tags are sorted lexicographically for deterministic comparison
4. SSH to host and check current status:
   - Run: `tailscale status --json`
   - Parse JSON to get current tailnet, tags, online status
5. If already joined correctly:
   - Check tailnet matches config:
     - Compare `TailnetName` from `tailscale status --json` against the expected tailnet name
     - The `tailnet_domain` config field is used only for FQDN generation via `NodeFQDN`
     - For status comparison, the provider compares the `TailnetName` field from status JSON
     - **Expected tailnet name derivation**: The expected tailnet name SHALL be the portion of `tailnet_domain` before the first dot
       - Example: `tailnet_domain = "bartekus.ts.net"` → expected `TailnetName = "bartekus"`
       - Example: `tailnet_domain = "example.ts.net"` → expected `TailnetName = "example"`
     - If `TailnetName` does not match the expected tailnet name (derived from config), return `ErrTailnetMismatch`
   - Check tags match expected tags:
     - All expected tags must be present in the actual tags (subset match)
     - Extra tags beyond the expected set are allowed (Tailscale may add system tags)
     - Tag comparison is case-sensitive and exact string match
     - Tags are sorted lexicographically for deterministic comparison
   - If all match, return nil (already joined)
6. If not joined or wrong configuration:
   - Run: `tailscale up --authkey=${AUTHKEY} --hostname=${hostname} --advertise-tags=${tags}`
   - Re-check status to verify join succeeded
7. Validate final state:
   - Tailnet matches config (using TailnetName comparison as above)
   - Tags match expected tags (subset match: all expected tags present, extras allowed)
   - Node configuration is correct (online status is checked but offline nodes with correct config are acceptable)

**Tag Computation:**

Tags are computed by Stagecraft core (not the provider):
- Default tags from config (`default_tags`)
- Role-specific tags from config (`role_tags[role]`)
- Tags passed in `opts.Tags` from the deployment plan

The provider receives the final computed tag list in `opts.Tags` and applies it directly. Tags are sorted lexicographically for deterministic comparison.

**Idempotency:**

- If host is already joined to correct tailnet with correct tags, does nothing and returns nil
- Running EnsureJoined multiple times produces identical results

**Error Cases:**

- Auth key missing from environment variable
- Invalid or expired auth key
- Wrong tailnet (node already in different tailnet)
- Tag mismatch (cannot apply required tags)
- SSH connection failures

**Error Messages:**

- Auth key missing: `"tailscale provider: auth key missing from environment variable {env_var}"`
- Auth key invalid: `"tailscale provider: invalid or expired auth key"`
- Tailnet mismatch: `"tailscale provider: host is in tailnet {actual}, expected {expected}"`
- Tag mismatch: `"tailscale provider: host tags {actual} do not match expected {expected}"`

### 2.4 NodeFQDN

**Behavior:**

Returns the fully qualified domain name for a node in the Tailscale mesh network.

**Input:**

- `host`: Hostname or logical host ID (e.g., "app-1", "db-1")

**Output:**

- FQDN string (e.g., "app-1.mytailnet.ts.net")
- Error if config is invalid

**Guarantees:**

- Pure function with respect to config: no network calls, no side effects
- Deterministic: same inputs always produce same output
- Format: `{hostname}.{tailnet_domain}`
- **Note**: Requires config to be set via EnsureInstalled or EnsureJoined first. NodeFQDN uses provider config loaded from those methods and is pure with respect to that config (no network calls or side effects).

**Flow:**

1. Parse config (or use cached config from previous calls)
2. Validate tailnet_domain is set
3. Return: `fmt.Sprintf("%s.%s", host, config.TailnetDomain)`

**Examples:**

- `host="db-1"`, `tailnet_domain="example.ts.net"` → `"db-1.example.ts.net"`
- `host="gateway-1"`, `tailnet_domain="mytailnet.ts.net"` → `"gateway-1.mytailnet.ts.net"`

**Error Cases:**

- Config not loaded (should not happen in practice)
- Tailnet domain missing from config

**Error Messages:**

- Config invalid: `"tailscale provider: invalid config: tailnet_domain is required"`

⸻

## 3. Config Schema

### 3.1 Config Structure

Network provider configuration in `stagecraft.yml`:

```yaml
network:
  provider: tailscale
  providers:
    tailscale:
      # Required - env var name on the orchestrator machine
      auth_key_env: TS_AUTHKEY

      # Required - MagicDNS tailnet domain
      tailnet_domain: "mytailnet.ts.net"

      # Optional - default tags applied to all nodes
      default_tags:
        - "tag:stagecraft"
        - "tag:stagecraft-env-prod"

      # Optional - per-role tags
      role_tags:
        app:
          - "tag:stagecraft-app"
        db:
          - "tag:stagecraft-db"
        gateway:
          - "tag:stagecraft-gateway"

      # Optional - install settings
      install:
        method: "auto"    # "auto" or "skip" (default: "auto")
        min_version: "1.78.0"  # Optional: minimum Tailscale version
```

### 3.2 Config Fields

**auth_key_env** (required):

- Environment variable name containing Tailscale auth key
- Auth key is never stored in `stagecraft.yml` (security requirement)
- Must be set on the machine running Stagecraft
- Example: `TS_AUTHKEY`

**tailnet_domain** (required):

- Tailscale tailnet domain (MagicDNS domain)
- Used for FQDN generation
- Example: `"mytailnet.ts.net"`

**default_tags** (optional):

- List of tags applied to all nodes
- Tags must start with `tag:` prefix
- Tags without the `tag:` prefix are considered invalid config and cause `ErrConfigInvalid`
- **Validation timing**: Tag validation occurs during provider config parsing inside `EnsureInstalled` and `EnsureJoined`
- Config with invalid tags MUST fail before any remote operations (SSH or Tailscale calls)
- Example: `["tag:stagecraft", "tag:stagecraft-env-prod"]`

**role_tags** (optional):

- Map of role-specific tags
- Key is host role (e.g., "app", "db", "gateway")
- Value is list of tags for that role
- Tags must start with `tag:` prefix (same validation as `default_tags`)
- Tags are combined with default_tags and opts.Tags by Stagecraft core (not the provider)

**install.method** (optional):

- `"auto"`: Automatically install Tailscale if not present (default)
- `"skip"`: Skip installation (assume Tailscale is already installed)

**install.min_version** (optional):

- Minimum Tailscale version required
- Format: semantic version (e.g., "1.78.0")
- If not set, any installed version is acceptable

### 3.3 Config Validation

**Required Fields:**

- `auth_key_env`: Must be non-empty string
- `tailnet_domain`: Must be non-empty string

**Tag Validation:**

- All tags in `default_tags` and `role_tags` must start with `tag:` prefix
- Tags without prefix cause: `"tailscale provider: invalid config: tag {tag} must start with tag: prefix"`
- Validation is strict: invalid tags cause config parsing to fail
- Validation occurs during provider config parsing, before any remote operations

**Validation Errors:**

- Missing required field: `"tailscale provider: invalid config: {field} is required"`
- Invalid YAML: `"tailscale provider: invalid config: {error}"`
- Invalid tag format: `"tailscale provider: invalid config: tag {tag} must start with tag: prefix"`

⸻

## 4. Environment Variables

### 4.1 Required Environment Variables

**TS_AUTHKEY** (or value of `auth_key_env`):

- Tailscale auth key for joining hosts to tailnet
- Must be set on the machine running Stagecraft
- Never stored in config files
- Can be a one-time auth key or reusable auth key

**Usage:**

- Provider reads auth key from environment variable specified in `auth_key_env`
- If environment variable is not set, EnsureJoined returns error

### 4.2 Integration Test Environment Variables

**STAGECRAFT_TAILSCALE_INTEGRATION** (optional):

- Set to `"1"` to enable integration tests
- Integration tests are gated by this variable

⸻

## 5. Behavioral Invariants

### 5.1 Success Invariants

For a host that is "successfully ensured" by the Tailscale provider:

- Tailscale is installed and `tailscale version` returns a version >= configured minimum
- `tailscale status --json`:
  - Reports being logged into the configured tailnet (TailnetName matches expected tailnet name derived from config)
  - Shows tags where all expected tags are present (subset match: expected tags must be present, extras allowed)
    - Expected tags are the union of:
      - `default_tags`
      - Role-specific tags for the host (computed by core)
      - Tags passed in EnsureJoinedOptions
- `NodeFQDN(host)` returns a string that resolves in MagicDNS to that host's Tailscale IP

### 5.2 Idempotency Invariants

- Running `EnsureInstalled` multiple times produces identical results
- Running `EnsureJoined` multiple times produces identical results
- `NodeFQDN` is a pure function (no side effects)

### 5.3 Determinism Invariants

- `NodeFQDN` output is deterministic for given inputs
- Tag computation is deterministic (sorted union)
- Error messages are stable (no timestamps or random data)
- Config parsing is deterministic

⸻

## 6. Error Handling

### 6.1 Error Categories

**Config Errors:**

- Missing required fields
- Invalid YAML
- Invalid field values

**Install Errors:**

- Tailscale package install failed (non-zero exit code)
- OS not supported by v1 install method
- SSH connection failures

**Join Errors:**

- Invalid or expired auth key
- Tailnet mismatch (status shows different tailnet than expected)
- Tag mismatch (node is joined but tags differ from config)
- SSH connection failures

### 6.2 Error Messages

All errors must be clear and actionable:

- Config errors: Include field name and reason
- Install errors: Include OS and install script output
- Join errors: Include actual vs expected values (tailnet, tags)
- Version too old: `"tailscale provider: installation failed: installed version {actual} is below minimum {min_version}"`
- Version parse failure: `"tailscale provider: installation failed: cannot parse installed version {version}"`

### 6.3 Error Types

Provider defines specific error values for different failure categories:

```go
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

⸻

## 7. Integration with Stagecraft Core

### 7.1 Provider Registration

Provider registers itself during package initialization:

```go
// internal/providers/network/tailscale/tailscale.go

func init() {
    network.Register(&TailscaleProvider{})
}
```

Package must be imported in main binary to trigger registration:

```go
// cmd/stagecraft/main.go

import _ "stagecraft/internal/providers/network/tailscale"
```

### 7.2 Usage in Core

Core uses provider via network registry:

```go
import networkproviders "stagecraft/pkg/providers/network"

// Get provider
provider, err := networkproviders.Get("tailscale")
if err != nil {
    return err
}

// Ensure installed
opts := networkproviders.EnsureInstalledOptions{
    Config: providerCfg,
    Host:   "app-1",
}
err = provider.EnsureInstalled(ctx, opts)

// Ensure joined
joinOpts := networkproviders.EnsureJoinedOptions{
    Config: providerCfg,
    Host:   "app-1",
    Tags:   []string{"tag:app"},
}
err = provider.EnsureJoined(ctx, joinOpts)

// Get FQDN
fqdn, err := provider.NodeFQDN("app-1")
// fqdn = "app-1.mytailnet.ts.net"
```

### 7.3 Integration with Phase 7

**CLI_INFRA_UP flow:**

1. Cloud provider ensures droplet exists
2. Some connectivity (public IP + SSH) is available
3. Network provider is invoked:
   ```go
   networkProvider.EnsureInstalled(ctx, optsForHost)
   networkProvider.EnsureJoined(ctx, optsForHost)
   fqdn := networkProvider.NodeFQDN(hostID)
   ```
4. Compose generation or infra metadata uses fqdn

**INFRA_HOST_BOOTSTRAP:**

- Uses network provider to ensure Tailscale is installed and joined
- Generates host metadata containing `tailscale_fqdn`

⸻

## 8. SSH Connectivity

### 8.1 Host Reachability

For v1, hosts must be reachable via SSH before Tailscale is up:

- Hosts are typically provisioned by cloud provider with public IP
- SSH access is required for Tailscale installation and configuration
- After Tailscale is up, hosts can communicate via MagicDNS FQDN

### 8.2 SSH Details

SSH connection details are not part of NetworkProvider interface:

- Stagecraft core handles SSH connectivity separately
- Core passes host identifier to provider
- Provider uses injected Commander interface for SSH commands

**Commander Interface:**

```go
type Commander interface {
    Run(ctx context.Context, host string, cmd string, args ...string) (stdout, stderr string, err error)
}
```

Production implementation uses `executil` + SSH.

⸻

## 9. Determinism

### 9.1 Deterministic Operations

- **NodeFQDN**: Pure function, deterministic output
- **Tag computation**: Deterministic union (sorted)
- **Config parsing**: Deterministic YAML unmarshaling
- **Status parsing**: Deterministic JSON parsing (no ordering dependencies)

### 9.2 No Randomness

- No random identifiers in hostnames or tags
- No timestamps in error messages
- No machine-specific data in outputs

### 9.3 Test Determinism

- Unit tests use mocked SSH/exec commands
- Tests use golden JSON fixtures for `tailscale status --json`
- Tests are deterministic and reproducible

⸻

## 10. Testing

### 10.1 Unit Tests

**File:** `internal/providers/network/tailscale/tailscale_test.go`

**Coverage Target:** Approximately 70%, with all critical paths and error modes covered

**Test Cases:**

1. **ID()**: Returns "tailscale"
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
   - Simple pattern test
   - Config validation errors
5. **Config Parsing**:
   - Valid config
   - Missing required fields
   - Invalid YAML

**Mocking:**

- Use fake Commander implementation
- Use golden JSON fixtures for `tailscale status --json`
- Mock environment variable access

### 10.2 Integration Tests (Optional)

**File:** `internal/providers/network/tailscale/tailscale_integration_test.go`

**Build Tag:** `//go:build integration`

**Requirements:**

- `STAGECRAFT_TAILSCALE_INTEGRATION=1`
- `TS_AUTHKEY` (valid Tailscale auth key)
- Test tailnet configured

**Test Cases:**

- Real Tailscale installation on localhost (via SSH)
- Real Tailscale join with auth key
- Verify FQDN resolution

⸻

## 11. Non-Goals (v1)

- Managing Tailscale ACLs or tailnet configuration (handled by Tailscale admin console)
- Supporting every OS (Linux Debian/Ubuntu only)
- Supporting macOS or Windows as target hosts (Linux Debian/Ubuntu only)
- Requiring Tailscale to be installed on the orchestrator machine
- Managing auth key creation or rotation (user responsibility)
- Dynamic network reconfiguration (static configuration only)
- Multiple network providers per project (single provider only)
- Tailscale API integration (CLI-based approach only)
- Support for Headscale (self-hosted Tailscale) - deferred to future

⸻

## 12. Related Features

- `PROVIDER_NETWORK_INTERFACE` - Network provider interface definition
- `CLI_DEPLOY` - Deploy command (has TODOs for network provider integration)
- `CLI_INFRA_UP` - Infra up command that uses network providers (Phase 7)
- `CLI_INFRA_DOWN` - Infra down command that uses network providers (Phase 7)
- `INFRA_HOST_BOOTSTRAP` - Host bootstrap that uses network providers (Phase 7)
- `CORE_CONFIG` - Config system that validates network provider config

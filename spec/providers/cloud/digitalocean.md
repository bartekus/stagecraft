---
feature: PROVIDER_CLOUD_DO
version: v1
status: done
domain: providers
inputs:
  flags: []
outputs:
  exit_codes: {}
---

# DigitalOcean CloudProvider Implementation

⸻

## 1. Overview

PROVIDER_CLOUD_DO implements the CloudProvider interface for DigitalOcean, enabling Stagecraft to provision and manage DigitalOcean droplets for deployment environments.

It covers:

- Generating infrastructure plans from config and environment (dry-run)
- Creating DigitalOcean droplets
- Deleting DigitalOcean droplets
- Ensuring idempotent operations
- Deterministic planning

PROVIDER_CLOUD_DO does not:

- Manage other DigitalOcean resources (load balancers, volumes, databases, etc.)
- Support other cloud providers (AWS, GCP, Azure)
- Create or manage SSH keys in DigitalOcean account
- Provide cost estimation or budgeting
- Support multi-region deployments (single region per environment for v1)
- Support automatic scaling or auto-scaling groups

⸻

## 2. Interface Contract

The provider implements the CloudProvider interface from `spec/providers/cloud/interface.md`:

```go
type CloudProvider interface {
    ID() string
    Plan(ctx context.Context, opts PlanOptions) (InfraPlan, error)
    Apply(ctx context.Context, opts ApplyOptions) error
}
```

### 2.1 ID

**Behavior:**

- Returns `"digitalocean"` as the provider identifier
- Must match the key used in config: `cloud.provider: digitalocean`

### 2.2 Plan

**Behavior:**

Generates an infrastructure plan for the given environment. This is a dry-run operation that does not modify infrastructure.

**Input:**

```go
type PlanOptions struct {
    Config      any    // Provider-specific config (unmarshaled from cloud.providers.digitalocean)
    Environment string // Environment name (e.g., "staging", "prod")
}
```

**Output:**

```go
type InfraPlan struct {
    ToCreate []HostSpec // Hosts to be created (sorted lexicographically by Name)
    ToDelete []HostSpec // Hosts to be deleted (sorted lexicographically by Name)
}

type HostSpec struct {
    Name   string // Hostname (e.g., "app-1", "db-1")
    Role   string // Role (e.g., "gateway", "app", "db", "cache")
    Size   string // Instance size (e.g., "s-2vcpu-4gb")
    Region string // Region (e.g., "nyc1")
}
```

**Guarantees:**

- Side-effect free: No API calls that modify infrastructure
- Deterministic: Same config and environment produce same plan
- Sorted output: ToCreate and ToDelete lists sorted lexicographically by Name
- Pure function: No timestamps, no random data, no hidden state

**Flow:**

1. Parse config from `opts.Config`
2. Validate config (token_env, ssh_key_name, hosts required)
3. Get DigitalOcean API token from environment variable (`config.TokenEnv`)
   - If missing, return error
4. Validate SSH key exists in DigitalOcean account
   - List SSH keys via API
   - Find key matching `config.SSHKeyName`
   - If not found, return error
5. Get desired hosts from config for given environment (`opts.Environment`)
   - Look up `config.Hosts[environment]`
   - If environment not found, return empty plan (no hosts to create/delete)
6. List actual droplets from DigitalOcean API
   - Filter by environment tag or name pattern (`{environment}-*`)
   - Map droplets to HostSpec by parsing names
7. Compare desired vs actual:
   - ToCreate: desired hosts not in actual (by name)
   - ToDelete: actual hosts not in desired (by name)
8. Sort ToCreate and ToDelete lexicographically by Name
9. Return InfraPlan

**Determinism:**

- Same config and environment always produces same plan
- Hosts sorted lexicographically by Name
- No timestamps or random data in plan
- Plan is side-effect free (no infrastructure modifications)

**Error Cases:**

- Config validation errors (missing required fields)
- API token missing from environment variable
- SSH key not found in DigitalOcean account
- API failures (rate limits, network errors)
- Invalid region or size values

**Error Messages:**

- Config invalid: `"digitalocean provider: invalid config: {reason}"`
- Token missing: `"digitalocean provider: API token missing from environment variable {env_var}"`
- SSH key not found: `"digitalocean provider: SSH key '{name}' not found in DigitalOcean account"`
- API error: `"digitalocean provider: API error: {error}"`

### 2.3 Apply

**Behavior:**

Applies the given infrastructure plan, creating and deleting droplets as needed.

**Input:**

```go
type ApplyOptions struct {
    Config any      // Provider-specific config
    Plan   InfraPlan // Infrastructure plan to apply
}
```

**Guarantees:**

- Creates droplets specified in plan.ToCreate
- Deletes droplets specified in plan.ToDelete
- Handles already-existing droplets gracefully (idempotent)
- Handles already-deleted droplets gracefully (idempotent)
- Waits for droplets to be ready before returning (for create operations)
- Processes operations in deterministic order (sorted by Name)

**Flow:**

1. Parse config from `opts.Config`
2. Validate config
3. Get DigitalOcean API token from environment variable
4. Validate SSH key exists in DigitalOcean account
5. Get SSH key ID from DigitalOcean API
6. Process plan.ToCreate (in order, sorted by Name):
   - For each host spec:
     - Check if droplet exists (by name: `{environment}-{hostname}`)
     - If exists and matches spec (region, size), skip (idempotent)
     - If exists but doesn't match spec, return error (reconciliation needed)
     - If doesn't exist, create droplet via DigitalOcean API:
       - Name: `{environment}-{hostname}`
       - Region: `hostspec.Region`
       - Size: `hostspec.Size`
       - Image: `ubuntu-22-04-x64` (hardcoded for v1)
       - SSH Keys: `[sshKeyID]`
       - Tags: `["stagecraft", "stagecraft-env-{environment}"]`
     - Poll for droplet status until "active"
     - Return error if creation fails or times out
7. Process plan.ToDelete (in order, sorted by Name):
   - For each host spec:
     - Find droplet by name (`{environment}-{hostname}`)
     - If doesn't exist, skip (idempotent)
     - If exists, delete droplet via DigitalOcean API
     - Poll for droplet deletion until confirmed
     - Return error if deletion fails or times out
8. Return nil on success

**Droplet Naming Convention:**

- Format: `{environment}-{hostname}`
- Examples: `staging-app-1`, `prod-gateway-1`
- Rationale: Enables filtering by environment, prevents name conflicts

**Idempotency:**

- Already-existing droplets are skipped (if they match spec)
- Already-deleted droplets are skipped
- Running Apply() multiple times produces identical results

**Error Cases:**

- Config validation errors
- API token missing
- SSH key not found
- Droplet creation failures
- Droplet deletion failures
- Timeout waiting for droplet status
- API rate limits (with retry logic)
- Droplet exists but doesn't match spec (reconciliation needed)

**Error Messages:**

- Droplet exists: `"digitalocean provider: droplet '{name}' already exists"`
- Droplet not found: `"digitalocean provider: droplet '{name}' not found"`
- Droplet create failed: `"digitalocean provider: droplet creation failed: {error}"`
- Droplet delete failed: `"digitalocean provider: droplet deletion failed: {error}"`
- Droplet timeout: `"digitalocean provider: droplet operation timeout"`
- Rate limit: `"digitalocean provider: API rate limit exceeded, retrying..."`

⸻

## 3. Config Schema

### 3.1 Config Structure

```yaml
cloud:
  provider: digitalocean
  providers:
    digitalocean:
      token_env: DO_TOKEN              # Required: env var name for DO API token
      ssh_key_name: "my-ssh-key"       # Required: SSH key name in DO account
      default_region: "nyc1"           # Optional: default region
      default_size: "s-2vcpu-4gb"      # Optional: default size
      regions:                         # Optional: allowed regions
        - "nyc1"
        - "nyc3"
      sizes:                           # Optional: allowed sizes
        - "s-2vcpu-4gb"
        - "s-4vcpu-8gb"
      hosts:                           # Required: host definitions per environment
        staging:
          app-1:
            role: app
            size: "s-2vcpu-4gb"
            region: "nyc1"
          db-1:
            role: db
            size: "s-4vcpu-8gb"
            region: "nyc1"
        prod:
          gateway-1:
            role: gateway
            size: "s-2vcpu-4gb"
            region: "nyc3"
          app-1:
            role: app
            size: "s-4vcpu-8gb"
            region: "nyc3"
          app-2:
            role: app
            size: "s-4vcpu-8gb"
            region: "nyc3"
          db-1:
            role: db
            size: "s-8vcpu-16gb"
            region: "nyc3"
```

### 3.2 Config Validation

**Required Fields:**

- `token_env`: Environment variable name for DigitalOcean API token
- `ssh_key_name`: SSH key name in DigitalOcean account
- `hosts`: Host definitions per environment

**Optional Fields:**

- `default_region`: Default region for hosts (used if host doesn't specify region)
- `default_size`: Default size for hosts (used if host doesn't specify size)
- `regions`: Allowed regions (validated against DigitalOcean API)
- `sizes`: Allowed sizes (validated against DigitalOcean API)

**Validation Rules:**

- `token_env` must be non-empty string (environment variable name, not the token itself)
- `ssh_key_name` must be non-empty string (must exist in DigitalOcean account)
- `hosts` must be non-empty map (at least one environment defined)
- Each host config must have `role` field (required)
- Host `size` defaults to `default_size` if not specified (or fails if no default)
- Host `region` defaults to `default_region` if not specified (or fails if no default)
- Region and size values validated against DigitalOcean API (if `regions`/`sizes` specified)
- Token is never stored in config or state; only environment variable name is stored

**Error Messages:**

- Missing required field: `"digitalocean provider: invalid config: {field} is required"`
- Invalid region: `"digitalocean provider: invalid config: region '{region}' is not allowed"`
- Invalid size: `"digitalocean provider: invalid config: size '{size}' is not allowed"`

**SSH Key Behavior:**

- If `ssh_key_name` does not exist in DigitalOcean account, Plan() and Apply() fail with `ErrSSHKeyNotFound`
- Provider does not create SSH keys; user must pre-configure SSH key in DigitalOcean account
- SSH key lookup happens during Plan() and Apply() via DigitalOcean API

⸻

## 4. Determinism Requirements

### 4.1 Plan() Determinism

- Same config and environment always produces same plan
- Hosts sorted lexicographically by Name in ToCreate and ToDelete
- No timestamps or random data in plan
- Plan is side-effect free (no infrastructure modifications)

### 4.2 Apply() Determinism

- Operations processed in deterministic order (sorted by Name)
- Idempotent operations produce identical results
- Droplet naming is deterministic (`{environment}-{hostname}`)

### 4.3 Error Messages

- Stable error messages without timestamps or random data
- Error messages include relevant context (droplet name, region, size)

⸻

## 5. Error Handling

### 5.1 Error Types

**Config Errors** (local, deterministic, no API calls):
- `ErrConfigInvalid`: Config validation errors

**Authentication Errors** (API calls required):
- `ErrTokenMissing`: API token missing from environment
- `ErrSSHKeyNotFound`: SSH key not found in DigitalOcean account

**Resource Errors** (API operations):
- `ErrDropletExists`: Droplet already exists (when reconciliation needed)
- `ErrDropletNotFound`: Droplet not found
- `ErrDropletCreateFailed`: Droplet creation failed
- `ErrDropletDeleteFailed`: Droplet deletion failed
- `ErrDropletTimeout`: Droplet operation timeout

**API Errors** (infrastructure/rate limiting):
- `ErrAPIError`: DigitalOcean API error (wraps underlying API errors)
- `ErrRateLimit`: API rate limit exceeded (with retry logic)

### 5.2 Error Message Format

All error messages are prefixed with `"digitalocean provider: "` for consistency.

Examples:

- `"digitalocean provider: invalid config: token_env is required"`
- `"digitalocean provider: API token missing from environment variable DO_TOKEN"`
- `"digitalocean provider: SSH key 'my-ssh-key' not found in DigitalOcean account"`
- `"digitalocean provider: droplet 'staging-app-1' already exists"`
- `"digitalocean provider: API error: rate limit exceeded"`

⸻

## 6. Testing Requirements

### 6.1 Unit Tests

- Provider registration and ID() method
- Plan() with various configs and environments
- Plan() determinism (same inputs produce same output)
- Plan() sorting (ToCreate and ToDelete sorted lexicographically)
- Apply() create operations
- Apply() delete operations
- Apply() idempotency (already-existing/deleted droplets)
- Config parsing and validation
- Error handling for all error cases

**Coverage Target**: Approximately 70% coverage, with all critical paths and error modes covered.

### 6.2 Integration Tests (Optional)

**Build Tag**: `//go:build integration`

**Requirements:**

- Environment variable: `STAGECRAFT_DO_INTEGRATION=1`
- Environment variable: `DO_TOKEN` (valid DigitalOcean API token)
- SSH key configured in DigitalOcean account

**Test Cases:**

- Real droplet creation
- Real droplet deletion
- Real plan generation
- Verify idempotency

⸻

## 7. Implementation Notes

### 7.1 DigitalOcean API Client

- Uses `github.com/digitalocean/godo` for API client
- API client injected via dependency injection for testability
- Handles API rate limits with retry logic and exponential backoff

### 7.2 Droplet Image

- Uses `ubuntu-22-04-x64` image (hardcoded for v1)
- Future versions may support configurable images

### 7.3 Droplet Tags

- Automatically adds tags: `["stagecraft", "stagecraft-env-{environment}"]`
- Enables filtering droplets by environment
- Tags are deterministic and derived from environment name

### 7.4 Async Operations

- Droplet creation and deletion are async operations
- Provider polls for droplet status until "active" (for creation) or confirmed deletion
- Polling interval: 5 seconds (configurable in future)
- Timeout: 10 minutes (configurable in future)
- If timeout occurs, Apply() returns `ErrDropletTimeout` with actionable error message
- Partial failures: If some droplets are created/deleted but others fail, Apply() returns error describing partial state
- Best-effort rollback: v1 does not automatically rollback partial failures; operator must manually reconcile

⸻

## 8. Related Features

- `PROVIDER_CLOUD_INTERFACE` - Cloud provider interface definition
- `CLI_INFRA_UP` - Infrastructure provisioning command
- `CLI_INFRA_DOWN` - Infrastructure teardown command
- `CORE_PLAN` - Planning engine for infrastructure planning
- `CORE_CONFIG` - Config loading and validation

⸻

## 9. Non-Goals (v1)

- Managing other DigitalOcean resources (load balancers, volumes, databases, etc.)
- Supporting other cloud providers (AWS, GCP, Azure)
- Creating or managing SSH keys in DigitalOcean account
- Cost estimation or budgeting
- Infrastructure monitoring or alerting
- Multi-region deployments (single region per environment)
- Automatic scaling or auto-scaling groups
- Droplet resizing or modification

## 10. Cost and Billing Responsibility

**Operators are responsible for DigitalOcean billing and cost control.**

Stagecraft does not perform cost estimation, quota enforcement, or billing management. When Apply() creates droplets, they immediately incur DigitalOcean charges. Operators must:

- Review Plan() output before applying to understand infrastructure changes
- Monitor DigitalOcean billing dashboard for costs
- Set up DigitalOcean billing alerts independently
- Ensure sufficient quota/limits before running Apply()

Stagecraft provides no safeguards against accidental cost overruns beyond the dry-run Plan() operation.

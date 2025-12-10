# PROVIDER_CLOUD_DO Implementation Outline

> This document defines the v1 implementation plan for PROVIDER_CLOUD_DO. It translates the feature analysis brief into a concrete, testable, spec aligned delivery plan.

> All details in this outline must be reflected in `spec/providers/cloud/digitalocean.md` before any tests or code are written.

⸻

## 1. Feature Summary

**Feature ID:** PROVIDER_CLOUD_DO  
**Domain:** providers/cloud

**Goal:**

Implement the CloudProvider interface for DigitalOcean, enabling Stagecraft to provision and manage DigitalOcean droplets for deployment environments. The provider generates infrastructure plans, creates droplets, deletes droplets, and ensures idempotent operations.

**v1 Scope:**

- Single cloud provider per project (DigitalOcean only)
- Droplet creation and deletion only (no other DigitalOcean resources)
- Plan() operation (dry-run, side-effect free)
- Apply() operation (create/delete droplets)
- SSH key management (validate key exists, use for droplet creation)
- Region and size validation against DigitalOcean API
- Idempotent operations (handle already-existing/deleted droplets gracefully)
- Deterministic planning (same inputs produce same output)
- Async operation handling (poll for droplet status until ready)

**Out of scope for v1:**

- Other DigitalOcean resources (load balancers, volumes, databases, etc.)
- Other cloud providers (AWS, GCP, Azure)
- SSH key creation/management in DigitalOcean account
- Cost estimation or budgeting
- Infrastructure monitoring or alerting
- Multi-region deployments (single region per environment)
- Automatic scaling or auto-scaling groups
- Droplet resizing or modification

**Future extensions (not implemented in v1):**

- Support for other cloud providers (AWS, GCP, Azure)
- Support for other DigitalOcean resources (load balancers, volumes, etc.)
- Multi-region deployments
- Cost estimation and budgeting
- Infrastructure monitoring integration
- Auto-scaling groups

⸻

## 2. Problem Definition and Motivation

Stagecraft needs a way to:

- Provision actual infrastructure (droplets) on cloud providers based on deployment plans
- Plan infrastructure changes before applying them (dry-run)
- Manage infrastructure lifecycle (create, delete, reconcile)
- Integrate infrastructure provisioning with deployment workflows

PROVIDER_CLOUD_DO provides this by:

- Implementing the CloudProvider interface for DigitalOcean
- Generating infrastructure plans from config and environment
- Creating and deleting DigitalOcean droplets
- Ensuring idempotent operations

This enables Stagecraft to support infrastructure provisioning workflows, unblocking Phase 7 infrastructure features.

⸻

## 3. User Stories (v1)

### Platform Engineer

- As a platform engineer, I want Stagecraft to automatically provision DigitalOcean droplets based on deployment plans so I don't have to manually create infrastructure.

- As a platform engineer, I want to see what infrastructure changes will be made before applying them so I can review costs and changes.

- As a platform engineer, I want Stagecraft to handle droplet lifecycle (create/delete) automatically so infrastructure matches deployment state.

### Developer

- As a developer deploying to staging/production, I want Stagecraft to provision the required infrastructure automatically so I can focus on application deployment.

- As a developer, I want clear error messages when infrastructure provisioning fails so I can diagnose issues quickly.

- As a developer, I want to preview infrastructure changes before applying them so I can avoid unexpected costs.

### CI / Automation

- As a CI pipeline, I want Stagecraft to provision infrastructure deterministically so deployments are reproducible.

- As a CI pipeline, I want Stagecraft to clean up infrastructure after tests complete so costs are controlled.

⸻

## 4. Inputs and API Contract

### 4.1 CloudProvider Interface

The provider implements the CloudProvider interface from `pkg/providers/cloud/cloud.go`:

```go
type CloudProvider interface {
    ID() string
    Plan(ctx context.Context, opts PlanOptions) (InfraPlan, error)
    Apply(ctx context.Context, opts ApplyOptions) error
}
```

### 4.2 Plan() Operation

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

**Behavior:**

- Side-effect free: No API calls that modify infrastructure
- Deterministic: Same config and environment produce same plan
- Sorted output: ToCreate and ToDelete lists sorted lexicographically by Name
- Reads desired state from config (hosts defined per environment)
- Compares desired state with actual state (via DigitalOcean API)
- Returns plan with differences (ToCreate, ToDelete)

**Config Structure:**

```go
type Config struct {
    TokenEnv     string            `yaml:"token_env"`      // Required: env var name for DO API token (token never stored)
    SSHKeyName   string            `yaml:"ssh_key_name"`   // Required: SSH key name in DO account (must exist, validated via API)
    DefaultRegion string           `yaml:"default_region"` // Optional: default region (e.g., "nyc1")
    DefaultSize   string           `yaml:"default_size"`   // Optional: default size (e.g., "s-2vcpu-4gb")
    Regions       []string         `yaml:"regions"`        // Optional: allowed regions
    Sizes         []string         `yaml:"sizes"`          // Optional: allowed sizes
    Hosts         map[string]HostConfig `yaml:"hosts"`    // Required: host definitions per environment
}
```

**Config Ownership Model:**

- Token is never stored in config or state; only environment variable name (`token_env`) is stored
- SSH key must exist in DigitalOcean account; provider validates existence via API, fails if not found
- Provider does not create SSH keys; user must pre-configure

type HostConfig struct {
    Role   string `yaml:"role"`   // Required: role (e.g., "gateway", "app", "db")
    Size   string `yaml:"size"`    // Optional: size (defaults to default_size)
    Region string `yaml:"region"` // Optional: region (defaults to default_region)
}
```

**Example Config:**

```yaml
cloud:
  provider: digitalocean
  providers:
    digitalocean:
      token_env: DO_TOKEN
      ssh_key_name: "my-ssh-key"
      default_region: "nyc1"
      default_size: "s-2vcpu-4gb"
      hosts:
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

### 4.3 Apply() Operation

**Input:**

```go
type ApplyOptions struct {
    Config any      // Provider-specific config
    Plan   InfraPlan // Infrastructure plan to apply
}
```

**Behavior:**

- Creates droplets specified in plan.ToCreate
- Deletes droplets specified in plan.ToDelete
- Handles already-existing droplets gracefully (idempotent)
- Handles already-deleted droplets gracefully (idempotent)
- Waits for droplets to be ready before returning (for create operations)
- Processes operations in deterministic order (sorted by Name)

**Droplet Creation Flow:**

1. For each droplet in plan.ToCreate:
   - Check if droplet already exists (by name)
   - If exists and matches spec (region, size), skip (idempotent)
   - If exists but doesn't match spec, return error (reconciliation needed)
   - If doesn't exist, create droplet via DigitalOcean API
   - Poll for droplet status until "active" (async operation)
   - Polling interval: 5 seconds
   - Timeout: 10 minutes
   - Return error if creation fails or times out
   - Partial failures: If some droplets created but others fail, return error describing partial state

**Droplet Deletion Flow:**

1. For each droplet in plan.ToDelete:
   - Check if droplet exists (by name)
   - If doesn't exist, skip (idempotent)
   - If exists, delete droplet via DigitalOcean API
   - Poll for droplet deletion until confirmed
   - Return error if deletion fails or times out

⸻

## 5. Data Structures

### 5.1 DigitalOceanProvider

```go
// internal/providers/cloud/digitalocean/do.go

type DigitalOceanProvider struct {
    client APIClient // DigitalOcean API client (injected for testing)
}

func (p *DigitalOceanProvider) ID() string {
    return "digitalocean"
}

func NewDigitalOceanProvider() *DigitalOceanProvider {
    return &DigitalOceanProvider{
        client: NewDOClient(), // Production client
    }
}

func NewDigitalOceanProviderWithClient(client APIClient) *DigitalOceanProvider {
    return &DigitalOceanProvider{
        client: client, // Test client
    }
}
```

### 5.2 Config

```go
// internal/providers/cloud/digitalocean/config.go

type Config struct {
    TokenEnv     string            `yaml:"token_env"`
    SSHKeyName   string            `yaml:"ssh_key_name"`
    DefaultRegion string           `yaml:"default_region"`
    DefaultSize   string           `yaml:"default_size"`
    Regions       []string         `yaml:"regions"`
    Sizes         []string         `yaml:"sizes"`
    Hosts         map[string]HostConfig `yaml:"hosts"`
}

type HostConfig struct {
    Role   string `yaml:"role"`
    Size   string `yaml:"size"`
    Region string `yaml:"region"`
}
```

### 5.3 APIClient Interface (for testing)

```go
// internal/providers/cloud/digitalocean/client.go

type APIClient interface {
    ListDroplets(ctx context.Context) ([]Droplet, error)
    GetDroplet(ctx context.Context, name string) (*Droplet, error)
    CreateDroplet(ctx context.Context, req CreateDropletRequest) (*Droplet, error)
    DeleteDroplet(ctx context.Context, id int) error
    ListSSHKeys(ctx context.Context) ([]SSHKey, error)
    GetSSHKey(ctx context.Context, name string) (*SSHKey, error)
    WaitForDroplet(ctx context.Context, id int, status string) error
}

type Droplet struct {
    ID     int    `json:"id"`
    Name   string `json:"name"`
    Region string `json:"region"`
    Size   string `json:"size"`
    Status string `json:"status"`
    Networks Networks `json:"networks"`
}

type Networks struct {
    V4 []NetworkV4 `json:"v4"`
}

type NetworkV4 struct {
    IPAddress string `json:"ip_address"`
    Type      string `json:"type"`
}

type CreateDropletRequest struct {
    Name     string
    Region   string
    Size     string
    Image    string // e.g., "ubuntu-22-04-x64"
    SSHKeys  []int  // SSH key IDs
    Tags     []string
}

type SSHKey struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}
```

**Production Implementation:**

```go
type DOClient struct {
    client *godo.Client // github.com/digitalocean/godo
}

func NewDOClient() *DOClient {
    token := os.Getenv("DO_TOKEN") // From config.token_env
    return &DOClient{
        client: godo.NewFromToken(token),
    }
}
```

⸻

## 6. Implementation Details

### 6.1 Plan() Implementation

**Flow:**

1. Parse config from `opts.Config`
2. Validate config (token_env, ssh_key_name, hosts required)
3. Get DigitalOcean API token from environment variable
4. Validate SSH key exists in DigitalOcean account
5. Get desired hosts from config for given environment
6. List actual droplets from DigitalOcean API (filtered by environment tag or name pattern)
7. Compare desired vs actual:
   - ToCreate: desired hosts not in actual
   - ToDelete: actual hosts not in desired
8. Sort ToCreate and ToDelete lexicographically by Name
9. Return InfraPlan

**Determinism:**

- Same config and environment always produces same plan
- Hosts sorted lexicographically by Name
- No timestamps or random data in plan
- Plan is side-effect free (no infrastructure modifications)

**Error Cases:**

- Config validation errors (missing required fields)
- API token missing from environment
- SSH key not found in DigitalOcean account
- API failures (rate limits, network errors)
- Invalid region or size values

### 6.2 Apply() Implementation

**Flow:**

1. Parse config from `opts.ApplyOptions.Config`
2. Validate config
3. Get DigitalOcean API token from environment variable
4. Validate SSH key exists in DigitalOcean account
5. Get SSH key ID from DigitalOcean API
6. Process plan.ToCreate:
   - For each host spec:
     - Check if droplet exists (by name)
     - If exists and matches spec, skip (idempotent)
     - If exists but doesn't match spec, return error
     - If doesn't exist, create droplet
     - Wait for droplet to be "active"
7. Process plan.ToDelete:
   - For each host spec:
     - Find droplet by name
     - If doesn't exist, skip (idempotent)
     - If exists, delete droplet
     - Wait for deletion to complete
8. Return nil on success

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

### 6.3 Config Parsing

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
        return nil, fmt.Errorf("invalid digitalocean provider config: %w", err)
    }
    
    // Validate required fields
    if config.TokenEnv == "" {
        return nil, fmt.Errorf("token_env is required")
    }
    if config.SSHKeyName == "" {
        return nil, fmt.Errorf("ssh_key_name is required")
    }
    if len(config.Hosts) == 0 {
        return nil, fmt.Errorf("hosts configuration is required")
    }
    
    return &config, nil
}
```

### 6.4 Droplet Naming Convention

**Format:**

- Environment prefix: `{environment}-{hostname}`
- Example: `staging-app-1`, `prod-gateway-1`

**Rationale:**

- Enables filtering droplets by environment
- Prevents name conflicts across environments
- Makes droplet identification deterministic

⸻

## 7. Error Handling

### 7.1 Error Types

**Error Taxonomy:**

**Config Errors** (local, deterministic, no API calls):
```go
ErrConfigInvalid = errors.New("invalid config")
```

**Authentication Errors** (API calls required):
```go
ErrTokenMissing   = errors.New("API token missing from environment")
ErrSSHKeyNotFound = errors.New("SSH key not found")
```

**Resource Errors** (API operations):
```go
ErrDropletExists      = errors.New("droplet already exists")
ErrDropletNotFound    = errors.New("droplet not found")
ErrDropletCreateFailed = errors.New("droplet creation failed")
ErrDropletDeleteFailed = errors.New("droplet deletion failed")
ErrDropletTimeout     = errors.New("droplet operation timeout")
```

**API Errors** (infrastructure/rate limiting):
```go
ErrAPIError  = errors.New("DigitalOcean API error")  // wraps underlying API errors
ErrRateLimit = errors.New("API rate limit exceeded")   // with retry logic
```

### 7.2 Error Messages

- Config invalid: `"digitalocean provider: invalid config: {reason}"`
- Token missing: `"digitalocean provider: API token missing from environment variable {env_var}"`
- SSH key not found: `"digitalocean provider: SSH key '{name}' not found in DigitalOcean account"`
- Droplet exists: `"digitalocean provider: droplet '{name}' already exists"`
- Droplet not found: `"digitalocean provider: droplet '{name}' not found"`
- API error: `"digitalocean provider: API error: {error}"`
- Rate limit: `"digitalocean provider: API rate limit exceeded, retrying..."`

⸻

## 8. Testing Strategy

### 8.1 Unit Tests

**File:** `internal/providers/cloud/digitalocean/do_test.go`

**Test Cases:**

1. **ID()**:
   - Returns "digitalocean"

2. **Plan()**:
   - Generates plan with ToCreate for new hosts
   - Generates plan with ToDelete for removed hosts
   - Generates empty plan when desired matches actual
   - Handles config validation errors
   - Handles API failures
   - Sorts ToCreate and ToDelete lexicographically
   - Is deterministic (same inputs produce same output)

3. **Apply()**:
   - Creates droplets in plan.ToCreate
   - Deletes droplets in plan.ToDelete
   - Handles already-existing droplets (idempotent)
   - Handles already-deleted droplets (idempotent)
   - Waits for droplet status
   - Handles creation failures
   - Handles deletion failures
   - Handles API rate limits (with retry)

4. **Config Parsing**:
   - Valid config
   - Missing required fields
   - Invalid YAML
   - Missing hosts configuration

**Mocking:**

- Use fake APIClient implementation for DigitalOcean API
- Mock API responses for ListDroplets, CreateDroplet, DeleteDroplet
- Mock SSH key lookup
- Mock droplet status polling

### 8.2 Integration Tests (Optional)

**File:** `internal/providers/cloud/digitalocean/do_integration_test.go`

**Build Tag:** `//go:build integration`

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

## 9. Package Structure

```
internal/providers/cloud/digitalocean/
├── do.go              # Provider implementation
├── do_test.go         # Unit tests
├── config.go          # Config struct and parsing
├── client.go          # DigitalOcean API client interface and implementation
└── errors.go          # Error definitions
```

⸻

## 10. Registry Integration

**Registration:**

```go
// internal/providers/cloud/digitalocean/do.go

func init() {
    cloud.Register(NewDigitalOceanProvider())
}
```

**Import:**

The package must be imported somewhere in the main binary to trigger `init()`. This is typically done in `cmd/stagecraft/main.go` with an anonymous import:

```go
import _ "stagecraft/internal/providers/cloud/digitalocean"
```

⸻

## 11. Config Schema Example

**stagecraft.yml:**

```yaml
cloud:
  provider: digitalocean
  providers:
    digitalocean:
      token_env: DO_TOKEN
      ssh_key_name: "my-ssh-key"
      default_region: "nyc1"
      default_size: "s-2vcpu-4gb"
      hosts:
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

⸻

## 12. Determinism Guarantees

- **Plan()**: Pure function, deterministic output for given inputs
- **Config parsing**: Deterministic YAML unmarshaling
- **Host sorting**: ToCreate and ToDelete sorted lexicographically by Name
- **Error messages**: Stable error messages without timestamps or random data
- **Droplet naming**: Deterministic naming convention (environment-hostname)

⸻

## 13. Implementation Checklist

- [ ] Create package structure: `internal/providers/cloud/digitalocean/`
- [ ] Implement Config struct and parsing
- [ ] Implement APIClient interface and DOClient implementation
- [ ] Implement DigitalOceanProvider struct with ID()
- [ ] Implement Plan() with deterministic planning
- [ ] Implement Apply() with idempotent operations
- [ ] Define error types
- [ ] Register provider in init()
- [ ] Write unit tests (approximately 70% coverage, with all critical paths and error modes covered)
- [ ] Write integration tests (optional, gated by env var)
- [ ] Update config validation if needed
- [ ] Document usage in spec file

⸻

## 14. Dependencies

### Go Dependencies

- `gopkg.in/yaml.v3` - Config parsing (already in go.mod)
- `github.com/digitalocean/godo` - DigitalOcean API client (to be added)
- `stagecraft/pkg/providers/cloud` - Cloud provider interface
- `stagecraft/pkg/logging` - Logging helpers

### External Dependencies

- DigitalOcean API v2 accessible
- DigitalOcean API token in environment variable
- SSH key configured in DigitalOcean account

⸻

## 15. Success Metrics

- Provider registers successfully and can be retrieved from registry
- Plan() generates deterministic infrastructure plans
- Apply() creates and deletes droplets successfully
- Idempotent operations work correctly
- Async operations handled correctly (polling, timeouts)
- Partial failures surfaced correctly (no silent failures)
- Unit tests achieve approximately 70% coverage, with all critical paths and error modes covered
- All error paths are tested
- Integration tests pass (when run with valid API token)

⸻

## 16. Cost and Billing Responsibility

**Operators are responsible for DigitalOcean billing and cost control.**

Stagecraft does not perform cost estimation, quota enforcement, or billing management. When Apply() creates droplets, they immediately incur DigitalOcean charges. Operators must:

- Review Plan() output before applying to understand infrastructure changes
- Monitor DigitalOcean billing dashboard for costs
- Set up DigitalOcean billing alerts independently
- Ensure sufficient quota/limits before running Apply()

Stagecraft provides no safeguards against accidental cost overruns beyond the dry-run Plan() operation.

## 16. Cost and Billing Responsibility

**Operators are responsible for DigitalOcean billing and cost control.**

Stagecraft does not perform cost estimation, quota enforcement, or billing management. When Apply() creates droplets, they immediately incur DigitalOcean charges. Operators must:

- Review Plan() output before applying to understand infrastructure changes
- Monitor DigitalOcean billing dashboard for costs
- Set up DigitalOcean billing alerts independently
- Ensure sufficient quota/limits before running Apply()

Stagecraft provides no safeguards against accidental cost overruns beyond the dry-run Plan() operation.

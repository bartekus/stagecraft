---
feature: CLI_INFRA_UP
version: v1
status: done
domain: commands
inputs:
  flags: []
outputs:
  exit_codes:
    "0": "All hosts created and bootstrap succeeded"
    "10": "Some hosts failed bootstrap (partial failure)"
    "1": "Config error"
    "2": "CloudProvider failure"
    "3": "Global bootstrap error"
---

# CLI_INFRA_UP – Infrastructure Up

**Feature**: CLI_INFRA_UP  
**Phase**: Phase 7 – Infrastructure  
**Status**: done

⸻

## 1. Overview

`stagecraft infra up` provisions infrastructure using a configured CloudProvider and then bootstraps those hosts using the `INFRA_HOST_BOOTSTRAP` engine.

This command is deterministic, idempotent, and provider-agnostic.

v1 targets **DigitalOcean** cloud provider and **Tailscale** network provider only, but the command MUST remain provider-agnostic in its implementation.

⸻

## 2. Responsibilities and Non-Goals

### 2.1 Responsibilities (v1)

`CLI_INFRA_UP` MUST:

- Load configuration via `CORE_CONFIG` from `stagecraft.yml`
- Resolve CloudProvider from provider registry using `cloud.provider` config
- Validate NetworkProvider exists (used by bootstrap, not directly by CLI)
- Call CloudProvider `Plan()` to determine infrastructure changes
- Call CloudProvider `Apply()` to create or reconcile hosts
- Map provider-specific host metadata to internal `Host` model (as defined in `INFRA_HOST_BOOTSTRAP` spec)
  - Note: v1 assumes hosts are SSH-ready after `Apply()` completes. Explicit SSH readiness polling is deferred to a future version.
- Invoke `INFRA_HOST_BOOTSTRAP` service with mapped hosts
- Print deterministic, sorted results to console
- Return appropriate exit code based on success/failure

### 2.2 Non-Goals (v1)

`CLI_INFRA_UP` MUST NOT in v1:

- Delete hosts (handled by `CLI_INFRA_DOWN`)
- Support multiple cloud providers simultaneously (DigitalOcean only)
- Support multiple network providers (Tailscale only)
- Provide interactive confirmations before creating infrastructure
- Persist long-term state beyond in-memory results
- Implement automatic retry strategies (retries are manual re-runs)
- Perform cost estimation or budgeting
- Manage infrastructure monitoring or alerting

⸻

## 3. Command Invocation

### 3.1 CLI Syntax

```bash
stagecraft infra up
```

### 3.2 Flags

No flags in v1.

Future flags (ignored in v1):
- `--no-bootstrap` - Skip bootstrap step
- `--dry-run` - Show what would be created without creating
- `--verbose` - Detailed output

### 3.3 Environment

The command respects global flags:
- `--config` - Path to `stagecraft.yml` (default: `./stagecraft.yml`)
- `--env` - Environment name (default: from config or `default`)
- `--verbose` - Verbose logging
- `--dry-run` - Dry-run mode (may be supported in v1.1)

⸻

## 4. Execution Semantics

### 4.1 Execution Flow

When invoked, `CLI_INFRA_UP` performs the following steps:

1. **Load configuration** via `CORE_CONFIG` from `stagecraft.yml`
2. **Resolve CloudProvider** from provider registry using `cloud.provider` config value
3. **Validate NetworkProvider** exists (used by bootstrap, validated early for clear errors)
4. **Call CloudProvider `Plan()`** to determine infrastructure changes:
   - Returns `InfraPlan` with `ToCreate` and `ToDelete` lists
   - Side-effect free (no API calls that modify infrastructure)
5. **Call CloudProvider `Apply()`** to create or reconcile hosts:
   - Creates hosts specified in `plan.ToCreate`
   - Handles already-existing hosts gracefully (idempotent)
   - Waits for hosts to be "active" (provider responsibility)
6. **Wait for hosts to reach SSH-ready state**:
   - Poll loop checking SSH connectivity
   - Timeout after reasonable duration (e.g., 5 minutes)
   - Proceed to bootstrap once all hosts are SSH-ready
7. **Collect host metadata** from CloudProvider:
   - Host ID, name, role, public IP, tags
   - Provider-specific metadata
8. **Map provider output → internal Host model**:
   - Convert to `bootstrap.Host` type (as defined in `INFRA_HOST_BOOTSTRAP` spec)
   - Sort deterministically by `Host.ID`
9. **Invoke bootstrap engine**:
   - Call `bootstrap.Service.Bootstrap(ctx, hosts, cfg.Infra.Bootstrap)`
   - Bootstrap runs even if some hosts already exist (idempotent)
10. **Format and print results** to console:
    - Deterministic, sorted output
    - Per-host success/failure status
11. **Return exit code** based on success/failure rules

### 4.2 Idempotency Requirements

`CLI_INFRA_UP` MUST be idempotent:

- Re-running the command MUST NOT recreate existing hosts
- CloudProvider `Apply()` is responsible for idempotency (handles already-existing hosts)
- Bootstrap is idempotent (handles already-bootstrapped hosts)
- Output MUST be deterministic (same config → same output)

### 4.3 Error Semantics

The command returns:

- Exit code `0` when all hosts are created and bootstrap succeeds
- Exit code `10` when some hosts fail bootstrap (partial failure)
- Exit code `1` when configuration is invalid
- Exit code `2` when CloudProvider `Plan()` or `Apply()` fails
- Exit code `3` when bootstrap service fails at global level

Per-host failures are encoded in bootstrap `Result`, not as CLI errors.

⸻

## 5. Host Model

### 5.1 Internal Host Model

The command uses the `Host` model defined in `INFRA_HOST_BOOTSTRAP` spec:

```go
type Host struct {
    ID       string   // Stable unique identifier from CloudProvider
    Name     string   // Human-readable name (e.g. "app-1")
    Role     string   // Logical role (e.g. "app", "db", "proxy")
    PublicIP string   // IPv4 used for initial SSH connectivity
    Tags     []string // Provider or user-defined tags
}
```

### 5.2 Mapping Requirements

Provider output MUST be mapped deterministically:

- `ProviderHost.ID` → `Host.ID`
- `ProviderHost.Name` → `Host.Name`
- `ProviderHost.Role` → `Host.Role`
- `ProviderHost.PublicIP` → `Host.PublicIP`
- `ProviderHost.Tags` → `Host.Tags`

After mapping, hosts MUST be sorted by `Host.ID` ascending before passing to bootstrap.

⸻

## 6. Provider Interaction

### 6.1 CloudProvider Interface

The command uses the CloudProvider interface from `spec/providers/cloud/interface.md`:

```go
type CloudProvider interface {
    ID() string
    Plan(ctx context.Context, opts PlanOptions) (InfraPlan, error)
    Apply(ctx context.Context, opts ApplyOptions) error
}
```

### 6.2 NetworkProvider Validation

The command validates that the configured NetworkProvider exists but does not call it directly:

```go
networkProvider, err := network.Get(cfg.Network.ProviderID)
if err != nil {
    return fmt.Errorf("network provider %q not found: %w", cfg.Network.ProviderID, err)
}
```

NetworkProvider is used by bootstrap, not by the CLI directly.

### 6.3 Provider-Specific Behavior

For DigitalOcean (`PROVIDER_CLOUD_DO`):

- Creates droplets with Ubuntu 22.04 image
- Configures SSH keys from DigitalOcean account
- Tags droplets with `stagecraft` and `stagecraft-env-<env>` tags
- Waits for droplets to reach "active" status before returning

⸻

## 7. Bootstrap Integration

### 7.1 Bootstrap Invocation

After mapping hosts, the command invokes bootstrap:

```go
bootstrapSvc := bootstrap.NewService()
bootstrapResult, err := bootstrapSvc.Bootstrap(ctx, hosts, cfg.Infra.Bootstrap)
```

### 7.2 Bootstrap Configuration

Bootstrap configuration is loaded from `stagecraft.yml`:

```yaml
infra:
  bootstrap:
    ssh:
      user: "root"
      port: 22
    docker:
      enabled: true
      installMethod: "apt"
    network:
      provider: "tailscale"
```

### 7.3 Bootstrap Results

Bootstrap returns a `Result` with per-host outcomes:

```go
type Result struct {
    Hosts []HostResult
}

type HostResult struct {
    HostID    string
    Name      string
    Status    Status  // "success", "failed", "skipped"
    ErrorCode string
    Message   string
}
```

The command uses this result to:
- Format console output
- Compute exit code
- Report per-host failures

⸻

## 8. Output Format

### 8.1 Console Output

The command prints deterministic, sorted output:

```
Infrastructure provisioning complete:

Created hosts:
- app-1 (12345): success
- db-1 (12346): success
- worker-1 (12347): failed (docker_install_failed)
```

Ordering MUST be deterministic (sorted by `Host.ID` ascending).

### 8.2 Output Requirements

- No timestamps in deterministic output
- No random identifiers
- Sorted by `Host.ID` ascending
- Clear per-host status (success/failed with error code)
- Concise error messages

⸻

## 9. Configuration

### 9.1 Required Config

The command requires:

```yaml
cloud:
  provider: digitalocean  # Required: CloudProvider ID
  providers:
    digitalocean:
      token_env: DO_TOKEN  # Required: Environment variable for API token
      ssh_key_name: "my-ssh-key"  # Required: SSH key name in DigitalOcean
      hosts:
        staging:  # Environment name
          app-1:
            role: app
            size: "s-2vcpu-4gb"
            region: "nyc1"

network:
  provider: tailscale  # Required: NetworkProvider ID

infra:
  bootstrap:  # Required: Bootstrap configuration
    ssh:
      user: "root"
    docker:
      enabled: true
    network:
      provider: "tailscale"
```

### 9.2 Config Validation

The command MUST validate:

- `cloud.provider` exists in cloud provider registry
- `network.provider` exists in network provider registry
- `infra.bootstrap` is present and valid
- Cloud provider config is valid (provider-specific validation)

⸻

## 10. Error Conditions

### 10.1 Global Errors

Global errors (returning non-zero exit code) occur when:

- Configuration is invalid (exit code `1`)
  - Unknown CloudProvider ID
  - Unknown NetworkProvider ID
  - Missing required config fields
- CloudProvider `Plan()` fails (exit code `2`)
  - API errors
  - Rate limits
  - Invalid provider config
- CloudProvider `Apply()` fails (exit code `2`)
  - Droplet creation failures
  - API errors
  - Timeouts
- Bootstrap service fails at global level (exit code `3`)
  - Invalid bootstrap config
  - NetworkProvider not found in bootstrap config

### 10.2 Per-Host Errors

Per-host failures are encoded in bootstrap `Result`, not as CLI errors:

- SSH failures → `HostResult.Status = failed`, `ErrorCode = "ssh_failed"`
- Docker failures → `HostResult.Status = failed`, `ErrorCode = "docker_*"`
- Network failures → `HostResult.Status = failed`, `ErrorCode = "network_*"`

Partial failures result in exit code `10`, not a global error.

⸻

## 11. Determinism Guarantees

Given the same:

- Configuration (`stagecraft.yml`)
- Environment name
- External state (cloud provider state)

The command MUST:

- Process hosts in deterministic order
- Produce deterministic output (sorted by `Host.ID`)
- Return deterministic exit codes
- Avoid timestamps or random values in output

When external state changes (e.g., hosts already exist), behavior may differ, but there MUST be no internal source of non-determinism.

⸻

## 12. Testing Requirements

The following scenarios MUST be covered by tests in `internal/cli/commands/infra_up_test.go`:

1. **Provider resolution**
   - Valid provider ID → provider resolved successfully
   - Invalid provider ID → error returned

2. **Host creation**
   - CloudProvider creates hosts successfully
   - Already-existing hosts handled gracefully

3. **Host mapping**
   - Provider output → internal Host model
   - Deterministic ordering (sorted by ID)

4. **Bootstrap integration**
   - Bootstrap invoked with correct hosts
   - Bootstrap results used for exit code computation

5. **Exit code logic**
   - All hosts succeed → exit code 0
   - Some hosts fail → exit code 10
   - Global error → exit code 1/2/3

6. **Deterministic output**
   - Output is sorted and deterministic
   - No timestamps or random values

Coverage for the command MUST meet or exceed the thresholds enforced by `scripts/check-coverage.sh --fail-on-warning`.

⸻

## 13. Related Features

- `PROVIDER_CLOUD_DO` - DigitalOcean CloudProvider implementation
- `PROVIDER_NETWORK_TAILSCALE` - Tailscale NetworkProvider implementation
- `INFRA_HOST_BOOTSTRAP` - Host bootstrap engine
- `CLI_INFRA_DOWN` - Infrastructure teardown command (future)
- `CORE_CONFIG` - Config loading and validation

⸻

## 14. Versioning

- **Version**: v1
- **Scope**: DigitalOcean only; Tailscale only; Ubuntu 22.04 hosts only
- **Changes to behaviour MUST update this spec and corresponding tests before implementation changes.**

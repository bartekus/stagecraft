# CLI_INFRA_UP – Implementation Outline

> This document defines the v1 implementation plan for `CLI_INFRA_UP`. It translates the feature analysis brief into a concrete, testable, spec aligned delivery plan.

> All details in this outline must be reflected in `spec/commands/infra-up.md` before any tests or code are written.

**Status:** v1 implemented. Further work:
- SSH readiness polling for new hosts (currently assumes hosts are SSH-ready after Apply)
- Richer network config integration
- Future Phase 7 features (INFRA_FIREWALL, INFRA_VOLUME_MGMT, etc.)

⸻

## 1. Feature Summary

**Feature ID:** CLI_INFRA_UP  
**Domain:** commands

**Goal:**

Implement the `stagecraft infra up` command that provisions infrastructure hosts through a cloud provider and bootstraps them. The command performs:

1. Config loading
2. Provider resolution
3. Host creation / reconciliation
4. Mapping cloud provider output → internal host model
5. Delegation to `INFRA_HOST_BOOTSTRAP`
6. Display of deterministic results

**v1 Scope:**

- Single cloud provider per project (DigitalOcean only)
- Single network provider per project (Tailscale only)
- Ubuntu 22.04 hosts only
- Deterministic host creation and naming
- Idempotent operations
- Integration with bootstrap engine
- Minimal flags (no interactive confirmations)

**Out of scope for v1:**

- Host deletion (handled by `CLI_INFRA_DOWN`)
- Multi-cloud configurations
- Interactive confirmations
- Long-lived state persistence
- Autoscaling or dynamic host counts

**Future extensions (not implemented in v1):**

- Support for AWS, GCP, Azure
- Multi-region deployments
- Interactive mode with confirmations
- Dry-run mode (may be added in v1.1)

⸻

## 2. Problem Definition and Motivation

Stagecraft needs a way to:

- Create infrastructure hosts deterministically
- Map provider-specific output to a normalized host model
- Integrate with bootstrap engine seamlessly
- Provide clear, deterministic output to users

Without `CLI_INFRA_UP`, Stagecraft cannot provision infrastructure, blocking all Phase 7 features.

⸻

## 3. Execution Model and Orchestration

### 3.1 Command Structure

```go
// internal/cli/commands/infra_up.go

// Feature: CLI_INFRA_UP
// Spec: spec/commands/infra-up.md

func newInfraUpCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "up",
        Short: "Provision infrastructure for an environment",
        Long:  "Create infrastructure hosts using the configured cloud provider and bootstrap them.",
        RunE:  runInfraUp,
    }
    
    // No flags in v1
    
    return cmd
}
```

### 3.2 Execution Flow

```go
func runInfraUp(cmd *cobra.Command, args []string) error {
    // 1. Load config via CORE_CONFIG
    cfg, err := config.Load(cmd.Flag("config").Value.String())
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    
    // 2. Resolve CloudProvider
    cloudProvider, err := cloud.Get(cfg.Cloud.ProviderID)
    if err != nil {
        return fmt.Errorf("cloud provider %q not found: %w", cfg.Cloud.ProviderID, err)
    }
    
    // 3. Resolve NetworkProvider (validate, not used directly)
    _, err = network.Get(cfg.Network.ProviderID)
    if err != nil {
        return fmt.Errorf("network provider %q not found: %w", cfg.Network.ProviderID, err)
    }
    
    // 4. Create hosts via CloudProvider
    plan, err := cloudProvider.Plan(ctx, cloud.PlanOptions{
        Config:      cfg.Cloud.Providers[cfg.Cloud.ProviderID],
        Environment: cfg.Environment,
    })
    if err != nil {
        return fmt.Errorf("failed to plan infrastructure: %w", err)
    }
    
    err = cloudProvider.Apply(ctx, cloud.ApplyOptions{
        Config:      cfg.Cloud.Providers[cfg.Cloud.ProviderID],
        Environment: cfg.Environment,
        Plan:        plan,
    })
    if err != nil {
        return fmt.Errorf("failed to create infrastructure: %w", err)
    }
    
    // 5. Wait for hosts to reach SSH-ready state (poll loop)
    hosts, err := waitForHostsReady(ctx, cloudProvider, cfg)
    if err != nil {
        return fmt.Errorf("failed to wait for hosts: %w", err)
    }
    
    // 6. Map provider output → internal Host model
    infraHosts := mapToInfraHosts(hosts)
    
    // 7. Invoke bootstrap engine
    bootstrapSvc := bootstrap.NewService()
    bootstrapResult, err := bootstrapSvc.Bootstrap(ctx, infraHosts, cfg.Infra.Bootstrap)
    if err != nil {
        return fmt.Errorf("bootstrap failed: %w", err)
    }
    
    // 8. Format + print results to console
    printResults(infraHosts, bootstrapResult)
    
    // 9. Return exit code based on success/failure rules
    return computeExitCode(bootstrapResult)
}
```

### 3.3 Host Mapping

```go
// mapToInfraHosts converts provider-specific host metadata to internal Host model
func mapToInfraHosts(providerHosts []ProviderHost) []bootstrap.Host {
    hosts := make([]bootstrap.Host, 0, len(providerHosts))
    
    for _, ph := range providerHosts {
        hosts = append(hosts, bootstrap.Host{
            ID:       ph.ID,
            Name:     ph.Name,
            Role:     ph.Role,
            PublicIP: ph.PublicIP,
            Tags:     ph.Tags,
        })
    }
    
    // Sort deterministically by ID
    sort.Slice(hosts, func(i, j int) bool {
        return hosts[i].ID < hosts[j].ID
    })
    
    return hosts
}
```

⸻

## 4. Config and Data Structures

### 4.1 Config Schema

Config is read via `CORE_CONFIG` from `stagecraft.yml`:

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

network:
  provider: tailscale
  providers:
    tailscale:
      auth_key_env: TAILSCALE_AUTH_KEY
      tailnet_domain: "mytailnet.ts.net"

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

### 4.2 Host Model

Matches `INFRA_HOST_BOOTSTRAP` spec exactly:

```go
// internal/infra/bootstrap/host.go (from bootstrap spec)

type Host struct {
    ID       string   // Stable unique identifier from CloudProvider
    Name     string   // Human-readable name (e.g. "app-1")
    Role     string   // Logical role (e.g. "app", "db", "proxy")
    PublicIP string   // IPv4 used for initial SSH connectivity
    Tags     []string // Provider or user-defined tags
}
```

### 4.3 Provider Output Model

```go
// ProviderHost represents host metadata from CloudProvider
type ProviderHost struct {
    ID       string
    Name     string
    Role     string
    PublicIP string
    Tags     []string
}
```

⸻

## 5. Determinism Rules and Side Effect Guarantees

### 5.1 Deterministic Inputs

- Config values are immutable during a single run
- Provider `Plan()` output is deterministic (provider responsibility)
- Host list must be sorted before processing

### 5.2 Deterministic Outputs

- `Result.Hosts` must be returned in sorted order (by `Host.ID`)
- Console output must be deterministic (no timestamps, no random values)
- Exit codes must be deterministic given the same inputs

### 5.3 Idempotency Guarantees

- Re-running `infra up` must not recreate existing hosts (provider responsibility)
- Host mapping must be deterministic (same provider output → same Host model)
- Bootstrap invocation must be idempotent (bootstrap responsibility)

⸻

## 6. Error Handling and Reporting

### 6.1 Global Errors

Global errors (returning non-nil error from command) are reserved for:

- Invalid configuration (e.g., unknown CloudProvider ID)
- CloudProvider `Plan()` or `Apply()` failures
- NetworkProvider validation failures
- Bootstrap service initialization failures

### 6.2 Per-Host Errors

Per-host failures are encoded in bootstrap `Result`, not as CLI errors:

- SSH failures → `HostResult.Status = failed`, `ErrorCode = "ssh_failed"`
- Docker failures → `HostResult.Status = failed`, `ErrorCode = "docker_*"`
- Network failures → `HostResult.Status = failed`, `ErrorCode = "network_*"`

### 6.3 Exit Codes

Provisional exit codes (final numbers may change once `GOV_CLI_EXIT_CODES` lands):

- `0` - All hosts created and bootstrap succeeded
- `10` - Some hosts failed bootstrap (partial failure)
- `1` - Config error
- `2` - CloudProvider failure
- `3` - Global bootstrap error

⸻

## 7. Required Tests

All tests live under `internal/cli/commands/infra_up_test.go`.

### 7.1 Unit Tests

1. **Provider resolution**
   - Valid provider ID → provider resolved successfully
   - Invalid provider ID → error returned

2. **Host mapping**
   - Provider output → internal Host model
   - Deterministic ordering (sorted by ID)
   - All fields mapped correctly

3. **Exit code computation**
   - All hosts succeed → exit code 0
   - Some hosts fail → exit code 10
   - Global error → exit code 1/2/3

4. **Error classification**
   - Config errors → exit code 1
   - Provider errors → exit code 2
   - Bootstrap errors → exit code 3

### 7.2 Integration Tests

Using fake providers:

1. **All hosts created successfully**
   - CloudProvider creates hosts
   - Bootstrap succeeds for all hosts
   - Exit code 0

2. **Partial bootstrap failure**
   - CloudProvider creates hosts
   - Bootstrap fails for some hosts
   - Exit code 10

3. **Provider Apply() failure**
   - CloudProvider.Apply() fails
   - Command returns error (exit code 2)

4. **Config error**
   - Invalid config → error returned (exit code 1)

### 7.3 Golden Output Tests

- Deterministic, sorted output snapshot tests
- Verify output format matches spec

⸻

## 8. Completion Criteria

`CLI_INFRA_UP` is considered done when:

1. `docs/engine/analysis/CLI_INFRA_UP.md` (Analysis Brief) is complete and stable.
2. This Implementation Outline is complete and consistent with the Analysis Brief.
3. `spec/commands/infra-up.md` is written and matches this outline for v1.
4. Implementation exists in `internal/cli/commands/infra_up.go` with required header comments:
   - `// Feature: CLI_INFRA_UP`
   - `// Spec: spec/commands/infra-up.md`
5. Tests in `internal/cli/commands/infra_up_test.go` cover:
   - Provider resolution
   - Host mapping
   - Exit code logic
   - Error classification
   - Integration with bootstrap (using fakes)
   - Deterministic ordering
6. `go test ./...` passes.
7. `./scripts/check-coverage.sh --fail-on-warning` passes and coverage for `internal/cli/commands` meets thresholds.
8. `spec/features.yaml` is updated to mark `CLI_INFRA_UP` as `done` only after all of the above.

⸻

## 9. Implementation Slices

### Slice 1: Command Structure and Config Loading
- Create command structure
- Load config via `CORE_CONFIG`
- Validate config structure

### Slice 2: Provider Resolution
- Resolve CloudProvider from registry
- Validate NetworkProvider exists
- Error handling for missing providers

### Slice 3: Host Creation
- Call CloudProvider `Plan()` and `Apply()`
- Wait for hosts to be SSH-ready
- Error handling for provider failures

### Slice 4: Host Mapping
- Map provider output → internal Host model
- Deterministic sorting
- Tests for mapping correctness

### Slice 5: Bootstrap Integration
- Invoke bootstrap service
- Handle bootstrap results
- Error handling for bootstrap failures

### Slice 6: Output and Exit Codes
- Format console output
- Compute exit codes
- Golden output tests

⸻

## 10. Related Documentation

- **Feature Catalog**: `spec/features.yaml`
- **Analysis Brief**: `docs/engine/analysis/CLI_INFRA_UP.md`
- **Spec**: `spec/commands/infra-up.md`
- **Cloud Provider Interface**: `spec/providers/cloud/interface.md`
- **DigitalOcean Provider Spec**: `spec/providers/cloud/digitalocean.md`
- **Bootstrap Spec**: `spec/infra/bootstrap.md`

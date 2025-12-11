# INFRA_HOST_BOOTSTRAP – Implementation Outline

This document defines the v1 implementation plan for `INFRA_HOST_BOOTSTRAP`.

It is derived from the Analysis Brief and will drive the spec (`spec/infra/bootstrap.md`), tests, and implementation.

**Status:** v1 implemented. Further work:
- Additional bootstrap steps (firewall, volumes, etc.)
- Support for non-Ubuntu distributions
- Enhanced error recovery and retry strategies

⸻

## 1. Feature Summary and v1 Scope

### 1.1 Summary

`INFRA_HOST_BOOTSTRAP` is the infrastructure-level feature responsible for preparing newly created hosts for Stagecraft deployments.

Given a set of infrastructure hosts (created by `CLI_INFRA_UP` via `CloudProvider`), bootstrap will:

- Connect to each host via SSH
- Ensure Docker is installed and running
- Ensure Tailscale is installed and the host is joined to the tailnet via `NetworkProvider`
- Report per-host success/failure in a deterministic way

This feature is **not** a CLI command itself; it is an internal infra engine invoked by `CLI_INFRA_UP` (and possibly by future commands such as `CLI_INFRA_DOWN` or `CLI_STATUS`).

### 1.2 v1 Scope (Included)

- **OS baseline**: Ubuntu 22.04 LTS only
- **Host discovery**: Operates on a host list provided by `CLI_INFRA_UP` (from `CloudProvider`)
- **Connectivity**: Uses public IPv4 and SSH keys for initial access
- **Docker**:
  - Detect whether Docker is installed
  - Install Docker if missing (using apt-based script)
  - Verify Docker is functional (`docker version` / `docker ps` or equivalent)
- **Tailscale** (via `PROVIDER_NETWORK_TAILSCALE`):
  - Ensure Tailscale client is installed (`EnsureInstalled`)
  - Ensure host is joined to tailnet (`EnsureJoined`)
- **Idempotency**:
  - Safe to re-run bootstrap on already bootstrapped hosts
- **Per-host results**:
  - Expose a structured result set (success, failure, reason per host)

### 1.3 Out of Scope (v1)

- Non-Ubuntu distributions (e.g. Debian, CentOS, Alpine)
- Direct firewall management (belongs to `INFRA_FIREWALL`)
- Volume creation and attachment (belongs to `INFRA_VOLUME_MGMT`)
- SSH key management (assumed preconfigured by `PROVIDER_CLOUD_DO` + infra config)
- Persistent bootstrap state tracking beyond in-memory results
- Automatic retry strategies (v1 is single-pass; retries are manual re-runs)
- Cloud-specific bootstrap (cloud-init, user-data, etc.)

### 1.4 Reserved for Future Versions

- OS-specific bootstrap variants (Debian, RHEL, Alpine)
- Pluggable bootstrap steps (user-defined pre/post hooks)
- Persistent infra state including bootstrap history
- More advanced health checks (e.g. disk space, CPU flags, kernel modules)

⸻

## 2. Execution Model and Orchestration

### 2.1 Entry Point

Bootstrap will be implemented as an internal infra service / package, e.g.:

```go
// internal/infra/bootstrap.go

// Feature: INFRA_HOST_BOOTSTRAP
// Spec: spec/infra/bootstrap.md

type Service interface {
    Bootstrap(ctx context.Context, hosts []Host, cfg Config) (*Result, error)
}
```

`CLI_INFRA_UP` will:

1. Call CloudProvider to create hosts or reconcile existing ones.
2. Construct the `[]Host` slice using metadata from CloudProvider.
3. Call `bootstrap.Service.Bootstrap(...)` with the host set and infra config.

### 2.2 Processing Model

- All hosts are sorted deterministically before processing (by `Host.ID` or `Host.Name`).
- Bootstrap may process hosts sequentially for v1 to avoid concurrency complexity.
- Future versions can add controlled concurrency if needed.
- For each host:
  1. Establish SSH connection (using host public IP and configured SSH user).
  2. Run Docker detection/install sequence.
  3. Run Tailscale install/join via NetworkProvider.
  4. Run verification checks.
  5. Append a `HostResult` to the overall `Result` structure.

Bootstrap returns:

- A `Result` structure (list of per-host results).
- A single error only for catastrophic conditions (e.g. invalid configuration, systematic provider failure).
- Per-host failures are encoded inside `Result`, not via top-level error.

### 2.3 Idempotency Rules

- Docker install sequence must:
  - Check for existing Docker before installing.
  - Treat "already installed" as success.
- Tailscale operations must:
  - Use NetworkProvider's idempotent `EnsureInstalled` and `EnsureJoined` operations.
- Verification steps (e.g. `docker version`) must not modify host state.

⸻

## 3. Config and Data Structures

### 3.1 Config Schema (Go Types)

Config is read via `CORE_CONFIG` from `stagecraft.yml` and mapped into:

```go
// internal/infra/bootstrap/config.go

// Config defines bootstrap-level configuration derived from stagecraft.yml.
type Config struct {
    // SSH configuration for initial bootstrap (before Tailscale).
    SSH SSHConfig

    // Docker installation configuration.
    Docker DockerConfig

    // Network configuration (delegated to NetworkProvider).
    Network NetworkConfig
}

type SSHConfig struct {
    User           string   // e.g. "root" or "ubuntu"
    PrivateKeyPath string   // path to private key, if needed
    Port           int      // usually 22
    KnownHostsFile string   // optional, for strict host key checking
    ExtraArgs      []string // for advanced usage, if needed in v1
}

type DockerConfig struct {
    Enabled       bool   // if false, docker bootstrap is skipped
    InstallMethod string // "apt" | "script" | "skip"
    Version       string // "latest" or specific version (informational in v1)
}

type NetworkConfig struct {
    ProviderID string // e.g. "tailscale"
    // Provider-specific config is opaque and lives in core config / providers.
}
```

### 3.2 Host Model

```go
// internal/infra/bootstrap/host.go

type Host struct {
    ID       string   // stable unique id (from CloudProvider)
    Name     string   // human-readable name (e.g. "app-1")
    Role     string   // logical role, e.g. "app", "db", "proxy"
    PublicIP string   // IPv4 used for initial SSH
    Tags     []string // provider tags, used for logging or targeting
}
```

`CLI_INFRA_UP` is responsible for mapping provider-level host info into this `Host` type.

### 3.3 Result Model

```go
// internal/infra/bootstrap/result.go

type Status string

const (
    StatusSuccess Status = "success"
    StatusFailed  Status = "failed"
    StatusSkipped Status = "skipped"
)

type HostResult struct {
    HostID string // Host.ID
    Name   string // Host.Name
    Status Status
    // Short, deterministic error category, e.g. "ssh_failed", "docker_install_failed".
    ErrorCode string
    // Human-readable detail; may include provider message, but avoid unstable data.
    Message string
}

type Result struct {
    Hosts []HostResult
}
```

`Result` ordering must be deterministic (sorted by `HostID` or `Name` before returning).

⸻

## 4. Provider Boundaries and Interactions

### 4.1 CloudProvider (PROVIDER_CLOUD_DO)

- Used only by `CLI_INFRA_UP` to create hosts and fetch metadata.
- `INFRA_HOST_BOOTSTRAP` does not call CloudProvider directly.
- Interface for host metadata is defined at `spec/providers/cloud/interface.md` and is consumed by `CLI_INFRA_UP`.

### 4.2 NetworkProvider (PROVIDER_NETWORK_TAILSCALE)

Bootstrap will use the NetworkProvider registry:

```go
// pkg/providers/network/registry.go

provider, err := network.Get(cfg.Network.ProviderID)
if err != nil {
    // fail fast: misconfiguration
}
```

Expected methods (based on the existing interface):

- `EnsureInstalled(ctx, host)` - install Tailscale client via SSH or provider-specific mechanism
- `EnsureJoined(ctx, host)` - join host to tailnet
- `NodeFQDN(host)` - derive deterministic FQDN (used for logs only in v1)

Important: No Tailscale-specific keys, tags, or tailnet configuration live in bootstrap. Those stay in provider config per `PROVIDER_NETWORK_TAILSCALE` spec.

### 4.3 SSH Execution

All SSH execution should reuse existing infra/exec abstractions if available (or be introduced in a reusable way):

- Prefer a small `SSHCommander` interface to abstract `exec.Command` / SSH invocation.
- This interface should live in `internal/infra` and be easily faked for tests.

Example:

```go
type SSHCommander interface {
    Run(ctx context.Context, host Host, script string) (string, error)
}
```

⸻

## 5. Determinism Rules and Side Effect Guarantees

### 5.1 Deterministic Inputs

- `hosts` slice must be sorted before processing.
- Config values are immutable during a single run.

### 5.2 Deterministic Outputs

- `Result.Hosts` must be returned in sorted order.
- `ErrorCode` must be one of a small, predefined set:
  - `""` (no error)
  - `"ssh_failed"`
  - `"docker_detect_failed"`
  - `"docker_install_failed"`
  - `"docker_verify_failed"`
  - `"network_install_failed"`
  - `"network_join_failed"`
  - `"config_invalid"`
- `Message` should be concise and avoid:
  - Timestamps
  - Random IDs
  - Host-specific ephemeral data unless necessary for debugging

### 5.3 Idempotency Guarantees

- Docker detection before installation:
  - If Docker is installed and working, skip install step and mark as success.
- Network operations rely on idempotent provider methods:
  - If already installed or joined, these calls must not fail.

⸻

## 6. Error Handling and Reporting

### 6.1 Per-Host vs Global Errors

- Global errors (returning non-nil error from `Bootstrap`) are reserved for:
  - Invalid configuration (e.g. unknown NetworkProvider ID)
  - Internal programming errors (nil arguments, etc.)
- Per-host errors are recorded as `HostResult.Status = StatusFailed` with a specific `ErrorCode`.

### 6.2 SSH Failure Modes

- Connection timeout
- Authentication failure
- Host unreachable

All of the above map to `ErrorCode = "ssh_failed"` with a short message summarizing the underlying error.

### 6.3 Docker Failure Modes

- Could not determine if Docker is installed → `"docker_detect_failed"`
- Installation command failed → `"docker_install_failed"`
- Verification command failed → `"docker_verify_failed"`

### 6.4 Network Failure Modes

- `EnsureInstalled` failure → `"network_install_failed"`
- `EnsureJoined` failure → `"network_join_failed"`

⸻

## 7. Required Tests

All tests live under `internal/infra/bootstrap_test.go` (and supporting `_test` files if needed).

### 7.1 Unit Tests

1. **Single host – happy path**
   - Docker not installed → install + verify
   - Network install + join succeed
   - Result: one `HostResult` with `StatusSuccess`.

2. **Already bootstrapped host**
   - Docker detection finds Docker already installed and working.
   - Network provider reports idempotent success.
   - Result: success; no errors.

3. **SSH failure**
   - `SSHCommander` returns error on initial connection.
   - Result: one `HostResult` with `StatusFailed`, `ErrorCode = "ssh_failed"`.

4. **Docker install failure**
   - Docker detection says "not installed".
   - Install script returns error.
   - Result: `StatusFailed`, `ErrorCode = "docker_install_failed"`.

5. **Network join failure**
   - Docker succeeds.
   - `EnsureInstalled` succeeds.
   - `EnsureJoined` fails.
   - Result: `StatusFailed`, `ErrorCode = "network_join_failed"`.

6. **Multiple hosts – partial failure**
   - Host A bootstraps successfully.
   - Host B fails Docker install.
   - Result: two `HostResult`s with mixed success/failure; order deterministic.

7. **Deterministic ordering**
   - Provide hosts in random order.
   - Verify that `Result.Hosts` is always sorted by `Host.Name` or `Host.ID`.

### 7.2 Integration Tests (Future / Optional v1.1)

- End-to-end test with fake CloudProvider + NetworkProvider and real `SSHCommander` into a Dockerized VM could be added later. For v1, keep to unit-level tests with fakes.

⸻

## 8. Completion Criteria

`INFRA_HOST_BOOTSTRAP` is considered done when:

1. `docs/engine/analysis/INFRA_HOST_BOOTSTRAP.md` (Analysis Brief) is complete and stable.
2. This Implementation Outline is complete and consistent with the Analysis Brief.
3. `spec/infra/bootstrap.md` is written and matches this outline for v1.
4. Implementation exists in `internal/infra/bootstrap.go` (and related files) with required header comments:
   - `// Feature: INFRA_HOST_BOOTSTRAP`
   - `// Spec: spec/infra/bootstrap.md`
5. Tests in `internal/infra/bootstrap_test.go` cover:
   - Happy path (Docker + Tailscale)
   - Already-bootstrapped host
   - SSH failure
   - Docker failure
   - Network failure
   - Partial failure across multiple hosts
   - Deterministic ordering
6. `go test ./...` passes.
7. `./scripts/check-coverage.sh --fail-on-warning` passes and coverage for `internal/infra` meets thresholds.
8. `spec/features.yaml` is updated to mark `INFRA_HOST_BOOTSTRAP` as `done` only after all of the above.

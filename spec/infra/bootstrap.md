---
feature: INFRA_HOST_BOOTSTRAP
version: v1
status: done
domain: infra
inputs:
  flags: []
outputs:
  exit_codes: {}
---

# SPDX-License-Identifier: AGPL-3.0-or-later
#
# Feature: INFRA_HOST_BOOTSTRAP
#
# Host bootstrap (Docker, Tailscale, etc.) for infrastructure hosts created via CLI_INFRA_UP.
#
# This spec defines the v1 contract for the INFRA_HOST_BOOTSTRAP feature.

# INFRA_HOST_BOOTSTRAP – Host bootstrap for infra hosts

## 1. Overview

`INFRA_HOST_BOOTSTRAP` is the infrastructure-level feature responsible for preparing newly
created hosts for Stagecraft deployments.

Given:

- A set of infrastructure hosts created by `CLI_INFRA_UP` via `CloudProvider`, and
- Bootstrap configuration derived from `stagecraft.yml`,

the bootstrap engine:

1. Connects to each host via SSH using its public IP.
2. Ensures Docker is installed and running.
3. Ensures Tailscale is installed and the host is joined to the tailnet via `NetworkProvider`.
4. Produces a deterministic, per-host result set describing bootstrap outcomes.

This feature is **not** a CLI command on its own; it is an internal infra service that will be
invoked by `CLI_INFRA_UP` as part of the infra provisioning flow.

v1 targets **Ubuntu 22.04 LTS** hosts only.

---

## 2. Responsibilities and Non-Goals

### 2.1 Responsibilities (v1)

`INFRA_HOST_BOOTSTRAP` MUST:

- Accept a list of hosts created by a `CloudProvider` implementation (e.g. `PROVIDER_CLOUD_DO`).
- Use SSH to connect to each host using its public IPv4 address and configured SSH user.
- Detect whether Docker is installed and functional; if not, install it using an apt-based method.
- Verify Docker is usable (for example by running `docker version` or `docker ps`).
- Use a configured `NetworkProvider` (e.g. `PROVIDER_NETWORK_TAILSCALE`) to:
  - Ensure the Tailscale client is installed on each host.
  - Ensure each host is joined to the tailnet.
- Be **idempotent**:
  - Re-running bootstrap MUST NOT break already-bootstrapped hosts.
  - Existing Docker and Tailscale installations MUST be treated as success.
- Return a deterministic `Result` describing success or failure for each host.

### 2.2 Non-Goals (v1)

`INFRA_HOST_BOOTSTRAP` MUST NOT in v1:

- Support non-Ubuntu distributions (only Ubuntu 22.04 is in scope).
- Manage firewall configuration; that belongs to `INFRA_FIREWALL`.
- Manage Docker volumes; that belongs to `INFRA_VOLUME_MGMT`.
- Manage SSH keys, SSH agent configuration, or credential provisioning.
- Persist long-term bootstrap state beyond the in-memory `Result` of a run.
- Implement automatic retry strategies; retries are performed by re-running the caller (for example `CLI_INFRA_UP`).
- Perform cloud-provider specific bootstrapping such as cloud-init or user-data scripts.

---

## 3. Invocation and Execution Semantics

### 3.1 Invocation

Bootstrap is implemented as an internal infra service, for example:

```go
// internal/infra/bootstrap.go

// Feature: INFRA_HOST_BOOTSTRAP
// Spec: spec/infra/bootstrap.md

type Service interface {
    // Bootstrap prepares the given hosts for Stagecraft deployments.
    Bootstrap(ctx context.Context, hosts []Host, cfg Config) (*Result, error)
}
```

`CLI_INFRA_UP` MUST:

1. Call CloudProvider to create or reconcile infrastructure hosts.
2. Map provider-specific host metadata into `[]Host` as defined in this spec.
3. Load `Config` from `stagecraft.yml` (via `CORE_CONFIG`).
4. Call `Service.Bootstrap(ctx, hosts, cfg)`.

This spec does not define a standalone CLI; it defines the internal behaviour invoked by `CLI_INFRA_UP`.

### 3.2 Host Processing Model

For a single invocation of `Bootstrap`:

1. The `hosts` slice MUST be sorted deterministically (for example by `Host.ID` ascending) before any work is performed.
2. v1 MAY process hosts sequentially; concurrency is not required.
3. For each host:
   a. Establish an SSH session using `Host.PublicIP`, `cfg.SSH.User`, and configured SSH settings.
   b. Run Docker detection; if Docker is missing, run the Docker install sequence.
   c. Verify Docker is functional.
   d. Use the configured NetworkProvider to ensure Tailscale is installed and joined to the tailnet.
   e. Append a `HostResult` for this host to the `Result`.
4. After all hosts are processed, `Result.Hosts` MUST be sorted deterministically (for example by `Host.ID` ascending) before returning.

### 3.3 Idempotency Requirements

Bootstrap MUST be idempotent:

- If Docker is already installed and functional, the Docker step MUST succeed without altering host state.
- If Tailscale is already installed and the host is already joined to the tailnet, the NetworkProvider calls MUST succeed in an idempotent manner (the provider interface is responsible for idempotency).
- Re-running Bootstrap for the same host set and the same configuration MUST not produce additional side effects beyond those required to reconcile a partially configured host.

### 3.4 Error Semantics

Bootstrap returns:

- `(*Result, nil)` when configuration is valid and the engine runs to completion; individual host failures are represented as `HostResult.Status = failed`.
- `(nil, error)` only when there is a global error that prevents bootstrap from executing meaningfully. This includes:
  - Invalid configuration (for example NetworkProvider ID not found).
  - Programmatic or internal errors (for example nil dependencies).

Per-host failure MUST NOT result in a non-nil top-level error; it MUST be encoded in the `Result`.

⸻

## 4. Inputs

### 4.1 Host Model

The host model is the internal representation passed to bootstrap from `CLI_INFRA_UP`:

```go
type Host struct {
    ID       string   // Stable unique identifier from CloudProvider.
    Name     string   // Human-readable name (for example "app-1").
    Role     string   // Logical role (for example "app", "db", "proxy").
    PublicIP string   // IPv4 used for initial SSH connectivity.
    Tags     []string // Provider or user-defined tags.
}
```

Requirements:

- `ID` MUST be stable across runs for the same host.
- `PublicIP` MUST be routable from the machine running Stagecraft during initial bootstrap.
- `Role` and `Tags` MAY be used for logging and future targeting; v1 does not interpret them beyond that.

### 4.2 Config Model (Go Types)

The bootstrap configuration is loaded from `stagecraft.yml` via `CORE_CONFIG` and mapped into:

```go
type Config struct {
    SSH     SSHConfig
    Docker  DockerConfig
    Network NetworkConfig
}

type SSHConfig struct {
    User           string   // For example "root" or "ubuntu".
    PrivateKeyPath string   // Optional; empty means defaults from core executil or SSH config.
    Port           int      // Default 22 if zero.
    KnownHostsFile string   // Optional; empty disables strict host checking in v1.
    ExtraArgs      []string // Optional extra SSH arguments; v1 MAY ignore this if not needed.
}

type DockerConfig struct {
    Enabled       bool   // If false, Docker steps are skipped.
    InstallMethod string // "apt", "script", or "skip".
    Version       string // "latest" or a specific version; informational only in v1.
}

type NetworkConfig struct {
    ProviderID string // For example "tailscale".
    // Provider specific configuration remains opaque and lives in provider config.
}
```

If `DockerConfig.Enabled` is false, bootstrap MUST NOT attempt to install Docker but MUST still run the network step.

If `NetworkConfig.ProviderID` is empty or refers to a non-existent provider, Bootstrap MUST return `(nil, error)` with an error classified as `config_invalid`.

### 4.3 Config Schema (YAML Fragment)

The YAML representation in `stagecraft.yml` MUST serialize to the types above. A typical fragment:

```yaml
infra:
  bootstrap:
    ssh:
      user: "root"
      port: 22
      privateKeyPath: "~/.ssh/id_rsa"
      knownHostsFile: "~/.ssh/known_hosts"
    docker:
      enabled: true
      installMethod: "apt"    # "apt" | "script" | "skip"
      version: "latest"
    network:
      provider: "tailscale"
```

- The exact config path (for example `infra.bootstrap`) is defined in `CORE_CONFIG`; this spec defines the shape, not the wiring.
- `network.provider` MUST map to `NetworkConfig.ProviderID`.

⸻

## 5. Outputs

### 5.1 Status Type

Bootstrap uses a simple status enum to describe per-host outcomes:

```go
type Status string

const (
    StatusSuccess Status = "success"
    StatusFailed  Status = "failed"
    StatusSkipped Status = "skipped"
)
```

### 5.2 Error Codes

Per-host failures MUST be encoded using a small, fixed set of error codes:

- `""` (empty) – no error.
- `"ssh_failed"` – bootstrap could not establish SSH connectivity.
- `"docker_detect_failed"` – could not determine whether Docker is installed.
- `"docker_install_failed"` – Docker installation failed.
- `"docker_verify_failed"` – Docker verification failed (for example `docker version`).
- `"network_install_failed"` – NetworkProvider `EnsureInstalled` failed.
- `"network_join_failed"` – NetworkProvider `EnsureJoined` failed.
- `"config_invalid"` – configuration prevented bootstrap from running correctly for this host.

Specs MAY be extended in the future with additional codes; v1 MUST restrict itself to this set.

### 5.3 HostResult and Result Types

```go
type HostResult struct {
    HostID    string // Corresponds to Host.ID.
    Name      string // Corresponds to Host.Name.
    Status    Status
    ErrorCode string // One of the error codes defined above.
    Message   string // Human readable detail; kept concise and stable.
}

type Result struct {
    Hosts []HostResult
}
```

Requirements:

- `Result.Hosts` MUST be sorted deterministically (for example by `HostID` ascending) when returned.
- `Message` SHOULD avoid embedding timestamps, random identifiers, or highly variable content.
- `StatusSkipped` MAY be used when Docker or network bootstrap is disabled via configuration.

⸻

## 6. Behaviour Details

### 6.1 SSH Connectivity

For each host:

1. Bootstrap MUST attempt to open an SSH connection using:
   - `Host.PublicIP`
   - `cfg.SSH.User`
   - `cfg.SSH.Port` (default 22 if zero)
   - `cfg.SSH.PrivateKeyPath` when provided, or default SSH configuration when empty.
2. On any connectivity or authentication error, bootstrap MUST:
   - Stop further steps for this host.
   - Emit a `HostResult` with:
     - `Status = StatusFailed`
     - `ErrorCode = "ssh_failed"`
     - `Message` summarizing the error.

Bootstrap MUST then proceed to the next host.

### 6.2 Docker Detection and Installation

For each host with an established SSH session and `cfg.Docker.Enabled = true`:

1. Bootstrap MUST attempt to detect Docker:
   - Typically by running a command like `docker version` or `docker info`.
2. If detection succeeds, the Docker step is considered successful; no installation is performed.
3. If detection fails and `cfg.Docker.InstallMethod = "skip"`, bootstrap MUST:
   - Emit `StatusFailed` with `ErrorCode = "docker_detect_failed"` for this host.
   - Skip network steps or continue; the outline defines that deployment requires Docker, so v1 SHOULD treat this as failure.
4. If detection fails and `cfg.Docker.InstallMethod` is `"apt"` or `"script"`:
   - Bootstrap MUST run the corresponding install sequence (for example apt-based install).
   - If installation fails, bootstrap MUST emit:
     - `StatusFailed` with `ErrorCode = "docker_install_failed"`.
   - If installation succeeds, bootstrap MUST re-run the detection or verification step; failure here MUST be reported as:
     - `StatusFailed` with `ErrorCode = "docker_verify_failed"`.

### 6.3 NetworkProvider Integration (Tailscale)

For each host that has:

- A successful SSH connection, and
- Either Docker installed or Docker disabled via config,

bootstrap MUST call the configured NetworkProvider instance:

1. `EnsureInstalled(ctx, host)`; failure MUST produce:
   - `StatusFailed` with `ErrorCode = "network_install_failed"`.
2. `EnsureJoined(ctx, host)`; failure MUST produce:
   - `StatusFailed` with `ErrorCode = "network_join_failed"`.

Both methods MUST be invoked through the `pkg/providers/network` registry using `cfg.Network.ProviderID`.
This spec does not prescribe the exact Go signatures; that is defined in the provider interface spec.

If the NetworkProvider cannot be resolved from `cfg.Network.ProviderID`, bootstrap MUST return
a global `(nil, error)` with an error classified conceptually as `config_invalid`.

### 6.4 Determinism

Given the same:

- `hosts` slice (same contents),
- `cfg` (same settings), and
- external state (for example host OS state, network availability),

bootstrap MUST:

- Process hosts in a deterministic order.
- Produce `Result.Hosts` in that same deterministic order.
- Restrict `ErrorCode` values to the enumerated set.
- Avoid timestamps or random values in `Message`.

When external state changes (for example host OS state), behaviour may differ, but there MUST be no
internal source of non-determinism.

⸻

## 7. Error Conditions and Validation

### 7.1 Global Errors

Bootstrap MUST return a non-nil error and a nil `Result` when:

- `cfg.Network.ProviderID` is empty or does not resolve to a known provider.
- Required dependencies (for example provider registry) are not initialized.
- Input arguments are invalid (for example `hosts == nil` when the implementation does not tolerate that).

Global errors SHOULD be treated by callers (for example `CLI_INFRA_UP`) as immediate failures of the infra step.

### 7.2 Per-Host Errors

All per-host failures MUST:

- Set `Status = StatusFailed`.
- Set `ErrorCode` appropriately.
- Set a concise `Message` with a human readable reason.

Per-host failure MUST NOT propagate as a global error; the caller is expected to inspect the `Result`
to decide how to proceed.

⸻

## 8. Testing Requirements

The following scenarios MUST be covered by tests in `internal/infra/bootstrap_test.go`:

1. **Single host – happy path**
   - SSH connects.
   - Docker not installed; install then verify succeeds.
   - Network `EnsureInstalled` and `EnsureJoined` succeed.
   - Result contains a single `HostResult` with `StatusSuccess`.

2. **Already bootstrapped host**
   - Docker detection succeeds immediately.
   - Network provider indicates idempotent success.
   - Result is success; no errors.

3. **SSH failure**
   - `SSHCommander` always fails for the host.
   - Result contains one `HostResult` with `StatusFailed`, `ErrorCode = "ssh_failed"`.

4. **Docker install failure**
   - Docker detection fails.
   - Install command fails.
   - Result contains one `HostResult` with `StatusFailed`, `ErrorCode = "docker_install_failed"`.

5. **Network join failure**
   - Docker steps succeed.
   - Network `EnsureInstalled` succeeds.
   - `EnsureJoined` fails.
   - Result contains one `HostResult` with `StatusFailed`, `ErrorCode = "network_join_failed"`.

6. **Multiple hosts – partial failure**
   - Host A completes successfully.
   - Host B fails Docker install.
   - Result contains two `HostResult` instances; order deterministic; mixed success/failure.

7. **Deterministic ordering**
   - Hosts are provided in random order.
   - Result ordering is deterministic (for example sorted by `HostID`).

Coverage for the bootstrap package MUST meet or exceed the thresholds enforced by
`scripts/check-coverage.sh --fail-on-warning`.

⸻

## 9. Versioning

- **Version**: v1
- **Scope**: Ubuntu 22.04 only; Docker and Tailscale bootstrap; no firewall or volume management.
- **Changes to behaviour MUST update this spec and corresponding tests before implementation changes.**

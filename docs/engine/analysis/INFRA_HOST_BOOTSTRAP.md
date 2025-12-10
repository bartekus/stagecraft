# INFRA_HOST_BOOTSTRAP – Analysis Brief

This document captures the high level motivation, constraints, and success definition for INFRA_HOST_BOOTSTRAP.

It is the starting point for the Implementation Outline and Spec.

This brief must be approved before outline work begins.

⸻

## Problem Statement

After `CLI_INFRA_UP` creates hosts via CloudProvider(s), those hosts are not yet ready for Stagecraft deployments. They need a deterministic, idempotent bootstrap step that installs Docker, configures mesh networking, and validates readiness.

Without host bootstrap, Stagecraft cannot:
- Ensure hosts have Docker installed for containerized deployments
- Configure mesh networking (Tailscale) for secure multi-host communication
- Validate that hosts are ready for deployment operations
- Provide a consistent baseline for all infrastructure hosts

INFRA_HOST_BOOTSTRAP fills this gap by providing a provider-agnostic bootstrap engine that prepares hosts for Stagecraft deployments after infrastructure provisioning.

⸻

## Motivation

### Decouple Infrastructure Creation from Host Preparation

- **Separation of concerns**: Infrastructure creation (via CloudProvider) and host preparation (via bootstrap) are distinct operations with different failure modes and retry semantics.

- **Provider-agnostic core**: Bootstrap logic should work with any CloudProvider and NetworkProvider, not be tied to DigitalOcean or Tailscale specifically.

- **Idempotent operations**: Bootstrap must be safely re-runnable without breaking already-configured hosts.

### Enable Deployment Readiness

- **Docker requirement**: All Stagecraft deployments use Docker Compose. Hosts must have Docker installed and running before deployment can proceed.

- **Mesh networking**: Multi-host deployments require secure private networking. Tailscale (via NetworkProvider) must be installed and configured before hosts can communicate.

- **Consistent baseline**: Every host created by Stagecraft should reach the same known-good state, regardless of which cloud provider created it.

### Enable Higher-Level Features

- **Phase 7 prerequisites**: INFRA_HOST_BOOTSTRAP unblocks Phase 7 infrastructure features:
  - `INFRA_VOLUME_MGMT` requires Docker to be installed
  - `INFRA_FIREWALL` requires hosts to be bootstrapped and reachable
  - `DEPLOY_COMPOSE_GEN` requires Docker and network connectivity

- **Deployment workflows**: Hosts must be bootstrapped before `CLI_DEPLOY` can generate and deploy Compose files.

⸻

## User Roles and Stories

### Platform Engineer

- "When I run `stagecraft infra up`, I want all new hosts to come back ready for deployment (Docker installed, Tailscale joined), without manual SSH."

- "I want bootstrap to be idempotent so re-running `stagecraft infra up` won't break hosts that are already configured."

- "When bootstrap fails on one host, I want clear error messages showing which host failed and why, so I can fix it without affecting other hosts."

### Application Developer

- "When I deploy to a Stagecraft-managed environment, I assume hosts already have Docker and network wiring; I don't want to care about host bootstrap details."

- "I want bootstrap failures to be clearly reported so I know when infrastructure isn't ready for deployment."

### Operator / SRE

- "I want bootstrap to be deterministic - the same hosts and config should always produce the same bootstrap result."

- "I want to see per-host bootstrap status (success/failed) so I can diagnose issues without SSHing to every host."

- "I want bootstrap to handle partial failures gracefully - if one host fails, other hosts should still be bootstrapped successfully."

⸻

## v1 Success Criteria (5–7)

1. **Idempotent Bootstrap**: For every host returned by `CLI_INFRA_UP`, bootstrap runs exactly once and is idempotent. Re-running bootstrap on an already-bootstrapped host must not change its state or fail.

2. **Docker Installation**: Docker is installed and running, or detected as already installed, without failing the run. Bootstrap verifies Docker is functional (e.g., `docker version` succeeds).

3. **Tailscale Configuration**: Tailscale is installed and joined to the tailnet via `PROVIDER_NETWORK_TAILSCALE`. Bootstrap uses NetworkProvider's `EnsureInstalled()` and `EnsureJoined()` methods, respecting provider boundaries.

4. **Per-Host Reporting**: Bootstrap reports a clear per-host result (success / failed with reason). Failures on one host do not prevent bootstrap of other hosts.

5. **Deterministic Behavior**: All bootstrap behavior is deterministic given the same hosts + config. No randomness, timestamps, or machine-dependent output.

6. **Test Coverage**: Tests cover happy path, partial failure (one host fails), and re-run on already bootstrapped host. Test coverage target: 70%+.

7. **Provider Boundaries**: No provider-specific logic leaks into core; all provider calls go through interfaces (`pkg/providers/network`, `pkg/providers/cloud`).

⸻

## Determinism and Side-Effect Constraints

### No Randomness

- No random names; all host IDs and tags come from CloudProvider.
- No random identifiers in error messages or logs.
- No timestamps in deterministic outputs.

### Deterministic Execution

- All SSH execution paths must be deterministic given the same inputs.
- Host list must be sorted before processing (deterministic order).
- SSH command execution order must be deterministic.

### Idempotency Guarantees

- Re-running bootstrap must not change a successfully bootstrapped host.
- Checking for existing installations (Docker, Tailscale) must happen before attempting installation.
- NetworkProvider operations are already idempotent; bootstrap must not break that guarantee.

### No Environment Variable "Magic"

- No environment-variable "magic" inside the core bootstrap engine.
- All configuration comes from `stagecraft.yml` or explicit function parameters.
- Provider-specific config (e.g., Tailscale auth keys) is handled by providers, not bootstrap core.

⸻

## Risks and Architectural Boundaries

### Risks

- **SSH connectivity flakiness**: Hosts may be unreachable via SSH (timeouts, network issues, SSH key misconfiguration). Bootstrap must handle connection failures gracefully and report clear errors.

- **OS/image drift**: Bootstrap assumes Ubuntu 22.04 for v1. Different OS versions or distributions may require different installation commands. Risk is mitigated by v1 scope (Ubuntu 22.04 only).

- **Partial bootstrap failures**: One host may fail while others succeed. Bootstrap must continue processing remaining hosts and report all results.

- **Docker installation failures**: Docker installation may fail due to network issues, package repository problems, or system configuration. Bootstrap must detect and report these failures clearly.

- **Tailscale join failures**: Tailscale join may fail due to invalid auth keys, network issues, or tailnet configuration. Bootstrap must surface NetworkProvider errors clearly.

### Architectural Boundaries

- **Host creation is owned by CloudProvider** (`PROVIDER_CLOUD_DO`), not by bootstrap. Bootstrap receives host information (ID, IP, role, tags) from `CLI_INFRA_UP`, which calls CloudProvider.

- **Tailnet configuration is owned by NetworkProvider** (`PROVIDER_NETWORK_TAILSCALE`). Bootstrap calls NetworkProvider methods but does not manage Tailscale auth keys, tags, or tailnet configuration.

- **Firewall configuration is a separate feature** (`INFRA_FIREWALL`). Bootstrap does not configure firewall rules.

- **Volume management is a separate feature** (`INFRA_VOLUME_MGMT`). Bootstrap does not create or manage Docker volumes.

- **SSH connectivity is assumed**: Bootstrap assumes hosts are reachable via SSH (public IP initially, Tailscale FQDN after network is configured). SSH key management and connectivity verification are outside bootstrap scope.

⸻

## Upstream Dependencies

### Required Features (all done ✅)

- **PROVIDER_CLOUD_DO**: Cloud provider that creates hosts ✅
- **PROVIDER_NETWORK_TAILSCALE**: Network provider for Tailscale installation and configuration ✅
- **CORE_CONFIG**: Config loading and validation ✅
- **CORE_EXECUTIL**: Process execution utilities for SSH commands ✅

### Required Feature (must be implemented first ⚠️)

- **CLI_INFRA_UP**: Must provide a concrete host list for bootstrap to operate on. `CLI_INFRA_UP` must:
  - Call CloudProvider to create hosts
  - Return a deterministic host list (IDs, IPs, tags/roles) that the infra layer can consume
  - Provide host metadata (public IP, role, tags) needed for bootstrap

### Provider Interface Dependencies

- **PROVIDER_CLOUD_DO** must expose host metadata (ID, public IP, tags/role) after host creation.
- **PROVIDER_NETWORK_TAILSCALE** must provide idempotent operations for install/join (`EnsureInstalled()`, `EnsureJoined()`).

### Spec Dependencies

- `spec/providers/cloud/interface.md` - Cloud provider interface spec (already exists)
- `spec/providers/network/interface.md` - Network provider interface spec (already exists)
- `spec/infra/bootstrap.md` - Bootstrap spec (to be created)

### Runtime Dependencies

- SSH access to hosts (public IP initially)
- NetworkProvider config (Tailscale auth key, tailnet domain)
- Docker installation method (apt for Ubuntu 22.04)

⸻

## Alternatives Considered

### Alternative 1: Include Bootstrap in CloudProvider

**Rejected because**: CloudProvider should focus on infrastructure provisioning (create/delete hosts). Host preparation (Docker, networking) is a separate concern with different failure modes and retry semantics.

### Alternative 2: Include Bootstrap in CLI_INFRA_UP

**Rejected because**: Separation of concerns. `CLI_INFRA_UP` orchestrates infrastructure provisioning; bootstrap is a distinct operation that can be tested and reused independently.

### Alternative 3: Use Cloud-Init or User Data Scripts

**Rejected because**: Provider-agnostic approach is preferred. Cloud-init is cloud-provider-specific and doesn't work across all providers. Bootstrap via SSH is provider-agnostic and works with any CloudProvider.

### Alternative 4: Skip Bootstrap, Require Pre-Configured Hosts

**Rejected because**: Defeats Stagecraft's goal of automated infrastructure management. Users would need to manually configure every host, breaking the "infrastructure as code" workflow.

⸻

## Non-Goals (v1)

- Supporting multiple OS distributions (Ubuntu 22.04 only for v1)
- Installing custom packages or system configuration beyond Docker and Tailscale
- Managing SSH keys or SSH configuration
- Configuring firewall rules (handled by `INFRA_FIREWALL`)
- Creating or managing Docker volumes (handled by `INFRA_VOLUME_MGMT`)
- Configuring system services or systemd units beyond Docker and Tailscale
- Handling cloud-provider-specific bootstrap steps (e.g., DigitalOcean-specific configuration)
- Automatic retry of failed bootstrap operations (manual retry via re-running `CLI_INFRA_UP`)
- Bootstrap status tracking or persistence (bootstrap runs on-demand, no state stored)

⸻

## Approval

- Author: [To be filled]
- Reviewer: [To be filled]
- Date: [To be filled]

Once approved, the Implementation Outline may begin.

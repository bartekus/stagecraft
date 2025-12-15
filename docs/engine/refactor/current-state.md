# Current State Extraction: Provider Model

**Date:** 2025-12-13
**Scope:** `pkg/providers`, `internal/providers`, `internal/infra/bootstrap`

## 1. Existing Provider Interfaces

The current abstraction follows a loose "Resource Manager" pattern, grouped by domain.

| Domain | Interface Location | Primary Responsibilities |
| :--- | :--- | :--- |
| **Cloud** | `pkg/providers/cloud/cloud.go` | `Plan(opts)`, `Apply(opts)`, `Hosts(opts)` |
| **Network** | `pkg/providers/network/network.go` | `EnsureInstalled(opts)`, `EnsureJoined(opts)`, `NodeFQDN(host)` |
| **Backend** | `pkg/providers/backend/backend.go` | `Plan`, `Apply`, `Logs`, `Status` (implied from known patterns) |
| **Bootstrap** | `internal/infra/bootstrap/bootstrap.go` | `Bootstrap(ctx, hosts, cfg)` (Service interface) |
| **Migration** | `pkg/migrations/interface.go` | `Plan`, `Apply`, `History` (implied) |

## 2. Implemented Providers & Features

| Provider | Path | Type | Feature ID | Notes |
| :--- | :--- | :--- | :--- | :--- |
| **DigitalOcean** | `internal/providers/cloud/digitalocean` | Cloud | `PROVIDER_CLOUD_DIGITALOCEAN` | Implements `Plan` by querying DO API. `Apply` calls generic DO Go SDK. |
| **Tailscale** | `internal/providers/network/tailscale` | Network | `PROVIDER_NETWORK_TAILSCALE` | `EnsureInstalled` runs shell scripts. `EnsureJoined` runs `tailscale up`. |
| **Bootstrap** | `internal/infra/bootstrap` | Infra | `INFRA_HOST_BOOTSTRAP` | "Service" that orchestrates `Docker` checks + `NetworkProvider` calls. |
| **Encore** | `internal/providers/backend/encore` | Backend | `PROVIDER_BACKEND_ENCORE` | Wraps Encore CLI. |
| **GitHub Actions** | `internal/providers/ci/github` | CI | `PROVIDER_CI_GITHUB` | Generates `.github/workflows`. |
| **EnvFile** | `internal/providers/secrets/envfile` | Secrets | `PROVIDER_SECRETS_ENVFILE` | Reads/Writes `.env`. |

## 3. Determinism Leaks & Boundary Violations

### Critical Findings

1.  **Implicit Discovery in Planning**
    - `CloudProvider.Plan()` often makes live API calls to `list_droplets` (or similar) to calculate the diff.
    - **Violation:** Planning should be a pure function of `(DesiredState, CurrentState)`. Current implementation fetches `CurrentState` *inside* `Plan`.
    - **Consequence:** `Plan` is not reproducible. Flaky network = Flaky Plan.

2.  **Execution Logic inside Providers**
    - `NetworkProvider.EnsureInstalled()` checks for binary existence *and* installs it.
    - **Violation:** Execution should be separated. The "Check" is Discovery. The "Install" is Execution.
    - **Consequence:** Cannot dry-run "Installation" logic. Cannot verify "Check" logic without a host.

3.  **Bootstrap is an Orchestration Orphan**
    - `bootstrap.Service` is effectively a mini-engine that manually calls `executil` and `NetworkProvider`.
    - It bypasses the core `Plan/Apply` loop for "Bootstrapping" tasks.

4.  **No Direct Execution Substrate**
    - Most providers import `exec` or `commands` directly or use `executil`.
    - They don't use a unified `ExecutionProvider` (SSH/Agent/Container).
    - **Consequence:** Hard to swap local execution for SSH execution (though `bootstrap` has a `CommandExecutor` abstraction, it's local to that package).

## 4. Conclusion
The current state is "functional but rigid". To support agentic behavior (pure reasoning + delegated action), we must extract the "Discovery" (Facts) and "Execution" (Runner) concerns out of the opaque Provider logic.

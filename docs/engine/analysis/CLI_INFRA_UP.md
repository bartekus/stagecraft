# CLI_INFRA_UP – Analysis Brief

This document captures the high level motivation, constraints, and success definition for CLI_INFRA_UP.

It is the starting point for the Implementation Outline and Spec.

This brief must be approved before outline work begins.

⸻

## 1. Problem Statement

Stagecraft currently has the ability to:

- Load configuration
- Resolve cloud and network providers
- Bootstrap hosts once they exist

However, Stagecraft cannot yet create infrastructure such as droplets/VMs. This blocks subsequent features including:

- Host bootstrap (`INFRA_HOST_BOOTSTRAP`)
- Deployment workflows
- Volume and firewall management
- Multi-host networking

`CLI_INFRA_UP` fills this gap by providing a deterministic, provider-agnostic entry point for creating infrastructure required for Stagecraft-managed deployments.

⸻

## 2. Motivation

### 2.1 Create Infrastructure in a Provider-Agnostic Way

Cloud provider APIs differ (DigitalOcean, AWS, GCP, etc.), but Stagecraft needs a consistent, deterministic abstraction over:

- Host creation
- Host lookup and reconciliation
- Metadata mapping
- Request/response modeling

`CLI_INFRA_UP` is responsible for normalizing this.

### 2.2 Deterministic, Idempotent Infrastructure Creation

Infrastructure creation must be safe to re-run:

- Existing hosts should not be recreated
- Only missing or divergent hosts should be reconciled
- Output must be deterministic for both humans and automation

### 2.3 Provide Host Metadata to Bootstrap

`INFRA_HOST_BOOTSTRAP` cannot run without:

- Public IP
- Host ID
- Tags/roles
- Provider output

`CLI_INFRA_UP` produces this host model and hands it to bootstrap.

### 2.4 Enable Future Infrastructure Features

`CLI_INFRA_UP` is the root of the entire Phase 7 tree:

```
CLI_INFRA_UP
  └─> INFRA_HOST_BOOTSTRAP
       ├─> INFRA_VOLUME_MGMT
       ├─> INFRA_FIREWALL
       └─> DEPLOY_COMPOSE_GEN
```

Nothing in Phase 7 can begin without `CLI_INFRA_UP`.

⸻

## 3. User Roles and Stories

### Platform Engineer

- "When I run `stagecraft infra up`, I want all required hosts created consistently and predictably."
- "I want `infra up` to print exactly what was created, unchanged, or reconciled."

### Application Developer

- "I do not want to understand cloud provider APIs; I want Stagecraft to create everything needed to deploy my app."

### SRE / Operator

- "I want deterministic host naming and metadata so I can correlate logs and troubleshoot."

⸻

## 4. v1 Success Criteria (5–7)

1. **Deterministic Host Creation**: Same input configuration → same provider request → same host naming and metadata.

2. **Idempotent Reconciliation**: Re-running the command should not recreate or duplicate hosts.

3. **Provider-Agnostic Execution**: All cloud logic goes through provider interfaces.

4. **Deterministic Output Model**: CLI prints a stable, sorted list of created hosts.

5. **Bootstrap Integration**: After creation, the CLI must immediately invoke `INFRA_HOST_BOOTSTRAP`.

6. **Clear Failure Semantics**: Distinguish between global failures vs per-host bootstrap failures.

7. **Test Coverage**: 70% minimum. Tests MUST cover:
   - Provider interactions
   - Mapping provider outputs → internal host model
   - Deterministic ordering
   - Integration with bootstrap

⸻

## 5. Determinism and Side-Effect Constraints

### No Randomness

- No random names; all hostnames must come from config or provider.
- No timestamps in deterministic outputs.
- No random identifiers in error messages or logs.

### Deterministic Execution

- Host list must be sorted before processing (deterministic order).
- Provider requests must be deterministic given the same config.
- Output formatting must be deterministic.

### Idempotency Guarantees

- Re-running `infra up` must not recreate existing hosts.
- Provider operations are already idempotent; CLI must not break that guarantee.

### No Environment Variable "Magic"

- No environment-variable "magic" inside the core CLI logic.
- All configuration comes from `stagecraft.yml` or explicit function parameters.
- Provider-specific config (e.g., API tokens) is handled by providers, not CLI core.

⸻

## 6. Risks and Architectural Boundaries

### Risks

- **Cloud API failures**: Provider APIs may fail or timeout. CLI must handle these gracefully and report clear errors.

- **Partial host creation**: Some hosts may be created while others fail. CLI must report partial state clearly.

- **Provider throttling**: Cloud providers may rate-limit requests. CLI must surface these errors clearly.

- **SSH not ready immediately**: After droplet provisioning, SSH may not be immediately available. CLI must wait for SSH-ready state before invoking bootstrap.

### Architectural Boundaries

- **Host creation is owned by CloudProvider** (`PROVIDER_CLOUD_DO`). CLI orchestrates but does not implement cloud-specific logic.

- **Host bootstrap is owned by INFRA_HOST_BOOTSTRAP**. CLI invokes bootstrap but does not implement bootstrap logic.

- **Host teardown is a separate feature** (`CLI_INFRA_DOWN`). CLI does not delete hosts in v1.

- **Network configuration is owned by NetworkProvider** (`PROVIDER_NETWORK_TAILSCALE`). CLI validates network provider exists but does not configure networking directly.

⸻

## 7. Upstream Dependencies

### Required Features (all done ✅)

- **PROVIDER_CLOUD_DO**: Cloud provider that creates hosts ✅
- **PROVIDER_NETWORK_TAILSCALE**: Network provider for mesh networking ✅
- **CORE_CONFIG**: Config loading and validation ✅
- **INFRA_HOST_BOOTSTRAP**: Analysis Brief, Outline, and Spec complete ✅

### Spec Dependencies

- `spec/providers/cloud/interface.md` - Cloud provider interface spec (already exists)
- `spec/providers/network/interface.md` - Network provider interface spec (already exists)
- `spec/infra/bootstrap.md` - Bootstrap spec (already exists)
- `spec/commands/infra-up.md` - CLI command spec (to be created)

### Runtime Dependencies

- Cloud provider API access (DigitalOcean API token)
- SSH access to created hosts (for bootstrap)
- Network provider config (Tailscale auth key, tailnet domain)

⸻

## 8. Alternatives Considered

### Alternative 1: Include Infrastructure Creation in CloudProvider Only

**Rejected because**: CloudProvider should focus on infrastructure provisioning (create/delete hosts). CLI orchestration, host mapping, and bootstrap integration are separate concerns.

### Alternative 2: Include Bootstrap Logic in CLI_INFRA_UP

**Rejected because**: Separation of concerns. `CLI_INFRA_UP` orchestrates infrastructure provisioning; bootstrap is a distinct operation that can be tested and reused independently.

### Alternative 3: Use External Tools (Terraform, Pulumi)

**Rejected because**: Stagecraft needs tight integration with deployment workflows. External tools add complexity and break the unified CLI experience.

### Alternative 4: Skip Host Mapping, Use Provider Output Directly

**Rejected because**: Provider outputs are provider-specific. A normalized internal host model enables provider-agnostic bootstrap and future features.

⸻

## 9. Non-Goals (v1)

- Supporting multiple cloud providers simultaneously (DigitalOcean only for v1)
- Deleting hosts (handled by `CLI_INFRA_DOWN`)
- Interactive confirmations before creating infrastructure
- Long-lived state persistence beyond in-memory results
- Automatic retry of failed operations (manual retry via re-running command)
- Cost estimation or budgeting (operators are responsible for billing)
- Infrastructure monitoring or alerting (handled by cloud provider or external tools)
- Multi-region deployments (single region per environment for v1)
- Automatic scaling or auto-scaling groups (static infrastructure only)

⸻

## 10. Approval

- Author: [To be filled]
- Reviewer: [To be filled]
- Date: [To be filled]

Once approved, the Implementation Outline may begin.

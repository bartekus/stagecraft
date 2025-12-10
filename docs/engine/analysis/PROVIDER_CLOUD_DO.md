# PROVIDER_CLOUD_DO Feature Analysis Brief

This document captures the high level motivation, constraints, and success definition for PROVIDER_CLOUD_DO.

It is the starting point for the Implementation Outline and Spec.

This brief must be approved before outline work begins.

⸻

## 1. Problem Statement

Stagecraft core can generate deployment plans that reference logical hosts (app-1, db-1, gateway-1), but currently lacks the ability to:

- Provision actual infrastructure (droplets, VMs, instances) on cloud providers
- Plan infrastructure changes before applying them (dry-run)
- Manage infrastructure lifecycle (create, delete, reconcile)
- Integrate infrastructure provisioning with deployment workflows

Without a cloud provider implementation, Stagecraft cannot:
- Execute `stagecraft infra up` to provision infrastructure
- Execute `stagecraft infra down` to tear down infrastructure
- Support Phase 7 infrastructure features that require cloud provisioning
- Enable multi-host deployments that need infrastructure provisioning

PROVIDER_CLOUD_DO fills this gap by implementing the CloudProvider interface for DigitalOcean, enabling Stagecraft to provision and manage DigitalOcean droplets for deployment environments.

⸻

## 2. Motivation

### Infrastructure as Code Integration

- **Automated provisioning**: Developers and platform engineers need Stagecraft to automatically provision infrastructure (droplets) based on deployment plans, eliminating manual infrastructure setup.

- **Dry-run planning**: Before creating expensive infrastructure, users need to see what will be created/deleted via `Plan()` operation. This enables cost estimation and change review.

- **Lifecycle management**: Stagecraft must handle the full lifecycle of infrastructure - creating droplets for new deployments, deleting droplets for teardown, and reconciling desired vs actual state.

### Integration with Deployment Workflows

- **Phase 7 prerequisites**: PROVIDER_CLOUD_DO unblocks Phase 7 infrastructure features:
  - `CLI_INFRA_UP` needs cloud provider to create droplets
  - `CLI_INFRA_DOWN` needs cloud provider to delete droplets
  - `INFRA_HOST_BOOTSTRAP` uses cloud provider to ensure infrastructure exists before bootstrapping

- **Multi-host deployments**: Enables Stagecraft to provision multiple hosts (gateway, app servers, databases) as part of deployment workflows.

- **Provider-agnostic core**: By implementing the CloudProvider interface, Stagecraft core remains provider-agnostic while gaining DigitalOcean-specific capabilities.

### Operational Reliability

- **Idempotent operations**: Cloud provider operations must be idempotent - running Plan() multiple times must produce identical results, and Apply() must handle already-existing or already-deleted resources gracefully.

- **Deterministic planning**: Plan() must produce deterministic output - same config and environment must always produce the same plan.

- **Error handling**: Clear error messages for common failure modes (API failures, rate limits, invalid config, insufficient quota, etc.).

- **Cost awareness**: Plan() enables users to review infrastructure changes before incurring costs.

⸻

## 3. Users and User Stories

### Platform Engineers

- As a platform engineer, I want Stagecraft to automatically provision DigitalOcean droplets based on deployment plans so I don't have to manually create infrastructure.

- As a platform engineer, I want to see what infrastructure changes will be made before applying them so I can review costs and changes.

- As a platform engineer, I want Stagecraft to handle droplet lifecycle (create/delete) automatically so infrastructure matches deployment state.

### Developers

- As a developer deploying to staging/production, I want Stagecraft to provision the required infrastructure automatically so I can focus on application deployment.

- As a developer, I want clear error messages when infrastructure provisioning fails so I can diagnose issues quickly.

- As a developer, I want to preview infrastructure changes before applying them so I can avoid unexpected costs.

### CI / Automation

- As a CI pipeline, I want Stagecraft to provision infrastructure deterministically so deployments are reproducible.

- As a CI pipeline, I want Stagecraft to clean up infrastructure after tests complete so costs are controlled.

⸻

## 4. Success Criteria (v1)

1. **Provider Registration**:
   - Provider registers successfully with ID "digitalocean"
   - Provider can be retrieved from cloud registry
   - Config validation works correctly

2. **Plan() Operation**:
   - Can generate infrastructure plan from config and environment
   - Plan is deterministic (same inputs produce same output)
   - Plan includes ToCreate and ToDelete lists
   - Plan is side-effect free (no API calls that modify infrastructure)
   - Plan output is sorted lexicographically by host name
   - Returns clear errors for invalid config or API failures

3. **Apply() Operation**:
   - Can create droplets specified in plan.ToCreate
   - Can delete droplets specified in plan.ToDelete
   - Handles already-existing droplets gracefully (idempotent)
   - Handles already-deleted droplets gracefully (idempotent)
   - Waits for droplets to be ready before returning (for create operations)
   - Returns clear errors for API failures, rate limits, or quota issues

4. **Config Schema**:
   - Supports required fields: `token_env`, `ssh_key_name`
   - Supports optional fields: `default_region`, `default_size`, `regions`, `sizes`
   - Validates config and returns clear errors for missing required fields
   - Validates region and size values against DigitalOcean API

5. **DigitalOcean API Integration**:
   - Uses DigitalOcean API v2 to create/delete/list droplets
   - Handles API rate limits with retry logic
   - Handles async operations (droplet creation) with polling
   - Uses dependency injection for API client (testable)

6. **Error Handling**:
   - Clear error messages for all failure modes
   - Specific error types for different failure categories (config, API, quota, etc.)
   - No partial state hidden silently

7. **Testing**:
   - Unit tests with approximately 70% coverage, with all critical paths and error modes covered
   - Tests use mocked DigitalOcean API client
   - Tests cover all error paths
   - Optional integration tests with real DigitalOcean API (gated by env var)

⸻

## 5. Risks and Constraints

### External Dependencies

- **DigitalOcean API availability**: Provider depends on DigitalOcean API v2. API outages or rate limits will affect operations.

- **API rate limits**: DigitalOcean API has rate limits. Provider must implement retry logic with exponential backoff.

- **Async operations**: Droplet creation and deletion are async. Provider must poll for droplet status until "active" (creation) or confirmed deletion. If timeout occurs (10 minutes default), Apply() returns error with actionable message. Partial failures are surfaced; v1 does not automatically rollback.

- **Cost implications**: Creating droplets incurs costs. **Operators are responsible for DigitalOcean billing and cost control.** Stagecraft does not perform cost estimation or quota enforcement. Apply() immediately creates billable resources.

### Determinism Constraints

- **No randomness**: Host names, regions, sizes must be deterministic and derived from config.

- **Deterministic Plan()**: Plan() must be a pure function with respect to config and environment (no timestamps, no random data).

- **Idempotent Apply()**: Apply() must be idempotent - running multiple times must produce identical results.

- **Sorted output**: All lists (ToCreate, ToDelete) must be sorted lexicographically by host name.

### Platform Constraints

- **DigitalOcean only**: v1 supports DigitalOcean only. Other cloud providers are deferred to future versions.

- **Droplets only**: v1 supports droplet creation/deletion only. Other resources (load balancers, volumes, etc.) are deferred.

- **SSH key requirement**: Droplets require SSH keys for access. Provider must validate SSH key exists in DigitalOcean account.

### Integration Constraints

- **Phase 7 dependency**: This provider is a prerequisite for Phase 7 infrastructure features.

- **Network provider integration**: Infrastructure provisioning may need to integrate with network provider (e.g., Tailscale) for multi-host deployments.

⸻

## 6. Alternatives Considered

### Alternative 1: Manual Infrastructure Management

**Rejected because**: Requires manual setup for each deployment, defeating Stagecraft's goal of automated orchestration.

### Alternative 2: Use Terraform/Pulumi Instead

**Rejected because**: Stagecraft needs tight integration with deployment workflows. External tools add complexity and break the unified CLI experience.

### Alternative 3: Support Multiple Cloud Providers Simultaneously in v1

**Rejected because**: v1 scope is single provider (DigitalOcean). Multi-provider support is deferred to future versions.

### Alternative 4: Use DigitalOcean CLI Instead of API

**Rejected because**: API-based approach is more reliable, testable, and provides better error handling than CLI-based approach.

⸻

## 7. Dependencies

### Required Features (all done)

- **PROVIDER_CLOUD_INTERFACE**: Cloud provider interface definition ✅
- **CORE_PLAN**: Planning engine for infrastructure planning ✅
- **CORE_CONFIG**: Config loading and validation ✅

### Spec Dependencies

- `spec/providers/cloud/interface.md` - Cloud provider interface spec (already exists)
- `spec/providers/cloud/digitalocean.md` - DigitalOcean provider spec (to be created)

### Runtime Dependencies

- DigitalOcean API token in environment variable
- SSH key configured in DigitalOcean account
- DigitalOcean API v2 accessible

### External Dependencies

- DigitalOcean API Go client library (e.g., `github.com/digitalocean/godo`)

⸻

## 8. Non-Goals (v1)

- Managing other DigitalOcean resources (load balancers, volumes, databases, etc.) - droplets only
- Supporting other cloud providers (AWS, GCP, Azure) - DigitalOcean only
- Managing SSH keys in DigitalOcean account - user must pre-configure (provider validates existence, fails if not found)
- Cost estimation or budgeting - **operators are responsible for billing and cost control**
- Infrastructure monitoring or alerting - handled by DigitalOcean or external tools
- Multi-region deployments - single region per environment for v1
- Automatic scaling or auto-scaling groups - static infrastructure only
- Automatic rollback of partial failures - operator must manually reconcile

⸻

## 9. Approval

- Author: [To be filled]
- Reviewer: [To be filled]
- Date: [To be filled]

Once approved, the Implementation Outline may begin.

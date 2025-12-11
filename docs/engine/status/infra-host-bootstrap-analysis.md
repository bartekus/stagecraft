# INFRA_HOST_BOOTSTRAP Feature Analysis and Next Steps

> **Generated**: Analysis of current feature completion state and next steps for INFRA_HOST_BOOTSTRAP  
> **Source**: `spec/features.yaml` and implementation artifacts  
> **Last Updated**: See `spec/features.yaml` for current status

⸻

## Executive Summary

**Current Project Status**: 43/75 features complete (57.3%)

**INFRA_HOST_BOOTSTRAP Status**: `todo` - Not started (Phase 7: Infrastructure)

**Dependencies**: All dependencies satisfied ✅
- `PROVIDER_CLOUD_DO` - Complete (creates hosts)
- `PROVIDER_NETWORK_TAILSCALE` - Complete (mesh networking)
- `CLI_INFRA_UP` - Required but not yet implemented (will be implemented first)

**Recommendation**: Implement `CLI_INFRA_UP` first, then `INFRA_HOST_BOOTSTRAP` as the next logical step in Phase 7.

⸻

## Current Feature Completion State

### Overall Progress

| Metric | Value |
|--------|-------|
| **Total Features** | 75 |
| **Completed** | 43 (57.3%) |
| **In Progress** | 0 (0.0%) |
| **Planned** | 32 (42.7%) |

### Phase-by-Phase Completion

| Phase | Features | Done | WIP | Todo | Completion | Status |
|-------|----------|------|-----|------|------------|--------|
| **Phase 0: Foundation** | 8 | 8 | 0 | 0 | 100% | ✅ Complete |
| **Phase 1: Provider Interfaces** | 6 | 6 | 0 | 0 | 100% | ✅ Complete |
| **Phase 2: Core Orchestration** | 7 | 7 | 0 | 0 | 100% | ✅ Complete |
| **Phase 3: Local Development** | 10 | 10 | 0 | 0 | 100% | ✅ Complete |
| **Phase 4: Provider Implementations** | 2 | 2 | 0 | 0 | 100% | ✅ Complete |
| **Phase 5: Build and Deploy** | 6 | 4 | 0 | 2 | 67% | 🔄 In progress |
| **Phase 6: Migration System** | 10 | 3 | 0 | 7 | 30% | 🔄 In progress |
| **Phase 7: Infrastructure** | 5 | 0 | 0 | 5 | 0% | ⚠️ Not started |
| **Phase 8: Operations** | 6 | 0 | 0 | 6 | 0% | ⚠️ Not started |
| **Phase 9: CI Integration** | 3 | 0 | 0 | 3 | 0% | ⚠️ Not started |
| **Phase 10: Project Scaffold** | 5 | 0 | 0 | 5 | 0% | ⚠️ Not started |
| **Governance** | 4 | 3 | 0 | 1 | 75% | 🔄 In progress |

### Phase 7: Infrastructure Features

| Feature ID | Title | Status | Dependencies |
|------------|-------|--------|--------------|
| `CLI_INFRA_UP` | `stagecraft infra up` command | todo | `PROVIDER_CLOUD_DO`, `PROVIDER_NETWORK_TAILSCALE` |
| `CLI_INFRA_DOWN` | `stagecraft infra down` command | todo | `PROVIDER_CLOUD_DO` |
| **`INFRA_HOST_BOOTSTRAP`** | **Host bootstrap (Docker, Tailscale, etc.)** | **todo** | **`CLI_INFRA_UP`** |
| `INFRA_VOLUME_MGMT` | Volume management | todo | `CLI_INFRA_UP`, `PROVIDER_CLOUD_DO` |
| `INFRA_FIREWALL` | Firewall configuration | todo | `CLI_INFRA_UP`, `PROVIDER_CLOUD_DO` |

**Current Phase 7 Status**: 0% complete (0/5 features done)

⸻

## INFRA_HOST_BOOTSTRAP Context

### Feature Definition

**Feature ID**: `INFRA_HOST_BOOTSTRAP`  
**Title**: Host bootstrap (Docker, Tailscale, etc.)  
**Status**: `todo`  
**Phase**: Phase 7: Infrastructure  
**Owner**: bart  
**Spec**: `spec/infra/bootstrap.md` (not yet created)  
**Tests**: `internal/infra/bootstrap_test.go` (not yet created)

### Purpose

After cloud providers create infrastructure (droplets/VMs), hosts need to be bootstrapped with:
- **Docker**: Required for containerized deployments
- **Tailscale**: Mesh networking (via NetworkProvider)
- **System configuration**: Any other prerequisites for deployment

This feature ensures hosts are ready for deployment after infrastructure provisioning.

### Dependencies

**Required (all satisfied ✅)**:
- ✅ `PROVIDER_CLOUD_DO` - Creates DigitalOcean droplets
- ✅ `PROVIDER_NETWORK_TAILSCALE` - Provides Tailscale mesh networking
- ⚠️ `CLI_INFRA_UP` - Infrastructure provisioning command (not yet implemented, but required)

**Dependency Chain**:
```
CLI_INFRA_UP (creates hosts)
  └─> INFRA_HOST_BOOTSTRAP (bootstraps hosts)
      └─> INFRA_VOLUME_MGMT (manages volumes)
      └─> INFRA_FIREWALL (configures firewall)
```

### Integration Points

**1. CloudProvider Integration**
- CloudProvider (`PROVIDER_CLOUD_DO`) creates droplets with:
  - Public IP address (for initial SSH access)
  - SSH keys configured
  - Ubuntu 22.04 image
  - Tags for identification

**2. NetworkProvider Integration**
- NetworkProvider (`PROVIDER_NETWORK_TAILSCALE`) provides:
  - `EnsureInstalled()` - Installs Tailscale client
  - `EnsureJoined()` - Joins host to tailnet
  - `NodeFQDN()` - Generates deterministic FQDN

**3. SSH Connectivity**
- Hosts are initially reachable via public IP
- SSH is used to execute bootstrap commands
- After Tailscale is configured, hosts can use Tailscale FQDN

⸻

## Next Steps for INFRA_HOST_BOOTSTRAP

### Prerequisites

**Must be completed first**:
1. ✅ `PROVIDER_CLOUD_DO` - Complete
2. ✅ `PROVIDER_NETWORK_TAILSCALE` - Complete
3. ⚠️ `CLI_INFRA_UP` - **Must be implemented first** (creates hosts before bootstrap)

### Implementation Sequence

**Step 1: Implement CLI_INFRA_UP** (Required first)
- Creates infrastructure via CloudProvider
- Returns host information (IP addresses, hostnames)
- Provides foundation for bootstrap operations

**Step 2: Implement INFRA_HOST_BOOTSTRAP** (This feature)
- Bootstrap hosts after creation
- Install Docker
- Configure Tailscale via NetworkProvider
- Verify bootstrap success

**Step 3: Integrate with CLI_INFRA_UP**
- Call bootstrap after host creation
- Handle bootstrap failures gracefully
- Provide clear error messages

### Implementation Approach

Following the proven pattern from Phase 4 providers:

**1. Feature Analysis Brief**
- File: `docs/engine/analysis/INFRA_HOST_BOOTSTRAP.md`
- Define: Problem, motivation, success criteria, risks, dependencies

**2. Implementation Outline**
- File: `docs/engine/outlines/INFRA_HOST_BOOTSTRAP_IMPLEMENTATION_OUTLINE.md`
- Define: v1 scope, API contract, data structures, testing plan

**3. Spec Creation**
- File: `spec/infra/bootstrap.md`
- Define: Behavioral contract, config schema, error conditions

**4. Implementation (Sliced Approach)**
- **Slice 1**: Config & validation
- **Slice 2**: Docker installation
- **Slice 3**: Tailscale integration (via NetworkProvider)
- **Slice 4**: Bootstrap orchestration

**5. Testing**
- Unit tests for each component
- Integration tests with mock providers
- Test coverage target: 70%+

### Key Design Decisions

**1. Bootstrap Scope (v1)**
- ✅ Docker installation (required for deployments)
- ✅ Tailscale configuration (via NetworkProvider)
- ✅ Basic system verification
- ❌ Out of scope: Custom packages, complex system config

**2. Idempotency**
- Bootstrap operations must be idempotent
- Check if Docker is already installed before installing
- Use NetworkProvider's idempotent operations

**3. Error Handling**
- Clear error messages for bootstrap failures
- Partial failures (some hosts succeed, others fail)
- Retry logic for transient failures

**4. SSH Connectivity**
- Use public IP initially (before Tailscale)
- Switch to Tailscale FQDN after network is configured
- Handle SSH connection failures gracefully

### Expected Workflow

**CLI_INFRA_UP → INFRA_HOST_BOOTSTRAP flow**:

```
1. CLI_INFRA_UP calls CloudProvider.Apply()
   └─> Creates droplets (app-1, db-1, etc.)
   └─> Returns host information (IP addresses)

2. For each created host:
   a. INFRA_HOST_BOOTSTRAP connects via SSH (public IP)
   b. Installs Docker (if not already installed)
   c. Calls NetworkProvider.EnsureInstalled()
   d. Calls NetworkProvider.EnsureJoined()
   e. Verifies Docker is running
   f. Verifies Tailscale is connected

3. Hosts are now ready for deployment
   └─> Docker available for containers
   └─> Tailscale mesh network configured
   └─> Hosts reachable via Tailscale FQDN
```

### Configuration Schema (Expected)

```yaml
infra:
  bootstrap:
    docker:
      version: "latest"  # or specific version
      install_method: "apt"  # apt, script, skip
    network:
      provider: "tailscale"  # Uses NetworkProvider
    # Future: custom packages, system config
```

### Success Criteria

**INFRA_HOST_BOOTSTRAP is complete when**:
- ✅ Docker is installed and running on bootstrapped hosts
- ✅ Tailscale is installed and joined to tailnet
- ✅ Hosts are reachable via Tailscale FQDN
- ✅ Bootstrap operations are idempotent
- ✅ Clear error messages for failures
- ✅ Test coverage meets target (70%+)
- ✅ Spec document is complete and accurate

⸻

## Related Features and Context

### Completed Prerequisites

**PROVIDER_CLOUD_DO** (Complete ✅)
- Creates DigitalOcean droplets
- Returns droplet information (ID, IP, status)
- Handles SSH key configuration
- Test coverage: 80.3%

**PROVIDER_NETWORK_TAILSCALE** (Complete ✅)
- Installs Tailscale client via SSH
- Joins hosts to tailnet with tags
- Generates deterministic FQDNs
- Test coverage: ~68%

### Upcoming Features

**CLI_INFRA_UP** (Must implement first)
- Orchestrates infrastructure provisioning
- Calls CloudProvider to create hosts
- Will call INFRA_HOST_BOOTSTRAP after host creation

**INFRA_VOLUME_MGMT** (Depends on INFRA_HOST_BOOTSTRAP)
- Manages Docker volumes
- Requires Docker to be installed (via bootstrap)

**INFRA_FIREWALL** (Depends on INFRA_HOST_BOOTSTRAP)
- Configures firewall rules
- Requires hosts to be bootstrapped

⸻

## Recommendations

### Immediate Next Steps

1. **Implement CLI_INFRA_UP first** (Required dependency)
   - Creates infrastructure foundation
   - Provides host information for bootstrap

2. **Then implement INFRA_HOST_BOOTSTRAP**
   - Follows proven Phase 4 provider pattern
   - Integrates with existing providers
   - Enables host readiness for deployment

### Implementation Priority

**High Priority** (Unblocks other work):
- `CLI_INFRA_UP` - Foundation for infrastructure features
- `INFRA_HOST_BOOTSTRAP` - Enables host readiness

**Medium Priority** (Depends on bootstrap):
- `INFRA_VOLUME_MGMT` - Requires Docker
- `INFRA_FIREWALL` - Requires bootstrapped hosts

**Lower Priority**:
- `CLI_INFRA_DOWN` - Teardown (can be done later)

### Risk Mitigation

**1. SSH Connectivity**
- Ensure SSH keys are properly configured
- Handle connection timeouts gracefully
- Provide clear error messages

**2. Bootstrap Failures**
- Implement retry logic for transient failures
- Handle partial failures (some hosts succeed)
- Provide detailed error reporting

**3. Idempotency**
- Verify operations are idempotent
- Check state before making changes
- Use NetworkProvider's idempotent operations

⸻

## Related Documentation

- **Feature Catalog**: `spec/features.yaml`
- **Feature Completion Analysis**: `docs/engine/status/feature-completion-analysis.md`
- **Implementation Roadmap**: `docs/narrative/implementation-roadmap.md`
- **Phase 4 Analysis**: `docs/engine/status/phase4-analysis.md`
- **Cloud Provider Interface**: `spec/providers/cloud/interface.md`
- **DigitalOcean Provider Spec**: `spec/providers/cloud/digitalocean.md`
- **Tailscale Provider Spec**: `spec/providers/network/tailscale.md`

⸻

## Summary

**Current State**: Project is 57.3% complete with strong foundation (Phases 0-4 complete). Phase 7 infrastructure features are ready to begin.

**INFRA_HOST_BOOTSTRAP Status**: Ready to implement after `CLI_INFRA_UP` is completed. All dependencies are satisfied.

**Next Action**: Implement `CLI_INFRA_UP` first, then proceed with `INFRA_HOST_BOOTSTRAP` following the proven Phase 4 provider pattern.

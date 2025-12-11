---
adr: 0003
title: Removal of DRIVER_DO – Provider Layer Simplification
status: accepted
date: 2025-12-10
deciders: [bart]
---

# ADR 0003: Removal of DRIVER_DO – Provider Layer Simplification

## Status

**Accepted** - 2025-12-10

## Context

The Stagecraft roadmap originally included `DRIVER_DO` as a separate feature in Phase 4, alongside `PROVIDER_CLOUD_DO`. The relationship between these two features was ambiguous:

- Was DRIVER_DO a thin wrapper around CloudProvider?
- Was it legacy naming that should be consolidated?
- Did it provide additional functionality beyond infrastructure provisioning?

After completing `PROVIDER_CLOUD_DO`, we reviewed the architecture and determined that DRIVER_DO was redundant.

## Decision

**Cancel DRIVER_DO as a feature** and merge its intended functionality into existing features:

1. **Infrastructure Provisioning**: Fully handled by `PROVIDER_CLOUD_DO`
   - Droplet creation/deletion
   - SSH key management
   - Region and size configuration
   - Tagging and metadata

2. **Host-Level Configuration**: Belongs in Phase 7 infrastructure features
   - Docker installation → `INFRA_HOST_BOOTSTRAP`
   - Volume management → `INFRA_VOLUME_MGMT`
   - System configuration → Phase 7 infrastructure features

## Rationale

### Architectural Consistency

- No driver pattern exists for other providers (no DRIVER_AWS, DRIVER_TAILSCALE, etc.)
- The `CloudProvider` interface provides sufficient abstraction
- Adding a driver layer would introduce unnecessary complexity

### Clear Separation of Concerns

- **CloudProvider** (Phase 4): Provisions infrastructure (droplets, networks, etc.)
- **Infrastructure Features** (Phase 7): Configures hosts (Docker, volumes, firewall, etc.)

This separation aligns with the single responsibility principle and makes the architecture more maintainable.

### Phase 4 Completion

Cancelling DRIVER_DO allows Phase 4 to be marked as 100% complete, enabling progression to Phase 7 infrastructure workflows.

## Consequences

### Positive

- ✅ Cleaner architecture with single responsibility per provider
- ✅ Phase 4 marked as complete (100%)
- ✅ Clear path forward for Phase 7 infrastructure features
- ✅ No redundant abstraction layers

### Negative

- None identified - DRIVER_DO was never implemented, so no code removal needed

## Implementation

- Updated `spec/features.yaml`: `DRIVER_DO` status changed to `cancelled` with explanatory notes
- Updated `docs/engine/status/phase4-analysis.md`: Reflects cancellation and Phase 4 completion
- Regenerated status docs: `./scripts/generate-implementation-status.sh`

## References

- `PROVIDER_CLOUD_DO` implementation: `internal/providers/cloud/digitalocean/`
- Phase 7 infrastructure features: `spec/features.yaml` (INFRA_HOST_BOOTSTRAP, INFRA_VOLUME_MGMT, etc.)
- Cloud Provider Interface: `spec/providers/cloud/interface.md`

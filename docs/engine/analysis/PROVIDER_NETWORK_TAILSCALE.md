# PROVIDER_NETWORK_TAILSCALE Feature Analysis Brief

This document captures the high level motivation, constraints, and success definition for PROVIDER_NETWORK_TAILSCALE.

It is the starting point for the Implementation Outline and Spec.

This brief must be approved before outline work begins.

⸻

## 1. Problem Statement

Stagecraft core deals with logical hosts (app-1, db-1, gateway-1) in deployment plans, but currently lacks the ability to:

- Ensure hosts are connected to a private mesh network for secure multi-host communication
- Provide stable FQDNs for hosts that can be used in Compose generation, infra bootstrapping, and operations
- Handle the lifecycle of network client installation and joining hosts to mesh networks

Without a network provider implementation, Stagecraft cannot:
- Deploy applications across multiple hosts with secure private networking
- Generate Compose files that reference hosts by stable FQDNs
- Bootstrap infrastructure where hosts need to communicate privately
- Support Phase 7 infrastructure features that require host networking

PROVIDER_NETWORK_TAILSCALE fills this gap by implementing the NetworkProvider interface for Tailscale, enabling Stagecraft to manage Tailscale mesh networking for deployment hosts.

⸻

## 2. Motivation

### Multi-Host Deployment Support

- **Secure mesh networking**: Hosts need to communicate privately without exposing services to the public internet. Tailscale provides a zero-trust mesh network that Stagecraft can manage.

- **Stable FQDNs**: Stagecraft needs deterministic FQDNs for hosts (e.g., `app-1.mytailnet.ts.net`) that can be used in:
  - Compose generation for cross-host service communication
  - Infrastructure bootstrapping (SSH via Tailscale)
  - Future operations commands (status, logs, ssh)

- **Lifecycle management**: Stagecraft must ensure Tailscale is installed and configured on each host before deployment can proceed.

### Integration with Infrastructure Features

- **Phase 7 prerequisites**: PROVIDER_NETWORK_TAILSCALE unblocks Phase 7 infrastructure features:
  - `CLI_INFRA_UP` needs network provider to join hosts to mesh network
  - `INFRA_HOST_BOOTSTRAP` uses network provider to ensure Tailscale is ready
  - Cross-host Compose generation requires stable FQDNs

- **Provider-agnostic core**: By implementing the NetworkProvider interface, Stagecraft core remains provider-agnostic while gaining Tailscale-specific capabilities.

### Operational Reliability

- **Idempotent operations**: Network provider operations must be idempotent - running EnsureInstalled/EnsureJoined multiple times must produce the same result.

- **Deterministic FQDNs**: NodeFQDN must be a pure function that generates stable, predictable FQDNs without network calls.

- **Error handling**: Clear error messages for common failure modes (SSH failures, invalid auth keys, wrong tailnet, etc.).

⸻

## 3. Users and User Stories

### Platform Engineers

- As a platform engineer, I want Stagecraft to automatically ensure Tailscale is installed and configured on deployment hosts so I don't have to manually manage mesh networking.

- As a platform engineer, I want Stagecraft to generate stable FQDNs for hosts so I can reference them in Compose files and infrastructure configs.

- As a platform engineer, I want Tailscale provider to handle host tagging automatically so ACLs can be applied correctly.

### Developers

- As a developer deploying to multi-host environments, I want Stagecraft to handle Tailscale setup automatically so I can focus on application deployment.

- As a developer, I want clear error messages when Tailscale setup fails so I can diagnose issues quickly.

### CI / Automation

- As a CI pipeline, I want Stagecraft to ensure hosts are on the mesh network before deployment so services can communicate securely.

⸻

## 4. Success Criteria (v1)

1. **Provider Registration**:
   - Provider registers successfully with ID "tailscale"
   - Provider can be retrieved from network registry
   - Config validation works correctly

2. **EnsureInstalled**:
   - Can detect if Tailscale is already installed on a host
   - Can install Tailscale client on Linux hosts (Debian/Ubuntu) via official install script
   - Checks OS compatibility (Linux Debian/Ubuntu) before attempting install
   - Returns ErrUnsupportedOS for unsupported operating systems
   - Is idempotent (does nothing if already installed)
   - Returns clear errors for unsupported OS or install failures

3. **EnsureJoined**:
   - Can detect if host is already joined to correct tailnet with correct tags
   - Can join host to tailnet using auth key from environment variable
   - Can apply tags to nodes (default tags + role-specific tags)
   - Is idempotent (does nothing if already correctly configured)
   - Returns clear errors for invalid auth keys, wrong tailnet, or tag mismatches

4. **NodeFQDN**:
   - Generates deterministic FQDNs from host name and tailnet domain
   - Returns FQDNs in format: `{hostname}.{tailnet_domain}`
   - Is a pure function with respect to config (no network calls, no side effects)
   - Requires config to be set via EnsureInstalled or EnsureJoined first

5. **Config Schema**:
   - Supports required fields: `auth_key_env`, `tailnet_domain`
   - Supports optional fields: `default_tags`, `role_tags`, `install.method`, `install.min_version`
   - Validates config and returns clear errors for missing required fields

6. **Error Handling**:
   - Clear error messages for all failure modes
   - Specific error types for different failure categories (config, install, join, tailnet mismatch)
   - No partial state hidden silently

7. **Testing**:
   - Unit tests with approximately 70% coverage, with all critical paths and error modes covered
   - Tests use mocked SSH/exec commands
   - Tests cover all error paths
   - Optional integration tests with real Tailscale (gated by env var)

⸻

## 5. Risks and Constraints

### External Dependencies

- **Tailscale CLI availability**: Tailscale must be installable on target hosts. v1 supports Linux (Debian/Ubuntu) only.

- **SSH access requirement**: Network provider requires SSH access to hosts before Tailscale is up. This must be documented clearly.

- **Auth key management**: Auth keys must be provided via environment variables, never stored in config files. This is a security requirement.

### Determinism Constraints

- **No randomness**: Hostnames, tags, and FQDNs must be deterministic and derived from config.

- **Pure NodeFQDN**: NodeFQDN must be a pure function with no network calls or side effects.

- **Idempotent operations**: All operations must be idempotent and produce identical results when run multiple times.

### Platform Constraints

- **OS support**: v1 supports Linux (Debian/Ubuntu) only. Other OS support is deferred to future versions.

- **Install method**: v1 uses Tailscale's official install script. Custom install methods are deferred.

### Integration Constraints

- **Phase 7 dependency**: This provider is a prerequisite for Phase 7 infrastructure features.

- **CLI_DEPLOY integration**: CLI_DEPLOY currently has TODOs for network provider integration that will be addressed in future work.

⸻

## 6. Alternatives Considered

### Alternative 1: Manual Tailscale Management

**Rejected because**: Requires manual setup on each host, defeating Stagecraft's goal of automated orchestration.

### Alternative 2: Use Tailscale API Instead of CLI

**Rejected because**: CLI-based approach is simpler for v1, requires less API complexity, and matches the interface contract which assumes SSH-based operations.

### Alternative 3: Support Multiple Network Providers Simultaneously

**Rejected because**: v1 scope is single provider per project. Multi-provider support is deferred to future versions.

⸻

## 7. Dependencies

### Required Features (all done)

- **PROVIDER_NETWORK_INTERFACE**: Network provider interface definition ✅
- **CORE_PLAN**: Planning engine for infrastructure planning ✅
- **CORE_CONFIG**: Config loading and validation ✅
- **CORE_EXECUTIL**: Process execution utilities (for SSH/commands) ✅

### Spec Dependencies

- `spec/providers/network/interface.md` - Network provider interface spec (already exists)
- `spec/providers/network/tailscale.md` - Tailscale provider spec (to be created)

### Runtime Dependencies

- SSH access to target hosts (before Tailscale is up)
- Tailscale auth key in environment variable
- Tailscale CLI installable on target hosts (Linux Debian/Ubuntu for v1)

⸻

## 8. Non-Goals (v1)

- Managing Tailscale ACLs or tailnet configuration (handled by Tailscale admin console)
- Supporting every OS under the sun (Linux Debian/Ubuntu only for v1)
- Managing auth key creation or rotation (user responsibility)
- Dynamic network reconfiguration (static configuration only)
- Multiple network providers per project (single provider only)
- Tailscale API integration (CLI-based approach only)

⸻

## 9. Approval

- Author: [To be filled]
- Reviewer: [To be filled]
- Date: [To be filled]

Once approved, the Implementation Outline may begin.

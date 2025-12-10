# DEV_HOSTS Feature Analysis Brief

This document captures the high level motivation, constraints, and success definition for DEV_HOSTS.

It is the starting point for the Implementation Outline and Spec.

This brief must be approved before outline work begins.

⸻

## 1. Problem Statement

CLI_DEV currently:

- Computes a deterministic dev topology.
- Generates dev files under `.stagecraft/dev` (Compose + Traefik).
- Starts the dev stack via DEV_PROCESS_MGMT and Docker Compose.
- Provisions HTTPS certificates via DEV_MKCERT.

However, hosts file management is not yet implemented. Developers must still:

- Manually add dev domain entries (e.g., `app.localdev.test`, `api.localdev.test`) to `/etc/hosts`.
- Manually remove these entries when done developing.
- Remember which entries belong to Stagecraft versus other tools.

This creates friction and makes the dev experience less seamless. It also weakens the portfolio story for Stagecraft as an orchestrator that provides a complete, production-like dev environment from a single command.

DEV_HOSTS fills this gap by providing a deterministic, spec-governed way to manage `/etc/hosts` entries for dev domains used by CLI_DEV.

⸻

## 2. Motivation

### Developer Experience

- **One command**: Developers should be able to run `stagecraft dev` and have dev domains automatically resolve to `127.0.0.1` without manual hosts file editing.

- **Automatic cleanup**: When `stagecraft dev` exits, hosts file entries should be automatically removed, preventing accumulation of stale entries.

- **Opt-out flexibility**: Developers should be able to use `--no-hosts` to skip hosts file modification when they prefer manual DNS management or other DNS solutions.

### Operational Reliability

- **Deterministic entries**: Hosts file entries should be added in a stable, lexicographically sorted format.

- **Idempotent operations**: Running `stagecraft dev` multiple times must not create duplicate entries.

- **Safe cleanup**: Only Stagecraft-managed entries should be removed; other entries must be preserved.

- **Cross-platform support**: DEV_HOSTS must work on Linux, macOS, and Windows (with appropriate paths and permissions handling).

### CI and Automation

- **Skippable in CI**: CI environments often use DNS or other resolution mechanisms. DEV_HOSTS must be disabled when `--no-hosts` is set so pipelines can run without hosts file modification.

- **Deterministic behavior**: When enabled, DEV_HOSTS must always produce the same hosts file structure for given inputs.

⸻

## 3. Users and User Stories

### Developers

- As a developer, I want `stagecraft dev` to automatically add dev domain entries to my hosts file so I can access `app.localdev.test` and `api.localdev.test` without manual configuration.

- As a developer, I want hosts file entries to be automatically removed when `stagecraft dev` exits so I don't accumulate stale entries.

- As a developer, I want to disable hosts file modification with `--no-hosts` when I prefer to manage DNS manually or use other DNS solutions.

### Platform Engineers

- As a platform engineer, I want hosts file entries to be clearly marked as Stagecraft-managed so I can identify and audit them.

- As a platform engineer, I want DEV_HOSTS to preserve existing hosts file entries and only modify Stagecraft-managed entries.

- As a platform engineer, I want clear error messages when hosts file modification fails (e.g., permission errors).

### CI / Automation

- As a CI pipeline, I want to run `stagecraft dev --no-hosts` without requiring hosts file modification, so tests can run in environments where hosts file editing is restricted.

- As a CI pipeline, I want clear and deterministic failures when `--no-hosts` is not set but hosts file modification is not possible.

⸻

## 4. Success Criteria (v1)

1. When hosts file management is enabled (`--no-hosts` not set), DEV_HOSTS:

   - Adds entries for dev domains (e.g., `app.localdev.test`, `api.localdev.test`) pointing to `127.0.0.1`.
   - Marks entries as Stagecraft-managed with a comment.
   - Preserves all existing hosts file entries.
   - Writes entries in a deterministic, lexicographically sorted format.

2. DEV_HOSTS is idempotent:

   - If entries already exist for the requested domains, it does not create duplicates.
   - Running `stagecraft dev` multiple times produces the same hosts file state.

3. Cleanup on exit:

   - When `stagecraft dev` exits (normal or interrupted), DEV_HOSTS removes only Stagecraft-managed entries.
   - Other hosts file entries are preserved unchanged.

4. Cross-platform support:

   - Linux: `/etc/hosts`
   - macOS: `/etc/hosts`
   - Windows: `C:\Windows\System32\drivers\etc\hosts`
   - Handles permission errors gracefully with clear messages.

5. Error handling:

   - Permission errors return clear messages about sudo/elevation requirements.
   - Invalid hosts file format errors preserve what's possible and report clearly.
   - File locking errors are handled with retry logic or clear error messages.

6. Integration with CLI_DEV:

   - CLI_DEV calls DEV_HOSTS after domain computation.
   - CLI_DEV calls DEV_HOSTS cleanup on exit/interrupt.
   - The `--no-hosts` flag correctly bypasses DEV_HOSTS.

⸻

## 5. Risks and Constraints

### Determinism Constraints

- **Hosts file format must be stable**: Entries must be written in a deterministic format with lexicographically sorted domains.

- **No random data**: Entries must not include timestamps, UUIDs, or machine-specific identifiers.

- **Idempotent operations**: Adding entries multiple times must produce identical results.

### Cross-Platform Constraints

- **Different paths**: Linux/macOS use `/etc/hosts`, Windows uses `C:\Windows\System32\drivers\etc\hosts`.

- **Permission requirements**: Modifying `/etc/hosts` typically requires sudo/elevation on Unix systems.

- **File locking**: Hosts files may be locked by other processes; DEV_HOSTS must handle this gracefully.

### Safety Constraints

- **Preserve existing entries**: DEV_HOSTS must never remove or modify entries that are not Stagecraft-managed.

- **Clear marking**: Stagecraft-managed entries must be clearly marked so they can be identified and removed safely.

- **Atomic operations**: Hosts file writes should be atomic where possible to avoid corruption.

### Integration Constraints

- **CLI_DEV dependency**: DEV_HOSTS depends on CLI_DEV for domain computation and lifecycle management.

- **Cleanup timing**: DEV_HOSTS cleanup must be called during CLI_DEV shutdown, similar to DEV_PROCESS_MGMT teardown.

⸻

## 6. Alternatives Considered

### Alternative 1: Use DNS server (dnsmasq, etcd, etc.)

**Rejected because**: Adds external dependencies and complexity. Hosts file modification is simpler and more portable for local development.

### Alternative 2: Require manual hosts file editing

**Rejected because**: Creates friction and reduces the "one command" value proposition of CLI_DEV.

### Alternative 3: Use systemd-resolved or other system DNS overrides

**Rejected because**: Platform-specific and requires system-level configuration. Hosts file modification is universal and simpler.

⸻

## 7. Dependencies

### Required Features (all done)

- **CLI_DEV**: Provides domain computation and lifecycle management.
- **DEV_DOMAINS**: Domain computation logic (part of CLI_DEV).
- **DEV_PROCESS_MGMT**: Process lifecycle management (cleanup patterns).

### Spec Dependencies

- `spec/commands/dev.md` (CLI_DEV spec) - already exists
- `spec/dev/hosts.md` (DEV_HOSTS spec) - to be created

⸻

## 8. Approval

- Author: [To be filled]
- Reviewer: [To be filled]
- Date: [To be filled]

Once approved, the Implementation Outline may begin.

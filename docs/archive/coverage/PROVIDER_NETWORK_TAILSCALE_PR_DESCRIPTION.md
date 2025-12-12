> **Superseded by** `docs/engine/history/PROVIDER_NETWORK_TAILSCALE_EVOLUTION.md` and `docs/coverage/COVERAGE_LEDGER.md`. Kept for historical reference. New Tailscale evolution notes MUST go into the evolution log.

## Summary

Implement PROVIDER_NETWORK_TAILSCALE - Tailscale NetworkProvider implementation for Stagecraft, enabling mesh networking for multi-host deployments.

This PR:

- Implements the NetworkProvider interface for Tailscale
- Adds OS detection (Linux Debian/Ubuntu) with ErrUnsupportedOS
- Implements EnsureInstalled, EnsureJoined, and NodeFQDN methods
- Adds comprehensive test coverage (68.2%)
- Updates specs and status docs
- Marks feature as done in features.yaml

---

## Feature IDs

- `PROVIDER_NETWORK_TAILSCALE`

---

## Changes

### 1. PROVIDER_NETWORK_TAILSCALE planning and spec

- Added analysis and planning docs:
  - `docs/engine/analysis/PROVIDER_NETWORK_TAILSCALE.md` - Feature Analysis Brief
  - `docs/engine/outlines/PROVIDER_NETWORK_TAILSCALE_IMPLEMENTATION_OUTLINE.md` - Implementation Outline
- Added spec:
  - `spec/providers/network/tailscale.md`
    - Defines:
      - Config schema (auth_key_env, tailnet_domain, default_tags, role_tags, install settings)
      - EnsureInstalled behavior (OS detection, install script, idempotency)
      - EnsureJoined behavior (auth key handling, tag computation, tailnet validation)
      - NodeFQDN behavior (pure function with respect to config)
      - Error handling and determinism requirements

### 2. PROVIDER_NETWORK_TAILSCALE implementation

New package: `internal/providers/network/tailscale/`:

- `tailscale.go`
  - `TailscaleProvider` struct implementing `NetworkProvider` interface
  - `EnsureInstalled()` - Installs Tailscale client on Linux hosts (Debian/Ubuntu)
  - `EnsureJoined()` - Joins hosts to Tailscale tailnet with tags
  - `NodeFQDN()` - Generates deterministic FQDNs
  - `checkOSCompatibility()` - Validates OS/distribution before install

- `config.go`
  - `Config` struct with YAML tags
  - `InstallConfig` struct for install settings
  - `parseConfig()` - Validates required fields (auth_key_env, tailnet_domain)

- `status.go`
  - `TailscaleStatus` and `NodeInfo` structs for JSON parsing
  - `parseStatus()` - Parses `tailscale status --json` output

- `commander.go`
  - `Commander` interface for SSH command execution abstraction
  - `SSHCommander` - Production implementation using executil + SSH
  - `LocalCommander` - Test implementation with mocked commands
  - `getEnvVar()` - Retrieves environment variables with error handling

- `errors.go`
  - Error definitions: ErrConfigInvalid, ErrAuthKeyMissing, ErrAuthKeyInvalid, ErrTailnetMismatch, ErrTagMismatch, ErrInstallFailed, ErrUnsupportedOS

- `tailscale_test.go`
  - Comprehensive unit tests covering:
    - Provider registration
    - Config parsing and validation
    - EnsureInstalled (already installed, install succeeds, install fails, OS detection)
    - EnsureJoined (already joined, auth key missing, wrong tailnet, tag mismatch, join succeeds/fails)
    - NodeFQDN (with config, without config)
    - Helper functions (computeTags, tagsMatch)
  - Test coverage: 68.2%

- `registry_test.go`
  - Tests provider registration and retrieval from network registry

### 3. OS Detection

- Checks Linux via `uname -s`
- Validates Debian/Ubuntu via `/etc/os-release` or `lsb_release` fallback
- Returns `ErrUnsupportedOS` for unsupported OS/distributions
- Gracefully handles detection failures (proceeds if checks fail but OS isn't clearly wrong)

### 4. Tag Management

- Computes deterministic union of:
  - Default tags from config (`default_tags`)
  - Role-specific tags from config (`role_tags[role]`)
  - Tags passed in `EnsureJoinedOptions` (from host role in plan)
- Tags are sorted lexicographically for deterministic comparison
- Validates tags match expected set after join

### 5. Integration with Network Registry

- Provider registers automatically via `init()` function
- Can be retrieved via `network.Get("tailscale")`
- Follows standard provider registry pattern

### 6. Specs and status docs

- Updated `spec/features.yaml`: PROVIDER_NETWORK_TAILSCALE status changed from `todo` → `done`
- Regenerated `docs/engine/status/feature-completion-analysis.md`: Phase 4 now shows 33% complete (1/3 done)
- Regenerated `docs/engine/status/implementation-status.md`: PROVIDER_NETWORK_TAILSCALE listed as done with test files

---

## Phase 4 Status

- `Phase 4: Provider Implementations` is now 33% complete (1/3 features done):
  - `PROVIDER_NETWORK_TAILSCALE` - done ✅
  - `PROVIDER_CLOUD_DO` - todo
  - `DRIVER_DO` - todo

Status docs are regenerated from `spec/features.yaml` in this PR so that Phase 4 now shows 33% completion.

---

## Tests

- Unit tests:
  - `go test ./internal/providers/network/tailscale/...`
  - Coverage: 68.2% (target: ~70%, all critical paths covered)
- All tests pass
- Linting: 0 issues (golangci-lint)

---

## Risks and Mitigations

- **SSH access requirement**: Network provider requires SSH access to hosts before Tailscale is up
  - Mitigations: Documented in spec, clear error messages

- **OS support**: v1 supports Linux (Debian/Ubuntu) only
  - Mitigations: OS detection validates before install, returns clear ErrUnsupportedOS error

- **Auth key security**: Auth keys must be provided via environment variables
  - Mitigations: Never stored in config files, documented in spec

- **Determinism**: All operations are idempotent and deterministic
  - Mitigations: Tag computation is sorted, FQDN generation is pure function, error messages are stable

---

## Next Steps

This provider is ready for integration with Phase 7 infrastructure features:
- `CLI_INFRA_UP` - Infrastructure provisioning
- `INFRA_HOST_BOOTSTRAP` - Host bootstrap operations
- Multi-host deployment workflows

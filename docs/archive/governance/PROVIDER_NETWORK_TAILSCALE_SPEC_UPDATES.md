# PROVIDER_NETWORK_TAILSCALE Spec Updates - Proposed Changes

**Feature**: PROVIDER_NETWORK_TAILSCALE  
**Spec**: `spec/providers/network/tailscale.md`  
**Purpose**: Codify decisions from requirements analysis before Slice 2

This document contains proposed spec updates to resolve open decisions identified in `docs/engine/analysis/PROVIDER_NETWORK_TAILSCALE_REQUIREMENTS.md`.

---

## Decision 1: Tailnet Name vs MagicDNS Domain Matching

**Location**: Section 2.3 EnsureJoined, step 5

**Current spec text** (line ~154):
```
5. If already joined correctly:
   - Check tailnet matches config (TailnetName matches tailnet_domain or tailnet name)
```

**Proposed update**:
```
5. If already joined correctly:
   - Check tailnet matches config:
     - Compare `TailnetName` from `tailscale status --json` against the expected tailnet name
     - The `tailnet_domain` config field is used only for FQDN generation via `NodeFQDN`
     - For status comparison, the provider compares the `TailnetName` field from status JSON
     - **Expected tailnet name derivation**: The expected tailnet name SHALL be the portion of `tailnet_domain` before the first dot
       - Example: `tailnet_domain = "bartekus.ts.net"` → expected `TailnetName = "bartekus"`
       - Example: `tailnet_domain = "example.ts.net"` → expected `TailnetName = "example"`
     - If `TailnetName` does not match the expected tailnet name (derived from config), return `ErrTailnetMismatch`
```

**Rationale**: Makes explicit that `tailnet_domain` is for FQDN generation, while status comparison uses `TailnetName` from the status JSON. Provides deterministic mapping rule to derive expected tailnet name from config.

---

## Decision 2: Tag Equality Semantics

**Location**: Section 2.3 EnsureJoined, step 5 and step 7

**Current spec text** (line ~155):
```
   - Check tags match expected tags (all required tags present)
```

**Proposed update**:
```
   - Check tags match expected tags:
     - All expected tags must be present in the actual tags (subset match)
     - Extra tags beyond the expected set are allowed (Tailscale may add system tags)
     - Tag comparison is case-sensitive and exact string match
     - Tags are sorted lexicographically for deterministic comparison
```

**Also update** step 7 (line ~162):
```
7. Validate final state:
   - Tailnet matches config (using TailnetName comparison as above)
   - Tags match expected tags (subset match: all expected tags present, extras allowed)
   - Node is online (or at least successfully configured - see offline node handling below)
```

**Rationale**: Clarifies that tag matching is a subset operation, not exact equality, which is more robust for real-world Tailscale deployments.

---

## Decision 3: Role Handling in EnsureJoinedOptions

**Location**: Section 2.3 EnsureJoined, step 3

**Current spec text** (line ~146-149):
```
3. Compute tags: union of:
   - Default tags from config (default_tags)
   - Role-specific tags from config (role_tags[role])
   - Tags passed in opts.Tags (from host role in plan)
```

**Proposed update**:
```
3. Compute expected tags:
   - The provider receives the final tag list in `opts.Tags`
   - Stagecraft core (the caller) is responsible for computing the union of:
     - Default tags from config (`default_tags`)
     - Role-specific tags from config (`role_tags[role]`) where role is determined by core
     - Any additional tags from the deployment plan
   - **Forbidden behavior**: The provider MUST NOT attempt to compute `role_tags` or inspect roles directly
   - Only the final tag list in `opts.Tags` is authoritative
   - Tags are sorted lexicographically for deterministic comparison
```

**Also add note to Section 2.3 Input** (around line ~126):
```
**Input:**

```go
type EnsureJoinedOptions struct {
    Config any     // Provider-specific config
    Host   string  // Hostname or logical host ID
    Tags   []string // Final computed tags to apply (includes default_tags, role_tags, and plan tags)
}
```

**Note**: The `Tags` field contains the final computed tag list. Stagecraft core computes this by combining `default_tags`, `role_tags[role]`, and any plan-specific tags before calling `EnsureJoined`. The provider does not access `role_tags` directly.
```

**Rationale**: Simplifies provider interface by having core handle tag computation. Provider focuses on applying the final tag list.

---

## Decision 4: Tag Validation (tag: prefix requirement)

**Location**: Section 3.2 Config Fields, default_tags and role_tags descriptions

**Current spec text** (line ~293):
```
- Tags must start with `tag:` prefix
```

**Proposed update**:
```
- Tags must start with `tag:` prefix
- Tags without the `tag:` prefix are considered invalid config and cause `ErrConfigInvalid`
- **Validation timing**: Tag validation occurs during provider config parsing inside `EnsureInstalled` and `EnsureJoined`
- Config with invalid tags MUST fail before any remote operations (SSH or Tailscale calls)
```

**Also add to Section 3.3 Config Validation** (after line ~324):
```
**Tag Validation:**
- All tags in `default_tags` and `role_tags` must start with `tag:` prefix
- Tags without prefix cause: `"tailscale provider: invalid config: tag {tag} must start with tag: prefix"`
- Validation is strict: invalid tags cause config parsing to fail
- Validation occurs during provider config parsing, before any remote operations
```

**Rationale**: Makes tag validation explicit and strict, matching Tailscale's tag format requirements.

---

## Decision 5: Minimum Version Enforcement

**Location**: Section 2.2 EnsureInstalled, step 4

**Current spec text** (line ~87):
```
   - If command succeeds and version >= min_version (if configured), return nil (already installed)
```

**Proposed update**:
```
   - If command succeeds:
     - Parse installed version string as semantic version
     - **Version parsing rules**:
       - Strip build metadata (e.g., `1.44.0-123-gabcd` → `1.44.0`)
       - Accept patch suffixes (e.g., `1.78.0-1` → `1.78.0`)
       - If version cannot be parsed as semantic version, return `ErrInstallFailed` with message: `"tailscale provider: installation failed: cannot parse installed version {version}"`
     - If `install.min_version` is configured and installed version < min_version:
       - Return error: `"tailscale provider: installation failed: installed version {actual} is below minimum {min_version}"`
       - Do not attempt automatic upgrade
     - Otherwise, return nil (already installed at acceptable version)
```

**Also add to Section 2.2 Error Cases** (after line ~109):
```
- Installed version below minimum version requirement (if `install.min_version` is configured)
```

**Also add to Section 6.2 Error Messages** (after line ~412):
```
- Version too old: `"tailscale provider: installation failed: installed version {actual} is below minimum {min_version}"`
```

**Rationale**: Makes version enforcement explicit and predictable. No automatic upgrades in v1.

---

## Decision 6: Offline but Configured Nodes

**Location**: Section 2.3 EnsureJoined, Guarantees and step 7

**Current spec text** (line ~139):
```
- Host's Tailscale node is online or at least successfully configured
```

**Proposed update**:
```
- Host's Tailscale node is online or at least successfully configured
- **Offline node handling**: If a node is offline but has the correct tailnet and tags configured, `EnsureJoined` considers this a success
- The provider checks `Self.Online` from status JSON but does not fail if the node is offline, as long as configuration is correct
- This allows deployment workflows to proceed even if nodes are temporarily offline
```

**Also update step 7** (line ~163):
```
7. Validate final state:
   - Tailnet matches config (using TailnetName comparison)
   - Tags match expected tags (subset match: all expected tags present, extras allowed)
   - Node configuration is correct (online status is checked but offline nodes with correct config are acceptable)
```

**Rationale**: Clarifies that offline nodes with correct configuration are acceptable, which is more practical for deployment scenarios.

---

## Decision 7: macOS Support Status

**Location**: Section 2.2 EnsureInstalled, Supported OS (v1)

**Current spec text** (line ~98-102):
```
**Supported OS (v1):**

- Linux (Debian/Ubuntu) only
- Uses Tailscale's official install script
- Other OS support is deferred to future versions
```

**Proposed update**:
```
**Supported OS (v1):**

- **Target hosts**: Linux (Debian/Ubuntu) only
- Uses Tailscale's official install script
- **Unsupported target OS**: macOS, Windows, Alpine, CentOS, and all other Linux distributions
  - For unsupported target OS, `EnsureInstalled` MUST return `ErrUnsupportedOS`
- **Orchestrator OS**: Completely irrelevant - the provider operates on remote hosts via SSH, not the orchestrator machine
  - The orchestrator (local machine running Stagecraft) MUST NOT be inspected for OS compatibility
  - Only remote hosts are checked for OS compatibility
- Other OS support is deferred to future versions

**Note**: The orchestrator (the machine running Stagecraft) does not need Tailscale installed. The provider manages Tailscale on remote Linux hosts via SSH. You may install Tailscale on your Mac for your own network access, but this is not a requirement for provider operation.
```

**Also add to Section 11. Non-Goals (v1)** (after line ~622):
```
- Supporting macOS or Windows as target hosts (Linux Debian/Ubuntu only)
- Requiring Tailscale to be installed on the orchestrator machine
```

**Rationale**: Makes explicit that macOS is unsupported as a target host, and clarifies that orchestrator OS is irrelevant.

---

## Summary of Changes

1. **Tailnet matching**: Explicitly document that `tailnet_domain` is for FQDN generation, while status comparison uses `TailnetName`.
2. **Tag matching**: Document subset match semantics (expected tags must be present, extras allowed).
3. **Role handling**: Clarify that core computes tags; provider receives final list.
4. **Tag validation**: Make tag prefix requirement strict and explicit.
5. **Version enforcement**: Document that version < min_version causes failure, no auto-upgrade.
6. **Offline nodes**: Clarify that offline nodes with correct config are acceptable.
7. **macOS support**: Explicitly document that macOS is unsupported as target host.

---

## Implementation Notes

After updating the spec with these changes:

1. Update implementation to match spec:
   - Tag matching logic (subset vs exact)
   - Version enforcement behavior
   - Offline node handling
2. Update tests to reflect new behavior
3. Proceed with Slice 2 coverage work

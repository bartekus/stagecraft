# PROVIDER_NETWORK_TAILSCALE v1 – Requirements & Open Decisions

**Feature**: PROVIDER_NETWORK_TAILSCALE  
**Spec**: `spec/providers/network/tailscale.md`  
**Domain**: providers

This document enumerates what v1 of the Tailscale NetworkProvider must do, what is explicitly out of scope, and what decisions / unknowns must be resolved before further implementation and coverage work (Slice 2+).

---

## 1. Behavioral Contract (What the provider must do)

The provider implements the NetworkProvider interface:

```go
type NetworkProvider interface {
    ID() string
    EnsureInstalled(ctx context.Context, opts EnsureInstalledOptions) error
    EnsureJoined(ctx context.Context, opts EnsureJoinedOptions) error
    NodeFQDN(host string) (string, error)
}
```

### 1.1 ID

- Must return the stable identifier: `"tailscale"`.
- Must match `network.provider: tailscale` in config.
- Used as the key in the network provider registry.

### 1.2 EnsureInstalled

**Goal**: On a target host (Linux Debian/Ubuntu), ensure:
- `tailscaled` is installed and enabled.
- `tailscale` CLI is available on PATH.
- Installed version meets (optional) minimum version requirement.

**Required behavior:**

1. Parse provider config from `opts.Config` into a typed config struct.
2. Validate config (see section 2):
   - `auth_key_env` required.
   - `tailnet_domain` required and valid.
3. Respect install method:
   - `install.method == "skip"`: do no installation attempts; return nil (but may log a debug note).
   - Default (or `"auto"`): perform install if needed.
4. Detect existing installation via Commander:
   - Run something like `tailscale version` or `which tailscale`.
   - If present and (if configured) version >= `install.min_version` → return nil.
5. If not installed:
   - Execute official install script:
   - Conceptually: `curl -fsSL https://tailscale.com/install.sh | sh` (exact command should be centralized and testable).
   - If install fails: wrap as `ErrInstallFailed` with context.
6. Re-verify by calling `tailscale version` again.
7. Enforce idempotency:
   - Multiple calls with same config and host must converge on same final state.

**Non-goals for v1:**
- No support for non-Linux hosts.
- No OS-specific install flows beyond Debian/Ubuntu via official script.
- No dynamic downgrade / upgrade logic beyond minimal version check.

### 1.3 EnsureJoined

**Goal**: Ensure the host is joined to the correct tailnet with the correct tags.

**Required behavior:**

1. Parse config from `opts.Config`.
2. Read auth key from environment variable named by `config.AuthKeyEnv`.
   - If missing → `ErrAuthKeyMissing`.
3. Compute expected tags:
   - `default_tags` from config.
   - Role-specific tags from config (if role is present in opts – see "open decisions").
   - Explicit `opts.Tags`.
   - Deterministic union; sorted lexicographically.
4. Discover current state:
   - Run `tailscale status --json` via Commander.
   - Parse with `parseStatus`.
   - Extract:
     - Tailnet name / MagicDNS domain.
     - Node tags.
     - Node online / offline state.
5. Determine if the node is already correct:
   - Tailnet matches expected (see mapping decision in section 9).
   - Tags contain at least the expected set (exact vs superset – decision below).
   - If all invariants satisfied → return nil, do not run `tailscale up`.
6. If not correct:
   - Build `tailscale up` command via `buildTailscaleUpCommand`.
   - Run with auth key and tags.
   - Re-run `tailscale status --json`.
   - Re-validate final state.
7. Enforce idempotency:
   - Repeated `EnsureJoined` calls with the same config/host lead to the same configuration.
   - "Already correct" must not cause errors.

### 1.4 NodeFQDN

**Goal**: Compute the Tailscale FQDN for a host given a tailnet MagicDNS domain.

**Required behavior:**

1. Use loaded provider config (carried from `EnsureInstalled`/`EnsureJoined`).
2. Validate `tailnet_domain` via `validateTailnetDomain`.
3. Produce deterministic FQDN via `buildNodeFQDN(host, tailnetDomain)`:
   - `fmt.Sprintf("%s.%s", host, tailnetDomain)`.
4. No network calls, no side effects.

**Error behavior:**
- If tailnet domain invalid / missing → return error wrapping `ErrConfigInvalid`.

---

## 2. Configuration Schema & Validation

**Source**: `spec/providers/network/tailscale.md` (section 3).

### 2.1 Required fields

- **`auth_key_env`** (string, non-empty)
  - Name of env var holding Tailscale auth key.
- **`tailnet_domain`** (string, non-empty)
  - Must contain at least one dot (enforced via `validateTailnetDomain`).
  - MagicDNS domain, e.g. `mytailnet.ts.net`.

### 2.2 Optional fields

- **`default_tags`** (list of strings)
  - Each should start with `tag:`.
- **`role_tags`** (map string → []string)
  - Key: logical role, e.g. `app`, `db`, `gateway`.
  - Value: tags (also `tag:...`).
- **`install.method`** (`"auto"` or `"skip"`)
  - Default: `"auto"` if unspecified.
- **`install.min_version`** (string, semantic version, optional)
  - If set, `EnsureInstalled` checks that tailscale version meets threshold.

### 2.3 Validation rules

- Missing required fields → `ErrConfigInvalid`.
- Invalid `tailnet_domain` (empty or no dot) → `ErrConfigInvalid`.
- Empty `auth_key_env` → `ErrConfigInvalid`.
- Tags not starting with `tag:` – either:
  - Strict: considered invalid config (decision), or
  - Lenient: allowed but not recommended (spec currently leans "must start with `tag:`").

---

## 3. Provider Boundaries (Core vs Provider)

### 3.1 Stagecraft core responsibilities

- Resolve and load `network.provider` config section.
- Handle SSH / connection details via Commander abstraction.
- Maintain registry of providers (`network.Register` / `Get`).
- Decide when to call `EnsureInstalled`, `EnsureJoined`, `NodeFQDN` in workflows:
  - Phase 7: `CLI_INFRA_UP`, `INFRA_HOST_BOOTSTRAP`, etc.

### 3.2 Tailscale provider responsibilities

- Interpret only its own config (`network.providers.tailscale`).
- Implement Tailscale-specific logic:
  - OS compatibility check (Debian/Ubuntu v1).
  - Install via official script.
  - Tailnet join via `tailscale up`.
  - Status discovery via `tailscale status --json`.
  - Read `auth_key_env` from orchestrator environment.
- **Never:**
  - Touch non-Tailscale provider config.
  - Implement cloud provisioning (droplets, firewalls, etc.).
  - Modify Stagecraft core behavior.

---

## 4. Installation Semantics (Who installs Tailscale? When?)

### 4.1 Target hosts in v1

- **Supported**: Linux Debian / Ubuntu (via `/etc/os-release` → `parseOSRelease`).
- **Unsupported:**
  - macOS.
  - Windows.
  - Other Linux distros (Alpine, CentOS, etc.) – treated as unsupported OS.

### 4.2 Installer behavior

- For Debian/Ubuntu: run Tailscale official `install.sh` via Commander.
- For unsupported OS:
  - Return `ErrUnsupportedOS` with a clear message, no install attempt.

### 4.3 Your original question: install on your Mac vs CLI installing

- **For provider development and v1 behavior:**
  - The provider is concerned with installing/joining **remote Linux hosts** via SSH, not your Mac as the orchestrator.
  - The orchestrator does **not** need Tailscale installed for v1 provider behavior to work.
- **For convenience (outside provider scope):**
  - You can absolutely install Tailscale on your Mac for your own network access, but:
    - It should not be a hidden dependency of tests or provider logic.
    - Provider tests should assume "remote Linux host + Commander abstraction," not "Stagecraft host has tailscale."

**Requirement:**
All behavior that matters for Stagecraft must be testable with no Tailscale installed on the developer's machine; only via mocked Commander and, optionally, integration tests against Linux VMs/containers.

---

## 5. State Discovery & Reconciliation

### 5.1 Status model

- **Source of truth**: `tailscale status --json`.
- `parseStatus` must:
  - Parse JSON structurally.
  - Expose at least:
    - `TailnetName`
    - `Self.Tags`
    - `Self.Online`
    - (Optionally more fields as needed later).

### 5.2 Conceptual state machine

Given a host:

1. **Not installed** – `tailscale version` fails.
2. **Installed but not logged in** – status has no tailnet or indicates logged out.
3. **Logged into wrong tailnet** – tailnet does not match expected.
4. **Logged into correct tailnet but wrong tags**.
5. **Correct** – tailnet and tags match invariants.

`EnsureInstalled` handles 1 → 2.  
`EnsureJoined` handles 2/3/4 → 5.

### 5.3 Reconciliation rules

- **Wrong tailnet:**
  - Must fail fast with `ErrTailnetMismatch` (no auto-migration in v1).
- **Wrong tags:**
  - Run `tailscale up` with desired tags, then re-validate.
- **Offline node:**
  - Spec says "online or at least successfully configured."
  - **Decision**: treat "offline but correctly configured" as success for `EnsureJoined` in v1, with clear logging.

---

## 6. Error Handling Requirements

Sentinel errors defined in spec:

```go
var (
    ErrConfigInvalid     = errors.New("invalid config")
    ErrAuthKeyMissing    = errors.New("auth key missing from environment")
    ErrAuthKeyInvalid    = errors.New("invalid or expired auth key")
    ErrTailnetMismatch   = errors.New("tailnet mismatch")
    ErrTagMismatch       = errors.New("tag mismatch")
    ErrInstallFailed     = errors.New("tailscale installation failed")
    ErrUnsupportedOS     = errors.New("unsupported operating system")
)
```

### 6.1 Mapping conditions → errors

- Config parse/validation failures → `ErrConfigInvalid`.
- Missing `auth_key_env` variable → `ErrAuthKeyMissing`.
- Auth key causes login failure → `ErrAuthKeyInvalid`.
- Tailnet mismatch after status → `ErrTailnetMismatch`.
- Tag mismatch after join attempt → `ErrTagMismatch`.
- Install script failure → `ErrInstallFailed`.
- Non-Debian/Ubuntu OS → `ErrUnsupportedOS`.

Errors must be wrapped with `fmt.Errorf("tailscale provider: <context>: %w", err)` and be deterministic (no random strings, no timestamps).

---

## 7. Testing Strategy

### 7.1 Unit tests (package-local, deterministic)

Already started with Slice 1:
- `buildTailscaleUpCommand`
- `parseOSRelease`
- `validateTailnetDomain`
- `buildNodeFQDN`
- `parseStatus` edge cases

**Next slices should cover:**
- Config parsing and validation.
- `checkOSCompatibility` paths (supported, unsupported, empty).
- `EnsureInstalled` control flow using fake Commander:
  - Installed / not-installed / install failure / unsupported OS.
- `EnsureJoined` control flow using fake Commander:
  - Already correct.
  - Needs join.
  - Wrong tailnet.
  - Tag mismatch.
  - Missing auth key.

### 7.2 Integration tests (optional, build-tagged)

- Behind `//go:build integration` and env guard like `STAGECRAFT_TAILSCALE_INTEGRATION=1`.
- Deployed against throwaway Linux VM/container.
- Validate a full "install + join + NodeFQDN" roundtrip.

### 7.3 Determinism

All tests must pass:
- `go test -cover ./internal/providers/network/tailscale`
- `go test -race ./internal/providers/network/tailscale`
- `go test -count=20 ./internal/providers/network/tailscale`
- No reliance on local Tailscale install.

---

## 8. Integration Points With Stagecraft Flows

**Key consumers (current and future):**

- **Phase 7 / CLI_INFRA_UP:**
  - Cloud provider ensures droplet/VM.
  - Network provider `EnsureInstalled` and `EnsureJoined`.
  - Returns host metadata including FQDN (`NodeFQDN`).
- **INFRA_HOST_BOOTSTRAP:**
  - Uses Tailscale provider to prepare networking on newly provisioned hosts.
- **Topology / plan execution:**
  - FQDNs used by compose / Traefik / other providers (e.g. backend frontend providers) as hostnames.

**Requirement**: All usages must treat Tailscale provider as a pure NetworkProvider with no extra side channels.

---

## 9. Open Decisions / Unknowns To Resolve Before Slice 2

These are the things we should explicitly decide (and reflect in the spec) before deeper implementation:

### 9.1 Tailnet name vs MagicDNS domain matching

**Issue**: Tailscale's `TailnetName` can differ from MagicDNS domain (`<org>.github` vs `<org>.ts.net`).

**Decision needed**: What exactly do we compare `tailnet_domain` to in status?

- **Option A**: Treat `tailnet_domain` as canonical and only use it in `NodeFQDN`, not for status comparison.
- **Option B**: Support mapping or config field for tailnet name vs MagicDNS domain.

**Recommendation**: **Option A** for v1. `tailnet_domain` is used for FQDN generation. For status comparison, we compare against `TailnetName` from status JSON. If they don't match, we fail with `ErrTailnetMismatch`. This is explicit and clear.

### 9.2 Tag equality semantics

**Issue**: Do we require exact match, or "expected is subset of actual"?

**Real-world**: Tailscale may add system tags; exact equality might be brittle.

**Decision needed**: Tag matching strategy.

- **Option A**: Exact match (all expected tags must be present, no extras allowed).
- **Option B**: Subset match (all expected tags must be present, extras allowed).

**Recommendation**: **Option B** for v1. Required tags must be a subset of actual tags; extra tags allowed. This is more robust and matches real-world Tailscale behavior where system tags may be added.

### 9.3 Role injection into EnsureJoinedOptions

**Issue**: Spec mentions `role_tags[role]`, but `EnsureJoinedOptions` currently only has `Host` and `Tags`.

**Decision needed**: Where does role come from?

- **Option A**: Extend `EnsureJoinedOptions` with `Role string`.
- **Option B**: Have caller pre-compute tags (provider only sees final tag list).

**Recommendation**: **Option B** for v1. The caller (Stagecraft core) should compute the full tag list including role-specific tags before calling `EnsureJoined`. This keeps the provider interface simpler and more focused. The provider receives the final tag list and applies it.

### 9.4 Handling of invalid tags (without `tag:` prefix)

**Issue**: What happens if tags don't start with `tag:`?

**Decision needed**: Strict vs lenient behavior.

- **Option A**: Strict – reject tags without `tag:` prefix as invalid config.
- **Option B**: Lenient – allow but warn/log.

**Recommendation**: **Option A** for v1. Tags must start with `tag:` prefix. This is explicit in the spec and matches Tailscale's tag format requirements. Invalid tags should cause `ErrConfigInvalid`.

### 9.5 Minimum version enforcement

**Issue**: Do we fail if version < `install.min_version`, or attempt an upgrade?

**Decision needed**: Version check behavior.

- **Option A**: Fail with clear message (`ErrInstallFailed` or new error) and do not auto-upgrade.
- **Option B**: Attempt upgrade automatically.

**Recommendation**: **Option A** for v1. If `install.min_version` is set and the installed version is below it, fail with a clear error message. Do not attempt automatic upgrade. This is explicit and predictable.

### 9.6 Offline but configured nodes

**Issue**: What if a node is offline but correctly configured?

**Decision needed**: Success criteria for `EnsureJoined`.

- **Option A**: Offline nodes with correct config are considered success.
- **Option B**: Nodes must be online to be considered successfully joined.

**Recommendation**: **Option A** for v1. Offline nodes with correct tailnet and tags are considered success. The spec already says "online or at least successfully configured." This is more practical for deployment scenarios where nodes may be temporarily offline.

### 9.7 macOS as target OS

**Issue**: Is macOS officially "unsupported in v1"?

**Decision needed**: macOS support status.

**Recommendation**: **macOS is officially unsupported in v1**, even if you personally run Stagecraft on macOS as orchestrator. The provider targets remote Linux hosts (Debian/Ubuntu) only. The orchestrator OS is irrelevant to provider behavior.

---

## 10. Summary: Decisions Made

Based on the analysis above, here are the concrete decisions for v1:

1. **Tailnet matching**: Compare `TailnetName` from status JSON against expected tailnet. `tailnet_domain` is used only for FQDN generation.
2. **Tag matching**: Subset match – all expected tags must be present, extras allowed.
3. **Role handling**: Caller pre-computes tags; provider receives final tag list.
4. **Tag validation**: Strict – tags must start with `tag:` prefix.
5. **Version enforcement**: Fail if version < `install.min_version`, no auto-upgrade.
6. **Offline nodes**: Offline but correctly configured nodes are considered success.
7. **macOS support**: Officially unsupported in v1 (Linux Debian/Ubuntu only).

These decisions should be codified in the spec before Slice 2 implementation.

---

## 11. Next Steps

1. **Update spec** with decisions from section 10.
2. **Proceed with Slice 2** focusing on:
   - `EnsureInstalled` / `EnsureJoined` control-flow tests with fake Commander.
   - Error-path coverage based on the sentinel errors.
   - Config validation tests.
3. **Continue coverage work** toward ≥80% target.

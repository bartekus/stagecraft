# PROVIDER_NETWORK_TAILSCALE Evolution Log

> Canonical evolution history for the Tailscale network provider.
> This document replaces per slice plans, readiness docs, checklists, and ad hoc notes.

## 1. Purpose and Scope

This document captures the end to end evolution of `PROVIDER_NETWORK_TAILSCALE`:

- Design intent and constraints
- Slice plans and execution notes
- Coverage movement over time
- Governance and spec changes
- Open questions and deferred work

It consolidates content that previously lived in:

- `docs/governance/PROVIDER_NETWORK_TAILSCALE_SLICE1_PLAN.md`
- `docs/governance/PROVIDER_NETWORK_TAILSCALE_SLICE1_CHECKLIST.md`
- `docs/governance/PROVIDER_NETWORK_TAILSCALE_SLICE1_READY.md`
- `docs/engine/agents/PROVIDER_NETWORK_TAILSCALE_SLICE1_AGENT.md`
- `docs/governance/PROVIDER_NETWORK_TAILSCALE_SLICE2_PLAN.md`
- `docs/governance/PROVIDER_NETWORK_TAILSCALE_SLICE2_CHECKLIST.md`
- `docs/governance/PROVIDER_NETWORK_TAILSCALE_SLICE2_COMPLETENESS.md`
- `docs/governance/PROVIDER_NETWORK_TAILSCALE_SLICE2_DEPENDENCIES.md`
- `docs/governance/PROVIDER_NETWORK_TAILSCALE_SLICE2_COVERAGE_EXPECTATIONS.md`
- `docs/coverage/PROVIDER_NETWORK_TAILSCALE_PR_DESCRIPTION.md`
- Any other Tailscale specific coverage or governance notes

All future Tailscale evolution notes should be added here instead of creating new standalone docs.

---

## 2. Feature References

- **Feature ID:** `PROVIDER_NETWORK_TAILSCALE`
- **Spec:** `spec/providers/network/tailscale.md`
- **Core analysis:** `docs/engine/analysis/PROVIDER_NETWORK_TAILSCALE.md`
- **Implementation outline:** `docs/engine/outlines/PROVIDER_NETWORK_TAILSCALE_IMPLEMENTATION_OUTLINE.md`
- **Status:** see `docs/engine/status/PROVIDER_COVERAGE_STATUS.md` and `docs/coverage/COVERAGE_LEDGER.md`

---

## 3. Design Intent and Constraints

> Short summary of why this provider exists and the constraints it must respect.
> Migrate the high level intent from the analysis and governance docs here.

- **Purpose**: Enable Stagecraft to manage Tailscale mesh networking for deployment hosts, providing secure private networking and stable FQDNs for multi-host deployments.

- **Primary responsibilities**:
  - Ensure Tailscale is installed and configured on Linux hosts (Debian/Ubuntu)
  - Join hosts to the correct Tailscale tailnet with appropriate tags
  - Generate deterministic FQDNs for use in Compose generation and infrastructure operations
  - Validate OS compatibility before installation attempts
  - Handle version enforcement when minimum version is configured

- **Non goals**:
  - Managing Tailscale ACLs or tailnet configuration (handled by Tailscale admin console)
  - Supporting every OS (Linux Debian/Ubuntu only for v1)
  - Supporting macOS or Windows as target hosts (Linux Debian/Ubuntu only)
  - Requiring Tailscale to be installed on the orchestrator machine
  - Managing auth key creation or rotation (user responsibility)
  - Dynamic network reconfiguration (static configuration only)
  - Multiple network providers per project (single provider only)

- **Determinism constraints**:
  - All operations must be idempotent (running multiple times produces same result)
  - NodeFQDN must be a pure function (no network calls, deterministic output)
  - No time.Sleep in tests
  - No real SSH or Tailscale CLI calls in tests (use LocalCommander mock)
  - All tests must pass with `-race` and `-count=20` for determinism verification

- **Provider boundary rules**:
  - Core is always provider-agnostic
  - Provider implements NetworkProvider interface, never core logic
  - Config is opaque to core (provider-specific validation)
  - Tags are computed by Stagecraft core, provider receives final list in `opts.Tags`

- **External dependencies**:
  - Tailscale auth key from environment variable (never stored in config)
  - Tailscale admin console for ACL management (provider does not manage ACLs)
  - SSH access to target hosts (required before Tailscale is installed)
  - Tailscale install script (official script from tailscale.com)

---

## 4. Slice Timeline Overview

> High level table that shows each slice and its role.

| Slice | Status        | Focus                             | Coverage before | Coverage after | Date range        | Notes |
|-------|---------------|-----------------------------------|-----------------|----------------|-------------------|-------|
| 1     | complete      | Helper extraction + unit tests   | 68.2%           | 71.3%          | 2025-XX-XX        | Extracted 4 pure helper functions |
| 2     | complete      | Error paths + version enforcement | 71.3%           | 79.6%          | 2025-XX-XX        | Added 34 new test cases |
| 3     | planned / n/a | [reserved for future work]        | -               | -              | -                 | Final push to ≥80% if needed |

---

## 5. Slice 1 - Plan and Execution

### 5.1 Slice 1 Objectives

> Move the key objectives from SLICE1_PLAN here.

- Extract pure helper functions from orchestration code to enable deterministic unit testing
- Add comprehensive unit tests for extracted helpers without touching CLI invocation tests
- Improve coverage from 68.2% to ~75% through helper extraction and focused test additions

### 5.2 Scope

- **Included**:
  - Extract 4 pure helper functions: `buildTailscaleUpCommand`, `parseOSRelease`, `validateTailnetDomain`, `buildNodeFQDN`
  - Add unit tests for all 4 helpers (table-driven tests)
  - Add edge case tests for `parseStatus()` (invalid JSON, empty JSON, missing fields)
  - Refactor existing code to use extracted helpers (behavior unchanged)

- **Excluded**:
  - CLI invocation tests (deferred to Slice 2)
  - Error path tests for `EnsureInstalled()` (deferred to Slice 2)
  - Version parsing logic (deferred to Slice 2)

- **Deferred to future slices**:
  - Error path coverage for `EnsureInstalled()`
  - Version enforcement logic
  - OS compatibility comprehensive testing

### 5.3 Execution Notes

> Migrate checklists, "ready" conditions, and agent instructions here.

- **Preconditions**:
  - `ErrConfigInvalid` exists in `errors.go` ✅
  - Error message style confirmed: `"tailscale provider: %w: ..."`
  - Test patterns reviewed: table-driven with `t.Parallel()`
  - Package structure confirmed: helpers stay in `package tailscale`

- **Steps executed**:
  1. Extracted `buildTailscaleUpCommand()` from `EnsureJoined()` line 153
  2. Extracted `parseOSRelease()` from `checkOSCompatibility()` lines 283-296
  3. Extracted `validateTailnetDomain()` and `buildNodeFQDN()` from `NodeFQDN()`
  4. Refactored existing code to use helpers (behavior unchanged)
  5. Added unit tests: `TestBuildTailscaleUpCommand`, `TestParseOSRelease`, `TestValidateTailnetDomain`, `TestBuildNodeFQDN`
  6. Added edge case tests: `TestParseStatus_InvalidJSON`, `TestParseStatus_EmptyJSON`, `TestParseStatus_MissingFields`

- **Edge cases covered**:
  - Empty os-release content
  - Quoted ID values in os-release
  - Missing ID field in os-release
  - Invalid JSON in status parsing
  - Empty JSON in status parsing
  - Missing fields in status JSON

- **Failure modes handled**:
  - All helpers are pure functions (no external dependencies)
  - Tests use table-driven patterns for comprehensive coverage
  - All tests pass with `-race` and `-count=20`

### 5.4 Coverage and Outcomes

- **Starting coverage**: 68.2%
- **Ending coverage**: 71.3% (+3.1 percentage points)
- **New tests added**: 7 test functions (4 for helpers, 3 for parseStatus edge cases)
- **Known limitations**: CLI invocation tests and error paths still need coverage (addressed in Slice 2)

---

## 6. Slice 2 - Plan and Execution

### 6.1 Slice 2 Objectives

> Move the key objectives from SLICE2_PLAN here.

- Add comprehensive error path and behavioral coverage for `EnsureInstalled()` using mock Commander
- Focus on config validation, OS compatibility, version enforcement, and install flows
- Improve coverage from 71.3% to ~78-80% without touching real SSH or Tailscale
- Implement version parsing helper per spec requirements

### 6.2 Dependencies and Constraints

> Consolidate SLICE2_DEPENDENCIES and any external assumptions.

- **Tailscale client version assumptions**:
  - Versions may include build metadata (e.g., `1.44.0-123-gabcd`) which must be stripped
  - Versions may include patch suffixes (e.g., `1.78.0-1`) which must be handled
  - Version comparison uses simple string comparison for v1 (lexicographic)

- **OS support matrix**:
  - Supported: Linux Debian/Ubuntu only
  - Unsupported: macOS, Windows, Alpine, CentOS, and all other Linux distributions
  - OS detection uses `uname -s`, `/etc/os-release`, and `lsb_release -i -s` as fallback

- **Config schema requirements**:
  - `auth_key_env`: Required, non-empty string
  - `tailnet_domain`: Required, non-empty string, must contain a dot
  - `install.method`: Optional, defaults to "auto", can be "skip"
  - `install.min_version`: Optional, if set, enforces minimum version

- **External network assumptions**:
  - SSH access to target hosts required before Tailscale is installed
  - Tailscale install script fetched from tailscale.com (official script)
  - No network calls in tests (all Commander calls mocked via LocalCommander)

### 6.3 Execution Notes

> Combine SLICE2_CHECKLIST, SLICE2_COMPLETENESS, and slice 2 agent doc content.

- **Preconditions**:
  - `ErrInstallFailed` and `ErrUnsupportedOS` exist in `errors.go` ✅
  - Slice 1 helpers available: `parseOSRelease()` used in OS compatibility tests
  - `LocalCommander` supports all needed command patterns ✅
  - Spec requirements reviewed for version parsing rules ✅

- **Steps executed** (5 micro-slices):
  1. **Micro-slice 1**: Implemented `parseTailscaleVersion()` helper with 11 test cases (coverage: 71.3% → 73.0%)
  2. **Micro-slice 2**: Added config validation tests - 5 test cases (coverage: 73.0% → 73.0%, no change)
  3. **Micro-slice 3**: Added OS compatibility tests - 9 test cases (coverage: 73.0% → 75.4%)
  4. **Micro-slice 4**: Added version enforcement tests - 7 test cases (coverage: 75.4% → 77.7%)
  5. **Micro-slice 5**: Added install flow tests - 2 test cases (coverage: 77.7% → 79.6%)

- **Error handling flows**:
  - Config validation: Missing `auth_key_env`, missing `tailnet_domain`, invalid YAML, install method "skip"
  - OS compatibility: Debian/Ubuntu (supported), Alpine/CentOS/Darwin (unsupported), uname fails, os-release missing
  - Version enforcement: Version meets/exceeds/below minimum, version with build metadata, version with patch suffix, unparseable version
  - Install flows: Already installed, install succeeds, install fails, verification fails

- **Regression coverage added**:
  - All existing tests continue to pass
  - No behavior changes beyond version parsing implementation (per spec)
  - All tests pass with `-race` and `-count=5` for determinism

### 6.4 Coverage and Outcomes

> Consolidate SLICE2_COVERAGE_EXPECTATIONS and PR coverage descriptions.

- **Starting coverage**: 71.3%
- **Ending coverage (current baseline)**: 79.6% (+8.3 percentage points)
- **New tests added**: 34 test cases total
  - `TestParseTailscaleVersion`: 11 test cases
  - `TestEnsureInstalled_ConfigValidation`: 5 test cases
  - `TestEnsureInstalled_OSCompatibility`: 9 test cases
  - `TestEnsureInstalled_VersionEnforcement`: 7 test cases
  - `TestTailscaleProvider_EnsureInstalled_VerificationFails`: 1 test case
  - Updated: `TestTailscaleProvider_EnsureInstalled_InstallFails`: improved
- **Remaining gaps (if any)**: Very close to 80% target (79.6%), may need small additional slice to reach ≥80%

---

## 7. Coverage Evolution Summary

> High level history that will match or cross reference `COVERAGE_LEDGER.md`.

| Date       | Change source                | Coverage before | Coverage after | Notes                                  |
|------------|-----------------------------|-----------------|----------------|----------------------------------------|
| 2025-XX-XX | Slice 1 implementation      | 68.2%           | 71.3%          | Extracted 4 helpers, added 7 test functions |
| 2025-XX-XX | Slice 2 implementation      | 71.3%           | 79.6%          | Added 34 test cases, version parsing, error paths |
| ...        | Future improvements         | -               | -              | Final push to ≥80% if needed           |

---

## 8. Governance and Spec Adjustments

> Capture any spec and governance changes specific to this provider.

- **Spec version changes**:
  - Version parsing rules added (section 2.2): Must strip build metadata and patch suffixes, return errors for unparseable versions
  - OS compatibility clarification: Only target hosts checked, orchestrator OS irrelevant
  - Tag computation clarification: Provider receives final computed tags in `opts.Tags`, does not compute `role_tags` directly
  - Tailnet name matching: Expected tailnet name derived from `tailnet_domain` (portion before first dot)
  - Tag matching semantics: Subset match (all expected tags must be present, extras allowed)
  - Offline node handling: Offline nodes with correct config are acceptable

- **Breaking or behavioural changes**: None - all changes align with existing behavior or implement previously missing spec requirements

- **Governance decisions that impact this provider**:
  - Coverage target: ≥80% for v1 (currently 79.6%, very close)
  - Test determinism: All tests must pass with `-race` and `-count=20`
  - No flaky patterns: No `time.Sleep`, no real network calls in tests

- **Links to relevant ADRs**: None specific to this provider

---

## 9. Open Questions and Future Work

> Reserved for post v1 or follow up slices.

- **Potential Slice 3 focus**: Final push to ≥80% coverage if needed (currently 79.6%, very close to target)

- **Known tradeoffs that might be revisited**:
  - Version comparison uses simple string comparison for v1 (lexicographic). Future: use `golang.org/x/mod/semver.Compare()` for proper semantic version comparison
  - OS support limited to Linux Debian/Ubuntu for v1. Future: Support macOS, Windows, and other Linux distributions

- **Integration with future network providers**:
  - Headscale (self-hosted Tailscale) support may be added in future
  - Provider remains provider-agnostic, core does not hardcode Tailscale-specific logic

- **Tooling or UX improvements around Tailscale**:
  - Auth key rotation automation (currently user responsibility)
  - Custom install methods beyond official script
  - Tailscale API integration for advanced operations

---

## 10. Migration Notes

> Use this section temporarily while consolidating existing docs into this log.

- [x] Migrated SLICE1 plan content
- [x] Migrated SLICE1 checklist and ready content
- [x] Migrated SLICE1 agent content
- [x] Migrated SLICE2 plan content
- [x] Migrated SLICE2 checklist and completeness content
- [x] Migrated SLICE2 dependencies and coverage expectations
- [x] Migrated Tailscale specific coverage and PR description notes

Once migration is complete this checklist can be removed or marked as complete.

# DEV_HOSTS Implementation Outline

> This document defines the v1 implementation plan for DEV_HOSTS. It translates the feature analysis brief into a concrete, testable, spec aligned delivery plan.

> All details in this outline must be reflected in `spec/dev/hosts.md` before any tests or code are written.

⸻

## 1. Feature Summary

**Feature ID:** DEV_HOSTS  
**Domain:** dev

**Goal:**

Manage `/etc/hosts` entries for dev domains used by `stagecraft dev`. Automatically add entries when dev stack starts and remove them when it exits, providing a seamless local development experience.

**v1 Scope:**

- Single host dev environments.
- Dev domains:
  - Frontend dev domain (e.g., `app.localdev.test`).
  - Backend dev domain (e.g., `api.localdev.test`).
- Hosts file toggle via `--no-hosts`:
  - When disabled, no hosts file modification occurs.
- Cross-platform support:
  - Linux: `/etc/hosts`
  - macOS: `/etc/hosts`
  - Windows: `C:\Windows\System32\drivers\etc\hosts`
- Deterministic entry format:
  - Entries marked with Stagecraft comment.
  - Lexicographically sorted domains.
  - Idempotent add/remove operations.

**Out of scope for v1:**

- Multi-host dev environments.
- DNS server integration (dnsmasq, etcd, etc.).
- System-level DNS overrides (systemd-resolved, etc.).
- Advanced entry management (TTL, multiple IPs, etc.).

**Future extensions (not implemented in v1):**

- Support for multiple IP addresses per domain.
- TTL or expiration policies for entries.
- Integration with DNS servers for multi-host scenarios.

⸻

## 2. Problem Definition and Motivation

CLI_DEV needs a simple, deterministic way to:

- Add dev domain entries to the hosts file when the dev stack starts.
- Remove dev domain entries when the dev stack exits.
- Preserve all existing hosts file entries.
- Handle cross-platform paths and permissions.

DEV_HOSTS provides this by:

- Parsing and modifying the hosts file in a safe, deterministic manner.
- Marking entries as Stagecraft-managed for safe cleanup.
- Providing idempotent add/remove operations.
- Handling cross-platform differences transparently.

This allows `stagecraft dev` to deliver a complete dev experience where domains resolve automatically without manual configuration.

⸻

## 3. User Stories (v1)

### Developer

- As a developer, I want `stagecraft dev` to automatically add dev domain entries to my hosts file so I can access dev domains without manual configuration.

- As a developer, I want hosts file entries to be automatically removed when `stagecraft dev` exits so I don't accumulate stale entries.

- As a developer, I want to disable hosts file modification with `--no-hosts` when I prefer manual DNS management.

### Platform Engineer

- As a platform engineer, I want hosts file entries to be clearly marked as Stagecraft-managed so I can identify and audit them.

- As a platform engineer, I want DEV_HOSTS to preserve existing hosts file entries and only modify Stagecraft-managed entries.

### CI / Automation

- As a CI pipeline, I want to run `stagecraft dev --no-hosts` without requiring hosts file modification, so tests can run in restricted environments.

⸻

## 4. Inputs and API Contract

### 4.1 API Surface (v1)

```go
// internal/dev/hosts/manager.go (new file)

package hosts

import (
    "context"
    "fmt"
)

// Manager manages hosts file entries for dev domains.
type Manager interface {
    // AddEntries adds dev domain entries to the hosts file.
    // Entries are marked as Stagecraft-managed and are idempotent.
    AddEntries(ctx context.Context, domains []string) error

    // RemoveEntries removes Stagecraft-managed entries from the hosts file.
    // Only entries marked as Stagecraft-managed are removed.
    RemoveEntries(ctx context.Context, domains []string) error

    // Cleanup removes all Stagecraft-managed entries from the hosts file.
    Cleanup(ctx context.Context) error
}

// NewManager creates a new hosts file manager for the current platform.
func NewManager() Manager {
    // Platform-specific implementation
}

// Options captures hosts file behavior (for future use).
type Options struct {
    HostsFilePath string // Optional override, default is platform-specific
    Verbose       bool
}
```

### 4.2 Entry Format

Stagecraft-managed entries use the following format:

```text
127.0.0.1    app.localdev.test api.localdev.test    # Stagecraft managed
```

Rules:

- IP address: `127.0.0.1` (hardcoded for v1).
- Domains: Lexicographically sorted.
- Comment: `# Stagecraft managed` (exact match for identification).
- Multiple domains per line are allowed (space-separated).

### 4.3 Inputs

- From CLI_DEV:
  - Dev domains (frontend and backend domains from `dev.ComputeDomains()`).
  - `--no-hosts` flag value.

⸻

## 5. Data Structures

### 5.1 Entry Representation

```go
// internal/dev/hosts/parser.go (new file)

package hosts

// Entry represents a single hosts file entry.
type Entry struct {
    IP       string   // IP address (e.g., "127.0.0.1")
    Domains  []string // Domain names (lexicographically sorted)
    Comment  string   // Optional comment
    Managed  bool     // true if this is a Stagecraft-managed entry
}

// File represents a parsed hosts file.
type File struct {
    Entries []Entry
    // Preserve original formatting where possible
}
```

### 5.2 Platform Detection

```go
// internal/dev/hosts/platform.go (new file)

package hosts

import "runtime"

// HostsFilePath returns the platform-specific hosts file path.
func HostsFilePath() string {
    switch runtime.GOOS {
    case "windows":
        return `C:\Windows\System32\drivers\etc\hosts`
    default: // linux, darwin, etc.
        return "/etc/hosts"
    }
}
```

⸻

## 6. Behavior and Determinism Rules

### 6.1 Adding Entries

1. Parse the hosts file.
2. Check if entries already exist for the requested domains.
3. If entries exist and are Stagecraft-managed, do nothing (idempotent).
4. If entries exist but are not Stagecraft-managed, preserve them and add new Stagecraft entries.
5. If entries don't exist, add them in a new line with Stagecraft comment.
6. Sort domains lexicographically within the entry.
7. Write the file atomically (write to temp file, then rename).

### 6.2 Removing Entries

1. Parse the hosts file.
2. Identify Stagecraft-managed entries (by comment `# Stagecraft managed`).
3. Remove only Stagecraft-managed entries.
4. Preserve all other entries unchanged.
5. Write the file atomically.

### 6.3 Determinism Guarantees

- Entries are always written in lexicographically sorted order.
- File format is stable (consistent whitespace, comment placement).
- Idempotent operations: running add/remove multiple times produces identical results.
- No timestamps or random data in entries.

⸻

## 7. Error Handling

### 7.1 Permission Errors

- **Error**: Cannot write to hosts file (permission denied).
- **Behavior**: Return error with clear message about sudo/elevation requirements.
- **Message**: `"dev: cannot modify hosts file: permission denied. Run with sudo or administrator privileges, or use --no-hosts to skip hosts file modification"`.

### 7.2 File Locking Errors

- **Error**: Hosts file is locked by another process.
- **Behavior**: Retry with exponential backoff (up to 3 retries), then return error.
- **Message**: `"dev: hosts file is locked by another process. Please close other applications that may be using the hosts file"`.

### 7.3 Invalid Format Errors

- **Error**: Hosts file has invalid format that cannot be parsed.
- **Behavior**: Preserve what's possible, add new entries, return warning (not error).
- **Message**: `"dev: hosts file has invalid format. Stagecraft entries added, but some existing entries may not be preserved correctly"`.

### 7.4 Missing File Errors

- **Error**: Hosts file does not exist.
- **Behavior**: Create the file with Stagecraft entries only.
- **Message**: None (silent creation is acceptable).

⸻

## 8. Integration with CLI_DEV

### 8.1 Integration Points

1. **After domain computation**:
   - CLI_DEV calls `devhosts.Manager.AddEntries()` with computed domains.
   - This happens after `dev.ComputeDomains()` and before starting processes.

2. **On exit/interrupt**:
   - CLI_DEV calls `devhosts.Manager.Cleanup()` during shutdown.
   - This happens in the same cleanup path as DEV_PROCESS_MGMT teardown.

3. **Flag handling**:
   - When `--no-hosts` is set, CLI_DEV skips all DEV_HOSTS calls.

### 8.2 Code Changes

**File**: `internal/cli/commands/dev.go`

```go
// After domain computation:
if !opts.NoHosts {
    hostsMgr := devhosts.NewManager()
    if err := hostsMgr.AddEntries(ctx, []string{domains.Frontend, domains.Backend}); err != nil {
        return fmt.Errorf("dev: add hosts entries: %w", err)
    }
    // Store manager for cleanup
}

// In cleanup handler (defer or signal handler):
if !opts.NoHosts && hostsMgr != nil {
    if err := hostsMgr.Cleanup(ctx); err != nil {
        // Log error but don't fail (best-effort cleanup)
        log.Errorf("dev: cleanup hosts entries: %v", err)
    }
}
```

⸻

## 9. Testing Plan

### 9.1 Unit Tests

**File**: `internal/dev/hosts/hosts_test.go`

Test cases:

1. **AddEntries**:
   - Add entries to empty hosts file.
   - Add entries to existing hosts file (preserve others).
   - Idempotent add (no duplicates).
   - Lexicographic sorting of domains.

2. **RemoveEntries**:
   - Remove Stagecraft entries (preserve others).
   - Idempotent remove (no errors if already removed).
   - Only remove Stagecraft-managed entries.

3. **Cleanup**:
   - Remove all Stagecraft entries.
   - Preserve non-Stagecraft entries.

4. **Parser**:
   - Parse valid hosts file.
   - Handle comments correctly.
   - Handle multiple domains per line.
   - Handle empty lines.

5. **Platform**:
   - Correct path detection for each platform.
   - Path override via options.

6. **Error handling**:
   - Permission errors (mock file system).
   - File locking errors (mock locking).
   - Invalid format handling.

### 9.2 Integration Tests

**File**: `internal/cli/commands/dev_test.go` (extend existing)

Test cases:

1. **CLI_DEV integration**:
   - `--no-hosts` false: entries added.
   - `--no-hosts` true: entries not added.
   - Cleanup on exit (mock hosts file).

2. **Error propagation**:
   - Hosts file errors surface correctly.
   - `--no-hosts` bypasses errors.

### 9.3 Golden Tests (Optional)

**File**: `internal/dev/hosts/testdata/`

Golden files for:

- Empty hosts file → after adding entries.
- Existing hosts file → after adding entries (preserve format).
- Hosts file with Stagecraft entries → after cleanup.

⸻

## 10. Implementation Checklist

### Before coding:

- [x] Analysis brief approved
- [ ] This outline approved
- [ ] Spec updated to match outline (`spec/dev/hosts.md`)
- [ ] Platform detection logic designed

### During implementation:

1. **Phase 1: Core parser and manager**
   - Create `internal/dev/hosts/` package.
   - Implement hosts file parser.
   - Implement entry manager (add/remove).
   - Add unit tests.

2. **Phase 2: Platform support**
   - Implement platform detection.
   - Handle cross-platform paths.
   - Test on multiple platforms (or mock).

3. **Phase 3: CLI_DEV integration**
   - Wire DEV_HOSTS into CLI_DEV.
   - Add cleanup on exit.
   - Add integration tests.

4. **Phase 4: Error handling**
   - Implement permission error handling.
   - Implement file locking handling.
   - Add error tests.

### After implementation:

- [ ] Update docs if tests cause outline changes
- [ ] Ensure lifecycle completion in `spec/features.yaml`
- [ ] Run full test suite and verify no regressions
- [ ] Test on Linux, macOS, and Windows (or document limitations)

⸻

## 11. Completion Criteria

The feature is complete only when:

- [ ] All v1 API methods implemented and tested
- [ ] Hosts file parsing is deterministic
- [ ] Add/remove operations are idempotent
- [ ] Cross-platform support works (or limitations documented)
- [ ] CLI_DEV integration complete
- [ ] All tests pass
- [ ] Spec and outline match actual behavior
- [ ] Determinism guarantees enforced
- [ ] Feature status updated to `done` in `spec/features.yaml`

⸻

## 12. Implementation Order

1. **Parser and data structures** (foundation)
   - Enables hosts file reading/writing
   - Can be tested in isolation

2. **Manager implementation** (core)
   - Add/remove entry logic
   - Idempotent operations

3. **Platform support** (cross-platform)
   - Path detection
   - Platform-specific handling

4. **CLI_DEV integration** (integration)
   - Wire into CLI_DEV
   - Cleanup on exit

5. **Error handling** (robustness)
   - Permission errors
   - File locking
   - Invalid format

This order ensures each component can be built and tested independently before integration.

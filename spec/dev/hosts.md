---
feature: DEV_HOSTS
version: v1
status: done
domain: dev
---

# DEV_HOSTS - Hosts File Management

⸻

## 1. Overview

DEV_HOSTS defines how Stagecraft manages `/etc/hosts` entries for dev domains used by `stagecraft dev`.

It covers:

- When and how hosts file entries are added.
- Where entries are stored (platform-specific paths).
- How entries are marked as Stagecraft-managed.
- How entries are removed on cleanup.
- How hosts file management is toggled via CLI flags.

DEV_HOSTS does not:

- Modify DNS server configuration.
- Manage system-level DNS overrides.
- Handle multi-host DNS resolution.

⸻

## 2. Behavior

### 2.1 Hosts File Toggle

DEV_HOSTS is controlled by the CLI flag:

- `--no-hosts` (boolean)

Semantics:

- When `--no-hosts` is set:

  - Hosts file management is disabled for the dev run.

  - DEV_HOSTS is effectively bypassed.

  - No hosts file modification occurs.

  - CLI_DEV skips all DEV_HOSTS calls.

- When `--no-hosts` is not set:

  - Hosts file management is enabled.

  - DEV_HOSTS adds entries for dev domains when `stagecraft dev` starts.

  - DEV_HOSTS removes entries when `stagecraft dev` exits.

### 2.2 Hosts File Paths

DEV_HOSTS uses platform-specific hosts file paths:

- **Linux**: `/etc/hosts`
- **macOS**: `/etc/hosts`
- **Windows**: `C:\Windows\System32\drivers\etc\hosts`

The path is detected automatically based on `runtime.GOOS`. An optional override may be provided via `Options.HostsFilePath` for testing.

### 2.3 Entry Format

Stagecraft-managed entries use the following format:

```text
127.0.0.1    app.localdev.test api.localdev.test    # Stagecraft managed
```

Rules:

- **IP address**: `127.0.0.1` (hardcoded for v1, always localhost).
- **Domains**: Space-separated, lexicographically sorted.
- **Comment**: `# Stagecraft managed` (exact match required for identification).
- **Multiple domains per line**: Allowed (space-separated).

Example with multiple domains:

```text
127.0.0.1    api.localdev.test app.localdev.test    # Stagecraft managed
```

### 2.4 Adding Entries

When `stagecraft dev` starts and `--no-hosts` is not set:

1. DEV_HOSTS parses the hosts file.
2. DEV_HOSTS checks if entries already exist for the requested domains.
3. If entries exist and are Stagecraft-managed, no action is taken (idempotent).
4. If entries exist but are not Stagecraft-managed, they are preserved and new Stagecraft entries are added.
5. If entries don't exist, they are added in a new line with the Stagecraft comment.
6. Domains are sorted lexicographically within the entry.
7. The file is written atomically (write to temp file, then rename).

**Timing**: Entries are added after domain computation and before starting processes.

### 2.5 Removing Entries

When `stagecraft dev` exits (normal or interrupted):

1. DEV_HOSTS parses the hosts file.
2. DEV_HOSTS identifies Stagecraft-managed entries (by comment `# Stagecraft managed`).
3. DEV_HOSTS removes only Stagecraft-managed entries.
4. All other entries are preserved unchanged.
5. The file is written atomically.

**Timing**: Entries are removed during CLI_DEV cleanup, in the same path as DEV_PROCESS_MGMT teardown.

### 2.6 Idempotent Operations

DEV_HOSTS operations are idempotent:

- Running `AddEntries()` multiple times with the same domains produces identical results (no duplicates).
- Running `RemoveEntries()` or `Cleanup()` multiple times produces identical results (no errors if already removed).

### 2.7 Determinism

DEV_HOSTS must satisfy the following determinism guarantees:

- Entries are always written in lexicographically sorted order.
- File format is stable (consistent whitespace, comment placement).
- Idempotent operations produce identical results.
- No timestamps or random data in entries.

⸻

## 3. CLI Integration

CLI_DEV integrates DEV_HOSTS as follows:

1. CLI_DEV:

   - Computes dev domains via `dev.ComputeDomains()`.

   - Checks the `--no-hosts` flag.

2. If `--no-hosts` is not set:

   - CLI_DEV creates a DEV_HOSTS manager: `devhosts.NewManager()`.

   - CLI_DEV calls `manager.AddEntries(ctx, []string{domains.Frontend, domains.Backend})`.

   - CLI_DEV stores the manager for cleanup.

3. On exit/interrupt:

   - CLI_DEV calls `manager.Cleanup(ctx)` to remove Stagecraft-managed entries.

   - Cleanup errors are logged but do not fail the command (best-effort cleanup).

4. If `--no-hosts` is set:

   - CLI_DEV skips all DEV_HOSTS calls.

⸻

## 4. Error Handling and Exit Codes

DEV_HOSTS uses the following error handling via CLI_DEV:

| Error Condition | Behavior | Exit Code (via CLI_DEV) |
|----------------|----------|------------------------|
| Permission denied | Return error with clear message about sudo/elevation | 1 (invalid input) |
| File locked | Retry with backoff, then return error | 2 (external failure) |
| Invalid format | Preserve what's possible, add entries, return warning | 0 (success with warning) |
| Missing file | Create file with Stagecraft entries only | 0 (success) |

**Error Messages:**

- Permission denied: `"dev: cannot modify hosts file: permission denied. Run with sudo or administrator privileges, or use --no-hosts to skip hosts file modification"`

- File locked: `"dev: hosts file is locked by another process. Please close other applications that may be using the hosts file"`

- Invalid format: `"dev: hosts file has invalid format. Stagecraft entries added, but some existing entries may not be preserved correctly"`

⸻

## 5. Determinism

DEV_HOSTS must satisfy the following determinism guarantees:

- For a given set of domains and hosts file state:

  - The resulting hosts file structure is always identical.

  - Entries are written in lexicographically sorted order.

  - Idempotent operations produce identical results.

- DEV_HOSTS does not:

  - Introduce random identifiers.

  - Include timestamps or machine-specific data.

  - Depend on file system ordering.

⸻

## 6. Cross-Platform Considerations

### 6.1 Path Detection

DEV_HOSTS automatically detects the platform and uses the correct hosts file path:

- Linux: `/etc/hosts`
- macOS: `/etc/hosts`
- Windows: `C:\Windows\System32\drivers\etc\hosts`

Detection is based on `runtime.GOOS`.

### 6.2 Permissions

- **Linux/macOS**: Modifying `/etc/hosts` typically requires sudo/elevation. DEV_HOSTS returns clear error messages when permissions are insufficient.

- **Windows**: Modifying the hosts file typically requires administrator privileges. DEV_HOSTS returns clear error messages when permissions are insufficient.

### 6.3 File Locking

Hosts files may be locked by other processes (e.g., antivirus software, system services). DEV_HOSTS:

- Retries with exponential backoff (up to 3 retries).
- Returns clear error messages if locking persists.
- Does not corrupt the hosts file if locking occurs during write.

⸻

## 7. Safety Guarantees

DEV_HOSTS provides the following safety guarantees:

- **Preserve existing entries**: Only Stagecraft-managed entries are removed; all other entries are preserved unchanged.

- **Clear marking**: Stagecraft-managed entries are clearly marked with `# Stagecraft managed` comment for identification.

- **Atomic writes**: Hosts file writes are atomic where possible (write to temp file, then rename) to avoid corruption.

- **Idempotent operations**: Running add/remove multiple times produces identical results.

⸻

## 8. Tests

Minimum required tests:

- Unit tests in `internal/dev/hosts/hosts_test.go` for:

  - Adding entries to empty hosts file.
  - Adding entries to existing hosts file (preserve others).
  - Removing Stagecraft entries (preserve others).
  - Idempotent operations.
  - Lexicographic sorting.
  - Platform path detection.
  - Error handling (permissions, locking, invalid format).

- Integration tests in `internal/cli/commands/dev_test.go` for:

  - CLI_DEV integration with `--no-hosts` false.
  - CLI_DEV integration with `--no-hosts` true.
  - Cleanup on exit.

All tests must be deterministic and not depend on real hosts file access (use temporary files or mocks).

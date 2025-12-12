> **Superseded by** `docs/coverage/COVERAGE_LEDGER.md` section 3 (Historical Coverage Timeline). Kept for historical reference. New coverage snapshots and summaries MUST go into the coverage ledger.

## Summary

Complete DEV_HOSTS and full CLI_DEV v1 integration, closing out Phase 3: Local Development.

This PR:

- Adds a deterministic, cross platform `/etc/hosts` management subsystem (DEV_HOSTS).
- Integrates DEV_HOSTS into `stagecraft dev` via CLI_DEV.
- Wires CLI_DEV to the full Phase 3 dev topology (domains, mkcert, compose, traefik, process management).
- Updates specs and status docs so Phase 3 is now 100% complete.

---

## Feature IDs

- `CLI_DEV`
- `DEV_HOSTS`

---

## Changes

### 1. DEV_HOSTS planning and spec

- Added analysis and planning docs:
  - `docs/engine/analysis/DEV_HOSTS.md` - Feature Analysis Brief
  - `docs/engine/outlines/DEV_HOSTS_IMPLEMENTATION_OUTLINE.md` - Implementation Outline
- Added spec:
  - `spec/dev/hosts.md`
    - Defines:
      - `--no-hosts` flag semantics
      - Platform specific hosts paths
      - Entry format and `# Stagecraft managed` marker
      - Idempotent add and cleanup semantics
      - Error handling and determinism requirements

### 2. DEV_HOSTS implementation

New package: `internal/dev/hosts/`:

- `platform.go`
  - `FilePath()` for Linux/macOS (`/etc/hosts`) and Windows (`C:\Windows\System32\drivers\etc\hosts`).

- `parser.go`
  - `ParseFile(path)` - tolerant parser that:
    - Preserves comments and unknown lines.
    - Identifies Stagecraft managed entries via `# Stagecraft managed`.
    - Sorts domains lexicographically within each entry.
  - `WriteFile(path, *File)` - atomic writer:
    - Writes to `path.tmp`, then `os.Rename`.
    - Deterministic output format.
  - `File` helpers:
    - `HasDomains(domains []string)`
    - `RemoveManagedEntries()`
    - `AddManagedEntry(domains []string)` - idempotent, `127.0.0.1`, Stagecraft comment.

- `manager.go`
  - `Manager` interface:
    - `AddEntries(ctx, []string)`
    - `RemoveEntries(ctx, []string)`
    - `Cleanup(ctx)`
  - `NewManager()` and `NewManagerWithOptions(Options)` for injection/testing.
  - `parseWithRetry`, `writeWithRetry`:
    - Simple retry with backoff for file locking.
    - Clear permission error message:
      - Suggests sudo/admin or `--no-hosts`.
    - File lock error surfaced after retries.

### 3. DEV_HOSTS tests

- `internal/dev/hosts/hosts_test.go`
  - `TestFilePath` - platform sanity check.
  - Parser tests:
    - Empty file, missing file, valid entries, Stagecraft managed entries, comments.
  - Writer tests:
    - Round trip parse → write → parse.
  - Behavior tests:
    - `File.AddManagedEntry`:
      - Adds `127.0.0.1`, Stagecraft comment, sorted domains.
      - Idempotent behavior on repeated calls.
    - `File.RemoveManagedEntries`:
      - Removes only managed entries, preserves others.
  - Manager tests (where applicable) for add and cleanup behavior.

---

### 4. CLI_DEV integration

File: `internal/cli/commands/dev.go`

- Flags:
  - Confirms presence and types of:
    - `--env`
    - `--config`
    - `--no-https`
    - `--no-hosts`
    - `--no-traefik`
    - `--detach`
    - `--verbose`

- `runDevWithOptions` now:

  1. Validates `Env` is non-empty.
  2. Loads config via `loadConfigForEnv`.
  3. Computes dev domains with `dev.ComputeDomains(cfg, opts.Env)`:
     - Config driven with deterministic defaults:
       - `frontend`: `app.localdev.test`
       - `backend`: `api.localdev.test`
  4. Hosts management (when `--no-hosts` is **not** set):
     - Creates `devhosts.Manager`.
     - Calls `AddEntries(ctx, []string{domains.Frontend, domains.Backend})`.
     - Defers `Cleanup` on `context.Background()` and logs errors to `stderr` without failing the command.
  5. HTTPS integration via DEV_MKCERT:
     - Uses `devmkcert.NewGenerator().EnsureCertificates(...)` with:
       - `DevDir: ".stagecraft/dev"`
       - Domains from computed dev domains.
       - `EnableHTTPS: !opts.NoHTTPS`
       - `Verbose: opts.Verbose`.
  6. Topology building:
     - Uses `dev.NewDefaultBuilder().ResolveServiceDefinitions(...)`.
     - Backend service is required; errors if `backendSvc == nil`.
     - Optionally includes Traefik service definition when `--no-traefik` is not set.
  7. Delegates to dev topology/compose/process management to actually launch the stack.

---

### 5. CLI_DEV tests

File: `internal/cli/commands/dev_test.go`

- `TestNewDevCommand_HasExpectedFlags`
  - Asserts presence and types of all CLI_DEV flags.

- `TestNewDevCommand_DefaultsAndRun`
  - Uses a temp `stagecraft.yml` with a minimal dev backend.
  - Runs:
    - `stagecraft dev --config <path> --no-https --no-hosts`
    - Allows docker compose failures (CI environment) but still asserts that `.stagecraft/dev/compose.yaml` is written.

- `TestRunDevWithOptions_EmptyEnvFails` and `TestNewDevCommand_EmptyEnvFails`
  - Ensure empty `--env` is rejected.

- `TestRunDevWithOptions_BuildsTopology`
  - Similar to the command level test but calls `runDevWithOptions` directly.
  - Confirms `.stagecraft/dev/compose.yaml` is written even if docker compose fails.

- Tests updated to include `--no-hosts` (or `NoHosts: true`) where needed so CI does not require `/etc/hosts` write permissions.

---

## Phase 3 status

- `Phase 3: Local Development` is now functionally complete:
  - `CLI_DEV_BASIC` - done
  - `CLI_DEV` - done
  - `DEV_MKCERT` - done
  - `DEV_TRAEFIK` - done
  - `DEV_COMPOSE_INFRA` - done
  - `DEV_PROCESS_MGMT` - done
  - `DEV_HOSTS` - done
  - `PROVIDER_BACKEND_ENCORE` - done
  - `PROVIDER_BACKEND_GENERIC` - done
  - `PROVIDER_FRONTEND_GENERIC` - done

Status docs are regenerated from `spec/features.yaml` in this PR so that Phase 3 now shows 100% completion.

---

## Tests

- Unit tests:
  - `go test ./internal/dev/hosts/...`
  - `go test ./internal/cli/commands/...`
- Smokes:
  - `go test ./test/e2e/...` (with `--no-https` and `--no-hosts` where needed)
- Full suite:
  - `./scripts/run-all-checks.sh`

---

## Risks and mitigations

- **Hosts file modification on developer machines**
  - Mitigations:
    - Opt out with `--no-hosts`.
    - Stagecraft managed entries are clearly marked and removed on cleanup.
    - Best effort cleanup. Errors are logged to `stderr` without blocking shutdown.

- **Permission issues on `/etc/hosts`**
  - Mitigations:
    - Clear error messages instructing to use sudo/admin or `--no-hosts`.
    - CI tests use `--no-hosts` to avoid requiring elevated privileges.

- **Determinism**
  - Deterministic domain defaults.
  - Lexicographically sorted domains in hosts entries.
  - Atomic write and idempotent add/cleanup semantics.

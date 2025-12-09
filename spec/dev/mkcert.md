---
feature: DEV_MKCERT
version: v1
status: done
domain: dev
---

# DEV_MKCERT - Dev Certificate Management

⸻

## 1. Overview

DEV_MKCERT defines how Stagecraft provisions and manages local TLS certificates for dev domains used by `stagecraft dev`.

It covers:

- When and how mkcert is invoked.

- Where certificates are stored.

- How certificate configuration is exposed to other features (CLI_DEV and DEV_TRAEFIK).

- How HTTPS is toggled via CLI flags.

DEV_MKCERT does not:

- Modify system or browser trust stores.

- Generate certificates for production environments.

- Manage certificate renewal policies beyond presence checks.

⸻

## 2. Behavior

### 2.1 HTTPS Toggle

DEV_MKCERT is controlled by the CLI flag:

- `--no-https` (boolean)

Semantics:

- When `--no-https` is set:

  - HTTPS is disabled for the dev run.

  - DEV_MKCERT is effectively bypassed.

  - No mkcert calls are made.

  - `CertConfig.Enabled` is false.

- When `--no-https` is not set:

  - HTTPS is enabled.

  - DEV_MKCERT ensures certificates exist for dev domains via mkcert.

### 2.2 Certificate Directory

DEV_MKCERT uses a deterministic directory layout:

- Dev root directory: `.stagecraft/dev` (controlled by CLI_DEV).

- Certificate directory:

  ```text
  .stagecraft/dev/certs/
  ```

This directory:

- Is created if it does not exist, with permissions `0o755`.

- Contains all certificate and key files managed by DEV_MKCERT.

### 2.3 Certificate Files

In v1, DEV_MKCERT manages a single certificate pair that covers all dev domains:

- Certificate file:

  ```
  .stagecraft/dev/certs/dev-local.pem
  ```

- Key file:

  ```
  .stagecraft/dev/certs/dev-local-key.pem
  ```

These files:

- Are used by Traefik as the TLS certificate and key for `https://` endpoints.

- Are referenced by full path from DEV_TRAEFIK.

DEV_MKCERT does not manage separate certificates per domain in v1.

### 2.4 Domains

CLI_DEV provides the list of dev domains to DEV_MKCERT, typically:

- Frontend dev domain (for example `app.localdev.test`).

- Backend dev domain (for example `api.localdev.test`).

Rules:

- DEV_MKCERT removes duplicate domains.

- DEV_MKCERT sorts domains lexicographically.

- Sorted domains are used as arguments to mkcert.

### 2.5 mkcert Invocation

When HTTPS is enabled and any expected certificate file is missing:

1. DEV_MKCERT checks for mkcert in PATH.

   - Binary name:

     - `mkcert` by default.

     - Overridable via configuration or options if needed.

2. If mkcert is not found:

   - DEV_MKCERT returns an error with a message similar to:

     ```
     dev: mkcert binary not found; install mkcert or run with --no-https
     ```

3. If mkcert is found:

   - DEV_MKCERT runs mkcert with the following structure, using `<CertDir>` as current working directory:

     ```
     mkcert -cert-file dev-local.pem -key-file dev-local-key.pem <sorted domains...>
     ```

   - On success, mkcert produces:

     - `dev-local.pem`

     - `dev-local-key.pem`

     in `<CertDir>`.

   - DEV_MKCERT sets:

     - `CertConfig.Enabled = true`

     - `CertConfig.CertDir = <CertDir>`

     - `CertConfig.Domains = <sorted domains>`

     - `CertConfig.CertFile = <CertDir>/dev-local.pem`

     - `CertConfig.KeyFile = <CertDir>/dev-local-key.pem`

4. If mkcert fails:

   - DEV_MKCERT surfaces an error containing mkcert stderr output, wrapped with Stagecraft context.

### 2.6 Idempotency

DEV_MKCERT is idempotent:

- If both `dev-local.pem` and `dev-local-key.pem` already exist:

  - DEV_MKCERT does not invoke mkcert.

  - DEV_MKCERT returns CertConfig pointing to existing files.

- DEV_MKCERT does not delete or rotate certificates in v1.

⸻

## 3. CLI Integration

### 3.1 CLI Flags

Relevant flags for `stagecraft dev`:

| Flag | Default | Description |
|------|---------|-------------|
| `--no-https` | `false` | Disable mkcert and HTTPS integration for dev. |

Other flags like `--env`, `--no-traefik`, `--no-hosts`, `--detach`, `--verbose` are handled by CLI_DEV and DEV_PROCESS_MGMT but influence when DEV_MKCERT is invoked.

### 3.2 CLI_DEV Flow

When running `stagecraft dev`:

1. CLI_DEV computes dev topology and dev domains.

2. CLI_DEV determines if HTTPS is enabled:

   - `enableHTTPS = !noHTTPSFlag`.

3. If `enableHTTPS` is false:

   - CLI_DEV sets `CertConfig.Enabled = false`.

   - DEV_TRAEFIK is configured without TLS.

4. If `enableHTTPS` is true:

   - CLI_DEV invokes DEV_MKCERT to get CertConfig.

   - CLI_DEV passes CertConfig to DEV_TRAEFIK to configure TLS cert and key paths.

⸻

## 4. DEV_TRAEFIK Integration

DEV_TRAEFIK uses CertConfig as follows:

- When `CertConfig.Enabled` is true:

  - Traefik dynamic configuration includes TLS configuration referencing:

    - `CertConfig.CertFile` as the certificate file.

    - `CertConfig.KeyFile` as the key file.

  - Traefik routers for HTTPS entry points use TLS configuration.

- When `CertConfig.Enabled` is false:

  - Traefik is configured for HTTP only.

  - No TLS configuration is added.

DEV_MKCERT does not modify Traefik configuration directly.

⸻

## 5. Error Handling and Exit Codes

DEV_MKCERT error behavior is surfaced through CLI_DEV.

Example error cases:

1. HTTPS enabled, mkcert missing:

   - CLI_DEV receives an error from DEV_MKCERT.

   - CLI_DEV exits with exit code 2 (external provider failure).

   - Error message includes guidance to install mkcert or use `--no-https`.

2. HTTPS enabled, mkcert fails during generation:

   - CLI_DEV receives an error with mkcert stderr data wrapped.

   - CLI_DEV exits with exit code 2.

3. Invalid dev directory:

   - If DevDir is empty or invalid, DEV_MKCERT returns an error that CLI_DEV treats as an internal error or invalid input depending on context.

DEV_MKCERT does not perform exit code mapping itself. It returns Go errors that CLI_DEV maps to exit codes according to CLI governance.

⸻

## 6. Determinism

DEV_MKCERT must satisfy the following determinism guarantees:

- For a given DevDir, Domains set, EnableHTTPS flag, and mkcert binary:

  - The certificate directory and filenames are always the same.

  - The mkcert command line arguments are always the same.

  - CertConfig is always the same (up to path normalization).

- DEV_MKCERT must not:

  - Add random suffixes to file names.

  - Depend on non deterministic mkcert behavior for paths.

Runtime contents of certificate files (cryptographic material) are intentionally non deterministic, but Stagecraft does not inspect them.

⸻

## 7. Non Goals

DEV_MKCERT explicitly does not:

- Modify system or browser trust stores.

- Ensure mkcert's local CA is trusted.

- Manage certificate expiration and rotation schedules.

- Generate certificates for production environments.

- Perform ACME or Let's Encrypt flows.

These are either user responsibilities or future features.

⸻

## 8. Testing Requirements

DEV_MKCERT must be covered by:

### Unit Tests

- HTTPS disabled:

  - `EnableHTTPS: false` results in a disabled CertConfig.

  - No mkcert invocation occurs.

- Existing certs:

  - With both `dev-local.pem` and `dev-local-key.pem` present:

    - EnsureCertificates returns enabled CertConfig and does not invoke mkcert.

- mkcert missing:

  - No existing certs and mkcert not present:

    - EnsureCertificates returns a clear error about mkcert not found.

- mkcert arguments:

  - With a set of domains:

    - EnsureCertificates deduplicates and sorts domains.

    - mkcert is invoked with deterministic argument order.

Tests use fake ExecCommander and Command implementations to avoid invoking real mkcert.

### Integration Tests (CLIDEV)

- When `--no-https` is set:

  - CLI_DEV does not attempt to use DEV_MKCERT.

- When `--no-https` is not set and mkcert is missing:

  - CLI_DEV surfaces a clear, deterministic error.

- When `--no-https` is not set and a fake mkcert succeeds:

  - CLI_DEV threads CertConfig through to Traefik generation (in future CLI tests).

Golden tests are not required for DEV_MKCERT because certificates are inherently non deterministic at the content level. Paths and logs may be asserted for determinism.

⸻

## 9. Lifecycle and Status

- Feature ID: DEV_MKCERT

- Initial state in `spec/features.yaml`: `todo` or `wip`.

- State becomes `done` only when:

  - DEV_MKCERT is implemented according to this spec.

  - Unit tests and CLI integration tests are complete and passing.

  - CLI_DEV and DEV_TRAEFIK correctly use CertConfig.

  - All behavior described in this spec is implemented and verified.


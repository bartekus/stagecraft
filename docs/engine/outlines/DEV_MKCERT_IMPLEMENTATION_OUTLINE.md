# DEV_MKCERT Implementation Outline

> This document defines the v1 implementation plan for DEV_MKCERT. It translates the feature analysis brief into a concrete, testable, spec aligned delivery plan.

> All details in this outline must be reflected in `spec/dev/mkcert.md` before any tests or code are written.

⸻

## 1. Feature Summary

**Feature ID:** DEV_MKCERT  
**Domain:** dev

**Goal:**

Provision and manage local TLS certificates for dev domains used by `stagecraft dev`, using `mkcert` as the external certificate generator. Provide deterministic certificate directory layout under `.stagecraft/dev/certs` and a `CertConfig` that can be consumed by DEV_TRAEFIK and CLI_DEV.

**v1 Scope:**

- Single host dev environments.

- Dev domains:

  - Frontend dev domain (for example `app.localdev.test`).

  - Backend dev domain (for example `api.localdev.test`).

- HTTPS toggle via `--no-https`:

  - When disabled, no mkcert calls are made.

- `mkcert` integration:

  - Check for mkcert in `PATH`.

  - Use mkcert to generate certificates if missing.

- Deterministic file layout:

  - `.stagecraft/dev/certs/` as root.

  - Stable filenames for certificates and keys.

**Out of scope for v1:**

- Certificate renewal policies beyond basic presence checks.

- Managing or automating trust store installation.

- Multi host or multi environment certificate sets.

- Let us Encrypt or any ACME integration.

- Complex domain sets beyond frontend and backend dev domains.

**Future extensions (not implemented in v1):**

- Supporting multiple domain profiles per environment.

- Expiration detection and regeneration policies.

- Automated guidance or helpers for trust store installation.

- Flexible CA management and alternate certificate tools.

⸻

## 2. Problem Definition and Motivation

CLI_DEV and DEV_TRAEFIK need a simple, deterministic way to get:

- A directory with certificates.

- Certificate and key paths associated with the dev domains.

- A clear on/off switch for HTTPS via `--no-https`.

DEV_MKCERT provides this by:

- Using mkcert to generate certificates for dev domains when required.

- Storing certificates in deterministic paths under `.stagecraft/dev/certs`.

- Returning a `CertConfig` that DEV_TRAEFIK can use to build TLS config.

This allows `stagecraft dev` to deliver an HTTPS enabled dev experience that mirrors production routing patterns.

⸻

## 3. User Stories (v1)

### Developer

- As a developer, I want `stagecraft dev` to automatically provision certificates for my dev domains so I can use HTTPS routes.

- As a developer, I want `stagecraft dev --no-https` to skip certificate provisioning when I do not need HTTPS or mkcert is unavailable.

- As a developer, I want cert files to be stored in a predictable `.stagecraft/dev/certs` directory so I can inspect or clean them if needed.

### Platform Engineer

- As a platform engineer, I want a deterministic certificate layout that is independent of machine specific paths.

- As a platform engineer, I want errors about mkcert or certificate generation to be clear and actionable.

### CI / Automation

- As a CI pipeline, I want to run `stagecraft dev --no-https` without requiring mkcert, to keep CI environments simple.

- As a CI pipeline, I want certificate generation to be deterministic when HTTPS is enabled in controlled environments.

⸻

## 4. Inputs and API Contract

### 4.1 API Surface (v1)

```go
// internal/dev/mkcert/generator.go (new file)

package mkcert

import (
    "context"

    "stagecraft/pkg/config"
)

// Options captures mkcert behavior.
type Options struct {
    DevDir       string   // Root dev directory, usually ".stagecraft/dev"
    Domains      []string // Dev domains to issue certs for
    EnableHTTPS  bool     // false when --no-https is set
    MkcertBinary string   // Optional override, default "mkcert"
    Verbose      bool
}

// CertConfig represents the resulting certificate configuration.
type CertConfig struct {
    Enabled bool
    CertDir string
    Domains []string

    // Paths to certificate and key files for Traefik (or other consumers).
    // For v1 we assume a single cert covering all domains.
    CertFile string
    KeyFile  string
}

// Generator provisions dev certificates using mkcert.
type Generator struct {
    exec ExecCommander
    log  Logger
}

// ExecCommander abstracts external command execution.
type ExecCommander interface {
    CommandContext(ctx context.Context, name string, args ...string) Command
}

// Command abstracts a running command.
type Command interface {
    Run() error
    SetStdout(w Writer)
    SetStderr(w Writer)
}

// Logger is a minimal logging abstraction.
type Logger interface {
    Infof(format string, args ...any)
    Errorf(format string, args ...any)
}

// Writer is a minimal writer abstraction.
type Writer interface {
    Write(p []byte) (n int, err error)
}

// NewGenerator creates a generator with default dependencies.
func NewGenerator() *Generator

// NewGeneratorWithDeps creates a generator with explicit dependencies (for tests).
func NewGeneratorWithDeps(execCmd ExecCommander, logger Logger) *Generator

// EnsureCertificates ensures certificates exist for the given domains.
//
// - If EnableHTTPS is false, returns a disabled CertConfig and does not call mkcert.
// - If EnableHTTPS is true, ensures CertDir exists under DevDir and certificates
//   exist for all requested domains, invoking mkcert if necessary.
func (g *Generator) EnsureCertificates(ctx context.Context, cfg *config.Config, opts Options) (*CertConfig, error)
```

### 4.2 Input Sources

- From CLI_DEV or dev topology:

  - DevDir - root dev directory, usually `.stagecraft/dev`.

  - Domains - dev domains used for frontend and backend routes.

  - EnableHTTPS - derived from `--no-https` flag (enable when `--no-https` is false).

  - Verbose - derived from `--verbose`.

- From config:

  - Optional overrides or future configuration for certificates.

### 4.3 Output

- CertConfig with:

  - Enabled - indicates whether HTTPS is active.

  - CertDir - absolute or project relative path to certificate directory.

  - Domains - domains covered by the certificate.

  - CertFile and KeyFile - paths for consumption by DEV_TRAEFIK.

- Created directories and certificate files under `.stagecraft/dev/certs`.

⸻

## 5. Behavior Details

### 5.1 HTTPS Toggle

- If EnableHTTPS is false:

  - EnsureCertificates returns a CertConfig with:

    - Enabled set to false.

    - CertDir empty.

    - Domains empty.

    - CertFile and KeyFile empty.

  - No filesystem or mkcert operations are performed.

  - Caller (CLI_DEV) will configure HTTP only routing.

### 5.2 Directory Layout

- Certificate root directory:

  ```
  <DevDir>/certs/
  ```

  For example:

  ```
  .stagecraft/dev/certs/
  ```

- v1 certificate file pattern (single cert covering all domains):

  ```
  .stagecraft/dev/certs/dev-local.pem    # certificate (full chain)
  .stagecraft/dev/certs/dev-local-key.pem
  ```

- Paths must be deterministic and documented so DEV_TRAEFIK can mount them.

### 5.3 mkcert Integration

When EnableHTTPS is true:

1. Generator checks if the certificate files already exist at:

   - `<CertDir>/dev-local.pem`

   - `<CertDir>/dev-local-key.pem`

2. If both files exist:

   - EnsureCertificates returns a CertConfig pointing to these files without invoking mkcert.

3. If any file is missing:

   - Generator checks for mkcert binary in PATH:

     - Name is `opts.MkcertBinary` if non empty, otherwise `mkcert`.

   - If mkcert is not found:

     - Returns an error indicating mkcert is required or suggests `--no-https`.

4. If mkcert is available:

   - Generator ensures `<CertDir>` exists (creates with `0o755`).

   - Generator invokes mkcert with deterministic arguments, for example:

     ```
     mkcert -cert-file dev-local.pem -key-file dev-local-key.pem <domains...>
     ```

   - Working directory is set to `<CertDir>` to keep outputs localized.

5. On success:

   - EnsureCertificates returns a CertConfig with:

     - Enabled: true

     - CertDir: `<CertDir>`

     - Domains: `opts.Domains` (deduplicated and sorted)

     - CertFile: `<CertDir>/dev-local.pem`

     - KeyFile: `<CertDir>/dev-local-key.pem`

6. On failure:

   - Returns an error with mkcert stderr wrapped with context (for example `dev: mkcert failed: ...`).

### 5.4 Determinism

- Domains are deduplicated and sorted lexicographically before invoking mkcert.

- Cert directory and filenames are fixed and documented.

- mkcert arguments follow a deterministic order:

  ```
  mkcert -cert-file dev-local.pem -key-file dev-local-key.pem <sorted domains...>
  ```

- No random flags or file name suffixes are used.

### 5.5 Logging

- When Verbose is true:

  - Log whether HTTPS is enabled.

  - Log the cert directory in use.

  - Log whether certificates are reused or being generated.

  - Log the exact mkcert command line (excluding sensitive data; here there are none beyond domains).

⸻

## 6. Determinism and Side Effects

### 6.1 Determinism Rules

- Given the same:

  - DevDir.

  - Domains set.

  - EnableHTTPS and MkcertBinary values.

- DEV_MKCERT must:

  - Produce the same CertConfig.

  - Invoke mkcert with identical arguments when certificates are missing.

  - Write certificates to the same deterministic paths.

### 6.2 Side Effect Constraints

- File system:

  - Creates directories under `.stagecraft/dev/certs`.

  - Writes certificate and key files in that directory.

- External commands:

  - Invokes mkcert only when HTTPS is enabled and certificates are missing.

- No network I/O:

  - mkcert may perform its own operations, but Stagecraft itself does not initiate network calls.

- No trust store modification:

  - DEV_MKCERT does not modify system or browser trust stores.

⸻

## 7. Integration with CLI_DEV and DEV_TRAEFIK

### CLI_DEV

- CLI_DEV determines:

  - DevDir: `.stagecraft/dev`.

  - Domains: dev domains for frontend and backend (for example `app.localdev.test`, `api.localdev.test`).

  - EnableHTTPS: inverted `--no-https` flag.

  - Verbose: `--verbose` flag.

- CLI_DEV calls:

  ```go
  certGen := mkcert.NewGenerator()
  certCfg, err := certGen.EnsureCertificates(ctx, cfg, mkcert.Options{...})
  ```

- CLI_DEV passes certCfg to the Traefik topology builder or DEV_TRAEFIK generator.

### DEV_TRAEFIK

- DEV_TRAEFIK consumes:

  - CertConfig.CertFile.

  - CertConfig.KeyFile.

  - CertConfig.Domains.

  - CertConfig.Enabled.

- When TLS is enabled:

  - DEV_TRAEFIK includes TLS configuration referencing these paths.

- When TLS is disabled:

  - DEV_TRAEFIK produces configs without TLS.

⸻

## 8. Testing Strategy

### Unit Tests

File: `internal/dev/mkcert/generator_test.go`.

- HTTPS disabled:

  - EnableHTTPS: false → returns disabled CertConfig, no filesystem or exec calls.

- Existing certs:

  - Pre create `dev-local.pem` and `dev-local-key.pem` in temp CertDir.

  - EnableHTTPS: true → EnsureCertificates reuses them and does not invoke mkcert.

- Missing mkcert:

  - No existing certs.

  - Fake ExecCommander that signals missing binary (or do not use it at all).

  - Expect helpful error message about mkcert not found.

- mkcert invocation:

  - Use fake ExecCommander and capture name and args.

  - Verify mkcert is invoked with:

    - `-cert-file dev-local.pem`.

    - `-key-file dev-local-key.pem`.

    - Sorted domain list.

### Integration Tests (CLIDEV level)

- Extend CLI_DEV tests to:

  - Use a fake mkcert generator or stub (via injection) to avoid real mkcert calls.

  - Assert that when HTTPS is enabled:

    - DEV_MKCERT is called.

    - Cert paths are threaded through to DEV_TRAEFIK (in future).

### Non Goals for Tests

- No tests should rely on mkcert being installed.

- No tests should modify real user directories or trust stores.

⸻

## 9. Implementation Plan Checklist

### Before coding:

- Analysis brief approved (`docs/engine/analysis/DEV_MKCERT.md`).

- This Implementation Outline approved.

- Spec created (`spec/dev/mkcert.md`) matching this outline.

### During implementation:

1. Create package structure

   - `internal/dev/mkcert/generator.go`

   - `internal/dev/mkcert/generator_test.go`

2. Implement types and generator

   - Define Options, CertConfig.

   - Implement Generator, NewGenerator, and NewGeneratorWithDeps.

   - Implement EnsureCertificates with behavior described above.

   - Implement default ExecCommander and Logger similar to DEV_PROCESS_MGMT.

3. Wire into CLI_DEV

   - Extend `runDevWithOptions` to:

     - Determine dev domains.

     - Call DEV_MKCERT before Traefik and Compose file generation (or as part of topology builder).

     - Pass CertConfig to Traefik generator / topology.

4. Add tests

   - Unit tests for DEV_MKCERT behavior.

   - CLI tests to ensure HTTPS disabled path does not require mkcert and error surfaces correctly if enabled without mkcert.

### After implementation:

- Update docs if tests cause outline changes.

- Ensure lifecycle completion in `spec/features.yaml` (DEV_MKCERT set to `done`).

- Run full test suite and verify no regressions.

⸻

## 10. Completion Criteria

DEV_MKCERT is considered complete when:

- EnsureCertificates returns deterministic CertConfig based on inputs.

- HTTPS disabled path is a pure no op and works in environments without mkcert.

- HTTPS enabled path generates or reuses certificates under `.stagecraft/dev/certs`.

- Integration with CLI_DEV and DEV_TRAEFIK is complete and tested.

- Spec `spec/dev/mkcert.md` matches actual behavior.

- DEV_MKCERT feature state in `spec/features.yaml` is updated to `done`.


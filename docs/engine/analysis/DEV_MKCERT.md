# DEV_MKCERT Feature Analysis Brief

This document captures the high level motivation, constraints, and success definition for DEV_MKCERT.

It is the starting point for the Implementation Outline and Spec.

This brief must be approved before outline work begins.

⸻

## 1. Problem Statement

CLI_DEV currently:

- Computes a deterministic dev topology.

- Generates dev files under `.stagecraft/dev` (Compose + Traefik).

- Starts the dev stack via DEV_PROCESS_MGMT and Docker Compose.

However, HTTPS is not yet wired in. Developers must still:

- Manually generate local TLS certificates.

- Manually configure Traefik to use those certificates.

- Manually trust the local CA in their system / browser.

This creates friction and makes the dev experience diverge from production style HTTPS routing. It also weakens the portfolio story for Stagecraft as an orchestrator that gives you a production like dev environment from a single command.

DEV_MKCERT fills this gap by providing a deterministic, spec governed way to provision and manage local TLS certificates for dev domains used by CLI_DEV.

⸻

## 2. Motivation

### Developer Experience

- One command: Developers should be able to run `stagecraft dev` and get HTTPS ready routes like `https://app.localdev.test` and `https://api.localdev.test` without manual certificate work.

- Production like routing: Dev environments should use HTTPS and host based routing similar to production.

- Fewer sharp edges: HTTPS should work reliably without repeated browser certificate warnings once the local CA is trusted.

### Operational Reliability

- Deterministic cert layout: Certificates should live in known, deterministic locations under `.stagecraft/dev`, so Traefik and Compose can mount them consistently.

- Idempotent operations: Running `stagecraft dev` multiple times must reuse existing certificates when valid, and only regenerate when strictly necessary.

- Clear behavior when mkcert is missing: The feature must fail with clear error messages if mkcert is required but not installed, and respect `--no-https` to avoid surprises.

### CI and Automation

- Skippable in CI: CI environments often do not need HTTPS for internal test traffic. DEV_MKCERT must be disabled when `--no-https` is set so pipelines can run without mkcert installed.

- Deterministic behavior: When enabled in a controlled environment, DEV_MKCERT must always produce the same certificate paths and structure for given inputs.

⸻

## 3. Users and User Stories

### Developers

- As a developer, I want `stagecraft dev` to provision local HTTPS certificates for my dev domains so I can use `https://` URLs that mirror production.

- As a developer, I want to disable HTTPS with `--no-https` when I only need HTTP or when mkcert is not available.

- As a developer, I want certificate generation to be idempotent so repeated runs do not thrash my system trust store or cause unnecessary changes.

### Platform Engineers

- As a platform engineer, I want certificate files to live in a deterministic directory under `.stagecraft/dev`, so I can review them and configure Traefik and Compose volumes consistently.

- As a platform engineer, I want DEV_MKCERT behavior to be driven by config and CLI flags, not hidden heuristics.

### CI / Automation

- As a CI pipeline, I want to run `stagecraft dev --no-https` without requiring mkcert, so tests can run in environments where certificate tooling is not installed.

- As a CI pipeline, I want clear and deterministic failures when `--no-https` is not set but mkcert is missing.

⸻

## 4. Success Criteria (v1)

1. When HTTPS is enabled (`--no-https` not set), DEV_MKCERT:

   - Ensures a certificate directory exists under `.stagecraft/dev/certs`.

   - Ensures certificates exist for the dev domains used by CLI_DEV (for example `app.localdev.test`, `api.localdev.test`), either as explicit host certs or as a wildcard.

   - Returns a `CertConfig` with `CertDir` and `Domains` that can be consumed by CLI_DEV and DEV_TRAEFIK.

2. DEV_MKCERT is idempotent:

   - If valid certs already exist for the requested domains, it reuses them.

   - It does not regenerate certificates on every run unless something is missing.

3. When `--no-https` is set:

   - DEV_MKCERT is not invoked.

   - CLI_DEV and DEV_TRAEFIK configure HTTP only routing.

   - No mkcert related errors are raised.

4. When mkcert is required but not available:

   - DEV_MKCERT fails with a clear error message (for example `dev: mkcert not found in PATH; install mkcert or use --no-https`).

   - The error is surfaced through CLI_DEV with appropriate context.

5. All file outputs from DEV_MKCERT are deterministic:

   - Certificate directory path is stable and documented.

   - File names and structure are deterministic for given domains.

6. DEV_MKCERT is implemented in a dedicated package under `internal/dev/mkcert` and is fully testable with exec and filesystem abstractions.

⸻

## 5. Risks and Constraints

### Determinism Constraints

- File paths for certificates must be deterministic and anchored under `.stagecraft/dev/certs`.

- DEV_MKCERT must not rely on non deterministic mkcert behaviors beyond its documented output. Where mkcert output is variable, DEV_MKCERT must normalize or wrap it in deterministic behavior (for example stable file naming).

- DEV_MKCERT must not introduce random identifiers or timestamps into the paths or configs that Stagecraft consumes.

### External Dependencies

- DEV_MKCERT depends on the `mkcert` binary being available when HTTPS is enabled.

- DEV_MKCERT must not silently install mkcert or modify system trust stores. Those operations are user responsibilities, and Stagecraft can only guide with clear messaging.

- DEV_MKCERT interacts with the local filesystem and external command execution; tests must use fakes to avoid side effects.

### Implementation Constraints

- v1 scope is limited to:

  - Single host dev environments.

  - A fixed set of dev domains (frontend and backend) for CLI_DEV.

- DEV_MKCERT must not modify Traefik configs directly. It only returns certificate paths and domains that DEV_TRAEFIK uses to build TLS configuration.

- DEV_MKCERT must respect `--no-https` flag and never invoke mkcert when HTTPS is disabled.

⸻

## 6. Alternatives Considered

### Alternative 1 - Do not manage certificates

Rejected because it creates a gap between the value proposition of `stagecraft dev` and what users actually experience. Developers would still need to manually generate and wire certificates, reducing Stagecraft's value.

### Alternative 2 - Embed a Go based self signed CA

Rejected for v1 because:

- mkcert is a widely used, purpose built tool for local certificates.

- Embedding a CA implementation adds significant complexity and security risk.

- Using mkcert is closer to user expectations and better aligned with existing ecosystem practices.

### Alternative 3 - Let Traefik terminate HTTP only and rely on browser HTTP

Rejected because the goal is explicit production like HTTPS routing for dev. While HTTP only is supported via `--no-https`, HTTPS should be the default where possible.

⸻

## 7. Dependencies

Required upstream work:

- CLI_DEV:

  - Implementation outline and spec exist.

  - DEV_PROCESS_MGMT is done.

  - Dev topology and dev files slices are implemented.

- DEV_TRAEFIK:

  - Traefik config generation spec and outline exist and define how TLS cert paths are used.

DEV_MKCERT must avoid:

- Direct Traefik config modification.

- Direct Compose modification.

It only returns certificate information to CLI_DEV and DEV_TRAEFIK.

⸻

## 8. Approval

- Author: [To be filled]

- Reviewer: [To be filled]

- Date: [To be filled]

Once approved, the DEV_MKCERT Implementation Outline may begin.


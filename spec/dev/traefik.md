---
feature: DEV_TRAEFIK
version: v1
status: done
domain: dev
---

# DEV_TRAEFIK

Generate and manage Traefik routing configuration for local development.

## Overview

DEV_TRAEFIK is responsible for:

- Producing deterministic Traefik configuration for `stagecraft dev`.
- Mapping frontend and backend services to HTTP/HTTPS routes.
- Integrating with mkcert based local HTTPS when enabled.

DEV_TRAEFIK does not run Traefik itself; it only produces configuration and wiring instructions.

## Scope - v1

### Included

- Static and dynamic configuration for Traefik suitable for docker provider.
- Routing rules for:
  - Frontend dev domains (for example `app.localdev.test`)
  - Backend APIs (for example `api.localdev.test`)
- Optional HTTPS termination using mkcert generated certificates.

### Excluded (future)

- Non Traefik routers.
- Kubernetes ingress definitions.
- Advanced routing (canary, blue green, rate limiting).

## Inputs

- Stagecraft config via `pkg/config`.
- Provider information:
  - Frontend services, ports, and domains.
  - Backend services and ports.
- `mkcert.CertConfig` from DEV_MKCERT containing:
  - `Enabled`: Whether HTTPS is enabled
  - `CertFile`: Path to certificate file (e.g., `.stagecraft/dev/certs/dev-local.pem`)
  - `KeyFile`: Path to key file (e.g., `.stagecraft/dev/certs/dev-local-key.pem`)
  - `Domains`: List of domains covered by the certificate

## Outputs

- Traefik configuration files under `.stagecraft/dev/traefik/`:
  - `traefik-static.yaml`
  - `traefik-dynamic.yaml`
- All files must be deterministic in content ordering.

## Behaviour

- Generate static config covering:
  - Entry points (http, https)
  - Providers (docker)
- Generate dynamic config covering:
  - Routers for each frontend and backend
  - Services mapping to compose services and ports
  - TLS configuration when `CertConfig.Enabled` is true, referencing `CertConfig.CertFile` and `CertConfig.KeyFile`

DEV_TRAEFIK must not start Traefik processes itself. That responsibility belongs to DEV_PROCESS_MGMT.

## Determinism

DEV_TRAEFIK must produce deterministic YAML output for identical inputs. This ensures:

- **Stable key ordering**: All YAML maps use lexicographically sorted keys.
- **Router ordering**: Routers are sorted lexicographically by name (e.g., "backend" before "frontend").
- **Service ordering**: Services are sorted lexicographically by name.
- **Middleware ordering**: Middlewares are sorted lexicographically by name (when used).
- **Entry point ordering**: Entry points are sorted lexicographically by name (e.g., "web" before "websecure").
- **Provider ordering**: Providers are sorted lexicographically by name.
- **Server ordering**: Within a service's load balancer, servers are sorted lexicographically by URL.
- **Entry point list ordering**: Within each router's entry points list, entry points are sorted lexicographically.

Implementation details:
- The generator uses `sortHTTPConfig()` and `sortEntryPoints()` to enforce ordering before YAML serialization.
- Certificate paths in TLS configuration use container-relative paths (`/certs/dev-local.pem`, `/certs/dev-local-key.pem`) rather than host-specific absolute paths.
- No timestamps, random identifiers, or host-specific paths are embedded in generated configs.

## Tests

- Unit tests in `internal/dev/traefik_*_test.go` (exact path to be finalised).
- Golden tests for:
  - `traefik-static.yaml`
  - `traefik-dynamic.yaml`

under `internal/dev/testdata/traefik_*`.


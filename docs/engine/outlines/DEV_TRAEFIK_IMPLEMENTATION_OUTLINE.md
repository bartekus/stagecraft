# DEV_TRAEFIK Implementation Outline

> This document defines the v1 implementation plan for DEV_TRAEFIK. It translates the feature spec into a concrete, testable delivery plan.

> All details in this outline must be reflected in `spec/dev/traefik.md` before any tests or code are written.

⸻

## 1. Feature Summary

**Feature ID:** DEV_TRAEFIK  
**Domain:** dev

**Goal:**

Generate and manage Traefik routing configuration for local development. Produce deterministic Traefik static and dynamic configuration files that map frontend and backend services to HTTP/HTTPS routes, with TLS wiring for mkcert-generated certificates.

**v1 Scope:**

- Static configuration for Traefik (entry points, providers)
- Dynamic configuration for Traefik (routers, services, middlewares)
- Routing rules for:
  - Frontend dev domains (e.g., `app.localdev.test`)
  - Backend APIs (e.g., `api.localdev.test`)
- **TLS wiring**: Consume `mkcert.CertConfig` and wire certificate paths into Traefik TLS configuration
- Generate config files under `.stagecraft/dev/traefik/`
- Deterministic file content ordering

**Out of scope for v1:**

- Non-Traefik routers (nginx, Caddy, etc.)
- Kubernetes ingress definitions
- Advanced routing (canary, blue-green, rate limiting)
- Traefik dashboard configuration
- Let's Encrypt integration (mkcert only for v1)

**Future extensions (not implemented in v1):**

- Non-Traefik routers
- Kubernetes ingress
- Advanced routing features
- Traefik dashboard access
- Let's Encrypt for production-like dev

⸻

## 2. Problem Definition and Motivation

CLI_DEV needs to route traffic from dev domains (e.g., `app.localdev.test`, `api.localdev.test`) to the correct services (frontend, backend). Traefik provides this routing capability, but requires configuration. DEV_TRAEFIK generates this configuration deterministically from Stagecraft config.

**The critical missing piece**: DEV_MKCERT generates certificates (`dev-local.pem`, `dev-local-key.pem`), but DEV_TRAEFIK does not yet wire them into Traefik TLS configuration. This means HTTPS routes do not work even though certificates exist.

DEV_TRAEFIK TLS wiring completes the HTTPS dev experience by:
- **Consuming CertConfig**: Taking `mkcert.CertConfig` as input and using `CertFile` and `KeyFile` paths
- **Wiring TLS**: Generating Traefik routers with TLS configuration that references the certificate files
- **Completing the stack**: Enabling `https://app.localdev.test` and `https://api.localdev.test` to work end-to-end

This matters because:
- **Local production parity**: Developers can test routing and HTTPS locally
- **Determinism**: Same config produces identical Traefik configs
- **Provider integration**: Frontend and backend providers don't need to know about Traefik
- **Complete HTTPS flow**: Certificates → Traefik config → routing → services all work together

⸻

## 3. User Stories (v1)

### Developer

- (Indirect) As a developer, I want to access my frontend at `https://app.localdev.test` and backend at `https://api.localdev.test`
- (Indirect) As a developer, I want Traefik to route requests correctly without manual configuration

### Platform Engineer

- As a platform engineer, I want to review the generated Traefik config to understand routing rules
- As a platform engineer, I want Traefik config generation to be deterministic so I can version control it

### CLI_DEV (consumer)

- As CLI_DEV, I want to generate Traefik configs that route frontend and backend services correctly
- As CLI_DEV, I want Traefik configs to integrate with mkcert certificates when HTTPS is enabled

⸻

## 4. Inputs and API Contract

### 4.1 API Surface (v1)

```go
// internal/dev/traefik/generator.go (new file)

package traefik

import (
    "stagecraft/pkg/config"
)

// Generator generates Traefik configuration for dev environments.
type Generator struct {
    // Configuration
}

// NewGenerator creates a new Traefik config generator.
func NewGenerator() *Generator

// GenerateConfig generates Traefik static and dynamic configuration.
//
// certCfg is the certificate configuration from DEV_MKCERT. When certCfg.Enabled
// is true, TLS configuration will reference certCfg.CertFile and certCfg.KeyFile.
func (g *Generator) GenerateConfig(
    cfg *config.Config,
    frontendDomain string,
    frontendService string,
    frontendPort string,
    backendDomain string,
    backendService string,
    backendPort string,
    certCfg *mkcert.CertConfig,
) (*Config, error)

// Config represents generated Traefik configuration.
type Config struct {
    Static  *StaticConfig
    Dynamic *DynamicConfig
}

// StaticConfig represents Traefik static configuration.
type StaticConfig struct {
    EntryPoints map[string]EntryPointConfig
    Providers   map[string]ProviderConfig
}

// DynamicConfig represents Traefik dynamic configuration.
type DynamicConfig struct {
    HTTP *HTTPConfig
}

// HTTPConfig contains HTTP routers, services, and middlewares.
type HTTPConfig struct {
    Routers     map[string]RouterConfig
    Services    map[string]ServiceConfig
    Middlewares map[string]MiddlewareConfig
}

// RouterConfig represents a Traefik router.
type RouterConfig struct {
    Rule    string
    Service string
    TLS     *TLSConfig
    EntryPoints []string
}

// ServiceConfig represents a Traefik service.
type ServiceConfig struct {
    LoadBalancer *LoadBalancerConfig
}

// LoadBalancerConfig represents load balancer configuration.
type LoadBalancerConfig struct {
    Servers []ServerConfig
}

// ServerConfig represents a backend server.
type ServerConfig struct {
    URL string
}

// TLSConfig represents TLS configuration.
type TLSConfig struct {
    CertFile string
    KeyFile  string
}
```

### 4.2 Input Sources

- **Stagecraft config**: `pkg/config.Config` - provides project-level settings
- **Frontend service info**: Domain, service name, port (from CLI_DEV)
- **Backend service info**: Domain, service name, port (from CLI_DEV)
- **Certificate config**: `mkcert.CertConfig` from DEV_MKCERT containing:
  - `Enabled`: Whether HTTPS is enabled
  - `CertFile`: Path to certificate file (e.g., `.stagecraft/dev/certs/dev-local.pem`)
  - `KeyFile`: Path to key file (e.g., `.stagecraft/dev/certs/dev-local-key.pem`)
  - `Domains`: List of domains covered by the certificate

### 4.3 Output

- **Static config file**: `.stagecraft/dev/traefik/traefik-static.yaml`
- **Dynamic config file**: `.stagecraft/dev/traefik/traefik-dynamic.yaml`
- **In-memory config**: `*Config` struct for programmatic access

⸻

## 5. TLS Wiring Behavior

### 5.1 Certificate Path Resolution

When `certCfg.Enabled` is true:

1. DEV_TRAEFIK uses `certCfg.CertFile` and `certCfg.KeyFile` directly from `CertConfig`.

2. Certificate paths in Traefik config must be container-relative paths:
   - Source: `.stagecraft/dev/certs/dev-local.pem` (on host)
   - Mount: `.stagecraft/dev/certs/` → `/certs/` (in Traefik container)
   - Config reference: `/certs/dev-local.pem` (in Traefik YAML)

3. Both routers (frontend and backend) use the same certificate pair since DEV_MKCERT v1 generates a single cert covering all domains.

### 5.2 TLS Configuration Generation

- When `certCfg.Enabled = true`:
  - Each router includes a `tls` section with `certFile` and `keyFile` set to the container-relative paths.
  - Routers are configured for both `web` (HTTP) and `websecure` (HTTPS) entry points.

- When `certCfg.Enabled = false`:
  - Routers have no `tls` section.
  - Routers are still configured for both entry points, but HTTPS will not work (Traefik will serve HTTP on both).

### 5.3 Error Handling

- If `certCfg` is nil when HTTPS is expected, DEV_TRAEFIK should return an error.
- If `certCfg.Enabled = false` but HTTPS is required (future validation), DEV_TRAEFIK should return an error.
- DEV_TRAEFIK does not validate that certificate files exist; Traefik validates at runtime.

⸻

## 6. Data Structures

### 6.1 Traefik Static Config

```yaml
# traefik-static.yaml
entryPoints:
  web:
    address: ":80"
  websecure:
    address: ":443"

providers:
  docker:
    endpoint: "unix:///var/run/docker.sock"
    exposedByDefault: false
    network: "stagecraft-dev"
```

### 6.2 Traefik Dynamic Config

```yaml
# traefik-dynamic.yaml (HTTPS enabled)
http:
  routers:
    backend:
      rule: "Host(`api.localdev.test`)"
      service: "backend"
      entryPoints:
        - "web"
        - "websecure"
      tls:
        certFile: "/certs/dev-local.pem"
        keyFile: "/certs/dev-local-key.pem"

    frontend:
      rule: "Host(`app.localdev.test`)"
      service: "frontend"
      entryPoints:
        - "web"
        - "websecure"
      tls:
        certFile: "/certs/dev-local.pem"
        keyFile: "/certs/dev-local-key.pem"

  services:
    frontend:
      loadBalancer:
        servers:
          - url: "http://frontend:3000"

    backend:
      loadBalancer:
        servers:
          - url: "http://backend:4000"

  middlewares:
    # Future: add middlewares if needed
```

### 6.3 Go Types

```go
// internal/dev/traefik/types.go (new file)

// StaticConfig represents Traefik static configuration.
type StaticConfig struct {
    EntryPoints map[string]EntryPointConfig
    Providers   map[string]ProviderConfig
}

// EntryPointConfig represents an entry point.
type EntryPointConfig struct {
    Address string
}

// ProviderConfig represents a provider configuration.
type ProviderConfig struct {
    Docker *DockerProviderConfig
}

// DockerProviderConfig represents Docker provider configuration.
type DockerProviderConfig struct {
    Endpoint        string
    ExposedByDefault bool
    Network         string
}

// DynamicConfig represents Traefik dynamic configuration.
type DynamicConfig struct {
    HTTP *HTTPConfig
}

// HTTPConfig contains HTTP configuration.
type HTTPConfig struct {
    Routers     map[string]RouterConfig
    Services    map[string]ServiceConfig
    Middlewares map[string]MiddlewareConfig
}

// RouterConfig represents a router.
type RouterConfig struct {
    Rule        string
    Service     string
    TLS         *TLSConfig
    EntryPoints []string
}

// ServiceConfig represents a service.
type ServiceConfig struct {
    LoadBalancer *LoadBalancerConfig
}

// LoadBalancerConfig represents load balancer configuration.
type LoadBalancerConfig struct {
    Servers []ServerConfig
}

// ServerConfig represents a server.
type ServerConfig struct {
    URL string
}

// TLSConfig represents TLS configuration.
type TLSConfig struct {
    CertFile string
    KeyFile  string
}
```

⸻

## 7. Determinism and Side Effects

### 7.1 Determinism Rules

- **Router ordering**: Routers sorted lexicographically by name
- **Service ordering**: Services sorted lexicographically by name
- **Middleware ordering**: Middlewares sorted lexicographically by name
- **Server ordering**: Servers sorted lexicographically by URL
- **Entry point ordering**: Entry points sorted lexicographically by name
- **Provider ordering**: Providers sorted lexicographically by name
- **YAML key ordering**: All YAML maps use stable key ordering

### 7.2 Side Effect Constraints

- **File writes**: DEV_TRAEFIK writes config files to `.stagecraft/dev/traefik/`
- **No Traefik execution**: DEV_TRAEFIK does not start or manage Traefik processes
- **No Docker interaction**: DEV_TRAEFIK does not talk to Docker directly
- **No network I/O**: Pure computation from inputs (except file writes)

⸻

## 8. Provider Boundaries

### Frontend Provider

- Frontend provider (via CLI_DEV) provides domain, service name, and port
- DEV_TRAEFIK does not interpret frontend provider config
- Frontend provider owns the service definition; DEV_TRAEFIK only generates routing rules

### Backend Provider

- Backend provider (via CLI_DEV) provides domain, service name, and port
- DEV_TRAEFIK does not interpret backend provider config
- Backend provider owns the service definition; DEV_TRAEFIK only generates routing rules

### DEV_MKCERT

- DEV_MKCERT (via CLI_DEV) provides `mkcert.CertConfig` containing:
  - `Enabled`: Whether HTTPS is enabled
  - `CertFile`: Absolute or relative path to certificate file (e.g., `.stagecraft/dev/certs/dev-local.pem`)
  - `KeyFile`: Absolute or relative path to key file (e.g., `.stagecraft/dev/certs/dev-local-key.pem`)
  - `Domains`: List of domains covered by the certificate
- DEV_TRAEFIK consumes `CertConfig` and wires `CertFile` and `KeyFile` into Traefik TLS configuration
- DEV_TRAEFIK does not generate certificates; that's DEV_MKCERT's responsibility
- DEV_TRAEFIK does not validate certificate files exist; Traefik validates at runtime

**Note**: DEV_TRAEFIK is a pure config generator. It does not:
- Start Traefik (that's DEV_PROCESS_MGMT or CLI_DEV's responsibility)
- Validate Traefik config (Traefik validates on startup)
- Monitor Traefik health (that's CLI_DEV's responsibility)

⸻

## 9. Testing Strategy

### Unit Tests

- **Config generation**: `internal/dev/traefik/generator_test.go`
  - Test static config generation
  - Test dynamic config generation
  - Test router rule generation
  - Test service URL generation
  - Test TLS config when HTTPS enabled
  - Test error cases (missing domains, invalid ports, etc.)

- **Deterministic ordering**: Test that all maps are sorted lexicographically
- **YAML serialization**: Test that YAML output is deterministic

### Golden Tests

- **Static config**: `internal/dev/traefik/testdata/traefik-static_*.yaml`
  - Golden files for static config with various settings
  - Verify deterministic YAML structure

- **Dynamic config**: `internal/dev/traefik/testdata/traefik-dynamic_*.yaml`
  - Golden files for dynamic config with various service combinations
  - Verify deterministic YAML structure
  - Verify lexicographic ordering

### Integration Tests

- **CLI_DEV integration**: Test that CLI_DEV can use DEV_TRAEFIK to generate configs
- **File generation**: Test that config files are written to correct location
- **HTTPS integration**: Test that TLS config is correct when mkcert certificates exist

⸻

## 10. Implementation Plan Checklist

### Before coding:

- [x] Spec exists (`spec/dev/traefik.md`)
- [ ] This outline approved
- [ ] Traefik version compatibility decided (v2.x vs v3.x)

### During implementation:

1. **Create package structure**
   - Create `internal/dev/traefik/` directory
   - Create `generator.go`, `types.go`, `generator_test.go`

2. **Implement types**
   - Define `StaticConfig`, `DynamicConfig`, `HTTPConfig`, etc.
   - Add YAML tags for correct serialization

3. **Implement Generator**
   - Implement `NewGenerator()`
   - Implement `GenerateConfig()` with router and service generation
   - Implement deterministic ordering logic

4. **Implement YAML serialization**
   - Use `gopkg.in/yaml.v3` for YAML encoding
   - Ensure deterministic key ordering
   - Add `ToYAML()` methods for static and dynamic configs

5. **Add file writing**
   - Implement file writing to `.stagecraft/dev/traefik/`
   - Create directory if it doesn't exist
   - Handle file write errors

6. **Add tests**
   - Unit tests for config generation
   - Golden tests for YAML output
   - Integration tests with CLI_DEV

### After implementation:

- [ ] Update docs if tests cause outline changes
- [ ] Ensure lifecycle completion in `spec/features.yaml`
- [ ] Run full test suite and verify no regressions

⸻

## 11. Completion Criteria

The feature is complete only when:

- [ ] Generator produces deterministic Traefik configs
- [ ] Static config includes entry points and Docker provider
- [ ] Dynamic config includes routers and services for frontend and backend
- [ ] HTTPS/TLS config works when `CertConfig.Enabled = true`:
  - TLS config references `CertConfig.CertFile` and `CertConfig.KeyFile`
  - Certificate paths use container-relative paths (`/certs/dev-local.pem`, `/certs/dev-local-key.pem`)
  - Both frontend and backend routers have TLS configured
- [ ] All ordering is lexicographic and stable
- [ ] Golden tests pass and YAML output is deterministic
- [ ] Unit tests cover config generation, router rules, service URLs
- [ ] Integration with CLI_DEV works
- [ ] Config files are written to correct location
- [ ] Spec and outline match actual behavior
- [ ] Feature status updated to `done` in `spec/features.yaml`

⸻

## 12. Implementation Notes

### Traefik Version

Traefik v2.x is the target version. Traefik v3.x may have different config format; v2.x is more widely used and stable.

### Router Rules

Router rules use Traefik's rule syntax:
- `Host(\`domain\`)` for host-based routing
- Future: `PathPrefix(\`/api\`)` for path-based routing

### Service URLs

Service URLs use Docker Compose service names:
- `http://frontend:3000` (service name from compose, port from provider)
- Assumes services are on the same Docker network (`stagecraft-dev`)

### TLS Configuration

When `certCfg.Enabled` is true:
- TLS config references mkcert certificate files via `certCfg.CertFile` and `certCfg.KeyFile`
- Certificate file names are deterministic: `dev-local.pem` (cert) and `dev-local-key.pem` (key)
- Certificate paths are relative to `.stagecraft/dev/certs/` (e.g., `/certs/dev-local.pem` when mounted in Traefik container)
- Traefik must have access to certificate files (bind mount in compose from `.stagecraft/dev/certs/` to `/certs/`)
- When `certCfg.Enabled` is false, no TLS configuration is generated (HTTP-only routers)

### File Paths

Config files are written to:
- `.stagecraft/dev/traefik/traefik-static.yaml`
- `.stagecraft/dev/traefik/traefik-dynamic.yaml`

These paths are relative to the project root (where `stagecraft.yml` is located).

### Docker Provider Configuration

Traefik's Docker provider is configured to:
- Use Docker socket: `unix:///var/run/docker.sock`
- Only expose services with labels (not all services)
- Use network: `stagecraft-dev`

This allows Traefik to discover services via Docker labels (future enhancement) or use file-based dynamic config (v1 approach).


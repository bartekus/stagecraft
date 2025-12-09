# CLI_DEV Implementation Outline

> This document defines the v1 implementation plan for CLI_DEV. It translates the feature analysis brief into a concrete, testable, spec aligned delivery plan.

> All details in this outline must be reflected in `spec/commands/dev.md` before any tests or code are written.

⸻

## 1. Feature Summary

**Feature ID:** CLI_DEV

**Domain:** commands

**Goal:**

Provide a complete `stagecraft dev` command that orchestrates backend services, frontend services, and supporting infrastructure (Traefik, mkcert, hosts) in a single deterministic operation.

**v1 Scope:**

- Single-host local development using Docker Compose
- Orchestration of:
  - Backend provider (generic or encore-ts) via existing provider interfaces
  - Frontend provider (PROVIDER_FRONTEND_GENERIC) via existing provider interfaces
  - Local Traefik reverse proxy (DEV_TRAEFIK)
  - Local HTTPS via mkcert certificates (DEV_MKCERT, may be stubbed)
  - Hosts file management for dev domains (DEV_HOSTS, may be stubbed)
- Deterministic process lifecycle:
  - Start all components in order (infra → backends → frontends)
  - Tear down on interruption (Ctrl+C) where possible
- Flag support: `--env`, `--config`, `--no-https`, `--no-hosts`, `--no-traefik`, `--detach`, `--verbose`

**Out of scope for v1:**

- Multi-host dev environments
- Remote dev containers
- Kubernetes based dev environments
- Hot reload integration beyond what providers already support
- Advanced process management (health checks, auto-restart, etc.)

**Future extensions (not implemented in v1):**

- Multi-host dev environments
- Kubernetes dev environments
- Remote dev containers (VS Code Remote, etc.)
- Advanced health checking and auto-restart
- Hot reload orchestration at CLI level

⸻

## 2. Problem Definition and Motivation

CLI_DEV_BASIC only handles backend services. Developers must manually start frontends, configure Traefik, set up HTTPS, and manage hosts files. CLI_DEV fills this gap by providing a single command that orchestrates the complete local development stack.

This matters because:
- **Developer productivity**: One command instead of multiple manual steps
- **Determinism**: Same config produces identical dev environments
- **Portfolio value**: Demonstrates full orchestration capabilities
- **CI integration**: Enables e2e testing with complete environments

⸻

## 3. User Stories (v1)

### Developer

- As a developer, I want to run `stagecraft dev` and have my entire stack (backend, frontend, Traefik, HTTPS) start automatically
- As a developer, I want `stagecraft dev` to clean up all processes when I press Ctrl+C
- As a developer, I want to disable HTTPS with `--no-https` if I don't need it
- As a developer, I want to see verbose output with `--verbose` to debug startup issues

### Platform Engineer

- As a platform engineer, I want `stagecraft dev` to generate deterministic compose files and Traefik configs
- As a platform engineer, I want the dev topology to be computed from config, not hardcoded

### CI / Automation

- As a CI pipeline, I want `stagecraft dev --detach` to start the full stack in the background for e2e tests

⸻

## 4. Inputs and CLI or API Contract

### 4.1 Command or API Surface (v1)

```
stagecraft dev [flags]
```

### 4.2 Flags or Arguments Implemented in v1

| Flag | Default | Description |
|------|---------|-------------|
| `--env` | `"dev"` | Environment name to use |
| `--config` | `""` | Path to config file (optional, defaults to standard search paths) |
| `--no-https` | `false` | Disable mkcert and HTTPS integration |
| `--no-hosts` | `false` | Do not modify /etc/hosts (user opts into manual DNS) |
| `--no-traefik` | `false` | Skip Traefik setup (providers must expose their own ports) |
| `--detach` | `false` | Run dev processes in background, return immediately |
| `--verbose` | `false` | Enable verbose output |

All flags must be documented in `internal/cli/commands/dev.go` help text and kept lexicographically sorted.

### 4.3 Flags or Arguments Reserved for Future Extensions

(Not implemented in v1)

| Flag | Planned Purpose |
|------|-----------------|
| `--host` | Future: target specific host in multi-host setup |
| `--watch` | Future: enable file watching and hot reload orchestration |
| `--profile` | Future: use named dev profile from config |

### 4.4 Exit Codes (v1)

| Code | Meaning |
|------|---------|
| 0 | Success - all components started successfully |
| 1 | Invalid user input or config |
| 2 | External provider failure (Docker, mkcert, etc.) |
| 3 | Internal error (unexpected panic, invariant violation) |

Note: Exact exit code mapping will be aligned with GOV_CLI_EXIT_CODES once that governance spec is finalized. For v1, we use the recommended structure.

⸻

## 5. Data Structures

### 5.1 Dev Topology

```go
// internal/dev/topology.go (new file)

// Topology represents the complete dev environment structure.
type Topology struct {
    Environment string
    Backend     *BackendService
    Frontend    *FrontendService
    Infrastructure *Infrastructure
}

// BackendService represents backend service configuration.
type BackendService struct {
    ProviderID string
    Config     any // Provider-specific config
    ComposeService *compose.Service // Generated compose service definition
}

// FrontendService represents frontend service configuration.
type FrontendService struct {
    ProviderID string
    Config     any // Provider-specific config
    ComposeService *compose.Service // Generated compose service definition
    Domain     string // Dev domain (e.g., "app.localdev.test")
}

// Infrastructure represents shared dev infrastructure.
type Infrastructure struct {
    Traefik *TraefikConfig
    Mkcert  *MkcertConfig
    Hosts   *HostsConfig
}

// TraefikConfig holds Traefik configuration.
type TraefikConfig struct {
    Enabled bool
    StaticConfigPath  string
    DynamicConfigPath string
    ComposeService    *compose.Service
}

// MkcertConfig holds mkcert certificate configuration.
type MkcertConfig struct {
    Enabled bool
    CertDir string
    Domains []string
}

// HostsConfig holds hosts file configuration.
type HostsConfig struct {
    Enabled bool
    Entries []HostEntry
}

// HostEntry represents a single hosts file entry.
type HostEntry struct {
    IP    string // Usually "127.0.0.1"
    Domain string
}
```

### 5.2 Compose Model (reused from CORE_COMPOSE)

```go
// internal/compose/compose.go (existing)

// ComposeFile represents a parsed Docker Compose file.
type ComposeFile struct {
    data map[string]any
    path string
}

// Service represents a Docker Compose service (to be added if not exists).
type Service struct {
    Name         string
    Image        string
    Build        map[string]any
    Ports        []string
    Volumes      []string
    Environment  map[string]string
    Networks     []string
    DependsOn    []string
    Labels       map[string]string
}
```

### 5.3 Traefik Configuration (generated by DEV_TRAEFIK)

```go
// internal/dev/traefik/config.go (new file)

// TraefikStaticConfig represents Traefik static configuration.
type TraefikStaticConfig struct {
    EntryPoints map[string]EntryPointConfig
    Providers   map[string]ProviderConfig
}

// TraefikDynamicConfig represents Traefik dynamic configuration.
type TraefikDynamicConfig struct {
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
```

⸻

## 6. Determinism and Side Effects

### 6.1 Determinism Rules

- **Topology computation**: Same config produces identical topology structure
- **Compose generation**: All services, networks, volumes sorted lexicographically
- **Traefik config**: Routers, services, middlewares sorted lexicographically
- **Process start order**: Always infra → backends → frontends (stable ordering)
- **No random identifiers**: All service names, network names, volume names come from config or use deterministic defaults
- **No timestamps**: Generated configs must not include timestamps

### 6.2 Side Effect Constraints

- **Config generation**: CLI_DEV generates configs (compose, Traefik) but does not execute Docker directly
- **Process execution**: In v1, we may use basic `exec.Command` for providers; future DEV_PROCESS_MGMT will handle this
- **File writes**: CLI_DEV writes config files to `.stagecraft/dev/` directory
- **Hosts file**: If `--no-hosts` is not set, CLI_DEV modifies `/etc/hosts` (requires appropriate permissions)
- **mkcert**: If `--no-https` is not set, CLI_DEV may invoke mkcert to generate certificates (external command)

⸻

## 7. Provider Boundaries (if applicable)

### Backend Provider

- CLI_DEV invokes `BackendProvider.Dev(ctx, opts)` for development mode
- CLI_DEV does NOT interpret backend provider config
- Backend provider returns compose service definitions via a new method (to be defined) or through metadata

### Frontend Provider

- CLI_DEV invokes `FrontendProvider.Dev(ctx, opts)` for development mode
- CLI_DEV does NOT interpret frontend provider config
- Frontend provider returns compose service definitions via a new method (to be defined) or through metadata

### Infrastructure Providers

- DEV_TRAEFIK generates config files; does not run Traefik
- DEV_MKCERT generates certificates; may invoke mkcert CLI
- DEV_HOSTS manages hosts file entries

**Note**: In v1, providers may need to expose compose service definitions. This may require extending provider interfaces or using a metadata pattern. Exact approach to be determined during implementation.

⸻

## 8. Testing Strategy

### Unit Tests

- **Topology computation**: `internal/dev/topology_test.go`
  - Test topology generation from config
  - Test deterministic ordering
  - Test error cases (missing providers, invalid config)

- **Flag parsing**: `internal/cli/commands/dev_test.go`
  - Test all flags
  - Test flag combinations
  - Test error cases (invalid flag values)

### Integration / CLI Tests

- **Command execution**: `internal/cli/commands/dev_test.go`
  - Test full command flow with mocked providers
  - Test deterministic output
  - Test error handling

### E2E Tests

- **Smoke test**: `test/e2e/dev_smoke_test.go`
  - Minimal fixture: backend + frontend
  - Run `stagecraft dev` and verify services are accessible
  - Verify compose file generation
  - Verify Traefik config generation

### Golden Tests

- **Compose output**: `internal/dev/testdata/dev_compose_*.yaml`
  - Golden files for generated compose YAML

- **Traefik config**: `internal/dev/testdata/traefik_*.yaml`
  - Golden files for static and dynamic Traefik configs

⸻

## 9. Implementation Plan Checklist

### Before coding:

- [x] Analysis brief approved (this document)
- [ ] This outline approved
- [x] Spec updated to match outline (`spec/commands/dev.md` exists)
- [ ] DEV_COMPOSE_INFRA implementation outline approved
- [ ] DEV_TRAEFIK implementation outline approved

### During implementation:

1. **Phase 1: DEV_COMPOSE_INFRA (thin slice)**
   - Create `internal/dev/compose/` package
   - Implement compose model generation for trivial app (backend only)
   - Add unit tests and golden files
   - Verify deterministic output

2. **Phase 2: DEV_TRAEFIK (thin slice)**
   - Create `internal/dev/traefik/` package
   - Implement Traefik config generation for trivial app
   - Add unit tests and golden files
   - Verify deterministic output

3. **Phase 3: CLI_DEV orchestration**
   - Extend `internal/cli/commands/dev.go` (or create new if CLI_DEV_BASIC is separate)
   - Implement topology computation
   - Wire DEV_COMPOSE_INFRA and DEV_TRAEFIK
   - Implement process lifecycle (basic exec for now)
   - Add flag parsing and validation
   - Add unit tests

4. **Phase 4: E2E smoke test**
   - Create minimal test fixture
   - Write `test/e2e/dev_smoke_test.go`
   - Verify end-to-end flow

### After implementation:

- [ ] Update docs if tests cause outline changes
- [ ] Ensure lifecycle completion in `spec/features.yaml`
- [ ] Run full test suite and verify no regressions
- [ ] Update CLI help text

⸻

## 10. Completion Criteria

The feature is complete only when:

- [ ] All v1 flags implemented and tested
- [ ] Topology computation is deterministic
- [ ] DEV_COMPOSE_INFRA generates deterministic compose files
- [ ] DEV_TRAEFIK generates deterministic Traefik configs
- [ ] CLI_DEV successfully orchestrates backend + frontend + Traefik
- [ ] E2E smoke test passes
- [ ] All tests pass
- [ ] Spec and outline match actual behavior
- [ ] Determinism guarantees enforced
- [ ] Feature status updated to `done` in `spec/features.yaml`

⸻

## 11. Implementation Order

1. **DEV_COMPOSE_INFRA** (foundation)
   - Enables compose model generation
   - Can be tested in isolation
   - Required by CLI_DEV

2. **DEV_TRAEFIK** (foundation)
   - Enables routing configuration
   - Can be tested in isolation
   - Required by CLI_DEV

3. **CLI_DEV** (orchestration)
   - Stitches DEV_COMPOSE_INFRA and DEV_TRAEFIK together
   - Coordinates providers
   - Handles process lifecycle

4. **E2E smoke test** (validation)
   - Validates end-to-end flow
   - Catches integration issues

This order ensures each component can be built and tested independently before integration.


# DEV_COMPOSE_INFRA Implementation Outline

> This document defines the v1 implementation plan for DEV_COMPOSE_INFRA. It translates the feature spec into a concrete, testable delivery plan.

> All details in this outline must be reflected in `spec/dev/compose-infra.md` before any tests or code are written.

⸻

## 1. Feature Summary

**Feature ID:** DEV_COMPOSE_INFRA

**Domain:** dev

**Goal:**

Generate and manage Docker Compose infrastructure for `stagecraft dev`. Synthesize compose models that combine backend, frontend, and shared infrastructure containers in a deterministic way.

**v1 Scope:**

- Generate a single deterministic compose model for:
  - Backend services (from backend provider) with real ports
  - Frontend services (from frontend provider) with real ports
  - Traefik service with:
    - Image: `traefik:v2.11` (or latest v2.x)
    - Ports: `80:80`, `443:443` (HTTP/HTTPS entrypoints)
    - Volumes:
      - `.stagecraft/dev/certs` → `/certs` (for mkcert certificates)
      - `.stagecraft/dev/traefik` → `/etc/traefik` (for Traefik config files)
    - Command: `--configfile=/etc/traefik/traefik-static.yaml --providers.file.directory=/etc/traefik`
    - Networks: `stagecraft-dev`
- Create shared network: `stagecraft-dev` (all services join this network)
- Enforce deterministic service ordering (lexicographic)
- Produce in-memory compose model (reuse types from CORE_COMPOSE)
- Generate compose file at `.stagecraft/dev/compose.yaml` (via DEV_FILES)

**Out of scope for v1:**

- Multi compose file layering (override stacks)
- Non Docker runtimes (Podman, containerd)
- Multi host compose swarms
- Compose file validation beyond basic structure

**Future extensions (not implemented in v1):**

- Multi compose file layering
- Podman/containerd support
- Compose swarm mode
- Advanced volume management (named volumes, external volumes)

⸻

## 2. Problem Definition and Motivation

CLI_DEV needs to generate Docker Compose configurations that combine services from multiple providers (backend, frontend) and infrastructure components (Traefik). DEV_COMPOSE_INFRA provides this capability by:

- Merging service definitions from multiple sources
- Enforcing deterministic naming and ordering
- Producing stable compose models that can be serialized to YAML

**The critical missing piece**: While DEV_TRAEFIK generates routing configs and DEV_MKCERT generates certificates, there's no compose file that:
- Wires Traefik with the correct mounts for certs and config
- Creates the shared network (`stagecraft-dev`) that all services need
- Ensures all services (backend, frontend, Traefik) are on the same network
- Uses real ports from providers instead of stubs

DEV_COMPOSE_INFRA v1 completes the dev stack by generating a compose file that:
- Defines Traefik service with mounts for `.stagecraft/dev/certs` → `/certs` and `.stagecraft/dev/traefik` → `/etc/traefik`
- Creates `stagecraft-dev` network that all services join
- Uses real container ports from backend/frontend providers
- Ensures deterministic ordering so compose files are stable and reviewable

This matters because:
- **Determinism**: Same inputs produce identical compose files
- **Provider integration**: Allows providers to contribute service definitions without CLI_DEV needing provider-specific knowledge
- **Testability**: Compose models can be generated and tested in isolation
- **Complete stack**: Makes `stagecraft dev` actually usable with real services and routing

⸻

## 3. User Stories (v1)

### Developer

- (Indirect) As a developer, I want `stagecraft dev` to generate compose files that work consistently across machines

### Platform Engineer

- As a platform engineer, I want to review the generated compose file to understand what services are running
- As a platform engineer, I want compose file generation to be deterministic so I can version control it

### CLI_DEV (consumer)

- As CLI_DEV, I want to merge backend, frontend, and Traefik services into a single compose model
- As CLI_DEV, I want deterministic service ordering so compose files are stable

⸻

## 4. Inputs and API Contract

### 4.1 API Surface (v1)

```go
// internal/dev/compose/generator.go (new file)

package compose

import (
    "stagecraft/pkg/config"
    "stagecraft/internal/compose"
)

// Generator generates compose models for dev environments.
type Generator struct {
    // Configuration
}

// NewGenerator creates a new compose generator.
func NewGenerator() *Generator

// GenerateCompose generates a compose model from config and service definitions.
func (g *Generator) GenerateCompose(
    cfg *config.Config,
    backendService *ServiceDefinition,
    frontendService *ServiceDefinition,
    traefikService *ServiceDefinition,
) (*compose.ComposeFile, error)

// ServiceDefinition represents a service to include in the compose model.
type ServiceDefinition struct {
    Name         string
    Image        string
    Build        map[string]any
    Ports        []PortMapping
    Volumes      []VolumeMapping
    Environment  map[string]string
    Networks     []string
    DependsOn    []string
    Labels       map[string]string
}

// PortMapping represents a port mapping.
type PortMapping struct {
    Host      string // e.g., "8080"
    Container string // e.g., "3000"
    Protocol  string // "tcp" or "udp"
}

// VolumeMapping represents a volume mapping.
type VolumeMapping struct {
    Type        string // "bind", "volume", "tmpfs"
    Source      string // Host path or volume name
    Target      string // Container path
    ReadOnly    bool
}
```

### 4.2 Input Sources

- **Stagecraft config**: `pkg/config.Config` - provides project-level settings
- **Backend service definition**: From backend provider (via CLI_DEV)
- **Frontend service definition**: From frontend provider (via CLI_DEV)
- **Traefik service definition**: From DEV_TRAEFIK

### 4.3 Output

- **In-memory compose model**: `*compose.ComposeFile` (reused from CORE_COMPOSE)
- **Optional on-disk file**: `.stagecraft/dev/compose.yaml` (when explicitly requested)

⸻

## 5. Data Structures

### 5.1 Service Definition (internal)

```go
// internal/dev/compose/types.go (new file)

// ServiceDefinition represents a service to include in compose.
type ServiceDefinition struct {
    Name         string
    Image        string
    Build        map[string]any
    Ports        []PortMapping
    Volumes      []VolumeMapping
    Environment  map[string]string
    Networks     []string
    DependsOn    []string
    Labels       map[string]string
}

// PortMapping represents a port mapping.
type PortMapping struct {
    Host      string
    Container string
    Protocol  string // "tcp" or "udp", default "tcp"
}

// VolumeMapping represents a volume mapping.
type VolumeMapping struct {
    Type     string // "bind", "volume", "tmpfs"
    Source   string
    Target   string
    ReadOnly bool
}
```

### 5.2 Compose Model (reused from CORE_COMPOSE)

```go
// internal/compose/compose.go (existing, may need extension)

// ComposeFile represents a parsed Docker Compose file.
type ComposeFile struct {
    data map[string]any
    path string
}

// Methods to be added or extended:
// - AddService(name string, service map[string]any) error
// - AddNetwork(name string, network map[string]any) error
// - AddVolume(name string, volume map[string]any) error
// - ToYAML() ([]byte, error) - with deterministic ordering
```

### 5.3 Generated Compose Structure

The generated compose file will have this structure (YAML representation):

```yaml
version: "3.8"

services:
  backend:      # From backend provider
    image: <from provider>
    ports:
      - "<host>:<container>/tcp"
    networks:
      - stagecraft-dev
    # ... other fields from provider

  frontend:     # From frontend provider
    image: <from provider>
    ports:
      - "<host>:<container>/tcp"
    networks:
      - stagecraft-dev
    # ... other fields from provider

  traefik:      # Generated by DEV_COMPOSE_INFRA
    image: traefik:v2.11
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./.stagecraft/dev/certs:/certs:ro
      - ./.stagecraft/dev/traefik:/etc/traefik:ro
    command:
      - --configfile=/etc/traefik/traefik-static.yaml
      - --providers.file.directory=/etc/traefik
      - --providers.file.watch=true
    networks:
      - stagecraft-dev

networks:
  stagecraft-dev:
    name: stagecraft-dev
```

**Key v1 details:**

- **Traefik service**: Always included when `traefikService != nil`
  - Image: `traefik:v2.11` (deterministic version)
  - Ports: `80:80` and `443:443` (HTTP/HTTPS entrypoints)
  - Volumes: Bind mounts for certs and config (relative to project root)
  - Command: File provider pointing to mounted config directory
  - Network: Always on `stagecraft-dev`

- **Network**: `stagecraft-dev` network is always created
  - All services (backend, frontend, Traefik) join this network
  - Network name is deterministic and documented

- **Service ordering**: Services sorted lexicographically (`backend`, `frontend`, `traefik`)

⸻

## 6. Traefik Service and Network Generation

### 6.1 Traefik Service Generation

When `traefikService != nil`, DEV_COMPOSE_INFRA generates a complete Traefik service definition:

**Image and Ports:**
- Image: `traefik:v2.11` (deterministic version for v1)
- Ports:
  - `80:80` (HTTP entrypoint)
  - `443:443` (HTTPS entrypoint)

**Volume Mounts:**
- `.stagecraft/dev/certs:/certs:ro` (read-only, for mkcert certificates)
- `.stagecraft/dev/traefik:/etc/traefik:ro` (read-only, for Traefik static/dynamic config)

**Command:**
```yaml
command:
  - --configfile=/etc/traefik/traefik-static.yaml
  - --providers.file.directory=/etc/traefik
  - --providers.file.watch=true
```

This tells Traefik to:
- Load static config from `/etc/traefik/traefik-static.yaml`
- Watch `/etc/traefik` directory for dynamic config changes
- Use file provider for dynamic configuration

**Network:**
- Always joins `stagecraft-dev` network

**v1 behavior:**
- If `traefikService` provides custom fields (e.g., image, ports), they are ignored for v1
- DEV_COMPOSE_INFRA owns the Traefik service definition structure
- Future versions may allow overrides via `traefikService` parameter

### 6.2 Network Creation

DEV_COMPOSE_INFRA always creates a `stagecraft-dev` network:

```yaml
networks:
  stagecraft-dev:
    name: stagecraft-dev
```

**Network assignment:**
- All services (backend, frontend, Traefik) must join `stagecraft-dev`
- If a service's `Networks` field is empty or doesn't include `stagecraft-dev`, DEV_COMPOSE_INFRA adds it
- If a service already has networks, `stagecraft-dev` is added to the list (deduplicated)

**Determinism:**
- Network name is fixed: `stagecraft-dev`
- Network configuration is minimal (bridge driver, default settings)
- Network appears in compose file even if no services explicitly reference it (for clarity)

### 6.3 Service Ordering

Services in the compose file must be ordered lexicographically:

1. `backend` (if present)
2. `frontend` (if present)
3. `traefik` (if present)

This ensures deterministic YAML output regardless of the order services are added to the generator.

⸻

## 7. Determinism and Side Effects

### 6.1 Determinism Rules

- **Service ordering**: Services sorted lexicographically by name
- **Port mapping ordering**: Ports sorted by host port (numeric, then lexicographic)
- **Volume mapping ordering**: Volumes sorted by target path (lexicographic)
- **Environment variable ordering**: Environment variables sorted by key (lexicographic)
- **Network ordering**: Networks sorted lexicographically
- **Volume ordering**: Volumes sorted lexicographically
- **Label ordering**: Labels sorted by key (lexicographic)
- **Dependency ordering**: Dependencies sorted lexicographically

### 6.2 Side Effect Constraints

- **No Docker interaction**: DEV_COMPOSE_INFRA only generates models; does not execute Docker
- **File writes**: Optional file write to `.stagecraft/dev/compose.yaml` (explicit opt-in)
- **No network I/O**: Pure computation from inputs
- **No state reads**: Does not read existing compose files or Docker state

⸻

## 8. Provider Boundaries

### Backend Provider

- Backend provider (via CLI_DEV) provides `ServiceDefinition` for backend service
- DEV_COMPOSE_INFRA does not interpret backend provider config
- Backend provider owns the service definition structure

### Frontend Provider

- Frontend provider (via CLI_DEV) provides `ServiceDefinition` for frontend service
- DEV_COMPOSE_INFRA does not interpret frontend provider config
- Frontend provider owns the service definition structure

### DEV_TRAEFIK

- DEV_TRAEFIK generates Traefik config files (static/dynamic YAML)
- DEV_COMPOSE_INFRA generates the Traefik **service** definition in compose
- DEV_COMPOSE_INFRA is responsible for:
  - Traefik service image, ports, volumes, command
  - Mounting `.stagecraft/dev/certs` → `/certs` for mkcert certificates
  - Mounting `.stagecraft/dev/traefik` → `/etc/traefik` for Traefik config files
  - Traefik command-line flags pointing to mounted config
- DEV_TRAEFIK does not define the compose service; that's DEV_COMPOSE_INFRA's responsibility

**v1 Traefik service generation:**

When `traefikService != nil`, DEV_COMPOSE_INFRA generates a Traefik service with:
- Image: `traefik:v2.11` (hardcoded for v1, configurable in future)
- Ports: `80:80`, `443:443` (deterministic)
- Volumes:
  - `.stagecraft/dev/certs:/certs:ro` (read-only, for certificates)
  - `.stagecraft/dev/traefik:/etc/traefik:ro` (read-only, for config files)
- Command: `--configfile=/etc/traefik/traefik-static.yaml --providers.file.directory=/etc/traefik --providers.file.watch=true`
- Networks: `stagecraft-dev`

If `traefikService` provides additional fields (e.g., custom image, ports, volumes), they are merged with the defaults above, with defaults taking precedence for v1.

### Provider Service Definitions

**v1 approach**: CLI_DEV constructs `ServiceDefinition` from provider config/metadata. For v1:
- Backend provider: CLI_DEV extracts service name, image/build, ports from provider config
- Frontend provider: CLI_DEV extracts service name, image/build, ports from provider config
- Future: Providers may expose `ComposeService() (*ServiceDefinition, error)` method, but v1 uses CLI_DEV extraction

⸻

## 9. Testing Strategy

### Unit Tests

- **Service merging**: `internal/dev/compose/generator_test.go`
  - Test merging backend + frontend + Traefik services
  - Test deterministic service ordering (lexicographic)
  - Test error cases (duplicate service names, invalid ports, etc.)

- **Traefik service generation**:
  - Test that Traefik service is generated with correct image (`traefik:v2.11`)
  - Test that Traefik service has correct ports (`80:80`, `443:443`)
  - Test that Traefik service has correct volume mounts (certs and traefik config)
  - Test that Traefik service has correct command flags
  - Test that Traefik service joins `stagecraft-dev` network

- **Network creation**:
  - Test that `stagecraft-dev` network is always created
  - Test that all services (backend, frontend, Traefik) join `stagecraft-dev`
  - Test that services with existing networks still get `stagecraft-dev` added

- **Port mapping**: Test port mapping conversion to compose format
- **Volume mapping**: Test volume mapping conversion to compose format
- **Environment variables**: Test environment variable handling

### Golden Tests

- **Compose YAML output**: `internal/dev/compose/testdata/dev_compose_*.yaml`
  - `dev_compose_backend_frontend_traefik.yaml` - Full stack with all services
  - `dev_compose_backend_only.yaml` - Backend only (no frontend, no Traefik)
  - `dev_compose_backend_traefik.yaml` - Backend + Traefik (no frontend)
  - Verify deterministic YAML structure
  - Verify lexicographic service ordering
  - Verify Traefik service structure (image, ports, volumes, command, network)
  - Verify `stagecraft-dev` network is present
  - Verify all services join `stagecraft-dev` network

### Integration Tests

- **CLI_DEV integration**: Test that CLI_DEV can use DEV_COMPOSE_INFRA to generate compose files
- **CORE_COMPOSE compatibility**: Test that generated compose files can be loaded by CORE_COMPOSE loader

⸻

## 10. Implementation Plan Checklist

### Before coding:

- [x] Spec exists (`spec/dev/compose-infra.md`)
- [ ] This outline approved
- [ ] Provider interface extension strategy decided (how providers expose ServiceDefinition)

### During implementation:

1. **Create package structure**
   - Create `internal/dev/compose/` directory
   - Create `generator.go`, `types.go`, `generator_test.go`

2. **Implement ServiceDefinition types**
   - Define `ServiceDefinition`, `PortMapping`, `VolumeMapping`
   - Add validation methods

3. **Implement Generator**
   - Implement `NewGenerator()`
   - Implement `GenerateCompose()` with service merging
   - **Traefik service generation**:
     - When `traefikService != nil`, generate Traefik service with:
       - Image: `traefik:v2.11`
       - Ports: `80:80`, `443:443`
       - Volumes: certs and traefik config mounts
       - Command: file provider flags
       - Network: `stagecraft-dev`
   - **Network creation**: Always create `stagecraft-dev` network
   - **Network assignment**: Ensure all services join `stagecraft-dev` network
   - **Service ordering**: Sort services map lexicographically before adding to compose
   - Implement deterministic ordering logic for all fields

4. **Extend CORE_COMPOSE (if needed)**
   - Add methods to `ComposeFile` for adding services, networks, volumes
   - Add `ToYAML()` method with deterministic ordering
   - Or create wrapper that uses existing CORE_COMPOSE types

5. **Add tests**
   - Unit tests for service merging
   - Golden tests for YAML output
   - Integration tests with CLI_DEV

### After implementation:

- [ ] Update docs if tests cause outline changes
- [ ] Ensure lifecycle completion in `spec/features.yaml`
- [ ] Run full test suite and verify no regressions

⸻

## 11. Completion Criteria

The feature is complete only when:

- [ ] Generator produces deterministic compose models
- [ ] Service merging works for backend + frontend + Traefik
- [ ] **Traefik service is generated with correct mounts**:
  - `.stagecraft/dev/certs` → `/certs` mount exists
  - `.stagecraft/dev/traefik` → `/etc/traefik` mount exists
  - Traefik command points to mounted config files
- [ ] **Network creation**: `stagecraft-dev` network is always created
- [ ] **Network assignment**: All services (backend, frontend, Traefik) join `stagecraft-dev` network
- [ ] **Service ordering**: Services map is sorted lexicographically (`backend`, `frontend`, `traefik`)
- [ ] All ordering is lexicographic and stable
- [ ] Golden tests pass and YAML output is deterministic
- [ ] Unit tests cover service merging, port mapping, volume mapping, Traefik service generation, network creation
- [ ] Integration with CLI_DEV works
- [ ] Spec and outline match actual behavior
- [ ] Feature status updated to `done` in `spec/features.yaml`

⸻

## 12. Implementation Notes

### Reusing CORE_COMPOSE

CORE_COMPOSE provides `ComposeFile` type with `Load()` method. We need to:
- Either extend `ComposeFile` with `AddService()`, `AddNetwork()`, `AddVolume()` methods
- Or create a builder pattern that constructs `ComposeFile` from `ServiceDefinition`s
- Ensure `ToYAML()` method produces deterministic output

### Service Name Collisions

If backend and frontend providers both want to use the same service name, DEV_COMPOSE_INFRA should:
- Detect the collision
- Return an error with clear message
- Let CLI_DEV handle resolution (e.g., by prefixing with provider name)

### Network and Volume Naming

- **Default network**: `stagecraft-dev` (deterministic, always created)
- **Network configuration**: Simple bridge network with explicit name
- **Volume mounts**: Use relative paths from project root:
  - `.stagecraft/dev/certs` → `/certs` (Traefik)
  - `.stagecraft/dev/traefik` → `/etc/traefik` (Traefik)
- All names and paths must be stable across runs

### Traefik Service Generation

For v1, DEV_COMPOSE_INFRA generates the Traefik service definition directly:

- **Image**: `traefik:v2.11` (hardcoded for v1)
- **Ports**: `80:80`, `443:443` (HTTP/HTTPS entrypoints)
- **Volumes**:
  - `.stagecraft/dev/certs:/certs:ro` (mkcert certificates, read-only)
  - `.stagecraft/dev/traefik:/etc/traefik:ro` (Traefik config files, read-only)
- **Command**: 
  ```
  --configfile=/etc/traefik/traefik-static.yaml
  --providers.file.directory=/etc/traefik
  --providers.file.watch=true
  ```
- **Networks**: `stagecraft-dev`

The `traefikService` parameter is used to signal that Traefik should be included, but DEV_COMPOSE_INFRA owns the service definition structure. Future versions may allow `traefikService` to override defaults (e.g., custom image version).

### Port Conflicts

DEV_COMPOSE_INFRA does not validate port conflicts (that's Docker's job). However, it should:
- Preserve port mappings as provided by providers
- Format port mappings consistently (e.g., "8080:3000/tcp")


---
feature: CORE_COMPOSE
version: v1
status: done
domain: core
inputs:
  flags: []
outputs:
  exit_codes: {}
---
# Docker Compose Integration

- Feature ID: `CORE_COMPOSE`
- Status: done
- Depends on: `CORE_CONFIG`

## Goal

Provide Docker Compose file loading, parsing, and manipulation capabilities for Stagecraft. This enables environment-specific override generation, service filtering, and per-host configuration generation.

## Architecture

### Core Responsibilities

The Compose integration MUST:

1. **Load and Parse Compose Files**
   - Load `docker-compose.yml` from project root
   - Parse Compose file structure (services, volumes, networks)
   - Support multiple Compose files (base + overrides)

2. **Generate Environment-Specific Overrides**
   - Generate override files based on environment configuration
   - Inject environment variables
   - Resolve volume paths per environment
   - Apply port exposure policies

3. **Service Management**
   - Identify service roles (infra vs app services)
   - Filter services by environment configuration
   - Support service mode (container vs external)

4. **Per-Host Configuration**
   - Generate Compose configs filtered by host roles
   - Support multi-host deployments

## API Design

### Package: `internal/compose`

```go
package compose

import (
    "context"
    "io"
    
    "stagecraft/pkg/config"
)

// ComposeFile represents a parsed Docker Compose file.
type ComposeFile struct {
    // Internal representation of compose structure
}

// Loader loads and parses Docker Compose files.
type Loader struct {
    // Configuration for loading
}

// NewLoader creates a new Compose file loader.
func NewLoader() *Loader

// Load loads a Compose file from the given path.
func (l *Loader) Load(path string) (*ComposeFile, error)

// GenerateOverride generates an environment-specific override file.
func (c *ComposeFile) GenerateOverride(env string, cfg *config.Config) ([]byte, error)

// GetServices returns all service names from the Compose file.
func (c *ComposeFile) GetServices() []string

// FilterServices filters services by role or environment configuration.
func (c *ComposeFile) FilterServices(roles []string) []string

// GetServiceRoles returns the role mapping for services.
func (c *ComposeFile) GetServiceRoles(cfg *config.Config) map[string]string
```

## Behavior

### Compose File Loading

- Load `docker-compose.yml` from project root (default)
- Support custom paths via configuration
- Parse YAML structure into internal representation
- Validate basic Compose file structure
- Handle Compose file version differences gracefully

### Environment-Specific Overrides

The override generation MUST:

1. **Volume Path Resolution**
   - Resolve volume paths from environment config
   - Example: `postgres_volume: postgres_data` (dev) vs `/var/lib/platform/postgres` (prod)
   - Apply to all volume references in services

2. **Port Exposure Policy**
   - Apply port publishing rules from environment config
   - Example: `db_port_publish: "5433:5432"` (dev) vs `""` (prod)
   - Remove or add ports based on environment

3. **Environment Variables**
   - Inject environment-specific env vars
   - Merge with service-level environment config
   - Support variable interpolation

4. **Service Mode Configuration**
   - Handle `mode: external` vs `mode: container`
   - For external mode, remove service from Compose (runs outside)
   - For container mode, ensure service is included

### Service Filtering

- Filter services by role (gateway, db, cache, app)
- Support per-host role assignments
- Generate filtered Compose configs for specific hosts

### Error Handling

- Clear error messages for missing files
- Validation errors for malformed Compose files
- Helpful errors for missing environment config

## Implementation Details

### Compose File Structure

The implementation should work with standard Docker Compose v3+ format:

```yaml
version: "3.9"
services:
  service-name:
    image: ...
    volumes:
      - ${VOLUME_VAR}:/path
    ports:
      - "${PORT_PUBLISH:-}"
    environment:
      - KEY=${VALUE}
```

### Override File Format

Generated override files follow the same Compose format:

```yaml
version: "3.9"
services:
  db:
    volumes:
      - /var/lib/platform/postgres:/var/lib/postgresql/data
    ports: []
  api:
    # Service removed if mode: external
```

### Volume Path Resolution

Volume paths are resolved from environment config:

```yaml
environments:
  dev:
    postgres_volume: postgres_data  # Named volume
  prod:
    postgres_volume: /var/lib/platform/postgres  # Host path
```

The implementation MUST:
- Replace `${POSTGRES_VOLUME:-postgres_data}` with resolved value
- Handle both named volumes and host paths
- Preserve volume mount structure

### Port Publishing

Port publishing is controlled by environment config:

```yaml
environments:
  dev:
    db_port_publish: "5433:5432"  # Publish to host
  prod:
    db_port_publish: ""  # No host publishing
```

The implementation MUST:
- Replace `${DB_PORT_PUBLISH:-}` with resolved value
- Remove ports array if empty string
- Preserve port format when specified

## Non-Goals (v1)

- Full Compose file validation (basic structure only)
- Compose file watching/reloading
- Advanced Compose features (profiles, extends, etc.)
- Compose file generation from scratch
- Kubernetes manifest generation

## Testing

Tests MUST cover:

1. **File Loading**
   - Loading valid Compose files
   - Error handling for missing files
   - Error handling for malformed YAML

2. **Override Generation**
   - Volume path resolution (named vs host path)
   - Port publishing (with and without ports)
   - Environment variable injection
   - Service mode handling (external vs container)

3. **Service Filtering**
   - Filtering by role
   - Filtering by environment config
   - Per-host service lists

4. **Edge Cases**
   - Empty Compose files
   - Services without volumes/ports
   - Missing environment config values

## Related Features

- `CORE_CONFIG` - Configuration loading
- `CORE_ENV_RESOLUTION` - Environment resolution
- `DEV_COMPOSE_INFRA` - Dev infrastructure orchestration
- `DEPLOY_COMPOSE_GEN` - Per-host Compose generation for deployment

## Future Enhancements

- Support for Compose profiles
- Compose file validation against schema
- Support for Compose extends
- Multi-file Compose support (docker-compose.override.yml)
- Compose file diff visualization


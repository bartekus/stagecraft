# Core Config – Loading and Validation

- Feature ID: `CORE_CONFIG`
- Status: todo

## Goal

Provide a single, well-defined entrypoint for loading and validating Stagecraft configuration.

The config should be:

- **Human-friendly** (YAML, with sane defaults).
- **Machine-validated** (schema + semantic checks).
- **Easy to test** (pure functions where possible).

## Format (Full Schema)

Config file: `stagecraft.yml` (default in repo root).

Full structure (from `docs/stagecraft-spec.md`):

```yaml
project:
  name: platform
  registry: ghcr.io/your-org/platform

backend:
  provider: generic
  providers:
    generic:
      dev:
        command: ["npm", "run", "dev"]
        workdir: "./backend"
    encore-ts:
      dev:
        secrets:
          types: ["dev", "preview", "local"]
          from_env:
            - DOMAIN
            - API_DOMAIN

frontend:
  provider: generic-dev-command
  dev:
    workdir: apps/web
    command: ["npm", "run", "dev", "--", "--host", "0.0.0.0", "--port", "5173"]

network:
  provider: tailscale
  tailscale:
    auth_key_env: TS_AUTHKEY
    tailnet_domain: "mytailnet.ts.net"

hosts:
  gateway:
    - plat-gw-1
  app:
    - plat-api-1
    - plat-web-1
  db:
    - plat-db-1
  cache:
    - plat-cache-1

services:
  traefik:
    role: gateway
  db:
    role: db
  redis:
    role: cache
  logto:
    role: app
  # ... more services

databases:
  primary:
    migrations:
      engine: raw
      path: ./migrations
      strategy: pre_deploy
    connection_env: DATABASE_URL
  # Additional databases can be defined here
  # analytics:
  #   migrations:
  #     engine: drizzle
  #     path: ./drizzle/migrations
  #     strategy: post_deploy

environments:
  dev:
    env_file: .env.local
    postgres_volume: postgres_data
    postgres_init_scripts: ./scripts/db
    redis_volume: redis_data
    db_port_publish: "5433:5432"
    redis_port_publish: "6379:6379"
    traefik:
      mode: mkcert
      mkcert_cert_dir: ./.local-infra/certs
      dynamic_config_dir: ./.local-infra/traefik/dynamic
      dashboard_auth: basic
      hsts: false
      cors_for_logto: true
    api:
      mode: external
      url: http://localhost:4000
    web:
      mode: external
      url: http://localhost:5173

  prod:
    env_file: /etc/platform/env
    postgres_volume: /var/lib/platform/postgres
    postgres_init_scripts: /opt/platform/scripts/db
    redis_volume: /var/lib/platform/redis
    db_port_publish: ""
    redis_port_publish: ""
    traefik:
      mode: acme
      acme_email_env: ACME_EMAIL
      acme_storage: /var/lib/platform/traefik/acme.json
      dashboard_auth: none
      hsts: true
      cors_for_logto: false
    api:
      mode: container
    web:
      mode: container
```

## Behavior

### Default Path
- `DefaultConfigPath()` returns `"stagecraft.yml"`

### Existence Check
- `Exists(path string) (bool, error)`:
  - Returns `true, nil` if file exists and is regular
  - Returns `false, nil` if file does not exist
  - Returns `false, error` for other I/O errors

### Loading
- `Load(path string) (*Config, error)`:
  - Returns `ErrConfigNotFound` if file doesn't exist
  - Returns validation error if YAML is invalid or fails validation
  - Returns populated `Config` on success

### Validation (Full)

#### Project
- `project.name` must be non-empty
- `project.registry` must be valid registry URL (if present)

#### Backend
- `backend.provider` is required and must be a registered backend provider ID
- `backend.providers` is required and must be a map
- `backend.providers[backend.provider]` must exist
- Validation checks that the provider exists in the backend provider registry
- Stagecraft core does not validate provider-specific fields; validation is delegated to the provider

#### Frontend
- `frontend.provider` is required and must be a registered frontend provider ID (when implemented)
- `frontend.providers` is required and must be a map (when implemented)
- `frontend.providers[frontend.provider]` must exist (when implemented)
- Validation will check that the provider exists in the frontend provider registry (when implemented)
- **Note**: Frontend provider registry not yet implemented; will follow same pattern as backend providers

#### Network
- `network.provider` is required and must be a registered network provider ID (when implemented)
- Network provider-specific config will be namespaced under `network.providers.<id>` (when implemented)
- Validation will check that the provider exists in the network provider registry (when implemented)
- **Note**: Network provider registry not yet implemented; will follow same pattern as backend providers

#### Hosts
- At least one host role must be defined
- Host names must be non-empty strings

#### Services
- Service names must match docker-compose.yml services
- Service roles must match host roles

#### Databases (Migration Configuration)
- `databases` is optional (only needed if migrations are used)
- Each database must have:
  - `migrations.engine` is required and must be a registered migration engine ID
  - `migrations.path` must be a valid path (relative to project root)
  - `migrations.strategy` must be one of: `pre_deploy`, `post_deploy`, `manual` (if present)
- Validation checks that the engine exists in the migration engine registry
- `connection_env` must be a valid environment variable name
- Environment-specific overrides can override migration strategy per environment

#### Environments
- At least one environment must be defined
- Environment names must be non-empty
- Environment-specific configs must be valid

## Non-Goals (initial version)

- Remote config loading
- Full environment variable interpolation (v1) - Note: Basic `${VAR}` interpolation is supported for migration config values only
- Advanced schema evolution/migrations
- Config file watching/reloading

## Tests

See `spec/features.yaml` entry for `CORE_CONFIG`:
- `pkg/config/config_test.go` – unit tests for:
  - `DefaultConfigPath`
  - `Exists`
  - `Load` (config not found, invalid YAML, validation errors, happy path)
  - All validation rules


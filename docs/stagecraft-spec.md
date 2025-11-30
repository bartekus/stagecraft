# Stagecraft CLI - Application Specification

> **Related Documents:**
> - [`docs/implementation-roadmap.md`](implementation-roadmap.md) - Implementation phases and feature catalog
> - [`spec/features.yaml`](../spec/features.yaml) - Feature tracking (source of truth)
> - [`docs/adr/0001-architecture.md`](adr/0001-architecture.md) - Architecture decisions
> - [`blog/01-why-not-kamal.md`](../blog/01-why-not-kamal.md) - Design rationale

## 0. Purpose and goals

Stagecraft is a Go based orchestration CLI for local development and deployment of multi service applications.

It is designed around:
•	Local first DX - one command to spin up full local infra, HTTPS, backend, frontend.
•	Docker Compose + docker-rollout for runtime orchestration.
•	Tailscale or Headscale for multi host networking.
•	Provider model so Encore.ts, Vite, DO CLI, GitHub CLI and others plug in cleanly.
•	Configuration driven through a single stagecraft.yml plus one canonical docker-compose.yml.

Stagecraft is roughly in the same problem space as Kamal, but:
•	Uses Compose instead of one off docker run.
•	Supports multi host via mesh network (Tailscale) instead of static IPs only.
•	Has a first class local dev story (Encore dev server, Vite dev server, mkcert HTTPS).

⸻

## 1. High level concepts

### 1.1 Environments

Stagecraft works with named environments, at minimum:
•	dev - local developer machine.
•	staging - remote multi host environment.
•	prod - remote multi host environment.

Each environment defines:
•	Path to env file (env_file).
•	Volume roots and paths.
•	Port exposure policy.
•	Traefik TLS mode and security posture.
•	Whether app services (api, web, etc.) run as containers or external dev servers.
•	Network settings (Tailscale configuration).

### 1.2 Services

Services are the logical units defined in docker-compose.yml.

Example services:
•	Infra: traefik, db, redis, logto, pgweb, dozzle, electric (optional).
•	App: api, web, worker, etc.

Services have:
•	A role (gateway, app, db, cache, infra) for host placement.
•	Optional profile (for example app-prod so some services only run in deploy environments).

Stagecraft does not own service definitions; it reads them from Compose and uses config metadata to decide where and how to run them.

### 1.3 Hosts and roles

Stagecraft maps service roles to hosts.
•	A host is a machine identified by a Tailscale node name or reachable hostname.
•	Example:

```yaml
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
```

	•	Services declare their role so Stagecraft can determine which host(s) should run them.

### 1.4 Providers

Stagecraft uses a plugin like provider model.

Core provider types:
•	BackendProvider - how to run and build the backend (Encore.ts initially).
•	FrontendProvider - how to run local frontend dev (Vite or other).
•	NetworkProvider - mesh network; Tailscale or Headscale.
•	CloudProvider - infra bootstrap; DO CLI first, later AWS, GCP, bare metal.
•	CIProvider - CI integration; GitHub Actions via gh first.
•	SecretsProvider (optional abstraction) - storing and syncing secrets.

Each provider is configured in stagecraft.yml and has a Go interface.

⸻

## 2. Configuration: stagecraft.yml

Stagecraft uses a single project config file at repo root:

```yaml
project:
  name: platform
  registry: ghcr.io/your-org/platform

backend:
  provider: encore-ts
  dev:
    env_file: .env.local
    listen: "0.0.0.0:4000"
    disable_telemetry: true
    node_extra_ca_certs: "./.local-infra/certs/mkcert-rootCA.pem"
    encore_secrets:
      types: ["dev", "preview", "local"]
      from_env:
        - DOMAIN
        - API_DOMAIN
        - LOGTO_DOMAIN
        - WEB_DOMAIN
        - LOGTO_APP_ID
        - LOGTO_APP_SECRET
        - LOGTO_MANAGEMENT_API_APPLICATION_ID
        - LOGTO_MANAGEMENT_API_APPLICATION_SECRET
        - LOGTO_APP_API_EVENT_WEBHOOK_SIGNING_KEY
        - STRIPE_API_KEY
        - STRIPE_WEBHOOK_SECRET
        - STRIPE_SERVICE_API_KEY
        - STRIPE_API_VERSION

frontend:
  provider: generic-dev-command
  dev:
    workdir: apps/web
    command: ["npm", "run", "dev", "--", "--host", "0.0.0.0", "--port", "5173"]

network:
  provider: tailscale      # or "headscale"
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
  pgweb:
    role: app
  dozzle:
    role: app
  api:
    role: app
  web:
    role: app

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
      mode: external   # Encore dev server
      url: http://localhost:4000
    web:
      mode: external   # Vite dev server
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

This file is the single source of truth Stagecraft uses to decide:
•	How to run services in dev vs prod.
•	Where to send traffic from Traefik.
•	Which hosts run which containers.
•	How to talk to providers.

⸻

## 3. Canonical docker-compose.yml expectations

Stagecraft assumes a canonical docker-compose.yml at repo root defined by the user.

Design goals:
•	Service graph is unified across environments.
•	Differences are handled through env vars and override files generated by Stagecraft.

Example skeleton:

```yaml
version: "3.9"

x-logging: &default-logging
  driver: json-file
  options:
    max-size: "50m"
    max-file: "6"

networks:
  net:
    driver: bridge

volumes:
  postgres_data:
  redis_data:

services:
  traefik:
    image: traefik:v3.5
    restart: always
    logging: *default-logging
    networks: [net]
    ports:
      - "80:80"
      - "443:443"
      - "8080:8080"
    env_file:
      - ${PLATFORM_ENV_FILE:-.env.local}
    command: []          # Filled by override files or CLI
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro

  db:
    image: postgres:16
    restart: always
    logging: *default-logging
    networks: [net]
    env_file:
      - ${PLATFORM_ENV_FILE:-.env.local}
    environment:
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: ${POSTGRES_DB}
      POSTGRES_MULTIPLE_DATABASES: ${POSTGRES_MULTIPLE_DATABASES}
      DB_ADMIN_USER: ${DB_ADMIN_USER}
      DB_ADMIN_PASSWORD: ${DB_ADMIN_PASSWORD}
      DB_ADMIN_ACCESS: ${DB_ADMIN_ACCESS}
    volumes:
      - ${POSTGRES_VOLUME:-postgres_data}:/var/lib/postgresql/data
      - ${POSTGRES_INIT_SCRIPTS:-./scripts/db}:/docker-entrypoint-initdb.d:ro
    ports:
      - "${DB_PORT_PUBLISH:-}"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U $$POSTGRES_USER"]
      interval: 5s
      timeout: 5s
      retries: 10
      start_period: 20s

  redis:
    image: redis:7
    restart: always
    logging: *default-logging
    networks: [net]
    volumes:
      - ${REDIS_VOLUME:-redis_data}:/data
    ports:
      - "${REDIS_PORT_PUBLISH:-}"
    healthcheck:
      test: ["CMD-SHELL", "redis-cli ping | grep PONG"]
      interval: 5s
      timeout: 3s
      retries: 10
      start_period: 10s

  logto:
    image: svhd/logto:latest
    restart: always
    logging: *default-logging
    networks: [net]
    env_file:
      - ${PLATFORM_ENV_FILE:-.env.local}
    environment:
      DB_URL: ${LOGTO_DB_URL}
      APP_NAME: ${LOGTO_APP_NAME}
      APP_DESCRIPTION: ${LOGTO_APP_DESCRIPTION}
      LOGTO_APP_API_EVENT_WEBHOOK_URL: ${LOGTO_APP_API_EVENT_WEBHOOK_URL}
      NODE_EXTRA_CA_CERTS: ${LOGTO_NODE_EXTRA_CA_CERTS:-}
    volumes:
      - ${LOGTO_SETUP_SCRIPTS:-./scripts/logto}:/etc/logto/packages/cli/custom-setup:ro

  pgweb:
    image: sosedoff/pgweb
    restart: always
    logging: *default-logging
    networks: [net]
    env_file:
      - ${PLATFORM_ENV_FILE:-.env.local}
    environment:
      DATABASE_URL: ${PGWEB_DATABASE_URL}
    labels:
      - "traefik.enable=true"

  dozzle:
    image: amir20/dozzle:latest
    restart: always
    logging: *default-logging
    networks: [net]
    env_file:
      - ${PLATFORM_ENV_FILE:-.env.local}
    environment:
      DOZZLE_NO_ANALYTICS: "true"
      DOZZLE_AUTH_PROVIDER: "none"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock

  api:
    profiles: ["app-prod"]
    image: ghcr.io/your-org/platform/api:${IMAGE_TAG:-latest}
    restart: always
    logging: *default-logging
    networks: [net]
    env_file:
      - ${PLATFORM_ENV_FILE:-/etc/platform/env}

  web:
    profiles: ["app-prod"]
    image: ghcr.io/your-org/platform/web:${IMAGE_TAG:-latest}
    restart: always
    logging: *default-logging
    networks: [net]
    env_file:
      - ${PLATFORM_ENV_FILE:-/etc/platform/env}
```

Stagecraft generates small override files per environment and per host as needed, but this base file stays under version control.

⸻

## 4. Command tree (Cobra)

Top level command: stagecraft

### 4.1 stagecraft init

Purpose:
•	Initialize Stagecraft in a project.

Behavior:
•	Create stagecraft.yml with sensible defaults.
•	Optionally scaffold:
    •	Base docker-compose.yml skeleton.
    •	Sample Encore app and Vite frontend (optional flags).
•	Validate presence of required tools and print hints.

Flags:
•	--with-encore
•	--with-frontend
•	--force overwrite existing files.

⸻

### 4.2 stagecraft dev

Purpose:
•	Run full local development environment.

Behavior:
1.	Load stagecraft.yml.
2.	Load environment config for dev.
3.	Ensure local prerequisites:
   •	Docker installed and running.
   •	Encore CLI is available.
   •	mkcert installed (if traefik.mode is mkcert).
   •	Tailscale (if connected dev is enabled later).
4.	Prepare local HTTPS:
   •	Generate mkcert root and certs if missing.
   •	Update /etc/hosts entries for local dev domains if configured.
   •	Generate Traefik dynamic config files.
5.	Start infra via Docker Compose:
   •	Build environment variables from environments.dev.
   •	Call:
         •  docker compose -f docker-compose.yml -f docker-compose.override.dev.yml up -d <infra services>
6.	Backend dev (Encore provider):
   •	Load .env.local from backend.dev.env_file.
   •	For each secret listed in encore_secrets.from_env:
         •  If variable exists:
             •	Run: encore secret set --type dev,preview,local SECRET_NAME and pipe value on stdin.
   •	Construct backend env:
       •	Base on .env.local.
       •	Add DISABLE_ENCORE_TELEMETRY if configured.
       •	Add NODE_EXTRA_CA_CERTS if configured.
   •	Spawn encore run --debug --verbose --watch --listen 0.0.0.0:4000.
7.	Frontend dev (generic provider):
   •	Spawn command defined in frontend.dev (for example Vite).
8.	Maintain process group:
   •	Keep running until user stops.
   •	On exit, optionally offer to stop containers.

Flags:
•	--no-infra (run only backend/frontend).
•	--no-backend or --no-frontend.
•	--connected (use Tailscale for other device access if configured).

⸻

### 4.3 stagecraft build

Purpose:
•	Build Docker images for deployment.

Baseline expectation: run primarily in CI.

Behavior:
1.	Load stagecraft.yml.
2.	For each buildable component:
    •	BackendProvider:
        •	Call encore build docker <backendImage:tag>.
    •	FrontendProvider:
        •	Call docker build or configured build steps (if configured).
3.	Push images to registry defined in project.registry.
4.	Emit an artifact describing:
    •	version (probably git SHA).
    •	image tags used.

Flags:
•	--env (defaults to staging).
•	--version (override version tag).
•	--skip-push.

⸻

### 4.4 stagecraft deploy

Purpose:
•	Deploy a given version to a named environment.

Behavior:
1.	Load stagecraft.yml.
2.	Load environment config for target env (staging or prod).
3.	Resolve version:
    •	Provided via --version, or default to CI injected value or current git SHA.
4.	Network bootstrap:
    •	Ensure target hosts exist and are reachable (assumes infra already up).
    •	For each host:
        •	Ensure Docker present.
        •	Ensure Tailscale or Headscale agent installed and running.
        •	Join tailnet if not already.
5.	Compose preparation:
    •	For each host role:
        •	Derive service subset for that host from services mapping.
        •	Generate a per host docker-compose.generated.yml and optionally docker-compose.override.prod.yml that:
            •	Sets env vars for image tags (for example IMAGE_TAG=version).
            •	Sets volume roots for that environment.
            •	Adds Traefik ACME config on gateway host.
            •	Injects DB and Redis hostnames using Tailscale DNS names.
6.	Deployment execution:
    •	For each host:
        •	Upload generated Compose files.
        •	Run docker-rollout up -d (or docker compose up -d if rollout not used initially).
    •	Orchestrate in order:
        •	DB + cache first.
        •	App services.
        •	Gateway Traefik last or rolling.
7.	Update release history store:
•	Record environment, version, timestamp, status.

Flags:
•	--env (staging, prod).
•	--version (Git SHA or tag).
•	--host or --role to limit scope.
•	--dry-run to show intended actions without executing.

⸻

### 4.5 stagecraft rollback

Purpose:
•	Roll back an environment to a previous version.

Behavior:
1.	Load last known good version for environment from release history store.
2.	Run stagecraft deploy --env ENV --version PREVIOUS_VERSION with a flag so that history is updated appropriately.

Flags:
•	--env.
•	--to VERSION explicit target instead of automatic previous.

⸻

### 4.6 stagecraft infra up / stagecraft infra down

Purpose:
•	Provision or destroy remote infrastructure.

Behavior for infra up:
1.	Load stagecraft.yml.
2.	Pick environment (staging or prod).
3.	Use CloudProvider (DO CLI first) to:
    •	Create droplets for roles defined in hosts (if absent).
    •	Attach volumes according to env settings.
    •	Configure DO Firewalls to:
        •	Allow 22 (optional) and 80/443 externally.
        •	Restrict other ports.
    •	Optionally setup DO Load Balancer.
4.	For each new droplet:
    •	Run bootstrap:
        •	Install Docker.
        •	Install docker-rollout (if used).
        •	Install Tailscale or Headscale client.
        •	Join tailnet with tags for role.
        •	Create /etc/platform/env with minimal content.

Behavior for infra down:
•	Tear down droplets and attached volumes for that environment.
•	Optionally clean up DNS.

Flags:
•	--env.
•	--plan (print Infra plan, no apply).
•	--keep-volumes (destroy droplets but keep volume data).

⸻

### 4.7 stagecraft ci init / stagecraft ci run

Purpose:
•	Initialize and interact with CI workflows.

ci init:
•	Generate .github/workflows/stagecraft.yml with:
    •	Jobs:
        •	test - run backend and frontend tests.
        •	build - run stagecraft build.
        •	deploy - run stagecraft deploy.
•	Use gh to set repository secrets needed:
    •	Registry credentials.
    •	Tailscale auth key env.
    •	Stripe keys and other required secrets.

ci run:
•	Trigger GitHub Actions workflow via gh workflow run.
•	Show link and optionally follow status.

⸻

### 4.8 stagecraft status

Purpose:
•	Show current status of an environment.

Behavior:
•	For each host in target environment:
    •	Use SSH (preferably via tailscale ssh) to:
        •	Run docker ps filtered by project, or
        •	Query a local agent (future enhancement).
•	Show:
    •	Running containers.
    •	Versions (from image tags).
    •	Healthcheck status if available.

⸻

### 4.9 stagecraft logs

Purpose:
•	Tail logs of a service in an environment.

Behavior:
•	For a given service and env:
    •	Determine host(s) running that service using role mapping.
    •	Use SSH to call docker logs -f or connect to Dozzle.
•	Show log stream.

Flags:
•	--service (required).
•	--env.
•	--host or --role.

⸻

### 4.10 stagecraft ssh

Purpose:
•	Open an SSH session to a host or role.

Behavior:
•	Prefer tailscale ssh HOST.
•	Fallback: ssh to public IP if configured.

⸻

### 4.11 stagecraft secrets sync (optional but recommended)

Purpose:
•	Sync secrets for a given environment.

Behavior:
•	For dev:
    •	Sync values from .env.local to Encore dev secrets and local env as already described.
•	For remote environments:
    •	Update /etc/platform/env on hosts or push secrets into a provider (Vault, Doppler, etc.) according to configuration.

⸻

## 5. Go architecture

High level package structure:

```text
stagecraft/
  cmd/
    root.go             # Cobra root command
    dev.go
    deploy.go
    init.go
    infra.go
    ci.go
    status.go
    logs.go
    ssh.go
    rollback.go
  docs/
    stagecraft-spec.md
  pkg/
    config/               # Load and validate stagecraft.yml
    compose/              # Compose file handling and override generation
    providers/
        backend/
          encorets/
        frontend/
          devcommand/
        network/
          tailscale/
          headscale/
        cloud/
          digitalocean/
        ci/
          github/
        secrets/
          envfile/
          encoredev/
    dev/                  # Orchestration logic for `stagecraft dev`
    deploy/               # Orchestration logic for build/deploy/rollback
    infra/                # Infra up/down workflows
    ci/                   # CI integration workflows
    logging/              # Structured logging helpers
    executil/             # Process execution and streaming
```

### 5.1 Core interfaces

BackendProvider
```go
type DevOptions struct {
    Env       map[string]string
    Listen    string
    WorkDir   string
}

type BuildOptions struct {
    ImageTag string
    WorkDir  string
}

type BackendProvider interface {
    ID() string
    Dev(ctx context.Context, opts DevOptions) error
    BuildDocker(ctx context.Context, opts BuildOptions) (string, error)
}
```

EncoreTsProvider implementation:
•	Dev:
    •	Sync secrets using encore secret set from env.
    •	Spawn encore run --debug --verbose --watch --listen <listen> with proper env.
•	BuildDocker:
    •	Run encore build docker IMAGE:TAG.

FrontendProvider
```go
type FrontendDevOptions struct {
    Env     map[string]string
    WorkDir string
    Command []string
}

type FrontendProvider interface {
    ID() string
    Dev(ctx context.Context, opts FrontendDevOptions) error
}
```

DevCommandProvider:
•	Runs the configured command in given workdir.

NetworkProvider
```go
type NetworkProvider interface {
    EnsureInstalled(ctx context.Context, host string) error
    EnsureJoined(ctx context.Context, host string, tags []string) error
    NodeFQDN(host string) (string, error) // for example plat-db-1.mytailnet.ts.net
}
```
Tailscale provider implementation shells out to tailscale CLI on the host.

CloudProvider
```go
type HostSpec struct {
    Name   string
    Role   string
    Size   string
    Region string
}

type InfraPlan struct {
    ToCreate []HostSpec
    ToDelete []HostSpec
}

type CloudProvider interface {
    Plan(ctx context.Context, env string) (InfraPlan, error)
    Apply(ctx context.Context, plan InfraPlan) error
}
```

DigitalOcean provider uses doctl or DO API directly.

CIProvider
```go
type CIProvider interface {
    Init(ctx context.Context) error
    Trigger(ctx context.Context, env string, version string) error
}
```
GitHub provider wraps gh or GitHub REST API.

⸻

## 6. Design principles for DX

    •	Single command mental model:
        •	stagecraft dev for local.
        •	stagecraft deploy for remote.
	•	Declarative config:
        •	Users describe environment strategy in stagecraft.yml.
        •	Stagecraft takes care of commands and wiring around it.
	•	Pluggable backends and infra:
	•	    Encore.ts is first backend provider, but everything goes through interfaces.
	•	Minimal magic inside compose:
        •	Compose describes services.
        •	Stagecraft manipulates env, override files and host selection.
	•	Mesh by default:
    	•	Multi host networking through Tailscale or Headscale is first class.
	•	CI first for production images:
	    •	Local builds are optional; CI is canonical for release artifacts.

⸻

## 7. Cross-cutting concerns

	•	Global CLI behavior
        •	Global flags: --env, --config, --verbose/--quiet, --dry-run.
        •	Consistent exit codes and error formatting.
	•	Config resolution rules
        •	Precedence: flags → env vars → stagecraft.yml → built-in defaults.
        •	Ability to override config path (STAGECRAFT_CONFIG env or --config).
	•	OS / platform assumptions
	    •	Officially support: macOS + Linux (Windows WSL later).
	•	Release metadata storage
	    •	Where “release history” lives (e.g. S3/DO Spaces JSON, or Git-tracked file).
	•	Testing hooks
	    •	Abstractions that make providers mockable (interfaces already help here).

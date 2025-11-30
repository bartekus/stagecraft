## Why not Kamal

### 1. **Kamal evaluation and decision**
- Decision: Do not adopt Kamal directly; build a custom tool inspired by Kamal
- Rationale:
    - Kamal uses Ruby and kamal-proxy; the project prefers Go/Rust and Traefik
    - Kamal assumes one app per host; the project needs multi-app/multi-tenant
    - Kamal’s proxy conflicts with the Traefik-first architecture

### 2. **Core architecture decisions**

**Language choice:**
- Decision: Build the main CLI in Go (not Rust)
- Rationale: Better ecosystem fit (Docker, Tailscale, Encore are Go-native), simpler for orchestration, easier cross-compilation

**Networking:**
- Decision: Use Tailscale/Headscale for mesh networking
- Rationale: Enables single-host → multi-host scaling without changing the service model; services keep the same hostnames across hosts

**Orchestration model:**
- Decision: Docker Compose as the canonical stack definition + docker-rollout for zero-downtime updates
- Rationale: Single source of truth for services; dev/prod parity; works with Traefik

**Proxy strategy:**
- Decision: Traefik as the edge proxy (not kamal-proxy)
- Rationale: Already central to the design; supports multi-tenant routing; works with existing local dev setup

### 3. **Deployment tool design**

**Tool name:**
- Decision: "Stagecraft" (Go-based CLI)

**Provider model:**
- Decision: Pluggable provider architecture
- Providers:
    - BackendProvider: Encore.ts (default)
    - FrontendProvider: Vite/generic dev commands
    - NetworkProvider: Tailscale/Headscale
    - CloudProvider: DigitalOcean (first)
    - CIProvider: GitHub Actions (first)

**Config strategy:**
- Decision: Single canonical `docker-compose.yml` + `stagecraft.yml` config
- Environment differences handled via:
    - Environment variables
    - Generated override files
    - CLI-driven configuration (not hardcoded in compose)

### 4. **Local development decisions**

**Tailscale in local dev:**
- Decision: Make Tailscale optional but first-class in dev bootstrap
- Two modes:
    - `--mode=local`: Pure local, no Tailscale required
    - `--mode=connected`: Tailscale-enabled for cross-device testing and shared dev resources

**HTTPS in dev:**
- Decision: Continue using mkcert for local HTTPS
- Integration: mkcert setup automated in dev bootstrap alongside Tailscale

**Encore integration:**
- Decision: Treat Encore.ts as a BackendProvider, not the top-level orchestrator
- Responsibilities:
    - Encore CLI: Run/test/build backend
    - Stagecraft CLI: Orchestrate infra, Traefik, multi-service deployment

### 5. **Scaling strategy**

**Single host → multi-host:**
- Decision: Use role-based host mapping in `platform.deploy.yml`
- Services can move between hosts by changing role assignments
- Tailscale DNS keeps service hostnames stable regardless of physical location

**State management:**
- Decision: Stateless orchestrator (no Terraform-style state backend)
- "State" lives in:
    - Git (config)
    - Container registry (images)
    - Running containers on hosts

### 6. **Implementation decisions**

**Compose file strategy:**
- Decision: One canonical compose file with profiles and env vars
- CLI generates environment-specific overrides rather than maintaining separate files

**Build strategy:**
- Decision: CI-first for production images (linux/amd64)
- Local builds are optional/experimental
- Encore `build docker` works cross-platform but production images built in CI

**Release management:**
- Decision: Simple release history tracking (version per environment)
- Rollback capability built into deploy command

### 7. **Feature scope decisions**

**Core commands:**
- `stagecraft init` - Scaffold config
- `stagecraft dev` - Local development
- `stagecraft build` - Build Docker images
- `stagecraft deploy` - Deploy to environments
- `stagecraft rollback` - Rollback to previous version
- `stagecraft infra up/down` - Infrastructure provisioning
- `stagecraft ci init/run` - CI integration
- `stagecraft status/logs/ssh` - Operations commands

**Free tier compatibility:**
- Decision: Design for Tailscale free tier (<100 devices)
- All planned features work within free tier limits
- Option to migrate to Headscale (fully open-source) later if needed

### 8. **Documentation and spec**

- Decision: Create `stagecraft-spec.md` as the canonical design document
- Keep implementation checklist in README
- Both documents stay in sync as the project evolves

These decisions shape Stagecraft as a Go-based, Compose-first, Tailscale-enabled deployment orchestrator that bridges local development and multi-host production while keeping the same mental model throughout.
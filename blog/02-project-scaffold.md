## Project Scaffold

> **Related Documents:**
> - [`spec/scaffold/stagecraft-dir.md`](../spec/scaffold/stagecraft-dir.md) - `.stagecraft/` directory structure
> - [`docs/implementation-roadmap.md`](../docs/implementation-roadmap.md) - Phase 10: Project Scaffold
> - [`spec/features.yaml`](../spec/features.yaml) - Scaffold-related features (CLI_INIT_TEMPLATE, CLI_NEW, etc.)

### 1. **Two modes, one engine**
- **Drop-in mode** (`stagecraft init` or `stagecraft attach`): Add Stagecraft to existing repos with minimal changes
- **Scaffold mode** (`stagecraft new --template=platform`): Generate a new project from a template
- Both modes converge on the same manifest and folder structure

### 2. **Core contract: stagecraft.yml + .stagecraft/**
- **Manifest file** (`stagecraft.yml`): Single source of truth describing:
    - Project metadata (name, type)
    - Environments (local, prod, etc.)
    - Services (paths, types, run/deploy configs)
    - Infrastructure providers (docker, digitalocean, etc.)
    - Database migration configuration
- **Note**: `.yaml` and `.toml` extensions are reserved for future support, but v1 uses `.yml` exclusively.
- **`.stagecraft/` directory**: CLI workspace containing:
    - Generated docker-compose files
    - Environment templates
    - `agent/Agent.md` (AI development guide)
    - Health checks and smoke tests
    - Project-specific codegen templates
    - Release history (`releases.json`)
    - State files and metadata

### 3. **Platform template approach**
- Embedded as a first-class template in Stagecraft (via `embed.FS` in Go or a well-known git repo)
- Includes pre-configured structure: `apps/`, `services/`, `infra/`
- Pre-filled `stagecraft.yml` with Encore backend, Traefik, Logto, Postgres, Redis
- Includes DigitalOcean and GitHub Actions hooks (even if stubs)

### 4. **Repository structure for CLI**
```
/cmd/stagecraft          # main CLI
/internal
  /core                  # orchestration engine
  /spec                  # manifest schema, loaders, validators
  /project               # project discovery, init/attach
  /providers             # DO, GH, etc adapters
  /templates             # embedded project templates
  /ai                    # Agent.md generation
/pkg
  /cli                   # user-facing command helpers
  /config                # shared config types
/testdata/projects       # golden tests for different scenarios
```

### 5. **Provider abstraction**
- Infrastructure providers (DigitalOcean, GitHub Actions, etc.) are backend adapters
- Defined in manifest via `provider:` fields
- Implemented as Go interfaces in `pkg/providers/`
- Extensible for future providers (GCP, Kubernetes, etc.)

### 6. **Implementation priorities**
- Define v1 `stagecraft.yml` schema in Go (types + validation)
- Sketch Platform template's `stagecraft.yml`
- Sketch what `stagecraft init` emits for generic, non-Platform repos

### 7. **Design principles**
- Spec-driven, testable CLI
- Greenfield-ready but not greenfield-only
- AI-guardrail-friendly (Agent.md generation)
- Explicit boundaries and contracts
- Bloggable implementation with clear testability

This design supports both new projects (via templates) and existing projects (via drop-in), with a consistent contract that enables all higher-level commands (`up`, `deploy`, `test`, `doctor`, etc.) to work uniformly.
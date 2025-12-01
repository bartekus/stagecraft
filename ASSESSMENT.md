## Stagecraft

### Project Overview

Stagecraft is a Go-based CLI for orchestrating local development and deployment of multi-service applications.
It aims to be a "A local-first tool that scales from single-host to multi-host deployments like Kamal, but for Docker Compose"

### Core Goals

1. Local-first DX: one command to spin up full local infrastructure, HTTPS, backend, and frontend
2. Smooth scaling: single host → multi-host without switching tools
3. Provider model: pluggable providers for Encore.ts, Vite, DigitalOcean, GitHub Actions, etc.
4. Configuration-driven: single `stagecraft.yml` plus canonical `docker-compose.yml`

### Architecture

The project follows a layered architecture (ADR 0001):

1. CLI layer (`internal/cli/`): User-facing commands using Cobra
2. Core layer (`internal/core/`): Planning, state management, environment resolution (planned)
3. Driver layer (`internal/drivers/`): Platform-specific implementations (planned)
4. Provider layer (`pkg/providers/`, `internal/providers/`): Pluggable backends, migrations, etc.
5. Support libraries (`pkg/`): Config loading, shared utilities

### Current Implementation Status

#### ✅ Completed Features

1. Architecture & documentation
    - Architecture docs and ADR process
    - Project structure defined

2. Core infrastructure
    - Config loading (`pkg/config/`) with validation
    - Backend provider registry system
    - Migration engine registry system
    - Provider registration via `init()` functions

3. CLI commands (basic implementations)
    - `stagecraft init` — stub (doesn't create config yet)
    - `stagecraft dev` — delegates to backend provider
    - `stagecraft migrate` — runs migrations using registered engines

4. Provider implementations
    - Generic backend provider (command-based)
    - Encore.ts backend provider (stub)
    - Raw SQL migration engine (with tracking table)

### Key Design Decisions

1. Registry pattern: Providers register themselves via `init()` functions
   ```go
   func init() {
       backend.Register(&GenericProvider{})
   }
   ```

2. Provider-agnostic config: Provider-specific config is stored as `any` and unmarshaled by each provider

3. Spec-driven development: Features tracked in `spec/features.yaml` with status, specs, and tests

4. Test-first approach: Test files defined before implementation

### Technology Stack

- Language: Go 1.23.3
- CLI framework: Cobra
- Config: YAML (via `gopkg.in/yaml.v3`)
- Database: PostgreSQL (via `pgx/v5` for migrations)
- Dependencies: Minimal, focused

### Project Structure Highlights

```
stagecraft/
├── cmd/stagecraft/          # Entry point
├── internal/
│   ├── cli/                 # CLI commands (init, dev, migrate)
│   └── providers/           # Provider implementations
├── pkg/
│   ├── config/              # Config loading & validation
│   └── providers/           # Provider interfaces & registries
├── spec/                    # Feature specifications
├── docs/                    # Documentation & ADRs
├── examples/                # Example projects
└── test/e2e/                # End-to-end tests
```

### Strengths

1. Clear architecture with separation of concerns
2. Extensible provider system
3. Spec-driven development with feature tracking
4. Test coverage targets (80%+ for core packages)
5. Good documentation structure (ADRs, specs, guides)
6. Minimal dependencies

### Areas for Improvement / Next Steps

1. Config creation: `init` command doesn't create `stagecraft.yml` yet
2. Global flags: `--env`, `--config`, `--dry-run` not fully implemented
3. Core orchestration: Planning engine, state management, Compose integration not implemented
4. Provider implementations: Many providers are stubs (Encore.ts, Tailscale, DigitalOcean, etc.)
5. Local dev features: mkcert, Traefik, `/etc/hosts` management not implemented
6. Deployment: Build, deploy, rollback commands not implemented

### Development Approach

- Spec-first: Features specified before implementation
- Test-first: Tests defined in `features.yaml` before coding
- Feature tracking: 61+ features organized in phases in `spec/features.yaml`
- Status: Most features are `todo`; only architecture/docs are `done`

### Notable Code Quality

- Error handling: Consistent error wrapping with context
- Type safety: Interfaces properly defined and validated
- Testing: Test files exist for core functionality
- Documentation: Code comments reference feature IDs and specs

### Summary

Stagecraft is an early-stage project with a solid foundation. The architecture is clear, the provider system is extensible, and the development process is structured. Most features are planned but not implemented. The current codebase provides:

- Working config loading and validation
- Functional `dev` and `migrate` commands
- Extensible provider registry system
- Basic migration engine (raw SQL)

The project is positioned to scale as features are implemented, with a clear roadmap in `spec/features.yaml` organized into 10 phases.

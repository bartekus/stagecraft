>Prompt: Please summarize the existing project architecture, structure, development style, implementation details and code quality.

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

**Layered architecture** (ADR 0001):
1. **CLI Layer** (`internal/cli/`) — Cobra commands, user input, output
2. **Core Layer** (`internal/core/`) — Planning, state, config interpretation (planned)
3. **Driver Layer** (`internal/drivers/`) — Platform-specific implementations (planned)
4. **Provider layer** (`pkg/providers/`, `internal/providers/`): Pluggable backends, migrations, etc.
5**Support Libraries** (`pkg/`) — Reusable components (config, providers, utilities)

**Provider model**:
- Plugin-based with registries for backend providers and migration engines
- Provider-agnostic core; provider-specific config is opaque
- Current providers: `generic`, `encore-ts` (backend); `raw` (migrations)
- Registry pattern with thread-safe registration/lookup

**Data flow**:
```
User Command → CLI Layer → Config Loading → Provider Resolution → Execution
```

### **Project Structure**

**Current structure**:
```
stagecraft/
├── cmd/stagecraft/          # Entry point
├── internal/
│   ├── cli/                 # CLI commands (dev, migrate, init)
│   └── providers/           # Provider implementations
│       ├── backend/         # generic, encorets
│       └── migration/       # raw
├── pkg/
│   ├── config/              # Config loading & validation
│   └── providers/           # Provider interfaces & registries
│       ├── backend/
│       └── migration/
├── spec/                    # Feature specifications
│   ├── commands/
│   ├── core/
│   └── providers/
├── docs/                    # Documentation
│   ├── adr/                 # Architecture Decision Records
│   └── guides/
├── test/e2e/                # End-to-end tests
└── examples/                # Example projects
```

**Design principles**:
- `internal/` for implementation details
- `pkg/` for reusable/public APIs
- `spec/` for feature specs (spec-first)
- Clear separation of concerns

### **Development Style**

**Spec-first, test-first**:
- Features tracked in `spec/features.yaml` with IDs, status, specs, tests
- Specs in `spec/` before implementation
- Tests written before or alongside implementation
- Feature IDs referenced in code comments

**Code organization**:
- Feature comments: `// Feature: CLI_DEV_BASIC`
- Interface-driven design (provider interfaces)
- Registry pattern for extensibility
- Error wrapping with context (`fmt.Errorf("...: %w", err)`)

**Documentation**:
- ADRs for architectural decisions
- Specs for feature behavior
- Inline comments explaining design choices
- `Agent.md` for AI-assisted development guidelines

### **Implementation Details**

**Current features**:

1. **Config system** (`pkg/config/`):
    - YAML config loading with validation
    - Registry-based provider/engine validation
    - Provider-scoped config under `backend.providers.<id>`
    - Database migration config support

2. **CLI commands**:
    - `stagecraft dev` — Runs backend provider in dev mode
    - `stagecraft migrate` — Executes database migrations (with `--plan`)
    - `stagecraft init` — Project bootstrap (stub)

3. **Provider system**:
    - Backend providers: `generic` (command-based), `encore-ts` (Encore.ts)
    - Migration engines: `raw` (SQL file-based with tracking table)
    - Thread-safe registries with panic on duplicate registration

4. **Testing approach**:
    - Unit tests for core logic
    - Integration tests for registry behavior
    - E2E smoke tests (gated behind `e2e` build tag)
    - Golden file tests for CLI output

**Technical stack**:
- Go 1.23.3
- Cobra for CLI
- YAML v3 for config
- PostgreSQL (pgx/v5) for migrations
- Minimal external dependencies

### **Code Quality**

**Testing**:
- 83+ test functions across the codebase
- Coverage targets: 80%+ for core packages, 60%+ overall
- Coverage script (`scripts/check-coverage.sh`) with thresholds
- Tests use `t.TempDir()` for isolation
- Integration tests verify registry behavior

**Code standards**:
- Go formatting (`gofmt`/`goimports`)
- Interface compliance checks: `var _ Interface = (*Type)(nil)`
- Error handling with wrapped errors
- Context propagation for cancellation
- Clear function/method naming

**Quality indicators**:
- ✅ Comprehensive test coverage
- ✅ Clear error messages with context
- ✅ Thread-safe registry implementations
- ✅ Provider-agnostic validation (no hardcoded provider checks)
- ✅ Spec-driven development
- ✅ Documentation in code and markdown

**Areas for improvement**:
- Some features marked `todo` in `spec/features.yaml`
- `init` command is a stub
- Core planning engine not yet implemented
- No structured logging yet (planned)

**Development workflow**:
- Git hooks via `scripts/install-hooks.sh`
- Spec validation script
- Coverage checking script
- Feature tracking in YAML

### **Overall Assessment**

**Strengths**:
1. Clear architecture with defined layers
2. Strong testing discipline
3. Spec-first approach
4. Extensible provider model
5. Good documentation
6. Clean code organization

**Maturity**:
- Early WIP/experimental (per README)
- Core foundation solid
- Basic commands functional
- Many planned features not yet implemented

**Code quality**: High
- Well-structured, tested, documented
- Follows Go best practices
- Maintainable and extensible

The project shows a solid foundation with a clear path for growth, emphasizing quality, testability, and maintainability.
>Prompt: Please summarize the existing project architecture, structure, development style, implementation details and code quality.

## Stagecraft

### Project Overview

Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

### Core Goals

1. Local-first DX: one command to spin up full local infrastructure, HTTPS, backend, and frontend
2. Smooth scaling: single host → multi-host without switching tools
3. Provider model: pluggable providers for Encore.ts, Vite, DigitalOcean, GitHub Actions, etc.
4. Configuration-driven: single `stagecraft.yml` plus canonical `docker-compose.yml`

### Architecture

**Layered architecture** (ADR 0001):
1. **CLI Layer** (`internal/cli/`) — Cobra commands, user input, output
2. **Core Layer** (`internal/core/`) — Planning, state, config interpretation (basic implementation)
3. **Driver Layer** (`internal/drivers/`) — Platform-specific implementations (planned, not yet implemented)
4. **Provider layer** (`pkg/providers/`, `internal/providers/`): Pluggable backends, migrations, etc.
5. **Support Libraries** (`pkg/`) — Reusable components (config, providers, utilities)

**Provider model**:
- Plugin-based with registries for backend providers and migration engines
- Provider-agnostic core; provider-specific config is opaque (`map[string]any`)
- Current providers: `generic` (command-based), `encore-ts` (Encore.ts) for backends; `raw` (SQL file-based) for migrations
- Registry pattern with thread-safe registration/lookup (panic on duplicate registration)
- Provider config scoped under `backend.providers.<id>` to maintain agnosticism
- Validation uses registry lookups, never hardcoded provider checks

**Data flow**:
```
User Command → CLI Layer → Config Loading → Provider Resolution → Execution
```

### Project Structure

**Current structure**:
```
stagecraft/
├── cmd/stagecraft/          # Entry point (main.go)
├── internal/
│   ├── cli/                 # CLI commands (dev, migrate, init)
│   │   └── commands/        # Individual command implementations
│   ├── core/                # Core domain logic (planning engine)
│   └── providers/           # Provider implementations
│       ├── backend/         # generic, encorets
│       └── migration/       # raw
├── pkg/
│   ├── config/              # Config loading & validation
│   ├── logging/             # Structured logging (implemented)
│   └── providers/           # Provider interfaces & registries
│       ├── backend/
│       └── migration/
├── spec/                    # Feature specifications
│   ├── commands/
│   ├── core/
│   └── providers/
├── docs/                    # Documentation
│   ├── adr/                 # Architecture Decision Records
│   ├── guides/
│   └── reference/
├── test/e2e/                # End-to-end tests (gated behind build tags)
└── examples/                # Example projects
```

**Design principles**:
- `internal/` for implementation details (not exported)
- `pkg/` for reusable/public APIs
- `spec/` for feature specs (spec-first development)
- Clear separation of concerns
- Feature IDs traceable in code comments

### Development Style

**Spec-first, test-first**:
- Features tracked in `spec/features.yaml` with IDs, status, specs, tests
- Specs in `spec/` before implementation
- Tests written before or alongside implementation
- Feature IDs referenced in code comments: `// Feature: CLI_DEV_BASIC`
- ADR process for architectural decisions

**Code organization**:
- Feature comments: `// Feature: CLI_DEV_BASIC` with spec references
- Interface-driven design (provider interfaces)
- Registry pattern for extensibility
- Error wrapping with context (`fmt.Errorf("...: %w", err)`)
- Context propagation for cancellation
- Interface compliance checks: `var _ Interface = (*Type)(nil)`

**Documentation**:
- ADRs for architectural decisions (`docs/adr/`)
- Specs for feature behavior (`spec/`)
- Inline comments explaining design choices
- `Agent.md` for AI-assisted development guidelines
- License headers enforced (AGPL-3.0-or-later) with automated checking

**Quality enforcement**:
- Git hooks via `scripts/install-hooks.sh`
- Coverage checking: `scripts/check-coverage.sh` (80% core, 60% overall)
- Spec validation: `scripts/validate-spec.sh`
- Linting: golangci-lint with curated ruleset
- License header enforcement via `addlicense` tool

### Implementation Details

**Current features**:

1. **Config system** (`pkg/config/`):
    - YAML config loading with validation
    - Registry-based provider/engine validation (no hardcoded checks)
    - Provider-scoped config under `backend.providers.<id>`
    - Database migration config support
    - Environment configuration

2. **CLI commands**:
    - `stagecraft dev` ✅ — Runs backend provider in dev mode
    - `stagecraft migrate` ✅ — Executes database migrations (with `--plan` flag)
    - `stagecraft init` ⚠️ — Project bootstrap (stub implementation)

3. **Provider system**:
    - Backend providers: `generic` (command-based), `encore-ts` (Encore.ts)
    - Migration engines: `raw` (SQL file-based with `stagecraft_migrations` tracking table)
    - Thread-safe registries with panic on duplicate registration
    - Auto-registration via `init()` functions

4. **Core planning engine** (`internal/core/`):
    - Basic deployment planning implemented
    - Creates plans with operations (migrations, builds, deploys, health checks)
    - Supports migration strategies (pre_deploy, post_deploy, manual)
    - Dependency tracking structure in place (not yet fully utilized)

5. **Structured logging** (`pkg/logging/`):
    - Implemented with levels (Debug, Info, Warn, Error)
    - Structured fields support
    - Verbose mode for debug output
    - Timestamped output

6. **Testing approach**:
    - Unit tests for core logic
    - Integration tests for registry behavior
    - E2E smoke tests (gated behind `e2e` build tag)
    - Golden file tests for CLI output
    - Tests use `t.TempDir()` for isolation

**Technical stack**:
- Go 1.23.3
- Cobra for CLI framework
- YAML v3 for config parsing
- PostgreSQL (pgx/v5) for migrations
- Minimal external dependencies (focused dependency set)

### Code Quality

**Testing**:
- **93 test functions** across 16 test files
- Coverage targets: **80%+ for core packages** (`pkg/config`, `internal/core`), **60%+ overall**
- Coverage script (`scripts/check-coverage.sh`) with thresholds and per-package reporting
- Tests use `t.TempDir()` for isolation
- Integration tests verify registry behavior
- E2E tests structured with build tags for CI gating

**Code standards**:
- Go formatting (`gofmt`/`goimports`) enforced
- Interface compliance checks: `var _ Interface = (*Type)(nil)`
- Error handling with wrapped errors (`fmt.Errorf("...: %w", err)`)
- Context propagation for cancellation
- Clear function/method naming
- License headers in all source files (AGPL-3.0-or-later)

**Quality indicators**:
- ✅ Comprehensive test coverage (93 test functions)
- ✅ Clear error messages with context
- ✅ Thread-safe registry implementations
- ✅ Provider-agnostic validation (no hardcoded provider checks)
- ✅ Spec-driven development with feature tracking
- ✅ Documentation in code and markdown
- ✅ Structured logging implemented
- ✅ Planning engine foundation in place
- ✅ License compliance enforced

**Areas for improvement**:
- Many features marked `todo` in `spec/features.yaml`
- `init` command is a stub (needs full implementation)
- Core planning engine is basic (needs dependency resolution and execution)
- Driver layer not yet implemented (planned for future phases)
- Some E2E tests may need expansion

**Development workflow**:
- Git hooks via `scripts/install-hooks.sh` (format, license checks)
- Spec validation script (`scripts/validate-spec.sh`)
- Coverage checking script (`scripts/check-coverage.sh`)
- Feature tracking in YAML (`spec/features.yaml`)
- CLI docs generation (`scripts/generate-cli-docs.sh`)

### Overall Assessment

**Strengths**:
1. **Clear architecture** with defined layers and separation of concerns
2. **Strong testing discipline** with 93 test functions and coverage enforcement
3. **Spec-first approach** with feature tracking and traceability
4. **Extensible provider model** with registry pattern and agnostic validation
5. **Good documentation** including ADRs, specs, and inline comments
6. **Clean code organization** following Go best practices
7. **Quality enforcement** via tooling, hooks, and CI integration
8. **License compliance** with automated header checking

**Maturity**:
- **Status**: Early WIP/experimental (per README)
- **Core foundation**: Solid and well-structured
- **Basic commands**: Functional (`dev`, `migrate`)
- **Planning engine**: Basic implementation in place
- **Structured logging**: Implemented and functional
- **Many planned features**: Not yet implemented (driver layer, full init, etc.)

**Code quality**: **High**
- Well-structured, tested, and documented
- Follows Go best practices consistently
- Maintainable and extensible architecture
- Strong emphasis on quality gates and validation
- Production-ready foundation with clear growth path

The project demonstrates a **solid foundation** with a clear path for growth, emphasizing quality, testability, and maintainability. The spec-first, test-first approach provides excellent traceability and ensures features are well-documented before implementation.

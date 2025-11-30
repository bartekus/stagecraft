# 0001 – Stagecraft Architecture and Project Structure

- Status: Accepted
- Date: 2025-11-29

## Context

We are building Stagecraft as a Go-based CLI to orchestrate deployments and infrastructure workflows.
The codebase should:

- Be easy to navigate and extend.
- Support strong testing discipline.
- Play nicely with AI-assisted development (Cursor, ChatGPT).
- Be a public showcase of engineering quality.

We need a clear architectural and directory structure to support these goals.

## Decision

We adopt the following high-level structure:

### Core Structure (v1)

- `cmd/` – CLI entrypoints (Cobra root and subcommands).
- `internal/`
    - `core/` – Core domain logic (planning, state, environment resolution).
    - `drivers/` – Platform-specific implementations (e.g. DigitalOcean, GitHub Actions).
    - `cli/` – CLI-specific wiring and UX (commands, prompts, output formatting).
- `pkg/` – Reusable libraries (e.g. config loader, plugin interfaces).
- `spec/` – Machine- and human-readable specifications of features and commands.
- `docs/`
    - `adr/` – Architecture Decision Records.
    - `guides/` – How-to documents.
    - `reference/` – CLI reference and internal APIs where needed.
- `test/`
    - `e2e/` – End-to-end and integration tests.
- `Agent.md` – Guardrails and instructions for AI-assisted development.
- `blog/` – Drafts of public-facing posts documenting the journey.

### Extended Structure (v1+ and v2)

The following directories are reserved for future phases and will be added as features are implemented:

- `internal/`
    - `project/` – Project manifest handling, discovery, init/attach (Phase 10)
    - `runtime/` – Environment orchestration (Phase 2-3)
    - `providers/` – Provider implementations (Phase 1, 4)
    - `deploy/` – Deployment workflow logic (Phase 5)
    - `compose/` – Docker Compose handling (Phase 2)
    - `state/` – State backend (Phase 2)
    - `logging/` – Logging and UX output (Phase 0)
    - `ui/` – UI components (future TUI/dashboard, v2)
    - `templates/` – Embedded project templates (Phase 10)
    - `ai/` – Agent.md generation (Phase 10)
- `pkg/`
    - `providers/` – Provider interfaces (Phase 1)
    - `logging/` – Logging helpers (Phase 0)
    - `executil/` – Process execution utilities (Phase 0)
    - `migrations/` – Migration interfaces (Phase 6)
- `examples/` – Sample projects for dogfooding (v2)

Additionally:

- Each feature is tracked in `spec/features.yaml` with an ID, status, spec path, owner, and associated test files.
- We prefer spec- and test-first development for core functionality.
- See `docs/implementation-roadmap.md` for phase-by-phase directory structure evolution.

## Alternatives Considered

1. **Flat structure with everything under `pkg/`**
    - Pros: Simplifies imports.
    - Cons: Blurs distinction between internal and public APIs; harder to maintain encapsulation.

2. **Single `internal` monolith (no `core`/`drivers`/`cli` separation)**
    - Pros: Less upfront structure.
    - Cons: Makes layering and responsibilities unclear; harder to manage over time.

3. **Framework-style monorepo with multiple binaries**
    - Pros: Could support many tools.
    - Cons: Premature complexity for current scope; Stagecraft is initially a single CLI with plugins.

## Consequences

- Positive:
    - Clear layering and extensibility: CLI ↔ core ↔ drivers.
    - Easy to orient humans and AI in the codebase.
    - Supports strong testing discipline via clear seams.
- Negative:
    - Slight upfront overhead to maintain structure.
    - Some features may initially feel “over-structured” until the codebase grows into it.

We accept these trade-offs in favour of clarity and long-term maintainability.

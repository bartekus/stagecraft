Analyzing the document and summarizing the decisions.


[1 tool called]


## Summary of Decisions

### 1. Feature prioritization
- All 15 new feature ideas are deferred to v2.
- Focus v1 on core functionality (Kamal-inspired, DO droplet + Docker Compose).
- Top 5 v2 candidates:
    1. Ephemeral environments
    2. Build/deploy replay & audit ledger
    3. AI-enhanced test harness
    4. Unified secrets orchestrator
    5. Infrastructure topology map

### 2. Architecture decision
- Structure the codebase now to support v2 without a rewrite.
- Use a directory structure that:
    - Keeps v1 focused and clean
    - Provides extension points for v2 features
    - Works with Cursor/Agent.md/spec-driven development

### 3. Proposed directory structure
- `cmd/stagecraft/` — Cobra command entrypoints
- `internal/` — Implementation details (not for external import):
    - `cli/` — Cobra glue layer
    - `config/` — Config loading/validation
    - `project/` — Project manifest handling
    - `runtime/` — Environment orchestration
    - `providers/` — Cloud provider abstractions (DO, GitHub)
    - `deploy/` — Deployment workflow logic
    - `compose/` — Docker Compose handling
    - `state/` — State backend (v1: local JSON; v2: audit ledger)
    - `logging/` — Logging and UX output
    - `ui/` — UI components (future TUI/dashboard)
- `pkg/` — Minimal, stable public APIs (schema, API)
- `docs/` — Specs, ADRs, progress tracking
- `examples/` — Sample projects for dogfooding

### 4. Next steps (suggested)
1. Lock in the directory structure (create directories + placeholder files)
2. Create `docs/spec/v1-overview.md` with v1 goals and non-goals
3. Create `docs/progress/feature-matrix.md` for tracking implementation status
4. Update README.md to reflect v1 scope and v2 readiness

The structure is designed to allow adding v2 features (ephemeral envs, recipes, topology map, etc.) without major refactoring.
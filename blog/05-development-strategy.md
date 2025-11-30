Analyzing the document and summarizing the decisions.


[1 tool called]


## Summary of Decisions

### 1. **Development methodology**
- Spec-first, TDD-heavy, ADR-driven workflow
- Four loops:
    1. Spec → Slice → Tests → Code → Docs
    2. ADRs for major decisions
    3. Strict CI quality gates
    4. Public narrative: blog posts per milestone
- Vertical slices: one CLI command/sub-feature at a time
- TDD for core libraries; pragmatic testing for outer layers (CLI cosmetics)

### 2. **Repository structure**
- `cmd/` - CLI entrypoints (Cobra root and subcommands)
- `internal/` - Implementation details:
    - `core/` - Core domain logic (planning, state, environment resolution)
    - `drivers/` - Platform-specific implementations (DigitalOcean, GitHub Actions, etc.)
    - `cli/` - CLI-specific wiring and UX
- `pkg/` - Reusable libraries (config loader, plugin interfaces)
- `spec/` - Machine- and human-readable specifications
- `docs/` - ADRs, guides, reference
- `test/e2e/` - End-to-end tests
- `blog/` - Draft blog posts

### 3. **Spec and traceability**
- Machine-readable `spec/features.yaml` with feature IDs, status, specs, and tests
- Human-oriented Markdown specs per feature
- Code comments linking to feature IDs and specs (e.g., `// Feature: CLI_INIT`)
- Each feature must have: spec item, tests, ADR (if needed), docs

### 4. **Testing strategy**
- Unit tests: high coverage, TDD where possible, table-driven for config/parsing
- CLI behavior: golden file tests for output
- Integration/E2E: in `test/e2e/`, tagged as slow for CI
- Drivers: abstract cloud APIs behind interfaces, use fakes for unit tests

### 5. **Quality gates and tooling**
- Formatting: `gofmt`/`goimports` enforced in CI
- Linting: `staticcheck` or `golangci-lint` with curated ruleset
- Static analysis: `go vet` in CI
- Coverage: minimum 80%+ on core packages
- Pre-commit hooks: fmt, lint, unit tests on touched packages
- GitHub Actions: lint, test, optional docs-check jobs

### 6. **Documentation approach**
- Code-level: package docs, public types/functions documented
- CLI docs: generated from Cobra into `docs/reference/cli.md`
- Architecture: system overview in `docs/architecture.md`
- ADRs: one per non-trivial design decision, auto-numbered (0001, 0002, etc.)

### 7. **Blogging strategy**
- Blog drafts in-repo under `/blog/drafts`
- Each major milestone: 1 ADR + 1 devlog draft
- Posts reference ADR IDs, PRs, and commits
- Early posts planned: founding vision, CLI UX design, spec-first development

### 8. **AI/Cursor workflow**
- `Agent.md` at repo root defining:
    - Project purpose and non-goals
    - Rules for AI contributions
    - Spec-first, test-first requirements
    - Feature traceability requirements
    - Workflow expectations
- Folder-level `Agent.*.md` files allowed for granular instructions
- AI workflow: update `features.yaml` → write spec → create tests → implement → update docs

### 9. **Cobra command pattern**
- Decision: use `NewRootCommand()` constructor pattern instead of Cobra template's global `rootCmd`
- Rationale: better testability, no hidden side effects, clear dependency injection path, fits "world-class" goals

### 10. **Config package design**
- `pkg/config` package with:
    - `DefaultConfigPath()` - returns "stagecraft.yml"
    - `Exists(path)` - checks if config file exists
    - `Load(path)` - loads and validates config
    - `Config` struct with `ProjectConfig` and `EnvironmentConfig`
- Initial validation: project.name must be non-empty, environments must have non-empty driver
- Uses `gopkg.in/yaml.v3` for YAML parsing

### 11. **Initial feature priorities**
- `CLI_INIT` - Project bootstrap command (first to implement)
- `CORE_CONFIG` - Config loading and validation
- `CORE_PLAN` - Deployment planning engine
- `DRIVER_DO` - DigitalOcean driver
- `CLI_PLAN` and `CLI_DEPLOY` - Commands for planning and deployment

### 12. **CI baseline**
- Minimal GitHub Actions workflow with:
    - Formatting verification (`gofmt`)
    - `go vet`
    - Test execution with coverage
- Can expand later with `golangci-lint` and E2E jobs

These decisions establish a spec-first, test-driven, well-documented development process with clear AI collaboration guidelines and quality gates.
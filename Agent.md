⸻

# Agent Guide for Stagecraft

_Deterministic development protocol for AI assistants and human contributors._

>Audience: AI assistants (Cursor, ChatGPT, Copilot, Claude, etc.) and human collaborators.
Purpose: Guarantee spec-driven, test-first, provider-agnostic, registry-based, and deterministic contributions to Stagecraft.

⸻

# 🔥 Project Purpose

Stagecraft is a Go-based CLI orchestrating local-first development, single-host, and multi-host deployments of multi-service applications using Docker Compose.
It reimagines tools like Kamal with:

  *	 A clean, composable, registry-driven architecture (planning, drivers, providers, plugins)
  *  First-class developer UX for both local and remote environments
  *	 Strong correctness guarantees enforced by specs, tests, and documentation
  *	 Extensibility through pluggable providers and migration engines

>Stagecraft is not an orchestration platform, config management engine, or IaC DSL.
It delegates infrastructure heavy lifting to providers, not core.

This repository is both a production-grade tool and a public engineering portfolio.
Clarity, determinism, traceability, and documentation matter as much as functionality.

⸻

# ⭐ Architectural Principles
1.	Spec-driven behaviour – No behaviour exists without a spec.
2.	Test-first change flow – Tests precede implementation.
3.	Registry-based extensibility – No hardcoding, no special cases.
4.	Opaque provider configuration – Core never interprets provider-specific config.
5.	Predictable and idempotent operations
6.	Strict package boundaries
7.	Minimal diffs, maximal clarity
8.	Traceability from spec → tests → code → docs → git
9.	Determinism over convenience
10.	No non-deterministic behaviour – No random data, timestamps, or environment-dependent logic unless specified.
11.	Reproducibility – Running the same command twice must produce identical results unless external state has changed.

These principles override ambiguous instructions.

⸻

# 🧭 Golden Rules

## 1. Spec-first, Test-first

Before modifying or creating behaviour, locate the relevant section of:
  *	spec/features.yaml
  *	spec/<domain>/<feature>.md

New behaviour must follow this order:
1.	Write or update the spec
2.	Write failing tests
3.	Implement the smallest behavioural change
4.	Make tests pass
5.	Update docs
6.	Update feature status in spec/features.yaml

AI must never skip steps.
AI must never fill in missing specs by guessing.

⸻

## 2. Feature ID Rules

All meaningful changes must reference a Feature ID:
```go
// Feature: CLI_INIT
// Spec: spec/commands/init.md
```

#### Creating a Feature ID

Create a new Feature ID when:
  *	Adding user-facing behaviour
  *	Adding a CLI command
  *	Adding a provider or migration engine
  *	Changing config schema with behavioural impact

Do NOT create a new Feature ID for:
  *	Refactors
  *	Bug fixes
  *	Docs-only changes

#### Feature ID Naming Rules
  *	SCREAMING_SNAKE_CASE
  *	Must map directly to a spec file
  *	Must be unique and stable
  *	Never renamed after merge
  *	Never fork feature development across branches

>Feature definitions must never be placeholders; a new feature must contain at least one explicit behavioural statement.

⸻

## 3. Feature Lifecycle

Feature states live in spec/features.yaml:
```
todo → wip → done → deprecated → removed
```

  * deprecated = behaviour still exists but is slated for removal
  *	removed = behaviour no longer exists; docs updated

A feature becomes done only when:
  *	Spec is complete
  *	Tests are complete and passing
  *	Implementation is complete
  *	Docs are complete
  *	No ambiguity remains

The contributor (AI or human) must update the feature state.

⸻

### 🧪 Test Discipline

#### Core rules
  *	Tests MUST be written before implementation.
  *	Every behaviour change must include tests.
  *	Tests must cover:
    *	Happy path
    *	Failure path
    *	Edge conditions
    *	CLI behaviour where appropriate
    *	Registry integration
  *	Avoid non-determinism:
    *	No timestamps
    *	No random UUIDs
    *	No environment-dependent paths
  * All output lists must use deterministic ordering (prefer lexicographical ascending).

#### Golden Tests

Used when testing CLI output, config generation, or structured text.

Rules:
  *	Golden files must live in testdata/
  *	Golden files updated only when behaviour changes and spec is updated
  *	Always review golden diffs carefully

#### Parallelism
  *	Tests MUST NOT use t.Parallel() unless explicitly allowed in the spec.

⸻

## 4. Package Boundaries
  *	internal/ contains implementation details — no public APIs
  *	pkg/ contains stable, reusable packages for external use
  *	cmd/ must stay thin; wiring only
  * Packages MUST NOT form cyclic imports.

Directional rule:
  *	internal/ MAY import pkg/
  *	pkg/ MUST NOT import internal/

Never place business logic in cmd/.

⸻

## 5. File Modification Restrictions

Do not modify these without explicit human approval:
  *	LICENSE
  *	High-level README.md
  *	ADRs (never rewrite history; append new ADRs)
  *	Global governance files
  *	NOTICE
  *	CHANGELOG.md (if present)

>Tooling configs (.golangci.yml, .gitignore, .goreleaser.yml) may be modified, but require justification and minimal diffs.

If modification is necessary:
  *	Justify in commit & PR
  *	Keep diffs minimal

⸻

## 6. Go Style and Quality Standards
  * go build ./... must pass
  *	Format with gofmt, goimports, and gofumpt
  *	go test ./... must fully pass
  * All exported symbols must include GoDoc comments.
  *	Fix all golangci-lint warnings unless suppressed with justification:
```go
// nolint:gocritic // interface requires value
```

⸻

## 7. Provider and Engine Agnosticism

### Absolute Rules
*	Never hardcode provider or engine IDs
*	Never treat Encore.ts or Drizzle as special

__Bad__:
```go
if provider == "encore-ts" { ... }
```
__Good__:
```go
if !backendproviders.Has(provider) { ... }
```

### Provider/Engine Boundaries
  *	Core defines interfaces and registries only
  *	Providers/engines implement interfaces
  *	Core must never interpret provider-specific config
  *	Provider-specific logic must never leak into core
  *	Providers must never modify core behaviour
  * Providers MUST NOT read environment variables directly unless explicitly documented in their spec. All configuration must enter through provider config maps.

### Provider Registration Rules
  *	Registration MUST occur via init() in the provider’s own package
  *	Core MUST NOT instantiate providers manually
  *	Registration occurs through import side-effects in pkg/config/config.go
  *	Duplicate registration must be tested
  * Providers MUST NOT write to stdout/stderr directly; they must use structured logging.
  * Provider and engine registry iteration MUST be lexicographically sorted to ensure deterministic behaviour.

### Config Schema Rules

Provider configuration keys must follow:
```code
backend.providers.<id>.<env>.<configkey>
```

__Non-determinism policy__

Provider/engine loading must not depend on environment ordering or file system randomness.

⸻

## 📁 Folder-Level Instructions
  *	Some folders may contain additional Agent.md
  *	Both top-level and local rules apply
  *	If they conflict, human maintainer overrides all
       * Clarify order of precedence:
       1.	Human maintainer
       2.	Local Agent.md
       3.	Top-level Agent.md

Nested Agent.md files do not override parent definitions unless explicitly stated.

⸻

## 🧱 Naming Conventions
  *	Go types: __PascalCase__
  *	Interfaces: end with er (Provider, Planner)
  *	Packages: short, lowercase, no underscores
  *	Test files: <name>_test.go
  *	Spec files: spec/<domain>/<feature>.md
  *	Feature IDs: SCREAMING_SNAKE_CASE
  *	Errors: prefix with domain/feature
  *	Config keys: __kebab-case__
  * All generated files must be placed under .stagecraft/ unless specified, and must use kebab-case.

__CLI Command Names__
  *	MUST use dashed names: stagecraft deploy-plan
  *	NEVER camelCase or snake_case

⸻

## 🧩 Error Handling Rules
  *	Wrap all errors using fmt.Errorf("context: %w", err)
  *	Never return plain strings
  *	Avoid shadowed variables
  *	Errors must be deterministic and structured
  * Error messages MUST NOT include full system paths unless essential for debugging.

__Sentinel Errors__
  *	Used only when multiple packages must detect the same condition
  *	Must live in the lowest-level appropriate package
  *	Must be stable and documented

⸻

## 🧲 Behavioural Guardrails for AI

AI MUST:
  *	Make minimal diffs
  *	Never refactor unless explicitly instructed
  *	Never reorganize directories without approval
  *	Stay within scope of the task
  *	Always reference the Feature ID
  *	Always follow spec → tests → code → docs → commit order
  *	Ask for clarification when the spec is ambiguous
  *	Prefer precision over creativity
  *	Never introduce new dependencies without explicit approval
  * Ensure that a task affecting one feature does not cause incidental changes to unrelated features.

AI MUST NOT:
  *	Guess behaviour
  *	Invent features
  *	Generate large speculative changes
  *	Modify protected files
  *	Change registry loading behaviour
  *	Add non-deterministic code paths
  * Create new files unless explicitly required by the Feature ID or spec.

⸻

## 🧵 Git Workflow Rules (Critical)

### 1. Every task ends with a commit message

For each completed task, output:

A. Human summary
B. Commit message (strict format)

Commit message:
```code
<type>(<FEATURE_ID>): <short summary>

Optional longer explanation.
Spec: <path/to/spec.md>
Tests: <path/to/tests>
```

Allowed types: feat, fix, refactor, docs, test, ci, chore

>Commits must be as small and isolated as possible; avoid bundling unrelated changes.

>PRs MUST be merged using “squash and merge” unless the maintainer requests otherwise.

⸻

### 2. Each behavioural feature MUST be implemented in its own PR

__PR Title__
```code
[FEATURE_ID] <Short human description>
```

__PR Description__
  *	Feature:
  *	Spec:
  *	Tests:
  *	Summary
  *	Rationale
  *	Constraints

__PR Requirements__
  *	Atomic, spec-driven
  *	No mixing multiple features
  *	All tests pass
  *	Spec updated
  *	Docs updated
  *	Feature status updated

__Branch Naming Rules__
```code
feature/<FEATURE_ID>-short-desc
fix/<FEATURE_ID>-short-desc
chore/<short-desc>
docs/<short-desc>
```

  *	No uppercase
  *	No spaces

__PR Metadata__
  *	Must include a label (feature, fix, docs, test, ci, chore)
  *	Must be in draft until tests pass
  *	Human reviewer required
  *	Default target branch: main

⸻

## 🚫 Non-Goals
  *	Stagecraft is NOT a general-purpose automation tool
  *	Do not add speculative features without an ADR
  *	All new behaviour must be anchored to a Feature ID and spec
  * Stagecraft is not a plugin framework or workflow engine.

⸻

## ✔ When in Doubt

Prefer:
  *	clarity
  *	simplicity
  *	determinism
  *	traceability

over cleverness or abstraction.

When behaviour is ambiguous:
  *	Do not modify code
  *	Produce a clarification request summarizing options
  *	Wait for explicit human direction

⸻

✅ End of Agent Guide

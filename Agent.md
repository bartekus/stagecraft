# Agent Guide for Stagecraft

> Audience: AI assistants (Cursor, ChatGPT, Copilot, Claude, etc.) and human collaborators using them.  
> Purpose: Ensure deterministic, spec-driven, test-driven, provider-agnostic contributions to the Stagecraft codebase.

---

# 🔥 Project Purpose

Stagecraft is a Go-based CLI that orchestrates local-first application development, deployment, and infrastructure workflows.  
It reimagines tools like Kamal with:

- A clean, composable, registry-driven core (planning, drivers, providers, plugins)
- First-class developer UX for both local and remote workflows
- Strong correctness guarantees through specs, tests, and docs
- Extensibility through pluggable providers and migration engines

This repository is both a **production-grade tool** and a **public engineering portfolio**.  
Clarity, reasoning, determinism, and documentation matter as much as functionality.

---

# ⭐ Architectural Principles

1. **Spec-driven behaviour**  
2. **Test-first change flow**  
3. **Registry-based extensibility**  
4. **Opaque provider configuration**  
5. **Predictable and idempotent operations**  
6. **Strict package boundaries**  
7. **Minimal diffs, maximal clarity**  
8. **Traceability from spec → tests → code → docs → git**

These principles override ambiguous instructions.

---

# 🧭 Golden Rules

## 1. Spec-first, test-first
- Before implementing or modifying behaviour, inspect the relevant location in:
  - `spec/features.yaml`
  - The spec markdown under `spec/<domain>/<feature>.md`
- For new behaviour:
  1. Write or update the spec  
  2. Write failing tests  
  3. Implement code  
  4. Make tests pass  
  5. Update docs  

## 2. Every change MUST trace to a feature ID
Each meaningful change references a feature ID from `spec/features.yaml`:

```go
// Feature: CLI_INIT
// Spec: spec/commands/init.md
```

When a new behaviour is introduced:
	•	Add a feature entry with status: todo
	•	Add or update its spec file
	•	Write tests before implementation

Create a new feature ID when:
	•	Adding new user-facing behaviour
	•	Adding a new CLI command
	•	Adding a new provider or migration engine
	•	Changing config schema with behavioural impact

Do NOT create new feature IDs for:
	•	Pure refactors
	•	Bug fixes
	•	Docs-only changes

Feature ID Naming Rules:
  • Feature IDs MUST be unique and stable.
  • Format: SCREAMING_SNAKE_CASE.
  • Feature IDs must map directly to a spec file in spec/<domain>/.
  • Do not reuse or rename Feature IDs once merged.

⸻

# 3. Tests and docs are non-optional

Every behavioural change must:
	•	Add/update tests (*_test.go)
	•	Update or create the feature spec in spec/
	•	Update user docs in docs/ if applicable
	•	Update the feature's status (todo → wip → done) only when implementation + tests + docs are complete

Tests must fail before implementation.

Feature State Lifecycle:
  • Feature states live in spec/features.yaml.
  • Valid states: todo → wip → done.
  • State MUST be updated by the contributor completing the feature.
  • A feature is “done” only when:
  – Spec is complete
  – Tests are complete and passing
  – Implementation is complete
  – Docs are updated

⸻

# 4. Respect package boundaries
	•	internal/ contains implementation details — no public APIs should leak from here.
	•	pkg/ contains reusable and externally consumable packages.
	•	cmd/ must stay thin — command wiring only.
Never place business logic in cmd/.

⸻

# 5. Do not modify certain files unless explicitly asked

Only change the following files when the human explicitly requests it or when required to complete a clearly defined task:
	•	LICENSE
	•	High-level README.md positioning or messaging
	•	Existing ADRs (docs/adr/*) — append new ADRs instead of editing history
	•	Global governance files

If such a modification is necessary:
	•	Justify it in comments or commit messages
	•	Keep diffs minimal

⸻

# 6. Follow Go style and quality standards
	•	Run go build ./...
	•	Format code via gofmt and goimports
	•	Run go test ./... and ensure full pass
	•	Address golangci-lint findings unless explicitly suppressed with justification:

```go
// nolint:gocritic // explanation: interface requires value
```

⸻

# 7. Provider and Engine Agnosticism

Hard rule: Never hardcode provider or engine IDs

❌ Bad:
```go
if provider != "encore-ts" && provider != "generic" { ... }
```

✅ Good:
```go
if !backendproviders.Has(provider) { ... }
```

Provider-specific config must be scoped:
```code
backend.providers.<id>.<env>.<configkey>
```

Provider/engine rules:
	•	Provider configuration is opaque to core (map[string]any)
	•	Encore.ts is not special
	•	Drizzle is not special
	•	Provider-specific logic lives inside the provider implementation
	•	Migration engine-specific logic lives inside the engine implementation
	•	Core never contains exceptions for specific providers or engines

Provider Registration:
  • Providers MUST register themselves through init() side effects.
  • Registration must occur inside the provider's own package.
  • Core MUST NOT instantiate providers manually or via conditionals.

Registry wiring requirements:
	•	Reference:
	•	CORE_BACKEND_REGISTRY
	•	CORE_MIGRATION_REGISTRY
	•	CORE_BACKEND_PROVIDER_CONFIG_SCHEMA
	•	Update the spec before modifying code
	•	Ensure provider/engine registration happens via import side effects in pkg/config/config.go
	•	Never bypass the registry

Provider and Engine Boundaries:
  • Core defines interfaces and registries ONLY.
  • Providers implement interfaces, never adjust core.
  • No provider or engine is privileged (Encore.ts and Drizzle included).

⸻

📁 Folder-Level Instructions

Some folders may contain their own Agent.md.
When present:
	•	Follow both the top-level Agent.md and the local version
	•	If they conflict, defer to the human maintainer

Local Agent.md Precedence:
  • Local Agent.md files apply only to their folder subtree.
  • When rules conflict, human maintainer’s instructions override both.
⸻

🧪 Test Discipline

Write tests BEFORE implementation:
	1.	Add feature spec
	2.	Write failing tests
	3.	Implement smallest possible change
	4.	Make tests pass
	5.	Add regressions for discovered edge cases
	6.	Refactor only after green tests

Tests must cover:
	•	Happy path
	•	Failure cases
	•	Edge conditions
	•	CLI-level behaviour where appropriate
	•	Registry integration where applicable

Golden Tests:
  • Use golden files when testing CLI output, config generation, or structured text.
  • Golden files belong in testdata/ subfolders.
  • Update golden files only when behaviour changes AND after spec updates.

⸻

🔄 Multi-File Change Protocol

When a task requires modifying multiple files:
	1.	Update the spec first
	2.	Write failing tests
	3.	Modify implementation
	4.	Adjust docs
	5.	Produce commit message
	6.	Prepare PR description

AI should not skip steps.
Minimal diffs preferred.

⸻

❓ Ambiguity Rule

When the spec is ambiguous or unclear:
	•	Do not guess.
	•	Leave existing behaviour unchanged.
	•	Produce a clarification request summarizing options.
	•	Never invent new behaviour without explicit human approval.

⸻

🧱 Naming Conventions
	•	Go types: PascalCase
	•	Interfaces: end with er (e.g., Provider, Planner)
	•	Package names: short, lower-case, no underscores
	•	Test files: <name>_test.go
	•	Spec files: spec/<domain>/<feature>.md
	•	Feature IDs: SCREAMING_SNAKE_CASE
	•	Errors: prefix with domain or feature:

fmt.Errorf("backend provider validation failed: %w", err)

CLI Command Names:
  • CLI commands MUST use dashed names (e.g., stagecraft deploy-plan).
  • Do not use underscores or camelCase for command names.

⸻

🧩 Error Handling Rules
	•	Wrap all errors (fmt.Errorf("context: %w", err))
	•	Never return plain strings
	•	Use deterministic, structured error messages
	•	Avoid shadowing variables

Sentinel Errors:
  • Use sentinel error variables when multiple packages must detect a specific error.
  • Sentinel errors MUST live in the lowest-level appropriate package.

⸻

🧲 Behavioural Guardrails for AI
	•	Make minimal diffs
	•	Do not refactor unless explicitly instructed
	•	Do not rewrite large blocks of code or reorganize directories without approval
	•	Stay within scope of the requested task
	•	Always reference the feature ID
	•	Always follow spec → tests → code → docs → commit order
	•	Ask for clarification when necessary
	•	Prefer precision over creativity

⸻

## 🧵 Git Workflow Rules (Critical)

### 1. Every task ends with a commit message

For each completed task, output:

A. Human summary (free-form)

B. Commit message (strict-form)

The commit message format:
```code
<type>(<feature_id>): <short summary>

Longer explanation if necessary.
Spec: <path/to/spec.md>
Tests: <path/to/tests>
```

Allowed types:
	•	feat
	•	fix
	•	refactor
	•	docs
	•	test
	•	ci
	•	chore

⸻

### 2. Each behavioural feature must be implemented in a dedicated PR

PR Title

[FEATURE_ID] <Short human-readable description>

PR Description

Feature: <id>
Spec: <path>
Tests: <list of test files>
Summary:
- What changed
- Why it changed
- Any constraints or alternatives considered

PR Requirements
	•	Small, atomic, spec-driven
	•	Behavioural changes must not mix multiple features
	•	Tests must pass
	•	Specs must be updated
	•	Docs must be updated
	•	Feature status must be updated

Branch Naming Rules:
  • Feature branches:
      feature/<FEATURE_ID>-short-desc
  • Bug fix branches:
      fix/<FEATURE_ID>-short-desc
  • Chore branches:
      chore/<short-desc>
  • Docs-only branches:
      docs/<short-desc>
  • Branch names MUST NOT contain spaces or uppercase letters.

PR Metadata Requirements:
  • Each PR MUST have:
    – Label: feature, fix, docs, test, ci, chore
    – Milestone: matching release cycle (if applicable)
    – Draft state until tests pass
  • Human reviewer required before merge.

⸻

🚫 Non-Goals
	•	Stagecraft is not a general-purpose automation framework
	•	Avoid experimental or speculative changes unless backed by an ADR
	•	Avoid adding behaviour not anchored to a feature

⸻

✔ When in doubt

Favor:
	•	clarity
	•	simplicity
	•	determinism
	•	traceability
over cleverness or abstraction.

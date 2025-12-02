⸻

# Agent Guide for Stagecraft

__Deterministic development protocol for AI assistants and human contributors.__

> Audience: AI assistants (Cursor, ChatGPT, Copilot, Claude, etc.) and human collaborators.
> Purpose: Guarantee spec‑driven, test‑first, provider‑agnostic, registry‑based, and deterministic contributions to
> Stagecraft.

⸻

# ⚡ AI Quickstart (TL;DR)

Before doing anything:

1. Identify the Feature ID for the task. If none exists, stop and ask.
2. Locate the spec: spec/features.yaml and spec/<domain>/<feature>.md.
3. Write or update tests first – they must fail before you write code.
4. Limit scope to a single feature – never mix multiple features in one change.
5. Respect provider boundaries – core is provider‑agnostic.
6. Do not touch protected files (LICENSE, top‑level README, ADRs, NOTICE, etc.).
7. Keep changes deterministic – no randomness, timestamps, or environment‑dependent behaviour.
8. Respond with the AI Response Format Contract (Summary, Diff Intent, File List, etc.).
9. End with a commit message that follows the Git Workflow Rules.

If any of the above is unclear, stop and ask for clarification instead of guessing.

⸻

# 🔥 Project Purpose

Stagecraft is a Go‑based CLI orchestrating local‑first development, single‑host, and multi‑host deployments of
multi‑service applications using Docker Compose.

It reimagines tools like Kamal with:

* A clean, composable, registry‑driven architecture (planning, drivers, providers, plugins)
* First‑class developer UX for both local and remote environments
* Strong correctness guarantees enforced by specs, tests, and documentation
* Extensibility through pluggable providers and migration engines

> Stagecraft is not an orchestration platform, config management engine, or IaC DSL. </br>
> It delegates infrastructure heavy lifting to providers, not core.

This repository is both a production‑grade tool and a public engineering portfolio. </br>
Clarity, determinism, traceability, and documentation matter as much as functionality.

⸻

# ⭐ Architectural Principles

1. Spec‑driven behaviour – no behaviour exists without a spec.
2. Test‑first change flow – tests precede implementation.
3. Registry‑based extensibility – no hardcoding, no special cases.
4. Opaque provider configuration – core never interprets provider‑specific config.
5. Predictable and idempotent operations.
6. Strict package boundaries.
7. Minimal diffs, maximal clarity.
8. Traceability from spec → tests → code → docs → git.
9. Determinism over convenience.
10. No non‑deterministic behaviour – no random data, timestamps, or environment‑dependent logic unless explicitly
    specified.
11. Reproducibility – running the same command twice must produce identical results unless external state has changed.

These principles override ambiguous instructions.

⸻

# 🧭 Golden Rules

## 1. Spec‑first, Test‑first

Before modifying or creating behaviour, locate the relevant section of:

* spec/features.yaml
* spec/<domain>/<feature>.md

New behaviour must follow this order:

1. Write or update the spec.
2. Write failing tests.
3. Implement the smallest behavioural change.
4. Make tests pass.
5. Update docs.
6. Update feature status in spec/features.yaml.

AI MUST NOT skip steps.
AI MUST NOT fill in missing specs by guessing.

Example: Full Feature Lifecycle

For CLI_INIT:

1. Add or update spec/commands/init.md (including version).
2. Mark CLI_INIT as wip in spec/features.yaml.
3. Add or update tests (e.g. cmd/init_test.go) that fail.
4. Implement minimal code in cmd/init.go and supporting packages.
5. Ensure go test ./... passes.
6. Update docs (usage, examples).
7. Mark CLI_INIT as done in spec/features.yaml.
8. Provide a commit message and PR description referencing CLI_INIT.

⸻

## 2. Feature ID Rules

All meaningful changes must reference a Feature ID:

```go
// Feature: CLI_INIT
// Spec: spec/commands/init.md
```

Creating a Feature ID

Create a new Feature ID when:

* Adding user‑facing behaviour
* Adding a CLI command
* Adding a provider or migration engine
* Changing config schema with behavioural impact

Do NOT create a new Feature ID for:

* Refactors
* Bug fixes
* Docs‑only changes

Feature ID Naming Rules

* SCREAMING_SNAKE_CASE
* Must map directly to a spec file
* Must be unique and stable
* Never renamed after merge
* Never fork feature development across branches

Feature definitions must never be placeholders; a new feature must contain at least one explicit behavioural statement.

⸻

## 3. Feature Lifecycle

Feature states live in spec/features.yaml:

```text
todo → wip → done → deprecated → removed
```

* deprecated = behaviour still exists but is slated for removal
* removed = behaviour no longer exists; docs updated

A feature becomes done only when:

* Spec is complete
* Tests are complete and passing
* Implementation is complete
* Docs are complete
* No ambiguity remains

The contributor (AI or human) must update the feature state.

⸻

### 🧪 Test Discipline

#### Core Rules

* Tests MUST be written before implementation.
* Every behaviour change must include tests.
* Tests must cover:
  * Happy path
  * Failure path
  * Edge conditions
  * CLI behaviour where appropriate
  * Registry integration
* Avoid non‑determinism:
  * No timestamps
  * No random UUIDs
  * No environment‑dependent paths
* All output lists must use deterministic ordering (prefer lexicographical ascending).

#### Golden Tests

Use golden tests for CLI output, config generation, or structured text.

Rules:

* Golden files must live in testdata/.
* Golden files are updated only when behaviour changes and the spec is updated.
* Always review golden diffs carefully.

#### Parallelism

* Tests MUST NOT use t.Parallel() unless explicitly allowed in the spec.

#### Test Layout

Unless explicitly specified otherwise:

* Tests SHOULD be colocated with the code they validate (e.g. internal/<pkg>/foo.go and internal/<pkg>/foo_test.go).
* Golden files MUST be stored under a testdata/ directory adjacent to the relevant tests.

⸻

## 4. Package Boundaries

     * internal/ contains implementation details – no public APIs.
     * pkg/ contains stable, reusable packages for external use.
     * cmd/ must stay thin; wiring only.
     * Packages MUST NOT form cyclic imports.

Directional rule:

* internal/ MAY import pkg/.
* pkg/ MUST NOT import internal/.

Never place business logic in cmd/.

⸻

## 5. File Modification Restrictions

Do not modify these without explicit human approval:

* LICENSE
* High‑level README.md
* ADRs (never rewrite history; append new ADRs)
* Global governance files
* NOTICE
* CHANGELOG.md (if present)

> Tooling configs (.golangci.yml, .gitignore, .goreleaser.yml) may be modified, but require justification and minimal
> diffs.

If modification is necessary:

* Justify in commit & PR.
* Keep diffs minimal.

⸻

## 6. Go Style and Quality Standards

* go build ./... must pass.
* Format with gofmt, goimports, and gofumpt.
* go test ./... must fully pass.
* All exported symbols must include GoDoc comments.
* Fix all golangci-lint warnings unless suppressed with justification:

```go
// nolint:gocritic // interface requires value
```

Mocking Policy

* AI MUST NOT introduce new mocking frameworks without explicit approval.
* Generated mocks MUST be deterministic.
* If mocks are generated, their generator version and invocation MUST be documented in the spec or an ADR.

⸻

## 7. Provider and Engine Agnosticism

Absolute Rules

* Never hardcode provider or engine IDs.
* Never treat Encore.ts, Drizzle, or any provider as special.

__Bad__:

```go
if provider == "encore-ts" {
...
}
```

__Good__:

```go
if !backendproviders.Has(provider) {
...
}
```

### Provider/Engine Boundaries

* Core defines interfaces and registries only.
* Providers/engines implement interfaces.
* Core must never interpret provider‑specific config.
* Provider‑specific logic must never leak into core.
* Providers must never modify core behaviour.
* Providers MUST NOT read environment variables directly unless explicitly documented in their spec. All configuration
  must enter through provider config maps.

### Provider Registration Rules

* Registration MUST occur via init() in the provider’s own package.
* Core MUST NOT instantiate providers manually.
* Registration occurs through import side‑effects in pkg/config/config.go.
* Duplicate registration must be tested.
* Providers MUST NOT write to stdout/stderr directly; they must use structured logging.
* Provider and engine registry iteration MUST be lexicographically sorted to ensure deterministic behaviour.

### Config Schema Rules

Provider configuration keys must follow:

```text
backend.providers.<id>.<env>.<configkey>
```

### Non‑determinism policy

Provider/engine loading must not depend on environment ordering or file system randomness.

Example: Duplicate Provider Registration

* Attempting to register provider docker-compose twice MUST return ErrProviderAlreadyRegistered.
* Tests MUST assert the exact error value.

⸻

## 🪪 Provider Registration Conflict Rules

If a provider attempts to register with an already‑registered ID:

* Registration MUST fail deterministically.
* The provider registry MUST return a sentinel error: ErrProviderAlreadyRegistered.
* Tests MUST cover duplicate registration.

Provider registration order MUST NOT depend on import order; iteration MUST sort keys lexicographically.

⸻

## 📁 Folder‑Level Instructions

* Some folders may contain additional Agent.md.
* Both top‑level and local rules apply.
* If they conflict, human maintainer overrides all.

Order of precedence:

1. Human maintainer.
2. Local Agent.md.
3. Top‑level Agent.md.

Nested Agent.md files do not override parent definitions unless explicitly stated.

⸻

## 🧱 Naming Conventions

* Go types: __PascalCase__.
* Interfaces: end with er (e.g. Provider, Planner).
* Packages: short, lowercase, no underscores.
* Test files: <name>_test.go.
* Spec files: spec/<domain>/<feature>.md.
* Feature IDs: SCREAMING_SNAKE_CASE.
* Errors: prefix with domain/feature.
* Config keys: __kebab-case__.
* All generated files must be placed under .stagecraft/ unless specified, and must use kebab-case.

__CLI Command Names__

* MUST use dashed names: stagecraft deploy-plan.
* NEVER camelCase or snake_case.

__CLI Determinism__

* CLI help output MUST be deterministic.
* Command registration MUST use stable and lexicographically sorted ordering.
* No terminal width–dependent formatting that changes output between environments.

⸻

## ☑️ AI Pre‑Change Safety Checklist

Before generating any code, AI MUST verify:

1. What Feature ID does the task belong to?
2. Is the spec present, valid, and complete?
3. Are tests already written? If not, produce failing tests first.
4. Are there protected files in the diff? If yes, halt.
5. Is the change limited to a single feature? If not, halt.
6. Are provider boundaries respected? (core ↔ provider)
7. Will the behaviour be deterministic?

If any answer is unclear, AI MUST stop and ask for clarification.

⸻

## 🧩 Error Handling Rules

* Wrap all errors using fmt.Errorf("context: %w", err).
* Never return plain strings.
* Avoid shadowed variables.
* Errors must be deterministic and structured.
* Error messages MUST NOT include full system paths unless essential for debugging.

__Sentinel Errors__

* Used only when multiple packages must detect the same condition.
* Must live in the lowest‑level appropriate package.
* Must be stable and documented.

⸻

## 🧲 Behavioural Guardrails for AI

AI MUST:

* Make minimal diffs.
* Never refactor unless explicitly instructed.
* Never reorganize directories without approval.
* Stay within scope of the task.
* Always reference the Feature ID.
* Always follow spec → tests → code → docs → commit order.
* Ask for clarification when the spec is ambiguous.
* Prefer precision over creativity.
* Never introduce new dependencies without explicit approval.
* Ensure that a task affecting one feature does not cause incidental changes to unrelated features.

AI MUST NOT:

* Guess behaviour.
* Invent features.
* Generate large speculative changes.
* Modify protected files.
* Change registry loading behaviour.
* Add non‑deterministic code paths.
* Create new files unless explicitly required by the Feature ID or spec.

⸻

## 💬 AI Response Format Contract

Unless explicitly instructed otherwise, every AI task response MUST include:

1. Summary – one paragraph describing what was done.
2. Diff Intent – a human‑readable description of the exact changes to be made.
3. File List – list of files to be created, modified, or deleted.
4. Patch – unified diff (if asked for), minimal and scope‑limited.
5. Commit Message – formatted per Git Workflow Rules.

If the task involves new behaviour:

6. Feature Reference – Feature ID and spec path.
7. Test Plan – list of failing tests to be written or updated.
8. Documentation Changes – list of sections to be updated.

AI MUST NOT produce fully applied diffs without explicit instruction.
AI MUST NOT produce hidden changes beyond the listed file set.

⸻

## 📄 Spec Interpretation Rules

AI MUST treat the written spec as the single source of truth.

When the spec is:

* Silent → AI MUST NOT assume behaviour or invent rules.
* Ambiguous → AI MUST request clarification before writing code.
* Internally inconsistent → AI MUST report and stop.

If the spec is incomplete but a Feature ID exists:

* AI may propose exact wording for missing spec lines.
* A human must approve before tests or code are produced.

⸻

## 🧵 Git Workflow Rules (Critical)

### 1. Every task ends with a commit message

For each completed task (a coherent, single‑feature change), output:

A. Human‑readable summary.
B. Commit message (strict format).

Commit message:

```text
<type>(<FEATURE_ID>): <short summary>

Optional longer explanation.
Spec: <path/to/spec.md>
Tests: <path/to/tests>
```

Allowed types: feat, fix, refactor, docs, test, ci, chore.

Rules:

* Commits must be as small and isolated as possible; avoid bundling unrelated changes.
* PRs MUST be merged using “squash and merge” unless the maintainer requests otherwise.
* AI MUST NOT rewrite commits once pushed to a PR branch, except when instructed.
* AI MUST NOT reorder commits unless instructed.

All commit messages MUST pass linting:

* Max 72 chars in subject.
* No trailing periods.

Commit Granularity

* One commit per completed, single‑feature task is preferred.
* Multi‑step user instructions (e.g. “update spec, tests, and code for CLI_INIT”) SHOULD result in a single commit when
  they all relate to the same feature and are completed within one task.
* If the user explicitly asks for separate commits, follow that instruction.

⸻

### 2. Each behavioural feature MUST be implemented in its own PR

PR Title

```text
[FEATURE_ID] <Short human description>
```

PR Description

* Feature:
* Spec:
* Tests:
* Summary
* Rationale
* Constraints

PR Requirements

* Atomic, spec‑driven.
* No mixing multiple features.
* All tests pass.
* Spec updated.
* Docs updated.
* Feature status updated.

Branch Naming Rules

```text
feature/<FEATURE_ID>-short-desc
fix/<FEATURE_ID>-short-desc
chore/<short-desc>
docs/<short-desc>
```

	  * No uppercase.
	  * No spaces.

PR Metadata

* Must include a label (feature, fix, docs, test, ci, chore).
* Must be in draft until tests pass.
* Human reviewer required.
* Default target branch: main.

AI and PRs

* AI MAY propose branch names and PR titles/descriptions.
* AI MUST NOT transition a PR to ready-for-review. Only humans can do this.
* AI MUST NOT merge PRs.

⸻

## 🧬 ADR Trigger Conditions

A new ADR MUST be created when:

* A design decision affects multiple domains (providers, registry, config).
* A behaviour introduces long‑term architectural constraints.
* Alternatives exist and the choice is not obvious.
* Changes affect performance, security, reproducibility, or provider boundaries.

ADRs MUST follow template:

1. Context
2. Decision
3. Rationale
4. Alternatives
5. Consequences (positive and negative)

⸻

## 🦺 AI Code Generation Safety Rules

AI MUST NOT:

* Introduce dependencies without explicit human approval.
* Generate network calls in core packages.
* Add hidden telemetry or analytics.
* Read external URLs unless part of provider spec.
* Use concurrency unless explicitly allowed.

AI SHOULD:

* Prefer pure functions when possible.
* Minimize side effects.
* Minimize allocations.
* Avoid reflection unless required.

⸻

## 📕 Core Design Invariants

These MUST hold at all times:

1. Deterministic execution given identical inputs.
2. Registry entries sorted lexicographically.
3. No behaviour depends on import order.
4. No environment variables are read by core.
5. provider.Config is opaque to core.
6. No timestamps unless explicitly part of the spec.

* Time MUST be injected via a deterministic clock interface.
* No direct use of time.Now() anywhere in core.

7. Core NEVER shells out.

⸻

## ❤️‍🩹 AI Error Correction Protocol

If AI generates incorrect diffs or behaviour:

1. Undo incorrect generated diffs.
2. Provide corrected minimal diffs.
3. Explain what went wrong and why.
4. Ensure tests cover the regression.

⸻

### 🏴 Feature Mapping Invariant

For every Feature ID:

1. There MUST exist exactly one spec file named in the Feature header.
2. All implementation code MUST reference the same Feature ID.
3. All tests MUST reference the same Feature ID.
4. No two features may share code paths without explicit ADR.
5. Cyclic feature dependencies are forbidden unless defined in an ADR.

If a change touches files mapped to different features, AI MUST halt and request human direction.

⸻

## 📐 Repository State Invariants

The repository MUST remain in a valid state at all times:

1. All specs MUST parse successfully.
2. spec/features.yaml MUST reflect the ground truth of implemented behaviour.
3. No dangling Feature IDs (referenced without spec).
4. No orphan specs (spec exists with no implementation reference).
5. No failing tests on main.
6. Golden test files MUST match code output when regenerated.

If an invariant is violated, AI MUST stop and request human guidance.

⸻

## 🧨 Deterministic Failure Mode Rules

All failure paths MUST:

* Produce stable error messages.
* Produce stable exit codes.
* Produce stable structured logs.
* Avoid multi‑source error ambiguity.

Tests MUST assert exact error values or exact string matches, never substrings.

⸻

## 🌐 Cross‑Cutting Change Rules

When behaviour affects multiple domains, AI MUST:

1. Confirm whether an ADR is required.
2. Identify all Feature IDs impacted.
3. Stop unless the human approves a multi‑feature change.
4. Separate the change into multiple PRs unless explicitly directed otherwise.

⸻

## 📈 Spec Versioning Rules

Spec changes fall into categories:

* additive (allowed)
* clarifying (allowed, no code changes)
* breaking (requires ADR)
* behavioural change (requires Feature ID)

Every spec file MUST contain:

* A version field (e.g. v1, v1.1, etc.).

⸻

## 🧾 Logging Determinism Rules

Logs MUST:

* Use a structured format (JSON).
* Never include timestamps unless injected via deterministic clock.
* Include Feature ID when behaviour is feature‑specific.
* Avoid machine‑specific metadata.

AI MUST NOT introduce new log fields without explicit human approval.

Provider log fields MUST be namespaced:

```text
provider.<id>.<field>
```

⸻

## ⏱ Context and Timeout Rules

context.Context MUST:

* Only be created through a deterministic constructor.
* Never include real‑time deadlines unless specified by the feature.
* Never be cancelled except through deterministic test logic.

No use of context.WithTimeout or context.WithDeadline in core.

⸻

## 🔌 Interface Evolution Rules

Interfaces in pkg/ MUST be stable. Changes require:

* ADR.
* Spec update.
* Migration guidance.
* Major version bump if breaking.

Interfaces in internal/ MAY evolve freely but MUST remain deterministic.

⸻

## 📩 Change Envelope

Every task defines a strict change envelope:

* Only files explicitly listed in the AI Response Format Contract may be modified.
* AI MUST NOT expand the envelope without explicit human permission.
* If an upstream or downstream dependency is impacted, AI MUST halt and request direction.

No side effects, no incidental refactors, no opportunistic cleanups.

⸻

## 🗄️ Deterministic Generation of Files

Generated files MUST:

1. Be reproducible from spec + code alone.
2. Not contain timestamps, UUIDs, or machine‑dependent paths.
3. Be identical when regenerated by different contributors.

Generated files MAY be committed only if:

* The spec explicitly states they must be versioned, or
* They are golden test files.

Generated files MUST be ignored via .gitignore unless versioning is required.

⸻

## 🔱 Human Override Doctrine

Human maintainer overrides apply ONLY to:

* Resolving ambiguity.
* Approving spec changes.
* Approving architecture/ADR changes.

Human overrides MUST NOT:

* Bypass deterministic rules.
* Skip the spec → tests → code flow.
* Introduce untracked behaviour.

Humans cannot add behaviour without updating the spec.

⸻

## 🤖 PR Lifecycle State Machine

draft → ready-for-review → changes-requested → approved → merged

* AI MUST NOT transition a PR to ready-for-review. Only humans can do this.
* A PR MUST NOT be merged unless:
* CI is green.
* Feature state is updated.
* Commit message conforms to rules.
* No protected files were modified.

⸻

## ⏳ Multi‑Step Task Rules

AI MUST NOT compress multi‑step tasks into a single output unless explicitly instructed.

If the user writes “do X then Y”:

* AI MUST stop after X.
* Wait for confirmation.
* Then perform Y.

This avoids premature implementation.

⸻

## 📚 Canonical Error Categories

* ErrInvalidConfig
* ErrProviderUnavailable
* ErrPlanFailed
* ErrRegistryConflict
* ErrFeatureIncomplete
* ErrSpecViolation

⸻

## 📦 Approved Dependencies List

Only the following external dependencies MAY be used without explicit approval:

* cobra
* testify
* go-yaml
* etc. (placeholder, to be replaced with a concrete list)

All other dependencies require explicit human approval, justification, and ADR if architectural.

⸻

## 📌 Toolchain Determinism Rule

All contributors must use the exact versions of:

* Go compiler.
* golangci-lint.
* Test harness.
* Build tools.

as defined in .tool-versions or go.mod.

⸻

## 🚫 Non‑Goals

* Stagecraft is NOT a general‑purpose automation tool.
* Do not add speculative features without an ADR.
* All new behaviour must be anchored to a Feature ID and spec.
* Stagecraft is not a plugin framework or workflow engine.

⸻

## ✔ When in Doubt

Prefer:

* Clarity.
* Simplicity.
* Determinism.
* Traceability.

over cleverness or abstraction.

When behaviour is ambiguous:

* Do not modify code.
* Produce a clarification request summarizing options.
* Wait for explicit human direction.

⸻

## 🛑 Zero‑Tolerance List

* No non‑determinism.
* No guessing.
* No refactors (unless requested).
* No modifying protected files.
* No implicit defaults.
* No implicit auto‑detection.
* No implicit environment reading.

⸻

✅ End of Agent Guide

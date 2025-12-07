⸻

# Agent Guide for Stagecraft

__Deterministic development protocol for AI assistants and human contributors.__

> Audience: AI assistants (Cursor, ChatGPT, Copilot, Claude, etc.) and human collaborators using them.
> Purpose: Guarantee spec‑driven, test‑first, provider‑agnostic, registry‑based, and deterministic contributions to
> Stagecraft.

⸻

# ⚡ AI Quickstart (TL;DR)

Before doing anything, AI MUST:

**Pre-Work Checklist:**
- [ ] Feature ID identified
- [ ] Git hooks verified (see Hook Verification section)
- [ ] Working directory clean
- [ ] On correct feature branch

**Post-Work Checklist (before finishing):**
- [ ] `./scripts/goformat.sh` run (for Go changes)
- [ ] `./scripts/run-all-checks.sh` passes
- [ ] Commit created with correct message format
- [ ] PR created or instructions provided to create PR
- [ ] Branch is clean (`git status` empty)

0. **Identify the Feature ID for the task**
  - If no Feature ID exists, STOP and ask
  - Feature ID is required before any branch operations

0.1. **If creating a new feature (new command, provider behavior, config semantics):**
  - AI MUST recommend running: `./scripts/new-feature.sh <FEATURE_ID> <DOMAIN> [feature-name]`
  - This creates the complete feature skeleton:
    - `docs/analysis/<FEATURE_ID>.md` (Analysis Brief)
    - `docs/<FEATURE_ID>_IMPLEMENTATION_OUTLINE.md` (Implementation Outline)
    - `spec/<domain>/<feature>.md` (Spec file)
  - AI MUST follow the Feature Planning Protocol (see section 1.1) before any implementation
  - After skeleton creation, AI MUST guide the user through filling in:
    1. Analysis Brief (problem, motivation, success criteria)
    2. Implementation Outline (v1 scope, data structures, testing plan)
    3. Spec file (behavioral contract)
  - Only after all three documents are complete and approved may implementation begin

1. **Create or verify feature branch**
  - Ensure a clean working directory
  - Check current branch (see Git Branch Workflow below)
  - If on main → create feature branch from main using FEATURE_ID
  - If on feature branch → verify it matches FEATURE_ID
  - Follow the naming rules (see Git Branch Workflow below)

2. **Locate the relevant spec**
  - spec/features.yaml
  - spec/<domain>/<feature>.md

3. **Write or update tests first**
  - Tests MUST fail before code is written

4. **Limit work to a single feature**
  - Never mix features in a single commit or PR

5. **Respect provider boundaries**
  - Core is always provider‑agnostic
  - Providers implement interfaces, never core logic

6. **Do not touch protected files**
  - LICENSE
  - Main README
  - ADRs
  - NOTICE
  - Any file explicitly marked as protected

7. **Keep changes deterministic**
  - No randomness
  - No timestamps
  - No machine-dependent output
  - No external network dependencies unless mocked

8. **Use the AI Response Format Contract**
  - Summary
  - Diff Intent
  - File List
  - Risks / Reasoning

9. **Stage → Format → Commit → Create/Update PR → Verify after completing work**
   - For Go changes: Run `./scripts/goformat.sh` before staging
   - Run `./scripts/run-all-checks.sh` to verify all CI checks pass
   - Commit message must follow strict rules
   - Create or update PR for the feature branch (see Git Workflow Rules section 3)
   - Summaries MUST be included
   - Branch state MUST be clean afterward

> (see Multi-Step Task Rules for when not to commit intermediate states)” so there’s zero chance of over-committing.

10. **If anything is unclear: STOP and ask.**

Guessing is forbidden.

⸻

# 🔧 Hook Verification (Required)

Git hooks are required for all development work. The pre-commit hook runs gofumpt and basic checks, and will block commits on formatting errors. See CONTRIBUTING.md "Git Hooks" section for details.

**Before any work begins, AI MUST:**

1. **Verify hook exists:**
   ```bash
   ls -la .git/hooks/pre-commit
   ```

2. **If hook is missing:**
   - Run: `./scripts/install-hooks.sh`
   - Verify again: `ls -la .git/hooks/pre-commit`
   - If installation fails: STOP and report error

⸻

# 🔀 Git Branch Workflow (Critical)

This section defines the mandatory Git workflow for all AI-assisted development.

## Pre-Work: Branch Creation

**Prerequisite:** Feature ID MUST be identified before branch creation (see AI Quickstart step 0).

**Prerequisite:** AI MUST complete Hook Verification steps above.

Before starting work, AI MUST:

1. **Ensure working directory is clean**

Run:

```bash
git status
```

If there are uncommitted changes:

- AI MUST NOT proceed
- AI MUST report the state and ask for instruction

2. **Check current branch**

Run:

```bash
git branch --show-current
```

Rules:

- If on `main` → MUST create new feature branch
- If on a feature branch:
  - MUST confirm it matches current FEATURE_ID
  - If mismatch → STOP and ask for direction

3. **Create feature branch from main**

```bash
git checkout main
git pull origin main
git checkout -b feature/<FEATURE_ID>-short-desc
```

Branch naming rules:

- `feature/<FEATURE_ID>-short-desc`
- Use lowercase for short-desc
- Use hyphens instead of spaces
- 3–5 word short description

4. **Verify branch creation**

AI MUST confirm:

```bash
git branch --show-current
```

Result MUST match the newly created branch.

## Error Handling

1. **If `git pull origin main` fails:**
  - Check network connectivity
  - Verify remote is configured: `git remote -v`
  - If still failing: STOP and report error

2. **If branch already exists locally:**
  - Check if it matches current FEATURE_ID
  - If match: Use existing branch
  - If mismatch: STOP and ask for direction

3. **If working directory is not clean:**
  - List uncommitted changes: `git status`
  - STOP and report state
  - Wait for user instruction (stash, commit, or discard)

⸻

## Branch Naming Rules (Enhanced)

```text
feature/<FEATURE_ID>-short-desc     # New features
fix/<FEATURE_ID>-short-desc         # Bug fixes
test/<FEATURE_ID>-short-desc        # Test improvements
docs/<short-desc>                   # Documentation-only changes
chore/<short-desc>                  # Maintenance and cleanup
```

Constraints:

- `<FEATURE_ID>` is uppercase (per spec)
- `<short-desc>` is lowercase
- Hyphens only
- No spaces
- Short, descriptive (3–5 words)

### Branch Naming Examples

**Correct:**

- `feature/PROVIDER_FRONTEND_INTERFACE-frontend-provider`
- `fix/CLI_DEV-bug-fix`
- `chore/update-dependencies`

**Incorrect:**

- `feature/provider_frontend_interface` (FEATURE_ID should be uppercase)
- `feature/PROVIDER_FRONTEND_INTERFACE Frontend Provider` (spaces not allowed)
- `Feature/PROVIDER_FRONTEND_INTERFACE-frontend` (prefix must be lowercase)

⸻

## Branch State Requirements

AI MUST obey the following:

- Feature branch MUST be based on latest main
- Working directory MUST be clean before generating or applying changes
- After operations, AI must run:
  ```bash
  git status
  ```
  If uncommitted work remains → STOP and report.

⸻

# 🔥 Project Purpose

Stagecraft is a Go‑based CLI orchestrating local‑first development, single‑host, and multi‑host deployments of
multi‑service applications using Docker Compose.

It reimagines tools like Kamal with:

* A clean, composable, registry‑driven architecture (planning, drivers, providers, plugins)
* First‑class developer UX for both local and remote environments
* Strong correctness guarantees enforced by specs, tests, and documentation
* Extensibility through pluggable providers and migration engines

> Stagecraft is not an orchestration platform, config management engine, or IaC DSL.

> It delegates infrastructure heavy lifting to providers, not core.

This repository is both a production‑grade tool and a public engineering portfolio.

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

If both the spec and tests are missing for a behaviour, AI MUST stop and request a new or updated spec (or explicit
human direction) before proceeding with any tests or code.

Example: Full Feature Lifecycle

For CLI_INIT:

**Pre-Work:** Create feature branch (see Git Branch Workflow section).

1. Add or update spec/commands/init.md (including version).
2. Mark CLI_INIT as wip in spec/features.yaml.
3. Add or update tests (e.g. cmd/init_test.go) that fail.

**CLI Command Test Location:** CLI command tests MUST live under `cmd/<name>_test.go` unless golden tests are stored in
`cmd/<name>/testdata/`. This ensures consistent test organization across all commands.

4. Implement minimal code in cmd/init.go and supporting packages.
5. Ensure go test ./... passes.
6. Update docs (usage, examples).
7. Mark CLI_INIT as done in spec/features.yaml.
8. Commit with proper message (see Git Workflow Rules).

⸻

## 1.1 Feature Planning Protocol

Before any tests or code are written for a feature, AI MUST perform a three stage planning loop. This loop guarantees spec first, test first, deterministic, and provider safe development. It also ensures that all documentation and specifications remain aligned with the implementation.

⸻

### Stage 1. Feature Analysis Brief

AI MUST check for and update:

```
docs/analysis/<FEATURE_ID>.md
```

This brief defines the problem and the intent of the feature. It MUST include:

- Problem statement

- Motivation for the feature

- User roles and user stories

- 5 to 7 v1 success criteria

- Determinism and side effect constraints

- Risks and architectural boundaries

- Upstream dependencies

If the brief is missing or incomplete, AI MUST create or refine it before continuing.

⸻

### Stage 2. Implementation Outline

AI MUST create or update:

```
docs/<FEATURE_ID>_IMPLEMENTATION_OUTLINE.md
```

This outline describes the full v1 plan for the feature.

It MUST include:

- Feature summary and v1 scope

- What is included, excluded, and reserved for the future

- CLI or API contract (flags, exit codes, examples)

- Data structures and JSON schemas

- Determinism rules and side effect guarantees

- Provider boundaries

- Required tests (unit, integration, golden)

- Completion criteria

The Implementation Outline MUST match the Analysis Brief and MUST be approved before writing tests.

⸻

### Stage 3. Spec Alignment

AI MUST create or update:

```
spec/<domain>/<feature>.md
```

The spec is the authoritative definition of the feature.

It MUST match the Implementation Outline exactly for v1 and clearly mark all future extensions as unimplemented.

The spec MUST define:

- Behavior

- Flag semantics

- Exit codes

- Data formats

- Error conditions

No implementation work may begin until the Spec and outline match line by line.

⸻

### Enforcement Rules

1. AI MUST NOT write code before finishing Analysis Brief, Outline, and Spec alignment.

2. If implementation requires a change to behavior, AI MUST:

- Pause coding

- Update Analysis Brief and Outline

- Update Spec

- Then resume tests and code

3. All planned v1 behavior MUST be covered by tests.

4. All features MUST complete the lifecycle defined in spec/features.yaml.

⸻

### Summary of Required Artifacts Per Feature

| Stage | Required File | Purpose |
|-------|---------------|---------|
| Analysis | docs/analysis/<FEATURE_ID>.md | Defines why the feature exists |
| Outline | docs/<FEATURE_ID>_IMPLEMENTATION_OUTLINE.md | Defines what v1 will deliver |
| Spec | spec/<domain>/<feature>.md | Defines the contract the implementation must follow |
| Tests | multiple | Validate v1 behavior |
| Implementation | Go code | Implements the spec |

AI MUST follow this protocol for every feature in Stagecraft.

⸻

## 2. Feature ID Rules

All meaningful changes must reference a Feature ID:

```go
// Feature: CLI_INIT
// Spec: spec/commands/init.md
```

### 2.1 Feature and Spec Header Comments

All code files implementing a feature MUST include header comments linking to the Feature ID and spec:

```go
// Feature: <FEATURE_ID>
// Spec: spec/<domain>/<feature>.md
```

**Enforcement:**

- These comments MUST appear at the top of the file, after the license header
- They MUST be present in the main implementation file(s) for the feature
- They MUST be present in test files for the feature
- The Spec path MUST match the entry in `spec/features.yaml`

**Purpose:**

- Guarantees traceability from code → spec
- Enables automated validation of feature completeness
- Makes feature boundaries explicit in code
- Supports tooling for feature dependency analysis

**Example:**

```go
// SPDX-License-Identifier: AGPL-3.0-or-later
// ... license header ...

package commands

// Feature: CLI_PLAN
// Spec: spec/commands/plan.md

// NewPlanCommand returns the `stagecraft plan` command.
func NewPlanCommand() *cobra.Command {
    // ...
}
```

AI MUST add these comments when creating new feature implementations.

### Creating a Feature ID

Create a new Feature ID when:

* Adding user‑facing behaviour
* Adding a CLI command
* Adding a provider or migration engine
* Changing config schema with behavioural impact

Do NOT create a new Feature ID for:

* Refactors
* Bug fixes
* Docs‑only changes

### Feature ID Naming Rules

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

**Timing Rule:** AI MUST update the feature state in spec/features.yaml only when the implementation is complete, tests
pass, docs are updated, and the task is ready for commit. AI MUST NOT update feature states prematurely (e.g., before
tests pass or before implementation is complete).

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

#### Coverage Requirements

* All tests MUST achieve minimum coverage thresholds as enforced by `scripts/check-coverage.sh`.
* Coverage check MUST pass with `--fail-on-warning` flag before committing.
* Overall coverage must meet minimum thresholds (see `scripts/check-coverage.sh` for current thresholds).
* Core packages (e.g., `pkg/config`, `internal/core`) MUST meet higher coverage requirements.
* Coverage thresholds are enforced in CI and MUST be maintained locally.

AI MUST run `./scripts/check-coverage.sh --fail-on-warning` as part of pre-commit verification.

#### Golden Tests

Use golden tests for CLI output, config generation, or structured text.

Rules:

* Golden files must live in testdata/.
* Golden files are updated only when behaviour changes and the spec is explicitly updated. Golden file updates MUST
  correspond exactly to explicit spec changes, not incidental formatting changes or test refactoring.
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

* `go build ./...` must pass (all packages must compile).
* Format with gofumpt (MANDATORY - stricter than gofmt). AI MUST use `./scripts/goformat.sh`, which uses the pinned gofumpt version.
* All code MUST pass `gofumpt -l .` check. If gofumpt is not available, AI MUST install the version specified in `scripts/goformat.sh`.
* For non-Go changes (docs, YAML, etc.) gofumpt is not required, but the pre-commit hook must still pass.
* `go test ./...` must fully pass.
* All exported symbols must include GoDoc comments.
* Fix all golangci-lint warnings unless suppressed with justification:

```go
// nolint:gocritic // interface requires value
```

### Mocking Policy

* AI MUST NOT introduce new mocking frameworks without explicit approval.
* Generated mocks MUST be deterministic.
* If mocks are generated, their generator version and invocation MUST be documented in the spec or an ADR.

⸻

## 🔍 Pre-Commit Verification Requirements

Before committing, AI MUST ensure all changes pass the complete CI check suite.

> **Note**: For a complete list of scripts and their usage, see [scripts/README.md](scripts/README.md).

### Mandatory Pre-Commit Checks

AI MUST run `./scripts/run-all-checks.sh` before committing to verify all checks pass. This script
matches the CI workflow and ensures local validation matches what CI will enforce.

The following checks MUST pass:

1. **Formatting Checks**
   * For Go changes: AI MUST run `./scripts/goformat.sh` before staging files
   * `gofumpt -l .` must return no unformatted files
   * All Go files must be properly formatted with gofumpt
   * For non-Go changes, formatting is handled by the pre-commit hook

2. **Build Checks**
   * `go build ./...` must pass (all packages must compile)
   * Main binary must build: `go build -o stagecraft ./cmd/stagecraft`

3. **Test Checks**
   * `go test ./...` must fully pass
   * Coverage thresholds must be met (see Coverage Requirements below)
   * Coverage check must pass with `--fail-on-warning` flag

4. **Lint Checks**
   * `golangci-lint run ./...` must pass with no warnings (unless suppressed with justification)

5. **Spec Validation**
   * `spec/features.yaml` must be valid YAML
   * `scripts/validate-spec.sh` must pass
   * All referenced spec files must exist

6. **License Headers**
   * `addlicense -check .` must pass
   * All files must have proper license headers

If any check fails, AI MUST NOT commit until all checks pass. AI MUST report which checks failed and
why.

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

* Registration MUST occur via init() in the provider's own package.
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
Any directory reads (for example, using os.ReadDir) MUST be lexicographically sorted before processing to avoid
filesystem ordering differences.

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

**Collision Prevention:** All generated files under `.stagecraft/` MUST include deterministic content and deterministic
filenames. AI MUST NOT generate multiple files with overlapping roles without explicit instruction. Filenames MUST be
unique and descriptive of their purpose.

__CLI Command Names__

* MUST use dashed names: stagecraft deploy-plan.
* NEVER camelCase or snake_case.

__CLI Determinism__

* CLI help output MUST be deterministic.
* Command registration MUST use stable and lexicographically sorted ordering.
* CLI flags MUST be registered and rendered in a stable, lexicographically sorted order; no reliance on Cobra's implicit
  ordering is allowed.
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
* Errors MUST NOT be combined using errors.Join unless explicitly specified by the spec or an ADR, and the ordering MUST
  be deterministic.
* Errors must be deterministic and structured.
* A single error value MUST NOT be wrapped multiple times in the same return path unless explicitly specified by the
  spec or an ADR.
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
6. Branch Status – current branch name and git status output.
7. Commit Summary – after committing, show commit hash and verification.

**Note:** When a task involves code modification, the commit message generated under Git Workflow Rules (section "🧵 Git
Workflow Rules") MUST appear in section 5 of this contract. The commit message is part of the AI response, not a
separate post-response step.

If the task involves new behaviour:

8. Feature Reference – Feature ID and spec path.
9. Test Plan – list of failing tests to be written or updated.
10. Documentation Changes – list of sections to be updated.

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
* A human must approve before tests or code are produced, and AI MUST provide proposed wording as explicit Markdown
  snippets in its response (not written to files) for review.

⸻

## 🧵 Git Workflow Rules (Critical)

These rules govern all commits, branches, and PRs.

### 0. Pre-Work: Branch Setup

**Prerequisite:** Feature ID MUST be identified before branch setup (see AI Quickstart step 0).

**Prerequisite:** AI MUST complete Hook Verification steps above.

Before modifying any file:

1. Ensure working directory is clean:
   ```bash
   git status
   ```

2. Check current branch:
   ```bash
   git branch --show-current
   ```

3. If on `main`:
  - Checkout main:
    ```bash
    git checkout main
    ```
  - Pull latest main:
    ```bash
    git pull origin main
    ```
  - Create feature branch:
    ```bash
    git checkout -b feature/<FEATURE_ID>-short-desc
    ```

4. If on a feature branch:
  - Verify it matches current FEATURE_ID
  - If mismatch → STOP and ask for direction
  - If match → proceed with work (no branch creation needed)

5. Verify branch:
   ```bash
   git branch --show-current
   ```

If any step fails: STOP, report the issue, request direction.

⸻

### 1. Every task ends with a commit message

Each task should normally end with a single cohesive commit. If multiple commits are necessary, they must still follow the same rules and summarize the work clearly.

Steps AI MUST follow:

A. **Format Go files (if Go changes were made)**

   ```bash
   ./scripts/goformat.sh
   git add .
   ```

B. **Run pre-commit verification**

   ```bash
   ./scripts/run-all-checks.sh
   ```

   If any check fails, AI MUST NOT proceed with commit. AI MUST fix all issues and re-run the
   checks until all pass.

C. **Stage all changes**

   ```bash
   git add .
   ```

D. **Generate commit message in strict format**

    ```text
    <type>(<FEATURE_ID>): <short summary>
    
    Longer explanation if needed.
    Spec: spec/<...>.md
    Tests: <paths>
    ```

   Allowed commit types:

   - feat
   - fix
   - refactor
   - docs
   - test
   - ci
   - chore

   Constraints:

   - Subject ≤ 72 characters
   - No trailing periods
   - Body lines wrap at 80 chars

E. **Commit**

   ```bash
   git commit -m "<message>"
   ```

F. **Provide structured commit summary**

   AI MUST output in this format:

   ```text
   Commit Summary:
   - Feature: <FEATURE_ID>
   - Commit: <hash>
   - Branch: <branch-name>
   - Scope: <packages/files changed>
   - Changes:
     - <change 1>
     - <change 2>
   - Tests:
     - <test files/paths>
   - Spec: <spec file path>
   ```

G. **Create or update PR**

   AI MUST ensure a PR exists for the feature branch:

   - **PR Title MUST include:** `[FEATURE_ID] <Short human description>`
   - **PR Body MUST include:**
     - Feature ID
     - Summary of changes
     - Testing performed
     - Spec reference

   - **If AI has access to GitHub CLI:**
     - Propose: `gh pr create --title "[FEATURE_ID] <description>" --body "<body>" --draft`
     
   - **If AI does not have GitHub CLI access:**
     - Provide PR title and body that the human can paste into the GitHub UI
     - Clearly state: "Please create a PR with the following title and body:"

   - **AI MUST NOT:**
     - Transition PR to ready-for-review (only humans can do this)
     - Merge PRs

   - **PR Requirement:** Every feature branch MUST have a PR, even for small changes. This ensures all work is reviewed and tracked.

H. **Verify final state**

   AI MUST confirm:
   - No uncommitted changes remain (`git status` must be clean)
   - Commit meets formatting rules
   - Commit contains exactly the intended changes
   - PR exists or instructions provided

⸻

### 2. Post-Commit Verification

AI MUST:

1. Run:
   ```bash
   git status
   ```

2. Run:
   ```bash
   git log --oneline -1
   ```

3. Confirm:
  - No uncommitted changes remain
  - Commit meets formatting rules
  - Commit contains exactly the intended changes

4. Provide a verbal summary of committed work.

⸻

### 3. Each behavioural feature MUST be implemented in its own PR

PR Title

```text
[FEATURE_ID] <Short human description>
```

PR Description (MUST include):

* Feature: `<FEATURE_ID>`
* Spec: `spec/<...>.md`
* Tests: `<test paths>`
* Summary: `<what was changed>`
* Rationale: `<why this change>`
* Constraints: `<any limitations or considerations>`

PR Requirements

* Atomic, spec‑driven.
* No mixing multiple features.
* All tests pass.
* Spec updated.
* Docs updated.
* Feature status updated.

PR Metadata

* Must include a label (feature, fix, docs, test, ci, chore).
* Must be in draft until tests pass.
* Human reviewer required.
* Default target branch: main.

AI and PRs

* AI MUST create PR or provide PR title/body for human to create.
* AI MAY propose branch names and PR titles/descriptions.
* AI MUST NOT transition a PR to ready-for-review. Only humans can do this.
* AI MUST NOT merge PRs.

⸻

### 4. Branch Cleanup (Post-Merge)

AI MUST NOT delete branches.

AI MAY suggest:

```bash
git checkout main
git pull origin main
```

Human collaborator handles lifecycle management.

⸻

## 🔀 Git Integration Protocol

This consolidates the mandatory Git behavior for AI.

### Workflow Sequence

1. **Pre-Work**
  - Verify git availability
  - Check git status
  - Ensure clean workspace
  - Create feature branch from main (if needed)

2. **During Work**
  - Follow strict order:
    1. Read spec
    2. Write/update tests
    3. Implement code
    4. Update docs
  - Keep changes focused on a single feature
  - Never modify outside the defined scope

3. **Post-Work**
  - Stage all changes:
    ```bash
    git add .
    ```
  - Generate commit message
  - Commit work
  - Verify commit
  - Summarize results

### Integration with Feature Lifecycle

This workflow integrates with the Feature Lifecycle (see "Golden Rules" section):

1. Pre-Work (Git) → Feature Planning (Spec)
2. During Work → Test-First Development
3. Post-Work (Git) → Feature Status Update

The commit message MUST reference the Feature ID, ensuring traceability
from spec → tests → code → docs → git commit.

⸻

### Commit Message Requirements

AI MUST generate messages that:

- Use: `<type>(<FEATURE_ID>): <summary>`
- Contain spec and test references
- Pass subject/body length rules
- Clearly describe the intent of the change
- Map to exactly one feature

⸻

### 🔥 Commit Message Enforcement & Discipline

**Purpose:** Ensure all commits in Stagecraft follow deterministic, spec-linked, traceable commit message rules.

**Agent Mission:**

You are responsible for enforcing Stagecraft's strict commit message discipline. Your goal is to guarantee absolute traceability between:

**spec → tests → code → docs → commit → PR**

You MUST ensure commit messages:

- Follow required format
- Reference correct Feature ID
- Pass all structural rules
- Integrate with branch naming rules
- Reinforce test-first, spec-first, deterministic workflows

**Commit Format Rule (Mandatory):**

AI MUST generate commit messages using:

```
<type>(<FEATURE_ID>): <summary>
```

Where:

- `<type>` ∈ { `feat`, `fix`, `refactor`, `docs`, `test`, `ci`, `chore` }
- `<FEATURE_ID>` = SCREAMING_SNAKE_CASE
- `<summary>` = ≤72 chars, no trailing period, lowercase after colon

**If the Feature ID is missing → STOP and ask.**

**AI MUST Perform These Steps Before Committing:**

1. **Verify `.git/hooks/commit-msg` exists**
   - If missing → run `./scripts/install-hooks.sh`
   - If installation fails → STOP and report error

2. **Validate commit message against required pattern**
   - Format MUST be: `<type>(<FEATURE_ID>): <summary>`
   - Type MUST be lowercase
   - Feature ID MUST be SCREAMING_SNAKE_CASE
   - Summary MUST be ≤72 characters
   - Summary MUST NOT have trailing period
   - Summary MUST NOT start with capital letter after type

3. **Verify FEATURE_ID matches the active feature branch**
   - Check current branch: `git branch --show-current`
   - Extract FEATURE_ID from branch name
   - Ensure commit message FEATURE_ID matches branch FEATURE_ID
   - If mismatch → STOP and report

4. **Verify no protected files are touched**
   - Protected files: LICENSE, README.md, ADRs, NOTICE
   - If protected files modified → STOP and report

5. **Run all CI checks**
   - Execute: `./scripts/run-all-checks.sh`
   - All checks MUST pass before committing
   - If any check fails → STOP, fix issues, re-run

6. **Only then create commit message and commit**

**If any check fails: STOP and report.**

**Commit Quality Rules:**

AI MUST ensure:

- Subject ≤72 chars
- No trailing period
- No unicode decorations (emojis, fancy characters)
- Summary is literal, precise, minimal
- Message describes exactly the scoped changes
- References both spec and tests in the body

Example body:

```
Spec: spec/commands/deploy.md
Tests: cmd/deploy_test.go
```

**AI MUST Reject These Commit Messages:**

- Missing Feature ID
- Wrong format (missing parentheses, missing colon)
- Messages starting with uppercase after type
- Multi-feature changes
- Vague descriptions
- Long subjects (>72 chars)
- Commit-msg hook bypassing (using `STAGECRAFT_SKIP_HOOKS=1` or `SKIP_HOOKS=1`)

**Outcome:**

By following these rules:

- Every commit becomes a deterministic artifact
- Full traceability is maintained
- Specs and features map 1:1 to history
- Git hygiene and review quality improve
- AI assistants remain aligned with Stagecraft's engineering doctrine

⸻

### Branch Verification

AI MUST confirm:

```bash
git status
git branch --show-current
```

Requirements:

- Working directory clean
- On correct feature branch
- Ahead of main by expected number of commits

If mismatch: STOP and report.

⸻

### Error Handling

If any git command fails, AI MUST:

- Stop immediately
- Report exact error
- Suggest steps to resolve
- Wait for user direction

Proceeding after errors is not allowed.

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
8. Build tags MUST NOT be introduced or modified without explicit human approval and, if they affect behaviour, an ADR.

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

**Template:**

```yaml
---
feature: CLI_INIT
version: v1
status: wip
---

# CLI_INIT

[ Feature description ]
```

⸻

## 🧾 Logging Determinism Rules

Logs MUST:

* Use a structured format (JSON).
* Never include timestamps unless injected via deterministic clock.
* Include Feature ID when behaviour is feature‑specific.
* Avoid machine‑specific metadata.

__Structured Output Determinism__

* Any YAML, JSON, or other structured configuration output MUST be produced using a deterministic encoder.
* Keys MUST be ordered lexicographically at all levels where ordering is not semantically defined.
* Tests MUST NOT rely on map iteration order or encoder-specific non-determinism.

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

If the user writes "do X then Y":

* AI MUST stop after X.
* Wait for confirmation.
* Then perform Y.

This avoids premature implementation.

Multi-step interaction rules govern how AI and humans collaborate, not commit boundaries. Commits SHOULD follow the
commit granularity rules: typically a single commit per completed, single-feature change after all agreed steps for that
feature are done, unless the user explicitly requests separate commits for intermediate steps.

**Commit Boundary Rule for Multi-Step Tasks:**

Multi-step tasks MAY NOT require commits between steps unless they produce completed, valid, spec-compliant work.
Intermediate steps that produce incomplete or failing states MUST NOT be committed. Commit only after the full feature's
agreed multi-step sequence is complete or when the user explicitly requests an intermediate commit.

⸻

## 📚 Canonical Error Categories

* ErrInvalidConfig
* ErrProviderUnavailable
* ErrPlanFailed
* ErrRegistryConflict
* ErrFeatureIncomplete
* ErrSpecViolation

**Error Naming Rule:**

Canonical errors MUST be namespaced by the lowest-level package (e.g., `registry.ErrRegistryConflict`,
`config.ErrInvalidConfig`). Test assertions MUST reference the qualified identifier. This ensures clear error ownership
and prevents naming conflicts.

⸻

## 📦 Approved Dependencies List

Only the following external dependencies MAY be used without explicit approval:

* `github.com/spf13/cobra`
* `github.com/stretchr/testify`
* `gopkg.in/yaml.v3`

**Important:** The "etc." placeholder MUST NOT be interpreted as permission to introduce additional dependencies. Only
explicitly listed dependencies are allowed. All others require explicit human approval, justification, and ADR if
architectural.

⸻

## 📌 Toolchain Determinism Rule

All contributors must use the exact versions of:

* Go compiler.
* golangci-lint.
* Test harness.
* Build tools.

as defined in .tool-versions or go.mod.

Build tags and compilation flags MUST be documented and kept stable across environments.

⸻

## 🚫 Non‑Goals

* Stagecraft is NOT a general-purpose automation tool.
* Do not add speculative features without an ADR.
* All new behaviour must be anchored to a Feature ID and spec.
* Stagecraft is not a plugin framework or workflow engine.
* Stagecraft is not a general-purpose YAML/JSON transformer; any structured output MUST be directly justified by the
  spec.

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

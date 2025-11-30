# Agent Guide for Stagecraft

> Audience: AI assistants (Cursor, ChatGPT, Copilot, etc.) and human collaborators using them.

## Project Purpose

Stagecraft is a Go-based CLI that orchestrates application deployment and infrastructure workflows.
It aims to reimagine tools like Kamal with:

- A clean, composable core (planning, drivers, plugins).
- First-class developer UX for local and remote dev.
- Strong guarantees via tests, docs, and specs.

This repository is both a **production-grade tool** and a **public portfolio piece**. Code quality, reasoning, and documentation matter as much as functionality.

## Golden Rules

1. **Spec-first, test-first**
    - Before adding or changing behaviour, look at `spec/features.yaml` and the relevant spec in `spec/`.
    - Prefer writing or updating tests before implementation, especially in `internal/core` and `pkg/config`.

2. **Every change must trace to a feature ID**
    - Each meaningful change should be associated with a feature `id` from `spec/features.yaml` (e.g. `CLI_INIT`, `DEPLOY_PLAN`).
    - Add a short comment in new Go files referencing the feature ID and spec:
      ```go
      // Feature: CLI_INIT
      // Spec: spec/commands/init.md
      ```

3. **Tests and docs are non-optional**
    - When changing behaviour, also:
        - Update or add tests (`*_test.go`).
        - Update the relevant spec markdown in `spec/`.
        - Update docs in `docs/` if user-facing behaviour changed.
    - Do not mark a feature as `done` in `spec/features.yaml` unless tests and docs are in place.

4. **Respect boundaries and structure**
    - `internal/` contains implementation details; avoid exposing them as public APIs.
    - `pkg/` is for pieces that can be reused or imported externally.
    - `cmd/` should remain thin: wire CLI flags and arguments to `internal/` or `pkg/` logic, but not contain business logic.

5. **Do not modify certain files without explicit intent**
    - Only change these when the human author clearly asks for it:
        - `LICENSE`
        - `README.md` top-level positioning (minor edits ok, but no radical reframing)
        - Existing ADRs in `docs/adr/` (append follow-up ADRs instead of rewriting history).
    - If a modification to these files seems required to complete a task, clearly explain why in comments or commit messages.

6. **Follow Go style and quality standards**
    - Code should compile (`go build ./...`).
    - Code should be formatted with `gofmt`/`goimports`.
    - All tests must pass (`go test ./...`).
    - Address linter findings unless explicitly documented as exceptions.

7. **Provider and Engine Agnosticism**
    - **Never hardcode provider IDs or engine IDs in validation**
        - ❌ Wrong: `if provider != "encore-ts" && provider != "generic" { return error }`
        - ✅ Right: `if !backendproviders.Has(provider) { return error }`
        - Always validate against the appropriate registry (`pkg/providers/backend` or `pkg/providers/migration`)
    - **Provider-specific config belongs under `backend.providers.<id>`**
        - ❌ Wrong: `backend.dev.encore_secrets` (Encore-specific at top level)
        - ✅ Right: `backend.providers.encore-ts.dev.secrets` (provider-scoped)
        - Stagecraft core treats provider config as opaque (`map[string]any`)
    - **Encore.ts and Drizzle are implementations, not special cases**
        - Encore.ts is a `BackendProvider` implementation, not a core feature
        - Drizzle is a `MigrationEngine` implementation, not a core feature
        - If you need provider/engine-specific logic, it belongs in the provider/engine implementation
    - **When adding new providers or engines**
        - Register them in the appropriate registry (`backend.Register()` or `migration.Register()`)
        - Implement the interface (`BackendProvider` or `Engine`)
        - Add provider/engine-specific config under the scoped namespace
        - Do not modify core validation logic

## Workflow Expectations

When implementing or modifying a feature:

1. **Locate the feature in `spec/features.yaml`**
    - If it does not exist, add a new entry with `status: todo`.

2. **Review or create the feature spec**
    - Example: `spec/commands/init.md` for `CLI_INIT`.

3. **Update or add tests**
    - For core logic: unit tests in `internal/core/..._test.go`.
    - For CLI behaviour: tests in `internal/cli/commands/..._test.go` or golden file tests.
    - For end-to-end behaviour: tests in `test/e2e/` if needed.

4. **Implement code**
    - Keep functions small and focused.
    - Depend on interfaces where integration with external systems is involved.

5. **Run tests and tooling**
    - `go test ./...`
    - Lint (e.g. `golangci-lint run ./...` if configured).

6. **Update documentation and feature status**
    - Adjust relevant markdown in `spec/` and `docs/`.
    - Set the feature’s `status` in `spec/features.yaml` to `wip` or `done` as appropriate.

## Folder-Level Instructions

Some folders may include their own `Agent.md` with more specific guidance (e.g. for `internal/core` or `pkg/config`).  
When working in such folders, follow both the root `Agent.md` and the local one. If they conflict, defer to the human maintainer.

## Non-Goals

- Stagecraft is not intended to be a generic framework for arbitrary automation beyond deployment and infra workflows.
- Avoid speculative or experimental patterns unless justified in an ADR and wired to a clear feature.

When in doubt, favour clarity, simplicity, and traceability over cleverness.


---
feature: CLI_INIT
version: v1
status: done
domain: commands
inputs:
  flags: []
outputs:
  exit_codes:
    success: 0
    error: 1
---
# `stagecraft init` – Project Bootstrap Command

- Feature ID: `CLI_INIT`
- Status: todo

## Goal

Bootstrap Stagecraft into an existing project with minimal friction.

## User Story

As a developer,
I want to run `stagecraft init` in my project,
so that I get a minimal, valid Stagecraft configuration
and can start using other commands (`plan`, `deploy`, etc.) quickly.

## Behaviour

- When run with no existing config:
    - Creates a default Stagecraft config file (e.g. `stagecraft.yml`) in the current directory.
    - Asks a minimal set of interactive questions (with sensible defaults), such as:
        - project name
        - primary environment (e.g. `dev`, `staging`, `prod`)
        - initial provider/driver (e.g. `digitalocean`, `none`)

- When run with an existing config:
    - Validates the config.
    - Offers to:
        - print a summary; and/or
        - guide the user through an update / migration path (future enhancement).

## CLI Usage

- `stagecraft init`
- Options (future):
    - `--non-interactive` – generate config with defaults only.
    - `--config path/to/file` – choose an alternative config path.

## Outputs

- A config file written to the repo (default: `stagecraft.yml`).
- Informative CLI output indicating what was created or updated.

## Non-Goals (for initial version)

- No provider-specific configuration wizard (those can come later).
- No remote state setup.

## Tests

See `spec/features.yaml` entry for `CLI_INIT`:
- `internal/cli/commands/init_test.go` – unit/CLI behaviour tests.
- `test/e2e/init_smoke_test.go` – end-to-end smoke test.
- 
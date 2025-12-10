---
feature: GOV_CLI_EXIT_CODES
version: v1
status: todo
domain: governance
---

# GOV_CLI_EXIT_CODES

Governance rules for documenting and standardising CLI exit codes.

## Problem

Current CLI command specs do not consistently document exit codes. This makes it hard to:

- Reason about error handling in scripts and CI.
- Guarantee deterministic failure semantics.
- Align implementation with user facing contracts.

Warnings from spec validation already highlight missing exit codes for several CLI features:

- CLI_INIT
- CLI_GLOBAL_FLAGS
- CLI_PHASE_EXECUTION_COMMON
- CLI_DEV_BASIC
- CLI_DEPLOY
- CLI_ROLLBACK
- CLI_MIGRATE_BASIC
- CLI_RELEASES
- CLI_COMMIT_SUGGEST

## Goal

Define a governance standard for:

- How exit codes are documented in specs.
- How exit codes are chosen for new commands.
- How to gradually retrofit existing specs without breaking behaviour.

## Scope - v1

### Included

- Documentation rules for CLI exit codes in command specs.
- A common exit code structure for:
  - Success
  - User or config errors
  - External provider failures
  - Internal errors
- A migration plan for updating existing CLI specs.

### Excluded (v1)

- Behaviour changes to existing commands.
- Enforced numeric exit code mapping in code (that may come later as a follow up feature).

## Exit Code Documentation Rules

Every CLI command spec under `spec/commands/*.md` MUST:

1. Include a dedicated `## Exit Codes` section.
2. List all exit codes used by that command.
3. Provide a one line description for each exit code.

Example format:

```markdown
## Exit Codes

- `0`  - success
- `1`  - invalid user input or config
- `2`  - external provider failure (network, cloud, etc.)
- `3`  - internal error (unexpected panic, invariant violation)
```

Commands may define additional command specific exit codes, but they must be documented with the same format.

## Common Exit Code Semantics

For new commands and future refactors, the following semantics are recommended:

- `0` - success
- `1` - user error or invalid config
- `2` - external provider failure (Docker, cloud APIs, CI, etc.)
- `3` - internal error (bugs, invariant violations, unexpected panics)

Existing commands should be aligned with this structure where practical. Where behaviour already differs, specs must reflect the current behaviour and note any planned changes.

## Migration Plan

### Phase 1 - Documentation only:

- Update specs for the following existing CLI features to include exit codes:
  - CLI_INIT - `spec/commands/init.md`
  - CLI_GLOBAL_FLAGS - `spec/core/global-flags.md` (later moved to commands domain)
  - CLI_PHASE_EXECUTION_COMMON - `spec/core/phase-execution-common.md`
  - CLI_DEV_BASIC - `spec/commands/dev-basic.md`
  - CLI_DEPLOY - `spec/commands/deploy.md`
  - CLI_ROLLBACK - `spec/commands/rollback.md`
  - CLI_MIGRATE_BASIC - `spec/commands/migrate-basic.md`
  - CLI_RELEASES - `spec/commands/releases.md`
  - CLI_COMMIT_SUGGEST - `spec/commands/commit-suggest.md`
- Documentation must reflect current implementation as closely as possible.
- Where behaviour is unclear, the spec should mark the exit code as TBD with a note to validate in a later phase.

### Phase 2 - Behaviour alignment (future feature):

- Add tests that assert exit codes for key scenarios.
- Adjust implementations to match documented exit codes where needed.
- Update GOV_CLI_EXIT_CODES status to done once all targeted commands have:
  - Documented exit codes
  - Tests asserting exit codes
  - Implementations aligned with specs

## Determinism

Exit code usage must be deterministic:

- The same input and environment must always produce the same exit code.
- Multiple error conditions should be classified to the most specific applicable exit code.

## Tests

GOV_CLI_EXIT_CODES is primarily a documentation and governance feature. Tests are expected in two places:

- Spec validation:
  - Extend spec validation scripts to assert that all CLI command specs have an `## Exit Codes` section.
- Command tests:
  - For each CLI command, add tests that verify exit codes for:
    - Success path
    - Representative failure paths

These tests will be introduced as part of follow up features once this governance spec is approved.


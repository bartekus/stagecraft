<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<!--

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

-->

# Stagecraft Scripts

This directory contains utility scripts for development, testing, and maintenance tasks.

## Quick Reference

| Script | Purpose | When to Use |
|--------|---------|-------------|
| `run-all-checks.sh` | Run all CI checks locally | Before committing, before PR |
| `goformat.sh` | Format Go code | After making Go changes |
| `install-hooks.sh` | Install git hooks | First-time setup, after cloning |
| `new-feature.sh` | Generate feature skeleton | When starting a new feature |
| `add-headers.sh` | Add license headers | When adding new files |
| `check-header-comments.sh` | Verify license headers | Before committing |
| `check-coverage.sh` | Check test coverage | During development |
| `check-required-tests.sh` | Verify required tests exist | Before committing |
| `generate-cli-docs.sh` | Generate CLI reference docs | When CLI changes |
| `spec-sync-check.sh` | Verify spec/CLI alignment | Before committing |
| `validate-spec.sh` | Validate spec files | When editing specs |
| `validate-feature-integrity.sh` | Validate feature metadata | When editing features.yaml |
| `check-orphan-docs.sh` | Find orphan Analysis/Outline files | Periodically, before PR |
| `check-orphan-specs.sh` | Find orphan spec files | Periodically, before PR |
| `check-doc-patterns.sh` | Fail CI if new docs match forbidden patterns | In CI, before commit |
| `generate-evolution-log.sh` | Scaffold new provider evolution log | When creating new provider |
| `append-coverage-snapshot.sh` | Append coverage snapshot to ledger | In CI, after coverage runs |

## Essential Scripts

### `run-all-checks.sh`

**Purpose**: Runs all checks that match the CI workflow.

**Usage**:
```bash
./scripts/run-all-checks.sh
```

**What it does**:
- Formats Go code
- Runs linters
- Runs tests
- Checks coverage
- Validates specs
- Checks spec/CLI alignment
- Builds binaries

**When to use**: Before every commit, especially before creating a PR.

---

### `goformat.sh`

**Purpose**: Format all Go code using `gofumpt` and organize imports.

**Usage**:
```bash
./scripts/goformat.sh
```

**What it does**:
- Formats all `.go` files with `gofumpt`
- Organizes imports with `goimports`

**When to use**: After making any Go code changes, or let git hooks handle it automatically.

---

### `install-hooks.sh`

**Purpose**: Install git pre-commit hooks.

**Usage**:
```bash
./scripts/install-hooks.sh
```

**What it does**:
- Installs pre-commit hook that runs formatting and basic checks
- Ensures consistent code style across contributors

**When to use**: 
- First time setting up the repository
- After cloning the repository
- If hooks are missing or broken

**Note**: Required for all contributors. See [CONTRIBUTING.md](../CONTRIBUTING.md).

---

### `new-feature.sh`

**Purpose**: Generate a complete feature skeleton from templates.

**Usage**:
```bash
./scripts/new-feature.sh <FEATURE_ID> <DOMAIN> [feature-name]
```

**Examples**:
```bash
# Create a new CLI command
./scripts/new-feature.sh CLI_DEPLOY commands deploy

# Create a new provider
./scripts/new-feature.sh PROVIDER_BACKEND_NEW backend new-backend-provider
```

**What it creates**:
- `docs/engine/analysis/<FEATURE_ID>.md` - Analysis brief template
- `docs/engine/outlines/<FEATURE_ID>_IMPLEMENTATION_OUTLINE.md` - Implementation outline template
- `spec/<domain>/<feature>.md` - Spec file template
- Updates `spec/features.yaml` with new feature entry

**When to use**: When starting work on a new feature. See [Agent.md](../Agent.md) Feature Planning Protocol.

---

## Maintenance Scripts

### `add-headers.sh`

**Purpose**: Add license headers to files that are missing them.

**Usage**:
```bash
./scripts/add-headers.sh
```

**What it does**:
- Scans for files missing SPDX license identifiers
- Adds appropriate headers based on file type
- Uses full headers for entry files, short headers for others

**When to use**: When adding new files, or if headers are missing.

---

### `check-header-comments.sh`

**Purpose**: Verify all files have proper license headers.

**Usage**:
```bash
./scripts/check-header-comments.sh
```

**What it does**:
- Checks all source files for required headers
- Validates SPDX identifiers
- Ensures entry files have full headers

**When to use**: Before committing, or as part of CI checks.

---

### `check-coverage.sh`

**Purpose**: Check test coverage thresholds.

**Usage**:
```bash
./scripts/check-coverage.sh
```

**What it does**:
- Runs tests with coverage
- Checks if coverage meets minimum thresholds
- Reports coverage by package

**When to use**: During development to ensure adequate test coverage.

---

### `check-required-tests.sh`

**Purpose**: Verify that required test files exist.

**Usage**:
```bash
./scripts/check-required-tests.sh
```

**What it does**:
- Checks that implementation files have corresponding test files
- Validates test file naming conventions

**When to use**: Before committing to ensure tests are present.

---

## Documentation Scripts

### `generate-cli-docs.sh`

**Purpose**: Generate CLI reference documentation from Cobra commands.

**Usage**:
```bash
./scripts/generate-cli-docs.sh
```

**What it does**:
- Builds the stagecraft binary
- Extracts help text from all commands
- Generates `docs/reference/cli.md`

**When to use**: After adding or modifying CLI commands.

**Note**: The generated file should not be edited manually.

---

## Validation Scripts

### `spec-sync-check.sh`

**Purpose**: Verify that CLI implementation matches specifications.

**Usage**:
```bash
./scripts/spec-sync-check.sh
```

**What it does**:
- Compares CLI commands/flags with spec definitions
- Reports mismatches between spec and implementation
- Ensures ALIGNED discipline (spec-first development)

**When to use**: Before committing, especially after CLI changes.

---

### `validate-spec.sh`

**Purpose**: Validate specification files.

**Usage**:
```bash
./scripts/validate-spec.sh
```

**What it does**:
- Validates YAML structure in spec files
- Checks for required frontmatter
- Verifies spec file format

**When to use**: When editing spec files.

---

### `validate-feature-integrity.sh`

**Purpose**: Validate feature metadata in `spec/features.yaml`.

**Usage**:
```bash
./scripts/validate-feature-integrity.sh
```

**What it does**:
- Validates `spec/features.yaml` structure
- Checks feature dependencies
- Verifies feature IDs and status values

**When to use**: When editing `spec/features.yaml`.

---

### `check-orphan-docs.sh`

**Purpose**: Find Analysis and Outline files that don't have a matching feature entry in `spec/features.yaml`.

**Usage**:
```bash
./scripts/check-orphan-docs.sh
```

**What it does**:
- Scans `docs/engine/analysis/` for analysis files
- Scans `docs/engine/outlines/` for outline files
- Checks each file's Feature ID against `spec/features.yaml`
- Reports any files without matching entries

**When to use**: 
- Periodically to catch orphaned documentation
- Before creating a PR
- After removing features from `spec/features.yaml`

**What to do with orphans**:
- Remove if feature was cancelled
- Move to `docs/archive/` if feature is historical
- Add Feature ID to `spec/features.yaml` if feature is active

---

### `check-doc-patterns.sh`

**Purpose**: Fail CI if new documentation files match forbidden patterns that should use canonical docs instead.

**Usage**:
```bash
./scripts/check-doc-patterns.sh
```

**What it does**:
- Checks for forbidden patterns:
  - `*_COVERAGE_V1_COMPLETE.md` → Use `COVERAGE_LEDGER.md` and `<FEATURE_ID>_EVOLUTION.md`
  - `*_SLICE*_PLAN.md` → Use `<FEATURE_ID>_EVOLUTION.md`
  - `COMMIT_*_PHASE*.md` → Use `GOVERNANCE_ALMANAC.md`
- Only flags new files (not existing legacy files)
- Provides guidance on which canonical doc to use

**When to use**: 
- Automatically in CI (enforced as first-class check in `.github/workflows/docs-governance.yml`)
- Before committing documentation changes
- To enforce canonical documentation homes

**Integration**: 
- First-class CI check in `.github/workflows/docs-governance.yml` (un-skippable in CI)
- Included in `run-all-checks.sh` and `gov-pre-commit.sh`
- Can be bypassed locally with `STAGECRAFT_SKIP_DOC_PATTERNS=1` (blocked in CI)

---

### `generate-evolution-log.sh`

**Purpose**: Scaffold a new provider evolution log with standard structure and migration checklist.

**Usage**:
```bash
./scripts/generate-evolution-log.sh <FEATURE_ID>
```

**Example**:
```bash
./scripts/generate-evolution-log.sh PROVIDER_NETWORK_TAILSCALE
```

**What it does**:
- Creates `docs/engine/history/<FEATURE_ID>_EVOLUTION.md`
- Populates with standard sections:
  - Purpose and scope
  - Feature references
  - Design intent and constraints
  - Coverage timeline overview
  - Migration checklist
- Attempts to auto-detect spec and analysis file paths

**When to use**: 
- When starting a new provider that will have slices or coverage progression
- When establishing evolution tracking for an existing provider

**Next steps after generation**:
1. Review and populate sections with actual content
2. Migrate content from legacy docs (use migration checklist)
3. Mark legacy docs as superseded
4. Update cross-references in `COVERAGE_LEDGER.md` and `GOVERNANCE_ALMANAC.md`

---

### `append-coverage-snapshot.sh`

**Purpose**: Append coverage snapshot to `COVERAGE_LEDGER.md` in an append-only manner (never rewrites history).

**Usage**:
```bash
# Automatic (reads from go test -cover)
./scripts/append-coverage-snapshot.sh --event "Provider X slice 2 complete"

# Manual (specify values)
./scripts/append-coverage-snapshot.sh --event "Coverage snapshot" --overall 75.2 --core 82.1 --providers 73.5
```

**What it does**:
- Runs `go test -cover` to get current coverage (if not provided)
- Calculates overall, core, and providers coverage
- Appends a new row to the Historical Coverage Timeline table
- Never edits or rewrites existing entries (append-only)
- Uses current date automatically

**When to use**: 
- In CI after coverage runs
- After completing a slice or phase
- After significant coverage improvements

**Integration**: Can be added to CI workflows to automatically track coverage over time

**Note**: This script ensures the ledger remains an audit log - historical entries are never modified.

---

### `check-orphan-specs.sh`

**Purpose**: Find spec files that don't have a corresponding entry in `spec/features.yaml`.

**Usage**:
```bash
./scripts/check-orphan-specs.sh
```

**What it does**:
- Scans `spec/` directory for `.md` files
- Checks each spec file against `spec/features.yaml` entries
- Reports any spec files without matching entries

**When to use**:
- Periodically to catch orphaned specifications
- Before creating a PR
- After removing features from `spec/features.yaml`

**What to do with orphans**:
- Remove if feature was cancelled
- Move to `docs/archive/` if spec is historical
- Add entry to `spec/features.yaml` if feature is active

---

## Development Workflow

### Typical Workflow

1. **Start new feature**:
   ```bash
   ./scripts/new-feature.sh <FEATURE_ID> <DOMAIN> [name]
   ```

2. **During development**:
   ```bash
   ./scripts/goformat.sh          # Format code
   ./scripts/check-coverage.sh    # Check tests
   ```

3. **Before committing**:
   ```bash
   ./scripts/run-all-checks.sh    # Run all checks
   ```

4. **After CLI changes**:
   ```bash
   ./scripts/generate-cli-docs.sh # Update docs
   ```

### Pre-Commit Hooks

Most of these checks run automatically via git hooks (installed by `install-hooks.sh`). The hooks:
- Format code automatically
- Check headers
- Run basic validations

If hooks are installed, you typically only need to run `run-all-checks.sh` before creating a PR.

---

## Script Requirements

All scripts:
- Are bash scripts (`.sh`)
- Include proper license headers
- Use `set -e` or `set -euo pipefail` for error handling
- Are executable (`chmod +x`)

## Troubleshooting

**Script fails with "command not found"**:
- Ensure you're in the project root
- Check that required tools are installed (see script comments)

**Pre-commit hooks not running**:
- Run `./scripts/install-hooks.sh`
- Verify `.git/hooks/pre-commit` exists

**Coverage check fails**:
- Add more tests to increase coverage
- Check `scripts/check-coverage.sh` for threshold settings

For more information, see [CONTRIBUTING.md](../CONTRIBUTING.md) and [Agent.md](../Agent.md).


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


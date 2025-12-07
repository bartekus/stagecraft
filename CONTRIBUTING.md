<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<!--

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

-->

# Contributing to Stagecraft

Thank you for your interest in contributing to Stagecraft! This document outlines the contribution process and requirements.

## License Requirements

Stagecraft is licensed under the **GNU Affero General Public License v3 or later (AGPL-3.0-or-later)**. All contributions must comply with this license.

### License Header Requirements

Every source file in Stagecraft **must** include proper licensing and attribution. This is enforced automatically and is a requirement for all pull requests.

#### Required Elements

1. **SPDX License Identifier** (mandatory for all files)
   - Must be the first non-empty line in every source file
   - Format: `// SPDX-License-Identifier: AGPL-3.0-or-later` (for Go files)
   - Format: `# SPDX-License-Identifier: AGPL-3.0-or-later` (for shell scripts, YAML)
   - Format: `<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->` (for Markdown)

2. **Short Header** (for most source files)
   - Required for all `.go`, `.sh`, `.yaml`, `.yml`, `.json`, `.ts`, `.tsx` files
   - See examples below

3. **Full Header** (for entry files only)
   - Required only for:
     - `/cmd/stagecraft/main.go`
     - `/cmd/stagecraftd/main.go` (if exists)
     - `/internal/version/version.go` (if exists)
     - Any root-level entry point files

#### Examples

**Short Header (for most Go files):**
```go
// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package mypackage
```

**Full Header (for entry files):**
```go
// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft is a Go-based CLI for orchestrating local development and deployment of multi-service

applications. It aims to be "A local-first tool that scales from single-host to multi-host deployments

like Kamal, but for Docker Compose".

Copyright (C) 2025  Bartek Kus

This program is free software: you can redistribute it and/or modify it under the terms of the

GNU Affero General Public License as published by the Free Software Foundation, either version 3

of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without

even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU

Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License along with this program.

If not, see <https://www.gnu.org/licenses/>.

*/

package main
```

**Shell Script Header:**
```bash
#!/bin/bash
# SPDX-License-Identifier: AGPL-3.0-or-later
#
# Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.
#
# Copyright (C) 2025  Bartek Kus
#
# This program is free software licensed under the terms of the GNU AGPL v3 or later.
#
# See https://www.gnu.org/licenses/ for license details.
#
```

**Markdown Header:**
```markdown
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<!--

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

-->
```

### Automated License Checking

We use the [`addlicense`](https://github.com/google/addlicense) tool to enforce license headers:

1. **Pre-commit Hook**: Automatically checks license headers before each commit
2. **CI/CD**: GitHub Actions workflow validates all files on every PR
3. **Manual Check**: Run `addlicense -check .` to verify locally

### Adding Headers to New Files

Before committing new files:

1. **Install addlicense** (if not already installed):
   ```bash
   go install github.com/google/addlicense@latest
   ```

2. **Add headers automatically**:
   ```bash
   addlicense -c "Bartek Kus" -l agpl -y 2025 .
   ```

   Or use our helper script:
   ```bash
   ./scripts/add-headers.sh
   ```

3. **For entry files**, manually update to use the full header format (see examples above)

4. **Verify**:
   ```bash
   addlicense -check .
   ```

### PR Requirements

Every pull request must satisfy:

1. ✅ All files include the SPDX license identifier
2. ✅ Go/TS/YAML/etc files include the short header
3. ✅ Entry files include the full header
4. ✅ CI license-check workflow passes
5. ✅ No contributor may remove or alter the copyright holder
6. ✅ New files have headers added before commit

**PRs that fail the license check will not be merged.**

## Development Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/your-org/stagecraft.git
   cd stagecraft
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Install git hooks** (required):
   ```bash
   ./scripts/install-hooks.sh
   ```
   
   The pre-commit hook runs gofumpt and basic checks, and will block commits on formatting errors. See the "Git Hooks" section below for details.
   
   For a complete list of scripts and their usage, see [scripts/README.md](scripts/README.md).

4. **Run tests**:
   ```bash
   go test ./...
   ```

## Git Hooks

Stagecraft requires git hooks to be installed for all contributors. The pre-commit hook:

- Automatically formats Go files with gofumpt
- Organizes imports with goimports
- Adds license headers to new files
- Runs basic build checks

**Installation:**
```bash
./scripts/install-hooks.sh
```

**Verification:**
```bash
ls -la .git/hooks/pre-commit
```

If the hook is missing, formatting or basic checks may fail in CI and PRs will be blocked.

## Code Style

- Follow Go standard formatting (`gofmt`)
- Run `golangci-lint` before committing
- Write tests for new features
- Update documentation as needed

See [docs/README.md](docs/README.md) for a map of all project documentation.

## Submitting Changes

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/your-feature-name`
3. **Make your changes** (remember to add license headers!)
4. **Run tests and linters**: `go test ./... && golangci-lint run ./...`
5. **Commit your changes**: `git commit -m "Add feature: description"`
6. **Push to your fork**: `git push origin feature/your-feature-name`
7. **Open a Pull Request**

## AI-Assisted Development with Cursor

If you're using Cursor or other AI coding assistants, we have a dedicated guide for efficient AI workflows:

📖 **[Cursor Contributor Workflow Guide](docs/CONTRIBUTING_CURSOR.md)**

This guide covers:
- Thread hygiene (one feature per thread)
- File hygiene (what to open vs. attach)
- Using STRUC-C/L methodology in Cursor
- Cost-efficient AI usage patterns
- Examples of good vs. bad AI interactions

For a quick reference on which specs and docs to open for different feature types, see:

📖 **[Engine Documentation Index](docs/engine-index.md)**

## Questions?

If you have questions about contributing or the license requirements, please open an issue or contact the maintainers.

Thank you for contributing to Stagecraft! 🎭


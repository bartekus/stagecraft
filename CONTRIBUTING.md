<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<!--
Stagecraft - Human Contribution Guidelines
Copyright (C) 2025  Bartek Kus
Licensed under the GNU AGPL v3 or later.
-->

# Contributing to Stagecraft

> [!IMPORTANT]
> **This document defines the human contribution workflow.**
>
> It does **not** define enforcement rules for AI agents.
> *   **AI Agents**: STOP. Read [`Agent.md`](Agent.md) immediately.
> *   **Humans using AI**: Read [`Agent.md`](Agent.md) to understand the protocol your tool must follow.

## Core Principles

1.  **Spec-First**: No code without a spec. Update `spec/` before strictly implementing.
2.  **Test-First**: Write tests that fail before writing code that passes.
3.  **Deterministic**: No race conditions, no random outputs, no unseeded timestamps.
4.  **Governance-Aligned**: All features must be tracked in `spec/features.yaml`.

---

## License Requirements

Stagecraft is licensed under the **GNU Affero General Public License v3 or later (AGPL-3.0-or-later)**.

### License Headers

Every source file **must** include the SPDX header.
*   **Go**: `// SPDX-License-Identifier: AGPL-3.0-or-later`
*   **Shell/YAML**: `# SPDX-License-Identifier: AGPL-3.0-or-later`
*   **Markdown**: `<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->`

We use [`addlicense`](https://github.com/google/addlicense) and CI checks to enforce this.

---

## Development Setup

### 1. Prerequisites
*   Go 1.22+
*   Docker (for integration tests)
*   PostgreSQL (optional, for migration tests)

### 2. Clone & Build
```bash
git clone https://github.com/bartekus/stagecraft.git
cd stagecraft
go build ./cmd/stagecraft
```

### 3. Install Git Hooks (Mandatory)
We strictly enforce formatting and linting via git hooks.

```bash
./scripts/install-hooks.sh
```

**What the hook does:**
*   Runs `gofumpt` (stricter gofmt)
*   Runs `goimports`
*   Checks license headers
*   Blocks commits if checks fail

### 4. Verify Environment
```bash
./scripts/run-all-checks.sh
```
This runs the full CI suite locally (Tests, Lint, Build, Licenses).

---

## Workflow: Spec-First Development

**Do not start coding until you have updated the specification.**

### 1. Feature Registry
Check `spec/features.yaml`. Every non-trivial change must map to a `FEATURE_ID`.

### 2. Update Specs
*   **Existing Feature**: Update `spec/<domain>/<feature>.md`.
*   **New Feature**:
    1.  Create `docs/engine/analysis/<FEATURE_ID>.md`
    2.  Create `spec/<domain>/<feature>.md`
    3.  Register in `spec/features.yaml`

### 3. Branching Strategy
Naming convention: `type/feature-name` or `type/FEATURE_ID-name`
*   `feature/cli-dev-hot-reload`
*   `fix/deployment-race-condition`
*   `docs/update-readme`

### 4. Implementation
1.  Write failing tests in `internal/<pkg>/..._test.go`.
2.  Implement minimal code to pass tests.
3.  Ensure 100% determinism (no `map` iteration without sorting).

---

## Testing & Verification

*   **Unit Tests**: `go test ./...`
*   **Coverage**: `./scripts/check-coverage.sh` (Enforced strict thresholds)
*   **Specs**: `./scripts/validate-spec.sh` (Ensures `features.yaml` is valid)

---

## Submitting Changes

1.  **Commit**: descriptive messages.
2.  **Push**: `git push origin feature/...`
3.  **PR**: Open Pull Request against `main`.
    *   CI will run `run-all-checks.sh`.
    *   License check must pass.

---

## For AI Agents

If you are an AI agent (Cursor, Windsurf, Cline, etc.) operating on this repository:

**You are bound by the [Agent Protocol](Agent.md).**

Do not guess. Do not hallucinate specs. Follow the strict execution loop defined in `Agent.md`.

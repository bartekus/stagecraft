---
status: canonical
scope: meta
---

<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<!--

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

-->

# Stagecraft + Cursor Contributor Workflow

This guide explains how to use Cursor efficiently when working on Stagecraft,
so we keep AI costs low while preserving ALIGNED / STRUC-C discipline.

> **Note**: This guide is subordinate to [Agent.md](../Agent.md) – if anything conflicts, Agent.md wins.

The high-level principles:

- One **feature per thread**

- Keep **only relevant files open**

- Prefer **attachments over giant context blobs**

- Use **spec → tests → state → runtime → composition → chain** as the working order

---

## 1. Thread Hygiene

### 1.1 One thread per feature / change

Create a new Cursor chat for each feature, bugfix, or refactor:

- Use a name like:  

  `FTR-CLI_BUILD: Implement build command`  

  `CORE_STATE: Improve state consistency checks`  

  `CLI_PLAN: Refine plan behavior`

Avoid "eternal" threads; they accumulate massive context and spike token use.

### 1.2 Thread lifecycle

For each feature:

1. **Spec Phase**

   - Attach the relevant spec(s), do not paste them inline:

     - `spec/commands/<command>.md`

     - `spec/core/*.md`

     - `spec/governance/GOV_CORE.md`

   - Ask the model to *summarize and confirm understanding*.

2. **Tests Phase**

   - Open only:

     - The relevant `_test.go` files under:

       - `internal/cli/commands`

       - `internal/core`

       - `internal/tools`

       - `pkg/*`

     - `test/e2e/*` if needed.

   - Let Cursor suggest tests, but you control which ones to actually create / keep.

3. **Code Phase**

   - Open only the Go files directly impacted by the feature:

     - `internal/cli/commands/*.go`

     - `internal/core/*.go` + subdirs

     - `pkg/**` as needed.

   - Explicitly tell Cursor:  

     > Only modify the open files, do not touch any other files.

4. **Integration / Chain Phase**

   - Use a **fresh thread** when the previous one gets long / noisy.

   - Attach a short summary (or the context-handoff doc) instead of dragging old chat history forward.

When a feature is implemented and tests pass — close the thread. Do not reuse it for the next feature.

---

## 2. File Hygiene (What to Open vs Attach)

### 2.1 Files to open during normal work

Open only the files you are actively editing for this feature:

- **Specs**

  - `spec/commands/*.md`

  - `spec/core/*.md`

  - `spec/governance/GOV_CORE.md`

- **Code**

  - `internal/cli/root.go`

  - `internal/cli/commands/*.go`

  - `internal/core/*.go` and subpackages (`env`, `state`, etc.)

  - `internal/providers/**`

  - `pkg/**`

- **Tests**

  - Matching `_test.go` files in `internal/**` and `pkg/**`

  - `test/e2e/*` if you're touching user-visible behavior.

- **Docs (as needed)**

  - `docs/CLI_*_ANALYSIS.md`

  - `docs/CLI_*_IMPLEMENTATION_OUTLINE.md`

  - `docs/context-handoff/*.md`

  - `docs/narrative/stagecraft-spec.md`

  - `docs/features/OVERVIEW.md`

  - `spec/features.yaml`

### 2.2 Files to avoid opening (let `.cursorignore` handle most)

Avoid opening:

- Golden files and fixtures:

  - `internal/cli/commands/testdata/*`

  - `internal/compose/testdata/*`

- Generated / binary / transient:

  - `bin/*`

  - `coverage.out`

  - `.DS_Store`

- Non-essential docs for core engine work:

  - `blog/**`

  - `discussions/**`

  - `examples/**` (unless you're specifically working on examples)

You can still manually open these if needed, but try not to keep them open in general feature work.

---

## 3. Using STRUC-C/L in Cursor

For each feature, explicitly walk Cursor through STRUC-C/L:

1. **Spec**

   - Open/attach only the relevant spec(s).

   - Ask: "Restate this spec in your own words, list behavior and edge cases."

2. **Tests**

   - Open `<thing>_test.go`.

   - Ask: "Given the spec above, propose tests; do not edit code yet."

3. **State**

   - For stateful features, open:

     - `spec/core/state*.md`

     - `internal/core/state/state.go`

     - `internal/core/state/state_test.go`

   - Ask for state transitions, invariants, and failure modes.

4. **Runtime**

   - Open only the core implementation files being changed.

   - Ask for code changes (in clear patches) that satisfy the spec + tests.

5. **Composition**

   - Open relevant `internal/providers/**`, `pkg/providers/**`, `internal/compose/**`.

   - Ask for a short summary of how this feature composes with existing providers / registries / CLI commands.

6. **Chain**

   - Point Cursor to the relevant context-handoff doc(s) in `docs/context-handoff`.

   - Ask it to verify the chain: spec → tests → code → state → providers → CLI.

---

## 4. Golden Files & E2E Tests

- Golden files live in:

  - `internal/cli/commands/testdata/*`

  - `internal/compose/testdata/*`

Workflow:

1. Ask Cursor to **describe** the change needed in a golden file, not to regenerate all goldens.

2. If needed, let Cursor propose updated content for a *single* golden file.

3. Run the tests locally:

   - `go test ./...`

   - Or use `scripts/run-all-checks.sh`

   > **Note**: See [scripts/README.md](../../scripts/README.md) for a complete list of scripts and their usage.

Avoid letting Cursor bulk-edit golden/testdata files across the repo in one go.

---

## 5. Examples of "Good" vs "Bad" Interactions

### Good

> I'm working on CLI_BUILD. Opened:

> - spec/commands/build.md  

> - internal/cli/commands/build.go  

> - internal/core/phases_build.go  

> - internal/cli/commands/build_test.go  

> 

> Please review the spec and tests, then suggest a minimal patch to `build.go` only.

### Bad

> Here's the whole repo tree; redesign build, deploy, rollback, and plan behavior together.

The "bad" example forces Cursor to load and reason over the entire tree, exploding token usage.

---

By following this guide, Stagecraft development with Cursor remains:

- ALIGNED (spec-first, test-aware)

- STRUC-C/L-compliant

- Cost-efficient enough to scale out to multiple contributors using AI heavily


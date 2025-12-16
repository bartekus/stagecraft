# Context Log

> Rolling log of context handoff notes for AI assisted work.
> This document replaces per task context handoff files.

## 1. Purpose and Scope

The Context Log:

- Captures multi step task context for AI agents
- Records how features and phases were handed over between tasks
- Provides a chronological view of how complex work was coordinated
- Avoids the need for many small context handoff files

It replaces separate files such as:

- `docs/context-handoff/COMMIT_DISCIPLINE_PHASE3*.md`
- `docs/context-handoff/CLI_*` to `CLI_*` chains
- `docs/context-handoff/CORE_STATE_*` chains
- `docs/context-handoff/GOV_STATUS_ROADMAP_COMPLETE.md`
- Other similar context handoff docs

All new context handoff notes should be added here as dated entries.

---

## 2. Usage Rules

- Each new context handoff gets its own subsection
- Use ISO like dates for titles: `YYYY-MM-DD`
- Keep entries focused on:
  - What was done
  - What is next
  - Constraints and invariants
- Avoid duplicating full specs or code; link them instead

---

## 3. Index of Entries

> Keep this short and high level. Add entries as you go.

- 2025-01-XX - GOV_STATUS_ROADMAP Complete (section 4.2)
- 2025-XX-XX - COMMIT_DISCIPLINE_PHASE3 (section 4.3)
- 2025-XX-XX - CORE_STATE to CLI_DEPLOY (section 4.4)
- 2025-XX-XX - CLI_DEV and DEV_HOSTS integration context
- 2025-XX-XX - PROVIDER_NETWORK_TAILSCALE Slice 2 handoff
- ...

---

## 4. Entries

### 4.1 2025-12-XX - Example Entry Template

> Use this as a template for new entries.

**Context label:** SHORT_NAME_FOR_CONTEXT  
**Related features:**

- `FEATURE_ID_1`
- `FEATURE_ID_2` (if cross feature)

**Source docs:**

- `spec/...`
- `docs/engine/analysis/...`
- `docs/engine/outlines/...`
- `docs/engine/history/...` (if relevant)

**Current state:**

- What has already been implemented
- What tests and docs already exist
- What decisions are locked in

**Next steps for the agent:**

- Step 1:
- Step 2:
- Step 3:

**Constraints and guardrails:**

- No behavioural changes outside feature scope
- No changes to protected files
- Spec and outline must not drift

**Notes:**

- Any clarifications, caveats, or follow up items

---

### 4.2 2025-01-XX - GOV_STATUS_ROADMAP Complete

**Context label:** GOV_STATUS_ROADMAP Complete  
**Related features:**

- `GOV_STATUS_ROADMAP`
- `PROVIDER_FRONTEND_GENERIC`

**Source docs:**

- `spec/commands/status-roadmap.md`
- `docs/engine/analysis/GOV_STATUS_ROADMAP.md`
- `docs/engine/outlines/GOV_STATUS_ROADMAP_IMPLEMENTATION_OUTLINE.md`

**Current state:**

- GOV_STATUS_ROADMAP feature implemented and complete (PR #21)
- Roadmap analysis engine in `internal/tools/roadmap/` with comprehensive tests
- CLI command `cortex status roadmap` functional
- PROVIDER_FRONTEND_GENERIC coverage improved from 70.2% → 80.2%
- All tests passing, linting clean

**Next steps:**

- Review PR feedback if needed
- Address any CI issues
- Future: GOV_STATUS_ROADMAP enhancements (v2), PROVIDER_FRONTEND_GENERIC Phase 2 coverage improvements

**Constraints:**

- No changes to protected files
- Spec and outline must not drift

**Notes:**

- Coverage improvements focused on error paths for `runWithReadyPattern`, `runWithShutdown`, `shutdownProcess`
- All code quality fixes applied (linting issues resolved)

---

### 4.3 2025-XX-XX - COMMIT_DISCIPLINE_PHASE3

**Context label:** Commit Discipline Phase 3  
**Related features:**

- `PROVIDER_FRONTEND_GENERIC`
- `GOV_CORE`

**Source docs:**

- `docs/context-handoff/COMMIT_DISCIPLINE_PHASE3.md`
- `docs/context-handoff/COMMIT_DISCIPLINE_PHASE3B.md`
- `docs/context-handoff/COMMIT_DISCIPLINE_PHASE3C.md`

**Current state:**

- Phase 1 (local + hooks) and Phase 2 (CI + CLI) completed
- CI validation and `stagecraft validate-commit` functional

**Next steps:**

- Phase 3.A: Implement commit report Go types and golden tests
- Phase 3.B: Implement commit health generators and CLI integration
- Phase 3.C: Wire CLI commands for commit discipline reports

**Constraints:**

- All logic MUST be deterministic and testable
- No history rewriting
- No automatic fixes (read-only analysis)

**Notes:**

- Phase 3 elevates commit discipline from validation to intelligent insight
- Focus on feature-aware commit suggestions and historical commit health analysis

---

### 4.4 2025-XX-XX - CORE_STATE to CLI_DEPLOY

**Context label:** CORE_STATE Complete  
**Related features:**

- `CORE_STATE`
- `CLI_DEPLOY`

**Source docs:**

- `spec/core/state.md`
- `spec/commands/deploy.md`

**Current state:**

- CORE_STATE implemented, fully tested, and merged (PR #3)
- State manager with deterministic behavior, 88.5% test coverage
- APIs available: CreateRelease, GetRelease, GetCurrentRelease, ListReleases, UpdatePhase

**Next steps:**

- Implement CLI_DEPLOY command
- Follow tests-first workflow
- Use CORE_STATE APIs for release management

**Constraints:**

- All work scoped strictly to CLI_DEPLOY
- Tests must be written before implementation code
- No changes to protected files

**Notes:**

- Dependencies ready: CORE_STATE, CORE_PLAN, CORE_COMPOSE
- PROVIDER_NETWORK_TAILSCALE not required for first pass

---

## 5. Archived Source Documents

The following sections contain references to previously scattered context handoff documentation files, preserved here for historical reference. Original files have been moved to `docs/archive/context-handoff/`.

### 5.1 Commit Discipline Phase 3 Documents

- **Phase 3**: `docs/context-handoff/COMMIT_DISCIPLINE_PHASE3.md` → `docs/archive/context-handoff/`
- **Phase 3A**: `docs/context-handoff/COMMIT_REPORT_TYPES_PHASE3.md` → `docs/archive/context-handoff/`
- **Phase 3B**: `docs/context-handoff/COMMIT_DISCIPLINE_PHASE3B.md` → `docs/archive/context-handoff/`
- **Phase 3C**: `docs/context-handoff/COMMIT_DISCIPLINE_PHASE3C.md` → `docs/archive/context-handoff/`

[Full content preserved in archived files - see section 4.3 for summary]

### 5.2 Feature Handoff Chains

- **CORE_STATE to CLI_DEPLOY**: `docs/context-handoff/CORE_STATE-to-CLI_DEPLOY.md` → `docs/archive/context-handoff/`
- **CORE_STATE_TEST_ISOLATION to CORE_STATE_CONSISTENCY**: `docs/context-handoff/CORE_STATE_TEST_ISOLATION-to-CORE_STATE_CONSISTENCY.md` → `docs/archive/context-handoff/`
- **CLI_DEPLOY to CLI_RELEASES**: `docs/context-handoff/CLI_DEPLOY-to-CLI_RELEASES.md` → `docs/archive/context-handoff/`
- **CLI_RELEASES to CLI_ROLLBACK**: `docs/context-handoff/CLI_RELEASES-to-CLI_ROLLBACK.md` → `docs/archive/context-handoff/`
- **CLI_ROLLBACK to CLI_BUILD**: `docs/context-handoff/CLI_ROLLBACK-to-CLI_BUILD.md` → `docs/archive/context-handoff/`
- **CLI_ROLLBACK to CLI_PHASE_EXECUTION_COMMON**: `docs/context-handoff/CLI_ROLLBACK-to-CLI_PHASE_EXECUTION_COMMON.md` → `docs/archive/context-handoff/`
- **CLI_PHASE_EXECUTION_COMMON to CORE_STATE_TEST_ISOLATION**: `docs/context-handoff/CLI_PHASE_EXECUTION_COMMON-to-CORE_STATE_TEST_ISOLATION.md` → `docs/archive/context-handoff/`
- **CLI_ROLLBACK-CLI_RELEASES-CORE_STATE_CONSISTENCY to CLI_DEPLOY**: `docs/context-handoff/CLI_ROLLBACK-CLI_RELEASES-CORE_STATE_CONSISTENCY-to-CLI_DEPLOY.md` → `docs/archive/context-handoff/`
- **GOV_STATUS_ROADMAP Complete**: `docs/context-handoff/GOV_STATUS_ROADMAP_COMPLETE.md` → `docs/archive/context-handoff/`
- **GOV_CORE to FRONTMATTER**: `docs/context-handoff/GOV_CORE-to-FRONTMATTER.md` → `docs/archive/context-handoff/`

[Full content preserved in archived files - see relevant sections above for summaries]

---

## 6. Migration Notes

- [x] Migrated COMMIT_DISCIPLINE_PHASE3*.md
- [x] Migrated GOV_STATUS_ROADMAP_COMPLETE.md
- [x] Migrated CLI_* to CLI_* chains
- [x] Migrated CORE_STATE_* chains
- [x] Archived all source files to `docs/archive/context-handoff/`

Migration complete. All context handoff documentation is now consolidated in this log.

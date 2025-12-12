> **Superseded by** `docs/context-handoff/CONTEXT_LOG.md`. New context handoffs MUST be added to the context log. Kept for historical reference.

# Context Handoff Documents

This directory contains handoff documents that provide deterministic context for AI agents transitioning between feature implementations.

## Purpose

These documents serve as a **complete, self-contained context** for the next agent session, ensuring:

- ✅ Clear understanding of what was just completed
- ✅ Explicit next task with all dependencies listed
- ✅ Mandatory workflows (tests-first, spec-first)
- ✅ Canonical constraints (what MUST and MUST NOT be done)
- ✅ Architectural context for understanding design decisions
- ✅ Output expectations and verification checklists

## Template Structure

Each handoff document follows this structure:

### Layer 1: What Just Happened
- Feature ID, status, PR/commit references
- What now exists (APIs, files, coverage)
- Files created/updated

### Layer 2: Immediate Next Task
- Feature ID and status
- Dependencies (with readiness indicators)
- Scope reminders
- **Mandatory workflow** (tests-first)
- Implementation outline
- **Constraints** (canonical list of MUST/MUST NOT)

### Layer 3: Secondary Tasks
- Related features that should NOT be started yet
- Dependencies and prerequisites

### Architectural Context
- Why design decisions matter
- Integration pattern examples
- Reference implementations

### Output Expectations
- Commit message format
- Verification checklist
- Test requirements

### Quick Start for Next Agent
- Bootloader instructions
- File reading order
- Step-by-step workflow

### Final Checklist
- Pre-work verification items

## Creating a New Handoff Document

When completing a feature, create a handoff document using this template:

1. **Copy the template structure** from `TEMPLATE.md` (or use `CORE_STATE-to-CLI_DEPLOY.md` as a reference example)

2. **Update Layer 1** with:
   - Completed feature details
   - PR number and commit hash
   - APIs/files created
   - Test coverage achieved

3. **Update Layer 2** with:
   - Next feature ID
   - Dependencies (mark readiness: ✅ ready, ⏸ todo, ❌ blocked)
   - Specific implementation requirements
   - Updated constraints list

4. **Update Layer 3** with:
   - Related features that depend on the next task
   - Features that should NOT be started yet

5. **Update Quick Start** with:
   - Specific files to read
   - Feature branch name
   - Test file locations

6. **Naming convention**: `{COMPLETED_FEATURE}-to-{NEXT_FEATURE}.md`

## Principles

These documents follow Stagecraft's Agent.md principles:

- **Spec-first**: Reference spec locations explicitly
- **Test-first**: Mandatory workflow section enforces tests before code
- **Feature-bounded**: Clear scope reminders and constraints
- **Deterministic**: Complete context, no ambiguity

## Example Usage

```bash
# After completing a feature
# Option 1: Use the generic template
cp docs/context-handoff/TEMPLATE.md \
   docs/context-handoff/CURRENT_FEATURE-to-NEXT_FEATURE.md

# Option 2: Use the example as a reference
cp docs/context-handoff/CORE_STATE-to-CLI_DEPLOY.md \
   docs/context-handoff/CLI_DEPLOY-to-CLI_ROLLBACK.md

# Then edit the new document, replacing all placeholders with actual values:
# - <CURRENT_FEATURE_ID> → actual feature ID
# - <NEXT_FEATURE_ID> → next feature ID
# - <PR_NUMBER>, <COMMIT_HASH>, etc.
```

## Notes

- Documents are **timeless** - no meta-commentary about creation
- Documents are **complete** - next agent needs no additional context
- Documents are **deterministic** - clear constraints prevent scope creep
- Documents are **actionable** - specific file paths, APIs, and workflows


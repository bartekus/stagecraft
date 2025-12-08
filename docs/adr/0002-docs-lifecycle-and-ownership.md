# 0002 – Documentation Lifecycle and Ownership

- Status: Accepted
- Date: 2025-12-07

## Context

Stagecraft maintains extensive documentation across multiple purposes:
- **Specifications** (`spec/`) - Source of truth for feature behavior
- **Implementation guides** - Analysis briefs, outlines, status tracking
- **Narrative docs** - Roadmaps, architecture overviews, planning
- **Governance docs** - Process, discipline, workflow guides
- **Historical docs** - Completed work, superseded implementations

Without clear structure and lifecycle management, documentation can:
- Drift from specifications
- Accumulate orphaned files
- Create confusion about what's canonical vs historical
- Make it unclear where new docs should live

We need a formal model for:
- Where different types of docs belong
- How docs transition through lifecycle states
- Who owns what
- How to prevent drift and orphan accumulation

## Decision

We adopt a structured documentation model with explicit lifecycle states and automated enforcement.

### Documentation Structure

```
docs/
├── engine/              # Implementation-aligned, AI-critical
│   ├── analysis/       # Feature analysis briefs (docs/engine/analysis/<FEATURE_ID>.md)
│   ├── outlines/        # Implementation outlines (docs/engine/outlines/<FEATURE_ID>_IMPLEMENTATION_OUTLINE.md)
│   ├── status/          # Generated status tracking
│   └── engine-index.md  # File location guide
├── narrative/           # Human-facing planning and overview
│   ├── architecture.md
│   ├── implementation-roadmap.md
│   ├── stagecraft-spec.md (index)
│   └── ...
├── governance/         # Process, discipline, workflow
│   ├── CONTRIBUTING_CURSOR.md
│   ├── COMMIT_MESSAGE_ANALYSIS.md
│   └── ...
└── archive/            # Historical, superseded docs
    └── ...
```

### Lifecycle States

All non-spec documentation uses frontmatter to track lifecycle:

```yaml
---
status: active | canonical | archived
scope: v1 | v2 | meta
feature: CLI_PLAN          # optional
spec: ../spec/commands/plan.md  # optional
superseded_by: ../Agent.md # optional
---
```

**Status meanings:**
- **active** - Guiding current or imminent work (e.g., implementation outlines for features in progress)
- **canonical** - Stable, long-lived reference (e.g., architecture.md, governance docs)
- **archived** - Historical "how we did it" (work complete, truth lives in spec/tests/code)

### Ownership Model

**Source of Truth Hierarchy:**
1. **`spec/`** - Behavioral truth (what the system does)
2. **`docs/engine/`** - Implementation truth (how to build it)
3. **`docs/narrative/`** - Planning truth (why and when)
4. **`docs/governance/`** - Process truth (how we work)
5. **`docs/archive/`** - Historical record (how we did it)

**Ownership Rules:**
- **Spec files** - Owned by feature implementers, must match implementation
- **Analysis/Outlines** - Owned by feature implementers, created during planning
- **Narrative docs** - Owned by project maintainers, updated as roadmap evolves
- **Governance docs** - Owned by project maintainers, updated as process evolves
- **Archived docs** - Frozen, never edited (only moved to archive/)

### Drift Prevention

**Automated Enforcement:**
- `implementation-status.md` is generated from `spec/features.yaml` (CI enforced)
- `stagecraft-spec.md` is an index (no spec duplication)
- CI checks for legacy path patterns (former docs/analysis path, now docs/engine/analysis, and old outline patterns)
- Scripts validate feature integrity and detect orphans

**Manual Processes:**
- When feature is complete: Move analysis/outline to `archive/` if historical value
- When spec changes: Regenerate `implementation-status.md`
- When process changes: Update governance docs, mark old versions as archived

### Orphan Detection

Two validation scripts ensure consistency:

1. **`check-orphan-docs.sh`** - Finds analysis/outline files without matching `spec/features.yaml` entries
2. **`check-orphan-specs.sh`** - Finds spec files without matching `spec/features.yaml` entries

These should be run periodically and can be integrated into CI.

## Alternatives Considered

1. **Flat docs structure**
   - Pros: Simpler navigation
   - Cons: No clear separation of purpose, harder to know what's canonical

2. **No lifecycle tracking**
   - Pros: Less overhead
   - Cons: Docs accumulate, unclear what's current vs historical

3. **Manual drift prevention only**
   - Pros: No automation complexity
   - Cons: Drift inevitable, requires constant vigilance

## Consequences

**Positive:**
- Clear separation of concerns (engine vs narrative vs governance)
- Automated enforcement prevents common drift patterns
- Lifecycle states make doc purpose explicit
- Orphan detection catches inconsistencies early
- AI assistants have clear guidance on what to open

**Negative:**
- Requires discipline to maintain frontmatter
- Scripts need periodic updates as structure evolves
- Some overhead in moving docs through lifecycle

We accept these trade-offs in favor of long-term documentation quality and consistency.

## Related Decisions

- [0001 – Stagecraft Architecture and Project Structure](./0001-architecture.md) - Established `spec/` and `docs/` structure
- `docs/governance/STRATEGIC_DOC_MIGRATION.md` - Strategic document handling
- `docs/README.md` - Complete documentation navigation guide


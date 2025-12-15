# Docs-to-Generate Backlog

Prioritized list of new documentation pages to generate under `docs/generated/ai-agent/`. Sources are restricted to the four canonical generated inputs.

1. `COMMAND_CATALOG.md`
   - Purpose: Central index of all CLI command specifications.
   - Sources: `SPEC_CATALOG.md` -> `Files` table; filter paths starting with `spec/commands/`.
   - Rule: Sort by `Path` ascending.
2. `PROVIDER_CATALOG.md`
   - Purpose: Central index of all provider specifications and interfaces.
   - Sources: `SPEC_CATALOG.md` -> `Files` table; filter paths starting with `spec/providers/`.
   - Rule: Group by provider domain (directory after `spec/providers/`); within each group sort by `Path`.
3. `CORE_SPEC_INDEX.md`
   - Purpose: Focused index of core engine and architecture specs.
   - Sources: `SPEC_CATALOG.md` -> `Files` table; filter paths under `spec/core/` and `spec/adr/`.
   - Rule: ADR group first then core group; within each group sort by `Path`.
4. `GOVERNANCE_INDEX.md`
   - Purpose: Single entry point for governance and ADR documentation.
   - Sources: `DOCS_CATALOG.md` -> `Files by Directory`; include `docs/governance/` and `docs/adr/`.
   - Rule: Group by directory; within each group sort by path.
5. `ENGINE_ANALYSIS_INDEX.md`
   - Purpose: Index of problem statements and engineering analyses.
   - Sources: `DOCS_CATALOG.md` -> `Files by Directory`; include `docs/engine/analysis/`.
   - Rule: Sort by path.
6. `ENGINE_OUTLINE_INDEX.md`
   - Purpose: Inventory of implementation outlines.
   - Sources: `DOCS_CATALOG.md` -> `Files by Directory`; include `docs/engine/outlines/`.
   - Rule: Sort by path.
7. `CONTEXT_HANDOFF_INDEX.md`
   - Purpose: Traceable list of context-handoff documents.
   - Sources: `DOCS_CATALOG.md` -> `Files by Directory`; include `docs/context-handoff/`.
   - Rule: Sort by path.
8. `COVERAGE_INDEX.md`
   - Purpose: Aggregated view of coverage planning and reports.
   - Sources: `DOCS_CATALOG.md` -> `Files by Directory`; include `docs/coverage/`.
   - Rule: Sort by path.
9. `ARCHIVE_INDEX.md`
   - Purpose: Visibility into archived and historical documents.
   - Sources: `DOCS_CATALOG.md` -> `Files by Directory`; include `docs/archive/` and `docs/engine/history/`.
   - Rule: Group by directory; within each group sort by path.
10. `GUIDES_AND_NARRATIVES_INDEX.md`
    - Purpose: Discovery index for guides and narrative documentation.
    - Sources: `DOCS_CATALOG.md` -> `Files by Directory`; include `docs/guides/` and `docs/narrative/`.
    - Rule: Group by directory; within each group sort by path.

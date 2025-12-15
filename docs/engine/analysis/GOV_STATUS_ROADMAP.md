# GOV_STATUS_ROADMAP Feature Analysis Brief

This document captures the high level motivation, constraints, and success definition for GOV_STATUS_ROADMAP.

It is the starting point for the Implementation Outline and Spec.

This brief must be approved before outline work begins.

⸻

## 1. Problem Statement

Stagecraft governance requires a deterministic, always-current view of feature completion across all implementation phases (0-10). Currently, the phase-level completion analysis exists only as a manually maintained document (`docs/engine/status/feature-completion-analysis.md`), which:

- Requires manual updates when `spec/features.yaml` changes
- Can drift from the source of truth
- Lacks integration with governance tooling
- Cannot be used programmatically by dev agents or CI

This creates governance risk: strategic planning decisions may be based on stale or inaccurate completion data.

⸻

## 2. Motivation

### Governance Compliance

- **GOV_CORE Requirement**: The governance core now references `feature-completion-analysis.md` as the authoritative implementation audit. This document must be automatically generated to maintain governance integrity.
- **Deterministic State**: Governance tools must operate on machine-readable, always-current data, not manual snapshots.
- **CI Integration**: CI workflows need programmatic access to completion metrics for blocking merges, generating reports, and tracking progress.

### Developer Experience

- **Strategic Planning**: Developers and project managers need instant visibility into phase completion and critical-path blockers.
- **Feature Prioritization**: Dev agents and human developers need to know which features to work on next based on dependencies and phase completion.
- **Progress Tracking**: Clear metrics on overall project completion (currently ~51% complete) help set expectations and plan sprints.

### Multi-Agent Workflow

- **AI Agent Prerequisite**: Dev agents must check completion status before choosing which feature to implement next.
- **Externalized Memory**: The analysis document serves as externalized strategic memory that persists across agent sessions.
- **Dependency Validation**: Agents can validate feature dependencies against completion status before starting work.

⸻

## 3. Goals

### Primary Goal

Provide a deterministic CLI command (`stagecraft status roadmap`) that:

1. Reads `spec/features.yaml` as the source of truth
2. Computes phase-level completion statistics
3. Identifies critical-path blockers (features blocked by incomplete dependencies)
4. Generates a markdown document identical in structure to the current manual analysis
5. Overwrites `docs/engine/status/feature-completion-analysis.md` with deterministic output

### Secondary Goals

- Enable programmatic access to completion metrics (future: JSON output format)
- Support CI integration (exit codes, deterministic output)
- Provide human-readable strategic planning artifact
- Maintain backward compatibility with existing analysis document structure

⸻

## 4. Constraints

### Technical Constraints

- **Deterministic Output**: Output must be identical across runs (sorted keys, stable formatting, no timestamps)
- **Source of Truth**: Must read from `spec/features.yaml` only, no other sources
- **Phase Mapping**: Must correctly map features to phases based on comments in `spec/features.yaml`
- **Dependency Analysis**: Must traverse `depends_on` fields to identify blockers

### Governance Constraints

- **GOV_CORE Alignment**: Must align with governance core requirements for implementation tracking
- **File Location**: Output must be written to `docs/engine/status/feature-completion-analysis.md` (referenced by GOV_CORE)
- **No Breaking Changes**: Generated document must maintain same structure as current manual version

### Operational Constraints

- **No External Dependencies**: Should not require network access or external APIs
- **Fast Execution**: Should complete in < 1 second for typical `spec/features.yaml` size
- **Error Handling**: Must fail gracefully with clear error messages if `spec/features.yaml` is invalid

⸻

## 5. Success Criteria

### Functional Success

1. ✅ Command reads `spec/features.yaml` successfully
2. ✅ Computes accurate phase completion percentages
3. ✅ Identifies all critical-path blockers correctly
4. ✅ Generates markdown identical to current manual analysis structure
5. ✅ Output is deterministic (identical across runs)
6. ✅ Overwrites target file successfully

### Quality Success

1. ✅ Golden test validates markdown output format
2. ✅ Unit tests validate phase calculation logic
3. ✅ Integration test validates CLI command execution
4. ✅ Test coverage meets targets (80%+ for core logic, 70%+ for CLI)
5. ✅ No linter errors

### Governance Success

1. ✅ Command can be integrated into CI workflows
2. ✅ Generated document is referenced correctly by GOV_CORE
3. ✅ Dev agents can use output for feature prioritization
4. ✅ Manual analysis document can be replaced by generated version

⸻

## 6. Out of Scope (v1)

- **JSON Output Format**: v1 focuses on markdown generation only
- **Interactive Mode**: No TUI or interactive exploration
- **Historical Tracking**: No comparison with previous completion states
- **Velocity Metrics**: No calculation of completion rate over time
- **Custom Phase Definitions**: Phases are hardcoded based on `spec/features.yaml` comments
- **Multi-Repository Analysis**: Single repository only

⸻

## 7. Dependencies

### Required Features

- **GOV_CORE**: Governance core provides the framework for this tool
- **CORE_CONFIG**: Config loading utilities may be used (if needed)
- **Internal Tools**: `internal/tools/features` provides YAML parsing utilities

### Blocking Features

None. This feature can be implemented independently and enhances governance tooling.

### Enables Features

- **Future Dev Agent Features**: Provides strategic memory for AI agents
- **CI Integration**: Enables automated progress tracking in CI
- **Strategic Planning Tools**: Foundation for future planning and roadmap tools

⸻

## 8. Risks

### Technical Risks

- **Phase Detection**: Features are mapped to phases via comments in `spec/features.yaml`. If comment format changes, phase detection may break.
  - **Mitigation**: Use explicit phase detection logic, validate against known phase names
- **Dependency Cycles**: Circular dependencies in `depends_on` could cause infinite loops
  - **Mitigation**: Detect cycles and report as errors, don't attempt to resolve

### Governance Risks

- **Output Format Drift**: If manual analysis document structure changes, generated version may not match
  - **Mitigation**: Use golden tests to lock output format, update tests when structure changes intentionally
- **Stale References**: If GOV_CORE references change, this tool may generate incorrect paths
  - **Mitigation**: Validate file paths exist, fail fast with clear errors

⸻

## 9. Future Enhancements

- **JSON Output Format**: Add `--format=json` flag for programmatic access
- **Historical Comparison**: Compare current state with previous runs
- **Velocity Metrics**: Calculate completion rate over time
- **Custom Phase Views**: Allow filtering by phase or status
- **Dependency Graph Visualization**: Generate DOT format for graph visualization
- **Integration with CI**: Auto-generate on `spec/features.yaml` changes

⸻

## 10. Approval

This analysis brief must be approved before proceeding to implementation outline.

**Next Steps:**
1. Review and approve this brief
2. Generate implementation outline (`docs/engine/outlines/GOV_STATUS_ROADMAP_IMPLEMENTATION_OUTLINE.md`)
3. Generate spec (`spec/commands/status-roadmap.md`)
4. Implement feature following spec-first, TDD workflow

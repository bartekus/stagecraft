---
feature: GOV_STATUS_ROADMAP
version: v1
status: todo
domain: commands
inputs:
  flags: []
outputs:
  exit_codes:
    success: 0
    validation_failed: 1
    internal_error: 2
---

# GOV_STATUS_ROADMAP

`stagecraft status roadmap` command that generates a phase-level feature completion analysis document from `spec/features.yaml`.

## Overview

The `stagecraft status roadmap` command reads `spec/features.yaml`, computes phase-level completion statistics, identifies critical-path blockers, and generates a deterministic markdown document at `docs/engine/status/feature-completion-analysis.md`.

This command enables:
- Governance oversight (referenced by GOV_CORE)
- Strategic planning (phase completion visibility)
- Dev-agent prioritization (blocker identification)
- CI integration (programmatic progress tracking)

## Command Structure

### Parent Command

If `CLI_STATUS` is not yet implemented, create a minimal stub `status` command that contains only the `roadmap` subcommand.

### Subcommand

```bash
stagecraft status roadmap
```

**Usage:**
```bash
stagecraft status roadmap
```

**Description:**
Generates a phase-level feature completion analysis document from `spec/features.yaml`.

**Flags:**
- None for v1 (all behavior is deterministic)

**Exit Codes:**
- `0`: Success, document generated
- `1`: Validation error (invalid `spec/features.yaml`, file I/O error)
- `2`: Internal error (parsing failure, unexpected error)

## Behavior

### Input

- **Source File**: `spec/features.yaml` (relative to repository root)
- **Format**: YAML with feature definitions and phase comments

### Processing

1. **Parse YAML**: Read and parse `spec/features.yaml` using `internal/tools/features` utilities
2. **Detect Phases**: Map features to phases based on comments in YAML:
   - `# Architecture & Documentation` → Architecture phase
   - `# Phase 0: Foundation` → Phase 0
   - `# Phase 1: Provider Interfaces` → Phase 1
   - `# Phase 2: Core Orchestration` → Phase 2
   - `# Phase 3: Local Development` → Phase 3
   - `# Phase 4: Provider Implementations` → Phase 4
   - `# Phase 5: Build and Deploy` → Phase 5
   - `# Phase 6: Migration System` → Phase 6
   - `# Phase 7: Infrastructure` → Phase 7
   - `# Phase 8: Operations` → Phase 8
   - `# Phase 9: CI Integration` → Phase 9
   - `# Phase 10: Project Scaffold` → Phase 10
   - `# Governance` → Governance phase
3. **Calculate Statistics**: For each phase:
   - Count total features
   - Count done/wip/todo features
   - Calculate completion percentage
4. **Identify Blockers**: For each `todo`/`wip` feature:
   - Check `depends_on` fields
   - If any dependency is not `done`, mark as blocked
5. **Generate Markdown**: Create markdown document with:
   - Executive summary
   - Phase-by-phase completion table
   - Roadmap alignment analysis
   - Priority recommendations
   - Critical path analysis
   - Next steps

### Output

- **Output File**: `docs/engine/status/feature-completion-analysis.md` (relative to repository root)
- **Format**: Markdown
- **Deterministic**: Output must be identical across runs (sorted keys, stable formatting, no timestamps)

## Output Format

### Document Structure

The generated markdown follows this structure:

1. **Header**: Title and metadata
2. **Executive Summary**: Overall statistics (total features, completion percentage)
3. **Phase-by-Phase Completion Table**: Detailed breakdown per phase
4. **Roadmap Alignment**: Analysis of alignment with roadmap
5. **Priority Recommendations**: Next steps based on blockers
6. **Detailed Phase Analysis**: Per-phase breakdown with feature lists
7. **Critical Path Analysis**: Blocker dependencies
8. **Next Steps**: Prioritized action items

### Formatting Rules

- Phases sorted numerically (0-10, then Architecture, Governance)
- Feature IDs sorted alphabetically within phases
- Consistent table formatting
- No timestamps or generation metadata
- Stable spacing and alignment

### Example Output

```markdown
# Feature Completion Analysis

> **Source**: Generated from `spec/features.yaml` by `stagecraft status roadmap`
> **Last Updated**: See `spec/features.yaml` for the source of truth

## Executive Summary

- **Total Features**: 73
- **Completed**: 37 (50.7%)
- **In Progress**: 1 (1.4%)
- **Planned**: 35 (47.9%)

## Phase-by-Phase Completion

| Phase | Features | Done | WIP | Todo | Completion | Status |
|-------|----------|------|-----|------|------------|--------|
| Architecture & Docs | 2 | 0 | 0 | 2 | 0% | ⚠️ Not started |
| Phase 0: Foundation | 8 | 8 | 0 | 0 | 100% | ✅ Complete |
...
```

## Error Handling

### Validation Errors (Exit Code 1)

- **Invalid YAML**: `spec/features.yaml` has syntax errors
  - Error message: `invalid YAML syntax in spec/features.yaml: <error>`
- **Missing File**: `spec/features.yaml` does not exist
  - Error message: `spec/features.yaml not found: <path>`
- **File I/O Error**: Cannot read input or write output
  - Error message: `failed to <read|write> <file>: <error>`
- **Invalid Status**: Feature has invalid status value
  - Error message: `invalid status for feature <id>: <status> (expected: done|wip|todo)`

### Internal Errors (Exit Code 2)

- **Parsing Failure**: YAML parsing fails unexpectedly
  - Error message: `failed to parse spec/features.yaml: <error>`
- **Unexpected Structure**: YAML structure doesn't match expected format
  - Error message: `unexpected YAML structure: <details>`
- **Panic Recovery**: Unexpected panic (should not happen)
  - Error message: `internal error: <panic details>`

## Testing

### Unit Tests

- **Phase Detection**: Test phase detection from comments
- **Statistics Calculation**: Test per-phase and overall statistics
- **Blocker Detection**: Test blocker identification logic
- **Markdown Generation**: Golden test for markdown output

### Integration Tests

- **CLI Execution**: Test command execution end-to-end
- **File Generation**: Test output file creation
- **Error Handling**: Test error cases and exit codes

### Test Files

- `internal/tools/roadmap/phase_test.go`
- `internal/tools/roadmap/stats_test.go`
- `internal/tools/roadmap/generator_test.go`
- `internal/cli/commands/status_test.go`

### Golden Test Data

- Input: `testdata/features.yaml` (sample features.yaml)
- Expected: `testdata/feature-completion-analysis.md.golden`
- Validation: Generated markdown must match golden file exactly

## Dependencies

### Internal Dependencies

- `internal/tools/features`: YAML parsing utilities
- `internal/cli/commands`: Command structure patterns

### External Dependencies

- `gopkg.in/yaml.v3`: YAML parsing (already in use)
- Standard library: `fmt`, `os`, `path/filepath`, `sort`, `strings`

## Implementation Notes

### Phase Detection

Features are mapped to phases based on YAML comments. The parser must:
1. Track the current phase while iterating through features
2. Update phase when encountering a phase comment
3. Map each feature to the current phase
4. Handle features without explicit phase (map to "Uncategorized")

### Deterministic Output

To ensure deterministic output:
- Sort phases numerically (0-10, then Architecture, Governance)
- Sort feature IDs alphabetically within phases
- Use consistent table formatting
- Avoid timestamps or random elements
- Use stable string formatting

### Blocker Detection

A feature is a blocker if:
1. It has `status: todo` or `status: wip`
2. It has `depends_on` fields
3. Any dependency has `status: todo` or `status: wip`

The blocker detection must traverse the dependency graph to identify all blocked features.

## Future Enhancements

- `--format=json` flag for programmatic access
- `--output=<path>` flag to specify output file
- `--phase=<phase>` flag to filter by phase
- Historical comparison with previous runs
- Velocity metrics (completion rate over time)
- Dependency graph visualization (DOT format)

## References

- **Analysis Brief**: `docs/engine/analysis/GOV_STATUS_ROADMAP.md`
- **Implementation Outline**: `docs/engine/outlines/GOV_STATUS_ROADMAP_IMPLEMENTATION_OUTLINE.md`
- **Governance Reference**: `spec/governance/GOV_CORE.md` (section 4.5)
- **Feature Tracking**: `spec/features.yaml`

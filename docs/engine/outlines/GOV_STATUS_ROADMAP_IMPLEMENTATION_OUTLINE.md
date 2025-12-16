# GOV_STATUS_ROADMAP Implementation Outline

> This document defines the v1 implementation plan for GOV_STATUS_ROADMAP. It translates the feature analysis brief into a concrete, testable, spec aligned delivery plan.

> All details in this outline must be reflected in `spec/commands/status-roadmap.md` before any tests or code are written.

⸻

## 1. Feature Summary

**Feature ID:** GOV_STATUS_ROADMAP

**Domain:** commands

**Goal:**

Provide a deterministic CLI command (`stagecraft status roadmap`) that generates a phase-level feature completion analysis document from `spec/features.yaml`, enabling governance oversight, strategic planning, and dev-agent prioritization.

**v1 Scope:**

- CLI command: `cortex status roadmap`
- Reads `spec/features.yaml` as source of truth
- Computes phase completion statistics (done/wip/todo counts per phase)
- Identifies critical-path blockers (features with incomplete dependencies)
- Generates markdown document: `docs/engine/status/feature-completion-analysis.md`
- Deterministic output (sorted keys, stable formatting, no timestamps)
- Golden test for markdown output validation
- Unit tests for phase calculation logic
- Integration test for CLI command execution

**Out of scope for v1:**

- JSON output format
- Interactive mode or TUI
- Historical tracking or comparison
- Velocity metrics
- Custom phase definitions
- Multi-repository analysis

**Future extensions (not implemented in v1):**

- `--format=json` flag for programmatic access
- Historical comparison with previous runs
- Velocity metrics (completion rate over time)
- Custom phase filtering
- Dependency graph visualization (DOT format)
- CI integration hooks

⸻

## 2. Command Structure

### Command Hierarchy

```
stagecraft
  └── status (parent command, may be stub for v1)
      └── roadmap (subcommand)
```

**Note:** If `CLI_STATUS` (the parent `status` command) is not yet implemented, create a minimal stub command that only contains the `roadmap` subcommand. The stub should follow the same pattern as other parent commands (e.g., `gov`).

### Command Signature

```bash
stagecraft status roadmap [flags]
```

**Flags:**

- None for v1 (all behavior is deterministic based on `spec/features.yaml`)

**Future flags (v2):**

- `--format=markdown|json` (default: markdown)
- `--output=<path>` (default: `docs/engine/status/feature-completion-analysis.md`)
- `--phase=<phase>` (filter by phase)

**Exit Codes:**

- `0`: Success, document generated
- `1`: Validation error (invalid `spec/features.yaml`, file I/O error)
- `2`: Internal error (parsing failure, unexpected error)

⸻

## 3. Phase Detection Logic

### Phase Mapping

Features are organized into phases based on comments in `spec/features.yaml`:

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

### Implementation Strategy

1. Parse `spec/features.yaml` using `internal/tools/features` utilities
2. Track current phase while iterating through features
3. Map each feature to its phase based on preceding comment
4. Group features by phase for statistics calculation

### Edge Cases

- Features without explicit phase comment → Map to "Uncategorized" phase
- Multiple phase comments → Use the last comment before the feature
- Features before any phase comment → Map to "Uncategorized"

⸻

## 4. Statistics Calculation

### Per-Phase Statistics

For each phase, calculate:

- **Total Features**: Count of all features in phase
- **Done**: Count of features with `status: done`
- **WIP**: Count of features with `status: wip`
- **Todo**: Count of features with `status: todo`
- **Completion Percentage**: `(done / total) * 100`

### Overall Statistics

- **Total Features**: Sum across all phases
- **Done**: Sum of done features across all phases
- **WIP**: Sum of wip features across all phases
- **Todo**: Sum of todo features across all phases
- **Overall Completion**: `(done / total) * 100`

### Status Legend

- ✅ Complete (100% done)
- 🔄 In Progress (has wip features)
- ⚠️ Not Started (0% done)

⸻

## 5. Critical-Path Blocker Detection

### Blocker Identification

A feature is a critical-path blocker if:

1. It has `status: todo` or `status: wip`
2. It has `depends_on` fields
3. Any dependency has `status: todo` or `status: wip`

### Implementation Strategy

1. Build dependency graph from `depends_on` fields
2. For each `todo`/`wip` feature, check if all dependencies are `done`
3. If any dependency is not `done`, mark feature as blocked
4. Report blockers in analysis document

### Blocker Reporting

- List blocked features by phase
- Show dependency chain for each blocker
- Prioritize blockers that block the most downstream features

⸻

## 6. Markdown Output Format

### Document Structure

The generated markdown must match the current manual analysis structure:

1. **Header**: Title and metadata
2. **Executive Summary**: Overall statistics
3. **Phase-by-Phase Completion Table**: Detailed breakdown
4. **Roadmap Alignment**: Analysis of alignment with roadmap
5. **Priority Recommendations**: Next steps based on blockers
6. **Detailed Phase Analysis**: Per-phase breakdown
7. **Critical Path Analysis**: Blocker dependencies
8. **Next Steps**: Prioritized action items

### Formatting Rules

- **Deterministic Sorting**: Phases sorted numerically (0-10, then Architecture, Governance)
- **Stable Formatting**: Consistent spacing, alignment, table formatting
- **No Timestamps**: Document should not include generation time
- **Sorted Keys**: Feature IDs sorted alphabetically within phases
- **Consistent Status Icons**: Use same emoji/status indicators throughout

### Example Structure

```markdown
# Feature Completion Analysis

## Executive Summary

- **Total Features**: 73
- **Completed**: 37 (50.7%)
- **In Progress**: 1 (1.4%)
- **Planned**: 35 (47.9%)

## Phase-by-Phase Completion

| Phase | Features | Done | WIP | Todo | Completion | Status |
|-------|----------|------|-----|------|------------|--------|
| Phase 0: Foundation | 8 | 8 | 0 | 0 | 100% | ✅ Complete |
| Phase 1: Provider Interfaces | 6 | 6 | 0 | 0 | 100% | ✅ Complete |
...

## Priority Recommendations

### Immediate (Unblocks Other Work)

1. Complete Phase 3:
   - Finish `CLI_DEV` (wip)
   - Implement `DEV_HOSTS` (todo)
...
```

⸻

## 7. Implementation Components

### Core Components

1. **Phase Detector** (`internal/tools/roadmap/phase.go`)
   - Parses `spec/features.yaml` comments to detect phases
   - Maps features to phases
   - Handles edge cases (uncategorized features)

2. **Statistics Calculator** (`internal/tools/roadmap/stats.go`)
   - Calculates per-phase statistics
   - Calculates overall statistics
   - Identifies critical-path blockers

3. **Markdown Generator** (`internal/tools/roadmap/generator.go`)
   - Generates markdown document from statistics
   - Ensures deterministic formatting
   - Handles edge cases (empty phases, etc.)

4. **CLI Command** (`internal/cli/commands/status.go`)
   - Creates `status` parent command (stub if CLI_STATUS not done)
   - Creates `roadmap` subcommand
   - Handles file I/O and error reporting

### File Structure

```
internal/
  cli/
    commands/
      status.go              # Parent status command + roadmap subcommand
      status_test.go         # Integration tests
  tools/
    roadmap/
      phase.go               # Phase detection logic
      phase_test.go          # Unit tests for phase detection
      stats.go               # Statistics calculation
      stats_test.go          # Unit tests for statistics
      generator.go           # Markdown generation
      generator_test.go      # Golden tests for markdown output
      model.go               # Data structures (PhaseStats, Blocker, etc.)
```

⸻

## 8. Testing Strategy

### Unit Tests

1. **Phase Detection Tests** (`phase_test.go`)
   - Test phase detection from comments
   - Test edge cases (no comments, multiple comments)
   - Test feature-to-phase mapping

2. **Statistics Tests** (`stats_test.go`)
   - Test per-phase calculation
   - Test overall statistics
   - Test blocker detection logic

3. **Generator Tests** (`generator_test.go`)
   - Golden test for markdown output
   - Test deterministic formatting
   - Test edge cases (empty phases, all done, etc.)

### Integration Tests

1. **CLI Command Tests** (`status_test.go`)
   - Test command execution
   - Test file generation
   - Test error handling (invalid YAML, missing file, etc.)
   - Test exit codes

### Golden Test Data

- **Input**: `testdata/features.yaml` (sample features.yaml)
- **Expected Output**: `testdata/feature-completion-analysis.md.golden`
- **Validation**: Generated markdown must match golden file exactly

⸻

## 9. Error Handling

### Validation Errors (Exit Code 1)

- Invalid YAML syntax in `spec/features.yaml`
- Missing `spec/features.yaml` file
- File I/O errors (permission denied, disk full, etc.)
- Invalid feature status values (not `done`/`wip`/`todo`)

### Internal Errors (Exit Code 2)

- YAML parsing failures
- Unexpected data structure
- Panic recovery (should not happen, but handle gracefully)

### Error Messages

- Clear, actionable error messages
- Include file paths and line numbers where possible
- Suggest fixes for common errors

⸻

## 10. Dependencies

### Internal Dependencies

- `internal/tools/features`: YAML parsing utilities
- `internal/cli/commands`: Command structure patterns

### External Dependencies

- `gopkg.in/yaml.v3`: YAML parsing (already in use)
- Standard library: `fmt`, `os`, `path/filepath`, `sort`, `strings`

### No New Dependencies

This feature should not introduce new external dependencies beyond what's already in use.

⸻

## 11. Implementation Phases

### Phase 1: Core Logic (TDD)

1. Implement phase detection logic with tests
2. Implement statistics calculation with tests
3. Implement blocker detection with tests
4. All unit tests passing

### Phase 2: Markdown Generation

1. Implement markdown generator
2. Create golden test with sample data
3. Validate output matches expected format
4. Golden test passing

### Phase 3: CLI Integration

1. Create `status` parent command (stub)
2. Create `roadmap` subcommand
3. Wire up core logic to CLI
4. Integration tests passing

### Phase 4: Validation & Polish

1. Error handling for all edge cases
2. Exit code validation
3. Documentation updates
4. Linter compliance

⸻

## 12. Success Metrics

### Functional Metrics

- ✅ Command executes successfully
- ✅ Generated document matches golden test
- ✅ All unit tests pass
- ✅ All integration tests pass
- ✅ Test coverage: 80%+ core logic, 70%+ CLI

### Quality Metrics

- ✅ No linter errors
- ✅ Deterministic output (identical across runs)
- ✅ Clear error messages
- ✅ Fast execution (< 1 second)

### Governance Metrics

- ✅ Generated document referenced by GOV_CORE
- ✅ Can be integrated into CI workflows
- ✅ Dev agents can use output for prioritization

⸻

## 13. Approval

This implementation outline must be approved before proceeding to spec generation.

**Next Steps:**
1. Review and approve this outline
2. Generate spec (`spec/commands/status-roadmap.md`)
3. Implement feature following spec-first, TDD workflow

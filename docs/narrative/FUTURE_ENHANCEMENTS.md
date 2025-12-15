# Future Enhancements: Next-Level Feature Pipeline

> **Status**: Mixed - Thin slice implemented (GOV_CORE), full enhancements remain future work
> **Purpose**: Document advanced enhancements that would bring Stagecraft's feature pipeline to "Google-grade" governance levels
> **Last Updated**: 2025-12-07

---

## Overview

Stagecraft now has a **thin-slice governance core** (GOV_CORE) that provides:
- ✅ Machine-verifiable spec schema (YAML frontmatter)
- ✅ Structural diff for CLI flags (spec vs implementation)
- ✅ Feature dependency graph with cycle detection and impact analysis
- ✅ Minimal feature overview page (auto-generated)

The following enhancements represent natural next steps that would further strengthen governance, automation, and traceability beyond the current thin slice.

**Note**: The thin slice (GOV_CORE) is **implemented and required for v1**. The enhancements below are **optional** and can be implemented incrementally as the project grows.

---

## 1. Machine-Verifiable Spec Schema

### Status

**✅ Thin Slice Implemented (GOV_CORE)**: YAML frontmatter with validation for feature, version, status, domain, flags, and exit codes. Full schema with data structures and JSON schema remains future work.

### Goal

Extend the current thin-slice schema to include full machine-readable definitions for data structures, JSON schemas, and output contracts.

### Current Implementation

The thin slice supports:
- YAML frontmatter with required fields (feature, version, status, domain)
- Optional flag definitions (name, type, default, description)
- Optional exit code definitions
- Validation of frontmatter structure and integrity

### Proposed Extended Schema

```yaml
---
feature: CLI_BUILD
version: v1
status: wip
domain: commands
---

inputs:
  flags:
    - name: --dry-run
      type: boolean
      default: false
      description: "Preview build without executing"
    - name: --json
      type: boolean
      default: false
      description: "Output in JSON format"

outputs:
  exit_codes:
    success: 0
    failure: 1
    invalid_config: 2

data_structures:
  - name: BuildPlan
    type: object
    fields:
      - name: services
        type: array
        items: ServiceBuildPlan
      - name: total_steps
        type: integer

json_schema:
  # OpenAPI-style JSON schema for output
```

### Benefits

- **No undocumented flags**: Schema enforces all flags are declared
- **All flags tested**: Automated test generation from schema
- **All exit codes tested**: Exit code coverage validation
- **Structural diff**: Compare spec schema vs implementation
- **API contract enforcement**: Similar to OpenAPI for CLI tools

### Implementation Notes

- ✅ **Done**: YAML frontmatter in markdown specs (GOV_CORE)
- ✅ **Done**: Validation script (`cmd/spec-validate`) parses schema and validates
- 🔄 **Future**: Test generator from schema
- 🔄 **Future**: Full JSON schema support for output contracts
- Similar to Rust's "stability attributes" system

---

## 2. Structural Diff for Spec vs Implementation

### Status

**✅ Thin Slice Implemented (GOV_CORE)**: Flag comparison (spec vs CLI) with type, default, and missing/extra flag detection. Exit code alignment and JSON schema comparison remain future work.

### Goal

Extend the current flag comparison to include exit code alignment and JSON schema comparison.

### Current Implementation

The thin slice supports:
- ✅ Flag comparison (spec ↔ CLI)
- ✅ Type alignment checking (with normalization)
- ✅ Default value comparison (warnings)
- ✅ Missing/extra flag detection
- ✅ Command → feature ID mapping (CLI_* convention)

### Proposed Extended Approach

1. **Extract spec definitions**:
   - ✅ Parse flags from spec frontmatter (done)
   - 🔄 Extract JSON schema from spec (future)
   - 🔄 Parse exit code definitions (future - schema exists, constants missing)

2. **Extract implementation definitions**:
   - ✅ Parse Cobra command flags via introspection (done)
   - 🔄 Extract JSON struct tags from Go code (future)
   - 🔄 Parse actual exit codes from shared constants (future)

3. **Compare and report**:
   - ✅ Flag mismatches (done)
   - 🔄 JSON schema mismatches (future)
   - 🔄 Exit code mismatches (future - see GOV_CORE_EXITCODES)

### Benefits

- **Guaranteed alignment**: Implementation exactly matches spec
- **Prevents drift**: Catches discrepancies automatically
- **CI enforcement**: Fails PRs when spec and code diverge
- **Documentation accuracy**: Spec is always ground truth

### Implementation Notes

- ✅ **Done**: Flag comparison via Cobra command introspection (`cmd/spec-vs-cli`)
- 🔄 **Future**: Go AST parsing for JSON struct tag extraction
- 🔄 **Future**: JSON schema comparison using existing libraries
- 🔄 **Future**: Exit code constants package and alignment (see GOV_CORE_EXITCODES)
- Similar to OpenAPI code generation validation

---

## 3. Feature Dependency Graph

### Status

**✅ Implemented (GOV_CORE)**: Dependency graph from `spec/features.yaml` with cycle detection, impact analysis, and DOT visualization. Header comment parsing remains future work.

### Goal

Extend the current graph to include header comment parsing and PR integration.

### Current Implementation

The thin slice supports:
- ✅ Dependency extraction from `spec/features.yaml`
- ✅ Dependency graph construction
- ✅ Cycle detection (DAG validation)
- ✅ Impact analysis (transitive dependencies)
- ✅ DOT visualization with status-based colors
- ✅ CI integration (graph validation in `run-all-checks.sh`)

### Proposed Extended Features

1. **Dependency Extraction**:
   - ✅ Extract dependencies from `spec/features.yaml` (done)
   - 🔄 Parse `// Feature:` comments in code (future - see GOV_CORE_HEADERS)
   - 🔄 Extract dependencies from spec frontmatter (future)

2. **Visualization**:
   - ✅ Generate DOT/Graphviz diagrams (done)
   - ✅ Feature dependency tree (done)
   - ✅ Impact analysis visualization (done)

3. **CI Integration**:
   - ✅ Cycle detection warnings (done)
   - ✅ Dependency completeness checks (done)
   - 🔄 GitHub PR comments: "Editing CORE_STATE will affect CLI_DEPLOY, CLI_ROLLBACK" (future)

### Benefits

- **Impact analysis**: Know what breaks when changing a feature
- **Cycle prevention**: Detect circular dependencies early
- **Documentation**: Visual representation of feature relationships
- **PR guidance**: Automatic warnings for reviewers

### Implementation Notes

- ✅ **Done**: Custom graph implementation with cycle detection (`internal/tools/features/`)
- ✅ **Done**: DOT file generation (`features.ToDOT()`)
- ✅ **Done**: Impact analysis (`features.Impact()`)
- 🔄 **Future**: GitHub API for PR comments
- 🔄 **Future**: Header comment parsing (see GOV_CORE_HEADERS)
- Similar to Bazel's dependency graph system

---

## 4. Full Feature Portal / Dashboard

### Status

**✅ Minimal Implementation (GOV_CORE)**: Auto-generated markdown overview (`docs/features/OVERVIEW.md`) with feature table, dependency graph, and status summary. Full interactive dashboard remains future work.

### Goal

Extend the current minimal overview to a comprehensive interactive dashboard.

### Current Implementation

The thin slice supports:
- ✅ Auto-generated markdown overview
- ✅ Features by domain table
- ✅ Dependency graph (textual)
- ✅ Status summary
- ✅ CI staleness check

### Proposed Extended Features

Generate a comprehensive dashboard showing:
- ✅ All features with status (done - in overview)
- ✅ Dependency relationships (done - in overview)
- 🔄 Spec versions (future)
- 🔄 Test completeness (future)
- 🔄 Coverage by feature (future)
- 🔄 Last updated commit (future)
- 🔄 Interactive search and filtering (future)

### Proposed Features

1. **Static Site Generation**:
   - Generate HTML dashboard from `spec/features.yaml`
   - Include spec excerpts, test coverage, commit history
   - Deploy to GitHub Pages or wiki

2. **Metrics**:
   - Feature completion rate
   - Average time from `todo` → `done`
   - Test coverage by feature
   - Spec drift metrics

3. **Search and Filter**:
   - Filter by domain, status, owner
   - Search by feature ID or description
   - Sort by completion date, priority

### Benefits

- **Transparency**: Clear view of project status
- **Accountability**: Track feature progress
- **Onboarding**: New contributors see feature landscape
- **Planning**: Identify bottlenecks and dependencies

### Implementation Notes

- Static site generator (Hugo, Jekyll, or custom)
- GitHub Actions to regenerate on spec changes
- GitHub Pages deployment
- Similar to Kubernetes feature tracking dashboards

---

## 5. Automated Changelog Generation

### Goal

Automatically generate changelogs from commit messages and feature lifecycle transitions.

### Proposed Approach

1. **Parse commit history**:
   - Extract `feat(<FEATURE_ID>):` commits
   - Track feature status transitions in `spec/features.yaml`
   - Group by feature and version

2. **Generate changelog**:
   - Group changes by feature
   - Include spec references
   - Link to PRs and commits
   - Format for release notes

### Benefits

- **Zero-effort changelogs**: Automatic from commit format
- **Feature-based organization**: Changes grouped by feature
- **Traceability**: Links from changelog to spec and code
- **Release notes**: Ready-to-publish release documentation

### Implementation Notes

- Git log parsing with commit message format
- Feature status tracking from `spec/features.yaml` history
- Markdown generation
- Similar to conventional-changelog but feature-aware

---

## 6. Behavioral Diff Tool

### Goal

Compare behavioral changes between versions by analyzing spec diffs and implementation changes.

### Proposed Features

1. **Spec Diff Analysis**:
   - Compare spec versions
   - Identify added/removed/changed flags
   - Detect behavioral changes

2. **Implementation Diff Analysis**:
   - Compare code changes
   - Identify new behaviors
   - Detect removed behaviors

3. **Alignment Check**:
   - Ensure spec changes match implementation changes
   - Flag behavioral changes without spec updates
   - Generate migration guides

### Benefits

- **Breaking change detection**: Identify breaking changes automatically
- **Migration guides**: Auto-generate upgrade instructions
- **Version compatibility**: Understand what changed between versions
- **Documentation accuracy**: Spec changes match code changes

---

## 7. Feature Completion Dashboard

### Goal

Track feature completion metrics and generate reports for planning and accountability.

### Proposed Metrics

1. **Completion Metrics**:
   - Features by status (todo/wip/done)
   - Average time in each status
   - Features blocked by dependencies
   - Test coverage by feature

2. **Velocity Metrics**:
   - Features completed per week/month
   - Average PR size by feature
   - Time from spec → done

3. **Quality Metrics**:
   - Spec completeness score
   - Test requirement fulfillment
   - Documentation completeness

### Benefits

- **Planning**: Data-driven feature planning
- **Accountability**: Track feature progress
- **Bottleneck identification**: Find stuck features
- **Resource allocation**: Identify where help is needed

---

## Implementation Priority

**Current Status**: Thin slice (GOV_CORE) is **implemented and required for v1**. The following represent natural evolution beyond the thin slice:

1. **✅ Implemented**: Feature Dependency Graph (#3) - thin slice complete
2. **✅ Partially Implemented**: Machine-Verifiable Spec Schema (#1) - thin slice complete, full schema future
3. **✅ Partially Implemented**: Structural Diff Tool (#2) - flags done, exit codes future
4. **✅ Minimal Implementation**: Feature Overview (#4) - markdown overview done, full portal future
5. **🔄 Future Work**: Exit Code Alignment (GOV_CORE_EXITCODES)
6. **🔄 Future Work**: Header Comment Validation (GOV_CORE_HEADERS)
7. **🔄 Future Work**: Automated Changelog (#5), Behavioral Diff (#6), Completion Dashboard (#7)

---

## References

- **Kubernetes KEP**: [Kubernetes Enhancement Proposals](https://github.com/kubernetes/enhancements)
- **Rust RFC**: [Rust RFC Process](https://github.com/rust-lang/rfcs)
- **Bazel Design Docs**: [Bazel Design Documents](https://github.com/bazelbuild/bazel/tree/master/designs)
- **TensorFlow Governance**: [TensorFlow RFC Process](https://github.com/tensorflow/community)

---

## Notes

**GOV_CORE (Implemented)** provides:

- ✅ Machine-readable spec schemas (YAML frontmatter)
- ✅ Automated structural diffs (flags: spec vs CLI)
- ✅ Dependency graph analysis (with cycle detection and impact)
- ✅ Minimal feature overview (auto-generated markdown)
- ✅ Spec integrity validation (features.yaml ↔ spec files)
- ✅ CI integration (governance checks in `run-all-checks.sh`)

**Future Enhancements** (beyond thin slice):

- 🔄 Full JSON schema support for output contracts
- 🔄 Exit code alignment (requires centralized constants - see GOV_CORE_EXITCODES)
- 🔄 Header comment validation (see GOV_CORE_HEADERS)
- 🔄 Interactive feature dashboard (beyond markdown overview)
- 🔄 Automated changelog generation
- 🔄 Behavioral diff tool
- 🔄 Completion metrics dashboard

**Related Features**:
- `GOV_CORE_EXITCODES`: Exit code constants and alignment (future)
- `GOV_CORE_HEADERS`: Header comment validation (future)

All future enhancements are **optional** and can be implemented incrementally as the project grows.


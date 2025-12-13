# AI-Agent Documentation

This directory contains human-readable documentation generated from AI-Agent pipeline outputs.

## What is AI-Agent?

AI-Agent is a repository analysis tooling paradigm that consists of:

- **XRAY**: Structural reverse-engineering scanner that analyzes codebase structure
- **Context Compiler**: Deterministic packaging tool that processes declared repo knowledge (docs/spec files)
- **AI-Agent**: AI consumption layer that provides searchable context snapshots

This pipeline enables AI tools to understand repository structure and documentation in a deterministic, reproducible way.

## Inputs

The documentation in this directory is generated from:

- `.ai-context/files/manifest.json`: List of processed files with chunk counts
- `.ai-context/files/chunks.ndjson`: All document chunks in NDJSON format
- `.ai-context/xray/stagecraft/data/index.json`: XRAY structural analysis data (if available)

## Regenerating Documentation

To regenerate this documentation:

```bash
# Using npm script
npm --prefix tools/context-compiler run context:docs

# Or using Stagecraft CLI
stagecraft context docs
```

The generator reads the files listed above and produces deterministic markdown output. Running the generator twice produces identical output (no timestamps or non-deterministic content).

## Generated Documentation

This directory contains the following files:

1. **README.md** (this file): Overview of AI-Agent and how to regenerate docs
2. **REPO_INDEX.md**: Repository structure summary from XRAY data
3. **DOCS_CATALOG.md**: Catalog of all documentation files with chunk counts and headings
4. **SPEC_CATALOG.md**: Focused catalog of specification files (`spec/**`)

Additional generated pages:

5. **AI_AGENT_NAVIGATION_PLAN.md**: 12-step human navigation checklist
6. **DOCS_TO_GENERATE_BACKLOG.md**: proposed generated-docs backlog and generation rules
7. **COMMAND_CATALOG.md**: filtered view of `spec/commands/**`
8. **PROVIDER_CATALOG.md**: filtered view of `spec/providers/**`
9. **CORE_SPEC_INDEX.md**: focused index of core engine and architecture specs
10. **GOVERNANCE_INDEX.md**: governance and ADR documentation index
11. **ENGINE_ANALYSIS_INDEX.md**: engineering analysis document index
12. **ENGINE_OUTLINE_INDEX.md**: implementation outline index

## Intended Audience

- **Onboarding**: New contributors learning the repository structure
- **Navigation**: Finding relevant documentation and specifications
- **AI Tools**: Providing structured context about the repository
- **Maintainers**: Understanding documentation coverage and organization

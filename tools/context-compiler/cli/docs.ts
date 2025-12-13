// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

import { resolve, join, basename } from "node:path";
import { promises as fs } from "node:fs";
import type { FileManifestEntry, ChunkEntry } from "./types";

interface XrayIndex {
    root?: string;
    indexedAt?: string;
    scannedAt?: string;
    target?: string;
    files?: Array<{ path: string; size: number; lines?: number; ext?: string }>;
    languages?: Record<string, number>;
    topDirs?: Record<string, number>;
    packages?: Array<{ root: string; name?: string; files?: Array<{ path: string; size: number }> }>;
    stats?: {
        files: number;
        bytes: number;
        languages: Record<string, number>;
    };
    biggestFiles?: Array<{ path: string; size: number }>;
    digest?: string;
}

async function readJSON<T>(path: string): Promise<T | null> {
    try {
        const content = await fs.readFile(path, "utf8");
        return JSON.parse(content) as T;
    } catch {
        return null;
    }
}

async function readNDJSON<T>(path: string): Promise<T[]> {
    try {
        const content = await fs.readFile(path, "utf8");
        return content
            .split("\n")
            .filter(line => line.trim())
            .map(line => JSON.parse(line) as T);
    } catch {
        return [];
    }
}

function formatBytes(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function extractMarkdownHeadings(text: string): string[] {
    const headings: string[] = [];
    const lines = text.split("\n");
    for (const line of lines) {
        const match = line.match(/^(#{1,6})\s+(.+)$/);
        if (match) {
            headings.push(match[2].trim());
        }
    }
    return headings;
}

function getFirstLine(text: string): string {
    const firstLine = text.split("\n")[0].trim();
    return firstLine.length > 80 ? firstLine.substring(0, 77) + "..." : firstLine;
}

// --- New helpers ---
/**
 * Writes file content with exactly one trailing \n (append if missing), UTF-8.
 */
async function writeDeterministicFile(path: string, content: string): Promise<void> {
    let c = content;
    // Remove trailing whitespace and newlines
    c = c.replace(/[\r\n]+$/, "");
    c += "\n";
    await fs.writeFile(path, c, "utf8");
}

/**
 * Escapes | as \| and replaces newlines with spaces for markdown table cells.
 */
function mdEscapeTableCell(s: string): string {
    return s.replace(/\|/g, "\\|").replace(/[\r\n]+/g, " ");
}

async function generateReadme(repoRoot: string, outDir: string): Promise<void> {
    const content = `# AI-Agent Documentation

This directory contains human-readable documentation generated from AI-Agent pipeline outputs.

## What is AI-Agent?

AI-Agent is a repository analysis tooling paradigm that consists of:

- **XRAY**: Structural reverse-engineering scanner that analyzes codebase structure
- **Context Compiler**: Deterministic packaging tool that processes declared repo knowledge (docs/spec files)
- **AI-Agent**: AI consumption layer that provides searchable context snapshots

This pipeline enables AI tools to understand repository structure and documentation in a deterministic, reproducible way.

## Inputs

The documentation in this directory is generated from:

- \`.ai-context/files/manifest.json\`: List of processed files with chunk counts
- \`.ai-context/files/chunks.ndjson\`: All document chunks in NDJSON format
- \`.ai-context/xray/stagecraft/data/index.json\`: XRAY structural analysis data (if available)

## Regenerating Documentation

To regenerate this documentation:

\`\`\`bash
# Using npm script
npm --prefix tools/context-compiler run context:docs

# Or using Stagecraft CLI
stagecraft context docs
\`\`\`

The generator reads the files listed above and produces deterministic markdown output. Running the generator twice produces identical output (no timestamps or non-deterministic content).

## Generated Documentation

This directory contains the following files:

1. **README.md** (this file): Overview of AI-Agent and how to regenerate docs
2. **REPO_INDEX.md**: Repository structure summary from XRAY data
3. **DOCS_CATALOG.md**: Catalog of all documentation files with chunk counts and headings
4. **SPEC_CATALOG.md**: Focused catalog of specification files (\`spec/**\`)

Additional generated pages:

5. **AI_AGENT_NAVIGATION_PLAN.md**: 12-step human navigation checklist
6. **DOCS_TO_GENERATE_BACKLOG.md**: proposed generated-docs backlog and generation rules
7. **COMMAND_CATALOG.md**: filtered view of \`spec/commands/**\`
8. **PROVIDER_CATALOG.md**: filtered view of \`spec/providers/**\`
9. **CORE_SPEC_INDEX.md**: focused index of core engine and architecture specs
10. **GOVERNANCE_INDEX.md**: governance and ADR documentation index
11. **ENGINE_ANALYSIS_INDEX.md**: engineering analysis document index
12. **ENGINE_OUTLINE_INDEX.md**: implementation outline index
13. **CONTEXT_HANDOFF_INDEX.md**: traceability index of context handoff docs
14. **COVERAGE_INDEX.md**: coverage planning and reports index
15. **ARCHIVE_INDEX.md**: index of archived and historical docs
16. **GUIDES_AND_NARRATIVES_INDEX.md**: guides and narratives discovery index

## Intended Audience

- **Onboarding**: New contributors learning the repository structure
- **Navigation**: Finding relevant documentation and specifications
- **AI Tools**: Providing structured context about the repository
- **Maintainers**: Understanding documentation coverage and organization
`;

    await writeDeterministicFile(join(outDir, "README.md"), content);
}
// --- Deterministic doc generators ---

/**
 * Writes AI_AGENT_NAVIGATION_PLAN.md (deterministic, verbatim content).
 */
async function generateNavigationPlan(outDir: string): Promise<void> {
    const content = `# AI-Agent Navigation Plan

Ordered checklist designed to be followed in 10–15 minutes.

1. Understand the AI-Agent pipeline
   - Open \`docs/generated/ai-agent/README.md\` and read the sections \`What is AI-Agent?\` and \`Inputs\`.
2. Confirm generated artifacts and scope
   - In \`docs/generated/ai-agent/README.md\`, read \`Generated Documentation\`.
3. Get a repository-wide snapshot
   - Open \`docs/generated/ai-agent/REPO_INDEX.md\` and review \`Statistics\` and \`Languages\`.
4. Locate primary entry points
   - In \`docs/generated/ai-agent/REPO_INDEX.md\`, read \`Where to Start\`.
5. Confirm Agent contract and workflow rules
   - Using \`docs/generated/ai-agent/DOCS_CATALOG.md\`, locate \`Root Files\` -> \`Agent.md\` and review \`Spec-first, Test-first\` and \`Feature Planning Protocol\`.
6. Survey documentation coverage and structure
   - In \`docs/generated/ai-agent/DOCS_CATALOG.md\`, read \`Summary\` and scan \`Files by Directory\`.
7. Identify governance and process documentation
   - From \`docs/generated/ai-agent/DOCS_CATALOG.md\`, locate \`docs/governance/GOVERNANCE_ALMANAC.md\` and consult \`Core Governance Principles\`.
8. Understand contribution and AI usage rules
   - Using \`docs/generated/ai-agent/DOCS_CATALOG.md\`, locate \`docs/governance/CONTRIBUTING_CURSOR.md\` and review \`Thread Hygiene\`.
9. Assess specification surface area
   - Open \`docs/generated/ai-agent/SPEC_CATALOG.md\` and read the \`Summary\` section.
10. Understand core architecture decisions
    - In \`docs/generated/ai-agent/SPEC_CATALOG.md\`, locate \`spec/adr/0001-architecture.md\` and review \`Core Structure (v1)\`.
11. Inspect command and provider patterns
    - From \`docs/generated/ai-agent/SPEC_CATALOG.md\`, locate \`spec/commands/build.md\` and \`spec/providers/backend/generic.md\` and review \`CLI Definition\` and \`Interface\`.
12. Verify generated-docs locality
    - Confirm the directory \`docs/generated/ai-agent/\` contains the canonical files listed in \`docs/generated/ai-agent/README.md\`.
`;
    await writeDeterministicFile(join(outDir, "AI_AGENT_NAVIGATION_PLAN.md"), content);
}

/**
 * Writes DOCS_TO_GENERATE_BACKLOG.md (deterministic, verbatim content).
 */
async function generateDocsBacklog(outDir: string): Promise<void> {
    const content = `# Docs-to-Generate Backlog

Prioritized list of new documentation pages to generate under \`docs/generated/ai-agent/\`. Sources are restricted to the four canonical generated inputs.

1. \`COMMAND_CATALOG.md\`
   - Purpose: Central index of all CLI command specifications.
   - Sources: \`SPEC_CATALOG.md\` -> \`Files\` table; filter paths starting with \`spec/commands/\`.
   - Rule: Sort by \`Path\` ascending.
2. \`PROVIDER_CATALOG.md\`
   - Purpose: Central index of all provider specifications and interfaces.
   - Sources: \`SPEC_CATALOG.md\` -> \`Files\` table; filter paths starting with \`spec/providers/\`.
   - Rule: Group by provider domain (directory after \`spec/providers/\`); within each group sort by \`Path\`.
3. \`CORE_SPEC_INDEX.md\`
   - Purpose: Focused index of core engine and architecture specs.
   - Sources: \`SPEC_CATALOG.md\` -> \`Files\` table; filter paths under \`spec/core/\` and \`spec/adr/\`.
   - Rule: ADR group first then core group; within each group sort by \`Path\`.
4. \`GOVERNANCE_INDEX.md\`
   - Purpose: Single entry point for governance and ADR documentation.
   - Sources: \`DOCS_CATALOG.md\` -> \`Files by Directory\`; include \`docs/governance/\` and \`docs/adr/\`.
   - Rule: Group by directory; within each group sort by path.
5. \`ENGINE_ANALYSIS_INDEX.md\`
   - Purpose: Index of problem statements and engineering analyses.
   - Sources: \`DOCS_CATALOG.md\` -> \`Files by Directory\`; include \`docs/engine/analysis/\`.
   - Rule: Sort by path.
6. \`ENGINE_OUTLINE_INDEX.md\`
   - Purpose: Inventory of implementation outlines.
   - Sources: \`DOCS_CATALOG.md\` -> \`Files by Directory\`; include \`docs/engine/outlines/\`.
   - Rule: Sort by path.
7. \`CONTEXT_HANDOFF_INDEX.md\`
   - Purpose: Traceable list of context-handoff documents.
   - Sources: \`DOCS_CATALOG.md\` -> \`Files by Directory\`; include \`docs/context-handoff/\`.
   - Rule: Sort by path.
8. \`COVERAGE_INDEX.md\`
   - Purpose: Aggregated view of coverage planning and reports.
   - Sources: \`DOCS_CATALOG.md\` -> \`Files by Directory\`; include \`docs/coverage/\`.
   - Rule: Sort by path.
9. \`ARCHIVE_INDEX.md\`
   - Purpose: Visibility into archived and historical documents.
   - Sources: \`DOCS_CATALOG.md\` -> \`Files by Directory\`; include \`docs/archive/\` and \`docs/engine/history/\`.
   - Rule: Group by directory; within each group sort by path.
10. \`GUIDES_AND_NARRATIVES_INDEX.md\`
    - Purpose: Discovery index for guides and narrative documentation.
    - Sources: \`DOCS_CATALOG.md\` -> \`Files by Directory\`; include \`docs/guides/\` and \`docs/narrative/\`.
    - Rule: Group by directory; within each group sort by path.
`;
    await writeDeterministicFile(join(outDir, "DOCS_TO_GENERATE_BACKLOG.md"), content);
}

/**
 * Writes COMMAND_CATALOG.md: table of all spec/commands/** files.
 */
async function generateCommandCatalog(
    repoRoot: string,
    outDir: string,
    manifest: FileManifestEntry[],
    chunks: ChunkEntry[]
): Promise<void> {
    let content = `# Command Catalog

This catalog lists command specs under \`spec/commands/\`.

| Path | Chunks | Primary Topics |
|------|--------|----------------|
`;
    // Filter for spec/commands/
    const commandFiles = manifest.filter(f => f.path.startsWith("spec/commands/"));
    commandFiles.sort((a, b) => a.path.localeCompare(b.path));
    for (const file of commandFiles) {
        const fileChunks = chunks.filter(c => c.path === file.path);
        const headings = extractHeadingsFromChunks(fileChunks);
        const topics = headings
            .filter(h => {
                const lower = h.toLowerCase();
                return !["overview", "introduction", "table of contents", "contents"].includes(lower);
            })
            .slice(0, 5)
            .join(", ") || "General";
        content += `| \`${file.path}\` | ${file.chunks} | ${mdEscapeTableCell(topics)} |\n`;
    }
    await writeDeterministicFile(join(outDir, "COMMAND_CATALOG.md"), content);
}

/**
 * Writes PROVIDER_CATALOG.md: grouped table of all spec/providers/** files.
 */
async function generateProviderCatalog(
    repoRoot: string,
    outDir: string,
    manifest: FileManifestEntry[],
    chunks: ChunkEntry[]
): Promise<void> {
    let content = `# Provider Catalog

This catalog lists provider specs under \`spec/providers/\`, grouped by provider domain.
`;
    // Filter for spec/providers/
    const providerFiles = manifest.filter(f => f.path.startsWith("spec/providers/"));
    // Group by provider domain (segment after spec/providers/)
    const byDomain: Record<string, FileManifestEntry[]> = {};
    for (const file of providerFiles) {
        const rel = file.path.slice("spec/providers/".length);
        const segs = rel.split("/");
        let domain = segs.length > 1 && segs[0] ? segs[0] : "other";
        if (!byDomain[domain]) byDomain[domain] = [];
        byDomain[domain].push(file);
    }
    const domains = Object.keys(byDomain).sort();
    for (const domain of domains) {
        content += `\n\n## ${domain}\n\n`;
        content += `| Path | Chunks | Primary Topics |\n`;
        content += `|------|--------|----------------|\n`;
        const files = byDomain[domain].sort((a, b) => a.path.localeCompare(b.path));
        for (const file of files) {
            const fileChunks = chunks.filter(c => c.path === file.path);
            const headings = extractHeadingsFromChunks(fileChunks);
            const topics = headings
                .filter(h => {
                    const lower = h.toLowerCase();
                    return !["overview", "introduction", "table of contents", "contents"].includes(lower);
                })
                .slice(0, 5)
                .join(", ") || "General";
            content += `| \`${file.path}\` | ${file.chunks} | ${mdEscapeTableCell(topics)} |\n`;
        }
    }
    await writeDeterministicFile(join(outDir, "PROVIDER_CATALOG.md"), content);
}

/**
 * Writes CORE_SPEC_INDEX.md:
 * - Group 1: spec/adr/
 * - Group 2: spec/core/
 */
async function generateCoreSpecIndex(
    repoRoot: string,
    outDir: string,
    manifest: FileManifestEntry[],
    chunks: ChunkEntry[]
): Promise<void> {
    let content = `# Core Spec Index

Focused index of core engine and architecture specifications.

`;

    // Filter relevant files
    const coreFiles = manifest.filter(f => f.path.startsWith("spec/core/") || f.path.startsWith("spec/adr/"));

    // Group them
    const groups = {
        adr: coreFiles.filter(f => f.path.startsWith("spec/adr/")).sort((a, b) => a.path.localeCompare(b.path)),
        core: coreFiles.filter(f => f.path.startsWith("spec/core/")).sort((a, b) => a.path.localeCompare(b.path))
    };

    // Helper to generate table
    const generateTable = (files: FileManifestEntry[]) => {
        let table = `| Path | Chunks | Primary Topics |\n`;
        table += `|------|--------|----------------|\n`;
        for (const file of files) {
            const fileChunks = chunks.filter(c => c.path === file.path);
            const headings = extractHeadingsFromChunks(fileChunks);
            const topics = headings
                .filter(h => {
                    const lower = h.toLowerCase();
                    return !["overview", "introduction", "table of contents", "contents"].includes(lower);
                })
                .slice(0, 5)
                .join(", ") || "General";
            table += `| \`${file.path}\` | ${file.chunks} | ${mdEscapeTableCell(topics)} |\n`;
        }
        return table;
    };

    // spec/adr/
    if (groups.adr.length > 0) {
        content += `## Architecture Decision Records (ADR)\n\n`;
        content += generateTable(groups.adr);
        content += `\n`;
    }

    // spec/core/
    if (groups.core.length > 0) {
        content += `## Core Specifications\n\n`;
        content += generateTable(groups.core);
        content += `\n`;
    }

    await writeDeterministicFile(join(outDir, "CORE_SPEC_INDEX.md"), content);
}

/**
 * Writes GOVERNANCE_INDEX.md:
 * - Group by directory (docs/governance/, docs/adr/)
 */
async function generateGovernanceIndex(
    repoRoot: string,
    outDir: string,
    manifest: FileManifestEntry[],
    chunks: ChunkEntry[]
): Promise<void> {
    let content = `# Governance Index

Single entry point for governance and ADR documentation.

`;

    // Filter relevant files
    const govFiles = manifest.filter(f => f.path.startsWith("docs/governance/") || f.path.startsWith("docs/adr/"));

    // Group by directory
    const byDir: Record<string, FileManifestEntry[]> = {};
    for (const file of govFiles) {
        // Simple directory extraction: docs/governance/foo.md -> docs/governance/
        // Actually, let's group by top-level dir under docs/ for likely stability, or just exact dir?
        // Task says: "Group by directory (docs/governance/, docs/adr/)"
        const dir = file.path.startsWith("docs/governance/") ? "docs/governance/" : "docs/adr/";
        if (!byDir[dir]) byDir[dir] = [];
        byDir[dir].push(file);
    }

    const sortedDirs = Object.keys(byDir).sort();

    for (const dir of sortedDirs) {
        content += `## \`${dir}\`\n\n`;
        content += `| Path | Chunks | Primary Topics |\n`;
        content += `|------|--------|----------------|\n`;

        const files = byDir[dir].sort((a, b) => a.path.localeCompare(b.path));
        for (const file of files) {
            const fileChunks = chunks.filter(c => c.path === file.path);
            const headings = extractHeadingsFromChunks(fileChunks);
            const topics = headings
                .filter(h => {
                    const lower = h.toLowerCase();
                    return !["overview", "introduction", "table of contents", "contents"].includes(lower);
                })
                .slice(0, 5)
                .join(", ") || "General";
            content += `| \`${file.path}\` | ${file.chunks} | ${mdEscapeTableCell(topics)} |\n`;
        }
        content += `\n`;
    }

    await writeDeterministicFile(join(outDir, "GOVERNANCE_INDEX.md"), content);
}

/**
 * Writes ENGINE_ANALYSIS_INDEX.md:
 * - docs/engine/analysis/
 */
async function generateEngineAnalysisIndex(
    repoRoot: string,
    outDir: string,
    manifest: FileManifestEntry[],
    chunks: ChunkEntry[]
): Promise<void> {
    let content = `# Engine Analysis Index

Index of engineering problem analysis documents.

| Path | Chunks | Primary Topics |
|------|--------|----------------|
`;

    // Filter and sort
    const files = manifest
        .filter(f => f.path.startsWith("docs/engine/analysis/"))
        .sort((a, b) => a.path.localeCompare(b.path));

    for (const file of files) {
        const fileChunks = chunks.filter(c => c.path === file.path);
        const headings = extractHeadingsFromChunks(fileChunks);
        const topics = headings
            .filter(h => {
                const lower = h.toLowerCase();
                return !["overview", "introduction", "table of contents", "contents"].includes(lower);
            })
            .slice(0, 5)
            .join(", ") || "General";
        content += `| \`${file.path}\` | ${file.chunks} | ${mdEscapeTableCell(topics)} |\n`;
    }

    await writeDeterministicFile(join(outDir, "ENGINE_ANALYSIS_INDEX.md"), content);
}

/**
 * Writes ENGINE_OUTLINE_INDEX.md:
 * - docs/engine/outlines/
 */
async function generateEngineOutlineIndex(
    repoRoot: string,
    outDir: string,
    manifest: FileManifestEntry[],
    chunks: ChunkEntry[]
): Promise<void> {
    let content = `# Engine Outline Index

Inventory of implementation outlines.

| Path | Chunks | Primary Topics |
|------|--------|----------------|
`;

    // Filter and sort
    const files = manifest
        .filter(f => f.path.startsWith("docs/engine/outlines/"))
        .sort((a, b) => a.path.localeCompare(b.path));

    for (const file of files) {
        const fileChunks = chunks.filter(c => c.path === file.path);
        const headings = extractHeadingsFromChunks(fileChunks);
        const topics = headings
            .filter(h => {
                const lower = h.toLowerCase();
                return !["overview", "introduction", "table of contents", "contents"].includes(lower);
            })
            .slice(0, 5)
            .join(", ") || "General";
        content += `| \`${file.path}\` | ${file.chunks} | ${mdEscapeTableCell(topics)} |\n`;
    }

    await writeDeterministicFile(join(outDir, "ENGINE_OUTLINE_INDEX.md"), content);
}

/**
 * Writes CONTEXT_HANDOFF_INDEX.md
 * - docs/context-handoff/
 */
async function generateContextHandoffIndex(
    repoRoot: string,
    outDir: string,
    manifest: FileManifestEntry[],
    chunks: ChunkEntry[]
): Promise<void> {
    let content = `# Context Handoff Index

Traceable index of context handoff documents.

| Path | Chunks | Primary Topics |
|------|--------|----------------|
`;
    const files = manifest
        .filter(f => f.path.startsWith("docs/context-handoff/"))
        .sort((a, b) => a.path.localeCompare(b.path));

    for (const file of files) {
        const fileChunks = chunks.filter(c => c.path === file.path);
        const headings = extractHeadingsFromChunks(fileChunks);
        const topics = headings
            .filter(h => {
                const lower = h.toLowerCase();
                return !["overview", "introduction", "table of contents", "contents"].includes(lower);
            })
            .slice(0, 5)
            .join(", ") || "General";
        content += `| \`${file.path}\` | ${file.chunks} | ${mdEscapeTableCell(topics)} |\n`;
    }
    await writeDeterministicFile(join(outDir, "CONTEXT_HANDOFF_INDEX.md"), content);
}

/**
 * Writes COVERAGE_INDEX.md
 * - docs/coverage/
 */
async function generateCoverageIndex(
    repoRoot: string,
    outDir: string,
    manifest: FileManifestEntry[],
    chunks: ChunkEntry[]
): Promise<void> {
    let content = `# Coverage Index

Aggregated index of coverage planning and reports.

| Path | Chunks | Primary Topics |
|------|--------|----------------|
`;
    const files = manifest
        .filter(f => f.path.startsWith("docs/coverage/"))
        .sort((a, b) => a.path.localeCompare(b.path));

    for (const file of files) {
        const fileChunks = chunks.filter(c => c.path === file.path);
        const headings = extractHeadingsFromChunks(fileChunks);
        const topics = headings
            .filter(h => {
                const lower = h.toLowerCase();
                return !["overview", "introduction", "table of contents", "contents"].includes(lower);
            })
            .slice(0, 5)
            .join(", ") || "General";
        content += `| \`${file.path}\` | ${file.chunks} | ${mdEscapeTableCell(topics)} |\n`;
    }
    await writeDeterministicFile(join(outDir, "COVERAGE_INDEX.md"), content);
}

/**
 * Writes ARCHIVE_INDEX.md
 * - docs/archive/
 * - docs/engine/history/
 */
async function generateArchiveIndex(
    repoRoot: string,
    outDir: string,
    manifest: FileManifestEntry[],
    chunks: ChunkEntry[]
): Promise<void> {
    let content = `# Archive Index

Visibility into archived and historical documents.

`;
    const relevantFiles = manifest.filter(f =>
        f.path.startsWith("docs/archive/") || f.path.startsWith("docs/engine/history/")
    );

    const byDir: Record<string, FileManifestEntry[]> = {};
    for (const file of relevantFiles) {
        // Determine group directory
        // Task says "Group by directory". docs/archive/foo.md -> docs/archive/
        // We will group by the known source directories: docs/archive/ or docs/engine/history/
        // Or strictly by directory name? "Group by directory" usually means immediate parent or the top-level grouping asked.
        // Given the inputs are docs/archive/ and docs/engine/history/, let's group by those prefixes/folders.
        let dir = "";
        if (file.path.startsWith("docs/archive/")) dir = "docs/archive/";
        else if (file.path.startsWith("docs/engine/history/")) dir = "docs/engine/history/";
        else dir = "other/"; // Should not happen based on filter

        if (!byDir[dir]) byDir[dir] = [];
        byDir[dir].push(file);
    }

    const sortedDirs = Object.keys(byDir).sort();

    for (const dir of sortedDirs) {
        content += `## \`${dir}\`\n\n`;
        content += `| Path | Chunks | Primary Topics |\n`;
        content += `|------|--------|----------------|\n`;
        const files = byDir[dir].sort((a, b) => a.path.localeCompare(b.path));
        for (const file of files) {
            const fileChunks = chunks.filter(c => c.path === file.path);
            const headings = extractHeadingsFromChunks(fileChunks);
            const topics = headings
                .filter(h => {
                    const lower = h.toLowerCase();
                    return !["overview", "introduction", "table of contents", "contents"].includes(lower);
                })
                .slice(0, 5)
                .join(", ") || "General";
            content += `| \`${file.path}\` | ${file.chunks} | ${mdEscapeTableCell(topics)} |\n`;
        }
        content += `\n`;
    }
    await writeDeterministicFile(join(outDir, "ARCHIVE_INDEX.md"), content);
}

/**
 * Writes GUIDES_AND_NARRATIVES_INDEX.md
 * - docs/guides/
 * - docs/narrative/
 */
async function generateGuidesAndNarrativesIndex(
    repoRoot: string,
    outDir: string,
    manifest: FileManifestEntry[],
    chunks: ChunkEntry[]
): Promise<void> {
    let content = `# Guides and Narratives Index

Discovery index for guides and narrative documentation.

`;
    const relevantFiles = manifest.filter(f =>
        f.path.startsWith("docs/guides/") || f.path.startsWith("docs/narrative/")
    );

    const byDir: Record<string, FileManifestEntry[]> = {};
    for (const file of relevantFiles) {
        let dir = "";
        if (file.path.startsWith("docs/guides/")) dir = "docs/guides/";
        else if (file.path.startsWith("docs/narrative/")) dir = "docs/narrative/";
        else dir = "other/";

        if (!byDir[dir]) byDir[dir] = [];
        byDir[dir].push(file);
    }

    const sortedDirs = Object.keys(byDir).sort();

    for (const dir of sortedDirs) {
        content += `## \`${dir}\`\n\n`;
        content += `| Path | Chunks | Primary Topics |\n`;
        content += `|------|--------|----------------|\n`;
        const files = byDir[dir].sort((a, b) => a.path.localeCompare(b.path));
        for (const file of files) {
            const fileChunks = chunks.filter(c => c.path === file.path);
            const headings = extractHeadingsFromChunks(fileChunks);
            const topics = headings
                .filter(h => {
                    const lower = h.toLowerCase();
                    return !["overview", "introduction", "table of contents", "contents"].includes(lower);
                })
                .slice(0, 5)
                .join(", ") || "General";
            content += `| \`${file.path}\` | ${file.chunks} | ${mdEscapeTableCell(topics)} |\n`;
        }
        content += `\n`;
    }
    await writeDeterministicFile(join(outDir, "GUIDES_AND_NARRATIVES_INDEX.md"), content);
}

async function generateRepoIndex(repoRoot: string, outDir: string, xrayIndex: XrayIndex | null): Promise<void> {
    let content = `# Repository Index

This document summarizes the repository structure based on XRAY analysis data.

`;

    if (!xrayIndex) {
        content += `> **Note**: XRAY index data not available. Run \`stagecraft context xray\` to generate it.\n\n`;
        await fs.writeFile(join(outDir, "REPO_INDEX.md"), content, "utf8");
        return;
    }

    // Repository root - use basename for portability
    const rootName = xrayIndex.root || (xrayIndex.target ? basename(xrayIndex.target) : basename(repoRoot));
    content += `## Repository: ${rootName}\n\n`;

    // File counts and sizes
    const totalFiles = xrayIndex.stats?.files || xrayIndex.files?.length || 0;
    const totalBytes = xrayIndex.stats?.bytes ||
        (xrayIndex.files ? xrayIndex.files.reduce((sum, f) => sum + f.size, 0) : 0);

    content += `### Statistics\n\n`;
    content += `- **Total Files**: ${totalFiles.toLocaleString()}\n`;
    content += `- **Total Size**: ${formatBytes(totalBytes)}\n\n`;

    // Languages
    const languages = xrayIndex.stats?.languages || xrayIndex.languages || {};
    if (Object.keys(languages).length > 0) {
        content += `### Languages\n\n`;
        const langEntries = Object.entries(languages)
            .sort((a, b) => b[1] - a[1])
            .slice(0, 10);
        for (const [lang, bytes] of langEntries) {
            content += `- **${lang}**: ${formatBytes(bytes)}\n`;
        }
        content += `\n`;
    }

    // Top directories
    const topDirs = xrayIndex.topDirs || {};
    if (Object.keys(topDirs).length > 0) {
        content += `### Top Directories\n\n`;
        const dirEntries = Object.entries(topDirs)
            .sort((a, b) => b[1] - a[1])
            .slice(0, 10);
        for (const [dir, bytes] of dirEntries) {
            content += `- **${dir}/**: ${formatBytes(bytes)}\n`;
        }
        content += `\n`;
    }

    // Biggest files
    let biggestFiles: Array<{ path: string; size: number }> = [];
    if (xrayIndex.biggestFiles) {
        biggestFiles = xrayIndex.biggestFiles;
    } else if (xrayIndex.files) {
        biggestFiles = [...xrayIndex.files]
            .sort((a, b) => b.size - a.size)
            .slice(0, 10);
    }

    if (biggestFiles.length > 0) {
        content += `### Largest Files\n\n`;
        for (const file of biggestFiles) {
            content += `- **${file.path}**: ${formatBytes(file.size)}\n`;
        }
        content += `\n`;
    }

    // Packages
    if (xrayIndex.packages && xrayIndex.packages.length > 0) {
        content += `### Packages\n\n`;
        for (const pkg of xrayIndex.packages.slice(0, 10)) {
            const name = pkg.name || pkg.root;
            const fileCount = pkg.files?.length || 0;
            content += `- **${name}** (\`${pkg.root}\`): ${fileCount} files\n`;
        }
        content += `\n`;
    }

    // Where to start
    content += `## Where to Start\n\n`;
    content += `Top 10 entry points by importance:\n\n`;

    const entryPoints = [
        "README.md",
        "Agent.md",
        "spec/overview.md",
        "docs/README.md",
        "docs/narrative/architecture.md",
        "docs/guides/getting-started.md",
        "CONTRIBUTING.md",
        "docs/governance/GOVERNANCE_ALMANAC.md",
        "spec/features.yaml",
        "docs/features/OVERVIEW.md"
    ];

    for (let i = 0; i < entryPoints.length; i++) {
        content += `${i + 1}. \`${entryPoints[i]}\`\n`;
    }

    await fs.writeFile(join(outDir, "REPO_INDEX.md"), content, "utf8");
}

async function generateDocsCatalog(
    repoRoot: string,
    outDir: string,
    manifest: FileManifestEntry[],
    chunks: ChunkEntry[]
): Promise<void> {
    let content = `# Documentation Catalog

This catalog lists all documentation files processed by the context compiler, grouped by top-level directory.

## Summary

- **Total Files**: ${manifest.length}
- **Total Chunks**: ${chunks.length}

## Files by Directory

`;

    // Group files by top-level directory
    const byDir: Record<string, FileManifestEntry[]> = {};
    const rootFiles: FileManifestEntry[] = [];

    for (const file of manifest) {
        const parts = file.path.split("/");
        if (parts.length === 1) {
            rootFiles.push(file);
        } else {
            const topDir = parts[0];
            if (!byDir[topDir]) {
                byDir[topDir] = [];
            }
            byDir[topDir].push(file);
        }
    }

    // Sort directories
    const sortedDirs = Object.keys(byDir).sort();

    // Root files
    if (rootFiles.length > 0) {
        content += `### Root Files\n\n`;
        rootFiles.sort((a, b) => a.path.localeCompare(b.path));
        for (const file of rootFiles) {
            const fileChunks = chunks.filter(c => c.path === file.path);
            const headings = extractHeadingsFromChunks(fileChunks);
            content += `#### \`${file.path}\`\n\n`;
            content += `- **Chunks**: ${file.chunks}\n`;
            if (headings.length > 0) {
                content += `- **Headings**: ${headings.slice(0, 5).join(", ")}${headings.length > 5 ? "..." : ""}\n`;
            }
            content += `\n`;
        }
        content += `\n`;
    }

    // Directory groups
    for (const dir of sortedDirs) {
        content += `### \`${dir}/\`\n\n`;
        const files = byDir[dir].sort((a, b) => a.path.localeCompare(b.path));

        for (const file of files) {
            const fileChunks = chunks.filter(c => c.path === file.path);
            const headings = extractHeadingsFromChunks(fileChunks);

            content += `#### \`${file.path}\`\n\n`;
            content += `- **Chunks**: ${file.chunks}\n`;
            if (headings.length > 0) {
                content += `- **Headings**: ${headings.slice(0, 5).join(", ")}${headings.length > 5 ? "..." : ""}\n`;
            }
            content += `\n`;
        }
        content += `\n`;
    }

    await fs.writeFile(join(outDir, "DOCS_CATALOG.md"), content, "utf8");
}

function extractHeadingsFromChunks(chunks: ChunkEntry[]): string[] {
    const allHeadings = new Set<string>();
    for (const chunk of chunks) {
        // Try to get headings from meta first
        if (chunk.meta?.headings && Array.isArray(chunk.meta.headings)) {
            for (const h of chunk.meta.headings) {
                if (typeof h === "string") allHeadings.add(h);
            }
        }
        // Fall back to extracting from text
        const headings = extractMarkdownHeadings(chunk.text);
        for (const h of headings) {
            allHeadings.add(h);
        }
    }
    return Array.from(allHeadings).sort();
}

async function generateSpecCatalog(
    repoRoot: string,
    outDir: string,
    manifest: FileManifestEntry[],
    chunks: ChunkEntry[]
): Promise<void> {
    let content = `# Specification Catalog

This catalog focuses on specification files in \`spec/**\`.

## Summary

`;

    const specFiles = manifest.filter(f => f.path.startsWith("spec/"));
    const specChunks = chunks.filter(c => c.path.startsWith("spec/"));

    content += `- **Total Spec Files**: ${specFiles.length}\n`;
    content += `- **Total Spec Chunks**: ${specChunks.length}\n\n`;
    content += `## Files\n\n`;
    content += `| Path | Chunks | Primary Topics |\n`;
    content += `|------|--------|----------------|\n`;

    // Sort spec files
    specFiles.sort((a, b) => a.path.localeCompare(b.path));

    for (const file of specFiles) {
        const fileChunks = chunks.filter(c => c.path === file.path);
        const headings = extractHeadingsFromChunks(fileChunks);

        // Infer primary topics from headings (first 3-5 main headings)
        const topics = headings
            .filter(h => {
                // Filter out very generic headings
                const lower = h.toLowerCase();
                return !["overview", "introduction", "table of contents", "contents"].includes(lower);
            })
            .slice(0, 5)
            .join(", ") || "General";

        content += `| \`${file.path}\` | ${file.chunks} | ${topics} |\n`;
    }

    await fs.writeFile(join(outDir, "SPEC_CATALOG.md"), content, "utf8");
}

export async function cmdDocs(repoRoot: string): Promise<void> {
    const aiContextDir = join(repoRoot, ".ai-context");
    const filesDir = join(aiContextDir, "files");
    const xrayDir = join(aiContextDir, "xray", "stagecraft", "data");
    const outDir = join(repoRoot, "docs", "generated", "ai-agent");

    // Read inputs
    const manifestPath = join(filesDir, "manifest.json");
    const chunksPath = join(filesDir, "chunks.ndjson");
    const xrayIndexPath = join(xrayDir, "index.json");

    const manifest = await readJSON<FileManifestEntry[]>(manifestPath);
    const chunks = await readNDJSON<ChunkEntry>(chunksPath);
    const xrayIndex = await readJSON<XrayIndex>(xrayIndexPath);

    if (!manifest || manifest.length === 0) {
        throw new Error(`No manifest found at ${manifestPath}. Run 'stagecraft context build' first.`);
    }

    // Ensure output directory exists
    await fs.mkdir(outDir, { recursive: true });

    // Generate all docs
    await generateReadme(repoRoot, outDir);
    await generateRepoIndex(repoRoot, outDir, xrayIndex);
    await generateDocsCatalog(repoRoot, outDir, manifest, chunks);
    await generateSpecCatalog(repoRoot, outDir, manifest, chunks);

    // --- New deterministic pages ---
    await generateNavigationPlan(outDir);
    await generateDocsBacklog(outDir);
    await generateCommandCatalog(repoRoot, outDir, manifest, chunks);
    await generateProviderCatalog(repoRoot, outDir, manifest, chunks);
    await generateCoreSpecIndex(repoRoot, outDir, manifest, chunks);
    await generateGovernanceIndex(repoRoot, outDir, manifest, chunks);
    await generateEngineAnalysisIndex(repoRoot, outDir, manifest, chunks);
    await generateEngineOutlineIndex(repoRoot, outDir, manifest, chunks);
    await generateContextHandoffIndex(repoRoot, outDir, manifest, chunks);
    await generateCoverageIndex(repoRoot, outDir, manifest, chunks);
    await generateArchiveIndex(repoRoot, outDir, manifest, chunks);
    await generateGuidesAndNarrativesIndex(repoRoot, outDir, manifest, chunks);

    console.log(`✅ Generated AI-Agent docs → ${outDir}`);
}


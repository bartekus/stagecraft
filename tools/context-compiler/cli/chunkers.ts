// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

import { basename, relative  } from "node:path";
import { countLines, fileSha, fileSize, readText, sha1 } from "./util";

import type { BuildOpts, ChunkRec, FileIngestPayload } from "./types";

export async function buildPayloadForFile(opts: BuildOpts): Promise<FileIngestPayload | null> {
    const { repoSlug, absPath, relRoot } = opts;

    // Use relRoot for stable repo-relative paths regardless of cwd
    const rel = normalizePath(relative(relRoot, absPath));
    if (rel.startsWith("..")) {
        // if file is outside root (e.g., from docs dir call), still accept but keep nice path
    }

    const text = await readText(absPath);
    const fsha = await fileSha(absPath);
    const fsize = await fileSize(absPath);

    // Route by extension and directory
    if (rel.endsWith(".md")) {
        return buildMarkdownPayload(repoSlug, rel, text, fsha, fsize);
    }
    if (rel.endsWith(".json")) {
        if (rel.includes("/graphs/")) return buildGraphJsonPayload(repoSlug, rel, text, fsha, fsize);
        if (rel.includes("/packages/")) return buildPackageJsonPayload(repoSlug, rel, text, fsha, fsize);
        // fallback: treat as one JSON chunk
        return singleBlobPayload(repoSlug, rel, text, fsha, fsize, "json", { kind: "json" });
    }

    // any other file → single chunk (plain text)
    return singleBlobPayload(repoSlug, rel, text, fsha, fsize, "text", { kind: "blob" });
}

function normalizePath(p: string) {
    return p.replaceAll("\\", "/");
}

/** Markdown → split by H2/H3 headings, compute line ranges */
function buildMarkdownPayload(repoSlug: string, path: string, text: string, sha: string, sizeBytes: number): FileIngestPayload {
    const lines = text.split(/\r?\n/);
    const chunks: ChunkRec[] = [];

    let startLine = 1; // 1-indexed line number
    let sectionTitle: string | undefined;

    for (let i = 0; i < lines.length; i++) {
        const l = lines[i];
        const isH2 = l.startsWith("## ");
        const isH3 = l.startsWith("### ");
        const isBoundary = isH2 || isH3;

        if (isBoundary && i > 0) {
            // Previous section ends at the line before this heading (i is 0-indexed, so line i+1 is the heading)
            // Previous section ends at line i (1-indexed), which is index i-1 (0-indexed)
            const endLine = i; // 1-indexed: line before the heading
            // Extract text from startLine-1 (0-indexed) to endLine (0-indexed, exclusive)
            const body = lines.slice(startLine - 1, endLine).join("\n");
            if (body.trim()) {
                chunks.push({
                    lang: "markdown",
                    symbol: sectionTitle,
                    kind: "section",
                    startLine: startLine,
                    endLine: endLine,
                    sha: sha1(body),
                    meta: { path, section: sectionTitle },
                    text: body,
                });
            }
            // New section starts at the heading line (i+1, 1-indexed)
            startLine = i + 1;
            sectionTitle = l.replace(/^###?\s+/, "").trim();
        } else if (isBoundary && i === 0) {
            // first heading defines sectionTitle but no previous chunk to flush
            sectionTitle = l.replace(/^###?\s+/, "").trim();
            startLine = 1; // heading is at line 1
        }
    }

    // flush tail - includes from startLine to end of file
    if (startLine <= lines.length) {
        const endLine = lines.length; // 1-indexed: last line
        const tail = lines.slice(startLine - 1).join("\n");
        if (tail.trim()) {
            chunks.push({
                lang: "markdown",
                symbol: sectionTitle,
                kind: "section",
                startLine: startLine,
                endLine: endLine,
                sha: sha1(tail),
                meta: { path, section: sectionTitle },
                text: tail,
            });
        }
    }

    // If no headings at all, one chunk
    if (chunks.length === 0) {
        const totalLines = lines.length;
        chunks.push({
            lang: "markdown",
            startLine: 1,
            endLine: totalLines,
            sha: sha1(text),
            meta: { path },
            text,
        });
    }

    return {
        repoSlug,
        path,
        sha,
        lang: "markdown",
        sizeBytes,
        chunks,
    };
}

/** Graph JSON → per-node chunks + an overview.
 *  We cannot reliably map to real JSON line numbers, so we assign synthetic line ranges (index-based).
 */
function buildGraphJsonPayload(repoSlug: string, path: string, raw: string, sha: string, sizeBytes: number): FileIngestPayload {
    let json: any = {};
    try { json = JSON.parse(raw); } catch { /* keep empty */ }

    const nodes = Array.isArray(json?.nodes) ? json.nodes : [];
    const edges = Array.isArray(json?.edges) ? json.edges : [];

    const chunks: ChunkRec[] = [];
    let line = 1;

    // Node chunks
    for (const n of nodes) {
        const name = n?.name || n?.id || "node";
        const neighbors = edges
            .filter((e: any) => e.from === n.id || e.to === n.id)
            .map((e: any) => e.from === n.id ? e.to : e.from);

        const body = [
            `# Graph Node: ${name}`,
            n.path ? `Path: ${n.path}` : "",
            n.kind ? `Kind: ${n.kind}` : "",
            neighbors?.length ? `Neighbors (${neighbors.length}): ${neighbors.join(", ")}` : "",
            n.summary ? `Summary:\n${n.summary}` : "",
        ].filter(Boolean).join("\n");

        const startLine = line;
        const endLine = line; // synthetic
        line++;

        chunks.push({
            lang: "json",
            symbol: String(name),
            kind: "graph-node",
            startLine,
            endLine,
            sha: sha1(body),
            meta: { path, nodeId: n?.id, neighborsCount: neighbors?.length ?? 0 },
            text: body,
        });
    }

    // Overview chunk
    const overview = `# Graph Overview\nNodes: ${nodes.length}\nEdges: ${edges.length}`;
    chunks.push({
        lang: "json",
        symbol: "graph-overview",
        kind: "graph-overview",
        startLine: line,
        endLine: line,
        sha: sha1(overview),
        meta: { path },
        text: overview,
    });

    return {
        repoSlug,
        path,
        sha,
        lang: "json",
        sizeBytes,
        chunks,
    };
}

/** Package JSON → single concise card. */
function buildPackageJsonPayload(repoSlug: string, path: string, raw: string, sha: string, sizeBytes: number): FileIngestPayload {
    let pkg: any = {};
    try { pkg = JSON.parse(raw); } catch {}

    const name = pkg?.name || pkg?.package || pkg?.id || basename(path);
    const deps = Object.keys(pkg?.dependencies ?? {}).slice(0, 200);
    const scripts = Object.keys(pkg?.scripts ?? {}).slice(0, 100);
    const targets = pkg?.targets || pkg?.platforms || pkg?.os || [];
    const files = (pkg?.files || pkg?.paths || []).slice?.(0, 100) ?? [];

    const text = [
        `# Package: ${name}`,
        pkg?.version ? `Version: ${pkg.version}` : "",
        deps.length ? `Dependencies (${deps.length}): ${deps.join(", ")}` : "",
        scripts.length ? `Scripts: ${scripts.join(", ")}` : "",
        targets.length ? `Targets: ${JSON.stringify(targets)}` : "",
        files.length ? `Files: ${files.join(", ")}` : "",
    ].filter(Boolean).join("\n");

    const lines = countLines(raw);

    return {
        repoSlug,
        path,
        sha,
        lang: "json",
        sizeBytes,
        chunks: [{
            lang: "json",
            kind: "package",
            symbol: String(name),
            startLine: 1,
            endLine: Math.max(1, lines),
            sha: sha1(text),
            meta: { path, name, hasScripts: scripts.length > 0, depsCount: deps.length },
            text,
        }],
    };
}

function singleBlobPayload(repoSlug: string, path: string, text: string, sha: string, sizeBytes: number, lang: string, meta?: Record<string, any>): FileIngestPayload {
    return {
        repoSlug,
        path,
        sha,
        lang,
        sizeBytes,
        chunks: [{
            lang,
            startLine: 1,
            endLine: countLines(text),
            sha: sha1(text),
            meta: { path, ...meta },
            text,
        }],
    };
}

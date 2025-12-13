// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

import { resolve, join, basename, relative } from "node:path";
import { promises as fs } from "node:fs";
import { createHash } from "node:crypto";
import { buildPayloadForFile } from "./chunkers";
import { SCHEMA_VERSION, type ContextMeta, type FileManifestEntry, type ChunkEntry } from "./types";
import { cmdDocs } from "./docs";

interface Args { [k: string]: string | boolean | undefined }
type Cmd = "build" | "docs";

function parseArgs(): { cmd: Cmd; args: Args } {
    const [cmd, ...rest] = process.argv.slice(2);
    if (!cmd || cmd === "build") {
        const args: Args = {};
        for (let i = 0; i < rest.length; i++) {
            const a = rest[i];
            if (a.startsWith("--")) {
                const k = a.slice(2);
                const v = rest[i + 1] && !rest[i + 1].startsWith("--") ? rest[++i] : true;
                args[k] = v as any;
            }
        }
        return { cmd: "build", args };
    }
    if (cmd === "docs") {
        return { cmd: "docs", args: {} };
    }
    console.error(`Usage:\n  build [--target <path>] [--out <path>] [--include <csv>] [--ext <csv>]\n  docs`);
    process.exit(1);
}

async function collectFiles(
    repoRoot: string,
    includes: string[],
    exts: string[]
): Promise<string[]> {
    const files: string[] = [];

    for (const inc of includes) {
        const absPath = resolve(repoRoot, inc);
        try {
            const stat = await fs.stat(absPath);
            if (stat.isFile()) {
                files.push(absPath);
            } else if (stat.isDirectory()) {
                for await (const f of walkDir(absPath)) {
                    files.push(f);
                }
            }
        } catch {
            // Skip if doesn't exist
        }
    }

    // Filter by extension
    const filtered = files.filter(f => {
        const ext = f.substring(f.lastIndexOf("."));
        return exts.includes(ext);
    });

    // Sort lexicographically by repo-relative path (posix)
    const repoRel = filtered.map(f => ({
        abs: f,
        rel: normalizePath(relative(repoRoot, f))
    }));
    repoRel.sort((a, b) => a.rel.localeCompare(b.rel, "en", { numeric: true }));

    return repoRel.map(r => r.abs);
}

async function* walkDir(dir: string): AsyncGenerator<string> {
    const ents = await fs.readdir(dir, { withFileTypes: true }).catch(() => []);
    for (const e of ents) {
        const p = join(dir, e.name);
        if (e.isDirectory()) {
            yield* walkDir(p);
        } else {
            yield p;
        }
    }
}

function normalizePath(p: string): string {
    return p.replaceAll("\\", "/");
}

function computeDigest(
    compilerVersion: string,
    options: { include: string[]; ext: string[] },
    filePayloads: Array<{ path: string; sha: string }>
): string {
    const h = createHash("sha256");
    h.update(compilerVersion);
    // Sort arrays for order-invariant hashing
    const sortedOptions = {
        include: [...options.include].sort((a, b) => a.localeCompare(b, "en", { numeric: true })),
        ext: [...options.ext].sort((a, b) => a.localeCompare(b, "en", { numeric: true }))
    };
    h.update(JSON.stringify(sortedOptions));
    for (const fp of filePayloads.sort((a, b) => a.path.localeCompare(b.path))) {
        h.update(fp.path);
        h.update(fp.sha);
    }
    return h.digest("hex");
}

async function cmdBuild(args: Args) {
    // Default target: if we're in tools/context-compiler, go up to repo root
    let defaultTarget = ".";
    const cwd = process.cwd();
    if (basename(cwd) === "context-compiler" && basename(resolve(cwd, "..")) === "tools") {
        defaultTarget = resolve(cwd, "../..");
    } else if (!args.target) {
        defaultTarget = ".";
    }

    const target = args.target ? String(args.target) : defaultTarget;
    const outDir = String(args.out ?? ".ai-context");
    const includeStr = String(args.include ?? "docs,spec,README.md,Agent.md");
    const extStr = String(args.ext ?? ".md,.json");

    const includes = includeStr.split(",").map(s => s.trim()).filter(Boolean);
    const exts = extStr.split(",").map(s => s.trim()).filter(Boolean);

    // Resolve paths
    const repoRoot = resolve(target);
    const repoSlug = sanitizeSlug(basename(repoRoot));
    const outAbs = join(repoRoot, outDir);

    // Collect files
    const files = await collectFiles(repoRoot, includes, exts);
    if (files.length === 0) {
        console.warn("No files found matching criteria.");
        return;
    }

    console.log(`Processing ${files.length} file(s)...`);

    // Build payloads
    const filePayloads: Array<{ path: string; sha: string; chunks: any[] }> = [];
    for (const absPath of files) {
        const payload = await buildPayloadForFile({ repoSlug, absPath, relRoot: repoRoot });
        if (payload) {
            filePayloads.push({
                path: payload.path,
                sha: payload.sha,
                chunks: payload.chunks
            });
        }
    }

    // Sort for determinism
    filePayloads.sort((a, b) => a.path.localeCompare(b.path));

    // Compute digest
    const digest = computeDigest(SCHEMA_VERSION, { include: includes, ext: exts },
        filePayloads.map(fp => ({ path: fp.path, sha: fp.sha })));

    // Prepare manifest
    const manifest: FileManifestEntry[] = filePayloads.map(fp => ({
        path: fp.path,
        sha: fp.sha,
        chunks: fp.chunks.length
    }));

    // Prepare chunks (sorted by path, then startLine, then sha)
    const allChunks: ChunkEntry[] = [];
    for (const fp of filePayloads) {
        for (const chunk of fp.chunks) {
            allChunks.push({
                repoSlug,
                path: fp.path,
                kind: chunk.kind || "section",
                startLine: chunk.startLine,
                endLine: chunk.endLine,
                sha: chunk.sha,
                text: chunk.text,
                meta: chunk.meta || {}
            });
        }
    }
    allChunks.sort((a, b) => {
        if (a.path !== b.path) return a.path.localeCompare(b.path);
        if (a.startLine !== b.startLine) return a.startLine - b.startLine;
        return a.sha.localeCompare(b.sha);
    });

    // Write outputs
    await fs.mkdir(join(outAbs, "files"), { recursive: true });

    // Write manifest.json
    await fs.writeFile(
        join(outAbs, "files", "manifest.json"),
        JSON.stringify(manifest, null, 2),
        "utf8"
    );

    // Write chunks.ndjson
    const chunksLines = allChunks.map(c => JSON.stringify(c)).join("\n");
    await fs.writeFile(
        join(outAbs, "files", "chunks.ndjson"),
        chunksLines,
        "utf8"
    );

    // Write digest.txt
    await fs.writeFile(
        join(outAbs, "digest.txt"),
        digest,
        "utf8"
    );

    // Write meta.json
    const meta: ContextMeta = {
        schemaVersion: SCHEMA_VERSION,
        repoSlug,
        targetRoot: ".",
        options: {
            include: includes,
            ext: exts
        },
        counts: {
            files: filePayloads.length,
            chunks: allChunks.length
        },
        digest,
        generatedAt: new Date().toISOString()
    };
    await fs.writeFile(
        join(outAbs, "meta.json"),
        JSON.stringify(meta, null, 2),
        "utf8"
    );

    console.log(`✅ Built context: ${filePayloads.length} files, ${allChunks.length} chunks`);
    console.log(`   Output: ${outAbs}`);
}

function sanitizeSlug(s: string): string {
    return s.replace(/[^a-zA-Z0-9._-]+/g, "-").replace(/^-+|-+$/g, "") || "repo";
}

async function cmdDocsHandler() {
    // Default target: if we're in tools/context-compiler, go up to repo root
    let defaultTarget = ".";
    const cwd = process.cwd();
    if (basename(cwd) === "context-compiler" && basename(resolve(cwd, "..")) === "tools") {
        defaultTarget = resolve(cwd, "../..");
    } else {
        defaultTarget = ".";
    }
    const repoRoot = resolve(defaultTarget);
    return cmdDocs(repoRoot);
}

(async function main() {
    const { cmd, args } = parseArgs();
    if (cmd === "build") return cmdBuild(args);
    if (cmd === "docs") return cmdDocsHandler();
})().catch((e) => { console.error(e); process.exit(1); });

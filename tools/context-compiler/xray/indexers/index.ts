#!/usr/bin/env node

// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

/**
 * Tiny repo indexer (TS). Walks the tree, applies ignore rules,
 * computes quick stats, detects modules, writes data/index.json.
 */
import { promises as fs } from "node:fs";
import * as path from "node:path";
import { createHash } from "node:crypto";

type FileEntry = {
    path: string;            // posix-relative
    size: number;            // bytes
    lines: number;           // crude LOC
    ext: string;             // like ".ts"
};

type ModuleHit =
    | { kind: "npm"; file: string; name?: string }
    | { kind: "go"; file: string; module?: string }
    | { kind: "cargo"; file: string; name?: string }
    | { kind: "git"; file: string };

type Index = {
    root: string;
    indexedAt: string;
    files: FileEntry[];
    languages: Record<string, number>; // ext -> bytes
    moduleFiles: ModuleHit[];
    topDirs: Record<string, number>;   // dirname -> bytes
    digest: string;
};

const ROOT = path.resolve(process.argv[2] || ".");
const OUT = path.resolve(ROOT, "data/index.json");
const IGNORE_FILE = path.resolve(ROOT, "tools/context-compiler/xray/ignore.rules");

const DEFAULT_IGNORES = [
    ".git","node_modules","dist","build","out","target","vendor",
    ".cache",".tmp","coverage"
];

async function readIgnores() {
    try {
        const t = await fs.readFile(IGNORE_FILE, "utf8");
        return [
            ...DEFAULT_IGNORES,
            ...t.split(/\r?\n/).map(s => s.trim()).filter(Boolean)
        ];
    } catch {
        return DEFAULT_IGNORES;
    }
}

function shouldIgnore(rel: string, ignores: string[]) {
    const parts = rel.split(path.sep);
    return parts.some(p =>
        ignores.some(rule => rule === p || (rule.endsWith("*") && p.startsWith(rule.slice(0, -1))))
    );
}

async function walk(dir: string, ignores: string[], acc: FileEntry[] = [], base = ROOT): Promise<FileEntry[]> {
    const entries = await fs.readdir(dir, { withFileTypes: true });
    for (const e of entries) {
        const abs = path.join(dir, e.name);
        const rel = path.relative(base, abs);
        if (shouldIgnore(rel, ignores)) continue;
        if (e.isDirectory()) {
            await walk(abs, ignores, acc, base);
        } else if (e.isFile()) {
            try {
                const st = await fs.stat(abs);
                const content = await fs.readFile(abs, "utf8").catch(() => Buffer.alloc(0));
                const lines = content.length ? content.split(/\r?\n/).length : 0;
                acc.push({
                    path: rel.split(path.sep).join("/"),
                    size: st.size,
                    lines,
                    ext: path.extname(e.name).toLowerCase()
                });
            } catch { /* skip unreadable */ }
        }
    }
    return acc;
}

async function detectModules(root: string): Promise<ModuleHit[]> {
    const hits: ModuleHit[] = [];
    async function maybe(file: string, fn: (txt: string) => void) {
        try { const t = await fs.readFile(path.join(root, file), "utf8"); fn(t); } catch {}
    }
    await maybe("package.json", t => {
        try { hits.push({ kind: "npm", file: "package.json", name: JSON.parse(t).name }); } catch {}
    });
    await maybe("go.mod", t => {
        const m = t.match(/^module\s+([^\s]+)\s*$/m);
        hits.push({ kind: "go", file: "go.mod", module: m?.[1] });
    });
    await maybe("Cargo.toml", t => {
        const m = t.match(/^\s*name\s*=\s*"(.*?)"\s*$/m);
        hits.push({ kind: "cargo", file: "Cargo.toml", name: m?.[1] });
    });
    try { await fs.access(path.join(root, ".git")); hits.push({ kind: "git", file: ".git" }); } catch {}
    return hits;
}

function summarize(files: FileEntry[]) {
    const languages: Record<string, number> = {};
    const topDirs: Record<string, number> = {};
    for (const f of files) {
        const dir = f.path.includes("/") ? f.path.split("/")[0] : ".";
        topDirs[dir] = (topDirs[dir] || 0) + f.size;
        languages[f.ext || ""] = (languages[f.ext || ""] || 0) + f.size;
    }
    return { languages, topDirs };
}

(async () => {
    const ignores = await readIgnores();
    const files = await walk(ROOT, ignores);
    const { languages, topDirs } = summarize(files);
    const moduleFiles = await detectModules(ROOT);
    const digest = createHash("sha256")
        .update(JSON.stringify({ files, languages, topDirs, moduleFiles })).digest("hex").slice(0, 16);

    const index: Index = {
        root: path.basename(ROOT),
        indexedAt: new Date().toISOString(),
        files, languages, topDirs, moduleFiles, digest
    };

    await fs.mkdir(path.join(ROOT, "data"), { recursive: true });
    await fs.writeFile(OUT, JSON.stringify(index, null, 2), "utf8");
    console.log(`Wrote ${path.relative(ROOT, OUT)} (${files.length} files)`);
})();

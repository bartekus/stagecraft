#!/usr/bin/env node
// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

/**
 * XRAY v1 — Multi-language reverse-engineering scanner.
 *
 * What’s new vs v0:
 *  - .gitignore merge (root + nested) + tools/context-compiler/xray/ignore.rules
 *  - On-disk cache: .xraycache/entries.json  (size + mtimeMs → cached LOC)
 *  - Hotlists: docs/hotlist.md (largest files/dirs, binaries-ish, generated pockets)
 *  - Encore-aware map: docs/encore-map.md when signals are detected
 *  - Flags: --cache on|off (default on), --hotlist-limit N (default 20), --deep=node,go,rust (stubs)
 *
 * Usage:
 *   npx tsx xray/xray.ts all <targetDir> [--json ./data] [--out ./docs] [--cache on] [--hotlist-limit 20] [--deep node,go,rust]
 *   npx tsx xray/xray.ts scan <targetDir> ...
 *   npx tsx xray/xray.ts docs ...
 */

import { spawn } from "node:child_process";
import { promises as fs } from "node:fs";
import * as path from "node:path";
import { createHash } from "node:crypto";

// ----------------------------- Types ---------------------------------

type PkgKind = "node" | "go" | "rust";

type FileEntry = {
    path: string;    // pkg-relative (posix)
    size: number;    // bytes
    lines: number;   // cached or measured LOC
    ext: string;     // ".ts", ".go", ".rs", ...
};

type PackageSummary = {
    kind: PkgKind;
    root: string;          // path relative to scanRoot
    name?: string;         // node/rust
    version?: string;      // node
    module?: string;       // go
    workspaceRoot?: string;// rust workspace
    files: number;
    bytes: number;
    languages: Record<string, number>;
    topDirs: Record<string, number>;
    // doc-side conveniences:
    biggestFiles?: { path: string; size: number }[];
};

type Index = {
    scannedAt: string;
    target: string;
    packages: PackageSummary[];
    stats: { files: number; bytes: number; languages: Record<string, number> };
    digest: string;
};

// --------------------------- CLI Args --------------------------------

const argv = process.argv.slice(2);
if (!argv.length) {
    console.error("Usage: xray {all|scan|docs} <targetDir?> [--json ./<repoName>/data] [--out ./<repoName>/docs] [--cache on|off] [--hotlist-limit 20] [--deep node,go,rust]");
    process.exit(2);
}

function takeFlag(name: string, def?: string) {
    const i = argv.findIndex((a) => a === name);
    if (i >= 0) {
        const v = argv[i + 1];
        argv.splice(i, v ? 2 : 1);
        return v ?? def;
    }
    return def;
}
function takeBoolFlag(name: string, def: boolean) {
    const v = takeFlag(name);
    if (v == null) return def;
    return /^(1|on|true|yes)$/i.test(v);
}

function hasFlag(name: string) {
    const i = argv.findIndex(a => a === name);
    if (i >= 0) { argv.splice(i, 1); return true; }
    return false;
}

const EXPORTS = hasFlag("--exports");
const CHURN_WINDOW = takeFlag("--churn", "");   // e.g. "90d", "12w", "2025-01-01"

const cmd = argv.shift()!;
const targetArg = argv[0] && !argv[0].startsWith("--") ? argv.shift()! : "";

// Extract --json and --out flags before they're consumed
const jsonFlag = takeFlag("--json", "");
const outFlag = takeFlag("--out", "");

// Find repo root: walk up from cwd or targetArg until we find .git/, spec/, or Agent.md
async function findRepoRoot(startDir: string): Promise<string> {
    let current = path.resolve(startDir);
    const root = path.parse(current).root;

    while (current !== root) {
        const gitDir = path.join(current, ".git");
        const specDir = path.join(current, "spec");
        const agentMd = path.join(current, "Agent.md");

        if (await exists(gitDir) || await exists(specDir) || await exists(agentMd)) {
            return current;
        }

        const parent = path.dirname(current);
        if (parent === current) break;
        current = parent;
    }

    // Fallback: if we're in tools/context-compiler, go up two levels
    const cwd = process.cwd();
    if (path.basename(cwd) === "context-compiler" && path.basename(path.dirname(cwd)) === "tools") {
        return path.resolve(cwd, "../..");
    }

    // Last resort: use the start directory
    return path.resolve(startDir);
}

// Derive a repo slug from the target directory (or best-effort hints).
function toSlug(s: string) {
      return s.replace(/\\/g, "/")
        .replace(/^.*\//, "")        // basename
        .replace(/\.git$/, "")
        .replace(/[^a-zA-Z0-9._-]+/g, "-")
        .replace(/^-+|-+$/g, "")
        || "repo";
}
function deriveRepoSlug(targetDir: string): string {
      if (!targetDir) return "repo";
      const abs = path.resolve(process.cwd(), targetDir);
      const base = toSlug(path.basename(abs));
      return base || "repo";
}

// Resolve repo root and slug (will be set in main after async findRepoRoot)
let REPO_ROOT = "";
let REPO_SLUG = "";
let JSON_DIR = "";
let DOCS_DIR = "";
let CACHE_DIR = "";

const CACHE_ON = takeBoolFlag("--cache", true);
const HOTLIST_LIMIT = parseInt(takeFlag("--hotlist-limit", "20")!, 10);
const DEEP = (takeFlag("--deep", "") || "")
    .split(",")
    .map(s => s.trim())
    .filter(Boolean) as ("node"|"go"|"rust")[];

// ------------------------- Utilities ---------------------------------

const DEFAULT_IGNORES = new Set([
    ".git", "node_modules", "dist", "build", "out", "vendor", "target",
    ".cache", ".tmp", "coverage",
    // XRAY outputs/cache (avoid self-scan & churn)
    ".xraycache", ".ai-context"
]);
const DEFAULT_GENERATED_HINTS = [
    "**/*.gen.*", "**/*.pb.*", "**/*.pb.*.go", "**/*.g.dart", "**/__generated__/**",
    "**/generated/**", "**/gen/**"
];

const BIG_FILE_CAP_BYTES = 2 * 1024 * 1024; // 2 MB: above this, skip LOC count

function toPosix(p: string) { return p.split(path.sep).join("/"); }
function fmtBytes(n: number) {
    const u = ["B","KB","MB","GB"]; let i = 0, v = n;
    while (v >= 1024 && i < u.length - 1) { v /= 1024; i++; }
    return `${v.toFixed(1)} ${u[i]}`;
}

// ---------------------- Ignore matcher (.gitignore + rules) ----------

type IgnoreSpec = { dir: string; patterns: string[] }; // dir with its .gitignore patterns (posix)

function globToRegExp(glob: string): RegExp {
    let g = glob.replace(/\\/g, "/");
    g = g.replace(/[.+^${}()|[\]\\]/g, "\\$&");
    g = g.replace(/\*\*/g, "§§DOUBLESTAR§§");
    g = g.replace(/\*/g, "[^/]*");
    g = g.replace(/\?/g, "[^/]");
    g = g.replace(/§§DOUBLESTAR§§/g, ".*");
    return new RegExp("^" + g + "$");
}

async function collectIgnoreSpecs(scanRoot: string): Promise<IgnoreSpec[]> {
    const specs: IgnoreSpec[] = [];
    // Resolve relative to the repo root so this works no matter where XRAY is invoked from.
    const localRules = path.join(REPO_ROOT || process.cwd(), "tools", "context-compiler", "xray", "ignore.rules");
    if (await exists(localRules)) {
        const t = await fs.readFile(localRules, "utf8");
        const patterns = t.split(/\r?\n/).map(s => s.trim()).filter(Boolean);
        specs.push({ dir: scanRoot, patterns });
    }
    // Walk for .gitignore
    const dirs = await listAllDirs(scanRoot);
    for (const d of [scanRoot, ...dirs]) {
        const gi = path.join(d, ".gitignore");
        if (await exists(gi)) {
            const t = await fs.readFile(gi, "utf8");
            const raw = t.split(/\r?\n/).map(s => s.replace(/\s+#.*$/, "").trim()).filter(Boolean);
            specs.push({ dir: d, patterns: raw });
        }
    }
    return specs;
}

function buildIgnoreTester(specs: IgnoreSpec[]) {
    // Pre-compile patterns into regex with base dir
    const compiled = specs.flatMap(s => s.patterns.map(p => {
        // gitignore semantics: patterns are relative to file location, but can have leading "/" to pin
        const rx = globToRegExp(toPosix(p));
        const base = toPosix(s.dir);
        return { base, rx, raw: p };
    }));

    return (absPath: string) => {
        const norm = toPosix(absPath);
        // check default directory name ignores first (quick short-circuit)
        const parts = norm.split("/");
        if (parts.some(p => DEFAULT_IGNORES.has(p))) return true;

        for (const { base, rx } of compiled) {
            if (norm.startsWith(base)) {
                const rel = norm.slice(base.length + (norm.length > base.length ? 1 : 0)); // base-relative
                if (rx.test(rel)) return true;
            }
        }
        return false;
    };
}

// ---------------------------- Cache ----------------------------------

type CacheEntry = { size: number; mtimeMs: number; lines: number };
type CacheDB = Record<string, CacheEntry>;
function getCacheFile(): string {
    return path.join(CACHE_DIR, "entries.json");
}

async function loadCache(): Promise<CacheDB> {
    if (!CACHE_ON) return {};
    try {
        const t = await fs.readFile(getCacheFile(), "utf8");
        return JSON.parse(t) as CacheDB;
    } catch { return {}; }
}
async function saveCache(db: CacheDB) {
    if (!CACHE_ON) return;
    await fs.mkdir(CACHE_DIR, { recursive: true });
    await fs.writeFile(getCacheFile(), JSON.stringify(db), "utf8");
}

// ---------------------------- FS helpers -----------------------------

async function exists(p: string) { try { await fs.access(p); return true; } catch { return false; } }

async function listDirs(root: string): Promise<string[]> {
    const out: string[] = [];
    const ent = await fs.readdir(root, { withFileTypes: true }).catch(() => []);
    for (const e of ent) if (e.isDirectory()) out.push(path.join(root, e.name));
    return out;
}
async function listAllDirs(root: string): Promise<string[]> {
    const out: string[] = [];
    async function rec(d: string) {
        const ent = await fs.readdir(d, { withFileTypes: true }).catch(() => []);
        for (const e of ent) {
            if (e.isDirectory()) {
                const abs = path.join(d, e.name);
                out.push(abs);
                await rec(abs);
            }
        }
    }
    await rec(root);
    return out;
}

// ---------------------- Deep analyzers (graphs) ----------------------

type NodeGraph = {
    packages: Array<{
        name: string;
        root: string;          // path relative to scan root
        private?: boolean;
        deps: Record<string, string>;        // name -> semver
        devDeps: Record<string, string>;     // name -> semver
        workspaceLinks: string[];            // other local packages this one depends on
    }>;
};

type GoGraph = {
    moduleRoot: string;      // path relative to scan root
    modulePath: string;      // go.mod module path
    requires: Array<{ path: string; version?: string }>;
    // lightweight import edges (import path -> module path) for local code
    imports: Array<{ from: string; to: string; module?: string }>;
};

type RustGraph = {
    workspaceRoot?: string;  // relative
    packages: Array<{
        name: string;
        id: string;
        root: string;          // relative
        kind: ("bin"|"lib"|"proc-macro"|"other")[];
        // edges among local workspace members (name -> name)
        localDeps: string[];
    }>;
};

// util
function run(cmd: string, args: string[], cwd: string): Promise<{ code: number; out: string; err: string }> {
    return new Promise((res) => {
        const child = spawn(cmd, args, { cwd, stdio: ["ignore", "pipe", "pipe"] });
        let out = ""; let err = "";
        child.stdout.on("data", d => out += String(d));
        child.stderr.on("data", d => err += String(d));
        child.on("close", code => res({ code: code ?? 0, out, err }));
    });
}

// ---------- Node (ts-morph-free basic: manifest + workspace wiring) -----------
// We keep this cheap by reading package.json; exports graph via ts-morph can be added later.
async function buildNodeGraph(scanRoot: string, pkgs: PackageSummary[]): Promise<NodeGraph> {
    const nodePkgs = pkgs.filter(p => p.kind === "node");
    const byRoot = new Map(nodePkgs.map(p => [p.root, p]));
    const byName = new Map<string, string>(); // name -> root
    for (const p of nodePkgs) if (p.name) byName.set(p.name, p.root);

    const out: NodeGraph = { packages: [] };
    for (const p of nodePkgs) {
        const abs = path.join(scanRoot, p.root, "package.json");
        const json = await safeJSON<any>(abs);
        const deps = json?.dependencies ?? {};
        const devDeps = json?.devDependencies ?? {};
        const workspaceLinks: string[] = [];
        for (const depName of Object.keys(deps)) {
            const maybeLocal = byName.get(depName);
            if (maybeLocal) workspaceLinks.push(maybeLocal);
        }
        out.packages.push({
            name: json?.name ?? p.root,
            root: p.root,
            private: json?.private ?? false,
            deps,
            devDeps,
            workspaceLinks,
        });
    }
    return out;
}

// ------------------------------- Go ----------------------------------
async function buildGoGraph(scanRoot: string, pkgs: PackageSummary[]): Promise<GoGraph[]> {
    const mods = pkgs.filter(p => p.kind === "go");
    const graphs: GoGraph[] = [];

    for (const m of mods) {
        const cwd = path.join(scanRoot, m.root);
        // module metadata
        const modules = await run("go", ["list", "-m", "-json", "all"], cwd);
        if (modules.code !== 0) continue;
        const requires: Array<{ path: string; version?: string }> = [];
        for (const blob of modules.out.split("\n}\n")) {
            try {
                const obj = JSON.parse(blob.endsWith("}") ? blob : blob + "}");
                if (obj.Path && obj.Version && !obj.Main) {
                    requires.push({ path: obj.Path, version: obj.Version });
                }
            } catch { /* skip non-json chunk */ }
        }
        // import edges (lightweight)
        const deps = await run("go", ["list", "-deps", "-f", "{{with .Module}}{{.Path}}{{end}} {{.ImportPath}}", "./..."], cwd);
        const imports: Array<{ from: string; to: string; module?: string }> = [];
        if (deps.code === 0) {
            for (const ln of deps.out.split("\n")) {
                const [modPath, impPath] = ln.trim().split(/\s+/, 2);
                if (!impPath) continue;
                imports.push({ from: impPath, to: modPath ? impPath : impPath, module: modPath || undefined });
            }
        }
        graphs.push({
            moduleRoot: m.root,
            modulePath: m.module || "",
            requires, imports
        });
    }
    return graphs;
}

// ------------------------------ Rust ---------------------------------
async function buildRustGraph(scanRoot: string, pkgs: PackageSummary[]): Promise<RustGraph[]> {
    // Group by workspaceRoot if present; standalone crates become their own group
    const rusts = pkgs.filter(p => p.kind === "rust");
    const groups = new Map<string, PackageSummary[]>();
    for (const r of rusts) {
        const key = r.workspaceRoot ?? r.root;
        const arr = groups.get(key) ?? [];
        arr.push(r);
        groups.set(key, arr);
    }

    const out: RustGraph[] = [];
    for (const [wsRoot, members] of groups) {
        const cwd = path.join(scanRoot, wsRoot);
        const meta = await run("cargo", ["metadata", "--format-version", "1", "--no-deps"], cwd);
        if (meta.code !== 0) continue;

        let obj: any;
        try { obj = JSON.parse(meta.out); } catch { continue; }

        const pkgs = obj.packages as any[];
        const resolve = obj.resolve;
        const nodeById = new Map<string, any>();
        if (resolve?.nodes) for (const n of resolve.nodes) nodeById.set(n.id, n);

        const byName = new Map<string, string>();
        for (const p of pkgs) byName.set(p.name, p.id);

        const records: RustGraph["packages"] = [];
        for (const p of pkgs) {
            // kind
            const kinds: ("bin"|"lib"|"proc-macro"|"other")[] = [];
            for (const t of p.targets ?? []) {
                for (const k of (t.kind ?? [])) {
                    if (k === "bin" || k === "lib" || k === "proc-macro") kinds.push(k);
                    else kinds.push("other");
                }
            }
            // local deps = intersection with same workspace
            const node = nodeById.get(p.id);
            const localDeps: string[] = [];
            for (const depId of (node?.deps ?? []).map((d: any) => d.pkg)) {
                const dep = pkgs.find(x => x.id === depId);
                if (!dep) continue;
                localDeps.push(dep.name);
            }
            records.push({
                name: p.name,
                id: p.id,
                root: toPosix(path.relative(scanRoot, p.manifest_path.replace(/Cargo\.toml$/, ""))),
                kind: Array.from(new Set(kinds)),
                localDeps: Array.from(new Set(localDeps)),
            });
        }
        out.push({ workspaceRoot: wsRoot !== members[0].root ? wsRoot : undefined, packages: records });
    }
    return out;
}

// ----------------------- Write graphs & docs -------------------------

async function writeGraphs(jsonDir: string, graphs: { node?: NodeGraph; go?: GoGraph[]; rust?: RustGraph[] }) {
    await fs.mkdir(path.join(jsonDir, "graphs"), { recursive: true });
    if (graphs.node) await fs.writeFile(path.join(jsonDir, "graphs", "node.json"), JSON.stringify(graphs.node, null, 2), "utf8");
    if (graphs.go)   await fs.writeFile(path.join(jsonDir, "graphs", "go.json"), JSON.stringify(graphs.go, null, 2), "utf8");
    if (graphs.rust) await fs.writeFile(path.join(jsonDir, "graphs", "rust.json"), JSON.stringify(graphs.rust, null, 2), "utf8");
}

function mermaidHeader(title: string) {
    return `### ${title}

\`\`\`mermaid
graph LR
`;
}
function mermaidFooter() { return "```\n"; }

// keep graphs readable: cap nodes/edges
const MAX_NODES = 40;
const MAX_EDGES = 80;

function truncate<T>(arr: T[], n: number): T[] {
    return arr.length > n ? arr.slice(0, n) : arr;
}

function safeNodeId(s: string) {
    return s.replace(/[^A-Za-z0-9_]/g, "_").slice(-60);
}

// Build Mermaid sections
function mermaidForNode(g: NodeGraph) {
    const lines: string[] = [mermaidHeader("Node Workspaces & Deps").trim()];
    const locals = new Map(g.packages.map(p => [p.name, p] as const));
    const edges: string[] = [];

    for (const p of truncate(g.packages, MAX_NODES)) {
        const pid = safeNodeId(p.name);
        lines.push(`${pid}["${p.name}"]`);
        for (const dep of truncate(Object.keys(p.deps ?? {}), 20)) {
            if (locals.has(dep)) {
                const nid = safeNodeId(dep);
                edges.push(`${pid} --> ${nid}`);
            }
        }
        for (const wk of truncate(p.workspaceLinks, 20)) {
            const name = Array.from(locals.values()).find(x => x.root === wk)?.name ?? wk;
            const nid = safeNodeId(name);
            edges.push(`${pid} -.-> ${nid}`);
        }
    }
    for (const e of truncate(edges, MAX_EDGES)) lines.push(e);
    lines.push(mermaidFooter());
    return lines.join("\n");
}

function mermaidForGo(gs: GoGraph[]) {
    const lines: string[] = [mermaidHeader("Go Modules (local edges)")];
    const nodes = new Set<string>();
    const edges: string[] = [];

    for (const g of gs) {
        const id = safeNodeId(g.modulePath || g.moduleRoot);
        nodes.add(`${id}["${g.modulePath || g.moduleRoot}"]`);
        // local import edges within module: compress by package prefix
        for (const e of truncate(g.imports, 100)) {
            if (!e.module || e.module === g.modulePath) continue; // local-only edges aren’t interesting here
            const tid = safeNodeId(e.module);
            nodes.add(`${tid}["${e.module}"]`);
            edges.push(`${id} --> ${tid}`);
        }
    }
    for (const n of truncate(Array.from(nodes), MAX_NODES)) lines.push(n);
    for (const e of truncate(edges, MAX_EDGES)) lines.push(e);
    lines.push(mermaidFooter());
    return lines.join("\n");
}

function mermaidForRust(gs: RustGraph[]) {
    const lines: string[] = [mermaidHeader("Rust Workspaces (local deps)")];
    const edges: string[] = [];
    const nodes = new Set<string>();

    for (const g of gs) {
        for (const p of truncate(g.packages, MAX_NODES)) {
            const id = safeNodeId(p.name);
            nodes.add(`${id}["${p.name}"]`);
            for (const d of truncate(p.localDeps, 12)) {
                const did = safeNodeId(d);
                edges.push(`${id} --> ${did}`);
            }
        }
    }
    for (const n of truncate(Array.from(nodes), MAX_NODES)) lines.push(n);
    for (const e of truncate(edges, MAX_EDGES)) lines.push(e);
    lines.push(mermaidFooter());
    return lines.join("\n");
}

async function writeDepsDoc(jsonDir: string, docsDir: string, graphs: { node?: NodeGraph; go?: GoGraph[]; rust?: RustGraph[] }) {
    const sections: string[] = ["# Dependency Map\n"];

    if (graphs.node) {
        sections.push(mermaidForNode(graphs.node));
    } else {
        sections.push("*(Node graph not enabled or no Node packages found.)*\n");
    }
    if (graphs.go?.length) {
        sections.push(mermaidForGo(graphs.go));
    } else {
        sections.push("*(Go graph not enabled or no Go modules found.)*\n");
    }
    if (graphs.rust?.length) {
        sections.push(mermaidForRust(graphs.rust));
    } else {
        sections.push("*(Rust graph not enabled or no Rust crates/workspaces found.)*\n");
    }

    // Tiny cross-view: show “local -> local” links across ecosystems by path name
    sections.push("\n### Cross-ecosystem notes (heuristic)\n- This section is intentionally light; deep cross-language edges need tooling bridges.\n");

    await fs.writeFile(path.join(docsDir, "deps.md"), sections.join("\n"), "utf8");
}

// ---------------------- Discovery (package roots) --------------------

type Marker = { file: "package.json" | "go.mod" | "Cargo.toml"; abs: string };

async function findMarkers(scanRoot: string, isIgnored: (abs: string)=>boolean): Promise<Marker[]> {
    const out: Marker[] = [];
    const walk = async (dir: string) => {
        const ent = await fs.readdir(dir, { withFileTypes: true }).catch(() => []);
        for (const e of ent) {
            const abs = path.join(dir, e.name);
            if (isIgnored(abs)) {
                if (e.isDirectory()) continue;
            }
            if (e.isDirectory()) {
                await walk(abs);
            } else if (e.isFile()) {
                if (e.name === "package.json" || e.name === "go.mod" || e.name === "Cargo.toml") {
                    out.push({ file: e.name as any, abs });
                }
            }
        }
    };
    await walk(scanRoot);
    return out;
}

type PkgRoot = { kind: PkgKind; rootAbs: string; meta?: any };

async function safeJSON<T=any>(p: string): Promise<T|null> {
    try { return JSON.parse(await fs.readFile(p, "utf8")); } catch { return null; }
}

async function resolveNodeWorkspaces(rootDir: string, ws: any): Promise<string[]> {
    let patterns: string[] = [];
    if (Array.isArray(ws)) patterns = ws as string[];
    else if (ws && Array.isArray(ws.packages)) patterns = ws.packages as string[];
    if (!patterns.length) return [];

    const all = await listAllDirs(rootDir);
    const rel = all.map(d => toPosix(path.relative(rootDir, d))).filter(Boolean);
    const rxs = patterns.map(globToRegExp);

    const out: string[] = [];
    for (const r of rel) {
        if (rxs.some(rx => rx.test(r))) {
            const abs = path.join(rootDir, r);
            if (await exists(path.join(abs, "package.json"))) out.push(abs);
        }
    }
    return out;
}

function parseCargoName(toml: string): string | undefined {
    const pkgIdx = toml.indexOf("[package]");
    const slab = pkgIdx >= 0 ? toml.slice(pkgIdx) : toml;
    const m = slab.match(/^\s*name\s*=\s*"(.*?)"\s*$/m);
    return m?.[1];
}
function parseCargoWorkspaceMembers(toml: string): string[] {
    const wsIdx = toml.indexOf("[workspace]");
    if (wsIdx < 0) return [];
    const tail = toml.slice(wsIdx);
    const arr = tail.match(/members\s*=\s*\[([\s\S]*?)\]/m);
    if (!arr) return [];
    return [...arr[1].matchAll(/"([^"]+)"/g)].map(m => m[1]);
}
async function globExpandUnder(rootDir: string, patterns: string[]): Promise<string[]> {
    const all = await listAllDirs(rootDir);
    const rel = all.map(d => toPosix(path.relative(rootDir, d))).filter(Boolean);
    const rxs = patterns.map(globToRegExp);
    return rel.filter(d => rxs.some(rx => rx.test(d))).map(d => path.join(rootDir, d));
}

async function discoverPackages(scanRoot: string, isIgnored: (abs: string)=>boolean): Promise<PkgRoot[]> {
    const markers = await findMarkers(scanRoot, isIgnored);
    const found: PkgRoot[] = [];

    // Node roots + workspaces
    for (const m of markers.filter(m => m.file === "package.json")) {
        const rootDir = path.dirname(m.abs);
        if (isIgnored(rootDir)) continue;
        const pkg = await safeJSON<any>(m.abs);
        if (!pkg) continue;
        found.push({ kind: "node", rootAbs: rootDir, meta: { name: pkg.name, version: pkg.version, private: pkg.private } });

        if (pkg.workspaces) {
            const wsDirs = await resolveNodeWorkspaces(rootDir, pkg.workspaces);
            for (const d of wsDirs) {
                if (isIgnored(d)) continue;
                const child = await safeJSON<any>(path.join(d, "package.json"));
                found.push({ kind: "node", rootAbs: d, meta: { name: child?.name, version: child?.version, private: child?.private } });
            }
        }
    }

    // Go modules
    for (const m of markers.filter(m => m.file === "go.mod")) {
        const rootDir = path.dirname(m.abs);
        if (isIgnored(rootDir)) continue;
        const txt = await fs.readFile(m.abs, "utf8").catch(() => "");
        const mod = (txt.match(/^module\s+([^\s]+)\s*$/m) || [])[1];
        found.push({ kind: "go", rootAbs: rootDir, meta: { module: mod } });
    }

    // Rust crates & workspaces
    for (const m of markers.filter(m => m.file === "Cargo.toml")) {
        const rootDir = path.dirname(m.abs);
        if (isIgnored(rootDir)) continue;
        const toml = await fs.readFile(m.abs, "utf8").catch(() => "");
        if (/\[workspace\]/.test(toml)) {
            const members = parseCargoWorkspaceMembers(toml);
            const memberAbs = await globExpandUnder(rootDir, members);
            for (const abs of memberAbs) {
                if (isIgnored(abs)) continue;
                const t = await fs.readFile(path.join(abs, "Cargo.toml"), "utf8").catch(() => "");
                found.push({ kind: "rust", rootAbs: abs, meta: { name: parseCargoName(t), workspaceRoot: rootDir } });
            }
        } else {
            found.push({ kind: "rust", rootAbs: rootDir, meta: { name: parseCargoName(toml) } });
        }
    }

    // De-dup by rootAbs
    const seen = new Set<string>();
    return found.filter(p => (seen.has(p.rootAbs) ? false : (seen.add(p.rootAbs), true)));
}

// ------------------------ Indexing (files → stats) -------------------

async function walkFiles(root: string, isIgnored: (abs: string)=>boolean): Promise<string[]> {
    const files: string[] = [];
    async function rec(dir: string) {
        const ent = await fs.readdir(dir, { withFileTypes: true }).catch(() => []);
        for (const e of ent) {
            const abs = path.join(dir, e.name);
            if (isIgnored(abs)) {
                if (e.isDirectory()) continue;
            }
            if (e.isDirectory()) {
                await rec(abs);
            } else if (e.isFile()) {
                files.push(abs);
            }
        }
    }
    await rec(root);
    return files;
}

async function fileEntry(abs: string, base: string, cache: CacheDB): Promise<FileEntry> {
    const st = await fs.stat(abs).catch(() => null as any);
    const rel = toPosix(path.relative(base, abs));
    const ext = path.extname(abs).toLowerCase();
    let lines = 0;

    if (st) {
        const key = abs;
        const prev = cache[key];
        if (prev && prev.size === st.size && Math.floor(prev.mtimeMs) === Math.floor(st.mtimeMs)) {
            lines = prev.lines;
        } else {
            if (st.size > 0 && st.size <= BIG_FILE_CAP_BYTES) {
                const txt = await fs.readFile(abs, "utf8").catch(() => "");
                lines = txt ? (txt.match(/\n/g)?.length ?? 0) + 1 : 0;
            } else {
                lines = 0;
            }
            cache[key] = { size: st.size, mtimeMs: st.mtimeMs, lines };
        }
        return { path: rel, size: st.size, lines, ext };
    }
    return { path: rel, size: 0, lines: 0, ext };
}

function summarize(files: FileEntry[]) {
    const languages: Record<string, number> = {};
    const topDirs: Record<string, number> = {};
    let bytes = 0;

    for (const f of files) {
        languages[f.ext || ""] = (languages[f.ext || ""] || 0) + f.size;
        const top = f.path.includes("/") ? f.path.split("/")[0] : ".";
        topDirs[top] = (topDirs[top] || 0) + f.size;
        bytes += f.size;
    }
    return { files: files.length, bytes, languages, topDirs };
}

function largestFiles(files: FileEntry[], limit: number) {
    return [...files].sort((a,b)=>b.size-a.size).slice(0, limit).map(f => ({ path: f.path, size: f.size }));
}
function largestDirs(files: FileEntry[], limit: number) {
    const agg: Record<string, { bytes: number; files: number }> = {};
    for (const f of files) {
        const dir = f.path.includes("/") ? f.path.split("/").slice(0, -1).join("/") : ".";
        agg[dir] ??= { bytes: 0, files: 0 };
        agg[dir].bytes += f.size;
        agg[dir].files++;
    }
    return Object.entries(agg).sort((a,b)=>b[1].bytes - a[1].bytes).slice(0, limit)
        .map(([dir, s]) => ({ dir, bytes: s.bytes, files: s.files }));
}

function detectGenerated(files: FileEntry[]) {
    const rxes = DEFAULT_GENERATED_HINTS.map(globToRegExp);
    return files.filter(f => rxes.some(rx => rx.test(f.path)));
}
function likelyBinaryByExt(ext: string) {
    return [".png",".jpg",".jpeg",".gif",".webp",".ico",".pdf",".wasm",".zip",".gz",".xz",".tgz",".mp4",".mov",".mp3",".wav"].includes(ext);
}

async function indexPackage(scanRoot: string, pkg: PkgRoot, isIgnored: (abs: string)=>boolean, cache: CacheDB): Promise<PackageSummary & { _files: FileEntry[] }> {
    const filesAbs = await walkFiles(pkg.rootAbs, isIgnored);
    const entries = await Promise.all(filesAbs.map((a) => fileEntry(a, pkg.rootAbs, cache)));
    const sum = summarize(entries);
    const relRoot = toPosix(path.relative(scanRoot, pkg.rootAbs)) || ".";
    const base: PackageSummary & { _files: FileEntry[] } = {
        kind: pkg.kind,
        root: relRoot,
        name: pkg.meta?.name,
        version: pkg.meta?.version,
        module: pkg.meta?.module,
        workspaceRoot: pkg.meta?.workspaceRoot ? toPosix(path.relative(scanRoot, pkg.meta.workspaceRoot)) : undefined,
        files: sum.files,
        bytes: sum.bytes,
        languages: sum.languages,
        topDirs: sum.topDirs,
        biggestFiles: largestFiles(entries, HOTLIST_LIMIT),
        _files: entries
    };
    return base;
}

// ------------------------- Docs generation ---------------------------

function tableKV(obj: Record<string, number>, headerA = "Item", headerB = "Size", limit = 16) {
    const rows = Object.entries(obj).sort((a,b) => b[1] - a[1]).slice(0, limit);
    const lines = [`| ${headerA} | ${headerB} |`, "|---|---:|"];
    for (const [k,v] of rows) lines.push(`| \`${k || "(none)"}\` | ${fmtBytes(v)} |`);
    return lines.join("\n");
}
function pkgRow(p: PackageSummary) {
    const name = p.name ?? p.module ?? "(unnamed)";
    const langs = Object.entries(p.languages)
        .sort((a,b)=>b[1]-a[1]).slice(0,3).map(([ext])=>ext.replace(/^\./,"")).join(", ");
    return `| \`${p.kind}\` | \`${p.root}\` | ${name} | ${p.files} | ${fmtBytes(p.bytes)} | ${langs} |`;
}

async function writeOverviewAndModules(idx: Index, churn?: ChurnTotals) {
    await fs.mkdir(DOCS_DIR, { recursive: true });

    const langTotals: Record<string, number> = {};
    for (const p of idx.packages) for (const [k,v] of Object.entries(p.languages)) langTotals[k] = (langTotals[k] || 0) + v;

    const overview = `# Repository Overview

- Scanned: ${idx.scannedAt}
- Target: \`${idx.target}\`
- Packages: ${idx.packages.length}
- Files: ${idx.stats.files.toLocaleString()}
- Bytes: ${fmtBytes(idx.stats.bytes)}
- Digest: \`${idx.digest}\`

## Languages (total, by size)
${tableKV(langTotals)}

> Generated by XRAY v1. Flags: --cache ${CACHE_ON ? "on" : "off"}${DEEP.length ? `, --deep ${DEEP.join(",")}`:""}${EXPORTS ? ", --exports" : ""}${CHURN_WINDOW ? `, --churn ${CHURN_WINDOW}` : ""}.
`;
    await fs.writeFile(path.join(DOCS_DIR, "overview.md"), overview, "utf8");

    const header = `# Modules & Packages

First-pass map of discovered packages. Refine iteratively.

| Kind | Root | Name/Module | Files | Bytes | Top Langs | ${churn ? "Churn (since "+churn.window+")" : ""} |
|---|---|---|---:|---:|---|${churn ? "---:" : ""}|
${idx.packages.map(p => {
        const name = p.name ?? p.module ?? "(unnamed)";
        const langs = Object.entries(p.languages).sort((a,b)=>b[1]-a[1]).slice(0,3).map(([ext])=>ext.replace(/^\./,"")).join(", ");
        const churnCell = churn ? fmtChurn(churn.byPkg[p.root]?.added||0, churn.byPkg[p.root]?.deleted||0, churn.byPkg[p.root]?.commits||0) : "";
        return `| \`${p.kind}\` | \`${p.root}\` | ${name} | ${p.files} | ${fmtBytes(p.bytes)} | ${langs} | ${churnCell} |`;
    }).join("\n")}
`;
    await fs.writeFile(path.join(DOCS_DIR, "modules.md"), header, "utf8");
}

async function writeHotlist(allFilesByPkg: { pkg: PackageSummary; files: FileEntry[] }[]) {
    const topFilesGlobal = allFilesByPkg.flatMap(({pkg, files}) =>
        files.map(f => ({ path: `${pkg.root}/${f.path}`, size: f.size, ext: f.ext }))
    );
    topFilesGlobal.sort((a,b)=>b.size-a.size);
    const biggest = topFilesGlobal.slice(0, HOTLIST_LIMIT);

    const dirAgg: Record<string, number> = {};
    for (const g of topFilesGlobal) {
        const dir = toPosix(path.dirname(g.path));
        dirAgg[dir] = (dirAgg[dir] || 0) + g.size;
    }
    const biggestDirs = Object.entries(dirAgg).sort((a,b)=>b[1]-a[1]).slice(0, HOTLIST_LIMIT);

    const binaries = topFilesGlobal.filter(f => f.size > 256*1024 && (likelyBinaryByExt(f.ext) || /\.(bin|dat|pack|wasm)$/.test(f.path))).slice(0, HOTLIST_LIMIT);

    const genRxes = DEFAULT_GENERATED_HINTS.map(globToRegExp);
    const generated = topFilesGlobal.filter(f => genRxes.some(rx => rx.test(f.path))).slice(0, HOTLIST_LIMIT);

    const lines = [
        `# Hotlist

Largest files, heaviest dirs, likely binaries, and generated-code pockets (first ${HOTLIST_LIMIT} each).

## Largest Files
| File | Size |
|---|---:|`,
        ...biggest.map(f => `| \`${f.path}\` | ${fmtBytes(f.size)} |`),

        `\n## Heaviest Directories
| Dir | Size |
|---|---:|`,
        ...biggestDirs.map(([d, sz]) => `| \`${d}\` | ${fmtBytes(sz)} |`),

        `\n## Likely Binaries (by extension/size)
| File | Size |
|---|---:|`,
        ...binaries.map(f => `| \`${f.path}\` | ${fmtBytes(f.size)} |`),

        `\n## Generated-Code Pockets (heuristic)
| File | Size |
|---|---:|`,
        ...generated.map(f => `| \`${f.path}\` | ${fmtBytes(f.size)} |`),
    ].join("\n");

    await fs.writeFile(path.join(DOCS_DIR, "hotlist.md"), lines, "utf8");
}

function detectEncoreSignals(pkgs: { pkg: PackageSummary; files: FileEntry[] }[]) {
    const buckets = {
        supervisor: [] as FileEntry[],
        daemon: [] as FileEntry[],
        appruntime: [] as FileEntry[],
        runtimes_js: [] as FileEntry[],
        services: [] as FileEntry[],
        pkg: [] as FileEntry[],
    };
    for (const { pkg, files } of pkgs) {
        for (const f of files) {
            const full = `${pkg.root}/${f.path}`;
            if (/\/supervisor\//i.test(full)) buckets.supervisor.push(f);
            else if (/\/daemon\//i.test(full)) buckets.daemon.push(f);
            else if (/\/appruntime\//i.test(full)) buckets.appruntime.push(f);
            else if (/\/runtimes\/js\//i.test(full)) buckets.runtimes_js.push(f);
            else if (/\/services\//i.test(full)) buckets.services.push(f);
            else if (/\/pkg\//i.test(full)) buckets.pkg.push(f);
        }
    }
    const totals = Object.fromEntries(Object.entries(buckets).map(([k, arr]) => [k, arr.reduce((a,c)=>a+c.size,0)]));
    const any = Object.values(totals).some(v => v > 0);
    return { any, buckets, totals };
}

async function writeEncoreMap(group: ReturnType<typeof detectEncoreSignals>) {
    if (!group.any) return;
    const rows = Object.entries(group.totals).sort((a,b)=>b[1]-a[1]);
    const md = `# Encore Map (Heuristic)

Size by component bucket (heuristic path matching).

| Component | Bytes |
|---|---:|
${rows.map(([k,v])=>`| ${k} | ${fmtBytes(v)} |`).join("\n")}

> Buckets: \`supervisor\`, \`daemon\`, \`appruntime\`, \`runtimes/js\`, \`services\`, \`pkg\`.
`;
    await fs.writeFile(path.join(DOCS_DIR, "encore-map.md"), md, "utf8");
}

// -------------------------- Git churn -------------------------------

type ChurnEntry = { file: string; added: number; deleted: number };
type ChurnTotals = {
    byPath: Record<string, { added: number; deleted: number; commits: number }>;
    byPkg: Record<string, { added: number; deleted: number; commits: number }>;
    window: string;
};

function parseWindowToSinceISO(window: string): string | null {
    if (!window) return null;
    if (/^\d+\s*[dDwWmM]$/.test(window)) {
        const now = new Date();
        const n = parseInt(window, 10);
        const unit = window.trim().slice(-1).toLowerCase();
        const d = new Date(now);
        if (unit === "d") d.setDate(now.getDate() - n);
        if (unit === "w") d.setDate(now.getDate() - n * 7);
        if (unit === "m") d.setMonth(now.getMonth() - n);
        return d.toISOString().slice(0, 10);
    }
    if (/^\d{4}-\d{2}-\d{2}$/.test(window)) return window; // explicit since date
    return null;
}

async function computeChurn(scanRoot: string, pkgs: PackageSummary[], window: string): Promise<ChurnTotals | null> {
    const since = parseWindowToSinceISO(window);
    if (!since) return null;

    // run git log --numstat
    const { code, out } = await run("git", ["log", `--since=${since}`, "--numstat", "--pretty=format:@@@%H"], scanRoot);
    if (code !== 0) return null;

    const byPath: ChurnTotals["byPath"] = {};
    const lines = out.split("\n");
    let currentCommitTouched = new Set<string>();
    let commitsCountedPaths = new Set<string>();

    for (const ln of lines) {
        if (ln.startsWith("@@@")) {
            // new commit boundary: increment commit counters for files seen in that commit
            for (const p of currentCommitTouched) {
                const t = byPath[p] || (byPath[p] = { added: 0, deleted: 0, commits: 0 });
                t.commits += 1;
            }
            currentCommitTouched = new Set();
            commitsCountedPaths = new Set();
            continue;
        }
        // numstat line: "added<TAB>deleted<TAB>path"
        const m = ln.match(/^(\d+|-)\s+(\d+|-)\s+(.+)$/);
        if (!m) continue;
        const added = m[1] === "-" ? 0 : parseInt(m[1], 10);
        const deleted = m[2] === "-" ? 0 : parseInt(m[2], 10);
        const p = toPosix(path.relative(scanRoot, path.join(scanRoot, m[3])));

        const t = byPath[p] || (byPath[p] = { added: 0, deleted: 0, commits: 0 });
        t.added += added;
        t.deleted += deleted;

        if (!commitsCountedPaths.has(p)) {
            currentCommitTouched.add(p);
            commitsCountedPaths.add(p);
        }
    }
    // flush last commit
    for (const p of currentCommitTouched) {
        const t = byPath[p] || (byPath[p] = { added: 0, deleted: 0, commits: 0 });
        t.commits += 1;
    }

    // roll up to package roots
    const byPkg: ChurnTotals["byPkg"] = {};
    const byRoot = new Map(pkgs.map(p => [p.root + "/", p]));
    for (const [p, t] of Object.entries(byPath)) {
        // find the package whose root is a prefix of p (longest match)
        let owner: string | null = null;
        for (const r of byRoot.keys()) if (p.startsWith(r) || p === r.slice(0, -1)) owner = (!owner || r.length > owner.length ? r : owner);
        const key = owner ? (owner.endsWith("/") ? owner.slice(0, -1) : owner) : "(root)";
        const agg = byPkg[key] || (byPkg[key] = { added: 0, deleted: 0, commits: 0 });
        agg.added += t.added; agg.deleted += t.deleted; agg.commits += t.commits;
    }

    return { byPath, byPkg, window: since };
}

function fmtChurn(a: number, d: number, c: number) {
    const delta = a + d;
    return `${delta.toLocaleString()} lines / ${c} commits`;
}

async function writeChurnDoc(docsDir: string, churn: ChurnTotals, pkgs: PackageSummary[]) {
    const rows = Object.entries(churn.byPkg)
        .sort((a,b) => (b[1].added + b[1].deleted) - (a[1].added + a[1].deleted))
        .map(([pkgRoot, t]) => `| \`${pkgRoot}\` | ${t.added.toLocaleString()} | ${t.deleted.toLocaleString()} | ${t.commits} | ${fmtChurn(t.added, t.deleted, t.commits)} |`);

    const md = `# Churn Map (since ${churn.window})

| Package | Added | Deleted | Commits | Summary |
|---|---:|---:|---:|---|
${rows.join("\n") || "_No recent activity in window._"}
`;
    await fs.writeFile(path.join(docsDir, "churn.md"), md, "utf8");
}

// -------------------------- API Exports ------------------------------

type NodeExports = {
    package: string;
    root: string;
    exports: Record<string, string[]>; // file -> exported identifiers
};
type GoExports = {
    moduleRoot: string;
    packagePath: string;
    exports: Record<string, string[]>; // file -> exported identifiers
};
type RustExports = {
    crateRoot: string;
    crateName: string;
    exports: Record<string, string[]>; // file -> exported identifiers
};

async function buildNodeExports(scanRoot: string, pkgs: PackageSummary[]): Promise<NodeExports[]> {
    const targets = pkgs.filter(p => p.kind === "node");
    if (!targets.length) return [];
    // dynamic import to avoid hard dependency if flag off
    const { Project, ScriptTarget } = await import("ts-morph");
    const out: NodeExports[] = [];

    for (const p of targets) {
        const absRoot = path.join(scanRoot, p.root);
        const tsconfig = path.join(absRoot, "tsconfig.json");
        const project = new Project({
            tsConfigFilePath: (await exists(tsconfig)) ? tsconfig : undefined,
            compilerOptions: (await exists(tsconfig)) ? undefined : { target: ScriptTarget.ES2020, allowJs: true },
            skipAddingFilesFromTsConfig: !await exists(tsconfig)
        });
        if (!await exists(tsconfig)) {
            project.addSourceFilesAtPaths([path.join(absRoot, "**/*.ts"), path.join(absRoot, "**/*.tsx"), path.join(absRoot, "**/*.js")]);
        }

        const map: Record<string, string[]> = {};
        for (const sf of project.getSourceFiles()) {
            const decls = sf.getExportedDeclarations();
            const names = Array.from(decls.keys());
            if (names.length) {
                const rel = toPosix(path.relative(absRoot, sf.getFilePath()));
                map[rel] = names.slice(0, 100); // cap a bit
            }
        }
        out.push({ package: p.name ?? p.root, root: p.root, exports: map });
    }
    return out;
}

// crude regex-based Go exports (public identifiers start with uppercase)
async function buildGoExports(scanRoot: string, pkgs: PackageSummary[]): Promise<GoExports[]> {
    const targets = pkgs.filter(p => p.kind === "go");
    const out: GoExports[] = [];
    for (const m of targets) {
        const cwd = path.join(scanRoot, m.root);
        const res = await run("go", ["list", "-f", "{{.Dir}}:{{.ImportPath}}", "./..."], cwd);
        if (res.code !== 0) continue;
        const map: Record<string, string[]> = {};
        const lines = res.out.trim().split("\n");
        for (const ln of lines) {
            const [dir, pkgPath] = ln.split(":");
            if (!dir || !pkgPath) continue;
            const files = await fs.readdir(dir).catch(() => []);
            for (const f of files) {
                if (!f.endsWith(".go") || f.endsWith("_test.go")) continue;
                const abs = path.join(dir, f);
                const txt = await fs.readFile(abs, "utf8").catch(() => "");
                const ids = [
                    ...txt.matchAll(/\bfunc\s+([A-Z]\w*)\s*\(/g),
                    ...txt.matchAll(/\btype\s+([A-Z]\w*)\b/g),
                    ...txt.matchAll(/\bvar\s+([A-Z]\w*)\b/g),
                    ...txt.matchAll(/\bconst\s+([A-Z]\w*)\b/g),
                ].map(m => m[1]);
                if (ids.length) {
                    const rel = toPosix(path.relative(path.join(scanRoot, m.root), abs));
                    map[rel] = Array.from(new Set(ids)).slice(0, 100);
                }
            }
        }
        out.push({ moduleRoot: m.root, packagePath: m.module || m.root, exports: map });
    }
    return out;
}

// crude regex-based Rust exports
async function buildRustExports(scanRoot: string, pkgs: PackageSummary[]): Promise<RustExports[]> {
    const targets = pkgs.filter(p => p.kind === "rust");
    const out: RustExports[] = [];
    for (const r of targets) {
        const absRoot = path.join(scanRoot, r.root);
        const src = path.join(absRoot, "src");
        if (!await exists(src)) continue;
        const map: Record<string, string[]> = {};
        // walk src
        const stack = [src];
        while (stack.length) {
            const d = stack.pop()!;
            const ent = await fs.readdir(d, { withFileTypes: true }).catch(() => []);
            for (const e of ent) {
                const pth = path.join(d, e.name);
                if (e.isDirectory()) { stack.push(pth); continue; }
                if (!e.isFile() || !e.name.endsWith(".rs")) continue;
                const txt = await fs.readFile(pth, "utf8").catch(() => "");
                const ids = [
                    ...txt.matchAll(/\bpub\s+fn\s+([a-zA-Z_]\w*)/g),
                    ...txt.matchAll(/\bpub\s+struct\s+([A-Z]\w*)/g),
                    ...txt.matchAll(/\bpub\s+enum\s+([A-Z]\w*)/g),
                    ...txt.matchAll(/\bpub\s+trait\s+([A-Z]\w*)/g),
                    ...txt.matchAll(/\bpub\s+mod\s+([a-zA-Z_]\w*)/g),
                ].map(m => m[1]);
                if (ids.length) {
                    const rel = toPosix(path.relative(absRoot, pth));
                    map[rel] = Array.from(new Set(ids)).slice(0, 100);
                }
            }
        }
        out.push({ crateRoot: r.root, crateName: r.name ?? r.root, exports: map });
    }
    return out;
}

async function writeExportsDocs(docsDir: string, nodeE: NodeExports[], goE: GoExports[], rustE: RustExports[]) {
    // Ensure destination exists (covers calls from "scan" in the `all` flow)
    await fs.mkdir(docsDir, { recursive: true });

    const sect = (title: string, rows: string[]) => `# ${title}\n\n${rows.join("\n\n") || "_No data._"}\n`;

    // Node
    const nrows = nodeE.map(n => {
        const items = Object.entries(n.exports).sort().slice(0, 200).map(([file, ids]) =>
            `- \`${n.root}/${file}\` — ${ids.slice(0, 20).map(s=>`\`${s}\``).join(", ")}`
        );
        return `## ${n.package} (\`${n.root}\`)\n${items.join("\n") || "_No exports found._"}`;
    });
    await fs.writeFile(path.join(docsDir, "api-node.md"), sect("Node/TS API Surface", nrows), "utf8");

    // Go
    const grows = goE.map(g => {
        const items = Object.entries(g.exports).sort().slice(0, 200).map(([file, ids]) =>
            `- \`${g.moduleRoot}/${file}\` — ${ids.slice(0, 20).map(s=>`\`${s}\``).join(", ")}`
        );
        return `## ${g.packagePath} (\`${g.moduleRoot}\`)\n${items.join("\n") || "_No public identifiers found._"}`;
    });
    await fs.writeFile(path.join(docsDir, "api-go.md"), sect("Go API Surface", grows), "utf8");

    // Rust
    const rrows = rustE.map(r => {
        const items = Object.entries(r.exports).sort().slice(0, 200).map(([file, ids]) =>
            `- \`${r.crateRoot}/${file}\` — ${ids.slice(0, 20).map(s=>`\`${s}\``).join(", ")}`
        );
        return `## ${r.crateName} (\`${r.crateRoot}\`)\n${items.join("\n") || "_No pub items found._"}`;
    });
    await fs.writeFile(path.join(docsDir, "api-rust.md"), sect("Rust API Surface", rrows), "utf8");
}

// ------------------------- Main commands -----------------------------

async function cmdScan(targetDir: string) {
    // If caller didn't pass a targetDir, default to repo root.
    const effective = (targetDir && targetDir.trim()) ? targetDir.trim() : ".";

    // Anchor all relative resolution to the detected repo root.
    // Special-case '.' so invoking from tools/context-compiler still scans the Stagecraft repo.
    const scanRoot = (effective === "." || effective === "")
        ? REPO_ROOT
        : (path.isAbsolute(effective) ? effective : path.resolve(REPO_ROOT, effective));

    if (!scanRoot) throw new Error("scan requires a resolved repo root");

    // Build ignore tester
    const specs = await collectIgnoreSpecs(scanRoot);
    const isIgnored = buildIgnoreTester(specs);

    // Cache
    const cache = await loadCache();

    // Discover + index
    const packages = await discoverPackages(scanRoot, isIgnored);

    const summaries: PackageSummary[] = [];
    const pkgFiles: { pkg: PackageSummary; files: FileEntry[] }[] = [];
    let totalFiles = 0;
    let totalBytes = 0;
    const langTotals: Record<string, number> = {};

    await fs.mkdir(JSON_DIR, { recursive: true });
    const pkgDir = path.join(JSON_DIR, "packages");
    await fs.mkdir(pkgDir, { recursive: true });

    for (const p of packages) {
        const sum = await indexPackage(scanRoot, p, isIgnored, cache);
        const { _files, ...pub } = sum;
        summaries.push(pub);
        pkgFiles.push({ pkg: pub, files: _files });

        totalFiles += pub.files;
        totalBytes += pub.bytes;
        for (const [k,v] of Object.entries(pub.languages)) langTotals[k] = (langTotals[k] || 0) + v;

        const out = path.join(pkgDir, `${pub.kind}-${pub.root.replace(/[\/]/g, "_")}.json`);
        await fs.writeFile(out, JSON.stringify(pub, null, 2), "utf8");
    }

    const index: Index = {
        scannedAt: new Date().toISOString(),
        target: scanRoot,
        packages: summaries.sort((a,b)=>a.root.localeCompare(b.root)),
        stats: { files: totalFiles, bytes: totalBytes, languages: langTotals },
        digest: ""
    };

    const raw = JSON.stringify(index);
    index.digest = createHash("sha256").update(raw).digest("hex").slice(0, 16);

    await fs.writeFile(path.join(JSON_DIR, "index.json"), JSON.stringify(index, null, 2), "utf8");

    // Save cache after a successful scan
    await saveCache(cache);

    console.log(`Indexed ${summaries.length} packages, ${totalFiles} files, ${fmtBytes(totalBytes)}`);
    console.log(`→ ${path.join(JSON_DIR, "index.json")}`);

    // Deep analyzers (existing)
    const graphs: { node?: NodeGraph; go?: GoGraph[]; rust?: RustGraph[] } = {};
    if (DEEP.includes("node")) graphs.node = await buildNodeGraph(scanRoot, index.packages);
    if (DEEP.includes("go"))   graphs.go   = await buildGoGraph(scanRoot, index.packages);
    if (DEEP.includes("rust")) graphs.rust = await buildRustGraph(scanRoot, index.packages);
    if (graphs.node || graphs.go || graphs.rust) await writeGraphs(JSON_DIR, graphs);

    // Exports (new)
    let nodeExports: NodeExports[] = [];
    let goExports: GoExports[] = [];
    let rustExports: RustExports[] = [];
    if (EXPORTS) {
        await fs.mkdir(DOCS_DIR, { recursive: true });
        nodeExports = await buildNodeExports(scanRoot, index.packages);
        goExports   = await buildGoExports(scanRoot, index.packages);
        rustExports = await buildRustExports(scanRoot, index.packages);
        await writeExportsDocs(DOCS_DIR, nodeExports, goExports, rustExports);
    }

    // Churn (new)
    let churnTotals: ChurnTotals | null = null;
    if (CHURN_WINDOW) churnTotals = await computeChurn(scanRoot, index.packages, CHURN_WINDOW);

    // Return details for docs phase
    return { index, pkgFiles };
}

async function cmdDocs(index?: Index, pkgFiles?: { pkg: PackageSummary; files: FileEntry[] }[]) {
    if (!index) {
        const idx = await safeJSON<Index>(path.join(JSON_DIR, "index.json"));
        if (!idx) throw new Error(`Cannot read ${path.join(JSON_DIR, "index.json")}`);
        index = idx;
    }

    // Attempt to recompute churn if flag is provided; otherwise, show without it.
    let churnTotals: ChurnTotals | null = null;
    if (CHURN_WINDOW) {
        churnTotals = await computeChurn(path.resolve(index.target), index.packages, CHURN_WINDOW);
    }

    await writeOverviewAndModules(index!, churnTotals || undefined);

    if (pkgFiles && pkgFiles.length) {
        await writeHotlist(pkgFiles);
        const encore = detectEncoreSignals(pkgFiles);
        if (encore.any) await writeEncoreMap(encore);
    }

    // Load any existing graphs to render deps.md (so docs can be regenerated standalone)
    let graphs: { node?: NodeGraph; go?: GoGraph[]; rust?: RustGraph[] } = {};
    const gdir = path.join(JSON_DIR, "graphs");
    const [n, g, r] = await Promise.all([
        safeJSON<NodeGraph>(path.join(gdir, "node.json")),
        safeJSON<GoGraph[]>(path.join(gdir, "go.json")),
        safeJSON<RustGraph[]>(path.join(gdir, "rust.json")),
    ]);
    graphs = { node: n ?? undefined, go: g ?? undefined, rust: r ?? undefined };
    if (graphs.node || graphs.go || graphs.rust) {
        await writeDepsDoc(JSON_DIR, DOCS_DIR, graphs);
    }

    // Try to (re)write exports docs if user passed --exports (even on docs-only run)
    if (EXPORTS) {
        const nodeExports = await buildNodeExports(path.resolve(index.target), index.packages);
        const goExports   = await buildGoExports(path.resolve(index.target), index.packages);
        const rustExports = await buildRustExports(path.resolve(index.target), index.packages);
        await writeExportsDocs(DOCS_DIR, nodeExports, goExports, rustExports);
    }

    console.log(`Docs written to ${DOCS_DIR}/overview.md, ${DOCS_DIR}/modules.md${CHURN_WINDOW ? ", churn.md" : ""}`);
}

(async function main() {
    try {
        // Resolve repo root and set up paths
        const scanTarget = targetArg || ".";
        REPO_ROOT = await findRepoRoot(scanTarget);
        REPO_SLUG = deriveRepoSlug(REPO_ROOT);

        // Default to namespaced paths under .ai-context/xray/<slug>/
        if (jsonFlag) {
            JSON_DIR = path.resolve(process.cwd(), jsonFlag);
        } else {
            JSON_DIR = path.join(REPO_ROOT, ".ai-context", "xray", REPO_SLUG, "data");
        }

        if (outFlag) {
            DOCS_DIR = path.resolve(process.cwd(), outFlag);
        } else {
            DOCS_DIR = path.join(REPO_ROOT, ".ai-context", "xray", REPO_SLUG, "docs");
        }

        // Cache is always at repo root
        CACHE_DIR = path.join(REPO_ROOT, ".xraycache");

        if (cmd === "scan") {
            console.log(`[xray] repo: ${REPO_SLUG}`);
            console.log(`[xray] json: ${JSON_DIR}`);
            console.log(`[xray] docs: ${DOCS_DIR}`);
            await cmdScan(scanTarget || ".");
        } else if (cmd === "docs") {
            console.log(`[xray] repo: ${REPO_SLUG}`);
            console.log(`[xray] json: ${JSON_DIR}`);
            console.log(`[xray] docs: ${DOCS_DIR}`);
            await cmdDocs();
        } else if (cmd === "all") {
            if (!targetArg) throw new Error("all requires <targetDir>");
            console.log(`[xray] repo: ${REPO_SLUG}`);
            console.log(`[xray] json: ${JSON_DIR}`);
            console.log(`[xray] docs: ${DOCS_DIR}`);
            const { index, pkgFiles } = await cmdScan(scanTarget || ".");
            await cmdDocs(index, pkgFiles);
        } else {
            throw new Error(`unknown command: ${cmd}`);
        }
    } catch (e: any) {
        console.error("XRAY error:", e.message || e);
        process.exit(1);
    }
})();

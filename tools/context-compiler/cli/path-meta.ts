// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

import type { Meta } from "./types";

export function inferMetaFromPath(repo: string, relPath: string): Partial<Meta> {
    if (relPath.startsWith("docs/")) {
        return { source_type: "doc", artifact_type: classifyDoc(relPath) };
    }
    if (relPath.includes("/graphs/")) {
        return { source_type: "graph", language: inferLang(relPath), artifact_type: "deps-graph" };
    }
    if (relPath.includes("/packages/")) {
        return { source_type: "package", language: inferLang(relPath), artifact_type: "manifest" };
    }
    return { source_type: "doc", artifact_type: "overview" };
}

function inferLang(p: string): Meta["language"] {
    if (p.includes("/go")) return "go";
    if (p.includes("/node")) return "node";
    if (p.includes("/rust")) return "rust";
    return "mixed";
}

function classifyDoc(p: string): Meta["artifact_type"] {
    if (p.endsWith("overview.md")) return "overview";
    if (p.endsWith("deps.md")) return "dependency-map";
    if (p.includes("api-")) return "api";
    return "overview";
}

export function mkBaseMeta(repo: string, path: string): Meta {
    return {
        repo,
        source: "xray",
        source_type: "doc",
        path,
    };
}

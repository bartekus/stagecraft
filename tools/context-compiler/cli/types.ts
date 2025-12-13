// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

export type SourceType = "doc" | "graph" | "package";

export type RetryOpts = { retries?: number; baseMs?: number; maxMs?: number };

// .ai-context schema definition
export const SCHEMA_VERSION = "1.0.0";

export interface ContextMeta {
    schemaVersion: string;
    repoSlug: string;
    targetRoot: string;
    options: {
        include: string[];
        ext: string[];
    };
    counts: {
        files: number;
        chunks: number;
    };
    digest: string;
    generatedAt: string;
}

export interface FileManifestEntry {
    path: string;
    sha: string;
    chunks: number;
}

export interface ChunkEntry {
    repoSlug: string;
    path: string;
    kind?: string;
    startLine: number;
    endLine: number;
    sha: string;
    text: string;
    meta?: Record<string, any>;
}

export interface Meta {
    repo: string;
    source: "xray";
    source_type: SourceType;
    language?: "go" | "node" | "rust" | "mixed";
    artifact_type?: "deps-graph" | "manifest" | "dependency-map" | "api" | "overview";
    path: string;
    section?: string;
    scanned_at?: string; // ISO (optional, only in meta.json, not in chunks)
    sha?: string;
}

export interface Chunk {
    id: string;
    text: string;
    meta: Meta;
}

export interface EmbeddedRow {
    id: string;
    vector: number[];
    text: string;
    meta: Meta;
}

export type ChunkRec = FileIngestPayload["chunks"][number];

export interface BuildOpts {
    repoSlug: string;     // e.g., "encore"
    absPath: string;      // absolute file path
    relRoot: string;      // absolute root dir to compute repo-relative path
}

export interface Embedder {
    name: "openai" | "ollama";
    embed: (input: string[]) => Promise<number[][]>;
}

export interface FileIngestPayload {
    repoSlug: string;
    path: string;
    sha: string;
    lang: string;
    sizeBytes?: number;
    chunks: Array<{
        lang: string;
        symbol?: string;
        kind?: string;
        startLine: number;
        endLine: number;
        sha: string;
        meta?: Record<string, any>;
        text: string;
    }>;
}

export type UpsertResp = {
    fileId: number;
    requestedChunks: number;
    insertedChunkIds: number[];
    skipped: { index: number; reason: string }[];
};

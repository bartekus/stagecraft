// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

import { createHash } from "node:crypto";
import { promises as fs } from "node:fs";
import { join } from "node:path";
import type { RetryOpts } from "./types";

export const sha1 = (s: string | Buffer) =>
    createHash("sha1").update(s).digest("hex");

export async function* walk(dir: string): AsyncGenerator<string> {
    const ents = await fs.readdir(dir, { withFileTypes: true }).catch(() => []);
    for (const e of ents) {
        const p = join(dir, e.name);
        if (e.isDirectory()) yield* walk(p);
        else yield p;
    }
}

export async function readText(path: string): Promise<string> {
    return fs.readFile(path, "utf8");
}

export async function fileSha(path: string): Promise<string> {
    const buf = await fs.readFile(path);
    return sha1(buf);
}

export async function fileSize(path: string): Promise<number> {
    const s = await fs.stat(path);
    return s.size;
}

export function countLines(s: string): number {
    if (!s) return 0;
    return s.split(/\r?\n/).length;
}

export function sleep(ms: number) {
    return new Promise(res => setTimeout(res, ms));
}

export async function retry<T>(fn: () => Promise<T>, { retries = 5, baseMs = 250, maxMs = 5000 }: RetryOpts = {}): Promise<T> {
    let attempt = 0;
    // full jitter backoff
    while (true) {
        try {
            return await fn();
        } catch (err: any) {
            attempt++;
            const status = err?.status ?? err?.response?.status;
            // Retry only on transient codes/timeouts/network resets
            const retryable = status === 429 || (status >= 500 && status <= 599) || err?.code === 'ECONNRESET' || err?.name === 'FetchError';
            if (!retryable || attempt > retries) throw err;
            const delay = Math.min(maxMs, baseMs * Math.pow(2, attempt - 1)) * Math.random();
            console.warn(`Embedding retry ${attempt}/${retries} after error ${status ?? err?.message} — waiting ${Math.round(delay)}ms`);
            await sleep(delay);
        }
    }
}

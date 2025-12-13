// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

/**
 * File size cap for LOC scanning: files larger than this are skipped.
 * Large binaries showing lines: 0 is expected and desirable. It avoids scanning huge files
 * and makes results stable across environments.
 */
export const LOC_BIG_FILE_CAP_BYTES = 2 * 1024 * 1024;

/**
 * Minimal file entry type for LOC summary computation.
 * Only size is needed to determine if a file should be skipped.
 */
export type FileEntry = { size: number };

/**
 * Summary of LOC scanning results.
 */
export type LocSummary = {
    scannedFiles: number;
    skippedFiles: number;
    skippedBytes: number;
};

/**
 * Computes LOC scanning summary from a list of FileEntry objects.
 * A file is considered "skipped" if its size exceeds LOC_BIG_FILE_CAP_BYTES.
 * Files at or below the cap are scanned; files strictly above are skipped.
 * 
 * Large binaries showing lines: 0 is expected and desirable. It avoids scanning huge files
 * and makes results stable across environments.
 * 
 * @param entries Array of FileEntry objects to analyze
 * @returns Summary with scannedFiles, skippedFiles, and skippedBytes
 */
export function computeLocSummary(entries: FileEntry[]): LocSummary {
    let scannedFiles = 0;
    let skippedFiles = 0;
    let skippedBytes = 0;

    for (const entry of entries) {
        if (entry.size > LOC_BIG_FILE_CAP_BYTES) {
            skippedFiles++;
            skippedBytes += entry.size;
        } else {
            scannedFiles++;
        }
    }

    return { scannedFiles, skippedFiles, skippedBytes };
}


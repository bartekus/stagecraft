// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

import { describe, it, expect } from "vitest";
import { computeLocSummary, LOC_BIG_FILE_CAP_BYTES, type FileEntry } from "./loc";

describe("computeLocSummary", () => {
    it("should correctly categorize files at cap, above cap, and below cap", () => {
        // Test case from requirements: entries with sizes at cap, cap+1, and 10
        const entries: FileEntry[] = [
            { size: LOC_BIG_FILE_CAP_BYTES },      // At cap: scanned
            { size: LOC_BIG_FILE_CAP_BYTES + 1 },  // Above cap: skipped
            { size: 10 }                            // Below cap: scanned
        ];

        const result = computeLocSummary(entries);

        // cap-sized counts as scanned, only strictly greater is skipped
        expect(result.scannedFiles).toBe(2);
        expect(result.skippedFiles).toBe(1);
        expect(result.skippedBytes).toBe(LOC_BIG_FILE_CAP_BYTES + 1);
    });

    it("should correctly categorize files below and above the cap", () => {
        const files: FileEntry[] = [
            { size: 1024 },                         // 1KB, below cap
            { size: LOC_BIG_FILE_CAP_BYTES + 1 }   // Above cap
        ];

        const result = computeLocSummary(files);

        expect(result.scannedFiles).toBe(1);
        expect(result.skippedFiles).toBe(1);
        expect(result.skippedBytes).toBe(LOC_BIG_FILE_CAP_BYTES + 1);
    });

    it("should handle files exactly at the cap boundary", () => {
        const files: FileEntry[] = [
            { size: LOC_BIG_FILE_CAP_BYTES },      // Exactly at cap
            { size: LOC_BIG_FILE_CAP_BYTES + 1 }  // Just over cap
        ];

        const result = computeLocSummary(files);

        expect(result.scannedFiles).toBe(1); // exact.ts is scanned (size <= cap)
        expect(result.skippedFiles).toBe(1); // just-over.ts is skipped (size > cap)
        expect(result.skippedBytes).toBe(LOC_BIG_FILE_CAP_BYTES + 1);
    });

    it("should handle empty file list", () => {
        const files: FileEntry[] = [];
        const result = computeLocSummary(files);

        expect(result.scannedFiles).toBe(0);
        expect(result.skippedFiles).toBe(0);
        expect(result.skippedBytes).toBe(0);
    });

    it("should handle all files scanned (none skipped)", () => {
        const files: FileEntry[] = [
            { size: 100 },
            { size: 500 },
            { size: 1000 }
        ];

        const result = computeLocSummary(files);

        expect(result.scannedFiles).toBe(3);
        expect(result.skippedFiles).toBe(0);
        expect(result.skippedBytes).toBe(0);
    });

    it("should handle all files skipped (none scanned)", () => {
        const files: FileEntry[] = [
            { size: LOC_BIG_FILE_CAP_BYTES + 1000 },
            { size: LOC_BIG_FILE_CAP_BYTES + 5000 }
        ];

        const result = computeLocSummary(files);

        expect(result.scannedFiles).toBe(0);
        expect(result.skippedFiles).toBe(2);
        expect(result.skippedBytes).toBe((LOC_BIG_FILE_CAP_BYTES + 1000) + (LOC_BIG_FILE_CAP_BYTES + 5000));
    });

    it("should correctly sum skipped bytes across multiple skipped files", () => {
        const files: FileEntry[] = [
            { size: 1000 },
            { size: LOC_BIG_FILE_CAP_BYTES + 10000 },
            { size: LOC_BIG_FILE_CAP_BYTES + 20000 },
            { size: 5000 }
        ];

        const result = computeLocSummary(files);

        expect(result.scannedFiles).toBe(2); // 1000 and 5000
        expect(result.skippedFiles).toBe(2); // two large files
        expect(result.skippedBytes).toBe((LOC_BIG_FILE_CAP_BYTES + 10000) + (LOC_BIG_FILE_CAP_BYTES + 20000));
    });
});


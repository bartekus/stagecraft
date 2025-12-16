# XRAY (Rust) Specification

## 1. Overview

**XRAY** (`ai.agent/rust/xray`) is the high-performance, deterministic file scanner and static analyzer for Stagecraft. It is a standalone Rust binary that serves as the "eyes" of the Cortex system.

### Core Tenets
- **Speed**: Written in Rust for maximum throughput.
- **Determinism**: Outputs are strictly deterministic. Running XRAY twice on the same repository state MUST produce byte-identical outputs.
- **Independence**: XRAY does not depend on Cortex or Node.js. It depends only on the filesystem.

> [!NOTE]
> **Implementation Phasing**:
> - Phase A: traversal + loc + deterministic index (Done)
> - Phase B: content hashes (SHA-256) (Done)
> - Phase C1: structural aggregation (modules, topDirs, extension-based languages)
> - Phase C2: semantic analysis (language detection, complexity)

## 2. CLI Contract

The binary name is `xray`. It supports the following commands:

### `xray scan`
Scans the current directory (or target) and updates the local cache.

- **Usage**: `xray scan [FLAGS] [PATH]`
- **Behavior**:
  - Recursively walks the directory tree.
  - Respects hardcoded ignore list (Phase A) and `.xrayignore` (planned).
  - Computes content hashes (SHA-256) for all files.
  - Updates `.xraycache/` with the new index.

### `xray docs`
(Optional) Generates internal documentation or AST dumps if supported.

### `xray all`
Runs `scan` followed by any secondary processing steps.

## 3. Output Schema: `.xraycache/`

XRAY owns the `.xraycache/` directory. No other tool should write to it.

```text
.xraycache/
  <repoSlug>/
    data/
      index.json    (The authoritative file index)
    docs/           (Optional generated artifacts)
```

### `index.json`
This is the single source of truth for the repository's file state.

**Constraint**: Must be Canonical JSON (keys sorted, no whitespace variation).

```json
{
  "schemaVersion": "1.0.0",
  "root": "stagecraft",
  "target": ".",
  "files": [
    {
      "path": "cmd/cortex/main.go",
      "size": 1647,
      "hash": "sha256:abcd1234...",
      "lang": "go",
      "loc": 62,
      "complexity": 5
    }
  ],
  "languages": {
    "Go": 1
  },
  "topDirs": {
    "cmd": 1
  },
  "moduleFiles": [],
  "stats": {
    "fileCount": 1,
    "totalSize": 1647
  },
  "digest": "abcd1234..."
}
```

#### Determinism Rules
1.  **Sorting**: The `files` array MUST be sorted alphabetically by `path`.
2.  **No Timestamps**: The output must NOT contain `created_at`, `modified_at`, or runtime durations.
3.  **Stable Paths**: All paths are relative to the repository root.

#### Digest Definition
XRAY calculates a global "Repo Digest" to detect changes.
- **Digest Algorithm**: `SHA-256( CanonicalJSON( index_with_empty_digest ) )`.
- **Purpose**: Cortex uses this digest to skip rebuilding context if the repo hasn't changed.

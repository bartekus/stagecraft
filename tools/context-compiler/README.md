# Context Compiler

Local-only deterministic repo context builder for Stagecraft. Builds a searchable context snapshot from your repository documentation and specifications.

## Installation

From the Stagecraft repo root:

```bash
npm --prefix tools/context-compiler install
```

## Usage

### Build Context

Build a deterministic context snapshot from your repository:

```bash
npm --prefix tools/context-compiler run context:build
```

This will:
- Scan `docs/`, `spec/`, `README.md`, and `Agent.md` by default
- Process `.md` and `.json` files
- Write output to `.ai-context/` at the repo root

**Output structure:**
```
.ai-context/
  meta.json          # Build metadata and options
  digest.txt         # Deterministic SHA256 digest
  files/
    manifest.json    # List of processed files
    chunks.ndjson    # All chunks in NDJSON format
```

**Customization:**

```bash
# Custom target directory
npm --prefix tools/context-compiler run context:build -- --target ./my-repo

# Custom output directory
npm --prefix tools/context-compiler run context:build -- --out ./custom-output

# Custom include paths (comma-separated)
npm --prefix tools/context-compiler run context:build -- --include "docs,spec,custom-docs"

# Custom file extensions (comma-separated)
npm --prefix tools/context-compiler run context:build -- --ext ".md,.txt,.json"
```

### XRAY Scanning (Optional)

XRAY is a multi-language reverse-engineering scanner that can analyze your codebase:

```bash
# Full scan with all features
npm --prefix tools/context-compiler run xray:all

# Quick scan
npm --prefix tools/context-compiler run xray:scan

# Generate documentation
npm --prefix tools/context-compiler run xray:docs
```

XRAY outputs are cached in `.xraycache/` at the repo root.

## Output Format

### meta.json

Contains build metadata:
- `schemaVersion`: Version of the output schema
- `repoSlug`: Sanitized repository name
- `targetRoot`: Target directory (usually ".")
- `options`: Build options used (include paths, extensions)
- `counts`: File and chunk counts
- `digest`: SHA256 digest of the build
- `generatedAt`: ISO timestamp (not included in digest)

### manifest.json

Array of processed files:
```json
[
  {
    "path": "docs/README.md",
    "sha": "abc123...",
    "chunks": 5
  }
]
```

### chunks.ndjson

Newline-delimited JSON, one chunk per line:
```json
{"repoSlug":"stagecraft","path":"docs/README.md","kind":"section","startLine":1,"endLine":10,"sha":"def456...","text":"...","meta":{}}
```

Each chunk includes:
- `repoSlug`: Repository identifier
- `path`: Repo-relative file path
- `kind`: Chunk type (section, package, graph-node, etc.)
- `startLine` / `endLine`: Line range in source file
- `sha`: SHA1 of chunk content
- `text`: Chunk content
- `meta`: Additional metadata

### digest.txt

SHA256 digest computed from:
- Compiler version
- Normalized options
- Sorted list of (path, sha) pairs for all files

The digest is deterministic and does not include timestamps.

## Deterministic Builds

Builds are deterministic:
- Files are processed in lexicographic order
- Chunks are sorted by path, then startLine, then sha
- Digest excludes timestamps
- Paths are normalized to POSIX format

Running the same build with the same inputs produces identical outputs.

## File Processing

### Markdown Files

- Split by H2/H3 headings
- Each section becomes a chunk
- Headings are included in chunk text for better retrieval

### JSON Files

- **Graph JSON** (`/graphs/`): Per-node chunks + overview
- **Package JSON** (`/packages/`): Single concise card with metadata
- **Other JSON**: Single blob chunk

### Other Files

- Treated as plain text, single chunk per file

## Ignore Rules

XRAY uses ignore rules from `tools/context-compiler/xray/ignore.rules` (if present). Default ignores include:
- `.git`, `node_modules`, `dist`, `build`, `out`, `target`, `vendor`
- `.cache`, `.tmp`, `coverage`

## Notes

- The context compiler is **local-only** - no backend ingestion or embedding
- Old embedding/uploader code is kept but unused
- XRAY cache is stored in `.xraycache/` at repo root
- Both `.ai-context/` and `.xraycache/` are gitignored

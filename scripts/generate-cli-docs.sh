#!/bin/bash
# generate-cli-docs.sh - Generate CLI reference documentation from Cobra
#
# This script builds the stagecraft binary and uses it to generate
# markdown documentation for all commands.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

OUTPUT_FILE="docs/reference/cli.md"
TEMP_BINARY="$PROJECT_ROOT/.stagecraft-temp"

# Ensure output directory exists
mkdir -p "$(dirname "$OUTPUT_FILE")"

# Build the binary
echo "Building stagecraft binary..."
if ! go build -o "$TEMP_BINARY" ./cmd/stagecraft; then
    echo "Error: Failed to build stagecraft binary" >&2
    exit 1
fi

# Generate markdown docs with header
cat > "$OUTPUT_FILE" << 'EOF'
# Stagecraft CLI Reference

This document is auto-generated from the Cobra CLI. For the most up-to-date
information, run `stagecraft --help` or `stagecraft <command> --help`.

> **Note**: This is a generated file. Do not edit manually.
> To regenerate, run: `./scripts/generate-cli-docs.sh`

---

EOF

# Append help output
if [ -f "$TEMP_BINARY" ]; then
    "$TEMP_BINARY" --help >> "$OUTPUT_FILE" 2>&1 || true
else
    echo "Error: Binary not found after build" >> "$OUTPUT_FILE"
fi

# If help command failed, add a note
if [ ! -s "$OUTPUT_FILE" ] || ! grep -q "Stagecraft" "$OUTPUT_FILE"; then
    cat >> "$OUTPUT_FILE" << 'EOF'
# CLI Help Output

Run `stagecraft --help` to see the latest command documentation.

To generate this file, ensure the binary builds successfully:
  go build ./cmd/stagecraft
  ./scripts/generate-cli-docs.sh

EOF
fi

# Clean up
rm -f "$TEMP_BINARY"

echo "✓ CLI documentation generated at $OUTPUT_FILE"


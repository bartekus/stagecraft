#!/bin/bash
# SPDX-License-Identifier: AGPL-3.0-or-later
#
# Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.
#
# Copyright (C) 2025  Bartek Kus
#
# This program is free software licensed under the terms of the GNU AGPL v3 or later.
#
# See https://www.gnu.org/licenses/ for license details.
#
# check-orphan-specs.sh - Check for orphan spec files
#
# This script finds spec files that don't have a corresponding entry in
# spec/features.yaml.
#
# Exit code: 0 if no orphans found, 1 if orphans are detected

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

FEATURES_YAML="spec/features.yaml"
ERRORS=0

if [ ! -f "$FEATURES_YAML" ]; then
    echo "Error: $FEATURES_YAML not found"
    exit 1
fi

# Extract all spec paths from features.yaml
# This regex matches "spec: "path"" lines and extracts the path
SPEC_PATHS=$(grep -E "^\s+spec:" "$FEATURES_YAML" | sed 's/.*spec: *"\([^"]*\)".*/\1/' | sed "s/.*spec: *'\([^']*\)'.*/\1/")

echo "Checking for orphan spec files..."
echo ""

# Find all .md files in spec/ directory (excluding features.yaml and README files)
while IFS= read -r -d '' SPEC_FILE; do
    # Get relative path from spec/ directory (e.g., spec/commands/deploy.md -> commands/deploy.md)
    REL_PATH="${SPEC_FILE#spec/}"
    
    # Skip README files and overview.md (these are special)
    if [[ "$REL_PATH" == "README.md" ]] || [[ "$REL_PATH" == "overview.md" ]]; then
        continue
    fi
    
    # Check if this spec path is referenced in features.yaml
    if ! echo "$SPEC_PATHS" | grep -q "^${REL_PATH}$"; then
        echo "❌ Orphan spec file: ${SPEC_FILE}"
        echo "   Not referenced in spec/features.yaml"
        echo ""
        ERRORS=$((ERRORS + 1))
    fi
done < <(find spec -type f -name "*.md" -not -name "features.yaml" -print0)

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $ERRORS -eq 0 ]; then
    echo "✓ No orphan spec files found"
    exit 0
else
    echo "❌ Found ${ERRORS} orphan spec file(s)"
    echo ""
    echo "Orphan spec files are specification documents that don't have a corresponding"
    echo "entry in spec/features.yaml. These should either be:"
    echo "  - Removed if the feature was cancelled"
    echo "  - Have an entry added to spec/features.yaml if the feature is active"
    echo "  - Moved to docs/archive/ if the spec is historical"
    exit 1
fi


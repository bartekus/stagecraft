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
# check-orphan-docs.sh - Check for orphan Analysis/Outline files
#
# This script finds analysis and outline files that don't have a matching
# feature entry in spec/features.yaml.
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

# Extract all Feature IDs from features.yaml
FEATURE_IDS=$(grep -E "^\s+- id:" "$FEATURES_YAML" | sed 's/.*id: //' | tr -d '"' | tr -d "'")

if [ -z "$FEATURE_IDS" ]; then
    echo "Warning: No Feature IDs found in $FEATURES_YAML"
    exit 0
fi

echo "Checking for orphan Analysis/Outline files..."
echo ""

# Check analysis files
ANALYSIS_DIR="docs/engine/analysis"
if [ -d "$ANALYSIS_DIR" ]; then
    for ANALYSIS_FILE in "$ANALYSIS_DIR"/*.md; do
        # Skip if no files match the pattern
        [ -f "$ANALYSIS_FILE" ] || continue

        # Skip TEMPLATE.md
        if [ "$(basename "$ANALYSIS_FILE")" = "TEMPLATE.md" ]; then
            continue
        fi

        # Extract FEATURE_ID from filename
        # Handle patterns: CLI_PLAN.md, CLI_PLAN_ANALYSIS.md, GOV_CORE_IMPLEMENTATION_ANALYSIS.md
        FILENAME=$(basename "$ANALYSIS_FILE" .md)
        # Remove _IMPLEMENTATION_ANALYSIS suffix first, then _ANALYSIS if present
        FEATURE_ID="${FILENAME%_IMPLEMENTATION_ANALYSIS}"
        FEATURE_ID="${FEATURE_ID%_ANALYSIS}"

        # Check if this FEATURE_ID exists in features.yaml
        if ! echo "$FEATURE_IDS" | grep -q "^${FEATURE_ID}$"; then
            echo "❌ Orphan analysis file: ${ANALYSIS_FILE}"
            echo "   Feature ID '${FEATURE_ID}' not found in spec/features.yaml"
            echo ""
            ERRORS=$((ERRORS + 1))
        fi
    done
fi

# Check outline files
OUTLINE_DIR="docs/engine/outlines"
if [ -d "$OUTLINE_DIR" ]; then
    for OUTLINE_FILE in "$OUTLINE_DIR"/*_IMPLEMENTATION_OUTLINE.md; do
        # Skip if no files match the pattern
        [ -f "$OUTLINE_FILE" ] || continue

        # Skip IMPLEMENTATION_OUTLINE_TEMPLATE.md
        if [ "$(basename "$OUTLINE_FILE")" = "IMPLEMENTATION_OUTLINE_TEMPLATE.md" ]; then
            continue
        fi

        # Extract FEATURE_ID from filename (e.g., CLI_PLAN_IMPLEMENTATION_OUTLINE.md -> CLI_PLAN)
        FEATURE_ID=$(basename "$OUTLINE_FILE" _IMPLEMENTATION_OUTLINE.md)

        # Check if this FEATURE_ID exists in features.yaml
        if ! echo "$FEATURE_IDS" | grep -q "^${FEATURE_ID}$"; then
            echo "❌ Orphan outline file: ${OUTLINE_FILE}"
            echo "   Feature ID '${FEATURE_ID}' not found in spec/features.yaml"
            echo ""
            ERRORS=$((ERRORS + 1))
        fi
    done
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $ERRORS -eq 0 ]; then
    echo "✓ No orphan Analysis/Outline files found"
    exit 0
else
    echo "❌ Found ${ERRORS} orphan file(s)"
    echo ""
    echo "Orphan files are Analysis/Outline documents that don't have a corresponding"
    echo "entry in spec/features.yaml. These should either be:"
    echo "  - Removed if the feature was cancelled"
    echo "  - Moved to docs/archive/ if the feature is historical"
    echo "  - Have their Feature ID added to spec/features.yaml if the feature is active"
    exit 1
fi


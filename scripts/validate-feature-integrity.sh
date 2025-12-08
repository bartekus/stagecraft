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
# validate-feature-integrity.sh - Validate that all Feature IDs have required artifacts
#
# This script ensures that every Feature ID in spec/features.yaml has:
#   - A spec file
#   - An implementation outline
#   - An analysis brief
#   - An entry in spec/features.yaml (checked by parsing the file)
#
# Exit code: 0 if all features are valid, 1 if any are missing artifacts

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

FEATURES_YAML="spec/features.yaml"
ERRORS=0
WARNINGS=0

if [ ! -f "$FEATURES_YAML" ]; then
    echo "Error: $FEATURES_YAML not found"
    exit 1
fi

# Extract all Feature IDs from features.yaml
# This regex matches "id: FEATURE_ID" lines
FEATURE_IDS=$(grep -E "^\s+- id:" "$FEATURES_YAML" | sed 's/.*id: //' | tr -d '"' | tr -d "'")

if [ -z "$FEATURE_IDS" ]; then
    echo "Warning: No Feature IDs found in $FEATURES_YAML"
    exit 0
fi

echo "Validating feature integrity..."
echo ""

for FEATURE_ID in $FEATURE_IDS; do
    MISSING=()
    
    # Check for spec file
    # Extract spec path from features.yaml for this feature
    SPEC_PATH=$(grep -A 10 "id: ${FEATURE_ID}" "$FEATURES_YAML" | grep "spec:" | head -1 | sed 's/.*spec: *"\([^"]*\)".*/\1/' | sed "s/.*spec: *'\([^']*\)'.*/\1/")
    
    if [ -z "$SPEC_PATH" ]; then
        MISSING+=("spec entry in features.yaml")
    else
        if [ ! -f "spec/${SPEC_PATH}" ]; then
            MISSING+=("spec file: spec/${SPEC_PATH}")
        fi
    fi
    
    # Check for analysis brief
    ANALYSIS_FILE="docs/engine/analysis/${FEATURE_ID}.md"
    if [ ! -f "$ANALYSIS_FILE" ]; then
        MISSING+=("analysis brief: ${ANALYSIS_FILE}")
    fi
    
    # Check for implementation outline
    OUTLINE_FILE="docs/engine/outlines/${FEATURE_ID}_IMPLEMENTATION_OUTLINE.md"
    if [ ! -f "$OUTLINE_FILE" ]; then
        MISSING+=("implementation outline: ${OUTLINE_FILE}")
    fi
    
    # Report results
    if [ ${#MISSING[@]} -gt 0 ]; then
        echo "❌ ${FEATURE_ID}:"
        for item in "${MISSING[@]}"; do
            echo "   Missing: ${item}"
        done
        echo ""
        ERRORS=$((ERRORS + 1))
    else
        echo "✓ ${FEATURE_ID}: all artifacts present"
    fi
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $ERRORS -eq 0 ]; then
    echo "✓ All features have required artifacts"
    exit 0
else
    echo "❌ Found ${ERRORS} feature(s) with missing artifacts"
    echo ""
    echo "Required artifacts per feature:"
    echo "  - Analysis Brief: docs/engine/analysis/<FEATURE_ID>.md"
    echo "  - Implementation Outline: docs/engine/outlines/<FEATURE_ID>_IMPLEMENTATION_OUTLINE.md"
    echo "  - Spec file: spec/<domain>/<feature>.md (as specified in features.yaml)"
    echo "  - Entry in spec/features.yaml"
    exit 1
fi


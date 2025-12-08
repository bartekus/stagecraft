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
# This script ensures that every Feature ID in spec/features.yaml has the appropriate
# artifacts based on its status:
#   - todo: No artifacts required (INFO only)
#   - wip: Spec required (hard), analysis/outline recommended (WARNING)
#   - done: All artifacts required (spec, analysis, outline) - hard failure if missing
#
# Exit code: 0 if all features are valid, 1 if any done/wip features are missing required artifacts

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

FEATURES_YAML="spec/features.yaml"
ERRORS=0
WARNINGS=0
INFO_COUNT=0

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

echo "Validating feature integrity (status-aware enforcement)..."
echo ""

for FEATURE_ID in $FEATURE_IDS; do
    # Extract status for this feature
    STATUS=$(grep -A 10 "id: ${FEATURE_ID}" "$FEATURES_YAML" | grep "status:" | head -1 | sed 's/.*status: *"\([^"]*\)".*/\1/' | sed "s/.*status: *'\([^']*\)'.*/\1/" | tr -d ' ')
    
    # Extract spec path from features.yaml for this feature
    SPEC_PATH=$(grep -A 10 "id: ${FEATURE_ID}" "$FEATURES_YAML" | grep "spec:" | head -1 | sed 's/.*spec: *"\([^"]*\)".*/\1/' | sed "s/.*spec: *'\([^']*\)'.*/\1/")
    
    MISSING_REQUIRED=()
    MISSING_RECOMMENDED=()
    
    # Status-aware validation
    case "$STATUS" in
        todo)
            # No artifacts required for todo features
            INFO_COUNT=$((INFO_COUNT + 1))
            continue
            ;;
        wip)
            # wip: Spec is required (hard), analysis/outline are recommended (warning)
            if [ -z "$SPEC_PATH" ]; then
                MISSING_REQUIRED+=("spec entry in features.yaml")
            else
                if [ ! -f "spec/${SPEC_PATH}" ]; then
                    MISSING_REQUIRED+=("spec file: spec/${SPEC_PATH}")
                fi
            fi
            
            ANALYSIS_FILE="docs/engine/analysis/${FEATURE_ID}.md"
            if [ ! -f "$ANALYSIS_FILE" ]; then
                MISSING_RECOMMENDED+=("analysis brief: ${ANALYSIS_FILE}")
            fi
            
            OUTLINE_FILE="docs/engine/outlines/${FEATURE_ID}_IMPLEMENTATION_OUTLINE.md"
            if [ ! -f "$OUTLINE_FILE" ]; then
                MISSING_RECOMMENDED+=("implementation outline: ${OUTLINE_FILE}")
            fi
            ;;
        done)
            # done: All artifacts required (hard failure)
            if [ -z "$SPEC_PATH" ]; then
                MISSING_REQUIRED+=("spec entry in features.yaml")
            else
                if [ ! -f "spec/${SPEC_PATH}" ]; then
                    MISSING_REQUIRED+=("spec file: spec/${SPEC_PATH}")
                fi
            fi
            
            ANALYSIS_FILE="docs/engine/analysis/${FEATURE_ID}.md"
            if [ ! -f "$ANALYSIS_FILE" ]; then
                MISSING_REQUIRED+=("analysis brief: ${ANALYSIS_FILE}")
            fi
            
            OUTLINE_FILE="docs/engine/outlines/${FEATURE_ID}_IMPLEMENTATION_OUTLINE.md"
            if [ ! -f "$OUTLINE_FILE" ]; then
                MISSING_REQUIRED+=("implementation outline: ${OUTLINE_FILE}")
            fi
            ;;
        deprecated|removed)
            # Historical features - no enforcement
            continue
            ;;
        *)
            # Unknown status - treat as todo (no enforcement)
            INFO_COUNT=$((INFO_COUNT + 1))
            continue
            ;;
    esac
    
    # Report results
    if [ ${#MISSING_REQUIRED[@]} -gt 0 ]; then
        echo "❌ ${FEATURE_ID} (status: ${STATUS}):"
        for item in "${MISSING_REQUIRED[@]}"; do
            echo "   Missing (required): ${item}"
        done
        if [ ${#MISSING_RECOMMENDED[@]} -gt 0 ]; then
            for item in "${MISSING_RECOMMENDED[@]}"; do
                echo "   Missing (recommended): ${item}"
            done
        fi
        echo ""
        ERRORS=$((ERRORS + 1))
    elif [ ${#MISSING_RECOMMENDED[@]} -gt 0 ]; then
        echo "⚠️  ${FEATURE_ID} (status: ${STATUS}):"
        for item in "${MISSING_RECOMMENDED[@]}"; do
            echo "   Missing (recommended): ${item}"
        done
        echo ""
        WARNINGS=$((WARNINGS + 1))
    else
        echo "✓ ${FEATURE_ID} (status: ${STATUS}): all artifacts present"
    fi
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $ERRORS -eq 0 ]; then
    if [ $WARNINGS -gt 0 ]; then
        echo "✓ All required artifacts present"
        echo "⚠️  ${WARNINGS} feature(s) missing recommended artifacts (wip features)"
        if [ $INFO_COUNT -gt 0 ]; then
            echo "ℹ️  ${INFO_COUNT} todo feature(s) skipped (no artifacts required)"
        fi
        exit 0
    else
        echo "✓ All features have required artifacts"
        if [ $INFO_COUNT -gt 0 ]; then
            echo "ℹ️  ${INFO_COUNT} todo feature(s) skipped (no artifacts required)"
        fi
        exit 0
    fi
else
    echo "❌ Found ${ERRORS} feature(s) with missing required artifacts"
    if [ $WARNINGS -gt 0 ]; then
        echo "⚠️  ${WARNINGS} feature(s) missing recommended artifacts (wip features)"
    fi
    if [ $INFO_COUNT -gt 0 ]; then
        echo "ℹ️  ${INFO_COUNT} todo feature(s) skipped (no artifacts required)"
    fi
    echo ""
    echo "Enforcement rules:"
    echo "  - todo: No artifacts required"
    echo "  - wip: Spec required (hard), analysis/outline recommended (warning)"
    echo "  - done: All artifacts required (spec, analysis, outline) - hard failure"
    exit 1
fi


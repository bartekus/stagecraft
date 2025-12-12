#!/usr/bin/env bash
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
# check-doc-patterns.sh - Fail CI if new docs match forbidden patterns
#
# Usage:
#   ./scripts/check-doc-patterns.sh
#
# This script:
# - Checks for forbidden documentation patterns that should use canonical docs instead
# - Fails CI if any forbidden patterns are found
# - Provides guidance on which canonical doc to use

set -euo pipefail

# Defensive check: ensure we're running under bash
if [ -z "${BASH_VERSION:-}" ]; then
    echo "Error: This script requires bash. Current shell: ${SHELL:-unknown}" >&2
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

# Check for skip flag (local dev only, never in CI)
if [ "${STAGECRAFT_SKIP_DOC_PATTERNS:-0}" = "1" ]; then
    if [ -n "${CI:-}" ] || [ -n "${GITHUB_ACTIONS:-}" ]; then
        echo "Error: STAGECRAFT_SKIP_DOC_PATTERNS cannot be used in CI"
        exit 1
    fi
    echo "Skipping documentation pattern checks (STAGECRAFT_SKIP_DOC_PATTERNS=1)"
    exit 0
fi

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

ERRORS=0

error() {
    printf '%bERROR:%b %s\n' "${RED}" "${NC}" "$1" >&2
    ((ERRORS+=1))
}

# Check for forbidden patterns in docs/
echo "Checking for forbidden documentation patterns..."

# Get list of new/added files (staged or untracked)
# Use portable approach: combine outputs and filter empty lines
STAGED_NEW=$(git diff --cached --name-only --diff-filter=A 2>/dev/null || printf '')
UNTRACKED=$(git ls-files --others --exclude-standard 2>/dev/null || printf '')
WORKING_NEW=$(git diff --name-only --diff-filter=A 2>/dev/null || printf '')

# Combine all new files (portable across BSD/GNU)
# Suppress grep exit code (empty input causes grep to return 1, which would exit with set -e)
ALL_NEW_FILES=$(printf '%s\n%s\n%s\n' "$STAGED_NEW" "$UNTRACKED" "$WORKING_NEW" | grep -v '^$' || true | sort -u)

# Helper function to check if file is in new files list (handles grep exit code safely)
is_new_file() {
    local file="$1"
    [ -z "$ALL_NEW_FILES" ] && return 1
    printf '%s\n' "$ALL_NEW_FILES" | grep -Fqx "$file" 2>/dev/null || return 1
    return 0
}

# Exclude docs/archive/** and .DS_Store from pattern checks
# Helper function to check if file should be excluded (safety net for other find scopes)
should_exclude_file() {
    local file="$1"
    # Exclude docs/archive/** (including all subdirectories) and .DS_Store
    # Use prefix check for archive (handles nested paths)
    if [[ "$file" == docs/archive/* ]] || [[ "$file" == */.DS_Store ]] || [[ "$file" == *.DS_Store ]]; then
        return 0  # Exclude
    fi
    return 1  # Don't exclude
}

# Helper function: find docs/ excluding docs/archive/** using -prune (faster, cleaner)
# Prune both the directory and anything under it (explicit intent for future readers)
find_docs_pruned() {
    local pattern="$1"
    find docs/ \( -path docs/archive -o -path 'docs/archive/*' \) -prune -o -name "$pattern" -print 2>/dev/null || printf ''
}

# Pattern 1: *_COVERAGE_V1_COMPLETE.md
# Use pruned find to exclude docs/archive/** at the source
FOUND_FILES=$(find_docs_pruned "*_COVERAGE_V1_COMPLETE.md")
if [ -n "$FOUND_FILES" ]; then
    printf '%s\n' "$FOUND_FILES" | while IFS= read -r file; do
        [ -z "$file" ] && continue
        # Safety net: double-check exclusion (shouldn't be needed, but defensive)
        should_exclude_file "$file" && continue
        # Check if file is new (in our new files list)
        if is_new_file "$file"; then
            error "Forbidden pattern found: $file"
            echo "  → Use: docs/coverage/COVERAGE_LEDGER.md and docs/engine/history/<FEATURE_ID>_EVOLUTION.md instead"
            echo "  → See: Agent.md section 'Canonical Documentation Homes'"
        fi
    done || true
fi

# Pattern 2: *_SLICE*_PLAN.md
# Use pruned find to exclude docs/archive/** at the source
FOUND_FILES=$(find_docs_pruned "*_SLICE*_PLAN.md")
if [ -n "$FOUND_FILES" ]; then
    printf '%s\n' "$FOUND_FILES" | while IFS= read -r file; do
        [ -z "$file" ] && continue
        # Safety net: double-check exclusion (shouldn't be needed, but defensive)
        should_exclude_file "$file" && continue
        if is_new_file "$file"; then
            error "Forbidden pattern found: $file"
            echo "  → Use: docs/engine/history/<FEATURE_ID>_EVOLUTION.md instead"
            echo "  → See: Agent.md section 'Canonical Documentation Homes'"
        fi
    done || true
fi

# Pattern 3: COMMIT_*_PHASE*.md (in docs/governance/ or docs/todo/)
FOUND_FILES=$(find docs/governance/ docs/todo/ -name "COMMIT_*_PHASE*.md" 2>/dev/null || printf '')
if [ -n "$FOUND_FILES" ]; then
    printf '%s\n' "$FOUND_FILES" | while IFS= read -r file; do
        [ -z "$file" ] && continue
        # Skip archived files
        should_exclude_file "$file" && continue
        if is_new_file "$file"; then
            error "Forbidden pattern found: $file"
            echo "  → Use: docs/governance/GOVERNANCE_ALMANAC.md instead"
            echo "  → See: Agent.md section 'Canonical Documentation Homes'"
        fi
    done || true
fi

# Summary
if [ $ERRORS -eq 0 ]; then
    printf '%b✓%b No forbidden documentation patterns found\n' "${GREEN}" "${NC}"
    exit 0
else
    echo ""
    printf '%b✗%b Found %s forbidden documentation pattern(s)\n' "${RED}" "${NC}" "$ERRORS"
    echo ""
    echo "New documents matching these patterns are forbidden. Use the canonical documentation files instead:"
    echo "  - Provider evolution: docs/engine/history/<FEATURE_ID>_EVOLUTION.md"
    echo "  - Coverage tracking: docs/coverage/COVERAGE_LEDGER.md"
    echo "  - Governance rules: docs/governance/GOVERNANCE_ALMANAC.md"
    echo ""
    echo "See Agent.md section 'Canonical Documentation Homes' for details."
    exit 1
fi

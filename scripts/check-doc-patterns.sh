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
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m'

ERRORS=0

error() {
    printf '%bERROR:%b %s\n' "${RED}" "${NC}" "$1" >&2
    ((ERRORS++))
}

# Check for forbidden patterns in docs/
echo "Checking for forbidden documentation patterns..."

# Get list of new/added files (staged or untracked)
# Use portable approach: combine outputs and filter empty lines
STAGED_NEW=$(git diff --cached --name-only --diff-filter=A 2>/dev/null || printf '')
UNTRACKED=$(git ls-files --others --exclude-standard 2>/dev/null || printf '')
WORKING_NEW=$(git diff --name-only --diff-filter=A 2>/dev/null || printf '')

# Combine all new files (portable across BSD/GNU)
ALL_NEW_FILES=$(printf '%s\n%s\n%s\n' "$STAGED_NEW" "$UNTRACKED" "$WORKING_NEW" | grep -v '^$' | sort -u)

# Pattern 1: *_COVERAGE_V1_COMPLETE.md
# Collect files first (avoid subshell issues with while loop)
FOUND_FILES=$(find docs/ -name "*_COVERAGE_V1_COMPLETE.md" 2>/dev/null || printf '')
if [ -n "$FOUND_FILES" ]; then
    printf '%s\n' "$FOUND_FILES" | while IFS= read -r file; do
        [ -z "$file" ] && continue
        # Check if file is new (in our new files list)
        # Use grep -F for fixed string matching (more portable)
        if printf '%s\n' "$ALL_NEW_FILES" | grep -Fqx "$file"; then
            error "Forbidden pattern found: $file"
            echo "  → Use: docs/coverage/COVERAGE_LEDGER.md and docs/engine/history/<FEATURE_ID>_EVOLUTION.md instead"
            echo "  → See: Agent.md section 'Canonical Documentation Homes'"
        fi
    done
fi

# Pattern 2: *_SLICE*_PLAN.md
FOUND_FILES=$(find docs/ -name "*_SLICE*_PLAN.md" 2>/dev/null || printf '')
if [ -n "$FOUND_FILES" ]; then
    printf '%s\n' "$FOUND_FILES" | while IFS= read -r file; do
        [ -z "$file" ] && continue
        if printf '%s\n' "$ALL_NEW_FILES" | grep -Fqx "$file"; then
            error "Forbidden pattern found: $file"
            echo "  → Use: docs/engine/history/<FEATURE_ID>_EVOLUTION.md instead"
            echo "  → See: Agent.md section 'Canonical Documentation Homes'"
        fi
    done
fi

# Pattern 3: COMMIT_*_PHASE*.md (in docs/governance/ or docs/todo/)
FOUND_FILES=$(find docs/governance/ docs/todo/ -name "COMMIT_*_PHASE*.md" 2>/dev/null || printf '')
if [ -n "$FOUND_FILES" ]; then
    printf '%s\n' "$FOUND_FILES" | while IFS= read -r file; do
        [ -z "$file" ] && continue
        if printf '%s\n' "$ALL_NEW_FILES" | grep -Fqx "$file"; then
            error "Forbidden pattern found: $file"
            echo "  → Use: docs/governance/GOVERNANCE_ALMANAC.md instead"
            echo "  → See: Agent.md section 'Canonical Documentation Homes'"
        fi
    done
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

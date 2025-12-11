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
# Pre-commit hook snippet: Governance checks
# 
# This snippet can be added to your .git/hooks/pre-commit hook (or .hooks/pre-commit)
# to run governance validation before commits.
#
# Escape hatch: Set SKIP_GOV_PRE_COMMIT=1 to bypass (e.g., SKIP_GOV_PRE_COMMIT=1 git commit)

# Escape hatch: allow bypassing governance checks
if [ "${SKIP_GOV_PRE_COMMIT:-}" = "1" ]; then
    echo "⚠️  Skipping governance pre-commit checks (SKIP_GOV_PRE_COMMIT=1)"
    exit 0
fi

# Also respect the general hook skip flag
if [ "${STAGECRAFT_SKIP_HOOKS:-}" = "1" ] || [ "${SKIP_HOOKS:-}" = "1" ]; then
    echo "⚠️  Skipping governance pre-commit checks (general hook skip enabled)"
    exit 0
fi

# Get project root
PROJECT_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$PROJECT_ROOT"

# Run governance pre-commit script if it exists
if [ -x "$PROJECT_ROOT/scripts/gov-pre-commit.sh" ]; then
    if ! bash "$PROJECT_ROOT/scripts/gov-pre-commit.sh"; then
        echo ""
        echo "❌ Governance pre-commit checks failed."
        echo "   Fix the issues above, or skip with: SKIP_GOV_PRE_COMMIT=1 git commit"
        exit 1
    fi
else
    echo "⚠️  Governance pre-commit script not found: scripts/gov-pre-commit.sh"
    echo "   Skipping governance checks (this is a warning, not an error)"
    exit 0
fi

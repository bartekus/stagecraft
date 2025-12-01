#!/bin/bash
# SPDX-License-Identifier: AGPL-3.0-or-later
#
# Stagecraft - A Go-based CLI for orchestrating local-first multi-service deployments using Docker Compose.
#
# Copyright (C) 2025  Bartek Kus
#
# This program is free software licensed under the terms of the GNU AGPL v3 or later.
#
# See https://www.gnu.org/licenses/ for license details.
#
# check-coverage.sh - Check test coverage against thresholds
#
# Usage:
#   ./scripts/check-coverage.sh [--fail-on-warning]
#
# This script:
# - Runs tests with coverage
# - Checks overall coverage (warns if < 60%, fails if < 50%)
# - Checks core package coverage (fails if < 80%)
# - Reports per-package coverage

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

FAIL_ON_WARNING=${1:-""}
COVERAGE_FILE="coverage.out"

# Colors
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m'

# Coverage thresholds
OVERALL_MIN=60
OVERALL_CRITICAL=50
CORE_PACKAGE_MIN=80

ERRORS=0
WARNINGS=0

error() {
    echo -e "${RED}ERROR:${NC} $1" >&2
    ((ERRORS++))
}

warning() {
    echo -e "${YELLOW}WARNING:${NC} $1" >&2
    ((WARNINGS++))
}

info() {
    echo -e "${GREEN}✓${NC} $1"
}

# Run tests with coverage
echo "Running tests with coverage..."
go test ./... -coverprofile="$COVERAGE_FILE" -covermode=atomic

if [ ! -f "$COVERAGE_FILE" ]; then
    error "Coverage file not generated"
    exit 1
fi

# Get overall coverage
OVERALL_COVERAGE=$(go tool cover -func="$COVERAGE_FILE" | grep total | awk '{print $3}' | sed 's/%//')

echo ""
echo "=== Coverage Report ==="
echo ""

# Check overall coverage
if (( $(echo "$OVERALL_COVERAGE < $OVERALL_CRITICAL" | bc -l 2>/dev/null || echo "0") )); then
    error "Overall coverage is ${OVERALL_COVERAGE}%, below critical threshold of ${OVERALL_CRITICAL}%"
elif (( $(echo "$OVERALL_COVERAGE < $OVERALL_MIN" | bc -l 2>/dev/null || echo "0") )); then
    warning "Overall coverage is ${OVERALL_COVERAGE}%, below target of ${OVERALL_MIN}%"
    if [ "$FAIL_ON_WARNING" = "--fail-on-warning" ]; then
        ((ERRORS++))
    fi
else
    info "Overall coverage: ${OVERALL_COVERAGE}%"
fi

# Check core packages
echo ""
echo "=== Core Package Coverage ==="
echo ""

CORE_PACKAGES=(
    "pkg/config"
    "internal/core"
)

CORE_FAILED=0
for pkg in "${CORE_PACKAGES[@]}"; do
    # Check if package exists
    if ! go list "./$pkg" &>/dev/null; then
        continue
    fi
    
    PKG_COVERAGE=$(go tool cover -func="$COVERAGE_FILE" | grep "$pkg" | awk '{print $3}' | sed 's/%//' || echo "0")
    
    if [ "$PKG_COVERAGE" = "0" ] || [ -z "$PKG_COVERAGE" ]; then
        warning "$pkg: No coverage data (package may not have tests)"
        continue
    fi
    
    if (( $(echo "$PKG_COVERAGE < $CORE_PACKAGE_MIN" | bc -l 2>/dev/null || echo "0") )); then
        error "$pkg: ${PKG_COVERAGE}% (below ${CORE_PACKAGE_MIN}% threshold)"
        CORE_FAILED=1
    else
        info "$pkg: ${PKG_COVERAGE}%"
    fi
done

# Per-package breakdown
echo ""
echo "=== Per-Package Coverage ==="
echo ""

go tool cover -func="$COVERAGE_FILE" | grep -v "^total:" | while IFS= read -r line; do
    PKG=$(echo "$line" | awk '{print $1}')
    COV=$(echo "$line" | awk '{print $3}' | sed 's/%//')
    
    if (( $(echo "$COV < 50" | bc -l 2>/dev/null || echo "0") )); then
        echo -e "${RED}$PKG: ${COV}%${NC}"
    elif (( $(echo "$COV < 70" | bc -l 2>/dev/null || echo "0") )); then
        echo -e "${YELLOW}$PKG: ${COV}%${NC}"
    else
        echo -e "${GREEN}$PKG: ${COV}%${NC}"
    fi
done

# Summary
echo ""
echo "=== Summary ==="
echo "Overall coverage: ${OVERALL_COVERAGE}%"
echo ""

if [ $ERRORS -gt 0 ]; then
    error "$ERRORS error(s) found"
    exit 1
elif [ $WARNINGS -gt 0 ]; then
    warning "$WARNINGS warning(s) found"
    if [ "$FAIL_ON_WARNING" = "--fail-on-warning" ]; then
        exit 1
    fi
    exit 0
else
    info "All coverage checks passed!"
    exit 0
fi


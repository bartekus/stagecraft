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
# run-all-checks.sh - Runs all checks that match CI workflow
#
# Usage:
#   ./scripts/run-all-checks.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info() {
    echo -e "${GREEN}✓${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1" >&2
}

section() {
    echo ""
    echo "=== $1 ==="
    echo ""
}

# === Lint Checks (matches CI lint job) ===
section "Lint and Format Checks"

info "Checking Go formatting..."
# Use canonical formatting script
if [ -x ./scripts/goformat.sh ]; then
    # Check formatting (dry run)
    if command -v gofumpt &> /dev/null; then
        format_out=$(gofumpt -l .)
        if [ -n "$format_out" ]; then
            error "The following files are not gofumpt'ed:"
            echo "$format_out"
            error "Run ./scripts/goformat.sh to fix formatting"
            exit 1
        fi
        info "All files are properly formatted"
    else
        error "gofumpt not found. Install with: go install mvdan.cc/gofumpt@v0.6.0"
        exit 1
    fi
else
    error "scripts/goformat.sh not found"
    exit 1
fi

info "Running golangci-lint..."
if ! command -v golangci-lint &> /dev/null; then
    error "golangci-lint is not installed. Install with: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.6.2"
    exit 1
fi
golangci-lint run ./...
info "golangci-lint passed"

# === Test Checks (matches CI test job) ===
section "Tests and Coverage"

info "Building all packages..."
if ! go build ./...; then
    error "Failed to build all packages"
    exit 1
fi
info "All packages build successfully"

info "Building stagecraft binary..."
rm -rf bin
go build -o ./bin/stagecraft ./cmd/stagecraft
info "Build successful"

info "Running tests with coverage..."
go test ./... -coverprofile=coverage.out -covermode=atomic
info "Tests passed"

info "Checking coverage thresholds..."
if [ -x ./scripts/check-coverage.sh ]; then
    ./scripts/check-coverage.sh --fail-on-warning
    info "Coverage check passed"
else
    error "scripts/check-coverage.sh not found"
    exit 1
fi

# === Docs and Spec Checks (matches CI docs-and-spec job) ===
section "Docs and Spec Validation"

info "Validating spec/features.yaml YAML syntax..."
if [ -f spec/features.yaml ]; then
    if ! python3 -c "import yaml; yaml.safe_load(open('spec/features.yaml'))" 2>/dev/null; then
        error "spec/features.yaml is not valid YAML"
        exit 1
    fi
    info "spec/features.yaml is valid YAML"
else
    error "spec/features.yaml not found"
    exit 1
fi

info "Running spec validation script..."
if [ -f "scripts/validate-spec.sh" ]; then
    bash scripts/validate-spec.sh
    info "Spec validation passed"
else
    error "Spec validation script not found"
    exit 1
fi

info "Checking for missing spec files..."
if ! go run ./cmd/spec-reference-check; then
    error "Spec reference validation failed"
    exit 1
fi
info "All spec file references are valid"

info "Checking for orphan Analysis/Outline files..."
if [ -x ./scripts/check-orphan-docs.sh ]; then
    ./scripts/check-orphan-docs.sh
    info "Orphan docs check passed"
else
    error "scripts/check-orphan-docs.sh not found"
    exit 1
fi

info "Checking for orphan spec files..."
if [ -x ./scripts/check-orphan-specs.sh ]; then
    ./scripts/check-orphan-specs.sh
    info "Orphan specs check passed"
else
    error "scripts/check-orphan-specs.sh not found"
    exit 1
fi

# === Governance Checks (GOV_V1_CORE) ===
section "Governance Checks"

info "Validating spec frontmatter..."
if ! go run ./cmd/spec-validate --check-integrity; then
    error "Spec frontmatter validation failed"
    exit 1
fi
info "Spec frontmatter validation passed"

info "Validating feature dependency graph..."
if ! go run ./cmd/features-tool graph; then
    error "Feature dependency graph validation failed"
    exit 1
fi
info "Feature dependency graph is valid"

info "Checking CLI vs Spec alignment (flags)..."
if ! go run ./cmd/spec-vs-cli; then
    error "CLI vs Spec alignment check failed"
    exit 1
fi
info "CLI vs Spec alignment check passed"

info "Generating feature overview..."
if ! go run ./cmd/gen-features-overview; then
    error "Failed to generate feature overview"
    exit 1
fi

info "Verifying feature overview is up to date..."
if ! git diff --exit-code docs/features/OVERVIEW.md 2>/dev/null; then
    error "docs/features/OVERVIEW.md is out of date; please regenerate and commit"
    error "Run: go run ./cmd/gen-features-overview"
    exit 1
fi
info "Feature overview is up to date"

# === License Checks (matches CI license job) ===
section "License Headers"

info "Checking license headers..."
if ! command -v addlicense &> /dev/null; then
    error "addlicense is not installed. Install with: go install github.com/google/addlicense@latest"
    exit 1
fi

addlicense -ignore 'internal/providers/backend/generic/test_script.sh' -ignore '.bin/vendor/**' -check .
info "License headers check passed"

# === Summary ===
section "Summary"
info "All checks passed!"
echo ""

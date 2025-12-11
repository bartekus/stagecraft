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
set -euo pipefail

echo "Running provider governance checks..."

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

FEATURES_FILE="spec/features.yaml"

if [ ! -f "$FEATURES_FILE" ]; then
    echo "spec/features.yaml not found, skipping provider governance checks"
    exit 0
fi

status=0

# Track which providers have coverage strategies
declare -A has_strategy

# Scan all provider coverage strategy files
for f in internal/providers/*/*/COVERAGE_STRATEGY.md; do
    [ -e "$f" ] || continue

    # Extract Feature ID from first heading: "# PROVIDER_X - Coverage Strategy (V1 Complete)"
    first_heading="$(grep -m1 '^# ' "$f" || true)"

    if [ -z "$first_heading" ]; then
        echo "WARN: $f has no top-level heading, cannot infer Feature ID"
        status=1
        continue
    fi

    # Assume pattern: "# FEATURE_ID ..." and take the second token
    feature_id="$(echo "$first_heading" | awk '{print $2}')"

    if [ -z "$feature_id" ]; then
        echo "WARN: Unable to parse Feature ID from heading in $f"
        status=1
        continue
    fi

    # Mark this provider as having a strategy
    has_strategy["$feature_id"]=1

    # Verify feature exists in spec/features.yaml
    if ! grep -q "id: ${feature_id}" "$FEATURES_FILE"; then
        echo "ERROR: Coverage strategy $f references Feature ID '${feature_id}' which is not present in spec/features.yaml"
        status=1
        continue
    fi

    # Check for V1 Complete label in coverage heading
    if echo "$first_heading" | grep -qi "V1 Complete"; then
        status_doc="docs/engine/status/${feature_id}_COVERAGE_V1_COMPLETE.md"
        if [ ! -f "$status_doc" ]; then
            echo "WARN: ${feature_id} coverage marked 'V1 Complete' in $f but status doc ${status_doc} is missing"
            # warn only, do not force failure yet
        fi
    fi
done

# Check for missing coverage strategies: all PROVIDER_* features with status: done must have COVERAGE_STRATEGY.md
echo "Checking for missing provider coverage strategies..."

# Map of known providers to their expected paths
declare -A provider_paths=(
    ["PROVIDER_BACKEND_ENCORE"]="internal/providers/backend/encorets/COVERAGE_STRATEGY.md"
    ["PROVIDER_BACKEND_GENERIC"]="internal/providers/backend/generic/COVERAGE_STRATEGY.md"
    ["PROVIDER_CLOUD_DO"]="internal/providers/cloud/digitalocean/COVERAGE_STRATEGY.md"
    ["PROVIDER_NETWORK_TAILSCALE"]="internal/providers/network/tailscale/COVERAGE_STRATEGY.md"
    ["PROVIDER_FRONTEND_GENERIC"]="internal/providers/frontend/generic/COVERAGE_STRATEGY.md"
)

missing_count=0

# Check each known provider
for provider_id in "${!provider_paths[@]}"; do
    # Check if this provider is marked as done in features.yaml
    if grep -A 5 "^  - id: ${provider_id}" "$FEATURES_FILE" | grep -q "status: done"; then
        expected_path="${provider_paths[$provider_id]}"
        
        if [ -z "${has_strategy[$provider_id]:-}" ] && [ ! -f "$expected_path" ]; then
            echo "ERROR: ${provider_id} (status: done) is missing COVERAGE_STRATEGY.md"
            echo "       Expected: ${expected_path}"
            echo "       Create from template: docs/coverage/PROVIDER_COVERAGE_TEMPLATE.md"
            status=1
            missing_count=$((missing_count + 1))
        fi
    fi
done

if [ $status -ne 0 ]; then
    echo ""
    echo "Provider governance checks FAILED"
    echo "Missing coverage strategies: ${missing_count}"
    exit $status
fi

echo "Provider governance checks passed"

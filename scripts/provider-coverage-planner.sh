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

# Provider Coverage Completion Planner
# Outputs a structured plan for bringing providers to V1 Complete coverage status.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

FEATURES_FILE="spec/features.yaml"

if [ ! -f "$FEATURES_FILE" ]; then
    echo "spec/features.yaml not found"
    exit 1
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Provider Coverage Completion Planner"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Map of known providers
declare -A provider_info=(
    ["PROVIDER_BACKEND_ENCORE"]="backend/encorets"
    ["PROVIDER_BACKEND_GENERIC"]="backend/generic"
    ["PROVIDER_CLOUD_DO"]="cloud/digitalocean"
    ["PROVIDER_NETWORK_TAILSCALE"]="network/tailscale"
    ["PROVIDER_FRONTEND_GENERIC"]="frontend/generic"
)

# Check each provider
for provider_id in "${!provider_info[@]}"; do
    # Check if provider is marked as done
    if ! grep -A 5 "^  - id: ${provider_id}" "$FEATURES_FILE" | grep -q "status: done"; then
        continue
    fi
    
    provider_path="${provider_info[$provider_id]}"
    strategy_file="internal/providers/${provider_path}/COVERAGE_STRATEGY.md"
    
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Provider: ${provider_id}"
    echo "Path: ${provider_path}"
    echo ""
    
    # Check for coverage strategy
    if [ -f "$strategy_file" ]; then
        echo "✅ Coverage Strategy: Present"
        
        # Extract status from heading
        first_heading="$(grep -m1 '^# ' "$strategy_file" 2>/dev/null || true)"
        if echo "$first_heading" | grep -qi "V1 Complete"; then
            echo "   Status: V1 Complete"
        elif echo "$first_heading" | grep -qi "V1 Plan\|In Progress"; then
            echo "   Status: V1 Plan / In Progress"
        else
            echo "   Status: Unknown (check heading)"
        fi
        
        # Try to extract coverage percentage
        coverage_line="$(grep -i "coverage.*%" "$strategy_file" | head -1 || true)"
        if [ -n "$coverage_line" ]; then
            echo "   ${coverage_line}"
        fi
    else
        echo "❌ Coverage Strategy: MISSING"
        echo "   Expected: ${strategy_file}"
        echo "   Action: Create from docs/coverage/PROVIDER_COVERAGE_TEMPLATE.md"
    fi
    
    # Check for status doc if V1 Complete
    if [ -f "$strategy_file" ]; then
        first_heading="$(grep -m1 '^# ' "$strategy_file" 2>/dev/null || true)"
        if echo "$first_heading" | grep -qi "V1 Complete"; then
            status_doc="docs/engine/status/${provider_id}_COVERAGE_V1_COMPLETE.md"
            if [ -f "$status_doc" ]; then
                echo "✅ Status Doc: Present"
            else
                echo "⚠️  Status Doc: MISSING (required for V1 Complete)"
                echo "   Expected: ${status_doc}"
            fi
        fi
    fi
    
    # Try to get actual coverage
    if [ -d "internal/providers/${provider_path}" ]; then
        coverage_output="$(go test -cover "./internal/providers/${provider_path}" 2>&1 | grep -E "coverage:" || true)"
        if [ -n "$coverage_output" ]; then
            echo "📊 Current Coverage: ${coverage_output}"
        fi
    fi
    
    echo ""
done

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Summary"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Count statuses
complete_count=0
plan_count=0
missing_count=0

for provider_id in "${!provider_info[@]}"; do
    if ! grep -A 5 "^  - id: ${provider_id}" "$FEATURES_FILE" | grep -q "status: done"; then
        continue
    fi
    
    provider_path="${provider_info[$provider_id]}"
    strategy_file="internal/providers/${provider_path}/COVERAGE_STRATEGY.md"
    
    if [ ! -f "$strategy_file" ]; then
        missing_count=$((missing_count + 1))
    elif grep -m1 '^# ' "$strategy_file" 2>/dev/null | grep -qi "V1 Complete"; then
        complete_count=$((complete_count + 1))
    else
        plan_count=$((plan_count + 1))
    fi
done

echo "V1 Complete: ${complete_count}"
echo "V1 Plan: ${plan_count}"
echo "Missing Strategy: ${missing_count}"
echo ""
echo "Next Actions:"
echo "  1. Create missing coverage strategies from template"
echo "  2. Review providers in 'V1 Plan' status for flakiness"
echo "  3. Bring providers to V1 Complete following PROVIDER_COVERAGE_AGENT.md"
echo ""

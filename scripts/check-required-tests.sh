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
# check-required-tests.sh - Validates that required tests from Implementation Outline exist
#
# This script checks:
# - If Implementation Outline requires golden tests, at least one golden test must exist
# - If Implementation Outline requires integration tests, integration tests must exist
# - Test files match the requirements stated in the outline

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

ERRORS=0
WARNINGS=0

# Colors for output
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m'

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

# Check if Python 3 is available
if ! command -v python3 &> /dev/null; then
    error "python3 is required for validation"
    exit 1
fi

echo "Checking required tests against Implementation Outlines..."

python3 << 'PYTHON_SCRIPT'
import yaml
import os
import sys
import re

errors = 0
warnings = 0

# Load features.yaml
with open('spec/features.yaml', 'r') as f:
    data = yaml.safe_load(f)

if 'features' not in data:
    print("ERROR: 'features' key not found in spec/features.yaml")
    sys.exit(1)

for feature in data['features']:
    feature_id = feature.get('id', 'UNKNOWN')
    status = feature.get('status', 'unknown')
    
    # Only check done/wip features
    if status not in ['done', 'wip']:
        continue
    
    outline_file = f"docs/{feature_id}_IMPLEMENTATION_OUTLINE.md"
    if not os.path.exists(outline_file):
        continue
    
    # Read outline
    try:
        with open(outline_file, 'r') as f:
            outline_content = f.read()
    except Exception as e:
        print(f"WARNING: {feature_id}: could not read outline: {e}")
        warnings += 1
        continue
    
    # Check for golden test requirements
    requires_golden = any(phrase in outline_content.lower() for phrase in [
        'golden test',
        'golden file',
        'golden tests',
        'golden files',
        'testdata/'
    ])
    
    if requires_golden:
        # Look for golden test files
        # Golden tests are typically in testdata/ directories
        has_golden = False
        
        # Check test files listed in features.yaml
        test_files = feature.get('tests', [])
        for test_file in test_files:
            test_dir = os.path.dirname(test_file)
            testdata_dir = os.path.join(test_dir, 'testdata')
            if os.path.exists(testdata_dir):
                # Check for .golden files
                golden_files = [f for f in os.listdir(testdata_dir) if f.endswith('.golden')]
                if golden_files:
                    has_golden = True
                    break
        
        # Also check common testdata locations
        common_testdata_paths = [
            f"internal/cli/commands/testdata",
            f"pkg/config/testdata",
        ]
        for testdata_path in common_testdata_paths:
            if os.path.exists(testdata_path):
                golden_files = [f for f in os.listdir(testdata_path) if f.endswith('.golden')]
                if golden_files:
                    has_golden = True
                    break
        
        if not has_golden:
            print(f"WARNING: {feature_id}: Implementation Outline requires golden tests, but no .golden files found")
            warnings += 1
        else:
            print(f"✓ {feature_id}: golden tests found")
    
    # Check for integration test requirements
    requires_integration = any(phrase in outline_content.lower() for phrase in [
        'integration test',
        'integration tests',
        'end-to-end',
        'e2e'
    ])
    
    if requires_integration:
        # Look for integration test files
        has_integration = False
        
        test_files = feature.get('tests', [])
        for test_file in test_files:
            if 'integration' in test_file.lower() or 'e2e' in test_file.lower() or 'smoke' in test_file.lower():
                if os.path.exists(test_file):
                    has_integration = True
                    break
        
        # Also check test/e2e directory
        if os.path.exists('test/e2e'):
            e2e_files = [f for f in os.listdir('test/e2e') if f.endswith('_test.go')]
            if e2e_files:
                # Check if any e2e test references this feature
                for e2e_file in e2e_files:
                    e2e_path = os.path.join('test/e2e', e2e_file)
                    try:
                        with open(e2e_path, 'r') as f:
                            content = f.read()
                        if feature_id in content:
                            has_integration = True
                            break
                    except Exception:
                        pass
        
        if not has_integration:
            print(f"WARNING: {feature_id}: Implementation Outline requires integration tests, but none found")
            warnings += 1
        else:
            print(f"✓ {feature_id}: integration tests found")
    
    # Check for unit test requirements (most features should have these)
    requires_unit = 'unit test' in outline_content.lower() or 'unit tests' in outline_content.lower()
    if requires_unit:
        test_files = feature.get('tests', [])
        has_unit = any(os.path.exists(tf) for tf in test_files if '_test.go' in tf)
        
        if not has_unit and not test_files:
            print(f"WARNING: {feature_id}: Implementation Outline requires unit tests, but none listed")
            warnings += 1
        elif has_unit:
            print(f"✓ {feature_id}: unit tests found")

sys.exit(0 if errors == 0 else 1)
PYTHON_SCRIPT

TEST_EXIT=$?

if [ $TEST_EXIT -ne 0 ]; then
    error "Required tests check failed"
    exit 1
fi

echo ""
if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    info "All required test checks passed!"
    exit 0
elif [ $ERRORS -eq 0 ]; then
    warning "$WARNINGS warning(s) found (non-blocking)"
    exit 0
else
    error "$ERRORS error(s) and $WARNINGS warning(s) found"
    exit 1
fi


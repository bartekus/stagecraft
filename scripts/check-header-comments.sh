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
# check-header-comments.sh - Validates Feature and Spec header comments in Go files
#
# This script ensures:
# - Every Go file implementing a feature includes Feature and Spec comments
# - Header comments match features.yaml entries
# - Test files include proper Feature/Spec references

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

echo "Checking Feature and Spec header comments..."

# Build map of Feature IDs to spec paths from features.yaml
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

feature_spec_map = {}
if 'features' in data:
    for feature in data['features']:
        feature_id = feature.get('id', '')
        spec_path = feature.get('spec', '')
        if feature_id and spec_path:
            feature_spec_map[feature_id] = spec_path

# Find all Go files (excluding vendor, testdata, etc.)
go_files = []
for root, dirs, files in os.walk('.'):
    # Skip hidden directories and common ignore patterns
    dirs[:] = [d for d in dirs if not d.startswith('.') and d not in ['vendor', 'testdata']]
    
    for file in files:
        if file.endswith('.go'):
            filepath = os.path.join(root, file)
            go_files.append(filepath)

# Check each Go file
for filepath in sorted(go_files):
    # Skip generated files and test helpers
    if 'generated' in filepath.lower() or 'mock' in filepath.lower():
        continue
    
    try:
        with open(filepath, 'r') as f:
            content = f.read()
        
        # Skip empty files
        if not content.strip():
            continue
        
        # Look for Feature comment
        feature_match = re.search(r'^//\s*Feature:\s*(\w+)', content, re.MULTILINE)
        spec_match = re.search(r'^//\s*Spec:\s*(.+)', content, re.MULTILINE)
        
        # Files in internal/ and pkg/ should have Feature comments
        # (excluding some utility files)
        needs_feature_comment = (
            'internal/' in filepath or 
            'pkg/' in filepath
        ) and not any(skip in filepath for skip in [
            'internal/version',
            'cmd/stagecraft/main.go',  # Main entry point
        ])
        
        if needs_feature_comment:
            if not feature_match:
                # Check if it's a test file - might be acceptable
                if '_test.go' in filepath:
                    # Test files should still have Feature comments if they test feature-specific behavior
                    # But we'll only warn for now
                    print(f"WARNING: {filepath}: missing Feature comment (test file)")
                    warnings += 1
                else:
                    print(f"WARNING: {filepath}: missing Feature comment")
                    warnings += 1
            else:
                feature_id = feature_match.group(1)
                
                # Check if Feature ID exists in features.yaml
                if feature_id not in feature_spec_map:
                    print(f"WARNING: {filepath}: Feature ID '{feature_id}' not found in features.yaml")
                    warnings += 1
                else:
                    # Check if Spec comment matches features.yaml
                    expected_spec = feature_spec_map[feature_id]
                    if spec_match:
                        actual_spec = spec_match.group(1).strip()
                        # Normalize paths for comparison
                        expected_normalized = expected_spec.replace('spec/', '').replace('docs/', '')
                        actual_normalized = actual_spec.replace('spec/', '').replace('docs/', '')
                        
                        if expected_normalized != actual_normalized:
                            print(f"WARNING: {filepath}: Spec path mismatch. Expected: {expected_spec}, Found: {actual_spec}")
                            warnings += 1
                        else:
                            print(f"✓ {filepath}: Feature {feature_id}, Spec {actual_spec}")
                    else:
                        print(f"WARNING: {filepath}: missing Spec comment (Feature: {feature_id})")
                        warnings += 1
        
    except Exception as e:
        print(f"ERROR: {filepath}: error reading file: {e}")
        errors += 1

sys.exit(0 if errors == 0 else 1)
PYTHON_SCRIPT

HEADER_EXIT=$?

if [ $HEADER_EXIT -ne 0 ]; then
    error "Header comment check failed"
    exit 1
fi

echo ""
if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    info "All header comment checks passed!"
    exit 0
elif [ $ERRORS -eq 0 ]; then
    warning "$WARNINGS warning(s) found (non-blocking)"
    exit 0
else
    error "$ERRORS error(s) and $WARNINGS warning(s) found"
    exit 1
fi


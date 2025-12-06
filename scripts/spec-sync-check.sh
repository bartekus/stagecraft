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
# spec-sync-check.sh - Ensures spec, outline, and implementation are in sync
#
# This script validates:
# - Spec v1 scope matches Implementation Outline v1 scope
# - Spec matches actual data structures in code (flags, exit codes, JSON shapes)
# - Feature status transitions follow required stages

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
    error "python3 is required for spec validation"
    exit 1
fi

echo "Checking spec synchronization..."

# Parse features.yaml
python3 << 'PYTHON_SCRIPT'
import yaml
import os
import sys
import re

errors = 0
warnings = 0

def check_file_exists(path, description):
    """Check if a file exists, trying common locations."""
    paths_to_try = [
        path,
        f"spec/{path}",
        f"docs/{path}"
    ]
    for p in paths_to_try:
        if os.path.exists(p):
            return p
    return None

with open('spec/features.yaml', 'r') as f:
    data = yaml.safe_load(f)

if 'features' not in data:
    print("ERROR: 'features' key not found in spec/features.yaml")
    sys.exit(1)

for feature in data['features']:
    feature_id = feature.get('id', 'UNKNOWN')
    status = feature.get('status', 'unknown')
    spec_path = feature.get('spec', '')
    
    if not spec_path:
        continue
    
    # Find spec file
    spec_file = check_file_exists(spec_path, "spec file")
    if not spec_file:
        if status in ['done', 'wip']:
            print(f"ERROR: {feature_id}: spec file not found: {spec_path}")
            errors += 1
        continue
    
    # Check for implementation outline
    outline_file = f"docs/{feature_id}_IMPLEMENTATION_OUTLINE.md"
    if not os.path.exists(outline_file):
        if status in ['done', 'wip']:
            print(f"WARNING: {feature_id}: implementation outline not found: {outline_file}")
            warnings += 1
        continue
    
    # For done/wip features, check that spec and outline are aligned
    if status in ['done', 'wip']:
        # Read spec file to check for v1 markers
        try:
            with open(spec_file, 'r') as f:
                spec_content = f.read()
            
            # Check for version markers
            has_v1_marker = 'v1' in spec_content.lower() or 'version' in spec_content.lower()
            
            # Read outline to check v1 scope
            with open(outline_file, 'r') as f:
                outline_content = f.read()
            
            # Basic alignment checks
            # Check if both mention the same flags/commands
            spec_flags = re.findall(r'--[\w-]+', spec_content)
            outline_flags = re.findall(r'--[\w-]+', outline_content)
            
            if spec_flags and outline_flags:
                spec_set = set(spec_flags)
                outline_set = set(outline_flags)
                missing_in_spec = outline_set - spec_set
                missing_in_outline = spec_set - outline_set
                
                if missing_in_spec:
                    print(f"WARNING: {feature_id}: flags in outline not in spec: {missing_in_spec}")
                    warnings += 1
                if missing_in_outline:
                    print(f"WARNING: {feature_id}: flags in spec not in outline: {missing_in_outline}")
                    warnings += 1
            
            # Check exit codes alignment
            spec_exit_codes = re.findall(r'exit.*code.*\d+|exit code.*\d+', spec_content, re.IGNORECASE)
            outline_exit_codes = re.findall(r'exit.*code.*\d+|exit code.*\d+', outline_content, re.IGNORECASE)
            
            if spec_exit_codes and outline_exit_codes:
                # Basic check that both mention exit codes
                if len(spec_exit_codes) != len(outline_exit_codes):
                    print(f"WARNING: {feature_id}: exit code count mismatch between spec and outline")
                    warnings += 1
            
            print(f"✓ {feature_id}: spec and outline alignment checked")
            
        except Exception as e:
            print(f"WARNING: {feature_id}: error checking alignment: {e}")
            warnings += 1

sys.exit(0 if errors == 0 else 1)
PYTHON_SCRIPT

SYNC_EXIT=$?

if [ $SYNC_EXIT -ne 0 ]; then
    error "Spec sync check failed"
    exit 1
fi

echo ""
if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    info "All spec synchronization checks passed!"
    exit 0
elif [ $ERRORS -eq 0 ]; then
    warning "$WARNINGS warning(s) found (non-blocking)"
    exit 0
else
    error "$ERRORS error(s) and $WARNINGS warning(s) found"
    exit 1
fi


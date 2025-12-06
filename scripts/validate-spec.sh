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
# validate-spec.sh - Validates spec/features.yaml and related files
#
# This script ensures:
# - spec/features.yaml is valid YAML
# - All referenced spec files exist
# - All referenced test files exist (or are planned)
# - Features marked as 'done' have associated tests

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
NC='\033[0m' # No Color

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

# Check if Python 3 is available (for YAML parsing)
if ! command -v python3 &> /dev/null; then
    error "python3 is required for YAML validation"
    exit 1
fi

# Validate YAML syntax
info "Validating YAML syntax..."
if ! python3 -c "import yaml; yaml.safe_load(open('spec/features.yaml'))" 2>/dev/null; then
    error "spec/features.yaml is not valid YAML"
    exit 1
fi

# Parse features.yaml and validate references
info "Validating feature references..."

python3 << 'PYTHON_SCRIPT'
import yaml
import os
import sys

errors = 0
warnings = 0

with open('spec/features.yaml', 'r') as f:
    data = yaml.safe_load(f)

if 'features' not in data:
    print("ERROR: 'features' key not found in spec/features.yaml")
    sys.exit(1)

for feature in data['features']:
    feature_id = feature.get('id', 'UNKNOWN')
    status = feature.get('status', 'unknown')
    spec_path = feature.get('spec', '')
    tests = feature.get('tests', [])

    # Check spec file exists
    # Only require spec files for features that are done or wip, not todo
    if spec_path:
        spec_full_paths = [
            f"spec/{spec_path}",
            spec_path,
            f"docs/{spec_path}"
        ]
        spec_exists = any(os.path.exists(p) for p in spec_full_paths)
        if not spec_exists:
            if status in ['done', 'wip']:
                print(f"ERROR: Feature {feature_id}: spec file not found: {spec_path} (required for {status} features)")
                errors += 1
            else:
                print(f"WARNING: Feature {feature_id}: spec file not found: {spec_path} (optional for {status} features)")
                warnings += 1
        else:
            found_path = next(p for p in spec_full_paths if os.path.exists(p))
            print(f"✓ Feature {feature_id}: spec found at {found_path}")

    # Check test files for 'done' features
    if status == 'done' and not tests:
        print(f"WARNING: Feature {feature_id} is marked 'done' but has no test files listed")
        warnings += 1

    # Check that test files exist (if listed)
    for test_file in tests:
        if not os.path.exists(test_file):
            # This is a warning, not an error, as tests might be planned
            print(f"WARNING: Feature {feature_id}: test file not found: {test_file}")
            warnings += 1
        else:
            print(f"✓ Feature {feature_id}: test file exists: {test_file}")

# Cross-checks: orphan specs, dangling features, domain-to-path mapping
print("\n=== Cross-checks ===")

# Find all spec files
spec_files = set()
for root, dirs, files in os.walk('spec'):
    # Skip hidden directories
    dirs[:] = [d for d in dirs if not d.startswith('.')]
    for file in files:
        if file.endswith('.md'):
            full_path = os.path.join(root, file)
            # Normalize path (remove spec/ prefix)
            rel_path = os.path.relpath(full_path, 'spec')
            spec_files.add(rel_path)

# Check for orphan specs (spec exists with no feature entry)
feature_spec_paths = set()
for feature in data['features']:
    spec_path = feature.get('spec', '')
    if spec_path:
        # Normalize path
        normalized = spec_path.replace('spec/', '').replace('docs/', '')
        feature_spec_paths.add(normalized)

orphan_specs = spec_files - feature_spec_paths
if orphan_specs:
    for orphan in sorted(orphan_specs):
        print(f"WARNING: Orphan spec file (no feature entry): spec/{orphan}")
        warnings += 1

# Check for dangling features (feature entry with no spec)
dangling_features = []
for feature in data['features']:
    feature_id = feature.get('id', 'UNKNOWN')
    spec_path = feature.get('spec', '')
    if spec_path:
        normalized = spec_path.replace('spec/', '').replace('docs/', '')
        spec_full_paths = [
            f"spec/{normalized}",
            normalized,
            f"docs/{normalized}"
        ]
        if not any(os.path.exists(p) for p in spec_full_paths):
            if feature.get('status') in ['done', 'wip']:
                print(f"ERROR: Dangling feature (spec missing): {feature_id} -> {spec_path}")
                errors += 1
            else:
                print(f"WARNING: Dangling feature (spec missing): {feature_id} -> {spec_path}")
                warnings += 1

# Check domain-to-path mapping consistency
for feature in data['features']:
    feature_id = feature.get('id', 'UNKNOWN')
    spec_path = feature.get('spec', '')
    if not spec_path:
        continue
    
    # Extract domain from feature ID prefix (CLI_, CORE_, PROVIDER_, etc.)
    if feature_id.startswith('CLI_'):
        expected_domain = 'commands'
    elif feature_id.startswith('CORE_'):
        expected_domain = 'core'
    elif feature_id.startswith('PROVIDER_'):
        # Provider features can be in providers/<type>/
        expected_domain = 'providers'
    else:
        continue
    
    # Check if spec path matches expected domain
    if expected_domain == 'commands' and 'commands/' not in spec_path:
        print(f"WARNING: {feature_id}: CLI feature spec should be in commands/ domain, found: {spec_path}")
        warnings += 1
    elif expected_domain == 'core' and 'core/' not in spec_path:
        print(f"WARNING: {feature_id}: CORE feature spec should be in core/ domain, found: {spec_path}")
        warnings += 1
    elif expected_domain == 'providers' and 'providers/' not in spec_path:
        print(f"WARNING: {feature_id}: PROVIDER feature spec should be in providers/ domain, found: {spec_path}")
        warnings += 1

# Check for missing exit code definitions in command specs
for feature in data['features']:
    feature_id = feature.get('id', 'UNKNOWN')
    spec_path = feature.get('spec', '')
    status = feature.get('status', 'unknown')
    
    if not spec_path or status not in ['done', 'wip']:
        continue
    
    # Check if it's a CLI command
    if feature_id.startswith('CLI_'):
        spec_full_paths = [
            f"spec/{spec_path}",
            spec_path,
            f"docs/{spec_path}"
        ]
        spec_file = next((p for p in spec_full_paths if os.path.exists(p)), None)
        
        if spec_file:
            try:
                with open(spec_file, 'r') as f:
                    spec_content = f.read()
                
                # Check for exit code section
                if 'exit' in spec_content.lower() and 'code' in spec_content.lower():
                    # Check for specific exit code definitions
                    import re
                    exit_code_patterns = [
                        r'exit.*code.*0',
                        r'exit.*code.*1',
                        r'code.*0.*success',
                        r'code.*1.*error'
                    ]
                    has_exit_codes = any(re.search(p, spec_content, re.IGNORECASE) for p in exit_code_patterns)
                    if not has_exit_codes:
                        print(f"WARNING: {feature_id}: CLI command spec should define exit codes")
                        warnings += 1
            except Exception:
                pass

sys.exit(0 if errors == 0 else 1)
PYTHON_SCRIPT

VALIDATION_EXIT=$?

if [ $VALIDATION_EXIT -ne 0 ]; then
    error "Feature validation failed"
    exit 1
fi

# Summary
echo ""
if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    info "All validations passed!"
    exit 0
elif [ $ERRORS -eq 0 ]; then
    warning "$WARNINGS warning(s) found (non-blocking)"
    exit 0
else
    error "$ERRORS error(s) and $WARNINGS warning(s) found"
    exit 1
fi


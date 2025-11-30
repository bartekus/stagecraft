#!/bin/bash
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
    if spec_path:
        spec_full_paths = [
            f"spec/{spec_path}",
            spec_path,
            f"docs/{spec_path}"
        ]
        spec_exists = any(os.path.exists(p) for p in spec_full_paths)
        if not spec_exists:
            print(f"ERROR: Feature {feature_id}: spec file not found: {spec_path}")
            errors += 1
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


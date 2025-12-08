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
# new-feature.sh - Generate feature skeleton from templates
#
# Usage: ./scripts/new-feature.sh <FEATURE_ID> <DOMAIN> [feature-name]
#
# Example: ./scripts/new-feature.sh CLI_DEPLOY commands deploy
# Example: ./scripts/new-feature.sh PROVIDER_BACKEND_NEW backend new-backend-provider

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

# Validate arguments
if [ $# -lt 2 ]; then
    echo "Error: Missing required arguments"
    echo ""
    echo "Usage: $0 <FEATURE_ID> <DOMAIN> [feature-name]"
    echo ""
    echo "Arguments:"
    echo "  FEATURE_ID    - Feature ID in SCREAMING_SNAKE_CASE (e.g., CLI_DEPLOY)"
    echo "  DOMAIN        - Domain directory (e.g., commands, core, providers/backend)"
    echo "  feature-name  - Optional: kebab-case name for spec file (defaults to FEATURE_ID converted to kebab-case)"
    echo ""
    echo "Examples:"
    echo "  $0 CLI_DEPLOY commands deploy"
    echo "  $0 PROVIDER_BACKEND_NEW providers/backend new-backend-provider"
    echo "  $0 CORE_STATE core state"
    exit 1
fi

FEATURE_ID="$1"
DOMAIN="$2"
FEATURE_NAME="${3:-$(echo "$FEATURE_ID" | tr '[:upper:]' '[:lower:]' | tr '_' '-')}"

# Validate FEATURE_ID format (should be SCREAMING_SNAKE_CASE)
if ! echo "$FEATURE_ID" | grep -qE '^[A-Z][A-Z0-9_]*$'; then
    echo "Error: FEATURE_ID must be in SCREAMING_SNAKE_CASE (e.g., CLI_DEPLOY, CORE_STATE)"
    echo "       Got: $FEATURE_ID"
    exit 1
fi

# Validate domain exists
DOMAIN_DIR="spec/$DOMAIN"
if [ ! -d "$DOMAIN_DIR" ]; then
    echo "Error: Domain directory does not exist: $DOMAIN_DIR"
    echo ""
    echo "Available domains:"
    find spec -type d -mindepth 1 -maxdepth 1 | sed 's|spec/|  - |' | sort
    echo ""
    echo "For provider domains, use: providers/<provider-type>"
    exit 1
fi

# Check if files already exist
ANALYSIS_FILE="docs/engine/analysis/${FEATURE_ID}.md"
OUTLINE_FILE="docs/engine/outlines/${FEATURE_ID}_IMPLEMENTATION_OUTLINE.md"
SPEC_FILE="${DOMAIN_DIR}/${FEATURE_NAME}.md"

if [ -f "$ANALYSIS_FILE" ] || [ -f "$OUTLINE_FILE" ] || [ -f "$SPEC_FILE" ]; then
    echo "Warning: One or more files already exist:"
    [ -f "$ANALYSIS_FILE" ] && echo "  - $ANALYSIS_FILE"
    [ -f "$OUTLINE_FILE" ] && echo "  - $OUTLINE_FILE"
    [ -f "$SPEC_FILE" ] && echo "  - $SPEC_FILE"
    echo ""
    read -p "Overwrite existing files? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborted."
        exit 1
    fi
fi

# Ensure analysis directory exists
mkdir -p "$(dirname "$ANALYSIS_FILE")"

# Create analysis brief from template
echo "Creating analysis brief: $ANALYSIS_FILE"
sed "s/<FEATURE_ID>/$FEATURE_ID/g" docs/engine/analysis/TEMPLATE.md > "$ANALYSIS_FILE"

# Create implementation outline from template
echo "Creating implementation outline: $OUTLINE_FILE"
sed "s/<FEATURE_ID>/$FEATURE_ID/g" docs/engine/outlines/IMPLEMENTATION_OUTLINE_TEMPLATE.md | \
    sed "s|<domain>|$DOMAIN|g" | \
    sed "s|<feature>|$FEATURE_NAME|g" > "$OUTLINE_FILE"

# Create spec file with header
echo "Creating spec file: $SPEC_FILE"
cat > "$SPEC_FILE" <<EOF
# ${FEATURE_ID}

- **Feature ID**: \`${FEATURE_ID}\`
- **Domain**: \`${DOMAIN}\`
- **Status**: \`todo\`
- **Related features**:
  - (list dependencies here)

---

## 1. Purpose

(Describe the purpose of this feature)

---

## 2. Behavior

(Define the exact behavior this feature will implement)

---

## 3. CLI or API Contract

(If applicable, define the CLI command or API surface)

---

## 4. Inputs and Outputs

(Define what inputs are required and what outputs are produced)

---

## 5. Error Handling

(Define error conditions and exit codes)

---

## 6. Determinism Requirements

(Define deterministic behavior guarantees)

---

## 7. Testing Requirements

(Define what tests are required)

---

## 8. Implementation Notes

(Any implementation-specific notes or constraints)

EOF

# Add entry to spec/features.yaml if it doesn't exist
FEATURES_YAML="spec/features.yaml"
SPEC_PATH="${DOMAIN}/${FEATURE_NAME}.md"

# Check if entry already exists
if grep -q "id: ${FEATURE_ID}" "$FEATURES_YAML"; then
    echo "Warning: Feature ID ${FEATURE_ID} already exists in $FEATURES_YAML"
    echo "Skipping automatic entry creation."
    FEATURES_ENTRY_ADDED=false
else
    # Find the last feature entry to determine insertion point
    # We'll append to the end of the file, before any trailing comments
    echo "" >> "$FEATURES_YAML"
    echo "  - id: ${FEATURE_ID}" >> "$FEATURES_YAML"
    echo "    title: \"TODO: Add feature title\"" >> "$FEATURES_YAML"
    echo "    status: todo" >> "$FEATURES_YAML"
    echo "    spec: \"${SPEC_PATH}\"" >> "$FEATURES_YAML"
    echo "    owner: TODO" >> "$FEATURES_YAML"
    echo "    tests: []" >> "$FEATURES_YAML"
    FEATURES_ENTRY_ADDED=true
fi

echo ""
echo "✓ Feature skeleton created successfully!"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "FEATURE SCAFFOLDING SUMMARY"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Feature ID:        ${FEATURE_ID}"
echo "Domain:            ${DOMAIN}"
echo "Feature Name:      ${FEATURE_NAME}"
echo ""
echo "Analysis Brief:    ${ANALYSIS_FILE}"
echo "Implementation:    ${OUTLINE_FILE}"
echo "Spec:              ${SPEC_FILE}"
if [ "$FEATURES_ENTRY_ADDED" = true ]; then
    echo "Lifecycle Entry:    ${FEATURES_YAML} ✓ (auto-added)"
else
    echo "Lifecycle Entry:    ${FEATURES_YAML} (already exists)"
fi
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Next steps:"
echo "  1. Fill in the analysis brief: ${ANALYSIS_FILE}"
echo "  2. Fill in the implementation outline: ${OUTLINE_FILE}"
echo "  3. Fill in the spec: ${SPEC_FILE}"
if [ "$FEATURES_ENTRY_ADDED" = true ]; then
    echo "  4. Update feature entry in ${FEATURES_YAML} (title, owner, dependencies)"
else
    echo "  4. Verify feature entry in ${FEATURES_YAML}"
fi
echo "  5. Begin implementation following Feature Planning Protocol"
echo ""
echo "See Agent.md 'Feature Planning Protocol' section for the complete workflow."


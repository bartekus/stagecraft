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
# generate-evolution-log.sh - Scaffold a new provider evolution log
#
# Usage:
#   ./scripts/generate-evolution-log.sh <FEATURE_ID>
#
# Example:
#   ./scripts/generate-evolution-log.sh PROVIDER_NETWORK_TAILSCALE
#
# This script:
# - Creates docs/engine/history/<FEATURE_ID>_EVOLUTION.md
# - Populates it with standard header and migration checklist
# - Provides guidance on next steps

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

if [ $# -eq 0 ]; then
    echo "Usage: $0 <FEATURE_ID>"
    echo ""
    echo "Example: $0 PROVIDER_NETWORK_TAILSCALE"
    exit 1
fi

FEATURE_ID="$1"
EVOLUTION_FILE="docs/engine/history/${FEATURE_ID}_EVOLUTION.md"

# Check if file already exists
if [ -f "$EVOLUTION_FILE" ]; then
    echo "Error: $EVOLUTION_FILE already exists"
    exit 1
fi

# Extract provider type and name from FEATURE_ID
# Pattern: PROVIDER_<TYPE>_<NAME>
# Fallback to neutral template if pattern doesn't match
PROVIDER_TYPE=""
PROVIDER_NAME=""
if [[ "$FEATURE_ID" =~ ^PROVIDER_([A-Z]+)_(.+)$ ]]; then
    PROVIDER_TYPE="${BASH_REMATCH[1]}"
    PROVIDER_NAME="${BASH_REMATCH[2]}"
    # Lowercase using portable tools (macOS ships bash 3.2)
    PROVIDER_TYPE_LC=$(printf '%s' "$PROVIDER_TYPE" | tr '[:upper:]' '[:lower:]')
    PROVIDER_NAME_LC=$(printf '%s' "$PROVIDER_NAME" | tr '[:upper:]' '[:lower:]')
    PROVIDER_DESCRIPTION="${PROVIDER_TYPE_LC} ${PROVIDER_NAME_LC} provider"
else
    # Fallback: use neutral description
    PROVIDER_DESCRIPTION="feature"
fi

# Try to find spec file (only if we detected provider type)
SPEC_FILE=""
if [ -n "$PROVIDER_TYPE" ] && [ -n "$PROVIDER_NAME" ]; then
    PROVIDER_TYPE_LC=$(printf '%s' "$PROVIDER_TYPE" | tr '[:upper:]' '[:lower:]')
    PROVIDER_NAME_LC=$(printf '%s' "$PROVIDER_NAME" | tr '[:upper:]' '[:lower:]')

    if [ -f "spec/providers/${PROVIDER_TYPE_LC}/${PROVIDER_NAME_LC}.md" ]; then
        SPEC_FILE="spec/providers/${PROVIDER_TYPE_LC}/${PROVIDER_NAME_LC}.md"
    elif [ -f "spec/providers/${PROVIDER_TYPE_LC}/${PROVIDER_NAME}.md" ]; then
        SPEC_FILE="spec/providers/${PROVIDER_TYPE_LC}/${PROVIDER_NAME}.md"
    fi
fi
if [ -z "$SPEC_FILE" ]; then
    # Try generic spec location
    FEATURE_ID_LC=$(printf '%s' "$FEATURE_ID" | tr '[:upper:]' '[:lower:]')
    if [ -f "spec/${FEATURE_ID_LC}.md" ]; then
        SPEC_FILE="spec/${FEATURE_ID_LC}.md"
    else
        SPEC_FILE="spec/... (update with actual path)"
    fi
fi

# Try to find analysis file
ANALYSIS_FILE=""
if [ -f "docs/engine/analysis/${FEATURE_ID}.md" ]; then
    ANALYSIS_FILE="docs/engine/analysis/${FEATURE_ID}.md"
else
    ANALYSIS_FILE="docs/engine/analysis/${FEATURE_ID}.md (if exists)"
fi

# Try to find outline file
OUTLINE_FILE=""
if [ -f "docs/engine/outlines/${FEATURE_ID}_IMPLEMENTATION_OUTLINE.md" ]; then
    OUTLINE_FILE="docs/engine/outlines/${FEATURE_ID}_IMPLEMENTATION_OUTLINE.md"
else
    OUTLINE_FILE="docs/engine/outlines/${FEATURE_ID}_IMPLEMENTATION_OUTLINE.md (if exists)"
fi

# Generate the evolution log
cat > "$EVOLUTION_FILE" <<EOF
# ${FEATURE_ID} Evolution Log

> Canonical evolution history for the ${PROVIDER_DESCRIPTION}.
> This document replaces per slice plans, readiness docs, checklists, and ad hoc notes.

## 1. Purpose and Scope

This document captures the end to end evolution of \`${FEATURE_ID}\`:

- Design intent and constraints
- Slice plans and execution notes (if applicable)
- Coverage movement over time
- Governance and spec changes
- Open questions and deferred work

It consolidates content that previously lived in:

- Coverage plans and status documents
- Slice plans, checklists, and readiness docs (if applicable)
- PR descriptions and coverage notes
- Any other ${FEATURE_ID} specific coverage or governance notes

All future ${FEATURE_ID} evolution notes should be added here instead of creating new standalone docs.

---

## 2. Feature References

- **Feature ID:** \`${FEATURE_ID}\`
- **Spec:** \`${SPEC_FILE}\`
- **Core analysis:** \`${ANALYSIS_FILE}\`
- **Implementation outline:** \`${OUTLINE_FILE}\`
- **Status:** see \`docs/engine/status/PROVIDER_COVERAGE_STATUS.md\` and \`docs/coverage/COVERAGE_LEDGER.md\`

---

## 3. Design Intent and Constraints

> Short summary of why this provider exists and the constraints it must respect.
> Migrate the high level intent from the analysis and governance docs here.

- **Purpose**: [Describe the provider's purpose]

- **Primary responsibilities**:
  - [Responsibility 1]
  - [Responsibility 2]
  - [Responsibility 3]

- **Non goals**:
  - [Non-goal 1]
  - [Non-goal 2]

- **Determinism constraints**:
  - [Constraint 1]
  - [Constraint 2]

- **Provider boundary rules**:
  - [Rule 1]
  - [Rule 2]

- **External dependencies**:
  - [Dependency 1]
  - [Dependency 2]

---

## 4. Coverage Timeline Overview

> High level table that shows each phase/slice and its role.

| Phase/Slice | Status        | Focus                             | Coverage before | Coverage after | Date range        | Notes |
|-------------|---------------|-----------------------------------|-----------------|----------------|-------------------|-------|
| Initial     | complete      | Initial implementation             | XX.X%           | XX.X%          | 2025-XX-XX        | ...   |
| ...         | ...           | ...                               | ...             | ...            | ...                | ...   |

---

## 5. Coverage Evolution Summary

> High level history that will match or cross reference \`COVERAGE_LEDGER.md\`.

| Date       | Change source                | Coverage before | Coverage after | Notes                                  |
|------------|-----------------------------|-----------------|----------------|----------------------------------------|
| 2025-XX-XX | Initial implementation      | XX.X%           | XX.X%          | ...                                    |
| ...        | ...                         | ...             | ...            | ...                                    |

---

## 6. Governance and Spec Adjustments

> Capture any spec and governance changes specific to this provider.

- **Spec version changes** (with brief notes):
  - [Change 1]
  - [Change 2]

- **Breaking or behavioural changes**: [None / List changes]

- **Governance decisions that impact this provider**:
  - [Decision 1]
  - [Decision 2]

- **Links to relevant ADRs**: [None / List ADRs]

---

## 7. Open Questions and Future Work

> Reserved for post v1 or follow up work.

- **Potential future improvements**:
  - [Improvement 1]
  - [Improvement 2]

- **Known tradeoffs that might be revisited**:
  - [Tradeoff 1]
  - [Tradeoff 2]

---

## 8. Migration Notes

> Use this section temporarily while consolidating existing docs into this log.

- [ ] Migrated coverage plan content
- [ ] Migrated slice plans (if applicable)
- [ ] Migrated checklists and readiness docs (if applicable)
- [ ] Migrated PR descriptions and coverage notes
- [ ] Updated cross-references in other docs

Once migration is complete this checklist can be removed or marked as complete.
EOF

echo "Created: $EVOLUTION_FILE"
echo ""
echo "⚠️  Review and fill: Feature references (section 2), Design intent and constraints (section 3), Coverage timeline (section 4)"
echo ""
echo "Next steps:"
echo "  1. Review and populate sections 3-7 with actual content"
echo "  2. Migrate content from legacy docs (use section 8 checklist)"
echo "  3. Mark legacy docs as superseded"
echo "  4. Update cross-references in COVERAGE_LEDGER.md and GOVERNANCE_ALMANAC.md"
echo ""
echo "See existing evolution logs for examples:"
echo "  - docs/engine/history/PROVIDER_NETWORK_TAILSCALE_EVOLUTION.md"
echo "  - docs/engine/history/PROVIDER_FRONTEND_GENERIC_EVOLUTION.md"

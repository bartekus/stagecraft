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
# append-coverage-snapshot.sh - Append coverage snapshot to COVERAGE_LEDGER.md (append-only)
#
# Usage:
#   ./scripts/append-coverage-snapshot.sh [--event "Event description"] [--overall XX.X] [--core XX.X] [--providers XX.X]
#
# This script:
# - Appends a new row to the Historical Coverage Timeline in COVERAGE_LEDGER.md
# - Never rewrites or edits existing entries (append-only)
# - Uses current date and coverage data from go test -cover (if not provided)
# - Ensures deterministic formatting

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

LEDGER_FILE="docs/coverage/COVERAGE_LEDGER.md"

# Parse arguments
EVENT_DESC=""
OVERALL_COV=""
CORE_COV=""
PROVIDERS_COV=""
ALLOW_DUPLICATE=0

while [[ $# -gt 0 ]]; do
    case $1 in
        --event)
            EVENT_DESC="$2"
            shift 2
            ;;
        --overall)
            OVERALL_COV="$2"
            shift 2
            ;;
        --core)
            CORE_COV="$2"
            shift 2
            ;;
        --providers)
            PROVIDERS_COV="$2"
            shift 2
            ;;
        --allow-duplicate)
            ALLOW_DUPLICATE=1
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--event \"Event description\"] [--overall XX.X] [--core XX.X] [--providers XX.X] [--allow-duplicate]"
            exit 1
            ;;
    esac
done

# Get current date in YYYY-MM-DD format
CURRENT_DATE=$(date +%Y-%m-%d)

# If coverage values not provided, try to get them from go test
if [ -z "$OVERALL_COV" ] || [ -z "$CORE_COV" ] || [ -z "$PROVIDERS_COV" ]; then
    echo "Running coverage analysis..."
    
    # Run tests with coverage
    go test ./... -coverprofile=coverage.out >/dev/null 2>&1 || {
        echo "Warning: Failed to generate coverage profile. Using placeholder values."
        OVERALL_COV="${OVERALL_COV:-...}"
        CORE_COV="${CORE_COV:-...}"
        PROVIDERS_COV="${PROVIDERS_COV:-...}"
    }
    
    if [ -f coverage.out ]; then
        # Parse coverage from go tool cover output
        COVERAGE_OUTPUT=$(go tool cover -func=coverage.out 2>/dev/null | tail -1)
        
        if [ -z "$OVERALL_COV" ]; then
            OVERALL_COV=$(echo "$COVERAGE_OUTPUT" | awk '{print $3}' | sed 's/%//')
        fi
        
        # Calculate core coverage (pkg/config + internal/core)
        if [ -z "$CORE_COV" ]; then
            CORE_CONFIG=$(go tool cover -func=coverage.out 2>/dev/null | grep "stagecraft/pkg/config" | awk '{print $3}' | sed 's/%//')
            CORE_INTERNAL=$(go tool cover -func=coverage.out 2>/dev/null | grep "stagecraft/internal/core/" | awk '{print $3}' | sed 's/%//')
            if [ -n "$CORE_CONFIG" ] && [ -n "$CORE_INTERNAL" ]; then
                # Simple average (could be improved with weighted average)
                CORE_COV=$(awk -v a="$CORE_CONFIG" -v b="$CORE_INTERNAL" 'BEGIN { printf "%.1f", (a+b)/2 }')
            else
                CORE_COV="..."
            fi
        fi
        
        # Calculate providers coverage (average of all provider packages)
        if [ -z "$PROVIDERS_COV" ]; then
            PROVIDER_COVERAGES=$(go tool cover -func=coverage.out 2>/dev/null | grep "stagecraft/internal/providers" | awk '{print $3}' | sed 's/%//')
            if [ -n "$PROVIDER_COVERAGES" ]; then
                PROVIDERS_COV=$(echo "$PROVIDER_COVERAGES" | awk '{sum+=$1; count++} END {if(count>0) printf "%.1f", sum/count; else print "..."}')
            else
                PROVIDERS_COV="..."
            fi
        fi
        
        rm -f coverage.out
    fi
fi

# Default event description if not provided
if [ -z "$EVENT_DESC" ]; then
    EVENT_DESC="Coverage snapshot"
fi

# Format coverage values (ensure they have % suffix for display)
format_coverage() {
    local val="$1"
    if [ "$val" != "..." ] && [[ ! "$val" =~ %$ ]]; then
        echo "${val}%"
    else
        echo "$val"
    fi
}

OVERALL_DISPLAY=$(format_coverage "$OVERALL_COV")
CORE_DISPLAY=$(format_coverage "$CORE_COV")
PROVIDERS_DISPLAY=$(format_coverage "$PROVIDERS_COV")

# Find the Historical Coverage Timeline section
TIMELINE_SECTION="## 3. Historical Coverage Timeline"
if ! grep -q "$TIMELINE_SECTION" "$LEDGER_FILE"; then
    echo "Error: Could not find '$TIMELINE_SECTION' section in $LEDGER_FILE"
    echo "Expected heading: $TIMELINE_SECTION"
    exit 1
fi

# Find the line number of the table header (portable: use awk for line numbers)
TABLE_HEADER="^| Date       | Event / Source"
TABLE_START=$(awk "/$TABLE_HEADER/ {print NR; exit}" "$LEDGER_FILE")

if [ -z "$TABLE_START" ] || [ "$TABLE_START" -eq 0 ]; then
    echo "Error: Could not find coverage timeline table in $LEDGER_FILE"
    echo "Expected table header: | Date       | Event / Source"
    echo "Found section heading: $TIMELINE_SECTION"
    exit 1
fi

# Check for duplicate (same date + event) unless --allow-duplicate is set
if [ "$ALLOW_DUPLICATE" -eq 0 ]; then
    # Refuse duplicates (same date + event) unless explicitly allowed.
    if awk -v start="$TABLE_START" -v date="$CURRENT_DATE" -v event="$EVENT_DESC" '
        BEGIN {FS="|"}
        NR > start && /^\| [0-9]{4}-[0-9]{2}-[0-9]{2} / {
            d=$2; e=$3
            gsub(/^[ \t]+|[ \t]+$/, "", d)
            gsub(/^[ \t]+|[ \t]+$/, "", e)
            if (d == date && e == event) { exit 1 }
        }
    ' "$LEDGER_FILE"; then
        :
    else
        echo "Error: Duplicate entry detected: date=$CURRENT_DATE, event=$EVENT_DESC"
        echo "Use --allow-duplicate to override this check"
        exit 1
    fi
fi

# Find the last row of the table (before the closing ---)
# Use awk for portability (works on both BSD and GNU)
TABLE_END=$(awk -v start="$TABLE_START" '
    NR >= start && /^---/ {print NR; exit}
' "$LEDGER_FILE")

if [ -n "$TABLE_END" ] && [ "$TABLE_END" -gt 0 ]; then
    INSERT_LINE=$TABLE_END
else
    # If no closing --- found, insert before the next section
    NEXT_SECTION=$(awk -v start="$TABLE_START" '
        NR > start && /^## / {print NR; exit}
    ' "$LEDGER_FILE")
    
    if [ -n "$NEXT_SECTION" ] && [ "$NEXT_SECTION" -gt 0 ]; then
        INSERT_LINE=$NEXT_SECTION
    else
        # Fallback: append to end of file
        INSERT_LINE=$(wc -l < "$LEDGER_FILE" | tr -d ' ')
        INSERT_LINE=$((INSERT_LINE + 1))
    fi
fi

# Create the new row
NEW_ROW="| $CURRENT_DATE | $EVENT_DESC | $OVERALL_DISPLAY | $CORE_DISPLAY | $PROVIDERS_DISPLAY | |"

# Insert the new row (append-only: insert before the closing --- or next section)
# Use awk for better portability (works on macOS and Linux)
awk -v line="$INSERT_LINE" -v new_row="$NEW_ROW" '
    NR == line {print new_row}
    {print}
' "$LEDGER_FILE" > "${LEDGER_FILE}.tmp" && mv "${LEDGER_FILE}.tmp" "$LEDGER_FILE"

echo "✓ Appended coverage snapshot to $LEDGER_FILE"
echo "  Date: $CURRENT_DATE"
echo "  Event: $EVENT_DESC"
echo "  Overall: $OVERALL_DISPLAY"
echo "  Core: $CORE_DISPLAY"
echo "  Providers: $PROVIDERS_DISPLAY"
echo ""
echo "Note: This is an append-only operation. Historical entries are preserved."

#!/bin/bash
# SPDX-License-Identifier: AGPL-3.0-or-later

# check-purity.sh
# Enforces dependency boundaries between Stagecraft Core and Cortex.
#
# Rule A: ai.agent/cortex/... must NEVER import stagecraft/internal/...
# Rule B: cmd/stagecraft/... must NEVER import ai.agent/...
# Rule C: cmd/stagecraft/... must NEVER import tools/... (context-compiler)

set -e

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FAILURES=0

echo "Running Purity Check..."

# Check Rule A: Cortex Library Purity
# Exclude test files if needed, but ideally tests should also be pure or use external_test package
echo "Checking Rule A: Cortex Library Purity..."
if grep -r --include="*.go" "stagecraft/internal/" "$REPO_ROOT/ai.agent/cortex"; then
  echo "ERROR: Cortex library imports stagecraft/internal/ (Rule A Violation)"
  FAILURES=$((FAILURES + 1))
else
  echo "PASS: Rule A"
fi

# Check Rule B: Stagecraft CLI Purity (vs Cortex)
echo "Checking Rule B: Stagecraft CLI Purity (vs Cortex)..."
if grep -r "stagecraft/ai.agent/" "$REPO_ROOT/cmd/stagecraft" "$REPO_ROOT/internal/cli"; then
    # We must exclude the case where we might have legacy comments or such, but grep "stagecraft/ai.agent/" finds string literals too?
    # Better to look for imports. grep -r is crude but effective if imports are standard.
    # Note: internal/cli/root.go previously imported commands which import moved files.
    # If imports are cleaned up, this should be clean.
    echo "ERROR: Stagecraft CLI imports ai.agent/ (Rule B Violation)"
    FAILURES=$((FAILURES + 1))
else
    echo "PASS: Rule B"
fi

# Check Rule C: Stagecraft CLI Purity (vs Tools)
echo "Checking Rule C: Stagecraft CLI Purity (vs Tools)..."
if grep -r "stagecraft/tools/" "$REPO_ROOT/cmd/stagecraft" "$REPO_ROOT/internal/cli"; then
    echo "ERROR: Stagecraft CLI imports tools/ (Rule C Violation)"
    FAILURES=$((FAILURES + 1))
else
    echo "PASS: Rule C"
fi

if [ "$FAILURES" -gt 0 ]; then
  echo "Purity Check FAILED with $FAILURES violation(s)."
  exit 1
fi

echo "Purity Check PASSED."

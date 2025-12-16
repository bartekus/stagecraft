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

# Stagecraft governance pre-commit wrapper
# Runs a small set of structural and governance checks before committing.
# This is not a replacement for CI, only a fast local sanity pass.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "$ROOT_DIR"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Stagecraft - Governance Pre Commit"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo

# 1. Feature mapping and spec alignment
echo "[1/4] Checking feature mapping and governance invariants..."
if [[ -x ./bin/stagecraft ]]; then
  go run ./ai.agent/cmd/cortex gov feature-mapping
else
  echo "stagecraft binary not found, building..."
  go build -o ./bin/stagecraft ./cmd/stagecraft
  go run ./ai.agent/cmd/cortex gov feature-mapping
fi
echo

# 2. Orphan spec check
echo "[2/4] Checking for orphan spec files..."
if [[ -x ./scripts/check-orphan-specs.sh ]]; then
  ./scripts/check-orphan-specs.sh
else
  echo "WARN: ./scripts/check-orphan-specs.sh not found or not executable"
fi
echo

# 3. Core coverage guardrail (config + core)
echo "[3/5] Running core coverage guardrail (pkg/config + internal/core)..."
go test -cover ./pkg/config ./internal/core
echo

# 4. Documentation pattern checks
echo "[4/5] Checking for forbidden documentation patterns..."
if [[ -x ./scripts/check-doc-patterns.sh ]]; then
  ./scripts/check-doc-patterns.sh || {
    echo "✗ Documentation pattern checks failed"
    echo "✗ Governance pre-commit checks failed"
    exit 1
  }
else
  echo "WARN: ./scripts/check-doc-patterns.sh not found or not executable"
fi
echo

# 5. Full checks (optional but recommended)
if [[ "${GOV_FAST:-0}" == "1" ]]; then
  echo "[5/5] GOV_FAST=1 set - skipping full run-all-checks.sh"
  echo "      Remember to run ./scripts/run-all-checks.sh before merging."
else
  echo "[5/5] Running full project checks..."
  if [[ -x ./scripts/run-all-checks.sh ]]; then
    ./scripts/run-all-checks.sh
  else
    echo "WARN: ./scripts/run-all-checks.sh not found or not executable"
    echo "      Fallback: running go test ./..."
    go test ./...
  fi
fi

echo
echo "✅ Governance pre commit checks completed successfully."
echo "You are clear to commit, subject to normal review."

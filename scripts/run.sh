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

set -euo pipefail

# Pass-through wrapper to the canonical cortex invocation.
# Usage examples:
#   scripts/run.sh
#   scripts/run.sh --state-dir .cortex-state
#   scripts/run.sh --fail-fast
#   scripts/run.sh resume   (if you support it in cortex; otherwise just pass it through)

# Ensure Cortex is built
if [ ! -f "bin/cortex" ]; then
    echo "Building Cortex..."
    if [ -d "cmd/cortex" ]; then
        go build -o bin/cortex ./cmd/cortex
    elif [ -d "../cortex" ]; then
        echo "Found sibling cortex repo..."
        (cd ../cortex && go build -o ../stagecraft/bin/cortex ./cmd/cortex)
    else
        echo "Error: Cannot find cortex source in ./cmd/cortex or ../cortex"
        echo "Please install cortex manually to bin/cortex"
        exit 1
    fi
fi

if [ $# -eq 0 ]; then
    exec ./bin/cortex run all
else
    exec ./bin/cortex run "$@"
fi

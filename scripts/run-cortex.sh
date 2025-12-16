#!/bin/bash
# SPDX-License-Identifier: AGPL-3.0-or-later

# run-cortex.sh
# Single entrypoint for invoking Cortex from Stagecraft.
# Relies on Go module resolution (go.work locally, or go.mod/go install in CI).

set -euo pipefail

# Try to run Cortex using the module path.
# 1. If go.work is present, Go matches the module path to the local directory.
# 2. If ai.agent exists (legacy), use it as fallback to keep CI green before deletion/publish.
# 3. Otherwise, fetch from remote.

if [ -f "go.work" ]; then
    exec go run github.com/bartekus/cortex/cmd/cortex "$@"
fi

if [ -d "ai.agent/cmd/cortex" ]; then
    # Fallback for CI/transition until Cortex is published and ai.agent deleted
    exec go run ./ai.agent/cmd/cortex "$@"
fi

if ! go run github.com/bartekus/cortex/cmd/cortex "$@"; then
    echo ""
    echo "Error: Failed to run Cortex." >&2
    echo "Ensure you have a valid go.work pointing to local Cortex repo," >&2
    echo "OR that github.com/bartekus/cortex is available." >&2
    exit 1
fi

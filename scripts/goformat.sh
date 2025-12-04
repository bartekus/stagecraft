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
# goformat.sh - Canonical Go formatting command using gofumpt
# This is the single source of truth for Go formatting in Stagecraft.

set -euo pipefail

# Pin gofumpt version for consistency
# Update this version when upgrading gofumpt
GOFUMPT_VERSION="v0.6.0"

# Check if gofumpt is installed
if ! command -v gofumpt &> /dev/null; then
    echo "gofumpt not found. Installing gofumpt@${GOFUMPT_VERSION}..."
    go install mvdan.cc/gofumpt@${GOFUMPT_VERSION}
fi

# Format all Go files
echo "Running gofumpt (version ${GOFUMPT_VERSION}) on all Go files..."
gofumpt -w .


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

# Generate implementation status from spec/features.yaml
go run ./cmd/gen-implementation-status \
  > docs/engine/status/implementation-status.md

echo "✓ Generated implementation status at docs/engine/status/implementation-status.md"


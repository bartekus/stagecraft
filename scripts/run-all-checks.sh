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
# run-all-checks.sh - Runs all checks
#
# Usage:
#   ./scripts/run-all-checks.sh

go build -o ./bin/stagecraft ./cmd/stagecraft
go vet ./... && go test ./... && staticcheck ./... && ./scripts/check-coverage.sh && ./scripts/validate-spec.sh

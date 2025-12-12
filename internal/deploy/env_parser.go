// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

package deploy

import (
	"strings"
)

// parseEnvFileInto parses a dotenv-format file and merges key-value pairs into env.
// Copied from internal/core/env/env.go to keep DEPLOY_COMPOSE_GEN independent.
// Semantics intentionally mirror the encorets provider parser for consistency.
// Handles: comments, export keyword, quoted values, inline comments,
// escaped characters in quoted strings, and empty values.
func parseEnvFileInto(env map[string]string, data []byte) {
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle export keyword (e.g., "export KEY=value")
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimPrefix(line, "export ")
			line = strings.TrimSpace(line)
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			// Skip malformed lines (no = found)
			continue
		}

		key := strings.TrimSpace(parts[0])
		if key == "" {
			// Skip lines with empty keys (e.g., "=value")
			continue
		}
		value := strings.TrimSpace(parts[1])

		// Handle inline comments (but preserve # inside quoted strings)
		commentIdx := -1
		inDoubleQuote := false
		inSingleQuote := false
		for i, r := range value {
			if r == '"' && (i == 0 || value[i-1] != '\\') {
				inDoubleQuote = !inDoubleQuote
			} else if r == '\'' && (i == 0 || value[i-1] != '\\') {
				inSingleQuote = !inSingleQuote
			} else if r == '#' && !inDoubleQuote && !inSingleQuote {
				commentIdx = i
				break
			}
		}
		if commentIdx >= 0 {
			value = strings.TrimSpace(value[:commentIdx])
		}

		// Handle quoted values with escaped characters
		if len(value) >= 2 {
			if value[0] == '"' && value[len(value)-1] == '"' {
				// Double-quoted string: handle escaped characters
				unquoted := value[1 : len(value)-1]
				unquoted = strings.ReplaceAll(unquoted, "\\\\", "\\")
				unquoted = strings.ReplaceAll(unquoted, "\\\"", "\"")
				unquoted = strings.ReplaceAll(unquoted, "\\n", "\n")
				unquoted = strings.ReplaceAll(unquoted, "\\t", "\t")
				unquoted = strings.ReplaceAll(unquoted, "\\r", "\r")
				value = unquoted
			} else if value[0] == '\'' && value[len(value)-1] == '\'' {
				// Single-quoted string: no escape sequences (remove quotes only)
				value = value[1 : len(value)-1]
			}
		}

		// Later values override earlier ones (map behavior)
		env[key] = value
	}
}

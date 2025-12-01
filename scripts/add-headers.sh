#!/bin/bash
# SPDX-License-Identifier: AGPL-3.0-or-later
#
# Stagecraft - A Go-based CLI for orchestrating local-first multi-service deployments using Docker Compose.
#
# Copyright (C) 2025  Bartek Kus
#
# This program is free software licensed under the terms of the GNU AGPL v3 or later.
#
# See https://www.gnu.org/licenses/ for license details.
#
# add-headers.sh - Helper script to add license headers to Go files

set -e

SHORT_HEADER='// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - A Go-based CLI for orchestrating local-first multi-service deployments using Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

'

# Files that should have full header (already done: cmd/stagecraft/main.go)
FULL_HEADER_FILES=("cmd/stagecraft/main.go")

# Find all Go files
find . -name "*.go" -type f ! -path "./vendor/*" ! -path "./node_modules/*" ! -path "./.git/*" | while read -r file; do
	# Skip if already has SPDX
	if grep -q "SPDX-License-Identifier" "$file"; then
		echo "Skipping $file (already has header)"
		continue
	fi
	
	# Check if it's a full header file
	is_full_header=false
	for full_file in "${FULL_HEADER_FILES[@]}"; do
		if [[ "$file" == "./$full_file" ]]; then
			is_full_header=true
			break
		fi
	done
	
	if [ "$is_full_header" = true ]; then
		echo "Skipping $file (full header file, should be done manually)"
		continue
	fi
	
	# Create temp file with header
	temp_file=$(mktemp)
	echo -n "$SHORT_HEADER" > "$temp_file"
	
	# Remove leading comments if they exist (like // internal/cli/commands/init.go)
	# and add the rest of the file
	if head -1 "$file" | grep -q "^//"; then
		# Skip first line if it's a comment
		tail -n +2 "$file" >> "$temp_file"
	else
		cat "$file" >> "$temp_file"
	fi
	
	mv "$temp_file" "$file"
	echo "Added header to $file"
done

echo "Done adding headers!"


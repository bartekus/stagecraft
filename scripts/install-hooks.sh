#!/bin/bash
# install-hooks.sh - Install git hooks for Stagecraft

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

HOOKS_DIR=".git/hooks"
HOOKS_SOURCE=".hooks"

if [ ! -d ".git" ]; then
    echo "Error: .git directory not found. Are you in the project root?"
    exit 1
fi

# Create hooks directory if it doesn't exist
mkdir -p "$HOOKS_DIR"

# Install pre-commit hook
if [ -f "$HOOKS_SOURCE/pre-commit" ]; then
    ln -sf "../../$HOOKS_SOURCE/pre-commit" "$HOOKS_DIR/pre-commit"
    chmod +x "$HOOKS_DIR/pre-commit"
    echo "✓ Installed pre-commit hook"
else
    echo "Warning: $HOOKS_SOURCE/pre-commit not found"
fi

echo "Git hooks installed successfully!"
echo ""
echo "To test, try making a formatting error and committing:"
echo "  echo 'package main' > test.go"
echo "  git add test.go"
echo "  git commit -m 'test'  # Should fail"


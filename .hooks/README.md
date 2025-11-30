# Git Hooks

This directory contains git hooks for Stagecraft development.

## Installation

To install the pre-commit hook:

```bash
# From the project root
ln -s ../../.hooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

Or use the install script:

```bash
./scripts/install-hooks.sh
```

## Available Hooks

### pre-commit

Runs before each commit:
- Format check (gofmt)
- Import check (goimports, if available)
- Go vet
- Tests on changed packages
- golangci-lint (if available, warning only)

The hook will prevent commits if:
- Files are not formatted
- go vet finds issues
- Tests fail

It will warn (but not block) if:
- Imports need formatting
- golangci-lint finds issues

## Bypassing Hooks

If you need to bypass hooks (not recommended):

```bash
git commit --no-verify
```

## Troubleshooting

### Hook not running

Ensure the hook is executable and linked correctly:

```bash
ls -la .git/hooks/pre-commit
# Should show: .git/hooks/pre-commit -> ../../.hooks/pre-commit
```

### Tests taking too long

The hook runs tests with `-short` flag. If tests are still slow, consider:
- Running only unit tests (not integration tests)
- Using `git commit --no-verify` for WIP commits (fix before pushing)


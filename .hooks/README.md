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

The pre-commit hook automatically fixes common issues and performs quick checks on staged files:

**Auto-fixes (non-blocking):**
- ✅ **Formatting**: Auto-formats Go files with `gofumpt` (or `gofmt` as fallback)
- ✅ **Imports**: Auto-organizes imports with `goimports`
- ✅ **License headers**: Auto-adds license headers to files missing them

**Checks (blocking):**
- ❌ **Build errors**: Quick build check on changed packages (fails if code doesn't compile)

**Edge case handling:**
- Entry files (like `cmd/stagecraft/main.go`) need full headers - the hook will warn but skip auto-adding to avoid incorrect format
- Files in ignore patterns (like `test_script.sh`) are skipped
- Missing tools (gofumpt, goimports, addlicense) result in warnings but don't block commits

**What gets processed:**
- Only staged files are processed (fast and focused)
- Go files: `.go`
- Script files: `.sh`
- Config files: `.yaml`, `.yml`

## Environment Variables

### STAGECRAFT_SKIP_HOOKS
Skip all pre-commit hook checks:
```bash
STAGECRAFT_SKIP_HOOKS=1 git commit -m "message"
```

**Note:** `SKIP_HOOKS=1` also works for backward compatibility.

### HOOK_VERBOSE
Enable verbose output to see detailed hook operations:
```bash
HOOK_VERBOSE=1 git commit -m "message"
```

## Tool Execution Order

The hook runs tools in this order to ensure consistent results:

1. **goimports** - Organizes imports first
2. **gofumpt** - Formats code (final formatting authority)
3. **addlicense** - Adds license headers
4. **Build check** - Verifies code compiles

This order ensures gofumpt's stricter formatting rules are the final authority.

## Bypassing Hooks

### Recommended: Environment variable
```bash
STAGECRAFT_SKIP_HOOKS=1 git commit -m "message"
```

### Alternative: Git flag
```bash
git commit --no-verify
```

**Note:** CI will still run all checks, so bypassing hooks should only be used for WIP commits that will be fixed before pushing.

## Troubleshooting

### Hook not running

Ensure the hook is executable and linked correctly:

```bash
ls -la .git/hooks/pre-commit
# Should show: .git/hooks/pre-commit -> ../../.hooks/pre-commit
```

### Missing tools

If you see warnings about missing tools:

```bash
# Install gofumpt (preferred formatter)
go install mvdan.cc/gofumpt@latest

# Install goimports (import organizer)
go install golang.org/x/tools/cmd/goimports@latest

# Install addlicense (license header tool)
go install github.com/google/addlicense@latest
```

### Entry files need manual header updates

Entry files (like `cmd/stagecraft/main.go`) require full headers, not short headers. The hook will warn you if an entry file is missing a header. See `CONTRIBUTING.md` for the full header format.

### Build check fails

If the build check fails, fix the compilation errors before committing. The hook only checks packages with staged changes, so it should be fast.

### Unstaged files modified

If you see a warning about unstaged files being modified, the formatting/import tools may have touched files that weren't staged. Review the changes and stage any you want to include:

```bash
git diff <file>  # Review changes
git add <file>   # Stage if desired
```

## Relationship to CI

The pre-commit hook provides fast, local feedback and auto-fixes. CI provides comprehensive verification:

- **Pre-commit**: Auto-fixes formatting/imports/headers, quick build check
- **CI**: Full test suite, coverage checks, comprehensive linting, format verification

This two-stage approach ensures:
1. Developers get immediate feedback and fixes locally
2. CI verifies everything is correct before merging

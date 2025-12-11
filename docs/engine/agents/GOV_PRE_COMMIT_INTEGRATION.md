# Governance Pre-Commit Hook Integration

This document explains how to integrate the governance pre-commit checks into your Git hooks.

---

## Quick Integration

### Option 1: Add to Existing Hook (Recommended)

Add this snippet to the end of `.hooks/pre-commit` (before the final `info "Pre-commit checks passed"` line):

```bash
# === Governance checks (optional but recommended) ===
# Source: .hooks/pre-commit-gov-snippet.sh
if [ "${SKIP_GOV_PRE_COMMIT:-}" != "1" ] && [ -x "$PROJECT_ROOT/scripts/gov-pre-commit.sh" ]; then
    if ! bash "$PROJECT_ROOT/scripts/gov-pre-commit.sh"; then
        error "Governance pre-commit checks failed"
        error "Fix the issues above, or skip with: SKIP_GOV_PRE_COMMIT=1 git commit"
        exit 1
    fi
fi
```

**Location**: Insert this block after line 307 (after the build check) and before line 309 (the final success message).

### Option 2: Standalone Hook

If you prefer to keep governance checks separate, you can create `.git/hooks/pre-commit-governance`:

```bash
#!/usr/bin/env bash
# Copy the contents of .hooks/pre-commit-gov-snippet.sh here
```

Then add this to your main `.hooks/pre-commit`:

```bash
# Run governance checks (if not skipped)
if [ "${SKIP_GOV_PRE_COMMIT:-}" != "1" ]; then
    if [ -x "$PROJECT_ROOT/.git/hooks/pre-commit-governance" ]; then
        bash "$PROJECT_ROOT/.git/hooks/pre-commit-governance" || exit 1
    fi
fi
```

---

## Escape Hatches

The governance checks can be bypassed with:

```bash
# Skip only governance checks
SKIP_GOV_PRE_COMMIT=1 git commit

# Skip all hooks (including governance)
STAGECRAFT_SKIP_HOOKS=1 git commit
```

---

## What It Checks

The `scripts/gov-pre-commit.sh` script runs:

1. **Feature mapping validation** (`stagecraft gov feature-mapping`)
   - Ensures Feature IDs match between specs and code
   - Validates spec paths are correct
   - Checks for missing implementations/tests

2. **Orphan spec detection** (`check-orphan-specs.sh`)
   - Finds spec files not referenced in `spec/features.yaml`
   - Detects dead references

3. **Core coverage guardrail** (`go test -cover ./pkg/config ./internal/core`)
   - Ensures core packages maintain ≥80% coverage threshold
   - Fast check on critical packages

4. **Full project checks** (`run-all-checks.sh` or `go test ./...`)
   - Complete validation suite
   - Can be skipped with `GOV_FAST=1` for faster local iteration

---

## Fast Mode

For faster local iteration, you can set `GOV_FAST=1` to skip the full `run-all-checks.sh`:

```bash
GOV_FAST=1 git commit
```

This still runs:
- Feature mapping validation
- Orphan spec checks
- Core coverage guardrail

But skips the full project check suite.

---

## Installation

After adding the snippet to `.hooks/pre-commit`, reinstall hooks:

```bash
./scripts/install-hooks.sh
```

Or manually:

```bash
chmod +x .hooks/pre-commit
ln -sf ../../.hooks/pre-commit .git/hooks/pre-commit
```

---

## Troubleshooting

### "stagecraft binary not found"

The script will automatically build the binary if missing. If build fails:

```bash
go build -o ./bin/stagecraft ./cmd/stagecraft
```

### "check-orphan-specs.sh not found"

This is a warning, not an error. The script will continue. Install the script or skip this check.

### Governance checks are too slow

Use fast mode:

```bash
GOV_FAST=1 git commit
```

Or skip entirely for this commit:

```bash
SKIP_GOV_PRE_COMMIT=1 git commit
```

---

## Integration with CI

These checks are designed to catch issues locally before CI. CI should run the same checks (or stricter) to ensure consistency.

The governance pre-commit hook is a **safety net**, not a replacement for CI validation.

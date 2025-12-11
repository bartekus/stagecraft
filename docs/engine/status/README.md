# Implementation Status

This directory contains generated documentation that tracks feature implementation status.

## Files

- `implementation-status.md` - Auto-generated from `spec/features.yaml`
- `PROVIDER_FRONTEND_GENERIC_COVERAGE_V1_COMPLETE.md` - Coverage completion tracking for PROVIDER_FRONTEND_GENERIC
- `PROVIDER_FRONTEND_GENERIC_COVERAGE_PR.md` - PR description and reviewer guide for coverage completion
- `PROVIDER_NETWORK_TAILSCALE_COVERAGE_PLAN.md` - Coverage plan tracking for PROVIDER_NETWORK_TAILSCALE

**Note**: Provider coverage status documents follow the pattern `PROVIDER_*_COVERAGE_*.md` (either `*_COVERAGE_V1_COMPLETE.md` for completed coverage or `*_COVERAGE_PLAN.md` for in-progress work).

## Generation

This file is generated from `spec/features.yaml` using the `gen-implementation-status` tool.

**To regenerate:**

```bash
./scripts/generate-implementation-status.sh
```

Or manually:

```bash
go run ./cmd/gen-implementation-status
```

## Source of Truth

The source of truth for feature status is **always** `spec/features.yaml`. This markdown file is a human-readable snapshot for quick reference.

**Never edit `implementation-status.md` manually.** All changes should be made to `spec/features.yaml`, then regenerate this file.

## CI Integration

In CI, you can verify the file is up-to-date:

```bash
# Generate the file
./scripts/generate-implementation-status.sh

# Check if it changed
git diff --exit-code docs/engine/status/implementation-status.md
```

If the diff is non-empty, the file is out of sync with `spec/features.yaml`.


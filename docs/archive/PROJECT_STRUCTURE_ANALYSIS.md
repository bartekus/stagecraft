<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<!--

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

-->

---
status: archived
scope: v1
superseded_by: ../adr/0001-architecture.md
---

# Project Structure Analysis & Improvement Recommendations

This document analyzes the current Stagecraft project structure and provides recommendations for improvements, alignments, and cleanups.

**Date**: 2025-12-07  
**Status**: Analysis Complete (Archived)

> **Note**: This analysis document has been superseded by [ADR 0001 - Architecture](../adr/0001-architecture.md), which formalizes the project structure decisions. This document is retained for historical reference.

---

## Executive Summary

The project structure is generally well-organized and follows Go best practices. However, there are several issues that need attention:

1. **Critical**: Missing `cmd/stagecraft/main.go` entry point (referenced everywhere but doesn't exist)
2. **High**: Binary artifacts in repository root (should be in `.gitignore` or `bin/`)
3. **Medium**: Inconsistent documentation organization
4. **Medium**: Some naming inconsistencies
5. **Low**: Opportunities for better alignment with Go conventions

---

## Critical Issues

### 1. Main Entry Point

**Status**: ✅ **RESOLVED** - `cmd/stagecraft/main.go` exists and is correctly implemented.

**Current State**: File exists with proper:
- Full license header (as required for entry files)
- Correct package declaration
- Proper error handling
- Clean integration with `cli.NewRootCommand()`

**Verification**: Build succeeds: `go build ./cmd/stagecraft`

**Note**: This was initially flagged as missing during analysis, but verification shows it exists and is correct.

---

## High Priority Issues

### 2. Binary Artifacts Tracked in Git

**Issue**: The following binary files are **currently tracked in git** (should not be):
- `cli-introspect` ⚠️ **TRACKED**
- `features-tool` ⚠️ **TRACKED**
- `gen-features-overview` ⚠️ **TRACKED**
- `spec-validate` ⚠️ **TRACKED**
- `spec-vs-cli` ⚠️ **TRACKED**
- `stagecraft` ⚠️ **TRACKED**
- `coverage.out` (not shown in tracked files, but should be excluded)

**Current State**: These are build artifacts that should not be committed.

**Impact**: 
- Repository bloat
- Potential conflicts
- Binary files in version control (bad practice)
- `.gitignore` excludes them going forward, but existing tracked files remain

**Recommendation**:
1. ✅ `.gitignore` properly excludes these (already done)
2. **Remove from git tracking** (keep local files):
   ```bash
   git rm --cached cli-introspect features-tool gen-features-overview \
     spec-validate spec-vs-cli stagecraft coverage.out 2>/dev/null || true
   ```
3. Commit the removal: `git commit -m "Remove binary artifacts from git tracking"`
4. Ensure build scripts output to `bin/` directory consistently

**Priority**: 🟠 High - Repository hygiene (should be fixed)

### 3. Inconsistent Build Output Locations

**Issue**: 
- Scripts reference `./bin/stagecraft` but also build to root
- Some tools may build to root directory

**Recommendation**: 
- Standardize all builds to `bin/` directory
- Update all build scripts to use `bin/` consistently
- Add `bin/` to `.gitignore` (✅ already done)

**Priority**: 🟠 High - Build consistency

---

## Medium Priority Issues

### 4. Documentation Organization

**Current Structure**:
```
docs/
├── adr/                    # Architecture Decision Records
├── analysis/               # Implementation analysis
├── architecture.md         # Architecture overview (root level)
├── CLI_*_ANALYSIS.md       # Command analysis (root level)
├── CLI_*_IMPLEMENTATION_OUTLINE.md  # Implementation outlines (root level)
├── context-handoff/        # Feature handoff docs
├── CONTRIBUTING_CURSOR.md  # Cursor workflow guide
├── engine-index.md         # Engine doc index
├── features/               # Feature docs
├── FUTURE_ENHANCEMENTS.md  # Future plans (root level)
├── guides/                 # User guides
├── IMPLEMENTATION_OUTLINE_TEMPLATE.md  # Template (root level)
├── implementation-roadmap.md  # Roadmap (root level)
├── implementation-status.md  # Status (root level)
├── providers/              # Provider docs
├── reference/              # Reference docs
├── registry-implementation-summary.md  # Registry summary (root level)
└── stagecraft-spec.md       # Main spec (root level)
```

**Issues**:
- Mix of root-level and subdirectory organization
- Some files could be better grouped
- Templates mixed with actual docs

**Recommendations**:

#### Option A: Minimal Reorganization (Recommended)
Keep current structure but add clear organization:

```
docs/
├── engine/                  # NEW: Core technical docs
│   ├── stagecraft-spec.md
│   ├── implementation-status.md
│   ├── registry-implementation-summary.md
│   ├── CLI_*_ANALYSIS.md
│   ├── CLI_*_IMPLEMENTATION_OUTLINE.md
│   └── IMPLEMENTATION_OUTLINE_TEMPLATE.md
├── narrative/               # NEW: Human-facing docs
│   ├── implementation-roadmap.md
│   ├── FUTURE_ENHANCEMENTS.md
│   └── architecture.md
├── guides/                  # User-facing guides
├── reference/              # API/reference docs
├── features/               # Feature documentation
├── providers/              # Provider docs
├── context-handoff/        # Handoff docs
├── analysis/               # Implementation analysis
├── adr/                    # Architecture decisions
└── CONTRIBUTING_CURSOR.md  # Contributor guide
```

#### Option B: Keep Current (Simpler)
- Add `docs/README.md` explaining organization
- Document which docs are "engine" vs "narrative" in comments
- Update `.cursorignore` to exclude narrative docs during core work

**Priority**: 🟡 Medium - Organization clarity

### 5. Test File Organization

**Current State**: Tests are co-located with source files (Go standard), which is good.

**Minor Issue**: Some test helpers in `internal/cli/commands/`:
- `test_helpers.go` - Good
- `phases_test_helpers_test.go` - Could be `phases_test_helpers.go` (not a test file)
- `golden_test.go` - Good
- `phases_golden_test.go` - Good

**Recommendation**: 
- Rename `phases_test_helpers_test.go` → `phases_test_helpers.go` (it's a helper, not a test)
- Or move to `testdata/helpers.go` if it's only used by tests

**Priority**: 🟡 Medium - Naming consistency

### 6. Provider Structure Naming

**Current State**:
- `internal/providers/backend/encorets/` - ✅ Good (matches spec name)
- `internal/providers/backend/generic/` - ✅ Good
- `internal/providers/migration/raw/` - ✅ Good

**Issue**: Spec file is `encore-ts.md` but directory is `encorets`. This is fine (Go package naming), but could be documented.

**Recommendation**: Add comment in `encorets.go` explaining the naming:
```go
// Package encorets implements the Encore.ts backend provider.
// The package name "encorets" (not "encore-ts") follows Go naming conventions
// which don't allow hyphens in package names.
// Spec: spec/providers/backend/encore-ts.md
```

**Priority**: 🟡 Medium - Documentation clarity

---

## Low Priority Improvements

### 7. Script Organization

**Current State**: All scripts in `scripts/` directory - ✅ Good

**Minor Improvement**: Could add `scripts/README.md` explaining:
- What each script does
- When to run them
- Dependencies

**Priority**: 🟢 Low - Nice to have

### 8. Example Structure

**Current State**: `examples/basic-node/` - ✅ Good

**Recommendation**: 
- Add `examples/README.md` explaining examples
- Consider adding more examples as features are implemented

**Priority**: 🟢 Low - Future enhancement

### 9. Root-Level Documentation Files

**Current Files**:
- `Agent.md` - ✅ Core governance, should stay
- `ASSESSMENT.md` - Could move to `docs/`
- `CONTRIBUTING.md` - ✅ Standard location
- `IMPROVEMENTS.md` - Could move to `docs/` or `docs/narrative/`
- `PR_SUMMARY.md` - Temporary? Could be in `.git/` or removed
- `README.md` - ✅ Standard location

**Recommendation**:
- Keep `Agent.md`, `CONTRIBUTING.md`, `README.md` in root
- Move `ASSESSMENT.md` → `docs/engine/analysis/ASSESSMENT.md`
- Move `IMPROVEMENTS.md` → `docs/narrative/IMPROVEMENTS.md` (if creating narrative/)
- `PR_SUMMARY.md` - Review if still needed, otherwise remove

**Priority**: 🟢 Low - Organization

### 10. Go Module Structure

**Current State**: 
- Module name: `stagecraft` (local)
- Imports use: `stagecraft/internal/...`, `stagecraft/pkg/...`

**Status**: ✅ Correct for local development

**Future Consideration**: If publishing, would need to change to full import path (e.g., `github.com/bartekus/stagecraft`)

**Priority**: 🟢 Low - Future consideration

---

## Alignment with Go Best Practices

### ✅ What's Good

1. **Package Structure**: Clear separation of `cmd/`, `internal/`, `pkg/`
2. **Test Organization**: Tests co-located with source (Go standard)
3. **Internal vs Public**: Good use of `internal/` for private code
4. **Provider Pattern**: Clean separation of interfaces (`pkg/providers/`) and implementations (`internal/providers/`)
5. **Command Structure**: Well-organized CLI commands in `internal/cli/commands/`

### 🔄 Could Be Improved

1. **Build Artifacts**: Some binaries tracked in git (should be removed from tracking)
2. **Documentation**: Could benefit from clearer organization

---

## Recommended Action Plan

### Phase 1: Critical Fixes (Do First)
1. ✅ **VERIFIED** - `cmd/stagecraft/main.go` exists and is correct
2. ⚠️ **ACTION NEEDED** - Remove tracked binary files from git:
   ```bash
   git rm --cached cli-introspect features-tool gen-features-overview \
     spec-validate spec-vs-cli stagecraft coverage.out 2>/dev/null || true
   git commit -m "Remove binary artifacts from git tracking"
   ```
3. ⏸️ Update build scripts to use `bin/` consistently

### Phase 2: Organization (Do Next)
1. ⏸️ Decide on documentation organization (Option A or B)
2. ⏸️ If Option A: Create `docs/engine/` and `docs/narrative/`, move files
3. ⏸️ Update `.cursorignore` if needed
4. ⏸️ Add `docs/README.md` explaining structure

### Phase 3: Polish (Do When Time Permits)
1. ⏸️ Rename `phases_test_helpers_test.go` if appropriate
2. ⏸️ Add package comments explaining naming conventions
3. ⏸️ Add `scripts/README.md`
4. ⏸️ Add `examples/README.md`
5. ⏸️ Review and organize root-level docs

---

## File Location Checklist

### ✅ Correctly Located
- `cmd/stagecraft/main.go` - Main CLI entry point ✅
- `cmd/*/main.go` - Tool entry points
- `internal/cli/` - CLI implementation
- `internal/core/` - Core domain logic
- `internal/providers/` - Provider implementations
- `pkg/providers/` - Provider interfaces
- `pkg/config/`, `pkg/logging/`, `pkg/executil/` - Public utilities
- `spec/` - Specifications
- `docs/` - Documentation
- `scripts/` - Build/utility scripts
- `test/e2e/` - End-to-end tests

### ⚠️ Needs Attention
- Binary files tracked in git - Should be removed from tracking (already in `.gitignore`)
- Some docs could be better organized (optional improvement)

---

## Summary

The project structure is fundamentally sound and follows Go conventions well. The main issues are:

1. ✅ **Main entry point** - Verified: `cmd/stagecraft/main.go` exists and is correct
2. **Binary artifacts tracked in git** - Should be removed from tracking (already excluded by `.gitignore`)
3. **Documentation organization** - Could be improved but not critical

Most recommendations are organizational improvements that can be done incrementally without breaking changes.

---

## Questions for Discussion

1. Should we implement Option A (reorganize docs) or Option B (document current structure)?
2. Is `PR_SUMMARY.md` still needed, or can it be removed?
3. Should we create a `Makefile` to standardize build commands?
4. Do we want to add `scripts/README.md` and `examples/README.md`?

---

**Next Steps**: 
1. ✅ **DONE** - `cmd/stagecraft/main.go` exists and is correct
2. ⚠️ **ACTION NEEDED** - Remove tracked binary files from git (see Phase 1, step 2 above)
3. ⏸️ Review and decide on documentation organization approach (optional)


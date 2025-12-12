> **Superseded by** `docs/governance/GOVERNANCE_ALMANAC.md` section 7 (Violation Handling and Fixes). Kept for historical reference. New governance rules MUST be recorded in the almanac.

# Phase 5 Violation Fix Checklist

**Generated:** 2025-01-XX  
**Total Violations:** 15  
**Status:** Ready for execution

---

## Overview

This checklist organizes the 15 current mapping violations into a prioritized, actionable fix plan for Phase 5 Workstream A.

**Fix Order Strategy:**
1. **SPEC_PATH_MISMATCH** (7 violations) - Quick header fixes, no logic changes
2. **MISSING_TESTS** (3 violations) - Add Feature headers to existing test files
3. **MISSING_IMPL** (4 violations) - Add headers or fix feature status
4. **MISSING_SPEC** (1 violation) - Fix spec path or downgrade feature

---

## Priority 1: SPEC_PATH_MISMATCH (7 violations)

**Impact:** High - These are simple header fixes that will immediately reduce violation count.  
**Effort:** Low - Just update `// Spec:` comments.

### 1.1 GOV_V1_CORE → Wrong Spec Path (4 files)

**Issue:** Files declare `spec/commands/commit-suggest.md` but canonical path is `spec/governance/GOV_V1_CORE.md`.

**Root Cause:** These files are actually implementing `CLI_COMMIT_SUGGEST` but have the wrong Feature ID. They should have `CLI_COMMIT_SUGGEST` as the Feature, not `GOV_V1_CORE`.

**Files to Fix:**
- [ ] `internal/cli/commands/commit_suggest.go` - Change `Feature: GOV_V1_CORE` → `Feature: CLI_COMMIT_SUGGEST` and `Spec: spec/commands/commit-suggest.md` (keep spec path)
- [ ] `internal/cli/commands/commit_suggest_test.go` - Change `Feature: GOV_V1_CORE` → `Feature: CLI_COMMIT_SUGGEST` and `Spec: spec/commands/commit-suggest.md` (keep spec path)
- [ ] `internal/reports/suggestions/suggestions.go` - Change `Feature: GOV_V1_CORE` → `Feature: CLI_COMMIT_SUGGEST` and `Spec: spec/commands/commit-suggest.md` (keep spec path)
- [ ] `internal/reports/suggestions/suggestions_test.go` - Change `Feature: GOV_V1_CORE` → `Feature: CLI_COMMIT_SUGGEST` and `Spec: spec/commands/commit-suggest.md` (keep spec path)

**Note:** This will also fix the `MISSING_IMPL` violation for `CLI_COMMIT_SUGGEST` since the files will then be correctly attributed.

### 1.2 MIGRATION_INTERFACE → Wrong Spec Path (1 file)

**Issue:** File declares `spec/core/migration-registry.md` but canonical path is `spec/migrations/interface.md`.

**Files to Fix:**
- [ ] `pkg/providers/migration/migration.go` - Change `Spec: spec/core/migration-registry.md` → `Spec: spec/migrations/interface.md`

### 1.3 PROVIDER_FRONTEND_GENERIC → Wrong Spec Path (2 files)

**Issue:** Files declare `spec/commands/deploy.md` but canonical path is `spec/providers/frontend/generic.md`.

**Files to Fix:**
- [ ] `internal/cli/commands/feature_traceability_test.go` - Change `Spec: spec/commands/deploy.md` → `Spec: spec/providers/frontend/generic.md`
- [ ] `internal/reports/featuretrace/scan_test.go` - Change `Spec: spec/commands/deploy.md` → `Spec: spec/providers/frontend/generic.md`

**Verification:**
```bash
./bin/stagecraft gov feature-mapping | grep SPEC_PATH_MISMATCH
# Should return 0 violations after fixes
```

---

## Priority 2: MISSING_TESTS (3 violations)

**Impact:** Medium - These features have implementation but mapping doesn't see test files with the Feature header.  
**Effort:** Low - Add `// Feature: FEATURE_ID` headers to existing test files.

### 2.1 CLI_GLOBAL_FLAGS

**Issue:** Feature is `done` but mapping reports no tests.

**Current State:**
- ✅ Has header in `internal/cli/root_test.go` (line 74)
- ❌ Mapping tool may not be recognizing it (check if it's in a `_test.go` file)

**Action:**
- [ ] Verify `internal/cli/root_test.go` has `// Feature: CLI_GLOBAL_FLAGS` in a test function context
- [ ] If header is at package level, ensure it's before the first test function
- [ ] Re-run mapping to confirm fix

### 2.2 CORE_STATE_TEST_ISOLATION

**Issue:** Feature is `done` but mapping reports no tests.

**Current State:**
- ✅ Has header in `internal/cli/commands/test_helpers.go` (line 26)
- ❌ File is `test_helpers.go`, not `*_test.go` - mapping may only scan `*_test.go` files

**Action:**
- [ ] Check if mapping tool scans `test_helpers.go` files
- [ ] If not, either:
  - Rename to `test_helpers_test.go`, or
  - Add Feature header to an actual `*_test.go` file that uses the test helpers (e.g., `deploy_test.go`, `rollback_test.go`)
- [ ] Re-run mapping to confirm fix

### 2.3 MIGRATION_ENGINE_RAW

**Issue:** Feature is `done` but mapping reports no tests.

**Current State:**
- ✅ Has header in `internal/providers/migration/raw/raw.go` (line 32)
- ❌ Test file `raw_test.go` has `// Feature: GOV_V1_CORE` instead of `MIGRATION_ENGINE_RAW`

**Action:**
- [ ] Update `internal/providers/migration/raw/raw_test.go` - Change `Feature: GOV_V1_CORE` → `Feature: MIGRATION_ENGINE_RAW`
- [ ] Re-run mapping to confirm fix

**Verification:**
```bash
./bin/stagecraft gov feature-mapping | grep MISSING_TESTS
# Should return 0 violations after fixes
```

---

## Priority 3: MISSING_IMPL (4 violations)

**Impact:** Medium - These features are marked `done` but mapping doesn't find implementation files with the Feature header.  
**Effort:** Medium - May require adding headers or reconsidering feature status.

### 3.1 CLI_COMMIT_SUGGEST

**Issue:** Feature is `done` but mapping reports no implementation.

**Root Cause:** Files exist but have wrong Feature ID (`GOV_V1_CORE` instead of `CLI_COMMIT_SUGGEST`).

**Action:**
- [ ] **Already covered in Priority 1.1** - Fixing SPEC_PATH_MISMATCH will also fix this
- [ ] After Priority 1.1 fixes, verify this violation is resolved

### 3.2 CORE_BACKEND_PROVIDER_CONFIG_SCHEMA

**Issue:** Feature is `done` but mapping reports no implementation.

**Current State:**
- ✅ Spec exists: `spec/core/backend-provider-config.md`
- ✅ Test listed in features.yaml: `pkg/config/config_test.go`
- ❌ No files have `// Feature: CORE_BACKEND_PROVIDER_CONFIG_SCHEMA` header

**Action:**
- [ ] Check `pkg/config/config.go` and `pkg/config/config_test.go` to identify where provider config schema is implemented
- [ ] Add `// Feature: CORE_BACKEND_PROVIDER_CONFIG_SCHEMA` header to the relevant function/type in `config.go`
- [ ] Add `// Feature: CORE_BACKEND_PROVIDER_CONFIG_SCHEMA` header to the relevant test in `config_test.go`
- [ ] Re-run mapping to confirm fix

**Alternative:** If this feature is truly just schema definition (no code), consider:
- Downgrade to `wip` with note, or
- Mark as documentation-only feature (may need governance spec update)

### 3.3 CORE_STATE_CONSISTENCY

**Issue:** Feature is `done` but mapping reports no implementation.

**Current State:**
- ✅ Has header in `internal/core/state/state.go` (line 220)
- ❌ Header is **inside a function** (`saveState` method) - mapping tool scans file headers, not function-level comments
- ❌ Mapping expects headers at package level or before first function

**Action:**
- [ ] Add `// Feature: CORE_STATE_CONSISTENCY` header at **package level** in `state.go` (after package declaration, before imports)
- [ ] Keep the existing header in `saveState` function for documentation, but package-level header is required for mapping
- [ ] Ensure `state_test.go` also has `// Feature: CORE_STATE_CONSISTENCY` header at package level or in a test function
- [ ] Re-run mapping to confirm fix

### 3.4 DOCS_ADR

**Issue:** Feature is `done` but mapping reports no implementation.

**Current State:**
- ✅ Spec file exists at `docs/adr/0001-architecture.md` (verified)
- ❌ Spec path in features.yaml is `adr/0001-architecture.md` (resolves to `spec/adr/0001-architecture.md`)
- ❌ File is actually in `docs/adr/`, not `spec/adr/`
- ❌ This is a documentation-only feature (no code implementation)

**Action:**
- [ ] **Option 1 (Recommended):** Move `docs/adr/0001-architecture.md` → `spec/adr/0001-architecture.md`
  - This aligns with governance expectation that specs live under `spec/`
  - Update any cross-references in docs
- [ ] **Option 2:** Update `spec/features.yaml` to point to `docs/adr/0001-architecture.md`
  - May break assumptions that all specs are under `spec/`
- [ ] For implementation requirement:
  - Since this is documentation-only, consider downgrading to `wip` with note
  - Or add minimal implementation file with header for governance compliance
- [ ] Re-run mapping to confirm fix

**Verification:**
```bash
./bin/stagecraft gov feature-mapping | grep MISSING_IMPL
# Should return 0 violations after fixes
```

---

## Priority 4: MISSING_SPEC (1 violation)

**Impact:** Low - Single violation, likely a path issue.  
**Effort:** Low - Fix path or create spec file.

### 4.1 DOCS_ADR

**Issue:** Feature declares spec path `spec/adr/0001-architecture.md` but file doesn't exist at that location.

**Current State:**
- ✅ File exists at `docs/adr/0001-architecture.md` (verified)
- ❌ Mapping expects it at `spec/adr/0001-architecture.md` (per features.yaml path `adr/0001-architecture.md`)

**Action:**
- [ ] **Recommended:** Move `docs/adr/0001-architecture.md` → `spec/adr/0001-architecture.md`
  - Create `spec/adr/` directory if needed
  - Update any cross-references in other docs
  - This aligns with governance that specs live under `spec/`
- [ ] **Alternative:** Update `spec/features.yaml` spec path to `docs/adr/0001-architecture.md`
  - May require mapping tool changes if it only scans `spec/` directory
- [ ] Re-run mapping to confirm fix

**Verification:**
```bash
./bin/stagecraft gov feature-mapping | grep MISSING_SPEC
# Should return 0 violations after fixes
```

---

## Execution Notes

### After Each Priority Group

1. **Commit changes:**
   ```bash
   git add -A
   git commit -m "fix(GOV_V1_CORE): resolve SPEC_PATH_MISMATCH violations"
   ```

2. **Verify fixes:**
   ```bash
   ./bin/stagecraft gov feature-mapping
   ```

3. **Run full checks:**
   ```bash
   ./scripts/run-all-checks.sh
   ```

### Final Verification

After all fixes:

```bash
# Should show 0 violations
./bin/stagecraft gov feature-mapping

# Should pass cleanly
./scripts/run-all-checks.sh
```

### Expected Outcome

- **Before:** 15 violations across 4 categories
- **After:** 0 violations
- **Features:** All `done` features have proper spec, impl, and test headers
- **Status:** Ready for Workstream B (align spec/features.yaml with reality)

---

## Next Steps After This Checklist

Once all violations are fixed:

1. **Workstream B:** Cross-check `spec/features.yaml` statuses against reality
2. **Workstream D:** Update GOV_V1_CORE spec and implementation status
3. **Final:** Mark Phase 5 complete


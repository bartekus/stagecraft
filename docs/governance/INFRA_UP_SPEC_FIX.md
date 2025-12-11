# CLI_INFRA_UP - Spec Frontmatter Fix

**Feature**: CLI_INFRA_UP  
**Issue**: YAML frontmatter `exit_codes` format mismatch  
**Priority**: Low (follow-up after governance commit)

---

## Problem

The `spec/commands/infra-up.md` file uses an incorrect format for `exit_codes`:

**Current (incorrect)**:
```yaml
exit_codes:
  "0": "All hosts created and bootstrap succeeded"
  "10": "Some hosts failed bootstrap (partial failure)"
  "1": "Config error"
  "2": "CloudProvider failure"
  "3": "Global bootstrap error"
```

**Issue**: The YAML parser expects `map[string]int`, but this provides `map[string]string` (string keys with string values).

**Error**:
```
yaml: unmarshal errors:
  line 9: cannot unmarshal !!str `All hos...` into int
  line 10: cannot unmarshal !!str `Some ho...` into int
  ...
```

---

## Solution

Change to use symbolic names as keys and integer values (matching other CLI specs):

**Correct format**:
```yaml
exit_codes:
  success: 0
  partial_failure: 10
  config_error: 1
  cloud_provider_failure: 2
  global_bootstrap_error: 3
```

**Reference**: See `spec/commands/build.md` for a similar multi-exit-code pattern:
```yaml
exit_codes:
  success: 0
  user_error: 1
  plan_failure: 2
  build_failure: 3
  push_failure: 4
```

---

## Implementation

### Step 1: Update Frontmatter

**File**: `spec/commands/infra-up.md`

**Change**:
```diff
--- a/spec/commands/infra-up.md
+++ b/spec/commands/infra-up.md
@@ -8,11 +8,11 @@ inputs:
   flags: []
 outputs:
   exit_codes:
-    "0": "All hosts created and bootstrap succeeded"
-    "10": "Some hosts failed bootstrap (partial failure)"
-    "1": "Config error"
-    "2": "CloudProvider failure"
-    "3": "Global bootstrap error"
+    success: 0
+    partial_failure: 10
+    config_error: 1
+    cloud_provider_failure: 2
+    global_bootstrap_error: 3
```

### Step 2: Update Documentation Section

If the spec has an "Exit Codes" section in the body, update it to reference the symbolic names:

```markdown
## Exit Codes

- `0` (success) - All hosts created and bootstrap succeeded
- `10` (partial_failure) - Some hosts failed bootstrap (partial failure)
- `1` (config_error) - Config error
- `2` (cloud_provider_failure) - CloudProvider failure
- `3` (global_bootstrap_error) - Global bootstrap error
```

### Step 3: Validate

```bash
# Validate spec frontmatter
go run ./cmd/spec-validate --check-integrity

# Verify feature mapping
./bin/stagecraft gov feature-mapping

# Full validation
./scripts/run-all-checks.sh
```

All should pass without the YAML unmarshal errors.

---

## Commit Message

```
docs(CLI_INFRA_UP): fix spec frontmatter exit codes

- Change exit_codes from map[string]string to map[string]int format
- Use symbolic names (success, partial_failure, etc.) as keys
- Matches format used in other CLI command specs
- Fixes YAML unmarshal errors in spec validation
```

---

## Reference

- Spec Schema: `internal/tools/specschema/model.go` (line 39: `ExitCodes map[string]int`)
- Example: `spec/commands/build.md` (multi-exit-code pattern)
- Governance: `spec/governance/GOV_V1_CORE.md` (section 4.1)

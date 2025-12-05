---

## 📋 NEXT AGENT CONTEXT — After Completing Feature <CURRENT_FEATURE_ID>

---

## 🎉 LAYER 1: What Just Happened

### Feature Complete: <CURRENT_FEATURE_ID>

**Feature ID**: `<CURRENT_FEATURE_ID>`

**Status**: ✅ Implemented, fully tested, and merged

**PR**: #<PR_NUMBER> (<PR_URL>)

**Commit**: `<COMMIT_HASH>` - `<COMMIT_SUMMARY>`

### What Now Exists

**Package**: `<PATH/TO/PACKAGE/>`

- <Bullet: Delivered capability 1>
- <Bullet: Delivered capability 2>
- <Bullet: Delivered capability 3>
- <Bullet: Architectural guarantee 1>
- <Bullet: Architectural guarantee 2>
- <Bullet: Test coverage, benchmarks, etc.>

**APIs Available**:

```go
<API_SIGNATURE_1>
<API_SIGNATURE_2>
<API_SIGNATURE_3>
```

**Files Created**:

- `<path/to/new_file.go>`
- `<path/to/new_file_test.go>`

**Files Updated**:

- `<path/to/spec_or_doc.md>`
- `spec/features.yaml` — Marked `<CURRENT_FEATURE_ID>` as `done`

---

## 🎯 LAYER 2: Immediate Next Task

### Implement <NEXT_FEATURE_ID>

**Feature ID**: `<NEXT_FEATURE_ID>`

**Status**: `todo`

**Spec**: `<path/to/spec_for_next_feature.md>` (create if missing)

**Dependencies**:

- <DEPENDENCY_1> <status: ready/todo>
- <DEPENDENCY_2> <status: ready/todo>
- <DEPENDENCY_3> <status: optional/not required>

**⚠️ SCOPE REMINDER**: All work in this handoff MUST be scoped strictly to `<NEXT_FEATURE_ID>`. Do not modify unrelated features, files, or behavior.

**Reference Spec**: <PATH/TO/RELEVANT_SPEC_SECTION.md>

---

### 🧪 MANDATORY WORKFLOW — Tests First

**Before writing ANY implementation code**:

1. **Create test file**: `<path/to/next_feature_test.go>`

2. **Write failing tests** describing:

   - <Behavior 1>
   - <Behavior 2>
   - <Failure semantics / error propagation>
   - <Integration with CURRENT_FEATURE_ID APIs>
   - <Edge cases>

3. **Run tests** - they MUST fail

4. **Only then** begin implementation

**Test Pattern** (follow existing test patterns):

- Follow: `<reference_test_file.go>`
- Use golden tests where applicable
- Mock external dependencies
- Test sequencing / state transitions explicitly

---

### 🛠 Implementation Outline

**1. Standard Initialization Pattern**:

```go
<BOOTSTRAP_SNIPPET_EXAMPLE>
```

**2. Main Behavior / Phase Sequence** (if applicable):

<STEP_1> → <STEP_2> → <STEP_3> → <STEP_4>

**3. Failure Semantics**:

- On failure in <PHASE> → mark <PHASE> as failed
- Mark downstream phases as skipped
- Abort execution (do NOT proceed)

**4. Required Files**:

- `<path/to/next_feature.go>`
- `<path/to/next_feature_test.go>`
- `<path/to/updated_or_new_spec.md>`
- `<path/to/orchestration_layer/>` (optional)

**5. Integration Points**:

- Uses <CORE_COMPONENT_1> from `<path/to/file.go>`
- Uses <CORE_COMPONENT_2>
- Calls <API_METHOD> at integration checkpoints

---

### 🧭 CONSTRAINTS (Canonical List)

**The next agent MUST NOT**:

- ❌ Modify existing `<CURRENT_FEATURE_ID>` behavior or code
- ❌ Modify persisted formats (JSON, schemas)
- ❌ Add/rename/remove identifiers (phases, enums)
- ❌ Implement <SECONDARY_FEATURE_ID> early
- ❌ Mix multiple features in one PR
- ❌ Skip tests-first workflow
- ❌ Write directly to persisted state files (use managers instead)

**The next agent MUST**:

- ✅ Write failing tests first
- ✅ Follow established CLI/core patterns (`<reference_file.go>`)
- ✅ Use abstractions (<Manager/Service>) correctly
- ✅ Keep changes strictly scoped to `<NEXT_FEATURE_ID>`
- ✅ Update or create spec files as needed

---

## 📌 LAYER 3: Secondary Tasks

### <SECONDARY_FEATURE_ID>

**Feature ID**: `<SECONDARY_FEATURE_ID>`

**Status**: `todo`

**Dependencies**: <DEPENDENCY_LIST>

**Do NOT begin until `<NEXT_FEATURE_ID>` is complete.** (See CONSTRAINTS section)

---

### <FUTURE_DESIGN_ONLY_FEATURE_ID> (Design Only)

**Feature ID**: `<FUTURE_DESIGN_ONLY_FEATURE_ID>`

**Status**: `todo`

**Dependencies**: <DEPENDENCY_LIST>

**Do NOT implement until prerequisites are complete.** (See CONSTRAINTS section)

Design can consider:

- <Idea 1>
- <Idea 2>
- <Idea 3>

---

## 🎓 Architectural Context (For Understanding)

**Why These Design Decisions Matter**:

- **<Invariant 1>**: <Explanation>
- **<Invariant 2>**: <Explanation>
- **<Invariant 3>**: <Explanation>

**Integration Pattern Example** (for reference, not required to copy exactly):

```go
// Example: How <NEXT_FEATURE_ID> should integrate with <CURRENT_FEATURE_ID>
// This is illustrative - adapt to your implementation needs
<OPTIONAL_PATTERN_SNIPPET>
```

---

## 📝 Output Expectations

**When you complete `<NEXT_FEATURE_ID>`**:

1. **Summary**: What was implemented

2. **Commit Message** (follow this format):

```
feat(<NEXT_FEATURE_ID>): <short description>

Summary:
- <Change 1>
- <Change 2>
- <Change 3>

Files:
- <file1>
- <file2>
- ...

Test Results:
- All tests pass
- Coverage meets targets
- No lint errors

Feature: <NEXT_FEATURE_ID>
Spec: <path/to/spec.md>
```

3. **Verification**:

   - ✅ Tests were written first (before implementation)
   - ✅ No unrelated changes were made
   - ✅ Feature boundaries respected (only `<NEXT_FEATURE_ID>` code)
   - ✅ All checks pass (`./scripts/run-all-checks.sh`)

---

## ⚡ Quick Start for Next Agent

**Bootloader Instructions**:

1. **Load Context**:

   - Read `<path/to/core/file.go>` to understand API
   - Read `<path/to/spec.md>` for semantics
   - Read `<reference_cli_or_test.go>` for pattern reference
   - Check if `<path/to/next_feature_spec.md>` exists

2. **Begin Work**:

   - Feature ID: `<NEXT_FEATURE_ID>`
   - Create feature branch: `feature/<NEXT_FEATURE_ID>`
   - Start with tests: `<path/to/next_feature_test.go>`
   - Write failing tests first
   - Then implement: `<path/to/next_feature.go>`

3. **Follow Semantics**:

   - Use existing <identifiers/phases> (see CONSTRAINTS section)
   - Follow order: <SEQUENCE>
   - Implement failure semantics as specified

4. **Respect Constraints**:

   - See CONSTRAINTS section (canonical list)
   - Do not modify `<CURRENT_FEATURE_ID>`
   - Do not implement <SECONDARY_FEATURE_ID> yet
   - Keep feature boundaries clean

---

## ✅ Final Checklist

Before starting work:

- [ ] Feature ID identified: `<NEXT_FEATURE_ID>`
- [ ] Git hooks verified
- [ ] Working directory clean
- [ ] On feature branch: `feature/<NEXT_FEATURE_ID>`
- [ ] Spec located/created: `<path/to/next_feature_spec.md>`
- [ ] Tests written first: `<path/to/next_feature_test.go>`
- [ ] Tests fail (as expected)
- [ ] Ready to implement

---

**Copy this entire document into your next agent session to continue development.**

This document is optimized for deterministic AI handoff and aligns with Stagecraft's Agent.md principles (spec-first, test-first, feature-bounded, deterministic).


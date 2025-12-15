# <FEATURE_ID> - Coverage Strategy (<STATUS_LABEL>)

> Template for provider coverage strategies. Copy this file to:
> `internal/providers/<kind>/<name>/COVERAGE_STRATEGY.md` and replace placeholders.

This document defines the coverage approach for the `<FEATURE_ID>` provider.
It MUST describe how v1 coverage is achieved (or will be achieved) in a deterministic, AATSE-aligned way.

---

## 🎯 Coverage Goals

The provider must:

1. Exercise all critical execution paths:
   - Happy path(s)
   - Error paths (including external dependency failures)
   - Cancellation / shutdown paths
2. Avoid test flakiness caused by:
   - OS-level behavior (pipe buffering, process scheduling)
   - Unbounded timing assumptions (raw `time.Sleep`)
   - Racy goroutine orchestration
3. Keep provider tests:
   - Deterministic
   - Side-effect-minimized
   - Aligned with AATSE "no broken glass" principles

If previous tests relied on fragile seams or OS-dependent behavior, this strategy MUST explain how the v1 design avoids those patterns.

---

## ✔️ V1 Coverage Status - <STATUS_LABEL>

**Overall Coverage: <COVERAGE_PERCENT>%** (target: ≥ 80% for providers)

> Replace this table with coverage pulled from `go test -cover ./internal/providers/<kind>/<name>`.

| Function                 | Coverage | Status       |
|--------------------------|----------|--------------|
| `<FuncName>`             | xx.x%    | ✅ / ⚠ / ❌   |
| `<FuncName>`             | xx.x%    | ✅ / ⚠ / ❌   |
| `<FuncName>`             | xx.x%    | ✅ / ⚠ / ❌   |

**Status legend:**

- ✅ Complete - all branches and edge cases covered
- ⚠ Partial - major path covered, some edge/error branches pending
- ❌ Missing - not yet covered, MUST be addressed before v1 is truly complete

---

## What Changed in V1

> Describe the refactor/testing approach that makes this provider deterministic and testable.
> Draw inspiration from `PROVIDER_FRONTEND_GENERIC` but keep it provider-specific.

### 1. Extracted Deterministic Primitives (if applicable)

If the provider includes complex logic (like streaming, retries, or concurrency), describe the primitives you extracted.

Examples:

- `scanStream()`-style helpers
- Connection/handshake helpers
- Retry or backoff helpers

For each primitive:

- Define its input/output shape (interfaces, channels, etc.)
- Confirm it is synchronous/pure enough to test deterministically in isolation.

### 2. Replaced Flaky Tests

List any removed/replaced tests that previously caused flakiness.

- Integration tests relying on OS pipe buffering
- Tests using raw `time.Sleep`
- Tests depending on non-deterministic ordering or scheduling
- Tests that required awkward global seams (like `newScanner`)

For each, briefly state:

- What was removed
- What deterministic tests now cover the same behavior

### 3. Integration Coverage Scope

Describe what integration tests still cover, and what they intentionally avoid.

Integration tests SHOULD:

- Validate provider orchestration with external tools/processes
- Assert correct behavior across lifecycle (startup → ready → shutdown / error)
- Focus on *wiring* and *semantics*, not low-level primitives already covered in unit tests

Integration tests SHOULD NOT:

- Re-test pure helper logic covered by unit tests
- Depend on arbitrary sleep durations
- Depend on OS-specific behavior where avoidable

---

## 🧪 Coverage Philosophy: AATSE + "No Broken Glass"

The provider MUST follow Stagecraft's test philosophy:

### Deterministic Primitives

- Critical logic extracted into pure/synchronous helpers where possible.
- These helpers are unit-tested and, where useful, benchmarked.

### Predictable Orchestration

- Higher-level functions (commands, runners, orchestrators) coordinate deterministic primitives.
- They are covered by:
  - Unit tests for small orchestration branches
  - Integration tests for end-to-end behavior

### Separation of Concerns

- Pure logic tested in isolation.
- IO and external interactions tested via clear, well-bounded integration tests.

### Minimal Concurrency Surface

- Concurrency is introduced only where necessary.
- Any concurrent behavior is synchronized using channels, `sync.WaitGroup`, or other deterministic mechanisms.
- Tests do not depend on raw sleeps to "wait long enough".

### No Fragile Legal Fictions

- Avoid fake states that cannot occur in real executions.
- When simulating error states, do so through realistic fake implementations or readers/writers, not impossible invariants.

> AATSE principle:
> **If you need a seam solely for testing, the design is probably incomplete.**

---

## 📈 Future Expansions (Post-V1)

Optional future enhancements (non-blocking for v1):

- Additional structured logging or metrics tests
- Extended configuration or edge-case scenarios
- More exhaustive error simulations for external integrations

These MUST be documented as "post-v1" so they don't blur the v1 completion criteria.

---

## ✅ Conclusion

**<FEATURE_ID> coverage status: <STATUS_LABEL>.**

All major branches, edge cases, and lifecycle transitions SHOULD be validated through deterministic tests that align with:

- Stagecraft governance
- GOV_CORE
- `docs/governance/GOV_V1_TEST_REQUIREMENTS.md`
- AATSE design standards

If `<STATUS_LABEL>` is:

- `V1 Complete`:
  - This document MUST be stable and updated only when behavior changes.
  - A matching status file MUST exist at:
    - `docs/engine/status/<FEATURE_ID>_COVERAGE_V1_COMPLETE.md`

- `V1 Plan` / `In Progress`:
  - This document MUST include a short list of remaining gaps and their priority.
  - The status file MAY be named:
    - `docs/engine/status/<FEATURE_ID>_COVERAGE_PLAN.md` or similar.

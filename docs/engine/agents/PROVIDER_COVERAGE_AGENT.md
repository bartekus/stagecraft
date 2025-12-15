# PROVIDER_COVERAGE_AGENT

Role: Provider Coverage Improvement Agent  
Scope: Bringing provider implementations from "V1 Plan" to "V1 Complete" coverage status.

This agent helps systematically improve provider test coverage to meet GOV_CORE and GOV_V1_TEST_REQUIREMENTS standards.

---

## 1. Purpose

The PROVIDER_COVERAGE_AGENT exists to:

1. Identify providers with incomplete coverage (marked as "V1 Plan" or missing `COVERAGE_STRATEGY.md`).
2. Systematically improve coverage using the deterministic test patterns established by `PROVIDER_FRONTEND_GENERIC`.
3. Update coverage strategy documents and status tracking as work progresses.
4. Ensure all providers meet the ≥80% coverage target with deterministic, AATSE-aligned tests.

---

## 2. Primary Inputs

The agent treats these files as sources of truth:

1. **Provider Coverage Strategy Files**
   - `internal/providers/<kind>/<name>/COVERAGE_STRATEGY.md`
   - Status labels: "V1 Plan", "In Progress", "V1 Complete"

2. **Template**
   - `docs/coverage/PROVIDER_COVERAGE_TEMPLATE.md` - Template for new coverage strategies

3. **Reference Implementation**
   - `internal/providers/frontend/generic/COVERAGE_STRATEGY.md` - Canonical example of V1 Complete coverage
   - `docs/engine/status/PROVIDER_FRONTEND_GENERIC_COVERAGE_V1_COMPLETE.md` - Completion tracking example

4. **Governance Requirements**
   - `docs/governance/GOV_V1_TEST_REQUIREMENTS.md` - Test strategy standards
   - `spec/features.yaml` - Feature status and test file references

5. **Current Coverage Metrics**
   - Run: `go test -cover ./internal/providers/<kind>/<name>`
   - Parse coverage output to identify gaps

---

## 3. Operating Mode

### 3.1 High Level Rules

The agent must:

- Work on **one provider at a time**.
- Follow the **PROVIDER_FRONTEND_GENERIC pattern** for deterministic test design.
- **Extract pure helpers** before writing tests (AATSE principle).
- **Remove flaky patterns** (time.Sleep, uncontrolled goroutines, OS dependencies).
- Update coverage strategy documents **as work progresses**.

The agent should always tell the user:

- Which provider it is working on.
- Current coverage percentage and target.
- Which functions/paths need coverage.
- What deterministic helpers it will extract (if any).
- What tests it will add or refactor.

### 3.2 Workflow for Each Provider

#### Step 1: Assessment

1. **Check for existing coverage strategy:**
   - If missing: Create from template
   - If exists: Read current status and gaps

2. **Run coverage analysis:**
   ```bash
   go test -cover ./internal/providers/<kind>/<name>
   ```

3. **Identify coverage gaps:**
   - Functions below 80% coverage
   - Missing error paths
   - Flaky test patterns (grep for `time.Sleep`, uncontrolled goroutines)

4. **Review existing tests:**
   - Identify flaky patterns
   - Find opportunities for deterministic helper extraction

#### Step 2: Design Deterministic Helpers (if needed)

Following PROVIDER_FRONTEND_GENERIC pattern:

1. **Identify complex logic** that could be extracted:
   - Streaming/scanning logic → `scanStream()`-style helpers
   - Command building → `buildArgs()`-style helpers
   - Output parsing → `parseOutput()`-style helpers

2. **Extract pure functions:**
   - Accept explicit inputs (config, readers, writers)
   - Return explicit outputs (results, errors, channels)
   - No side effects, no OS dependencies

3. **Write unit tests for helpers:**
   - Test all branches deterministically
   - Use in-memory readers/writers
   - No subprocesses, no sleeps

#### Step 3: Refactor Integration Tests

1. **Remove flaky patterns:**
   - Replace `time.Sleep` with deterministic synchronization
   - Replace uncontrolled goroutines with channels/WaitGroups
   - Remove OS-dependent assumptions where possible

2. **Focus integration tests on orchestration:**
   - Process lifecycle (start, ready, shutdown)
   - Error handling at orchestration level
   - **Do NOT** re-test pure helper logic

#### Step 4: Add Missing Coverage

1. **Error paths:**
   - Invalid configuration
   - Missing dependencies
   - External failures (network, binaries, etc.)

2. **Edge cases:**
   - Empty inputs
   - Boundary conditions
   - Cancellation/timeout scenarios

3. **Verify coverage:**
   ```bash
   go test -cover ./internal/providers/<kind>/<name>
   ```

#### Step 5: Update Documentation

1. **Update `COVERAGE_STRATEGY.md`:**
   - Change status to "V1 Complete" (if target met)
   - Update coverage table with final metrics
   - Document extracted helpers and test approach
   - Remove "plan" language

2. **Create status document** (if V1 Complete):
   - `docs/engine/status/<FEATURE_ID>_COVERAGE_V1_COMPLETE.md`
   - Follow `PROVIDER_FRONTEND_GENERIC_COVERAGE_V1_COMPLETE.md` pattern

3. **Verify governance:**
   ```bash
   ./scripts/check-provider-governance.sh
   ```

---

## 4. Allowed and Forbidden Actions

### 4.1 Allowed

The agent **may**:

- Extract pure helper functions from provider implementations.
- Add unit tests for extracted helpers.
- Refactor flaky integration tests to be deterministic.
- Add missing error path and edge case tests.
- Update coverage strategy documents.
- Create status tracking documents.

### 4.2 Forbidden (without explicit user request)

The agent must **not**:

- Change provider behavior or interfaces.
- Remove working tests without replacement.
- Introduce new dependencies or abstractions beyond pure helpers.
- Skip coverage for "hard to test" paths (find a way to test them deterministically).
- Mark coverage as "V1 Complete" without meeting ≥80% target.

---

## 5. Validation Commands

After completing work on a provider, run:

### 5.1 Coverage Verification

```bash
# Check coverage percentage
go test -cover ./internal/providers/<kind>/<name>

# Verify no race conditions
go test -race ./internal/providers/<kind>/<name>

# Verify no flakiness
go test -count=20 ./internal/providers/<kind>/<name>
```

### 5.2 Governance Validation

```bash
# Verify coverage strategy is valid
./scripts/check-provider-governance.sh

# Full project checks
./scripts/run-all-checks.sh
```

---

## 6. Example Interaction Pattern

A typical agent run for a provider:

1. **Scope**
   - Provider: `PROVIDER_NETWORK_TAILSCALE`
   - Current: 68.2% coverage (V1 Plan)
   - Target: ≥80% coverage (V1 Complete)

2. **Assessment**
   - Coverage strategy exists: ✅
   - Current gaps: Error paths in `EnsureInstalled`, missing tests for `parseStatus`
   - Flaky patterns: None identified

3. **Plan**
   - Extract `buildTailscaleUpArgs(config)` helper
   - Extract `parseTailscaleStatus(output)` helper
   - Add unit tests for helpers
   - Add error path tests for `EnsureInstalled`
   - Add tests for `parseStatus` edge cases

4. **Execution**
   - Extract helpers
   - Write unit tests
   - Add missing coverage
   - Verify: `go test -cover` shows ≥80%

5. **Documentation**
   - Update `COVERAGE_STRATEGY.md` to "V1 Complete"
   - Create `PROVIDER_NETWORK_TAILSCALE_COVERAGE_V1_COMPLETE.md`
   - Run governance checks

6. **Suggested commit message**
   - `test(PROVIDER_NETWORK_TAILSCALE): achieve v1 coverage (68.2% → 82.5%)`

---

## 7. Provider Priority Order

When multiple providers need coverage work, prioritize:

1. **Providers marked `done` in `spec/features.yaml`** without coverage strategies
2. **Providers with "V1 Plan" status** (already started)
3. **Providers closest to 80%** (quick wins)
4. **Providers with flaky tests** (stability first)

Current provider status (from `spec/features.yaml`):

- ✅ `PROVIDER_FRONTEND_GENERIC` - V1 Complete (reference)
- 🔄 `PROVIDER_NETWORK_TAILSCALE` - V1 Plan (in progress)
- ⏳ `PROVIDER_BACKEND_ENCORE` - done, needs coverage strategy
- ⏳ `PROVIDER_BACKEND_GENERIC` - done, needs coverage strategy
- ⏳ `PROVIDER_CLOUD_DO` - done, needs coverage strategy

---

## 8. Reference Patterns

### 8.1 Extracting Pure Helpers

**Before (flaky, hard to test):**
```go
func (p *Provider) runWithReadyPattern(ctx context.Context, cmd *exec.Cmd) error {
    // Inline scanner logic mixed with orchestration
    scanner := bufio.NewScanner(stdout)
    for scanner.Scan() {
        if pattern.MatchString(scanner.Text()) {
            return nil
        }
    }
    // ...
}
```

**After (deterministic, testable):**
```go
// Pure helper (unit testable)
func scanStream(r io.Reader, w io.Writer, pattern *regexp.Regexp) (<-chan struct{}, <-chan error) {
    // Pure logic, no side effects
}

// Orchestration (integration testable)
func (p *Provider) runWithReadyPattern(ctx context.Context, cmd *exec.Cmd) error {
    readyCh, errCh := scanStream(stdout, os.Stdout, pattern)
    // Orchestrate channels
}
```

### 8.2 Replacing Flaky Patterns

**Before:**
```go
time.Sleep(10 * time.Millisecond)
if !ready {
    t.Fatal("not ready")
}
```

**After:**
```go
select {
case <-readyCh:
    // proceed
case <-time.After(timeout):
    t.Fatal("timeout")
}
```

---

## 9. Success Criteria

A provider is "V1 Complete" when:

- ✅ Coverage ≥80% (verified via `go test -cover`)
- ✅ All tests pass with `-race` (no race conditions)
- ✅ All tests pass with `-count=20` (no flakiness)
- ✅ No `time.Sleep` in tests (except documented integration scenarios)
- ✅ No uncontrolled goroutines in tests
- ✅ Pure helpers extracted and unit tested
- ✅ Integration tests focus on orchestration, not logic
- ✅ `COVERAGE_STRATEGY.md` marked "V1 Complete"
- ✅ Status document created at `docs/engine/status/<FEATURE_ID>_COVERAGE_V1_COMPLETE.md`
- ✅ Governance checks pass: `./scripts/check-provider-governance.sh`

---

**End of PROVIDER_COVERAGE_AGENT specification.**

# CORE_STATE_TEST_ISOLATION - State Test Isolation for CLI Commands

- **Feature ID**: `CORE_STATE_TEST_ISOLATION`
- **Status**: done
- **Owner**: bart
- **Depends on**: `CORE_STATE`
- **Related**:
  - `CLI_PHASE_EXECUTION_COMMON`
  - `CORE_STATE_CONSISTENCY` (future)

## 1. Purpose

CLI commands that read or write release state must be testable in a deterministic, isolated way.

This feature defines a mandatory test pattern for any test that touches the Stagecraft state file so that:

- Tests never share state files or directories.
- Test failures are never caused by cross-test contamination.
- The same test suite can be run repeatedly or in parallel without flakiness caused by shared state.

This spec is about **test isolation**, not the state file semantics themselves (those are defined in `spec/core/state.md`).

## 2. Scope

**In scope:**

- Isolation rules for tests that touch `.stagecraft/releases.json` (or other state files managed by `state.Manager`).
- A standard helper for setting up isolated state environments in CLI tests.
- Rules for using `STAGECRAFT_STATE_FILE` in tests.
- Requirements for future tests that interact with state.

**Out of scope:**

- File format of the state file.
- State consistency and durability semantics (covered by `CORE_STATE` and `CORE_STATE_CONSISTENCY`).
- Behaviour of individual CLI commands beyond their interaction with state.

## 3. State Test Isolation Model

### 3.1 Isolation Constraints

All tests that interact with state MUST satisfy:

1. **Unique filesystem root per test**
   - Each test gets its own temporary directory.
   - The directory is not shared across tests, even within the same file.
   - No test may rely on any files outside its own temporary root except:
     - Go toolchain,
     - System binaries,
     - Read-only fixtures explicitly allowed by the test.

2. **Unique state file per test**
   - Each test has its own state file path (e.g., `<tmpdir>/.stagecraft/releases.json`).
   - The state file path is either:
     - Generated inside the helper, or
     - Explicitly derived from the helper's temp directory.

3. **Test-scoped environment variables**
   - `STAGECRAFT_STATE_FILE` MUST be set via `t.Setenv` inside each test.
   - The env var MUST point to an **absolute** path to avoid CWD-related surprises.
   - No test may rely on the process-level default path without explicit justification.

4. **Working directory safety**
   - If a test needs to change the working directory, it MUST:
     - Change to the test's temp directory, and
     - Restore the previous working directory via `t.Cleanup`.
   - Tests MUST NOT assume a particular working directory at start time.

5. **No hidden global state**
   - Tests MUST NOT mutate global package-level variables that affect state locations.
   - The only supported mechanism for state location override in tests is `STAGECRAFT_STATE_FILE`.

### 3.2 Helper Function: `setupIsolatedStateTestEnv`

A standard helper lives in `internal/cli/commands/test_helpers.go`:

```go
// setupIsolatedStateTestEnv creates a per-test isolated filesystem root
// and state manager.
//
// It:
//   - Creates a temporary directory
//   - Creates a .stagecraft/releases.json path inside it
//   - Sets STAGECRAFT_STATE_FILE to an absolute path (via t.Setenv)
//   - Changes the working directory to the temp dir
//   - Registers t.Cleanup to restore the working directory
//   - Returns a state.Manager bound to the isolated state file
func setupIsolatedStateTestEnv(t *testing.T) *isolatedStateTestEnv { ... }
```

**Behaviour requirements:**

- The temporary directory MUST be created with `t.TempDir()`.
- The state file path MUST be absolute.
- The function MUST call `t.Helper()` at the top.
- The helper MUST NOT panic; any setup error MUST cause `t.Fatalf` or equivalent.
- The helper MUST NOT depend on external environment variables.

### 3.3 Usage Rules

All CLI tests that touch state (directly or indirectly) MUST:

1. Call `setupIsolatedStateTestEnv(t)` at the beginning of the test.
2. Use the returned `state.Manager` when direct state access is required.
3. Avoid constructing `state.Manager` manually, unless the test is explicitly about testing `state.NewManager(path)`.

**Examples** (logical, not exact code):

```go
func TestDeployCommand_RecordsRelease(t *testing.T) {
    env := setupIsolatedStateTestEnv(t)
    // Run the command under test...
    cmd := newDeployCommand()
    // ...
    // Assert using env.Manager (bound to isolated state file)
    releases, err := env.Manager.ListReleases(context.Background(), "dev")
    require.NoError(t, err)
    // ...
}
```

### 3.4 Prohibited Patterns

The following patterns are explicitly disallowed in state-touching tests:

- Using a shared, hard-coded path for the state file (e.g., `/tmp/stagecraft-releases.json`).
- Using process-global env var setup for `STAGECRAFT_STATE_FILE` outside of test helpers.
- Mixing `state.NewDefaultManager()` and `state.NewManager(customPath)` in an ad hoc way without isolation.
- Relying on the order tests run in to determine state content.

## 4. CLI Integration

This spec applies to tests in:

- `internal/cli/commands/deploy_test.go`
- `internal/cli/commands/rollback_test.go`
- `internal/cli/commands/releases_test.go`
- Any future tests under `internal/cli/commands/` that:
  - Run commands which mutate or read release history, or
  - Directly use `state.Manager`.

Future CLI tests must adopt `setupIsolatedStateTestEnv` instead of re-implementing their own temp file helpers.

## 5. Non-goals

- Allowing tests to share state as a way to simulate multi-command workflows.
  - Multi-step workflows must be exercised within a single test that uses its own isolated state.
- Supporting test patterns that depend on the real default state location.

## 6. Testing Strategy

- **Unit tests for `setupIsolatedStateTestEnv` itself:**
  - Confirms that:
    - `STAGECRAFT_STATE_FILE` is set.
    - The path is absolute.
    - The working directory is changed into the temp dir.
    - The working directory is restored on cleanup.
- **CLI tests:**
  - All state-touching CLI tests MUST:
    - Use the helper.
    - Assert end-to-end semantics (e.g., that deploy and rollback record releases in the isolated state file).

## 7. Acceptance Criteria

`CORE_STATE_TEST_ISOLATION` is considered done when:

- All state-touching CLI tests use `setupIsolatedStateTestEnv`.
- No tests rely on shared state files.
- `STAGECRAFT_STATE_FILE` is documented in `spec/core/state.md` and respected by `state.NewDefaultManager()`.
- The full test suite runs repeatedly without cross-test contamination caused by shared state files.
- The helper itself is covered by tests.


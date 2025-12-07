---
feature: CORE_STATE_CONSISTENCY
version: v1
status: done
domain: core
inputs:
  flags: []
outputs:
  exit_codes: {}
---
# CORE_STATE_CONSISTENCY - State Durability and Read-after-write Guarantees

- **Feature ID**: `CORE_STATE_CONSISTENCY`
- **Status**: done
- **Owner**: bart
- **Depends on**:
  - `CORE_STATE`
  - `CORE_STATE_TEST_ISOLATION`
- **Related**:
  - `CLI_PHASE_EXECUTION_COMMON`
  - `CLI_ROLLBACK`
  - `CLI_RELEASES`

## 1. Purpose

Stagecraft stores release history and phase status in a local state file (e.g., `.stagecraft/releases.json`).

This feature defines explicit **consistency and durability guarantees** for state writes and reads, so that:

- After a command completes successfully, subsequent reads see the state that was just written.
- Commands and tests do not observe "stale" or partially updated state due to OS buffering or rename semantics.
- Rollback and phase execution behave deterministically, even when they perform multiple sequential updates.

This spec is focused on **file system semantics**, not on the business meaning of releases or phases.

## 2. Scope

**In scope:**

- The behaviour of `saveState` and related write paths in `internal/core/state`.
- The interaction between:
  - Commands that update state (deploy, rollback, releases),
  - State manager APIs (`ListReleases`, `UpdatePhase`, etc),
  - The underlying filesystem.
- Guarantees that apply inside a single process on a single host.

**Out of scope (v1):**

- Cross-process or distributed locking.
- Remote state backends (S3, DB, etc).
- Multi-host consistency.
- Crash consistency beyond best-effort fsync semantics.

## 3. Consistency Model

### 3.1 Read-after-write Guarantee (Single Process)

Within a single process:

- If a state-modifying operation:
  - Successfully returns `nil` error to the caller, and
  - Is followed by a state-reading operation in the same process,

then:

> The read operation MUST observe the state produced by the last successful write, provided it reads from the same state file path.

Formally, for any sequence within one process:

```text
1. mgr := NewManager(path) or NewDefaultManager()
2. op1 := mgr.UpdatePhase(...)
3. op2 := mgr.UpdatePhase(...)
4. err := mgr.saveState(...)
5. err == nil
6. snapshot, err2 := mgr.ListReleases(...)
```

`snapshot` MUST reflect the results of both `op1` and `op2`.

### 3.2 Multi-manager Behaviour

If multiple `state.Manager` instances in the same process point at the same state file:

- After a successful write, any subsequent read through any manager instance that uses the same file path MUST see the updated state.

**Example:**

```text
1. mgr1 := NewDefaultManager()          // uses path P
2. mgr1.UpdatePhase(...); mgr1.saveState()
3. mgr2 := NewManager(P)
4. snapshot, err := mgr2.ListReleases(...)
```

`snapshot` MUST include phase updates committed by `mgr1`.

### 3.3 Atomicity

`CORE_STATE` already defines atomic write via "write to temp file then rename". `CORE_STATE_CONSISTENCY` tightens this with explicit sync semantics:

- Writes MUST be performed as:
  1. Open a new temp file in the same directory as the target.
  2. Encode the full state into the temp file.
  3. Flush data to disk with `file.Sync()`.
  4. Close the temp file.
  5. Atomically rename the temp file to the target path.
  6. Sync the directory containing the target file (best effort; failures are ignored).

- At no point should a partial file be visible at the final path.

Directory sync is best-effort and non-fatal. Failures do not cause `saveState` to return an error, as many filesystems either do not support directory sync or expose platform-specific behavior.

## 4. Implementation Requirements

### 4.1 saveState Write Protocol

In `internal/core/state/state.go` (or equivalent), `saveState` MUST follow this protocol:

1. **Prepare destination**
   - Compute the directory `d = filepath.Dir(path)` and base filename.

2. **Create temp file**
   - Use `os.CreateTemp(d, ".releases-*.tmp")` or equivalent.
   - Ensure the temp file lives in the same directory as the final target.

3. **Write state**
   - Encode the complete state into the temp file.
   - After writing:
     - Call `tempFile.Sync()` and check for errors.
     - Close the file and check for errors.

4. **Atomic rename**
   - Call `os.Rename(tempPath, finalPath)` and check for errors.

5. **Directory sync**
   - After `os.Rename(tempPath, finalPath)`, Stagecraft attempts a best-effort `Sync()` on the containing directory.
   - Open the directory `d` with `os.Open(d)`.
   - Call `dirFile.Sync()` and check for errors.
   - Close the directory file.
   - Directory sync failures do not cause `saveState` to return an error. They are treated as non-fatal because many filesystems either do not support directory sync or expose platform-specific behavior. A successful `saveState` return means:
     - The state file was fully written and `Sync()`ed at the file level.
     - The file was atomically renamed into place.
     - A directory sync was attempted; failures are ignored.

6. **Deterministic error handling**
   - If any step fails:
     - Clean up temporary files when safe to do so.
     - Wrap errors with context and return them.
     - Under no circumstances should the function return `nil` if the write path did not complete successfully.

### 4.2 loadState and Read Behaviour

- `loadState` (and any read operation) MUST:
  - Open the state file at the final path.
  - Read and decode the full contents.
  - Not cache results across calls in a way that hides updates made by other managers in the same process.

Caching is allowed only if:

- The cache is invalidated whenever a write occurs through the same manager, and
- New `state.Manager` instances always perform a fresh read.

### 4.3 Rollback and Phase Execution

Rollback and phase execution must be structured so that:

- Phase updates for a single release are applied in a well-defined order.
- After the rollback command returns successfully:
  - A subsequent call to `ListReleases` observes all completed phases for the rollback release.

For example, for `TestRollbackCommand_SuccessfulRollback_AllPhasesCompleted`:

- The test may:
  - Run the rollback command.
  - Construct a fresh `state.Manager` pointing at the same state file.
  - Call `ListReleases` and identify the rollback release.
- It MUST see `PhaseStatus == completed` for all rollback phases.

If the test runs immediately after the command, the guarantees from 4.1 and 4.2 MUST be sufficient for consistency without artificial sleeps.

## 5. Testing Strategy

### 5.1 Unit Tests for saveState and loadState

- Tests MUST verify:
  - The temp file is created in the same directory.
  - `Sync()` is called on the file.
  - `os.Rename` is used for finalization.
  - The directory is synced after rename (where supported).
- Where direct verification of `Sync` is hard, use:
  - Fakes or wrappers around `os` primitives, or
  - OS-specific tests when possible.

### 5.2 Behavioural Tests

- Tests that simulate:

1. **Single manager**
   - Create manager, perform multiple updates, call `saveState`, then `ListReleases`.
   - Assert that the read reflects the final state.

2. **Multiple managers**
   - `mgr1` writes state, `mgr2` reads it.
   - `mgr2` writes state, `mgr1` reads updated state.
   - No stale snapshots allowed after successful writes.

3. **Rollback end-to-end**
   - Reproduce `TestRollbackCommand_SuccessfulRollback_AllPhasesCompleted` under isolated conditions.
   - Run the test multiple times (e.g., with `-count 25`) to ensure stability.

### 5.3 Regression Tests

- Add a focused regression test that previously would have exhibited the "phases completed but read back as pending" behaviour, and assert that it no longer occurs.

## 6. Non-goals

- Providing strict guarantees under system crash or sudden power loss.
- Providing cross-host or distributed consistency guarantees.
- Implementing explicit file locks or FS-level locking.

These can be addressed in future state backends or higher-level features.

## 7. Acceptance Criteria

`CORE_STATE_CONSISTENCY` is considered done when:

- `saveState` implements the fsync + directory sync protocol described above.
- Tests validate read-after-write consistency for:
  - Single manager,
  - Multiple managers,
  - Rollback commands.
- `TestRollbackCommand_SuccessfulRollback_AllPhasesCompleted` passes reliably when run:
  - Alone, and
  - As part of the full suite, repeated multiple times.
- No intermittent "pending vs completed" phase mismatches are observed due to filesystem behaviour.


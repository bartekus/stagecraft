---
status: canonical
scope: v1
---

# Commit Message Format Analysis

## Issue Summary

Commit messages were not following the required format:
```
<type>(<FEATURE_ID>): <summary>
```

## Why This Matters

Commit messages are a critical link in Stagecraft's deterministic traceability chain:

**spec → tests → code → docs → commit → PR**

Without proper commit message format:

- **Traceability breaks**: Feature IDs in commits enable automated linking between specs, code, tests, and PRs
- **Feature lifecycle integrity fails**: Single-feature change enforcement requires consistent Feature ID usage
- **Review quality degrades**: Ambiguous or missing Feature IDs make it impossible to verify PR scope
- **Automation fails**: CI/CD and tooling depend on structured commit messages for feature tracking

Commit messages are not just documentation—they are **deterministic artifacts** that guarantee the integrity of Stagecraft's engineering discipline.

## Root Cause

1. **Git hooks not installed**: The commit-msg hook in `.hooks/commit-msg` was not installed in `.git/hooks/`
2. **Hook bypass**: Even when hooks exist, they can be bypassed with `STAGECRAFT_SKIP_HOOKS=1` or `SKIP_HOOKS=1`
3. **AI not checking format**: When generating commit messages, the format requirement wasn't being validated before committing

## Required Format

From `Agent.md` and `.hooks/commit-msg`:

```
<type>(<FEATURE_ID>): <summary>
```

Where:
- `<type>`: One of `feat`, `fix`, `refactor`, `docs`, `test`, `ci`, `chore`
- `<FEATURE_ID>`: SCREAMING_SNAKE_CASE (e.g., `PROVIDER_FRONTEND_GENERIC`)
- `<summary>`: Short description (≤72 characters, no trailing period)

**Deterministic Style Constraints:**
- No emojis or unicode decorations
- No auto-formatting by git clients
- Literal, precise, minimal descriptions
- ASCII-only characters in subject line

## Examples

✅ **Valid:**
- `feat(PROVIDER_FRONTEND_GENERIC): implement provider`
- `fix(PROVIDER_FRONTEND_GENERIC): address review feedback`
- `docs(PROVIDER_FRONTEND_GENERIC): sync roadmap docs`

❌ **Invalid:**
- `feat: implement PROVIDER_FRONTEND_GENERIC` (missing parentheses)
- `fix: address linter errors` (missing FEATURE_ID)
- `docs: sync roadmap and status docs with spec/features.yaml` (missing FEATURE_ID, too long)
- `feat(CLI_DEPLOY) update deploy command` (missing colon after parentheses)
- `Feat(PROVIDER_FRONTEND_GENERIC): implement provider` (type must be lowercase)
- `feat(PROVIDER_FRONTEND_GENERIC): Implement provider.` (capital letter after type, trailing period)
- `feat(CLI_PLAN, CLI_DEPLOY): refactor planning and deployment` (multiple Feature IDs - Stagecraft forbids multiple Feature IDs per commit. Each commit must map to exactly one Feature ID and one PR)

## Prevention Strategy

### 1. Always Install Hooks
```bash
./scripts/install-hooks.sh
```

### 2. Verify Hook Installation
```bash
ls -la .git/hooks/commit-msg
# Should show a symlink to .hooks/commit-msg
```

### 3. Test Hook Before Committing
```bash
echo "test" > /tmp/test_msg
.git/hooks/commit-msg /tmp/test_msg
# Should fail with format error
```

### 4. AI Workflow (Mandatory Rules)

**AI MUST perform these steps before committing:**

1. **Verify `.git/hooks/commit-msg` exists**
   - If missing → run `./scripts/install-hooks.sh`
   - If installation fails → STOP and report error

2. **Validate commit message against required pattern**
   - Format MUST be: `<type>(<FEATURE_ID>): <summary>`
   - Type MUST be lowercase: `feat`, `fix`, `refactor`, `docs`, `test`, `ci`, `chore`
   - Feature ID MUST be SCREAMING_SNAKE_CASE
   - Summary MUST be ≤72 characters
   - Summary MUST NOT have trailing period
   - Summary MUST NOT start with capital letter after type

3. **Verify FEATURE_ID matches the active feature branch**
   - Check current branch: `git branch --show-current`
   - Extract FEATURE_ID from branch name
   - Ensure commit message FEATURE_ID matches branch FEATURE_ID
   - If mismatch → STOP and report

4. **Verify no protected files are touched**
   - Protected files: LICENSE, README.md, ADRs, NOTICE
   - If protected files modified → STOP and report

5. **Run all CI checks**
   - Execute: `./scripts/run-all-checks.sh`
   - All checks MUST pass before committing
   - If any check fails → STOP, fix issues, re-run

6. **Only then create commit message and commit**
   - Generate message following format
   - Execute: `git commit -m "<message>"`
   - Verify commit succeeded

**If any check fails: STOP and report.**

### 5. Feature Lifecycle Integration

Commit messages are a required link in the feature lifecycle:

- **Feature ID validation**: Commit message FEATURE_ID must match the feature branch and spec
- **Single-feature change enforcement**: Each commit must reference exactly one FEATURE_ID
- **PR workflow integration**: Branch naming and commit messages must align:
  - Branch: `feature/<FEATURE_ID>-short-desc`
  - Commit: `<type>(<FEATURE_ID>): <summary>`
- **Spec traceability**: Commit messages enable automated linking to:
  - `spec/features.yaml` (feature definition)
  - `spec/<domain>/<feature>.md` (spec file)
  - `docs/engine/analysis/<FEATURE_ID>.md` (analysis brief)
  - Related test files

**AI MUST ensure commit messages maintain this traceability chain.**

### 6. Pre-Commit Checklist
Before committing:
- [ ] Hook is installed (`ls .git/hooks/commit-msg`)
- [ ] Message follows format: `<type>(<FEATURE_ID>): <summary>`
- [ ] FEATURE_ID matches current feature branch
- [ ] FEATURE_ID matches spec/features.yaml entry
- [ ] Subject ≤72 characters
- [ ] No trailing period
- [ ] No emojis or unicode decorations
- [ ] Type is lowercase
- [ ] Summary starts with lowercase (after colon)
- [ ] All CI checks pass (`./scripts/run-all-checks.sh`)

## Fixed Commits

The following commits were rewritten to follow the correct format:

1. ✅ `feat(PROVIDER_FRONTEND_GENERIC): implement provider`
2. ✅ `fix(PROVIDER_FRONTEND_GENERIC): address review feedback` (2 commits)
3. ✅ `fix(PROVIDER_FRONTEND_GENERIC): address linter errors and coverage`
4. ✅ `docs(PROVIDER_FRONTEND_GENERIC): sync roadmap docs`

**All commits in feature branch are now valid!**

## Recommendations

1. **Add hook installation to CI/CD**: Ensure hooks are installed in CI
2. **Document in Agent.md**: Emphasize the commit message format requirement
3. **Add pre-commit check**: Verify hook is installed before allowing commits
4. **AI reminder**: Always validate commit message format before executing `git commit`

## Phase 3.C – CLI Wiring

Phase 3.C completes the commit discipline reporting system by exposing it as first-class Stagecraft CLI commands.

### Commands

#### `stagecraft commit report`

Generates a commit health report analyzing commit message discipline.

**Usage:**
```bash
stagecraft commit report [--from=origin/main] [--to=HEAD]
```

**Output:**
- Path: `.stagecraft/reports/commit-health.json`
- Format: JSON report with schema version 1.0
- Contents:
  - Repository metadata
  - Commit range analyzed
  - Summary statistics (total, valid, invalid commits)
  - Per-commit validation results
  - Violation details with severity levels

**Input Sources:**
- Git commit history (via `git log`)
- Feature registry from `spec/features.yaml`
- Repository metadata (name, default branch)

#### `stagecraft feature traceability`

Generates a feature traceability report analyzing feature presence across spec, implementation, tests, and commits.

**Usage:**
```bash
stagecraft feature traceability
```

**Output:**
- Path: `.stagecraft/reports/feature-traceability.json`
- Format: JSON report with schema version 1.0
- Contents:
  - Summary statistics (total features, status breakdown)
  - Per-feature traceability:
    - Spec file presence and path
    - Implementation files (sorted)
    - Test files (sorted)
    - Commit SHAs referencing the feature
    - Detected problems (missing spec, missing tests, etc.)

**Input Sources:**
- Repository tree scan (deterministic lexicographical walk)
- Feature ID extraction from file headers (`// Feature: <ID>`)
- File classification (spec, implementation, test)

### Report Locations

All reports are written atomically to `.stagecraft/reports/`:

- `commit-health.json` – Commit message discipline analysis
- `feature-traceability.json` – Feature presence and traceability analysis

Reports use atomic writes (temporary file + rename) to ensure they are either fully written or not present at all.

### Deterministic Guarantees

Both commands provide deterministic output:

- **Commit report**: Deterministic git log parsing, sorted by SHA
- **Feature traceability**: Deterministic tree traversal (lexicographical), sorted file lists
- **JSON output**: Consistent formatting, no timestamps or random values
- **Atomic writes**: Reports are either complete or absent (no partial files)

### Integration with Phase 3.A and 3.B

- **Phase 3.A**: Defined report types and schemas
- **Phase 3.B**: Implemented pure generators (no I/O)
- **Phase 3.C**: Wired generators into CLI commands with deterministic I/O

This completes the commit discipline reporting pipeline: **spec → types → generators → CLI → reports**.


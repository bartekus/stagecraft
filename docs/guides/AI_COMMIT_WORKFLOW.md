# AI Commit Workflow Guide

**Quick reference for AI assistants committing to Stagecraft.**

> For detailed analysis, see: [`docs/COMMIT_MESSAGE_ANALYSIS.md`](../COMMIT_MESSAGE_ANALYSIS.md)  
> For full agent rules, see: [`Agent.md`](../../Agent.md#-commit-message-enforcement--discipline)

⸻

## 🎯 Core Principle

Commit messages are **deterministic artifacts** in Stagecraft's traceability chain:

**spec → tests → code → docs → commit → PR**

Every commit MUST maintain this chain.

⸻

## 📋 Mandatory Commit Format

```
<type>(<FEATURE_ID>): <summary>
```

**Components:**
- `<type>`: `feat`, `fix`, `refactor`, `docs`, `test`, `ci`, `chore` (lowercase)
- `<FEATURE_ID>`: SCREAMING_SNAKE_CASE (e.g., `PROVIDER_FRONTEND_GENERIC`)
- `<summary>`: ≤72 chars, no trailing period, lowercase after colon

**If FEATURE_ID is missing → STOP and ask.**

⸻

## ✅ Pre-Commit Checklist (AI MUST)

Before committing, AI MUST:

1. **Verify hook exists**
   ```bash
   ls -la .git/hooks/commit-msg
   ```
   - If missing → run `./scripts/install-hooks.sh`
   - If installation fails → STOP and report

2. **Validate commit message format**
   - Format: `<type>(<FEATURE_ID>): <summary>`
   - Type is lowercase
   - FEATURE_ID is SCREAMING_SNAKE_CASE
   - Summary ≤72 chars, no trailing period, lowercase after colon

3. **Verify FEATURE_ID matches branch**
   ```bash
   git branch --show-current
   ```
   - Extract FEATURE_ID from branch name
   - Ensure commit FEATURE_ID matches branch FEATURE_ID
   - If mismatch → STOP and report

4. **Verify no protected files**
   - Protected: LICENSE, README.md, ADRs, NOTICE
   - If modified → STOP and report

5. **Run all CI checks**
   ```bash
   ./scripts/run-all-checks.sh
   ```
   - All checks MUST pass
   - If any fail → STOP, fix, re-run

6. **Only then commit**
   ```bash
   git commit -m "<message>"
   ```

**If any check fails: STOP and report.**

⸻

## 🌿 Branch Naming Rules

Feature branches MUST follow:
```
feature/<FEATURE_ID>-short-desc
```

**Examples:**
- ✅ `feature/PROVIDER_FRONTEND_GENERIC-implement-provider`
- ✅ `fix/CLI_DEV-bug-fix`
- ❌ `feature/provider_frontend_generic` (FEATURE_ID must be uppercase)
- ❌ `Feature/PROVIDER_FRONTEND_GENERIC-frontend` (prefix must be lowercase)

**Constraints:**
- FEATURE_ID is uppercase (SCREAMING_SNAKE_CASE)
- `short-desc` is lowercase, hyphenated
- No spaces, 3-5 words

⸻

## 📝 Commit Message Examples

### ✅ Valid

```
feat(PROVIDER_FRONTEND_GENERIC): implement provider
fix(PROVIDER_FRONTEND_GENERIC): address review feedback
docs(PROVIDER_FRONTEND_GENERIC): sync roadmap docs
```

### ❌ Invalid

```
feat: implement PROVIDER_FRONTEND_GENERIC          # Missing parentheses
fix: address linter errors                        # Missing FEATURE_ID
feat(CLI_DEPLOY) update deploy command           # Missing colon
Feat(PROVIDER_FRONTEND_GENERIC): implement        # Type must be lowercase
feat(PROVIDER_FRONTEND_GENERIC): Implement.       # Capital after colon, trailing period
feat(CLI_PLAN, CLI_DEPLOY): refactor             # Multiple Feature IDs (forbidden)
```

⸻

## 🚫 AI MUST Reject

- Missing FEATURE_ID
- Wrong format (missing parentheses, colon)
- Uppercase after type
- Multi-feature commits
- Vague descriptions
- Subjects >72 chars
- Trailing periods
- Unicode/emoji decorations
- Hook bypassing (`STAGECRAFT_SKIP_HOOKS=1`)

⸻

## 🔗 Commit Message Body (Optional but Recommended)

Include spec and test references:

```
Spec: spec/commands/deploy.md
Tests: cmd/deploy_test.go
```

⸻

## 🎯 Feature Lifecycle Integration

Commit messages MUST maintain traceability:

- **FEATURE_ID validation**: Must match branch and `spec/features.yaml`
- **Single-feature rule**: One FEATURE_ID per commit
- **PR alignment**: Branch name and commit message must align
- **Spec traceability**: Links to:
  - `spec/features.yaml` (feature definition)
  - `spec/<domain>/<feature>.md` (spec file)
  - `docs/analysis/<FEATURE_ID>.md` (analysis brief)

⸻

## 📚 Related Documentation

- **Full Analysis**: [`docs/COMMIT_MESSAGE_ANALYSIS.md`](../COMMIT_MESSAGE_ANALYSIS.md)
- **Agent Rules**: [`Agent.md`](../../Agent.md#-commit-message-enforcement--discipline)
- **Phase 1 Issue**: [`.github/ISSUE_TEMPLATE/commit_message_phase1.md`](../../.github/ISSUE_TEMPLATE/commit_message_phase1.md)
- **Phase 2 Issue**: [`.github/ISSUE_TEMPLATE/commit_message_phase2.md`](../../.github/ISSUE_TEMPLATE/commit_message_phase2.md)
- **Hook Implementation**: [`.hooks/commit-msg`](../../.hooks/commit-msg)

⸻

## 🔄 Enforcement Phases

- **Phase 1** (Current): Local enforcement via hooks + AI workflow discipline
- **Phase 2** (Future): CI-level validation + optional CLI tooling

See TODO docs:
- [`docs/todo/COMMIT_MESSAGE_ENFORCEMENT_PHASE1.md`](../todo/COMMIT_MESSAGE_ENFORCEMENT_PHASE1.md)
- [`docs/todo/COMMIT_MESSAGE_ENFORCEMENT_PHASE2.md`](../todo/COMMIT_MESSAGE_ENFORCEMENT_PHASE2.md)

⸻

## ⚡ Quick Decision Tree

```
Need to commit?
  ↓
Have FEATURE_ID? → NO → STOP, ask user
  ↓ YES
Hook installed? → NO → Run ./scripts/install-hooks.sh
  ↓ YES
Message format valid? → NO → Fix format
  ↓ YES
FEATURE_ID matches branch? → NO → STOP, report mismatch
  ↓ YES
Protected files touched? → YES → STOP, report
  ↓ NO
CI checks pass? → NO → Fix issues, re-run
  ↓ YES
COMMIT ✓
```

⸻

**Remember:** Commit messages are deterministic artifacts. Every commit maintains the traceability chain: **spec → tests → code → docs → commit → PR**

⸻

## 📊 Run the Reports

Stagecraft provides two CLI commands for analyzing commit discipline and feature traceability.

### When to Run

Run these reports:

- **Before creating a PR**: Verify commit message discipline and feature completeness
- **After merging a feature**: Check feature traceability (spec, implementation, tests, commits)
- **Periodically**: Monitor commit health trends and feature gaps
- **In CI/CD**: Integrate into automated quality checks

### Commands

#### Commit Health Report

```bash
stagecraft commit report
```

Analyzes commit messages in the current branch (default: `origin/main..HEAD`).

**What it checks:**
- Commit message format compliance
- Feature ID presence and validity
- Feature ID matches spec registry
- Summary length and formatting rules

**Output:** `.stagecraft/reports/commit-health.json`

**Interpretation:**
- `summary.valid_commits` / `summary.invalid_commits`: Overall discipline
- `commits.<sha>.is_valid`: Per-commit status
- `commits.<sha>.violations`: Specific rule violations
- `summary.violations_by_code`: Violation frequency

#### Feature Traceability Report

```bash
stagecraft feature traceability
```

Scans repository for feature presence across spec, implementation, tests, and commits.

**What it checks:**
- Feature spec files exist
- Implementation files present
- Test files present
- Commits reference feature IDs
- Status consistency (e.g., "done" features have tests)

**Output:** `.stagecraft/reports/feature-traceability.json`

**Interpretation:**
- `summary.total_features`: Total features found
- `summary.features_with_gaps`: Features missing components
- `features.<id>.problems`: Specific traceability issues
- `features.<id>.status`: Feature lifecycle state

#### Commit Suggestions

```bash
stagecraft commit suggest
```

Reads both `.stagecraft/reports/commit-health.json` and `.stagecraft/reports/feature-traceability.json` and generates actionable suggestions.

**What it does:**
- Aggregates commit message violations into human-readable guidance
- Highlights missing or invalid Feature IDs
- Surfaces summary formatting issues (length, punctuation, capitalization)
- Prioritizes higher-severity issues first (errors → warnings → info)

**Output formats:**
- Text (default): grouped by severity with a final summary section
- JSON: machine-readable report with suggestion objects and summary counts

**Examples:**

```bash
# Human-readable output with defaults (severity >= info, up to 10 suggestions)
stagecraft commit suggest

# Only show high-priority issues, no limit
stagecraft commit suggest --severity=warning --max-suggestions=0

# JSON output, suitable for tooling and CI
stagecraft commit suggest --format=json --severity=info --max-suggestions=50
```

**Interpretation:**
- **Errors** – violations that MUST be fixed before merging
- **Warnings** – issues that should be addressed to maintain commit discipline
- **Info** – low-severity hygiene improvements and guidance

### Workflow Integration

**Before PR:**
1. Run `stagecraft commit report` to verify commit discipline.
2. Run `stagecraft feature traceability` to verify feature completeness.
3. Run `stagecraft commit suggest` to get a prioritized list of actions.
4. Address suggestions:
   - Rewrite commits where necessary
   - Add missing specs, implementation, or tests
   - Re-run the reports until suggestions are either resolved or explicitly accepted

**After Feature Completion:**
1. Run `stagecraft feature traceability`
2. Verify feature status matches reality (spec + impl + tests = "done")
3. Ensure commits reference the feature ID

**In CI/CD:**
- Add `stagecraft commit report` to PR checks
- Fail on high violation rates or missing feature IDs
- Use `stagecraft feature traceability` to enforce test coverage

### Report Schema

Both reports follow deterministic JSON schemas (version 1.0):

- **No timestamps**: Reports are deterministic and comparable across runs
- **Sorted lists**: All arrays are sorted for consistency
- **Atomic writes**: Reports are written atomically (no partial files)

See Phase 3.A/3.B documentation for detailed schema definitions.

⸻

**Remember:** Commit messages are deterministic artifacts. Every commit maintains the traceability chain: **spec → tests → code → docs → commit → PR**


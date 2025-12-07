# Commit Message Format Analysis

## Issue Summary

Commit messages were not following the required format:
```
<type>(<FEATURE_ID>): <summary>
```

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

## Examples

✅ **Valid:**
- `feat(PROVIDER_FRONTEND_GENERIC): implement provider`
- `fix(PROVIDER_FRONTEND_GENERIC): address review feedback`
- `docs(PROVIDER_FRONTEND_GENERIC): sync roadmap docs`

❌ **Invalid:**
- `feat: implement PROVIDER_FRONTEND_GENERIC` (missing parentheses)
- `fix: address linter errors` (missing FEATURE_ID)
- `docs: sync roadmap and status docs with spec/features.yaml` (missing FEATURE_ID, too long)

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

### 4. AI Workflow
When generating commit messages:
1. **Always** include `(<FEATURE_ID>)` in the format
2. **Verify** the format matches: `<type>(<FEATURE_ID>): <summary>`
3. **Check** subject length ≤72 characters
4. **Ensure** no trailing period

### 5. Pre-Commit Checklist
Before committing:
- [ ] Hook is installed (`ls .git/hooks/commit-msg`)
- [ ] Message follows format: `<type>(<FEATURE_ID>): <summary>`
- [ ] FEATURE_ID matches current feature
- [ ] Subject ≤72 characters
- [ ] No trailing period

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


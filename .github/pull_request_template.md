## Feature Information

**Feature ID:** `FEATURE_ID`

**Spec:** `spec/<domain>/<feature>.md`

**Status:** `todo` | `wip` | `done`

---

## Summary

Brief description of what this PR implements.

---

## Changes

- [ ] Analysis Brief: `docs/engine/analysis/<FEATURE_ID>.md` (created/updated)
- [ ] Implementation Outline: `docs/engine/outlines/<FEATURE_ID>_IMPLEMENTATION_OUTLINE.md` (created/updated)
- [ ] Spec: `spec/<domain>/<feature>.md` (created/updated)
- [ ] Tests: (list test files added/updated)
- [ ] Implementation: (list code files added/updated)
- [ ] Documentation: (list doc files updated)
- [ ] Lifecycle: `spec/features.yaml` (status updated)

---

## Testing

Describe the testing performed:

- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Golden tests updated (if applicable)
- [ ] Manual testing performed (if applicable)

Test files:
- `path/to/test_file_test.go`

---

## Spec Compliance

- [ ] Implementation matches spec exactly
- [ ] All v1 behavior from Implementation Outline is implemented
- [ ] All flags/arguments match spec
- [ ] Exit codes match spec
- [ ] Error messages match spec requirements

---

## Determinism

- [ ] No timestamps in output
- [ ] No random data
- [ ] All lists are lexicographically sorted
- [ ] JSON output is stable across runs
- [ ] No machine-dependent behavior

---

## Provider Boundaries (if applicable)

- [ ] Core remains provider-agnostic
- [ ] Provider-specific logic isolated to provider packages
- [ ] No hardcoded provider IDs
- [ ] Registry-based provider resolution

---

## Documentation

- [ ] Spec file complete and accurate
- [ ] Implementation Outline matches actual behavior
- [ ] Analysis Brief reflects final implementation
- [ ] Code comments include Feature ID and Spec path
- [ ] User-facing docs updated (if applicable)

---

## Checklist

Before requesting review:

- [ ] All CI checks pass
- [ ] `./scripts/run-all-checks.sh` passes locally
- [ ] `./scripts/validate-feature-integrity.sh` passes
- [ ] `./scripts/validate-spec.sh` passes
- [ ] Feature status updated in `spec/features.yaml`
- [ ] Commit message follows format: `<type>(<FEATURE_ID>): <summary>`
- [ ] No protected files modified
- [ ] Branch is up to date with main

---

## Related

- Closes #(issue number)
- Related to #(issue number)
- Depends on #(issue number)

---

## Notes

Any additional context, implementation decisions, or future work notes.


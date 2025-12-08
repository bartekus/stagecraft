---
name: New Feature
about: Propose a new feature for Stagecraft
title: '[FEATURE_ID] Feature Title'
labels: feature
assignees: ''
---

## Feature Information

**Feature ID:** `FEATURE_ID` (SCREAMING_SNAKE_CASE)

**Domain:** `commands` | `core` | `providers/<type>` | etc.

**Feature Name:** `feature-name` (kebab-case, for spec file)

---

## Problem Statement

Describe the core problem this feature solves.

---

## Motivation

Why does this feature matter for:

- Developer experience
- Operational reliability
- CI workflows
- Provider ecosystems

---

## User Stories

### Developers

- As a developer, I want to ...

### Platform Engineers

- As a platform engineer, I want to ...

### Automation and CI

- As a CI pipeline, I want to ...

---

## Success Criteria (v1)

Provide 5 to 7 measurable statements that define when the feature is done:

- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Criterion 3
- [ ] Criterion 4
- [ ] Criterion 5

---

## Dependencies

List feature, spec, or provider dependencies:

- Depends on: `FEATURE_ID_1` (status)
- Depends on: `FEATURE_ID_2` (status)

---

## Constraints and Risks

### Determinism Constraints

- Must not vary across machines or runs
- Must not depend on timestamps

### Provider Constraints

- Provider interfaces must remain stable
- No breaking changes to existing provider behavior

### Architectural Constraints

- Must reuse core packages where possible
- Must avoid circular dependencies

---

## Alternatives Considered

(Optional) List rejected approaches and why.

---

## Implementation Notes

(Optional) Any preliminary thoughts on implementation approach.

---

## Next Steps

After this issue is approved:

1. Run `./scripts/new-feature.sh FEATURE_ID DOMAIN feature-name` to create skeleton
2. Fill in Analysis Brief (`docs/engine/analysis/<FEATURE_ID>.md`)
3. Fill in Implementation Outline (`docs/<FEATURE_ID>_IMPLEMENTATION_OUTLINE.md`)
4. Fill in Spec (`spec/<domain>/<feature>.md`)
5. Add feature entry to `spec/features.yaml`
6. Begin implementation following Feature Planning Protocol (see `Agent.md`)

---

## Related

- Related to: #(issue number)
- Blocks: #(issue number)
- Blocked by: #(issue number)


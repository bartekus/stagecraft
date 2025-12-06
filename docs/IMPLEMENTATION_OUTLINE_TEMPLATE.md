# <FEATURE_ID> Implementation Outline

> This document defines the v1 implementation plan for <FEATURE_ID>. It translates the feature analysis brief into a concrete, testable, spec aligned delivery plan.

> All details in this outline must be reflected in `spec/<domain>/<feature>.md` before any tests or code are written.

⸻

## 1. Feature Summary

**Feature ID:** <FEATURE_ID>

**Domain:** <commands | core | provider | registry | etc>

**Goal:**

Short description of the feature and what it enables.

**v1 Scope:**

Describe exactly what will exist in v1.

**Out of scope for v1:**

- Bullet list

- Bullet list

**Future extensions (not implemented in v1):**

- Bullet list

- Bullet list

⸻

## 2. Problem Definition and Motivation

Short explanation of the functional gap this feature fills.

Why it matters for developers, operators, CI, or providers.

⸻

## 3. User Stories (v1)

### Developer

- Story 1

- Story 2

### Platform Engineer

- Story 1

### CI / Automation

- Story 1

⸻

## 4. Inputs and CLI or API Contract

### 4.1 Command or API Surface (v1)

```
stagecraft <command> [flags]
```

### 4.2 Flags or Arguments Implemented in v1

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | text | Output format or behavior |
| ... | ... | ... |

### 4.3 Flags or Arguments Reserved for Future Extensions

(Not implemented in v1)

| Flag | Planned Purpose |
|------|-----------------|
| `--roles` | Future filtering |
| `--phases` | Future filtering |

### 4.4 Exit Codes (v1)

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Any error in v1 |

Future planned exit codes should be listed plainly here but marked as not implemented.

⸻

## 5. Data Structures

List the exact Go structures or JSON schemas used in v1.

Example:

```go
type ProviderPlan struct {
    Provider string
    Steps    []ProviderStep
}

type ProviderStep struct {
    Name        string
    Description string
}
```

For JSON output, show the full expected v1 shape:

```json
{
  "provider_plans": [
    { "provider": "generic", "steps": [...] }
  ]
}
```

⸻

## 6. Determinism and Side Effects

### 6.1 Determinism Rules

- All lists sorted lexicographically

- No timestamps unless documented

- No random seeds

- JSON output stable across runs

### 6.2 Side Effect Constraints

- No shelling out

- No network I/O

- No state writes

- File reads allowed: config, compose, env

⸻

## 7. Provider Boundaries (if applicable)

Describe exactly which provider interfaces are invoked.

Describe any required provider behaviors in v1.

Example:

BackendProvider.Plan(ctx, opts) returns ProviderPlan and MUST be side effect free.

⸻

## 8. Testing Strategy

### Unit Tests

- Validate struct generation

- Validate provider behavior

- Validate error cases

### Integration / CLI Tests

- Validate command output

- Validate deterministic ordering

### Golden Tests

- Required for human readable output

- Required for JSON rendering

⸻

## 9. Implementation Plan Checklist

### Before coding:

- Analysis brief approved

- This outline approved

- Spec updated to match outline

### During implementation:

- Write failing tests first

- Implement v1 behavior

- Produce passing tests

### After implementation:

- Update docs if tests cause outline changes

- Ensure lifecycle completion in spec/features.yaml

⸻

## 10. Completion Criteria

The feature is complete only when:

- Tests pass

- Spec and outline match actual behavior

- Determinism guarantees enforced

- All planned v1 behavior delivered

- Feature status updated to done in spec/features.yaml


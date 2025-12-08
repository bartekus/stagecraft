# <FEATURE_ID> Feature Analysis Brief

This document captures the high level motivation, constraints, and success definition for <FEATURE_ID>.

It is the starting point for the Implementation Outline and Spec.

This brief must be approved before outline work begins.

⸻

## 1. Problem Statement

Describe the core problem the feature is solving.

Example:

Developers need a deterministic preview of deployment behavior without executing side effects.

⸻

## 2. Motivation

Why does this feature matter for:

- Developer experience

- Operational reliability

- CI workflows

- Provider ecosystems

⸻

## 3. Users and User Stories

### Developers

- As a developer, I want …

### Platform Engineers

- As a platform engineer, I want …

### Automation and CI

- As a CI pipeline, I want …

⸻

## 4. Success Criteria (v1)

Provide 5 to 7 measurable statements that define when the feature is done.

Example:

- Running stagecraft plan produces deterministic output in text and JSON.

- No external commands executed during plan generation.

- Provider interfaces remain side effect free.

- Errors return exit code 1 in v1.

⸻

## 5. Risks and Constraints

### Determinism Constraints

- Must not vary across machines or runs

- Must not depend on timestamps

### Provider Constraints

- Provider interfaces must remain stable

- No breaking changes to existing provider behavior

### Architectural Constraints

- Must reuse core packages where possible

- Must avoid circular dependencies

⸻

## 6. Alternatives Considered (optional)

List rejected approaches and why.

⸻

## 7. Dependencies

List feature, spec, or provider dependencies.

Example:

- Depends on CORE_PLAN

- Depends on provider registry stability

⸻

## 8. Approval

- Author:

- Reviewer:

- Date:

Once approved, the Implementation Outline may begin.


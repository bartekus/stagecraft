# Cortex Primitives

> **CONTRACT DEFINITIONS**

This directory defines the core "Cortex Primitives" — the fundamental cognitive structures the AI Agent uses to reason about the repository.

## 1. Decision Log Primitive
**Backed by**: `spec/governance/decisions.md`

The authoritative log of architectural and governance decisions.
- **Contract**: Agents must cite `DECISION-###` when justifying non-obvious actions.
- **Invariant**: No ADRs as separate files. All decisions are appended to the log.

## 2. Failure Classification Primitive
**Backed by**: `spec/governance/GOV_CLI_EXIT_CODES.md` (and enforced by `failure_lens`)

The taxonomy for understanding system failure.
- **Classes**: `user_input`, `config_invalid`, `external_dependency`, `provider_failure`, `transient_environment`, `internal_invariant`, `unclassified`.
- **Contract**: All errors must ideally map to one of these classes.
- **Mapping**: Maps deterministically to Exit Codes (0, 1, 2, 3).

## 3. Skill Registry Primitive
**Backed by**: `ai.agent/skills/registry.json`

The catalog of executable capabilities available to the agent.
- **Contract**: A skill exists IF AND ONLY IF it is present in `registry.json`.
- **Invariant**: The registry is sorted lexicographically by ID and is purely deterministic (no timestamps).

## 4. Context Lenses
**Backed by**: `ai.agent/skills/src/**` (Source code)

Deterministic analyzers that project repository state into structured context.
- **Examples**: `git_history_lens`, `failure_lens`.
- **Contract**: Lenses accept defined inputs and produce rigid, schema-compliant outputs without side effects.

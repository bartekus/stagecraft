# AI Cortex

> **AUTHORITATIVE OPERATIONAL TRUTH**
> This directory contains the "brain" of the Stagecraft AI Agent system.
> It is **NOT** documentation for humans. It is executable configuration and logic for agents.

## Purpose

The `ai.agent/` directory establishes the operational mechanics of the repository's AI agents. It bridges the gap between:
- **Behavioral Truth**: `spec/` (What the system must do)
- **Execution Truth**: `code/` (How the system does it)
- **Human Projections**: `docs/__generated__/` (Read-only maps for humans and context loading)

## Invariants

1. **Not Documentation**: Documentation lives in `docs/` (generated or narrative). This directory contains **Skills**, **Lenses**, and **Cortex Primitives**.
2. **Determinism**: All files and tools in this directory must yield deterministic, timestamp-free output.
3. **Execution**: The contents here are meant to be *executed* or *parsed* by agents, not just read.

## Structure

- `cortex/`: Core primitives (Decision Log contracts, Failure Classifications).
- `skills/`: Executable skills and the deterministic Skill Registry (`registry.json`).
- `skills/src/`: Source code for installed skills (Go-based).

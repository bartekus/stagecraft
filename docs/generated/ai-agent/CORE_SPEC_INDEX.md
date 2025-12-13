# Core Spec Index

Focused index of core engine and architecture specifications.

## Architecture Decision Records (ADR)

| Path | Chunks | Primary Topics |
|------|--------|----------------|
| `spec/adr/0001-architecture.md` | 7 | 0001 – Stagecraft Architecture and Project Structure, Alternatives Considered, Consequences, Context, Core Structure (v1) |

## Core Specifications

| Path | Chunks | Primary Topics |
|------|--------|----------------|
| `spec/core/backend-provider-config.md` | 19 | Backend Provider Configuration Schema, Benefits, Config Struct, Core Validation (Stagecraft), Encore.ts Provider |
| `spec/core/backend-registry.md` | 12 | Architecture, Backend Provider Registry, Goal, Interface, Non-Goals |
| `spec/core/compose.md` | 20 | API Design, Architecture, Behavior, Compose File Loading, Compose File Structure |
| `spec/core/config.md` | 10 | Backend, Behavior, Core Config – Loading and Validation, Databases (Migration Configuration), Default Path |
| `spec/core/env-resolution.md` | 13 | Behavior, Config Schema, Env File Parser, Environment Context, Environment Resolution |
| `spec/core/executil.md` | 11 | API Design, Behavior, Command Execution, Core ExecUtil – Process Execution Utilities, Error Handling |
| `spec/core/global-flags.md` | 11 | Behavior, Environment Variable Support, Flag Precedence, Flag Resolution, Flag Validation |
| `spec/core/logging.md` | 11 | API Design, Behavior, Core Logging – Structured Logging Helpers, Global Flag Integration, Goal |
| `spec/core/migration-registry.md` | 12 | Architecture, Engine Registration, Goal, Interface, Migration Engine Registry |
| `spec/core/phase-execution-common.md` | 21 | 1. Phase Work Failure, 2. Planner Failure (Pre-Execution), 3. State Manager Failure (Status Updates), Behaviour, CORE_PHASE_EXECUTION_COMMON – Shared Phase Execution Semantics |
| `spec/core/plan.md` | 11 | Architecture, Deployment Planning Engine, Example Plan, Future Enhancements, Goal |
| `spec/core/state-consistency.md` | 17 | 1. Purpose, 2. Scope, 3. Consistency Model, 3.1 Read-after-write Guarantee (Single Process), 3.2 Multi-manager Behaviour |
| `spec/core/state-test-isolation.md` | 12 | 1. Purpose, 2. Scope, 3. State Test Isolation Model, 3.1 Isolation Constraints, 3.2 Helper Function: `setupIsolatedStateTestEnv` |
| `spec/core/state.md` | 13 | Behavior, Environment Variable Support, Goal, Interface, Non-Goals (v1) |

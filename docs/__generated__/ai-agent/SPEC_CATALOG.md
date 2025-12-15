# Specification Catalog

This catalog focuses on specification files in `spec/**`.

## Summary

- **Total Spec Files**: 51
- **Total Spec Chunks**: 899

## Files

| Path | Chunks | Primary Topics |
|------|--------|----------------|
| `spec/adr/0001-architecture.md` | 7 | 0001 – Stagecraft Architecture and Project Structure, Alternatives Considered, Consequences, Context, Core Structure (v1) |
| `spec/commands/build.md` | 24 | 1. Summary, 2. CLI Definition, 2.1 Usage, 2.2 Required Flags, 2.3 Optional Flags |
| `spec/commands/commit-suggest.md` | 40 | 1. Purpose, 10. Testing Requirements, 10.1 Unit Tests, 10.2 Integration Tests, 10.3 Golden Tests |
| `spec/commands/deploy.md` | 18 | 1. Purpose, 10. Dependencies, 2. Scope, 3. CLI Interface, 3.1 Usage |
| `spec/commands/dev-basic.md` | 24 | Basic Node.js App, Behaviour, CLI Usage, Command Structure, Config Resolution |
| `spec/commands/dev.md` | 16 | Behaviour, CLI_DEV, Command, Configuration sources, Determinism |
| `spec/commands/infra-up.md` | 37 | 1. Overview, 10. Error Conditions, 10.1 Global Errors, 10.2 Per-Host Errors, 11. Determinism Guarantees |
| `spec/commands/init.md` | 8 | Behaviour, CLI Usage, Goal, Non-Goals (for initial version), Outputs |
| `spec/commands/migrate-basic.md` | 26 | Basic Node.js App, Behaviour, CLI Usage, Command Structure, Config Resolution |
| `spec/commands/plan.md` | 29 | 1. Purpose, 10. Dependencies, 11. Testing Requirements, 11.1 Golden Test Layout, 12. Implementation Notes |
| `spec/commands/releases.md` | 21 | Behaviour, CLI Usage, Command Structure, Error Handling, Error Messages |
| `spec/commands/rollback.md` | 32 | Behaviour, CLI Usage, Command Structure, Deploy Integration, Dry-run Semantics |
| `spec/commands/status-roadmap.md` | 32 | Behavior, Blocker Detection, Command Structure, Dependencies, Deterministic Output |
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
| `spec/deploy/compose-gen.md` | 14 | 1. Purpose, 2. Scope, 3. Inputs and Outputs, 4. Behavior, 5. Integration |
| `spec/deploy/rollout.md` | 10 | 1. Purpose, 2. Scope, 3. Compatibility Matrix, 4. Configuration, 5. Error Handling |
| `spec/dev/compose-infra.md` | 10 | Behaviour, DEV_COMPOSE_INFRA, Determinism, Excluded (future), Included |
| `spec/dev/hosts.md` | 19 | 1. Overview, 2. Behavior, 2.1 Hosts File Toggle, 2.2 Hosts File Paths, 2.3 Entry Format |
| `spec/dev/mkcert.md` | 20 | 1. Overview, 2. Behavior, 2.1 HTTPS Toggle, 2.2 Certificate Directory, 2.3 Certificate Files |
| `spec/dev/process-mgmt.md` | 15 | 1. Overview, 2. Behaviour, 2.1 Dev Files as Source of Truth, 2.2 Command Execution, 2.3 Foreground Mode (default) |
| `spec/dev/traefik.md` | 10 | Behaviour, DEV_TRAEFIK, Determinism, Excluded (future), Included |
| `spec/engine/plan-actions.md` | 22 | 1. Wire Format, 2. Determinism Rules (apply to all actions), 3. Validation Rules (apply to all actions), 4. Defaults Rules, 5. Forward Compatibility Rules |
| `spec/governance/GOV_CLI_EXIT_CODES.md` | 14 | Common Exit Code Semantics, Determinism, Excluded (v1), Exit Code Documentation Rules, Exit Codes |
| `spec/governance/GOV_CORE.md` | 15 | 1. Summary, 2. Goals, 3. Non-Goals, 4. Design, 4.1 Spec Schema (Frontmatter) |
| `spec/infra/bootstrap.md` | 28 | 1. Overview, 2. Responsibilities and Non-Goals, 2.1 Responsibilities (v1), 2.2 Non-Goals (v1), 3. Invocation and Execution Semantics |
| `spec/overview.md` | 5 | Command Surface (Initial), Core Concepts, High-Level Goals, Non-Goals (for v0), Stagecraft – Project Overview |
| `spec/providers/backend/encore-ts.md` | 11 | 1. Goals and Non-Goals, 1.1 Goals, 1.2 Non-Goals, 2. Relationship to Core Backend Abstraction, 2.1 BackendProvider Interface |
| `spec/providers/backend/generic.md` | 20 | Build Mode Behavior, Comparison with Other Providers, Config Parsing, Configuration, Dev Mode Behavior |
| `spec/providers/ci/interface.md` | 11 | CI Provider Interface, Config Schema, Error Types, Goal, Interface |
| `spec/providers/cloud/digitalocean.md` | 27 | 1. Overview, 10. Cost and Billing Responsibility, 2. Interface Contract, 2.1 ID, 2.2 Plan |
| `spec/providers/cloud/interface.md` | 11 | Cloud Provider Interface, Config Schema, Error Types, Goal, Interface |
| `spec/providers/frontend/generic.md` | 21 | Comparison with Other Providers, Config Parsing, Configuration, Dev Mode Behavior, Error Handling |
| `spec/providers/frontend/interface.md` | 9 | Config Schema, Frontend Provider Interface, Goal, Interface, Non-Goals (v1) |
| `spec/providers/migration/raw.md` | 23 | Comparison with Other Engines, Configuration, Core Validation (Stagecraft), Database Support, Engine-Specific Validation |
| `spec/providers/network/interface.md` | 11 | Config Schema, Error Types, Goal, Interface, Network Provider Interface |
| `spec/providers/network/tailscale.md` | 38 | 1. Overview, 10. Testing, 10.1 Unit Tests, 10.2 Integration Tests (Optional), 11. Non-Goals (v1) |
| `spec/providers/secrets/interface.md` | 11 | Config Schema, Error Types, Goal, Interface, Non-Goals (v1) |
| `spec/scaffold/stagecraft-dir.md` | 17 | .stagecraft/ Directory Structure, Behavior, Creation, Directory Structure, File Descriptions |

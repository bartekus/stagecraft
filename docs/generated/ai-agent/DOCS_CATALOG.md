# Documentation Catalog

This catalog lists all documentation files processed by the context compiler, grouped by top-level directory.

## Summary

- **Total Files**: 189
- **Total Chunks**: 3620

## Files by Directory

### Root Files

#### `Agent.md`

- **Chunks**: 77
- **Headings**: 0. Pre-Work: Branch Setup, 1. Every task ends with a commit message, 1. Spec‑first, Test‑first, 1.1 Feature Planning Protocol, 2. Feature ID Rules...

#### `README.md`

- **Chunks**: 21
- **Headings**: AI Context Pipeline, Build Stagecraft, Build from source, Building & running, Clone the repository...


### `docs/`

#### `docs/adr/0002-docs-lifecycle-and-ownership.md`

- **Chunks**: 11
- **Headings**: 0002 – Documentation Lifecycle and Ownership, Alternatives Considered, Consequences, Context, Decision...

#### `docs/adr/0003-driver-do-removal.md`

- **Chunks**: 13
- **Headings**: ADR 0003: Removal of DRIVER_DO – Provider Layer Simplification, Architectural Consistency, Clear Separation of Concerns, Consequences, Context...

#### `docs/archive/agents/AGENT_BRIEF_COVERAGE_PHASE1.md`

- **Chunks**: 11
- **Headings**: 1. Fix 4 Failing Tests, 2. Improve `pkg/config` Coverage from 66.7% → ≥ 80%, 3. Add Missing Test Files for "done" Features, Agent Brief: Test Coverage Compliance - Phase 1, Constraints...

#### `docs/archive/agents/AGENT_BRIEF_COVERAGE_PHASE2.md`

- **Chunks**: 19
- **Headings**: 1. `internal/git` (Target ≥ 70%), 2. `internal/tools/docs` (Target ≥ 60%), 3. `internal/providers/migration/raw` (Target ≥ 70%), Agent Brief: Test Coverage Quality Lift - Phase 2, Common Setup...

#### `docs/archive/agents/AGENT_BRIEF_GOV_V1_CORE_PHASE3.md`

- **Chunks**: 20
- **Headings**: Agent Brief: Spec Reference Checker Hardening - Phase 3, Alternative: Shell Script Hardening, Approach, Constraints (From Agent.md & GOV_V1_CORE), Current Problem...

#### `docs/archive/agents/AGENT_BRIEF_GOV_V1_CORE_PHASE4.md`

- **Chunks**: 20
- **Headings**: AGENT BRIEF — GOV_V1_CORE — Phase 4, Follow-up Work (Enforcement & Rollout), Golden Tests (optional but recommended), Implementation (Scaffold — Complete), Integration Tests (via run-all-checks.sh)...

#### `docs/archive/agents/AGENT_BRIEF_GOV_V1_CORE_PHASE5.md`

- **Chunks**: 18
- **Headings**: AGENT BRIEF — GOV_V1_CORE — Phase 5, Behavior, Governance Golden Tests, Must do, Pre work...

#### `docs/archive/agents/PROVIDER_NETWORK_TAILSCALE_SLICE1_AGENT.md`

- **Chunks**: 13
- **Headings**: 1. Extract Pure Helpers, 2. Refactor Existing Code to Use Helpers, 3. Add Unit Tests for Helpers, 4. Add `parseStatus` Edge-Case Tests, 5. Run Verification Commands...

#### `docs/archive/agents/PROVIDER_NETWORK_TAILSCALE_SLICE2_AGENT.md`

- **Chunks**: 20
- **Headings**: 1. Implement Version Parsing Helper, 2. Update EnsureInstalled Version Logic, 3. Add Config Validation Tests, 4. Add OS Compatibility Tests, 5. Add Version Parsing and Enforcement Tests...

#### `docs/archive/context-handoff/CLI_DEPLOY-to-CLI_RELEASES.md`

- **Chunks**: 16
- **Headings**: CLI_ROLLBACK, Feature Complete: CLI_DEPLOY, Implement CLI_RELEASES, What Now Exists, ⚡ Quick Start for Next Agent...

#### `docs/archive/context-handoff/CLI_PHASE_EXECUTION_COMMON-to-CORE_STATE_TEST_ISOLATION.md`

- **Chunks**: 17
- **Headings**: CLI_PHASE_EXECUTION_COMMON (Complete Test Migration), Current Test Status, Feature Complete: CLI_PHASE_EXECUTION_COMMON, Implement CORE_STATE_TEST_ISOLATION, What Now Exists...

#### `docs/archive/context-handoff/CLI_RELEASES-to-CLI_ROLLBACK.md`

- **Chunks**: 15
- **Headings**: Feature Complete: CLI_RELEASES, Future Enhancements (v2), Implement CLI_ROLLBACK, ⚡ Quick Start for Next Agent, ✅ Final Checklist...

#### `docs/archive/context-handoff/CLI_ROLLBACK-CLI_RELEASES-CORE_STATE_CONSISTENCY-to-CLI_DEPLOY.md`

- **Chunks**: 33
- **Headings**: 1. Summary, 1. `TestDeployCommand_Success_AllPhasesRunInOrder`, 10. `TestDeployCommand_MigratePhasesExistInState`, 11. `TestDeployCommand_UsesBackendProviderBuild`, 12. `TestDeployCommand_Output_Golden`...

#### `docs/archive/context-handoff/CLI_ROLLBACK-to-CLI_BUILD.md`

- **Chunks**: 17
- **Headings**: CLI_PLAN, DEPLOY_COMPOSE_GEN (Design Only), Feature Complete: CLI_ROLLBACK, Implement CLI_BUILD, What Now Exists...

#### `docs/archive/context-handoff/CLI_ROLLBACK-to-CLI_PHASE_EXECUTION_COMMON.md`

- **Chunks**: 37
- **Headings**: Bootloader Instructions:, CLI_BUILD (Design Only), CLI_PLAN, Current State Analysis:, Feature Complete: CLI_ROLLBACK...

#### `docs/archive/context-handoff/COMMIT_DISCIPLINE_PHASE3.md`

- **Chunks**: 11
- **Headings**: 1. Historical Commit Health Analysis (Read-only), 2. Feature-Aware Commit Suggestions (Optional helper), 3. Traceability Gap Detection, 4. Documentation, ⚠️ Non-Goals (Phase 3)...

#### `docs/archive/context-handoff/COMMIT_DISCIPLINE_PHASE3B.md`

- **Chunks**: 11
- **Headings**: 1. Commit History Scanner → Commit Health Report, 2. Feature Traceability Index → Feature Traceability Report, 3. CLI Integration & I/O Adapters, 4. Optional: Feature-Aware Commit Suggestion Helper, 5. Documentation Updates...

#### `docs/archive/context-handoff/COMMIT_DISCIPLINE_PHASE3C.md`

- **Chunks**: 13
- **Headings**: 1. CLI Commands (spec-defined), 2. Git Adapter (Thin, Deterministic), 3. Repository Scanner, 4. File Output Layer (Stable, Atomic), 5. CLI Tests (Golden + Integration)...

#### `docs/archive/context-handoff/COMMIT_REPORT_TYPES_PHASE3.md`

- **Chunks**: 11
- **Headings**: 1. Implement `commithealth` report types, 2. Implement `featuretrace` report types, 3. Deterministic JSON roundtrip tests (golden), 4. Determinism Constraints, 5. Documentation Update...

#### `docs/archive/context-handoff/CORE_STATE_TEST_ISOLATION-to-CORE_STATE_CONSISTENCY.md`

- **Chunks**: 1

#### `docs/archive/context-handoff/CORE_STATE-to-CLI_DEPLOY.md`

- **Chunks**: 17
- **Headings**: CLI_RELEASES, CLI_ROLLBACK (Design Only), Feature Complete: CORE_STATE, Implement CLI_DEPLOY, What Now Exists...

#### `docs/archive/context-handoff/GOV_STATUS_ROADMAP_COMPLETE.md`

- **Chunks**: 14
- **Headings**: 1. GOV_STATUS_ROADMAP Enhancements (v2), 2. PROVIDER_FRONTEND_GENERIC Coverage Phase 2, 3. DEV_PROCESS_MGMT Feature Mapping Issue, Code Quality Fixes, Coverage Improvements...

#### `docs/archive/context-handoff/GOV_V1_CORE-to-FRONTMATTER.md`

- **Chunks**: 16
- **Headings**: Add Frontmatter to All Existing Spec Files, Complete GOV_V1_CORE (Phase 2 & 3), Feature Complete: GOV_V1_CORE (Phase 1), What Now Exists, ⚡ Quick Start for Next Agent...

#### `docs/archive/coverage/CLI_DEV_COMPLETE_PHASE3_PR_DESCRIPTION.md`

- **Chunks**: 12
- **Headings**: 1. DEV_HOSTS planning and spec, 2. DEV_HOSTS implementation, 3. DEV_HOSTS tests, 4. CLI_DEV integration, 5. CLI_DEV tests...

#### `docs/archive/coverage/COVERAGE_COMPLIANCE_PLAN_PHASE2.md`

- **Chunks**: 21
- **Headings**: Constraints, Coverage Requirements, Execution Order, Implementation Notes, In Scope...

#### `docs/archive/coverage/COVERAGE_COMPLIANCE_PLAN.md`

- **Chunks**: 17
- **Headings**: 1. `PROVIDER_BACKEND_INTERFACE`, 1.1 `internal/tools/cliintrospect` Tests, 1.2 `internal/cli/commands` Build Tests, 2. `CLI_DEPLOY`, Constraints...

#### `docs/archive/coverage/PROVIDER_FRONTEND_GENERIC_COVERAGE_PHASE1_OUTLINE.md`

- **Chunks**: 35
- **Headings**: 1. Feature Summary, 10. Approval, 2. Test Implementation Plan, 2.1 Test Files Structure, 2.2 Test Organization...

#### `docs/archive/coverage/PROVIDER_FRONTEND_GENERIC_COVERAGE_PHASE1.md`

- **Chunks**: 31
- **Headings**: 1. Problem Statement, 10. Approval, 2. Motivation, 3. Users and User Stories, 4. Success Criteria (Phase 1)...

#### `docs/archive/coverage/PROVIDER_FRONTEND_GENERIC_COVERAGE_PHASE2.md`

- **Chunks**: 13
- **Headings**: After Phase 2, Before Phase 2, Coverage Results, Files Modified, Implementation Details...

#### `docs/archive/coverage/PROVIDER_FRONTEND_GENERIC_DEFLAKE_FOLLOWUP.md`

- **Chunks**: 7
- **Headings**: Alternative Text (If Using Option B - Extract Helper), Follow-up: COVERAGE_STRATEGY.md Update After Deflaking Scanner Error Test, Resolved Test Debt, Scanner Error Test for `runWithReadyPattern` (Resolved), Section to Replace

#### `docs/archive/coverage/PROVIDER_FRONTEND_GENERIC_PR_DESCRIPTION.md`

- **Chunks**: 12
- **Headings**: Changes, Checklist, Coverage Results, Documentation, Governance Alignment...

#### `docs/archive/coverage/PROVIDER_FRONTEND_GENERIC_REVIEWER_GUIDE.md`

- **Chunks**: 15
- **Headings**: 1. Test Hardening Implementation, 2. Test Seam Addition, 3. Scanner Error Test, 4. Documentation, All tests pass...

#### `docs/archive/coverage/PROVIDER_FRONTEND_GENERIC_TEST_HARDENING_SUMMARY.md`

- **Chunks**: 11
- **Headings**: 1. Test Hardening Implementation, 2. Documentation Updates, 3. Test Seam Addition, Coverage Results, Files Changed...

#### `docs/archive/coverage/PROVIDER_NETWORK_TAILSCALE_PR_DESCRIPTION.md`

- **Chunks**: 14
- **Headings**: 1. PROVIDER_NETWORK_TAILSCALE planning and spec, 2. PROVIDER_NETWORK_TAILSCALE implementation, 3. OS Detection, 4. Tag Management, 5. Integration with Network Registry...

#### `docs/archive/coverage/TEST_COVERAGE_ANALYSIS.md`

- **Chunks**: 35
- **Headings**: 1. `PROVIDER_BACKEND_INTERFACE`, 1. `internal/tools/cliintrospect` - 2 Test Failures, 2. `CLI_DEPLOY`, 2. `internal/cli/commands` - 2 Test Failures, Appendix: Coverage Thresholds Reference...

#### `docs/archive/engine/analysis/PROVIDER_NETWORK_TAILSCALE_REQUIREMENTS.md`

- **Chunks**: 38
- **Headings**: 1. Behavioral Contract (What the provider must do), 1.1 ID, 1.2 EnsureInstalled, 1.3 EnsureJoined, 1.4 NodeFQDN...

#### `docs/archive/governance/CI_PROVIDER_COVERAGE_ENFORCEMENT.md`

- **Chunks**: 12
- **Headings**: Automated Coverage Reporting, CI Integration, CI Provider Coverage Enforcement, Check provider governance, Coverage Threshold Enforcement...

#### `docs/archive/governance/COMMIT_GUIDANCE_PROVIDER_GOVERNANCE.md`

- **Chunks**: 11
- **Headings**: 1. Provider governance check, 2. Coverage planner (visual check), 3. Full validation suite (optional but recommended), Commit Guidance: Provider Coverage Governance, Core Scripts...

#### `docs/archive/governance/COMMIT_MESSAGE_ANALYSIS.md`

- **Chunks**: 20
- **Headings**: 1. Always Install Hooks, 2. Verify Hook Installation, 3. Test Hook Before Committing, 4. AI Workflow (Mandatory Rules), 5. Feature Lifecycle Integration...

#### `docs/archive/governance/COMMIT_READY_SUMMARY.md`

- **Chunks**: 6
- **Headings**: 1. Quick Follow-Up: infra-up Spec Fix, 2. Provider Coverage Improvements, Commit, Governance Commit - Ready Summary, Next Steps After Commit...

#### `docs/archive/governance/GOV_V1_CORE_PHASE3_PLAN.md`

- **Chunks**: 28
- **Headings**: Alternative: Shell Script Hardening, Approach: Go-Based Validator, Comment Pattern, Current Implementation, False Positive Scenarios...

#### `docs/archive/governance/GOV_V1_TEST_REQUIREMENTS.md`

- **Chunks**: 22
- **Headings**: 1. Determinism First, 2. Separation of Concerns, 3. No Test Seams, 4. Coverage Targets, Anti-Patterns to Avoid...

#### `docs/archive/governance/INFRA_UP_SPEC_FIX.md`

- **Chunks**: 10
- **Headings**: CLI_INFRA_UP - Spec Frontmatter Fix, Commit Message, Exit Codes, Full validation, Implementation...

#### `docs/archive/governance/PHASE5_VIOLATION_FIX_CHECKLIST.md`

- **Chunks**: 22
- **Headings**: 1.1 GOV_V1_CORE → Wrong Spec Path (4 files), 1.2 MIGRATION_INTERFACE → Wrong Spec Path (1 file), 1.3 PROVIDER_FRONTEND_GENERIC → Wrong Spec Path (2 files), 2.1 CLI_GLOBAL_FLAGS, 2.2 CORE_STATE_TEST_ISOLATION...

#### `docs/archive/governance/PR_TEMPLATE_PROVIDER_COVERAGE.md`

- **Chunks**: 16
- **Headings**: Added, Alignment with Governance, Changes, Check coverage, Checklist...

#### `docs/archive/governance/PROVIDER_BACKEND_ENCORE_COVERAGE_PLAN.md`

- **Chunks**: 15
- **Headings**: 1. Flakiness Patterns, 2. Deterministic Design, 3. Test Organization, Current State, Estimated Effort...

#### `docs/archive/governance/PROVIDER_BACKEND_GENERIC_COVERAGE_PLAN.md`

- **Chunks**: 15
- **Headings**: 1. Flakiness Patterns, 2. Deterministic Design, 3. Test Organization, Current State, Estimated Effort...

#### `docs/archive/governance/PROVIDER_CLOUD_DO_COVERAGE_PLAN.md`

- **Chunks**: 16
- **Headings**: 1. API Client Error Paths, 2. Config Validation Edge Cases, 3. Plan Reconciliation Edge Cases, Current State, Estimated Effort...

#### `docs/archive/governance/PROVIDER_CLOUD_DO_MICRO_PLAN.md`

- **Chunks**: 11
- **Headings**: Commit Message, Coverage Analysis, Expected Outcome, Expected: ≥80%, Implementation Steps...

#### `docs/archive/governance/PROVIDER_GOVERNANCE_SUMMARY.md`

- **Chunks**: 22
- **Headings**: - 1 V1 Complete (PROVIDER_FRONTEND_GENERIC), - 4 V1 Plan (ready for improvement), CI Integration, Commit Ready, Coverage Improvement...

#### `docs/archive/governance/PROVIDER_NETWORK_TAILSCALE_SLICE1_CHECKLIST.md`

- **Chunks**: 11
- **Headings**: After helpers extracted - verify behavior unchanged, After tests added - verify coverage increase, Before starting - baseline coverage, Common Pitfalls to Avoid, Final verification...

#### `docs/archive/governance/PROVIDER_NETWORK_TAILSCALE_SLICE1_PLAN.md`

- **Chunks**: 20
- **Headings**: 1. `buildTailscaleUpCommand` (NEW), 2. `parseOSRelease` (NEW), 3. `validateTailnetDomain` (NEW), 4. `buildNodeFQDN` (NEW), 5. `parseStatus` (EXISTING - needs more tests)...

#### `docs/archive/governance/PROVIDER_NETWORK_TAILSCALE_SLICE1_READY.md`

- **Chunks**: 6
- **Headings**: Next Steps After Slice 1, PROVIDER_NETWORK_TAILSCALE Slice 1 - Ready to Execute, Pre-Verified, Quick Start, Success Criteria...

#### `docs/archive/governance/PROVIDER_NETWORK_TAILSCALE_SLICE2_CHECKLIST.md`

- **Chunks**: 21
- **Headings**: After tests added - verify coverage increase, After version parsing added - verify behavior unchanged, All tests pass, Before starting - baseline coverage, Commander Mock Setup Patterns...

#### `docs/archive/governance/PROVIDER_NETWORK_TAILSCALE_SLICE2_COMPLETENESS.md`

- **Chunks**: 43
- **Headings**: 1. Document Structure Completeness, 10. AATSE Compliance Check, 10.1 Test Design Principles, 10.2 Spec Alignment, 11. Implementation Readiness...

#### `docs/archive/governance/PROVIDER_NETWORK_TAILSCALE_SLICE2_COVERAGE_EXPECTATIONS.md`

- **Chunks**: 30
- **Headings**: 1. New Functions (100% Coverage Target), 2. Modified Functions (Improved Coverage), 2.1 `EnsureInstalled()`, 2.2 `checkOSCompatibility()`, 2.3 `parseConfig()`...

#### `docs/archive/governance/PROVIDER_NETWORK_TAILSCALE_SLICE2_DEPENDENCIES.md`

- **Chunks**: 20
- **Headings**: Breaking Changes, Code Modification Impact, Conclusion, Coverage Dependency Chain, Critical Path...

#### `docs/archive/governance/PROVIDER_NETWORK_TAILSCALE_SLICE2_PLAN.md`

- **Chunks**: 21
- **Headings**: 1. Config Parsing and Validation Paths, 2. OS Compatibility and Version Logic, 2.1 Supported OS Detection, 2.2 Version Enforcement, 3. Install vs Already-Installed Flows...

#### `docs/archive/governance/PROVIDER_NETWORK_TAILSCALE_SPEC_UPDATES.md`

- **Chunks**: 10
- **Headings**: Decision 1: Tailnet Name vs MagicDNS Domain Matching, Decision 2: Tag Equality Semantics, Decision 3: Role Handling in EnsureJoinedOptions, Decision 4: Tag Validation (tag: prefix requirement), Decision 5: Minimum Version Enforcement...

#### `docs/archive/governance/STRATEGIC_DOC_MIGRATION.md`

- **Chunks**: 20
- **Headings**: 1. Create Your Internal Strategic Document, 2. Verify Git Ignore, 3. Verify It's Not Tracked, 4. Use the Safe Public Version, Architecture Considerations...

#### `docs/archive/PROJECT_STRUCTURE_ANALYSIS.md`

- **Chunks**: 28
- **Headings**: 1. Main Entry Point, 10. Go Module Structure, 2. Binary Artifacts Tracked in Git, 3. Inconsistent Build Output Locations, 4. Documentation Organization...

#### `docs/archive/registry-implementation-summary.md`

- **Chunks**: 22
- **Headings**: Architecture Verification, Conclusion, Config Structure, Config Tests, Critical Success Factors - Verification...

#### `docs/archive/status/PROVIDER_BACKEND_ENCORE_COVERAGE_V1_COMPLETE.md`

- **Chunks**: 10
- **Headings**: Alignment with Governance, Coverage Metrics, Documentation, Documentation Updated, Next Steps...

#### `docs/archive/status/PROVIDER_BACKEND_GENERIC_COVERAGE_V1_COMPLETE.md`

- **Chunks**: 10
- **Headings**: Alignment with Governance, Coverage Metrics, Documentation, Documentation Updated, Next Steps...

#### `docs/archive/status/PROVIDER_CLOUD_DO_COVERAGE_V1_COMPLETE.md`

- **Chunks**: 10
- **Headings**: Added, Alignment with Governance, Coverage Metrics, Documentation, Next Steps...

#### `docs/archive/status/PROVIDER_FRONTEND_GENERIC_COVERAGE_PR.md`

- **Chunks**: 17
- **Headings**: Added, Alignment with Governance, Changes, Check coverage, Checklist...

#### `docs/archive/status/PROVIDER_FRONTEND_GENERIC_COVERAGE_V1_COMPLETE.md`

- **Chunks**: 10
- **Headings**: Added, Alignment with Governance, Coverage Metrics, Documentation, Next Steps...

#### `docs/archive/status/PROVIDER_NETWORK_TAILSCALE_COVERAGE_PLAN.md`

- **Chunks**: 4
- **Headings**: Alignment, PROVIDER_NETWORK_TAILSCALE - Coverage V1 Plan, Plan, Summary

#### `docs/archive/todo/COMMIT_MESSAGE_ENFORCEMENT_PHASE1.md`

- **Chunks**: 3
- **Headings**: Notes, Scope, TODO: Phase 1 – Commit Message Discipline Enforcement

#### `docs/archive/todo/COMMIT_MESSAGE_ENFORCEMENT_PHASE2.md`

- **Chunks**: 3
- **Headings**: Notes, Scope, TODO: Phase 2 – Commit Message CI Validation & CLI Tooling

#### `docs/context-handoff/CONTEXT_LOG.md`

- **Chunks**: 13
- **Headings**: 1. Purpose and Scope, 2. Usage Rules, 3. Index of Entries, 4. Entries, 4.1 2025-12-XX - Example Entry Template...

#### `docs/context-handoff/INDEX.md`

- **Chunks**: 10
- **Headings**: Available Handoff Documents, CLI Command Handoffs, Context Handoff Documents Index, Core Engine Handoffs, Governance Handoffs...

#### `docs/context-handoff/README.md`

- **Chunks**: 14
- **Headings**: - <CURRENT_FEATURE_ID> → actual feature ID, - <NEXT_FEATURE_ID> → next feature ID, - <PR_NUMBER>, <COMMIT_HASH>, etc., After completing a feature, Architectural Context...

#### `docs/context-handoff/TEMPLATE.md`

- **Chunks**: 17
- **Headings**: <FUTURE_DESIGN_ONLY_FEATURE_ID> (Design Only), <SECONDARY_FEATURE_ID>, Feature Complete: <CURRENT_FEATURE_ID>, Implement <NEXT_FEATURE_ID>, What Now Exists...

#### `docs/coverage/COVERAGE_LEDGER.md`

- **Chunks**: 18
- **Headings**: 1. Purpose and Scope, 2. Current Snapshot, 2.1 Coverage by Domain, 2.2 Coverage by Provider, 3. Historical Coverage Timeline...

#### `docs/coverage/PROVIDER_COVERAGE_TEMPLATE.md`

- **Chunks**: 15
- **Headings**: 1. Extracted Deterministic Primitives (if applicable), 2. Replaced Flaky Tests, 3. Integration Coverage Scope, <FEATURE_ID> - Coverage Strategy (<STATUS_LABEL>), Deterministic Primitives...

#### `docs/design/commit-reports-go-types.md`

- **Chunks**: 7
- **Headings**: Commit Health Report Types, Determinism Requirements, Example Usage, Feature Traceability Report Types, File Structure...

#### `docs/engine/agents/AGENT_BRIEF_COVERAGE.md`

- **Chunks**: 11
- **Headings**: Agent Brief: Test Coverage Compliance, During Work, How to Run Coverage Work, Invariants (Never Changes), Migration Notes...

#### `docs/engine/agents/AGENT_BRIEF_GOV_V1_CORE.md`

- **Chunks**: 12
- **Headings**: Agent Brief: GOV_V1_CORE Governance Hardening, During Work, How to Run GOV_V1_CORE Work, Invariants (Never Changes), Migration Notes...

#### `docs/engine/agents/GOV_PRE_COMMIT_INTEGRATION.md`

- **Chunks**: 13
- **Headings**: "check-orphan-specs.sh not found", "stagecraft binary not found", === Governance checks (optional but recommended) ===, Copy the contents of .hooks/pre-commit-gov-snippet.sh here, Escape Hatches...

#### `docs/engine/agents/PROVIDER_COVERAGE_AGENT.md`

- **Chunks**: 18
- **Headings**: 1. Purpose, 2. Primary Inputs, 3. Operating Mode, 3.1 High Level Rules, 3.2 Workflow for Each Provider...

#### `docs/engine/agents/STAGECRAFT_VALIDATION_AGENT.md`

- **Chunks**: 21
- **Headings**: 1. Purpose, 2. Primary Inputs, 3. Operating Mode, 3.1 High Level Rules, 3.2 Slice Based Execution...

#### `docs/engine/analysis/CLI_DEV.md`

- **Chunks**: 27
- **Headings**: 1. Problem Statement, 2. Motivation, 3. Users and User Stories, 4. Success Criteria (v1), 4.1. v1 Implementation Status and Limitations...

#### `docs/engine/analysis/CLI_INFRA_UP.md`

- **Chunks**: 31
- **Headings**: 1. Problem Statement, 10. Approval, 2. Motivation, 2.1 Create Infrastructure in a Provider-Agnostic Way, 2.2 Deterministic, Idempotent Infrastructure Creation...

#### `docs/engine/analysis/CLI_PLAN_ANALYSIS.md`

- **Chunks**: 14
- **Headings**: 1. Existing Infrastructure Review, 1. Specification Document ✅, 2. Features.yaml Update ✅, 2. Key Findings, 3. Implementation Outline ✅...

#### `docs/engine/analysis/DEV_HOSTS.md`

- **Chunks**: 24
- **Headings**: 1. Problem Statement, 2. Motivation, 3. Users and User Stories, 4. Success Criteria (v1), 5. Risks and Constraints...

#### `docs/engine/analysis/DEV_MKCERT.md`

- **Chunks**: 21
- **Headings**: 1. Problem Statement, 2. Motivation, 3. Users and User Stories, 4. Success Criteria (v1), 5. Risks and Constraints...

#### `docs/engine/analysis/DEV_PROCESS_MGMT.md`

- **Chunks**: 21
- **Headings**: 1. Problem Statement, 2. Motivation, 3. Users and User Stories, 4. Success Criteria (v1), 5. Risks and Constraints...

#### `docs/engine/analysis/DEV_TRAEFIK.md`

- **Chunks**: 21
- **Headings**: 1. Problem Statement, 2. Motivation, 3. Users and User Stories, 4. Success Criteria (v1), 5. Risks and Constraints...

#### `docs/engine/analysis/GOV_STATUS_ROADMAP.md`

- **Chunks**: 27
- **Headings**: 1. Problem Statement, 10. Approval, 2. Motivation, 3. Goals, 4. Constraints...

#### `docs/engine/analysis/GOV_V1_CORE_IMPLEMENTATION_ANALYSIS.md`

- **Chunks**: 24
- **Headings**: 1. What Was Actually Completed, 1.1 Spec Schema + Validation ✅, 1.2 Feature Dependency Graph + Impact + DOT ✅, 1.3 CLI Introspection ✅, 1.4 Feature Overview Generator ✅...

#### `docs/engine/analysis/INFRA_HOST_BOOTSTRAP.md`

- **Chunks**: 32
- **Headings**: Alternative 1: Include Bootstrap in CloudProvider, Alternative 2: Include Bootstrap in CLI_INFRA_UP, Alternative 3: Use Cloud-Init or User Data Scripts, Alternative 4: Skip Bootstrap, Require Pre-Configured Hosts, Alternatives Considered...

#### `docs/engine/analysis/PROVIDER_CLOUD_DO.md`

- **Chunks**: 28
- **Headings**: 1. Problem Statement, 2. Motivation, 3. Users and User Stories, 4. Success Criteria (v1), 5. Risks and Constraints...

#### `docs/engine/analysis/PROVIDER_NETWORK_TAILSCALE.md`

- **Chunks**: 26
- **Headings**: 1. Problem Statement, 2. Motivation, 3. Users and User Stories, 4. Success Criteria (v1), 5. Risks and Constraints...

#### `docs/engine/analysis/TEMPLATE.md`

- **Chunks**: 15
- **Headings**: 1. Problem Statement, 2. Motivation, 3. Users and User Stories, 4. Success Criteria (v1), 5. Risks and Constraints...

#### `docs/engine/engine-index.md`

- **Chunks**: 40
- **Headings**: CLI_* (Commands), CORE_* (Core Engine), Context Handoff, Core Principles, Example: Working on `CLI_BUILD`...

#### `docs/engine/history/PROVIDER_BACKEND_ENCORE_EVOLUTION.md`

- **Chunks**: 15
- **Headings**: 1. Purpose and Scope, 10. Migration Notes, 2. Feature References, 3. Design Intent and Constraints, 4. Coverage Timeline Overview...

#### `docs/engine/history/PROVIDER_BACKEND_GENERIC_EVOLUTION.md`

- **Chunks**: 15
- **Headings**: 1. Purpose and Scope, 10. Migration Notes, 2. Feature References, 3. Design Intent and Constraints, 4. Coverage Timeline Overview...

#### `docs/engine/history/PROVIDER_CLOUD_DO_EVOLUTION.md`

- **Chunks**: 15
- **Headings**: 1. Purpose and Scope, 10. Migration Notes, 2. Feature References, 3. Design Intent and Constraints, 4. Coverage Timeline Overview...

#### `docs/engine/history/PROVIDER_FRONTEND_GENERIC_EVOLUTION.md`

- **Chunks**: 33
- **Headings**: 1. Purpose and Scope, 10. Open Questions and Future Work, 11. Archived Source Documents, 11.1 Archived Source: Phase 1 Coverage Analysis Brief, 11.2 Archived Source: Phase 1 Implementation Outline...

#### `docs/engine/history/PROVIDER_NETWORK_TAILSCALE_EVOLUTION.md`

- **Chunks**: 23
- **Headings**: 1. Purpose and Scope, 10. Archived Source Documents, 10.1 Slice 1 Documents, 10.2 Slice 2 Documents, 10.3 Other Documents...

#### `docs/engine/outlines/CLI_DEV_IMPLEMENTATION_OUTLINE.md`

- **Chunks**: 34
- **Headings**: 1. Feature Summary, 10. Completion Criteria, 11. Implementation Order, 2. Problem Definition and Motivation, 3. User Stories (v1)...

#### `docs/engine/outlines/CLI_INFRA_UP_IMPLEMENTATION_OUTLINE.md`

- **Chunks**: 32
- **Headings**: 1. Feature Summary, 10. Related Documentation, 2. Problem Definition and Motivation, 3. Execution Model and Orchestration, 3.1 Command Structure...

#### `docs/engine/outlines/CLI_PLAN_IMPLEMENTATION_OUTLINE.md`

- **Chunks**: 1

#### `docs/engine/outlines/DEV_COMPOSE_INFRA_IMPLEMENTATION_OUTLINE.md`

- **Chunks**: 42
- **Headings**: 1. Feature Summary, 10. Implementation Plan Checklist, 11. Completion Criteria, 12. Implementation Notes, 2. Problem Definition and Motivation...

#### `docs/engine/outlines/DEV_HOSTS_IMPLEMENTATION_OUTLINE.md`

- **Chunks**: 36
- **Headings**: 1. Feature Summary, 10. Implementation Checklist, 11. Completion Criteria, 12. Implementation Order, 2. Problem Definition and Motivation...

#### `docs/engine/outlines/DEV_MKCERT_IMPLEMENTATION_OUTLINE.md`

- **Chunks**: 32
- **Headings**: 1. Feature Summary, 10. Completion Criteria, 2. Problem Definition and Motivation, 3. User Stories (v1), 4. Inputs and API Contract...

#### `docs/engine/outlines/DEV_PROCESS_MGMT_IMPLEMENTATION_OUTLINE.md`

- **Chunks**: 26
- **Headings**: 1. Feature Summary, 10. Completion Criteria, 2. Problem Definition and Motivation, 3. User Stories (v1), 4. Inputs and API Contract...

#### `docs/engine/outlines/DEV_TRAEFIK_IMPLEMENTATION_OUTLINE.md`

- **Chunks**: 42
- **Headings**: 1. Feature Summary, 10. Implementation Plan Checklist, 11. Completion Criteria, 12. Implementation Notes, 2. Problem Definition and Motivation...

#### `docs/engine/outlines/GOV_STATUS_ROADMAP_IMPLEMENTATION_OUTLINE.md`

- **Chunks**: 50
- **Headings**: 1. Feature Summary, 10. Dependencies, 11. Implementation Phases, 12. Success Metrics, 13. Approval...

#### `docs/engine/outlines/GOV_V1_CORE_IMPLEMENTATION_OUTLINE.md`

- **Chunks**: 24
- **Headings**: 1. Summary & v1 Scope, 1.1 Problem, 1.2 v1 Goal, 2. In / Out of Scope, 2.1 In Scope (v1)...

#### `docs/engine/outlines/GOV_V1_CORE_PHASE5_OUTLINE.md`

- **Chunks**: 20
- **Headings**: 1. Summary and v1 Scope, 1.1 Problem, 1.2 v1 Goal, 2. In / Out of Scope, 2.1 In Scope...

#### `docs/engine/outlines/IMPLEMENTATION_OUTLINE_TEMPLATE.md`

- **Chunks**: 26
- **Headings**: 1. Feature Summary, 10. Completion Criteria, 2. Problem Definition and Motivation, 3. User Stories (v1), 4. Inputs and CLI or API Contract...

#### `docs/engine/outlines/INFRA_HOST_BOOTSTRAP_IMPLEMENTATION_OUTLINE.md`

- **Chunks**: 31
- **Headings**: 1. Feature Summary and v1 Scope, 1.1 Summary, 1.2 v1 Scope (Included), 1.3 Out of Scope (v1), 1.4 Reserved for Future Versions...

#### `docs/engine/outlines/PROVIDER_CLOUD_DO_IMPLEMENTATION_OUTLINE.md`

- **Chunks**: 37
- **Headings**: 1. Feature Summary, 10. Registry Integration, 11. Config Schema Example, 12. Determinism Guarantees, 13. Implementation Checklist...

#### `docs/engine/outlines/PROVIDER_NETWORK_TAILSCALE_IMPLEMENTATION_OUTLINE.md`

- **Chunks**: 36
- **Headings**: 1. Feature Summary, 10. Registry Integration, 11. Config Schema Example, 12. Determinism Guarantees, 13. Implementation Checklist...

#### `docs/engine/status/feature-completion-analysis.md`

- **Chunks**: 24
- **Headings**: Architecture & Documentation, Critical Gaps, Critical Path Analysis, Detailed Phase Analysis, Executive Summary...

#### `docs/engine/status/implementation-status.md`

- **Chunks**: 14
- **Headings**: Architecture & Core, CLI Commands, Completed Features, Core Functionality, Coverage Status...

#### `docs/engine/status/infra-host-bootstrap-analysis.md`

- **Chunks**: 28
- **Headings**: Completed Prerequisites, Configuration Schema (Expected), Current Feature Completion State, Dependencies, Executive Summary...

#### `docs/engine/status/phase4-analysis.md`

- **Chunks**: 20
- **Headings**: 1. Planning Phase, 2. Implementation Phase (Sliced Approach), 3. Completion Phase, Cancelled Features, Completed Features...

#### `docs/engine/status/PROVIDER_COVERAGE_STATUS.md`

- **Chunks**: 13
- **Headings**: Coverage Status Summary, Governance Requirements, Next Actions, Priority 1: Complete PROVIDER_NETWORK_TAILSCALE, Provider Coverage Status...

#### `docs/engine/status/README.md`

- **Chunks**: 5
- **Headings**: CI Integration, Check if it changed, Files, Generate the file, Generation...

#### `docs/features/OVERVIEW.md`

- **Chunks**: 4
- **Headings**: Dependency Graph, Features by Domain, Stagecraft Features Overview, Status Summary

#### `docs/governance/CONTRIBUTING_CURSOR.md`

- **Chunks**: 12
- **Headings**: 1. Thread Hygiene, 1.1 One thread per feature / change, 1.2 Thread lifecycle, 2. File Hygiene (What to Open vs Attach), 2.1 Files to open during normal work...

#### `docs/governance/GOVERNANCE_ALMANAC.md`

- **Chunks**: 20
- **Headings**: 1. Purpose and Scope, 10. Migration Notes, 2. Core Governance Principles, 2.1 Governance Specs and ADRs, 3. Commit and PR Discipline...

#### `docs/guides/AI_COMMIT_WORKFLOW.md`

- **Chunks**: 19
- **Headings**: AI Commit Workflow Guide, Commands, Commit Health Report, Commit Suggestions, Feature Traceability Report...

#### `docs/guides/getting-started.md`

- **Chunks**: 18
- **Headings**: 1. Initialize a Project, 2. Configure Your Project, 3. Verify Configuration, Add to PATH (Optional), Add to your shell profile (~/.zshrc, ~/.bashrc, etc.)...

#### `docs/narrative/architecture.md`

- **Chunks**: 3
- **Headings**: Data Flow (High Level), Layers, Stagecraft Architecture

#### `docs/narrative/FUTURE_ENHANCEMENTS.md`

- **Chunks**: 47
- **Headings**: 1. Machine-Verifiable Spec Schema, 2. Structural Diff for Spec vs Implementation, 3. Feature Dependency Graph, 4. Full Feature Portal / Dashboard, 5. Automated Changelog Generation...

#### `docs/narrative/implementation-roadmap.md`

- **Chunks**: 42
- **Headings**: Architecture, Critical Path Features, Current Status Summary, Design Documents, Development Principles...

#### `docs/narrative/stagecraft-spec.md`

- **Chunks**: 14
- **Headings**: Backend, CI, Cloud, Commands, Core...

#### `docs/narrative/V2_FEATURES.md`

- **Chunks**: 13
- **Headings**: Contributing, Developer Experience, Environment Management, Implementation Timeline, Infrastructure Management...

#### `docs/providers/backend.md`

- **Chunks**: 20
- **Headings**: Architecture, Available Providers, Backend Providers, BackendProvider Interface, Best Practices...

#### `docs/providers/migrations.md`

- **Chunks**: 21
- **Headings**: Architecture, Available Engines, Best Practices, Configuration, Core Validation (Stagecraft)...

#### `docs/README.md`

- **Chunks**: 15
- **Headings**: Documentation Lifecycle, Documentation Structure, Documentation Workflow, Feature Governance Dashboard, Feature Mapping Check (GOV_V1_CORE Phase 4)...

#### `docs/reference/cli.md`

- **Chunks**: 1
- **Headings**: Stagecraft CLI Reference


### `spec/`

#### `spec/adr/0001-architecture.md`

- **Chunks**: 7
- **Headings**: 0001 – Stagecraft Architecture and Project Structure, Alternatives Considered, Consequences, Context, Core Structure (v1)...

#### `spec/commands/build.md`

- **Chunks**: 24
- **Headings**: 1. Summary, 2. CLI Definition, 2.1 Usage, 2.2 Required Flags, 2.3 Optional Flags...

#### `spec/commands/commit-suggest.md`

- **Chunks**: 40
- **Headings**: 1. Purpose, 10. Testing Requirements, 10.1 Unit Tests, 10.2 Integration Tests, 10.3 Golden Tests...

#### `spec/commands/deploy.md`

- **Chunks**: 18
- **Headings**: 1. Purpose, 10. Dependencies, 2. Scope, 3. CLI Interface, 3.1 Usage...

#### `spec/commands/dev-basic.md`

- **Chunks**: 24
- **Headings**: Basic Node.js App, Behaviour, CLI Usage, Command Structure, Config Resolution...

#### `spec/commands/dev.md`

- **Chunks**: 16
- **Headings**: Behaviour, CLI_DEV, Command, Configuration sources, Determinism...

#### `spec/commands/infra-up.md`

- **Chunks**: 37
- **Headings**: 1. Overview, 10. Error Conditions, 10.1 Global Errors, 10.2 Per-Host Errors, 11. Determinism Guarantees...

#### `spec/commands/init.md`

- **Chunks**: 8
- **Headings**: Behaviour, CLI Usage, Goal, Non-Goals (for initial version), Outputs...

#### `spec/commands/migrate-basic.md`

- **Chunks**: 26
- **Headings**: Basic Node.js App, Behaviour, CLI Usage, Command Structure, Config Resolution...

#### `spec/commands/plan.md`

- **Chunks**: 29
- **Headings**: 1. Purpose, 10. Dependencies, 11. Testing Requirements, 11.1 Golden Test Layout, 12. Implementation Notes...

#### `spec/commands/releases.md`

- **Chunks**: 21
- **Headings**: Behaviour, CLI Usage, Command Structure, Error Handling, Error Messages...

#### `spec/commands/rollback.md`

- **Chunks**: 32
- **Headings**: Behaviour, CLI Usage, Command Structure, Deploy Integration, Dry-run Semantics...

#### `spec/commands/status-roadmap.md`

- **Chunks**: 32
- **Headings**: Behavior, Blocker Detection, Command Structure, Dependencies, Deterministic Output...

#### `spec/core/backend-provider-config.md`

- **Chunks**: 19
- **Headings**: Backend Provider Configuration Schema, Benefits, Config Struct, Core Validation (Stagecraft), Encore.ts Provider...

#### `spec/core/backend-registry.md`

- **Chunks**: 12
- **Headings**: Architecture, Backend Provider Registry, Goal, Interface, Non-Goals...

#### `spec/core/compose.md`

- **Chunks**: 20
- **Headings**: API Design, Architecture, Behavior, Compose File Loading, Compose File Structure...

#### `spec/core/config.md`

- **Chunks**: 10
- **Headings**: Backend, Behavior, Core Config – Loading and Validation, Databases (Migration Configuration), Default Path...

#### `spec/core/env-resolution.md`

- **Chunks**: 13
- **Headings**: Behavior, Config Schema, Env File Parser, Environment Context, Environment Resolution...

#### `spec/core/executil.md`

- **Chunks**: 11
- **Headings**: API Design, Behavior, Command Execution, Core ExecUtil – Process Execution Utilities, Error Handling...

#### `spec/core/global-flags.md`

- **Chunks**: 11
- **Headings**: Behavior, Environment Variable Support, Flag Precedence, Flag Resolution, Flag Validation...

#### `spec/core/logging.md`

- **Chunks**: 11
- **Headings**: API Design, Behavior, Core Logging – Structured Logging Helpers, Global Flag Integration, Goal...

#### `spec/core/migration-registry.md`

- **Chunks**: 12
- **Headings**: Architecture, Engine Registration, Goal, Interface, Migration Engine Registry...

#### `spec/core/phase-execution-common.md`

- **Chunks**: 21
- **Headings**: 1. Phase Work Failure, 2. Planner Failure (Pre-Execution), 3. State Manager Failure (Status Updates), Behaviour, CORE_PHASE_EXECUTION_COMMON – Shared Phase Execution Semantics...

#### `spec/core/plan.md`

- **Chunks**: 11
- **Headings**: Architecture, Deployment Planning Engine, Example Plan, Future Enhancements, Goal...

#### `spec/core/state-consistency.md`

- **Chunks**: 17
- **Headings**: 1. Purpose, 2. Scope, 3. Consistency Model, 3.1 Read-after-write Guarantee (Single Process), 3.2 Multi-manager Behaviour...

#### `spec/core/state-test-isolation.md`

- **Chunks**: 12
- **Headings**: 1. Purpose, 2. Scope, 3. State Test Isolation Model, 3.1 Isolation Constraints, 3.2 Helper Function: `setupIsolatedStateTestEnv`...

#### `spec/core/state.md`

- **Chunks**: 13
- **Headings**: Behavior, Environment Variable Support, Goal, Interface, Non-Goals (v1)...

#### `spec/deploy/compose-gen.md`

- **Chunks**: 14
- **Headings**: 1. Purpose, 2. Scope, 3. Inputs and Outputs, 4. Behavior, 5. Integration...

#### `spec/deploy/rollout.md`

- **Chunks**: 10
- **Headings**: 1. Purpose, 2. Scope, 3. Compatibility Matrix, 4. Configuration, 5. Error Handling...

#### `spec/dev/compose-infra.md`

- **Chunks**: 10
- **Headings**: Behaviour, DEV_COMPOSE_INFRA, Determinism, Excluded (future), Included...

#### `spec/dev/hosts.md`

- **Chunks**: 19
- **Headings**: 1. Overview, 2. Behavior, 2.1 Hosts File Toggle, 2.2 Hosts File Paths, 2.3 Entry Format...

#### `spec/dev/mkcert.md`

- **Chunks**: 20
- **Headings**: 1. Overview, 2. Behavior, 2.1 HTTPS Toggle, 2.2 Certificate Directory, 2.3 Certificate Files...

#### `spec/dev/process-mgmt.md`

- **Chunks**: 15
- **Headings**: 1. Overview, 2. Behaviour, 2.1 Dev Files as Source of Truth, 2.2 Command Execution, 2.3 Foreground Mode (default)...

#### `spec/dev/traefik.md`

- **Chunks**: 10
- **Headings**: Behaviour, DEV_TRAEFIK, Determinism, Excluded (future), Included...

#### `spec/engine/plan-actions.md`

- **Chunks**: 22
- **Headings**: 1. Wire Format, 2. Determinism Rules (apply to all actions), 3. Validation Rules (apply to all actions), 4. Defaults Rules, 5. Forward Compatibility Rules...

#### `spec/governance/GOV_CLI_EXIT_CODES.md`

- **Chunks**: 14
- **Headings**: Common Exit Code Semantics, Determinism, Excluded (v1), Exit Code Documentation Rules, Exit Codes...

#### `spec/governance/GOV_V1_CORE.md`

- **Chunks**: 15
- **Headings**: 1. Summary, 2. Goals, 3. Non-Goals, 4. Design, 4.1 Spec Schema (Frontmatter)...

#### `spec/infra/bootstrap.md`

- **Chunks**: 28
- **Headings**: 1. Overview, 2. Responsibilities and Non-Goals, 2.1 Responsibilities (v1), 2.2 Non-Goals (v1), 3. Invocation and Execution Semantics...

#### `spec/overview.md`

- **Chunks**: 5
- **Headings**: Command Surface (Initial), Core Concepts, High-Level Goals, Non-Goals (for v0), Stagecraft – Project Overview

#### `spec/providers/backend/encore-ts.md`

- **Chunks**: 11
- **Headings**: 1. Goals and Non-Goals, 1.1 Goals, 1.2 Non-Goals, 2. Relationship to Core Backend Abstraction, 2.1 BackendProvider Interface...

#### `spec/providers/backend/generic.md`

- **Chunks**: 20
- **Headings**: Build Mode Behavior, Comparison with Other Providers, Config Parsing, Configuration, Dev Mode Behavior...

#### `spec/providers/ci/interface.md`

- **Chunks**: 11
- **Headings**: CI Provider Interface, Config Schema, Error Types, Goal, Interface...

#### `spec/providers/cloud/digitalocean.md`

- **Chunks**: 27
- **Headings**: 1. Overview, 10. Cost and Billing Responsibility, 2. Interface Contract, 2.1 ID, 2.2 Plan...

#### `spec/providers/cloud/interface.md`

- **Chunks**: 11
- **Headings**: Cloud Provider Interface, Config Schema, Error Types, Goal, Interface...

#### `spec/providers/frontend/generic.md`

- **Chunks**: 21
- **Headings**: Comparison with Other Providers, Config Parsing, Configuration, Dev Mode Behavior, Error Handling...

#### `spec/providers/frontend/interface.md`

- **Chunks**: 9
- **Headings**: Config Schema, Frontend Provider Interface, Goal, Interface, Non-Goals (v1)...

#### `spec/providers/migration/raw.md`

- **Chunks**: 23
- **Headings**: Comparison with Other Engines, Configuration, Core Validation (Stagecraft), Database Support, Engine-Specific Validation...

#### `spec/providers/network/interface.md`

- **Chunks**: 11
- **Headings**: Config Schema, Error Types, Goal, Interface, Network Provider Interface...

#### `spec/providers/network/tailscale.md`

- **Chunks**: 38
- **Headings**: 1. Overview, 10. Testing, 10.1 Unit Tests, 10.2 Integration Tests (Optional), 11. Non-Goals (v1)...

#### `spec/providers/secrets/interface.md`

- **Chunks**: 11
- **Headings**: Config Schema, Error Types, Goal, Interface, Non-Goals (v1)...

#### `spec/scaffold/stagecraft-dir.md`

- **Chunks**: 17
- **Headings**: .stagecraft/ Directory Structure, Behavior, Creation, Directory Structure, File Descriptions...



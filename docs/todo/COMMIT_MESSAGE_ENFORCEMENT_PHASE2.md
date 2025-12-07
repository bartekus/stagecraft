# TODO: Phase 2 – Commit Message CI Validation & CLI Tooling

- **Link:** `docs/COMMIT_MESSAGE_ANALYSIS.md`
- **Feature IDs:** `PROVIDER_FRONTEND_GENERIC`, `GOV_V1_CORE`
- **Status:** todo → wip
- **Prerequisite:** Phase 1 must be complete

## Scope

- Implement CI-level commit message validation (GitHub Actions job)
- Add optional CLI helper: `stagecraft validate-commit "<message>"`
- Integrate FEATURE_ID validation against `spec/features.yaml`
- Add orphan FEATURE_ID detection
- Add branch name / FEATURE_ID matching logic
- Enhance validation rules (multi-feature, format, length, etc.)
- Update documentation with Phase 2 behaviour

## Notes

- CI validation is mandatory for PRs
- CLI tool is optional (additive, not required)
- No changes to core Stagecraft runtime behaviour
- Fully backward-compatible with Phase 1 workflow
- Uses same validation logic in both CI and CLI for consistency

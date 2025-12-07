# Go Types for Commit Discipline Reports

**Purpose:** Define Go struct types that mirror the JSON schemas for commit-health.json and feature-traceability.json exactly.

**Location:** These types should be implemented in:
- `internal/reports/commithealth/types.go`
- `internal/reports/featuretrace/types.go`

⸻

## Commit Health Report Types

**Package:** `internal/reports/commithealth`

```go
package commithealth

// Report represents the complete commit health report.
type Report struct {
	SchemaVersion string            `json:"schema_version"`
	GeneratedAt   string            `json:"generated_at,omitempty"`
	Repo          RepoInfo          `json:"repo"`
	Range         CommitRange       `json:"range"`
	Summary       Summary           `json:"summary"`
	Rules         []Rule            `json:"rules"`
	Commits       map[string]Commit `json:"commits"`
}

// RepoInfo contains repository metadata.
type RepoInfo struct {
	Name          string `json:"name"`
	DefaultBranch string `json:"default_branch"`
}

// CommitRange describes the commit range analyzed.
type CommitRange struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Description string `json:"description"`
}

// Summary contains aggregate statistics.
type Summary struct {
	TotalCommits    int            `json:"total_commits"`
	ValidCommits    int            `json:"valid_commits"`
	InvalidCommits  int            `json:"invalid_commits"`
	ViolationsByCode map[string]int `json:"violations_by_code"`
}

// Rule describes a commit validation rule.
type Rule struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Severity    string `json:"severity"` // "info", "warning", "error"
}

// Commit represents a single commit's health status.
type Commit struct {
	Subject   string     `json:"subject"`
	IsValid   bool       `json:"is_valid"`
	Violations []Violation `json:"violations"`
}

// Violation represents a single validation violation.
type Violation struct {
	Code     string         `json:"code"`
	Severity string         `json:"severity"` // "info", "warning", "error"
	Message  string         `json:"message"`
	Details  map[string]any `json:"details"`
}

// ViolationCode represents known violation codes.
type ViolationCode string

const (
	ViolationCodeMissingFeatureID          ViolationCode = "MISSING_FEATURE_ID"
	ViolationCodeInvalidType               ViolationCode = "INVALID_TYPE"
	ViolationCodeInvalidFeatureIDFormat     ViolationCode = "INVALID_FEATURE_ID_FORMAT"
	ViolationCodeFeatureIDNotInSpec        ViolationCode = "FEATURE_ID_NOT_IN_SPEC"
	ViolationCodeFeatureIDBranchMismatch   ViolationCode = "FEATURE_ID_BRANCH_MISMATCH"
	ViolationCodeMultipleFeatureIDs        ViolationCode = "MULTIPLE_FEATURE_IDS"
	ViolationCodeSummaryTooLong            ViolationCode = "SUMMARY_TOO_LONG"
	ViolationCodeSummaryHasTrailingPeriod  ViolationCode = "SUMMARY_HAS_TRAILING_PERIOD"
	ViolationCodeSummaryStartsWithUppercase ViolationCode = "SUMMARY_STARTS_WITH_UPPERCASE"
	ViolationCodeSummaryHasUnicode         ViolationCode = "SUMMARY_HAS_UNICODE"
	ViolationCodeInvalidFormatGeneric      ViolationCode = "INVALID_FORMAT_GENERIC"
)

// Severity represents violation severity levels.
type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)
```

⸻

## Feature Traceability Report Types

**Package:** `internal/reports/featuretrace`

```go
package featuretrace

// Report represents the complete feature traceability report.
type Report struct {
	SchemaVersion string             `json:"schema_version"`
	GeneratedAt   string             `json:"generated_at,omitempty"`
	Summary       Summary            `json:"summary"`
	Features      map[string]Feature `json:"features"`
}

// Summary contains aggregate statistics.
type Summary struct {
	TotalFeatures     int `json:"total_features"`
	Done              int `json:"done"`
	WIP               int `json:"wip"`
	Todo              int `json:"todo"`
	Deprecated        int `json:"deprecated"`
	Removed           int `json:"removed"`
	FeaturesWithGaps int `json:"features_with_gaps"`
}

// Feature represents traceability information for a single feature.
type Feature struct {
	Status       FeatureStatus `json:"status"`
	Spec         SpecInfo      `json:"spec"`
	Implementation ImplementationInfo `json:"implementation"`
	Tests        TestsInfo     `json:"tests"`
	Commits      CommitsInfo   `json:"commits"`
	Problems     []Problem     `json:"problems"`
}

// FeatureStatus represents the lifecycle state of a feature.
type FeatureStatus string

const (
	FeatureStatusTodo       FeatureStatus = "todo"
	FeatureStatusWIP        FeatureStatus = "wip"
	FeatureStatusDone       FeatureStatus = "done"
	FeatureStatusDeprecated FeatureStatus = "deprecated"
	FeatureStatusRemoved    FeatureStatus = "removed"
)

// SpecInfo describes spec file presence and location.
type SpecInfo struct {
	Present bool   `json:"present"`
	Path    string `json:"path"` // empty string if not present
}

// ImplementationInfo describes implementation file presence and locations.
type ImplementationInfo struct {
	Present bool     `json:"present"`
	Files   []string `json:"files"` // sorted list of file paths
}

// TestsInfo describes test file presence and locations.
type TestsInfo struct {
	Present bool     `json:"present"`
	Files   []string `json:"files"` // sorted list of test file paths
}

// CommitsInfo describes commit presence and SHAs.
type CommitsInfo struct {
	Present bool     `json:"present"`
	SHAs    []string `json:"shas"` // sorted list of commit SHAs
}

// Problem represents a traceability problem for a feature.
type Problem struct {
	Code     string         `json:"code"`
	Severity string         `json:"severity"` // "info", "warning", "error"
	Message  string         `json:"message"`
	Details  map[string]any `json:"details"`
}

// ProblemCode represents known problem codes.
type ProblemCode string

const (
	ProblemCodeMissingSpec                    ProblemCode = "MISSING_SPEC"
	ProblemCodeMissingImplementation          ProblemCode = "MISSING_IMPLEMENTATION"
	ProblemCodeMissingTests                   ProblemCode = "MISSING_TESTS"
	ProblemCodeMissingCommits                 ProblemCode = "MISSING_COMMITS"
	ProblemCodeOrphanSpec                    ProblemCode = "ORPHAN_SPEC"
	ProblemCodeOrphanFeatureIDInCommits      ProblemCode = "ORPHAN_FEATURE_ID_IN_COMMITS"
	ProblemCodeStatusDoneButMissingTests     ProblemCode = "STATUS_DONE_BUT_MISSING_TESTS"
	ProblemCodeStatusDoneButMissingImplementation ProblemCode = "STATUS_DONE_BUT_MISSING_IMPLEMENTATION"
	ProblemCodeUnreferencedSpecPath          ProblemCode = "UNREFERENCED_SPEC_PATH"
)

// Severity represents problem severity levels.
type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)
```

⸻

## Implementation Notes

### Determinism Requirements

1. **Key Sorting:**
   - When marshaling to JSON, use `json.Marshal` with sorted keys
   - For maps, iterate keys in sorted order when building the report
   - For arrays, ensure consistent ordering (already sorted in struct definitions)

2. **Stable Output:**
   - Same inputs must produce identical JSON output
   - Use deterministic iteration order for maps
   - Sort all string arrays before marshaling

3. **Testing:**
   - Golden tests should compare exact JSON output
   - Use `json.Compact` to normalize whitespace for comparison
   - Do not assert on `generated_at` field values

### Example Usage

```go
// Generate commit health report
report := commithealth.Report{
    SchemaVersion: "1.0",
    Repo: commithealth.RepoInfo{
        Name:          "stagecraft",
        DefaultBranch: "main",
    },
    Range: commithealth.CommitRange{
        From:        "origin/main",
        To:          "HEAD",
        Description: "origin/main..HEAD",
    },
    Summary: commithealth.Summary{
        TotalCommits:   12,
        ValidCommits:   10,
        InvalidCommits: 2,
        ViolationsByCode: map[string]int{
            "MISSING_FEATURE_ID": 1,
            "MULTIPLE_FEATURE_IDS": 1,
        },
    },
    Rules: []commithealth.Rule{
        {
            Code:        "MISSING_FEATURE_ID",
            Description: "Commit message is missing a Feature ID",
            Severity:    "error",
        },
    },
    Commits: map[string]commithealth.Commit{
        "abc123": {
            Subject:   "feat(CLI_DEPLOY): add rollback support",
            IsValid:   true,
            Violations: []commithealth.Violation{},
        },
    },
}

// Marshal with sorted keys
data, err := json.Marshal(report)
```

⸻

## File Structure

```
internal/
  reports/
    commithealth/
      types.go          # Type definitions
      generator.go      # Report generation logic
      generator_test.go
    featuretrace/
      types.go          # Type definitions
      generator.go      # Report generation logic
      generator_test.go
```

These types provide a 1:1 mapping to the JSON schemas, making implementation straightforward and ensuring deterministic output.


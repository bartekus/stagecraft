# Agent: PROVIDER_NETWORK_TAILSCALE_SLICE1

## Scope

Implement **Slice 1** of the Tailscale provider coverage plan.

**Goal**:  
- Extract pure helper functions from the Tailscale provider.  
- Add deterministic unit tests (no network, no processes, no sleeps).  
- Improve coverage from ~68.2% to ~75% without touching CLI invocation tests.

**Reference plan**: `docs/governance/PROVIDER_NETWORK_TAILSCALE_SLICE1_PLAN.md`

---

## Rules

- Do **not** change behavior for existing public methods beyond factoring out helpers.
- No `time.Sleep`, no real Tailscale CLI calls in tests.
- Tests must be deterministic and pass with:
  - `go test -cover ./internal/providers/network/tailscale`
  - `go test -race ./internal/providers/network/tailscale`
  - `go test -count=20 ./internal/providers/network/tailscale`
- Follow existing code style and test patterns in `internal/providers/network/tailscale`.
- Keep helpers **package-local** unless there is a strong reason to expose them.

---

## Tasks

### 1. Extract Pure Helpers

In `internal/providers/network/tailscale/tailscale.go`:

**1.1 Add `buildTailscaleUpCommand`:**

```go
// buildTailscaleUpCommand builds the Tailscale "up" command string.
// This is a pure function that takes explicit inputs and returns a command string.
func buildTailscaleUpCommand(authKey, hostname string, tags []string) string {
	tagArgs := strings.Join(tags, ",")
	return fmt.Sprintf("tailscale up --authkey=%s --hostname=%s --advertise-tags=%s",
		authKey, hostname, tagArgs)
}
```

**1.2 Add `parseOSRelease`:**

```go
// parseOSRelease parses the ID field from /etc/os-release content.
// Returns the distribution ID (e.g., "debian", "ubuntu") or empty string if not found.
// This is a pure function that operates on string content only.
func parseOSRelease(osReleaseContent string) string {
	lines := strings.Split(osReleaseContent, "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "ID=") {
			continue
		}
		id := strings.TrimPrefix(line, "ID=")
		id = strings.Trim(id, `"`)
		id = strings.ToLower(id)
		return id
	}
	return ""
}
```

**1.3 Add `validateTailnetDomain`:**

```go
// validateTailnetDomain validates that a Tailnet domain is non-empty and has valid format.
// Returns an error if the domain is invalid.
func validateTailnetDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("tailscale provider: %w: tailnet_domain is required", ErrConfigInvalid)
	}
	// Basic validation: should contain at least one dot
	if !strings.Contains(domain, ".") {
		return fmt.Errorf("tailscale provider: %w: tailnet_domain %q must contain a dot", ErrConfigInvalid, domain)
	}
	return nil
}
```

**1.4 Add `buildNodeFQDN`:**

```go
// buildNodeFQDN builds the FQDN for a Tailscale node.
// This is a pure function: host + domain = FQDN.
func buildNodeFQDN(host, domain string) string {
	return fmt.Sprintf("%s.%s", host, domain)
}
```

**Note**: Ensure needed imports (`fmt`, `strings`) are already present or add them.

---

### 2. Refactor Existing Code to Use Helpers

**2.1 In `EnsureJoined()`:**

Replace line 153:
```go
// OLD:
tagArgs := strings.Join(tags, ",")
joinCmd := fmt.Sprintf("tailscale up --authkey=%s --hostname=%s --advertise-tags=%s",
	authKey, opts.Host, tagArgs)

// NEW:
joinCmd := buildTailscaleUpCommand(authKey, opts.Host, tags)
```

**2.2 In `checkOSCompatibility()`:**

Replace lines 283-296 (the inline parsing):
```go
// OLD:
lines := strings.Split(osRelease, "\n")
for _, line := range lines {
	if !strings.HasPrefix(line, "ID=") {
		continue
	}
	id := strings.TrimPrefix(line, "ID=")
	id = strings.Trim(id, `"`)
	id = strings.ToLower(id)
	if id == "debian" || id == "ubuntu" {
		return nil
	}
	return fmt.Errorf("tailscale provider: %w: detected distribution %q, v1 supports Debian/Ubuntu only",
		ErrUnsupportedOS, id)
}

// NEW:
id := parseOSRelease(osRelease)
if id == "debian" || id == "ubuntu" {
	return nil
}
if id != "" {
	return fmt.Errorf("tailscale provider: %w: detected distribution %q, v1 supports Debian/Ubuntu only",
		ErrUnsupportedOS, id)
}
```

**2.3 In `NodeFQDN()`:**

Replace lines 201-205:
```go
// OLD:
if p.config.TailnetDomain == "" {
	return "", fmt.Errorf("tailscale provider: %w: tailnet_domain is required", ErrConfigInvalid)
}
return fmt.Sprintf("%s.%s", host, p.config.TailnetDomain), nil

// NEW:
if err := validateTailnetDomain(p.config.TailnetDomain); err != nil {
	return "", err
}
return buildNodeFQDN(host, p.config.TailnetDomain), nil
```

**Ensure behavior remains equivalent**; only move logic into helpers.

---

### 3. Add Unit Tests for Helpers

In `internal/providers/network/tailscale/tailscale_test.go`:

**3.1 `TestBuildTailscaleUpCommand`:**

```go
func TestBuildTailscaleUpCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		authKey  string
		hostname string
		tags     []string
		want     string
	}{
		{
			name:     "single tag",
			authKey:  "tskey-auth-123",
			hostname: "app-1",
			tags:     []string{"tag:web"},
			want:     "tailscale up --authkey=tskey-auth-123 --hostname=app-1 --advertise-tags=tag:web",
		},
		{
			name:     "multiple tags",
			authKey:  "tskey-auth-123",
			hostname: "app-1",
			tags:     []string{"tag:web", "tag:prod"},
			want:     "tailscale up --authkey=tskey-auth-123 --hostname=app-1 --advertise-tags=tag:web,tag:prod",
		},
		{
			name:     "no tags",
			authKey:  "tskey-auth-123",
			hostname: "app-1",
			tags:     []string{},
			want:     "tailscale up --authkey=tskey-auth-123 --hostname=app-1 --advertise-tags=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildTailscaleUpCommand(tt.authKey, tt.hostname, tt.tags)
			if got != tt.want {
				t.Errorf("buildTailscaleUpCommand() = %q, want %q", got, tt.want)
			}
		})
	}
}
```

**3.2 `TestParseOSRelease`:**

```go
func TestParseOSRelease(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name: "debian",
			content: `PRETTY_NAME="Debian GNU/Linux 11 (bullseye)"
NAME="Debian GNU/Linux"
ID=debian
ID_LIKE=debian`,
			want: "debian",
		},
		{
			name: "ubuntu",
			content: `NAME="Ubuntu"
VERSION="22.04.3 LTS (Jammy Jellyfish)"
ID=ubuntu
ID_LIKE=debian`,
			want: "ubuntu",
		},
		{
			name:    "quoted ID",
			content: `ID="debian"`,
			want:    "debian",
		},
		{
			name:    "no ID field",
			content: `PRETTY_NAME="Some OS"`,
			want:    "",
		},
		{
			name:    "empty content",
			content: "",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseOSRelease(tt.content)
			if got != tt.want {
				t.Errorf("parseOSRelease() = %q, want %q", got, tt.want)
			}
		})
	}
}
```

**3.3 `TestValidateTailnetDomain`:**

```go
func TestValidateTailnetDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		domain  string
		wantErr bool
	}{
		{
			name:    "valid domain",
			domain:  "example.ts.net",
			wantErr: false,
		},
		{
			name:    "valid subdomain",
			domain:  "sub.example.ts.net",
			wantErr: false,
		},
		{
			name:    "empty domain",
			domain:  "",
			wantErr: true,
		},
		{
			name:    "no dot",
			domain:  "example",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTailnetDomain(tt.domain)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTailnetDomain() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
```

**3.4 `TestBuildNodeFQDN`:**

```go
func TestBuildNodeFQDN(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		host   string
		domain string
		want   string
	}{
		{
			name:   "simple host",
			host:   "app-1",
			domain: "example.ts.net",
			want:   "app-1.example.ts.net",
		},
		{
			name:   "host with dash",
			host:   "db-primary",
			domain: "example.ts.net",
			want:   "db-primary.example.ts.net",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildNodeFQDN(tt.host, tt.domain)
			if got != tt.want {
				t.Errorf("buildNodeFQDN() = %q, want %q", got, tt.want)
			}
		})
	}
}
```

---

### 4. Add `parseStatus` Edge-Case Tests

In `tailscale_test.go` (or create `status_test.go` if preferred):

**Note**: `parseStatus()` currently only validates JSON syntax, not required fields. Empty JSON `{}` will succeed but return empty struct. Test what actually happens:

```go
func TestParseStatus_InvalidJSON(t *testing.T) {
	t.Parallel()

	_, err := parseStatus("not json")
	if err == nil {
		t.Error("parseStatus() should return error for invalid JSON")
	}
}

func TestParseStatus_EmptyJSON(t *testing.T) {
	t.Parallel()

	// Empty JSON will unmarshal successfully but return empty struct
	status, err := parseStatus("{}")
	if err != nil {
		t.Fatalf("parseStatus() with empty JSON returned error: %v", err)
	}
	if status == nil {
		t.Error("parseStatus() returned nil status")
	}
	// Verify it's actually empty
	if status.TailnetName != "" {
		t.Errorf("parseStatus() with empty JSON: TailnetName = %q, want empty", status.TailnetName)
	}
}

func TestParseStatus_ValidStatus(t *testing.T) {
	t.Parallel()

	jsonData := `{
		"TailnetName": "example.ts.net",
		"Self": {
			"Online": true,
			"TailscaleIPs": ["100.64.0.1"],
			"Tags": ["tag:web"]
		}
	}`

	status, err := parseStatus(jsonData)
	if err != nil {
		t.Fatalf("parseStatus() returned error: %v", err)
	}
	if status.TailnetName != "example.ts.net" {
		t.Errorf("parseStatus() TailnetName = %q, want %q", status.TailnetName, "example.ts.net")
	}
	if len(status.Self.Tags) != 1 || status.Self.Tags[0] != "tag:web" {
		t.Errorf("parseStatus() Self.Tags = %v, want [tag:web]", status.Self.Tags)
	}
}
```

---

### 5. Run Verification Commands

From repo root:

```bash
# Check coverage increase
go test -cover ./internal/providers/network/tailscale

# Verify determinism
go test -race ./internal/providers/network/tailscale
go test -count=20 ./internal/providers/network/tailscale

# Verify governance
./scripts/check-provider-governance.sh
```

**Confirm**: Coverage increased from ~68.2% toward ~75%.

---

### 6. Update Documentation

**6.1 `internal/providers/network/tailscale/COVERAGE_STRATEGY.md`:**

- Update coverage percentage (use actual value from `go test -cover`)
- Add note: "Slice 1 complete: extracted 4 pure helper functions and added comprehensive unit tests"

**6.2 `docs/engine/status/PROVIDER_COVERAGE_STATUS.md`:**

- Update coverage percentage for `PROVIDER_NETWORK_TAILSCALE`

**6.3 `docs/engine/status/PROVIDER_NETWORK_TAILSCALE_COVERAGE_PLAN.md`:**

- Add note: "Slice 1 complete: helper extraction + deterministic unit tests"

---

### 7. Commit

```bash
git add \
  internal/providers/network/tailscale/tailscale.go \
  internal/providers/network/tailscale/tailscale_test.go \
  internal/providers/network/tailscale/status.go \
  internal/providers/network/tailscale/COVERAGE_STRATEGY.md \
  docs/engine/status/PROVIDER_COVERAGE_STATUS.md \
  docs/engine/status/PROVIDER_NETWORK_TAILSCALE_COVERAGE_PLAN.md

git commit -m "test(PROVIDER_NETWORK_TAILSCALE): extract helpers and add unit tests

- Extract buildTailscaleUpCommand() helper
- Extract parseOSRelease() helper
- Extract validateTailnetDomain() helper
- Extract buildNodeFQDN() helper
- Add comprehensive unit tests for all helpers
- Add edge case tests for parseStatus()
- Coverage: 68.2% → ~75% (Slice 1 complete)"
```

---

## Success Criteria

- ✅ 4 pure helper functions extracted
- ✅ All helpers have comprehensive unit tests
- ✅ Existing code refactored to use helpers (behavior unchanged)
- ✅ Coverage increases from 68.2% → ~75%
- ✅ All tests pass with `-race` and `-count=20`
- ✅ No CLI invocation tests added (deferred to later slices)
- ✅ Documentation updated

---

## Reference

- Plan: `docs/governance/PROVIDER_NETWORK_TAILSCALE_SLICE1_PLAN.md`
- Coverage Strategy: `internal/providers/network/tailscale/COVERAGE_STRATEGY.md`
- Test Requirements: `docs/governance/GOV_V1_TEST_REQUIREMENTS.md`
- Reference Model: `internal/providers/frontend/generic/COVERAGE_STRATEGY.md`

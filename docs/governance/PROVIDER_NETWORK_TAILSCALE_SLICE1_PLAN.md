# PROVIDER_NETWORK_TAILSCALE - Slice 1: Extract Helpers & Unit Tests

**Feature**: PROVIDER_NETWORK_TAILSCALE  
**Current Coverage**: 68.2%  
**Target for Slice 1**: ~75% (extract helpers, add unit tests)  
**Final Target**: ≥80% (in later slices)

---

## Goal

Extract pure helper functions from orchestration code and add deterministic unit tests. This front-loads easy coverage gains before tackling CLI invocation tests.

---

## Helper Functions to Extract

### 1. `buildTailscaleUpCommand` (NEW)

**Location**: Extract from `EnsureJoined()` line 153

**Current code**:
```go
joinCmd := fmt.Sprintf("tailscale up --authkey=%s --hostname=%s --advertise-tags=%s",
    authKey, opts.Host, tagArgs)
```

**Extracted function**:
```go
// buildTailscaleUpCommand builds the Tailscale "up" command string.
// This is a pure function that takes explicit inputs and returns a command string.
func buildTailscaleUpCommand(authKey, hostname string, tags []string) string {
    tagArgs := strings.Join(tags, ",")
    return fmt.Sprintf("tailscale up --authkey=%s --hostname=%s --advertise-tags=%s",
        authKey, hostname, tagArgs)
}
```

**Test skeleton**:
```go
func TestBuildTailscaleUpCommand(t *testing.T) {
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

### 2. `parseOSRelease` (NEW)

**Location**: Extract from `checkOSCompatibility()` lines 283-296

**Current code**: Inline parsing of `/etc/os-release` content

**Extracted function**:
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

**Test skeleton**:
```go
func TestParseOSRelease(t *testing.T) {
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
            name: "quoted ID",
            content: `ID="debian"`,
            want: "debian",
        },
        {
            name: "no ID field",
            content: `PRETTY_NAME="Some OS"`,
            want: "",
        },
        {
            name: "empty content",
            content: "",
            want: "",
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

### 3. `validateTailnetDomain` (NEW)

**Location**: Extract validation logic from `NodeFQDN()` and `EnsureJoined()`

**Extracted function**:
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

**Test skeleton**:
```go
func TestValidateTailnetDomain(t *testing.T) {
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

### 4. `buildNodeFQDN` (NEW)

**Location**: Extract from `NodeFQDN()` line 205

**Extracted function**:
```go
// buildNodeFQDN builds the FQDN for a Tailscale node.
// This is a pure function: host + domain = FQDN.
func buildNodeFQDN(host, domain string) string {
    return fmt.Sprintf("%s.%s", host, domain)
}
```

**Test skeleton**:
```go
func TestBuildNodeFQDN(t *testing.T) {
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

## Existing Helpers to Add Tests For

### 5. `parseStatus` (EXISTING - needs more tests)

**Location**: `status.go:40`

**Current coverage**: 75.0%

**Additional test cases needed**:
- Invalid JSON
- Missing required fields
- Empty JSON
- Malformed status structure

**Test skeleton**:
```go
func TestParseStatus_InvalidJSON(t *testing.T) {
    _, err := parseStatus("not json")
    if err == nil {
        t.Error("parseStatus() should return error for invalid JSON")
    }
}

func TestParseStatus_EmptyJSON(t *testing.T) {
    _, err := parseStatus("{}")
    if err == nil {
        t.Error("parseStatus() should return error for empty JSON")
    }
}

func TestParseStatus_MissingFields(t *testing.T) {
    _, err := parseStatus(`{"Self": {}}`)
    if err == nil {
        t.Error("parseStatus() should return error for missing required fields")
    }
}
```

---

## Implementation Steps

### Step 1: Extract Helper Functions

1. Add `buildTailscaleUpCommand()` to `tailscale.go`
2. Add `parseOSRelease()` to `tailscale.go`
3. Add `validateTailnetDomain()` to `tailscale.go`
4. Add `buildNodeFQDN()` to `tailscale.go`

### Step 2: Refactor Existing Code

1. Update `EnsureJoined()` to use `buildTailscaleUpCommand()`
2. Update `checkOSCompatibility()` to use `parseOSRelease()`
3. Update `NodeFQDN()` to use `validateTailnetDomain()` and `buildNodeFQDN()`

### Step 3: Add Unit Tests

1. Add tests for all 4 new helpers
2. Add additional tests for `parseStatus()` edge cases
3. Run tests to verify coverage increase

### Step 4: Verify

```bash
go test -cover ./internal/providers/network/tailscale
go test -race ./internal/providers/network/tailscale
go test -count=20 ./internal/providers/network/tailscale
./scripts/check-provider-governance.sh
```

### Step 5: Update Documentation

1. Update `COVERAGE_STRATEGY.md` with new coverage percentage
2. Update `PROVIDER_COVERAGE_STATUS.md`
3. Add note to `PROVIDER_NETWORK_TAILSCALE_COVERAGE_PLAN.md`: "Slice 1 complete: helper extraction + deterministic unit tests"

### Step 6: Commit

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

## Expected Coverage Increase

**Before**: 68.2%  
**After**: ~75% (target for Slice 1)

**Functions improved**:
- `buildTailscaleUpCommand()` - NEW, 100% coverage
- `parseOSRelease()` - NEW, 100% coverage
- `validateTailnetDomain()` - NEW, 100% coverage
- `buildNodeFQDN()` - NEW, 100% coverage
- `parseStatus()` - 75% → ~85% (with edge case tests)
- `checkOSCompatibility()` - 60% → ~70% (uses extracted `parseOSRelease()`)
- `EnsureJoined()` - 64.9% → ~70% (uses extracted `buildTailscaleUpCommand()`)
- `NodeFQDN()` - 80% → ~85% (uses extracted helpers)

---

## Success Criteria

- ✅ 4 new pure helper functions extracted
- ✅ All helpers have comprehensive unit tests
- ✅ Coverage increases from 68.2% → ~75%
- ✅ All tests pass with `-race` and `-count=20`
- ✅ No CLI invocation tests added (deferred to later slices)
- ✅ Documentation updated

---

## Next Slices

**Slice 2**: Add error path tests for `EnsureInstalled()` and `EnsureJoined()` (mock Commander)
**Slice 3**: Add integration-style tests for CLI invocation (if needed to reach ≥80%)

---

## Reference

- Coverage Strategy: `internal/providers/network/tailscale/COVERAGE_STRATEGY.md`
- Test Requirements: `docs/governance/GOV_V1_TEST_REQUIREMENTS.md`
- Coverage Agent: `docs/engine/agents/PROVIDER_COVERAGE_AGENT.md`
- Reference Model: `internal/providers/frontend/generic/COVERAGE_STRATEGY.md`

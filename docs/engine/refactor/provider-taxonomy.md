# Blueprint: Provider Taxonomy Refactor

**Status:** Draft
**Target Version:** v1-Refactor
**Related:** `docs/narrative/resources-are-the-wrong-abstraction.md`

## 1. Target Provider Interface Schema

### 1.1 Normative Concepts

We move from "Managing Resources" to "Synthesizing Capabilities".

*   **Fact:** An observed immutable truth about a target (e.g., "File /etc/hosts exists").
*   **Capability:** A high-level property that a target *can* possess (e.g., "Network Mesh Connectivity").
*   **Transition:** A pure function that generates an *action* to move from State A to State B.

### 1.2 Provider Manifest Schema (`provider.yaml`)

Every provider must expose a manifest describing what it offers.

```yaml
apiVersion: stagecraft.io/v1alpha1
kind: Provider
metadata:
  name: "tailscale"
  version: "1.0.0"

class: CapabilityProvider # or ExecutionProvider, DiscoveryProvider

capabilities:
  - id: "capability.network.mesh.v1"
    description: "Connectivity to the private mesh network"
    schema:
      type: object
      properties:
        auth_key: { type: string }

transitions:
  - id: "install-client"
    goal: "Install Tailscale Client"
    requires:
      - "capability.os.linux"
    ensures:
      - "fact.tailscale.installed"
    inputs:
      - allow_unsupported_distros: boolean

  - id: "join-mesh"
    goal: "Join Tailscale Mesh"
    requires:
      - "fact.tailscale.installed"
    ensures:
      - "capability.network.mesh.v1"
    failureModes:
      - "auth_key_invalid"
      - "network_timeout"
```

### 1.3 Engine-Facing Go Interfaces

These interfaces belong in `pkg/taxonomy/`.

#### Execution Provider (The Hands)
```go
// ExecutionProvider executes low-level instructions on a target.
type ExecutionProvider interface {
    ID() string
    // Execute runs a command and returns the result (stdout/err/exit).
    Execute(ctx context.Context, cmd Request) (Result, error)
}
```

#### Discovery Provider (The Eyes)
```go
// DiscoveryProvider gathers facts without side effects.
type DiscoveryProvider interface {
    ID() string
    // Discover runs probes to gather facts matching the criteria.
    Discover(ctx context.Context, executor ExecutionProvider, criteria ProbeCriteria) ([]Fact, error)
}
```

#### Capability Provider (The Brains)
```go
// CapabilityProvider defines the logic for transitions.
type CapabilityProvider interface {
    ID() string
    
    // Plan returns the set of transitions needed to achieve the desired state
    // given the current facts. It is PURE.
    Plan(desired DesiredState, observed Facts) ([]TransitionPlan, error)
    
    // We do NOT have an Apply() method. 
    // Instead, the Engine orchestrates the execution of the TransitionPlan
    // using the ExecutionProvider.
}
```

---

## 2. Mapping Table (Existing → New Taxonomy)

| Existing Component | Current Role | Target Taxonomy Class | Output / Facts | Refactor Difficulty |
| :--- | :--- | :--- | :--- | :--- |
| **Tailscale Provider** | Ops + Install | **Capability Provider** | *Facts:* `tailscale.installed`, `tailscale.joined`<br>*Caps:* `network.mesh` | Medium. Extract `EnsureInstalled` logic into pure plans + shell commands. |
| **DigitalOcean** | Cloud + API | **Capability Provider** (Droplet)<br>**Discovery Provider** (Read API) | *Facts:* `cloud.instance.id`, `cloud.ip.public` | Low. Split `Plan` (API Read) from `Apply` (API Write). |
| **Bootstrap** | Service Orchestrator | **Composite Capability** | *Ensures:* `os.ready`, `docker.installed` | High. Rewire as a set of logical requirements rather than procedural code. |
| **Container Runner** | (Planned) | **Execution Provider** | N/A | New Implementation. |
| **SSH Executor** | `executil` wrapper | **Execution Provider** | N/A | Medium. Formalize existing `executil` usage. |
| **Git Provider** | Content Gen | **Integration Provider** | *Artifact:* `.github/workflows/*` | Low. View as "File Generation" capability. |

---

## 3. Migration Slices (Minimal-Change Path)

We will execute this refactor in **vertical slices** to avoid a "big bang" rewrite.

### Slice 1: Taxonomy Foundation
**Goal:** Establish the types and interfaces without changing behavior.
- **Feature ID:** `TAXONOMY_CORE_TYPES`
- **Files:**
    - `[NEW] pkg/taxonomy/types.go` (Fact, Transition, structs)
    - `[NEW] pkg/taxonomy/interfaces.go` (Provider interfaces)
- **Tests:** Unit tests for `Fact` logic (merging, diffing).

### Slice 2: The Execution Substrate
**Goal:** Create the `ExecutionProvider` and a Local implementation.
- **Feature ID:** `TAXONOMY_EXECUTION_LOCAL`
- **Files:**
    - `[NEW] internal/taxonomy/execution/local/local.go`
- **Tests:**
    - Verify `Execute("echo hello")` returns structured result.
- **Backwards Compat:** `executil` remains for old code.

### Slice 3: Discovery "hello-world"
**Goal:** Implement a simple Discovery Provider that uses the Execution Provider.
- **Feature ID:** `TAXONOMY_DISCOVERY_BASIC`
- **Files:**
    - `[NEW] internal/taxonomy/discovery/shell/shell.go`
- **Logic:** `Discover` takes a shell command ("which docker"), runs it via ExecutionProvider, parses output into a Fact.

### Slice 4: Tailscale Migration (Pilot)
**Goal:** Convert **one** real provider to verify the model.
- **Feature ID:** `TAXONOMY_MIGRATION_TAILSCALE`
- **Files:**
    - `[MODIFY] internal/providers/network/tailscale/tailscale.go` (Implement `CapabilityProvider`)
    - `[NEW] internal/providers/network/tailscale/discovery.go` (Fact definitions)
- **Acceptance Criteria:**
    - `Plan()` returns "Install" transition if fact "tailscale.installed" is missing.
    - `Plan()` is deterministic (unit testable without host).
- **Rollback:** Keep `network.NetworkProvider` interface wrapper calling the new logic.

### Slice 5: Boostrap Re-Alignment
**Goal:** Make `infra/bootstrap` use the new Taxonomy for the `Docker` check.
- **Feature ID:** `TAXONOMY_MIGRATION_BOOTSTRAP`
- **Files:**
    - `[MODIFY] internal/infra/bootstrap/service.go`
- **Changes:**
    - Replace direct `executil` checks with `DiscoveryProvider.Discover()`.
    - Replace procedural installs with `ExecutionProvider.Execute(Transition.Action)`.

### Slice 6: Remote Execution (SSH)
**Goal:** Unlock multi-host capability.
- **Feature ID:** `TAXONOMY_EXECUTION_SSH`
- **Files:**
    - `[NEW] internal/taxonomy/execution/ssh/ssh.go`
- **Verification:** Run Slice 4 (Tailscale) against a remote host using Slice 6 (SSH).

## 4. Feature ID Updates (`spec/features.yaml`)

We will add a new group `Phase 11: Taxonomy Refactor`.

```yaml
- id: TAXONOMY_CORE_TYPES
  title: "Core types for Capability/Transition taxonomy"
  status: todo

- id: TAXONOMY_EXECUTION_LOCAL
  title: "Local ExecutionProvider implementation"
  status: todo
  depends_on: [TAXONOMY_CORE_TYPES]

- id: TAXONOMY_EXECUTION_SSH
  title: "SSH ExecutionProvider implementation"
  status: todo
  depends_on: [TAXONOMY_EXECUTION_LOCAL]

- id: TAXONOMY_DISCOVERY_BASIC
  title: "Shell-based DiscoveryProvider"
  status: todo
  depends_on: [TAXONOMY_EXECUTION_LOCAL]

# ... etc
```

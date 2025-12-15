# Resources Are The Wrong Abstraction

**Or: Why Stagecraft needs a Vocabulary of Capabilities and Transitions**

## The "Resource" Trap

In the beginning, we treated everything like a Cloud Provider. We had `Host` resources, `Network` resources, `Droplet` resources. The verbs were `Create`, `Read`, `Update`, `Delete`. This CRUD-centric view—the "Resource" abstraction—works beautifully for static infrastructure APIs like AWS or DigitalOcean. They own the state; we just reconcile it.

But Stagecraft is not Terraform. Stagecraft is an **agentic orchestrator** that spans local dev environments, ephemeral CI containers, and long-lived production hosts. It doesn't just "configure resources"; it **mutates reality** across vastly different substrates.

When we try to shoehorn "Ensure Docker is Installed" or "Run a DB Migration" into a `Resource` model, we fail. These are not "resources" with stable IDs and steady states. They are **Transitions**: dynamic operational arcs that take a system from State A to State B.

## The Determinism Gap

Resources imply that state is static and discoverable. But in a local-first world, state is messy.
- A `Host` might be a local Mac machine, a Docker container, or a remote VM.
- A "Resource" check for `Tailscale` might fail because the network is flaky, not because the configuration is wrong.

Our current providers (`CloudProvider`, `NetworkProvider`) mix three distinct concerns:
1.  **Discovery:** "What is true right now?" (e.g., `Hosts()`)
2.  **Planning:** "What should be true?" (e.g., `Plan()`)
3.  **Execution:** "Make it true." (e.g., `Apply()`, `EnsureInstalled()`)

This mixing kills determinism. `Plan()` often implicitly calls `Hosts()`. If `Hosts()` fails (flake), planning fails. If `Apply()` has side effects, they are buried in the provider.

## A New Taxonomy

To build a robust, agentic engine, we must explode the "Resource" and separate concerns into a new taxonomy:

### 1. Capabilities (The "What")
A **Capability** is a pure declaration of a feature or property. It has no side effects.
- *Examples:* `capability.tailscale.v1`, `capability.docker.v1`, `capability.postgres.v14`
- It defines: "I need this system to have the *Capability* to route traffic via Tailscale."

### 2. Transitions (The "How")
A **Transition** is the bridge between *lacking* a Capability and *having* it.
- *Shape:* `Requires: [CapA], Ensures: [CapB], FailureModes: [...]`
- It is a directed edge in our planning graph.

### 3. Execution (The "Doer")
Actual mutation is delegated to a dumb, isolated **Execution Provider**.
- It knows nothing about "Tailscale" or "Postgres".
- It knows only: `Execute(Command)`, `Copy(File)`, `HTTP(Request)`.
- It returns structured, deterministic results.

### 4. Discovery (The "Observer")
Facts determine which Capabilities are currently held. **Discovery Providers** are read-only observers.
- They emit **Facts**: `Fact{Subject: "host-1", Predicate: "has_binary", Object: "docker"}`.
- Planning is a pure function: `f(DesiredCapabilities, ObservedFacts) -> Plan`.

## Use Case: The "Tailscale" Example

**Old Way (Resource):**
- `NetworkProvider.EnsureInstalled()` checks if binary exists (Discovery) and installs it (Execution).
- If installation fails, the provider returns a generic error.
- We don't know if it failed because of network, disk space, or bad config.

**New Way (Taxonomy):**
1.  **Discovery:** `DiscoveryProvider` runs `which tailscale` via `ExecutionProvider`. Emits `Fact(TailscaleInstalled=False)`.
2.  **Planning:** Engine sees `Desired(Tailscale)` mismatch. Finds `Transition(InstallTailscale)`.
    - `Transition` declares: `Requires: [InternetAccess, Sudo]`.
3.  **Execution:** Engine instructs `ExecutionProvider` to run the installation script.
4.  **Integration:** Engine records the result.

## Conclusion

We are moving Stagecraft from a **Resource Manager** to a **Capability Synthesizer**.
- **Resources** manage nouns.
- **Capabilities** manage potential.

This shift allows us to verify plans without networks, test logic without hosts, and execute with surgical precision. It is the only way to scale the agentic future.

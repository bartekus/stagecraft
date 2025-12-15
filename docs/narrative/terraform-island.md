# Terraform Island

## Terraform is not the problem

Terraform is excellent at provisioning infrastructure.
Stagecraft is not trying to replace it.

Terraform answers:
- What infrastructure should exist?
- How do I reconcile cloud APIs with desired state?

Stagecraft answers a different question:
- How do I *use* infrastructure once it exists?

These are complementary concerns.

## The Resource Trap (briefly)

Terraform operates on **resources**.
Stagecraft operates on **capabilities and transitions**.

Trying to ingest Terraform’s resource graph into Stagecraft would reintroduce the very problems we are explicitly fixing:
- Planning that depends on live API calls
- Execution mixed into reasoning
- Flaky, non-reproducible plans

So Stagecraft does not interpret Terraform resources.

## The Terraform Island model

Stagecraft treats Terraform as an **island**:

- Terraform runs deterministically in an execution sandbox
- Terraform provisions infrastructure and manages its own state
- Stagecraft reads only **Terraform outputs**
- Those outputs are normalized into:
  - **HostRefs** (engine-owned host inventory)
  - **Facts** (immutable observations)

Terraform owns provisioning.
Stagecraft owns orchestration.

## The one contract

There is exactly one required contract between Terraform and Stagecraft:

```text
terraform output -json
└── output "stagecraft_hosts"
```

```hcl
Below is a clean, copy-pasteable diagram you can use in docs. It’s deliberately conceptual, not implementation-noisy, and matches your taxonomy (Discovery, Planning, Execution separation).

You can drop this into:
	•	docs/narrative/terraform-island.md
	•	or docs/architecture/ as a reference diagram

⸻

Terraform Island → Capabilities → Execution (Conceptual Flow)

┌────────────────────────────────────────────────────────────┐
│                    Terraform Island                         │
│                                                            │
│   Terraform Repo                                            │
│   (HCL, Modules, State, Providers)                          │
│                                                            │
│   ┌────────────────────────────────────────────────────┐   │
│   │ ExecutionProvider                                   │   │
│   │ (local / container / agent)                         │   │
│   │                                                    │   │
│   │   terraform init                                   │   │
│   │   terraform workspace select|new                   │   │
│   │   terraform apply                                  │   │
│   │   terraform output -json                           │   │
│   │                                                    │   │
│   └────────────────────────────────────────────────────┘   │
│                                                            │
│                Normalized Outputs                           │
│         (Stagecraft-owned JSON contract)                    │
│                                                            │
│   stagecraft_hosts[]                                        │
│   └─ host_id, ips, roles, labels                            │
│                                                            │
└───────────────┬────────────────────────────────────────────┘
                │
                │ emits
                ▼
┌────────────────────────────────────────────────────────────┐
│                 Stagecraft Engine                            │
│                                                            │
│   HostRefs (Inventory)                                      │
│   ───────────────────                                      │
│   prod-web-1, prod-worker-1, …                              │
│                                                            │
│   Facts (Observed State)                                    │
│   ─────────────────────                                    │
│   fact.host.ip.public                                       │
│   fact.host.roles                                           │
│   fact.terraform.workspace                                  │
│                                                            │
└───────────────┬────────────────────────────────────────────┘
                │
                │ discovery (read-only)
                ▼
┌────────────────────────────────────────────────────────────┐
│              Discovery Providers                             │
│                                                            │
│   SSH / Agent / Local Execution                             │
│                                                            │
│   which docker        → fact.docker.installed               │
│   which tailscale     → fact.tailscale.installed            │
│   tailscale status    → fact.tailscale.joined               │
│                                                            │
│   (No mutations allowed here)                               │
│                                                            │
└───────────────┬────────────────────────────────────────────┘
                │
                │ pure function
                ▼
┌────────────────────────────────────────────────────────────┐
│              Capability Providers (Planning)                 │
│                                                            │
│   Inputs:                                                   │
│     - Desired Capabilities                                  │
│     - Observed Facts                                        │
│                                                            │
│   Output:                                                   │
│     Transition Plan                                         │
│                                                            │
│   Example:                                                  │
│     IF docker.installed == false                            │
│        → install-docker                                     │
│     IF tailscale.joined == false                            │
│        → join-tailscale                                     │
│                                                            │
│   (No execution, no probing)                                │
│                                                            │
└───────────────┬────────────────────────────────────────────┘
                │
                │ orchestrated execution
                ▼
┌────────────────────────────────────────────────────────────┐
│              Execution Providers (The Hands)                 │
│                                                            │
│   Local Exec / SSH / Agent                                  │
│                                                            │
│   install-docker                                            │
│   install-tailscale                                         │
│   tailscale up --auth-key=…                                 │
│                                                            │
│   Structured results → new Facts                            │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

⸻

How to read this diagram (key points)

1. Terraform is execution-only

Terraform never participates in planning.
Stagecraft never interprets Terraform’s resource graph.

The boundary is terraform output -json.

⸻

2. Facts are the currency

Everything past Terraform is driven by Facts:
	•	Terraform emits facts about existence and reachability
	•	Discovery emits facts about actual host state
	•	Planning consumes facts, never APIs or shells

This is what fixes determinism.

⸻

3. Capabilities are intent, not infrastructure

Capabilities answer:

“What should this host be capable of?”

Not:

“What resource should exist?”

That’s why Docker install, Tailscale join, agent enrollment all fit cleanly.

⸻

4. Execution is swappable

The bottom box is intentionally generic:
	•	Today: SSH
	•	Tomorrow: systemd agent
	•	In CI: container runner

Same plan, same facts, different hands.

⸻

Optional caption you can add under the diagram

Terraform provisions. Stagecraft reasons. Execution mutates.
Each phase has a single responsibility, and no phase is allowed to cheat.

If you want next, I can:
	•	convert this into a Mermaid diagram for rendered docs,
	•	align it visually with your existing architecture diagrams,
	•	or annotate it with feature IDs so it directly maps to your implementation roadmap.


```mermaid

flowchart TD
  %% ------------------------------------------------------------
  %% Terraform Island → Capabilities → Execution
  %% Aligned to Stagecraft "phases" framing: Inputs → Facts → Plan → Execute
  %% Feature mapping is annotated in node labels as [FEATURE_ID]
  %% ------------------------------------------------------------

  %% === TERRAFORM ISLAND ===
  subgraph TF["Terraform Island"]
    direction TB

    TRepo["Terraform Repo\n(HCL, Modules, State, Providers)"]
    TRunner["Deterministic Terraform Runner\n(pinned version, sandbox, structured logs)\n[PROVIDER_INTEGRATION_TERRAFORM_DETERMINISTIC_RUNNER]"]
    TOut["terraform output -json\nNormalized Outputs Contract\n(stagecraft_hosts)\n[PROVIDER_INTEGRATION_TERRAFORM_OUTPUTS_CONTRACT]"]

    TRepo --> TRunner --> TOut
  end

  %% === ENGINE INGEST ===
  subgraph ENG["Stagecraft Engine"]
    direction TB

    HostRefs["HostRefs (Inventory)\n(host_id, address, roles, labels)\n[PROVIDER_INTEGRATION_TERRAFORM_FACTS_HOSTREFS]"]
    Facts0["Facts (from Terraform)\n(fact.host.ip.*, roles, labels,\nfact.terraform.*)\n[PROVIDER_INTEGRATION_TERRAFORM_FACTS_HOSTREFS]"]

    TOut --> HostRefs
    TOut --> Facts0
  end

  %% === DISCOVERY ===
  subgraph DISC["Discovery Providers (read-only)"]
    direction TB

    ExecSub["ExecutionProvider Substrate\n(local / container / SSH / agent)\n[TAXONOMY_EXECUTION_LOCAL]\n(+ later: TAXONOMY_EXECUTION_SSH)"]
    Probes["Discovery Probes\nwhich docker → fact.docker.installed\nwhich tailscale → fact.tailscale.installed\ntailscale status → fact.tailscale.joined\n[TAXONOMY_DISCOVERY_BASIC]"]

    ExecSub --> Probes
  end

  %% === CAPABILITY PLANNING ===
  subgraph PLAN["Capability Providers (pure planning)"]
    direction TB

    Types["Core Taxonomy Types\n(Fact, Capability, Transition)\n[TAXONOMY_CORE_TYPES]"]
    Desired["Desired Capabilities\n(docker.installed, tailscale.mesh, …)\n[TAXONOMY_MIGRATION_BOOTSTRAP]\n(+ pilot: TAXONOMY_MIGRATION_TAILSCALE)"]
    Planner["Plan(desired, observed_facts) → TransitionPlan\n(pure; no probing)\n[TAXONOMY_MIGRATION_TAILSCALE]"]

    Types --> Planner
    Desired --> Planner
  end

  %% === EXECUTION ===
  subgraph EXEC["Execution (mutations)"]
    direction TB

    Transitions["Transition Plan\ninstall-docker\ninstall-tailscale\njoin-tailscale\n[TAXONOMY_MIGRATION_TAILSCALE]"]
    Apply["Execute via ExecutionProvider\n(local / SSH / agent)\n[TAXONOMY_EXECUTION_LOCAL]\n(+ later: TAXONOMY_EXECUTION_SSH)"]
    Facts1["New Facts (post-exec)\n(fact.docker.installed=true,\nfact.tailscale.joined=true)\n[TAXONOMY_CORE_TYPES]"]

    Transitions --> Apply --> Facts1
  end

  %% === MAIN FLOW ===
  Facts0 --> Probes
  Probes --> Planner
  Planner --> Transitions

  %% === STYLES (visually aligned: clean boxes + neutral palette) ===
  classDef island fill:#f7f7f7,stroke:#333,stroke-width:1px,color:#111;
  classDef engine fill:#f7f7f7,stroke:#333,stroke-width:1px,color:#111;
  classDef phase fill:#ffffff,stroke:#333,stroke-width:1px,color:#111;
  classDef emphasis fill:#ffffff,stroke:#111,stroke-width:2px,color:#111;

  class TF island;
  class ENG engine;
  class DISC,PLAN,EXEC phase;

  class TRunner,TOut,HostRefs,Facts0,ExecSub,Types,Planner,Apply emphasis;
```
Notes on alignment and roadmap mapping:
	•	Boxes and grouping mirror your architecture narrative: External island → Engine ingest → Discovery → Pure planning → Execution.
	•	Feature IDs are placed directly on the nodes they implement so this diagram can be used as a roadmap index.
	•	The diagram assumes you’ll add the taxonomy feature IDs you already referenced in your refactor blueprint:
	•	TAXONOMY_CORE_TYPES
	•	TAXONOMY_EXECUTION_LOCAL
	•	TAXONOMY_DISCOVERY_BASIC
	•	TAXONOMY_MIGRATION_TAILSCALE
	•	TAXONOMY_MIGRATION_BOOTSTRAP
	•	and (later) TAXONOMY_EXECUTION_SSH

If you want this to render exactly like your other diagrams, tell me what Mermaid theme settings you’re using (or paste one existing Mermaid block from your docs) and I’ll match the style precisely.
```

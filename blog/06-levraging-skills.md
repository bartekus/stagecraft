Levraging Skills 

Instructions shape behavior.
Skills produce outcomes.
Confusing them is the root of most “AI agent” chaos.

⸻

1. What a “Skill” actually is (in the AI-agent sense)

A Skill is a bounded, executable capability that an agent can invoke to achieve an outcome.

Think of a skill as:

“Given this input and state, I can reliably perform this transformation or action.”

Core properties of a Skill:
	•	Operational – it does something, not just “thinks”
	•	Callable – can be selected or invoked by an agent
	•	Scoped – has a narrow, explicit responsibility
	•	Composable – can be chained with other skills
	•	Testable – you can assert its behavior
	•	Portable – usable across agents or contexts

Examples:
	•	SummarizeDocument
	•	GenerateTerraformOutputs
	•	ValidateSpecFrontmatter
	•	JoinTailscaleNetwork
	•	ComputeFeatureDependencyGraph

In Stagecraft terms, a Skill is closer to:
	•	a capability than a resource
	•	a plan action than a document
	•	a function than a policy

⸻

2. What *.instruction.md, AGENT.md, Copilot prompts actually are

These are instructional control planes, not capabilities.

They answer:

“How should the agent behave, reason, and constrain itself?”

AGENT.md / instruction files

They are:
	•	Normative – define rules, preferences, style, governance
	•	Global or semi-global – affect all actions
	•	Non-executable – they don’t “run”
	•	Contextual – interpreted, not invoked
	•	Soft-enforced – compliance depends on model behavior

Examples from your ecosystem:
	•	“Spec-first, deterministic outputs”
	•	“One feature per file”
	•	“Always annotate Feature IDs”
	•	“No timestamps or random ordering”

These are constitution-level documents.

⸻

3. The fundamental difference (the key insight)

Here’s the clean distinction:

Dimension	Skill	Instruction / AGENT.md
Ontology	Capability	Constraint / Policy
Invoked	Yes	No
Executable	Yes (conceptually or literally)	No
Scope	Narrow	Broad
Determinism	Enforced by design	Aspirational
Testability	High	Low
Composability	High	Low
Drift risk	Low	High

A skill is “what the agent can do.”
Instructions are “how the agent should behave while doing anything.”

They operate at different layers and should never be conflated.

⸻

4. Why Copilot-style prompts feel similar (but aren’t)

Copilot prompt files look skill-like because:
	•	They’re task-specific
	•	They’re sometimes named after actions
	•	They’re colocated with code

But they are still:
	•	Passive
	•	Non-addressable
	•	Non-composable
	•	Non-verifiable

You cannot say:

“Invoke the CopilotTerraformPlan prompt.”

You can say:

“Invoke the TerraformPlanSkill with inputs X.”

That is a categorical difference.

⸻

5. Where Skills sit in a clean architecture (your mental model)

In your world (Stagecraft / AATSE), the layering looks like this:

┌────────────────────────────┐
│ Governance / Constitution  │  ← AGENT.md, CONTRIBUTING.md
├────────────────────────────┤
│ Reasoning & Planning       │  ← Plans, outlines, analysis
├────────────────────────────┤
│ Skills / Capabilities      │  ← Executable units
├────────────────────────────┤
│ Tools / Runtimes           │  ← Go, Terraform, Docker, APIs
└────────────────────────────┘

Skills are the hinge layer between cognition and execution.

⸻

6. Why this matters (and why you’re right to ask)

If you treat instructions as skills:
	•	You get non-deterministic behavior
	•	You cannot test or evolve capabilities safely
	•	You bake policy into execution
	•	You get agent drift

If you treat skills as instructions:
	•	You lose composability
	•	You lose reusability
	•	You lose governance clarity

Your instinct toward:
	•	capability taxonomy
	•	feature IDs
	•	plan actions
	•	execution substrates

is exactly the move from instruction-driven AI to capability-driven systems.

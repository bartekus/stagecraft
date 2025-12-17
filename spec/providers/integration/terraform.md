---
id: PROVIDER_INTEGRATION_TERRAFORM
title: Terraform Integration Provider (Terraform Island)
status: draft
type: spec
group: providers
requires:
- TAXONOMY_CORE_TYPES
- TAXONOMY_EXECUTION_LOCAL
---

# Terraform Integration Provider (Terraform Island)

## 1. Purpose

Enable Stagecraft to integrate with existing customer Terraform repositories and workflows without requiring a rewrite.

This provider:
- Runs Terraform deterministically via an engine execution substrate (ExecutionProvider).
- Extracts Terraform outputs (`terraform output -json`) and normalizes them into Stagecraft-owned data:
  - `HostRef` inventory for subsequent bootstrap/deploy steps.
  - `Facts` for planning and verification.

Non-goal:
- Stagecraft does not parse HCL for correctness or interpret Terraform resource graphs as Stagecraft plans.

## 2. Definitions

### 2.1 Terraform Root
A directory containing Terraform configuration that can be initialized and executed with Terraform CLI.

### 2.2 Terraform Island
Terraform is treated as an opaque, external system that:
- Receives inputs (vars/workspace/backend env).
- Produces outputs (JSON).
  Stagecraft consumes outputs as facts and host references only.

### 2.3 HostRef
An engine-owned host identity record used across bootstrap, network join, deploy, and later agent enrollment.

### 2.4 Fact
An immutable, observed statement about a subject (host, project, environment). Facts are deterministically serializable.

## 3. Provider Class and Responsibilities

Provider taxonomy class: Integration Provider (execution + normalization).

Responsibilities:
1. Deterministic terraform execution:
  - init/plan/apply/destroy/output with pinned terraform version.
  - structured capture of stdout/stderr/exit for audit.
2. Output normalization:
  - Convert `terraform output -json` to Stagecraft `TerraformOutputs` model.
  - Convert normalized outputs into `HostRef[]` and `Fact[]`.
3. Optional inspection for UX:
  - Detect terraform root(s), required_version, backend type, provider constraints.
  - Must not be required for correctness.

## 4. Determinism Requirements

The provider MUST:
- Run terraform with a pinned version (by config).
- Capture and store:
  - terraform version
  - working directory
  - workspace
  - vars (redacted) and var-file names
  - init/plan/apply outputs (stdout/stderr), exit codes
- Produce deterministic JSON outputs:
  - Sort object keys and arrays where order is not semantically meaningful.
  - No timestamps in normalized artifacts unless explicitly requested by user and gated behind a flag.
- Never rely on nondeterministic filesystem ordering.

The provider MUST NOT:
- Read remote state during the Plan phase of Stagecraft without an explicit execution step.
- Perform HCL evaluation beyond light inspection for UX warnings.

## 5. Configuration

### 5.1 Terraform Provider Config

```yaml
terraform:
  enabled: true
  root: "./infra/terraform"
  terraform_version: "1.6.6"
  workspace: "stagecraft-${env}"
  var_files:
    - "./infra/terraform/env/${env}.tfvars"
  vars:
    env: "prod"
    project: "acme"
  env:
    TF_LOG: "WARN"
  backend:
    mode: "repo" # "repo" | "override"
    override: {} # backend config map when mode="override"
  outputs_contract:
    mode: "strict" # "strict" | "best-effort"
    required:
      - "stagecraft_hosts"
```
Notes:
	•	vars are passed via -var (or an auto tfvars json file in the runner sandbox).
	•	var_files are passed via -var-file.
	•	backend.mode="repo" uses whatever backend config exists in the repository.
	•	backend.mode="override" injects backend config via init flags (implementation-specific).

6. Execution Model

Terraform runs are executed as explicit engine steps using ExecutionProvider:
	•	LocalExecutionProvider initially.
	•	Container runner / remote agent later (same interface).

6.1 Commands

Required commands:
	•	terraform init
	•	terraform workspace select or terraform workspace new
	•	terraform plan -out=<planfile> (optional but recommended)
	•	terraform apply <planfile> (or terraform apply -auto-approve)
	•	terraform output -json

Optional commands:
	•	terraform destroy (when tearing down)
	•	terraform validate (pre-flight)

6.2 Runner Sandbox

The provider SHOULD run in an isolated workspace directory:
	•	Copy or mount terraform root into sandbox.
	•	Place ephemeral files (plans, auto tfvars json) in sandbox.
	•	Ensure secrets never end up in persisted artifacts.

7. Output Normalization Contract

Stagecraft consumes Terraform outputs via:
	•	terraform output -json

The provider normalizes outputs into a Stagecraft-owned JSON structure.

7.1 Normalized Output Schema
```
{
  "hosts": [
    {
      "host_id": "prod-web-001",
      "public_ip": "203.0.113.10",
      "private_ip": "10.0.1.10",
      "ssh_port": 22,
      "roles": ["web"],
      "labels": { "region": "nyc3", "provider": "digitalocean" }
    }
  ],
  "network": {
    "vpc_cidr": "10.0.0.0/16"
  },
  "artifacts": {
    "kubeconfig": null
  }
}
```
Semantics:
	•	hosts[].host_id MUST be stable across runs for the same environment.
	•	roles[] order is not semantically meaningful; MUST be sorted.
	•	labels keys MUST be sorted.

7.2 Terraform Output Mapping

This spec defines one canonical output name for host inventory:
	•	stagecraft_hosts: an output whose value is an array of host objects compatible with hosts[] above.

If outputs_contract.mode=strict, missing stagecraft_hosts is an error.
If best-effort, the provider will attempt to infer hosts from common outputs (non-normative).

8. Facts Emission

The provider MUST emit facts derived from normalized outputs.

Minimum fact set per host:
	•	fact.host.id (subject: host_id)
	•	fact.host.ip.public (if present)
	•	fact.host.ip.private (if present)
	•	fact.host.ssh.port (if present)
	•	fact.host.roles (if present)
	•	fact.host.labels.<k> for each label key

Additional global facts:
	•	fact.terraform.workspace
	•	fact.terraform.root
	•	fact.terraform.version

Facts MUST be deterministically serialized.

9. HostRef Emission

The provider MUST emit HostRef[] derived from normalized outputs.

Minimum HostRef fields:
	•	id = host_id
	•	address = public_ip (preferred) else private_ip (if reachable in current execution substrate)
	•	ssh_port default 22 if missing
	•	roles array
	•	labels map

HostRefs are engine-owned inventory used by bootstrap/network/deploy.

10. Error Handling

Terraform CLI failures must be classified into stable error categories.

Minimum categories:
	•	terraform_init_failed
	•	terraform_plan_failed
	•	terraform_apply_failed
	•	terraform_output_failed
	•	outputs_contract_violation
	•	workspace_failed
	•	terraform_binary_missing
	•	terraform_version_mismatch

Provider must return structured error including:
	•	category
	•	exit code
	•	redacted stderr snippet
	•	action hint (non-normative string)

11. Security
	•	Provider must redact sensitive values from logs and persisted artifacts.
	•	Terraform state handling:
	•	This provider does not manage state; it follows the repo/backend configuration.
	•	If local state exists in sandbox, it must not be persisted by Stagecraft unless explicitly requested.

12. Acceptance Criteria
	1.	Given a valid terraform root with output stagecraft_hosts, Stagecraft can:
	•	init, select workspace, apply
	•	read outputs
	•	produce normalized JSON
	•	produce HostRefs and Facts deterministically.
	2.	If stagecraft_hosts missing and contract is strict, provider errors with outputs_contract_violation.
	3.	Output normalization is deterministic:
	•	stable ordering
	•	no timestamps
	•	identical normalized output for identical terraform output JSON.

13. Implementation Notes (Non-normative)
	•	Prefer a small normalizer package:
	•	parse terraform output -json
	•	map output name stagecraft_hosts to TerraformOutputs.Hosts
	•	sort roles and label keys
	•	Execute via ExecutionProvider so this integrates with container runner and agent execution later.

---

If you want, next I can generate the **matching feature entries** for `spec/features.yaml` (with dependencies on your taxonomy slices) and a small **“customer quickstart”** snippet you can drop into docs that shows the exact Terraform output they need to add (`output "stagecraft_hosts" { value = ... }`).

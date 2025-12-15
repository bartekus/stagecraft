Example: Simple Host

```hcl
# Stagecraft Terraform Island: required output
#
# Stagecraft reads: `terraform output -json`
# and expects an output named `stagecraft_hosts` whose value is a list of objects.

output "stagecraft_hosts" {
  description = "Host inventory for Stagecraft (normalized into HostRefs + Facts)"
  value = [
    {
      host_id    = "prod-web-001"          # MUST be stable across runs for the same environment
      public_ip  = digitalocean_droplet.web.ipv4_address
      private_ip = digitalocean_droplet.web.ipv4_address_private
      ssh_port   = 22
      roles      = ["web"]
      labels = {
        provider = "digitalocean"
        region   = digitalocean_droplet.web.region
        env      = var.env
      }
    }
  ]
}
```

⸻

Example: Multiple Hosts with Roles (Web + Worker)
```hcl
########################################
# Variables
########################################

variable "env" {
  type    = string
  default = "prod"
}

variable "region" {
  type    = string
  default = "nyc3"
}

variable "web_count" {
  type    = number
  default = 2
}

variable "worker_count" {
  type    = number
  default = 3
}

########################################
# Resources
########################################

resource "digitalocean_droplet" "web" {
  count  = var.web_count
  name   = "${var.env}-web-${count.index + 1}"
  region = var.region
  size   = "s-2vcpu-2gb"
  image  = "ubuntu-22-04-x64"
}

resource "digitalocean_droplet" "worker" {
  count  = var.worker_count
  name   = "${var.env}-worker-${count.index + 1}"
  region = var.region
  size   = "s-2vcpu-4gb"
  image  = "ubuntu-22-04-x64"
}

########################################
# Stagecraft output
########################################

output "stagecraft_hosts" {
  description = "Host inventory for Stagecraft (web + worker roles)"
  value = concat(
    [
      for idx, d in digitalocean_droplet.web : {
      host_id    = "${var.env}-web-${idx + 1}"
      public_ip  = d.ipv4_address
      private_ip = d.ipv4_address_private
      ssh_port   = 22
      roles      = ["web"]
      labels = {
        env      = var.env
        role     = "web"
        region   = var.region
        provider = "digitalocean"
      }
    }
    ],
    [
      for idx, d in digitalocean_droplet.worker : {
      host_id    = "${var.env}-worker-${idx + 1}"
      public_ip  = d.ipv4_address
      private_ip = d.ipv4_address_private
      ssh_port   = 22
      roles      = ["worker"]
      labels = {
        env      = var.env
        role     = "worker"
        region   = var.region
        provider = "digitalocean"
      }
    }
    ]
  )
}
```
⸻

Why this pattern works well with Stagecraft

A few points worth calling out explicitly in the docs (this is the “aha” for customers):
1.	Stable host identity
   •	host_id is derived from intent (prod-web-1), not Terraform resource addresses.
   •	This lets Stagecraft track hosts across bootstrap, deploy, network join, and future agent enrollment.
2.	Role-driven orchestration
   •	roles = ["web"] / ["worker"] lets Stagecraft target capabilities and deployments by role.
   •	No coupling to Terraform resource names.
3.	Terraform stays Terraform
   •	Customers keep modules, state, providers, and CI exactly as-is.
   •	Stagecraft only consumes terraform output -json.
4.	Deterministic normalization
   •	Arrays (roles) and map keys (labels) are naturally sortable.
   •	Re-running Terraform with no changes yields identical normalized host inventory.

⸻

1) Module-based variant (common repo shape)

This pattern assumes a modules/droplets module that returns a list of droplet objects.
Stagecraft wants a single output named stagecraft_hosts.

modules/droplets/outputs.tf:
```hcl
output "droplets" {
  description = "Droplet inventory emitted by the module"
  value = [
    for d in digitalocean_droplet.this : {
      id         = d.id
      name       = d.name
      region     = d.region
      public_ip  = d.ipv4_address
      private_ip = d.ipv4_address_private
    }
  ]
}
```
Root module main.tf:
```hcl
variable "env"    { type = string }
variable "region" { type = string }

module "web" {
  source = "./modules/droplets"

  env    = var.env
  region = var.region
  role   = "web"
  count  = 2
  size   = "s-2vcpu-2gb"
  image  = "ubuntu-22-04-x64"
}

module "worker" {
  source = "./modules/droplets"

  env    = var.env
  region = var.region
  role   = "worker"
  count  = 3
  size   = "s-2vcpu-4gb"
  image  = "ubuntu-22-04-x64"
}
```
Root module outputs.tf (Stagecraft contract):
```hcl
output "stagecraft_hosts" {
  description = "Host inventory for Stagecraft (normalized into HostRefs + Facts)"
  value = concat(
    [
      for idx, d in module.web.droplets : {
        host_id    = "${var.env}-web-${idx + 1}"    # stable identity
        public_ip  = d.public_ip
        private_ip = d.private_ip
        ssh_port   = 22
        roles      = ["web"]
        labels = {
          env      = var.env
          role     = "web"
          region   = d.region
          provider = "digitalocean"
        }
      }
    ],
    [
      for idx, d in module.worker.droplets : {
        host_id    = "${var.env}-worker-${idx + 1}"
        public_ip  = d.public_ip
        private_ip = d.private_ip
        ssh_port   = 22
        roles      = ["worker"]
        labels = {
          env      = var.env
          role     = "worker"
          region   = d.region
          provider = "digitalocean"
        }
      }
    ]
  )
}
```
This is the sweet spot:
	•	Customers keep module boundaries and don’t rewire their infra.
	•	Stagecraft gets stable host_id, roles, IPs, and labels.

⸻

2) End-to-end mapping: Terraform → HostRefs/Facts → Bootstrap → Tailscale join

This is the conceptual pipeline that aligns with your taxonomy direction.

Step A: Terraform execution (Integration Provider)

Stagecraft runs Terraform (deterministically) and reads outputs.
	•	Execute:
	•	terraform init
	•	terraform workspace select|new <workspace>
	•	terraform apply ...
	•	terraform output -json
	•	Normalize:
	•	Parse stagecraft_hosts
	•	Produce:
	•	HostRefs[]
	•	Facts[]

HostRef example produced:
```
{
  "id": "prod-web-1",
  "address": "203.0.113.10",
  "ssh_port": 22,
  "roles": ["web"],
  "labels": { "env":"prod", "role":"web", "region":"nyc3", "provider":"digitalocean" }
}
```
Fact examples produced (per host):
	•	fact.host.id = prod-web-1
	•	fact.host.ip.public = 203.0.113.10
	•	fact.host.roles = ["web"]
	•	fact.host.labels.role = web
	•	plus global facts:
	•	fact.terraform.workspace
	•	fact.terraform.version

Step B: Bootstrap (Composite Capability)

Bootstrap becomes “desired capabilities per host”, not procedural logic.

For example, for all hosts:
	•	capability.os.linux (assumed/validated)
	•	capability.docker.installed.v1

Optionally, for all hosts (or only certain roles):
	•	capability.network.mesh.tailscale.v1

So bootstrap in Stagecraft terms is basically:

DesiredState
	•	For role=web|worker: require docker
	•	For everything: require tailscale mesh (if enabled)

Step C: Discovery (read-only)

For each host, engine schedules discovery probes via ExecutionProvider (SSH execution substrate later, local for tests).

Examples:
	•	which docker → emits fact.docker.installed = true|false
	•	docker version → emits version facts
	•	which tailscale → emits fact.tailscale.installed
	•	tailscale status --json (or a safe equivalent) → emits fact.tailscale.joined, tailnet name, node id

Step D: Planning (pure)

CapabilityProvider logic (pure function):
	•	If fact.docker.installed missing → add transition install-docker
	•	If fact.tailscale.installed missing → add transition install-tailscale
	•	If fact.tailscale.joined missing and desired mesh capability present → add transition join-tailscale

Step E: Execution (mutations)

Engine orchestrates transitions, executing actions via ExecutionProvider:
	•	install-docker runs distro-appropriate script
	•	install-tailscale runs official install method
	•	join-tailscale runs tailscale up --auth-key=... (or tagged auth)

This directly fixes your determinism leaks:
	•	Planning never calls DO API or runs shell.
	•	Discovery is explicit and schedulable.
	•	Exec is unified and swappable (SSH/Agent/Container).

⸻

3) Docs page structure (so it lands cleanly)

You want this to be approachable for customers with existing Terraform, while keeping your narrative (capabilities/transitions) intact.

Recommended docs layout

A) Narrative (1 page, short)
	•	docs/narrative/terraform-island.md
	•	Goal: explain the philosophy in 2–3 minutes.
	•	Sections:
	1.	“Stagecraft is not Terraform”
	2.	“Terraform Island: we run it; we don’t interpret it”
	3.	“What Stagecraft adds: bootstrap, mesh, deploy, verification”
	4.	“The one contract: output stagecraft_hosts”

B) Provider docs (practical quickstart)
	•	docs/providers/terraform.md
	•	Goal: copy/paste onboarding.
	•	Sections:
	1.	Requirements (Terraform version pin, credentials)
	2.	Minimal contract output (single host)
	3.	Realistic multi-host example (the one you asked for)
	4.	Module-based variant (the one above)
	5.	Troubleshooting (missing outputs, wrong types, unstable host_id)

C) End-to-end guide (optional but powerful)
	•	docs/guides/terraform-to-mesh-and-deploy.md
	•	Goal: show the full story without going deep into engine internals.
	•	Sections:
	1.	“Provision with Terraform”
	2.	“Stagecraft reads outputs and builds host inventory”
	3.	“Bootstrap installs Docker”
	4.	“Join Tailscale mesh”
	5.	“Deploy Compose app”
	•	Keep it role-based: web/worker.

D) Spec (already have it)
	•	spec/providers/integration/terraform.md
	•	Pure spec; referenced by features.

How to avoid overwhelming people

Rule of thumb:
	•	Narrative page: 10–20 lines per section max.
	•	Provider quickstart: show code first, explain after.
	•	End-to-end guide: present as a single happy path; move edge cases to troubleshooting.



```

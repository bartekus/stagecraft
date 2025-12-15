# Provider Catalog

This catalog lists provider specs under `spec/providers/`, grouped by provider domain.


## backend

| Path | Chunks | Primary Topics |
|------|--------|----------------|
| `spec/providers/backend/encore-ts.md` | 11 | 1. Goals and Non-Goals, 1.1 Goals, 1.2 Non-Goals, 2. Relationship to Core Backend Abstraction, 2.1 BackendProvider Interface |
| `spec/providers/backend/generic.md` | 20 | Build Mode Behavior, Comparison with Other Providers, Config Parsing, Configuration, Dev Mode Behavior |


## ci

| Path | Chunks | Primary Topics |
|------|--------|----------------|
| `spec/providers/ci/interface.md` | 11 | CI Provider Interface, Config Schema, Error Types, Goal, Interface |


## cloud

| Path | Chunks | Primary Topics |
|------|--------|----------------|
| `spec/providers/cloud/digitalocean.md` | 27 | 1. Overview, 10. Cost and Billing Responsibility, 2. Interface Contract, 2.1 ID, 2.2 Plan |
| `spec/providers/cloud/interface.md` | 11 | Cloud Provider Interface, Config Schema, Error Types, Goal, Interface |


## frontend

| Path | Chunks | Primary Topics |
|------|--------|----------------|
| `spec/providers/frontend/generic.md` | 21 | Comparison with Other Providers, Config Parsing, Configuration, Dev Mode Behavior, Error Handling |
| `spec/providers/frontend/interface.md` | 9 | Config Schema, Frontend Provider Interface, Goal, Interface, Non-Goals (v1) |


## migration

| Path | Chunks | Primary Topics |
|------|--------|----------------|
| `spec/providers/migration/raw.md` | 23 | Comparison with Other Engines, Configuration, Core Validation (Stagecraft), Database Support, Engine-Specific Validation |


## network

| Path | Chunks | Primary Topics |
|------|--------|----------------|
| `spec/providers/network/interface.md` | 11 | Config Schema, Error Types, Goal, Interface, Network Provider Interface |
| `spec/providers/network/tailscale.md` | 38 | 1. Overview, 10. Testing, 10.1 Unit Tests, 10.2 Integration Tests (Optional), 11. Non-Goals (v1) |


## secrets

| Path | Chunks | Primary Topics |
|------|--------|----------------|
| `spec/providers/secrets/interface.md` | 11 | Config Schema, Error Types, Goal, Interface, Non-Goals (v1) |

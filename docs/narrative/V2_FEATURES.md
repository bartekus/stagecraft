<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<!--
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
-->

# Stagecraft v2 Features

> **Status**: Future Planning
> **Purpose**: High-level overview of planned v2 features
> **Last Updated**: 2025-01-XX

---

## Overview

This document provides a high-level overview of features planned for Stagecraft v2. These features are **deferred** from v1 and will be considered after v1 is complete and stable.

> **Note**: This is a public-facing overview. Detailed strategic planning, prioritization, and competitive analysis are maintained internally.

---

## v2 Feature Categories

### Environment Management
- **Ephemeral Environments**: Support for temporary, on-demand environments for testing and preview deployments
- **Environment Templates**: Reusable environment configurations

### Observability & Monitoring
- **Health Monitoring**: Proactive health checks and alerting
- **Observability Stack**: Comprehensive monitoring, logging, and tracing integration
- **Topology Visualization**: Visual representation of infrastructure and service relationships

### Developer Experience
- **Editor Integration**: IDE and editor plugin support for enhanced workflows
- **Sync Capabilities**: Local and remote synchronization primitives
- **Pipeline Composition**: Advanced pipeline configuration and composition

### Infrastructure Management
- **Infrastructure Recipes**: Reusable infrastructure templates and patterns
- **Snapshot Management**: Infrastructure snapshot and restore capabilities
- **Cost Management**: Budget tracking and guardrails

### Security & Compliance
- **Secrets Management**: Enhanced secrets orchestration and rotation
- **Audit Capabilities**: Comprehensive build and deployment history tracking
- **Multi-tenant Support**: Organization and team features with RBAC

### Migration & Testing
- **Migration Tooling**: Advanced migration planning and simulation
- **Testing Enhancements**: Enhanced testing capabilities and integration

---

## Implementation Timeline

v2 planning will begin when:
- All v1 features are complete and tested
- v1 has been used in production for at least one project
- User feedback identifies clear v2 priorities
- Core architecture is stable and extensible

Specific features and prioritization will be determined based on:
- User feedback and feature requests
- Technical feasibility and architecture readiness
- Business priorities and market needs

---

## Contributing

If you have ideas for v2 features or want to discuss future directions:
- Open an issue with the `v2` label
- Start a discussion in the discussions forum
- Contribute to the specification in `spec/` when features are approved

---

## Related Documents

- [`docs/implementation-roadmap.md`](./implementation-roadmap.md) - Complete implementation roadmap including v2 feature list
- [`spec/features.yaml`](../spec/features.yaml) - Feature registry (source of truth)
- [`docs/implementation-status.md`](./implementation-status.md) - Current implementation status

---

## Notes

This document provides a **high-level overview** of v2 features. It does not include:
- Detailed prioritization or rankings
- Competitive positioning analysis
- Implementation sequencing details
- Strategic differentiators
- Moat-defining capabilities

For detailed strategic planning, see internal documentation (not publicly available).


<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<!--
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
-->

# Strategic Document Migration Guide

This guide explains how to handle strategic documents that contain competitive intelligence and should not be public.

---

## Quick Summary

✅ **Created**:
- `internal-strategy/` directory (git-ignored, cursor-ignored)
- `docs/V2_FEATURES.md` - Safe public version template
- Updated `docs/implementation-roadmap.md` - Removed detailed v2 references

📝 **You Need To**:
1. Manually create `internal-strategy/04-new-feature-ideas.md` with your strategic content
2. Add the warning header from the template below
3. Verify it's not tracked by git

---

## Step-by-Step Migration

### 1. Create Your Internal Strategic Document

Manually create `internal-strategy/04-new-feature-ideas.md` with this header:

```markdown
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<!--
⚠️ INTERNAL STRATEGIC DOCUMENT - DO NOT PUBLISH ⚠️

This document contains:
- Strategic differentiators and competitive positioning
- Detailed feature prioritization and rankings
- Moat-defining capabilities
- Multi-year product roadmap
- Architecture decisions that reveal competitive advantages

This document should NEVER be:
- Committed to git (it's git-ignored)
- Shared publicly
- Referenced in public documentation
- Included in AI context (it's cursor-ignored)

For public-facing content, see:
- `docs/implementation-roadmap.md` (v2 features section - high-level only)
- `docs/V2_FEATURES.md` (safe public overview)
-->

# New Feature Ideas: Strategic Roadmap

> **Status**: Internal Strategic Planning
> **Classification**: INTERNAL — STRATEGIC — DO NOT PUBLISH
> **Last Updated**: [DATE]

[Your full strategic content here - rankings, prioritization, competitive analysis, moat positioning, etc.]
```

### 2. Verify Git Ignore

Check that `.gitignore` includes:
```
internal-strategy/
```

### 3. Verify It's Not Tracked

```bash
git status
# Should NOT show internal-strategy/ files

# Try to add it (should be ignored):
git add internal-strategy/
git status
# Should show nothing new
```

### 4. Use the Safe Public Version

The safe public version is already created:
- `docs/V2_FEATURES.md` - High-level, vague, no prioritization

Update `docs/implementation-roadmap.md` to reference this instead of the strategic doc.

---

## What's Safe vs Dangerous?

### ✅ Safe to Publish (Public)
- High-level feature names: "Ephemeral Environments"
- Generic descriptions: "Support for temporary environments"
- General categories: "Environment Management"
- Vague timelines: "Planned for v2"

### ❌ Keep Private (Internal Strategy)
- Rankings: "Top 5 most valuable features"
- Prioritization: "This is our #1 differentiator"
- Competitive analysis: "vs Kamal, Coolify, etc."
- Moat positioning: "Hard to copy because..."
- Implementation sequencing: "Do this first, then that"
- Business strategy: "Revenue driver", "SaaS positioning"

---

## Template: Internal Strategic Document

When creating `internal-strategy/04-new-feature-ideas.md`, use this structure:

```markdown
⚠️ INTERNAL STRATEGIC DOCUMENT - DO NOT PUBLISH ⚠️

# New Feature Ideas: Strategic Roadmap

## Top 5 Most Valuable Features

1. [Feature] - [Why it's #1, competitive advantage, moat strength]
2. [Feature] - [Strategic reasoning]
...

## Competitive Positioning

[How these features position Stagecraft vs competitors]

## Implementation Sequencing

[Detailed sequencing with business priorities]

## Architecture Considerations

[Future-proofing decisions]
```

---

## Template: Safe Public Version

The public version (`docs/V2_FEATURES.md`) should look like:

```markdown
# v2 Features

## Environment Management
- **Ephemeral Environments**: Support for temporary, on-demand environments

[No rankings, no prioritization, no competitive analysis]
```

---

## Verification Checklist

- [ ] `internal-strategy/04-new-feature-ideas.md` created manually
- [ ] Warning header added to internal document
- [ ] `.gitignore` includes `internal-strategy/`
- [ ] `.cursorignore` includes `internal-strategy/**`
- [ ] `git status` shows no untracked files in `internal-strategy/`
- [ ] Public version (`docs/V2_FEATURES.md`) is vague and high-level
- [ ] `docs/implementation-roadmap.md` references safe public version
- [ ] All rankings/prioritization removed from public docs

---

## Why This Matters

Strategic documents containing:
- Feature prioritization
- Competitive differentiators  
- Moat-defining capabilities
- Multi-year roadmaps

...are **valuable intelligence** for competitors. They reveal:
- What you're building next
- What you think is most valuable
- How you're positioning against competitors
- Your strategic direction for years ahead

**Never commit these to a public repository.**

---

## Questions?

If unsure whether content should be public or private:
- **When in doubt, keep it private**
- Ask: "Would a competitor benefit from seeing this?"
- Ask: "Does this reveal our strategic direction?"
- Ask: "Does this show our prioritization?"

If the answer to any is "yes", it belongs in `internal-strategy/`.


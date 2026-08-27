# Learnings

Project-specific knowledge that isn't in `SKILL.md` or `references/`. A learning **overrides** the defaults — the agent reads this folder first.

## When to add one

After you hit a convention, constraint, or trap specific to this project that isn't already captured. Prefer the `/learn` command — it greps for overlap first and decides **Save / Improve / Absorb / Drop** so this folder doesn't fill with duplicates.

## Format

One file per learning: `YYYY-MM-DD-short-slug.md`, body under ~30 lines. Frontmatter says when it applies, then the rule and why.

```markdown
---
description: <one line — when this applies>
---

<the rule>. **Why:** <the reason>. **How to apply:** <what to do>.
```

Keep it specific and current — if a learning becomes wrong, delete it.

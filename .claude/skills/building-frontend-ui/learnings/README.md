# Learnings — building-frontend-ui

Project-specific frontend knowledge that isn't in `SKILL.md` or `references/`. A learning **overrides** the defaults — the agent reads this folder first.

## When to add one

After you hit a convention, constraint, or trap specific to this project: the design-token source and naming, the component library in use, the styling approach, the state/data libraries, banned dependencies, the breakpoints, the test setup, or a recurring UI bug class.

## Format

One file per learning: `YYYY-MM-DD-short-slug.md`, body under ~30 lines. Start with frontmatter that says when it applies, then the rule and why.

```markdown
---
description: <one line — when this applies>
---

<the rule>. **Why:** <the reason>. **How to apply:** <what to do>.
```

Keep it specific and current — if a learning becomes wrong, delete it.

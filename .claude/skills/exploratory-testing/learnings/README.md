# Learnings — read before, append after

This folder accumulates project-specific or user-specific lessons that aren't obvious from the main SKILL.md. It is the **memory of the skill**.

The protocol below applies to **every** skill that has a `learnings/` folder.

---

## Before generating: read

List this folder and read every file whose name looks relevant to the current session. Product risk areas, fragile flows, environment quirks, charter-scoping conventions, and "this area is always broken under condition X" rules live here. If a learning contradicts a default in `SKILL.md` or `references/`, the learning wins — mention it to the user.

If `learnings/` is empty except for this README, proceed with the SKILL.md defaults.

---

## After completing: append (only when warranted)

Append a new learning only if you discovered something that:

- Was **not obvious** from `SKILL.md` or `references/`.
- Would have **changed how you scoped the charter or where you looked** if you had known it.
- Is **likely to apply again** to this user / project / domain.
- Is a **convention, constraint, risk area, or trap** — not a one-off fact about the bug you just found.

Do **not** append:

- A summary of the session you just ran or the bug you just filed.
- Restatements of the SKILL.md or references content.
- Ephemeral facts ("today checkout was broken").
- Anything you can't state as a rule applicable to a future session.

---

## Append protocol

**Filename:** `YYYY-MM-DD-short-kebab-slug.md` — use the real date.

**Body:**

```markdown
---
description: <one-line: when does this learning apply?>
---

# <Title — what the rule is, in one short phrase>

**Context:** <one or two sentences — the situation where the lesson surfaced>

**Lesson:** <the rule itself, stated so it can be applied without re-reading the context>

**Why:** <the reason — incident, preference, convention, technical constraint>

**Apply when:** <the trigger condition, more precise than the description>
```

Keep the body under 30 lines.

---

## How learnings interact with the SKILL.md

- `SKILL.md` is the **default behavior**.
- `references/` are the **detailed defaults**, loaded on demand.
- `learnings/` are the **overrides** — narrower, project-specific, accumulated over time.

When a learning conflicts with the SKILL.md, follow the learning **and mention it** to the user so they can correct or refine it. A learning overrides the defaults.

---

## Examples of good learnings for this skill (illustrative — delete when real ones land)

- `2026-02-10-import-flow-is-the-risk-hotspot.md` — "The CSV import flow is chronically fragile (RCRCRC: Chronic). Any session touching imports should start with malformed-file and Unicode data attacks before anything else."
- `2026-03-22-staging-clock-runs-utc-only.md` — "Staging ignores user time zone and renders everything in UTC, so time-zone bugs never reproduce there. Charter time/scheduling sessions against a prod-like build, not staging."
- `2026-04-15-charters-must-name-a-persona.md` — "This team's charters always specify a persona resource (free vs paid vs admin); sessions without one consistently miss role-gating bugs."

Examples of bad learnings (do not write these):

- `2026-05-01-found-discount-bug.md` — describes a specific bug, not a reusable rule.
- `2026-05-12-use-proof-debrief.md` — already in SKILL.md; duplication.
- `2026-05-15-exploring-is-fun.md` — opinion without an observed trigger.

---

## Maintenance

When a learning becomes stale (the flow got rewritten, the hot zone moved), **delete the file** rather than editing it in place. Git history preserves the older version; the folder reflects what is currently true.

When two learnings overlap, merge them and delete the duplicate.

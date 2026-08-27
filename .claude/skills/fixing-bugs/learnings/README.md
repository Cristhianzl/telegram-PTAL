# Learnings — read before, append after

This folder accumulates project-specific or user-specific lessons that aren't obvious from the main SKILL.md. It is the **memory of the skill**.

The protocol below applies to **every** skill that has a `learnings/` folder.

---

## Before generating: read

List this folder and read every file whose name looks relevant to the current bug. Recurring bug patterns, hot zones, framework-specific traps, or "this exception always means X here" rules live here. If a learning contradicts a default in `SKILL.md`, the learning wins — mention it to the user.

If `learnings/` is empty except for this README, proceed with the SKILL.md defaults.

---

## After completing: append (only when warranted)

Append a new learning only if you discovered something that:

- Was **not obvious** from `SKILL.md` or `references/`.
- Would have **changed how you started investigating** if you had known it.
- Is **likely to apply again** to this user / project / domain.
- Is a **convention, constraint, or trap** — not a one-off fact about the current bug.

Do **not** append:

- A summary of the bug you just fixed.
- Restatements of the SKILL.md content.
- Ephemeral facts ("today the bug was in cart.py").
- Anything you can't state as a rule applicable to a future task.

---

## Append protocol

**Filename:** `YYYY-MM-DD-short-kebab-slug.md` — use the real date.

**Body:**

```markdown
---
trigger: <one-line: when does this learning apply?>
---

# <Title — what the rule is, in one short phrase>

**Context:** <one or two sentences — the situation where the lesson surfaced>

**Lesson:** <the rule itself, stated so it can be applied without re-reading the context>

**Why:** <the reason — incident, preference, convention, technical constraint>

**Apply when:** <the trigger condition, more precise than the frontmatter>
```

Keep the body under 30 lines.

---

## How learnings interact with the SKILL.md

- `SKILL.md` is the **default behavior**.
- `references/` are the **detailed defaults**, loaded on demand.
- `learnings/` are the **overrides** — narrower, project-specific, accumulated over time.

When a learning conflicts with the SKILL.md, follow the learning **and mention it** to the user so they can correct or refine it.

---

## Examples of good learnings for this skill (illustrative — delete when real ones land)

- `2026-02-12-streaming-uses-generator-exit-not-cancellederror.md` — "On the `/conversations/stream` endpoint, client disconnect surfaces as `GeneratorExit`, not `CancelledError`. Any bug report about 'stream crash' must test `aclose()`, not `athrow(CancelledError)`."
- `2026-03-04-billing-bugs-need-postgres-container.md` — "Bugs in `services/billing/` cannot be reproduced against SQLite. Use the test Postgres container — the bug is almost always in a JSONB query or a transaction boundary that SQLite hides."
- `2026-04-21-cluster-timezone-off-by-one.md` — "Off-by-one timezone bugs in this repo are a cluster: see TECH-450. Default policy is UTC at storage, local only at presentation — see `developing-features/references/observability.md` for logging."

Examples of bad learnings (do not write these):

- `2026-05-01-fixed-discount-bug.md` — describes a specific bug, not a reusable rule.
- `2026-05-12-use-tdd-for-bugs.md` — already in SKILL.md; duplication.
- `2026-05-15-bugs-are-frustrating.md` — opinion without an observed trigger.

---

## Maintenance

When a learning becomes stale (the framework changed, the hot zone got rewritten), **delete the file** rather than editing it in place. Git history preserves the older version; the folder reflects what is currently true.

When two learnings overlap, merge them and delete the duplicate.

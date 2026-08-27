# Learnings — read before, append after

This folder accumulates project-specific or user-specific lessons that aren't obvious from the main SKILL.md. It is the **memory of the skill**.

The protocol below applies to **every** skill that has a `learnings/` folder.

---

## Before generating: read

List this folder and read every file whose name looks relevant to the current review. Severity adjustments for this codebase, recurring violation patterns, banned anti-patterns specific to the project, or "in this repo, X is fine because Y" exceptions live here. If a learning contradicts a default in `SKILL.md`, the learning wins — mention it to the user.

If `learnings/` is empty except for this README, proceed with the SKILL.md defaults.

---

## After completing: append (only when warranted)

Append a new learning only if you discovered something that:

- Was **not obvious** from `SKILL.md` or `references/`.
- Would have **changed how you started reviewing** if you had known it.
- Is **likely to apply again** to this user / project / review area.
- Is a **convention, constraint, or trap** — not a one-off fact about the current PR.

Do **not** append:

- A summary of the review you just produced.
- Restatements of the SKILL.md content.
- Ephemeral facts ("this PR had 14 findings").
- Anything you can't state as a rule applicable to a future review.

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

- `2026-02-12-no-comprehension-block-on-storybook.md` — "Reviews of `apps/storybook/` are exempt from the Comprehension Audit blocker — the area is a demo sandbox. Flag as `R` (Recommended) at most."
- `2026-03-04-billing-service-no-direct-orm.md` — "Any `db.query(` or raw ORM call inside `services/billing/` is a Blocker (`B`). Billing must go through `BillingRepository`. Documented in ADR-0019."
- `2026-04-21-line-budget-relaxed-graphql.md` — "Generated GraphQL resolvers under `src/graphql/__generated__/` are exempt from the 500-line file gate. Don't open findings against generated code."
- `2026-05-08-comment-template-i18n.md` — "When the PR has a `lang=` label, the review summary section must still be in English (per repo policy), but a one-line translation can be added beneath each `Blocker` heading."

Examples of bad learnings (do not write these):

- `2026-05-01-reviewed-a-pr.md` — describes a specific review, not a reusable rule.
- `2026-05-12-use-b-i-r-labels.md` — already in SKILL.md; duplication.
- `2026-05-15-this-pr-was-messy.md` — opinion without an observed trigger.

---

## Maintenance

When a learning becomes stale (the codebase evolved, the convention changed), **delete the file** rather than editing it in place. Git history preserves the older version; the folder reflects what is currently true.

When two learnings overlap, merge them and delete the duplicate.

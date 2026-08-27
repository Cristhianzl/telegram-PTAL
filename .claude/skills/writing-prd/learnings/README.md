# Learnings — read before, append after

This folder accumulates project-specific or user-specific lessons that aren't obvious from the main SKILL.md. It is the **memory of the skill**.

The protocol below applies to **every** skill that has a `learnings/` folder.

---

## Before generating: read

List this folder and read every file whose name looks relevant to the current product area, team, or artifact type. Project-specific section templates, required metrics, approval workflows, non-goal conventions, or PR-FAQ variants live here. If a learning contradicts a default in `SKILL.md`, the learning wins — mention it to the user.

If `learnings/` is empty except for this README, proceed with the SKILL.md defaults.

---

## After completing: append (only when warranted)

Append a new learning only if you discovered something that:

- Was **not obvious** from `SKILL.md` or `references/`.
- Would have **changed how you started writing the PRD** if you had known it.
- Is **likely to apply again** to this user / project / product area.
- Is a **convention, constraint, or trap** — not a one-off fact about the current PRD.

Do **not** append:

- A summary of the PRD you just wrote.
- Restatements of the SKILL.md content.
- Ephemeral facts ("this PRD had 3 open questions").
- Anything you can't state as a rule applicable to a future PRD.

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

**Why:** <the reason — incident, preference, convention, business constraint>

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

- `2026-02-10-prd-template-lives-in-notion.md` — "PRDs for this org use the fixed Notion template, not the skeleton in `references/prd-structure.md`. Map our sections onto theirs: their 'Context' = our problem statement + evidence."
- `2026-03-02-north-star-is-fixed.md` — "The North Star for the growth team is 'weekly activated teams' — every PRD in that area must tie its success metric to it, never invent a new top-line metric."
- `2026-04-15-non-goals-need-rationale.md` — "Reviewers here reject a non-goals list without a one-line *why* per item. Always pair each non-goal with its reason."
- `2026-05-06-strategic-bets-require-prfaq.md` — "Anything labeled a 'bet' in the quarterly plan must ship a PR-FAQ before a PRD; PRDs for bets are rejected without one."

Examples of bad learnings (do not write these):

- `2026-05-01-wrote-checkout-prd.md` — describes a specific PRD, not a reusable rule.
- `2026-05-12-use-rice.md` — already in `references/`; duplication.
- `2026-05-15-prds-feel-long.md` — opinion without an observed trigger.

---

## Maintenance

When a learning becomes stale (the template changed, the North Star moved), **delete the file** rather than editing it in place. Git history preserves the older version; the folder reflects what is currently true.

When two learnings overlap, merge them and delete the duplicate.

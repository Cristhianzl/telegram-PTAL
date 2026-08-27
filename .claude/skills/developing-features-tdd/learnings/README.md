# Learnings — read before, append after

This folder accumulates project-specific or user-specific lessons that aren't obvious from the main SKILL.md. It is the **memory of the skill**.

The protocol below applies to **every** skill that has a `learnings/` folder.

---

## Before generating: read

List this folder and read every file whose name looks relevant to the current feature. If a learning contradicts a default in `SKILL.md`, the learning wins — mention it to the user.

If `learnings/` is empty except for this README, proceed with the SKILL.md defaults.

---

## After completing: append (only when warranted)

Append a new learning only if you discovered something that:

- Was **not obvious** from `SKILL.md` or `references/`.
- Would have **changed how you started** if you had known it.
- Is **likely to apply again** to this user / this project / this domain.
- Is a **convention, constraint, or trap** — not a one-off fact about the current feature.

Do **not** append:

- The summary of what you just built.
- Restatements of the SKILL.md content.
- Ephemeral facts ("today the feature had 4 slices").
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

- `2026-02-12-pytest-anyio-not-asyncio.md` — "This repo uses `pytest-anyio`, not `pytest-asyncio`. Use `@pytest.mark.anyio` and the `anyio_backend` fixture — `@pytest.mark.asyncio` will silently skip."
- `2026-03-04-no-mocking-billing.md` — "Tests touching `services/billing/` use a Postgres test container, never mocks. We had two outages caused by mock/prod divergence in the migration plan."
- `2026-04-21-claude-haiku-fallback.md` — "For the support-bot feature, the fallback when Claude Opus times out is Claude Haiku (not a templated response). Documented in ADR-0014 because of cost trade-off."

Examples of bad learnings (do not write these):

- `2026-05-01-implemented-checkout.md` — describes a specific feature, not a reusable rule.
- `2026-05-12-use-aaa.md` — already in SKILL.md; duplication.
- `2026-05-15-tdd-feels-slow.md` — opinion without an observed trigger.

---

## Maintenance

When a learning becomes stale, **delete the file** rather than editing it in place. Git history preserves the older version; the folder reflects what is currently true.

When two learnings overlap, merge them and delete the duplicate.

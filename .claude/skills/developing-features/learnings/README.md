# Learnings — read before, append after

This folder accumulates project-specific or user-specific lessons that aren't obvious from the main SKILL.md. It is the **memory of the skill**.

The protocol below applies to **every** skill that has a `learnings/` folder.

---

## Before generating: read

When the skill activates, **list this folder and read any file whose name looks relevant** to the current task. Use `ls learnings/` and pattern-match on filenames — they are short and descriptive.

If a learning contradicts a default in `SKILL.md`, the learning wins. The learnings folder is where the user's reality overrides the generic best practice.

If `learnings/` is empty except for this README, that's fine — proceed with the SKILL.md defaults.

---

## After completing: append (only when warranted)

Append a new learning **only if** during the task you discovered something that:

- Was **not obvious** from `SKILL.md` or `references/`.
- Would have **changed how you started** if you had known it.
- Is **likely to apply again** to this user / this project / this domain.
- Is a **convention, constraint, or trap** — not a one-off fact about the current task.

Do **not** append:

- The summary of what you just did.
- Restatements of the SKILL.md content.
- Ephemeral facts about the current diff or file count.
- Anything you can't state as a rule that would apply on a future task.

---

## Append protocol

Create one file per learning, with this exact format.

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

Keep the body under 30 lines. If the learning grows beyond that, it's probably two — split it.

---

## How learnings interact with the SKILL.md

- `SKILL.md` is the **default behavior**.
- `references/` are the **detailed defaults**, loaded on demand.
- `learnings/` are the **overrides** — narrower, project- or user-specific, accumulated over time.

When a learning conflicts with the SKILL.md, follow the learning **and mention it** to the user so they can correct or refine it.

---

## Examples of good learnings for this skill (illustrative — delete when real ones land)

- `2026-02-12-no-redis-in-domain.md` — "The `services/billing/domain/` module must not import Redis or any cache client. Cache is wired up in `application/` only. Adding Redis to domain breaks the layered architecture review gate."
- `2026-03-04-webhook-provider-uses-jws.md` — "Provider X signs webhooks with detached JWS, not HMAC. Verify with their public key (`config/x_webhook_pubkey.pem`), not with `hmac.compare_digest`."
- `2026-04-21-no-warn-level.md` — "This project deletes `warn`-level logs as noise. Use `info` for recoverable anomalies and `error` for failures. Anything that would be `warn` either goes to `info` with `outcome=recovered` or is escalated to `error`."

Examples of bad learnings (do not write these):

- `2026-05-01-fixed-the-bug.md` — describes a specific bug, not a reusable rule.
- `2026-05-12-use-solid.md` — already in SKILL.md; duplication.
- `2026-05-15-i-prefer-dataclasses.md` — opinion without an observed trigger or constraint.

---

## Maintenance

When a learning becomes stale, **delete the file** rather than editing it in place. The git history preserves the older version; the folder reflects what is currently true.

When two learnings overlap, merge them and delete the duplicate.

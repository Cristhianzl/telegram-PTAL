# Learnings — read before, append after

This folder accumulates project-specific or user-specific lessons that aren't obvious from the main SKILL.md. It is the **memory of the skill**.

The protocol below applies to **every** skill that has a `learnings/` folder, not just this one.

---

## Before generating: read

When the skill activates, **list this folder and read any file whose name looks relevant to the current task**. Use `ls learnings/` and pattern-match on filenames — they are short and descriptive.

If you see a learning that contradicts a default in `SKILL.md`, the learning wins. The learnings folder is where the user's reality overrides the generic best practice.

If `learnings/` is empty except for this README, that's fine — proceed with the SKILL.md defaults.

---

## After completing: append (only when warranted)

Append a new learning **only if** during the task you discovered something that:

- Was **not obvious** from `SKILL.md` or `references/`.
- Would have **changed how you started** if you had known it.
- Is **likely to apply again** to this user / this project / this domain.
- Is a **convention, constraint, or trap** — not a one-off fact about the current diff.

Do **not** append:

- The summary of what you just did. That belongs in the PR description, not here.
- Restatements of the SKILL.md content.
- "Today I learned that the diff had 3 files" — ephemeral, not a lesson.
- Anything you can't state as a rule that would apply on a future task.

---

## Append protocol

Create one file per learning, in this folder, with this exact format:

**Filename:** `YYYY-MM-DD-short-kebab-slug.md` — use the real date.

**Body:**

```markdown
---
trigger: <one-line: when does this learning apply? what would make a future Claude need to read it?>
---

# <Title — what the rule is, in one short phrase>

**Context:** <one or two sentences — the situation where the lesson surfaced>

**Lesson:** <the rule itself, stated so it can be applied without re-reading the context>

**Why:** <the reason — incident, preference, convention, technical constraint>

**Apply when:** <the trigger condition, more precise than the frontmatter — e.g., "the repo has a `scopes.txt` at the root", or "the user is generating a PR for the `payments` service">
```

Keep the body under 30 lines. If the learning grows beyond that, it's probably two learnings — split it.

---

## How learnings interact with the SKILL.md

- `SKILL.md` is the **default behavior**.
- `references/` are the **detailed defaults**, loaded on demand.
- `learnings/` are the **overrides** — narrower, project- or user-specific, accumulated over time.

When a learning conflicts with the SKILL.md, follow the learning **and mention it** to the user in your response, so they can correct or refine it.

---

## Examples of good learnings (illustrative — delete these when real ones land)

- `2026-02-12-scope-allowlist.md` — "Use scopes only from `scopes.txt` at repo root; reject anything else."
- `2026-03-04-no-co-authored-by-trailer.md` — "Drop the `Co-Authored-By: Claude` trailer; this repo strips trailers in CI."
- `2026-04-21-payments-pr-requires-pii-checklist.md` — "PRs touching `services/payments/**` must include a `## PII Impact` section before merge."

Examples of bad learnings (do not write these):

- `2026-05-01-cool-pr-i-wrote.md` — describes a specific PR, not a reusable rule.
- `2026-05-12-conventional-commits.md` — already in `SKILL.md`; duplication.
- `2026-05-15-i-think-we-should.md` — opinion without an observed trigger.

---

## Maintenance

When a learning becomes stale (the convention changed, the constraint was lifted, the project evolved), **delete the file** rather than editing it in place. The git history preserves the older version; the folder should reflect what is **currently true**.

When two learnings overlap, merge them into one and delete the duplicate.

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

- The summary of what you just did. That belongs in the task output, not here.
- Restatements of the SKILL.md content.
- "Today I learned X had three files" — ephemeral, not a lesson.
- Anything you can't state as a rule that would apply on a future task.

---

## Append protocol

Create one file per learning, in this folder, with this exact format.

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

**Apply when:** <the trigger condition, more precise than the frontmatter>
```

Keep the body under 30 lines. If the learning grows beyond that, it's probably two learnings — split it.

---

## How learnings interact with the SKILL.md

- `SKILL.md` is the **default behavior**.
- `references/` are the **detailed defaults**, loaded on demand.
- `learnings/` are the **overrides** — narrower, project- or user-specific, accumulated over time.

When a learning conflicts with the SKILL.md, follow the learning **and mention it** to the user in your response, so they can correct or refine it.

---

## Examples of good learnings for this skill (illustrative — delete when real ones land)

- `2026-02-12-windows-not-supported.md` — "This monorepo declared Windows out of scope on 2026-02; the matrix is Linux + macOS only. Stop flagging POSIX-only paths as defects in `services/*`."
- `2026-03-04-pathlib-banned-in-legacy.md` — "Module `legacy/parser/` predates `pathlib`; do not introduce `pathlib` there — the consumers expect `str` paths. Use `os.path` instead."
- `2026-04-21-arm64-only-for-prod.md` — "Prod runs only on `linux/arm64` (Graviton). Skip multi-arch buildx when generating Dockerfiles; pin base images to arm64 digests."

Examples of bad learnings (do not write these):

- `2026-05-01-fixed-a-path-bug.md` — describes a specific bug, not a reusable rule.
- `2026-05-12-use-pathlib.md` — already in SKILL.md; duplication.
- `2026-05-15-windows-feels-clunky.md` — opinion without an observed trigger.

---

## Maintenance

When a learning becomes stale (the convention changed, the constraint was lifted, the project evolved), **delete the file** rather than editing it in place. The git history preserves the older version; the folder should reflect what is **currently true**.

When two learnings overlap, merge them into one and delete the duplicate.

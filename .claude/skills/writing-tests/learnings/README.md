# Learnings — read before, append after

This folder accumulates project-specific or user-specific lessons that aren't obvious from the main SKILL.md. It is the **memory of the skill**.

The protocol below applies to **every** skill that has a `learnings/` folder.

---

## Before generating: read

List this folder and read every file whose name looks relevant to the current task. Framework conventions, test-helper locations, flaky-zone alerts, coverage carve-outs, or fixture quirks live here. If a learning contradicts a default in `SKILL.md`, the learning wins — mention it to the user.

If `learnings/` is empty except for this README, proceed with the SKILL.md defaults.

---

## After completing: append (only when warranted)

Append a new learning only if you discovered something that:

- Was **not obvious** from `SKILL.md` or `references/`.
- Would have **changed how you started** if you had known it.
- Is **likely to apply again** to this user / project / test type.
- Is a **convention, constraint, or trap** — not a one-off fact about the current tests.

Do **not** append:

- A summary of the tests you just wrote.
- Restatements of the SKILL.md content.
- Ephemeral facts ("the suite has 423 tests today").
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

- `2026-02-12-pytest-anyio-not-asyncio.md` — "Project uses `pytest-anyio`, not `pytest-asyncio`. Use `@pytest.mark.anyio` and the `anyio_backend` fixture. `@pytest.mark.asyncio` silently skips."
- `2026-03-04-billing-uses-postgres-container.md` — "Tests in `tests/integration/billing/` use a real Postgres test container, never SQLite. Bugs in JSONB queries and transaction boundaries don't reproduce against SQLite."
- `2026-04-21-no-snapshot-tests.md` — "Project policy: no snapshot tests in `apps/web/`. They flake every UI tweak. Use explicit field-level assertions instead. Documented in ADR-0022."
- `2026-05-08-coverage-carve-out-generated-code.md` — "Generated code under `src/proto/` and `src/graphql/__generated__/` is excluded from the 80% coverage gate. See `pyproject.toml [tool.coverage.run] omit`."

Examples of bad learnings (do not write these):

- `2026-05-01-wrote-15-tests.md` — describes a specific test session, not a reusable rule.
- `2026-05-12-use-aaa.md` — already in SKILL.md; duplication.
- `2026-05-15-coverage-is-annoying.md` — opinion without an observed trigger.

---

## Maintenance

When a learning becomes stale, **delete the file** rather than editing it in place. Git history preserves the older version; the folder reflects what is currently true.

When two learnings overlap, merge them and delete the duplicate.

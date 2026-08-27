# Learnings — read before, append after

This folder accumulates project-specific or user-specific lessons that aren't obvious from the main SKILL.md. It is the **memory of the skill**.

The protocol below applies to **every** skill that has a `learnings/` folder.

---

## Before generating: read

List this folder and read every file whose name looks relevant to the current feature. Project-specific naming conventions, ADR numbering schemes, glossary inheritance rules, dashboard URLs, or template variants live here. If a learning contradicts a default in `SKILL.md`, the learning wins — mention it to the user.

If `learnings/` is empty except for this README, proceed with the SKILL.md defaults.

---

## After completing: append (only when warranted)

Append a new learning only if you discovered something that:

- Was **not obvious** from `SKILL.md` or `references/`.
- Would have **changed how you started writing the doc** if you had known it.
- Is **likely to apply again** to this user / project / domain.
- Is a **convention, constraint, or trap** — not a one-off fact about the current feature.

Do **not** append:

- A summary of the doc you just wrote.
- Restatements of the SKILL.md content.
- Ephemeral facts ("this feature had 6 ADRs").
- Anything you can't state as a rule applicable to a future doc.

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

- `2026-02-12-adrs-live-in-docs-adr.md` — "ADRs are stored as separate files under `docs/adr/NNNN-slug.md`, NOT inline in Section 5 of the feature doc. Section 5 lists `ADR-001: Title → docs/adr/0014-...md` and stops there."
- `2026-03-04-payments-glossary-inherited.md` — "All features in `services/payments/` inherit terms from `docs/payments/glossary.md`. Section 2 of a payments feature doc lists only **new** terms and links to the inherited glossary."
- `2026-04-21-c4-must-use-mermaid-only.md` — "No image links in Section 9. Renderers in our docs site only support Mermaid. PlantUML, .drawio, .png embeds break the build."
- `2026-05-08-section-10-required-for-cli.md` — "Section 10 (Platform Compatibility) is mandatory for any feature in `apps/cli/`. Optional for backend-only features that ship as Docker only."

Examples of bad learnings (do not write these):

- `2026-05-01-wrote-orders-feature.md` — describes a specific doc, not a reusable rule.
- `2026-05-12-use-ddd.md` — already in SKILL.md; duplication.
- `2026-05-15-templates-feel-rigid.md` — opinion without an observed trigger.

---

## Maintenance

When a learning becomes stale (the template changed, the ADR location moved), **delete the file** rather than editing it in place. Git history preserves the older version; the folder reflects what is currently true.

When two learnings overlap, merge them and delete the duplicate.

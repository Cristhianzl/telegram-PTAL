# Learnings — read before, append after

This folder accumulates project-specific or user-specific lessons that aren't obvious from the main SKILL.md. It is the **memory of the skill**.

The protocol below applies to **every** skill that has a `learnings/` folder.

---

## Before generating: read

List this folder and read every file whose name looks relevant to the current threat model. Project assets, established trust boundaries, known attacker personas, accepted risks, compliance constraints, and "in this system X is out of scope because Y" decisions live here. If a learning contradicts a default in `SKILL.md` or a reference, the learning wins — mention it to the user.

If `learnings/` is empty except for this README, proceed with the SKILL.md defaults.

---

## After completing: append (only when warranted)

Append a new learning only if you discovered something that:

- Was **not obvious** from `SKILL.md` or `references/`.
- Would have **changed how you started the threat model** if you had known it.
- Is **likely to apply again** to this system / user / domain.
- Is an **asset, boundary, attacker persona, accepted-risk decision, compliance rule, or recurring threat pattern** — not a one-off observation about today's model.

Do **not** append:

- A summary of the threat model you just produced.
- Restatements of the SKILL.md or reference content.
- Ephemeral facts ("this feature had 9 threats").
- Anything you can't state as a rule applicable to a future model.

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

**Why:** <the reason — incident, convention, compliance constraint, architecture fact>

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

- `2026-02-10-sso-provider-out-of-scope.md` — "The corporate SSO/IdP is trusted and out of scope; model only our integration with it (token validation, claims trust). Documented in ADR-0007."
- `2026-03-03-tenant-isolation-is-primary-asset.md` — "Cross-tenant data leakage is the highest-impact threat in this product; every data store and flow must be analyzed for IDOR / tenant-scoping. Rank any tenant-isolation gap as top priority."
- `2026-04-18-pii-triggers-linddun.md` — "Any feature touching `users.profile` or `payments.*` data also requires a LINDDUN pass, not just STRIDE — GDPR applies (EU users)."
- `2026-05-06-accepted-risk-internal-metrics.md` — "DoS on the internal metrics endpoint is an accepted risk (internal-only, behind VPN). Don't re-open it each pass; note it as accepted."

Examples of bad learnings (do not write these):

- `2026-05-01-modeled-the-login-flow.md` — describes a specific model, not a reusable rule.
- `2026-05-12-stride-has-six-categories.md` — already in references; duplication.
- `2026-05-15-this-system-is-risky.md` — opinion without an observed trigger.

---

## Maintenance

When a learning becomes stale (the architecture evolved, the boundary moved, the risk was finally mitigated), **delete the file** rather than editing it in place. Git history preserves the older version; the folder reflects what is currently true.

When two learnings overlap, merge them and delete the duplicate.

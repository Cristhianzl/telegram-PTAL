# PRD section skeleton

Purpose: the canonical PRD section order with one line of guidance each, plus the artifact ladder for matching weight to scope.

Use this skeleton for a full PRD. For a one-pager, keep the bold sections and collapse or drop the rest. Order matters: the problem and non-goals come *before* requirements so reviewers anchor on the *why* before the *what*.

## The sections

1. **Header / metadata** — owner (single DRI), status (Draft / In review / Approved / Shipped), target date, and links (designs, research, tickets, the engineering design doc). One glance tells a reader who owns it and how fresh it is.

2. **Problem statement / context** — the problem in the user's words, backed by **evidence** (data, research, support volume, quotes). Frame from the user's point of view, not the company's. No solution here.

3. **Goals / objective** — the outcome you want and how it ties to a **strategic objective** (company OKR, product vision). One to three goals; more than that means the PRD is doing too much.

4. **Non-goals / out-of-scope** — *emphasize this.* List explicitly what this will **not** do, and where useful, why. This is the #1 differentiator between a good and a great PRD: it prevents scope creep and short-circuits "but what about…" debates.

5. **Target users / personas** — who this is for, specifically. Segment if behavior differs. If "everyone", you have not narrowed enough.

6. **User stories / jobs-to-be-done** — what users are trying to accomplish, as stories ("As a … I want … so that …") or JTBD ("When [situation], I want to [motivation], so I can [outcome]").

7. **Requirements** — **functional** (what the system must do) and **non-functional** (performance, security, accessibility, reliability, privacy). Each requirement should be testable and traceable to a goal.

8. **Success metrics** — how you will know the outcome moved: North Star + leading/lagging, each with a target and timeframe, plus guardrails. Detail in [metrics-and-acceptance.md](metrics-and-acceptance.md).

9. **Scope** — what ships **now** vs. what is explicitly **deferred** to a later phase. Pairs with non-goals: non-goals are never; deferred is later.

10. **Dependencies / assumptions** — what must be true or available (teams, services, data, legal), and **how each assumption was validated** (or that it is still unvalidated).

11. **Risks** — what could go wrong (technical, market, adoption, compliance) and the mitigation or trigger for each.

12. **Open questions** — unresolved decisions, each with an **owner and a resolution date**. An open question with no owner is a hidden blocker.

13. **Rollout / launch** — phasing, flags, beta/GA gates, comms, and what "launched" means.

14. **Appendix** — supporting research, mockups, prior art, glossary, decision log.

## The artifact ladder

Match the artifact's weight to the bet's scope, risk, and audience. Escalate only when warranted; start lighter and grow.

```
one-pager  →  mini-PRD  →  full PRD  →  engineering design doc
(small,       (small,       (new product,   (the technical HOW —
 1 team)       defined        cross-functional, architecture, data,
               scope)         higher risk)      APIs; separate doc)
```

The first three own the **what/why**. The design doc owns the **how** and is a separate artifact — the PRD links to it. Collapsing the design doc into the PRD buries product decisions under technical ones.

## Sources

- [Atlassian — How to write a product requirements document](https://www.atlassian.com/agile/product-management/requirements)
- [Aha! — Product requirements documents](https://www.aha.io/roadmapping/guide/requirements-management/what-is-a-good-product-requirements-document)
- [Lenny Rachitsky — My favorite product templates](https://www.lennysnewsletter.com/p/my-favorite-templates-issue-37)
- [SVPG / Marty Cagan — Assessing product opportunities](https://www.svpg.com/assessing-product-opportunities/)

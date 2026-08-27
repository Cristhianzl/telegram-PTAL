---
name: writing-prd
description: Author product requirements — PRDs, product specs, one-pagers, PR-FAQs, product briefs — that lead with the problem, define testable success metrics and acceptance criteria, and make non-goals explicit. Use when the user asks to write a PRD, write/define product requirements, write a product spec, write a one-pager, write a PR-FAQ, "write a spec", "define requirements", scope a feature, or draft a product brief. Also use when the request is a half-baked feature idea that needs a problem framed before any solution. Not for engineering design docs / the technical "how" — that is a separate design doc (the PRD owns the what/why, the design doc owns the how).
license: MIT
---

# Writing a PRD

Produce a product requirements artifact that states the **problem before the solution**, ties to a measurable outcome, makes **non-goals explicit**, and defines acceptance criteria a team can test. The PRD is the *what* and *why*; the *how* belongs in a separate engineering design doc. Match the artifact's weight to the bet's scope and risk — a one-pager for a small change, a full PRD for a new product, a PR-FAQ for a strategic bet whose desirability is unproven.

## Read first (always)

List `learnings/` and read every file relevant to the current product area, team, or artifact type. Project-specific section templates, required metrics, approval workflows, naming, or "what counts as a non-goal here" live there and override the defaults in this SKILL.md. If a learning conflicts with this file, **the learning wins** — mention it to the user.

If `learnings/` holds only its README, proceed with the defaults below.

## Workflow

1. **Decide whether — and what — to write.** Run a lightweight opportunity assessment before committing to a heavy artifact. Use the SVPG/Cagan 10 questions ([Inspired, Marty Cagan](https://www.svpg.com/assessing-product-opportunities/)): what problem, for whom, how big is the opportunity, what alternatives exist, why are we best suited, why now, how will we go to market, how will we measure success, what are the critical factors to succeed, and — given all that — what is the recommendation. For strategic or ambiguous bets where desirability/viability is unproven, write the Amazon **PR-FAQ first** (see [references/pr-faq.md](references/pr-faq.md)).
   → verify: you can state the problem and the target user in one sentence each, without naming a solution.

2. **Pick the artifact weight.** One-pager / mini-PRD for a small, well-understood change (roughly < 2 weeks, one team). Full PRD for a new product, cross-functional work, or higher risk/ambiguity. PR-FAQ for a strategic bet where the value is not yet validated. When in doubt, start lighter and grow the artifact — never start with a heavyweight template for a small change.
   → verify: the weight matches scope, risk, and audience; you can justify the choice in one line.

3. **Problem before solution, always.** Write the section skeleton from [references/prd-structure.md](references/prd-structure.md). Fill the problem statement, evidence, and target users *before* describing any solution. If you cannot articulate the problem with evidence, stop and go back to step 1.
   → verify: the problem statement cites evidence (data, research, support tickets, quotes) — not an assumption.

4. **Define success metrics and testable acceptance criteria.** Pick a North Star and leading/lagging metrics with a **target and timeframe each**, plus guardrail/counter-metrics. Write acceptance criteria in Given/When/Then, behavior-not-implementation, covering edge/boundary/error cases. Keep **acceptance** (did we ship it right?) separate from **success** (did it move the outcome?). See [references/metrics-and-acceptance.md](references/metrics-and-acceptance.md).
   → verify: every metric has a number and a date; every requirement maps to at least one testable acceptance criterion.

5. **Lint against the anti-patterns (below) before sharing.** Walk the list explicitly. The most common gap between a good and a great PRD is a missing or weak **non-goals** section.
   → verify: each anti-pattern checked off; non-goals section is present and specific.

## The artifact ladder

Match weight to scope, risk, and audience — escalate only when the bet warrants it.

| Artifact            | When                                                        | Owns          |
|---------------------|------------------------------------------------------------|---------------|
| One-pager / mini-PRD| Small, well-understood change; one team; low risk           | what / why    |
| Full PRD            | New product; cross-functional; higher risk or ambiguity     | what / why    |
| PR-FAQ              | Strategic bet; desirability/viability unproven              | what / why    |
| Engineering design doc | The technical *how* — architecture, data model, APIs     | how (separate)|

Never collapse the design doc into the PRD. The PRD answers *what should be true and why*; the design doc answers *how we will build it*. Conflating them buries product decisions under technical ones and makes both harder to review.

## Anti-patterns — lint before sharing

- **Solutionizing before a validated problem.** A solution written before the problem is evidenced. Fix: lead with problem + evidence; defer the solution.
- **No metrics, weak metrics, or vanity metrics.** "Increase engagement" with no number, or counting signups that never activate. Fix: North Star + leading/lagging, each with a target and timeframe; add guardrails. See [references/metrics-and-acceptance.md](references/metrics-and-acceptance.md).
- **Missing non-goals / out-of-scope.** The #1 gap between a good and a great PRD. Without it, scope creeps and reviewers argue about things you never intended to do. Fix: list explicitly what this will *not* do — and, where useful, why.
- **Ambiguous or untestable requirements.** "It should be fast", "intuitive UX". Fix: rewrite as Given/When/Then with concrete, measurable conditions.
- **Conflating acceptance with success criteria.** Treating "we shipped the feature" as proof the outcome moved. Fix: acceptance = built correctly (binary, per requirement); success = outcome moved (metric vs. target over time). Keep them in separate sections.
- **Over-rigid spec with no room to learn.** A locked spec that forbids discovery during build. Fix: state the problem and outcome firmly; hold the solution detail loosely where the team should still learn. Mark open questions with an owner and a resolution date.
- **Smuggling the *how* into the PRD.** Architecture, schemas, library choices. Fix: move them to the design doc; the PRD references it.

## Hard rules

- **Problem before solution.** No solution section is complete until the problem section cites evidence.
- **Non-goals are mandatory** in a full PRD; recommended in a one-pager.
- **Every metric needs a target and a timeframe.** No bare directions ("more", "better").
- **Acceptance ≠ success.** Two distinct sections, never merged.
- **The PRD is the what/why.** Defer the how to a linked engineering design doc.
- **English only**, regardless of conversation language.
- **Don't invent numbers.** If you lack a real target, mark it `TBD` and flag it — don't fabricate.
- **No commits.** Print the PRD in chat or write to the planned path only if the user asks to save; never run `git commit`/`add`/`push`.

## Capture a learning (final step)

Before closing, ask: *did I encounter a section template, required metric, approval workflow, non-goal convention, or product-area quirk not in this SKILL.md or `references/`?* If yes, append a `learnings/YYYY-MM-DD-slug.md` (protocol in [learnings/README.md](learnings/README.md)). If no, skip.

## See also

- [references/prd-structure.md](references/prd-structure.md) — the canonical PRD section skeleton with one line of guidance each, plus the artifact ladder.
- [references/metrics-and-acceptance.md](references/metrics-and-acceptance.md) — success metrics, acceptance vs. success criteria, and prioritization frameworks (RICE, MoSCoW, Kano, Impact/Effort).
- [references/pr-faq.md](references/pr-faq.md) — the Amazon Working Backwards PR-FAQ template and when to use it.
- `../documenting-features/references/communication.md` — answer-first structure (Minto Pyramid / SCQA); lead the PRD with the recommendation, not the build-up.
- `documenting-features` skill — the living-doc companion: once the feature ships, the PRD's problem/scope/criteria feed the feature doc.
- `developing-features` skill — turns the PRD's acceptance criteria into implementation and tests.
- `learnings/` — project-specific PRD conventions accumulated over time.

# Communicating clearly (answer-first)

How to structure any prose this project produces — docs, PRDs, PR descriptions, review summaries, status updates, proposals, post-mortems, even chat answers. Based on the **Minto Pyramid Principle** (Barbara Minto).

## The one rule: lead with the answer

Put the **conclusion / recommendation / "so what" first**, then the reasons that support it, then the detail. The reader should get the takeaway in the first sentence and be able to stop there. Never make them read to the end to find the point.

## The pyramid

1. **Governing thought** — one sentence: the answer / recommendation / main message.
2. **Key arguments** — 3–5 grouped points that support it, each a mini-conclusion. Group related ideas; keep the groups distinct (no overlap, no obvious gaps).
3. **Detail / evidence** — data, steps, code, caveats — placed under the argument it supports.

Order the arguments logically (by importance, by sequence, or by structure) — not in the order you happened to discover them.

## Open with SCQA (docs, proposals, PRDs)

A short intro that earns the answer:

- **Situation** — the stable context the reader already accepts.
- **Complication** — what changed / the problem / the tension.
- **Question** — the question it raises (often left implicit).
- **Answer** — your governing thought, which becomes the top of the pyramid.

## Per artifact

- **PR description** — first line: what it does and why it's safe to merge; then grouped changes; then test notes / risks.
- **Code review** — verdict first (block / approve-with-nits / approve), then findings grouped by severity.
- **PRD** — the recommendation/goal (or PR-FAQ) up front; then problem, scope, metrics.
- **Doc / RFC** — a TL;DR with the decision; then context (SCQA) and detail.
- **Status update / incident** — current state + impact first; then timeline and detail.
- **Chat answer** — the answer in the first line; reasoning after.

## Anti-patterns

- Burying the conclusion at the bottom ("…and therefore we should X").
- Narrating how you figured it out chronologically instead of leading with the result.
- A flat wall of bullets with no governing thought.
- More than ~5 top-level points → they're ungrouped; regroup them.

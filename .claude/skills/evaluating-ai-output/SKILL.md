---
name: evaluating-ai-output
description: Evaluate non-deterministic LLM/AI output with evals instead of one-shot "it worked" — define expected behavior first, measure pass@k / pass^k, and grade with code / model / human graders. Use when building or changing an AI/LLM feature, an agent, a prompt, a RAG pipeline, or a classifier, where a single good run is not proof of correctness. Complements writing-tests (deterministic logic) and developing-features-tdd.
license: MIT
---

# Evaluating AI output

Code is deterministic; LLM output isn't. A feature that "worked once" can fail the next call on the same input. **Evals are the unit tests of AI work** — they measure how *often* and how *well* the output meets the bar, not just that it can.

## Read first (always)

List `learnings/` and read anything relevant — provider quirks, rubric calibration, and known-flaky cases for this project belong there.

## Define expected behavior BEFORE you implement

Write the eval first: the inputs, what a good output looks like, and what must never happen. If you can't state how you'd grade it, you don't yet understand the feature.

## Measure across repeated trials

Run each case **k times** (LLM output varies) and report:

- **pass@k** — *at least one* of k attempts succeeds. Measures **capability** ("can it do this at all?"). Typical target: pass@3 > 90%.
- **pass^k** — *all* k attempts succeed. Measures **stability/reliability** ("does it do this every time?"). Use for **critical paths** (auth, money, irreversible actions). pass^3 means 3 consecutive clean runs.

A feature can have high pass@k but low pass^k — impressive once, unreliable in production. Match the metric to the risk.

## Three graders (use the cheapest that's trustworthy)

| Grader | How | Use for |
|---|---|---|
| **Code-based** | Deterministic check — regex/`grep`, schema/JSON validation, an assertion, a tool call that must appear | Anything machine-verifiable (format, presence, exact values). Always prefer this. |
| **Model-based** | A model scores the output 1–5 against a written rubric | Quality/judgment that code can't check (relevance, tone, reasoning). Calibrate the rubric on a few human-labeled examples. |
| **Human** | A person reviews, tagged risk LOW / MED / HIGH | High-stakes or ambiguous cases. **Never fully automate security review** — keep a human in the loop there. |

## Build the eval set

Cover **representative** cases, **adversarial/edge** cases (the failure modes from your threat model — see `skills/threat-modeling`), and **regression** cases (every bug becomes a permanent eval). Keep a **baseline** (committed scores) and gate changes on "no regression vs. baseline". Store the eval set, the baseline, and run logs alongside the feature.

## Boundary with the testing skills

- **Deterministic logic** (parsing, math, control flow around the model) → `skills/writing-tests` / `developing-features-tdd`. Don't write an eval for what a unit test can assert.
- **Non-deterministic output quality** (does the model's answer meet the bar, reliably) → here.
- For AI runtime resilience (timeouts, fallbacks, circuit breakers, kill switch), see `developing-features-tdd/references/ai-runtime.md`.

## Capture a learning

If you find a rubric that calibrates well, a recurring failure class, or a provider quirk, append a `learnings/YYYY-MM-DD-slug.md` (or use `/learn`).

## See also

- `skills/writing-tests`, `skills/developing-features-tdd` — deterministic coverage.
- `skills/threat-modeling` — abuse cases feed the adversarial eval set.
- `skills/building-langflow-components` — when the AI feature is a Langflow component.

---
name: running-agent-loops
description: Run a multi-step or unattended agent loop safely — sequential pipelines, implement→review→fix (PR) loops, parallel fan-out, and RFC→DAG orchestration. Use when the user wants to "loop", "run autonomously", "iterate until it's done", chain `claude -p` calls, fan out parallel agents on a spec, or set up an implement→verify→commit cycle. Covers the cross-iteration context bridge and the de-sloppify pass. Not for a single one-shot task.
license: MIT
---

# Running agent loops

Patterns for driving work across **many steps or iterations** — possibly unattended — without the run drifting, looping, or shipping slop. The core trade is **isolation vs. continuity**: a fresh context per step avoids bleed, but you must deliberately carry forward what matters.

## Read first (always)

List `learnings/` and read anything relevant. Project-specific loop conventions, budgets, and stop signals live there and override this file.

## Principles (apply to every loop)

- **Fresh context per step beats one long context.** Each `claude -p` / subagent call starts clean — no bleed, no drift. The cost: it forgets. Bridge it deliberately (below).
- **The reviewer is never the author.** Put review/verify in a separate step/context so it isn't anchored to the implementation.
- **Two focused passes beat one constrained pass.** Don't pile negative instructions onto the implementer — add a separate **de-sloppify** pass (below). Quality from constraints decays; quality from a dedicated cleanup step doesn't.
- **Every loop has an explicit stop condition.** Bound it by max-runs, max-cost, max-duration, or a completion signal — never "until it feels done".

## The patterns (pick the simplest that fits)

| Pattern | Use when | Shape |
|---|---|---|
| **Sequential pipeline** | A known series of steps on one unit of work | implement → de-sloppify → verify → commit, each a fresh context |
| **PR loop** | Iterate a branch toward green | branch → implement → review → fix → run checks → repeat until checks pass or budget hits |
| **Parallel fan-out** | N independent variations/units of the same spec | an orchestrator **assigns** each agent a distinct direction + index (don't rely on agents to self-differentiate); run in waves of 3–5 |
| **RFC → DAG** | A large feature with dependencies | decompose into work units + a dependency DAG; run each in an isolated worktree; land via a merge queue; deeper review tier for riskier units |

## The cross-iteration context bridge

Independent steps lose memory. Persist a single **`SHARED_TASK_NOTES.md`** (or the project's plan file): the agent **reads it at the start of each iteration and updates it at the end** — current objective, what's done, what's next, decisions, and gotchas. This is what turns isolated calls into a coherent run.

## The de-sloppify pass

After each implement step, run a **dedicated cleanup pass** (separate agent/context) whose only job is to remove slop: dead code, stray comments, debug prints, half-finished branches, inconsistent naming, over-engineering. Keep it separate from the implementer so neither job is diluted.

## Safety before you launch an unattended loop

- Baseline **tests pass** before iteration 1.
- An explicit **stop condition** and a **budget** (runs / cost / time).
- **Rollback** ready: work on a branch or in a worktree, commit per green step.
- **Checkpoints** so you can diff now-vs-then.
- **Escalate to the human** when: no progress across two consecutive checkpoints, the same error/stack repeats, the diff drifts off-objective, or the budget is exceeded. Stalling silently is the failure mode.

If a loop is wedged (looping, repeating an identical failed action, burning budget), switch to `skills/debugging-agent-runs`.

## Capture a learning

If you hit a loop convention, stop-signal, or trap not covered here, append a `learnings/YYYY-MM-DD-slug.md` (or use `/learn`).

## See also

- `/dual-review` command — the adversarial two-reviewer convergence step for a PR loop.
- `skills/debugging-agent-runs` — recover a wedged run.
- `skills/reviewing-code`, `skills/writing-pull-requests` — the review/PR steps a loop calls.

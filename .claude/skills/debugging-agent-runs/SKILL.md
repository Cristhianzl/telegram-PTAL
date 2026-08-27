---
name: debugging-agent-runs
description: Recover when the AGENT itself is stuck — looping, repeating the same failed action, burning tokens/budget, overflowing context, drifting from the goal, or acting on stale state. Use when a run isn't converging and reworded retries aren't helping. This is a workflow you follow, not a hidden runtime that auto-heals. For bugs in the product code use fixing-bugs.
license: MIT
---

# Debugging agent runs

When the problem is the **run**, not the code: the agent keeps trying variations of the same thing, talks itself in circles, or acts on what it remembers instead of what's true. This skill is a recovery loop to break that. Be honest: it's a discipline you apply — it can't magically fix a stuck process.

## Read first (always)

List `learnings/` and read anything relevant — past failure modes specific to this project belong there.

## Recognize the failure mode

| Symptom | Likely cause |
|---|---|
| Same action retried with slightly reworded prompts | Loop / no new information between attempts |
| Output truncated, "forgetting" earlier context | Context overflow → compact or restart with a tighter scope |
| Repeated 429 / timeouts | Rate limit → back off, batch, slow down |
| Edits land on the wrong branch/files; conflicts | Branch / diff drift → re-verify git state |
| Confident claims that don't match reality | Stale state → trusting memory over the world |
| Scope keeps widening while nothing lands | Scope creep → shrink to one verifiable thing |

## The recovery loop

Run these **in order** — most stuck runs are fixed by step 1–2:

1. **Restate the real objective in one sentence.** If you can't, that's the bug. Re-anchor to the user's actual goal, not the sub-task you spiraled into.
2. **Verify world state — don't trust memory.** Re-read the files, `git status`/`diff`, and the actual test/command output. The model's recollection is often stale; reality isn't.
3. **Shrink the failing scope.** Reduce to the smallest reproduction or the single failing unit. Big scope hides the cause.
4. **Run one discriminating check** — the single command/test whose result tells you which branch of the problem you're in. Don't guess; get one new bit of information.
5. **Only then retry — and change the approach,** not just the wording. If the new info didn't change your plan, you're still looping.

## Anti-patterns

- Retrying the same action three times with reworded prompts (no new information = a loop).
- Trusting your memory of the code/state instead of re-reading it.
- Widening scope when stuck (the opposite of step 3).
- Fabricating that something "should now work" without a check that proves it.

## When to stop and ask the human

After one full recovery pass with no progress — or on a budget/rate-limit wall — **stop and surface it**: the objective, what you verified, the one check you ran, and where it's blocked. A clear handoff beats more silent retries.

## Capture a learning

If a project-specific failure mode or recovery trick shows up, append a `learnings/YYYY-MM-DD-slug.md` (or use `/learn`).

## See also

- `skills/fixing-bugs` — once you've isolated a real code bug, switch to the reproduce-first fix workflow.
- `skills/running-agent-loops` — stop conditions and escalation for unattended loops.

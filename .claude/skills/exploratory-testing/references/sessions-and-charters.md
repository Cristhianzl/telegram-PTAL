# Sessions & charters

Purpose: how to manage exploratory testing as accountable work — how to scope a charter, time-box a session, record it, and debrief — without scripting away the exploration.

The canonical management method is **Session-Based Test Management (SBTM)**, from Jonathan and James Bach at [Satisfice](https://www.satisfice.com/sbtm).

---

## SBTM — Session-Based Test Management

SBTM makes exploration trackable while keeping it open-ended. Three pieces:

1. **Charter** — a short mission statement that bounds one session (see below).
2. **Session** — an uninterrupted, time-boxed block (45–90 min) spent pursuing one charter. Uninterrupted is the point: an interrupted session isn't a session.
3. **Session sheet** — the reviewable record: charter, notes, bugs, issues, and the time breakdown. It's what makes exploration auditable.

### The TBS metrics

At debrief, estimate how the session time split across three activities (they overlap; estimate percentages that sum to ~100%):

- **T — Test design and execution** — actually exploring and running tests. You want this high.
- **B — Bug investigation and reporting** — chasing down and writing up problems found.
- **S — Setup** — getting the environment, data, and tools ready to test.

A session that's 70% setup is telling you the product is hard to test (low **Testability**) — that's a finding in itself.

---

## Thread-Based Test Management (TBTM)

The lighter-weight sibling, also from the Bachs, for chaotic or interrupt-heavy environments where fixed time-boxes don't survive contact with reality.

- The unit is a **thread** — a line of investigation, not a fixed block of time. Threads are *interruptible*: you pause one and resume it later.
- Threads are tracked on a **mind-map** rather than session sheets, branching as new questions appear.
- Use TBTM when you can't protect a 90-minute block, when work is highly reactive, or early in learning a product when you don't yet know enough to write tight charters. Use SBTM when you can secure focused time and want clean coverage accounting.

---

## Writing and scoping a charter

Use Hendrickson's formula from *Explore It!*:

> **Explore (target) with (resources) to discover (information).**

- **Target** — the area, feature, or flow. One mission per charter.
- **Resources** — tools, data, conditions, personas, or a tour to apply. This is the lever that creates pressure: "with a slow network", "with malformed CSVs", "with a saboteur tour", "as a screen-reader user".
- **Information** — the class of problem you're hunting. Keeps you from sliding into the happy path.

Goldilocks scope:
- Too broad — "Explore the app" — produces an aimless click-around.
- Too narrow — a list of exact steps with an exact expected value — is a script; hand that to `writing-tests`.
- Just right names a target, applies a resource, and states what you're trying to learn.

A good charter is something you could not fully satisfy in 90 minutes, but could make real progress on. If you'd finish it in 5 minutes, broaden it; if it'd take a week, split it.

---

## Session-sheet template (copy-paste)

```markdown
# Session sheet

Charter:        Explore ___ with ___ to discover ___.
Tester:         <name>
Date / time:    <UTC start–end>
Build / env:    <build id, OS, browser/device, config, test data set>

## Areas covered
- <feature / screen / data path actually touched>

## Notes (chronological)
- <what I tried, what I saw, what surprised me — the running log>

## Bugs
- BUG-### <one-line summary>  (severity: __, priority: __)  → full report filed separately

## Issues / open questions
- <things I couldn't resolve, missing oracle, "is this intended?">

## Risks
- <areas that feel fragile and deserve more attention>

## Follow-up charters
- Explore ___ with ___ to discover ___.
- Explore ___ with ___ to discover ___.

## TBS breakdown (must total ~100%)
- Test design & execution: __%
- Bug investigation & reporting: __%
- Setup: __%
```

---

## PROOF debrief

Close every session with a short structured summary — the basis for a debrief conversation and for deciding what to charter next.

- **P — Past** — what you did this session.
- **R — Results** — what you found (bugs, risks, observations).
- **O — Obstacles** — what got in the way of testing (blocked features, missing data, poor testability).
- **O — Outlook** — what's left unexplored; where the risk still is.
- **F — Feelings** — your gut read on the quality and risk of the area. Subjective, and valuable — uneasiness is often the first signal of a deeper problem.

Most debriefs end by spawning follow-up charters. The search is never "done"; it's bounded by the risk you've covered and the time you have.

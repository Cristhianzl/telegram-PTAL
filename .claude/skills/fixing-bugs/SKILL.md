---
name: fixing-bugs
description: Fix a bug using strict TDD — UNDERSTAND → REPRODUCE (RED) → VERIFY RED → FIX → VERIFY GREEN → VALIDATE → REFACTOR. Use when the user reports a bug, asks to fix something, mentions a Sentry/error/regression, or pastes a stack trace. Bugs require a failing test that proves the bug existed before any code change. For building new features use developing-features-tdd; for non-TDD code work use developing-features.
license: MIT
---

# Fixing Bugs

Never fix a bug before proving it exists with a failing test. The fix is only valid if a test transitions from RED (failing) to GREEN (passing).

## Read first (always)

List `learnings/` and read every file relevant to the current bug. Project-specific test conventions, framework quirks, recurring bug patterns, or hot zones live there and override the defaults in this SKILL.md. If a learning conflicts with this file, **the learning wins** — mention it to the user.

## Tradeoff — when to apply, when to lighten up

Apply the full cycle for **any defect in production code** the user is reporting (Sentry alert, support ticket, "this broke", regression). The discipline pays off because the test becomes a permanent sentinel against the same bug returning.

Lighten the formality only when fixing a typo, a build script, or a one-off script the user explicitly labels as throwaway. Even then, run the test suite afterwards.

## The cycle (strict order)

```
1. UNDERSTAND   →  Read the bug report; identify expected vs actual; locate the OS/platform
2. REPRODUCE    →  Write a failing test that triggers the EXACT error path
3. VERIFY RED   →  Run the test; it MUST fail for the reason in the bug report
4. FIX          →  Write the MINIMUM code change to make the test pass
5. VERIFY GREEN →  Run the same test; it MUST now pass
6. VALIDATE     →  Run the FULL suite (CI matrix if platform-suspect); nothing else breaks
7. REFACTOR     →  Optional; only if the fix introduced duplication. Tests stay GREEN
```

If you skip steps or change the order, the fix is invalid. Full per-phase detail in `references/cycle.md`.

## Phase 1 — UNDERSTAND

Decompose the bug report:

```
GIVEN:    <precondition / initial state>
WHEN:     <action that triggers the bug>
THEN:     <what actually happens — the bug>
EXPECTED: <what should happen — the fix target>
```

If you can't extract all four → **ask for clarification**, don't guess.

**Identify the platform dimension.** Bug reports rarely state the OS. Check user agent, Sentry tags, container image, support metadata. Bugs in the following areas are platform-suspect by default — filesystem operations, text encoding / line endings, subprocess / shell, temp / config dir resolution, process signals, networking. If the bug is platform-suspect, **the test must run on the affected OS** in Phase 3.

**Pick an investigation technique** (full list in `references/investigation.md`):

| Technique                | When                                                        |
|--------------------------|-------------------------------------------------------------|
| Read error logs          | Always start here                                            |
| Reproduce locally        | Always — if you can't reproduce, you can't fix              |
| `git bisect`             | The bug is a regression and you don't know which commit broke it |
| `git log` / `git blame`  | You need the history of the affected code                    |
| Debugger / breakpoints   | You need to inspect runtime state                            |
| 5 Whys                   | Bug is recurring or the surface cause looks too simple       |

Do **not** open the editor and start changing things. Systematic investigation beats trial-and-error.

## Phase 2 — REPRODUCE

The test is the proof the bug exists. No test = no proof = no fix.

The test must:

1. Reproduce the exact scenario from the bug report.
2. Exercise the **exact** error path — not a similar one. If the report says `GeneratorExit`, your test triggers `GeneratorExit`, not `CancelledError`.
3. Assert the **expected** behavior, not the buggy behavior.
4. Fail before the fix.
5. Pass after the fix.
6. Fail with a message consistent with the bug report.

**Naming:** `should_[expected_behavior]_when_[condition_that_triggered_bug]`. Examples and full quality rules in `references/cycle.md`.

→ verify: the test exercises the precise code path described in the stack trace or repro steps, and the failure message names the bug.

## Phase 3 — VERIFY RED

Run the test. It MUST fail. Document:

```
TEST: <name>
STATUS: FAILED
ERROR: <actual error or assertion message>
MATCHES BUG REPORT: Yes — <one-line tie-back to the report>
```

If the test passes before the fix, one of three things is true:

- The bug isn't reproducible with your scenario → revisit Phase 1.
- The assertions are wrong → revisit Phase 2.
- The bug is already fixed in the current code → confirm against the latest and close the ticket.

If the test fails for a **different** reason than the report:

- Setup error in Arrange → fix the fixture.
- The bug has a different root cause → revisit Phase 1.
- Mocks or fakes are wrong → check them.

Don't proceed to FIX until the failure reason is clean.

## Phase 4 — FIX

Change the **minimum** code necessary. Do **not** refactor, optimize, or add features in the same diff.

**Root cause analysis** (write this down before coding):

```
ROOT CAUSE:
  File: <path>
  Function: <name>
  Line: <number>
  Issue: <one-line explanation of why the bug occurs>
  Fix: <one-line description of the minimum change>
```

**Follow existing codebase patterns.** Before writing the fix, search for similar error handling or patterns already established. If prior art exists (e.g., another function already handles the same exception type), follow that pattern.

**Pre-existing issues discovered during the fix:**

| Situation                                                                            | Action                                                          |
|--------------------------------------------------------------------------------------|-----------------------------------------------------------------|
| Pre-existing issue is in the lines you're touching and trivial to fix                | Fix it in the same PR, note "pre-existing fix" in description   |
| Pre-existing issue is adjacent and involves moderate refactoring                     | Separate commit in the same PR, or a follow-up ticket           |
| Pre-existing issue is in a different area                                            | Separate ticket — don't touch it here                           |

The test: does fixing this pre-existing issue make my bug fix **safer or easier to review**? If yes, include it (separate commit). If no, ticket it.

**Scope rules:**

| Allowed                                                              | Not allowed                                            |
|----------------------------------------------------------------------|--------------------------------------------------------|
| Fix the root cause of the reported bug                               | Refactor the entire module                             |
| Add a guard clause for the edge case                                 | Rewrite the function "for clarity"                     |
| Correct a conditional logic error                                    | Change the API contract                                |
| Fix an off-by-one error                                              | Add new parameters or endpoints                        |
| Handle a missing null check                                          | Reorganize the file structure                          |
| Fix a pre-existing issue in the lines you're touching (note in PR)   | Fix pre-existing issues in unrelated areas             |

## Phase 5 — VERIFY GREEN

Run the same test from Phase 2. It MUST now pass. Document:

```
TEST: <name>
STATUS: PASSED
BEFORE FIX: <failure message>
AFTER FIX: <pass>
ROOT CAUSE: <one-line>
```

If it still fails, the fix is wrong or incomplete. Don't proceed to VALIDATE.

## Phase 6 — VALIDATE

Run the **full** test suite. All tests must pass.

If a pre-existing test breaks:

1. **Analyze** — is the breaking test correct, or was it asserting the buggy behavior as correct?
2. **If the test was wrong** (it locked in the bug) → fix the test in the same PR, document why.
3. **If the test was right** → your fix has a side effect. Return to Phase 4.
4. **Never delete a failing test** to make CI green.

**Cross-platform validation.** If the bug was platform-suspect, the test must be green on **every supported OS**, not just yours. Run through CI on the affected OS before declaring done. If the suite didn't previously run on that OS, extending the CI matrix is part of **this** fix, not "a follow-up PR". See `ensuring-cross-platform` skill.

**Regression checklist:**

- [ ] New bug-fix test passes (GREEN).
- [ ] All existing unit tests pass.
- [ ] All existing integration tests pass (if applicable).
- [ ] No new warnings.
- [ ] Linter, formatter, type checker pass.
- [ ] CI matrix passes on every supported OS.

## Phase 7 — REFACTOR (optional)

Only if the fix introduced duplication or reduced clarity. Tests must stay GREEN. Small steps. Re-run tests after each. Separate commit from the fix.

## The bug-fix test is permanent

The test you wrote in Phase 2 is **the permanent sentinel against this bug returning**. Recurring bugs are one of the most common problems in software. Without a test, a future refactor can silently reintroduce the exact same defect.

- Never delete a bug-fix test unless the feature it tests is removed.
- Bug-fix tests must run in CI/CD on every build and PR.
- If a bug comes back, either the test was deleted, the test was too narrow, or someone overrode it. Investigate, don't just re-fix.

## Logging quality in bug fixes

A bug fix often adds log statements (error handlers, fallback paths, disconnection events). These logs **must** meet the same quality bar as any other production code.

- **Correlation context is mandatory.** Every new log includes IDs (request, conversation, user, resource) that let a future investigator trace the event.
- **Event names are unique.** Two different error conditions must NOT log the same event name — collisions break filtering in Sentry/CloudWatch/Datadog.
- **Follow the codebase convention.** Check the existing logging style in the file before adding a new statement.
- **Platform context for platform bugs.** If the root cause was OS-specific, include `platform=sys.platform` (or equivalent) in the log.

Full guidance in `developing-features/references/observability.md`.

## Code quality still applies

A bug fix does NOT exempt you from project standards. All `developing-features` rules apply: no PII in logs, no secrets, file structure hard limits, strong typing, no silent failures, linters/formatters pass.

If your fix would violate a code quality rule → find a different approach.

## Final step: capture a learning

Before closing the task, ask: *did I encounter a bug pattern, framework quirk, OS trap, or recurring shape that wasn't in this SKILL.md or `references/`?* If yes, append a `learnings/YYYY-MM-DD-slug.md`. If no, skip.

Pay particular attention to **bug clusters**: if the same kind of bug keeps recurring (off-by-one, null pointer, timezone, encoding), the learning should describe the **systemic root cause and where to prevent it**, not just this instance.

## Output format

1. **Root cause analysis** — WHERE, WHY, WHAT (concise, 3–5 lines). Include 5 Whys for complex or recurring bugs.
2. **Characterization tests** (if applicable) — capture existing behavior as a safety net.
3. **Failing test** — the test that reproduces the bug (RED).
4. **Fix** — the minimum code change.
5. **Passing test** — confirmation the same test now passes (GREEN).
6. **Regression check** — confirmation all existing tests pass.
7. **Brief explanation** — what was changed and why, in 2–3 bullets.
8. **Investigation notes** (complex bugs) — what was investigated and ruled out.

## See also

- `skills/validating-in-reality` — when the user provides a cURL/repro, also prove the fix against the running system (request before/after, DB state, E2E).
- `references/cycle.md` — phase-by-phase detail and examples (test naming, AAA structure, exact-error-path rule).
- `references/investigation.md` — `git bisect` automation, 5 Whys, debugging journal, bug clustering.
- `references/checklist.md` — full bug-fix checklist + commit format + PR description rules + edge cases (multiple bugs, untested code, integration-only, regressions).
- `developing-features` skill — code quality / SOLID / file structure / security / observability rules that apply during the fix.
- `developing-features-tdd` skill — broader TDD reference; bugs are TDD's narrow case.
- `ensuring-cross-platform` skill — required when the bug is platform-suspect.
- `learnings/` — project-specific bug patterns, hot zones, recurring traps accumulated over time.

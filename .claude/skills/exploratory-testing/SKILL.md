---
name: exploratory-testing
description: Run structured exploratory testing — write a charter, time-box a session, apply heuristics and oracles to discover bugs the scripted suite never looks for, then debrief and file solid bug reports. Use when the user says "test this manually", "find bugs", "explore the feature", "edge cases", "risk-based testing", "session-based", or asks for a charter, a test session, or a bug hunt on a flow that already "works". Framework- and stack-agnostic. Not for writing automated tests — use writing-tests for unit/integration/E2E coverage and regression checks.
license: MIT
---

# Exploratory Testing

Exploratory testing is simultaneous **learning, test design, and execution** — you study the product, decide what to try next based on what you just saw, and run the test, all in one tight loop. It is *structured, not random*. The structure comes from a charter that bounds the search, heuristics that generate ideas, and oracles that tell you when something is wrong.

The goal is to find the bugs nobody wrote a test for: the unknown unknowns. A passing scripted suite proves the known behaviors still hold; it says nothing about the behaviors no one thought to check. That gap is your job.

## Read first (always)

List `learnings/` and read every file relevant to the current task. Product risk areas, known fragile flows, environment quirks, and "this is always broken under condition X" notes live there and sharpen the charter before you start. If a learning conflicts with this file, **the learning wins** — mention it to the user.

If `learnings/` is empty except for its README, proceed with the defaults below.

## The loop

1. **Charter.** Frame the session with Elisabeth Hendrickson's formula (["Explore It!"](https://pragprog.com/titles/ehxta/explore-it/)):

   > **Explore (target) with (resources) to discover (information).**

   Scope it Goldilocks — not too loose, not too tight:
   - Too vague: *"Test checkout."* The session drifts straight to the happy path and finds nothing.
   - Too tight: *"Click add-to-cart, then click pay, then assert total = $10."* That's a script; it kills the agency that makes exploration work.
   - Just right: *"Explore the **payment flow** under **slow-network conditions** to discover **state-management failures between discount application and address confirmation**."* It names a target, a resource that creates pressure, and the class of information you're hunting.

2. **Time-box a session.** 45–90 minutes, uninterrupted. Take reviewable notes as you go (what you tried, what you saw, what surprised you). Notes are the deliverable when no bug is found — they record coverage. See `references/sessions-and-charters.md`.

3. **Apply heuristics + oracles.** Heuristics generate test ideas (what to try); oracles tell you whether the result is right (how you recognize a problem). Don't improvise from a blank page — run the product through SFDIPOT, vary data deliberately, take a tour. Check every observation against several FEW HICCUPPS oracles. Full catalog in `references/heuristics.md`.

4. **File bugs well.** A bug nobody can reproduce is noise. Concise summary, numbered steps from a known state, expected vs actual, environment, evidence, and severity *and* priority noted separately. Template in `references/bug-reports.md`.

5. **Debrief with PROOF** and branch. After the session, summarize **P**ast (what you did), **R**esults (what you found), **O**bstacles (what got in the way), **O**utlook (what's left to explore), **F**eelings (your read on the risk). Most sessions spawn one or more follow-up charters — file them so the search continues.

## Charter discipline

- One charter = one mission. If you find yourself testing three unrelated things, you have three charters; split and queue the others.
- Resources are the lever. "with slow network", "with a screen reader", "with malformed import files", "with two users editing the same record" — the resource is what turns a flat walkthrough into a probe.
- The information clause keeps you honest. If you can't say what class of problem you're hunting, the charter is too vague to start.

## Heuristics are the engine

You will not find interesting bugs by clicking around. Pull ideas from structured lists and *write down which one you're applying* so coverage is visible:

- **SFDIPOT** — sweep the product's elements: Structure, Function, Data, Interfaces, Platform, Operations, Time.
- **Data variation** — for every input: boundaries (at / just-below / just-above), counts (zero / one / many), and attacks (empty, null, very long, special chars, Unicode, emoji, negative, leading/trailing spaces, injection).
- **Tours** — use them as charter generators: follow the money, follow the data (FedEx), the saboteur tour (break things on purpose), the supermodel tour (UI surface only).

Each is detailed, with sources, in `references/heuristics.md`.

## Oracles decide what counts as a bug

An oracle is a principle by which you recognize a problem. Every one is **fallible** — apply several before you call something a bug. The mnemonic is **FEW HICCUPPS** (Michael Bolton / [DevelopSense](https://developsense.com/blog/2012/07/few-hiccupps)): Familiarity, Explainability, World, History, Image, Comparable products, Claims, User expectations, Product (internal consistency), Purpose, Standards/Statutes. When a behavior contradicts one of these, you likely have a bug — describe the inconsistency, not your guess at the cause.

## Boundary with automated tests

This is the testing-vs-checking distinction (Bach & Bolton, [DevelopSense](https://developsense.com/blog/2009/08/testing-vs-checking)):

- **Checking** is confirming a known fact about the product mechanically — that's what automated tests do, and it's what `writing-tests` covers. Checks guard behavior you already understand against regression.
- **Testing** is exploring to *discover* whether the product has problems you didn't anticipate. That's this skill.

Practical rules:
- Don't burn an exploratory session re-running the scripted happy path — the suite already covers it. Go where the suite isn't.
- When exploration finds a recurring or important bug, feed a new **automated check** back into the suite via `writing-tests` so it never regresses again. Exploration discovers; automation defends.
- If a flow can be fully described as fixed inputs and one expected output, it's a check, not an exploration. Hand it to `writing-tests`.

## Capture a learning (final step)

After the task, ask: *did I hit a product risk area, an environment trap, a flaky flow, or a charter-scoping convention not in this skill?* If yes, append a `learnings/YYYY-MM-DD-slug.md` (frontmatter `description:`, body with the rule + Why + How to apply). A learning overrides the defaults here. If no, skip — don't log a summary of the session you just ran.

## See also

- `references/heuristics.md` — SFDIPOT, CRUSSPIC STMPL, RCRCRC, FEW HICCUPPS, data-variation patterns, Whittaker's tours, with sources.
- `references/sessions-and-charters.md` — SBTM vs Thread-Based TM, charter writing, session-sheet template, PROOF debrief.
- `references/bug-reports.md` — bug report template, severity vs priority, triage basics.
- `writing-tests` skill — the boundary: automated checks that defend known behavior (hand off regressions discovered here).
- `fixing-bugs` skill — once a bug is reproduced, the TDD fix loop (failing test first).
- `reviewing-code` skill — static review of a diff; exploratory testing is the dynamic complement.
- `learnings/` — project-specific risk areas, fragile flows, environment quirks.

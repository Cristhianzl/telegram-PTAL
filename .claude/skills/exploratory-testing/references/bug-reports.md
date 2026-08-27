# Bug reports

Purpose: turn a discovery into an actionable report. The test of a good bug report is that someone else can reproduce, judge, and route the bug without talking to you.

See the [Ministry of Testing](https://www.ministryoftesting.com/dojo/lessons/how-to-write-a-good-bug-report) community guidance for the conventions encoded here.

---

## Anatomy of a great bug report

- **Summary** — one concise line. A reader scanning a list should grasp what's broken and where. "Discount resets to $0 when address is edited after applying a coupon" beats "checkout broken".
- **Repro steps** — numbered, starting from a **known state** ("From a logged-out homepage..."). Each step is a single action. Anyone should be able to follow them blind. Note the reproduction rate if it's intermittent (e.g. "3 of 5 attempts").
- **Expected vs actual** — what *should* happen, then what *did* happen. Describe **observed behavior**, not your theory of the cause. "Total shows $0" not "the discount calculator is dividing by zero" — the guessed cause is often wrong and misleads the fixer. If you can name the oracle violated (see `heuristics.md` FEW HICCUPPS), do: "inconsistent with **Claims**".
- **Environment** — build/version, OS, browser/device, configuration, and the **test data** in play. Many bugs are environment- or data-specific; omitting this is the top cause of "cannot reproduce".
- **Evidence** — screenshots, a short screen recording, console/server logs, network traces, the exact request/response. A 10-second video of an intermittent bug is worth more than a paragraph.
- **Severity and priority** — record **both, separately** (see below). Don't collapse them into one field.

---

## Template (copy-paste)

```markdown
**Summary:** <one line — what's broken, where>

**Environment:**
- Build / version: <id>
- OS / device: <e.g. macOS 15.3 / iPhone 14>
- Browser / app: <e.g. Chrome 122>
- Config / data: <feature flags, account type, test data set>

**Steps to reproduce:** (from <known starting state>)
1.
2.
3.

**Expected:** <what should happen>

**Actual:** <what was observed — behavior, not assumed cause>

**Reproduction rate:** <always | N of M attempts>

**Evidence:** <screenshots / video / logs / network capture>

**Severity:** <critical | major | moderate | minor>  (technical impact)
**Priority:** <P0 | P1 | P2 | P3>  (business urgency)

**Notes / oracle violated:** <e.g. inconsistent with Claims: help text says X>
```

---

## Severity vs priority — they are different axes

- **Severity = technical impact** of the bug if it occurs. How badly does it damage the product when hit? Typically set by the **developer or tester**.
- **Priority = business urgency** of fixing it. How soon must it be addressed relative to other work? Typically set by the **product manager / owner**.

They are independent, and all four combinations exist:

| | High priority | Low priority |
|---|---|---|
| **High severity** | Payment fails for all users — fix now | Data loss only when importing a deprecated format no one uses anymore |
| **Low severity** | Company logo misspelled on the landing page — cosmetic but embarrassing, fix fast | Tooltip text slightly misaligned on a rarely-visited settings page |

Recording both prevents the classic mistakes: deprioritizing a high-severity bug because it's "just a crash in an edge case", or scrambling to fix a typo because someone marked the lone severity field "critical".

---

## Triage basics

When a report lands, before it goes to a developer:

1. **Validate completeness** — can you reproduce it from the steps and environment given? If not, send it back for detail rather than guessing.
2. **Deduplicate** — search existing bugs; link to the original instead of opening a duplicate. Duplicates fragment discussion and inflate counts.
3. **Assign severity and priority** — severity from technical impact, priority from business urgency, set by the right roles.
4. **Route** — assign to the team or owner of the affected area, with enough context that they don't have to re-triage.

A well-formed report makes triage fast; a vague one ("it's broken") burns a triager's time before any fixing starts.

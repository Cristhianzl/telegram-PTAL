# Investigation techniques

Use when the root cause is not obvious from the bug report. Systematic investigation beats trial-and-error.

---

## The 5 Whys

For complex or recurring bugs, dig past the surface symptom until you reach a systemic cause.

```
Bug: "Discount is not applied when cart has exactly 3 items"

Why 1: Why is the discount not applied?
  → The condition in calculate_discount() excludes carts with exactly 3 items.

Why 2: Why does the condition exclude exactly 3 items?
  → It uses `>` instead of `>=`.

Why 3: Why was `>` used instead of `>=`?
  → The original requirement said "more than 3 items" but was later changed to "3 or more".

Why 4: Why wasn't the code updated when the requirement changed?
  → The requirement change was documented in the ticket but the developer only read the original spec.

Why 5: Why did the developer rely on the original spec?
  → There is no process to flag requirement changes on in-progress tickets.

ROOT CAUSE: Process gap — requirement changes are not communicated to developers working on affected tickets.
CODE FIX: Change `>` to `>=` in calculate_discount().
PROCESS FIX: Flag requirement changes on in-progress tickets (separate action item).
```

**Rules for 5 Whys:**

- Stop when you reach something you can **act on** (code fix, process change, or both).
- Don't force exactly 5 iterations — stop at 3 if the root cause is clear, go to 7 if needed.
- Distinguish between the **code fix** (immediate, this PR) and the **process fix** (follow-up ticket).
- Focus on **systems and processes**, not on blaming individuals.

---

## `git bisect` for regressions

If the bug is a regression (it used to work and now doesn't), `git bisect` finds the exact commit that introduced it.

### Manual bisect

```bash
# Start the bisect session
git bisect start

# Mark the current commit as bad (bug exists)
git bisect bad

# Mark a known good commit (bug didn't exist)
git bisect good v1.2.0

# Git checks out a commit in the middle — test it
# If the bug is present:
git bisect bad

# If the bug is NOT present:
git bisect good

# Repeat until git identifies the first bad commit, then:
git bisect reset
```

### Automated bisect with your failing test

Once you have the failing test from Phase 2, you can let git do the work:

```bash
git bisect start HEAD v1.2.0
git bisect run python -m pytest tests/test_cart.py::test_should_apply_discount_when_cart_has_exactly_three_items
```

The test script exits `0` for "good" (test passes) and non-zero for "bad" (test fails). Git automatically finds the exact commit that introduced the bug.

**When to use `git bisect`:**

- The bug is a regression (it used to work).
- You don't know when or why it broke.
- There are many commits between the known good state and the current bad state.
- The root cause is not obvious from the bug report alone.

---

## Debugging journal

For complex bugs, maintain a short log of what you investigated and ruled out. This prevents circular debugging (trying the same thing twice) and helps if someone else picks up the work.

```
BUG-1234: Discount not applied for 3 items
─────────────────────────────────────────
[10:15] Reproduced locally with test — confirmed RED
[10:20] Checked validate_cart() — not the issue, validation passes
[10:35] Checked apply_promotions() — promotion is loaded correctly
[10:42] Found it: calculate_discount() line 47 uses `>` instead of `>=`
[10:45] Applied fix — test is GREEN
[10:50] Full suite passes — no regressions
```

**When to keep a debugging journal:**

- The bug takes more than 30 minutes to diagnose.
- Multiple team members are investigating.
- The bug is intermittent or environment-specific.
- You want to document the investigation in the PR description.

For routine bugs (10-minute fix), skip the journal — it's overhead, not insight.

---

## Bug clustering — finding common root causes

If you keep fixing the same kind of bug, you are treating symptoms, not the disease.

**How to identify a bug cluster:**

- The same file or module appears in multiple bug reports.
- The same type of error (null, boundary, encoding, timezone) keeps recurring.
- The same developer pattern (missing validation, wrong operator, late binding) keeps showing up.

**What to do with a cluster:**

1. Fix each individual bug with its own test (standard TDD cycle).
2. Create a **follow-up ticket** to address the systemic issue:
   - Add a shared validation helper.
   - Enforce stricter types at the boundary.
   - Add a linting rule that detects the pattern.
   - Add a test category (e.g., "all currency operations must round half-up") to the test suite.
3. Document the pattern in the PR so reviewers understand the bigger picture.

The point: **the next instance of the cluster doesn't reach production**.

---

## Anti-patterns

### Shotgun fix — changing multiple things and hoping

```
# Forbidden — no root cause analysis, no understanding
- Changed condition on line 47
- Added null check on line 23
- Modified default value on line 12
- Updated type definition on line 5
- Changed error message on line 89
```

Each change must have a clear, documented reason tied to the root cause. If you don't know which change fixed the bug, you haven't fixed the bug — you've changed something that masked it.

### "Works on my machine"

Reproducing the symptom on Linux when the bug was reported on Windows proves nothing — you may have found a *different* bug that happens to look similar. The test must run on the affected OS in Phase 3.

### Jumping to the code

Reading the bug report, hopping into the editor, and "I'll figure it out as I go." This is how shotgun fixes happen. Investigate first.

### Believing "it's intermittent"

Intermittent bugs are bugs with a missing piece of context — timing, concurrency, environment, input shape. They are not random. Find the missing piece. If you can't reproduce it locally, gather more context from production (Sentry tags, request IDs, log fragments) until you can.

# Bug-fix checklist, commit format, and edge cases

The end-of-fix gate, the commit shape, and the special cases.

---

## Checklist

```
PHASE 1 — UNDERSTAND
[ ] Bug report is clear and complete (GIVEN/WHEN/THEN/EXPECTED)
[ ] Root cause identified (file, function, line, reason)
[ ] 5 Whys applied for complex or recurring bugs
[ ] Scope of impact known
[ ] If regression: git bisect used to find the offending commit
[ ] Existing codebase patterns searched for prior art

PHASE 2 — REPRODUCE
[ ] A new test was written BEFORE any code change
[ ] Test name follows should_[expected]_when_[condition]
[ ] Test reproduces the EXACT scenario from the bug report
[ ] Test exercises the EXACT error path (not a similar one)
[ ] Test asserts the EXPECTED (correct) behavior
[ ] Test follows Arrange-Act-Assert structure
[ ] For untested code: characterization tests written FIRST as safety net

PHASE 3 — VERIFY RED
[ ] Test FAILS before the fix
[ ] Failure reason is CONSISTENT with the bug report
[ ] RED state documented (error message + match confirmation)

PHASE 4 — FIX
[ ] Root cause documented (WHERE, WHY, WHAT)
[ ] Fix follows existing codebase patterns for consistency
[ ] Fix changes the MINIMUM amount of code
[ ] No dead code — every line justified by RED → GREEN transition
[ ] No unrelated refactoring in the fix
[ ] No new features in the fix
[ ] Pre-existing issues handled correctly (fix in place, separate commit, or follow-up ticket)

PHASE 5 — VERIFY GREEN
[ ] Previously failing test now PASSES
[ ] GREEN state documented

PHASE 6 — VALIDATE
[ ] ALL existing tests still pass
[ ] Characterization tests still pass (if written)
[ ] No new warnings
[ ] Linter, formatter, and type checker pass
[ ] If platform-suspect: test passes on every supported OS in CI
[ ] Any test updates are justified and documented

QUALITY & PREVENTION
[ ] All developing-features standards met (SOLID, file structure, security)
[ ] No PII in logs
[ ] No secrets in code
[ ] Strong typing maintained
[ ] New log statements have correlation context (IDs, not just event names)
[ ] New log event names are unique — no collisions with existing events
[ ] New log statements follow the codebase's logging convention
[ ] Bug-fix test added to CI/CD as permanent regression guard
[ ] Bug cluster pattern checked — is this a recurring issue type?
[ ] Debugging journal included in PR description (for complex bugs)
[ ] PR description is accurate — every claim matches the diff
[ ] Pre-existing issues NOT fixed are documented as follow-up tickets
```

---

## Commit format

```
fix: <short description of what was fixed>

Bug: <bug ticket ID or reference>
Root Cause: <one-line explanation of why it happened>
Fix: <one-line explanation of what was changed>
Test: <name of the test that proves the fix>
```

Example:

```
fix: apply discount when cart has exactly 3 items

Bug: BUG-1234
Root Cause: off-by-one in discount condition (> instead of >=)
Fix: changed item_count > 3 to item_count >= 3 in calculate_discount()
Test: test_should_apply_discount_when_cart_has_exactly_three_items
```

For complex bugs, include the debugging journal in the PR description, not the commit.

### PR description accuracy

Every claim in the PR description must be verifiable in the diff. The description is a contract with the reviewer.

- List only what **you** changed — don't claim credit for pre-existing code.
- Be precise — "changed `>` to `>=` on line 47" is better than "fixed the condition".
- Separate your fix from pre-existing fixes — if you also fixed a pre-existing issue, label it clearly.
- Don't overstate the scope — if you added a `try/except`, don't say "redesigned the error handling".
- Include what you did **not** change and why — "Did not refactor the DRY violation in `post_execute`/`on_error` — created follow-up ticket TECH-890."

---

## Edge case: multiple bugs in one report

If a report contains multiple distinct bugs:

1. Separate them — each bug gets its own test and its own fix.
2. Fix one at a time — complete the full RED→GREEN cycle for each.
3. Do NOT batch fixes — one bug, one test, one fix, one commit.
4. Order by dependency — if Bug B only appears after Bug A is fixed, fix A first.

---

## Edge case: bug in untested code (characterization tests)

If the buggy code has no existing tests:

1. **Write characterization tests first** — capture the **current** behavior of the surrounding code before touching anything. These tests document what the code actually does (even if some of it is buggy) and protect you from accidentally breaking something else during the fix.
2. Write the bug-fix test (Phase 2 — the RED test that asserts the **expected** behavior).
3. Fix the bug (Phase 4).
4. Verify GREEN (Phase 5) — both the bug-fix test AND the characterization tests must pass.
5. Consider adding additional tests for the now-fixed function (success cases, other edges, error cases) in a **separate commit** from the bug fix.

**Characterization test rules:**

- They assert what the code **currently does**, not what it should do.
- They are a safety net, not a validation of correctness.
- If a characterization test starts failing during your fix, investigate — you may be introducing a regression.
- Do NOT fix other bugs you discover while writing characterization tests — create separate tickets.

```python
# Characterization test — documents current behavior
def test_characterize_calculate_discount_with_five_items():
    # This captures the existing (correct) behavior so we know if the fix
    # accidentally breaks other scenarios
    cart = Cart(items=[item1, item2, item3, item4, item5])
    assert cart.discount == 4.99  # currently works — must stay working

# Bug-fix test — asserts the EXPECTED (correct) behavior
def test_should_apply_discount_when_cart_has_exactly_three_items():
    cart = Cart(items=[item1, item2, item3])
    assert cart.discount == 2.99  # currently FAILS — this is the bug
```

---

## Edge case: bug cannot be unit-tested

If the bug requires integration or end-to-end coverage:

1. Still write the test first — even if it's an integration test.
2. Document why a unit test is insufficient (e.g., "bug occurs in DB transaction boundary").
3. Keep the test as focused as possible — minimize the integration surface.
4. Mock what you can; integrate only what you must.
5. The RED → GREEN cycle still applies regardless of test type.

---

## Edge case: regression — using `git bisect`

See `investigation.md` for the full guide. Quick recipe:

```bash
git bisect start HEAD v1.2.0
git bisect run python -m pytest tests/test_cart.py::test_should_apply_discount_when_cart_has_exactly_three_items
```

Use when:

- The bug is a regression (it used to work).
- You don't know which commit broke it.
- Many commits stand between the known good state and the current bad state.
- The root cause is not obvious from the report.

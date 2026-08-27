# The bug-fix cycle — phase-by-phase detail

Companion to `SKILL.md`. Read this when you need the longer explanation of a phase, especially Phase 2's "exact error path" rule and the platform-specific dimension.

---

## Phase 1 — UNDERSTAND

Don't open the editor before completing this phase.

Answers you must produce:

| Question                                          | Why                                                  |
|---------------------------------------------------|------------------------------------------------------|
| What is the expected behavior?                    | Defines what "fixed" means                            |
| What is the actual (buggy) behavior?              | Defines what the test must reproduce                  |
| What are the exact steps to reproduce?            | Defines the test scenario (Arrange + Act)             |
| What error, wrong value, or side effect occurs?   | Defines the test assertion (Assert)                   |
| What is the scope of impact?                      | Defines if more tests are needed                      |
| Which OS does the report describe?                | Defines the platform dimension of the test            |

Bug report decomposition:

```
GIVEN:   <precondition / initial state>
WHEN:    <action that triggers the bug>
THEN:    <what actually happens — the bug>
EXPECTED:<what should happen instead — the fix target>
```

If you cannot extract all four → **ask for clarification, do not guess.**

### Platform dimension

Bug reports rarely state the OS. Check the user agent, Sentry tags, container image, support ticket metadata. Bugs in these areas are platform-suspect by default:

- Filesystem operations (paths, permissions, atomic moves, symlinks, case sensitivity).
- Text encoding or line endings.
- Subprocess / shell invocation.
- Temp file / config dir resolution.
- Process signals, fork vs spawn.
- Networking (localhost resolution, port allocation).

If the bug is platform-specific, the test must run on the affected OS in Phase 3 — see `ensuring-cross-platform`.

If a platform-specific bug shipped to production, the test suite did not run on that OS. That is a separate finding: extending the CI matrix is part of **this** fix, not a follow-up.

---

## Phase 2 — REPRODUCE — the exact-error-path rule

> The most common mistake in bug-fix tests is testing a **similar** error path instead of the **exact** one.

If the bug report says "crash caused by `GeneratorExit` during streaming", your test MUST trigger a `GeneratorExit`. Testing `CancelledError` (a similar but different exception) leaves the primary bug path unproven, and the PR may pass code review with "tests added ✓" even though the actual bug is uncovered.

**Wrong — tests a similar error, not the exact one from the bug report:**
```python
# Bug: "Sentry Error: GeneratorExit crashes the streaming endpoint"
async def test_should_handle_cancelled_error_during_streaming():
    gen = stream_conversations(request)
    async for chunk in gen:
        break
    gen.throw(asyncio.CancelledError)  # Not what Sentry reported
```

**Right — tests the exact error path from the bug report:**
```python
async def test_should_handle_generator_exit_during_streaming():
    gen = stream_conversations(request)
    async for chunk in gen:
        break
    await gen.aclose()  # Sends GeneratorExit — the actual production error
```

**How to identify the exact error path:**

- Read the stack trace from production (Sentry, CloudWatch, etc.).
- Note the exact exception type, not a parent class or sibling.
- Note the exact function and line where it occurred.
- Your test must trigger that same exception through that same code path.

---

### Test naming convention

```
should_[expected_behavior]_when_[condition_that_triggered_bug]
```

Examples:

```python
# Bug: "User receives 500 when email has uppercase characters"
def test_should_validate_email_successfully_when_email_has_uppercase_characters():

# Bug: "Discount is not applied when cart has exactly 3 items"
def test_should_apply_discount_when_cart_has_exactly_three_items():

# Bug: "API returns null instead of empty list when user has no orders"
def test_should_return_empty_list_when_user_has_no_orders():
```

```typescript
// Bug: "Payment fails silently when currency is EUR"
it('should_process_payment_successfully_when_currency_is_EUR', () => {});

// Bug: "Date filter ignores timezone offset"
it('should_filter_by_date_respecting_timezone_offset', () => {});
```

---

### Arrange / Act / Assert

```python
def test_should_[expected]_when_[condition]():
    # Arrange — set up the EXACT scenario from the bug report
    input_data = create_scenario_from_bug_report()

    # Act — execute the action that triggers the bug
    result = execute_buggy_operation(input_data)

    # Assert — verify the EXPECTED (correct) behavior, not the buggy behavior
    assert result == expected_correct_value
    # This assertion MUST FAIL before the fix.
    # This assertion MUST PASS after the fix.
```

```typescript
it('should_[expected]_when_[condition]', () => {
    // Arrange — set up the EXACT scenario
    const inputData = createScenarioFromBugReport();

    // Act
    const result = executeBuggyOperation(inputData);

    // Assert — expected behavior, not buggy
    expect(result).toEqual(expectedCorrectValue);
});
```

---

### What the test must NOT do

- Assert the buggy behavior (the test would pass before the fix — useless).
- Test a similar error path instead of the exact one (see above).
- Test implementation details instead of observable behavior.
- Depend on other tests or shared mutable state.
- Be flaky or non-deterministic.
- Cover unrelated scenarios — one bug, one focused test.

---

## Phase 3 — VERIFY RED

Run the test. It MUST fail.

| Check                                                 | Expected                                          |
|-------------------------------------------------------|---------------------------------------------------|
| Does the test fail?                                   | Yes                                                |
| Is the failure reason consistent with the bug?        | Yes — error matches the report                     |
| Is the assertion checking the right thing?            | Yes — expected vs. actual makes sense              |

If the test passes before the fix:

1. The bug isn't reproducible with your scenario → revisit Phase 1.
2. The assertions are wrong → revisit Phase 2.
3. The bug is already fixed → confirm against the latest code.

If the test fails for a different reason than the bug:

1. Setup error → fix Arrange.
2. The bug has a different root cause → revisit Phase 1.
3. Mocks/fakes are wrong → check them.

---

## Phase 4 — FIX — minimum change

Fix scope rules:

| Allowed                                                              | Not allowed                                            |
|----------------------------------------------------------------------|--------------------------------------------------------|
| Fix the root cause of the reported bug                               | Refactor the entire module                             |
| Add a guard clause for the edge case                                 | Rewrite the function "for clarity"                     |
| Correct a conditional logic error                                    | Change the API contract                                |
| Fix an off-by-one error                                              | Add new parameters or endpoints                        |
| Handle a missing null check                                          | Reorganize the file structure                          |
| Fix a pre-existing issue in the lines you're touching (note in PR)   | Fix pre-existing issues in unrelated areas             |
| Rename a colliding log event name your fix interacts with            | Rename all log events across the module                |

Document the root cause:

```
ROOT CAUSE:
  File: cart_service.py, line 47
  Function: calculate_discount()
  Issue: condition uses `>` instead of `>=`, excluding carts with exactly 3 items
  Fix: change `if item_count > 3` to `if item_count >= 3`
```

---

### Cross-platform fix verification

Common platform-fix mistakes (catch in Phase 6):

| Mistake                                                                 | Detection                                              |
|-------------------------------------------------------------------------|--------------------------------------------------------|
| Adding `if sys.platform == "win32"` without testing the other branches  | All branches must have test coverage                   |
| Replacing `/` with `os.sep` in one place but not another                | Grep the diff for residual `/` literals in path contexts |
| "Fixing" line endings by force-writing `\n` in binary mode              | Verify file opens correctly on Windows readers          |
| Catching a Windows-specific exception silently                          | Re-raise with context; never swallow OS-specific errors |

---

## Phase 5 — VERIFY GREEN

Run the same test. It MUST now pass.

| Check                                                                  | Expected                          |
|------------------------------------------------------------------------|-----------------------------------|
| Does the previously failing test now pass?                             | Yes                                |
| Did the fix change the minimum amount of code?                         | Yes — no scope creep               |
| Does the test assertion match the expected behavior from the report?   | Yes — exact match                  |

If the test still fails, the fix is wrong or incomplete. Don't proceed.

---

## Phase 6 — VALIDATE

Full suite + CI matrix.

If existing tests break:

1. **Analyze** — is the breaking test correct, or was it relying on buggy behavior?
2. **If the test was wrong** → fix the test (it was asserting the buggy behavior as correct). Document why in the PR.
3. **If the test was right** → your fix has a side effect. Return to Phase 4.

Never delete a failing test to "make CI green".

---

## Phase 7 — REFACTOR (optional)

Refactoring rules:

- Small, incremental changes — one transformation at a time.
- Run tests after each change.
- No behavior changes — observable behavior must not alter.
- Separate commit from the fix.

# Test quality

How to name, structure, and judge a test. Applies to every test written during a RED→GREEN slice.

---

## Naming convention

```
should_[expected_behavior]_when_[condition]
```

Examples:

```python
# Happy path
def test_should_create_user_when_input_is_valid():

# Edge case
def test_should_return_empty_list_when_user_has_no_orders():

# Error case
def test_should_raise_validation_error_when_email_is_malformed():

# Adversarial
def test_should_reject_request_when_signature_is_invalid():
def test_should_deny_access_when_user_is_not_owner_of_resource():
```

```typescript
// Happy path
it('should_create_user_when_input_is_valid', () => {

// Adversarial
it('should_reject_payload_when_hmac_signature_does_not_match', () => {
```

The test name **encodes the acceptance criterion**. A future reader of the test file should know what's being verified without reading the body.

---

## Structure — Arrange / Act / Assert

```python
def test_should_persist_user_when_input_is_valid():
    # Arrange — set up the EXACT scenario for this slice
    repository = FakeUserRepository()
    service = UserService(repository=repository)
    request = CreateUserRequest(email="user@example.com", name="Alice")

    # Act — execute the behavior under test
    result = service.create_user(request)

    # Assert — verify observable behavior (return value, side effect, error)
    assert result.id is not None
    assert repository.find_by_id(result.id).email == "user@example.com"
```

```typescript
it('should_persist_user_when_input_is_valid', () => {
    // Arrange
    const repo = new FakeUserRepository();
    const service = new UserService(repo);
    const req = { email: 'user@example.com', name: 'Alice' };

    // Act
    const result = service.createUser(req);

    // Assert
    expect(result.id).toBeDefined();
    expect(repo.findById(result.id).email).toBe('user@example.com');
});
```

One Act per test. If you find yourself writing two Acts, you have two tests.

---

## What to test (every slice)

A complete slice's tests cover all five categories:

1. **Happy path** — the smallest valuable behavior in normal conditions.
2. **Edge cases** — empty inputs, null, boundary values (0, 1, MAX, MIN), maximum allowed sizes, off-by-one.
3. **Error cases** — invalid input, downstream service failure, timeout, conflict.
4. **Adversarial cases** — forged input, replay, injection patterns, authorization bypass, scope escape.
5. **State cases** — before/after for any state transition; idempotency; concurrent invocation if applicable.

Happy-path tests alone are insufficient. They prove the code does **a** right thing, not **the** right thing.

---

## Tests must challenge the code, not only confirm it

- Write tests based on **requirements / spec**, not on what the source code currently does.
- When a test fails: **first ask if the CODE is wrong**, not the test. Never silently change an assertion to match buggy code.
- Include tests that actively try to break the code: boundary values, malformed input, concurrent scenarios, failure of every external dependency, authorization violations.

---

## What tests must NOT do

- Assert implementation details (private methods, internal call order) instead of observable behavior.
- Depend on other tests, run order, or shared mutable fixtures.
- Hit real external services (DB, HTTP, filesystem) in unit tests — use fakes or mocks.
- Use real time / random — inject a clock / RNG and control them.
- Be flaky — if a test is flaky, it's **broken**; fix the test, never retry-loop the failure away.
- Cover unrelated scenarios — one slice, one focused test.

---

## Anti-patterns

### Test that passes before the code exists

```python
# Useless — assertion is true regardless of implementation
def test_create_user():
    service = UserService(repo=FakeRepo())
    assert service is not None  # trivially true
```

Fix: assert observable behavior. Construct a scenario where the absence of the feature makes the assertion impossible.

### Tests that only confirm the happy path

```python
# Incomplete — only proves the easy case
def test_create_user_happy():
    service.create_user(valid_request)
```

Fix: add at least one edge case, one error case, one adversarial case per slice.

### Tests asserting implementation details

```python
# Couples test to internal structure — refactoring will break it
def test_create_user_calls_validator_then_repo():
    spy = MockValidator()
    service.create_user(req)
    assert spy.called_before(repo_spy)
```

Fix: assert observable behavior.

```python
def test_should_persist_user_when_input_is_valid():
    service.create_user(req)
    assert repo.find_by_email(req.email) is not None
```

### Tests with hidden coupling

```python
# Order-dependent — test_b passes only if test_a ran first
class TestUserService:
    def test_a_create(self):
        self.service.create("alice")  # mutates shared state
    def test_b_find(self):
        assert self.service.find("alice") is not None  # depends on test_a
```

Fix: each test creates its own state. Use fixtures with explicit setup or fresh fakes per test.

### Mockery test (testing the mock, not the code)

```python
# All this proves is "the mock returned what you told it to return"
def test_find_user():
    mock_repo = Mock()
    mock_repo.find_by_id.return_value = User(id="u1", name="Alice")
    user = mock_repo.find_by_id("u1")
    assert user.name == "Alice"
```

Fix: the System Under Test (SUT) must execute real code that does something with the mock's return value. If the test never invokes SUT code, delete it.

### Snowball test

A test that calls a chain of methods, each producing input for the next. When it fails, you don't know which step broke. Symptom: the test has 30+ lines of setup.

Fix: split into multiple tests, each isolating one decision.

### Liar test

A test that passes but doesn't actually exercise the behavior named in its title.

```python
def test_should_charge_card():
    service = PaymentService(processor=NoOpProcessor())  # always returns success
    result = service.charge(...)
    assert result.success  # passes trivially because the processor is a no-op
```

Fix: the fake must be capable of representing the failure mode the test is meant to detect.

---

## Edge case: feature touches untested legacy code

If the feature must extend code that has no existing tests:

1. **Write characterization tests first** — capture the current behavior as a safety net. Do NOT assert what it *should* do; assert what it *does*.
2. Confirm characterization tests are GREEN against the current code.
3. Then proceed with the normal RED→GREEN cycle for the new feature.
4. If a characterization test starts failing during your work → you've introduced a regression. Investigate before proceeding.
5. Do NOT fix unrelated bugs you discover while writing characterization tests — log them as follow-ups.

---

## Edge case: feature cannot be unit-tested

If a slice genuinely requires integration / E2E coverage (DB transaction boundary, real browser interaction):

1. Still write the test first.
2. Document in the test file why a unit test is insufficient.
3. Keep the test as focused as possible — minimize the integration surface.
4. Mock what you can; integrate only what you must.
5. The RED → GREEN cycle still applies.

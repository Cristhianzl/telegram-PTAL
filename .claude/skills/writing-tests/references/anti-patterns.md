# Test anti-patterns

Eight named test smells. Every one of them shows up in real code. Each entry shows the symptom, the wrong example, and the correction.

---

## Mirror (Confirmation Bias)

A test that reads the source code and asserts exactly what the code does. Finds zero bugs.

```python
# Wrong — copied the implementation into the assertion
# Source: get_status() returns "active" when enabled=True
def test_status():
    user = create_user(enabled=True)
    assert user.get_status() == "active"
```

Right: assert from the **spec**, not the code. Test what the code *should* do for other inputs.

```python
def test_should_return_inactive_when_account_is_expired():
    user = create_user(enabled=True, expires_at=datetime(2020, 1, 1))
    assert user.get_status() == "inactive"

def test_should_return_pending_when_email_unverified():
    user = create_user(enabled=True, email_verified=False)
    assert user.get_status() == "pending"
```

---

## Liar

A test that passes but doesn't verify the behavior its name claims.

```python
# Wrong — passes regardless of whether the user was actually created
def test_create_user():
    user = create_user(name="Alice")
    assert user is not None
```

Right: assert the **observable** behavior the test name implies.

```python
def test_should_persist_user_when_input_is_valid():
    create_user(name="Alice")
    assert repository.find_by_name("Alice") is not None
```

---

## Giant

50+ lines of setup, multiple Acts, dozens of unrelated assertions.

```python
# Wrong — too much in one test
def test_entire_checkout_flow():
    # 30 lines of setup
    # create user, products, cart, apply coupon, calc tax,
    # process payment, send email, update inventory…
    assert result.status == "completed"
    assert result.total == 42.50
    assert len(result.items) == 3
    assert email_sent == True
    assert inventory_updated == True
```

Right: 5+ separate tests, each with one Act and one focused assertion group.

---

## Mockery

So many mocks that the test only exercises the mock setup.

```python
# Wrong — mocking everything; what is even tested?
def test_over_mocked(mocker):
    mocker.patch("app.service_a")
    mocker.patch("app.service_b")
    mocker.patch("app.service_c")
    mocker.patch("app.repository")
    mocker.patch("app.validator")
    result = do_something()
    assert result == "expected"
```

Right: if the function under test is just wiring, **delete the unit test** and write an integration test that exercises real components. Or refactor the production code to have fewer collaborators.

---

## Inspector

A test that knows too much about internal structure and breaks on every refactor.

```python
# Wrong — coupled to implementation details
def test_inspector(mocker):
    mock_repo = mocker.patch("app.repository.save")
    create_user(name="Alice")
    mock_repo.assert_called_once_with(User(name="Alice", id=ANY, created_at=ANY))
```

Right: assert the observable outcome.

```python
def test_should_persist_user_when_input_is_valid():
    create_user(name="Alice")
    assert repository.find_by_name("Alice").name == "Alice"
```

Exception: when the **call itself** is the behavior being tested (e.g., "should send a welcome email"), it's fine to assert the call.

```python
def test_should_send_welcome_email_when_user_registers(mocker):
    send_email = mocker.patch("app.notifier.send_email")
    register_user(email="alice@example.com")
    send_email.assert_called_once_with(to="alice@example.com", template="welcome")
```

---

## Chain Gang

Tests depend on each other's execution order or shared mutable state.

```python
# Wrong — test_b only passes if test_a ran first
class TestUserService:
    def test_a_create(self):
        global shared_user
        shared_user = create_user(name="Alice")

    def test_b_update(self):
        update_user(shared_user, name="Bob")
```

Right: each test creates its own state. Use fixtures with explicit setup, or fresh fakes per test.

```python
def test_should_create_user_when_input_is_valid():
    user = create_user(name="Alice")
    assert user.name == "Alice"

def test_should_update_user_name_when_called_with_new_name():
    user = create_user(name="Alice")
    update_user(user, name="Bob")
    assert user.name == "Bob"
```

---

## Flaky

A test that passes sometimes, fails sometimes, with no code changes.

```python
# Wrong — depends on machine speed
def test_flaky_time():
    token = generate_token()
    time.sleep(0.001)
    assert token.is_valid()
```

Right: inject a clock, control time deterministically.

```python
def test_should_consider_token_expired_when_clock_advances_past_ttl():
    clock = FakeClock(now=datetime(2026, 1, 1, 12, 0, 0))
    token = generate_token(clock=clock, ttl_seconds=60)
    clock.advance(seconds=61)
    assert not token.is_valid()
```

A flaky test is **broken**. Fix it immediately or quarantine with a ticket reference. Never retry-loop a flake away.

---

## Snowball

A giant snapshot or serialization test that breaks on every minor change.

```python
# Wrong — couples to every field, breaks on any addition
def test_snapshot():
    result = get_user_response(user_id="123")
    assert result == {
        "id": "123",
        "name": "Alice",
        "email": "alice@example.com",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z",
        "settings": {"theme": "dark", "language": "en"},
        # 20 more fields
    }
```

Right: assert only the fields the test claims to verify.

```python
def test_should_return_user_with_settings_when_user_exists():
    result = get_user_response(user_id="123")
    assert result["id"] == "123"
    assert result["name"] == "Alice"
    assert result["settings"]["theme"] == "dark"
```

When a full-shape snapshot is genuinely useful (e.g., API contract test), use a contract-testing tool (`pact`, `schemathesis`) rather than a freeform equality check.

---

## Adversarial test examples (good practice)

For comparison, what **good** adversarial tests look like:

```python
# Boundary
def test_should_apply_exact_cap_at_boundary():
    # 20% of 250 = 50 (exactly at the $50 cap)
    result = calculate_discount(user=premium_user, price=250)
    assert result == 50.0

# Cap enforcement
def test_should_cap_premium_discount_at_50():
    # 20% of 500 = 100, but spec caps at $50
    result = calculate_discount(user=premium_user, price=500)
    assert result == 50.0  # if code forgot the cap, this catches the bug

# Negative
def test_should_reject_negative_price():
    with pytest.raises(ValueError, match="price must be positive"):
        calculate_discount(user=premium_user, price=-100)
```

```typescript
// What should NOT happen
it('should not render delete button for read-only users', () => {
  render(<Header user={readOnlyUser} />);
  expect(screen.queryByRole('button', { name: /delete/i })).not.toBeInTheDocument();
});

// Boundary
it('should truncate title to 50 characters with ellipsis', () => {
  const longTitle = 'A'.repeat(60);
  render(<Header title={longTitle} />);
  expect(screen.getByText('A'.repeat(50) + '...')).toBeInTheDocument();
});

// Doesn't crash on empty
it('should handle empty title without crashing', () => {
  expect(() => render(<Header title="" />)).not.toThrow();
});
```

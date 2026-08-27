# Test file structure

How to lay out test files, what to name them, where to put fixtures and helpers.

---

## Mirror the source structure

```
src/
├── users/
│   ├── user_service.py
│   ├── user_repository.py
│   └── helpers/
│       └── validation.py
tests/
├── users/
│   ├── test_user_service.py
│   ├── test_user_repository.py
│   └── helpers/
│       └── test_validation.py
```

In JS/TS, **colocated** tests are also fine:

```
src/
├── users/
│   ├── user.service.ts
│   ├── user.service.test.ts
│   ├── user.repository.ts
│   └── user.repository.test.ts
```

What is NOT fine:

```
tests/
├── test_everything.py
├── test_misc.py
├── test_utils.py
```

A flat, unorganized test directory is a smell. So is a catch-all `test_utils.py`.

---

## File naming convention

| Language          | Source file          | Test file                                                     |
|-------------------|----------------------|----------------------------------------------------------------|
| Python            | `user_service.py`    | `test_user_service.py`                                         |
| JS / TS           | `user.service.ts`    | `user.service.test.ts` or `user.service.spec.ts`               |
| Java              | `UserService.java`   | `UserServiceTest.java`                                         |
| Go                | `user_service.go`    | `user_service_test.go`                                         |
| C#                | `UserService.cs`     | `UserServiceTests.cs`                                          |
| Rust              | `user_service.rs`    | `mod tests {}` inside the file, or `tests/user_service.rs`     |
| Ruby (RSpec)      | `user_service.rb`    | `user_service_spec.rb`                                         |
| Ruby (Minitest)   | `user_service.rb`    | `test_user_service.rb`                                         |

**Always follow the project's existing convention.** If none exists, use the table above.

---

## File size limits

| Metric                                 | Guideline                                                                                                  |
|----------------------------------------|------------------------------------------------------------------------------------------------------------|
| Lines of test code per file            | ~1000 lines guideline — above this, consider splitting; **not required** if the file covers one module     |
| Test cases per file                    | No hard limit                                                                                              |
| describe/context blocks per file       | No hard limit                                                                                              |
| Setup complexity (Arrange phase)       | ≤ 20 lines per test — extract to helpers/factories if exceeded                                              |

Split based on **logical separation**, not arbitrary line counts. A 1200-line file thoroughly covering one complex module is better than 5 small files that fragment related tests.

The only hard rule: **one file = one subject (one module/service)**.

### When to split

Only split when:

- The file covers **unrelated** modules or services.
- The file mixes unit and integration tests.
- The file exceeds ~1000 lines AND the behaviors are logically separable.

```
# Good — one file covers one module thoroughly
test_user_service.py     (40 tests, 800 lines — all about UserService)

# Also good — split by logical area only when it makes sense
test_user_service_creation.py        (create scenarios)
test_user_service_authentication.py  (login, logout, session)

# Bad — over-split into tiny files for no reason
test_user_service_create.py   (3 tests)
test_user_service_update.py   (2 tests)
test_user_service_delete.py   (2 tests)
test_user_service_get.py      (2 tests)
```

---

## Shared test infrastructure

### Factories / builders

Use factory functions to create test data. Never hardcode complex objects inline across many tests.

**Python:**
```python
# tests/factories/user_factory.py
def create_user(**overrides):
    defaults = {
        "id": "user-123",
        "name": "Alice",
        "email": "alice@example.com",
        "tier": "standard",
        "is_active": True,
    }
    return User(**{**defaults, **overrides})

# Usage
def test_should_apply_discount_for_premium_users():
    user = create_user(tier="premium")
    # ...
```

**TypeScript:**
```typescript
// tests/factories/userFactory.ts
export const createUser = (overrides: Partial<User> = {}): User => ({
  id: 'user-123',
  name: 'Alice',
  email: 'alice@example.com',
  tier: 'standard',
  isActive: true,
  ...overrides,
});

// Usage
it('should apply discount for premium users', () => {
  const user = createUser({ tier: 'premium' });
  // ...
});
```

**Rules for factories:**

- Provide sensible defaults for **all** required fields.
- Allow overriding any field.
- Produce **valid** objects by default.
- Do NOT contain business logic.
- Do NOT make I/O calls.

### Shared fixtures / setup

| Scope        | Where to define                                                | When to use                              |
|--------------|----------------------------------------------------------------|------------------------------------------|
| Single test  | Inline in the test (Arrange phase)                              | Simple, one-off setup                    |
| Single file  | `beforeEach` / `setUp` in the same file                          | Shared across tests in one file          |
| Single module| Shared helper file in the same test directory                   | Shared across files in one module        |
| Global       | Root `conftest.py`, `jest.setup.ts`, or equivalent              | Truly global (DB connection, env vars)   |

**Rule:** keep fixtures as close to the tests as possible. Promote to a wider scope only when reuse is real, not hypothetical.

---

## Directory layout (recommended)

```
tests/                              # or __tests__/, spec/, test/
├── unit/                           # fast, isolated tests
│   ├── users/
│   │   ├── test_user_service.py
│   │   └── helpers/
│   │       └── test_validation.py
│   └── orders/
│       └── test_order_service.py
├── integration/                    # component interaction tests
│   ├── test_user_repository.py
│   └── test_payment_gateway.py
├── e2e/                            # end-to-end tests (if applicable)
│   └── test_checkout_flow.py
├── factories/                      # shared test data factories
│   ├── user_factory.py
│   └── order_factory.py
├── fixtures/                       # shared test fixtures and data
│   ├── sample_responses.json
│   └── seed_data.sql
└── helpers/                        # shared test utilities
    ├── assertions.py               # custom assertion helpers
    └── mock_builders.py            # reusable mock setup
```

If the project uses a different structure, follow that.

---

## DRY in tests — with nuance

DRY applies, but with a caveat: **prefer readability over DRY in tests**. A test should be fully understandable by reading it top to bottom, without jumping to five different files.

**Do** extract when:

- Same object creation appears in 3+ tests → factory function.
- Same identical preconditions across 5+ tests in a file → `setUp` / `beforeEach`.
- Same multi-step assertion appears in 3+ tests → custom assertion helper.
- Same mock setup appears in 3+ tests → setup helper.

**Do NOT** extract when:

- The Arrange phase is 2–3 lines and inline reads better.
- Extracting hides what's being verified.
- Each test needs slightly different setup — duplication beats a parameterized helper monster.

```python
# Bad — over-DRY, test is unreadable without opening helpers
def test_should_reject_expired_coupon():
    order = setup_standard_order()          # what's in this order?
    coupon = setup_expired_coupon()          # when did it expire?
    assert_coupon_rejection(order, coupon)   # what exactly is asserted?

# Good — some duplication, but the test tells a complete story
def test_should_reject_expired_coupon():
    # Arrange
    order = create_order(total=100.00)
    coupon = create_coupon(code="SAVE10", expires_at=datetime(2023, 1, 1))

    # Act
    result = apply_coupon(order, coupon)

    # Assert
    assert result.is_error
    assert result.error_code == "COUPON_EXPIRED"
```

**Rule of thumb:** extract repeated logic (5+ lines, 3+ occurrences) into named helpers, BUT keep the test's **intent** readable at the call site.

---

## Security in test code

PII and security rules apply to test code too:

- Use **fake / anonymized data**. Generic emails: `user@example.com`, `alice@test.com`. Placeholder tokens: `test-api-key-123`, `fake-token`.
- Never commit real API keys, passwords, or PII in test fixtures.
- Never hardcode production URLs or credentials in test files.
- Never log PII in test setup or teardown.

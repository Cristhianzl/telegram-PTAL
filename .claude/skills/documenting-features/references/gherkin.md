# Gherkin scenarios

Behavior specifications in Given/When/Then form. Every scenario maps to a test in the suite. The Gherkin in Section 4 of the doc is the **executable specification** translated to natural language.

---

## Feature header

```gherkin
Feature: Place an order

  As a customer
  I want to submit a cart with payment details
  So that I receive the items I selected

  Background:
    Given the customer is authenticated
    And the cart has at least one item
```

The header has:

- **Feature** — the name of the capability (matches the bounded context).
- **As a / I want / So that** — the user story. The `So that` is critical: it's the **business value**. A scenario without a `So that` is hard to prioritize.
- **Background** — preconditions shared by every scenario below. Keep it short — if Background grows, the scenarios are probably testing too much.

---

## Scenario structure

```gherkin
Scenario: Place an order with valid items and payment
  Given the cart contains the SKU "PRO-LICENSE" with quantity 1
  And the customer has a valid payment method on file
  When the customer submits the order
  Then the order is created with status "placed"
  And the customer receives an order confirmation email
  And the inventory for "PRO-LICENSE" is decremented by 1
```

Rules:

- **One observable behavior per scenario.** Multiple `Then` lines are OK if they describe the same outcome from different angles. Multiple unrelated `Then`s = two scenarios.
- **One `When` per scenario.** Multiple `When`s = you have two scenarios that should be split.
- **`Given`** sets up the world. **`When`** is the action under test. **`Then`** is the observable outcome.
- **`And`** chains additional steps of the same kind. **`But`** is also valid for negations.

---

## Coverage rules — scenarios per feature

Every feature must have at least:

- **One happy-path scenario** (the primary flow with valid inputs).
- **One edge-case scenario** (boundary value, empty input, near-the-limit).
- **One error-case scenario** (invalid input rejected with the correct error).
- **One authorization scenario** (user without permission is denied).

If the feature is purely internal (no user-facing actor), replace the authorization scenario with a **failure-mode scenario** (downstream dependency unavailable, retried, etc.).

---

## Translating acceptance criteria from TDD

If the feature was built with the `developing-features-tdd` skill, the acceptance criteria already exist in `Given/When/Then/So-That` form (from Phase 1 — UNDERSTAND). Translate them directly into Gherkin — same content, same structure.

```
# From TDD acceptance criteria
GIVEN:    a registered user with valid credentials
WHEN:     they submit a login request with the wrong password
THEN:     the system returns 401 with a generic error message
SO THAT:  attackers cannot enumerate valid usernames via timing or error content
```

becomes:

```gherkin
Scenario: Wrong password returns a generic 401
  Given a registered user with valid credentials
  When they submit a login request with the wrong password
  Then the system returns 401 with a generic error message
  # Rationale (from TDD): prevents username enumeration via timing or error content
```

Put the `So that` in the feature header (it's a feature-level value) and in the scenario comment when it's scenario-specific (preventing enumeration is a security driver for this particular scenario).

---

## Anti-patterns

### The implementation leak

```gherkin
# Wrong — asserts implementation, not behavior
Scenario: Place an order
  When the customer calls POST /api/v2/orders
  And the OrderHandler.create method is invoked
  Then the OrdersRepository.save method is called with the right arguments
```

```gherkin
# Right — asserts behavior, not method calls
Scenario: Place an order
  When the customer submits a cart with a valid payment method
  Then the order is created with status "placed"
  And the customer receives an order confirmation email
```

The scenario should make sense to Product. Method names don't.

### The novel

```gherkin
# Wrong — too much in one scenario
Scenario: Customer journey
  Given the customer browses to the storefront
  And finds a product
  And adds it to the cart
  And enters payment details
  And selects shipping
  When the customer submits the order
  Then the order is placed
  And the confirmation page is shown
  And the email is sent
  And the inventory is updated
  And the analytics event fires
  And the customer can see the order in their history
```

Split this into 4–5 smaller scenarios, each testing one decision. The "novel" hides which step actually broke when the scenario fails.

### The doppelganger

```gherkin
# Wrong — two scenarios that are nearly identical, varying only in one input
Scenario: Place an order with 1 item
  Given the cart has 1 item
  When the customer submits the order
  Then the order is placed

Scenario: Place an order with 2 items
  Given the cart has 2 items
  When the customer submits the order
  Then the order is placed

Scenario: Place an order with 3 items
  ...
```

Use a `Scenario Outline` with `Examples`:

```gherkin
Scenario Outline: Place an order with valid item counts
  Given the cart has <count> items
  When the customer submits the order
  Then the order is placed

  Examples:
    | count |
    | 1     |
    | 2     |
    | 3     |
    | 50    |
```

### The vague

```gherkin
# Wrong — what does "successfully" mean? what's "appropriate"?
Scenario: Order placement
  When the customer places an order
  Then the order is successfully created
  And appropriate notifications are sent
```

Right:

```gherkin
Scenario: Place an order with valid items and payment
  Given the cart contains the SKU "PRO-LICENSE" with quantity 1
  When the customer submits the order
  Then the order is created with status "placed"
  And the customer receives an order confirmation email at their on-file address
  And the inventory for "PRO-LICENSE" is decremented by 1
```

Every `Then` clause is verifiable — there's an observable thing to check.

---

## Mapping Gherkin to tests

Every scenario in Section 4 maps to a test in the test suite. The test name follows the `should_X_when_Y` convention from `writing-tests`:

- Scenario `Place an order with valid items and payment` → test `should_place_order_when_cart_and_payment_are_valid`.
- Scenario `Reject duplicate idempotency key` → test `should_reject_with_409_when_idempotency_key_is_duplicate`.

If the mapping is not 1:1, the scenario is probably too coarse or the test is testing too much.

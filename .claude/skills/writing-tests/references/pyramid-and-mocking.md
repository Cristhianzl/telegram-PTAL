# Testing pyramid and mocking

The longer explanation of the layers and the doubles. Use when you need to justify a choice or rule out an approach.

---

## The pyramid

```
        /  E2E  \          ~10% — critical user journeys only
       /----------\
      / Integration \       ~20% — component interactions, API contracts
     /----------------\
    /    Unit Tests     \   ~80% — core logic, validation, transformations
   /____________________\
```

The percentages are a guideline, not a dogma — adapt to your project's complexity, but always **bias toward the base**.

| Layer        | Speed         | Cost      | Scope                     | Stability |
|--------------|---------------|-----------|---------------------------|-----------|
| Unit         | Milliseconds  | Cheap     | Single function/class     | Very stable |
| Integration  | Seconds       | Moderate  | Multiple components       | Stable    |
| E2E          | Minutes       | Expensive | Full system               | Fragile   |

### Rules per layer

**Unit**
- MUST be fast (< 100 ms each), isolated, deterministic.
- MUST NOT touch DB, network, filesystem, or real clocks.
- Should make up the majority of the suite.

**Integration**
- MUST test real component interaction.
- MUST NOT test business logic that belongs in unit tests.
- Use containerized services (Postgres, Redis, etc.) — never mock the boundary you're trying to integrate against.

**E2E**
- MUST cover critical user journeys only.
- MUST NOT cover edge cases, error conditions, or variations — those go in lower layers.
- Be very stingy with E2E tests. They are slow, fragile, expensive to maintain.

---

## Coverage priority matrix

| Priority | What to test                                                              | Test type    | Goal                          |
|----------|---------------------------------------------------------------------------|--------------|-------------------------------|
| P0       | Core business logic, domain rules, calculations, state transitions        | Unit         | 100% of branches              |
| P0       | Input validation at system boundaries                                     | Unit         | All valid + invalid inputs    |
| P1       | Error handling and failure paths                                          | Unit         | All expected error types      |
| P1       | Integration between internal components                                   | Integration  | Key interaction points        |
| P2       | External service communication contracts                                  | Integration  | Happy + main failure modes    |
| P2       | Data persistence operations (read/write)                                  | Integration  | CRUD + edge cases             |
| P3       | Critical user journeys (end-to-end)                                       | E2E          | Happy path only               |
| P3       | Simple getters, setters, trivial mappers                                  | None         | Not worth testing             |

---

## Mocking terminology

| Type    | Purpose                                              | Example                                                              |
|---------|------------------------------------------------------|----------------------------------------------------------------------|
| Dummy   | Fills a parameter; never used                         | `null` or empty object passed to satisfy a signature                  |
| Stub    | Returns predefined data                               | `getUser()` always returns a fixed user                              |
| Fake    | Working implementation with shortcuts                 | In-memory database instead of real DB                                |
| Spy     | Records calls for later verification                  | Verify `sendEmail()` was called once                                 |
| Mock    | Stub + spy with assertions                            | Expects `save()` to be called with specific args                     |

Use the right one. A mock is appropriate when the **call itself** is the behavior being tested (e.g., "sends a welcome email"). For everything else, prefer fakes or stubs that let you assert observable behavior on returned values.

---

## Mocking rules

### Do mock
- External HTTP APIs and third-party services
- Databases and data stores (in unit tests)
- Filesystem and I/O operations
- Clocks, timers, and random generators
- Email / SMS / notification services
- Message queues and event buses

### Don't mock
- The system under test itself
- Simple value objects or DTOs
- Pure functions with no side effects
- Things you don't own (prefer wrapping and mocking the wrapper)
- Implementation details (method call order, internal state)

### Mock with caution
- Don't over-mock. If your test setup has more mocks than assertions, reconsider.
- Prefer fakes over mocks for complex dependencies. An in-memory DB beats a fully mocked repository.
- When mocking, model BOTH the happy path AND failure modes. A mock that always succeeds teaches you nothing.
- Keep mocks close to real behavior. If the real API rejects a malformed payload with a 400, your mock should do the same — otherwise the test misses real defects.
- If you need to mock more than 3 dependencies, the production code has too many dependencies — refactor the code, then test.

---

## The fundamental rule — test the spec, not the code

> When writing tests, base your expected values on the **requirements, specification, or expected behavior** — NOT on what the source code currently does.

**Wrong (Confirmation Testing):**

```
1. Read the source code.
2. See that function returns X for input Y.
3. Write test: assert function(Y) == X.
4. Test passes ✓ (obviously — you copied the answer).
5. You learned NOTHING. You found ZERO bugs.
```

**Right (Defect Detection Testing):**

```
1. Read the REQUIREMENTS / SPEC.
2. Determine what the function SHOULD return for input Y.
3. Write test: assert function(Y) == expected_from_spec.
4. If test passes → great, code matches spec.
5. If test FAILS → you found a BUG.
```

### How to apply in practice

| Step                                              | What to do                                                                                                                 |
|---------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------|
| 1. Understand the requirement                      | Before reading the source code, understand WHAT the feature should do (docstrings, comments, PR descriptions, tickets, or ask the user). |
| 2. Define expected behavior independently          | Based on the requirement, decide what the correct output should be for each input — BEFORE looking at the implementation.   |
| 3. Write tests from the spec                       | Write your assertions based on step 2, not on what the code does.                                                           |
| 4. Include adversarial tests                       | Intentionally write tests that TRY to break the code: unexpected inputs, edge cases, boundary values.                       |
| 5. Expect some failures                            | If every test passes on the first run, ask yourself: "Am I actually testing behavior, or just mirroring the code?"          |

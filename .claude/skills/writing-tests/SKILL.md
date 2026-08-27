---
name: writing-tests
description: Write unit, integration, and E2E tests that follow the testing pyramid, the Arrange-Act-Assert pattern, the should_X_when_Y naming convention, and meet a 75% (target 80%) branch coverage gate per supported OS. Use when the user asks to write tests, add coverage, build a test suite, fix flaky tests, or verify a feature with tests. For TDD bug fixes use fixing-bugs; for TDD feature work use developing-features-tdd.
license: MIT
---

# Writing Tests

Test code is production code. Same quality bar — naming, structure, immutability, file structure, security. The point of the test suite is to **find bugs**, not to confirm what the code already does.

## Read first (always)

List `learnings/` and read every file relevant to the current task. Project-specific test framework choices, fixture conventions, flaky-test hotspots, or coverage carve-outs live there and override the defaults in this SKILL.md. If a learning conflicts with this file, **the learning wins** — mention it to the user.

## Tradeoff — when to apply, when to lighten up

Apply the full discipline (pyramid, AAA, coverage gate, multi-platform matrix) for **production code**. Lighten the formality for one-off scripts, exploration, or examples under 30 lines — but even then, at least one happy-path test and one adversarial test.

## Pre-test analysis (mandatory)

Before writing a single test:

1. **Identify the test framework.** Look for `pytest.ini` / `jest.config.*` / `vitest.config.*` / `*_test.go` / `build.gradle` test deps / `*.Tests.csproj` / `Cargo.toml [dev-dependencies]`. Use the framework already in the project. Do NOT introduce a new one.
2. **Identify the mocking library.** `unittest.mock` / `pytest-mock`, `jest.mock` / `vitest.mock` / `sinon` / `msw`, `Mockito`, `gomock` / `testify/mock`, `Moq` / `NSubstitute`. Use what's already there.
3. **Identify existing test patterns.** How are files named and organized? Are there factories, fixtures, helpers, `conftest.py`, `__mocks__`? Reuse them. Do NOT duplicate.
4. **Identify the linter, formatter, type checker** used on test code. Run them in pre-commit (Step 5).

If you introduce a new pattern without checking existing ones, **stop and refactor**.

## Workflow

1. **Plan coverage from the spec, not the code.** List the behaviors the feature exhibits per the requirements. Categorize each as success, error, edge, or boundary. Map each to unit / integration / E2E using the pyramid below.
   → verify: every behavior from the spec maps to at least one planned test; pure infrastructure glue maps to ≤1.

2. **Write tests in priority order.** P0 critical first (core business logic, input validation), P1 next (error handling, internal integration), P2 (external service contracts, persistence), P3 (E2E happy paths only).
   → verify: the next test you're about to write is at or above the priority of any uncovered higher-priority item.

3. **Each test follows AAA / Given-When-Then.** One Act per test. One logical assertion (multiple `assert` lines OK if they verify a single outcome). Clear visual separation between phases.
   → verify: the test reads top-to-bottom without jumping to helpers to understand intent.

4. **Name with `should_[expected]_when_[condition]`** (or the language's idiomatic equivalent). Describe behavior, not implementation.
   → verify: a reader can guess the test body from the name.

5. **Mock at boundaries; fake at depth.** Mock external HTTP, DB (in unit tests), filesystem, clocks, RNG, queues. Don't mock the SUT, value objects, or pure functions. Prefer fakes over mocks for complex dependencies (in-memory DB beats mocked repository).
   → verify: the test setup has fewer mocks than assertions; nothing under test is mocked.

6. **Make tests independent and deterministic.** No shared mutable state. No real time, no real RNG, no real network in unit tests. Each test creates its own state.
   → verify: the test passes in isolation, in random order, and concurrently with siblings.

7. **Run pre-commit validation** (the 8 steps in `references/pre-commit.md`).
   → verify: every step gate passes; the coverage gate is met per OS.

8. **Capture a learning (final step).** Ask: *did I encounter a testing convention, framework quirk, flaky pattern, or coverage carve-out not in this SKILL.md or `references/`?* If yes, append a `learnings/YYYY-MM-DD-slug.md`. If no, skip.

## Testing pyramid

```
        /  E2E  \          ~10% — Critical user journeys only
       /----------\
      / Integration \       ~20% — Component interactions, API contracts
     /----------------\
    /    Unit Tests     \   ~80% — Core logic, validation, transformations
   /____________________\
```

| Layer        | Speed         | Cost      | Scope                     | Stability |
|--------------|---------------|-----------|---------------------------|-----------|
| Unit         | Milliseconds  | Cheap     | Single function/class     | Very stable |
| Integration  | Seconds       | Moderate  | Multiple components       | Stable    |
| E2E          | Minutes       | Expensive | Full system               | Fragile   |

**Per-layer rules:**

| Layer        | MUST                                                | MUST NOT                                                       |
|--------------|-----------------------------------------------------|----------------------------------------------------------------|
| Unit         | Be fast (<100ms each), isolated, deterministic       | Touch DB, network, filesystem, or real clocks                  |
| Integration  | Test real component interaction                      | Test business logic that belongs in unit tests                 |
| E2E          | Cover critical user journeys only                    | Cover edge cases, error conditions, or variations              |

## What to test — the behavior checklist

For every feature, systematically cover these four categories:

- **Success** — primary expected behavior with valid inputs; boundary values still valid (min/max allowed); multiple valid variations.
- **Error** — invalid inputs rejected; missing required fields; unauthorized actions; external dependency failures; domain rule violations.
- **Edge** — empty, null, zero, very large, Unicode, special chars, concurrent/idempotent calls.
- **State transitions** (if applicable) — valid transitions, invalid transitions, transition side effects.

Happy-path tests alone are **insufficient**. Coverage of all four categories is the goal of each feature's test set.

## Tests must challenge the code, not confirm it

> A test suite where everything passes on the first try is suspicious. Good tests are adversarial — they actively look for problems.

- Write tests based on **requirements / spec / expected behavior**, NOT on what the source code currently does. If you copy what the code does into the assertion, you found zero bugs.
- Include tests that intentionally try to break the code: unexpected types, boundary values, malformed data, concurrent calls.
- When a test fails: **first ask if the CODE is wrong**, not the test. Never silently change an assertion to match buggy code.

Full anti-patterns and adversarial examples in `references/anti-patterns.md`.

## Mocking — terminology and rules

| Type    | Purpose                                           |
|---------|---------------------------------------------------|
| Dummy   | Fills a parameter; never used                     |
| Stub    | Returns predefined data                           |
| Fake    | Working implementation with shortcuts             |
| Spy     | Records calls for later verification              |
| Mock    | Stub + spy with assertions                         |

**Do** mock external HTTP, DB (unit tests), filesystem, clocks, timers, RNG, email/SMS, queues.
**Don't** mock the SUT, value objects, pure functions, things you don't own (wrap them and mock the wrapper).

If you need to mock more than three dependencies, the code under test probably has too many — refactor the production code, then test.

## Coverage — meaningful, not vanity

- **Target: 80% branch coverage. Minimum: 75%.** Below 75% the task is not complete.
- Focus on **branch** coverage, not just line coverage. Both sides of every `if/else`, every `catch`, every error path.
- Coverage is a **diagnostic** for what's missing, not a goal for what exists.

Per-category targets:

| Code category                       | Floor | Target | Ideal     |
|-------------------------------------|-------|--------|-----------|
| Core business logic                 | 75%   | 80%    | 90-100%   |
| Input validation                    | 75%   | 80%    | 90-100%   |
| Error handling paths                | 75%   | 80%    | 85-95%    |
| Data transformation / mapping       | 75%   | 80%    | 85-95%    |
| API/HTTP handlers                   | 75%   | 80%    | 80-90%    |
| Simple DTOs, getters, setters       | —     | —      | don't test |

**Honest vs vanity coverage:** every test must have at least one meaningful assertion. Coverage reports are reviewed for **uncovered branches**, not just line percentages. Pair with mutation testing where available (`mutmut`, `Stryker`, `pitest`).

## Multi-platform — the matrix is mandatory

A test suite that only runs on one OS lies about coverage. CI on Linux says nothing about Windows.

```yaml
strategy:
  fail-fast: false      # critical — one OS failing must not abort the others
  matrix:
    os: [ubuntu-latest, macos-latest, windows-latest]
runs-on: ${{ matrix.os }}
```

Tests are implicitly Linux-only when they hardcode `/tmp`, depend on `\n`-only line endings, call `subprocess.run("ls")`, assume case-sensitive filesystems, or fork worker processes. Fix patterns in `references/multi-platform.md`.

Coverage counts **per OS**. Combine reports with `coverage combine` / `nyc merge` before comparing against the gate.

## Test file structure

- **Mirror source structure.** `src/users/user_service.py` → `tests/users/test_user_service.py` (or colocated in JS/TS).
- **One subject per file.** A 1200-line file covering one module thoroughly is better than 5 small files that fragment related tests.
- **Setup ≤ 20 lines per test.** Extract to factories/helpers if exceeded.
- **No catch-all helper files.** Name by responsibility: `factories/user_factory.py`, `assertions/order_assertions.py`. Never `test_utils.py`.
- **Separate unit from integration** by directory: `tests/unit/`, `tests/integration/`, `tests/e2e/`.

Detail and naming conventions per language in `references/file-structure.md`.

## Test isolation — hard rules

| Rule                                                | Why                                                 |
|-----------------------------------------------------|-----------------------------------------------------|
| Each test independent                               | No reliance on another test's execution             |
| Deterministic                                       | Same code → same result, every time, every machine  |
| No shared mutable state                             | Fresh setup per test                                 |
| No execution-order dependency                       | Shuffled order must still pass                       |
| No real time                                        | Mock clocks, timers, `Date.now()`, `time.time()`     |
| No real randomness                                  | Seed or mock                                         |
| No real network in unit tests                       | Integration may use containerized services           |
| Tests clean up after themselves                     | Teardown / `afterEach` resets state                  |

**Flaky-test policy:** a flaky test is **broken**. Fix it immediately or quarantine with a ticket reference. Common causes: time-dependent logic, shared mutable state, race conditions, external service dependencies, hardcoded sleeps instead of proper async handling.

## Pre-commit validation (8 steps)

Full bash commands per language in `references/pre-commit.md`. Summary:

1. Run each new test **in isolation**.
2. Run all new tests in **random order**.
3. Run all new tests **together** (catches shared-state leaks).
4. Verify coverage on changed code (intermediate check).
5. Run project linter, formatter, type checker on **test files** (test code is production code).
6. **All created/modified tests pass.** Zero failures, zero skips. Never `@skip` to hide a failure.
7. **Run coverage report for ALL created tests** (backend + frontend if both), **show output to the user**, verify ≥ 75% (target 80%).
8. Final checklist verified (`references/pre-commit.md` § Step 8).

If any step fails → fix before declaring done.

## Anti-patterns to avoid

Eight named anti-patterns (Mirror, Liar, Giant, Mockery, Inspector, Chain Gang, Flaky, Snowball) — full examples in `references/anti-patterns.md`. Quick summary:

| Anti-pattern   | Symptom                                                          |
|----------------|-------------------------------------------------------------------|
| Mirror         | Reads code, asserts what code does — finds zero bugs              |
| Liar           | Passes but doesn't verify the behavior its name claims            |
| Giant          | 50+ lines of setup, multiple Acts, dozens of unrelated assertions |
| Mockery        | More mocks than assertions; testing the mocks, not the code       |
| Inspector      | Asserts call counts and method order; breaks on every refactor    |
| Chain Gang     | Tests depend on each other's state or order                        |
| Flaky          | Passes sometimes, fails sometimes, no code changes                 |
| Snowball       | Giant snapshot that breaks on every minor change                   |

## Output format

1. **Test code** — clean, complete, no placeholders, following all rules above.
2. **Coverage summary** — what is covered (success / error / edge / state) and what is intentionally not covered (with justification).
3. **Brief explanation** — testing decisions and trade-offs in 3–5 bullets. Concise.

## See also

- `references/pyramid-and-mocking.md` — pyramid detail; mock terminology; mocking rules; over-mocking smell.
- `references/anti-patterns.md` — the 8 named anti-patterns with wrong/right examples.
- `references/file-structure.md` — directory layout, naming convention, factories, shared infrastructure.
- `references/pre-commit.md` — 8-step pre-commit validation with bash commands per language.
- `references/multi-platform.md` — CI matrix, platform-conditional tests, line endings, Docker tests.
- `developing-features` skill — code quality rules that apply to test code (SOLID, file structure, security).
- `developing-features-tdd` skill — when tests drive the implementation (RED → GREEN cycle).
- `fixing-bugs` skill — bug-fix tests (the narrow TDD case for defects).
- `ensuring-cross-platform` skill — full platform-agnostic rules feeding the test matrix.
- `learnings/` — project-specific test conventions, framework quirks, flaky-zone alerts.

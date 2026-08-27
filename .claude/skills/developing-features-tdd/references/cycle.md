# The TDD cycle — phase-by-phase detail

Companion to `SKILL.md`. Read this when you need the longer explanation of a phase.

---

## Phase 1 — UNDERSTAND

Do **not** open the editor until you can answer the seven questions in SKILL.md.

Decompose each acceptance criterion into Given/When/Then/So-That:

```
GIVEN:    a registered user with valid credentials
WHEN:     they submit a login request with the wrong password
THEN:     the system returns 401 with a generic error message
SO THAT:  attackers cannot enumerate valid usernames via timing or error content
```

`SO THAT` is what gives the test a reason to exist beyond "the code does what the code does". It surfaces the business or security driver. Without it, the test becomes a tautology.

If a criterion contains "and" or "or" → split it. Each test asserts one observable behavior.

### Pre-implementation security check

Five questions, every feature:

1. **What am I trusting here that I haven't verified?**
2. **What is the blast radius if this goes wrong?**
3. **Am I implementing this from the authoritative source?**
4. **What happens in the failure path?**
5. **Who controls this value, and can they lie?**

If the feature touches AI/LLM, webhooks, auth, or data persistence, the answers **must** become adversarial tests in Phase 3 before any production code exists.

---

## Phase 2 — DESIGN

Why this phase exists: TDD has a well-known failure mode where the developer writes 10 GREEN tests on top of an architecture that has six different responsibilities glued together in one file. Phase 2 prevents that.

The design output (in scratch notes, plan tool, or PR draft) is in SKILL.md.

### Layer rules

| Layer                       | CAN                                                    | CANNOT                                                       |
|-----------------------------|--------------------------------------------------------|--------------------------------------------------------------|
| Handler / Controller        | Receive input, delegate to service, return output      | Contain business logic, call DB directly, call external APIs directly |
| Service / Orchestrator      | Coordinate operations, apply business rules            | Know about HTTP/transport, execute SQL/queries directly      |
| Repository / Data access    | Execute queries, map data                              | Make business decisions, call external APIs                  |
| Helper                      | Transform data, validate, format                       | Have side effects, do I/O, maintain state                    |
| External client             | Communicate with external services                     | Contain business logic, access the database                  |

### Durability anchor

Most LLM-assisted work loses intent. The session that produced the code ends, the PR description gets archived, and only the code remains. If the *why* of a non-obvious design choice is not anchored to a durable location, it is lost.

Durable locations, in order of preference:

1. **Test name** — `test_should_use_redis_when_payload_exceeds_1MB_to_avoid_postgres_TOAST_overhead`.
2. **`Why:` code comment** at the decision point, one line.
3. **PR description** `## Design Decisions` section.
4. **ADR** (`docs/adr/NNNN-<slug>.md`) for big choices — write one only when the choice affects more than the immediate file.

Capture each non-obvious decision in at least one of these. Otherwise it's gone.

---

## Phase 3 — RED

Slice rules:

- One slice = one valuable observable behavior. Smaller is better.
- Slice order: happy → edge → error → adversarial → integration → AI failure-mode.
- Each test exercises the slice through its public surface, mocking all external dependencies.
- Each test fails with a reason that screams "this feature isn't built yet" — `AttributeError`, `NotImplementedError`, missing module, wrong return value.

Common mistake: writing a test that exercises three slices at once. Split it. Each test asserts one decision.

### Adversarial slice examples

```python
# Webhook integration
def test_should_reject_request_when_signature_is_invalid(): ...
def test_should_reject_request_when_timestamp_is_outside_tolerance(): ...
def test_should_reject_request_when_body_is_modified_after_signing(): ...

# Auth boundary
def test_should_deny_when_user_is_not_owner_of_resource(): ...
def test_should_deny_when_token_is_expired(): ...
def test_should_deny_when_role_lacks_required_permission(): ...

# LLM integration
def test_should_invoke_fallback_when_llm_times_out(): ...
def test_should_return_domain_error_when_llm_output_is_unparseable(): ...
def test_should_skip_llm_call_when_circuit_breaker_is_open(): ...
def test_should_bypass_llm_when_kill_switch_disabled(): ...
```

See `test-quality.md` for full naming and structure rules.

---

## Phase 4 — VERIFY RED

```
TEST: test_should_create_user_when_input_is_valid
STATUS: FAILED
REASON: AttributeError: 'UserService' object has no attribute 'create_user'
EXPECTED REASON: Yes — production code for this slice does not exist yet
```

If the test passes before code exists, three possible causes:

1. The test asserts something already true by accident → tighten the assertion.
2. The test hits a stub or default that satisfies it → make the slice more specific.
3. The behavior already exists somewhere → re-evaluate the slice scope.

If the test fails for the wrong reason, three possible causes:

1. Setup error in Arrange → fix the fixture.
2. Wrong dependency wiring → revisit Phase 2 design.
3. Test depends on something not yet decided → split the slice further.

Don't proceed to GREEN until the failure reason is clean.

---

## Phase 5 — GREEN

Minimum code. No extra methods. No optional parameters "for future use". No retry/cache "while I'm here".

If a slice forces you to violate an architectural rule (SRP, DIP, cyclomatic complexity ≤ 10, nesting ≤ 4), the **slice is too big** or the **design from Phase 2 is wrong**. Return to Phase 2.

### Architectural rules that apply even in GREEN

These cannot be deferred to REFACTOR:

| Rule                                            | Meaning                                                                          |
|-------------------------------------------------|-----------------------------------------------------------------------------------|
| SRP                                             | Function/class describable in ONE sentence without "and"/"or"                     |
| DIP                                             | High-level code depends on abstractions, not concrete infrastructure              |
| No global state                                 | No singletons-as-state, no module-level mutable dicts                             |
| Strong typing                                   | No `any`, `object`, `dynamic`                                                     |
| No magic values                                 | Numbers/strings used as flags or thresholds → named constants                     |
| Early returns / guard clauses                   | Reduce nesting; fail fast at boundaries                                           |
| No silent failures                              | Every error path returns or raises a domain-meaningful error                      |
| No PII/secrets in logs                          | Even in "temporary" debug logs — strip or hash                                    |
| Cyclomatic complexity ≤ 10                      | New functions never exceed CC=10                                                  |
| Nesting depth ≤ 4                               | If a 5th level appears, extract the inner block to its own function               |

---

## Phase 6 — VERIFY GREEN

Run the same test. It must pass.

If it still fails:

- The implementation is incomplete → finish it.
- The test is asserting the wrong thing → refine the assertion (rare — usually the code is wrong).
- The wiring is broken → check dependency injection / configuration.

---

## Phase 7 — REFACTOR

Refactor is **only** allowed when tests are GREEN.

| Smell                                                       | Refactor                                                  |
|-------------------------------------------------------------|------------------------------------------------------------|
| Duplicated logic across 2+ places (5+ lines, same intent)   | Extract to a named helper                                  |
| `if/elif` chain on a type/category with 3+ branches         | Strategy / polymorphism (OCP)                              |
| Function describable only with "and"/"or"                   | Split into two (SRP)                                       |
| Long call chain                                             | Encapsulate behind a method on the direct collaborator     |
| Method does work that doesn't belong to its class           | Move to the class that owns the data                       |
| Magic number/string                                         | Named constant                                             |
| Deep nesting                                                | Early returns / guard clauses                              |
| Mixed responsibility prefixes in one file                   | Split file by responsibility                               |
| Bloated interface (most implementors stub methods)          | Split into smaller interfaces (ISP)                        |

### Rules

- Small, incremental steps — one transformation at a time.
- Run tests after each change. Never batch refactors.
- No new behavior — refactors must not alter what tests assert.
- Refactor commit is **separate** from slice implementation.
- Prefer duplication over wrong abstraction.

### When to apply OCP

| Apply OCP                                                              | Skip OCP                                              |
|------------------------------------------------------------------------|--------------------------------------------------------|
| 3+ branches on type/category and more variants are expected            | Only 2 branches and stable                            |
| `if/elif` chain grows with each feature request                        | Branching logic is genuinely terminal                  |
| Different behaviors share an interface but vary in implementation      | Each branch is a one-liner                             |

---

## Phase 8 — VALIDATE

Full suite + coverage + linters + quality gates. See `checklist.md` for the full list.

If existing tests break:

1. Analyze — is the breaking test correct, or was it asserting behavior your feature legitimately changed?
2. If the test was right → your feature has a regression. Return to Phase 5 / 7.
3. If the requirement changed → update the test in the same PR with a clear `Why:` in the commit.
4. **Never delete a failing test to make CI green.**

---

## Phase 9 — REPEAT

Pick the next slice. Return to Phase 3.

When all acceptance criteria and all threat-model items have a corresponding passing test, the feature is **done**. Otherwise it isn't.

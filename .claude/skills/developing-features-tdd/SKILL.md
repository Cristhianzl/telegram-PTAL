---
name: developing-features-tdd
description: Build a new feature using strict test-driven development — UNDERSTAND → DESIGN → RED → VERIFY RED → GREEN → VERIFY GREEN → REFACTOR → VALIDATE → REPEAT. Use when building a new feature with TDD, when the user says "TDD this", "tests first", "red green refactor", or asks for feature work that needs to be verifiable from the spec. For bug fixes use fixing-bugs; for non-TDD production code use developing-features.
license: MIT
---

# Developing Features with TDD

Test-driven feature development. Every observable behavior is driven by a failing test that goes RED → GREEN, with architecture planned before the first test is written.

> **Never write production code for a new feature before proving its expected behavior with a failing test.**

## Read first (always)

List `learnings/` and read every file relevant to the current feature. Project-specific test conventions, framework quirks, or constraints live there and override the defaults in this SKILL.md. If a learning conflicts with this file, **the learning wins** — mention it to the user.

## Tradeoff — when to apply, when to lighten up

Apply the full nine-phase cycle for **production features** the user explicitly asks to build with TDD, or for code paths where regressions are expensive (payments, auth, data integrity, AI runtime, public APIs).

Lighten the formality for one-off scripts, throwaway exploration, prototypes the user has labeled as such, or features under 30 lines of production code. Still write at least one happy-path test and one adversarial test — but skip the explicit Phase 2 design document.

## Best practices — inherited from `developing-features` (this skill changes only the *how*)

Every engineering standard in `developing-features` applies here **in full and unchanged**. TDD changes *how* you implement (test-first: RED → GREEN → REFACTOR), not *what* "good" means — it never relaxes or replaces these:

- The **pre-implementation security check**, the **reuse ladder** (existing code → stdlib → platform → installed dep → only then new code), the **trade-off priority** (correctness → simplicity → testability → performance → reuse), and the **core code-quality rules**: SOLID, DRY/KISS/YAGNI/Law of Demeter, file-structure hard limits, intention-revealing naming, immutability, no global state, strong typing, guard clauses, complexity ≤ 10 / nesting ≤ 4, and no-WHAT-comments.
- **Error handling**, **security**, **observability**, **data layer & scale** (pooling, indexes, no-N+1, pagination), and **platform-agnostic** rules.
- Their full detail lives in `developing-features/references/{security,solid,pragmatic-principles,file-structure,observability,data-layer}.md` — read them as needed.

The phases below weave these into the RED/GREEN/REFACTOR/VALIDATE loop; they are the **same bar**, reached test-first.

## The cycle (strict order)

```
1. UNDERSTAND     →  Decompose requirements (Given/When/Then) + threat model
2. DESIGN         →  Plan architecture (files, layers, responsibilities) BEFORE coding
3. RED            →  Write a failing test for the next slice of behavior
4. VERIFY RED     →  Run the test — it MUST fail, for the right reason
5. GREEN          →  Write the MINIMUM production code that makes it pass
6. VERIFY GREEN   →  Run the test — it MUST now pass
7. REFACTOR       →  Clean up, apply SOLID/DRY/KISS — tests stay GREEN
8. VALIDATE       →  Run ALL tests; coverage ≥ 75% (target 80%); linters/formatters pass
9. REPEAT         →  Next slice (steps 3–8) until all acceptance criteria are covered
```

If you skip steps or change the order, the feature is invalid. The phases are detailed in `references/cycle.md`.

## Phase 1 — UNDERSTAND

Before opening the editor, you must be able to answer:

| Question                                                                | Why                                                  |
|-------------------------------------------------------------------------|------------------------------------------------------|
| What problem does this feature solve?                                   | Defines success criteria                              |
| Who is the user/actor?                                                  | Defines API surface and permissions                   |
| What are the explicit acceptance criteria?                              | Defines what tests must prove                         |
| What are the inputs, outputs, and side effects?                         | Defines function/service contracts                    |
| What can go wrong (errors, edge cases, abuse)?                          | Defines adversarial tests                             |
| What systems does this touch (DB, APIs, AI, queues)?                    | Defines integration boundaries                        |
| What is the security blast radius if this misbehaves?                   | Drives the threat model                               |

For each acceptance criterion, write:

```
GIVEN:    <precondition / initial state>
WHEN:     <action the user/system performs>
THEN:     <expected observable behavior>
SO THAT:  <business value>
```

If any criterion cannot be decomposed → **ask for clarification**. Don't guess scope.

**Pre-implementation security check** — the same questions from the `developing-features` skill apply (all of them). If the feature touches AI/LLM, webhooks, auth, or data persistence, the answers **must become adversarial tests** in Phase 3 before the corresponding production code exists.

## Phase 2 — DESIGN

TDD does not exempt you from designing first. Skipping design produces small green tests sitting on top of spaghetti architecture.

Before the first RED test, produce this design doc (in scratch notes, plan tool, or PR draft):

```
FEATURE: <name>

SLICES (in build order, smallest valuable behavior first):
  1. <slice 1>
  2. <slice 2>
  3. <...>

FILE STRUCTURE:
  - src/<area>/types/<feature>_types.<ext>
  - src/<area>/helpers/<feature>_validation.<ext>
  - src/<area>/services/<feature>_service.<ext>
  - src/<area>/repositories/<feature>_repository.<ext>
  - src/<area>/handlers/<feature>_handler.<ext>

DEPENDENCIES (injected, not instantiated):
  - <ServiceX> depends on abstraction <RepositoryY>

THREAT MODEL (must become tests):
  - <Threat 1> → test_should_<reject|deny|fail>_when_<condition>
  - <Threat 2> → ...

DURABILITY ANCHOR (where intent lives after this PR):
  - Test names encoding each acceptance criterion
  - PR description: ## Acceptance Criteria → Tests
  - ADR (only for non-obvious decisions): docs/adr/<NNNN>-<slug>.md

NON-OBVIOUS DESIGN DECISIONS (each must be anchored above):
  - <decision 1>: <why this over the alternative>
```

File-structure rules (hard limits, every language): see `developing-features` skill, `references/file-structure.md`. Layer rules (handler ≠ service ≠ repository ≠ helper ≠ client) apply throughout.

**If the feature calls an LLM at runtime**, the design also includes an AI runtime profile (timeout, circuit breaker, fallback, cost ceiling, kill switch, SLO). See `references/ai-runtime.md`. These items **become failure-mode tests** in Phase 3.

## Phase 3 — RED

Build the feature **one thin slice at a time**. One slice = one RED test = one minimal GREEN change.

**Slice order** (recommended):

1. Happy path of the smallest valuable behavior.
2. Edge cases (empty, null, boundary, max).
3. Error cases (invalid input, downstream failure, timeout).
4. Adversarial cases (forged input, replay, injection, scope bypass).
5. Integration seams (handler ↔ service ↔ repository) — only after unit slices are GREEN.
6. **AI runtime failure modes** (mandatory if the feature calls an LLM): timeout, malformed output, circuit open, rate limit / safety filter, kill switch toggled off.

**Each test must:**

- Assert observable behavior from the acceptance criteria, not implementation details.
- Be independent and deterministic — no shared mutable state, no order coupling.
- Mock or fake all external dependencies (DB, HTTP, filesystem, time, randomness).
- **Fail before the production code exists**, proving the behavior is missing.
- **Pass after the minimum production code is written**, proving the behavior is delivered.
- Use Arrange–Act–Assert (or Given–When–Then) structure.

**Test naming:** `should_[expected_behavior]_when_[condition]`. Examples and full quality rules in `references/test-quality.md`.

→ verify: the test runs the slice through its public surface, mocks all external deps, has one focused assertion (or one focused group), and the failure message clearly names the missing behavior.

## Phase 4 — VERIFY RED

Run the test. It MUST fail.

| Check                                                                 | Expected                                                       |
|-----------------------------------------------------------------------|----------------------------------------------------------------|
| Does the test fail?                                                   | Yes — must fail                                                |
| Is the failure reason "feature isn't built yet" shaped?               | `AttributeError`, `NotImplementedError`, missing module, etc.  |
| Is the assertion checking observable behavior, not internals?         | Yes                                                            |

If the test **passes** before the code exists, the assertion is trivially true — tighten it, or split the slice further.

If the test **fails for the wrong reason** (setup error, wrong wiring, unmet dependency from another slice), fix the test plumbing or split the slice — don't proceed to GREEN.

Document the RED state in scratch notes:
```
TEST: test_should_create_user_when_input_is_valid
STATUS: FAILED
REASON: AttributeError: 'UserService' object has no attribute 'create_user'
EXPECTED REASON: Yes — production code for this slice does not exist yet
```

## Phase 5 — GREEN

Write only enough code to make the current RED test pass. Do not implement future slices.

- Change/add the **minimum** code necessary.
- Place code in the file decided in Phase 2. If you find you need a file you didn't plan, pause and revise the plan — don't dump it in a generic `utils`.
- Follow the layer rules from Phase 2 (handler ≠ service ≠ repository ≠ helper ≠ client).
- Inject dependencies via constructors/parameters; never instantiate infrastructure inside business logic.
- Do **not** add features not driven by the current RED test.
- Do **not** pre-build retry policies, caches, plugin systems, or configuration knobs "for later" (YAGNI).

**Architectural rules that apply even in GREEN** (cannot be deferred to REFACTOR):

- SRP (one sentence without "and"/"or"), DIP (depend on abstractions), no global state, strong typing (no `any`/`object`/`dynamic`), no magic numbers, early returns / guard clauses, no silent failures, no secrets in logs.
- Cyclomatic complexity ≤ 10 in new functions; nesting depth ≤ 4. If a slice forces a violation, the design is wrong — return to Phase 2.

If GREEN forces violation of any of these, the **design is wrong** — return to Phase 2, don't push through.

## Phase 6 — VERIFY GREEN

Run the same test from Phase 3. It MUST now pass.

| Check                                                          | Expected |
|----------------------------------------------------------------|----------|
| Does the previously failing test now pass?                     | Yes      |
| Did you write the minimum code to make it pass?                | Yes — no scope creep |
| Did you respect the planned file structure and layers?         | Yes      |
| Are dependencies injected, not hard-wired?                     | Yes      |

## Phase 7 — REFACTOR

Refactor **only** when tests are GREEN. After every small step, re-run tests.

| Smell                                                                  | Refactor                                                  |
|------------------------------------------------------------------------|------------------------------------------------------------|
| Duplicated logic across 2+ places (5+ lines, same intent)              | Extract to a named helper                                  |
| `if/elif` chain on a type/category with 3+ branches                    | Strategy / polymorphism (OCP)                              |
| Function describable only with "and"/"or"                              | Split into two (SRP)                                       |
| Long call chain (`a.b.c.d.method()`)                                   | Encapsulate behind a method on the direct collaborator     |
| Magic number/string                                                    | Named constant                                             |
| Deep nesting                                                           | Early returns / guard clauses                              |
| Mixed responsibility prefixes in one file                              | Split file by responsibility                               |
| Bloated interface (most implementors stub methods)                     | Split into smaller interfaces (ISP)                        |

**Refactoring rules:**
- Small incremental steps — one transformation at a time.
- Run tests after each change.
- No new behavior — refactors must not alter what tests assert.
- Separate commit from slice implementation.
- Prefer duplication over wrong abstraction.

## Phase 8 — VALIDATE

Run the FULL test suite. All tests must pass. New + pre-existing.

- [ ] All new tests pass; all pre-existing tests still pass.
- [ ] Linter, formatter, type checker pass for every language touched.
- [ ] No `any`/`object`/`dynamic` introduced; no suppressed lint rules without a `Why:` comment.
- [ ] **Cyclomatic complexity ≤ 10** in new functions; **nesting ≤ 4**.
- [ ] **Coverage ≥ 75% (target 80%)** — branch coverage, shown to the user.
- [ ] Platform-agnostic checks pass (see `ensuring-cross-platform` skill).
- [ ] **If it touches a DB:** pooled connection with limits, indexes on FK / filter columns, no N+1, lists paginated (see `developing-features/references/data-layer.md`).
- [ ] **AI runtime resilience tests pass** if the feature calls an LLM (timeout → fallback, malformed output → graceful error, circuit-open → bypass, kill switch → deterministic disable).
- [ ] **Comprehension gate** — you can defend every significant block of the diff without re-reading it.
- [ ] **Architectural intent is captured** for every non-obvious decision (test name, `Why:` comment, PR `## Design Decisions`, or ADR).

If existing tests break:

1. **Analyze** — is the breaking test correct, or was it asserting behavior your feature legitimately changed?
2. If the test was right → your feature has a regression. Return to Phase 5 / 7.
3. If the requirement legitimately changed → update the test in the same PR with a clear `Why:` in the commit description.
4. **Never delete a failing test to make CI green.**

## Phase 9 — REPEAT

Pick the next slice from Phase 2 (or refine if you learned something). Return to Phase 3.

**The feature is done when:**

- Every acceptance criterion maps to ≥1 passing test.
- Every threat-model item maps to ≥1 passing adversarial test.
- All architectural rules are satisfied.
- Coverage ≥ 75% (target 80%).
- All linters/formatters/type checkers pass.
- CI matrix passes on every supported OS (see `ensuring-cross-platform`).

## Tests are permanent

The tests you wrote during the RED→GREEN cycle are **the executable specification** of this feature. Never delete a feature test unless the feature is removed. If a future PR breaks one, either fix the PR or update the test in the same PR with `Why:` documentation.

## Final step: capture a learning

Before closing the task, ask: *did I encounter a TDD-relevant constraint, framework quirk, or testing trap not in this SKILL.md or `references/`?* If yes, append a `learnings/YYYY-MM-DD-slug.md`. If no, skip.

## Output format

1. **Requirement decomposition** — Given/When/Then/So-That per acceptance criterion.
2. **Threat model** — risks → adversarial tests planned.
3. **Architecture plan** — file structure, layers, dependencies (Phase 2 output).
4. **Slice-by-slice build** — for each slice: RED test → GREEN code → REFACTOR notes (if any).
5. **Final test run** — full suite passes, coverage shown.
6. **Brief explanation** — architectural decisions in 3–5 bullets.

## See also

- `references/cycle.md` — phase-by-phase detail with examples for each step.
- `references/test-quality.md` — naming convention, AAA structure, what to test / what NOT to test, anti-patterns.
- `references/ai-runtime.md` — AI runtime profile, mandatory failure-mode tests for LLM features.
- `references/checklist.md` — the full feature checklist + PR description template + commit format.
- `developing-features` skill — non-TDD production code rules (SOLID, security, file structure, observability, data layer) referenced throughout this skill; incl. `references/data-layer.md` (connection pooling, indexes, no-N+1, pagination).
- `ensuring-cross-platform` skill — portability rules that apply at RED, GREEN, REFACTOR, and VALIDATE.
- `writing-tests` skill — broader testing strategy (pyramid, integration, E2E, coverage).
- `learnings/` — project-specific TDD conventions accumulated over time.

# Feature checklist, commit format, and PR description

The end-of-feature gate. Every item must be checked before the feature is considered done. Below the checklist are the per-slice commit format and the PR description template.

---

## Feature checklist

```
PHASE 1 — UNDERSTAND
[ ] Acceptance criteria are clear and decomposable (Given/When/Then/So-That)
[ ] Inputs, outputs, side effects, and integration points are documented
[ ] Threat model completed (what can go wrong, blast radius, trust boundaries)
[ ] Adversarial scenarios identified (will become tests in Phase 3)

PHASE 2 — DESIGN
[ ] Functions/classes listed and categorized by responsibility
[ ] File structure planned (semantic names, no utils/helpers/misc)
[ ] Layer responsibilities respected (handler / service / repo / helper / client)
[ ] Dependencies identified for injection (no infrastructure instantiated in business logic)
[ ] Slices ordered from smallest valuable behavior to last edge case
[ ] Durability anchor identified for each acceptance criterion (test name / PR section / ADR)
[ ] Non-obvious design decisions enumerated and assigned a capture location
[ ] If feature calls an LLM: AI runtime profile defined (see references/ai-runtime.md)

PHASES 3–6 — RED → GREEN (per slice, repeated)
[ ] Each slice started with a failing test (RED) BEFORE production code
[ ] Test names follow should_[expected]_when_[condition]
[ ] Tests assert observable behavior (not implementation details)
[ ] Tests are independent, deterministic, and mock external dependencies
[ ] Each RED was followed by minimum production code in the planned file/layer
[ ] No code added without a test driving it
[ ] Architectural rules respected at GREEN (SOLID, DIP, no globals, strong typing)
[ ] Cyclomatic complexity ≤ 10 and nesting ≤ 4 in every new function
[ ] If feature calls an LLM: AI runtime failure-mode tests written (timeout, malformed output, circuit-open, kill switch, cost ceiling)

PHASE 7 — REFACTOR
[ ] Refactoring done only with tests GREEN
[ ] Small steps with re-run tests between each
[ ] No new behavior introduced during refactor
[ ] OCP/strategy applied where 3+ type/category branches emerged
[ ] No premature abstraction (YAGNI respected)

PHASE 8 — VALIDATE
[ ] All new tests pass
[ ] All pre-existing tests still pass
[ ] Coverage ≥ 75% (target 80%) — branch coverage shown
[ ] Linter, formatter, type checker pass for every language touched
[ ] Cyclomatic complexity check passes (configured tool, not eyeballed)
[ ] No `any`/`object`/`dynamic` introduced
[ ] No suppressed lint rules without documented `Why:`
[ ] If it touches a DB: pooled connection with limits, indexes on FK/filter columns, no N+1, lists paginated (see developing-features/references/data-layer.md)

COMPREHENSION GATE (mandatory before "done")
[ ] I can defend every significant block of the diff without re-reading it
[ ] Every non-obvious design decision is captured durably (test name, `Why:` comment, PR section, or ADR)
[ ] No "the AI wrote it, I trust it" code in the diff — every line is defensible
[ ] No "future AI will refactor this" assumption baked in anywhere

ARCHITECTURE
[ ] No file >500 LOC AND mixed responsibilities (refactor required if both true)
[ ] No file with 6+ functions of DIFFERENT responsibilities
[ ] No file with 11+ functions even with same responsibility
[ ] No file with 2+ main classes of different responsibilities
[ ] No generic file names (utils/helpers/misc/common/shared)
[ ] Each file describable in ONE sentence without "and"/"or"

SECURITY
[ ] All inputs validated at system boundaries
[ ] No secrets, PII, or tokens in code or logs
[ ] Auth/AuthZ enforced server-side on every protected operation
[ ] AI/LLM features: input/output guardrails, no direct DB, scope enforcement in code
[ ] Third-party integrations: signature verification per official spec, timing-safe compare
[ ] Adversarial tests cover the threat-model items from Phase 1
[ ] Database access uses parameterized queries / ORM (no string concat)
[ ] No raw exceptions / stack traces leaked to external clients
[ ] AI-generated code in critical paths verified against the authoritative source

AI RUNTIME (only if feature calls an LLM in production)
[ ] Explicit timeout on every AI call (no SDK defaults)
[ ] Circuit breaker configured with explicit thresholds
[ ] Fallback path implemented and tested under failure
[ ] Cost ceiling defined with alerts
[ ] Kill switch implemented and verified
[ ] SLO documented (latency, success rate, fallback rate)
[ ] Observability emits the mandatory fields (operation, request_id, model, tokens, duration, outcome)

OBSERVABILITY
[ ] Structured logs at key decision points
[ ] Every log has correlation context (request ID, user ID, etc.)
[ ] Log event names are unique (no collisions)
[ ] No PII / secrets / non-action logs

ERROR HANDLING
[ ] Expected errors handled explicitly
[ ] No silent failures
[ ] Domain-meaningful errors (not generic Exception)
[ ] Errors documented as part of the API contract

PR HYGIENE
[ ] PR description lists every acceptance criterion and the test that proves it
[ ] PR description lists every threat-model item and the adversarial test that covers it
[ ] PR description has a ## Design Decisions section for every non-obvious choice
[ ] No claim in the PR description that isn't verifiable in the diff
[ ] Refactor commits are separate from feature-slice commits
```

---

## Per-slice commit format

Each RED→GREEN slice ideally produces one commit:

```
feat(<area>): <slice description>

Slice: <which acceptance criterion this slice covers>
Test:  <name of the test that drove this slice>
Why:   <one-line business reason>
```

Refactor commits are **separate**:

```
refactor(<area>): <what was cleaned up>

Driver: <which smell — duplication / OCP branch / SRP split / Demeter chain>
Tests:  unchanged, all GREEN
```

---

## Final PR description template

```markdown
## Feature
<one-paragraph summary of what was built and why>

## Acceptance Criteria → Tests
- AC1: <criterion> → `test_should_X_when_Y`
- AC2: <criterion> → `test_should_A_when_B`
- ...

## Threat Model → Adversarial Tests
- Threat: forged webhook → `test_should_reject_when_signature_is_invalid`
- Threat: cross-tenant access → `test_should_deny_when_user_is_not_owner`
- ...

## Architecture
- Files added: <list, grouped by layer>
- Layers: <handler> → <service> → <repository>
- Key abstractions: <interfaces injected, not concrete classes>

## Design Decisions
- <Non-obvious choice 1>: <why this approach over the alternative — here OR in docs/adr/NNNN-slug.md>
- <Non-obvious choice 2>: <...>
- (If none: "None — design follows established conventions")

## AI Runtime (only if feature calls an LLM in production)
- Call placement: <sync in request path | async via queue | background>
- Timeout: <e.g., 5s>
- Circuit breaker: <thresholds>
- Fallback: <what runs when the AI is unavailable>
- SLO: <P95 latency, success rate, fallback rate>
- Kill switch: <feature flag name>

## Coverage
- Backend: X% (≥75% required, target 80%)
- Frontend: X% (≥75% required, target 80%)

## Out of Scope (Follow-Ups)
- <anything explicitly NOT built, with reason>
```

---

## Anti-patterns to flag in the PR

- **Cowboy feature** — production code committed without a corresponding RED test.
- **Test passes before code exists** — assertion trivially true; the test proves nothing.
- **Happy-path-only coverage** — no edge, no error, no adversarial cases.
- **Designing after tests** — files keep appearing in unplanned locations; `utils.py` grows.
- **Implementing beyond the current RED test** — extra methods, configuration, retry/cache logic with no test driving them.
- **Tests asserting implementation details** — spying on internal method order; refactors will break the tests.
- **Refactoring while tests are RED** — no safety net.
- **Generic file names** — `utils`, `helpers`, `misc`, `common`, `shared`.
- **Logging anti-patterns** — colliding event names, PII/secrets in logs, log message duplicates the docstring.

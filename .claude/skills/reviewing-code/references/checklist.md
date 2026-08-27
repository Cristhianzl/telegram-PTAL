# Pre-submission checklist + perfect-score criteria + framework guides

The checklist a reviewer applies to verify completeness. Mirrors the author's pre-submission checklist so author and reviewer are looking at the same gate.

---

## Pre-submission checklist

```
CRITICAL (Blockers)
[ ] No PII in any logs, prints, or webhook messages
[ ] No secrets / credentials in code
[ ] No duplicate types, classes, or logic (DRY)
[ ] No file exceeds 500 lines (600–700 only if all other rules pass)
[ ] No file has more than 5 functions with DIFFERENT responsibilities
[ ] No file has more than 10 functions even with same responsibility
[ ] No file has more than 1 main class (grouped exceptions/DTOs OK)
[ ] No mixed responsibility prefixes in same file (validate* AND format* = VIOLATION)
[ ] No function exceeds cyclomatic complexity 10 (15 for legacy untouched-by-this-PR code)
[ ] No nesting depth exceeds 4 levels in any function

CRITICAL — Comprehension Audit (Blockers)
[ ] I can summarize each significant block in 1-3 sentences without re-reading the diff
[ ] I can answer "why this approach over the obvious alternative?" for every non-obvious choice
[ ] Every non-obvious design decision is captured durably (test name, Why: comment, PR ## Design Decisions, or ADR)
[ ] No "the AI wrote it, I trust it" code in the diff — every line is defensible
[ ] No "future AI will refactor this" assumption baked in anywhere

CRITICAL — AI-Generated Code in High-Risk Areas (Blockers, when applicable)
[ ] Every external API call verified against the provider's official documentation (not the AI's claim)
[ ] Every signature / HMAC / JWT verification compared line-by-line with the reference implementation
[ ] Every input validation covers what the SPEC requires, not just what the AI thought of
[ ] Every error path traced manually — no silent success on a failed check
[ ] No eval / exec / innerHTML / dangerouslySetInnerHTML / unsafe deserialization
[ ] A senior engineer (not the author, not the AI) reviewed the critical-path block

CRITICAL — AI / Chatbot Features Only (Blockers)
[ ] External integration follows the provider's official spec exactly
[ ] Signature comparison uses timing-safe function (not ===)
[ ] Timestamp / replay attack protection implemented
[ ] Signature verified BEFORE reading any payload field
[ ] Webhook secret in environment variable, not in code
[ ] AI / chatbot layer has NO direct database access
[ ] AI agent uses only tools / permissions explicitly required for its purpose
[ ] User input is validated and sanitized before reaching the LLM
[ ] System prompt contains NO secrets, credentials, or connection strings
[ ] LLM output is treated as untrusted (sanitized before render, never eval'd)
[ ] Irreversible agentic actions require human confirmation
[ ] DB user used by AI services has minimum required permissions
[ ] Authorization enforced in code (downstream functions), NOT delegated to AI judgment

CRITICAL — AI Runtime Resilience (Blockers, when feature calls an LLM in production)
[ ] Explicit timeout on every AI call (no SDK defaults)
[ ] Circuit breaker with explicit thresholds
[ ] Fallback path implemented AND tested under failure
[ ] Asynchronous by default unless latency budget explicitly justifies sync
[ ] Idempotency for state-changing operations
[ ] Cost ceiling defined with alerts
[ ] Kill switch implemented (no-redeploy disable)
[ ] Mandatory observability fields emitted (operation, request_id, model, tokens, duration, outcome)
[ ] SLO documented in PR description
[ ] No full prompts / completions logged verbatim

IMPORTANT (Must fix)
[ ] Each file / function has single responsibility (SRP)
[ ] No growing if/elif chains for type handling (OCP)
[ ] No subclass that breaks parent's contract (LSP)
[ ] No interface forcing useless implementations (ISP)
[ ] High-level modules depend on abstractions, not concretions (DIP)
[ ] No speculative code for future requirements (YAGNI)
[ ] Could this diff be smaller? Existing code / stdlib / platform feature / installed dependency instead of new code (reuse ladder)
[ ] No deep call chains reaching through object internals (Law of Demeter)
[ ] Proper error handling (no silent failures)
[ ] Strong typing (no any / object types)
[ ] Inputs validated at boundaries
[ ] Types in dedicated types file
[ ] Constants in dedicated constants file

IMPORTANT — Accessibility (when the diff touches UI; see building-frontend-ui/references/accessibility.md)
[ ] Keyboard operable both directions (Tab AND Shift+Tab — no trap); visible focus; focus lands somewhere useful after open/close
[ ] Semantic elements; labels on controls; alt on images; aria-label on icon-only buttons; visible label matches accessible name
[ ] No color-only meaning; composite widgets (grid/tree/menu) are one tab stop with roving tabindex
[ ] If the project has a11y scanners, they ran green on the changed states (not just default render) — all engines, not just one

IMPORTANT — i18n (when the repo has a locales/ dir or i18n config)
[ ] No hardcoded user-facing strings in the diff — new UI text goes through the translation function
[ ] Every new translation key exists in ALL locale files (missing in one language = broken UI for that language)
[ ] Removed/renamed keys cleaned up in all locales (no orphans)

IMPORTANT — Data layer & scale (when it touches a DB; see developing-features/references/data-layer.md)
[ ] Connection pooling with explicit limits (not a connection per request)
[ ] Index on every foreign key and on columns used in WHERE / JOIN / ORDER BY, added in the same migration
[ ] No N+1 queries (eager/batch load), and no query inside a loop
[ ] Every list endpoint is paginated
[ ] Transactions kept short; a query/statement timeout is set

RECOMMENDED (Should fix)
[ ] Appropriate logging at key points
[ ] No unnecessary comments
[ ] No over-engineering (no files with 1-2 trivial functions)

TESTING (see writing-tests skill)
[ ] Unit tests for core logic
[ ] Tests cover success, error, and edge cases
[ ] Tests have BOTH happy path AND adversarial tests
[ ] Adversarial tests: edge cases, invalid inputs, boundary values, error states
[ ] If feature calls an LLM: AI runtime failure-mode tests (timeout, malformed output, circuit-open, kill switch)
[ ] Coverage ran and shown — ≥75% minimum (target 80%) for ALL created tests
[ ] All created tests pass — zero failures
[ ] Not prolonging legacy bad patterns
```

---

## Perfect-score criteria

To achieve a perfect review score, the code must:

1. Zero PII in logs — verified by searching for email, name, phone patterns.
2. Zero security issues — no hardcoded secrets, proper input validation.
3. Zero DRY violations — no duplicate logic, types, or classes.
4. Zero file structure violations — ≤500 lines (600–700 only if all other rules pass), ≤5 functions (different responsibilities), ≤10 (same responsibility), ≤1 main class.
5. Zero mixed responsibilities — no different function prefixes in the same file.
6. Zero over-engineering — no unnecessary file fragmentation.
7. Zero SOLID violations — SRP, OCP, LSP, ISP, DIP all verified.
8. Zero pragmatic principle violations — YAGNI, KISS, Law of Demeter respected.
9. Clean architecture — proper layer separation.
10. Proper error handling — explicit, meaningful, no silent failures.
11. Strong typing — no loose types.
12. Appropriate tests — core logic covered with BOTH happy path AND adversarial tests.
13. Coverage validated — coverage ran and shown, ≥75% (target 80%) for all created tests.
14. AI security compliant *(when applicable)* — all AI guardrails verified.
15. Cyclomatic complexity gate passed — no function above CC 10 (15 for legacy untouched), nesting ≤ 4.
16. Comprehension Audit passed — author can defend every block; non-obvious decisions captured durably.
17. AI-untrusted-input lens applied *(when AI-generated code is in critical paths)* — every assumption verified against the authoritative source.
18. AI runtime resilience verified *(when feature calls an LLM in production)* — timeout, circuit breaker, fallback (tested), idempotency, cost ceiling, kill switch, observability, SLO.

---

## Framework-specific guidelines

### React / Frontend

```
featureName/
├── feature-name.tsx              # Orchestrates the feature (NOT index.tsx)
├── feature-name.types.ts         # Types for this context only
├── feature-name.constants.ts     # Constants for this context only
├── hooks/                        # One responsibility per hook
│   ├── use-feature-data.ts       # Responsibility: fetch / manage data
│   └── use-feature-actions.ts    # Responsibility: handle user actions
├── helpers/                      # Pure functions (no React dependencies)
│   ├── validation.ts             # ONLY validate* functions
│   └── formatting.ts             # ONLY format* functions
└── components/                   # Presentational sub-components
    ├── feature-header.tsx
    └── feature-list-item.tsx
```

**Anti-patterns:**

- `index.tsx` as main component file.
- Component that fetches, transforms, validates, AND renders.
- Business logic inside render.
- Giant `useEffect` doing multiple things.
- `validate*` and `format*` functions in the same file.

---

### Python / Backend

```
feature/
├── feature_service.py            # Orchestrates domain logic (≤500 lines)
├── feature_types.py              # Domain models (Pydantic, dataclasses)
├── feature_constants.py          # Domain constants
├── helpers/
│   ├── validation.py             # ONLY validate*, check*, is_valid* functions
│   └── formatting.py             # ONLY format*, build*, serialize* functions
├── repositories/
│   └── feature_repository.py     # ONLY save*, load*, find*, query* functions
└── handlers/
    └── feature_handler.py        # ONLY request handling
```

---

### NestJS / Node Backend

```
feature/
├── feature.module.ts             # Wires dependencies
├── feature.controller.ts         # HTTP layer (≤500 lines)
├── feature.service.ts            # Domain logic (≤500 lines)
├── feature.types.ts              # DTOs, domain interfaces
├── feature.constants.ts          # Domain constants
├── helpers/
│   ├── validation.helper.ts      # ONLY validate* functions
│   └── formatting.helper.ts      # ONLY format*, build* functions
└── repositories/
    └── feature.repository.ts     # ONLY data access functions
```

---

## When a finding doesn't fit a category

Default to `I` (Important) when uncertain — escalate to `B` only if the issue is genuinely blocker-grade (PII leak, security defect, file-structure violation, missing tests on a high-risk path). Demote to `R` if it's purely a style or readability suggestion.

If a finding doesn't fit any category, it might not be a finding at all — it might be a preference. Drop it unless the user explicitly asked for opinionated feedback.

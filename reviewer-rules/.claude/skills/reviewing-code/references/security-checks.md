# Security checks during review

The blocker categories. Apply this lens to every PR; the categories are zero-tolerance even when the PR is "obviously fine".

---

## PII in logs — zero tolerance

The #1 review priority. Any PII in a `logger` / `print` / `console.log` call is an immediate `B` finding.

**Forbidden** in any log call:

- Email addresses (`user.email`, `body.email`, `request.email`, any `*email*`)
- User names (`user.first_name`, `user.last_name`, `user.full_name`, `user.name`)
- Phone numbers (`user.phone`, `user.phone_number`)
- Addresses (`user.address`, `user.location`)
- Raw request bodies that may contain any of the above

**Approved identifiers** for logs:

| Identifier                | When to use                       |
|---------------------------|-----------------------------------|
| `user.auth_id`            | Primary — always prefer this      |
| `user.user_id`            | Acceptable                        |
| `user.stripe_id`          | Acceptable for payment context    |
| `user.id`                 | Acceptable if it's an internal ID |

Wrong vs right by language:

```python
# Wrong
logger.info(f"User {user.email} logged in")
logging.warning(f"Failed for {body.email}")
print(f"Contact sent: {data}")  # if data contains email

# Right
logger.info(f"User auth_id={user.auth_id} logged in")
logger.warning("Failed login", extra={"auth_id": user.auth_id})
```

```typescript
// Wrong
console.log(`User ${user.email} logged in`);
logger.warn(`Failed for ${body.email}`);

// Right
console.log(`User auth_id=${user.authId} logged in`);
logger.warn('Failed login', { authId: user.authId });
```

```java
// Wrong
logger.info("User {} logged in", user.getEmail());

// Right
logger.info("User authId={} logged in", user.getAuthId());
```

---

## General security checks

- All user inputs sanitized and validated at system boundaries.
- No secrets / credentials in code (env vars or secret managers).
- No internal details exposed in error messages to end users.
- External / untrusted data never trusted without validation.
- SQL queries use parameterized statements (no string concatenation).
- No hardcoded API keys, tokens, or passwords.
- Sensitive data not stored in plain text.
- HTTPS / TLS for all external communication.
- Rate limiting implemented on all public-facing endpoints.
- Passwords stored with bcrypt / Argon2 / scrypt — never MD5 / SHA-1.
- Sensitive data at rest (PII, credentials, payment info) encrypted.

---

## AI / chatbot / LLM security

> If the feature includes any LLM, chatbot, or AI agent, **all** items below are blockers. Aligned with OWASP Top 10 for LLM Applications 2025.

- AI / chatbot layer has **no direct database access** — all data flows through a service / repository layer that enforces its own access control.
- AI agent has access only to tools explicitly required for its stated purpose.
- Authorization is enforced **in downstream functions** (code), NOT delegated to the AI's own judgment.
- Database user / role used by AI-accessible services has read-only permissions unless write access is explicitly justified.
- User input is validated and sanitized **before** being forwarded to the LLM.
- System prompt and user input are separated at the architecture level (not string concatenation).
- Rate limiting per user / session.
- Known prompt-injection patterns detected and blocked.
- LLM output treated as **untrusted** — not rendered as raw HTML without sanitization.
- Output is scanned / filtered before use (SQL fragments, shell commands, PII patterns).
- If AI-generated code is executed, it runs in a **sandboxed environment** — never `eval()` / `exec()`.
- **No secrets, credentials, or connection strings in the system prompt.**
- System prompt is NOT used as an authorization boundary.
- Irreversible agentic actions require human confirmation.
- Every tool call by the agent is audited (logged with parameters and result).
- Execution time limits and token consumption limits set on agent runs.
- Repeated validation failures from the same source are logged and alerted (>5 failures in 10 min = probe attempt).

---

## Third-party integration security

> Any PR that integrates with an external service MUST pass all checks below. Implementing a generic pattern instead of the provider's exact specification is a security bug.

- Reviewer **verified the implementation against the provider's official documentation**.
- Signature / authentication format matches exactly what the provider specifies.
- Timestamp validation implemented to reject replayed requests (typical tolerance: 5 minutes).
- **Timing-safe comparison** used for signature comparison — never `===` / `==` / string comparison.
- Raw request body passed to the HMAC function — not a re-serialized version.
- Signature verified **before** any payload field is read.
- Failed signature verification throws / returns false — no silent pass on exception.
- Webhook secret stored in environment variable — not in code.
- Integration tested with real payloads from the provider's sandbox.

---

## AI-generated code in high-risk areas

> Treat AI-generated code in security-critical paths as **untrusted input**. The vulnerability lives in the assumption embedded in the code, not the syntax.

### High-risk areas

- Authentication, session, token handling
- Authorization checks and ownership filters
- Payment and financial transaction code
- Webhook signature verification, HMAC, JWT
- Cryptographic operations (signing, hashing, key derivation, encryption)
- Service / trust boundaries
- AI / LLM input/output guardrails (the guardrails themselves)
- File upload handling, deserialization (pickle, YAML, etc.)
- SQL query construction (verify the ORM is not bypassed by raw fragments)
- Anything touching secrets, credentials, or PII

### Mandatory checks (in addition to standard security)

- Every external API call has been verified against the **provider's official documentation** — not blog posts, not Stack Overflow, not the AI's claim about how the API works.
- Every signature / HMAC / JWT verification has been compared, line by line, against the provider's reference implementation or official SDK.
- Every input validation covers the case set the **spec** requires, not just the cases the AI happened to think of.
- Every error path traced manually: does it leak details? Does a failed check silently succeed (e.g., `except: return True`)?
- Every assumption embedded in a comment, default value, or constant has been challenged: is this actually true, or did the AI guess?
- No `eval`, `exec`, raw `innerHTML`, `dangerouslySetInnerHTML`, or unsafe deserialization in the diff.
- A senior engineer (not the author, not the AI) has reviewed the critical-path block with this lens applied.

### Where the lens does NOT apply

Trivial CRUD on non-sensitive data, internal tooling not exposed to users, prototypes behind feature flags not yet enabled in production. The lens is targeted, not universal — applying it to everything devalues it.

---

## AI runtime resilience (when feature calls an LLM in production)

> A call to an LLM is a network call to an unreliable, slow, non-deterministic external service. Treating it as a function call is a production incident waiting to happen.

All items below are blockers:

- Explicit timeout on every AI call (no SDK defaults).
- Circuit breaker with explicit thresholds (e.g., open after 5 failures in 30s).
- Fallback path implemented **and tested** under failure (kill the LLM endpoint, run the feature, confirm graceful degradation).
- Asynchronous by default unless the latency budget explicitly justifies a sync call.
- Idempotency for state-changing operations (idempotency keys or safe-to-retry semantics).
- Cost ceiling defined (tokens per request, requests per user/day, daily cap) with alerts.
- Kill switch (config flag or feature flag that disables the AI path without a redeploy).
- Observability emits mandatory fields: `operation`, `request_id`, `model`, `prompt_token_count`, `completion_token_count`, `duration_ms`, `outcome`, `error_class`.
- SLO documented in the PR description (P95 latency, success rate, fallback rate).
- No full prompts / completions logged verbatim — they leak PII and secrets. Redact, sample, or hash.
- No retry-forever loops — bounded retries with exponential backoff and a dead-letter for failures.

---

## Comprehension audit (apply to every PR)

> A PR where the author cannot defend every significant block is a PR you must block.

For every non-trivial block in the diff:

- The author can summarize the change in 1–3 sentences without scrolling.
- The author can answer "why this approach over the obvious alternative?" — and the answer is captured durably (test name, `Why:` comment, PR `## Design Decisions`, ADR), not just in the author's head.
- The author can name which test would fail if a given block were removed.
- The author can articulate the failure mode if the inputs were malformed or the dependency went down.
- No block is "the AI wrote it, I trust it." Every line is something the author can defend.

### Red flags

| Red flag                                                                       | What it usually means                                                              |
|--------------------------------------------------------------------------------|------------------------------------------------------------------------------------|
| Author can't summarize the diff in 1–3 sentences without scrolling             | They don't own the change — they hosted an AI session that produced it             |
| Generic test names: `test_create_user_works`, `test_handles_error`             | Intent was not encoded — future maintainers can't reconstruct it                   |
| Patterns inconsistent with the rest of the codebase, no explanation             | Either a deliberate departure (capture the reason) or AI-introduced inconsistency  |
| PR description says "AI-generated" with no design rationale                    | Author skipped the comprehension step                                              |
| `Why:` comments missing from non-obvious blocks                                | Architectural intent was not captured — lost the moment the author's memory fades  |
| "The tests pass, what more do you want?"                                       | Tests prove the *what*. The audit is about the *why*. Different things.            |
| "Future AI will clean it up"                                                   | LLMs can't reconstruct intent that was never captured                              |

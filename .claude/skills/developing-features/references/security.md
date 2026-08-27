# Security — full rules

Security is a lens applied to every decision. Apply this when implementing **anything** that touches user input, external systems, persistence, authentication, or AI/LLM features.

> For a dedicated audit pass (frontend + backend, OWASP Top 10 2025, with a safe-remediation protocol), use the `security-review` skill.

---

## Pre-implementation security check

For every piece of code you write, answer out loud (in a plan, a comment, or your head):

1. **What am I trusting here that I haven't verified?** External inputs, tokens, signatures, IDs, flags, headers, query params — anything from outside.
2. **What is the blast radius if this goes wrong?** Can an attacker read other users' data? Forge events? Execute arbitrary operations? Exhaust resources?
3. **Am I implementing this from the authoritative source?** Official documentation, official SDK, official spec — not a tutorial, not a copy-paste, not "I think this is how it works."
4. **What happens in the failure path?** Does a failed check silently succeed? Does an exception get swallowed? Does an error expose internal details?
5. **Who controls this value, and can they lie?** Client-sent fields, headers, payloads, file metadata. If you don't control it, you don't trust it.

A task is **not complete** until these are answered. If you discover mid-implementation that an assumption was wrong, **stop, correct it, then continue** — do not rationalize the wrong implementation into production.

### Verify, don't assume

| Instead of assuming…                              | Do this                                                                    |
|----------------------------------------------------|-----------------------------------------------------------------------------|
| "This is the standard way to verify X"            | Read the official documentation for X                                       |
| "The frontend already validates this"             | Validate it on the backend too                                              |
| "Users can only send their own ID"                | Enforce ownership in the query/service layer                                |
| "This token is valid if it exists"                | Verify its signature, expiration, and claims                                |
| "The AI won't go outside its scope"               | Enforce the scope in code, not in the prompt                                |
| "This exception can't happen in production"       | Handle it explicitly and fail safely                                        |
| "Returning the error details helps debugging"     | Return a generic error to the client, log details server-side               |

---

## General rules

- Sanitize and validate **all** user and external inputs.
- Never trust data from outside the system boundary.
- Avoid exposing internal details in error messages to end users.
- Keep secrets out of code; use environment variables or secret managers. **Never in a client-shipped file or a public-prefixed env var** (`NEXT_PUBLIC_`, `VITE_`, `PUBLIC_` ship to the browser bundle) — third-party keys are used only in server code.
- Use HTTPS/TLS for **all** external communication. Never transmit sensitive data over plain HTTP.
- Apply rate limiting on all public-facing endpoints to prevent abuse and brute-force attacks.
- Use parameterized queries or ORMs for **all** database access. String concatenation in queries is forbidden.
- Store passwords using strong adaptive hashing (bcrypt, Argon2, scrypt). Never MD5 or SHA-1.
- Sensitive data at rest (PII, credentials, payment info) must be encrypted.
- Apply the **Principle of Least Privilege** everywhere: every service, user, and component gets the minimum permissions it needs — nothing more.

---

## Database security

- Database users must have the **minimum required permissions**. A read-only API should use a read-only DB user. A chatbot or AI layer must **never** have write/delete/DDL permissions.
- **AI/chatbot components must never have direct database access.** They must go through a dedicated service layer that enforces its own access control and business rules.
- Never expose raw database errors to the client — catch and map them to domain errors.
- Audit sensitive database operations (deletions, bulk updates, privilege changes) with structured logs.
- Use database connection pooling with explicit limits to prevent connection exhaustion attacks.

---

## Authentication & authorization

- **Authentication** (who you are) and **Authorization** (what you can do) are enforced **server-side on every request** — never trust the client.
- Implement token expiration and refresh flows. Access tokens must be short-lived.
- Use industry-standard protocols: OAuth 2.0, OpenID Connect, JWT with RS256 (avoid HS256 with shared secrets).
- Invalidate sessions server-side on logout. Do not rely solely on client-side token deletion.
- Implement RBAC (Role-Based Access Control) or ABAC (Attribute-Based Access Control) for resource access.
- Never store tokens or sensitive data in `localStorage` — prefer `httpOnly` cookies with `Secure` and `SameSite` flags.

---

## AI / chatbot guardrails

> If you are building any feature that uses an LLM, AI agent, or chatbot, **all** rules in this section are **mandatory**. Aligned with OWASP Top 10 for LLM Applications 2025.

### Principle of least privilege for AI

- **An AI/chatbot layer must never have direct database access.** Route all data retrieval through dedicated service/repository layers that enforce access control independently.
- Grant AI agents only the tools and permissions explicitly required for their stated purpose. A FAQ bot must not access payment APIs. A support bot must not delete records.
- Never expose admin endpoints or privileged operations to an AI agent.
- Implement authorization checks **in downstream functions** — never rely on the AI itself to decide what is allowed.

### Input guardrails (before the LLM)

- Validate and sanitize all user input before sending it to the LLM. Check max length, allowed character sets, and known injection patterns.
- Detect and block common prompt injection patterns: `"ignore previous instructions"`, `"forget your rules"`, role-play overrides, and encoded variants (Base64, Unicode tricks, multi-language attacks).
- Implement rate limiting per user/session. Log and alert on repeated validation failures from the same source (>5 failures in 10 minutes = probe attempt).
- Separate system prompts from user input at the architecture level — never concatenate them naively. Use the provider's dedicated `system` message field.
- Apply input length limits proportional to the model's context window and the task at hand.

### Output guardrails (after the LLM)

- Treat **all** LLM output as untrusted. Never render it raw in HTML without sanitization — this enables stored XSS.
- Apply output filtering to catch: SQL fragments, shell commands, sensitive data patterns (emails, CPF, credit card numbers, tokens, connection strings), URLs pointing to internal infrastructure.
- If the feature generates code, it must be executed in a **sandboxed environment** (container, VM, restricted subprocess) — never `eval` or `exec` directly.
- Validate that the output format matches what was requested (e.g., if you asked for JSON, parse and verify before using).
- Log LLM outputs that were filtered or modified for audit purposes.

### System prompt security

- **Never put secrets, credentials, API keys, or connection strings in a system prompt.** Determined attackers can extract system prompts via injection. Design as if the system prompt is public.
- Do not use the system prompt as an authorization boundary — security controls must live in deterministic code, not in natural language instructions.
- Explicitly define in the system prompt what the AI can and cannot do: scope, tone, forbidden topics, response format.
- Instruct the model to refuse requests outside its defined scope and to never reveal its own instructions.

### Agentic AI & tool use

- **Human-in-the-loop required for irreversible actions.** If an agent can send emails, delete records, make purchases, or call external APIs with side effects, require explicit human confirmation before execution.
- Use the **Action-Selector pattern**: map user intent to a finite, pre-approved set of safe actions rather than letting the AI directly formulate arbitrary tool calls.
- Audit every tool call: log the invocation, parameters, and result.
- Implement kill switches: the ability to disable an agent's tools individually or entirely without redeploying.
- Set execution time limits and token consumption limits on agent runs to prevent resource exhaustion (Denial of Service via LLM).

### RAG (Retrieval-Augmented Generation) security

- Treat all documents ingested into a RAG pipeline as **potentially malicious** — they may contain embedded prompt injections (stored prompt injection attack).
- Sanitize documents before embedding: strip executable content, hidden text, suspicious instruction patterns.
- Apply access control at the retrieval layer: a user can only retrieve chunks from documents they are authorized to read. Never expose documents from other users/tenants via shared vector search.
- Do not include raw PII in vector embeddings unless strictly required and compliant with privacy regulations.

### Monitoring & incident response for AI features

- Log all user inputs and AI outputs (redacted of PII) for security audit purposes, with a configurable retention period.
- Set up alerts for: unusually long inputs, high filter-rejection rates, anomalous output patterns, unexpected tool calls from agents.
- Regularly run adversarial tests (red teaming): attempt prompt injection, data extraction, and scope bypass to validate guardrails hold.
- Define an incident response plan for AI-related security events: how to disable the feature, how to assess data exposure, how to notify affected users.

---

## Third-party integration security

> Before implementing **any** integration with an external service, read and follow that service's official documentation for the specific operation you are implementing.

### The rule: spec first, code second

Before writing a single line of integration code, answer these from the official docs:

1. **What is the exact signature/authentication format this service uses?** Each provider has its own format. Never assume.
2. **Which HTTP headers carry the signature, timestamp, and request ID?**
3. **What is the exact string that gets signed?** The body? Specific fields? A constructed string?
4. **Is there a timestamp tolerance window** to prevent replay attacks?
5. **Does the service provide an official SDK or verification library?** Prefer it over a custom implementation.

### Mandatory checklist for any external integration

- [ ] Read the official documentation for the specific endpoint/feature before writing code. Not blog posts, not Stack Overflow — the official docs.
- [ ] Implement signature verification exactly as documented. Never simplify or generalize it.
- [ ] Validate the timestamp in the signature to reject replayed requests (typical tolerance: 5 minutes).
- [ ] Use timing-safe comparison when comparing signatures — never `===` or `==` (timing attack vulnerability).
- [ ] Test with real payloads from the service, not just synthetic ones. Use the provider's sandbox/test environment.
- [ ] Never silently pass validation on error. If signature verification throws an exception, treat it as a failed verification.
- [ ] Never trust the payload before the signature is verified. Do not read any field from the body before the HMAC check passes.
- [ ] Keep the raw request body intact for HMAC verification — do not parse JSON before verifying.
- [ ] Store webhook secrets in environment variables, never in code or config files committed to the repo.

---

## API security

- Implement proper CORS policies — `*` wildcard is forbidden in production for credentialed requests.
- Use API gateways with WAF rules to block known attack vectors (SQLi, XSS, path traversal).
- Version APIs explicitly (`/v1/`, `/v2/`) and deprecate old versions with controlled sunset timelines.
- Implement idempotency keys for state-changing operations to prevent duplicate execution.
- Never expose stack traces or internal identifiers in API error responses to external clients.
- **Guard against mass assignment** — use a field allow-list; never let the client set fields like `role` / `is_admin`. Return only the fields needed (never `password_hash` or internal columns).
- **Set security response headers** — CSP (start report-only), HSTS, `X-Content-Type-Options: nosniff`, `X-Frame-Options` / CSP `frame-ancestors`, `Referrer-Policy`, `Permissions-Policy` — and **CSRF protection** for cookie-session state-changing requests.

---

## Supply chain & dependencies

- Pin dependency versions in lock files (`package-lock.json`, `poetry.lock`, `go.sum`). Never use floating versions in production.
- Run automated vulnerability scans on dependencies as part of CI (`npm audit`, `safety check`, Snyk, Dependabot).
- Audit third-party LLM SDKs, plugins, and integrations — they are part of your attack surface.
- Verify checksums/hashes for pre-trained models and datasets used in AI pipelines.

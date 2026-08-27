# API security

The non-negotiable security floor for any API surface. Organized around the [OWASP API Security Top 10](https://owasp.org/API-Security/editions/2023/en/0x11-t10/). This complements the broader `developing-features` security reference — the API is one boundary.

## Transport & authentication

- **TLS only.** Reject plaintext; redirect or refuse `http://`. Use HSTS.
- **Use OAuth2 / OIDC** (or a vetted token scheme); don't invent your own auth.
- **Validate tokens server-side on every request** — signature, issuer, audience, expiry. Never trust client-asserted identity, roles, or scopes without verification.

## Authorization — the top two API risks

- **Object-level authorization (OWASP API #1, BOLA).** On every request that names an object, verify the caller is allowed to act on **that specific object** — not merely that they're authenticated. `GET /orders/{id}` must check the order belongs to the caller. This is the most common and most damaging API flaw.
- **Function-level authorization (OWASP API #2, BFLA).** Verify the caller may invoke this operation at all (e.g. admin-only endpoints). Don't rely on the UI hiding the button.
- **Scopes for coarse access, per-object checks for fine access.** Scopes gate *which kinds* of operations; object checks gate *which instances*.
- **Least privilege.** Tokens, service accounts, and DB users get the minimum rights the endpoint needs.

## Data handling

- **No sensitive data in URLs.** Tokens, secrets, PII, and the like go in headers or the body — URLs leak into access logs, proxies, browser history, and `Referer`.
- **Avoid mass assignment.** Bind requests to an explicit allow-list of writable fields; never blindly map the request body onto your model (a client could set `is_admin`, `balance`, `owner_id`).
- **Avoid excessive data exposure.** Return only the fields the consumer needs; don't dump the whole object and rely on the client to hide fields. Don't leak internal IDs, structure, or other users' data.

## Input validation

- **Validate every input against the schema** at the boundary: types, formats, enums, **max lengths and sizes**, ranges. Reject, don't coerce silently.
- Bound payload and collection sizes; cap upload sizes; reject unexpected fields where strictness matters.

## Abuse resistance

- **Rate limiting and quotas.** Return `429` with `Retry-After` and limit headers (e.g. `RateLimit-Limit`/`RateLimit-Remaining`). Protects against brute force, scraping, and resource exhaustion (OWASP API #4).
- **CORS allow-list.** Enumerate allowed origins explicitly; **never** `Access-Control-Allow-Origin: *` together with credentials.
- **Don't leak internal structure** in errors (see `errors-idempotency-concurrency.md`) — no stack traces, SQL, or framework internals to clients.

# OWASP Top 10 (2025) checklist — frontend & backend

A review checklist based on OWASP Top 10 **2025** + OWASP API Top 10. Items marked **↗** are already enforced elsewhere — **verify** them, don't re-derive: the always-on review blockers in `../../reviewing-code/references/security-checks.md`, the build-time rules in `../../developing-features/references/security.md`. This file adds the breadth and the gaps.

## CRITICAL — do first

**Secrets**
- No service-role / admin / DB keys in frontend code or in any **public-prefixed env var** (`NEXT_PUBLIC_`, `VITE_`, `PUBLIC_`) — those ship in the browser bundle. Grep the repo for `KEY`/`SECRET`/`TOKEN`/`PASSWORD`/`service_role`, and search the built bundle (DevTools → Sources) for the same strings.
- Third-party API keys (Stripe, OpenAI, email) used **only in server code**. `.env` / `.pem` / creds are gitignored and absent from git history. If a secret ever leaked → **rotate the key** (removing it from code isn't enough — it's in history).

**Broken access control (A01) ↗**
- Authorization is server-side on every request; hiding a button is not access control.
- **IDOR/BOLA:** every endpoint that takes an ID also filters by owner (`WHERE id = :id AND user_id = :session`). Test: as user A, request user B's resource — if it returns data, that's a finding.
- **RLS** (Postgres/Supabase): Row Level Security enabled on every user-data table **with explicit policies** (RLS on + no policy = nobody sees anything; RLS off = everybody sees everything). No `SECURITY DEFINER` view/function silently bypassing policies.

**Injection (A05)**
- SQL: parameterized / ORM only, zero string-built SQL. ↗
- XSS: use the framework's escaping; no `dangerouslySetInnerHTML` / `v-html` / `innerHTML` with user content (sanitize with DOMPurify if truly needed). ↗
- The rest of the family — check what applies to the stack: **NoSQL** (reject operators like `$ne`/`$gt`/`$where`; validate types; never pass a raw request object into a query), **CSV/formula** (prefix fields starting with `= + - @` on export), **CRLF/header** (strip `\r\n` from anything placed in HTTP or email headers), **SSTI** (user input is *data* passed to the template, never part of the template string), **LDAP/XPath** (escape per interpreter), **GraphQL** (depth + complexity limits, introspection off in prod), **prompt injection** (see `../../developing-features/references/untrusted-content.md`).

**Password storage (A04) ↗**
- bcrypt / argon2id / scrypt with a per-password salt — never MD5 / SHA-1 / plain SHA-256, never reversible "encryption". Prefer a managed auth provider. Migrate old hashes **on next successful login** (transparent re-hash), never in bulk.

## HIGH

**Auth & session (A07)**
- Rate limit / lockout on login, password reset, and signup. 2FA/MFA available (ideally required for admins).
- JWT: short expiry, signed, **no sensitive data in the payload** (it's readable, only signed). Session cookies `HttpOnly` + `Secure` + `SameSite`. ↗ Logout invalidates the session server-side. ↗
- Password reset: single-use, expiring token, invalidated after use.
- **No user enumeration:** login and forgot-password return the same message whether or not the account exists ("If this email is registered, we'll send a link").

**Crypto in transit & at rest (A04) ↗+**
- HTTPS everywhere with HTTP→HTTPS redirect; **HSTS** enabled. Sensitive data encrypted at rest. `Cache-Control: no-store` on responses carrying sensitive data. Don't store data you don't need.

**Security response headers (A02)** — set at the server / proxy / CDN:
- **CSP** — the strongest XSS defense; **start in `Content-Security-Policy-Report-Only`**, watch the reports, then enforce.
- HSTS · `X-Content-Type-Options: nosniff` · `X-Frame-Options: DENY` (or CSP `frame-ancestors`) · `Referrer-Policy: strict-origin-when-cross-origin` · `Permissions-Policy` (disable unused browser APIs like camera/mic/geo).

**CORS & CSRF**
- `Access-Control-Allow-Origin` is an explicit allow-list (never `*` with credentials); don't reflect an arbitrary `Origin`. ↗
- State-changing requests authenticated by a **session cookie** need CSRF protection (a CSRF token or strict `SameSite`). APIs authenticated by an `Authorization` header token are naturally less exposed — confirm which you are.

**Supply chain (A03) ↗**
- Run `npm`/`pnpm audit` / `pip-audit` / equivalent; fix high/critical. Lockfile committed. Remove unused deps. Enable Dependabot/Renovate. Consider an SBOM. Update one dep at a time; prefer security patches over major upgrades.

## MEDIUM

**Misconfiguration (A02):** debug mode off in prod; no default credentials/keys; admin panel behind auth and not at an obvious route; directory listing, backup files, and `.git` not publicly accessible; unused ports/services closed.

**Error handling (A10) ↗:** generic error to the user (no stack trace / SQL / file path / framework version); the detailed log stays server-side; the app doesn't crash or enter an insecure state on unexpected/null/malformed input; **fail closed** — if an authorization check errors, deny.

**Software / data integrity (A08) ↗:** don't deserialize untrusted data without validation (RCE risk); **SRI** (Subresource Integrity hash) on CDN scripts; verify third-party updates; CI/CD deploy access controlled.

**File upload:** validate type and size **server-side** (never trust the extension or the sent `Content-Type`); store outside the web root (or in dedicated storage) with no execute permission; sanitize filenames (block `../` path traversal); re-encode images to drop embedded payloads.

**API (OWASP API Top 10):** BOLA ↗; **mass assignment** — don't let a user set fields they shouldn't (`"role":"admin"` on a self-profile update); use a **field allow-list**. **Excessive data exposure** — return only the fields needed, never `password_hash` / `is_admin` / internal columns. Rate-limit APIs, not just login.

## Logging, alerting & response (A09)
- Log security events: logins (success + failure), password/email changes, denied access, admin actions. **Never** log secrets / tokens / cards. ↗ Protect logs from tampering; define retention.
- **Alert** on suspicious patterns — not just log. Keep a minimal incident-response plan (rotate keys, invalidate sessions, notify affected users).

## Privacy (LGPD / GDPR)
- A legal basis per data collected; clear consent (no pre-ticked boxes); **data minimization**; data-subject rights (access, correction, deletion, portability); a published privacy policy; retention limits; a breach-notification plan (to the regulator + affected people); vetted third-party processors (analytics, AI, email).

## Continuity
- Automated, **tested** backups (a backup never restored isn't a backup); encrypted and stored apart from production; point-in-time recovery where possible.

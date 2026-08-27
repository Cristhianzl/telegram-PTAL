---
name: security-review
description: Review an application's code for security across frontend AND backend and produce a prioritized, safe-to-apply remediation plan — OWASP Top 10 (2025), OWASP API Top 10, plus privacy (LGPD/GDPR). Use when the user asks for a "security review", "security audit", "harden this", "is this secure?", "check for vulnerabilities", "OWASP", "pentest-style review", or wants a focused security pass on a feature/PR/app (secrets, access control, injection, auth/session, headers/CORS/CSRF, file upload, supply chain, logging, privacy). Applies a strict anti-regression protocol so fixes don't break the app. For designing security up front use threat-modeling; for the always-on security blockers in a normal PR use reviewing-code.
license: MIT
---

# Security review

Audit the application's code for real vulnerabilities — frontend and backend — and hand back a prioritized, **safe-to-apply** plan. The goal is to shrink the attack surface **without changing what a legitimate user sees or breaking the app**.

## Read first (always)

List `learnings/` and read anything relevant. **Reuse, don't re-derive:**

- The **always-on blocker set** applied to every PR lives in `../reviewing-code/references/security-checks.md` (PII in logs, injection, HMAC/signature, the AI-generated-code lens). Apply it — don't re-list it.
- The **build-time prevention rules** live in `../developing-features/references/security.md` (auth/session, DB least-privilege, TLS, supply chain, and the full AI/LLM guardrails) and `../developing-features/references/untrusted-content.md` (prompt injection).
- This skill adds the **OWASP-2025 breadth**, the **frontend / infra / privacy gaps**, and the **safe-remediation protocol** (see `references/`).

## How to run a security review

1. **Scope it.** What are you auditing — a PR, a feature, the whole app? What's the stack (framework, DB, auth provider, does it use AI)? Read `AGENTS.md`/README and grep the code to map the inputs, auth checks, persistence, external calls, and any AI/LLM surface.
2. **Go by priority, top-down** (`references/owasp-2025-checklist.md`): **CRITICAL** (secrets, broken access control / IDOR, injection, password storage) → **HIGH** (auth/session, security headers, CORS/CSRF, crypto, supply chain) → **MEDIUM** (misconfiguration, error handling, file upload, API mass-assignment, logging/alerting) → **privacy (LGPD/GDPR)** → **continuity (backups)**.
3. **Report findings — don't blind-fix.** For each: severity, the exact risk, the concrete fix, and how to verify it caused no regression. Prove access-control findings (logged in as user A, try user B's resource).
4. **Fix under the anti-regression protocol** (`references/safe-remediation.md`): one change per branch/commit, backup before DB changes, test the legitimate flow before and after, report-only rollout for CSP/rate-limit, transparent re-hash for password migration, staging before prod, fail closed.

## The one principle under all of it

Never mix **user data** with **command/structure** — parametrize, escape, or allow-list at the right interpreter (DB, shell, template, browser, AI model). Enforce **authorization server-side on every request**, keep **secrets off the client**, and **fail closed**. Every rule below is an application of these.

## Output

A prioritized report: findings grouped **Critical / High / Medium**, each with *what's wrong · why it matters · how to fix safely · how to verify no regression*, ending with the smallest first step. This is a review — **never run git or deploy**; the human applies the fixes.

## Capture a learning

If you find a project-specific vulnerability class or a fix that needed a special rollout, append a `learnings/YYYY-MM-DD-slug.md` (or use `/learn`).

## See also

- `references/owasp-2025-checklist.md` — the full frontend + backend + privacy checklist (OWASP Top 10 2025 + API Top 10).
- `references/safe-remediation.md` — apply security fixes without breaking the app.
- `skills/threat-modeling` — design-time: what could go wrong *before* building.
- `skills/reviewing-code` — the always-on security blockers in a normal PR.
- `../developing-features/references/security.md` — build-time prevention; `../developing-features/references/untrusted-content.md` — prompt-injection guardrail.

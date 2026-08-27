---
description: Project security audit
allowed-tools: Read, Glob, Grep
---

# /security

Static and dependency audit. Use the `Grep` tool (not `bash grep`).

## 1. Dependencies

- Run the project's dependency audit (`<dependency audit command>`)
- Review reported vulnerabilities by severity

## 2. Secrets

Grep for:

- `password\s*=\s*["']`
- `api_key\s*=\s*["']`
- `(secret|token)\s*=\s*["'][^"']+["']`

Verify `.env*` is in `.gitignore`.

## 3. SQL injection

Grep for string-built SQL:

- `execute\(f["']` or `execute\(["'].*\{` (f-string in SQL)
- Query builders without bind params

OK: parameterized queries with bind params. Bad: f-string interpolation into SQL.

## 4. XSS

Grep `dangerouslySetInnerHTML`. Each occurrence must be sanitized.

## 5. Auth

List private endpoints without an auth dependency. Flag the ones that look private.

## 6. CORS / Headers

- `allow_origins=["*"]` in production = critical
- Security headers: CSP, X-Frame-Options, X-Content-Type-Options

## 7. Sensitive data

- Personal data (emails, IDs) in logs (Grep `logger.*email` and similar)
- Sensitive fields hashed/encrypted in models

## Report

```text
Security
Deps:    <✅ | ⚠️ N>
Secrets: ✅
SQLi:    ✅
XSS:     ✅
Auth:    ⚠️ N public endpoints
CORS:    ✅
Data:    ✅

Critical:
- ...

Recommendations:
- ...
```

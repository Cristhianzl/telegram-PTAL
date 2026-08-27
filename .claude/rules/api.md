---
description: API — enforced essentials for HTTP/REST endpoints (contract, resources, status, compatibility, errors, idempotency, security)
globs: "**/{api,routes,controllers,handlers}/**"
---

# API

> Prerequisite: `CLAUDE.md` (general baseline). Depth: `skills/api-design` (and its `references/`).

The short, non-negotiable subset. The skill carries the full reasoning, method/status tables, pagination, versioning, and security detail.

## Contract & resources

- **Contract-first.** Write/update the OpenAPI (or proto/SDL) spec before the code; lint it in CI.
- **Resources, not verbs, in paths.** `POST /users`, not `POST /createUser`. Pluralized collections; shallow nesting for ownership only.
- **Consistent conventions** — one casing for paths, one for fields, everywhere. Opaque IDs; never expose DB keys/structure.

## HTTP semantics

- **Correct status codes** — the status is the outcome. **Never `200 OK` with an error in the body.** `5xx` is server fault only.
- Correct method semantics (GET/HEAD safe+idempotent, PUT/DELETE idempotent, PATCH partial, POST create/non-idempotent).
- **Paginate every list from day one** (cursor preferred). Adding pagination later is a breaking change.

## Compatibility

- **Never break consumers.** Additive-only within a version; breaking changes need a new version.
- Renaming/removing a field, changing a type/default/format, or adding pagination to an unpaginated list are **breaking**.
- Deprecate, don't silently break — `Deprecation` + `Sunset` headers and a documented support window. Clients must ignore unknown fields/enum values.

## Errors & write safety

- **Standard errors** — `application/problem+json` (RFC 9457) with a stable machine-readable `code`; HTTP status matches the body. **Never leak stack traces.**
- **Idempotency-Key** for non-idempotent POST (store and replay the first response).
- **Optimistic concurrency** with `ETag` + `If-Match` → `412` on conflict.

## Security (floor)

- **Per-object authorization** on every request that names an object (BOLA), plus function-level checks (BFLA). Validate tokens server-side.
- **No secrets/PII in URLs.** Validate input against the schema (types, sizes, max lengths).
- **Rate-limit** (`429` + `Retry-After`); **CORS allow-list** — never `*` with credentials. Avoid mass assignment and excessive data exposure.

---
name: api-design
description: Design and review HTTP/REST (and choose between REST/GraphQL/gRPC) APIs that are contract-first, resource-oriented, backward-compatible, and secure by default. Use whenever the task touches an API surface — designing an API, adding or changing an endpoint or route, picking an HTTP method or status code, naming resources, paginating a list, versioning, deprecating a field, modeling errors, idempotency, concurrency, webhooks, or writing an OpenAPI spec. Triggers: "design an API", "add an endpoint", API, REST, endpoint, route, HTTP, versioning, pagination, webhook, OpenAPI, idempotency. Not for pure UI work — see building-frontend-ui.
license: MIT
---

# API Design

A good API is a **contract** you can evolve without breaking the people who depend on it. Design the contract first, model resources not actions, use HTTP the way it is specified, stay backward-compatible, and make security and operability the default — not an afterthought.

This skill is protocol-agnostic at the workflow level; the canonical detail lives in `references/` and is loaded on demand.

## Read first (always)

List `learnings/` and read every file whose name looks relevant to the current task — the project's versioning policy, its error envelope, its auth model, its naming conventions, and provider-specific webhook quirks live there and override the defaults in this file. If a learning conflicts with this SKILL.md, **the learning wins** — mention it to the user.

## Tradeoff — when to apply, when to lighten up

Apply the full discipline for any API with **external or cross-team consumers, or anything that persists state**. Lighten the ceremony for a private, single-consumer endpoint you fully control on both sides and can change atomically — but never lighten the security floor (authorization, input validation, no secrets in URLs).

## Workflow

1. **Choose the protocol deliberately.** REST for resource CRUD and broad client reach; GraphQL when clients need flexible, client-shaped queries over a graph; gRPC for low-latency internal service-to-service. Don't default to GraphQL/gRPC because they're new. State the choice and why.
   → verify: you can name the consumer and why this protocol fits it.

2. **Design contract-first.** Write the [OpenAPI](https://www.openapis.org/) (or proto/SDL) spec **before** the code; review it, then generate or hand-write handlers against it. Lint the spec in CI (e.g. Spectral) so drift fails the build.
   → verify: the spec exists and is the source of truth, not reverse-engineered from code.

3. **Model resources (nouns), not actions.** A resource is a thing with an identity; collections are plural. Unusual operations become sub-resources or state, not verbs in the path. See `references/rest-conventions.md`.
   → verify: no verb in any path.

4. **Use HTTP method + status-code semantics correctly.** GET/HEAD safe and idempotent, PUT/DELETE idempotent, PATCH partial, POST creates/non-idempotent. Status code carries the outcome; the body carries the data or the error. See `references/rest-conventions.md`.
   → verify: no `200 OK` with an error in the body.

5. **Apply consistent conventions everywhere.** One casing for paths, one for fields; pluralized collections; opaque IDs (don't leak DB sequence/structure). Inconsistency is a permanent tax on every consumer.
   → verify: a reader can predict the shape of an endpoint they haven't seen.

6. **Never break consumers.** Additive-only changes within a version; breaking changes need a new version and a deprecation window with `Deprecation`/`Sunset` headers. Adding pagination to a previously-unpaginated list **is** a breaking change. See `references/versioning-and-compatibility.md`.
   → verify: every change is classified additive or breaking before you ship it.

7. **Standardize errors.** Use [RFC 9457](https://www.rfc-editor.org/rfc/rfc9457) `application/problem+json` with a stable machine-readable `code`. HTTP status must match the body. Never leak stack traces. See `references/errors-idempotency-concurrency.md`.
   → verify: every error path returns the same envelope shape and a stable code.

8. **Paginate every list from day one.** Cursor-based (`page_size` + `page_token` / `next` link) preferred over offset. Retrofitting pagination later breaks clients.
   → verify: no unbounded list endpoint exists.

9. **Make writes safe to retry.** `Idempotency-Key` for non-idempotent POST (store and replay the first response); optimistic concurrency with `ETag` + `If-Match` → `412` on conflict. See `references/errors-idempotency-concurrency.md`.
   → verify: a client that retries a POST after a timeout cannot double-charge / double-create.

10. **Security by default.** TLS only; validate tokens server-side; per-object authorization (not just role); no secrets in URLs; validate input against the schema; rate-limit; CORS allow-list. See `references/security.md`.
    → verify: object-level authorization is checked on every request that names an object.

11. **Capture a learning (final step).** If you hit a convention, constraint, or trap not already in this file or `references/`, append a `learnings/YYYY-MM-DD-slug.md`. Otherwise skip.

## Anti-patterns (reject on sight)

- **Verbs in paths** — `POST /createUser`, `/getOrders`. Use `POST /users`, `GET /orders`.
- **`200 OK` with an error in the body.** The status code is the outcome; clients and proxies rely on it.
- **Leaking internal/DB models** — exposing column names, internal enums, sequence IDs, or join structure as the public shape. Map to a deliberate DTO.
- **Chatty endpoints / N+1** — forcing clients to make many round-trips to assemble one view. Offer expansion (`expand=`) or a coarser resource.
- **Breaking changes without a new version** — including renaming/removing a field, changing a type or default, or **adding pagination to a previously-unpaginated list**.
- **Inconsistent conventions** — mixed casing, singular and plural collections, sometimes-wrapped responses.
- **Non-idempotent POST with no idempotency key** — any retry double-creates.
- **Sensitive data in URLs** — tokens, PII, secrets in path or query; they land in logs, history, and referers.
- **Missing object-level authorization (BOLA)** — checking authentication but not whether *this* caller may touch *this* object. OWASP API #1.
- **Exposing stack traces** — internal class names, file paths, SQL in the error body.
- **Unbounded list endpoints** — `GET /events` returning everything.

## Capture a learning

When you finish, ask: *did I encounter a convention, constraint, or trap that wasn't in this SKILL.md or `references/`?* (A project-specific versioning policy, a non-standard error envelope the org has standardized on, a webhook provider's signing scheme, a mandated header.) If yes, append a `learnings/YYYY-MM-DD-slug.md` following `learnings/README.md`. If no, skip — don't write noise.

## See also

- `references/rest-conventions.md` — resource modeling, naming, methods, status codes, pagination, filtering, caching.
- `references/versioning-and-compatibility.md` — versioning strategies, additive vs breaking changes, deprecation policy.
- `references/errors-idempotency-concurrency.md` — problem+json, idempotency keys, optimistic concurrency, retries, long-running ops, webhooks.
- `references/security.md` — TLS, OAuth2/OIDC, BOLA/BFLA, input validation, rate limiting, CORS, data exposure.
- `developing-features` skill — general production rules and the full security reference (the API surface is one boundary among several).
- `reviewing-code` skill — review a diff that changes an API surface.
- `documenting-features` skill — capture the API contract, ADRs, and the domain model alongside the code.
- `learnings/` — project-specific API conventions accumulated over time.

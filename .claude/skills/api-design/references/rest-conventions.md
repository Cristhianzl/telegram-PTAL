# REST conventions — resources, methods, status, pagination, caching

The detailed defaults for HTTP/REST design. Sources: [Google AIPs](https://google.aip.dev), [Zalando RESTful API Guidelines](https://opensource.zalando.com/restful-api-guidelines/), [Microsoft / Azure REST API Guidelines](https://github.com/microsoft/api-guidelines).

## Resource modeling & naming

- **Collections are plural nouns.** `/users`, `/orders`, `/payment-intents`. An item is `/users/{id}`. ([Google AIP-122/132](https://google.aip.dev/132), [Zalando 134](https://opensource.zalando.com/restful-api-guidelines/#134))
- **Model things, not actions.** The URL identifies a resource; the method says what to do with it. No verbs in paths.
- **Shallow nesting for ownership only.** `/users/{id}/orders` is fine when an order belongs to a user; don't nest more than one level deep — prefer `/orders?user_id={id}` for filtering. Deep nesting (`/a/{}/b/{}/c/{}/d`) couples clients to your hierarchy.
- **Consistent casing.** Pick one for paths (kebab-case is common: `/payment-intents`) and one for JSON fields (`snake_case` or `camelCase`) and never mix. ([Zalando 118/129](https://opensource.zalando.com/restful-api-guidelines/#118))
- **Opaque IDs.** Treat IDs as opaque strings; don't expose auto-increment DB keys or encode internal structure that leaks volume or lets enumeration. Clients must not parse them. ([Zalando 144](https://opensource.zalando.com/restful-api-guidelines/#144))

## HTTP methods

| Method | Purpose | Safe | Idempotent | Body |
|--------|---------|------|-----------|------|
| GET | Read a resource/collection | yes | yes | no request body |
| HEAD | GET headers only | yes | yes | no |
| POST | Create / non-idempotent action | no | no | yes |
| PUT | Create-or-replace at a known URI | no | yes | full representation |
| PATCH | Partial update | no | no* | partial / patch doc |
| DELETE | Remove a resource | no | yes | usually none |

*PATCH is not inherently idempotent; make it so where you can. **Safe** = no observable side effects. **Idempotent** = repeating the same request has the same effect as doing it once.

### Unusual actions → resources, not verbs

When an operation doesn't map to CRUD (cancel, publish, search, batch), model it as state or a sub-resource, not a verb:

- Cancel an order → `POST /orders/{id}/cancellations` or `PATCH /orders/{id}` with `{ "status": "canceled" }`.
- Search → `GET /orders?...` (or `POST /orders/search` only when the query is too large/complex for a query string).
- Custom methods, when truly unavoidable, use a clearly marked colon form per [Google AIP-136](https://google.aip.dev/136) (`/orders/{id}:cancel`) — last resort.

## Status codes

Use the narrowest code that's true. The status is the outcome; never return `200` with an error inside.

**Success:** `200 OK` (read / update with body) · `201 Created` (+ `Location` header pointing at the new resource) · `202 Accepted` (async accepted, not yet done) · `204 No Content` (success, nothing to return).

**Client errors:** `400` malformed request · `401` unauthenticated · `403` authenticated but not allowed · `404` not found (or to hide existence) · `409` conflict (state clash) · `412` precondition failed (`If-Match` mismatch) · `422` semantically invalid (well-formed but fails business rules) · `429` rate-limited (+ `Retry-After`).

**Batch:** `207 Multi-Status` when a single request contains per-item outcomes.

**Server errors:** `5xx` are *server* faults only. Never blame the client with a `5xx`, and never hide a server fault behind a `2xx`/`4xx`.

## Pagination — design it in from the start

Every collection endpoint paginates from day one. Retrofitting it later breaks every client (see `versioning-and-compatibility.md`).

- **Cursor (keyset) preferred over offset.** Cursors are stable under concurrent inserts and scale; offset (`?offset=N`) drifts and degrades on large tables. ([Zalando 159](https://opensource.zalando.com/restful-api-guidelines/#159))
- **Shape:** request `page_size` + `page_token`; response returns items plus `next_page_token` (or a `next` hypermedia link). The `next` link/token must carry forward all the parameters (filters, sort) needed to continue. ([Google AIP-158](https://google.aip.dev/158))
- **Bound `page_size`** with a sane default and a hard max; never let a client request unbounded data.

## Filtering, sorting, sparse fields, expansion

- **Filtering:** query params (`?status=active&user_id=...`); for complex queries consider a documented filter syntax ([Google AIP-160](https://google.aip.dev/160)).
- **Sorting:** `?sort=` as a comma list with direction prefixes — `sort=-created_at,name` (descending created, then ascending name). ([Zalando 137](https://opensource.zalando.com/restful-api-guidelines/#137))
- **Sparse fieldsets:** `?fields=id,name,email` lets clients trim the payload and avoids over-fetching.
- **Expand / embed:** `?expand=customer,line_items` inlines related resources to kill N+1 round-trips — the cure for chatty APIs. Keep the default response lean and let clients opt in.

## Caching ([RFC 9111](https://www.rfc-editor.org/rfc/rfc9111))

- **`Cache-Control`** governs freshness: `max-age=<s>`, `no-cache` (revalidate before use), `no-store` (never cache — use for anything sensitive), `public` vs `private`.
- **Validators:** return `ETag` (and/or `Last-Modified`). On the next request the client sends `If-None-Match: <etag>` (or `If-Modified-Since`); if unchanged, respond `304 Not Modified` with no body.
- The same `ETag` doubles as the concurrency token for writes (`If-Match`) — see `errors-idempotency-concurrency.md`.

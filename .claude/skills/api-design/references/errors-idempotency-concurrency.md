# Errors, idempotency & concurrency

How writes stay correct under retries, conflicts, and partial failure. Sources: [RFC 9457 (Problem Details)](https://www.rfc-editor.org/rfc/rfc9457), [Stripe idempotency](https://stripe.com/docs/api/idempotent_requests), [Azure REST Guidelines](https://github.com/microsoft/api-guidelines/blob/vNext/azure/Guidelines.md), [Google AIP-151 (long-running operations)](https://google.aip.dev/151).

## Errors — RFC 9457 problem+json

Return a single, consistent error envelope with content type `application/problem+json`:

```json
{
  "type": "https://api.acme.com/errors/insufficient-funds",
  "title": "Insufficient funds",
  "status": 402,
  "detail": "Wallet 9f3 has balance 10.00, required 25.00.",
  "instance": "/wallets/9f3/charges/abc",
  "code": "INSUFFICIENT_FUNDS",
  "errors": [{ "field": "amount", "code": "TOO_LARGE" }],
  "traceId": "0af7651a..."
}
```

- `type`, `title`, `status`, `detail`, `instance` are the standard members; `code`, `errors[]`, `traceId` are extension members.
- **Carry a stable, machine-readable `code`** clients can branch on — `title`/`detail` are human-facing and may change wording.
- The **HTTP status must match the body's `status`**; never `200` with a problem body.
- **Clients ignore unknown extension members** (tolerant reader).
- **Never leak stack traces, SQL, internal class names, or file paths.** Log those server-side keyed by `traceId`; return only the public envelope.

## Idempotency

- **Safe** = no side effects (GET). **Idempotent** = repeating == doing once (PUT, DELETE).
- **POST is non-idempotent**, so a client that times out and retries can double-create. Fix with an **`Idempotency-Key`** header (client-generated unique key):
  - On first receipt, process and **store the response keyed by the key** (with a TTL).
  - On a retry with the same key, **replay the stored response** instead of re-executing — including replaying a stored *failure*, so the client sees the same outcome.
  - Scope keys per endpoint/account; reject reuse of a key with a different request body.

## Optimistic concurrency

Prevent lost updates without locking:

1. GET returns the resource with an `ETag`.
2. The client sends the mutation with `If-Match: <etag>`.
3. If the resource changed since (ETag mismatch) → **`412 Precondition Failed`**; the client re-reads and retries. If matched, apply.

Require `If-Match` on updates to resources where concurrent edits are possible; a missing precondition can be rejected with `428 Precondition Required`.

## Retries

- Clients retry **idempotent** requests (and POSTs carrying an idempotency key) with **exponential backoff + jitter** to avoid thundering herds.
- **Honor `Retry-After`** on `429`/`503`.
- Servers should be defensive: assume any request may arrive more than once.

## Long-running operations

When work can't finish within the request:

- **`202 Accepted`** + a **status monitor**: return a `Location`/operation resource the client polls (`GET /operations/{id}` → `pending|done|failed` with the result/error). ([Google AIP-151](https://google.aip.dev/151))
- Or **webhooks** for push notification of completion.

## Webhooks

- **Sign the payload** (HMAC over the raw body + timestamp); the receiver **verifies the signature with a timing-safe compare** and rejects stale timestamps (replay protection).
- **Put no sensitive data in the payload** — send IDs and let the receiver fetch over the authenticated API.
- **Deliver at-least-once**; receivers must be **idempotent** (dedupe on event ID) and tolerate out-of-order, duplicate, and retried deliveries.
- Respond `2xx` fast and process asynchronously; the sender retries on non-`2xx` with backoff.

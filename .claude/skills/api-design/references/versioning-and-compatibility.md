# Versioning & backward compatibility

Backward compatibility is the core discipline of API design: you ship a contract and then evolve it without breaking the consumers who already depend on it. Sources: [Google AIP-180](https://google.aip.dev/180), [Stripe API upgrades](https://stripe.com/docs/upgrades), [Zalando 106/107/189](https://opensource.zalando.com/restful-api-guidelines/#106).

## Versioning strategies (and their tradeoffs)

| Strategy | Example | Pros | Cons |
|----------|---------|------|------|
| URI path | `/v1/orders` | Visible, easy to route/cache, trivial to test in a browser | Couples version to URL; "version" of the whole surface |
| Header / media-type | `Accept: application/vnd.acme.v2+json` | Keeps URLs stable; per-representation versioning | Harder to test/debug; caching needs `Vary` |
| Query param | `/orders?api-version=2` | Simple, explicit | Easy to omit; muddies caching |
| Date-based | `Stripe-Version: 2024-06-20` | Fine-grained, per-account pinning; lets you ship many small breaking changes | Needs server-side version translation infra |

Pick one and use it consistently. URI-path versioning is the safe default for most teams; date-based pinning (Stripe model) is worth the infrastructure only when you change the contract often and must not force synchronized client upgrades.

## Backward-compatible (additive / safe) changes

These do **not** require a new version, provided clients follow the tolerant-reader rule:

- Add a **new** resource or endpoint.
- Add a **new OPTIONAL** request parameter or field (with a safe default).
- Add a **new field to a response**.
- Reorder fields in a JSON object (order is not significant).
- Add a value to a **request-only** enum that the server accepts (not to a response enum clients switch on — that can surprise them).

## Breaking changes (require a new version)

- **Remove or rename** any field, parameter, endpoint, or enum value.
- **Change a field's type** (`string` → `number`, scalar → array, `null`ability).
- **Change a default** value or default behavior.
- **Change the format** of a value (date format, ID scheme, units, precision).
- **Move a field into or out of a `oneof`** / make a previously-optional field required.
- **Tighten validation** so previously-accepted input is now rejected.
- **Add pagination to a previously-unpaginated list endpoint** — clients that read the whole array now silently miss data. This is the classic "looks additive but isn't" trap.

When in doubt, ask: *could a correctly-built existing client misbehave after this change?* If yes, it's breaking.

## Tolerant-reader rule

Both sides cooperate so additive changes stay additive:

- **Clients must ignore unknown fields** and unknown enum values rather than erroring or crashing.
- **Clients must not assume field order** or that the set of fields is closed.
- **Servers must not require** fields/params they didn't require before.

> The same forward/backward-compatibility discipline applies to **anything you persist or publish**, not just HTTP responses — stored rows/JSON, event and message payloads, queue jobs, cache entries. Old and new code versions run at the same time, so writers must keep old readers working and readers must tolerate fields they don't recognize. Some serialization formats (e.g. Avro, Protocol Buffers) build these evolution rules in for you; with plain JSON you apply the same discipline by hand. The discipline is what matters — the formats are just examples, not a requirement.

Document this expectation in your API docs; it's what makes additive evolution possible.

## Deprecation policy

Never silently break or remove. When retiring a field or endpoint:

1. **Announce** in docs and changelog with a replacement and a timeline.
2. **Signal in responses** with the [`Deprecation`](https://www.rfc-editor.org/rfc/rfc9745) header and a [`Sunset`](https://www.rfc-editor.org/rfc/rfc8594) header carrying the removal date; optionally a `Link` to the migration guide.
3. **Honor a published support window** (e.g. a minimum N months) so consumers can migrate.
4. **Monitor usage** of the deprecated surface; only remove after it falls to zero (or the window expires with notice).

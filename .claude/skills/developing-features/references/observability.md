# Observability — structured logging

Logs are an interface, not a debug aid. They are read by humans during incidents and by tools during normal operation. Design them.

---

## What to log

Add a structured log at every **decision point** — anywhere the code branches based on input, time, or external state — and at every **boundary** — anywhere the code crosses into a different process, service, or storage layer.

Each log entry has, at minimum:

- **Operation name** — what the code is doing (`user.create`, `payment.charge`, `webhook.verify`).
- **Relevant IDs** — user ID, request ID, correlation ID, resource ID. Whatever lets a future investigator follow the breadcrumb.
- **Outcome** — `success`, `failure`, `skipped`, plus the reason on failure.
- **Duration** if the operation is non-trivial (DB query, external call, expensive computation).

---

## Log levels

| Level   | When to use                                                                              |
|---------|-------------------------------------------------------------------------------------------|
| `debug` | Tracing — variable values, branch decisions. Off in production.                          |
| `info`  | Normal operations — successful business events, lifecycle transitions.                    |
| `warn`  | Recoverable anomalies — retry succeeded after failure, fallback path taken, deprecated input. |
| `error` | Operation failed and the failure is being propagated. Always actionable.                  |
| `fatal` | The process can't continue. Rare. Pairs with a crash.                                    |

`warn` is the most-abused level. If the code did its job and there's nothing to act on, it's `info`. If a human needs to look, it's `warn` or `error`.

---

## Structured format

Use the language's structured logging API. Never build log lines by string concatenation.

**Wrong:**
```python
logger.info(f"User {user_id} created order {order_id} for ${amount}")
```

**Right:**
```python
logger.info(
    "order.created",
    extra={
        "user_id": user_id,
        "order_id": order_id,
        "amount_cents": amount_cents,
        "currency": currency,
    },
)
```

Why: structured logs are queryable. `extra` becomes JSON in modern log shippers; concatenated strings require regex and break the moment a developer changes the wording.

---

## Never log

- **Passwords, tokens, API keys, session IDs.** Ever. Even at `debug`.
- **PII without redaction** — emails, phone numbers, CPF, credit card numbers, addresses. If the field is needed for debugging, log the **hash** or a **partial mask** (`emi***@example.com`).
- **Raw request bodies** when they may contain secrets. Log the shape (field names, sizes), not the content.
- **Connection strings**, even with the password masked. The host and DB name are still attack surface.

If a logging library has a `sanitize` or `redact` hook, configure it. Don't rely on every call site to remember.

---

## Don't log

- **No-ops.** If the code took no action because a precondition wasn't met, log only if the user could have expected an action. Don't fill the log with "skipping cleanup because flag is off".
- **Redundant context already in the docstring.** The log line says **what happened**, not **what the function does**.
- **Loop-internal `info` logs** that fire thousands of times per request. Aggregate (`processed 1,247 records`) instead.
- **Every entry/exit of every function.** Trace-level instrumentation belongs in `debug` or in a tracing library, not `info`.

---

## Log shape examples

**Webhook signature verification:**
```python
logger.info("webhook.received", extra={"provider": "stripe", "event_id": event_id})

if not signature_valid:
    logger.warn(
        "webhook.signature.invalid",
        extra={
            "provider": "stripe",
            "event_id": event_id,
            "reason": "timestamp_outside_tolerance",
        },
    )
    return error_response(401)

logger.info("webhook.verified", extra={"provider": "stripe", "event_id": event_id})
```

**External API call:**
```python
start = time.monotonic()
try:
    response = client.charge(...)
except PaymentProviderError as exc:
    logger.error(
        "payment.charge.failed",
        extra={
            "provider": "stripe",
            "user_id": user_id,
            "amount_cents": amount_cents,
            "duration_ms": int((time.monotonic() - start) * 1000),
            "error_code": exc.code,
        },
    )
    raise
logger.info(
    "payment.charge.succeeded",
    extra={
        "provider": "stripe",
        "user_id": user_id,
        "amount_cents": amount_cents,
        "duration_ms": int((time.monotonic() - start) * 1000),
        "charge_id": response.id,
    },
)
```

---

## Correlation IDs

Every inbound request should generate (or accept) a correlation ID. Every log entry inside that request includes the correlation ID. Every outbound call propagates it (`X-Correlation-Id` header or equivalent).

Without this, an incident investigator stitches logs by hand from timestamp and user_id — slow and error-prone.

---

## Metrics complement logs

Logs answer "what happened to this specific request". Metrics answer "what's happening across all requests".

For any operation worth logging, also emit a counter (`payment.charge.attempted`, `payment.charge.succeeded`, `payment.charge.failed`) and, if non-trivial, a histogram (`payment.charge.duration_ms`). Don't try to compute SLOs from logs.

**Measure latency by percentiles (p95 / p99), never the average.** The average hides the slow tail that users actually feel — one slow request in a hundred is invisible to the mean but is exactly what you get paged for. Set SLOs and alerts on a percentile, not the mean.

Percentiles come from production traffic — at **development time** you can't measure them, so catch latency before it ships instead: (1) read the code for the known killers (N+1, unindexed `WHERE`/`JOIN`, a query in a loop, unbounded lists, sync I/O on the hot path — see `data-layer.md`); (2) run `EXPLAIN ANALYZE` against a **realistic data volume** (problems hide at 1k rows, surface at 1M); (3) for a hot path, run a quick local load test (`k6` / `wrk` / `ab` / `hey`) to see the distribution; (4) profile if it's CPU-bound. Emit the `duration` histogram now so production hands you the real p95/p99 later.

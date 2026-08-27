# Data layer & scale

Cheap "born-scaled" defaults for any service that talks to a database. Apply them from day one — retrofitting pooling and indexes under load is an incident, not a refactor. Stack-agnostic: the idioms differ per ORM/driver, the rules don't.

## Connection pooling

- **Always use a pooled connection with explicit bounds.** Never open a connection per request/operation without a pool.
- Set them deliberately, don't accept blind defaults: **pool size**, **max overflow**, **acquire timeout**, and **recycle / pre-ping** (so a dead connection is dropped, not handed out).
- **Size the pool to the database, not the app.** Total connections across all instances × workers must stay under the DB's `max_connections`, with headroom for migrations/admin. Rough budget: `pool_size × instances × workers ≤ max_connections − reserve`.
- **Serverless / many short-lived instances → put a pooler in front** (e.g. PgBouncer in transaction mode). Per-instance pools don't bound total connections in that topology.
- One pool per process, created at startup — never per request.

## Indexes

- **Index every foreign key**, and every column used in `WHERE`, `JOIN`, `ORDER BY`, `GROUP BY`, or a uniqueness lookup. An unindexed filter is a full scan that looks fine at 1k rows and falls over at 1M.
- **Composite index order = predicate order:** equality columns first, then the range/sort column — `(tenant_id, status, created_at)` serves `WHERE tenant_id=? AND status=? ORDER BY created_at`. A composite can serve a left-prefix query; it can't serve one that skips its leading column.
- **Add the index with the feature**, in the same migration — don't wait for a slow query in prod.
- **Indexes cost writes and storage.** Don't index low-selectivity columns (e.g. a boolean) or duplicate a prefix already covered by a composite. Consider **partial** indexes (`WHERE active`) and **covering** indexes for hot read paths.
- On large/live tables, build **without locking** (Postgres `CREATE INDEX CONCURRENTLY`) in its own migration step.
- Enforce real invariants with a **unique index/constraint**, not only app-side checks.

## Query hygiene (the common scale killers)

- **No N+1.** Loading a list then a related row per item is N+1 — use eager/batch loading (a join, or a batched `IN (...)` / `selectinload`-style fetch).
- **No query inside a loop.** Hoist it to one set-based query.
- **Paginate every list from day one** (cursor preferred) — see `../../api-design/references/rest-conventions.md`.
- **Select only the columns you need** on hot paths; avoid shipping unused blobs.
- For hot counts/rankings at scale, prefer a **maintained counter/aggregate** over `COUNT(*)` per request.

## Writes & write scaling

**Cheap defaults (day one)** — the write-side twins of the read rules; they cost nothing:

- **Batch / bulk insert** — one statement/transaction for many rows, not N round-trips (one connection, one commit).
- **No write inside a loop** — accumulate and flush once.
- **Make retryable writes idempotent** (a natural/unique key or an idempotency key) so a retry can't double-insert.
- **Async only for genuine fire-and-forget** (emails, webhooks, derived/analytics events) — never for data the caller must read back immediately.

**When writes become the bottleneck — measure first, then pick.** A single relational DB handles ~**10–20k writes/s**; past that *sustained*, reach for one of these. The decisive question is **constant load vs spike** — it picks the strategy:

| Strategy | Use when | Tradeoff |
|---|---|---|
| **Queue** (Kafka / SQS / RabbitMQ) | **Spikes** (launch, viral, rush hour) — absorb the burst, workers drain at DB pace | Data isn't in the DB instantly (fine for a feed, not a balance). **Not for constant load** — the queue just grows until it falls over. |
| **Batching** (buffer, e.g. in Redis, flush periodically) | Many small inserts (counters, metrics, logs) | A small delay before it lands; works with complex/JSON rows. |
| **Sharding** | **Constant** volume above one DB's ceiling | Lose cross-shard joins; pick the **partition key from your queries**. Use a DB built for it (Cassandra / ScyllaDB / DynamoDB / ClickHouse), not vanilla Postgres. A bad key → **hot shard**. |
| **Hierarchical aggregation** (edge → regional → central) | Truly global, **purely numeric/aggregable** data (dashboard counters) | Aggregables only — no JSON/strings. Overkill for most. |

**The mistakes that signal over- (or under-) engineering:** sharding before the volume demands it (a relational DB is stronger than it looks — "not worth the complexity yet" is a valid, senior answer); a partition key chosen without knowing the queries (hot shards); using a **queue for constant write load** (it only smooths spikes). Measure your writes/s and constant-vs-spike first, add the cheapest thing that fits, and say the tradeoff out loud.

## Transactions & timeouts

- Keep transactions **short**; never hold one open across an external/network call.
- Set a **statement/query timeout** so one pathological query can't pin a connection forever.
- Rely on DB constraints for correctness (uniqueness, FKs, checks), not only app code; pick the isolation level the invariant actually needs.

### Isolation & concurrency anomalies

Most subtle data bugs are concurrency races, not logic errors. Pick the isolation level by the **anomaly you must prevent**:

- **Read Committed** (typical default) — no dirty reads, but allows non-repeatable reads and **lost updates**.
- **Snapshot / Repeatable Read** — a consistent snapshot; still allows **write skew** and **phantoms**.
- **Serializable** — prevents all of them; use it (or a lock / atomic op) when correctness depends on data you just read.

Recognize and guard the common races:

- **Lost update** — two read-modify-write on the same row; one clobbers the other. Fix with an atomic update (`SET x = x + 1`), `SELECT … FOR UPDATE`, or compare-and-set (a version column).
- **Write skew** — two transactions each read a set, both pass a check, both write, and together they break an invariant (e.g. "at least one approver on call"). Snapshot isolation does **not** catch this — use Serializable or lock the rows the decision depends on.
- **Phantom** — a row appearing/disappearing changes a query a concurrent transaction relied on.

Rule of thumb: **if a write's correctness depends on what you just read, make that read-decide-write atomic** (serializable, a row lock, or a DB constraint) — don't assume the default isolation protects it.

## What WOULD be over-engineering

Born-scaled ≠ premature distributed systems. Read replicas, sharding, a distributed cache, or CQRS come **only when a measured limit demands it**. Pooling, indexes, no-N+1, and pagination are the cheap baseline that buys you the runway to not need those yet.

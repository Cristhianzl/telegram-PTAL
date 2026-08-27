# C4 diagrams with Mermaid

Three useful levels. Use Mermaid because it renders natively in GitHub and most Markdown viewers — no external assets.

---

## Level 1: Context

Who uses the system and what external systems it talks to. The audience is Product and stakeholders. Keep it abstract — no boxes for internal services.

```mermaid
C4Context
  title System Context diagram for Orders

  Person(customer, "Customer", "Places orders via web or mobile")
  Person(ops, "Ops engineer", "Handles incidents and escalations")
  System(orders, "Orders System", "Creates, tracks, and fulfills orders")
  System_Ext(payments, "Payments Provider", "Stripe / PIX")
  System_Ext(inventory, "Inventory System", "Stock and reservations")
  System_Ext(crm, "CRM", "Customer profile, segmentation")

  Rel(customer, orders, "Places orders")
  Rel(ops, orders, "Investigates incidents")
  Rel(orders, payments, "Captures payment")
  Rel(orders, inventory, "Reserves stock")
  Rel(orders, crm, "Reads customer segment")
```

Rules:

- One `System` per box (yours). Multiple `System_Ext` for third parties.
- `Person` for human roles.
- Each `Rel` includes the verb and the data direction.
- No technology labels at Level 1 — that's Level 2.

---

## Level 2: Container

The internal pieces of your system. The audience is Product and Engineering. Each container is a separately deployable unit: API, worker, database, queue, web app, mobile app.

```mermaid
C4Container
  title Container diagram for Orders

  Person(customer, "Customer")

  Container(web, "Web App", "Next.js", "Customer-facing storefront")
  Container(api, "Orders API", "Python / FastAPI", "Exposes /orders endpoints")
  ContainerDb(db, "Orders DB", "Postgres 16", "Orders, line items, status history")
  Container(worker, "Order Worker", "Python / Celery", "Async pricing, fulfillment")
  Container_Ext(queue, "Message Queue", "RabbitMQ", "order.placed, order.shipped")

  System_Ext(payments, "Payments Provider", "Stripe / PIX")

  Rel(customer, web, "Browses and orders", "HTTPS")
  Rel(web, api, "Calls", "JSON / HTTPS")
  Rel(api, db, "Reads / Writes", "SQL / pgvector")
  Rel(api, queue, "Publishes events", "AMQP")
  Rel(worker, queue, "Subscribes to events", "AMQP")
  Rel(worker, db, "Updates status", "SQL")
  Rel(api, payments, "Captures payment", "HTTPS / official SDK")
```

Rules:

- Include the technology choice in each container (`Postgres 16`, `Python / Celery`, `Next.js`).
- `Rel` includes the protocol (`HTTPS`, `gRPC`, `AMQP`, `SQL`).
- Show data direction by where the arrow originates.
- Keep it under 12 containers — beyond that, split into multiple Level 2 diagrams by bounded context.

---

## Level 3: Component (optional)

The internal structure of one container. The audience is Engineering. Useful when the container has > 5 distinct components doing materially different things.

```mermaid
C4Component
  title Component diagram for Orders API

  Container(web, "Web App", "Next.js")
  ContainerDb(db, "Orders DB", "Postgres")
  Container_Ext(queue, "Message Queue", "RabbitMQ")

  Component(handler, "Orders Handler", "FastAPI router", "HTTP layer, request/response shaping")
  Component(service, "Orders Service", "Domain layer", "Orchestrates order placement, applies invariants")
  Component(pricing, "Pricing Service", "Domain layer", "Computes totals, taxes, discounts")
  Component(repo, "Orders Repository", "Data access", "Persists orders and events")
  Component(publisher, "Event Publisher", "Outbox pattern", "Publishes domain events to the queue")

  Rel(web, handler, "POST /orders", "JSON / HTTPS")
  Rel(handler, service, "Delegates")
  Rel(service, pricing, "Computes totals")
  Rel(service, repo, "Reads / Writes")
  Rel(service, publisher, "Emits OrderPlaced")
  Rel(repo, db, "SQL")
  Rel(publisher, queue, "AMQP")
```

Rules:

- Skip this level for features with a simple internal structure.
- Each `Component` is something a developer can grep for (a class, a module, a service object).
- Show how the components inside the container relate to the **external** containers (`web`, `db`, `queue`) at the boundary.

---

## When to skip Mermaid entirely

If the feature is purely a backend job, a refactor, or a config change with no architectural shape change, you don't need a diagram. **The diagram exists to communicate; if there's nothing new to communicate, skip it.**

Section 9 may legitimately read:

> No new architectural elements. See the [Orders System](../orders.md#9-architecture-diagrams) docs for the unchanged Context and Container diagrams.

---

## Common mistakes

- **Diagrams that lie.** Boxes that don't correspond to any actual service or table. The diagram is a contract — every box must exist.
- **Diagrams that duplicate the code.** A diagram listing every class is a class diagram, not a Component diagram. Aggregate at the right level.
- **Tech labels at Level 1.** Level 1 is for stakeholders. They don't care that it's FastAPI.
- **Missing arrows.** Every container should have at least one inbound or outbound relationship. A floating box is a smell.
- **Stale diagrams.** When the architecture changes, the diagram changes in the same PR. Stale diagrams are worse than no diagrams.

# Metrics, acceptance criteria, and prioritization

Purpose: how to define success metrics, how to separate acceptance from success criteria, and which prioritization frameworks to reference inside a PRD.

## Success metrics

- **North Star metric** — the one metric that best captures the value delivered to users. Make it **leading** (predicts future success), **influenceable** (the team can move it), and **simple** (everyone understands it). One per product, not per feature.
- **Leading vs. lagging** — leading metrics move early and predict (activation rate, time-to-first-value); lagging metrics confirm later (retention, revenue). Track both: leading to steer, lagging to confirm.
- **Guardrail / counter-metrics** — the metrics you must *not* harm while chasing the goal (latency, error rate, churn, support volume, unit cost). Without them, a team can win the target metric and lose the product.
- **Avoid vanity metrics** — numbers that look good but don't connect to value or decisions (raw pageviews, total signups that never activate, cumulative downloads). Ask: *if this number doubled, would we do anything differently?* If not, it's vanity.
- **State a target and a timeframe per metric.** "Increase activation from 32% to 45% within 90 days of GA" — not "increase activation".

## Acceptance criteria vs. success criteria

These answer two different questions and belong in **separate sections**.

| | Acceptance criteria | Success criteria |
|---|---|---|
| Question | Did we build it **right**? | Did it **move the outcome**? |
| Shape | Binary, per requirement | Metric vs. target over time |
| When | At ship / QA | After launch, measured |
| Becomes | Test cases + definition of done | The metrics review |

Shipping the feature satisfies acceptance. It does **not** prove success — only the metric, measured against its target, does.

## Writing acceptance criteria

- **Testable** — a tester (or test) can unambiguously decide pass/fail.
- **Behavior, not implementation** — describe what the user observes, not how the code does it.
- **Given / When / Then** — `Given <context>, When <action>, Then <observable outcome>`.
- **Atomic and independent** — one behavior per criterion; criteria don't depend on each other's order.
- **Cover edge, boundary, and error cases** — not just the happy path: empty input, max values, permission denied, network failure.
- They become the **test cases** and the **definition of done**. If a criterion can't become a test, rewrite it until it can.

Example:

```
Given a signed-in user with an empty cart
When they click "Checkout"
Then they see "Your cart is empty" and remain on the cart page
```

## Prioritization frameworks (reference inside a PRD)

- **RICE** — score = (Reach × Impact × Confidence) / Effort. Best for ranking a backlog of comparable items by expected return.
- **MoSCoW** — Must have / Should have / Could have / Won't have (this time). Best for scoping a single release: what's in, what's explicitly out.
- **Kano** — basic (expected; absence angers) / performance (more is better) / delighters (unexpected; drive love). Best for sanity-checking that the set actually creates user value, not just output.
- **Impact / Effort 2×2** — quick visual triage: do high-impact/low-effort first, schedule high-impact/high-effort, drop low-impact/high-effort.

**Common pairing:** use MoSCoW or Impact/Effort to **shortlist** what's in scope → RICE to **rank** the shortlist → Kano to **sanity-check** that the result delights, not just ships.

## Sources

- [Amplitude — North Star metric](https://amplitude.com/north-star)
- [Intercom — RICE prioritization](https://www.intercom.com/blog/rice-simple-prioritization-for-product-managers/)
- [ProductPlan — MoSCoW](https://www.productplan.com/glossary/moscow-prioritization/) · [Kano model](https://www.productplan.com/glossary/kano-model/)

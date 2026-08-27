---
name: documenting-features
description: Produce living, DDD-aligned feature documentation as a single Markdown file with 10 fixed sections — overview, ubiquitous language, domain model, Gherkin behaviors, ADRs, technical spec, observability, deployment & rollback, C4 diagrams, platform compatibility. Use when the user asks to document a feature, write feature docs, create ADRs, document the domain model, or produce living docs alongside the code.
license: MIT
---

# Documenting Features

Generate one Markdown file per feature, covering all 10 sections, written so Product can validate the first half and Engineering can maintain the whole thing. Lives next to the code, version-controlled, kept in sync with implementation.

## Read first (always)

List `learnings/` and read every file relevant to the current feature (the domain, the bounded context, the team). Project-specific naming conventions, ADR numbering, glossary inheritance rules, or template variants live there and override the defaults in this SKILL.md. If a learning conflicts with this file, **the learning wins** — mention it to the user.

## Tradeoff — when to apply, when to lighten up

Apply the full 10-section template for **shipped features** that cross team boundaries (Product ↔ Engineering ↔ Ops) or touch the domain model.

Lighten formality (skip sections, not principles) for: internal tooling not exposed to users; refactors with no behavior change (document the ADR only); spike work and prototypes behind feature flags. Never skip Section 1 (Overview) and Section 2 (Ubiquitous Language) — the glossary is the single source of truth for cross-team communication.

## Core principles

1. **Ubiquitous language is the single source of truth.** The same term across docs, code, database, and conversation. If the business calls it "Enrollment", the code has `class Enrollment` and the database has the `enrollments` table — no synonyms.
2. **Documentation lives next to the code.** Stored in the repo, version-controlled with the feature, kept in sync via PR review.
3. **One feature = one file.** All 10 sections in a single `.md` file. Use horizontal rules (`---`) and clear headers to separate sections. One-file-per-feature beats N-files-per-feature for navigation and synchronization.
4. **C4 model for diagrams.** Level 1 (Context) for Product/Stakeholders. Level 2 (Container) for both. Level 3 (Component) for Engineering.

## Workflow

1. **Analyze the implemented feature.** Read the production code, the tests, the PR description, the ADRs already in the repo. Identify the domain (bounded context), the actors, the inputs and outputs.
   → verify: you can name the bounded context and the related contexts.

2. **Extract the ubiquitous language.** List every domain-specific term used in code, comments, function names, table names, API endpoints. Pair each with a business definition and the code symbol it maps to.
   → verify: every term you list appears in the code with the same spelling.

3. **Identify the domain model.** Aggregates (root entity, child entities, value objects), invariants (rules that must always be true), domain events (name, trigger, payload, consumers).
   → verify: each invariant has a corresponding test or guard in the code.

4. **Translate acceptance criteria into Gherkin** (Given/When/Then) for: every happy path, every edge case, every error scenario, every permission/authorization case.
   → verify: each Gherkin scenario maps to a test in the suite.

5. **Record non-obvious design decisions as ADRs.** Each ADR has Status, Context, Decision, Consequences (Benefits, Trade-offs, Impact on Product). Number them sequentially in the repo (`docs/adr/NNNN-slug.md`) or inline within this file (`ADR-001`, `ADR-002`).
   → verify: every ADR is a decision the team had to debate — not a restatement of best practice.

6. **Document the technical spec** — dependencies, API contracts (request/response/errors), error codes, recovery actions.
   → verify: every API contract matches what the code accepts and returns.

7. **Document observability** — metrics (counter / gauge / histogram), important logs (level, event, fields), alert thresholds, dashboards.
   → verify: every metric and log named here is emitted by the code.

8. **Document deployment & rollback** — feature flags (with default and rollout strategy), database migrations (with reversibility notes), rollback plan (step-by-step), smoke tests.
   → verify: the rollback plan would actually work if executed today.

9. **Add C4 diagrams** — Context (Level 1), Container (Level 2), and Component (Level 3) where useful. Use Mermaid syntax (see `references/c4-diagrams.md`).
   → verify: every box in a diagram corresponds to something that exists.

10. **Document platform compatibility** — supported OS list, per-platform implementations (when behavior diverges), known limitations, installation instructions, CI matrix.
    → verify: every "Supported" cell in the matrix is backed by an actually-running CI lane.

11. **Capture a learning (final step).** Ask: *did I encounter a documentation convention, glossary inheritance, ADR numbering scheme, or domain quirk not in this SKILL.md or `references/`?* If yes, append a `learnings/YYYY-MM-DD-slug.md`. If no, skip.

## The 10 required sections (single file)

```
docs/
└── features/
    └── [feature-name].md    # one file, all sections below
```

| Section                              | Audience                       | Source skill / dependency       |
|--------------------------------------|--------------------------------|----------------------------------|
| 1. Overview                          | Product, Stakeholders          | PR description                   |
| 2. Ubiquitous Language Glossary      | All                            | Code review                      |
| 3. Domain Model                      | Engineering                    | Code review                      |
| 4. Behavior Specifications (Gherkin) | Product, QA, Engineering       | `developing-features-tdd` acceptance criteria |
| 5. Architecture Decision Records     | Engineering                    | PR `## Design Decisions`, ADRs   |
| 6. Technical Specification           | Engineering, Ops               | Code review                      |
| 7. Observability                     | Engineering, Ops               | `developing-features` § Observability |
| 8. Deployment & Rollback             | Engineering, Ops, Release Mgmt | Migrations, feature flags        |
| 9. C4 Diagrams                       | Engineering, Onboarding        | Architecture                     |
| 10. Platform Compatibility           | Engineering, Ops, End Users    | `ensuring-cross-platform`        |

Full template with placeholders in `references/template.md`. Diagram syntax in `references/c4-diagrams.md`. Gherkin guidance in `references/gherkin.md`.

## Section quick rules

**1. Overview** — 2–3 sentences from the business perspective; business context; bounded context; related contexts (with relationship type: Partnership, Customer-Supplier, Conformist, etc.).

**2. Ubiquitous Language Glossary** — table of `Term | Definition | Code Reference`. Every domain term that appears in the code. No synonyms — pick one and use it everywhere.

**3. Domain Model** — aggregates with root entity, child entities, value objects, invariants. Domain events with `Event | Trigger | Payload | Consumers`.

**4. Behavior Specifications (Gherkin)** — Feature / As a / I want / So that. Background, then Scenarios. One scenario per behavior (happy, edge, error, authorization). See `references/gherkin.md` for naming and structure rules.

**5. ADRs** — Status (Proposed / Accepted / Deprecated / Superseded). Context. Decision. Consequences (Benefits, Trade-offs, Impact on Product). Only write an ADR for a decision the team had to debate — not for best-practice defaults.

**6. Technical Specification** — Dependencies table. API contracts (request / success response / error response). Error handling table (`Error Code | Condition | User Message | Recovery Action`).

**7. Observability** — Metrics (`metric_name | Type | Description | Alert Threshold`). Logs (`Level | Event | Fields | When`). Dashboard links.

**8. Deployment & Rollback** — Feature flags (`Flag | Purpose | Default | Rollout Strategy`). Migrations (with reversibility). Rollback plan (numbered steps, data considerations, dependencies). Smoke tests (checkbox list).

**9. C4 Diagrams** — Mermaid `C4Context` and `C4Container` blocks. Add `C4Component` only when the feature has non-trivial internal structure.

**10. Platform Compatibility** — Supported platforms table. Per-capability platform implementation differences. Known limitations. Per-platform installation steps. CI matrix (`OS | Unit | Integration | E2E | Smoke (Docker)`).

## Hard rules

- **One feature = one file.** Don't split into multiple files. Use horizontal rules and sub-headers.
- **Documentation in English.** Regardless of conversation language.
- **No commits.** Per `writing-pull-requests` rules — print the doc in the chat or write to the planned `docs/features/<feature-name>.md` path, but never run `git commit`, `git add`, or `git push`.
- **Don't invent.** If you don't have a number for a metric threshold, mark it `TBD` and flag it to the user — don't fabricate.
- **Code references must exist.** Every `ClassName` or `method_name` cited in the glossary must be searchable in the current code.
- **Mermaid only for diagrams.** No external image links, no PlantUML, no ASCII art beyond the C4 Mermaid blocks.

## Final step: capture a learning

Before closing, ask: *did I encounter a documentation pattern, domain shape, or template adjustment not in this SKILL.md or `references/`?* If yes, append a `learnings/YYYY-MM-DD-slug.md`. If no, skip.

## Output

The final feature doc, ready to save at `docs/features/<feature-name>.md`. Print the full doc in the chat for the user to save (or save it directly only if the user has asked for "save the file" explicitly — by default, print).

## See also

- `references/communication.md` — answer-first structure (Minto Pyramid / SCQA) for every doc and summary; lead with the conclusion.
- `references/template.md` — the complete 10-section template with placeholders, ready to fill in.
- `references/c4-diagrams.md` — Mermaid syntax for Context, Container, and Component diagrams with concrete examples.
- `references/gherkin.md` — how to write Gherkin scenarios that translate to tests; common smells; coverage rules.
- `developing-features` skill — code structure and observability that feed Sections 3, 6, 7.
- `developing-features-tdd` skill — acceptance criteria and threat-model items that feed Section 4 (Gherkin scenarios).
- `ensuring-cross-platform` skill — the source of Section 10 (Platform Compatibility).
- `writing-pull-requests` skill — the PR description is the seed for Section 1 (Overview) and the source of ADR rationales for Section 5.
- `learnings/` — project-specific documentation conventions accumulated over time.

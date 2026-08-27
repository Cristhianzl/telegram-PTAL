---
name: threat-modeling
description: Structured security design analysis — find what can go wrong before an attacker does. Build a Data Flow Diagram with trust boundaries, apply STRIDE per element, add abuse/misuse cases, and map every threat to a control and a test. Use when the user asks to threat model a feature/system, run STRIDE, map the attack surface, do a security design review, write abuse cases, identify trust boundaries, ask "what could go wrong" / "what are the security risks of this feature", or draw a data flow diagram. Use at design time (before building) and again on any significant change (new data flow, new trust boundary, new third party, new auth path, new PII). A living activity, not a one-time document.
license: MIT
---

# Threat Modeling

Threat modeling is structured anticipation: you reason about how a system can be attacked **while it is still cheap to change** — at design time, on paper or in a diagram, before the code exists. The output is not a document; it is a set of mitigations wired to tests and a habit of asking "what could go wrong?" continuously.

This skill runs on **Shostack's Four Question Framework** ([Adam Shostack, *Threat Modeling: Designing for Security*](https://shostack.org/resources/threat-modeling)): What are we working on? What can go wrong? What are we going to do about it? Did we do a good job? Everything else — STRIDE, DFDs, risk ranking — serves those four questions.

## Read first (always)

List `learnings/` and read every file relevant to the current system (the domain, the trust boundaries, the third parties, the compliance regime). Project-specific assets, known attacker personas, accepted risks, and "in this system X is out of scope because Y" decisions live there and override the defaults in this SKILL.md. If a learning conflicts with this file, **the learning wins** — mention it to the user.

If `learnings/` holds only its README, proceed with the defaults below.

## Why this matters

Most vulnerabilities are design flaws, not coding bugs — a missing trust boundary, an unverified assumption, an authorization check that lives on the client. Code review and tests catch implementation defects; they rarely catch "we never decided who is allowed to do this." Threat modeling is the only activity that surfaces the flaw before the architecture sets. The [Threat Modeling Manifesto](https://www.threatmodelingmanifesto.org) frames it as a *culture*: a team that continuously finds and fixes, values a journey over a one-time artifact, and treats the model as living.

## The Four Question Framework (the workflow)

### 1. What are we working on?

Define **scope and assets**, then draw the system.

- **Assets:** what is worth protecting? Data (PII, credentials, payment data, secrets), capabilities (admin actions, money movement), and reputation/availability. If nothing is worth stealing or breaking, you have no threat model.
- **Diagram:** build a **Data Flow Diagram (DFD)** — external entities, processes, data stores, data flows — and draw **trust boundaries** where privilege or control changes (network edge, process boundary, third-party call, user↔server). The boundaries are where the threats concentrate.
- **Altitude:** scope the system at the right zoom — one feature or service, not the whole company. Too big and you boil the ocean; too small and you miss the boundary crossing that matters.

→ Full method, element types, boundary rules, and a worked example: `references/dfd-and-scope.md`.
→ verify: you can name the assets, every data flow crosses a labeled boundary or stays inside one, and the diagram fits on one screen.

### 2. What can go wrong?

Walk the diagram and **apply STRIDE per element** along each data flow and every boundary crossing.

- **STRIDE** maps each threat to the security property it violates: **S**poofing↔authentication, **T**ampering↔integrity, **R**epudiation↔non-repudiation, **I**nformation disclosure↔confidentiality, **D**enial of service↔availability, **E**levation of privilege↔authorization ([Microsoft STRIDE](https://learn.microsoft.com/en-us/azure/security/develop/threat-modeling-tool-threats)).
- **STRIDE-per-element:** not every category applies to every element type — processes face all six; data stores rarely "spoof." Use the element-type table so you ask the *right* questions and don't waste effort.
- **Abuse / misuse cases:** flip each user story into an attacker's goal ("as an attacker, I want to…"). Add attacker personas to keep the analysis grounded in who would actually try.

→ STRIDE definitions, per-element applicability, example threats and mitigations: `references/stride.md`.
→ Abuse cases, attacker personas, and other lenses (LINDDUN for privacy): `references/methodologies-and-risk.md`.
→ verify: every element and every boundary crossing has been questioned against its applicable STRIDE categories; threats are written as concrete scenarios, not category names.

### 3. What are we going to do about it?

For **each threat**, decide a response and bind it to something verifiable.

- **Response (pick one):** *mitigate* (add a control), *eliminate* (remove the feature/flow), *transfer* (push to a provider/insurer), or *accept* (document who accepted it and why). Accepting is a valid answer — silently ignoring is not.
- **Map threat → control:** every mitigated threat names a concrete control (input validation, authz check, encryption in transit, rate limit, audit log, signed token…).
- **Map control → test:** every control names a test or abuse-case that would fail if the control regressed. A control with no test will rot.
- **Track the rest:** unresolved threats become work items with an owner and a risk rank. Prioritize by **risk = likelihood × impact** (see `references/methodologies-and-risk.md` for ranking options and the criticism of DREAD).

→ verify: no threat is left in "what can go wrong" without a response; every "mitigate" has a control *and* a test reference; accepted risks name an accepter.

### 4. Did we do a good job?

Review coverage and treat the model as **living**.

- **Coverage:** did you cover every element, every boundary, every applicable STRIDE category? Did the highest-risk threats get the strongest controls?
- **Continuity:** threat modeling is not done once. Redo it on **significant change** — a new data flow, a new trust boundary, a new third party, a new auth path, new categories of data (especially PII). Keep each pass lightweight.
- **Feedback:** did findings actually become controls and tests in the codebase? A model that never changes the code did nothing.

→ verify: the model is dated, the diff from last time is clear, and follow-through (controls + tests landed) is confirmed — not just promised.

## Anti-patterns (these defeat the purpose)

- **Boiling the ocean.** Modeling the entire system at maximum detail. You will never finish and you will model things that don't matter. Scope tight, model the risky boundary.
- **One-time-only with no follow-through.** A threat model produced once, filed, and never revisited — and whose threats never became controls or tests. The Manifesto's core warning: value the journey, not the artifact.
- **Modeling without a DFD or assets.** "Brainstorming threats" with no diagram and no asset list produces a random list, not coverage. You can't reason about boundary crossings you never drew.
- **Threats with no controls or tests.** A long "what can go wrong" list and an empty "what are we going to do about it." The list is theater unless each item maps to a mitigation and a test.
- **Wrong altitude.** Modeling at the level of individual functions (too low) or the whole company (too high). Model a feature/service.

## Capture a learning

After completing a threat model, ask: *did I encounter an asset, attacker persona, trust boundary, accepted-risk decision, compliance constraint, or recurring threat pattern not already in this SKILL.md or `references/`?* If yes, append a `learnings/YYYY-MM-DD-slug.md` (protocol in `learnings/README.md`). If no, skip — don't log the model you just produced.

## See also

- `references/stride.md` — STRIDE category definitions, the security property each violates, STRIDE-per-element applicability, and example threats with example mitigations.
- `references/dfd-and-scope.md` — how to build a Data Flow Diagram (four element types + trust boundaries), how to scope assets and altitude, with a worked mini-example.
- `references/methodologies-and-risk.md` — when to reach for PASTA, Attack Trees, LINDDUN (privacy), VAST, OCTAVE; risk ranking (likelihood × impact, DREAD and its criticism, CVSS); abuse cases and attacker personas.
- `developing-features` skill — security implementation depth once threats are known; see its `references/security.md` for the controls (validation, authz, secrets, encryption) you map threats to.
- `reviewing-code` skill — the security checks that verify, at PR time, that the controls this model specified actually shipped.
- `api-design` skill — API security (authn/authz, input validation, rate limiting) for threats found on API trust boundaries.
- `learnings/` — project-specific assets, attacker personas, accepted risks, and recurring threat patterns.

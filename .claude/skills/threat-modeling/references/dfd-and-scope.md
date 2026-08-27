# Data Flow Diagrams, trust boundaries, and scope

A threat model needs a picture. The Data Flow Diagram (DFD) is the canonical one: it shows how data moves through the system and — crucially — where it crosses a **trust boundary**. Threats cluster at boundaries, so the diagram is what makes coverage possible ([OWASP Threat Modeling](https://owasp.org/www-community/Threat_Modeling); [OWASP Threat Modeling Process](https://owasp.org/www-community/Threat_Modeling_Process)).

## The four element types

A DFD uses exactly four kinds of node, plus boundaries drawn over them:

1. **External entity** — an actor or system outside your control that interacts with the system: an end user, an admin, a third-party API, an upstream service. You cannot change its behavior; you can only decide how much to trust it. (Conventionally drawn as a rectangle.)
2. **Process** — anything that acts on data: a service, a web app, a function, a worker. Processes are where authentication, authorization, and logic live, so they face every STRIDE category. (Circle.)
3. **Data store** — anywhere data rests: a database, a file, a cache, a message queue, a bucket, a config store. (Two parallel lines / open-ended rectangle.)
4. **Data flow** — the movement of data between the above: a request, a response, a message, a stream, a file read. Every flow has a direction and carries something worth labeling. (Arrow.)

## Trust boundaries

A **trust boundary** is any place where the level of privilege, control, or trust changes — where data or a request moves from a less-trusted zone to a more-trusted one (or vice versa). Draw a boundary wherever:

- **Network changes** — internet ↔ DMZ ↔ internal network; public endpoint ↔ private VPC.
- **Process / privilege changes** — user-mode ↔ kernel; unprivileged service ↔ privileged service; tenant A ↔ tenant B.
- **A third party is involved** — your service ↔ a payment provider, an identity provider, an external API. You control one side, not the other.
- **The user is on the other side** — browser/mobile client ↔ your server. Anything past this boundary toward the client is attacker-controllable and cannot be trusted.

The rule that makes the diagram pay off: **every data flow that crosses a trust boundary is a place to apply STRIDE.** A flow that stays entirely inside one zone is lower priority. Boundaries turn a vague "analyze the system" into a finite list of crossings to question.

## How to scope

- **Assets first.** List what is worth protecting: PII, credentials, payment data, secrets/keys, money-movement capability, admin actions, availability of the service. The assets justify the effort and rank the threats.
- **In / out of scope.** Write down explicitly what you are *not* modeling this pass (e.g., "the corporate SSO provider is trusted and out of scope; we model only our integration with it") so the boundary is a decision, not an omission.
- **Right altitude.** Model one feature or service at a meaningful zoom — small enough to finish in a session, large enough to contain at least one interesting trust boundary. Too high (the whole platform) boils the ocean; too low (a single function) misses the boundary crossings. If the diagram doesn't fit on one screen, it's probably too big.

## Worked mini-example (in prose)

Consider a "reset password" feature.

- **External entity:** the user (untrusted — anyone can hit the endpoint, including someone who isn't the account owner).
- **Process:** the web/API service that handles the reset request.
- **Data store:** the user database (where the email and password hash live) and a token store for reset tokens.
- **Data flows:** user → service (the reset request carrying an email); service → token store (write a reset token); service → email provider (send the reset link); user → service (submit the new password with the token).
- **Trust boundaries:** one between the user and the service (internet edge — everything from the user is attacker-controllable); one between the service and the email provider (a third party you don't control).

Now walk the boundary crossings with STRIDE:
- **Spoofing** at the user boundary — can an attacker request a reset for someone else's email? (Mitigation: don't reveal whether the email exists; rate-limit.)
- **Information disclosure** on the email flow — does the response or timing reveal whether an account exists? Is the reset link logged anywhere?
- **Tampering / Elevation** on the token submit flow — is the token unguessable, single-use, expiring, and bound to the account? Could a tampered token reset another user's password? (Mitigation: cryptographically random, short-lived, single-use, server-side validated tokens.)
- **Denial of service** at the user boundary — can an attacker flood reset emails to harass a victim or exhaust the email quota? (Mitigation: rate limit per account and per IP.)
- **Repudiation** — is the reset action logged with actor, time, and source for audit?

Five elements and two boundaries produced a concrete, prioritized threat list — and each threat already points at a control. That is the diagram doing its job.

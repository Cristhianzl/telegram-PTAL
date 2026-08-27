# STRIDE — categories, properties, examples, per-element

STRIDE is a mnemonic for six categories of threat, each the violation of one security property. It was created at Microsoft and popularized by Adam Shostack ([Microsoft STRIDE](https://learn.microsoft.com/en-us/azure/security/develop/threat-modeling-tool-threats); [Shostack, *Threat Modeling: Designing for Security*](https://shostack.org/resources/threat-modeling)). Use it as a checklist so you ask the right question at every element and boundary, instead of relying on inspiration.

## The six categories

| Category | Violates | The question to ask |
|---|---|---|
| **S**poofing | Authentication | Can someone pretend to be a user, service, or component they are not? |
| **T**ampering | Integrity | Can someone modify data in transit, at rest, or in memory? |
| **R**epudiation | Non-repudiation | Can someone deny having done an action, and could we prove otherwise? |
| **I**nformation disclosure | Confidentiality | Can someone read data they should not? |
| **D**enial of service | Availability | Can someone make the system unavailable or unusably slow? |
| **E**levation of privilege | Authorization | Can someone do something they are not authorized to do? |

## Per category — example threats and example mitigations

### Spoofing (authentication)
- **Threats:** stolen/guessed credentials; forged identity tokens; phishing; replayed session cookies; one service impersonating another; spoofed source addresses.
- **Mitigations:** strong authentication (MFA); signed and short-lived tokens (verify signature, issuer, audience, expiry); mutual TLS between services; anti-replay nonces; secure session management; don't trust client-asserted identity.

### Tampering (integrity)
- **Threats:** modified request payloads or query params; altered data in a store; cache poisoning; tampered configuration or supply-chain artifacts; man-in-the-middle changes.
- **Mitigations:** input validation and canonicalization; integrity checks (HMAC, checksums, digital signatures); TLS for data in transit; parameterized queries; write-access controls on stores; signed artifacts and dependency pinning.

### Repudiation (non-repudiation)
- **Threats:** a user denies performing an action; no record of who did what; logs that can be edited or deleted by the actor.
- **Mitigations:** tamper-evident audit logs with secure timestamps and actor identity; append-only / centralized logging the actor can't alter; digital signatures on critical actions; correlation IDs.

### Information disclosure (confidentiality)
- **Threats:** unencrypted data in transit or at rest; verbose error messages leaking internals; over-broad API responses; secrets in logs or source; insecure direct object references exposing other users' data.
- **Mitigations:** encryption in transit (TLS) and at rest; least privilege on data access; field-level minimization (return only what's needed); generic error messages to clients; secrets in a manager, never in code or logs; scrub PII from logs.

### Denial of service (availability)
- **Threats:** flooding an endpoint; expensive queries / algorithmic complexity attacks; resource exhaustion (connections, memory, disk); unbounded uploads or recursion; a slow third party blocking the request thread.
- **Mitigations:** rate limiting and quotas; timeouts on every external call; circuit breakers and bulkheads; input size limits and pagination; autoscaling and load shedding; CDN / upstream DDoS protection.

### Elevation of privilege (authorization)
- **Threats:** missing authorization check; client-side-only access control; IDOR (acting on another user's resource by changing an ID); path traversal; privilege escalation via injection.
- **Mitigations:** server-side authorization on every request; least privilege and deny-by-default; object-level (not just route-level) authz checks; validated and sandboxed inputs; separation of duties for sensitive operations.

## STRIDE-per-element

Not every category applies to every element type. Use the element type (from `dfd-and-scope.md`) to scope the questions you ask — this is "STRIDE-per-element" ([Shostack](https://shostack.org/resources/threat-modeling)). An `X` means the category is typically worth analyzing for that element type.

| Element type | S | T | R | I | D | E |
|---|---|---|---|---|---|---|
| **External entity** (user, third-party system) | X |  | X |  |  |  |
| **Process** (service, function, app) | X | X | X | X | X | X |
| **Data store** (DB, file, cache, queue) |  | X | X* | X | X |  |
| **Data flow** (request, message, stream) |  | X |  | X | X |  |

\* Repudiation applies to data stores chiefly when the store *is* the log/audit trail — tampering with or deleting log records.

How to read it: **processes face all six** (they authenticate, authorize, can be flooded, can leak). **Data flows** are mostly tampering / disclosure / DoS (intercept, modify, block). **Data stores** are tampering / disclosure / DoS, plus repudiation for log stores. **External entities** are spoofing / repudiation — you can't tamper with a user, but a user can be impersonated and can deny an action.

Walk the DFD element by element; for each, ask only the categories marked above, written as concrete scenarios for *this* system. That is the difference between coverage and a brainstorm.

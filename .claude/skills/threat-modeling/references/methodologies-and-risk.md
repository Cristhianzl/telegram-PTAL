# Other methodologies, risk ranking, and abuse cases

STRIDE is the default lens, but it isn't the only one. Pick the method that matches what you're worried about, then rank the threats so mitigation effort goes where the risk is. Sources: [OWASP Threat Modeling](https://owasp.org/www-community/Threat_Modeling); [Threat Modeling Manifesto](https://www.threatmodelingmanifesto.org); [LINDDUN](https://linddun.org); [Adam Shostack's resources](https://shostack.org/resources/threat-modeling).

## When to reach for another method

- **PASTA** (Process for Attack Simulation and Threat Analysis) — a 7-stage, risk- and business-centric process: define objectives, define technical scope, decompose the application, analyze threats, analyze vulnerabilities, model attacks, analyze risk & impact. Use when stakeholders need threats tied to **business impact** and you have time for a heavyweight, evidence-driven analysis.
- **Attack Trees** — model a single attacker **goal** at the root, decompose into sub-goals and concrete attack steps as branches (AND/OR nodes). Use when you want to explore *how* a specific high-value goal ("steal the signing key", "drain an account") could be reached, and to reason about the cheapest path.
- **LINDDUN** — the privacy counterpart to STRIDE. Seven categories: **L**inkability, **I**dentifiability, **N**on-repudiation, **D**etectability, **D**isclosure of information, **U**nawareness, **N**on-compliance ([linddun.org](https://linddun.org)). Use when the system handles personal data and the risk is to **privacy** (re-identification, profiling, lack of consent, regulatory non-compliance) rather than classic security.
- **VAST** (Visual, Agile, and Simple Threat modeling) — designed to **scale across an org** and integrate with Agile/DevOps using application and operational threat models. Use when you need threat modeling embedded in many teams' pipelines, not a one-off expert exercise.
- **OCTAVE** (Operationally Critical Threat, Asset, and Vulnerability Evaluation) — an organizational, **risk-management-centric** approach focused on operational/strategic risk to critical assets. Use for enterprise-level risk assessment rather than per-feature design review.

Most teams do well with STRIDE for security plus LINDDUN when privacy is in play, escalating to PASTA/attack trees for high-stakes targets.

## Ranking risk

Threats are not equal; rank them so you mitigate the worst first.

- **Likelihood × Impact** — the simple, defensible default. Estimate how likely a threat is to be exploited and how bad it would be, multiply (or use a small qualitative matrix: Low/Medium/High each). Good enough for most design reviews and easy to explain.
- **DREAD** — Damage, Reproducibility, Exploitability, Affected users, Discoverability, each scored and summed. **Use with caution:** DREAD is widely criticized for subjectivity — scores vary wildly between raters and the numbers imply false precision. If you use it, agree on a rubric per dimension and treat the total as a rough ordering, not a measurement.
- **CVSS** — the Common Vulnerability Scoring System. Designed for **known, concrete vulnerabilities** (a specific CVE-like finding), not abstract design threats. Use it to rank vulnerabilities found in implementation/scanning, not the speculative threats from a design-time model.

Whatever scale you use, the point is the same: **prioritize mitigations by risk.** The highest-likelihood, highest-impact threats get the strongest controls and the soonest work items; low/low threats may be accepted and documented.

## Abuse cases and attacker personas

Threats become concrete when you frame them from the attacker's side.

- **Abuse / misuse cases** — for each user story, write its inverse: "As an attacker, I want to … so that …". A login story spawns "as an attacker, I want to brute-force credentials"; a file-upload story spawns "as an attacker, I want to upload an executable / a zip bomb / a file that overwrites another user's data." Abuse cases feed directly into "what can go wrong" and become adversarial tests in "what are we going to do about it."
- **Attacker personas** — name who would realistically attack: an opportunistic script-kiddie, a malicious insider, a competitor, a motivated nation-state, an abusive user targeting another user. Each persona has different capability and motivation, which calibrates **likelihood** in the risk ranking — a threat only a nation-state can pull off ranks differently from one any anonymous user can.

Keep abuse cases tied to assets: an abuse case that doesn't threaten a real asset is noise.

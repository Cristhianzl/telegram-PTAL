# PR templates

Variants for different PR shapes. Pick the closest, adapt the rest. All output goes in the chat — never write a file.

---

## Feature

```
TITLE: feat(<scope>): <Add|Enable|Support> <capability> for <user/system>

COMMIT MESSAGE: feat(<scope>): <Add|Enable|Support> <capability> for <user/system>

DESCRIPTION:
## Objective
<One sentence: who can now do what, or what's now possible>

## Changes
- <Add the public surface — endpoint, command, UI element>
- <Wire to the underlying domain logic>
- <Add tests covering the happy path and at least one edge case>

## Notes
<Feature flag name, rollout plan, follow-up tickets — omit heading if empty>
```

---

## Bug fix

```
TITLE: fix(<scope>): Resolve <symptom> when <condition>

COMMIT MESSAGE: fix(<scope>): Resolve <symptom> when <condition>

DESCRIPTION:
## Objective
<One sentence: the user-visible defect being corrected>

## Changes
- Add failing test reproducing <condition>
- <Minimum code change that makes it pass>
- <Any guard added to prevent recurrence>

## Notes
- Root cause: <one line — e.g., race between A and B, off-by-one in pagination>
- Affected versions: <if known>
```

---

## Refactor (behavior-preserving)

```
TITLE: refactor(<scope>): <Outcome — readability, decoupling, perf-neutral cleanup>

COMMIT MESSAGE: refactor(<scope>): <Outcome>

DESCRIPTION:
## Objective
<One sentence: what's structurally better, and why now>

## Changes
- <Structural change, named at the shape level, not line level>
- Existing tests still pass; no new tests required (or: <test added to lock in the new boundary>)

## Notes
<Performance characteristics if relevant; otherwise omit>
```

---

## Performance

```
TITLE: perf(<scope>): <Reduce|Improve> <metric> by <approach>

COMMIT MESSAGE: perf(<scope>): <Reduce|Improve> <metric> by <approach>

DESCRIPTION:
## Objective
<One sentence with a number: e.g., "Cut /search p95 from 480ms to 90ms.">

## Changes
- <The change that drives the improvement>
- <Benchmark or load test added/updated>

## Notes
- Before / after numbers (with method): <link or inline>
- Tradeoffs: <memory ↑? cache invalidation cost? complexity?>
```

---

## Test-only

```
TITLE: test(<scope>): <Cover|Add tests for> <area>

COMMIT MESSAGE: test(<scope>): <Cover|Add tests for> <area>

DESCRIPTION:
## Objective
<One sentence: which behavior is now covered and why it mattered>

## Changes
- <New tests added — by behavior, not by file>
- <Fixtures, helpers, or test infra changes if any>

## Notes
<Coverage delta if measured; otherwise omit>
```

---

## Docs-only

```
TITLE: docs(<scope>): <Document|Clarify|Correct> <topic>

COMMIT MESSAGE: docs(<scope>): <Document|Clarify|Correct> <topic>

DESCRIPTION:
## Objective
<One sentence: which gap or error is closed>

## Changes
- <Doc additions / corrections by topic>
```

---

## Chore (build / CI / deps)

```
TITLE: chore(<scope>): <Bump|Pin|Replace> <thing>

COMMIT MESSAGE: chore(<scope>): <Bump|Pin|Replace> <thing>

DESCRIPTION:
## Objective
<One sentence: motivation — security advisory, compatibility, removal of unmaintained dep>

## Changes
- <Version moves or config changes>
- <Anything CI-visible: workflow rename, runner image bump, etc.>

## Notes
- Breaking changes from the upstream changelog that affect us: <list or "none">
```

---

## Revert

```
TITLE: revert: <original title>

COMMIT MESSAGE: revert: <original title>

DESCRIPTION:
## Objective
Roll back <commit-sha / PR #> because <observed failure or business decision>.

## Changes
- Revert <PR #> in full
- <Any forward-fix or follow-up linked here>

## Notes
- Root cause from incident: <one line>
- Re-apply plan: <ticket / conditions before re-attempt>
```

---

## Breaking change

When the change breaks a public contract (API, CLI flag, config schema, DB column with consumers), use the matching type **and** add `!` after the scope:

```
TITLE: feat(api)!: Replace /v1/users/me with /v2/users/me

COMMIT MESSAGE: feat(api)!: Replace /v1/users/me with /v2/users/me

DESCRIPTION:
## Objective
<One sentence on the new behavior>

## Changes
- <The breaking change, named precisely>
- <Migration path / deprecation shim if any>

## Notes
- BREAKING CHANGE: <exact consumer impact — what stops working, when>
- Migration steps: <numbered, ≤5>
- Deprecation timeline: <e.g., shim removed in v3.0, scheduled YYYY-MM>
```

The `!` is what most release-note tools look for to flag a major bump.

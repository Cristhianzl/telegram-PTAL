---
name: reviewing-code
description: Review a pull request and produce a GitHub-comment-shaped review document with severity-labeled findings (Blocker / Important / Recommended / Nice-to-have) plus a copy-paste-safe action checklist for the author. Use when the user asks to review a PR, "code review", "review this diff", "verify the changes before merge", or asks for a second opinion on someone else's branch. Output is the review comment in chat — never run git or post to GitHub directly.
license: MIT
---

# Reviewing Code

Produce a PR review optimized for posting as a single GitHub comment. The review applies a security lens, a comprehension audit, and a structural check, then labels findings by severity with a copy-paste-safe action checklist.

## Read first (always)

List `learnings/` and read every file relevant to the current PR (the touched modules, frameworks, or risk areas). Project-specific review conventions, severity adjustments, banned patterns, or scope rules live there and override the defaults in this SKILL.md. If a learning conflicts with this file, **the learning wins** — mention it to the user.

## Tradeoff — when to apply, when to lighten up

Apply the full discipline (security lens + comprehension audit + checklist + grep) for **production-bound PRs** touching shared services, payments, auth, user data, AI runtime, or anything externally observable.

Lighten formality for: docs-only changes, lockfile bumps with no API change, single-line typo fixes, internal-tooling-only changes behind a feature flag not yet enabled. Still apply the security lens — even doc PRs can leak secrets in examples.

## Hard rules — output is a chat message, not a side effect

Apply the same restrictions as the `writing-pull-requests` skill, plus review-specific ones:

- **Never** run `gh pr review`, `gh pr comment`, or any GitHub-mutating command. The user posts the review themselves.
- **Never** run `git commit`, `git add`, `git push`. Only the human commits.
- **Never** write the review to a file. Output goes in the chat.
- **Never** use `#N` to label findings — GitHub auto-links it to PR/issue `N`. Use `B1`, `I2`, `R3`, etc.
- **Never** `@mention` users unless the user explicitly asks.
- **Never** include `Fixes #N`, `Closes #N`, `Resolves #N` — they auto-close issues on merge.
- **Never** paste full file contents — link with `path/to/file.ts:42` and quote 3-10 relevant lines max.
- **Never** use absolute local paths (`/Users/...`, `/home/...`). Use repo-relative paths only.
- All review documents are written **entirely in English**, regardless of the conversation language.

## Workflow

1. **Read the PR end-to-end before commenting.** Files, diffs, test changes, and the PR description. If the description claims something not in the diff, that's a finding by itself.
   → verify: you can summarize the PR in 1–3 sentences and name the dominant risk area (auth, payments, AI, infra, UI, refactor, etc.).

2. **Apply the security lens** to every significant block (see `references/security-checks.md`). Five questions: what is being trusted unverified? what's the authoritative source for this behavior? what happens on the failure path? who controls each value and could they lie? what is the blast radius if this is wrong?
   → verify: every "unverified assumption" has been mapped to either a finding or a defensible justification in the diff.

3. **Apply the comprehension audit.** For every non-trivial block: can the author summarize it in 1–3 sentences? Is the *why* captured durably (test name, `Why:` comment, PR `## Design Decisions`, ADR)? Or is the block "AI wrote it, I trust it"?
   → verify: every non-obvious choice has a durable anchor for its rationale. Red flags trigger a finding (see `references/security-checks.md` § Comprehension Audit).

4. **Apply the structural checks** (see `references/structural-checks.md`). File-structure hard limits (500 LOC / 5 different-prefix functions / 10 same-prefix / 1 main class / cyclomatic ≤ 10 / nesting ≤ 4). SOLID. Pragmatic principles (DRY, KISS, YAGNI, Demeter). Architecture / layer separation.
   → verify: every violation is labeled by severity; "would a senior call this overengineered/duplicated?" is your gut check.

5. **Apply the platform-agnostic checks.** Paths, encoding, shell, temp/config dirs, line endings, time, CI matrix. Full grep recipes in `references/grep-recipes.md`. See `ensuring-cross-platform` for the rationale.
   → verify: no Windows-only bug ships through this PR.

6. **Apply the correctness checks** (see `references/correctness-checks.md`). The third class of bug: code that is right on the happy path and silently wrong on an edge/error path. New enum member → audit every enumeration site. Control-flow/signal exceptions (pause/cancel/retry) propagate **unwrapped** through every layer. Degrade-don't-crash contracts never hard-raise.
   → verify: every new state and every failure branch has a test that would fail if its line were deleted.

7. **Apply the testing checks.** Coverage gate (≥75%, target 80%) shown by the author. Tests challenge the code, not confirm it. Both happy path AND adversarial. Anti-patterns absent (Mirror, Liar, Giant, Mockery, Inspector, Chain Gang, Flaky, Snowball). See `writing-tests` for the full list.
   → verify: if you removed a line of business logic, at least one test would fail.

8. **Compile the findings** by severity (B = Blocker, I = Important, R = Recommended, N = Nice-to-have). Output format in `references/output-format.md`.
   → verify: every finding has a label, a file:line reference, a problem statement, an impact statement, and a suggested fix.

9. **Render the review** in the structure below (one `## Code Review Summary` section, severity sections in order, then the Action Checklist).
   → verify: the rendered output passes the copy-paste safety check (no `#N`, no `@mentions`, no absolute paths, length under cap).

10. **Capture a learning (final step).** Ask: *did I encounter a review pattern, codebase quirk, recurring violation, or severity adjustment not in this SKILL.md or `references/`?* If yes, append a `learnings/YYYY-MM-DD-slug.md`. If no, skip.

## Severity scoring

| Severity         | Label | Meaning                                                          |
|------------------|-------|------------------------------------------------------------------|
| Blocker          | `B1`  | Must be fixed before merge. PII in logs, security defect, file-structure violation, missing test for a high-risk path. |
| Important        | `I1`  | Preferably this PR. SOLID violations, architecture leaks, weak error handling, missing adversarial tests. |
| Recommended      | `R1`  | Can ship as a follow-up. Observability gaps, naming clarity, minor duplication. |
| Nice-to-have     | `N1`  | Polish. Idiomatic suggestions, refactor opportunities that don't affect correctness. |

Use these labels consistently throughout the review. **Never `#N` — GitHub auto-links it.**

## Required output structure

```markdown
## Code Review Summary

<2-4 sentence verdict: ship / changes-requested / blocked, and the headline reason>

**Verdict:** <Approve | Approve with comments | Request changes | Block>
**Findings:** <N blockers, M important, K recommended>

---

## ⛔ Blockers (resolve before merge)

### B1 — <Short title>
**File:** `path/to/file.ts:42-58`
**Issue:** <1-3 sentence problem statement>
**Why it matters:** <impact / blast radius>
**Suggested fix:** <concrete action>

<details>
<summary>Code reference</summary>

```ts
// quote only the 3-10 most relevant lines
```
</details>

### B2 — <Short title>
...

---

## ⚠️ Important (preferably this PR)

### I1 — <Short title>
...

---

## 💡 Recommended (can ship as a follow-up)

### R1 — <Short title>
...

---

## ✅ Action checklist for the author

**Blockers (resolve before merge):**
- [ ] **B1** — <one-line restatement>
- [ ] **B2** — <one-line restatement>

**Important (preferably this PR):**
- [ ] **I1** — <one-line restatement>

**Recommended (can ship as a follow-up PR):**
- [ ] **R1** — <one-line restatement>
```

Target length: 300–800 lines of markdown. Hard cap: 1500 lines (GitHub truncates long comments). One finding = one section.

## The five questions to ask on every PR

For every significant block of code:

1. **What is this code trusting without verifying?** — inputs, tokens, signatures, IDs, headers, flags from outside the system.
2. **What is the authoritative source for this behavior?** — is the implementation based on official docs/SDK, or on assumption/tutorial?
3. **What happens in every failure path?** — exceptions swallowed? failed checks that silently pass? errors leaking internals?
4. **Who controls each value, and could they lie?** — client-sent fields, URL params, headers are all forgeable unless verified server-side.
5. **What is the blast radius if this is wrong?** — data exposure? privilege escalation? forged events? resource exhaustion?

Unverified assumptions are where vulnerabilities live. They look like reasonable code.

## Blocker categories (zero-tolerance review)

These get a `B` label automatically — full detail in `references/security-checks.md` and `references/structural-checks.md`:

- **PII in logs.** Any email, name, phone, address in a `logger`/`print`/`console` call. Approved identifiers: `auth_id`, `user_id`, `stripe_id`, internal `id`.
- **Secrets in code.** Hardcoded API keys, tokens, passwords. Webhook secrets not in env vars.
- **DRY violation.** Duplicate types, classes, or logic (5+ lines × 2+ occurrences).
- **File structure violation.** > 500 LOC / > 5 different-prefix functions / > 10 same-prefix / > 1 main class / cyclomatic > 10 / nesting > 4.
- **Mixed responsibility prefixes in one file.** `validate*` AND `format*` in the same file = violation.
- **Comprehension debt.** Author can't defend a block. "AI wrote it, I trust it." `Why:` missing on a non-obvious choice.
- **AI security violation** (when applicable). Direct DB access from AI layer. Secrets in system prompt. Output rendered raw. No human-in-the-loop on irreversible actions.
- **AI runtime missing resilience** (when applicable). No timeout, no circuit breaker, no fallback, no kill switch, no cost ceiling.
- **Third-party integration done wrong.** Custom signature verification instead of provider spec. `==` instead of timing-safe compare. Reading payload fields before verification.
- **AI-generated code in high-risk paths unverified.** Auth, payments, signatures, crypto — every assumption needs to be checked against the authoritative source.

## Output format

The review comment, ready to paste. Print it in the chat. Don't write a file. Don't run `gh`.

Detailed rules — labeling schemes, length, length, link format, copy-paste safety check — in `references/output-format.md`.

## See also

- `../documenting-features/references/communication.md` — answer-first structure (Minto Pyramid); lead the review with the verdict, then findings by severity.
- `references/output-format.md` — GitHub-comment formatting rules, label schemes, structure skeleton, copy-paste safety check.
- `references/security-checks.md` — PII zero-tolerance, AI security, third-party integration, AI-untrusted-code lens, AI runtime resilience.
- `skills/security-review` — for a dedicated, OWASP-2025 security audit (frontend + backend) beyond the always-on blockers here.
- `references/structural-checks.md` — file structure hard limits, SOLID verification, pragmatic principles, layer separation, error handling.
- `references/correctness-checks.md` — enum/state-machine completeness, control-flow exceptions propagate unwrapped, degrade-don't-crash contracts + the failure-path test spec.
- `references/grep-recipes.md` — copy-paste grep commands for catching common violations across languages.
- `references/checklist.md` — pre-submission checklist + perfect-score criteria + framework-specific guidelines.
- `developing-features` skill — the code quality rules being reviewed; if you find yourself rebuilding the rationale, link to that skill in the finding.
- `writing-tests` skill — what "good tests" means in a review context; the 8 anti-patterns map directly to review findings.
- `ensuring-cross-platform` skill — the portability rules; platform red flags are in the structural checks here.
- `learnings/` — project-specific review conventions, recurring violation patterns, severity adjustments.
